// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package tern

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// BenchmarkStreamLatency measures round-trip time for a single stream message.
// The backend echoes each message back; we measure the full round trip.
func BenchmarkStreamLatency(b *testing.B) {
	for _, env := range benchRelays(b) {
		b.Run(env.name, func(b *testing.B) {
			ctx := context.Background()
			backend, client := benchPair(b, env.setup(b))

			// Backend echo loop.
			go func() {
				for {
					data, err := backend.Recv(ctx)
					if err != nil {
						return
					}
					backend.Send(ctx, data)
				}
			}()

			msg := []byte("ping")
			b.ResetTimer()
			for range b.N {
				client.Send(ctx, msg)
				client.Recv(ctx)
			}
		})
	}
}

// BenchmarkDatagramLatency measures round-trip time for a single datagram.
func BenchmarkDatagramLatency(b *testing.B) {
	for _, env := range benchRelays(b) {
		b.Run(env.name, func(b *testing.B) {
			ctx := context.Background()
			backend, client := benchPair(b, env.setup(b))

			// Backend echo loop.
			go func() {
				for {
					data, err := backend.RecvDatagram(ctx)
					if err != nil {
						return
					}
					backend.SendDatagram(data)
				}
			}()

			msg := []byte("ping")
			b.ResetTimer()
			for range b.N {
				client.SendDatagram(msg)
				client.RecvDatagram(ctx)
			}
		})
	}
}

// BenchmarkStreamThroughput measures one-way throughput at various payload sizes.
func BenchmarkStreamThroughput(b *testing.B) {
	for _, size := range []int{64, 1024, 16384, 65536} {
		for _, env := range benchRelays(b) {
			name := fmt.Sprintf("%s/%dB", env.name, size)
			b.Run(name, func(b *testing.B) {
				ctx := context.Background()
				backend, client := benchPair(b, env.setup(b))

				// Backend drain loop.
				go func() {
					for {
						if _, err := backend.Recv(ctx); err != nil {
							return
						}
					}
				}()

				msg := make([]byte, size)
				b.SetBytes(int64(size))
				b.ResetTimer()
				for range b.N {
					client.Send(ctx, msg)
				}
			})
		}
	}
}

// BenchmarkDatagramThroughput measures one-way datagram throughput.
func BenchmarkDatagramThroughput(b *testing.B) {
	for _, size := range []int{64, 1024} {
		for _, env := range benchRelays(b) {
			name := fmt.Sprintf("%s/%dB", env.name, size)
			b.Run(name, func(b *testing.B) {
				ctx := context.Background()
				backend, client := benchPair(b, env.setup(b))

				// Backend drain loop.
				go func() {
					for {
						if _, err := backend.RecvDatagram(ctx); err != nil {
							return
						}
					}
				}()

				msg := make([]byte, size)
				b.SetBytes(int64(size))
				b.ResetTimer()
				for range b.N {
					client.SendDatagram(msg)
				}
			})
		}
	}
}

// --- helpers ---

type benchRelay struct {
	name  string
	setup func(b *testing.B) relayEnv
}

// benchRelays returns relay environments for benchmarking. Each entry's
// setup is deferred until the sub-benchmark actually runs, so
// `-bench=/live` doesn't start a local server and vice versa.
func benchRelays(b *testing.B) []benchRelay {
	b.Helper()
	relays := []benchRelay{{
		name: "local",
		setup: func(b *testing.B) relayEnv {
			return localRelayB(b)
		},
	}}

	token, url := liveRelayEnv()
	if token != "" {
		relays = append(relays, benchRelay{
			name: "live",
			setup: func(b *testing.B) relayEnv {
				env := relayEnv{
					url: url,
					cfg: Config{
						Token:        token,
						WebTransport: true,
					},
				}
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				probe, err := Register(ctx, env.url, env.cfg)
				if err != nil {
					b.Skipf("live relay not reachable: %v", err)
				}
				probe.CloseNow()
				return env
			},
		})
	}

	return relays
}

func benchPair(b *testing.B, env relayEnv) (*Conn, *Conn) {
	b.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	b.Cleanup(cancel)

	backend, err := Register(ctx, env.url, env.cfg)
	if err != nil {
		b.Fatal("register:", err)
	}
	b.Cleanup(func() { backend.CloseNow() })

	client, err := Connect(ctx, env.url, backend.InstanceID(), env.cfg)
	if err != nil {
		b.Fatal("connect:", err)
	}
	b.Cleanup(func() { client.CloseNow() })

	return backend, client
}
