// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon_test

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/marcelocantos/pigeon"
	"github.com/marcelocantos/pigeon/crypto"
)

func pairedSession(t *testing.T, streams []string) (backend, client *pigeon.Session, ctx context.Context, cancel context.CancelFunc) {
	t.Helper()
	relayURL, teardown := startRelay(t)
	t.Cleanup(teardown)
	time.Sleep(200 * time.Millisecond)

	bid, cid, brec, crec := pairingPair(t, relayURL)
	pairings := map[string]*crypto.PairingRecord{cid.InstanceID(): brec}
	ctx, cancel = context.WithTimeout(context.Background(), 20*time.Second)
	t.Cleanup(cancel)
	tlsCfg := &tls.Config{InsecureSkipVerify: true}

	listener, _, err := pigeon.Register(ctx, &pigeon.RegisterArgs{
		Identity: bid,
		Pairing: func(id string) (*crypto.PairingRecord, error) {
			rec, ok := pairings[id]
			if !ok {
				return nil, fmt.Errorf("unknown client %q", id)
			}
			return rec, nil
		},
		Relay:   relayURL,
		TLS:     tlsCfg,
		Streams: streams,
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	got := make(chan *pigeon.Session, 1)
	errCh := make(chan error, 1)
	go func() {
		sess, err := listener.Accept(ctx)
		if err != nil {
			errCh <- err
			return
		}
		got <- sess
	}()

	cs, err := pigeon.Connect(ctx, &pigeon.ConnectArgs{
		InstanceID: bid.InstanceID(),
		Record:     crec,
		Identity:   cid,
		Relay:      relayURL,
		TLS:        tlsCfg,
		Streams:    streams,
	})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	select {
	case backend = <-got:
	case err := <-errCh:
		t.Fatalf("accept: %v", err)
	case <-ctx.Done():
		t.Fatal("accept timeout")
	}
	t.Cleanup(func() { _ = backend.Close(); _ = cs.Close() })
	return backend, cs, ctx, cancel
}

func TestE2EUnmatchedStreamReset(t *testing.T) {
	t.Parallel()
	backend, client, ctx, _ := pairedSession(t, nil)

	chatC, err := client.OpenStream(ctx, "chat")
	if err != nil {
		t.Fatalf("open chat: %v", err)
	}
	chatB, err := backend.AcceptStream(ctx, "chat")
	if err != nil {
		t.Fatalf("accept chat: %v", err)
	}
	if err := chatC.Send([]byte("ok")); err != nil {
		t.Fatalf("chat send: %v", err)
	}
	if _, err := chatB.Recv(ctx); err != nil {
		t.Fatalf("chat recv: %v", err)
	}

	for _, name := range []string{"nope", "nope2", "nope3"} {
		bad, err := client.OpenStream(ctx, name)
		if err != nil {
			t.Fatalf("open unmatched %q: %v", name, err)
		}
		rctx, rcancel := context.WithTimeout(ctx, pigeon.DefaultStreamRendezvous+2*time.Second)
		if _, err := bad.Recv(rctx); err == nil {
			rcancel()
			t.Fatalf("unmatched stream %q Recv succeeded; want reset/EOF", name)
		}
		rcancel()
	}
	if n := backend.BufferedStreamCount(); n != 0 {
		t.Fatalf("backend parked %d unmatched streams", n)
	}
}

func TestE2EStreamAdmissionAllowlist(t *testing.T) {
	t.Parallel()
	backend, client, ctx, _ := pairedSession(t, []string{"chat"})

	chatC, err := client.OpenStream(ctx, "chat")
	if err != nil {
		t.Fatalf("open chat: %v", err)
	}
	chatB, err := backend.AcceptStream(ctx, "chat")
	if err != nil {
		t.Fatalf("accept chat: %v", err)
	}
	if err := chatC.Send([]byte("ok")); err != nil {
		t.Fatalf("chat send: %v", err)
	}
	if _, err := chatB.Recv(ctx); err != nil {
		t.Fatalf("chat recv: %v", err)
	}

	bad, err := client.OpenStream(ctx, "nope")
	if err != nil {
		t.Fatalf("open undeclared: %v", err)
	}
	rctx, rcancel := context.WithTimeout(ctx, 3*time.Second)
	defer rcancel()
	if _, err := bad.Recv(rctx); err == nil {
		t.Fatal("undeclared stream Recv succeeded; want immediate reset")
	}
	if n := backend.BufferedStreamCount(); n != 0 {
		t.Fatalf("backend parked undeclared stream (count=%d)", n)
	}
}

func TestE2EStreamCloseWriteRPC(t *testing.T) {
	t.Parallel()
	backend, client, ctx, _ := pairedSession(t, nil)

	errCh := make(chan error, 1)
	go func() {
		st, err := backend.AcceptStream(ctx, "rpc")
		if err != nil {
			errCh <- err
			return
		}
		req, err := st.Recv(ctx)
		if err != nil {
			errCh <- fmt.Errorf("backend recv: %w", err)
			return
		}
		if string(req) != "ping" {
			errCh <- fmt.Errorf("got %q", req)
			return
		}
		errCh <- st.Send([]byte("pong"))
	}()

	st, err := client.OpenStream(ctx, "rpc")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := st.Send([]byte("ping")); err != nil {
		t.Fatalf("send: %v", err)
	}
	if err := st.CloseWrite(); err != nil {
		t.Fatalf("CloseWrite: %v", err)
	}
	got, err := st.Recv(ctx)
	if err != nil {
		t.Fatalf("recv reply: %v", err)
	}
	if string(got) != "pong" {
		t.Fatalf("reply %q", got)
	}
	if err := <-errCh; err != nil {
		t.Fatal(err)
	}
}

func TestE2EStreamResetAbortsPeerRecv(t *testing.T) {
	t.Parallel()
	backend, client, ctx, _ := pairedSession(t, nil)

	errCh := make(chan error, 1)
	go func() {
		st, err := backend.AcceptStream(ctx, "abort")
		if err != nil {
			errCh <- err
			return
		}
		_, err = st.Recv(ctx)
		if err == nil {
			errCh <- errors.New("Recv succeeded after Reset")
			return
		}
		if !pigeon.IsStreamReset(err) {
			errCh <- fmt.Errorf("Recv error is not a stream reset: %v", err)
			return
		}
		errCh <- nil
	}()

	st, err := client.OpenStream(ctx, "abort")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	// Give the backend Recv a chance to park, then abort.
	time.Sleep(100 * time.Millisecond)
	if err := st.Reset(); err != nil {
		t.Fatalf("Reset: %v", err)
	}
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("peer Recv did not return after Reset")
	}
}

func TestE2EGoAwayDrainsThenRefuses(t *testing.T) {
	t.Parallel()
	backend, client, ctx, _ := pairedSession(t, nil)

	chatC, err := client.OpenStream(ctx, "chat")
	if err != nil {
		t.Fatalf("open chat: %v", err)
	}
	chatB, err := backend.AcceptStream(ctx, "chat")
	if err != nil {
		t.Fatalf("accept chat: %v", err)
	}

	if err := backend.GoAway(); err != nil {
		t.Fatal(err)
	}
	// Give the go-away control stream time to land.
	time.Sleep(200 * time.Millisecond)

	st, err := client.OpenStream(ctx, "x")
	if err == nil {
		rctx, rcancel := context.WithTimeout(ctx, 3*time.Second)
		defer rcancel()
		if _, err := st.Recv(rctx); err == nil {
			t.Fatal("stream opened after GoAway was not reset")
		}
	} else if !errors.Is(err, pigeon.ErrSessionDraining) {
		t.Logf("OpenStream after GoAway: %v", err)
	}

	if err := chatC.Send([]byte("still")); err != nil {
		t.Fatalf("in-flight send after GoAway: %v", err)
	}
	got, err := chatB.Recv(ctx)
	if err != nil {
		t.Fatalf("in-flight recv after GoAway: %v", err)
	}
	if string(got) != "still" {
		t.Fatalf("got %q", got)
	}
	if err := backend.Close(); err != nil {
		t.Fatal(err)
	}
}
