#!/usr/bin/env bash
# Copyright 2026 Marcelo Cantos
# SPDX-License-Identifier: Apache-2.0
#
# Exercise a RELEASED, TAGGED version of this module the way an unrelated
# external consumer actually would: `go get` it from the real Go module
# proxy (proxy.golang.org, falling back to direct VCS), checksummed against
# the real public sumdb, into a private module cache that has never seen
# this checkout.
#
# This is deliberately NOT scripts/go-module-smoke.sh, which validates the
# *archive shape* of the current (possibly unreleased) checkout via a
# synthetic local file:// proxy and GOSUMDB=off. That script answers "is
# the tree's zip export well-formed"; this one answers "does the tagged
# release a stranger would `go get` actually build and pass its tests",
# with no local shortcuts:
#   - no git archive / local checkout content of any kind
#   - no file:// proxy
#   - no `replace` directive
#   - GOSUMDB left ON (real checksum verification against sum.golang.org)
#   - a private GOPATH/GOMODCACHE/GOWORK=off so nothing in $HOME or this
#     repo's own module cache can mask a failure
#
# Fails closed: refuses to run against anything that is not a real,
# existing git tag, and verifies after `go get` that the module actually
# resolved to that exact tagged version from outside the repo tree — not a
# pseudo-version, not a replace, not a path under this checkout.
set -euo pipefail
cd "$(dirname "$0")/.."
repo_root=$(pwd -P)

usage() { echo "usage: go-module-release-smoke.sh [vX.Y.Z]" >&2; exit 2; }
[[ $# -le 1 ]] || usage

version="${1:-}"
if [[ -z "$version" ]]; then
	# Default to the highest vX.Y.Z tag reachable from HEAD. Reachable,
	# not merely present, so a tag cut on an abandoned branch can't be
	# picked by accident.
	version=$(git tag --merged HEAD -l 'v[0-9]*.[0-9]*.[0-9]*' | sort -V | tail -1)
fi
[[ -n "$version" ]] || { echo 'no released vX.Y.Z tag found (reachable from HEAD)' >&2; exit 1; }
[[ "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo "not a semver tag: $version" >&2; exit 1; }
# Fail closed: the tag must be a real, existing git tag object, not merely
# a string that happens to look like one.
git rev-parse -q --verify "refs/tags/$version" >/dev/null || {
	echo "refs/tags/$version does not exist — not released yet, nothing to smoke-test" >&2
	exit 1
}

for tool in git python3 pkg-config go curl; do command -v "$tool" >/dev/null || { echo "missing $tool" >&2; exit 1; }; done
pkg-config --exists libsodium || { echo 'Install libsodium development headers and pkg-config first' >&2; exit 1; }

module=github.com/marcelocantos/pigeon
work=$(mktemp -d)
trap 'chmod -R u+w "$work" 2>/dev/null; rm -rf "$work"' EXIT
mkdir -p "$work/gopath" "$work/consumer"

# No replace, no local proxy, no vendored checkout: force the real public
# proxy (direct VCS fallback only, never a file:// anything), the real
# sumdb, and strip any ambient override that could quietly defeat either.
unset GOPRIVATE GONOSUMCHECK GOINSECURE GOFLAGS GONOSUMPATTERNS 2>/dev/null || true
export GOWORK=off
export GOPATH="$work/gopath"
export GOMODCACHE="$work/gopath/pkg/mod"
export GOPROXY='https://proxy.golang.org,direct'
export GOSUMDB=sum.golang.org
export PKG_CONFIG_PATH="${PKG_CONFIG_PATH:-}"

(
	cd "$work/consumer"
	go mod init example.com/pigeon-release-smoke >/dev/null
	go get "$module@$version"
)

# Fail closed: go.mod must record exactly the requested tagged version —
# not a different version the proxy silently substituted, and no replace
# directive sneaking the module back to this checkout or anywhere local.
if grep -q '^replace ' "$work/consumer/go.mod"; then
	echo "consumer go.mod carries a replace directive — not an honest external-consumer test" >&2
	grep '^replace ' "$work/consumer/go.mod" >&2
	exit 1
fi
got_version=$(cd "$work/consumer" && go list -m -f '{{.Version}}' "$module" 2>/dev/null || true)
[[ "$got_version" == "$version" ]] || {
	echo "resolved to $module@$got_version, expected $module@$version" >&2
	exit 1
}

moddir=$(cd "$work/consumer" && go list -m -f '{{.Dir}}' "$module")
case "$moddir" in
"$GOMODCACHE"/*) ;;
*)
	echo "module resolved outside the private module cache ($moddir) — refusing" >&2
	exit 1
	;;
esac
case "$moddir" in
"$repo_root"*)
	echo "module resolved inside this checkout ($moddir) — refusing" >&2
	exit 1
	;;
esac
# The downloaded module content must be the proxy's immutable zip for this
# tag, not anything derived from the working tree: it has no .git at all.
[[ ! -e "$moddir/.git" ]] || { echo "downloaded module carries a .git dir — not the proxy zip" >&2; exit 1; }

cat >"$work/consumer/consumer_test.go" <<GO
package consumer

import (
	"testing"

	"github.com/marcelocantos/pigeon"
	"github.com/marcelocantos/pigeon/cwire"
)

func TestReleasedModule(t *testing.T) {
	_ = pigeon.Connect
	if b := cwire.EncodeUvarint(300); len(b) != 2 {
		t.Fatalf("wire bridge: %x", b)
	}
}
GO
(cd "$work/consumer" && go test ./...)

# Archive conformance on the REAL released zip: native build outputs must
# not have leaked into the tagged release.
test -d "$moddir/dist" && test ! -e "$moddir/c/vendor/build" || {
	echo "released module $version leaked native build outputs" >&2
	exit 1
}

# Exercise the actual published API and the cwire/crypto bridges from
# inside the downloaded module tree — never the local checkout — against
# the full package test suites those packages carry at the tagged version.
# Deliberately excludes the root package's parallel QUIC E2E stress suite
# and csdke2e's untagged cgo relay oracle for the same reason
# scripts/go-module-smoke.sh does: that coverage belongs to the
# source-checkout `go test ./...` CI step, not a module-consumer smoke.
(cd "$moddir" && go test ./cwire ./crypto)

echo "go-module-release-smoke: $module@$version OK (resolved $moddir)"
