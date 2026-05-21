// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Package backchannel carries the five typed pairing-ceremony events
// between a short-lived CLI process and a long-lived daemon over a Unix
// domain socket on the same host.
//
// The transport is Unix-socket only (macOS and Linux); Windows support
// via named pipes is out of scope.
//
// The wire format mirrors the pigeon relay stream framing — a 4-byte
// big-endian length prefix followed by a JSON payload — so the daemon's
// event loop can plumb backchannel and relay events through the same
// dispatch path without distinguishing transport.
//
// Socket-path discovery is per-app:
//
//	$XDG_RUNTIME_DIR/pigeon/<app-name>.sock        (Linux)
//	~/Library/Application Support/pigeon/<app-name>.sock  (macOS)
//
// Permissions: parent directory 0700, socket file 0600. The daemon
// unlinks any stale socket left by a previously-crashed instance before
// binding (test-connect first, ECONNREFUSED ⇒ safe to unlink).
//
// Lifecycle:
//
//	# Daemon side
//	srv, err := backchannel.Listen(&backchannel.ListenArgs{AppName: "pairdroid"})
//	defer srv.Close()
//	for {
//	    sess, err := srv.Accept(ctx)
//	    if err != nil { return }
//	    go handle(sess) // sess.Close() on exit
//	}
//
//	# CLI side
//	cli, err := backchannel.Dial(&backchannel.DialArgs{AppName: "pairdroid"})
//	defer cli.Close()
//	_ = cli.Send(ctx, backchannel.Message{Type: backchannel.MsgPairBegin})
//	msg, err := cli.Recv(ctx)
//
// Every lifecycle type is `Close() error`, idempotent, defer-callable.
package backchannel
