// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Code generated from protocol/*.yaml. DO NOT EDIT.

package pigeon

import (
	"github.com/arr-ai/frozen"
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

var _ frozen.Set[string] // suppress unused import

// SessionProtocol backend states.
const (
	SessionProtocolBackendIdle                 State = "Idle"
	SessionProtocolBackendGenerateToken        State = "GenerateToken"
	SessionProtocolBackendRegisterRelay        State = "RegisterRelay"
	SessionProtocolBackendWaitingForClient     State = "WaitingForClient"
	SessionProtocolBackendDeriveSecret         State = "DeriveSecret"
	SessionProtocolBackendSendAck              State = "SendAck"
	SessionProtocolBackendWaitingForCode       State = "WaitingForCode"
	SessionProtocolBackendValidateCode         State = "ValidateCode"
	SessionProtocolBackendStorePaired          State = "StorePaired"
	SessionProtocolBackendPaired               State = "Paired"
	SessionProtocolBackendAuthCheck            State = "AuthCheck"
	SessionProtocolBackendSessionActive        State = "SessionActive"
	SessionProtocolBackendRelayConnected       State = "RelayConnected"
	SessionProtocolBackendCandidatesAdvertised State = "CandidatesAdvertised"
	SessionProtocolBackendAltActive            State = "AltActive"
	SessionProtocolBackendRelayBackoff         State = "RelayBackoff"
	SessionProtocolBackendAltDegraded          State = "AltDegraded"
)

// SessionProtocol client states.
const (
	SessionProtocolClientIdle                    State = "Idle"
	SessionProtocolClientObtainBackchannelSecret State = "ObtainBackchannelSecret"
	SessionProtocolClientConnectRelay            State = "ConnectRelay"
	SessionProtocolClientGenKeyPair              State = "GenKeyPair"
	SessionProtocolClientWaitAck                 State = "WaitAck"
	SessionProtocolClientE2EReady                State = "E2EReady"
	SessionProtocolClientShowCode                State = "ShowCode"
	SessionProtocolClientWaitPairComplete        State = "WaitPairComplete"
	SessionProtocolClientPaired                  State = "Paired"
	SessionProtocolClientReconnect               State = "Reconnect"
	SessionProtocolClientSendAuth                State = "SendAuth"
	SessionProtocolClientSessionActive           State = "SessionActive"
	SessionProtocolClientRelayConnected          State = "RelayConnected"
	SessionProtocolClientPairDialing             State = "PairDialing"
	SessionProtocolClientPairChecking            State = "PairChecking"
	SessionProtocolClientAltActive               State = "AltActive"
	SessionProtocolClientRelayFallback           State = "RelayFallback"
)

// SessionProtocol relay states.
const (
	SessionProtocolRelayIdle              State = "Idle"
	SessionProtocolRelayBackendRegistered State = "BackendRegistered"
	SessionProtocolRelayBridged           State = "Bridged"
)

// SessionProtocol message types.
const (
	SessionProtocolMsgPairHello    MsgType = "pair_hello"
	SessionProtocolMsgPairHelloAck MsgType = "pair_hello_ack"
	SessionProtocolMsgPairConfirm  MsgType = "pair_confirm"
	SessionProtocolMsgPairComplete MsgType = "pair_complete"
	SessionProtocolMsgAuthRequest  MsgType = "auth_request"
	SessionProtocolMsgAuthOk       MsgType = "auth_ok"
	SessionProtocolMsgCandidates   MsgType = "candidates"
	SessionProtocolMsgPairCheck    MsgType = "pair_check"
	SessionProtocolMsgPairCheckAck MsgType = "pair_check_ack"
	SessionProtocolMsgPathPing     MsgType = "path_ping"
	SessionProtocolMsgPathPong     MsgType = "path_pong"
)

// SessionProtocol guards.
const (
	SessionProtocolGuardTokenValid               GuardID = "token_valid"
	SessionProtocolGuardTokenInvalid             GuardID = "token_invalid"
	SessionProtocolGuardCodeCorrect              GuardID = "code_correct"
	SessionProtocolGuardCodeWrong                GuardID = "code_wrong"
	SessionProtocolGuardDeviceKnown              GuardID = "device_known"
	SessionProtocolGuardDeviceUnknown            GuardID = "device_unknown"
	SessionProtocolGuardNonceFresh               GuardID = "nonce_fresh"
	SessionProtocolGuardChallengeValid           GuardID = "challenge_valid"
	SessionProtocolGuardChallengeInvalid         GuardID = "challenge_invalid"
	SessionProtocolGuardAltEnabled               GuardID = "alt_enabled"
	SessionProtocolGuardAltDisabled              GuardID = "alt_disabled"
	SessionProtocolGuardLocalCandidatesAvailable GuardID = "local_candidates_available"
	SessionProtocolGuardUnderMaxFailures         GuardID = "under_max_failures"
	SessionProtocolGuardAtMaxFailures            GuardID = "at_max_failures"
)

// SessionProtocol actions.
const (
	SessionProtocolActionActivateLan     ActionID = "activate_lan"
	SessionProtocolActionBridgeStreams   ActionID = "bridge_streams"
	SessionProtocolActionDeriveSecret    ActionID = "derive_secret"
	SessionProtocolActionDialCandidate   ActionID = "dial_candidate"
	SessionProtocolActionFallbackToRelay ActionID = "fallback_to_relay"
	SessionProtocolActionGenerateToken   ActionID = "generate_token"
	SessionProtocolActionRegisterRelay   ActionID = "register_relay"
	SessionProtocolActionResetFailures   ActionID = "reset_failures"
	SessionProtocolActionSendPairHello   ActionID = "send_pair_hello"
	SessionProtocolActionStoreDevice     ActionID = "store_device"
	SessionProtocolActionStoreSecret     ActionID = "store_secret"
	SessionProtocolActionUnbridge        ActionID = "unbridge"
	SessionProtocolActionVerifyDevice    ActionID = "verify_device"
)

// SessionProtocol events.
const (
	SessionProtocolEventAltDatagram           EventID = "alt_datagram"
	SessionProtocolEventAltError              EventID = "alt_error"
	SessionProtocolEventAltStreamData         EventID = "alt_stream_data"
	SessionProtocolEventAltStreamError        EventID = "alt_stream_error"
	SessionProtocolEventAppClose              EventID = "app_close"
	SessionProtocolEventAppForceFallback      EventID = "app_force_fallback"
	SessionProtocolEventAppLaunch             EventID = "app_launch"
	SessionProtocolEventAppRecv               EventID = "app_recv"
	SessionProtocolEventAppRecvDatagram       EventID = "app_recv_datagram"
	SessionProtocolEventAppSend               EventID = "app_send"
	SessionProtocolEventAppSendDatagram       EventID = "app_send_datagram"
	SessionProtocolEventBackchannelReceived   EventID = "backchannel_received"
	SessionProtocolEventBackendDisconnect     EventID = "backend_disconnect"
	SessionProtocolEventBackendRegister       EventID = "backend_register"
	SessionProtocolEventBackoffExpired        EventID = "backoff_expired"
	SessionProtocolEventCandidatesChanged     EventID = "candidates_changed"
	SessionProtocolEventCandidatesGathered    EventID = "candidates_gathered"
	SessionProtocolEventCandidatesRefreshTick EventID = "candidates_refresh_tick"
	SessionProtocolEventCandidatesTimeout     EventID = "candidates_timeout"
	SessionProtocolEventCheckCode             EventID = "check_code"
	SessionProtocolEventCliCodeEntered        EventID = "cli_code_entered"
	SessionProtocolEventCliInitPair           EventID = "cli_init_pair"
	SessionProtocolEventClientConnect         EventID = "client_connect"
	SessionProtocolEventClientDisconnect      EventID = "client_disconnect"
	SessionProtocolEventCodeDisplayed         EventID = "code_displayed"
	SessionProtocolEventDialFailed            EventID = "dial_failed"
	SessionProtocolEventDialOk                EventID = "dial_ok"
	SessionProtocolEventDisconnect            EventID = "disconnect"
	SessionProtocolEventEcdhComplete          EventID = "ecdh_complete"
	SessionProtocolEventFinalise              EventID = "finalise"
	SessionProtocolEventKeyPairGenerated      EventID = "key_pair_generated"
	SessionProtocolEventPairCheckOk           EventID = "pair_check_ok"
	SessionProtocolEventPingTick              EventID = "ping_tick"
	SessionProtocolEventPingTimeout           EventID = "ping_timeout"
	SessionProtocolEventRecvAuthOk            EventID = "recv_auth_ok"
	SessionProtocolEventRecvAuthRequest       EventID = "recv_auth_request"
	SessionProtocolEventRecvCandidates        EventID = "recv_candidates"
	SessionProtocolEventRecvPairCheck         EventID = "recv_pair_check"
	SessionProtocolEventRecvPairCheckAck      EventID = "recv_pair_check_ack"
	SessionProtocolEventRecvPairComplete      EventID = "recv_pair_complete"
	SessionProtocolEventRecvPairConfirm       EventID = "recv_pair_confirm"
	SessionProtocolEventRecvPairHello         EventID = "recv_pair_hello"
	SessionProtocolEventRecvPairHelloAck      EventID = "recv_pair_hello_ack"
	SessionProtocolEventRecvPathPing          EventID = "recv_path_ping"
	SessionProtocolEventRecvPathPong          EventID = "recv_path_pong"
	SessionProtocolEventRelayConnected        EventID = "relay_connected"
	SessionProtocolEventRelayDatagram         EventID = "relay_datagram"
	SessionProtocolEventRelayOk               EventID = "relay_ok"
	SessionProtocolEventRelayRegistered       EventID = "relay_registered"
	SessionProtocolEventRelayStreamData       EventID = "relay_stream_data"
	SessionProtocolEventRelayStreamError      EventID = "relay_stream_error"
	SessionProtocolEventSecretParsed          EventID = "secret_parsed"
	SessionProtocolEventSessionEstablished    EventID = "session_established"
	SessionProtocolEventSignalCodeDisplay     EventID = "signal_code_display"
	SessionProtocolEventTokenCreated          EventID = "token_created"
	SessionProtocolEventVerify                EventID = "verify"
	SessionProtocolEventVerifyTimeout         EventID = "verify_timeout"
)

// SessionProtocol commands.
const (
	SessionProtocolCmdWriteActiveStream    CmdID = "write_active_stream"
	SessionProtocolCmdSendActiveDatagram   CmdID = "send_active_datagram"
	SessionProtocolCmdSendPathPing         CmdID = "send_path_ping"
	SessionProtocolCmdSendPathPong         CmdID = "send_path_pong"
	SessionProtocolCmdSendCandidates       CmdID = "send_candidates"
	SessionProtocolCmdSendPairCheck        CmdID = "send_pair_check"
	SessionProtocolCmdSendPairCheckAck     CmdID = "send_pair_check_ack"
	SessionProtocolCmdDialCandidate        CmdID = "dial_candidate"
	SessionProtocolCmdDeliverRecv          CmdID = "deliver_recv"
	SessionProtocolCmdDeliverRecvError     CmdID = "deliver_recv_error"
	SessionProtocolCmdDeliverRecvDatagram  CmdID = "deliver_recv_datagram"
	SessionProtocolCmdStartAltStreamReader CmdID = "start_alt_stream_reader"
	SessionProtocolCmdStopAltStreamReader  CmdID = "stop_alt_stream_reader"
	SessionProtocolCmdStartAltDgReader     CmdID = "start_alt_dg_reader"
	SessionProtocolCmdStopAltDgReader      CmdID = "stop_alt_dg_reader"
	SessionProtocolCmdStartMonitor         CmdID = "start_monitor"
	SessionProtocolCmdStopMonitor          CmdID = "stop_monitor"
	SessionProtocolCmdStartPongTimeout     CmdID = "start_pong_timeout"
	SessionProtocolCmdCancelPongTimeout    CmdID = "cancel_pong_timeout"
	SessionProtocolCmdStartBackoffTimer    CmdID = "start_backoff_timer"
	SessionProtocolCmdCloseAltPath         CmdID = "close_alt_path"
	SessionProtocolCmdSignalAltReady       CmdID = "signal_alt_ready"
	SessionProtocolCmdResetAltReady        CmdID = "reset_alt_ready"
	SessionProtocolCmdSetCryptoDatagram    CmdID = "set_crypto_datagram"
)

// Wire constants — protocol-level values shared across all platforms.
// DatagramFraming
const (
	DgConnWhole    byte = 0x00 // conn-level single-frame datagram
	DgPing         byte = 0x10 // health ping on direct path
	DgPong         byte = 0x11 // health pong on direct path
	DgConnFragment byte = 0x40 // conn-level multi-frame datagram
	DgChanWhole    byte = 0x80 // channel single-frame datagram
	DgChanFragment byte = 0xC0 // channel multi-frame datagram
	FragHeaderSize      = 8    // fragment header: msgID(4) + fragIdx(2) + totalFrags(2)
	ChanIdSize          = 2    // channel ID prefix size in bytes
)

// DatagramLimits
const (
	MaxDatagramPayload = 1200 // max payload per QUIC datagram (bytes)
	FragmentTimeoutMs  = 5000 // ms // fragment reassembly timeout
)

// MessageFraming
const (
	FrameApp          byte = 0x00 // application data
	FrameCandidates   byte = 0x01 // candidate set advertisement (relay control channel)
	FrameCutover      byte = 0x02 // transport cutover marker
	FramePairCheck    byte = 0x03 // candidate-pair connectivity-check challenge (on candidate pipe)
	FramePairCheckAck byte = 0x04 // candidate-pair connectivity-check ack / implicit nomination (on candidate pipe)
)

// Candidates
const (
	CandHost  = "host"  // host candidate (local interface address; what LAN-direct uses)
	CandSrflx = "srflx" // server-reflexive candidate (STUN-derived public address; Phase 1 STUN target)
)

// MessageFraming
const (
	MaxMessageSize   = 1048576 // max stream message size (1 MiB)
	LengthPrefixSize = 4       // big-endian length prefix size
)

// Health
const (
	PingIntervalMs  = 5000 // ms // health ping interval
	PongTimeoutMs   = 4000 // ms // pong reply timeout
	MaxPingFailures = 3    // consecutive failures before fallback
	MaxBackoffLevel = 5    // exponential backoff cap
)

// ChannelKeys
const (
	StreamChannelOpenerSuffix = ":o2a"     // HKDF info suffix for opener→acceptor stream key
	StreamChannelAcceptSuffix = ":a2o"     // HKDF info suffix for acceptor→opener stream key
	DgChannelSendSuffix       = ":dg:send" // HKDF info suffix for datagram send key
	DgChannelRecvSuffix       = ":dg:recv" // HKDF info suffix for datagram recv key
	ChannelIdHashMultiplier   = 31         // hash multiplier for channel name → uint16 ID
)

func SessionProtocol() *Protocol {
	return &Protocol{
		Name: "Session",
		Actors: []Actor{
			{Name: "backend", Initial: "Idle", Transitions: []Transition{
				{From: "Idle", To: "GenerateToken", On: Internal("cli_init_pair"), Do: "generate_token", Updates: []VarUpdate{{Var: "current_token", Expr: "\"tok_1\""}, {Var: "active_tokens", Expr: "active_tokens \\union {\"tok_1\"}"}}},
				{From: "GenerateToken", To: "RegisterRelay", On: Internal("token_created"), Do: "register_relay"},
				{From: "RegisterRelay", To: "WaitingForClient", On: Internal("relay_registered"), Updates: []VarUpdate{{Var: "secret_published", Expr: "TRUE"}}},
				{From: "WaitingForClient", To: "DeriveSecret", On: Recv("pair_hello"), Guard: "token_valid", Do: "derive_secret", Updates: []VarUpdate{{Var: "received_client_pub", Expr: "recv_msg.pubkey"}, {Var: "backend_ecdh_pub", Expr: "\"backend_pub\""}, {Var: "backend_shared_key", Expr: "DeriveKey(\"backend_pub\", recv_msg.pubkey)"}, {Var: "backend_code", Expr: "DeriveCode(\"backend_pub\", recv_msg.pubkey)"}}},
				{From: "WaitingForClient", To: "Idle", On: Recv("pair_hello"), Guard: "token_invalid"},
				{From: "DeriveSecret", To: "SendAck", On: Internal("ecdh_complete"), Sends: []Send{{To: "client", Msg: "pair_hello_ack", Fields: map[string]string{"pubkey": "backend_ecdh_pub"}}}},
				{From: "SendAck", To: "WaitingForCode", On: Internal("signal_code_display"), Sends: []Send{{To: "client", Msg: "pair_confirm"}}},
				{From: "WaitingForCode", To: "ValidateCode", On: Internal("cli_code_entered"), Updates: []VarUpdate{{Var: "received_code", Expr: "cli_entered_code"}}},
				{From: "ValidateCode", To: "StorePaired", On: Internal("check_code"), Guard: "code_correct"},
				{From: "ValidateCode", To: "Idle", On: Internal("check_code"), Guard: "code_wrong", Updates: []VarUpdate{{Var: "code_attempts", Expr: "code_attempts + 1"}}},
				{From: "StorePaired", To: "Paired", On: Internal("finalise"), Do: "store_device", Sends: []Send{{To: "client", Msg: "pair_complete", Fields: map[string]string{"key": "backend_shared_key", "secret": "\"dev_secret_1\""}}}, Updates: []VarUpdate{{Var: "device_secret", Expr: "\"dev_secret_1\""}, {Var: "paired_devices", Expr: "paired_devices \\union {\"device_1\"}"}, {Var: "active_tokens", Expr: "active_tokens \\ {current_token}"}, {Var: "used_tokens", Expr: "used_tokens \\union {current_token}"}}},
				{From: "Paired", To: "AuthCheck", On: Recv("auth_request"), Updates: []VarUpdate{{Var: "received_device_id", Expr: "recv_msg.device_id"}, {Var: "received_auth_nonce", Expr: "recv_msg.nonce"}}},
				{From: "AuthCheck", To: "SessionActive", On: Internal("verify"), Guard: "device_known", Do: "verify_device", Sends: []Send{{To: "client", Msg: "auth_ok"}}, Updates: []VarUpdate{{Var: "auth_nonces_used", Expr: "auth_nonces_used \\union {received_auth_nonce}"}}},
				{From: "AuthCheck", To: "Idle", On: Internal("verify"), Guard: "device_unknown"},
				{From: "SessionActive", To: "RelayConnected", On: Internal("session_established")},
				{From: "RelayConnected", To: "CandidatesAdvertised", On: Internal("candidates_gathered"), Sends: []Send{{To: "client", Msg: "candidates", Fields: map[string]string{"challenge": "local_pair_check_challenge", "set": "local_candidates"}}}},
				{From: "CandidatesAdvertised", To: "AltActive", On: Recv("pair_check"), Guard: "challenge_valid", Do: "activate_lan", Sends: []Send{{To: "client", Msg: "pair_check_ack"}}, Updates: []VarUpdate{{Var: "ping_failures", Expr: "0"}, {Var: "backoff_level", Expr: "0"}, {Var: "b_active_path", Expr: "\"alt\""}, {Var: "b_dispatcher_path", Expr: "\"alt\""}, {Var: "monitor_target", Expr: "\"alt\""}, {Var: "alt_signal", Expr: "\"ready\""}}},
				{From: "CandidatesAdvertised", To: "RelayConnected", On: Recv("pair_check"), Guard: "challenge_invalid"},
				{From: "CandidatesAdvertised", To: "RelayBackoff", On: Internal("candidates_timeout"), Updates: []VarUpdate{{Var: "backoff_level", Expr: "Min(backoff_level + 1, max_backoff_level)"}, {Var: "alt_signal", Expr: "\"pending\""}}},
				{From: "AltActive", To: "AltActive", On: Internal("ping_tick"), Sends: []Send{{To: "client", Msg: "path_ping"}}},
				{From: "AltActive", To: "AltDegraded", On: Internal("ping_timeout"), Updates: []VarUpdate{{Var: "ping_failures", Expr: "1"}}},
				{From: "AltDegraded", To: "AltDegraded", On: Internal("ping_tick"), Sends: []Send{{To: "client", Msg: "path_ping"}}},
				{From: "AltActive", To: "RelayBackoff", On: Internal("alt_stream_error"), Do: "fallback_to_relay", Updates: []VarUpdate{{Var: "backoff_level", Expr: "Min(backoff_level + 1, max_backoff_level)"}, {Var: "b_active_path", Expr: "\"relay\""}, {Var: "b_dispatcher_path", Expr: "\"relay\""}, {Var: "monitor_target", Expr: "\"none\""}, {Var: "alt_signal", Expr: "\"pending\""}, {Var: "ping_failures", Expr: "0"}}},
				{From: "AltDegraded", To: "RelayBackoff", On: Internal("alt_stream_error"), Do: "fallback_to_relay", Updates: []VarUpdate{{Var: "backoff_level", Expr: "Min(backoff_level + 1, max_backoff_level)"}, {Var: "b_active_path", Expr: "\"relay\""}, {Var: "b_dispatcher_path", Expr: "\"relay\""}, {Var: "monitor_target", Expr: "\"none\""}, {Var: "alt_signal", Expr: "\"pending\""}, {Var: "ping_failures", Expr: "0"}}},
				{From: "AltDegraded", To: "AltActive", On: Recv("path_pong"), Do: "reset_failures", Updates: []VarUpdate{{Var: "ping_failures", Expr: "0"}}},
				{From: "AltDegraded", To: "AltDegraded", On: Internal("ping_timeout"), Guard: "under_max_failures", Updates: []VarUpdate{{Var: "ping_failures", Expr: "ping_failures + 1"}}},
				{From: "AltDegraded", To: "RelayBackoff", On: Internal("ping_timeout"), Guard: "at_max_failures", Do: "fallback_to_relay", Updates: []VarUpdate{{Var: "backoff_level", Expr: "Min(backoff_level + 1, max_backoff_level)"}, {Var: "b_active_path", Expr: "\"relay\""}, {Var: "b_dispatcher_path", Expr: "\"relay\""}, {Var: "monitor_target", Expr: "\"none\""}, {Var: "alt_signal", Expr: "\"pending\""}, {Var: "ping_failures", Expr: "0"}}},
				{From: "RelayBackoff", To: "CandidatesAdvertised", On: Internal("backoff_expired"), Sends: []Send{{To: "client", Msg: "candidates", Fields: map[string]string{"challenge": "local_pair_check_challenge", "set": "local_candidates"}}}},
				{From: "RelayBackoff", To: "CandidatesAdvertised", On: Internal("candidates_changed"), Sends: []Send{{To: "client", Msg: "candidates", Fields: map[string]string{"challenge": "local_pair_check_challenge", "set": "local_candidates"}}}, Updates: []VarUpdate{{Var: "backoff_level", Expr: "0"}}},
				{From: "RelayConnected", To: "CandidatesAdvertised", On: Internal("candidates_refresh_tick"), Guard: "local_candidates_available", Sends: []Send{{To: "client", Msg: "candidates", Fields: map[string]string{"challenge": "local_pair_check_challenge", "set": "local_candidates"}}}},
				{From: "CandidatesAdvertised", To: "RelayConnected", On: Internal("app_force_fallback"), Updates: []VarUpdate{{Var: "alt_signal", Expr: "\"pending\""}}},
				{From: "AltActive", To: "RelayBackoff", On: Internal("app_force_fallback"), Do: "fallback_to_relay", Updates: []VarUpdate{{Var: "backoff_level", Expr: "Min(backoff_level + 1, max_backoff_level)"}, {Var: "b_active_path", Expr: "\"relay\""}, {Var: "b_dispatcher_path", Expr: "\"relay\""}, {Var: "monitor_target", Expr: "\"none\""}, {Var: "alt_signal", Expr: "\"pending\""}, {Var: "ping_failures", Expr: "0"}}},
				{From: "AltDegraded", To: "RelayBackoff", On: Internal("app_force_fallback"), Do: "fallback_to_relay", Updates: []VarUpdate{{Var: "backoff_level", Expr: "Min(backoff_level + 1, max_backoff_level)"}, {Var: "b_active_path", Expr: "\"relay\""}, {Var: "b_dispatcher_path", Expr: "\"relay\""}, {Var: "monitor_target", Expr: "\"none\""}, {Var: "alt_signal", Expr: "\"pending\""}, {Var: "ping_failures", Expr: "0"}}},
				{From: "RelayConnected", To: "Paired", On: Internal("disconnect")},
				{From: "RelayConnected", To: "RelayConnected", On: Internal("app_send")},
				{From: "CandidatesAdvertised", To: "CandidatesAdvertised", On: Internal("app_send")},
				{From: "AltActive", To: "AltActive", On: Internal("app_send")},
				{From: "AltDegraded", To: "AltDegraded", On: Internal("app_send")},
				{From: "RelayBackoff", To: "RelayBackoff", On: Internal("app_send")},
				{From: "RelayConnected", To: "RelayConnected", On: Internal("relay_stream_data")},
				{From: "CandidatesAdvertised", To: "CandidatesAdvertised", On: Internal("relay_stream_data")},
				{From: "AltActive", To: "AltActive", On: Internal("relay_stream_data")},
				{From: "AltDegraded", To: "AltDegraded", On: Internal("relay_stream_data")},
				{From: "RelayBackoff", To: "RelayBackoff", On: Internal("relay_stream_data")},
				{From: "RelayConnected", To: "RelayConnected", On: Internal("relay_stream_error")},
				{From: "CandidatesAdvertised", To: "CandidatesAdvertised", On: Internal("relay_stream_error")},
				{From: "AltActive", To: "AltActive", On: Internal("relay_stream_error")},
				{From: "AltDegraded", To: "AltDegraded", On: Internal("relay_stream_error")},
				{From: "RelayBackoff", To: "RelayBackoff", On: Internal("relay_stream_error")},
				{From: "RelayConnected", To: "RelayConnected", On: Internal("app_send_datagram")},
				{From: "CandidatesAdvertised", To: "CandidatesAdvertised", On: Internal("app_send_datagram")},
				{From: "AltActive", To: "AltActive", On: Internal("app_send_datagram")},
				{From: "AltDegraded", To: "AltDegraded", On: Internal("app_send_datagram")},
				{From: "RelayBackoff", To: "RelayBackoff", On: Internal("app_send_datagram")},
				{From: "RelayConnected", To: "RelayConnected", On: Internal("relay_datagram")},
				{From: "CandidatesAdvertised", To: "CandidatesAdvertised", On: Internal("relay_datagram")},
				{From: "AltActive", To: "AltActive", On: Internal("relay_datagram")},
				{From: "AltDegraded", To: "AltDegraded", On: Internal("relay_datagram")},
				{From: "RelayBackoff", To: "RelayBackoff", On: Internal("relay_datagram")},
				{From: "AltActive", To: "AltActive", On: Internal("alt_stream_data")},
				{From: "AltDegraded", To: "AltDegraded", On: Internal("alt_stream_data")},
				{From: "AltActive", To: "AltActive", On: Internal("alt_datagram")},
				{From: "AltDegraded", To: "AltDegraded", On: Internal("alt_datagram")},
			}},
			{Name: "client", Initial: "Idle", Transitions: []Transition{
				{From: "Idle", To: "ObtainBackchannelSecret", On: Internal("backchannel_received")},
				{From: "ObtainBackchannelSecret", To: "ConnectRelay", On: Internal("secret_parsed")},
				{From: "ConnectRelay", To: "GenKeyPair", On: Internal("relay_connected")},
				{From: "GenKeyPair", To: "WaitAck", On: Internal("key_pair_generated"), Do: "send_pair_hello", Sends: []Send{{To: "backend", Msg: "pair_hello", Fields: map[string]string{"pubkey": "\"client_pub\"", "token": "current_token"}}}},
				{From: "WaitAck", To: "E2EReady", On: Recv("pair_hello_ack"), Do: "derive_secret", Updates: []VarUpdate{{Var: "received_backend_pub", Expr: "recv_msg.pubkey"}, {Var: "client_shared_key", Expr: "DeriveKey(\"client_pub\", recv_msg.pubkey)"}}},
				{From: "E2EReady", To: "ShowCode", On: Recv("pair_confirm"), Updates: []VarUpdate{{Var: "client_code", Expr: "DeriveCode(received_backend_pub, \"client_pub\")"}}},
				{From: "ShowCode", To: "WaitPairComplete", On: Internal("code_displayed")},
				{From: "WaitPairComplete", To: "Paired", On: Recv("pair_complete"), Do: "store_secret"},
				{From: "Paired", To: "Reconnect", On: Internal("app_launch")},
				{From: "Reconnect", To: "SendAuth", On: Internal("relay_connected"), Sends: []Send{{To: "backend", Msg: "auth_request", Fields: map[string]string{"device_id": "\"device_1\"", "key": "client_shared_key", "nonce": "\"nonce_1\"", "secret": "device_secret"}}}},
				{From: "SendAuth", To: "SessionActive", On: Recv("auth_ok")},
				{From: "SessionActive", To: "RelayConnected", On: Internal("session_established")},
				{From: "RelayConnected", To: "PairDialing", On: Recv("candidates"), Guard: "alt_enabled", Do: "dial_candidate"},
				{From: "RelayConnected", To: "RelayConnected", On: Recv("candidates"), Guard: "alt_disabled"},
				{From: "PairDialing", To: "PairChecking", On: Internal("dial_ok"), Sends: []Send{{To: "backend", Msg: "pair_check", Fields: map[string]string{"challenge": "received_pair_check_challenge", "instance_id": "instance_id"}}}},
				{From: "PairDialing", To: "RelayConnected", On: Internal("dial_failed")},
				{From: "PairChecking", To: "AltActive", On: Recv("pair_check_ack"), Do: "activate_lan", Updates: []VarUpdate{{Var: "c_active_path", Expr: "\"alt\""}, {Var: "c_dispatcher_path", Expr: "\"alt\""}, {Var: "alt_signal", Expr: "\"ready\""}}},
				{From: "PairChecking", To: "RelayConnected", On: Internal("verify_timeout"), Updates: []VarUpdate{{Var: "c_dispatcher_path", Expr: "\"relay\""}}},
				{From: "AltActive", To: "AltActive", On: Recv("path_ping"), Sends: []Send{{To: "backend", Msg: "path_pong"}}},
				{From: "AltActive", To: "RelayFallback", On: Internal("alt_error"), Do: "fallback_to_relay", Updates: []VarUpdate{{Var: "c_active_path", Expr: "\"relay\""}, {Var: "c_dispatcher_path", Expr: "\"relay\""}, {Var: "alt_signal", Expr: "\"pending\""}}},
				{From: "AltActive", To: "RelayFallback", On: Internal("alt_stream_error"), Do: "fallback_to_relay", Updates: []VarUpdate{{Var: "c_active_path", Expr: "\"relay\""}, {Var: "c_dispatcher_path", Expr: "\"relay\""}, {Var: "alt_signal", Expr: "\"pending\""}}},
				{From: "RelayFallback", To: "RelayConnected", On: Internal("relay_ok")},
				{From: "AltActive", To: "PairDialing", On: Recv("candidates"), Guard: "alt_enabled", Do: "dial_candidate"},
				{From: "PairDialing", To: "RelayConnected", On: Internal("app_force_fallback")},
				{From: "PairChecking", To: "RelayConnected", On: Internal("app_force_fallback"), Updates: []VarUpdate{{Var: "c_dispatcher_path", Expr: "\"relay\""}}},
				{From: "AltActive", To: "RelayConnected", On: Internal("app_force_fallback"), Do: "fallback_to_relay", Updates: []VarUpdate{{Var: "c_active_path", Expr: "\"relay\""}, {Var: "c_dispatcher_path", Expr: "\"relay\""}, {Var: "alt_signal", Expr: "\"pending\""}}},
				{From: "RelayConnected", To: "Paired", On: Internal("disconnect")},
				{From: "RelayConnected", To: "RelayConnected", On: Internal("app_send")},
				{From: "PairDialing", To: "PairDialing", On: Internal("app_send")},
				{From: "PairChecking", To: "PairChecking", On: Internal("app_send")},
				{From: "AltActive", To: "AltActive", On: Internal("app_send")},
				{From: "RelayFallback", To: "RelayFallback", On: Internal("app_send")},
				{From: "RelayConnected", To: "RelayConnected", On: Internal("relay_stream_data")},
				{From: "PairDialing", To: "PairDialing", On: Internal("relay_stream_data")},
				{From: "PairChecking", To: "PairChecking", On: Internal("relay_stream_data")},
				{From: "AltActive", To: "AltActive", On: Internal("relay_stream_data")},
				{From: "RelayFallback", To: "RelayFallback", On: Internal("relay_stream_data")},
				{From: "RelayConnected", To: "RelayConnected", On: Internal("relay_stream_error")},
				{From: "PairDialing", To: "PairDialing", On: Internal("relay_stream_error")},
				{From: "PairChecking", To: "PairChecking", On: Internal("relay_stream_error")},
				{From: "AltActive", To: "AltActive", On: Internal("relay_stream_error")},
				{From: "RelayFallback", To: "RelayFallback", On: Internal("relay_stream_error")},
				{From: "RelayConnected", To: "RelayConnected", On: Internal("app_send_datagram")},
				{From: "PairDialing", To: "PairDialing", On: Internal("app_send_datagram")},
				{From: "PairChecking", To: "PairChecking", On: Internal("app_send_datagram")},
				{From: "AltActive", To: "AltActive", On: Internal("app_send_datagram")},
				{From: "RelayFallback", To: "RelayFallback", On: Internal("app_send_datagram")},
				{From: "RelayConnected", To: "RelayConnected", On: Internal("relay_datagram")},
				{From: "PairDialing", To: "PairDialing", On: Internal("relay_datagram")},
				{From: "PairChecking", To: "PairChecking", On: Internal("relay_datagram")},
				{From: "AltActive", To: "AltActive", On: Internal("relay_datagram")},
				{From: "RelayFallback", To: "RelayFallback", On: Internal("relay_datagram")},
				{From: "AltActive", To: "AltActive", On: Internal("alt_stream_data")},
				{From: "AltActive", To: "AltActive", On: Internal("alt_datagram")},
			}},
			{Name: "relay", Initial: "Idle", Transitions: []Transition{
				{From: "Idle", To: "BackendRegistered", On: Internal("backend_register")},
				{From: "BackendRegistered", To: "Bridged", On: Internal("client_connect"), Do: "bridge_streams", Updates: []VarUpdate{{Var: "relay_bridge", Expr: "\"active\""}}},
				{From: "Bridged", To: "BackendRegistered", On: Internal("client_disconnect"), Do: "unbridge", Updates: []VarUpdate{{Var: "relay_bridge", Expr: "\"idle\""}}},
				{From: "BackendRegistered", To: "Idle", On: Internal("backend_disconnect")},
			}},
		},
		Messages: []Message{
			{Type: "pair_hello", From: "client", To: "backend", Desc: "ECDH pubkey + pairing token"},
			{Type: "pair_hello_ack", From: "backend", To: "client", Desc: "ECDH pubkey"},
			{Type: "pair_confirm", From: "backend", To: "client", Desc: "signal to compute and display code"},
			{Type: "pair_complete", From: "backend", To: "client", Desc: "encrypted device secret"},
			{Type: "auth_request", From: "client", To: "backend", Desc: "encrypted auth with nonce"},
			{Type: "auth_ok", From: "backend", To: "client", Desc: "session established"},
			{Type: "candidates", From: "backend", To: "client", Desc: "set of local candidates (host, srflx, ...) + a per-candidate pair_check challenge (sent via relay)"},
			{Type: "pair_check", From: "client", To: "backend", Desc: "challenge echo + instance ID + pair_id under test (sent on the candidate-pair pipe)"},
			{Type: "pair_check_ack", From: "backend", To: "client", Desc: "pair_check accepted; this candidate-pair is nominated as the active path (sent on the candidate-pair pipe)"},
			{Type: "path_ping", From: "backend", To: "client", Desc: "health check on active direct path"},
			{Type: "path_pong", From: "client", To: "backend", Desc: "health check response"},
		},
		Vars: []VarDef{
			{Name: "current_token", Initial: "\"none\"", Desc: "pairing token currently in play"},
			{Name: "active_tokens", Initial: "{}", Desc: "set of valid (non-revoked) tokens"},
			{Name: "used_tokens", Initial: "{}", Desc: "set of revoked tokens"},
			{Name: "backend_ecdh_pub", Initial: "\"none\"", Desc: "backend ECDH public key"},
			{Name: "received_client_pub", Initial: "\"none\"", Desc: "pubkey backend received in pair_hello"},
			{Name: "received_backend_pub", Initial: "\"none\"", Desc: "pubkey client received in pair_hello_ack"},
			{Name: "backend_shared_key", Initial: "<<\"none\">>", Desc: "ECDH key derived by backend"},
			{Name: "client_shared_key", Initial: "<<\"none\">>", Desc: "ECDH key derived by client"},
			{Name: "backend_code", Initial: "<<\"none\">>", Desc: "code computed by backend"},
			{Name: "client_code", Initial: "<<\"none\">>", Desc: "code computed by client"},
			{Name: "received_code", Initial: "<<\"none\">>", Desc: "code entered via CLI"},
			{Name: "cli_entered_code", Initial: "<<\"none\">>", Desc: "staging for CLI code input"},
			{Name: "code_attempts", Initial: "0", Desc: "failed code submission attempts"},
			{Name: "device_secret", Initial: "\"none\"", Desc: "persistent device secret"},
			{Name: "paired_devices", Initial: "{}", Desc: "device IDs that completed pairing"},
			{Name: "received_device_id", Initial: "\"none\"", Desc: "device_id from auth_request"},
			{Name: "auth_nonces_used", Initial: "{}", Desc: "set of consumed auth nonces"},
			{Name: "received_auth_nonce", Initial: "\"none\"", Desc: "nonce from auth_request"},
			{Name: "secret_published", Initial: "FALSE", Desc: "whether token has been published via backchannel"},
			{Name: "recv_msg", Initial: "[type |-> \"none\"]", Desc: "last received message (staging)"},
			{Name: "adversary_keys", Initial: "{}", Desc: "encryption keys the adversary knows"},
			{Name: "adv_ecdh_pub", Initial: "\"adv_pub\"", Desc: "adversary's ECDH public key"},
			{Name: "adv_saved_client_pub", Initial: "\"none\"", Desc: "real client pubkey saved during MitM"},
			{Name: "adv_saved_server_pub", Initial: "\"none\"", Desc: "real backend pubkey saved during MitM"},
			{Name: "local_pair_check_challenge", Initial: "\"none\"", Desc: "32-byte random challenge for pair_check exchange"},
			{Name: "received_pair_check_challenge", Initial: "\"none\"", Desc: "challenge most recently received in pair_check"},
			{Name: "instance_id", Initial: "\"none\"", Desc: "relay instance ID"},
			{Name: "ping_failures", Initial: "0", Desc: "consecutive failed pings"},
			{Name: "max_ping_failures", Initial: "3", Desc: "threshold before fallback"},
			{Name: "backoff_level", Initial: "0", Desc: "exponential backoff level"},
			{Name: "max_backoff_level", Initial: "5", Desc: "backoff cap"},
			{Name: "local_listen_addr", Initial: "\"none\"", Desc: "local listening address (host candidate)"},
			{Name: "local_candidates", Initial: "{}", Desc: "candidate IDs this peer gathered locally and advertised via candidates message"},
			{Name: "remote_candidates", Initial: "{}", Desc: "candidate IDs this peer received in the peer's candidates message"},
			{Name: "pair_check_acks", Initial: "{}", Desc: "pair IDs for which a pair_check_ack has been observed (nomination universe)"},
			{Name: "active_pair_id", Initial: "\"none\"", Desc: "pair ID currently nominated as the active alt path; \"none\" when on relay"},
			{Name: "b_active_path", Initial: "\"relay\"", Desc: "backend active path"},
			{Name: "c_active_path", Initial: "\"relay\"", Desc: "client active path"},
			{Name: "b_dispatcher_path", Initial: "\"relay\"", Desc: "backend datagram dispatcher binding"},
			{Name: "c_dispatcher_path", Initial: "\"relay\"", Desc: "client datagram dispatcher binding"},
			{Name: "monitor_target", Initial: "\"none\"", Desc: "health monitor target"},
			{Name: "alt_signal", Initial: "\"pending\"", Desc: "AltReady notification state"},
			{Name: "relay_bridge", Initial: "\"idle\"", Desc: "relay bridge state"},
		},
		Guards: []GuardDef{
			{ID: "token_valid", Expr: "recv_msg.token \\in active_tokens"},
			{ID: "token_invalid", Expr: "recv_msg.token \\notin active_tokens"},
			{ID: "code_correct", Expr: "received_code = backend_code"},
			{ID: "code_wrong", Expr: "received_code /= backend_code"},
			{ID: "device_known", Expr: "received_device_id \\in paired_devices"},
			{ID: "device_unknown", Expr: "received_device_id \\notin paired_devices"},
			{ID: "nonce_fresh", Expr: "received_auth_nonce \\notin auth_nonces_used"},
			{ID: "challenge_valid", Expr: "received_pair_check_challenge = local_pair_check_challenge"},
			{ID: "challenge_invalid", Expr: "received_pair_check_challenge /= local_pair_check_challenge"},
			{ID: "alt_enabled", Expr: "TRUE"},
			{ID: "alt_disabled", Expr: "FALSE"},
			{ID: "local_candidates_available", Expr: "local_listen_addr /= \"none\""},
			{ID: "under_max_failures", Expr: "ping_failures + 1 < max_ping_failures"},
			{ID: "at_max_failures", Expr: "ping_failures + 1 >= max_ping_failures"},
		},
		Operators: []Operator{
			{Name: "KeyRank", Params: "k", Expr: "CASE k = \"adv_pub\" -> 0 [] k = \"client_pub\" -> 1 [] k = \"backend_pub\" -> 2 [] OTHER -> 3", Desc: "deterministic ordering for ECDH"},
			{Name: "DeriveKey", Params: "a, b", Expr: "IF KeyRank(a) <= KeyRank(b) THEN <<\"ecdh\", a, b>> ELSE <<\"ecdh\", b, a>>", Desc: "symbolic ECDH"},
			{Name: "DeriveCode", Params: "a, b", Expr: "IF KeyRank(a) <= KeyRank(b) THEN <<\"code\", a, b>> ELSE <<\"code\", b, a>>", Desc: "confirmation code from pubkeys"},
			{Name: "Min", Params: "a, b", Expr: "IF a < b THEN a ELSE b", Desc: "minimum of two values"},
			{Name: "ValidPairingRecord", Params: "r", Expr: "r.peer_instance_id /= \"none\" /\\ r.peer_eph_pub /= \"none\"", Desc: "Interface contract with PairingCeremony.tla: a PairingRecord is well-formed iff it carries a non-default peer instance ID and ephemeral pubkey. SessionMachine treats this as an axiom over its initial state — the session begins with a record that PairingCeremony.tla's PairingProducesValidRecord postcondition has already established."},
		},
		AdvActions: []AdvAction{
			{Name: "QR_shoulder_surf", Desc: "observe QR code content", Code: "      await current_token /= \"none\";\n      adversary_knowledge := adversary_knowledge \\union {[type |-> \"qr_token\", token |-> current_token]};"},
			{Name: "MitM_pair_hello", Desc: "intercept pair_hello and substitute adversary pubkey", Code: "      await Len(chan_client_backend) > 0 /\\ Head(chan_client_backend).type = MSG_pair_hello;\n      adv_saved_client_pub := Head(chan_client_backend).pubkey;\n      chan_client_backend := <<[type |-> MSG_pair_hello, token |-> Head(chan_client_backend).token, pubkey |-> adv_ecdh_pub]>> \\o Tail(chan_client_backend);"},
			{Name: "MitM_pair_hello_ack", Desc: "intercept pair_hello_ack and substitute adversary pubkey", Code: "      await Len(chan_backend_client) > 0 /\\ Head(chan_backend_client).type = MSG_pair_hello_ack;\n      adv_saved_server_pub := Head(chan_backend_client).pubkey;\n      adversary_keys := adversary_keys \\union {DeriveKey(adv_ecdh_pub, adv_saved_server_pub), DeriveKey(adv_ecdh_pub, adv_saved_client_pub)};\n      chan_backend_client := <<[type |-> MSG_pair_hello_ack, pubkey |-> adv_ecdh_pub]>> \\o Tail(chan_backend_client);"},
			{Name: "MitM_reencrypt_secret", Desc: "decrypt pair_complete with MitM key", Code: "      await Len(chan_backend_client) > 0 /\\ Head(chan_backend_client).type = MSG_pair_complete /\\ Head(chan_backend_client).key \\in adversary_keys;\n      with msg = Head(chan_backend_client) do\n        adversary_knowledge := adversary_knowledge \\union {[type |-> \"plaintext_secret\", secret |-> msg.secret]};\n        chan_backend_client := <<[type |-> MSG_pair_complete, key |-> DeriveKey(adv_ecdh_pub, adv_saved_client_pub), secret |-> msg.secret]>> \\o Tail(chan_backend_client);\n      end with;"},
			{Name: "concurrent_pair", Desc: "race a forged pair_hello using shoulder-surfed token", Code: "      await \\E m \\in adversary_knowledge : m = [type |-> \"qr_token\", token |-> current_token];\n      await Len(chan_client_backend) < 3;\n      chan_client_backend := Append(chan_client_backend, [type |-> MSG_pair_hello, token |-> current_token, pubkey |-> adv_ecdh_pub]);"},
			{Name: "token_bruteforce", Desc: "send pair_hello with fabricated token", Code: "      await Len(chan_client_backend) < 3;\n      chan_client_backend := Append(chan_client_backend, [type |-> MSG_pair_hello, token |-> \"fake_token\", pubkey |-> adv_ecdh_pub]);"},
			{Name: "code_guess", Desc: "submit fabricated confirmation code", Code: "      await backend_state = backend_WaitingForCode;\n      cli_entered_code := <<\"guess\", \"000000\">>;"},
			{Name: "session_replay", Desc: "replay captured auth_request with stale nonce", Code: "      await Len(chan_client_backend) < 3;\n      await \\E m \\in adversary_knowledge : m.type = MSG_auth_request;\n      with msg \\in {m \\in adversary_knowledge : m.type = MSG_auth_request} do\n        chan_client_backend := Append(chan_client_backend, msg);\n      end with;"},
		},
		Properties: []Property{
			{Name: "NoTokenReuse", Kind: Invariant, Expr: "used_tokens \\intersect active_tokens = {}", Desc: "A revoked pairing token is never accepted again"},
			{Name: "MitMDetectedByCodeMismatch", Kind: Invariant, Expr: "(backend_shared_key \\in adversary_keys /\\ backend_code /= <<\"none\">> /\\ client_code /= <<\"none\">>) => backend_code /= client_code", Desc: "MitM produces mismatched codes"},
			{Name: "MitMPrevented", Kind: Invariant, Expr: "backend_shared_key \\in adversary_keys => backend_state \\notin {backend_StorePaired, backend_Paired, backend_AuthCheck, backend_SessionActive}", Desc: "Compromised key prevents pairing completion"},
			{Name: "AuthRequiresCompletedPairing", Kind: Invariant, Expr: "backend_state = backend_SessionActive => received_device_id \\in paired_devices", Desc: "Session requires completed pairing"},
			{Name: "NoNonceReuse", Kind: Invariant, Expr: "backend_state = backend_SessionActive => received_auth_nonce \\notin (auth_nonces_used \\ {received_auth_nonce})", Desc: "Each auth nonce accepted at most once"},
			{Name: "DeviceSecretSecrecy", Kind: Invariant, Expr: "\\A m \\in adversary_knowledge : \"type\" \\in DOMAIN m => m.type /= \"plaintext_secret\"", Desc: "Adversary never learns device secret"},
			{Name: "PathConsistency", Kind: Invariant, Expr: "b_active_path \\in {\"relay\", \"alt\"} /\\ c_active_path \\in {\"relay\", \"alt\"}", Desc: "Paths are always valid"},
			{Name: "BackoffBounded", Kind: Invariant, Expr: "backoff_level <= max_backoff_level", Desc: "Backoff never exceeds cap"},
			{Name: "BackoffResetsOnSuccess", Kind: Invariant, Expr: "backend_state = backend_AltActive => backoff_level = 0", Desc: "alt-pair success resets backoff"},
			{Name: "DispatcherAlwaysBound", Kind: Invariant, Expr: "b_dispatcher_path \\in {\"relay\", \"alt\"} /\\ c_dispatcher_path \\in {\"relay\", \"alt\"}", Desc: "Dispatchers always bound to valid path"},
			{Name: "BackendDispatcherMatchesActive", Kind: Invariant, Expr: "backend_state = backend_AltActive => b_dispatcher_path = \"alt\"", Desc: "Backend dispatcher on alt-pair when alt-pair active"},
			{Name: "ClientDispatcherMatchesActive", Kind: Invariant, Expr: "client_state = client_AltActive => c_dispatcher_path = \"alt\"", Desc: "Client dispatcher on alt-pair when alt-pair active"},
			{Name: "MonitorOnlyWhenAlt", Kind: Invariant, Expr: "monitor_target = \"alt\" => backend_state \\in {backend_AltActive, backend_AltDegraded}", Desc: "Monitor only pings when alt-pair is active or degraded"},
			{Name: "FallbackLeadsToReadvertise", Kind: Invariant, Expr: "", Desc: "After fallback, backend eventually re-advertises candidates"},
			{Name: "DegradedLeadsToResolutionOrFallback", Kind: Invariant, Expr: "", Desc: "Degraded state eventually resolves (recovery or fallback)"},
		},
		ChannelBound: 3,
		OneShot:      false,
	}
}

type ECDHState struct {
	BackendPub string // backend ECDH public key
	ClientPub  string // pubkey received from client
	SharedKey  string // ECDH-derived shared key
	Code       string // confirmation code derived from pubkeys
}

type TokenState struct {
	Current string             // pairing token currently in play
	Active  frozen.Set[string] // set of valid (non-revoked) tokens
	Used    frozen.Set[string] // set of revoked tokens
}

type BackendPathState struct {
	ActivePath     string // which path carries traffic
	DispatcherPath string // datagram dispatcher binding
	MonitorTarget  string // health monitor target
	AltSignal      string // AltReady notification state
}

type ClientPathState struct {
	ActivePath     string // which path carries traffic
	DispatcherPath string // datagram dispatcher binding
}

type Candidate struct {
	Kind     string // candidate kind discriminator: cand_host / cand_srflx / ...
	Addr     string // transport address (host:port) the peer should dial
	Priority int    // pair-priority used during nomination (higher wins; ICE-style numbering recommended)
	CandId   string // stable per-session identifier; pair IDs are <<local_cand_id, remote_cand_id>>
}

// SessionProtocolBackendMachine is the generated state machine for the backend actor.
type SessionProtocolBackendMachine struct {
	State             State
	CurrentToken      string             // pairing token currently in play
	ActiveTokens      frozen.Set[string] // set of valid (non-revoked) tokens
	UsedTokens        frozen.Set[string] // set of revoked tokens
	BackendEcdhPub    string             // backend ECDH public key
	ReceivedClientPub string             // pubkey backend received in pair_hello
	BackendSharedKey  string             // ECDH key derived by backend
	BackendCode       string             // code computed by backend
	ReceivedCode      string             // code entered via CLI
	CodeAttempts      int                // failed code submission attempts
	DeviceSecret      string             // persistent device secret
	PairedDevices     frozen.Set[string] // device IDs that completed pairing
	ReceivedDeviceId  string             // device_id from auth_request
	AuthNoncesUsed    frozen.Set[string] // set of consumed auth nonces
	ReceivedAuthNonce string             // nonce from auth_request
	SecretPublished   bool               // whether token has been published via backchannel
	PingFailures      int                // consecutive failed pings
	BackoffLevel      int                // exponential backoff level
	BActivePath       string             // backend active path
	BDispatcherPath   string             // backend datagram dispatcher binding
	MonitorTarget     string             // health monitor target
	AltSignal         string             // AltReady notification state

	Guards   map[GuardID]func() bool
	Actions  map[ActionID]func() error
	OnChange func(varName string)
}

func NewSessionProtocolBackendMachine() *SessionProtocolBackendMachine {
	return &SessionProtocolBackendMachine{
		State:             SessionProtocolBackendIdle,
		CurrentToken:      "none",
		BackendEcdhPub:    "none",
		ReceivedClientPub: "none",
		BackendSharedKey:  "",
		BackendCode:       "",
		ReceivedCode:      "",
		CodeAttempts:      0,
		DeviceSecret:      "none",
		ReceivedDeviceId:  "none",
		ReceivedAuthNonce: "none",
		SecretPublished:   false,
		PingFailures:      0,
		BackoffLevel:      0,
		BActivePath:       "relay",
		BDispatcherPath:   "relay",
		MonitorTarget:     "none",
		AltSignal:         "pending",
		Guards:            make(map[GuardID]func() bool),
		Actions:           make(map[ActionID]func() error),
	}
}

func (m *SessionProtocolBackendMachine) HandleMessage(msg MsgType) (bool, error) {
	switch {
	case m.State == SessionProtocolBackendWaitingForClient && msg == SessionProtocolMsgPairHello && m.Guards[SessionProtocolGuardTokenValid] != nil && m.Guards[SessionProtocolGuardTokenValid]():
		if fn := m.Actions[SessionProtocolActionDeriveSecret]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		// received_client_pub: recv_msg.pubkey (set by action)
		m.BackendEcdhPub = "backend_pub"
		if m.OnChange != nil {
			m.OnChange("backend_ecdh_pub")
		}
		// backend_shared_key: DeriveKey("backend_pub", recv_msg.pubkey) (set by action)
		// backend_code: DeriveCode("backend_pub", recv_msg.pubkey) (set by action)
		m.State = SessionProtocolBackendDeriveSecret
		return true, nil
	case m.State == SessionProtocolBackendWaitingForClient && msg == SessionProtocolMsgPairHello && m.Guards[SessionProtocolGuardTokenInvalid] != nil && m.Guards[SessionProtocolGuardTokenInvalid]():
		m.State = SessionProtocolBackendIdle
		return true, nil
	case m.State == SessionProtocolBackendPaired && msg == SessionProtocolMsgAuthRequest:
		// received_device_id: recv_msg.device_id (set by action)
		// received_auth_nonce: recv_msg.nonce (set by action)
		m.State = SessionProtocolBackendAuthCheck
		return true, nil
	case m.State == SessionProtocolBackendCandidatesAdvertised && msg == SessionProtocolMsgPairCheck && m.Guards[SessionProtocolGuardChallengeValid] != nil && m.Guards[SessionProtocolGuardChallengeValid]():
		if fn := m.Actions[SessionProtocolActionActivateLan]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		m.PingFailures = 0
		if m.OnChange != nil {
			m.OnChange("ping_failures")
		}
		m.BackoffLevel = 0
		if m.OnChange != nil {
			m.OnChange("backoff_level")
		}
		m.BActivePath = "alt"
		if m.OnChange != nil {
			m.OnChange("b_active_path")
		}
		m.BDispatcherPath = "alt"
		if m.OnChange != nil {
			m.OnChange("b_dispatcher_path")
		}
		m.MonitorTarget = "alt"
		if m.OnChange != nil {
			m.OnChange("monitor_target")
		}
		m.AltSignal = "ready"
		if m.OnChange != nil {
			m.OnChange("alt_signal")
		}
		m.State = SessionProtocolBackendAltActive
		return true, nil
	case m.State == SessionProtocolBackendCandidatesAdvertised && msg == SessionProtocolMsgPairCheck && m.Guards[SessionProtocolGuardChallengeInvalid] != nil && m.Guards[SessionProtocolGuardChallengeInvalid]():
		m.State = SessionProtocolBackendRelayConnected
		return true, nil
	case m.State == SessionProtocolBackendAltDegraded && msg == SessionProtocolMsgPathPong:
		if fn := m.Actions[SessionProtocolActionResetFailures]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		m.PingFailures = 0
		if m.OnChange != nil {
			m.OnChange("ping_failures")
		}
		m.State = SessionProtocolBackendAltActive
		return true, nil
	}
	return false, nil
}

func (m *SessionProtocolBackendMachine) Step(event EventID) (bool, error) {
	switch {
	case m.State == SessionProtocolBackendIdle && event == SessionProtocolEventCliInitPair:
		if fn := m.Actions[SessionProtocolActionGenerateToken]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		m.CurrentToken = "tok_1"
		if m.OnChange != nil {
			m.OnChange("current_token")
		}
		// active_tokens: active_tokens \union {"tok_1"} (set by action)
		m.State = SessionProtocolBackendGenerateToken
		return true, nil
	case m.State == SessionProtocolBackendGenerateToken && event == SessionProtocolEventTokenCreated:
		if fn := m.Actions[SessionProtocolActionRegisterRelay]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		m.State = SessionProtocolBackendRegisterRelay
		return true, nil
	case m.State == SessionProtocolBackendRegisterRelay && event == SessionProtocolEventRelayRegistered:
		m.SecretPublished = true
		if m.OnChange != nil {
			m.OnChange("secret_published")
		}
		m.State = SessionProtocolBackendWaitingForClient
		return true, nil
	case m.State == SessionProtocolBackendDeriveSecret && event == SessionProtocolEventEcdhComplete:
		m.State = SessionProtocolBackendSendAck
		return true, nil
	case m.State == SessionProtocolBackendSendAck && event == SessionProtocolEventSignalCodeDisplay:
		m.State = SessionProtocolBackendWaitingForCode
		return true, nil
	case m.State == SessionProtocolBackendWaitingForCode && event == SessionProtocolEventCliCodeEntered:
		// received_code: cli_entered_code (set by action)
		m.State = SessionProtocolBackendValidateCode
		return true, nil
	case m.State == SessionProtocolBackendValidateCode && event == SessionProtocolEventCheckCode && m.Guards[SessionProtocolGuardCodeCorrect] != nil && m.Guards[SessionProtocolGuardCodeCorrect]():
		m.State = SessionProtocolBackendStorePaired
		return true, nil
	case m.State == SessionProtocolBackendValidateCode && event == SessionProtocolEventCheckCode && m.Guards[SessionProtocolGuardCodeWrong] != nil && m.Guards[SessionProtocolGuardCodeWrong]():
		m.CodeAttempts = m.CodeAttempts + 1
		if m.OnChange != nil {
			m.OnChange("code_attempts")
		}
		m.State = SessionProtocolBackendIdle
		return true, nil
	case m.State == SessionProtocolBackendStorePaired && event == SessionProtocolEventFinalise:
		if fn := m.Actions[SessionProtocolActionStoreDevice]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		m.DeviceSecret = "dev_secret_1"
		if m.OnChange != nil {
			m.OnChange("device_secret")
		}
		// paired_devices: paired_devices \union {"device_1"} (set by action)
		// active_tokens: active_tokens \ {current_token} (set by action)
		// used_tokens: used_tokens \union {current_token} (set by action)
		m.State = SessionProtocolBackendPaired
		return true, nil
	case m.State == SessionProtocolBackendAuthCheck && event == SessionProtocolEventVerify && m.Guards[SessionProtocolGuardDeviceKnown] != nil && m.Guards[SessionProtocolGuardDeviceKnown]():
		if fn := m.Actions[SessionProtocolActionVerifyDevice]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		// auth_nonces_used: auth_nonces_used \union {received_auth_nonce} (set by action)
		m.State = SessionProtocolBackendSessionActive
		return true, nil
	case m.State == SessionProtocolBackendAuthCheck && event == SessionProtocolEventVerify && m.Guards[SessionProtocolGuardDeviceUnknown] != nil && m.Guards[SessionProtocolGuardDeviceUnknown]():
		m.State = SessionProtocolBackendIdle
		return true, nil
	case m.State == SessionProtocolBackendSessionActive && event == SessionProtocolEventSessionEstablished:
		m.State = SessionProtocolBackendRelayConnected
		return true, nil
	case m.State == SessionProtocolBackendRelayConnected && event == SessionProtocolEventCandidatesGathered:
		m.State = SessionProtocolBackendCandidatesAdvertised
		return true, nil
	case m.State == SessionProtocolBackendCandidatesAdvertised && event == SessionProtocolEventCandidatesTimeout:
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.AltSignal = "pending"
		if m.OnChange != nil {
			m.OnChange("alt_signal")
		}
		m.State = SessionProtocolBackendRelayBackoff
		return true, nil
	case m.State == SessionProtocolBackendAltActive && event == SessionProtocolEventPingTick:
		m.State = SessionProtocolBackendAltActive
		return true, nil
	case m.State == SessionProtocolBackendAltActive && event == SessionProtocolEventPingTimeout:
		m.PingFailures = 1
		if m.OnChange != nil {
			m.OnChange("ping_failures")
		}
		m.State = SessionProtocolBackendAltDegraded
		return true, nil
	case m.State == SessionProtocolBackendAltDegraded && event == SessionProtocolEventPingTick:
		m.State = SessionProtocolBackendAltDegraded
		return true, nil
	case m.State == SessionProtocolBackendAltActive && event == SessionProtocolEventAltStreamError:
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.BActivePath = "relay"
		if m.OnChange != nil {
			m.OnChange("b_active_path")
		}
		m.BDispatcherPath = "relay"
		if m.OnChange != nil {
			m.OnChange("b_dispatcher_path")
		}
		m.MonitorTarget = "none"
		if m.OnChange != nil {
			m.OnChange("monitor_target")
		}
		m.AltSignal = "pending"
		if m.OnChange != nil {
			m.OnChange("alt_signal")
		}
		m.PingFailures = 0
		if m.OnChange != nil {
			m.OnChange("ping_failures")
		}
		m.State = SessionProtocolBackendRelayBackoff
		return true, nil
	case m.State == SessionProtocolBackendAltDegraded && event == SessionProtocolEventAltStreamError:
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.BActivePath = "relay"
		if m.OnChange != nil {
			m.OnChange("b_active_path")
		}
		m.BDispatcherPath = "relay"
		if m.OnChange != nil {
			m.OnChange("b_dispatcher_path")
		}
		m.MonitorTarget = "none"
		if m.OnChange != nil {
			m.OnChange("monitor_target")
		}
		m.AltSignal = "pending"
		if m.OnChange != nil {
			m.OnChange("alt_signal")
		}
		m.PingFailures = 0
		if m.OnChange != nil {
			m.OnChange("ping_failures")
		}
		m.State = SessionProtocolBackendRelayBackoff
		return true, nil
	case m.State == SessionProtocolBackendAltDegraded && event == SessionProtocolEventPingTimeout && m.Guards[SessionProtocolGuardUnderMaxFailures] != nil && m.Guards[SessionProtocolGuardUnderMaxFailures]():
		m.PingFailures = m.PingFailures + 1
		if m.OnChange != nil {
			m.OnChange("ping_failures")
		}
		m.State = SessionProtocolBackendAltDegraded
		return true, nil
	case m.State == SessionProtocolBackendAltDegraded && event == SessionProtocolEventPingTimeout && m.Guards[SessionProtocolGuardAtMaxFailures] != nil && m.Guards[SessionProtocolGuardAtMaxFailures]():
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.BActivePath = "relay"
		if m.OnChange != nil {
			m.OnChange("b_active_path")
		}
		m.BDispatcherPath = "relay"
		if m.OnChange != nil {
			m.OnChange("b_dispatcher_path")
		}
		m.MonitorTarget = "none"
		if m.OnChange != nil {
			m.OnChange("monitor_target")
		}
		m.AltSignal = "pending"
		if m.OnChange != nil {
			m.OnChange("alt_signal")
		}
		m.PingFailures = 0
		if m.OnChange != nil {
			m.OnChange("ping_failures")
		}
		m.State = SessionProtocolBackendRelayBackoff
		return true, nil
	case m.State == SessionProtocolBackendRelayBackoff && event == SessionProtocolEventBackoffExpired:
		m.State = SessionProtocolBackendCandidatesAdvertised
		return true, nil
	case m.State == SessionProtocolBackendRelayBackoff && event == SessionProtocolEventCandidatesChanged:
		m.BackoffLevel = 0
		if m.OnChange != nil {
			m.OnChange("backoff_level")
		}
		m.State = SessionProtocolBackendCandidatesAdvertised
		return true, nil
	case m.State == SessionProtocolBackendRelayConnected && event == SessionProtocolEventCandidatesRefreshTick && m.Guards[SessionProtocolGuardLocalCandidatesAvailable] != nil && m.Guards[SessionProtocolGuardLocalCandidatesAvailable]():
		m.State = SessionProtocolBackendCandidatesAdvertised
		return true, nil
	case m.State == SessionProtocolBackendCandidatesAdvertised && event == SessionProtocolEventAppForceFallback:
		m.AltSignal = "pending"
		if m.OnChange != nil {
			m.OnChange("alt_signal")
		}
		m.State = SessionProtocolBackendRelayConnected
		return true, nil
	case m.State == SessionProtocolBackendAltActive && event == SessionProtocolEventAppForceFallback:
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.BActivePath = "relay"
		if m.OnChange != nil {
			m.OnChange("b_active_path")
		}
		m.BDispatcherPath = "relay"
		if m.OnChange != nil {
			m.OnChange("b_dispatcher_path")
		}
		m.MonitorTarget = "none"
		if m.OnChange != nil {
			m.OnChange("monitor_target")
		}
		m.AltSignal = "pending"
		if m.OnChange != nil {
			m.OnChange("alt_signal")
		}
		m.PingFailures = 0
		if m.OnChange != nil {
			m.OnChange("ping_failures")
		}
		m.State = SessionProtocolBackendRelayBackoff
		return true, nil
	case m.State == SessionProtocolBackendAltDegraded && event == SessionProtocolEventAppForceFallback:
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.BActivePath = "relay"
		if m.OnChange != nil {
			m.OnChange("b_active_path")
		}
		m.BDispatcherPath = "relay"
		if m.OnChange != nil {
			m.OnChange("b_dispatcher_path")
		}
		m.MonitorTarget = "none"
		if m.OnChange != nil {
			m.OnChange("monitor_target")
		}
		m.AltSignal = "pending"
		if m.OnChange != nil {
			m.OnChange("alt_signal")
		}
		m.PingFailures = 0
		if m.OnChange != nil {
			m.OnChange("ping_failures")
		}
		m.State = SessionProtocolBackendRelayBackoff
		return true, nil
	case m.State == SessionProtocolBackendRelayConnected && event == SessionProtocolEventDisconnect:
		m.State = SessionProtocolBackendPaired
		return true, nil
	case m.State == SessionProtocolBackendRelayConnected && event == SessionProtocolEventAppSend:
		m.State = SessionProtocolBackendRelayConnected
		return true, nil
	case m.State == SessionProtocolBackendCandidatesAdvertised && event == SessionProtocolEventAppSend:
		m.State = SessionProtocolBackendCandidatesAdvertised
		return true, nil
	case m.State == SessionProtocolBackendAltActive && event == SessionProtocolEventAppSend:
		m.State = SessionProtocolBackendAltActive
		return true, nil
	case m.State == SessionProtocolBackendAltDegraded && event == SessionProtocolEventAppSend:
		m.State = SessionProtocolBackendAltDegraded
		return true, nil
	case m.State == SessionProtocolBackendRelayBackoff && event == SessionProtocolEventAppSend:
		m.State = SessionProtocolBackendRelayBackoff
		return true, nil
	case m.State == SessionProtocolBackendRelayConnected && event == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolBackendRelayConnected
		return true, nil
	case m.State == SessionProtocolBackendCandidatesAdvertised && event == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolBackendCandidatesAdvertised
		return true, nil
	case m.State == SessionProtocolBackendAltActive && event == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolBackendAltActive
		return true, nil
	case m.State == SessionProtocolBackendAltDegraded && event == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolBackendAltDegraded
		return true, nil
	case m.State == SessionProtocolBackendRelayBackoff && event == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolBackendRelayBackoff
		return true, nil
	case m.State == SessionProtocolBackendRelayConnected && event == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolBackendRelayConnected
		return true, nil
	case m.State == SessionProtocolBackendCandidatesAdvertised && event == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolBackendCandidatesAdvertised
		return true, nil
	case m.State == SessionProtocolBackendAltActive && event == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolBackendAltActive
		return true, nil
	case m.State == SessionProtocolBackendAltDegraded && event == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolBackendAltDegraded
		return true, nil
	case m.State == SessionProtocolBackendRelayBackoff && event == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolBackendRelayBackoff
		return true, nil
	case m.State == SessionProtocolBackendRelayConnected && event == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolBackendRelayConnected
		return true, nil
	case m.State == SessionProtocolBackendCandidatesAdvertised && event == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolBackendCandidatesAdvertised
		return true, nil
	case m.State == SessionProtocolBackendAltActive && event == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolBackendAltActive
		return true, nil
	case m.State == SessionProtocolBackendAltDegraded && event == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolBackendAltDegraded
		return true, nil
	case m.State == SessionProtocolBackendRelayBackoff && event == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolBackendRelayBackoff
		return true, nil
	case m.State == SessionProtocolBackendRelayConnected && event == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolBackendRelayConnected
		return true, nil
	case m.State == SessionProtocolBackendCandidatesAdvertised && event == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolBackendCandidatesAdvertised
		return true, nil
	case m.State == SessionProtocolBackendAltActive && event == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolBackendAltActive
		return true, nil
	case m.State == SessionProtocolBackendAltDegraded && event == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolBackendAltDegraded
		return true, nil
	case m.State == SessionProtocolBackendRelayBackoff && event == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolBackendRelayBackoff
		return true, nil
	case m.State == SessionProtocolBackendAltActive && event == SessionProtocolEventAltStreamData:
		m.State = SessionProtocolBackendAltActive
		return true, nil
	case m.State == SessionProtocolBackendAltDegraded && event == SessionProtocolEventAltStreamData:
		m.State = SessionProtocolBackendAltDegraded
		return true, nil
	case m.State == SessionProtocolBackendAltActive && event == SessionProtocolEventAltDatagram:
		m.State = SessionProtocolBackendAltActive
		return true, nil
	case m.State == SessionProtocolBackendAltDegraded && event == SessionProtocolEventAltDatagram:
		m.State = SessionProtocolBackendAltDegraded
		return true, nil
	}
	return false, nil
}

func (m *SessionProtocolBackendMachine) HandleEvent(ev EventID) ([]CmdID, error) {
	switch {
	case m.State == SessionProtocolBackendIdle && ev == SessionProtocolEventCliInitPair:
		if fn := m.Actions[SessionProtocolActionGenerateToken]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		m.CurrentToken = "tok_1"
		if m.OnChange != nil {
			m.OnChange("current_token")
		}
		// active_tokens: active_tokens \union {"tok_1"} (set by action)
		m.State = SessionProtocolBackendGenerateToken
		return nil, nil
	case m.State == SessionProtocolBackendGenerateToken && ev == SessionProtocolEventTokenCreated:
		if fn := m.Actions[SessionProtocolActionRegisterRelay]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		m.State = SessionProtocolBackendRegisterRelay
		return nil, nil
	case m.State == SessionProtocolBackendRegisterRelay && ev == SessionProtocolEventRelayRegistered:
		m.SecretPublished = true
		if m.OnChange != nil {
			m.OnChange("secret_published")
		}
		m.State = SessionProtocolBackendWaitingForClient
		return nil, nil
	case m.State == SessionProtocolBackendWaitingForClient && ev == SessionProtocolEventRecvPairHello && m.Guards[SessionProtocolGuardTokenValid] != nil && m.Guards[SessionProtocolGuardTokenValid]():
		if fn := m.Actions[SessionProtocolActionDeriveSecret]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		// received_client_pub: recv_msg.pubkey (set by action)
		m.BackendEcdhPub = "backend_pub"
		if m.OnChange != nil {
			m.OnChange("backend_ecdh_pub")
		}
		// backend_shared_key: DeriveKey("backend_pub", recv_msg.pubkey) (set by action)
		// backend_code: DeriveCode("backend_pub", recv_msg.pubkey) (set by action)
		m.State = SessionProtocolBackendDeriveSecret
		return nil, nil
	case m.State == SessionProtocolBackendWaitingForClient && ev == SessionProtocolEventRecvPairHello && m.Guards[SessionProtocolGuardTokenInvalid] != nil && m.Guards[SessionProtocolGuardTokenInvalid]():
		m.State = SessionProtocolBackendIdle
		return nil, nil
	case m.State == SessionProtocolBackendDeriveSecret && ev == SessionProtocolEventEcdhComplete:
		m.State = SessionProtocolBackendSendAck
		return nil, nil
	case m.State == SessionProtocolBackendSendAck && ev == SessionProtocolEventSignalCodeDisplay:
		m.State = SessionProtocolBackendWaitingForCode
		return nil, nil
	case m.State == SessionProtocolBackendWaitingForCode && ev == SessionProtocolEventCliCodeEntered:
		// received_code: cli_entered_code (set by action)
		m.State = SessionProtocolBackendValidateCode
		return nil, nil
	case m.State == SessionProtocolBackendValidateCode && ev == SessionProtocolEventCheckCode && m.Guards[SessionProtocolGuardCodeCorrect] != nil && m.Guards[SessionProtocolGuardCodeCorrect]():
		m.State = SessionProtocolBackendStorePaired
		return nil, nil
	case m.State == SessionProtocolBackendValidateCode && ev == SessionProtocolEventCheckCode && m.Guards[SessionProtocolGuardCodeWrong] != nil && m.Guards[SessionProtocolGuardCodeWrong]():
		m.CodeAttempts = m.CodeAttempts + 1
		if m.OnChange != nil {
			m.OnChange("code_attempts")
		}
		m.State = SessionProtocolBackendIdle
		return nil, nil
	case m.State == SessionProtocolBackendStorePaired && ev == SessionProtocolEventFinalise:
		if fn := m.Actions[SessionProtocolActionStoreDevice]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		m.DeviceSecret = "dev_secret_1"
		if m.OnChange != nil {
			m.OnChange("device_secret")
		}
		// paired_devices: paired_devices \union {"device_1"} (set by action)
		// active_tokens: active_tokens \ {current_token} (set by action)
		// used_tokens: used_tokens \union {current_token} (set by action)
		m.State = SessionProtocolBackendPaired
		return nil, nil
	case m.State == SessionProtocolBackendPaired && ev == SessionProtocolEventRecvAuthRequest:
		// received_device_id: recv_msg.device_id (set by action)
		// received_auth_nonce: recv_msg.nonce (set by action)
		m.State = SessionProtocolBackendAuthCheck
		return nil, nil
	case m.State == SessionProtocolBackendAuthCheck && ev == SessionProtocolEventVerify && m.Guards[SessionProtocolGuardDeviceKnown] != nil && m.Guards[SessionProtocolGuardDeviceKnown]():
		if fn := m.Actions[SessionProtocolActionVerifyDevice]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		// auth_nonces_used: auth_nonces_used \union {received_auth_nonce} (set by action)
		m.State = SessionProtocolBackendSessionActive
		return nil, nil
	case m.State == SessionProtocolBackendAuthCheck && ev == SessionProtocolEventVerify && m.Guards[SessionProtocolGuardDeviceUnknown] != nil && m.Guards[SessionProtocolGuardDeviceUnknown]():
		m.State = SessionProtocolBackendIdle
		return nil, nil
	case m.State == SessionProtocolBackendSessionActive && ev == SessionProtocolEventSessionEstablished:
		m.State = SessionProtocolBackendRelayConnected
		return nil, nil
	case m.State == SessionProtocolBackendRelayConnected && ev == SessionProtocolEventCandidatesGathered:
		m.State = SessionProtocolBackendCandidatesAdvertised
		return []CmdID{SessionProtocolCmdSendCandidates}, nil
	case m.State == SessionProtocolBackendCandidatesAdvertised && ev == SessionProtocolEventRecvPairCheck && m.Guards[SessionProtocolGuardChallengeValid] != nil && m.Guards[SessionProtocolGuardChallengeValid]():
		if fn := m.Actions[SessionProtocolActionActivateLan]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		m.PingFailures = 0
		if m.OnChange != nil {
			m.OnChange("ping_failures")
		}
		m.BackoffLevel = 0
		if m.OnChange != nil {
			m.OnChange("backoff_level")
		}
		m.BActivePath = "alt"
		if m.OnChange != nil {
			m.OnChange("b_active_path")
		}
		m.BDispatcherPath = "alt"
		if m.OnChange != nil {
			m.OnChange("b_dispatcher_path")
		}
		m.MonitorTarget = "alt"
		if m.OnChange != nil {
			m.OnChange("monitor_target")
		}
		m.AltSignal = "ready"
		if m.OnChange != nil {
			m.OnChange("alt_signal")
		}
		m.State = SessionProtocolBackendAltActive
		return []CmdID{SessionProtocolCmdSendPairCheckAck, SessionProtocolCmdStartAltStreamReader, SessionProtocolCmdStartAltDgReader, SessionProtocolCmdStartMonitor, SessionProtocolCmdSignalAltReady, SessionProtocolCmdSetCryptoDatagram}, nil
	case m.State == SessionProtocolBackendCandidatesAdvertised && ev == SessionProtocolEventRecvPairCheck && m.Guards[SessionProtocolGuardChallengeInvalid] != nil && m.Guards[SessionProtocolGuardChallengeInvalid]():
		m.State = SessionProtocolBackendRelayConnected
		return nil, nil
	case m.State == SessionProtocolBackendCandidatesAdvertised && ev == SessionProtocolEventCandidatesTimeout:
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.AltSignal = "pending"
		if m.OnChange != nil {
			m.OnChange("alt_signal")
		}
		m.State = SessionProtocolBackendRelayBackoff
		return []CmdID{SessionProtocolCmdResetAltReady, SessionProtocolCmdStartBackoffTimer}, nil
	case m.State == SessionProtocolBackendAltActive && ev == SessionProtocolEventPingTick:
		m.State = SessionProtocolBackendAltActive
		return []CmdID{SessionProtocolCmdSendPathPing, SessionProtocolCmdStartPongTimeout}, nil
	case m.State == SessionProtocolBackendAltActive && ev == SessionProtocolEventPingTimeout:
		m.PingFailures = 1
		if m.OnChange != nil {
			m.OnChange("ping_failures")
		}
		m.State = SessionProtocolBackendAltDegraded
		return nil, nil
	case m.State == SessionProtocolBackendAltDegraded && ev == SessionProtocolEventPingTick:
		m.State = SessionProtocolBackendAltDegraded
		return []CmdID{SessionProtocolCmdSendPathPing, SessionProtocolCmdStartPongTimeout}, nil
	case m.State == SessionProtocolBackendAltActive && ev == SessionProtocolEventAltStreamError:
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.BActivePath = "relay"
		if m.OnChange != nil {
			m.OnChange("b_active_path")
		}
		m.BDispatcherPath = "relay"
		if m.OnChange != nil {
			m.OnChange("b_dispatcher_path")
		}
		m.MonitorTarget = "none"
		if m.OnChange != nil {
			m.OnChange("monitor_target")
		}
		m.AltSignal = "pending"
		if m.OnChange != nil {
			m.OnChange("alt_signal")
		}
		m.PingFailures = 0
		if m.OnChange != nil {
			m.OnChange("ping_failures")
		}
		m.State = SessionProtocolBackendRelayBackoff
		return []CmdID{SessionProtocolCmdStopMonitor, SessionProtocolCmdStopAltStreamReader, SessionProtocolCmdStopAltDgReader, SessionProtocolCmdCloseAltPath, SessionProtocolCmdResetAltReady, SessionProtocolCmdStartBackoffTimer}, nil
	case m.State == SessionProtocolBackendAltDegraded && ev == SessionProtocolEventAltStreamError:
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.BActivePath = "relay"
		if m.OnChange != nil {
			m.OnChange("b_active_path")
		}
		m.BDispatcherPath = "relay"
		if m.OnChange != nil {
			m.OnChange("b_dispatcher_path")
		}
		m.MonitorTarget = "none"
		if m.OnChange != nil {
			m.OnChange("monitor_target")
		}
		m.AltSignal = "pending"
		if m.OnChange != nil {
			m.OnChange("alt_signal")
		}
		m.PingFailures = 0
		if m.OnChange != nil {
			m.OnChange("ping_failures")
		}
		m.State = SessionProtocolBackendRelayBackoff
		return []CmdID{SessionProtocolCmdStopMonitor, SessionProtocolCmdStopAltStreamReader, SessionProtocolCmdStopAltDgReader, SessionProtocolCmdCloseAltPath, SessionProtocolCmdResetAltReady, SessionProtocolCmdStartBackoffTimer}, nil
	case m.State == SessionProtocolBackendAltDegraded && ev == SessionProtocolEventRecvPathPong:
		if fn := m.Actions[SessionProtocolActionResetFailures]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		m.PingFailures = 0
		if m.OnChange != nil {
			m.OnChange("ping_failures")
		}
		m.State = SessionProtocolBackendAltActive
		return []CmdID{SessionProtocolCmdCancelPongTimeout}, nil
	case m.State == SessionProtocolBackendAltDegraded && ev == SessionProtocolEventPingTimeout && m.Guards[SessionProtocolGuardUnderMaxFailures] != nil && m.Guards[SessionProtocolGuardUnderMaxFailures]():
		m.PingFailures = m.PingFailures + 1
		if m.OnChange != nil {
			m.OnChange("ping_failures")
		}
		m.State = SessionProtocolBackendAltDegraded
		return nil, nil
	case m.State == SessionProtocolBackendAltDegraded && ev == SessionProtocolEventPingTimeout && m.Guards[SessionProtocolGuardAtMaxFailures] != nil && m.Guards[SessionProtocolGuardAtMaxFailures]():
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.BActivePath = "relay"
		if m.OnChange != nil {
			m.OnChange("b_active_path")
		}
		m.BDispatcherPath = "relay"
		if m.OnChange != nil {
			m.OnChange("b_dispatcher_path")
		}
		m.MonitorTarget = "none"
		if m.OnChange != nil {
			m.OnChange("monitor_target")
		}
		m.AltSignal = "pending"
		if m.OnChange != nil {
			m.OnChange("alt_signal")
		}
		m.PingFailures = 0
		if m.OnChange != nil {
			m.OnChange("ping_failures")
		}
		m.State = SessionProtocolBackendRelayBackoff
		return []CmdID{SessionProtocolCmdStopMonitor, SessionProtocolCmdStopAltStreamReader, SessionProtocolCmdStopAltDgReader, SessionProtocolCmdCloseAltPath, SessionProtocolCmdResetAltReady, SessionProtocolCmdStartBackoffTimer}, nil
	case m.State == SessionProtocolBackendRelayBackoff && ev == SessionProtocolEventBackoffExpired:
		m.State = SessionProtocolBackendCandidatesAdvertised
		return []CmdID{SessionProtocolCmdSendCandidates}, nil
	case m.State == SessionProtocolBackendRelayBackoff && ev == SessionProtocolEventCandidatesChanged:
		m.BackoffLevel = 0
		if m.OnChange != nil {
			m.OnChange("backoff_level")
		}
		m.State = SessionProtocolBackendCandidatesAdvertised
		return []CmdID{SessionProtocolCmdSendCandidates}, nil
	case m.State == SessionProtocolBackendRelayConnected && ev == SessionProtocolEventCandidatesRefreshTick && m.Guards[SessionProtocolGuardLocalCandidatesAvailable] != nil && m.Guards[SessionProtocolGuardLocalCandidatesAvailable]():
		m.State = SessionProtocolBackendCandidatesAdvertised
		return []CmdID{SessionProtocolCmdSendCandidates}, nil
	case m.State == SessionProtocolBackendCandidatesAdvertised && ev == SessionProtocolEventAppForceFallback:
		m.AltSignal = "pending"
		if m.OnChange != nil {
			m.OnChange("alt_signal")
		}
		m.State = SessionProtocolBackendRelayConnected
		return []CmdID{SessionProtocolCmdResetAltReady}, nil
	case m.State == SessionProtocolBackendAltActive && ev == SessionProtocolEventAppForceFallback:
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.BActivePath = "relay"
		if m.OnChange != nil {
			m.OnChange("b_active_path")
		}
		m.BDispatcherPath = "relay"
		if m.OnChange != nil {
			m.OnChange("b_dispatcher_path")
		}
		m.MonitorTarget = "none"
		if m.OnChange != nil {
			m.OnChange("monitor_target")
		}
		m.AltSignal = "pending"
		if m.OnChange != nil {
			m.OnChange("alt_signal")
		}
		m.PingFailures = 0
		if m.OnChange != nil {
			m.OnChange("ping_failures")
		}
		m.State = SessionProtocolBackendRelayBackoff
		return []CmdID{SessionProtocolCmdStopMonitor, SessionProtocolCmdCancelPongTimeout, SessionProtocolCmdStopAltStreamReader, SessionProtocolCmdStopAltDgReader, SessionProtocolCmdCloseAltPath, SessionProtocolCmdResetAltReady, SessionProtocolCmdStartBackoffTimer}, nil
	case m.State == SessionProtocolBackendAltDegraded && ev == SessionProtocolEventAppForceFallback:
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.BActivePath = "relay"
		if m.OnChange != nil {
			m.OnChange("b_active_path")
		}
		m.BDispatcherPath = "relay"
		if m.OnChange != nil {
			m.OnChange("b_dispatcher_path")
		}
		m.MonitorTarget = "none"
		if m.OnChange != nil {
			m.OnChange("monitor_target")
		}
		m.AltSignal = "pending"
		if m.OnChange != nil {
			m.OnChange("alt_signal")
		}
		m.PingFailures = 0
		if m.OnChange != nil {
			m.OnChange("ping_failures")
		}
		m.State = SessionProtocolBackendRelayBackoff
		return []CmdID{SessionProtocolCmdStopMonitor, SessionProtocolCmdCancelPongTimeout, SessionProtocolCmdStopAltStreamReader, SessionProtocolCmdStopAltDgReader, SessionProtocolCmdCloseAltPath, SessionProtocolCmdResetAltReady, SessionProtocolCmdStartBackoffTimer}, nil
	case m.State == SessionProtocolBackendRelayConnected && ev == SessionProtocolEventDisconnect:
		m.State = SessionProtocolBackendPaired
		return nil, nil
	case m.State == SessionProtocolBackendRelayConnected && ev == SessionProtocolEventAppSend:
		m.State = SessionProtocolBackendRelayConnected
		return []CmdID{SessionProtocolCmdWriteActiveStream}, nil
	case m.State == SessionProtocolBackendCandidatesAdvertised && ev == SessionProtocolEventAppSend:
		m.State = SessionProtocolBackendCandidatesAdvertised
		return []CmdID{SessionProtocolCmdWriteActiveStream}, nil
	case m.State == SessionProtocolBackendAltActive && ev == SessionProtocolEventAppSend:
		m.State = SessionProtocolBackendAltActive
		return []CmdID{SessionProtocolCmdWriteActiveStream}, nil
	case m.State == SessionProtocolBackendAltDegraded && ev == SessionProtocolEventAppSend:
		m.State = SessionProtocolBackendAltDegraded
		return []CmdID{SessionProtocolCmdWriteActiveStream}, nil
	case m.State == SessionProtocolBackendRelayBackoff && ev == SessionProtocolEventAppSend:
		m.State = SessionProtocolBackendRelayBackoff
		return []CmdID{SessionProtocolCmdWriteActiveStream}, nil
	case m.State == SessionProtocolBackendRelayConnected && ev == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolBackendRelayConnected
		return []CmdID{SessionProtocolCmdDeliverRecv}, nil
	case m.State == SessionProtocolBackendCandidatesAdvertised && ev == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolBackendCandidatesAdvertised
		return []CmdID{SessionProtocolCmdDeliverRecv}, nil
	case m.State == SessionProtocolBackendAltActive && ev == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolBackendAltActive
		return []CmdID{SessionProtocolCmdDeliverRecv}, nil
	case m.State == SessionProtocolBackendAltDegraded && ev == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolBackendAltDegraded
		return []CmdID{SessionProtocolCmdDeliverRecv}, nil
	case m.State == SessionProtocolBackendRelayBackoff && ev == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolBackendRelayBackoff
		return []CmdID{SessionProtocolCmdDeliverRecv}, nil
	case m.State == SessionProtocolBackendRelayConnected && ev == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolBackendRelayConnected
		return []CmdID{SessionProtocolCmdDeliverRecvError}, nil
	case m.State == SessionProtocolBackendCandidatesAdvertised && ev == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolBackendCandidatesAdvertised
		return []CmdID{SessionProtocolCmdDeliverRecvError}, nil
	case m.State == SessionProtocolBackendAltActive && ev == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolBackendAltActive
		return []CmdID{SessionProtocolCmdDeliverRecvError}, nil
	case m.State == SessionProtocolBackendAltDegraded && ev == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolBackendAltDegraded
		return []CmdID{SessionProtocolCmdDeliverRecvError}, nil
	case m.State == SessionProtocolBackendRelayBackoff && ev == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolBackendRelayBackoff
		return []CmdID{SessionProtocolCmdDeliverRecvError}, nil
	case m.State == SessionProtocolBackendRelayConnected && ev == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolBackendRelayConnected
		return []CmdID{SessionProtocolCmdSendActiveDatagram}, nil
	case m.State == SessionProtocolBackendCandidatesAdvertised && ev == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolBackendCandidatesAdvertised
		return []CmdID{SessionProtocolCmdSendActiveDatagram}, nil
	case m.State == SessionProtocolBackendAltActive && ev == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolBackendAltActive
		return []CmdID{SessionProtocolCmdSendActiveDatagram}, nil
	case m.State == SessionProtocolBackendAltDegraded && ev == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolBackendAltDegraded
		return []CmdID{SessionProtocolCmdSendActiveDatagram}, nil
	case m.State == SessionProtocolBackendRelayBackoff && ev == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolBackendRelayBackoff
		return []CmdID{SessionProtocolCmdSendActiveDatagram}, nil
	case m.State == SessionProtocolBackendRelayConnected && ev == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolBackendRelayConnected
		return []CmdID{SessionProtocolCmdDeliverRecvDatagram}, nil
	case m.State == SessionProtocolBackendCandidatesAdvertised && ev == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolBackendCandidatesAdvertised
		return []CmdID{SessionProtocolCmdDeliverRecvDatagram}, nil
	case m.State == SessionProtocolBackendAltActive && ev == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolBackendAltActive
		return []CmdID{SessionProtocolCmdDeliverRecvDatagram}, nil
	case m.State == SessionProtocolBackendAltDegraded && ev == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolBackendAltDegraded
		return []CmdID{SessionProtocolCmdDeliverRecvDatagram}, nil
	case m.State == SessionProtocolBackendRelayBackoff && ev == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolBackendRelayBackoff
		return []CmdID{SessionProtocolCmdDeliverRecvDatagram}, nil
	case m.State == SessionProtocolBackendAltActive && ev == SessionProtocolEventAltStreamData:
		m.State = SessionProtocolBackendAltActive
		return []CmdID{SessionProtocolCmdDeliverRecv}, nil
	case m.State == SessionProtocolBackendAltDegraded && ev == SessionProtocolEventAltStreamData:
		m.State = SessionProtocolBackendAltDegraded
		return []CmdID{SessionProtocolCmdDeliverRecv}, nil
	case m.State == SessionProtocolBackendAltActive && ev == SessionProtocolEventAltDatagram:
		m.State = SessionProtocolBackendAltActive
		return []CmdID{SessionProtocolCmdDeliverRecvDatagram}, nil
	case m.State == SessionProtocolBackendAltDegraded && ev == SessionProtocolEventAltDatagram:
		m.State = SessionProtocolBackendAltDegraded
		return []CmdID{SessionProtocolCmdDeliverRecvDatagram}, nil
	}
	return nil, nil
}

// SessionProtocolClientMachine is the generated state machine for the client actor.
type SessionProtocolClientMachine struct {
	State              State
	ReceivedBackendPub string // pubkey client received in pair_hello_ack
	ClientSharedKey    string // ECDH key derived by client
	ClientCode         string // code computed by client
	CActivePath        string // client active path
	CDispatcherPath    string // client datagram dispatcher binding
	AltSignal          string // AltReady notification state

	Guards   map[GuardID]func() bool
	Actions  map[ActionID]func() error
	OnChange func(varName string)
}

func NewSessionProtocolClientMachine() *SessionProtocolClientMachine {
	return &SessionProtocolClientMachine{
		State:              SessionProtocolClientIdle,
		ReceivedBackendPub: "none",
		ClientSharedKey:    "",
		ClientCode:         "",
		CActivePath:        "relay",
		CDispatcherPath:    "relay",
		AltSignal:          "pending",
		Guards:             make(map[GuardID]func() bool),
		Actions:            make(map[ActionID]func() error),
	}
}

func (m *SessionProtocolClientMachine) HandleMessage(msg MsgType) (bool, error) {
	switch {
	case m.State == SessionProtocolClientWaitAck && msg == SessionProtocolMsgPairHelloAck:
		if fn := m.Actions[SessionProtocolActionDeriveSecret]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		// received_backend_pub: recv_msg.pubkey (set by action)
		// client_shared_key: DeriveKey("client_pub", recv_msg.pubkey) (set by action)
		m.State = SessionProtocolClientE2EReady
		return true, nil
	case m.State == SessionProtocolClientE2EReady && msg == SessionProtocolMsgPairConfirm:
		// client_code: DeriveCode(received_backend_pub, "client_pub") (set by action)
		m.State = SessionProtocolClientShowCode
		return true, nil
	case m.State == SessionProtocolClientWaitPairComplete && msg == SessionProtocolMsgPairComplete:
		if fn := m.Actions[SessionProtocolActionStoreSecret]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		m.State = SessionProtocolClientPaired
		return true, nil
	case m.State == SessionProtocolClientSendAuth && msg == SessionProtocolMsgAuthOk:
		m.State = SessionProtocolClientSessionActive
		return true, nil
	case m.State == SessionProtocolClientRelayConnected && msg == SessionProtocolMsgCandidates && m.Guards[SessionProtocolGuardAltEnabled] != nil && m.Guards[SessionProtocolGuardAltEnabled]():
		if fn := m.Actions[SessionProtocolActionDialCandidate]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		m.State = SessionProtocolClientPairDialing
		return true, nil
	case m.State == SessionProtocolClientRelayConnected && msg == SessionProtocolMsgCandidates && m.Guards[SessionProtocolGuardAltDisabled] != nil && m.Guards[SessionProtocolGuardAltDisabled]():
		m.State = SessionProtocolClientRelayConnected
		return true, nil
	case m.State == SessionProtocolClientPairChecking && msg == SessionProtocolMsgPairCheckAck:
		if fn := m.Actions[SessionProtocolActionActivateLan]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		m.CActivePath = "alt"
		if m.OnChange != nil {
			m.OnChange("c_active_path")
		}
		m.CDispatcherPath = "alt"
		if m.OnChange != nil {
			m.OnChange("c_dispatcher_path")
		}
		m.AltSignal = "ready"
		if m.OnChange != nil {
			m.OnChange("alt_signal")
		}
		m.State = SessionProtocolClientAltActive
		return true, nil
	case m.State == SessionProtocolClientAltActive && msg == SessionProtocolMsgPathPing:
		m.State = SessionProtocolClientAltActive
		return true, nil
	case m.State == SessionProtocolClientAltActive && msg == SessionProtocolMsgCandidates && m.Guards[SessionProtocolGuardAltEnabled] != nil && m.Guards[SessionProtocolGuardAltEnabled]():
		if fn := m.Actions[SessionProtocolActionDialCandidate]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		m.State = SessionProtocolClientPairDialing
		return true, nil
	}
	return false, nil
}

func (m *SessionProtocolClientMachine) Step(event EventID) (bool, error) {
	switch {
	case m.State == SessionProtocolClientIdle && event == SessionProtocolEventBackchannelReceived:
		m.State = SessionProtocolClientObtainBackchannelSecret
		return true, nil
	case m.State == SessionProtocolClientObtainBackchannelSecret && event == SessionProtocolEventSecretParsed:
		m.State = SessionProtocolClientConnectRelay
		return true, nil
	case m.State == SessionProtocolClientConnectRelay && event == SessionProtocolEventRelayConnected:
		m.State = SessionProtocolClientGenKeyPair
		return true, nil
	case m.State == SessionProtocolClientGenKeyPair && event == SessionProtocolEventKeyPairGenerated:
		if fn := m.Actions[SessionProtocolActionSendPairHello]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		m.State = SessionProtocolClientWaitAck
		return true, nil
	case m.State == SessionProtocolClientShowCode && event == SessionProtocolEventCodeDisplayed:
		m.State = SessionProtocolClientWaitPairComplete
		return true, nil
	case m.State == SessionProtocolClientPaired && event == SessionProtocolEventAppLaunch:
		m.State = SessionProtocolClientReconnect
		return true, nil
	case m.State == SessionProtocolClientReconnect && event == SessionProtocolEventRelayConnected:
		m.State = SessionProtocolClientSendAuth
		return true, nil
	case m.State == SessionProtocolClientSessionActive && event == SessionProtocolEventSessionEstablished:
		m.State = SessionProtocolClientRelayConnected
		return true, nil
	case m.State == SessionProtocolClientPairDialing && event == SessionProtocolEventDialOk:
		m.State = SessionProtocolClientPairChecking
		return true, nil
	case m.State == SessionProtocolClientPairDialing && event == SessionProtocolEventDialFailed:
		m.State = SessionProtocolClientRelayConnected
		return true, nil
	case m.State == SessionProtocolClientPairChecking && event == SessionProtocolEventVerifyTimeout:
		m.CDispatcherPath = "relay"
		if m.OnChange != nil {
			m.OnChange("c_dispatcher_path")
		}
		m.State = SessionProtocolClientRelayConnected
		return true, nil
	case m.State == SessionProtocolClientAltActive && event == SessionProtocolEventAltError:
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		m.CActivePath = "relay"
		if m.OnChange != nil {
			m.OnChange("c_active_path")
		}
		m.CDispatcherPath = "relay"
		if m.OnChange != nil {
			m.OnChange("c_dispatcher_path")
		}
		m.AltSignal = "pending"
		if m.OnChange != nil {
			m.OnChange("alt_signal")
		}
		m.State = SessionProtocolClientRelayFallback
		return true, nil
	case m.State == SessionProtocolClientAltActive && event == SessionProtocolEventAltStreamError:
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		m.CActivePath = "relay"
		if m.OnChange != nil {
			m.OnChange("c_active_path")
		}
		m.CDispatcherPath = "relay"
		if m.OnChange != nil {
			m.OnChange("c_dispatcher_path")
		}
		m.AltSignal = "pending"
		if m.OnChange != nil {
			m.OnChange("alt_signal")
		}
		m.State = SessionProtocolClientRelayFallback
		return true, nil
	case m.State == SessionProtocolClientRelayFallback && event == SessionProtocolEventRelayOk:
		m.State = SessionProtocolClientRelayConnected
		return true, nil
	case m.State == SessionProtocolClientPairDialing && event == SessionProtocolEventAppForceFallback:
		m.State = SessionProtocolClientRelayConnected
		return true, nil
	case m.State == SessionProtocolClientPairChecking && event == SessionProtocolEventAppForceFallback:
		m.CDispatcherPath = "relay"
		if m.OnChange != nil {
			m.OnChange("c_dispatcher_path")
		}
		m.State = SessionProtocolClientRelayConnected
		return true, nil
	case m.State == SessionProtocolClientAltActive && event == SessionProtocolEventAppForceFallback:
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		m.CActivePath = "relay"
		if m.OnChange != nil {
			m.OnChange("c_active_path")
		}
		m.CDispatcherPath = "relay"
		if m.OnChange != nil {
			m.OnChange("c_dispatcher_path")
		}
		m.AltSignal = "pending"
		if m.OnChange != nil {
			m.OnChange("alt_signal")
		}
		m.State = SessionProtocolClientRelayConnected
		return true, nil
	case m.State == SessionProtocolClientRelayConnected && event == SessionProtocolEventDisconnect:
		m.State = SessionProtocolClientPaired
		return true, nil
	case m.State == SessionProtocolClientRelayConnected && event == SessionProtocolEventAppSend:
		m.State = SessionProtocolClientRelayConnected
		return true, nil
	case m.State == SessionProtocolClientPairDialing && event == SessionProtocolEventAppSend:
		m.State = SessionProtocolClientPairDialing
		return true, nil
	case m.State == SessionProtocolClientPairChecking && event == SessionProtocolEventAppSend:
		m.State = SessionProtocolClientPairChecking
		return true, nil
	case m.State == SessionProtocolClientAltActive && event == SessionProtocolEventAppSend:
		m.State = SessionProtocolClientAltActive
		return true, nil
	case m.State == SessionProtocolClientRelayFallback && event == SessionProtocolEventAppSend:
		m.State = SessionProtocolClientRelayFallback
		return true, nil
	case m.State == SessionProtocolClientRelayConnected && event == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolClientRelayConnected
		return true, nil
	case m.State == SessionProtocolClientPairDialing && event == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolClientPairDialing
		return true, nil
	case m.State == SessionProtocolClientPairChecking && event == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolClientPairChecking
		return true, nil
	case m.State == SessionProtocolClientAltActive && event == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolClientAltActive
		return true, nil
	case m.State == SessionProtocolClientRelayFallback && event == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolClientRelayFallback
		return true, nil
	case m.State == SessionProtocolClientRelayConnected && event == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolClientRelayConnected
		return true, nil
	case m.State == SessionProtocolClientPairDialing && event == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolClientPairDialing
		return true, nil
	case m.State == SessionProtocolClientPairChecking && event == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolClientPairChecking
		return true, nil
	case m.State == SessionProtocolClientAltActive && event == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolClientAltActive
		return true, nil
	case m.State == SessionProtocolClientRelayFallback && event == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolClientRelayFallback
		return true, nil
	case m.State == SessionProtocolClientRelayConnected && event == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolClientRelayConnected
		return true, nil
	case m.State == SessionProtocolClientPairDialing && event == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolClientPairDialing
		return true, nil
	case m.State == SessionProtocolClientPairChecking && event == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolClientPairChecking
		return true, nil
	case m.State == SessionProtocolClientAltActive && event == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolClientAltActive
		return true, nil
	case m.State == SessionProtocolClientRelayFallback && event == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolClientRelayFallback
		return true, nil
	case m.State == SessionProtocolClientRelayConnected && event == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolClientRelayConnected
		return true, nil
	case m.State == SessionProtocolClientPairDialing && event == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolClientPairDialing
		return true, nil
	case m.State == SessionProtocolClientPairChecking && event == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolClientPairChecking
		return true, nil
	case m.State == SessionProtocolClientAltActive && event == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolClientAltActive
		return true, nil
	case m.State == SessionProtocolClientRelayFallback && event == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolClientRelayFallback
		return true, nil
	case m.State == SessionProtocolClientAltActive && event == SessionProtocolEventAltStreamData:
		m.State = SessionProtocolClientAltActive
		return true, nil
	case m.State == SessionProtocolClientAltActive && event == SessionProtocolEventAltDatagram:
		m.State = SessionProtocolClientAltActive
		return true, nil
	}
	return false, nil
}

func (m *SessionProtocolClientMachine) HandleEvent(ev EventID) ([]CmdID, error) {
	switch {
	case m.State == SessionProtocolClientIdle && ev == SessionProtocolEventBackchannelReceived:
		m.State = SessionProtocolClientObtainBackchannelSecret
		return nil, nil
	case m.State == SessionProtocolClientObtainBackchannelSecret && ev == SessionProtocolEventSecretParsed:
		m.State = SessionProtocolClientConnectRelay
		return nil, nil
	case m.State == SessionProtocolClientConnectRelay && ev == SessionProtocolEventRelayConnected:
		m.State = SessionProtocolClientGenKeyPair
		return nil, nil
	case m.State == SessionProtocolClientGenKeyPair && ev == SessionProtocolEventKeyPairGenerated:
		if fn := m.Actions[SessionProtocolActionSendPairHello]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		m.State = SessionProtocolClientWaitAck
		return nil, nil
	case m.State == SessionProtocolClientWaitAck && ev == SessionProtocolEventRecvPairHelloAck:
		if fn := m.Actions[SessionProtocolActionDeriveSecret]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		// received_backend_pub: recv_msg.pubkey (set by action)
		// client_shared_key: DeriveKey("client_pub", recv_msg.pubkey) (set by action)
		m.State = SessionProtocolClientE2EReady
		return nil, nil
	case m.State == SessionProtocolClientE2EReady && ev == SessionProtocolEventRecvPairConfirm:
		// client_code: DeriveCode(received_backend_pub, "client_pub") (set by action)
		m.State = SessionProtocolClientShowCode
		return nil, nil
	case m.State == SessionProtocolClientShowCode && ev == SessionProtocolEventCodeDisplayed:
		m.State = SessionProtocolClientWaitPairComplete
		return nil, nil
	case m.State == SessionProtocolClientWaitPairComplete && ev == SessionProtocolEventRecvPairComplete:
		if fn := m.Actions[SessionProtocolActionStoreSecret]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		m.State = SessionProtocolClientPaired
		return nil, nil
	case m.State == SessionProtocolClientPaired && ev == SessionProtocolEventAppLaunch:
		m.State = SessionProtocolClientReconnect
		return nil, nil
	case m.State == SessionProtocolClientReconnect && ev == SessionProtocolEventRelayConnected:
		m.State = SessionProtocolClientSendAuth
		return nil, nil
	case m.State == SessionProtocolClientSendAuth && ev == SessionProtocolEventRecvAuthOk:
		m.State = SessionProtocolClientSessionActive
		return nil, nil
	case m.State == SessionProtocolClientSessionActive && ev == SessionProtocolEventSessionEstablished:
		m.State = SessionProtocolClientRelayConnected
		return nil, nil
	case m.State == SessionProtocolClientRelayConnected && ev == SessionProtocolEventRecvCandidates && m.Guards[SessionProtocolGuardAltEnabled] != nil && m.Guards[SessionProtocolGuardAltEnabled]():
		if fn := m.Actions[SessionProtocolActionDialCandidate]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		m.State = SessionProtocolClientPairDialing
		return []CmdID{SessionProtocolCmdDialCandidate}, nil
	case m.State == SessionProtocolClientRelayConnected && ev == SessionProtocolEventRecvCandidates && m.Guards[SessionProtocolGuardAltDisabled] != nil && m.Guards[SessionProtocolGuardAltDisabled]():
		m.State = SessionProtocolClientRelayConnected
		return nil, nil
	case m.State == SessionProtocolClientPairDialing && ev == SessionProtocolEventDialOk:
		m.State = SessionProtocolClientPairChecking
		return []CmdID{SessionProtocolCmdSendPairCheck}, nil
	case m.State == SessionProtocolClientPairDialing && ev == SessionProtocolEventDialFailed:
		m.State = SessionProtocolClientRelayConnected
		return nil, nil
	case m.State == SessionProtocolClientPairChecking && ev == SessionProtocolEventRecvPairCheckAck:
		if fn := m.Actions[SessionProtocolActionActivateLan]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		m.CActivePath = "alt"
		if m.OnChange != nil {
			m.OnChange("c_active_path")
		}
		m.CDispatcherPath = "alt"
		if m.OnChange != nil {
			m.OnChange("c_dispatcher_path")
		}
		m.AltSignal = "ready"
		if m.OnChange != nil {
			m.OnChange("alt_signal")
		}
		m.State = SessionProtocolClientAltActive
		return []CmdID{SessionProtocolCmdStartAltStreamReader, SessionProtocolCmdStartAltDgReader, SessionProtocolCmdSignalAltReady, SessionProtocolCmdSetCryptoDatagram}, nil
	case m.State == SessionProtocolClientPairChecking && ev == SessionProtocolEventVerifyTimeout:
		m.CDispatcherPath = "relay"
		if m.OnChange != nil {
			m.OnChange("c_dispatcher_path")
		}
		m.State = SessionProtocolClientRelayConnected
		return nil, nil
	case m.State == SessionProtocolClientAltActive && ev == SessionProtocolEventRecvPathPing:
		m.State = SessionProtocolClientAltActive
		return []CmdID{SessionProtocolCmdSendPathPong}, nil
	case m.State == SessionProtocolClientAltActive && ev == SessionProtocolEventAltError:
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		m.CActivePath = "relay"
		if m.OnChange != nil {
			m.OnChange("c_active_path")
		}
		m.CDispatcherPath = "relay"
		if m.OnChange != nil {
			m.OnChange("c_dispatcher_path")
		}
		m.AltSignal = "pending"
		if m.OnChange != nil {
			m.OnChange("alt_signal")
		}
		m.State = SessionProtocolClientRelayFallback
		return []CmdID{SessionProtocolCmdStopAltStreamReader, SessionProtocolCmdStopAltDgReader, SessionProtocolCmdCloseAltPath, SessionProtocolCmdResetAltReady}, nil
	case m.State == SessionProtocolClientAltActive && ev == SessionProtocolEventAltStreamError:
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		m.CActivePath = "relay"
		if m.OnChange != nil {
			m.OnChange("c_active_path")
		}
		m.CDispatcherPath = "relay"
		if m.OnChange != nil {
			m.OnChange("c_dispatcher_path")
		}
		m.AltSignal = "pending"
		if m.OnChange != nil {
			m.OnChange("alt_signal")
		}
		m.State = SessionProtocolClientRelayFallback
		return []CmdID{SessionProtocolCmdStopAltStreamReader, SessionProtocolCmdStopAltDgReader, SessionProtocolCmdCloseAltPath, SessionProtocolCmdResetAltReady}, nil
	case m.State == SessionProtocolClientRelayFallback && ev == SessionProtocolEventRelayOk:
		m.State = SessionProtocolClientRelayConnected
		return nil, nil
	case m.State == SessionProtocolClientAltActive && ev == SessionProtocolEventRecvCandidates && m.Guards[SessionProtocolGuardAltEnabled] != nil && m.Guards[SessionProtocolGuardAltEnabled]():
		if fn := m.Actions[SessionProtocolActionDialCandidate]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		m.State = SessionProtocolClientPairDialing
		return []CmdID{SessionProtocolCmdStopAltStreamReader, SessionProtocolCmdStopAltDgReader, SessionProtocolCmdCloseAltPath, SessionProtocolCmdDialCandidate}, nil
	case m.State == SessionProtocolClientPairDialing && ev == SessionProtocolEventAppForceFallback:
		m.State = SessionProtocolClientRelayConnected
		return nil, nil
	case m.State == SessionProtocolClientPairChecking && ev == SessionProtocolEventAppForceFallback:
		m.CDispatcherPath = "relay"
		if m.OnChange != nil {
			m.OnChange("c_dispatcher_path")
		}
		m.State = SessionProtocolClientRelayConnected
		return []CmdID{SessionProtocolCmdStopAltStreamReader, SessionProtocolCmdStopAltDgReader, SessionProtocolCmdCloseAltPath}, nil
	case m.State == SessionProtocolClientAltActive && ev == SessionProtocolEventAppForceFallback:
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		m.CActivePath = "relay"
		if m.OnChange != nil {
			m.OnChange("c_active_path")
		}
		m.CDispatcherPath = "relay"
		if m.OnChange != nil {
			m.OnChange("c_dispatcher_path")
		}
		m.AltSignal = "pending"
		if m.OnChange != nil {
			m.OnChange("alt_signal")
		}
		m.State = SessionProtocolClientRelayConnected
		return []CmdID{SessionProtocolCmdStopAltStreamReader, SessionProtocolCmdStopAltDgReader, SessionProtocolCmdCloseAltPath, SessionProtocolCmdResetAltReady}, nil
	case m.State == SessionProtocolClientRelayConnected && ev == SessionProtocolEventDisconnect:
		m.State = SessionProtocolClientPaired
		return nil, nil
	case m.State == SessionProtocolClientRelayConnected && ev == SessionProtocolEventAppSend:
		m.State = SessionProtocolClientRelayConnected
		return []CmdID{SessionProtocolCmdWriteActiveStream}, nil
	case m.State == SessionProtocolClientPairDialing && ev == SessionProtocolEventAppSend:
		m.State = SessionProtocolClientPairDialing
		return []CmdID{SessionProtocolCmdWriteActiveStream}, nil
	case m.State == SessionProtocolClientPairChecking && ev == SessionProtocolEventAppSend:
		m.State = SessionProtocolClientPairChecking
		return []CmdID{SessionProtocolCmdWriteActiveStream}, nil
	case m.State == SessionProtocolClientAltActive && ev == SessionProtocolEventAppSend:
		m.State = SessionProtocolClientAltActive
		return []CmdID{SessionProtocolCmdWriteActiveStream}, nil
	case m.State == SessionProtocolClientRelayFallback && ev == SessionProtocolEventAppSend:
		m.State = SessionProtocolClientRelayFallback
		return []CmdID{SessionProtocolCmdWriteActiveStream}, nil
	case m.State == SessionProtocolClientRelayConnected && ev == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolClientRelayConnected
		return []CmdID{SessionProtocolCmdDeliverRecv}, nil
	case m.State == SessionProtocolClientPairDialing && ev == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolClientPairDialing
		return []CmdID{SessionProtocolCmdDeliverRecv}, nil
	case m.State == SessionProtocolClientPairChecking && ev == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolClientPairChecking
		return []CmdID{SessionProtocolCmdDeliverRecv}, nil
	case m.State == SessionProtocolClientAltActive && ev == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolClientAltActive
		return []CmdID{SessionProtocolCmdDeliverRecv}, nil
	case m.State == SessionProtocolClientRelayFallback && ev == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolClientRelayFallback
		return []CmdID{SessionProtocolCmdDeliverRecv}, nil
	case m.State == SessionProtocolClientRelayConnected && ev == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolClientRelayConnected
		return []CmdID{SessionProtocolCmdDeliverRecvError}, nil
	case m.State == SessionProtocolClientPairDialing && ev == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolClientPairDialing
		return []CmdID{SessionProtocolCmdDeliverRecvError}, nil
	case m.State == SessionProtocolClientPairChecking && ev == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolClientPairChecking
		return []CmdID{SessionProtocolCmdDeliverRecvError}, nil
	case m.State == SessionProtocolClientAltActive && ev == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolClientAltActive
		return []CmdID{SessionProtocolCmdDeliverRecvError}, nil
	case m.State == SessionProtocolClientRelayFallback && ev == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolClientRelayFallback
		return []CmdID{SessionProtocolCmdDeliverRecvError}, nil
	case m.State == SessionProtocolClientRelayConnected && ev == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolClientRelayConnected
		return []CmdID{SessionProtocolCmdSendActiveDatagram}, nil
	case m.State == SessionProtocolClientPairDialing && ev == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolClientPairDialing
		return []CmdID{SessionProtocolCmdSendActiveDatagram}, nil
	case m.State == SessionProtocolClientPairChecking && ev == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolClientPairChecking
		return []CmdID{SessionProtocolCmdSendActiveDatagram}, nil
	case m.State == SessionProtocolClientAltActive && ev == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolClientAltActive
		return []CmdID{SessionProtocolCmdSendActiveDatagram}, nil
	case m.State == SessionProtocolClientRelayFallback && ev == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolClientRelayFallback
		return []CmdID{SessionProtocolCmdSendActiveDatagram}, nil
	case m.State == SessionProtocolClientRelayConnected && ev == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolClientRelayConnected
		return []CmdID{SessionProtocolCmdDeliverRecvDatagram}, nil
	case m.State == SessionProtocolClientPairDialing && ev == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolClientPairDialing
		return []CmdID{SessionProtocolCmdDeliverRecvDatagram}, nil
	case m.State == SessionProtocolClientPairChecking && ev == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolClientPairChecking
		return []CmdID{SessionProtocolCmdDeliverRecvDatagram}, nil
	case m.State == SessionProtocolClientAltActive && ev == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolClientAltActive
		return []CmdID{SessionProtocolCmdDeliverRecvDatagram}, nil
	case m.State == SessionProtocolClientRelayFallback && ev == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolClientRelayFallback
		return []CmdID{SessionProtocolCmdDeliverRecvDatagram}, nil
	case m.State == SessionProtocolClientAltActive && ev == SessionProtocolEventAltStreamData:
		m.State = SessionProtocolClientAltActive
		return []CmdID{SessionProtocolCmdDeliverRecv}, nil
	case m.State == SessionProtocolClientAltActive && ev == SessionProtocolEventAltDatagram:
		m.State = SessionProtocolClientAltActive
		return []CmdID{SessionProtocolCmdDeliverRecvDatagram}, nil
	}
	return nil, nil
}

// SessionProtocolRelayMachine is the generated state machine for the relay actor.
type SessionProtocolRelayMachine struct {
	State       State
	RelayBridge string // relay bridge state

	Guards   map[GuardID]func() bool
	Actions  map[ActionID]func() error
	OnChange func(varName string)
}

func NewSessionProtocolRelayMachine() *SessionProtocolRelayMachine {
	return &SessionProtocolRelayMachine{
		State:       SessionProtocolRelayIdle,
		RelayBridge: "idle",
		Guards:      make(map[GuardID]func() bool),
		Actions:     make(map[ActionID]func() error),
	}
}

func (m *SessionProtocolRelayMachine) HandleMessage(msg MsgType) (bool, error) {
	switch {
	}
	return false, nil
}

func (m *SessionProtocolRelayMachine) Step(event EventID) (bool, error) {
	switch {
	case m.State == SessionProtocolRelayIdle && event == SessionProtocolEventBackendRegister:
		m.State = SessionProtocolRelayBackendRegistered
		return true, nil
	case m.State == SessionProtocolRelayBackendRegistered && event == SessionProtocolEventClientConnect:
		if fn := m.Actions[SessionProtocolActionBridgeStreams]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		m.RelayBridge = "active"
		if m.OnChange != nil {
			m.OnChange("relay_bridge")
		}
		m.State = SessionProtocolRelayBridged
		return true, nil
	case m.State == SessionProtocolRelayBridged && event == SessionProtocolEventClientDisconnect:
		if fn := m.Actions[SessionProtocolActionUnbridge]; fn != nil {
			if err := fn(); err != nil {
				return false, err
			}
		}
		m.RelayBridge = "idle"
		if m.OnChange != nil {
			m.OnChange("relay_bridge")
		}
		m.State = SessionProtocolRelayBackendRegistered
		return true, nil
	case m.State == SessionProtocolRelayBackendRegistered && event == SessionProtocolEventBackendDisconnect:
		m.State = SessionProtocolRelayIdle
		return true, nil
	}
	return false, nil
}

func (m *SessionProtocolRelayMachine) HandleEvent(ev EventID) ([]CmdID, error) {
	switch {
	case m.State == SessionProtocolRelayIdle && ev == SessionProtocolEventBackendRegister:
		m.State = SessionProtocolRelayBackendRegistered
		return nil, nil
	case m.State == SessionProtocolRelayBackendRegistered && ev == SessionProtocolEventClientConnect:
		if fn := m.Actions[SessionProtocolActionBridgeStreams]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		m.RelayBridge = "active"
		if m.OnChange != nil {
			m.OnChange("relay_bridge")
		}
		m.State = SessionProtocolRelayBridged
		return nil, nil
	case m.State == SessionProtocolRelayBridged && ev == SessionProtocolEventClientDisconnect:
		if fn := m.Actions[SessionProtocolActionUnbridge]; fn != nil {
			if err := fn(); err != nil {
				return nil, err
			}
		}
		m.RelayBridge = "idle"
		if m.OnChange != nil {
			m.OnChange("relay_bridge")
		}
		m.State = SessionProtocolRelayBackendRegistered
		return nil, nil
	case m.State == SessionProtocolRelayBackendRegistered && ev == SessionProtocolEventBackendDisconnect:
		m.State = SessionProtocolRelayIdle
		return nil, nil
	}
	return nil, nil
}
