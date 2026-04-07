// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/marcelocantos/pigeon/faultproxy"
)

// --- Shared chaos infrastructure ---

// pairStats tracks per-pair traffic counters.
type pairStats struct {
	streamSent   atomic.Int64
	streamRecvd  atomic.Int64
	streamErrors atomic.Int64
	dgSent       atomic.Int64
	dgRecvd      atomic.Int64
}

// chaosWorkload runs a continuous send/recv workload on a single pair
// until the deadline. Returns stats.
func chaosWorkload(t *testing.T, ctx context.Context, b, c *Conn, deadline time.Time, pairID int) *pairStats {
	t.Helper()
	var stats pairStats
	var wg sync.WaitGroup

	// Stream sender: sequential numbered messages.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; time.Now().Before(deadline); i++ {
			msg := make([]byte, 12)
			binary.BigEndian.PutUint32(msg[:4], uint32(pairID))
			binary.BigEndian.PutUint64(msg[4:], uint64(i))

			sendCtx, sendCancel := context.WithTimeout(ctx, 5*time.Second)
			err := c.Send(sendCtx, msg)
			sendCancel()
			if err != nil {
				stats.streamErrors.Add(1)
				time.Sleep(100 * time.Millisecond)
				continue
			}
			stats.streamSent.Add(1)
			time.Sleep(time.Duration(20+rand.IntN(80)) * time.Millisecond)
		}
	}()

	// Stream receiver: verify ordering per pair.
	wg.Add(1)
	go func() {
		defer wg.Done()
		var lastSeq int64 = -1
		outOfOrder := 0
		for time.Now().Before(deadline) {
			recvCtx, recvCancel := context.WithTimeout(ctx, 5*time.Second)
			data, err := b.Recv(recvCtx)
			recvCancel()
			if err != nil {
				stats.streamErrors.Add(1)
				continue
			}
			stats.streamRecvd.Add(1)
			if len(data) == 12 {
				seq := int64(binary.BigEndian.Uint64(data[4:]))
				if seq <= lastSeq {
					outOfOrder++
				}
				lastSeq = seq
			}
		}
		if outOfOrder > 0 {
			t.Errorf("pair %d: %d out-of-order stream messages", pairID, outOfOrder)
		}
	}()

	// Datagram sender.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for time.Now().Before(deadline) {
			if err := c.SendDatagram(fmt.Appendf(nil, "dg-%d-%d", pairID, stats.dgSent.Load())); err == nil {
				stats.dgSent.Add(1)
			}
			time.Sleep(time.Duration(50+rand.IntN(100)) * time.Millisecond)
		}
	}()

	// Datagram receiver.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for time.Now().Before(deadline) {
			recvCtx, recvCancel := context.WithTimeout(ctx, 2*time.Second)
			_, err := b.RecvDatagram(recvCtx)
			recvCancel()
			if err == nil {
				stats.dgRecvd.Add(1)
			}
		}
	}()

	// Large message sender (tests fragmentation).
	wg.Add(1)
	go func() {
		defer wg.Done()
		for time.Now().Before(deadline) {
			size := 2000 + rand.IntN(8000)
			payload := make([]byte, size)
			crand.Read(payload)
			sendCtx, sendCancel := context.WithTimeout(ctx, 10*time.Second)
			err := c.Send(sendCtx, payload)
			sendCancel()
			if err == nil {
				stats.streamSent.Add(1)
			} else {
				stats.streamErrors.Add(1)
			}
			time.Sleep(time.Duration(500+rand.IntN(2000)) * time.Millisecond)
		}
	}()

	wg.Wait()
	return &stats
}

type scenario struct {
	name string
	opts []faultproxy.Option
}

var scenarios = []scenario{
	{"clean", nil},
	{"latency-50ms", []faultproxy.Option{faultproxy.WithLatency(50 * time.Millisecond, 20 * time.Millisecond)}},
	{"latency-200ms", []faultproxy.Option{faultproxy.WithLatency(200 * time.Millisecond, 50 * time.Millisecond)}},
	{"loss-5%", []faultproxy.Option{faultproxy.WithPacketLoss(0.05)}},
	{"loss-15%", []faultproxy.Option{faultproxy.WithPacketLoss(0.15)}},
	{"corrupt-2%", []faultproxy.Option{faultproxy.WithCorrupt(0.02)}},
	{"latency+loss", []faultproxy.Option{
		faultproxy.WithLatency(100 * time.Millisecond, 30 * time.Millisecond),
		faultproxy.WithPacketLoss(0.05),
	}},
	{"bandwidth-20KB", []faultproxy.Option{faultproxy.WithBandwidth(20000)}},
	{"blackhole-brief", []faultproxy.Option{faultproxy.WithPacketLoss(1.0)}},
}

// chaosController randomly cycles through fault scenarios on the proxy.
func chaosController(proxy *faultproxy.Proxy, deadline time.Time, done <-chan struct{}) int64 {
	var switches atomic.Int64
	for time.Now().Before(deadline) {
		s := scenarios[rand.IntN(len(scenarios))]
		switches.Add(1)

		if s.name == "blackhole-brief" {
			proxy.UpdateProfile(s.opts...)
			dur := 500*time.Millisecond + time.Duration(rand.Int64N(int64(1500*time.Millisecond)))
			select {
			case <-time.After(dur):
			case <-done:
				return switches.Load()
			}
			proxy.UpdateProfile()
		} else {
			proxy.UpdateProfile(s.opts...)
		}

		hold := 2*time.Second + time.Duration(rand.Int64N(int64(6*time.Second)))
		select {
		case <-time.After(hold):
		case <-done:
			return switches.Load()
		}
	}
	return switches.Load()
}

// chaosRelay sets up a local relay with a fault proxy, returns the
// relay env, proxy, and a cleanup function.
func chaosRelay(t *testing.T) (relayEnv, *faultproxy.Proxy) {
	t.Helper()

	cert, pool := generateTestCert(t)
	tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	srv, err := NewWebTransportServer("127.0.0.1:0", tlsCfg, "")
	if err != nil {
		t.Fatal(err)
	}
	wtUDP, _ := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	go srv.Serve(wtUDP)
	t.Cleanup(func() { srv.Close() })

	qsrv := NewQUICServer("127.0.0.1:0", tlsCfg, "", srv.Hub())
	qUDP, _ := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	go qsrv.ServeWithTLS(qUDP, tlsCfg)
	t.Cleanup(func() { qsrv.Close() })

	qPort := qUDP.LocalAddr().(*net.UDPAddr).Port
	wtPort := wtUDP.LocalAddr().(*net.UDPAddr).Port

	proxy, err := faultproxy.New("127.0.0.1:" + strconv.Itoa(qPort))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { proxy.Close() })
	proxyAddr, _ := net.ResolveUDPAddr("udp", proxy.Addr())

	env := relayEnv{
		url: "https://127.0.0.1:" + strconv.Itoa(wtPort),
		cfg: Config{
			TLS:      &tls.Config{RootCAs: pool},
			QUICPort: strconv.Itoa(proxyAddr.Port),
		},
	}
	return env, proxy
}

func reportPairStats(t *testing.T, label string, stats *pairStats) {
	t.Helper()
	t.Logf("  %s: stream sent=%d recv=%d errors=%d loss=%.1f%% | dg sent=%d recv=%d loss=%.1f%%",
		label,
		stats.streamSent.Load(), stats.streamRecvd.Load(), stats.streamErrors.Load(),
		100*float64(stats.streamSent.Load()-stats.streamRecvd.Load())/float64(max(stats.streamSent.Load(), 1)),
		stats.dgSent.Load(), stats.dgRecvd.Load(),
		100*float64(stats.dgSent.Load()-stats.dgRecvd.Load())/float64(max(stats.dgSent.Load(), 1)))
}

func assertPairStats(t *testing.T, label string, stats *pairStats, maxLoss float64) {
	t.Helper()
	if stats.streamSent.Load() > 0 && stats.streamRecvd.Load() == 0 {
		t.Errorf("%s: no stream messages received", label)
	}
	loss := float64(stats.streamSent.Load()-stats.streamRecvd.Load()) / float64(max(stats.streamSent.Load(), 1))
	if loss > maxLoss {
		t.Errorf("%s: stream loss %.1f%% exceeds %.0f%%", label, loss*100, maxLoss*100)
	}
	if stats.dgSent.Load() > 10 && stats.dgRecvd.Load() == 0 {
		t.Errorf("%s: no datagrams received", label)
	}
}

// --- Test functions ---

// TestChaos runs a single-pair chaos test against a local relay.
//
// Run with: go test -run TestChaos$ -timeout 600s -v
// Override duration: CHAOS_DURATION=5m
func TestChaos(t *testing.T) {
	if testing.Short() {
		t.Skip("chaos test skipped in short mode")
	}

	duration := parseDuration("CHAOS_DURATION", 60*time.Second)
	t.Logf("chaos: duration=%v, 1 pair, local relay", duration)

	env, proxy := chaosRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), duration+30*time.Second)
	defer cancel()

	b, err := Register(ctx, env.url, env.cfg)
	if err != nil {
		t.Fatal("register:", err)
	}
	defer b.CloseNow()

	c, err := Connect(ctx, env.url, b.InstanceID(), env.cfg)
	if err != nil {
		t.Fatal("connect:", err)
	}
	defer c.CloseNow()

	deadline := time.Now().Add(duration)
	done := make(chan struct{})

	var switches int64
	go func() {
		switches = chaosController(proxy, deadline, done)
	}()

	stats := chaosWorkload(t, ctx, b, c, deadline, 0)
	close(done)

	proxy.UpdateProfile()
	proxyStats := proxy.GetStats()

	t.Logf("=== Chaos results (%v) ===", duration)
	t.Logf("Scenarios: %d switches", switches)
	reportPairStats(t, "pair-0", stats)
	t.Logf("Proxy: forwarded=%d dropped=%d corrupted=%d",
		proxyStats.PacketsForwarded.Load(),
		proxyStats.PacketsDropped.Load(),
		proxyStats.PacketsCorrupted.Load())
	assertPairStats(t, "pair-0", stats, 0.5)
}

// TestChaosMultiPair runs multiple concurrent pairs through the same
// relay and fault proxy, testing relay contention and cross-pair isolation.
//
// Run with: go test -run TestChaosMultiPair -timeout 600s -v
// Override: CHAOS_PAIRS=10 CHAOS_DURATION=2m
func TestChaosMultiPair(t *testing.T) {
	if testing.Short() {
		t.Skip("chaos test skipped in short mode")
	}

	duration := parseDuration("CHAOS_DURATION", 30*time.Second)
	numPairs := parseIntEnv("CHAOS_PAIRS", 5)
	t.Logf("chaos: duration=%v, %d pairs, local relay", duration, numPairs)

	env, proxy := chaosRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), duration+30*time.Second)
	defer cancel()

	done := make(chan struct{})

	// Connect all pairs first (no chaos yet).
	type pair struct{ b, c *Conn }
	pairs := make([]pair, 0, numPairs)
	for i := range numPairs {
		b, err := Register(ctx, env.url, env.cfg)
		if err != nil {
			t.Fatalf("pair %d: register: %v", i, err)
		}
		t.Cleanup(func() { b.CloseNow() })

		c, err := Connect(ctx, env.url, b.InstanceID(), env.cfg)
		if err != nil {
			t.Fatalf("pair %d: connect: %v", i, err)
		}
		t.Cleanup(func() { c.CloseNow() })

		pairs = append(pairs, pair{b, c})
	}
	t.Logf("all %d pairs connected, starting chaos", numPairs)

	// Set deadline AFTER connecting pairs so connection time doesn't eat
	// into the chaos duration.
	deadline := time.Now().Add(duration)

	// Now start chaos controller.
	var switches atomic.Int64
	go func() {
		switches.Store(chaosController(proxy, deadline, done))
	}()

	// Run workloads.
	allStats := make([]*pairStats, numPairs)
	var wg sync.WaitGroup
	for i, p := range pairs {
		wg.Add(1)
		go func(pairID int, p pair) {
			defer wg.Done()
			allStats[pairID] = chaosWorkload(t, ctx, p.b, p.c, deadline, pairID)
		}(i, p)
	}

	wg.Wait()
	close(done)

	proxy.UpdateProfile()
	proxyStats := proxy.GetStats()

	t.Logf("=== Multi-pair chaos results (%v, %d pairs) ===", duration, numPairs)
	t.Logf("Scenarios: %d switches", switches.Load())
	for i, s := range allStats {
		if s != nil {
			reportPairStats(t, fmt.Sprintf("pair-%d", i), s)
			assertPairStats(t, fmt.Sprintf("pair-%d", i), s, 0.8)
		}
	}
	t.Logf("Proxy: forwarded=%d dropped=%d corrupted=%d",
		proxyStats.PacketsForwarded.Load(),
		proxyStats.PacketsDropped.Load(),
		proxyStats.PacketsCorrupted.Load())
}

// TestChaosLive runs a chaos workload against the deployed relay at
// carrier-pigeon.fly.dev. Requires PIGEON_TOKEN to be set. No fault proxy —
// tests real network conditions.
//
// Run with: PIGEON_TOKEN=<tok> go test -run TestChaosLive -timeout 600s -v
func TestChaosLive(t *testing.T) {
	token := os.Getenv("PIGEON_TOKEN")
	if token == "" {
		t.Skip("PIGEON_TOKEN not set; skipping live chaos test")
	}
	if testing.Short() {
		t.Skip("chaos test skipped in short mode")
	}

	duration := parseDuration("CHAOS_DURATION", 60*time.Second)
	numPairs := parseIntEnv("CHAOS_PAIRS", 3)
	t.Logf("chaos live: duration=%v, %d pairs, carrier-pigeon.fly.dev", duration, numPairs)

	ctx, cancel := context.WithTimeout(context.Background(), duration+30*time.Second)
	defer cancel()

	deadline := time.Now().Add(duration)

	allStats := make([]*pairStats, numPairs)
	var wg sync.WaitGroup
	for i := range numPairs {
		wg.Add(1)
		go func(pairID int) {
			defer wg.Done()

			cfg := Config{Token: token}

			b, err := Register(ctx, "https://carrier-pigeon.fly.dev", cfg)
			if err != nil {
				t.Errorf("pair %d: register: %v", pairID, err)
				return
			}
			defer b.CloseNow()
			t.Logf("pair %d: registered as %s", pairID, b.InstanceID())

			c, err := Connect(ctx, "https://carrier-pigeon.fly.dev", b.InstanceID(), cfg)
			if err != nil {
				t.Errorf("pair %d: connect: %v", pairID, err)
				return
			}
			defer c.CloseNow()

			allStats[pairID] = chaosWorkload(t, ctx, b, c, deadline, pairID)
		}(i)
	}

	wg.Wait()

	t.Logf("=== Live chaos results (%v, %d pairs) ===", duration, numPairs)
	for i, s := range allStats {
		if s != nil {
			reportPairStats(t, fmt.Sprintf("pair-%d", i), s)
			// Live tests: more lenient — real network has its own loss.
			if s.streamSent.Load() > 0 && s.streamRecvd.Load() == 0 {
				t.Errorf("pair-%d: no stream messages received", i)
			}
		}
	}
}

// TestChaosWithLAN runs chaos with LAN path switching.
func TestChaosWithLAN(t *testing.T) {
	if testing.Short() {
		t.Skip("chaos test skipped in short mode")
	}

	duration := parseDuration("CHAOS_DURATION", 30*time.Second)
	t.Logf("chaos+LAN: duration=%v", duration)

	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), duration+15*time.Second)
	defer cancel()

	b, c, _ := lanPair(t, env)
	waitForLAN(t, ctx, b, c)

	var sent, recvd atomic.Int64
	deadline := time.Now().Add(duration)

	// Periodically force fallback and re-establishment.
	go func() {
		for time.Now().Before(deadline) {
			time.Sleep(3*time.Second + time.Duration(rand.Int64N(int64(4*time.Second))))
			c.fallbackToRelay()
			slog.Info("chaos: forced fallback")
			time.Sleep(2 * time.Second)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; time.Now().Before(deadline); i++ {
			msg := fmt.Appendf(nil, "chaos-%d", i)
			sendCtx, sendCancel := context.WithTimeout(ctx, 5*time.Second)
			err := c.Send(sendCtx, msg)
			sendCancel()
			if err != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			sent.Add(1)
			time.Sleep(50 * time.Millisecond)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for time.Now().Before(deadline) {
			recvCtx, recvCancel := context.WithTimeout(ctx, 5*time.Second)
			data, err := b.Recv(recvCtx)
			recvCancel()
			if err != nil {
				continue
			}
			if bytes.HasPrefix(data, []byte("chaos-")) {
				recvd.Add(1)
			}
		}
	}()

	wg.Wait()

	t.Logf("chaos+LAN: sent=%d recv=%d loss=%.1f%%",
		sent.Load(), recvd.Load(),
		100*float64(sent.Load()-recvd.Load())/float64(max(sent.Load(), 1)))

	if sent.Load() > 0 && recvd.Load() == 0 {
		t.Fatal("no messages received")
	}
}

// --- Helpers ---

func parseDuration(envKey string, def time.Duration) time.Duration {
	if v := os.Getenv(envKey); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func parseIntEnv(envKey string, def int) int {
	if v := os.Getenv(envKey); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
