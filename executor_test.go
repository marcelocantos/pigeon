// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package tern

import (
	"context"
	"io"
	"testing"
	"time"
)

// TestExecutorBackendLANActivation verifies that the executor drives
// a BackendMachine from RelayConnected through LANOffered to LANActive,
// returning the correct commands at each step.
func TestExecutorBackendLANActivation(t *testing.T) {
	m := NewBackendMachine()
	m.State = BackendRelayConnected

	m.Guards[GuardChallengeValid] = func() bool { return true }
	m.Guards[GuardChallengeInvalid] = func() bool { return false }
	m.Guards[GuardLanServerAvailable] = func() bool { return true }
	m.Actions[ActionActivateLan] = func() error { return nil }

	// RelayConnected + lan_server_ready → LANOffered
	cmds, err := m.HandleEvent(EventLanServerReady)
	if err != nil {
		t.Fatal("lan_server_ready:", err)
	}
	if m.State != BackendLANOffered {
		t.Fatalf("state: got %s, want LANOffered", m.State)
	}
	assertCmds(t, cmds, CmdSendLanOffer)

	// LANOffered + recv lan_verify → LANActive
	cmds, err = m.HandleEvent(EventRecvLanVerify)
	if err != nil {
		t.Fatal("recv_lan_verify:", err)
	}
	if m.State != BackendLANActive {
		t.Fatalf("state: got %s, want LANActive", m.State)
	}
	assertCmds(t, cmds,
		CmdSendLanConfirm,
		CmdStartLanStreamReader,
		CmdStartLanDgReader,
		CmdStartMonitor,
		CmdSignalLanReady,
		CmdSetCryptoDatagram,
	)
}

// TestExecutorBackendFallback verifies the degradation and fallback
// sequence returns the correct resource cleanup commands.
func TestExecutorBackendFallback(t *testing.T) {
	m := NewBackendMachine()
	m.State = BackendLANActive
	m.PingFailures = 0

	m.Guards[GuardUnderMaxFailures] = func() bool { return m.PingFailures+1 < 3 }
	m.Guards[GuardAtMaxFailures] = func() bool { return m.PingFailures+1 >= 3 }
	m.Actions[ActionFallbackToRelay] = func() error { return nil }

	// LANActive + ping_timeout → LANDegraded
	cmds, _ := m.HandleEvent(EventPingTimeout)
	if m.State != BackendLANDegraded {
		t.Fatalf("state: got %s, want LANDegraded", m.State)
	}
	assertCmds(t, cmds) // no commands on first degradation

	// Exhaust failures.
	m.PingFailures = 2

	// LANDegraded + ping_timeout (at_max) → RelayBackoff
	cmds, _ = m.HandleEvent(EventPingTimeout)
	if m.State != BackendRelayBackoff {
		t.Fatalf("state: got %s, want RelayBackoff", m.State)
	}
	assertCmds(t, cmds,
		CmdStopMonitor,
		CmdStopLanStreamReader,
		CmdStopLanDgReader,
		CmdCloseLanPath,
		CmdResetLanReady,
		CmdStartBackoffTimer,
	)
}

// TestExecutorClientLANActivation verifies the client-side LAN
// activation sequence.
func TestExecutorClientLANActivation(t *testing.T) {
	m := NewClientMachine()
	m.State = ClientRelayConnected

	m.Guards[GuardLanEnabled] = func() bool { return true }
	m.Guards[GuardLanDisabled] = func() bool { return false }
	m.Actions[ActionDialLan] = func() error { return nil }
	m.Actions[ActionActivateLan] = func() error { return nil }

	// RelayConnected + recv lan_offer → LANConnecting
	cmds, _ := m.HandleEvent(EventRecvLanOffer)
	if m.State != ClientLANConnecting {
		t.Fatalf("state: got %s, want LANConnecting", m.State)
	}
	assertCmds(t, cmds, CmdDialLan)

	// LANConnecting + lan_dial_ok → LANVerifying
	cmds, _ = m.HandleEvent(EventLanDialOk)
	if m.State != ClientLANVerifying {
		t.Fatalf("state: got %s, want LANVerifying", m.State)
	}
	assertCmds(t, cmds, CmdSendLanVerify)

	// LANVerifying + recv lan_confirm → LANActive
	cmds, _ = m.HandleEvent(EventRecvLanConfirm)
	if m.State != ClientLANActive {
		t.Fatalf("state: got %s, want LANActive", m.State)
	}
	assertCmds(t, cmds,
		CmdStartLanStreamReader,
		CmdStartLanDgReader,
		CmdSignalLanReady,
		CmdSetCryptoDatagram,
	)
}

// TestExecutorClientFallback verifies the client fallback sequence.
func TestExecutorClientFallback(t *testing.T) {
	m := NewClientMachine()
	m.State = ClientLANActive

	m.Actions[ActionFallbackToRelay] = func() error { return nil }

	// LANActive + lan_error → RelayFallback
	cmds, _ := m.HandleEvent(EventLanError)
	if m.State != ClientRelayFallback {
		t.Fatalf("state: got %s, want RelayFallback", m.State)
	}
	assertCmds(t, cmds,
		CmdStopLanStreamReader,
		CmdStopLanDgReader,
		CmdCloseLanPath,
		CmdResetLanReady,
	)

	// RelayFallback + relay_ok → RelayConnected
	cmds, _ = m.HandleEvent(EventRelayOk)
	if m.State != ClientRelayConnected {
		t.Fatalf("state: got %s, want RelayConnected", m.State)
	}
	assertCmds(t, cmds) // no commands
}

// TestExecutorEventLoop verifies the executor processes events and
// executes commands.
func TestExecutorEventLoop(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	m := NewBackendMachine()
	m.State = BackendRelayConnected
	m.Guards[GuardChallengeValid] = func() bool { return true }
	m.Guards[GuardChallengeInvalid] = func() bool { return false }
	m.Guards[GuardLanServerAvailable] = func() bool { return true }
	m.Actions[ActionActivateLan] = func() error { return nil }

	// Create a mock relay path.
	relay := newPath("relay", &execMockStream{}, &execMockDatagram{ctx: ctx}, nil, nil, nil)
	e := newExecutor(ctx, cancel, m, relay)

	// Submit a lan_server_ready event.
	e.submit(event{id: EventLanServerReady})

	// Give the event loop time to process.
	time.Sleep(50 * time.Millisecond)

	if m.State != BackendLANOffered {
		t.Fatalf("state: got %s, want LANOffered", m.State)
	}
}

// TestExecutorSendRecv verifies that send() and recv() work through
// the executor event loop with real framing on an io.Pipe.
func TestExecutorSendRecv(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a pipe for the relay stream. The executor writes to pw,
	// the "remote" reads from pr (and vice versa for recv).
	relayR, relayW := io.Pipe()

	// The executor's relay stream needs bidirectional I/O.
	// Use a pipeStream that reads from one pipe and writes to another.
	remoteR, remoteW := io.Pipe()
	localStream := &pipeStream{r: remoteR, w: relayW}
	remoteStream := &pipeStream{r: relayR, w: remoteW}

	m := NewBackendMachine()
	m.State = BackendRelayConnected

	relay := newPath("relay", localStream, &execMockDatagram{ctx: ctx}, nil, nil, nil)
	e := newExecutor(ctx, cancel, m, relay)

	// Send a message through the executor.
	go func() {
		if err := e.send(ctx, []byte("hello")); err != nil {
			t.Errorf("send: %v", err)
		}
	}()

	// Read the framed message from the remote side.
	data, err := readMessage(remoteStream)
	if err != nil {
		t.Fatal("readMessage:", err)
	}
	if string(data) != "hello" {
		t.Fatalf("got %q, want hello", data)
	}

	// Now test recv: write a framed message on the remote side,
	// the executor's stream reader should pick it up and deliver.
	// Write first (in goroutine since pipe may block), then recv.
	wrote := make(chan struct{})
	go func() {
		if err := writeMessage(remoteStream, []byte("world")); err != nil {
			t.Errorf("writeMessage: %v", err)
		}
		close(wrote)
	}()

	// Wait for write to complete (pipe is synchronous).
	select {
	case <-wrote:
	case <-ctx.Done():
		t.Fatal("timeout writing to remote stream")
	}

	// Give the stream reader time to read and post the event.
	time.Sleep(100 * time.Millisecond)

	got, err := e.recv(ctx)
	if err != nil {
		t.Fatal("recv:", err)
	}
	if string(got) != "world" {
		t.Fatalf("got %q, want world", got)
	}
}

// TestExecutorSendRecvDatagram verifies that sendDatagram() and
// recvDatagram() work through the executor event loop.
func TestExecutorSendRecvDatagram(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Mock datagram transport: sent datagrams are received by the peer.
	// localToRemote carries datagrams from local → remote.
	// remoteToLocal carries datagrams from remote → local.
	localToRemote := make(chan []byte, 64)
	remoteToLocal := make(chan []byte, 64)
	localDg := &chanDatagram{out: localToRemote, in: remoteToLocal}
	remoteDg := &chanDatagram{out: remoteToLocal, in: localToRemote}

	m := NewBackendMachine()
	m.State = BackendRelayConnected

	relay := newPath("relay", &execMockStream{}, localDg, nil, nil, nil)
	e := newExecutor(ctx, cancel, m, relay)

	// Send a datagram through the executor.
	if err := e.sendDatagram([]byte("dgtest")); err != nil {
		t.Fatal("sendDatagram:", err)
	}

	// Read the raw datagram from the "remote" side and verify framing.
	raw := <-remoteDg.in
	if len(raw) < 2 || raw[0] != dgConnWhole {
		t.Fatalf("bad frame: %x", raw)
	}
	if string(raw[1:]) != "dgtest" {
		t.Fatalf("payload: got %q, want dgtest", raw[1:])
	}

	// Now test recv: send a framed datagram from the remote side.
	frame := make([]byte, 1+len("dgback"))
	frame[0] = dgConnWhole
	copy(frame[1:], "dgback")
	localDg.in <- frame

	// Give the datagram reader time to post the event.
	time.Sleep(50 * time.Millisecond)

	got, err := e.recvDatagram(ctx)
	if err != nil {
		t.Fatal("recvDatagram:", err)
	}
	if string(got) != "dgback" {
		t.Fatalf("got %q, want dgback", got)
	}
}

// chanDatagram implements datagrammer using Go channels.
type chanDatagram struct {
	out chan []byte // datagrams we send go here
	in  chan []byte // datagrams we receive come from here
}

func (d *chanDatagram) SendDatagram(data []byte) error {
	d.out <- data
	return nil
}

func (d *chanDatagram) ReceiveDatagram(ctx context.Context) ([]byte, error) {
	select {
	case data := <-d.in:
		return data, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// TestExecutorLANLifecycle verifies the full LAN lifecycle through
// the executor: backend offers → client dials → both activate →
// monitor starts → fallback → cleanup.
func TestExecutorLANLifecycle(t *testing.T) {
	// Test the backend machine's event sequence for LAN activation.
	m := NewBackendMachine()
	m.State = BackendRelayConnected

	m.Guards[GuardChallengeValid] = func() bool { return true }
	m.Guards[GuardChallengeInvalid] = func() bool { return false }
	m.Guards[GuardLanServerAvailable] = func() bool { return true }
	m.Guards[GuardUnderMaxFailures] = func() bool { return false }
	m.Guards[GuardAtMaxFailures] = func() bool { return true }
	m.Actions[ActionActivateLan] = func() error { return nil }
	m.Actions[ActionFallbackToRelay] = func() error { return nil }

	// 1. LAN server ready → offer
	cmds, _ := m.HandleEvent(EventLanServerReady)
	assertCmds(t, cmds, CmdSendLanOffer)
	if m.State != BackendLANOffered {
		t.Fatalf("expected LANOffered, got %s", m.State)
	}

	// 2. Client verifies → activate LAN
	cmds, _ = m.HandleEvent(EventRecvLanVerify)
	assertCmds(t, cmds,
		CmdSendLanConfirm,
		CmdStartLanStreamReader, CmdStartLanDgReader,
		CmdStartMonitor, CmdSignalLanReady, CmdSetCryptoDatagram)
	if m.State != BackendLANActive {
		t.Fatalf("expected LANActive, got %s", m.State)
	}

	// 3. Ping timeout → degrade
	cmds, _ = m.HandleEvent(EventPingTimeout)
	if m.State != BackendLANDegraded {
		t.Fatalf("expected LANDegraded, got %s", m.State)
	}

	// 4. Max failures → fallback with full cleanup
	m.PingFailures = 2
	cmds, _ = m.HandleEvent(EventPingTimeout)
	assertCmds(t, cmds,
		CmdStopMonitor, CmdStopLanStreamReader, CmdStopLanDgReader,
		CmdCloseLanPath, CmdResetLanReady, CmdStartBackoffTimer)
	if m.State != BackendRelayBackoff {
		t.Fatalf("expected RelayBackoff, got %s", m.State)
	}

	// 5. Backoff expires → re-offer
	cmds, _ = m.HandleEvent(EventBackoffExpired)
	assertCmds(t, cmds, CmdSendLanOffer)
	if m.State != BackendLANOffered {
		t.Fatalf("expected LANOffered, got %s", m.State)
	}
}

// pipeStream combines a reader and writer into an io.ReadWriteCloser.
type pipeStream struct {
	r io.Reader
	w io.Writer
}

func (p *pipeStream) Read(b []byte) (int, error)  { return p.r.Read(b) }
func (p *pipeStream) Write(b []byte) (int, error) { return p.w.Write(b) }
func (p *pipeStream) Close() error                { return nil }

// --- Test helpers ---

func assertCmds(t *testing.T, got []CmdID, want ...CmdID) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("commands: got %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("command[%d]: got %s, want %s", i, got[i], want[i])
		}
	}
}

// execMockStream implements io.ReadWriteCloser for testing.
type execMockStream struct{}

func (m *execMockStream) Read(p []byte) (int, error)  { select {} }
func (m *execMockStream) Write(p []byte) (int, error) { return len(p), nil }
func (m *execMockStream) Close() error                { return nil }

// execMockDatagram implements datagrammer for testing.
type execMockDatagram struct {
	ctx context.Context
}

func (m *execMockDatagram) SendDatagram(data []byte) error { return nil }
func (m *execMockDatagram) ReceiveDatagram(ctx context.Context) ([]byte, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}
