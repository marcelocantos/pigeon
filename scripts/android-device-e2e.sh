#!/usr/bin/env bash
# Copyright 2026 Marcelo Cantos
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail
cd "$(dirname "$0")/.."
# Keep Gradle/Android metadata inside this isolated checkout; fleet sandboxes
# cannot write ~/.gradle or ~/.android. Callers can override either directory.
export ANDROID_HOME="${ANDROID_HOME:-$HOME/Library/Android/sdk}"
export GRADLE_USER_HOME="${GRADLE_USER_HOME:-$PWD/android/build/.gradle-user-home}"
export ANDROID_USER_HOME="${ANDROID_USER_HOME:-$PWD/android/build/.android-user-home}"
mkdir -p "$GRADLE_USER_HOME" "$ANDROID_USER_HOME"
serial=${ANDROID_SERIAL:-}
if [[ -z "$serial" ]]; then
  for candidate in 66211FDDJ000ZX 3419105H80658D; do
    if adb devices | grep -q "^${candidate}[[:space:]]\+device$"; then serial=$candidate; break; fi
  done
fi
[[ -n "$serial" ]] || { echo 'No Fold/Tablet device; device oracle FAIL (not skipped)' >&2; exit 1; }
adb -s "$serial" get-state | grep -qx device || { echo "Device $serial unavailable" >&2; exit 1; }
# adb reverse only forwards TCP. QUIC is UDP: discover the LAN egress address
# towards the device and give the instrumented test that reachable address.
device_ip=$(adb -s "$serial" shell ip -4 addr show wlan0 | sed -n 's/.*inet \([0-9.]*\)\/.*/\1/p' | head -1)
[[ -n "$device_ip" ]] || { echo "Device $serial has no wlan0 IPv4 address" >&2; exit 1; }
host=${PIGEON_DEVICE_HOST:-$(python3 - "$device_ip" <<'PY'
import socket,sys
s=socket.socket(socket.AF_INET,socket.SOCK_DGRAM)
s.connect((sys.argv[1],4433))
print(s.getsockname()[0])
s.close()
PY
)}
[[ "$host" != 127.* ]] || { echo 'LAN address resolved to loopback' >&2; exit 1; }
read -r quic_port http_port <<< "$(python3 - <<'PY'
import socket
ports=[]
for typ in (socket.SOCK_DGRAM,socket.SOCK_STREAM):
 s=socket.socket(socket.AF_INET,typ);s.bind(('0.0.0.0',0));ports.append(s.getsockname()[1]);s.close()
print(*ports)
PY
)"
tmp=$(mktemp -d)
relay_pid= fixture_pid=
cleanup() {
  [[ -z "$fixture_pid" ]] || kill "$fixture_pid" 2>/dev/null || true
  [[ -z "$relay_pid" ]] || kill "$relay_pid" 2>/dev/null || true
  wait "$fixture_pid" "$relay_pid" 2>/dev/null || true
  rm -rf "$tmp"
}
trap cleanup EXIT
# Compile/install before starting the fixture deadline. Missing SDK/NDK or
# vendored libraries are hard failures, not skipped device tests.
(cd android && ./gradlew -Pkotlin.compiler.execution.strategy=in-process :pigeon-host:assembleDebugAndroidTest :pigeon-host:assembleDebug)
make c/vendor/build/lib/libsodium.a
go build -o "$tmp/relay" ./cmd/pigeon
go build -o "$tmp/fixture" ./cmd/pigeon-android-fixture
# The local relay's self-signed certificate is accepted ONLY by explicit
# insecure test configuration; the production Android path verifies peers.
PIGEON_TOKEN= "$tmp/relay" -port "$http_port" -quic-port "$quic_port" >"$tmp/relay.log" 2>&1 & relay_pid=$!
# Development relay serves /health on WebTransport (HTTP/3 over UDP),
# not HTTPS over TCP. A TCP curl would fail even for a healthy relay.
# Let the fixture's real raw-QUIC registration be the readiness probe.
kill -0 "$relay_pid" 2>/dev/null || { cat "$tmp/relay.log" >&2; exit 1; }
"$tmp/fixture" -relay "https://$host:$quic_port" -control "0.0.0.0:0" >"$tmp/fixture.log" 2>&1 & fixture_pid=$!
for _ in {1..80}; do
  control=$(sed -n 's/^CONTROL=0.0.0.0:\([0-9]*\)$/\1/p' "$tmp/fixture.log" | head -1)
  [[ -z "$control" ]] || break
  kill -0 "$fixture_pid" 2>/dev/null || { cat "$tmp/fixture.log" >&2; exit 1; }
  sleep .25
done
[[ -n "${control:-}" ]] || { cat "$tmp/fixture.log" >&2; exit 1; }
token=$(curl -fsS --max-time 40 "http://127.0.0.1:$control/token")
[[ -n "$token" ]] || { echo 'Fixture returned empty pairing token' >&2; exit 1; }
kill -0 "$relay_pid" 2>/dev/null || { cat "$tmp/relay.log" >&2; exit 1; }
# The :pigeon library produces only an AAR, never a debug APK; the
# :pigeon-host application module (🎯T66.6) is the real installable
# target the instrumentation runs against.
test_apk=$(find android/pigeon-host/build/outputs/apk/androidTest -name '*.apk' -type f | head -1)
[[ -n "$test_apk" ]] || { echo 'Missing instrumentation APK' >&2; exit 1; }
target_apk=$(find android/pigeon-host/build/outputs/apk/debug -name '*.apk' -type f | head -1)
[[ -n "$target_apk" ]] || { echo 'Missing pigeon-host target APK' >&2; exit 1; }
adb -s "$serial" install -r "$target_apk"
adb -s "$serial" install -r "$test_apk"
# -e values are passed as arguments rather than interpolated into shell script.
result=$(adb -s "$serial" shell am instrument -w -r \
  -e class com.marcelocantos.pigeon.DeviceEchoTest \
  -e token "$token" -e relay "https://$host:$quic_port" \
  -e control "http://$host:$control" \
  com.marcelocantos.pigeon.host.test/androidx.test.runner.AndroidJUnitRunner) || { echo "$result"; exit 1; }
echo "$result"
grep -q '^OK (1 test)' <<< "$result" || { echo 'Device instrumentation did not pass exactly one test' >&2; exit 1; }
wait "$fixture_pid" || { cat "$tmp/fixture.log" >&2; exit 1; }
fixture_pid=
grep -q '^DEVICE_ORACLE_PASS$' "$tmp/fixture.log" || { cat "$tmp/fixture.log" >&2; exit 1; }
echo "Device oracle GREEN on $serial via $host:$quic_port"
