// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package tern

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/marcelocantos/tern/faultproxy"
)

// faultyRelay starts a local relay and returns a relayEnv that routes
// through a fault proxy. The proxy sits between the client and the
// relay's QUIC port.
func faultyRelay(t *testing.T, opts ...faultproxy.Option) (relayEnv, *faultproxy.Proxy) {
	t.Helper()

	cert, pool := generateTestCert(t)
	tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	srv, err := NewWebTransportServer("127.0.0.1:0", tlsCfg, "")
	if err != nil {
		t.Fatal(err)
	}
	wtUDP, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	go srv.Serve(wtUDP)
	t.Cleanup(func() { srv.Close() })

	qsrv := NewQUICServer("127.0.0.1:0", tlsCfg, "", srv.Hub())
	qUDP, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	go qsrv.ServeWithTLS(qUDP, tlsCfg)
	t.Cleanup(func() { qsrv.Close() })

	qPort := qUDP.LocalAddr().(*net.UDPAddr).Port
	wtPort := wtUDP.LocalAddr().(*net.UDPAddr).Port

	// Proxy sits in front of the QUIC port.
	proxy, err := faultproxy.New("127.0.0.1:"+strconv.Itoa(qPort), opts...)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { proxy.Close() })

	// Parse proxy address to get its port.
	proxyAddr, _ := net.ResolveUDPAddr("udp", proxy.Addr())

	return relayEnv{
		url: "https://127.0.0.1:" + strconv.Itoa(wtPort),
		opts: []Option{
			WithTLS(&tls.Config{RootCAs: pool}),
			WithQUICPort(strconv.Itoa(proxyAddr.Port)),
		},
	}, proxy
}

// TestHighLatencyStreamRoundTrip verifies stream messaging works under
// 100ms latency with 30ms jitter.
func TestHighLatencyStreamRoundTrip(t *testing.T) {
	env, _ := faultyRelay(t,
		faultproxy.WithLatency(100*time.Millisecond, 30*time.Millisecond),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	start := time.Now()
	if err := c.Send(ctx, []byte("high-latency")); err != nil {
		t.Fatal(err)
	}
	data, err := b.Recv(ctx)
	if err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(start)

	if string(data) != "high-latency" {
		t.Fatalf("got %q", data)
	}
	// Should take noticeably longer than without proxy.
	t.Logf("high-latency round-trip: %v", elapsed)
}

// TestPacketLossStreamRecovery verifies that QUIC's reliability layer
// recovers from packet loss on the stream path.
func TestPacketLossStreamRecovery(t *testing.T) {
	env, proxy := faultyRelay(t,
		faultproxy.WithPacketLoss(0.1), // 10% loss
	)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	// Send 20 messages — QUIC retransmits should deliver all of them.
	for i := range 20 {
		msg := []byte("msg-" + strconv.Itoa(i))
		if err := c.Send(ctx, msg); err != nil {
			t.Fatalf("send %d: %v", i, err)
		}
	}

	for i := range 20 {
		data, err := b.Recv(ctx)
		if err != nil {
			t.Fatalf("recv %d: %v", i, err)
		}
		expected := "msg-" + strconv.Itoa(i)
		if string(data) != expected {
			t.Fatalf("message %d: got %q, want %q", i, data, expected)
		}
	}

	stats := proxy.GetStats()
	t.Logf("packet loss recovery: forwarded=%d dropped=%d",
		stats.PacketsForwarded.Load(), stats.PacketsDropped.Load())
}

// TestDatagramLossUnderFault verifies that datagrams degrade gracefully
// under packet loss — some are lost, no crashes, no hangs.
func TestDatagramLossUnderFault(t *testing.T) {
	env, proxy := faultyRelay(t,
		faultproxy.WithPacketLoss(0.2), // 20% loss
	)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	// Send 50 datagrams.
	sent := 0
	for range 50 {
		if err := c.SendDatagram([]byte("dg")); err != nil {
			continue
		}
		sent++
	}

	// Receive whatever arrives.
	received := 0
	recvCtx, recvCancel := context.WithTimeout(ctx, 3*time.Second)
	defer recvCancel()
	for {
		_, err := b.RecvDatagram(recvCtx)
		if err != nil {
			break
		}
		received++
	}

	stats := proxy.GetStats()
	t.Logf("datagram under 20%% loss: sent=%d received=%d dropped=%d",
		sent, received, stats.PacketsDropped.Load())

	// Should receive some but not all.
	if received == 0 {
		t.Fatal("no datagrams received at all")
	}
}

// TestCorruptionHandledByQUIC verifies that QUIC rejects corrupted
// packets and the connection survives.
func TestCorruptionHandledByQUIC(t *testing.T) {
	env, _ := faultyRelay(t,
		faultproxy.WithCorrupt(0.05), // 5% corruption
	)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	// QUIC should detect and retransmit corrupted packets.
	for i := range 10 {
		msg := []byte("integrity-" + strconv.Itoa(i))
		if err := c.Send(ctx, msg); err != nil {
			t.Fatalf("send %d: %v", i, err)
		}
	}

	for i := range 10 {
		data, err := b.Recv(ctx)
		if err != nil {
			t.Fatalf("recv %d: %v", i, err)
		}
		expected := "integrity-" + strconv.Itoa(i)
		if string(data) != expected {
			t.Fatalf("message %d: got %q, want %q", i, data, expected)
		}
	}
}

// TestChannelUnderLatencyAndLoss verifies that streaming channels work
// correctly under combined latency and loss.
func TestChannelUnderLatencyAndLoss(t *testing.T) {
	env, _ := faultyRelay(t,
		faultproxy.WithLatency(50*time.Millisecond, 20*time.Millisecond),
		faultproxy.WithPacketLoss(0.05),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	ch, err := c.OpenChannel("fault-test")
	if err != nil {
		t.Fatal(err)
	}
	defer ch.Close()

	bch, err := b.AcceptChannel(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer bch.Close()

	// Bidirectional messaging on named channel.
	for i := range 10 {
		msg := []byte("ch-" + strconv.Itoa(i))
		if err := ch.Send(ctx, msg); err != nil {
			t.Fatalf("send %d: %v", i, err)
		}
		data, err := bch.Recv(ctx)
		if err != nil {
			t.Fatalf("recv %d: %v", i, err)
		}
		if string(data) != string(msg) {
			t.Fatalf("message %d: got %q, want %q", i, data, msg)
		}
	}
}

// TestMidConversationFaultChange verifies that changing the fault
// profile mid-conversation (e.g., sudden packet loss spike) doesn't
// break the connection.
func TestMidConversationFaultChange(t *testing.T) {
	env, proxy := faultyRelay(t) // start clean
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	// Send 5 messages cleanly.
	for i := range 5 {
		c.Send(ctx, []byte("clean-"+strconv.Itoa(i)))
		data, err := b.Recv(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "clean-"+strconv.Itoa(i) {
			t.Fatalf("got %q", data)
		}
	}

	// Inject 10% loss and 80ms latency mid-conversation.
	proxy.UpdateProfile(
		faultproxy.WithPacketLoss(0.1),
		faultproxy.WithLatency(80*time.Millisecond, 20*time.Millisecond),
	)

	// Send 10 more messages — should still arrive (QUIC retransmits).
	for i := range 10 {
		if err := c.Send(ctx, []byte("fault-"+strconv.Itoa(i))); err != nil {
			t.Fatalf("send under fault %d: %v", i, err)
		}
	}
	for i := range 10 {
		data, err := b.Recv(ctx)
		if err != nil {
			t.Fatalf("recv under fault %d: %v", i, err)
		}
		if string(data) != "fault-"+strconv.Itoa(i) {
			t.Fatalf("got %q", data)
		}
	}
}

// TestBandwidthThrottledTransfer verifies that large messages complete
// under bandwidth throttling (just slower).
func TestBandwidthThrottledTransfer(t *testing.T) {
	env, _ := faultyRelay(t,
		faultproxy.WithBandwidth(50000), // 50KB/s
	)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	// Send a 10KB message.
	payload := make([]byte, 10000)
	for i := range payload {
		payload[i] = byte(i % 256)
	}

	start := time.Now()
	if err := c.Send(ctx, payload); err != nil {
		t.Fatal(err)
	}
	data, err := b.Recv(ctx)
	if err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(start)

	if len(data) != len(payload) {
		t.Fatalf("got %d bytes, want %d", len(data), len(payload))
	}
	for i := range data {
		if data[i] != payload[i] {
			t.Fatalf("byte %d: got %d, want %d", i, data[i], payload[i])
		}
	}
	t.Logf("10KB at 50KB/s: %v", elapsed)
}

// TestBlackholeRecovery verifies that a QUIC connection survives a
// brief network blackout and resumes communication afterward.
func TestBlackholeRecovery(t *testing.T) {
	env, proxy := faultyRelay(t) // start clean
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	// Verify connectivity.
	c.Send(ctx, []byte("before"))
	data, err := b.Recv(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "before" {
		t.Fatalf("got %q", data)
	}

	// Enable blackhole for 500ms.
	proxy.UpdateProfile(faultproxy.WithPacketLoss(1.0))
	time.Sleep(500 * time.Millisecond)
	proxy.UpdateProfile(faultproxy.WithPacketLoss(0))

	// Connection should recover — QUIC keeps the session alive.
	if err := c.Send(ctx, []byte("after")); err != nil {
		t.Fatal("send after blackhole:", err)
	}
	data, err = b.Recv(ctx)
	if err != nil {
		t.Fatal("recv after blackhole:", err)
	}
	if string(data) != "after" {
		t.Fatalf("got %q", data)
	}
}
