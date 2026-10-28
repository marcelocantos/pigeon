#!/usr/bin/env python3
"""Fail-closed remote JitPack publication oracle; pass a published commit SHA.

Usage: python3 scripts/check-android-publication.py <40-character-git-sha>
The SHA must exist on GitHub already. This never tests an unpushed worktree.
"""
# Copyright 2026 Marcelo Cantos
# SPDX-License-Identifier: Apache-2.0
import io
import re
import sys
import urllib.error
import urllib.request
import xml.etree.ElementTree as ET
import zipfile

BASE = "https://jitpack.io/com/github/marcelocantos/pigeon/pigeon"
MAX_AAR = 100 * 1024 * 1024


def fetch(url, limit):
    try:
        with urllib.request.urlopen(
            urllib.request.Request(url, headers={"User-Agent": "pigeon-publication-smoke/1"}),
            timeout=120,
        ) as response:
            if response.status != 200:
                raise RuntimeError(f"HTTP {response.status}: {url}")
            data = response.read(limit + 1)
            if not data or len(data) > limit:
                raise RuntimeError(f"empty/oversized response: {url}")
            return data
    except urllib.error.URLError as exc:
        raise RuntimeError(f"remote artifact unavailable: {url}: {exc}") from exc


def check(version):
    if not re.fullmatch(r"[a-fA-F0-9]{40}", version):
        raise ValueError("expected a published full 40-character commit SHA (not a mutable branch or old tag)")
    url = f"{BASE}/{version}/pigeon-{version}"
    pom = ET.fromstring(fetch(url + ".pom", 2 * 1024 * 1024))
    def field(name):
        for child in pom:
            if child.tag.rsplit('}', 1)[-1] == name:
                return child.text
        return None
    for key, expected in (
        ("groupId", "com.github.marcelocantos.pigeon"),
        ("artifactId", "pigeon"),
        ("version", version),
        ("packaging", "aar"),
    ):
        actual = field(key)
        # Gradle POMs may omit packaging for AAR; the extension and ZIP below
        # prove the actual type, but require the rest of the coordinates.
        if key == "packaging" and actual is None:
            continue
        if actual != expected:
            raise RuntimeError(f"POM {key}: expected {expected}, got {actual!r}")
    with zipfile.ZipFile(io.BytesIO(fetch(url + ".aar", MAX_AAR))) as aar:
        names = set(aar.namelist())
        native = "jni/arm64-v8a/libpigeon-jni.so"
        if native not in names or aar.getinfo(native).file_size < 100000:
            raise RuntimeError(f"missing/empty {native}")
        if "classes.jar" not in names:
            raise RuntimeError("missing Kotlin classes.jar")
        with zipfile.ZipFile(io.BytesIO(aar.read("classes.jar"))) as classes:
            if "com/marcelocantos/pigeon/crypto/E2EChannel.class" not in classes.namelist():
                raise RuntimeError("missing E2EChannel Kotlin class")
    print(f"GREEN remote POM + arm64-v8a JNI AAR + Kotlin classes: {url}")


if __name__ == "__main__":
    try:
        if len(sys.argv) != 2:
            raise ValueError("usage: check-android-publication.py <published-40-char-sha>")
        check(sys.argv[1])
    except (ValueError, RuntimeError, ET.ParseError, zipfile.BadZipFile) as exc:
        print(f"RED Android publication: {exc}", file=sys.stderr)
        sys.exit(1)
