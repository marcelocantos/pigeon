// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Code generated from protocol/*.yaml. DO NOT EDIT.

package pairing

import (
	"github.com/marcelocantos/pigeon/protocol"
)

type (
	State      = protocol.State
	MsgType    = protocol.MsgType
	GuardID    = protocol.GuardID
	ActionID   = protocol.ActionID
	EventID    = protocol.EventID
	CmdID      = protocol.CmdID
	Protocol   = protocol.Protocol
	Actor      = protocol.Actor
	Transition = protocol.Transition
	Send       = protocol.Send
	Message    = protocol.Message
	VarDef     = protocol.VarDef
	VarUpdate  = protocol.VarUpdate
	GuardDef   = protocol.GuardDef
	Operator   = protocol.Operator
	AdvAction  = protocol.AdvAction
	Property   = protocol.Property
)

var (
	Recv      = protocol.Recv
	Internal  = protocol.Internal
	Invariant = protocol.Invariant
	Liveness  = protocol.Liveness
)

// PairingCeremonyProtocol acceptor states.
const (
	PairingCeremonyProtocolAcceptorIdle                State = "Idle"
	PairingCeremonyProtocolAcceptorGeneratingEphemeral State = "GeneratingEphemeral"
	PairingCeremonyProtocolAcceptorRegisteringRelay    State = "RegisteringRelay"
	PairingCeremonyProtocolAcceptorWaitingForHello     State = "WaitingForHello"
	PairingCeremonyProtocolAcceptorDerivingCode        State = "DerivingCode"
	PairingCeremonyProtocolAcceptorAwaitingUserConfirm State = "AwaitingUserConfirm"
	PairingCeremonyProtocolAcceptorAwaitingPeerConfirm State = "AwaitingPeerConfirm"
	PairingCeremonyProtocolAcceptorPaired              State = "Paired"
	PairingCeremonyProtocolAcceptorAborted             State = "Aborted"
)

// PairingCeremonyProtocol initiator states.
const (
	PairingCeremonyProtocolInitiatorIdle                State = "Idle"
	PairingCeremonyProtocolInitiatorDecodingToken       State = "DecodingToken"
	PairingCeremonyProtocolInitiatorGeneratingEphemeral State = "GeneratingEphemeral"
	PairingCeremonyProtocolInitiatorConnectingRelay     State = "ConnectingRelay"
	PairingCeremonyProtocolInitiatorAwaitingWelcome     State = "AwaitingWelcome"
	PairingCeremonyProtocolInitiatorDerivingCode        State = "DerivingCode"
	PairingCeremonyProtocolInitiatorAwaitingUserConfirm State = "AwaitingUserConfirm"
	PairingCeremonyProtocolInitiatorAwaitingPeerConfirm State = "AwaitingPeerConfirm"
	PairingCeremonyProtocolInitiatorPaired              State = "Paired"
	PairingCeremonyProtocolInitiatorAborted             State = "Aborted"
)

// PairingCeremonyProtocol message types.
const (
	PairingCeremonyProtocolMsgHello              MsgType = "hello"
	PairingCeremonyProtocolMsgWelcome            MsgType = "welcome"
	PairingCeremonyProtocolMsgConfirmToInitiator MsgType = "confirm_to_initiator"
	PairingCeremonyProtocolMsgConfirmToAcceptor  MsgType = "confirm_to_acceptor"
)

// PairingCeremonyProtocol guards.
const ()

// PairingCeremonyProtocol actions.
const (
	PairingCeremonyProtocolActionDecodeToken   ActionID = "decode_token"
	PairingCeremonyProtocolActionDeriveCode    ActionID = "derive_code"
	PairingCeremonyProtocolActionDialRelay     ActionID = "dial_relay"
	PairingCeremonyProtocolActionEmitToken     ActionID = "emit_token"
	PairingCeremonyProtocolActionGenEphemeral  ActionID = "gen_ephemeral"
	PairingCeremonyProtocolActionRegisterRelay ActionID = "register_relay"
	PairingCeremonyProtocolActionStoreRecord   ActionID = "store_record"
)

// PairingCeremonyProtocol events.
const (
	PairingCeremonyProtocolEventCodeReady              EventID = "code_ready"
	PairingCeremonyProtocolEventEphemeralReady         EventID = "ephemeral_ready"
	PairingCeremonyProtocolEventPairBegin              EventID = "pair_begin"
	PairingCeremonyProtocolEventRecvConfirmToAcceptor  EventID = "recv_confirm_to_acceptor"
	PairingCeremonyProtocolEventRecvConfirmToInitiator EventID = "recv_confirm_to_initiator"
	PairingCeremonyProtocolEventRecvHello              EventID = "recv_hello"
	PairingCeremonyProtocolEventRecvWelcome            EventID = "recv_welcome"
	PairingCeremonyProtocolEventRelayConnected         EventID = "relay_connected"
	PairingCeremonyProtocolEventRelayRegistered        EventID = "relay_registered"
	PairingCeremonyProtocolEventTokenDecoded           EventID = "token_decoded"
	PairingCeremonyProtocolEventTokenReceived          EventID = "token_received"
	PairingCeremonyProtocolEventUserCancel             EventID = "user_cancel"
	PairingCeremonyProtocolEventUserConfirm            EventID = "user_confirm"
)

func PairingCeremonyProtocol() *Protocol {
	return &Protocol{
		Name: "PairingCeremony",
		Actors: []Actor{
			{Name: "acceptor", Initial: "Idle", Transitions: []Transition{
				{From: "Idle", To: "GeneratingEphemeral", On: Internal("pair_begin"), Do: "gen_ephemeral", Updates: []VarUpdate{{Var: "acceptor_eph_pub", Expr: "\"acceptor_eph\""}}},
				{From: "GeneratingEphemeral", To: "RegisteringRelay", On: Internal("ephemeral_ready"), Do: "register_relay"},
				{From: "RegisteringRelay", To: "WaitingForHello", On: Internal("relay_registered"), Do: "emit_token"},
				{From: "WaitingForHello", To: "DerivingCode", On: Recv("hello"), Do: "derive_code", Updates: []VarUpdate{{Var: "acceptor_received_eph_pub", Expr: "recv_msg.eph_pub"}, {Var: "acceptor_received_identity", Expr: "recv_msg.identity_pub"}, {Var: "acceptor_received_instance", Expr: "recv_msg.instance_id"}, {Var: "acceptor_code", Expr: "DeriveCode(acceptor_eph_pub, recv_msg.eph_pub)"}}},
				{From: "DerivingCode", To: "AwaitingUserConfirm", On: Internal("code_ready"), Sends: []Send{{To: "initiator", Msg: "welcome", Fields: map[string]string{"eph_pub": "acceptor_eph_pub", "identity_pub": "acceptor_identity_pub", "instance_id": "acceptor_instance_id"}}}},
				{From: "AwaitingUserConfirm", To: "AwaitingPeerConfirm", On: Internal("user_confirm"), Sends: []Send{{To: "initiator", Msg: "confirm_to_initiator"}}, Updates: []VarUpdate{{Var: "acceptor_user_confirmed", Expr: "\"true\""}}},
				{From: "AwaitingPeerConfirm", To: "Paired", On: Recv("confirm_to_acceptor"), Do: "store_record", Updates: []VarUpdate{{Var: "acceptor_received_confirm", Expr: "\"true\""}}},
				{From: "AwaitingUserConfirm", To: "Aborted", On: Internal("user_cancel")},
				{From: "AwaitingPeerConfirm", To: "Aborted", On: Internal("user_cancel")},
			}},
			{Name: "initiator", Initial: "Idle", Transitions: []Transition{
				{From: "Idle", To: "DecodingToken", On: Internal("token_received"), Do: "decode_token", Updates: []VarUpdate{{Var: "received_acceptor_eph_pub", Expr: "\"acceptor_eph\""}, {Var: "received_acceptor_identity", Expr: "\"acceptor_id\""}, {Var: "received_acceptor_instance", Expr: "\"acceptor_instance\""}}},
				{From: "DecodingToken", To: "GeneratingEphemeral", On: Internal("token_decoded"), Do: "gen_ephemeral", Updates: []VarUpdate{{Var: "initiator_eph_pub", Expr: "\"initiator_eph\""}}},
				{From: "GeneratingEphemeral", To: "ConnectingRelay", On: Internal("ephemeral_ready"), Do: "dial_relay"},
				{From: "ConnectingRelay", To: "AwaitingWelcome", On: Internal("relay_connected"), Sends: []Send{{To: "acceptor", Msg: "hello", Fields: map[string]string{"eph_pub": "initiator_eph_pub", "identity_pub": "initiator_identity_pub", "instance_id": "initiator_instance_id"}}}},
				{From: "AwaitingWelcome", To: "DerivingCode", On: Recv("welcome"), Do: "derive_code", Updates: []VarUpdate{{Var: "initiator_received_eph_pub", Expr: "recv_msg.eph_pub"}, {Var: "initiator_received_identity", Expr: "recv_msg.identity_pub"}, {Var: "initiator_received_instance", Expr: "recv_msg.instance_id"}, {Var: "initiator_code", Expr: "DeriveCode(initiator_eph_pub, recv_msg.eph_pub)"}}},
				{From: "DerivingCode", To: "AwaitingUserConfirm", On: Internal("code_ready")},
				{From: "AwaitingUserConfirm", To: "AwaitingPeerConfirm", On: Internal("user_confirm"), Sends: []Send{{To: "acceptor", Msg: "confirm_to_acceptor"}}, Updates: []VarUpdate{{Var: "initiator_user_confirmed", Expr: "\"true\""}}},
				{From: "AwaitingPeerConfirm", To: "Paired", On: Recv("confirm_to_initiator"), Do: "store_record", Updates: []VarUpdate{{Var: "initiator_received_confirm", Expr: "\"true\""}}},
				{From: "AwaitingUserConfirm", To: "Aborted", On: Internal("user_cancel")},
				{From: "AwaitingPeerConfirm", To: "Aborted", On: Internal("user_cancel")},
			}},
		},
		Messages: []Message{
			{Type: "hello", From: "initiator", To: "acceptor", Desc: "JSON {kind: \"hello\", eph_pub, identity_pub, instance_id}"},
			{Type: "welcome", From: "acceptor", To: "initiator", Desc: "JSON {kind: \"welcome\", eph_pub, identity_pub, instance_id}"},
			{Type: "confirm_to_initiator", From: "acceptor", To: "initiator", Desc: "JSON {kind: \"confirm\"} — sent after acceptor's local user pressed y"},
			{Type: "confirm_to_acceptor", From: "initiator", To: "acceptor", Desc: "JSON {kind: \"confirm\"} — sent after initiator's local user pressed y"},
		},
		Vars: []VarDef{
			{Name: "acceptor_eph_pub", Initial: "\"none\"", Desc: "acceptor's ephemeral X25519 public key"},
			{Name: "acceptor_identity_pub", Initial: "\"acceptor_id\"", Desc: "acceptor's long-term Identity public key"},
			{Name: "acceptor_instance_id", Initial: "\"acceptor_instance\"", Desc: "acceptor's stable instance ID"},
			{Name: "acceptor_received_eph_pub", Initial: "\"none\"", Desc: "ephemeral pubkey acceptor saw in hello (may be adversary's)"},
			{Name: "acceptor_received_identity", Initial: "\"none\"", Desc: "identity pubkey acceptor saw in hello"},
			{Name: "acceptor_received_instance", Initial: "\"none\"", Desc: "instance ID acceptor saw in hello"},
			{Name: "acceptor_code", Initial: "<<\"none\">>", Desc: "confirmation code acceptor derived from its (ephA, ephB) view"},
			{Name: "acceptor_user_confirmed", Initial: "\"false\"", Desc: "has the acceptor's local human pressed y?"},
			{Name: "acceptor_received_confirm", Initial: "\"false\"", Desc: "has the acceptor received initiator's confirm message?"},
			{Name: "initiator_eph_pub", Initial: "\"none\"", Desc: "initiator's ephemeral X25519 public key"},
			{Name: "initiator_identity_pub", Initial: "\"initiator_id\"", Desc: "initiator's long-term Identity public key"},
			{Name: "initiator_instance_id", Initial: "\"initiator_instance\"", Desc: "initiator's stable instance ID"},
			{Name: "received_acceptor_eph_pub", Initial: "\"none\"", Desc: "acceptor ephemeral pubkey from token (trusted, out-of-band)"},
			{Name: "received_acceptor_identity", Initial: "\"none\"", Desc: "acceptor identity pubkey from token"},
			{Name: "received_acceptor_instance", Initial: "\"none\"", Desc: "acceptor instance ID from token"},
			{Name: "initiator_received_eph_pub", Initial: "\"none\"", Desc: "ephemeral pubkey initiator saw in welcome (may be adversary's)"},
			{Name: "initiator_received_identity", Initial: "\"none\"", Desc: "identity pubkey initiator saw in welcome"},
			{Name: "initiator_received_instance", Initial: "\"none\"", Desc: "instance ID initiator saw in welcome"},
			{Name: "initiator_code", Initial: "<<\"none\">>", Desc: "confirmation code initiator derived from its (ephA, ephB) view"},
			{Name: "initiator_user_confirmed", Initial: "\"false\"", Desc: "has the initiator's local human pressed y?"},
			{Name: "initiator_received_confirm", Initial: "\"false\"", Desc: "has the initiator received acceptor's confirm message?"},
			{Name: "adversary_keys", Initial: "{}", Desc: "ephemeral pubkeys the adversary has injected"},
			{Name: "adv_eph_pub", Initial: "\"adv_eph\"", Desc: "adversary's ephemeral X25519 public key"},
			{Name: "adv_saved_acceptor_eph", Initial: "\"none\"", Desc: "real acceptor eph pubkey saved during MitM substitution"},
			{Name: "adv_saved_initiator_eph", Initial: "\"none\"", Desc: "real initiator eph pubkey saved during MitM substitution"},
			{Name: "recv_msg", Initial: "[type |-> \"none\"]", Desc: "last received message (staging slot)"},
		},
		Guards: []GuardDef{},
		Operators: []Operator{
			{Name: "KeyRank", Params: "k", Expr: "CASE k = \"adv_eph\" -> 0 [] k = \"initiator_eph\" -> 1 [] k = \"acceptor_eph\" -> 2 [] OTHER -> 3", Desc: "Stable rank for symbolic pubkeys so DeriveCode is order-independent."},
			{Name: "DeriveCode", Params: "a, b", Expr: "IF KeyRank(a) <= KeyRank(b) THEN <<\"code\", a, b>> ELSE <<\"code\", b, a>>", Desc: "Confirmation code derived from both ephemeral pubkeys (order-independent)."},
		},
		AdvActions: []AdvAction{
			{Name: "MitM_hello", Desc: "intercept hello and substitute adversary ephemeral pubkey", Code: "      await Len(chan_initiator_acceptor) > 0 /\\ Head(chan_initiator_acceptor).type = MSG_hello;\n      adv_saved_initiator_eph := Head(chan_initiator_acceptor).eph_pub;\n      adversary_keys := adversary_keys \\union {adv_eph_pub};\n      chan_initiator_acceptor := <<[type |-> MSG_hello, eph_pub |-> adv_eph_pub, identity_pub |-> Head(chan_initiator_acceptor).identity_pub, instance_id |-> Head(chan_initiator_acceptor).instance_id]>> \\o Tail(chan_initiator_acceptor);"},
			{Name: "MitM_welcome", Desc: "intercept welcome and substitute adversary ephemeral pubkey", Code: "      await Len(chan_acceptor_initiator) > 0 /\\ Head(chan_acceptor_initiator).type = MSG_welcome;\n      adv_saved_acceptor_eph := Head(chan_acceptor_initiator).eph_pub;\n      adversary_keys := adversary_keys \\union {adv_eph_pub};\n      chan_acceptor_initiator := <<[type |-> MSG_welcome, eph_pub |-> adv_eph_pub, identity_pub |-> Head(chan_acceptor_initiator).identity_pub, instance_id |-> Head(chan_acceptor_initiator).instance_id]>> \\o Tail(chan_acceptor_initiator);"},
		},
		Properties: []Property{
			{Name: "MitMDetectedByCodeMismatch", Kind: Invariant, Expr: "(adv_eph_pub \\in adversary_keys /\\ acceptor_code /= <<\"none\">> /\\ initiator_code /= <<\"none\">>) => acceptor_code /= initiator_code", Desc: "If the adversary has injected its ephemeral pubkey AND both sides have computed codes, the codes differ — so the human comparison would catch the MitM."},
			{Name: "MitMPreventsPairing", Kind: Invariant, Expr: "(adv_eph_pub \\in adversary_keys /\\ acceptor_code /= initiator_code) => (acceptor_state /= acceptor_Paired \\/ initiator_state /= initiator_Paired)", Desc: "When codes differ (MitM detected), at least one side never reaches Paired (because the human cancels)."},
			{Name: "HonestPairingMatchesCodes", Kind: Invariant, Expr: "(adversary_keys = {} /\\ acceptor_code /= <<\"none\">> /\\ initiator_code /= <<\"none\">>) => acceptor_code = initiator_code", Desc: "Without adversary interference, both sides derive the same 6-digit code."},
			{Name: "HonestPairingCompletes", Kind: Liveness, Expr: "acceptor_state = acceptor_Paired /\\ initiator_state = initiator_Paired", Desc: "Without adversary interference and with both users pressing y, both sides eventually reach Paired."},
		},
		ChannelBound: 3,
		OneShot:      true,
	}
}

// PairingCeremonyProtocolAcceptorMachine is the generated state machine for the acceptor actor.
type PairingCeremonyProtocolAcceptorMachine struct {
	State                    State
	AcceptorEphPub           string // acceptor's ephemeral X25519 public key
	AcceptorReceivedEphPub   string // ephemeral pubkey acceptor saw in hello (may be adversary's)
	AcceptorReceivedIdentity string // identity pubkey acceptor saw in hello
	AcceptorReceivedInstance string // instance ID acceptor saw in hello
	AcceptorCode             string // confirmation code acceptor derived from its (ephA, ephB) view
	AcceptorUserConfirmed    string // has the acceptor's local human pressed y?
	AcceptorReceivedConfirm  string // has the acceptor received initiator's confirm message?

	Guards   map[GuardID]func() bool
	Actions  map[ActionID]func() error
	OnChange func(varName string)
}

func NewPairingCeremonyProtocolAcceptorMachine() *PairingCeremonyProtocolAcceptorMachine {
	return &PairingCeremonyProtocolAcceptorMachine{
		State:                    PairingCeremonyProtocolAcceptorIdle,
		AcceptorEphPub:           "none",
		AcceptorReceivedEphPub:   "none",
		AcceptorReceivedIdentity: "none",
		AcceptorReceivedInstance: "none",
		AcceptorCode:             "",
		AcceptorUserConfirmed:    "false",
		AcceptorReceivedConfirm:  "false",
		Guards:                   make(map[GuardID]func() bool),
		Actions:                  make(map[ActionID]func() error),
	}
}

func (m *PairingCeremonyProtocolAcceptorMachine) HandleMessage(msg MsgType) (bool, error) {
	switch {
	case m.State == PairingCeremonyProtocolAcceptorWaitingForHello && msg == PairingCeremonyProtocolMsgHello:
		if fn := m.Actions[PairingCeremonyProtocolActionDeriveCode]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		// acceptor_received_eph_pub: recv_msg.eph_pub (set by action)
		// acceptor_received_identity: recv_msg.identity_pub (set by action)
		// acceptor_received_instance: recv_msg.instance_id (set by action)
		// acceptor_code: DeriveCode(acceptor_eph_pub, recv_msg.eph_pub) (set by action)
		m.State = PairingCeremonyProtocolAcceptorDerivingCode
		return true, nil
	case m.State == PairingCeremonyProtocolAcceptorAwaitingPeerConfirm && msg == PairingCeremonyProtocolMsgConfirmToAcceptor:
		if fn := m.Actions[PairingCeremonyProtocolActionStoreRecord]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		m.AcceptorReceivedConfirm = "true"
		if m.OnChange != nil {
			m.OnChange("acceptor_received_confirm")
		}
		m.State = PairingCeremonyProtocolAcceptorPaired
		return true, nil
	}
	return false, nil
}

func (m *PairingCeremonyProtocolAcceptorMachine) Step(event EventID) (bool, error) {
	switch {
	case m.State == PairingCeremonyProtocolAcceptorIdle && event == PairingCeremonyProtocolEventPairBegin:
		if fn := m.Actions[PairingCeremonyProtocolActionGenEphemeral]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		m.AcceptorEphPub = "acceptor_eph"
		if m.OnChange != nil {
			m.OnChange("acceptor_eph_pub")
		}
		m.State = PairingCeremonyProtocolAcceptorGeneratingEphemeral
		return true, nil
	case m.State == PairingCeremonyProtocolAcceptorGeneratingEphemeral && event == PairingCeremonyProtocolEventEphemeralReady:
		if fn := m.Actions[PairingCeremonyProtocolActionRegisterRelay]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		m.State = PairingCeremonyProtocolAcceptorRegisteringRelay
		return true, nil
	case m.State == PairingCeremonyProtocolAcceptorRegisteringRelay && event == PairingCeremonyProtocolEventRelayRegistered:
		if fn := m.Actions[PairingCeremonyProtocolActionEmitToken]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		m.State = PairingCeremonyProtocolAcceptorWaitingForHello
		return true, nil
	case m.State == PairingCeremonyProtocolAcceptorDerivingCode && event == PairingCeremonyProtocolEventCodeReady:
		m.State = PairingCeremonyProtocolAcceptorAwaitingUserConfirm
		return true, nil
	case m.State == PairingCeremonyProtocolAcceptorAwaitingUserConfirm && event == PairingCeremonyProtocolEventUserConfirm:
		m.AcceptorUserConfirmed = "true"
		if m.OnChange != nil {
			m.OnChange("acceptor_user_confirmed")
		}
		m.State = PairingCeremonyProtocolAcceptorAwaitingPeerConfirm
		return true, nil
	case m.State == PairingCeremonyProtocolAcceptorAwaitingUserConfirm && event == PairingCeremonyProtocolEventUserCancel:
		m.State = PairingCeremonyProtocolAcceptorAborted
		return true, nil
	case m.State == PairingCeremonyProtocolAcceptorAwaitingPeerConfirm && event == PairingCeremonyProtocolEventUserCancel:
		m.State = PairingCeremonyProtocolAcceptorAborted
		return true, nil
	}
	return false, nil
}

func (m *PairingCeremonyProtocolAcceptorMachine) HandleEvent(ev EventID) ([]CmdID, error) {
	switch {
	case m.State == PairingCeremonyProtocolAcceptorIdle && ev == PairingCeremonyProtocolEventPairBegin:
		if fn := m.Actions[PairingCeremonyProtocolActionGenEphemeral]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		m.AcceptorEphPub = "acceptor_eph"
		if m.OnChange != nil {
			m.OnChange("acceptor_eph_pub")
		}
		m.State = PairingCeremonyProtocolAcceptorGeneratingEphemeral
		return nil, nil
	case m.State == PairingCeremonyProtocolAcceptorGeneratingEphemeral && ev == PairingCeremonyProtocolEventEphemeralReady:
		if fn := m.Actions[PairingCeremonyProtocolActionRegisterRelay]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		m.State = PairingCeremonyProtocolAcceptorRegisteringRelay
		return nil, nil
	case m.State == PairingCeremonyProtocolAcceptorRegisteringRelay && ev == PairingCeremonyProtocolEventRelayRegistered:
		if fn := m.Actions[PairingCeremonyProtocolActionEmitToken]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		m.State = PairingCeremonyProtocolAcceptorWaitingForHello
		return nil, nil
	case m.State == PairingCeremonyProtocolAcceptorWaitingForHello && ev == PairingCeremonyProtocolEventRecvHello:
		if fn := m.Actions[PairingCeremonyProtocolActionDeriveCode]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		// acceptor_received_eph_pub: recv_msg.eph_pub (set by action)
		// acceptor_received_identity: recv_msg.identity_pub (set by action)
		// acceptor_received_instance: recv_msg.instance_id (set by action)
		// acceptor_code: DeriveCode(acceptor_eph_pub, recv_msg.eph_pub) (set by action)
		m.State = PairingCeremonyProtocolAcceptorDerivingCode
		return nil, nil
	case m.State == PairingCeremonyProtocolAcceptorDerivingCode && ev == PairingCeremonyProtocolEventCodeReady:
		m.State = PairingCeremonyProtocolAcceptorAwaitingUserConfirm
		return nil, nil
	case m.State == PairingCeremonyProtocolAcceptorAwaitingUserConfirm && ev == PairingCeremonyProtocolEventUserConfirm:
		m.AcceptorUserConfirmed = "true"
		if m.OnChange != nil {
			m.OnChange("acceptor_user_confirmed")
		}
		m.State = PairingCeremonyProtocolAcceptorAwaitingPeerConfirm
		return nil, nil
	case m.State == PairingCeremonyProtocolAcceptorAwaitingPeerConfirm && ev == PairingCeremonyProtocolEventRecvConfirmToAcceptor:
		if fn := m.Actions[PairingCeremonyProtocolActionStoreRecord]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		m.AcceptorReceivedConfirm = "true"
		if m.OnChange != nil {
			m.OnChange("acceptor_received_confirm")
		}
		m.State = PairingCeremonyProtocolAcceptorPaired
		return nil, nil
	case m.State == PairingCeremonyProtocolAcceptorAwaitingUserConfirm && ev == PairingCeremonyProtocolEventUserCancel:
		m.State = PairingCeremonyProtocolAcceptorAborted
		return nil, nil
	case m.State == PairingCeremonyProtocolAcceptorAwaitingPeerConfirm && ev == PairingCeremonyProtocolEventUserCancel:
		m.State = PairingCeremonyProtocolAcceptorAborted
		return nil, nil
	}
	return nil, nil
}

// PairingCeremonyProtocolInitiatorMachine is the generated state machine for the initiator actor.
type PairingCeremonyProtocolInitiatorMachine struct {
	State                     State
	InitiatorEphPub           string // initiator's ephemeral X25519 public key
	ReceivedAcceptorEphPub    string // acceptor ephemeral pubkey from token (trusted, out-of-band)
	ReceivedAcceptorIdentity  string // acceptor identity pubkey from token
	ReceivedAcceptorInstance  string // acceptor instance ID from token
	InitiatorReceivedEphPub   string // ephemeral pubkey initiator saw in welcome (may be adversary's)
	InitiatorReceivedIdentity string // identity pubkey initiator saw in welcome
	InitiatorReceivedInstance string // instance ID initiator saw in welcome
	InitiatorCode             string // confirmation code initiator derived from its (ephA, ephB) view
	InitiatorUserConfirmed    string // has the initiator's local human pressed y?
	InitiatorReceivedConfirm  string // has the initiator received acceptor's confirm message?

	Guards   map[GuardID]func() bool
	Actions  map[ActionID]func() error
	OnChange func(varName string)
}

func NewPairingCeremonyProtocolInitiatorMachine() *PairingCeremonyProtocolInitiatorMachine {
	return &PairingCeremonyProtocolInitiatorMachine{
		State:                     PairingCeremonyProtocolInitiatorIdle,
		InitiatorEphPub:           "none",
		ReceivedAcceptorEphPub:    "none",
		ReceivedAcceptorIdentity:  "none",
		ReceivedAcceptorInstance:  "none",
		InitiatorReceivedEphPub:   "none",
		InitiatorReceivedIdentity: "none",
		InitiatorReceivedInstance: "none",
		InitiatorCode:             "",
		InitiatorUserConfirmed:    "false",
		InitiatorReceivedConfirm:  "false",
		Guards:                    make(map[GuardID]func() bool),
		Actions:                   make(map[ActionID]func() error),
	}
}

func (m *PairingCeremonyProtocolInitiatorMachine) HandleMessage(msg MsgType) (bool, error) {
	switch {
	case m.State == PairingCeremonyProtocolInitiatorAwaitingWelcome && msg == PairingCeremonyProtocolMsgWelcome:
		if fn := m.Actions[PairingCeremonyProtocolActionDeriveCode]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		// initiator_received_eph_pub: recv_msg.eph_pub (set by action)
		// initiator_received_identity: recv_msg.identity_pub (set by action)
		// initiator_received_instance: recv_msg.instance_id (set by action)
		// initiator_code: DeriveCode(initiator_eph_pub, recv_msg.eph_pub) (set by action)
		m.State = PairingCeremonyProtocolInitiatorDerivingCode
		return true, nil
	case m.State == PairingCeremonyProtocolInitiatorAwaitingPeerConfirm && msg == PairingCeremonyProtocolMsgConfirmToInitiator:
		if fn := m.Actions[PairingCeremonyProtocolActionStoreRecord]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		m.InitiatorReceivedConfirm = "true"
		if m.OnChange != nil {
			m.OnChange("initiator_received_confirm")
		}
		m.State = PairingCeremonyProtocolInitiatorPaired
		return true, nil
	}
	return false, nil
}

func (m *PairingCeremonyProtocolInitiatorMachine) Step(event EventID) (bool, error) {
	switch {
	case m.State == PairingCeremonyProtocolInitiatorIdle && event == PairingCeremonyProtocolEventTokenReceived:
		if fn := m.Actions[PairingCeremonyProtocolActionDecodeToken]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		m.ReceivedAcceptorEphPub = "acceptor_eph"
		if m.OnChange != nil {
			m.OnChange("received_acceptor_eph_pub")
		}
		m.ReceivedAcceptorIdentity = "acceptor_id"
		if m.OnChange != nil {
			m.OnChange("received_acceptor_identity")
		}
		m.ReceivedAcceptorInstance = "acceptor_instance"
		if m.OnChange != nil {
			m.OnChange("received_acceptor_instance")
		}
		m.State = PairingCeremonyProtocolInitiatorDecodingToken
		return true, nil
	case m.State == PairingCeremonyProtocolInitiatorDecodingToken && event == PairingCeremonyProtocolEventTokenDecoded:
		if fn := m.Actions[PairingCeremonyProtocolActionGenEphemeral]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		m.InitiatorEphPub = "initiator_eph"
		if m.OnChange != nil {
			m.OnChange("initiator_eph_pub")
		}
		m.State = PairingCeremonyProtocolInitiatorGeneratingEphemeral
		return true, nil
	case m.State == PairingCeremonyProtocolInitiatorGeneratingEphemeral && event == PairingCeremonyProtocolEventEphemeralReady:
		if fn := m.Actions[PairingCeremonyProtocolActionDialRelay]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		m.State = PairingCeremonyProtocolInitiatorConnectingRelay
		return true, nil
	case m.State == PairingCeremonyProtocolInitiatorConnectingRelay && event == PairingCeremonyProtocolEventRelayConnected:
		m.State = PairingCeremonyProtocolInitiatorAwaitingWelcome
		return true, nil
	case m.State == PairingCeremonyProtocolInitiatorDerivingCode && event == PairingCeremonyProtocolEventCodeReady:
		m.State = PairingCeremonyProtocolInitiatorAwaitingUserConfirm
		return true, nil
	case m.State == PairingCeremonyProtocolInitiatorAwaitingUserConfirm && event == PairingCeremonyProtocolEventUserConfirm:
		m.InitiatorUserConfirmed = "true"
		if m.OnChange != nil {
			m.OnChange("initiator_user_confirmed")
		}
		m.State = PairingCeremonyProtocolInitiatorAwaitingPeerConfirm
		return true, nil
	case m.State == PairingCeremonyProtocolInitiatorAwaitingUserConfirm && event == PairingCeremonyProtocolEventUserCancel:
		m.State = PairingCeremonyProtocolInitiatorAborted
		return true, nil
	case m.State == PairingCeremonyProtocolInitiatorAwaitingPeerConfirm && event == PairingCeremonyProtocolEventUserCancel:
		m.State = PairingCeremonyProtocolInitiatorAborted
		return true, nil
	}
	return false, nil
}

func (m *PairingCeremonyProtocolInitiatorMachine) HandleEvent(ev EventID) ([]CmdID, error) {
	switch {
	case m.State == PairingCeremonyProtocolInitiatorIdle && ev == PairingCeremonyProtocolEventTokenReceived:
		if fn := m.Actions[PairingCeremonyProtocolActionDecodeToken]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		m.ReceivedAcceptorEphPub = "acceptor_eph"
		if m.OnChange != nil {
			m.OnChange("received_acceptor_eph_pub")
		}
		m.ReceivedAcceptorIdentity = "acceptor_id"
		if m.OnChange != nil {
			m.OnChange("received_acceptor_identity")
		}
		m.ReceivedAcceptorInstance = "acceptor_instance"
		if m.OnChange != nil {
			m.OnChange("received_acceptor_instance")
		}
		m.State = PairingCeremonyProtocolInitiatorDecodingToken
		return nil, nil
	case m.State == PairingCeremonyProtocolInitiatorDecodingToken && ev == PairingCeremonyProtocolEventTokenDecoded:
		if fn := m.Actions[PairingCeremonyProtocolActionGenEphemeral]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		m.InitiatorEphPub = "initiator_eph"
		if m.OnChange != nil {
			m.OnChange("initiator_eph_pub")
		}
		m.State = PairingCeremonyProtocolInitiatorGeneratingEphemeral
		return nil, nil
	case m.State == PairingCeremonyProtocolInitiatorGeneratingEphemeral && ev == PairingCeremonyProtocolEventEphemeralReady:
		if fn := m.Actions[PairingCeremonyProtocolActionDialRelay]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		m.State = PairingCeremonyProtocolInitiatorConnectingRelay
		return nil, nil
	case m.State == PairingCeremonyProtocolInitiatorConnectingRelay && ev == PairingCeremonyProtocolEventRelayConnected:
		m.State = PairingCeremonyProtocolInitiatorAwaitingWelcome
		return nil, nil
	case m.State == PairingCeremonyProtocolInitiatorAwaitingWelcome && ev == PairingCeremonyProtocolEventRecvWelcome:
		if fn := m.Actions[PairingCeremonyProtocolActionDeriveCode]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		// initiator_received_eph_pub: recv_msg.eph_pub (set by action)
		// initiator_received_identity: recv_msg.identity_pub (set by action)
		// initiator_received_instance: recv_msg.instance_id (set by action)
		// initiator_code: DeriveCode(initiator_eph_pub, recv_msg.eph_pub) (set by action)
		m.State = PairingCeremonyProtocolInitiatorDerivingCode
		return nil, nil
	case m.State == PairingCeremonyProtocolInitiatorDerivingCode && ev == PairingCeremonyProtocolEventCodeReady:
		m.State = PairingCeremonyProtocolInitiatorAwaitingUserConfirm
		return nil, nil
	case m.State == PairingCeremonyProtocolInitiatorAwaitingUserConfirm && ev == PairingCeremonyProtocolEventUserConfirm:
		m.InitiatorUserConfirmed = "true"
		if m.OnChange != nil {
			m.OnChange("initiator_user_confirmed")
		}
		m.State = PairingCeremonyProtocolInitiatorAwaitingPeerConfirm
		return nil, nil
	case m.State == PairingCeremonyProtocolInitiatorAwaitingPeerConfirm && ev == PairingCeremonyProtocolEventRecvConfirmToInitiator:
		if fn := m.Actions[PairingCeremonyProtocolActionStoreRecord]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		m.InitiatorReceivedConfirm = "true"
		if m.OnChange != nil {
			m.OnChange("initiator_received_confirm")
		}
		m.State = PairingCeremonyProtocolInitiatorPaired
		return nil, nil
	case m.State == PairingCeremonyProtocolInitiatorAwaitingUserConfirm && ev == PairingCeremonyProtocolEventUserCancel:
		m.State = PairingCeremonyProtocolInitiatorAborted
		return nil, nil
	case m.State == PairingCeremonyProtocolInitiatorAwaitingPeerConfirm && ev == PairingCeremonyProtocolEventUserCancel:
		m.State = PairingCeremonyProtocolInitiatorAborted
		return nil, nil
	}
	return nil, nil
}
