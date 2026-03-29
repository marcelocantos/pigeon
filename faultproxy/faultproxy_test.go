// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package faultproxy

import (
	"net"
	"testing"
	"time"
)

func echoServer(t *testing.T) *net.UDPConn {
	t.Helper()
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	go func() {
		buf := make([]byte, 65536)
		for {
			n, addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				return
			}
			conn.WriteToUDP(buf[:n], addr)
		}
	}()
	return conn
}

func TestPassthrough(t *testing.T) {
	echo := echoServer(t)
	proxy, err := New(echo.LocalAddr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer proxy.Close()

	client, err := net.DialUDP("udp", nil, proxy.conn.LocalAddr().(*net.UDPAddr))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	msg := []byte("hello")
	client.Write(msg)
	client.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 1024)
	n, err := client.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	if string(buf[:n]) != "hello" {
		t.Fatalf("got %q, want hello", buf[:n])
	}

	stats := proxy.GetStats()
	if stats.PacketsForwarded.Load() < 2 {
		t.Fatalf("expected at least 2 forwarded packets, got %d", stats.PacketsForwarded.Load())
	}
}

func TestLatency(t *testing.T) {
	echo := echoServer(t)
	proxy, err := New(echo.LocalAddr().String(), WithLatency(100*time.Millisecond, 0))
	if err != nil {
		t.Fatal(err)
	}
	defer proxy.Close()

	client, err := net.DialUDP("udp", nil, proxy.conn.LocalAddr().(*net.UDPAddr))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	start := time.Now()
	client.Write([]byte("ping"))
	client.SetReadDeadline(time.Now().Add(5 * time.Second))
	buf := make([]byte, 1024)
	_, err = client.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(start)

	// Should take at least 200ms (100ms each way).
	if elapsed < 180*time.Millisecond {
		t.Fatalf("round-trip took %v, expected >= 200ms", elapsed)
	}
}

func TestPacketLoss(t *testing.T) {
	echo := echoServer(t)
	proxy, err := New(echo.LocalAddr().String(), WithPacketLoss(1.0))
	if err != nil {
		t.Fatal(err)
	}
	defer proxy.Close()

	client, err := net.DialUDP("udp", nil, proxy.conn.LocalAddr().(*net.UDPAddr))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	// Send 10 packets — all should be dropped.
	for range 10 {
		client.Write([]byte("test"))
	}

	client.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	buf := make([]byte, 1024)
	_, err = client.Read(buf)
	if err == nil {
		t.Fatal("expected timeout, got a packet")
	}

	stats := proxy.GetStats()
	if stats.PacketsDropped.Load() < 10 {
		t.Fatalf("expected >= 10 dropped, got %d", stats.PacketsDropped.Load())
	}
}

func TestCorruption(t *testing.T) {
	echo := echoServer(t)
	proxy, err := New(echo.LocalAddr().String(), WithCorrupt(1.0))
	if err != nil {
		t.Fatal(err)
	}
	defer proxy.Close()

	client, err := net.DialUDP("udp", nil, proxy.conn.LocalAddr().(*net.UDPAddr))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	// Send 20 identical packets. The echo server returns what it received
	// (already corrupted). At least some should differ from the original.
	msg := []byte("AAAAAAAAAA")
	corrupted := 0
	for range 20 {
		client.Write(msg)
		client.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		buf := make([]byte, 1024)
		n, err := client.Read(buf)
		if err != nil {
			continue
		}
		if string(buf[:n]) != string(msg) {
			corrupted++
		}
	}

	if corrupted == 0 {
		t.Fatal("expected some corrupted packets")
	}

	stats := proxy.GetStats()
	if stats.PacketsCorrupted.Load() == 0 {
		t.Fatal("corruption counter is 0")
	}
}

func TestUpdateProfile(t *testing.T) {
	echo := echoServer(t)
	proxy, err := New(echo.LocalAddr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer proxy.Close()

	client, err := net.DialUDP("udp", nil, proxy.conn.LocalAddr().(*net.UDPAddr))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	// Initially no loss — packets go through.
	client.Write([]byte("before"))
	client.SetReadDeadline(time.Now().Add(time.Second))
	buf := make([]byte, 1024)
	_, err = client.Read(buf)
	if err != nil {
		t.Fatal("expected packet before profile change:", err)
	}

	// Enable 100% loss.
	proxy.UpdateProfile(WithPacketLoss(1.0))

	client.Write([]byte("after"))
	client.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	_, err = client.Read(buf)
	if err == nil {
		t.Fatal("expected timeout after 100% loss")
	}
}

func TestBlackhole(t *testing.T) {
	echo := echoServer(t)
	proxy, err := New(echo.LocalAddr().String(),
		WithBlackhole(200*time.Millisecond, 100*time.Millisecond),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer proxy.Close()

	client, err := net.DialUDP("udp", nil, proxy.conn.LocalAddr().(*net.UDPAddr))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	// Over 2 seconds, send packets every 50ms. Some should get through,
	// some should be dropped during blackhole periods.
	sent := 0
	received := 0
	for range 40 {
		client.Write([]byte("x"))
		sent++
		client.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		buf := make([]byte, 1024)
		_, err := client.Read(buf)
		if err == nil {
			received++
		}
		time.Sleep(50 * time.Millisecond)
	}

	// Should have a mix of delivered and dropped.
	if received == 0 {
		t.Fatal("no packets received at all")
	}
	if received == sent {
		t.Fatal("all packets received — blackhole didn't drop any")
	}
	t.Logf("blackhole: sent=%d received=%d dropped=%d", sent, received, sent-received)
}

func TestBandwidthThrottle(t *testing.T) {
	echo := echoServer(t)
	// 10KB/s — sending 10KB should take about 1 second.
	proxy, err := New(echo.LocalAddr().String(), WithBandwidth(10000))
	if err != nil {
		t.Fatal(err)
	}
	defer proxy.Close()

	client, err := net.DialUDP("udp", nil, proxy.conn.LocalAddr().(*net.UDPAddr))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	// Send 10 × 1KB packets.
	payload := make([]byte, 1000)
	start := time.Now()
	for range 10 {
		client.Write(payload)
	}

	// Wait for all echoes.
	client.SetReadDeadline(time.Now().Add(5 * time.Second))
	buf := make([]byte, 2000)
	for range 10 {
		_, err := client.Read(buf)
		if err != nil {
			break
		}
	}
	elapsed := time.Since(start)

	// Should take at least 500ms (throttle applies to both directions).
	if elapsed < 400*time.Millisecond {
		t.Fatalf("10KB at 10KB/s took only %v", elapsed)
	}
	t.Logf("bandwidth: 10KB at 10KB/s took %v", elapsed)
}
