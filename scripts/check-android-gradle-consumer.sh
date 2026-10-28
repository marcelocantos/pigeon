#!/usr/bin/env bash
# Copyright 2026 Marcelo Cantos
# SPDX-License-Identifier: Apache-2.0
# Live Gradle resolution from a fresh consumer outside this repository.
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
sha="${1:?usage: check-android-gradle-consumer.sh <published-full-sha>}"
[[ "$sha" =~ ^[[:xdigit:]]{40}$ ]] || { echo 'expected a published full commit SHA' >&2; exit 1; }
python3 "$root/scripts/check-android-publication.py" "$sha"
consumer="$(mktemp -d)"
trap 'rm -rf "$consumer"' EXIT
printf 'rootProject.name = "pigeon-external-consumer"\n' > "$consumer/settings.gradle.kts"
cat > "$consumer/build.gradle.kts" <<'GRADLE'
plugins { base }
repositories { mavenCentral(); maven("https://jitpack.io") }
val pigeon by configurations.creating
val version = providers.gradleProperty("pigeonVersion").get()
dependencies { pigeon("com.github.marcelocantos:pigeon:$version") }
tasks.register("resolvePigeon") {
    doLast {
        val files = pigeon.resolve()
        require(files.any { it.extension == "aar" && it.length() > 0 }) {
            "Gradle did not resolve a nonempty pigeon AAR: $files"
        }
        println("GREEN independent Gradle consumer resolved: $files")
    }
}
GRADLE
"$root/android/gradlew" -p "$consumer" resolvePigeon "-PpigeonVersion=$sha" \
    --no-daemon --console=plain --refresh-dependencies
