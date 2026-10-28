#!/usr/bin/env bash
# Copyright 2026 Marcelo Cantos
# SPDX-License-Identifier: Apache-2.0
# Exercise a git-exported module through Go's normal zip/proxy/cache path.
# This intentionally has no c/vendor/build and does NOT use a replace directive.
set -euo pipefail
cd "$(dirname "$0")/.."
full=${1:-}
if [[ -n "$full" && "$full" != --full ]]; then echo 'usage: go-module-smoke.sh [--full]' >&2; exit 2; fi
for tool in git python3 pkg-config go; do command -v "$tool" >/dev/null || { echo "missing $tool" >&2; exit 1; }; done
pkg-config --exists libsodium || { echo 'Install libsodium development headers and pkg-config first' >&2; exit 1; }
work=$(mktemp -d)
trap 'chmod -R u+w "$work" 2>/dev/null; rm -rf "$work"' EXIT
# The archive is the tagged-source equivalent: gitlinks/submodule content and
# untracked native build outputs cannot enter the zip. Go validates the zip.
git archive HEAD | tar -xf - -C "$work"
version=v0.0.0-go-module-smoke
module=github.com/marcelocantos/pigeon
proxy="$work/proxy/$module/@v"
mkdir -p "$proxy" "$work/gopath" "$work/consumer"
python3 - "$work" "$module" "$version" <<'PY'
import os, sys, zipfile
root, module, version = sys.argv[1:]
prefix = module + '@' + version + '/'
with zipfile.ZipFile(root + '/proxy/' + module + '/@v/' + version + '.zip', 'w', zipfile.ZIP_DEFLATED) as z:
    for dirpath, dirs, files in os.walk(root):
        if dirpath.startswith(root + '/proxy') or dirpath.startswith(root + '/gopath') or dirpath.startswith(root + '/consumer'):
            dirs[:] = []
            continue
        for name in files:
            path = os.path.join(dirpath, name)
            rel = os.path.relpath(path, root)
            if os.path.islink(path) or rel.startswith('.git/'):
                continue
            z.write(path, prefix + rel)
PY
cp "$work/go.mod" "$proxy/$version.mod"
printf '{"Version":"%s","Time":"2026-01-01T00:00:00Z"}\n' "$version" > "$proxy/$version.info"
printf '%s\n' "$version" > "$proxy/list"
# Dependencies still come from the public proxy, but this module MUST come
# from the local proxy zip. GOSUMDB=off applies only to this synthetic version.
export GOWORK=off GOPATH="$work/gopath" GOMODCACHE="$work/gopath/pkg/mod"
export GOPROXY="file://$work/proxy,https://proxy.golang.org" GOSUMDB=off
export PKG_CONFIG_PATH="${PKG_CONFIG_PATH:-}"
(cd "$work/consumer" && go mod init example.com/pigeon-module-smoke >/dev/null && go get "$module@$version")
cat > "$work/consumer/consumer_test.go" <<'GO'
package consumer
import (
    "testing"
    "github.com/marcelocantos/pigeon"
    "github.com/marcelocantos/pigeon/cwire"
)
func TestPublishedModule(t *testing.T) {
    _ = pigeon.Connect
    if b := cwire.EncodeUvarint(300); len(b) != 2 { t.Fatalf("wire bridge: %x", b) }
}
GO
(cd "$work/consumer" && go test ./...)
# Check packages within the actual downloaded module, never the checkout.
moddir="$GOMODCACHE/$module@$version"
test -d "$moddir/dist" && test ! -e "$moddir/c/vendor/build" || { echo 'module archive leaked native build outputs' >&2; exit 1; }
if [[ "$full" == --full ]]; then
    (cd "$moddir" && go test ./...)
else
    (cd "$moddir" && go test ./cwire ./crypto ./c/test/csdke2e ./)
fi
