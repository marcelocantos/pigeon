// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Package faultproxy provides a transparent UDP proxy that injects
// network faults between a QUIC client and server. It forwards packets
// bidirectionally while applying configurable latency, jitter, packet
// loss, reordering, corruption, and bandwidth throttling.
//
// Usage:
//
//	proxy, _ := faultproxy.New(relayAddr,
//	    faultproxy.WithLatency(50*time.Millisecond, 20*time.Millisecond),
//	    faultproxy.WithPacketLoss(0.05),
//	)
//	defer proxy.Close()
//	// Connect to proxy.Addr() instead of the relay.
package faultproxy

import (
	"math/rand/v2"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Profile configures the fault injection behaviour.
type Profile struct {
	// Latency adds a fixed delay to every packet.
	Latency time.Duration
	// Jitter adds uniform random delay in [-Jitter, +Jitter] on top of Latency.
	Jitter time.Duration
	// PacketLoss is the probability [0,1] that a packet is silently dropped.
	PacketLoss float64
	// Reorder is the probability [0,1] that a packet is delayed by an
	// extra 0-3× Latency, causing it to arrive out of order.
	Reorder float64
	// Corrupt is the probability [0,1] that a random byte in the packet
	// is flipped. The packet is still delivered (QUIC will reject it).
	Corrupt float64
	// BandwidthBytesPerSec limits throughput. 0 means unlimited.
	BandwidthBytesPerSec int
	// BlackholeDuration, if set, periodically drops all packets for this
	// duration. Combined with BlackholeInterval to simulate connectivity
	// gaps.
	BlackholeDuration time.Duration
	// BlackholeInterval is the time between blackhole periods.
	BlackholeInterval time.Duration
}

// Option configures a Proxy.
type Option func(*Profile)

// WithLatency sets fixed latency and jitter.
func WithLatency(base, jitter time.Duration) Option {
	return func(p *Profile) { p.Latency = base; p.Jitter = jitter }
}

// WithPacketLoss sets the packet loss probability.
func WithPacketLoss(rate float64) Option {
	return func(p *Profile) { p.PacketLoss = rate }
}

// WithReorder sets the reordering probability.
func WithReorder(rate float64) Option {
	return func(p *Profile) { p.Reorder = rate }
}

// WithCorrupt sets the corruption probability.
func WithCorrupt(rate float64) Option {
	return func(p *Profile) { p.Corrupt = rate }
}

// WithBandwidth sets the bandwidth limit in bytes per second.
func WithBandwidth(bytesPerSec int) Option {
	return func(p *Profile) { p.BandwidthBytesPerSec = bytesPerSec }
}

// WithBlackhole configures periodic connectivity blackouts.
func WithBlackhole(duration, interval time.Duration) Option {
	return func(p *Profile) { p.BlackholeDuration = duration; p.BlackholeInterval = interval }
}

// Stats tracks proxy traffic counters.
type Stats struct {
	PacketsForwarded atomic.Int64
	PacketsDropped   atomic.Int64
	PacketsCorrupted atomic.Int64
	PacketsReordered atomic.Int64
	BytesForwarded   atomic.Int64
}

// Proxy is a transparent UDP fault-injection proxy.
type Proxy struct {
	conn     *net.UDPConn // listens for client packets
	target   *net.UDPAddr // the real relay
	profile  Profile
	stats    Stats
	done     chan struct{}
	wg       sync.WaitGroup
	mu       sync.Mutex
	clients  map[string]*net.UDPConn // client addr -> upstream conn
	throttle *throttle
	blackholed atomic.Bool
}

// New creates and starts a fault-injection proxy that forwards UDP
// packets to target. The proxy listens on a random local port.
func New(target string, opts ...Option) (*Proxy, error) {
	targetAddr, err := net.ResolveUDPAddr("udp", target)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		return nil, err
	}

	var profile Profile
	for _, o := range opts {
		o(&profile)
	}

	p := &Proxy{
		conn:    conn,
		target:  targetAddr,
		profile: profile,
		done:    make(chan struct{}),
		clients: make(map[string]*net.UDPConn),
	}

	if profile.BandwidthBytesPerSec > 0 {
		p.throttle = newThrottle(profile.BandwidthBytesPerSec)
	}

	if profile.BlackholeInterval > 0 && profile.BlackholeDuration > 0 {
		p.wg.Add(1)
		go p.blackholeLoop()
	}

	p.wg.Add(1)
	go p.readLoop()

	return p, nil
}

// Addr returns the proxy's listen address. Pass this to the client
// instead of the real relay address.
func (p *Proxy) Addr() string {
	return p.conn.LocalAddr().String()
}

// GetStats returns a pointer to the live traffic counters.
func (p *Proxy) GetStats() *Stats {
	return &p.stats
}

// UpdateProfile atomically replaces the fault profile. This allows
// changing fault parameters mid-test.
func (p *Proxy) UpdateProfile(opts ...Option) {
	p.mu.Lock()
	for _, o := range opts {
		o(&p.profile)
	}
	if p.profile.BandwidthBytesPerSec > 0 {
		p.throttle = newThrottle(p.profile.BandwidthBytesPerSec)
	} else {
		p.throttle = nil
	}
	p.mu.Unlock()
}

// Close stops the proxy and releases resources.
func (p *Proxy) Close() error {
	close(p.done)
	p.conn.Close()
	p.mu.Lock()
	for _, c := range p.clients {
		c.Close()
	}
	p.mu.Unlock()
	p.wg.Wait()
	return nil
}

// readLoop reads packets from clients and forwards them to the target.
func (p *Proxy) readLoop() {
	defer p.wg.Done()
	buf := make([]byte, 65536)
	for {
		n, clientAddr, err := p.conn.ReadFromUDP(buf)
		if err != nil {
			select {
			case <-p.done:
				return
			default:
				continue
			}
		}

		pkt := make([]byte, n)
		copy(pkt, buf[:n])

		// Get or create an upstream connection for this client.
		upstream := p.getOrCreateUpstream(clientAddr)

		// Apply faults and forward client → target.
		p.forward(pkt, upstream, nil, true)
	}
}

// getOrCreateUpstream returns the upstream UDP connection for a client,
// creating one if needed. Each client gets its own upstream so replies
// can be routed back.
func (p *Proxy) getOrCreateUpstream(clientAddr *net.UDPAddr) *net.UDPConn {
	key := clientAddr.String()
	p.mu.Lock()
	upstream, ok := p.clients[key]
	if ok {
		p.mu.Unlock()
		return upstream
	}

	upstream, err := net.DialUDP("udp", nil, p.target)
	if err != nil {
		p.mu.Unlock()
		return nil
	}
	p.clients[key] = upstream
	p.mu.Unlock()

	// Start a goroutine to relay target → client replies.
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		buf := make([]byte, 65536)
		for {
			n, err := upstream.Read(buf)
			if err != nil {
				select {
				case <-p.done:
					return
				default:
					return
				}
			}
			pkt := make([]byte, n)
			copy(pkt, buf[:n])
			p.forward(pkt, nil, clientAddr, false)
		}
	}()

	return upstream
}

// forward applies the fault profile and sends the packet.
// If upstream is non-nil, sends to upstream (client→target).
// If clientAddr is non-nil, sends back to client (target→client).
func (p *Proxy) forward(pkt []byte, upstream *net.UDPConn, clientAddr *net.UDPAddr, _ bool) {
	p.mu.Lock()
	profile := p.profile
	throttle := p.throttle
	p.mu.Unlock()

	// Blackhole: drop everything.
	if p.blackholed.Load() {
		p.stats.PacketsDropped.Add(1)
		return
	}

	// Packet loss.
	if profile.PacketLoss > 0 && rand.Float64() < profile.PacketLoss {
		p.stats.PacketsDropped.Add(1)
		return
	}

	// Corruption: flip a random byte.
	if profile.Corrupt > 0 && rand.Float64() < profile.Corrupt && len(pkt) > 0 {
		corrupted := make([]byte, len(pkt))
		copy(corrupted, pkt)
		idx := rand.IntN(len(corrupted))
		corrupted[idx] ^= byte(rand.IntN(255) + 1)
		pkt = corrupted
		p.stats.PacketsCorrupted.Add(1)
	}

	// Bandwidth throttle.
	if throttle != nil {
		throttle.wait(len(pkt))
	}

	// Calculate delay.
	delay := profile.Latency
	if profile.Jitter > 0 {
		jitter := time.Duration(rand.Int64N(int64(2*profile.Jitter))) - profile.Jitter
		delay += jitter
	}

	// Reorder: add extra random delay.
	if profile.Reorder > 0 && rand.Float64() < profile.Reorder {
		extra := time.Duration(rand.Int64N(int64(3 * profile.Latency)))
		delay += extra
		p.stats.PacketsReordered.Add(1)
	}

	send := func() {
		if upstream != nil {
			upstream.Write(pkt)
		} else if clientAddr != nil {
			p.conn.WriteToUDP(pkt, clientAddr)
		}
		p.stats.PacketsForwarded.Add(1)
		p.stats.BytesForwarded.Add(int64(len(pkt)))
	}

	if delay > 0 {
		go func() {
			timer := time.NewTimer(delay)
			select {
			case <-timer.C:
				send()
			case <-p.done:
				timer.Stop()
			}
		}()
	} else {
		send()
	}
}

// blackholeLoop periodically enables and disables the blackhole.
func (p *Proxy) blackholeLoop() {
	defer p.wg.Done()
	for {
		select {
		case <-p.done:
			return
		case <-time.After(p.profile.BlackholeInterval):
		}

		p.blackholed.Store(true)

		select {
		case <-p.done:
			return
		case <-time.After(p.profile.BlackholeDuration):
		}

		p.blackholed.Store(false)
	}
}

// throttle implements a simple token-bucket rate limiter.
type throttle struct {
	mu       sync.Mutex
	rate     float64 // bytes per nanosecond
	tokens   float64
	maxBurst float64
	lastTime time.Time
}

func newThrottle(bytesPerSec int) *throttle {
	burst := float64(bytesPerSec) // 1 second of burst
	return &throttle{
		rate:     float64(bytesPerSec) / 1e9,
		tokens:   burst,
		maxBurst: burst,
		lastTime: time.Now(),
	}
}

func (t *throttle) wait(n int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(t.lastTime)
	t.lastTime = now
	t.tokens += float64(elapsed.Nanoseconds()) * t.rate
	if t.tokens > t.maxBurst {
		t.tokens = t.maxBurst
	}

	t.tokens -= float64(n)
	if t.tokens < 0 {
		// Sleep until we have enough tokens.
		deficit := -t.tokens
		sleepNs := deficit / t.rate
		time.Sleep(time.Duration(sleepNs))
		t.tokens = 0
		t.lastTime = time.Now()
	}
}
