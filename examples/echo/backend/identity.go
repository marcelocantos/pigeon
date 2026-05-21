// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package main

import "github.com/marcelocantos/pigeon/crypto"

// loadIdentity returns the long-term Identity used by both the daemon and
// the `pair` subcommand. The default implementation reads/writes a JSON
// keypair under idPath; on macOS deployments swap in newKeychainIdentity
// (see keychain_identity_darwin.go) by replacing the body with:
//
//	return newKeychainIdentity("com.example.echo-backend", "default")
//
// That is the entire migration step. Existing pairings keep working
// after a one-time re-pair (see keychain_identity_darwin.go for why).
func loadIdentity(idPath string) (crypto.Identity, error) {
	return crypto.NewFileIdentity(idPath)
}
