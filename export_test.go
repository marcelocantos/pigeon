// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

// Test hooks for 🎯T63 stream-lifecycle oracles.

// BufferedStreamCount is the number of unmatched incoming streams
// currently parked for AcceptStream.
func (s *Session) BufferedStreamCount() int { return s.bufferedStreamCount() }

// IsStreamReset reports whether err is a QUIC/WebTransport stream reset
// (RESET_STREAM), as opposed to a clean FIN/EOF.
func IsStreamReset(err error) bool { return isStreamReset(err) }
