// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Code generated from protocol/*.yaml. DO NOT EDIT.

package protocol

// PairingCeremony acceptor states.
const (
	PairingCeremonyAcceptorIdle                State = "Idle"
	PairingCeremonyAcceptorGeneratingEphemeral State = "GeneratingEphemeral"
	PairingCeremonyAcceptorRegisteringRelay    State = "RegisteringRelay"
	PairingCeremonyAcceptorWaitingForHello     State = "WaitingForHello"
	PairingCeremonyAcceptorDerivingCode        State = "DerivingCode"
	PairingCeremonyAcceptorAwaitingUserConfirm State = "AwaitingUserConfirm"
	PairingCeremonyAcceptorAwaitingPeerConfirm State = "AwaitingPeerConfirm"
	PairingCeremonyAcceptorPaired              State = "Paired"
	PairingCeremonyAcceptorAborted             State = "Aborted"
)

// PairingCeremony initiator states.
const (
	PairingCeremonyInitiatorIdle                State = "Idle"
	PairingCeremonyInitiatorDecodingToken       State = "DecodingToken"
	PairingCeremonyInitiatorGeneratingEphemeral State = "GeneratingEphemeral"
	PairingCeremonyInitiatorConnectingRelay     State = "ConnectingRelay"
	PairingCeremonyInitiatorAwaitingWelcome     State = "AwaitingWelcome"
	PairingCeremonyInitiatorDerivingCode        State = "DerivingCode"
	PairingCeremonyInitiatorAwaitingUserConfirm State = "AwaitingUserConfirm"
	PairingCeremonyInitiatorAwaitingPeerConfirm State = "AwaitingPeerConfirm"
	PairingCeremonyInitiatorPaired              State = "Paired"
	PairingCeremonyInitiatorAborted             State = "Aborted"
)

// PairingCeremony message types.
const (
	PairingCeremonyMsgHello              MsgType = "hello"
	PairingCeremonyMsgWelcome            MsgType = "welcome"
	PairingCeremonyMsgConfirmToInitiator MsgType = "confirm_to_initiator"
	PairingCeremonyMsgConfirmToAcceptor  MsgType = "confirm_to_acceptor"
)

// PairingCeremony guards.
const ()

// PairingCeremony actions.
const (
	PairingCeremonyActionDecodeToken   ActionID = "decode_token"
	PairingCeremonyActionDeriveCode    ActionID = "derive_code"
	PairingCeremonyActionDialRelay     ActionID = "dial_relay"
	PairingCeremonyActionEmitToken     ActionID = "emit_token"
	PairingCeremonyActionGenEphemeral  ActionID = "gen_ephemeral"
	PairingCeremonyActionRegisterRelay ActionID = "register_relay"
	PairingCeremonyActionStoreRecord   ActionID = "store_record"
)

// PairingCeremony events.
const (
	PairingCeremonyEventCodeReady              EventID = "code_ready"
	PairingCeremonyEventEphemeralReady         EventID = "ephemeral_ready"
	PairingCeremonyEventPairBegin              EventID = "pair_begin"
	PairingCeremonyEventRecvConfirmToAcceptor  EventID = "recv_confirm_to_acceptor"
	PairingCeremonyEventRecvConfirmToInitiator EventID = "recv_confirm_to_initiator"
	PairingCeremonyEventRecvHello              EventID = "recv_hello"
	PairingCeremonyEventRecvWelcome            EventID = "recv_welcome"
	PairingCeremonyEventRelayConnected         EventID = "relay_connected"
	PairingCeremonyEventRelayRegistered        EventID = "relay_registered"
	PairingCeremonyEventTokenDecoded           EventID = "token_decoded"
	PairingCeremonyEventTokenReceived          EventID = "token_received"
	PairingCeremonyEventUserCancel             EventID = "user_cancel"
	PairingCeremonyEventUserConfirm            EventID = "user_confirm"
)

func PairingCeremony() *Protocol {
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
			{Name: "ValidPairingRecord", Params: "r", Expr: "r.peer_instance_id /= \"none\" /\\ r.peer_eph_pub /= \"none\"", Desc: "Interface contract with SessionMachine: a PairingRecord is well-formed iff it carries a non-default peer instance ID and ephemeral pubkey. Pairing proves it as a postcondition (PairingProducesValidRecord); SessionMachine assumes it over its initial state."},
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
			{Name: "PairingProducesValidRecord", Kind: Invariant, Expr: "(acceptor_state = acceptor_Paired => ValidPairingRecord([peer_instance_id |-> acceptor_received_instance, peer_eph_pub |-> acceptor_received_eph_pub])) /\\ (initiator_state = initiator_Paired => ValidPairingRecord([peer_instance_id |-> initiator_received_instance, peer_eph_pub |-> initiator_received_eph_pub]))", Desc: "Postcondition: whenever either side reaches Paired, the implicit PairingRecord built from received-peer fields satisfies the ValidPairingRecord interface contract that SessionMachine assumes."},
		},
		ChannelBound: 3,
		OneShot:      true,
	}
}

// PairingCeremonyAcceptorMachine is the generated state machine for the acceptor actor.
type PairingCeremonyAcceptorMachine struct {
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

func NewPairingCeremonyAcceptorMachine() *PairingCeremonyAcceptorMachine {
	return &PairingCeremonyAcceptorMachine{
		State:                    PairingCeremonyAcceptorIdle,
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

func (m *PairingCeremonyAcceptorMachine) HandleMessage(msg MsgType) (bool, error) {
	switch {
	case m.State == PairingCeremonyAcceptorWaitingForHello && msg == PairingCeremonyMsgHello:
		if fn := m.Actions[PairingCeremonyActionDeriveCode]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		// acceptor_received_eph_pub: recv_msg.eph_pub (set by action)
		// acceptor_received_identity: recv_msg.identity_pub (set by action)
		// acceptor_received_instance: recv_msg.instance_id (set by action)
		// acceptor_code: DeriveCode(acceptor_eph_pub, recv_msg.eph_pub) (set by action)
		m.State = PairingCeremonyAcceptorDerivingCode
		return true, nil
	case m.State == PairingCeremonyAcceptorAwaitingPeerConfirm && msg == PairingCeremonyMsgConfirmToAcceptor:
		if fn := m.Actions[PairingCeremonyActionStoreRecord]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		m.AcceptorReceivedConfirm = "true"
		if m.OnChange != nil {
			m.OnChange("acceptor_received_confirm")
		}
		m.State = PairingCeremonyAcceptorPaired
		return true, nil
	}
	return false, nil
}

func (m *PairingCeremonyAcceptorMachine) Step(event EventID) (bool, error) {
	switch {
	case m.State == PairingCeremonyAcceptorIdle && event == PairingCeremonyEventPairBegin:
		if fn := m.Actions[PairingCeremonyActionGenEphemeral]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		m.AcceptorEphPub = "acceptor_eph"
		if m.OnChange != nil {
			m.OnChange("acceptor_eph_pub")
		}
		m.State = PairingCeremonyAcceptorGeneratingEphemeral
		return true, nil
	case m.State == PairingCeremonyAcceptorGeneratingEphemeral && event == PairingCeremonyEventEphemeralReady:
		if fn := m.Actions[PairingCeremonyActionRegisterRelay]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		m.State = PairingCeremonyAcceptorRegisteringRelay
		return true, nil
	case m.State == PairingCeremonyAcceptorRegisteringRelay && event == PairingCeremonyEventRelayRegistered:
		if fn := m.Actions[PairingCeremonyActionEmitToken]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		m.State = PairingCeremonyAcceptorWaitingForHello
		return true, nil
	case m.State == PairingCeremonyAcceptorDerivingCode && event == PairingCeremonyEventCodeReady:
		m.State = PairingCeremonyAcceptorAwaitingUserConfirm
		return true, nil
	case m.State == PairingCeremonyAcceptorAwaitingUserConfirm && event == PairingCeremonyEventUserConfirm:
		m.AcceptorUserConfirmed = "true"
		if m.OnChange != nil {
			m.OnChange("acceptor_user_confirmed")
		}
		m.State = PairingCeremonyAcceptorAwaitingPeerConfirm
		return true, nil
	case m.State == PairingCeremonyAcceptorAwaitingUserConfirm && event == PairingCeremonyEventUserCancel:
		m.State = PairingCeremonyAcceptorAborted
		return true, nil
	case m.State == PairingCeremonyAcceptorAwaitingPeerConfirm && event == PairingCeremonyEventUserCancel:
		m.State = PairingCeremonyAcceptorAborted
		return true, nil
	}
	return false, nil
}

func (m *PairingCeremonyAcceptorMachine) HandleEvent(ev EventID) ([]CmdID, error) {
	switch {
	case m.State == PairingCeremonyAcceptorIdle && ev == PairingCeremonyEventPairBegin:
		if fn := m.Actions[PairingCeremonyActionGenEphemeral]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		m.AcceptorEphPub = "acceptor_eph"
		if m.OnChange != nil {
			m.OnChange("acceptor_eph_pub")
		}
		m.State = PairingCeremonyAcceptorGeneratingEphemeral
		return nil, nil
	case m.State == PairingCeremonyAcceptorGeneratingEphemeral && ev == PairingCeremonyEventEphemeralReady:
		if fn := m.Actions[PairingCeremonyActionRegisterRelay]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		m.State = PairingCeremonyAcceptorRegisteringRelay
		return nil, nil
	case m.State == PairingCeremonyAcceptorRegisteringRelay && ev == PairingCeremonyEventRelayRegistered:
		if fn := m.Actions[PairingCeremonyActionEmitToken]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		m.State = PairingCeremonyAcceptorWaitingForHello
		return nil, nil
	case m.State == PairingCeremonyAcceptorWaitingForHello && ev == PairingCeremonyEventRecvHello:
		if fn := m.Actions[PairingCeremonyActionDeriveCode]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		// acceptor_received_eph_pub: recv_msg.eph_pub (set by action)
		// acceptor_received_identity: recv_msg.identity_pub (set by action)
		// acceptor_received_instance: recv_msg.instance_id (set by action)
		// acceptor_code: DeriveCode(acceptor_eph_pub, recv_msg.eph_pub) (set by action)
		m.State = PairingCeremonyAcceptorDerivingCode
		return nil, nil
	case m.State == PairingCeremonyAcceptorDerivingCode && ev == PairingCeremonyEventCodeReady:
		m.State = PairingCeremonyAcceptorAwaitingUserConfirm
		return nil, nil
	case m.State == PairingCeremonyAcceptorAwaitingUserConfirm && ev == PairingCeremonyEventUserConfirm:
		m.AcceptorUserConfirmed = "true"
		if m.OnChange != nil {
			m.OnChange("acceptor_user_confirmed")
		}
		m.State = PairingCeremonyAcceptorAwaitingPeerConfirm
		return nil, nil
	case m.State == PairingCeremonyAcceptorAwaitingPeerConfirm && ev == PairingCeremonyEventRecvConfirmToAcceptor:
		if fn := m.Actions[PairingCeremonyActionStoreRecord]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		m.AcceptorReceivedConfirm = "true"
		if m.OnChange != nil {
			m.OnChange("acceptor_received_confirm")
		}
		m.State = PairingCeremonyAcceptorPaired
		return nil, nil
	case m.State == PairingCeremonyAcceptorAwaitingUserConfirm && ev == PairingCeremonyEventUserCancel:
		m.State = PairingCeremonyAcceptorAborted
		return nil, nil
	case m.State == PairingCeremonyAcceptorAwaitingPeerConfirm && ev == PairingCeremonyEventUserCancel:
		m.State = PairingCeremonyAcceptorAborted
		return nil, nil
	}
	return nil, nil
}

// PairingCeremonyInitiatorMachine is the generated state machine for the initiator actor.
type PairingCeremonyInitiatorMachine struct {
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

func NewPairingCeremonyInitiatorMachine() *PairingCeremonyInitiatorMachine {
	return &PairingCeremonyInitiatorMachine{
		State:                     PairingCeremonyInitiatorIdle,
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

func (m *PairingCeremonyInitiatorMachine) HandleMessage(msg MsgType) (bool, error) {
	switch {
	case m.State == PairingCeremonyInitiatorAwaitingWelcome && msg == PairingCeremonyMsgWelcome:
		if fn := m.Actions[PairingCeremonyActionDeriveCode]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		// initiator_received_eph_pub: recv_msg.eph_pub (set by action)
		// initiator_received_identity: recv_msg.identity_pub (set by action)
		// initiator_received_instance: recv_msg.instance_id (set by action)
		// initiator_code: DeriveCode(initiator_eph_pub, recv_msg.eph_pub) (set by action)
		m.State = PairingCeremonyInitiatorDerivingCode
		return true, nil
	case m.State == PairingCeremonyInitiatorAwaitingPeerConfirm && msg == PairingCeremonyMsgConfirmToInitiator:
		if fn := m.Actions[PairingCeremonyActionStoreRecord]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		m.InitiatorReceivedConfirm = "true"
		if m.OnChange != nil {
			m.OnChange("initiator_received_confirm")
		}
		m.State = PairingCeremonyInitiatorPaired
		return true, nil
	}
	return false, nil
}

func (m *PairingCeremonyInitiatorMachine) Step(event EventID) (bool, error) {
	switch {
	case m.State == PairingCeremonyInitiatorIdle && event == PairingCeremonyEventTokenReceived:
		if fn := m.Actions[PairingCeremonyActionDecodeToken]; fn != nil {
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
		m.State = PairingCeremonyInitiatorDecodingToken
		return true, nil
	case m.State == PairingCeremonyInitiatorDecodingToken && event == PairingCeremonyEventTokenDecoded:
		if fn := m.Actions[PairingCeremonyActionGenEphemeral]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		m.InitiatorEphPub = "initiator_eph"
		if m.OnChange != nil {
			m.OnChange("initiator_eph_pub")
		}
		m.State = PairingCeremonyInitiatorGeneratingEphemeral
		return true, nil
	case m.State == PairingCeremonyInitiatorGeneratingEphemeral && event == PairingCeremonyEventEphemeralReady:
		if fn := m.Actions[PairingCeremonyActionDialRelay]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		m.State = PairingCeremonyInitiatorConnectingRelay
		return true, nil
	case m.State == PairingCeremonyInitiatorConnectingRelay && event == PairingCeremonyEventRelayConnected:
		m.State = PairingCeremonyInitiatorAwaitingWelcome
		return true, nil
	case m.State == PairingCeremonyInitiatorDerivingCode && event == PairingCeremonyEventCodeReady:
		m.State = PairingCeremonyInitiatorAwaitingUserConfirm
		return true, nil
	case m.State == PairingCeremonyInitiatorAwaitingUserConfirm && event == PairingCeremonyEventUserConfirm:
		m.InitiatorUserConfirmed = "true"
		if m.OnChange != nil {
			m.OnChange("initiator_user_confirmed")
		}
		m.State = PairingCeremonyInitiatorAwaitingPeerConfirm
		return true, nil
	case m.State == PairingCeremonyInitiatorAwaitingUserConfirm && event == PairingCeremonyEventUserCancel:
		m.State = PairingCeremonyInitiatorAborted
		return true, nil
	case m.State == PairingCeremonyInitiatorAwaitingPeerConfirm && event == PairingCeremonyEventUserCancel:
		m.State = PairingCeremonyInitiatorAborted
		return true, nil
	}
	return false, nil
}

func (m *PairingCeremonyInitiatorMachine) HandleEvent(ev EventID) ([]CmdID, error) {
	switch {
	case m.State == PairingCeremonyInitiatorIdle && ev == PairingCeremonyEventTokenReceived:
		if fn := m.Actions[PairingCeremonyActionDecodeToken]; fn != nil {
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
		m.State = PairingCeremonyInitiatorDecodingToken
		return nil, nil
	case m.State == PairingCeremonyInitiatorDecodingToken && ev == PairingCeremonyEventTokenDecoded:
		if fn := m.Actions[PairingCeremonyActionGenEphemeral]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		m.InitiatorEphPub = "initiator_eph"
		if m.OnChange != nil {
			m.OnChange("initiator_eph_pub")
		}
		m.State = PairingCeremonyInitiatorGeneratingEphemeral
		return nil, nil
	case m.State == PairingCeremonyInitiatorGeneratingEphemeral && ev == PairingCeremonyEventEphemeralReady:
		if fn := m.Actions[PairingCeremonyActionDialRelay]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		m.State = PairingCeremonyInitiatorConnectingRelay
		return nil, nil
	case m.State == PairingCeremonyInitiatorConnectingRelay && ev == PairingCeremonyEventRelayConnected:
		m.State = PairingCeremonyInitiatorAwaitingWelcome
		return nil, nil
	case m.State == PairingCeremonyInitiatorAwaitingWelcome && ev == PairingCeremonyEventRecvWelcome:
		if fn := m.Actions[PairingCeremonyActionDeriveCode]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		// initiator_received_eph_pub: recv_msg.eph_pub (set by action)
		// initiator_received_identity: recv_msg.identity_pub (set by action)
		// initiator_received_instance: recv_msg.instance_id (set by action)
		// initiator_code: DeriveCode(initiator_eph_pub, recv_msg.eph_pub) (set by action)
		m.State = PairingCeremonyInitiatorDerivingCode
		return nil, nil
	case m.State == PairingCeremonyInitiatorDerivingCode && ev == PairingCeremonyEventCodeReady:
		m.State = PairingCeremonyInitiatorAwaitingUserConfirm
		return nil, nil
	case m.State == PairingCeremonyInitiatorAwaitingUserConfirm && ev == PairingCeremonyEventUserConfirm:
		m.InitiatorUserConfirmed = "true"
		if m.OnChange != nil {
			m.OnChange("initiator_user_confirmed")
		}
		m.State = PairingCeremonyInitiatorAwaitingPeerConfirm
		return nil, nil
	case m.State == PairingCeremonyInitiatorAwaitingPeerConfirm && ev == PairingCeremonyEventRecvConfirmToInitiator:
		if fn := m.Actions[PairingCeremonyActionStoreRecord]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		m.InitiatorReceivedConfirm = "true"
		if m.OnChange != nil {
			m.OnChange("initiator_received_confirm")
		}
		m.State = PairingCeremonyInitiatorPaired
		return nil, nil
	case m.State == PairingCeremonyInitiatorAwaitingUserConfirm && ev == PairingCeremonyEventUserCancel:
		m.State = PairingCeremonyInitiatorAborted
		return nil, nil
	case m.State == PairingCeremonyInitiatorAwaitingPeerConfirm && ev == PairingCeremonyEventUserCancel:
		m.State = PairingCeremonyInitiatorAborted
		return nil, nil
	}
	return nil, nil
}
