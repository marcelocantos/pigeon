// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package backchannel

import (
	"context"
	"errors"
	"fmt"
)

// EventSink is the executor's inbound event hook. The pairing state
// machine consumes events from many transports — relay streams, relay
// datagrams, and (with this package) the local backchannel — through
// the same interface. EventSink is a plain function value, not an
// interface, so that callers can adapt it from any closure that
// drives their executor.
//
// The Name argument is the event identifier in the protogen-generated
// event vocabulary (e.g. SessionProtocolEventCliInitPair). The Payload
// carries any per-event fields the transition needs; the backchannel
// adapter sets it to the originating Message verbatim so the executor
// can pull out InstanceID / Token / Code as required.
type EventSink func(ctx context.Context, name string, payload Message) error

// ServeSession reads frames from sess in a loop and feeds them into
// sink as executor events. Returns when sess is closed, ctx is
// cancelled, or sink returns an error. Errors from sink are propagated
// because they signal the executor has rejected the event (e.g. wrong
// state) and the daemon must decide policy.
//
// Outbound traffic (daemon → CLI) is sent directly through sess.Send by
// the daemon's normal action emitters; ServeSession only handles the
// inbound direction.
func ServeSession(ctx context.Context, sess *Session, sink EventSink) error {
	if sess == nil {
		return errors.New("backchannel.ServeSession: sess is nil")
	}
	if sink == nil {
		return errors.New("backchannel.ServeSession: sink is nil")
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		msg, err := sess.Recv(ctx)
		if err != nil {
			return err
		}
		name, ok := EventForMessage(msg.Type)
		if !ok {
			return fmt.Errorf("backchannel: unknown message type %q from CLI", msg.Type)
		}
		if err := sink(ctx, name, msg); err != nil {
			return fmt.Errorf("backchannel dispatch %s: %w", name, err)
		}
	}
}

// EventForMessage maps a backchannel MsgType to the protogen event
// name the executor expects. Returns ok=false for messages that the
// CLI must not originate (everything except pair_begin and
// code_submit). The reverse direction — daemon → CLI — is handled by
// the daemon emitting the Message directly via Session.Send when the
// machine reaches the corresponding state; that mapping is in
// MessageForEvent.
func EventForMessage(t MsgType) (string, bool) {
	switch t {
	case MsgPairBegin:
		return EventCliInitPair, true
	case MsgCodeSubmit:
		return EventCliCodeEntered, true
	default:
		return "", false
	}
}

// MessageForEvent is the daemon→CLI mirror of EventForMessage. The
// daemon's action emitter for a "tell the CLI X" transition calls this
// to obtain the wire MsgType for the event the machine just fired.
// Returns ok=false for events that do not cross the backchannel.
//
// pair_status is synthesised by the daemon when the machine reaches a
// terminal state (Paired or Aborted); the session protocol has no
// single event named pair_status, so callers map terminal-state entry
// to Send(Message{Type: MsgPairStatus, Status: "paired"}) themselves.
func MessageForEvent(name string) (MsgType, bool) {
	switch name {
	case EventTokenCreated:
		return MsgTokenResponse, true
	case EventSignalCodeDisplay:
		return MsgWaitingForCode, true
	default:
		return "", false
	}
}

// Event names mirror the protogen-generated SessionProtocol* /
// Session* constants used by the executor. Declared here as plain
// string constants so the backchannel package does not have to import
// the session-protocol generated code (which would create a cycle if
// the session protocol later references backchannel for transport
// selection). Keep these in lockstep with protocol/session.yaml.
const (
	EventCliInitPair       = "cli_init_pair"
	EventCliCodeEntered    = "cli_code_entered"
	EventTokenCreated      = "token_created"
	EventSignalCodeDisplay = "signal_code_display"
)
