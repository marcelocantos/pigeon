// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Code generated from protocol/*.yaml. DO NOT EDIT.

package pigeon

import (
	"github.com/marcelocantos/pigeon/protocol"
	"github.com/arr-ai/frozen"
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
	Recv     = protocol.Recv
	Internal = protocol.Internal
	Invariant = protocol.Invariant
	Liveness  = protocol.Liveness
)

var _ frozen.Set[string] // suppress unused import

// SessionProtocol backend states.
const (
	SessionProtocolBackendIdle State = "Idle"
	SessionProtocolBackendGenerateToken State = "GenerateToken"
	SessionProtocolBackendRegisterRelay State = "RegisterRelay"
	SessionProtocolBackendWaitingForClient State = "WaitingForClient"
	SessionProtocolBackendDeriveSecret State = "DeriveSecret"
	SessionProtocolBackendSendAck State = "SendAck"
	SessionProtocolBackendWaitingForCode State = "WaitingForCode"
	SessionProtocolBackendValidateCode State = "ValidateCode"
	SessionProtocolBackendStorePaired State = "StorePaired"
	SessionProtocolBackendPaired State = "Paired"
	SessionProtocolBackendAuthCheck State = "AuthCheck"
	SessionProtocolBackendSessionActive State = "SessionActive"
	SessionProtocolBackendRelayConnected State = "RelayConnected"
	SessionProtocolBackendLANOffered State = "LANOffered"
	SessionProtocolBackendLANActive State = "LANActive"
	SessionProtocolBackendRelayBackoff State = "RelayBackoff"
	SessionProtocolBackendLANDegraded State = "LANDegraded"
)

// SessionProtocol client states.
const (
	SessionProtocolClientIdle State = "Idle"
	SessionProtocolClientObtainBackchannelSecret State = "ObtainBackchannelSecret"
	SessionProtocolClientConnectRelay State = "ConnectRelay"
	SessionProtocolClientGenKeyPair State = "GenKeyPair"
	SessionProtocolClientWaitAck State = "WaitAck"
	SessionProtocolClientE2EReady State = "E2EReady"
	SessionProtocolClientShowCode State = "ShowCode"
	SessionProtocolClientWaitPairComplete State = "WaitPairComplete"
	SessionProtocolClientPaired State = "Paired"
	SessionProtocolClientReconnect State = "Reconnect"
	SessionProtocolClientSendAuth State = "SendAuth"
	SessionProtocolClientSessionActive State = "SessionActive"
	SessionProtocolClientRelayConnected State = "RelayConnected"
	SessionProtocolClientLANConnecting State = "LANConnecting"
	SessionProtocolClientLANVerifying State = "LANVerifying"
	SessionProtocolClientLANActive State = "LANActive"
	SessionProtocolClientRelayFallback State = "RelayFallback"
)

// SessionProtocol relay states.
const (
	SessionProtocolRelayIdle State = "Idle"
	SessionProtocolRelayBackendRegistered State = "BackendRegistered"
	SessionProtocolRelayBridged State = "Bridged"
)

// SessionProtocol message types.
const (
	SessionProtocolMsgPairHello MsgType = "pair_hello"
	SessionProtocolMsgPairHelloAck MsgType = "pair_hello_ack"
	SessionProtocolMsgPairConfirm MsgType = "pair_confirm"
	SessionProtocolMsgPairComplete MsgType = "pair_complete"
	SessionProtocolMsgAuthRequest MsgType = "auth_request"
	SessionProtocolMsgAuthOk MsgType = "auth_ok"
	SessionProtocolMsgLanOffer MsgType = "lan_offer"
	SessionProtocolMsgLanVerify MsgType = "lan_verify"
	SessionProtocolMsgLanConfirm MsgType = "lan_confirm"
	SessionProtocolMsgPathPing MsgType = "path_ping"
	SessionProtocolMsgPathPong MsgType = "path_pong"
)

// SessionProtocol guards.
const (
	SessionProtocolGuardTokenValid GuardID = "token_valid"
	SessionProtocolGuardTokenInvalid GuardID = "token_invalid"
	SessionProtocolGuardCodeCorrect GuardID = "code_correct"
	SessionProtocolGuardCodeWrong GuardID = "code_wrong"
	SessionProtocolGuardDeviceKnown GuardID = "device_known"
	SessionProtocolGuardDeviceUnknown GuardID = "device_unknown"
	SessionProtocolGuardNonceFresh GuardID = "nonce_fresh"
	SessionProtocolGuardChallengeValid GuardID = "challenge_valid"
	SessionProtocolGuardChallengeInvalid GuardID = "challenge_invalid"
	SessionProtocolGuardLanEnabled GuardID = "lan_enabled"
	SessionProtocolGuardLanDisabled GuardID = "lan_disabled"
	SessionProtocolGuardLanServerAvailable GuardID = "lan_server_available"
	SessionProtocolGuardUnderMaxFailures GuardID = "under_max_failures"
	SessionProtocolGuardAtMaxFailures GuardID = "at_max_failures"
)

// SessionProtocol actions.
const (
	SessionProtocolActionActivateLan ActionID = "activate_lan"
	SessionProtocolActionBridgeStreams ActionID = "bridge_streams"
	SessionProtocolActionDeriveSecret ActionID = "derive_secret"
	SessionProtocolActionDialLan ActionID = "dial_lan"
	SessionProtocolActionFallbackToRelay ActionID = "fallback_to_relay"
	SessionProtocolActionGenerateToken ActionID = "generate_token"
	SessionProtocolActionRegisterRelay ActionID = "register_relay"
	SessionProtocolActionResetFailures ActionID = "reset_failures"
	SessionProtocolActionSendPairHello ActionID = "send_pair_hello"
	SessionProtocolActionStoreDevice ActionID = "store_device"
	SessionProtocolActionStoreSecret ActionID = "store_secret"
	SessionProtocolActionUnbridge ActionID = "unbridge"
	SessionProtocolActionVerifyDevice ActionID = "verify_device"
)

// SessionProtocol events.
const (
	SessionProtocolEventAppClose EventID = "app_close"
	SessionProtocolEventAppForceFallback EventID = "app_force_fallback"
	SessionProtocolEventAppLaunch EventID = "app_launch"
	SessionProtocolEventAppRecv EventID = "app_recv"
	SessionProtocolEventAppRecvDatagram EventID = "app_recv_datagram"
	SessionProtocolEventAppSend EventID = "app_send"
	SessionProtocolEventAppSendDatagram EventID = "app_send_datagram"
	SessionProtocolEventBackchannelReceived EventID = "backchannel_received"
	SessionProtocolEventBackendDisconnect EventID = "backend_disconnect"
	SessionProtocolEventBackendRegister EventID = "backend_register"
	SessionProtocolEventBackoffExpired EventID = "backoff_expired"
	SessionProtocolEventCheckCode EventID = "check_code"
	SessionProtocolEventCliCodeEntered EventID = "cli_code_entered"
	SessionProtocolEventCliInitPair EventID = "cli_init_pair"
	SessionProtocolEventClientConnect EventID = "client_connect"
	SessionProtocolEventClientDisconnect EventID = "client_disconnect"
	SessionProtocolEventCodeDisplayed EventID = "code_displayed"
	SessionProtocolEventDisconnect EventID = "disconnect"
	SessionProtocolEventEcdhComplete EventID = "ecdh_complete"
	SessionProtocolEventFinalise EventID = "finalise"
	SessionProtocolEventKeyPairGenerated EventID = "key_pair_generated"
	SessionProtocolEventLanDatagram EventID = "lan_datagram"
	SessionProtocolEventLanDialFailed EventID = "lan_dial_failed"
	SessionProtocolEventLanDialOk EventID = "lan_dial_ok"
	SessionProtocolEventLanError EventID = "lan_error"
	SessionProtocolEventLanServerChanged EventID = "lan_server_changed"
	SessionProtocolEventLanServerReady EventID = "lan_server_ready"
	SessionProtocolEventLanStreamData EventID = "lan_stream_data"
	SessionProtocolEventLanStreamError EventID = "lan_stream_error"
	SessionProtocolEventLanVerifyOk EventID = "lan_verify_ok"
	SessionProtocolEventOfferTimeout EventID = "offer_timeout"
	SessionProtocolEventPingTick EventID = "ping_tick"
	SessionProtocolEventPingTimeout EventID = "ping_timeout"
	SessionProtocolEventReadvertiseTick EventID = "readvertise_tick"
	SessionProtocolEventRecvAuthOk EventID = "recv_auth_ok"
	SessionProtocolEventRecvAuthRequest EventID = "recv_auth_request"
	SessionProtocolEventRecvLanConfirm EventID = "recv_lan_confirm"
	SessionProtocolEventRecvLanOffer EventID = "recv_lan_offer"
	SessionProtocolEventRecvLanVerify EventID = "recv_lan_verify"
	SessionProtocolEventRecvPairComplete EventID = "recv_pair_complete"
	SessionProtocolEventRecvPairConfirm EventID = "recv_pair_confirm"
	SessionProtocolEventRecvPairHello EventID = "recv_pair_hello"
	SessionProtocolEventRecvPairHelloAck EventID = "recv_pair_hello_ack"
	SessionProtocolEventRecvPathPing EventID = "recv_path_ping"
	SessionProtocolEventRecvPathPong EventID = "recv_path_pong"
	SessionProtocolEventRelayConnected EventID = "relay_connected"
	SessionProtocolEventRelayDatagram EventID = "relay_datagram"
	SessionProtocolEventRelayOk EventID = "relay_ok"
	SessionProtocolEventRelayRegistered EventID = "relay_registered"
	SessionProtocolEventRelayStreamData EventID = "relay_stream_data"
	SessionProtocolEventRelayStreamError EventID = "relay_stream_error"
	SessionProtocolEventSecretParsed EventID = "secret_parsed"
	SessionProtocolEventSessionEstablished EventID = "session_established"
	SessionProtocolEventSignalCodeDisplay EventID = "signal_code_display"
	SessionProtocolEventTokenCreated EventID = "token_created"
	SessionProtocolEventVerify EventID = "verify"
	SessionProtocolEventVerifyTimeout EventID = "verify_timeout"
)

// SessionProtocol commands.
const (
	SessionProtocolCmdWriteActiveStream CmdID = "write_active_stream"
	SessionProtocolCmdSendActiveDatagram CmdID = "send_active_datagram"
	SessionProtocolCmdSendPathPing CmdID = "send_path_ping"
	SessionProtocolCmdSendPathPong CmdID = "send_path_pong"
	SessionProtocolCmdSendLanOffer CmdID = "send_lan_offer"
	SessionProtocolCmdSendLanVerify CmdID = "send_lan_verify"
	SessionProtocolCmdSendLanConfirm CmdID = "send_lan_confirm"
	SessionProtocolCmdDialLan CmdID = "dial_lan"
	SessionProtocolCmdDeliverRecv CmdID = "deliver_recv"
	SessionProtocolCmdDeliverRecvError CmdID = "deliver_recv_error"
	SessionProtocolCmdDeliverRecvDatagram CmdID = "deliver_recv_datagram"
	SessionProtocolCmdStartLanStreamReader CmdID = "start_lan_stream_reader"
	SessionProtocolCmdStopLanStreamReader CmdID = "stop_lan_stream_reader"
	SessionProtocolCmdStartLanDgReader CmdID = "start_lan_dg_reader"
	SessionProtocolCmdStopLanDgReader CmdID = "stop_lan_dg_reader"
	SessionProtocolCmdStartMonitor CmdID = "start_monitor"
	SessionProtocolCmdStopMonitor CmdID = "stop_monitor"
	SessionProtocolCmdStartPongTimeout CmdID = "start_pong_timeout"
	SessionProtocolCmdCancelPongTimeout CmdID = "cancel_pong_timeout"
	SessionProtocolCmdStartBackoffTimer CmdID = "start_backoff_timer"
	SessionProtocolCmdCloseLanPath CmdID = "close_lan_path"
	SessionProtocolCmdSignalLanReady CmdID = "signal_lan_ready"
	SessionProtocolCmdResetLanReady CmdID = "reset_lan_ready"
	SessionProtocolCmdSetCryptoDatagram CmdID = "set_crypto_datagram"
)

// Wire constants — protocol-level values shared across all platforms.
// DatagramFraming
const (
	DgConnWhole byte = 0x00 // conn-level single-frame datagram
	DgPing byte = 0x10 // health ping on direct path
	DgPong byte = 0x11 // health pong on direct path
	DgConnFragment byte = 0x40 // conn-level multi-frame datagram
	DgChanWhole byte = 0x80 // channel single-frame datagram
	DgChanFragment byte = 0xC0 // channel multi-frame datagram
	FragHeaderSize = 8 // fragment header: msgID(4) + fragIdx(2) + totalFrags(2)
	ChanIdSize = 2 // channel ID prefix size in bytes
)

// DatagramLimits
const (
	MaxDatagramPayload = 1200 // max payload per QUIC datagram (bytes)
	FragmentTimeoutMs = 5000 // ms // fragment reassembly timeout
)

// MessageFraming
const (
	FrameApp byte = 0x00 // application data
	FrameLanOffer byte = 0x01 // LAN address exchange
	FrameCutover byte = 0x02 // transport cutover marker
	MaxMessageSize = 1048576 // max stream message size (1 MiB)
	LengthPrefixSize = 4 // big-endian length prefix size
)

// Health
const (
	PingIntervalMs = 5000 // ms // health ping interval
	PongTimeoutMs = 4000 // ms // pong reply timeout
	MaxPingFailures = 3 // consecutive failures before fallback
	MaxBackoffLevel = 5 // exponential backoff cap
)

// ChannelKeys
const (
	StreamChannelOpenerSuffix = ":o2a" // HKDF info suffix for opener→acceptor stream key
	StreamChannelAcceptSuffix = ":a2o" // HKDF info suffix for acceptor→opener stream key
	DgChannelSendSuffix = ":dg:send" // HKDF info suffix for datagram send key
	DgChannelRecvSuffix = ":dg:recv" // HKDF info suffix for datagram recv key
	ChannelIdHashMultiplier = 31 // hash multiplier for channel name → uint16 ID
)

func SessionProtocol() *Protocol {
	return &Protocol{
		Name: "Session",
		Actors: []Actor{
			{Name: "backend", Initial: "Idle", Transitions: []Transition{
				{From: "Idle", To: "GenerateToken", On: Internal("cli_init_pair"), Do: "generate_token", Updates: []VarUpdate{{Var: "current_token", Expr: "\"tok_1\""}, {Var: "active_tokens", Expr: "active_tokens \\union {\"tok_1\"}"}, }},
				{From: "GenerateToken", To: "RegisterRelay", On: Internal("token_created"), Do: "register_relay"},
				{From: "RegisterRelay", To: "WaitingForClient", On: Internal("relay_registered"), Updates: []VarUpdate{{Var: "secret_published", Expr: "TRUE"}, }},
				{From: "WaitingForClient", To: "DeriveSecret", On: Recv("pair_hello"), Guard: "token_valid", Do: "derive_secret", Updates: []VarUpdate{{Var: "received_client_pub", Expr: "recv_msg.pubkey"}, {Var: "backend_ecdh_pub", Expr: "\"backend_pub\""}, {Var: "backend_shared_key", Expr: "DeriveKey(\"backend_pub\", recv_msg.pubkey)"}, {Var: "backend_code", Expr: "DeriveCode(\"backend_pub\", recv_msg.pubkey)"}, }},
				{From: "WaitingForClient", To: "Idle", On: Recv("pair_hello"), Guard: "token_invalid"},
				{From: "DeriveSecret", To: "SendAck", On: Internal("ecdh_complete"), Sends: []Send{{To: "client", Msg: "pair_hello_ack", Fields: map[string]string{"pubkey": "backend_ecdh_pub", }}, }},
				{From: "SendAck", To: "WaitingForCode", On: Internal("signal_code_display"), Sends: []Send{{To: "client", Msg: "pair_confirm"}, }},
				{From: "WaitingForCode", To: "ValidateCode", On: Internal("cli_code_entered"), Updates: []VarUpdate{{Var: "received_code", Expr: "cli_entered_code"}, }},
				{From: "ValidateCode", To: "StorePaired", On: Internal("check_code"), Guard: "code_correct"},
				{From: "ValidateCode", To: "Idle", On: Internal("check_code"), Guard: "code_wrong", Updates: []VarUpdate{{Var: "code_attempts", Expr: "code_attempts + 1"}, }},
				{From: "StorePaired", To: "Paired", On: Internal("finalise"), Do: "store_device", Sends: []Send{{To: "client", Msg: "pair_complete", Fields: map[string]string{"key": "backend_shared_key", "secret": "\"dev_secret_1\"", }}, }, Updates: []VarUpdate{{Var: "device_secret", Expr: "\"dev_secret_1\""}, {Var: "paired_devices", Expr: "paired_devices \\union {\"device_1\"}"}, {Var: "active_tokens", Expr: "active_tokens \\ {current_token}"}, {Var: "used_tokens", Expr: "used_tokens \\union {current_token}"}, }},
				{From: "Paired", To: "AuthCheck", On: Recv("auth_request"), Updates: []VarUpdate{{Var: "received_device_id", Expr: "recv_msg.device_id"}, {Var: "received_auth_nonce", Expr: "recv_msg.nonce"}, }},
				{From: "AuthCheck", To: "SessionActive", On: Internal("verify"), Guard: "device_known", Do: "verify_device", Sends: []Send{{To: "client", Msg: "auth_ok"}, }, Updates: []VarUpdate{{Var: "auth_nonces_used", Expr: "auth_nonces_used \\union {received_auth_nonce}"}, }},
				{From: "AuthCheck", To: "Idle", On: Internal("verify"), Guard: "device_unknown"},
				{From: "SessionActive", To: "RelayConnected", On: Internal("session_established")},
				{From: "RelayConnected", To: "LANOffered", On: Internal("lan_server_ready"), Sends: []Send{{To: "client", Msg: "lan_offer", Fields: map[string]string{"addr": "lan_addr", "challenge": "challenge_bytes", }}, }},
				{From: "LANOffered", To: "LANActive", On: Recv("lan_verify"), Guard: "challenge_valid", Do: "activate_lan", Sends: []Send{{To: "client", Msg: "lan_confirm"}, }, Updates: []VarUpdate{{Var: "ping_failures", Expr: "0"}, {Var: "backoff_level", Expr: "0"}, {Var: "b_active_path", Expr: "\"lan\""}, {Var: "b_dispatcher_path", Expr: "\"lan\""}, {Var: "monitor_target", Expr: "\"lan\""}, {Var: "lan_signal", Expr: "\"ready\""}, }},
				{From: "LANOffered", To: "RelayConnected", On: Recv("lan_verify"), Guard: "challenge_invalid"},
				{From: "LANOffered", To: "RelayBackoff", On: Internal("offer_timeout"), Updates: []VarUpdate{{Var: "backoff_level", Expr: "Min(backoff_level + 1, max_backoff_level)"}, {Var: "lan_signal", Expr: "\"pending\""}, }},
				{From: "LANActive", To: "LANActive", On: Internal("ping_tick"), Sends: []Send{{To: "client", Msg: "path_ping"}, }},
				{From: "LANActive", To: "LANDegraded", On: Internal("ping_timeout"), Updates: []VarUpdate{{Var: "ping_failures", Expr: "1"}, }},
				{From: "LANDegraded", To: "LANDegraded", On: Internal("ping_tick"), Sends: []Send{{To: "client", Msg: "path_ping"}, }},
				{From: "LANActive", To: "RelayBackoff", On: Internal("lan_stream_error"), Do: "fallback_to_relay", Updates: []VarUpdate{{Var: "backoff_level", Expr: "Min(backoff_level + 1, max_backoff_level)"}, {Var: "b_active_path", Expr: "\"relay\""}, {Var: "b_dispatcher_path", Expr: "\"relay\""}, {Var: "monitor_target", Expr: "\"none\""}, {Var: "lan_signal", Expr: "\"pending\""}, {Var: "ping_failures", Expr: "0"}, }},
				{From: "LANDegraded", To: "RelayBackoff", On: Internal("lan_stream_error"), Do: "fallback_to_relay", Updates: []VarUpdate{{Var: "backoff_level", Expr: "Min(backoff_level + 1, max_backoff_level)"}, {Var: "b_active_path", Expr: "\"relay\""}, {Var: "b_dispatcher_path", Expr: "\"relay\""}, {Var: "monitor_target", Expr: "\"none\""}, {Var: "lan_signal", Expr: "\"pending\""}, {Var: "ping_failures", Expr: "0"}, }},
				{From: "LANDegraded", To: "LANActive", On: Recv("path_pong"), Do: "reset_failures", Updates: []VarUpdate{{Var: "ping_failures", Expr: "0"}, }},
				{From: "LANDegraded", To: "LANDegraded", On: Internal("ping_timeout"), Guard: "under_max_failures", Updates: []VarUpdate{{Var: "ping_failures", Expr: "ping_failures + 1"}, }},
				{From: "LANDegraded", To: "RelayBackoff", On: Internal("ping_timeout"), Guard: "at_max_failures", Do: "fallback_to_relay", Updates: []VarUpdate{{Var: "backoff_level", Expr: "Min(backoff_level + 1, max_backoff_level)"}, {Var: "b_active_path", Expr: "\"relay\""}, {Var: "b_dispatcher_path", Expr: "\"relay\""}, {Var: "monitor_target", Expr: "\"none\""}, {Var: "lan_signal", Expr: "\"pending\""}, {Var: "ping_failures", Expr: "0"}, }},
				{From: "RelayBackoff", To: "LANOffered", On: Internal("backoff_expired"), Sends: []Send{{To: "client", Msg: "lan_offer", Fields: map[string]string{"addr": "lan_addr", "challenge": "challenge_bytes", }}, }},
				{From: "RelayBackoff", To: "LANOffered", On: Internal("lan_server_changed"), Sends: []Send{{To: "client", Msg: "lan_offer", Fields: map[string]string{"addr": "lan_addr", "challenge": "challenge_bytes", }}, }, Updates: []VarUpdate{{Var: "backoff_level", Expr: "0"}, }},
				{From: "RelayConnected", To: "LANOffered", On: Internal("readvertise_tick"), Guard: "lan_server_available", Sends: []Send{{To: "client", Msg: "lan_offer", Fields: map[string]string{"addr": "lan_addr", "challenge": "challenge_bytes", }}, }},
				{From: "LANOffered", To: "RelayConnected", On: Internal("app_force_fallback"), Updates: []VarUpdate{{Var: "lan_signal", Expr: "\"pending\""}, }},
				{From: "LANActive", To: "RelayBackoff", On: Internal("app_force_fallback"), Do: "fallback_to_relay", Updates: []VarUpdate{{Var: "backoff_level", Expr: "Min(backoff_level + 1, max_backoff_level)"}, {Var: "b_active_path", Expr: "\"relay\""}, {Var: "b_dispatcher_path", Expr: "\"relay\""}, {Var: "monitor_target", Expr: "\"none\""}, {Var: "lan_signal", Expr: "\"pending\""}, {Var: "ping_failures", Expr: "0"}, }},
				{From: "LANDegraded", To: "RelayBackoff", On: Internal("app_force_fallback"), Do: "fallback_to_relay", Updates: []VarUpdate{{Var: "backoff_level", Expr: "Min(backoff_level + 1, max_backoff_level)"}, {Var: "b_active_path", Expr: "\"relay\""}, {Var: "b_dispatcher_path", Expr: "\"relay\""}, {Var: "monitor_target", Expr: "\"none\""}, {Var: "lan_signal", Expr: "\"pending\""}, {Var: "ping_failures", Expr: "0"}, }},
				{From: "RelayConnected", To: "Paired", On: Internal("disconnect")},
				{From: "RelayConnected", To: "RelayConnected", On: Internal("app_send")},
				{From: "LANOffered", To: "LANOffered", On: Internal("app_send")},
				{From: "LANActive", To: "LANActive", On: Internal("app_send")},
				{From: "LANDegraded", To: "LANDegraded", On: Internal("app_send")},
				{From: "RelayBackoff", To: "RelayBackoff", On: Internal("app_send")},
				{From: "RelayConnected", To: "RelayConnected", On: Internal("relay_stream_data")},
				{From: "LANOffered", To: "LANOffered", On: Internal("relay_stream_data")},
				{From: "LANActive", To: "LANActive", On: Internal("relay_stream_data")},
				{From: "LANDegraded", To: "LANDegraded", On: Internal("relay_stream_data")},
				{From: "RelayBackoff", To: "RelayBackoff", On: Internal("relay_stream_data")},
				{From: "RelayConnected", To: "RelayConnected", On: Internal("relay_stream_error")},
				{From: "LANOffered", To: "LANOffered", On: Internal("relay_stream_error")},
				{From: "LANActive", To: "LANActive", On: Internal("relay_stream_error")},
				{From: "LANDegraded", To: "LANDegraded", On: Internal("relay_stream_error")},
				{From: "RelayBackoff", To: "RelayBackoff", On: Internal("relay_stream_error")},
				{From: "RelayConnected", To: "RelayConnected", On: Internal("app_send_datagram")},
				{From: "LANOffered", To: "LANOffered", On: Internal("app_send_datagram")},
				{From: "LANActive", To: "LANActive", On: Internal("app_send_datagram")},
				{From: "LANDegraded", To: "LANDegraded", On: Internal("app_send_datagram")},
				{From: "RelayBackoff", To: "RelayBackoff", On: Internal("app_send_datagram")},
				{From: "RelayConnected", To: "RelayConnected", On: Internal("relay_datagram")},
				{From: "LANOffered", To: "LANOffered", On: Internal("relay_datagram")},
				{From: "LANActive", To: "LANActive", On: Internal("relay_datagram")},
				{From: "LANDegraded", To: "LANDegraded", On: Internal("relay_datagram")},
				{From: "RelayBackoff", To: "RelayBackoff", On: Internal("relay_datagram")},
				{From: "LANActive", To: "LANActive", On: Internal("lan_stream_data")},
				{From: "LANDegraded", To: "LANDegraded", On: Internal("lan_stream_data")},
				{From: "LANActive", To: "LANActive", On: Internal("lan_datagram")},
				{From: "LANDegraded", To: "LANDegraded", On: Internal("lan_datagram")},
			}},
			{Name: "client", Initial: "Idle", Transitions: []Transition{
				{From: "Idle", To: "ObtainBackchannelSecret", On: Internal("backchannel_received")},
				{From: "ObtainBackchannelSecret", To: "ConnectRelay", On: Internal("secret_parsed")},
				{From: "ConnectRelay", To: "GenKeyPair", On: Internal("relay_connected")},
				{From: "GenKeyPair", To: "WaitAck", On: Internal("key_pair_generated"), Do: "send_pair_hello", Sends: []Send{{To: "backend", Msg: "pair_hello", Fields: map[string]string{"pubkey": "\"client_pub\"", "token": "current_token", }}, }},
				{From: "WaitAck", To: "E2EReady", On: Recv("pair_hello_ack"), Do: "derive_secret", Updates: []VarUpdate{{Var: "received_backend_pub", Expr: "recv_msg.pubkey"}, {Var: "client_shared_key", Expr: "DeriveKey(\"client_pub\", recv_msg.pubkey)"}, }},
				{From: "E2EReady", To: "ShowCode", On: Recv("pair_confirm"), Updates: []VarUpdate{{Var: "client_code", Expr: "DeriveCode(received_backend_pub, \"client_pub\")"}, }},
				{From: "ShowCode", To: "WaitPairComplete", On: Internal("code_displayed")},
				{From: "WaitPairComplete", To: "Paired", On: Recv("pair_complete"), Do: "store_secret"},
				{From: "Paired", To: "Reconnect", On: Internal("app_launch")},
				{From: "Reconnect", To: "SendAuth", On: Internal("relay_connected"), Sends: []Send{{To: "backend", Msg: "auth_request", Fields: map[string]string{"device_id": "\"device_1\"", "key": "client_shared_key", "nonce": "\"nonce_1\"", "secret": "device_secret", }}, }},
				{From: "SendAuth", To: "SessionActive", On: Recv("auth_ok")},
				{From: "SessionActive", To: "RelayConnected", On: Internal("session_established")},
				{From: "RelayConnected", To: "LANConnecting", On: Recv("lan_offer"), Guard: "lan_enabled", Do: "dial_lan"},
				{From: "RelayConnected", To: "RelayConnected", On: Recv("lan_offer"), Guard: "lan_disabled"},
				{From: "LANConnecting", To: "LANVerifying", On: Internal("lan_dial_ok"), Sends: []Send{{To: "backend", Msg: "lan_verify", Fields: map[string]string{"challenge": "offer_challenge", "instance_id": "instance_id", }}, }},
				{From: "LANConnecting", To: "RelayConnected", On: Internal("lan_dial_failed")},
				{From: "LANVerifying", To: "LANActive", On: Recv("lan_confirm"), Do: "activate_lan", Updates: []VarUpdate{{Var: "c_active_path", Expr: "\"lan\""}, {Var: "c_dispatcher_path", Expr: "\"lan\""}, {Var: "lan_signal", Expr: "\"ready\""}, }},
				{From: "LANVerifying", To: "RelayConnected", On: Internal("verify_timeout"), Updates: []VarUpdate{{Var: "c_dispatcher_path", Expr: "\"relay\""}, }},
				{From: "LANActive", To: "LANActive", On: Recv("path_ping"), Sends: []Send{{To: "backend", Msg: "path_pong"}, }},
				{From: "LANActive", To: "RelayFallback", On: Internal("lan_error"), Do: "fallback_to_relay", Updates: []VarUpdate{{Var: "c_active_path", Expr: "\"relay\""}, {Var: "c_dispatcher_path", Expr: "\"relay\""}, {Var: "lan_signal", Expr: "\"pending\""}, }},
				{From: "LANActive", To: "RelayFallback", On: Internal("lan_stream_error"), Do: "fallback_to_relay", Updates: []VarUpdate{{Var: "c_active_path", Expr: "\"relay\""}, {Var: "c_dispatcher_path", Expr: "\"relay\""}, {Var: "lan_signal", Expr: "\"pending\""}, }},
				{From: "RelayFallback", To: "RelayConnected", On: Internal("relay_ok")},
				{From: "LANActive", To: "LANConnecting", On: Recv("lan_offer"), Guard: "lan_enabled", Do: "dial_lan"},
				{From: "LANConnecting", To: "RelayConnected", On: Internal("app_force_fallback")},
				{From: "LANVerifying", To: "RelayConnected", On: Internal("app_force_fallback"), Updates: []VarUpdate{{Var: "c_dispatcher_path", Expr: "\"relay\""}, }},
				{From: "LANActive", To: "RelayConnected", On: Internal("app_force_fallback"), Do: "fallback_to_relay", Updates: []VarUpdate{{Var: "c_active_path", Expr: "\"relay\""}, {Var: "c_dispatcher_path", Expr: "\"relay\""}, {Var: "lan_signal", Expr: "\"pending\""}, }},
				{From: "RelayConnected", To: "Paired", On: Internal("disconnect")},
				{From: "RelayConnected", To: "RelayConnected", On: Internal("app_send")},
				{From: "LANConnecting", To: "LANConnecting", On: Internal("app_send")},
				{From: "LANVerifying", To: "LANVerifying", On: Internal("app_send")},
				{From: "LANActive", To: "LANActive", On: Internal("app_send")},
				{From: "RelayFallback", To: "RelayFallback", On: Internal("app_send")},
				{From: "RelayConnected", To: "RelayConnected", On: Internal("relay_stream_data")},
				{From: "LANConnecting", To: "LANConnecting", On: Internal("relay_stream_data")},
				{From: "LANVerifying", To: "LANVerifying", On: Internal("relay_stream_data")},
				{From: "LANActive", To: "LANActive", On: Internal("relay_stream_data")},
				{From: "RelayFallback", To: "RelayFallback", On: Internal("relay_stream_data")},
				{From: "RelayConnected", To: "RelayConnected", On: Internal("relay_stream_error")},
				{From: "LANConnecting", To: "LANConnecting", On: Internal("relay_stream_error")},
				{From: "LANVerifying", To: "LANVerifying", On: Internal("relay_stream_error")},
				{From: "LANActive", To: "LANActive", On: Internal("relay_stream_error")},
				{From: "RelayFallback", To: "RelayFallback", On: Internal("relay_stream_error")},
				{From: "RelayConnected", To: "RelayConnected", On: Internal("app_send_datagram")},
				{From: "LANConnecting", To: "LANConnecting", On: Internal("app_send_datagram")},
				{From: "LANVerifying", To: "LANVerifying", On: Internal("app_send_datagram")},
				{From: "LANActive", To: "LANActive", On: Internal("app_send_datagram")},
				{From: "RelayFallback", To: "RelayFallback", On: Internal("app_send_datagram")},
				{From: "RelayConnected", To: "RelayConnected", On: Internal("relay_datagram")},
				{From: "LANConnecting", To: "LANConnecting", On: Internal("relay_datagram")},
				{From: "LANVerifying", To: "LANVerifying", On: Internal("relay_datagram")},
				{From: "LANActive", To: "LANActive", On: Internal("relay_datagram")},
				{From: "RelayFallback", To: "RelayFallback", On: Internal("relay_datagram")},
				{From: "LANActive", To: "LANActive", On: Internal("lan_stream_data")},
				{From: "LANActive", To: "LANActive", On: Internal("lan_datagram")},
			}},
			{Name: "relay", Initial: "Idle", Transitions: []Transition{
				{From: "Idle", To: "BackendRegistered", On: Internal("backend_register")},
				{From: "BackendRegistered", To: "Bridged", On: Internal("client_connect"), Do: "bridge_streams", Updates: []VarUpdate{{Var: "relay_bridge", Expr: "\"active\""}, }},
				{From: "Bridged", To: "BackendRegistered", On: Internal("client_disconnect"), Do: "unbridge", Updates: []VarUpdate{{Var: "relay_bridge", Expr: "\"idle\""}, }},
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
			{Type: "lan_offer", From: "backend", To: "client", Desc: "LAN address + challenge (sent via relay)"},
			{Type: "lan_verify", From: "client", To: "backend", Desc: "challenge response + instance ID (sent via LAN)"},
			{Type: "lan_confirm", From: "backend", To: "client", Desc: "LAN verified, path is live (sent via LAN)"},
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
			{Name: "lan_addr", Initial: "\"none\"", Desc: "LAN server address (host:port)"},
			{Name: "challenge_bytes", Initial: "\"none\"", Desc: "32-byte random challenge for LAN verification"},
			{Name: "offer_challenge", Initial: "\"none\"", Desc: "challenge from the most recent LAN offer"},
			{Name: "instance_id", Initial: "\"none\"", Desc: "relay instance ID"},
			{Name: "ping_failures", Initial: "0", Desc: "consecutive failed pings"},
			{Name: "max_ping_failures", Initial: "3", Desc: "threshold before fallback"},
			{Name: "backoff_level", Initial: "0", Desc: "exponential backoff level"},
			{Name: "max_backoff_level", Initial: "5", Desc: "backoff cap"},
			{Name: "lan_server_addr", Initial: "\"none\"", Desc: "LAN server listen address"},
			{Name: "b_active_path", Initial: "\"relay\"", Desc: "backend active path"},
			{Name: "c_active_path", Initial: "\"relay\"", Desc: "client active path"},
			{Name: "b_dispatcher_path", Initial: "\"relay\"", Desc: "backend datagram dispatcher binding"},
			{Name: "c_dispatcher_path", Initial: "\"relay\"", Desc: "client datagram dispatcher binding"},
			{Name: "monitor_target", Initial: "\"none\"", Desc: "health monitor target"},
			{Name: "lan_signal", Initial: "\"pending\"", Desc: "LANReady notification state"},
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
			{ID: "challenge_valid", Expr: "offer_challenge = challenge_bytes"},
			{ID: "challenge_invalid", Expr: "offer_challenge /= challenge_bytes"},
			{ID: "lan_enabled", Expr: "TRUE"},
			{ID: "lan_disabled", Expr: "FALSE"},
			{ID: "lan_server_available", Expr: "lan_server_addr /= \"none\""},
			{ID: "under_max_failures", Expr: "ping_failures + 1 < max_ping_failures"},
			{ID: "at_max_failures", Expr: "ping_failures + 1 >= max_ping_failures"},
		},
		Operators: []Operator{
			{Name: "KeyRank", Params: "k", Expr: "CASE k = \"adv_pub\" -> 0 [] k = \"client_pub\" -> 1 [] k = \"backend_pub\" -> 2 [] OTHER -> 3", Desc: "deterministic ordering for ECDH"},
			{Name: "DeriveKey", Params: "a, b", Expr: "IF KeyRank(a) <= KeyRank(b) THEN <<\"ecdh\", a, b>> ELSE <<\"ecdh\", b, a>>", Desc: "symbolic ECDH"},
			{Name: "DeriveCode", Params: "a, b", Expr: "IF KeyRank(a) <= KeyRank(b) THEN <<\"code\", a, b>> ELSE <<\"code\", b, a>>", Desc: "confirmation code from pubkeys"},
			{Name: "Min", Params: "a, b", Expr: "IF a < b THEN a ELSE b", Desc: "minimum of two values"},
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
			{Name: "PathConsistency", Kind: Invariant, Expr: "b_active_path \\in {\"relay\", \"lan\"} /\\ c_active_path \\in {\"relay\", \"lan\"}", Desc: "Paths are always valid"},
			{Name: "BackoffBounded", Kind: Invariant, Expr: "backoff_level <= max_backoff_level", Desc: "Backoff never exceeds cap"},
			{Name: "BackoffResetsOnSuccess", Kind: Invariant, Expr: "backend_state = backend_LANActive => backoff_level = 0", Desc: "LAN success resets backoff"},
			{Name: "DispatcherAlwaysBound", Kind: Invariant, Expr: "b_dispatcher_path \\in {\"relay\", \"lan\"} /\\ c_dispatcher_path \\in {\"relay\", \"lan\"}", Desc: "Dispatchers always bound to valid path"},
			{Name: "BackendDispatcherMatchesActive", Kind: Invariant, Expr: "backend_state = backend_LANActive => b_dispatcher_path = \"lan\"", Desc: "Backend dispatcher on LAN when LAN active"},
			{Name: "ClientDispatcherMatchesActive", Kind: Invariant, Expr: "client_state = client_LANActive => c_dispatcher_path = \"lan\"", Desc: "Client dispatcher on LAN when LAN active"},
			{Name: "MonitorOnlyWhenLAN", Kind: Invariant, Expr: "monitor_target = \"lan\" => backend_state \\in {backend_LANActive, backend_LANDegraded}", Desc: "Monitor only pings when LAN is active/degraded"},
			{Name: "FallbackLeadsToReadvertise", Kind: Invariant, Expr: "", Desc: "After fallback, backend eventually re-advertises LAN"},
			{Name: "DegradedLeadsToResolutionOrFallback", Kind: Invariant, Expr: "", Desc: "Degraded state eventually resolves (recovery or fallback)"},
		},
		ChannelBound: 3,
		OneShot: false,
	}
}

type ECDHState struct {
	BackendPub string // backend ECDH public key
	ClientPub string // pubkey received from client
	SharedKey string // ECDH-derived shared key
	Code string // confirmation code derived from pubkeys
}

type TokenState struct {
	Current string // pairing token currently in play
	Active frozen.Set[string] // set of valid (non-revoked) tokens
	Used frozen.Set[string] // set of revoked tokens
}

type BackendPathState struct {
	ActivePath string // which path carries traffic
	DispatcherPath string // datagram dispatcher binding
	MonitorTarget string // health monitor target
	LanSignal string // LANReady notification state
}

type ClientPathState struct {
	ActivePath string // which path carries traffic
	DispatcherPath string // datagram dispatcher binding
}

// SessionProtocolBackendMachine is the generated state machine for the backend actor.
type SessionProtocolBackendMachine struct {
	State State
	CurrentToken string // pairing token currently in play
	ActiveTokens frozen.Set[string] // set of valid (non-revoked) tokens
	UsedTokens frozen.Set[string] // set of revoked tokens
	BackendEcdhPub string // backend ECDH public key
	ReceivedClientPub string // pubkey backend received in pair_hello
	BackendSharedKey string // ECDH key derived by backend
	BackendCode string // code computed by backend
	ReceivedCode string // code entered via CLI
	CodeAttempts int // failed code submission attempts
	DeviceSecret string // persistent device secret
	PairedDevices frozen.Set[string] // device IDs that completed pairing
	ReceivedDeviceId string // device_id from auth_request
	AuthNoncesUsed frozen.Set[string] // set of consumed auth nonces
	ReceivedAuthNonce string // nonce from auth_request
	SecretPublished bool // whether token has been published via backchannel
	PingFailures int // consecutive failed pings
	BackoffLevel int // exponential backoff level
	BActivePath string // backend active path
	BDispatcherPath string // backend datagram dispatcher binding
	MonitorTarget string // health monitor target
	LanSignal string // LANReady notification state

	Guards  map[GuardID]func() bool
	Actions map[ActionID]func() error
	OnChange func(varName string)
}

func NewSessionProtocolBackendMachine() *SessionProtocolBackendMachine {
	return &SessionProtocolBackendMachine{
		State: SessionProtocolBackendIdle,
		CurrentToken: "none",
		BackendEcdhPub: "none",
		ReceivedClientPub: "none",
		BackendSharedKey: "",
		BackendCode: "",
		ReceivedCode: "",
		CodeAttempts: 0,
		DeviceSecret: "none",
		ReceivedDeviceId: "none",
		ReceivedAuthNonce: "none",
		SecretPublished: false,
		PingFailures: 0,
		BackoffLevel: 0,
		BActivePath: "relay",
		BDispatcherPath: "relay",
		MonitorTarget: "none",
		LanSignal: "pending",
		Guards:  make(map[GuardID]func() bool),
		Actions: make(map[ActionID]func() error),
	}
}

func (m *SessionProtocolBackendMachine) HandleMessage(msg MsgType) (bool, error) {
	switch {
	case m.State == SessionProtocolBackendWaitingForClient && msg == SessionProtocolMsgPairHello && m.Guards[SessionProtocolGuardTokenValid] != nil && m.Guards[SessionProtocolGuardTokenValid]():
		if fn := m.Actions[SessionProtocolActionDeriveSecret]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		// received_client_pub: recv_msg.pubkey (set by action)
		m.BackendEcdhPub = "backend_pub"
		if m.OnChange != nil { m.OnChange("backend_ecdh_pub") }
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
	case m.State == SessionProtocolBackendLANOffered && msg == SessionProtocolMsgLanVerify && m.Guards[SessionProtocolGuardChallengeValid] != nil && m.Guards[SessionProtocolGuardChallengeValid]():
		if fn := m.Actions[SessionProtocolActionActivateLan]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.PingFailures = 0
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.BackoffLevel = 0
		if m.OnChange != nil { m.OnChange("backoff_level") }
		m.BActivePath = "lan"
		if m.OnChange != nil { m.OnChange("b_active_path") }
		m.BDispatcherPath = "lan"
		if m.OnChange != nil { m.OnChange("b_dispatcher_path") }
		m.MonitorTarget = "lan"
		if m.OnChange != nil { m.OnChange("monitor_target") }
		m.LanSignal = "ready"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = SessionProtocolBackendLANActive
		return true, nil
	case m.State == SessionProtocolBackendLANOffered && msg == SessionProtocolMsgLanVerify && m.Guards[SessionProtocolGuardChallengeInvalid] != nil && m.Guards[SessionProtocolGuardChallengeInvalid]():
		m.State = SessionProtocolBackendRelayConnected
		return true, nil
	case m.State == SessionProtocolBackendLANDegraded && msg == SessionProtocolMsgPathPong:
		if fn := m.Actions[SessionProtocolActionResetFailures]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.PingFailures = 0
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = SessionProtocolBackendLANActive
		return true, nil
	}
	return false, nil
}

func (m *SessionProtocolBackendMachine) Step(event EventID) (bool, error) {
	switch {
	case m.State == SessionProtocolBackendIdle && event == SessionProtocolEventCliInitPair:
		if fn := m.Actions[SessionProtocolActionGenerateToken]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.CurrentToken = "tok_1"
		if m.OnChange != nil { m.OnChange("current_token") }
		// active_tokens: active_tokens \union {"tok_1"} (set by action)
		m.State = SessionProtocolBackendGenerateToken
		return true, nil
	case m.State == SessionProtocolBackendGenerateToken && event == SessionProtocolEventTokenCreated:
		if fn := m.Actions[SessionProtocolActionRegisterRelay]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.State = SessionProtocolBackendRegisterRelay
		return true, nil
	case m.State == SessionProtocolBackendRegisterRelay && event == SessionProtocolEventRelayRegistered:
		m.SecretPublished = true
		if m.OnChange != nil { m.OnChange("secret_published") }
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
		if m.OnChange != nil { m.OnChange("code_attempts") }
		m.State = SessionProtocolBackendIdle
		return true, nil
	case m.State == SessionProtocolBackendStorePaired && event == SessionProtocolEventFinalise:
		if fn := m.Actions[SessionProtocolActionStoreDevice]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.DeviceSecret = "dev_secret_1"
		if m.OnChange != nil { m.OnChange("device_secret") }
		// paired_devices: paired_devices \union {"device_1"} (set by action)
		// active_tokens: active_tokens \ {current_token} (set by action)
		// used_tokens: used_tokens \union {current_token} (set by action)
		m.State = SessionProtocolBackendPaired
		return true, nil
	case m.State == SessionProtocolBackendAuthCheck && event == SessionProtocolEventVerify && m.Guards[SessionProtocolGuardDeviceKnown] != nil && m.Guards[SessionProtocolGuardDeviceKnown]():
		if fn := m.Actions[SessionProtocolActionVerifyDevice]; fn != nil {
			if err := fn(); err != nil { return false, err }
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
	case m.State == SessionProtocolBackendRelayConnected && event == SessionProtocolEventLanServerReady:
		m.State = SessionProtocolBackendLANOffered
		return true, nil
	case m.State == SessionProtocolBackendLANOffered && event == SessionProtocolEventOfferTimeout:
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = SessionProtocolBackendRelayBackoff
		return true, nil
	case m.State == SessionProtocolBackendLANActive && event == SessionProtocolEventPingTick:
		m.State = SessionProtocolBackendLANActive
		return true, nil
	case m.State == SessionProtocolBackendLANActive && event == SessionProtocolEventPingTimeout:
		m.PingFailures = 1
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = SessionProtocolBackendLANDegraded
		return true, nil
	case m.State == SessionProtocolBackendLANDegraded && event == SessionProtocolEventPingTick:
		m.State = SessionProtocolBackendLANDegraded
		return true, nil
	case m.State == SessionProtocolBackendLANActive && event == SessionProtocolEventLanStreamError:
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.BActivePath = "relay"
		if m.OnChange != nil { m.OnChange("b_active_path") }
		m.BDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("b_dispatcher_path") }
		m.MonitorTarget = "none"
		if m.OnChange != nil { m.OnChange("monitor_target") }
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.PingFailures = 0
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = SessionProtocolBackendRelayBackoff
		return true, nil
	case m.State == SessionProtocolBackendLANDegraded && event == SessionProtocolEventLanStreamError:
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.BActivePath = "relay"
		if m.OnChange != nil { m.OnChange("b_active_path") }
		m.BDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("b_dispatcher_path") }
		m.MonitorTarget = "none"
		if m.OnChange != nil { m.OnChange("monitor_target") }
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.PingFailures = 0
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = SessionProtocolBackendRelayBackoff
		return true, nil
	case m.State == SessionProtocolBackendLANDegraded && event == SessionProtocolEventPingTimeout && m.Guards[SessionProtocolGuardUnderMaxFailures] != nil && m.Guards[SessionProtocolGuardUnderMaxFailures]():
		m.PingFailures = m.PingFailures + 1
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = SessionProtocolBackendLANDegraded
		return true, nil
	case m.State == SessionProtocolBackendLANDegraded && event == SessionProtocolEventPingTimeout && m.Guards[SessionProtocolGuardAtMaxFailures] != nil && m.Guards[SessionProtocolGuardAtMaxFailures]():
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.BActivePath = "relay"
		if m.OnChange != nil { m.OnChange("b_active_path") }
		m.BDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("b_dispatcher_path") }
		m.MonitorTarget = "none"
		if m.OnChange != nil { m.OnChange("monitor_target") }
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.PingFailures = 0
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = SessionProtocolBackendRelayBackoff
		return true, nil
	case m.State == SessionProtocolBackendRelayBackoff && event == SessionProtocolEventBackoffExpired:
		m.State = SessionProtocolBackendLANOffered
		return true, nil
	case m.State == SessionProtocolBackendRelayBackoff && event == SessionProtocolEventLanServerChanged:
		m.BackoffLevel = 0
		if m.OnChange != nil { m.OnChange("backoff_level") }
		m.State = SessionProtocolBackendLANOffered
		return true, nil
	case m.State == SessionProtocolBackendRelayConnected && event == SessionProtocolEventReadvertiseTick && m.Guards[SessionProtocolGuardLanServerAvailable] != nil && m.Guards[SessionProtocolGuardLanServerAvailable]():
		m.State = SessionProtocolBackendLANOffered
		return true, nil
	case m.State == SessionProtocolBackendLANOffered && event == SessionProtocolEventAppForceFallback:
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = SessionProtocolBackendRelayConnected
		return true, nil
	case m.State == SessionProtocolBackendLANActive && event == SessionProtocolEventAppForceFallback:
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.BActivePath = "relay"
		if m.OnChange != nil { m.OnChange("b_active_path") }
		m.BDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("b_dispatcher_path") }
		m.MonitorTarget = "none"
		if m.OnChange != nil { m.OnChange("monitor_target") }
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.PingFailures = 0
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = SessionProtocolBackendRelayBackoff
		return true, nil
	case m.State == SessionProtocolBackendLANDegraded && event == SessionProtocolEventAppForceFallback:
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.BActivePath = "relay"
		if m.OnChange != nil { m.OnChange("b_active_path") }
		m.BDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("b_dispatcher_path") }
		m.MonitorTarget = "none"
		if m.OnChange != nil { m.OnChange("monitor_target") }
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.PingFailures = 0
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = SessionProtocolBackendRelayBackoff
		return true, nil
	case m.State == SessionProtocolBackendRelayConnected && event == SessionProtocolEventDisconnect:
		m.State = SessionProtocolBackendPaired
		return true, nil
	case m.State == SessionProtocolBackendRelayConnected && event == SessionProtocolEventAppSend:
		m.State = SessionProtocolBackendRelayConnected
		return true, nil
	case m.State == SessionProtocolBackendLANOffered && event == SessionProtocolEventAppSend:
		m.State = SessionProtocolBackendLANOffered
		return true, nil
	case m.State == SessionProtocolBackendLANActive && event == SessionProtocolEventAppSend:
		m.State = SessionProtocolBackendLANActive
		return true, nil
	case m.State == SessionProtocolBackendLANDegraded && event == SessionProtocolEventAppSend:
		m.State = SessionProtocolBackendLANDegraded
		return true, nil
	case m.State == SessionProtocolBackendRelayBackoff && event == SessionProtocolEventAppSend:
		m.State = SessionProtocolBackendRelayBackoff
		return true, nil
	case m.State == SessionProtocolBackendRelayConnected && event == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolBackendRelayConnected
		return true, nil
	case m.State == SessionProtocolBackendLANOffered && event == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolBackendLANOffered
		return true, nil
	case m.State == SessionProtocolBackendLANActive && event == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolBackendLANActive
		return true, nil
	case m.State == SessionProtocolBackendLANDegraded && event == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolBackendLANDegraded
		return true, nil
	case m.State == SessionProtocolBackendRelayBackoff && event == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolBackendRelayBackoff
		return true, nil
	case m.State == SessionProtocolBackendRelayConnected && event == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolBackendRelayConnected
		return true, nil
	case m.State == SessionProtocolBackendLANOffered && event == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolBackendLANOffered
		return true, nil
	case m.State == SessionProtocolBackendLANActive && event == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolBackendLANActive
		return true, nil
	case m.State == SessionProtocolBackendLANDegraded && event == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolBackendLANDegraded
		return true, nil
	case m.State == SessionProtocolBackendRelayBackoff && event == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolBackendRelayBackoff
		return true, nil
	case m.State == SessionProtocolBackendRelayConnected && event == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolBackendRelayConnected
		return true, nil
	case m.State == SessionProtocolBackendLANOffered && event == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolBackendLANOffered
		return true, nil
	case m.State == SessionProtocolBackendLANActive && event == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolBackendLANActive
		return true, nil
	case m.State == SessionProtocolBackendLANDegraded && event == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolBackendLANDegraded
		return true, nil
	case m.State == SessionProtocolBackendRelayBackoff && event == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolBackendRelayBackoff
		return true, nil
	case m.State == SessionProtocolBackendRelayConnected && event == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolBackendRelayConnected
		return true, nil
	case m.State == SessionProtocolBackendLANOffered && event == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolBackendLANOffered
		return true, nil
	case m.State == SessionProtocolBackendLANActive && event == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolBackendLANActive
		return true, nil
	case m.State == SessionProtocolBackendLANDegraded && event == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolBackendLANDegraded
		return true, nil
	case m.State == SessionProtocolBackendRelayBackoff && event == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolBackendRelayBackoff
		return true, nil
	case m.State == SessionProtocolBackendLANActive && event == SessionProtocolEventLanStreamData:
		m.State = SessionProtocolBackendLANActive
		return true, nil
	case m.State == SessionProtocolBackendLANDegraded && event == SessionProtocolEventLanStreamData:
		m.State = SessionProtocolBackendLANDegraded
		return true, nil
	case m.State == SessionProtocolBackendLANActive && event == SessionProtocolEventLanDatagram:
		m.State = SessionProtocolBackendLANActive
		return true, nil
	case m.State == SessionProtocolBackendLANDegraded && event == SessionProtocolEventLanDatagram:
		m.State = SessionProtocolBackendLANDegraded
		return true, nil
	}
	return false, nil
}

func (m *SessionProtocolBackendMachine) HandleEvent(ev EventID) ([]CmdID, error) {
	switch {
	case m.State == SessionProtocolBackendIdle && ev == SessionProtocolEventCliInitPair:
		if fn := m.Actions[SessionProtocolActionGenerateToken]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.CurrentToken = "tok_1"
		if m.OnChange != nil { m.OnChange("current_token") }
		// active_tokens: active_tokens \union {"tok_1"} (set by action)
		m.State = SessionProtocolBackendGenerateToken
		return nil, nil
	case m.State == SessionProtocolBackendGenerateToken && ev == SessionProtocolEventTokenCreated:
		if fn := m.Actions[SessionProtocolActionRegisterRelay]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.State = SessionProtocolBackendRegisterRelay
		return nil, nil
	case m.State == SessionProtocolBackendRegisterRelay && ev == SessionProtocolEventRelayRegistered:
		m.SecretPublished = true
		if m.OnChange != nil { m.OnChange("secret_published") }
		m.State = SessionProtocolBackendWaitingForClient
		return nil, nil
	case m.State == SessionProtocolBackendWaitingForClient && ev == SessionProtocolEventRecvPairHello && m.Guards[SessionProtocolGuardTokenValid] != nil && m.Guards[SessionProtocolGuardTokenValid]():
		if fn := m.Actions[SessionProtocolActionDeriveSecret]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		// received_client_pub: recv_msg.pubkey (set by action)
		m.BackendEcdhPub = "backend_pub"
		if m.OnChange != nil { m.OnChange("backend_ecdh_pub") }
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
		if m.OnChange != nil { m.OnChange("code_attempts") }
		m.State = SessionProtocolBackendIdle
		return nil, nil
	case m.State == SessionProtocolBackendStorePaired && ev == SessionProtocolEventFinalise:
		if fn := m.Actions[SessionProtocolActionStoreDevice]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.DeviceSecret = "dev_secret_1"
		if m.OnChange != nil { m.OnChange("device_secret") }
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
			if err := fn(); err != nil { return nil, err }
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
	case m.State == SessionProtocolBackendRelayConnected && ev == SessionProtocolEventLanServerReady:
		m.State = SessionProtocolBackendLANOffered
		return []CmdID{SessionProtocolCmdSendLanOffer}, nil
	case m.State == SessionProtocolBackendLANOffered && ev == SessionProtocolEventRecvLanVerify && m.Guards[SessionProtocolGuardChallengeValid] != nil && m.Guards[SessionProtocolGuardChallengeValid]():
		if fn := m.Actions[SessionProtocolActionActivateLan]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.PingFailures = 0
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.BackoffLevel = 0
		if m.OnChange != nil { m.OnChange("backoff_level") }
		m.BActivePath = "lan"
		if m.OnChange != nil { m.OnChange("b_active_path") }
		m.BDispatcherPath = "lan"
		if m.OnChange != nil { m.OnChange("b_dispatcher_path") }
		m.MonitorTarget = "lan"
		if m.OnChange != nil { m.OnChange("monitor_target") }
		m.LanSignal = "ready"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = SessionProtocolBackendLANActive
		return []CmdID{SessionProtocolCmdSendLanConfirm, SessionProtocolCmdStartLanStreamReader, SessionProtocolCmdStartLanDgReader, SessionProtocolCmdStartMonitor, SessionProtocolCmdSignalLanReady, SessionProtocolCmdSetCryptoDatagram}, nil
	case m.State == SessionProtocolBackendLANOffered && ev == SessionProtocolEventRecvLanVerify && m.Guards[SessionProtocolGuardChallengeInvalid] != nil && m.Guards[SessionProtocolGuardChallengeInvalid]():
		m.State = SessionProtocolBackendRelayConnected
		return nil, nil
	case m.State == SessionProtocolBackendLANOffered && ev == SessionProtocolEventOfferTimeout:
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = SessionProtocolBackendRelayBackoff
		return []CmdID{SessionProtocolCmdResetLanReady, SessionProtocolCmdStartBackoffTimer}, nil
	case m.State == SessionProtocolBackendLANActive && ev == SessionProtocolEventPingTick:
		m.State = SessionProtocolBackendLANActive
		return []CmdID{SessionProtocolCmdSendPathPing, SessionProtocolCmdStartPongTimeout}, nil
	case m.State == SessionProtocolBackendLANActive && ev == SessionProtocolEventPingTimeout:
		m.PingFailures = 1
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = SessionProtocolBackendLANDegraded
		return nil, nil
	case m.State == SessionProtocolBackendLANDegraded && ev == SessionProtocolEventPingTick:
		m.State = SessionProtocolBackendLANDegraded
		return []CmdID{SessionProtocolCmdSendPathPing, SessionProtocolCmdStartPongTimeout}, nil
	case m.State == SessionProtocolBackendLANActive && ev == SessionProtocolEventLanStreamError:
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.BActivePath = "relay"
		if m.OnChange != nil { m.OnChange("b_active_path") }
		m.BDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("b_dispatcher_path") }
		m.MonitorTarget = "none"
		if m.OnChange != nil { m.OnChange("monitor_target") }
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.PingFailures = 0
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = SessionProtocolBackendRelayBackoff
		return []CmdID{SessionProtocolCmdStopMonitor, SessionProtocolCmdStopLanStreamReader, SessionProtocolCmdStopLanDgReader, SessionProtocolCmdCloseLanPath, SessionProtocolCmdResetLanReady, SessionProtocolCmdStartBackoffTimer}, nil
	case m.State == SessionProtocolBackendLANDegraded && ev == SessionProtocolEventLanStreamError:
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.BActivePath = "relay"
		if m.OnChange != nil { m.OnChange("b_active_path") }
		m.BDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("b_dispatcher_path") }
		m.MonitorTarget = "none"
		if m.OnChange != nil { m.OnChange("monitor_target") }
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.PingFailures = 0
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = SessionProtocolBackendRelayBackoff
		return []CmdID{SessionProtocolCmdStopMonitor, SessionProtocolCmdStopLanStreamReader, SessionProtocolCmdStopLanDgReader, SessionProtocolCmdCloseLanPath, SessionProtocolCmdResetLanReady, SessionProtocolCmdStartBackoffTimer}, nil
	case m.State == SessionProtocolBackendLANDegraded && ev == SessionProtocolEventRecvPathPong:
		if fn := m.Actions[SessionProtocolActionResetFailures]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.PingFailures = 0
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = SessionProtocolBackendLANActive
		return []CmdID{SessionProtocolCmdCancelPongTimeout}, nil
	case m.State == SessionProtocolBackendLANDegraded && ev == SessionProtocolEventPingTimeout && m.Guards[SessionProtocolGuardUnderMaxFailures] != nil && m.Guards[SessionProtocolGuardUnderMaxFailures]():
		m.PingFailures = m.PingFailures + 1
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = SessionProtocolBackendLANDegraded
		return nil, nil
	case m.State == SessionProtocolBackendLANDegraded && ev == SessionProtocolEventPingTimeout && m.Guards[SessionProtocolGuardAtMaxFailures] != nil && m.Guards[SessionProtocolGuardAtMaxFailures]():
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.BActivePath = "relay"
		if m.OnChange != nil { m.OnChange("b_active_path") }
		m.BDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("b_dispatcher_path") }
		m.MonitorTarget = "none"
		if m.OnChange != nil { m.OnChange("monitor_target") }
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.PingFailures = 0
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = SessionProtocolBackendRelayBackoff
		return []CmdID{SessionProtocolCmdStopMonitor, SessionProtocolCmdStopLanStreamReader, SessionProtocolCmdStopLanDgReader, SessionProtocolCmdCloseLanPath, SessionProtocolCmdResetLanReady, SessionProtocolCmdStartBackoffTimer}, nil
	case m.State == SessionProtocolBackendRelayBackoff && ev == SessionProtocolEventBackoffExpired:
		m.State = SessionProtocolBackendLANOffered
		return []CmdID{SessionProtocolCmdSendLanOffer}, nil
	case m.State == SessionProtocolBackendRelayBackoff && ev == SessionProtocolEventLanServerChanged:
		m.BackoffLevel = 0
		if m.OnChange != nil { m.OnChange("backoff_level") }
		m.State = SessionProtocolBackendLANOffered
		return []CmdID{SessionProtocolCmdSendLanOffer}, nil
	case m.State == SessionProtocolBackendRelayConnected && ev == SessionProtocolEventReadvertiseTick && m.Guards[SessionProtocolGuardLanServerAvailable] != nil && m.Guards[SessionProtocolGuardLanServerAvailable]():
		m.State = SessionProtocolBackendLANOffered
		return []CmdID{SessionProtocolCmdSendLanOffer}, nil
	case m.State == SessionProtocolBackendLANOffered && ev == SessionProtocolEventAppForceFallback:
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = SessionProtocolBackendRelayConnected
		return []CmdID{SessionProtocolCmdResetLanReady}, nil
	case m.State == SessionProtocolBackendLANActive && ev == SessionProtocolEventAppForceFallback:
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.BActivePath = "relay"
		if m.OnChange != nil { m.OnChange("b_active_path") }
		m.BDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("b_dispatcher_path") }
		m.MonitorTarget = "none"
		if m.OnChange != nil { m.OnChange("monitor_target") }
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.PingFailures = 0
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = SessionProtocolBackendRelayBackoff
		return []CmdID{SessionProtocolCmdStopMonitor, SessionProtocolCmdCancelPongTimeout, SessionProtocolCmdStopLanStreamReader, SessionProtocolCmdStopLanDgReader, SessionProtocolCmdCloseLanPath, SessionProtocolCmdResetLanReady, SessionProtocolCmdStartBackoffTimer}, nil
	case m.State == SessionProtocolBackendLANDegraded && ev == SessionProtocolEventAppForceFallback:
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.BActivePath = "relay"
		if m.OnChange != nil { m.OnChange("b_active_path") }
		m.BDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("b_dispatcher_path") }
		m.MonitorTarget = "none"
		if m.OnChange != nil { m.OnChange("monitor_target") }
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.PingFailures = 0
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = SessionProtocolBackendRelayBackoff
		return []CmdID{SessionProtocolCmdStopMonitor, SessionProtocolCmdCancelPongTimeout, SessionProtocolCmdStopLanStreamReader, SessionProtocolCmdStopLanDgReader, SessionProtocolCmdCloseLanPath, SessionProtocolCmdResetLanReady, SessionProtocolCmdStartBackoffTimer}, nil
	case m.State == SessionProtocolBackendRelayConnected && ev == SessionProtocolEventDisconnect:
		m.State = SessionProtocolBackendPaired
		return nil, nil
	case m.State == SessionProtocolBackendRelayConnected && ev == SessionProtocolEventAppSend:
		m.State = SessionProtocolBackendRelayConnected
		return []CmdID{SessionProtocolCmdWriteActiveStream}, nil
	case m.State == SessionProtocolBackendLANOffered && ev == SessionProtocolEventAppSend:
		m.State = SessionProtocolBackendLANOffered
		return []CmdID{SessionProtocolCmdWriteActiveStream}, nil
	case m.State == SessionProtocolBackendLANActive && ev == SessionProtocolEventAppSend:
		m.State = SessionProtocolBackendLANActive
		return []CmdID{SessionProtocolCmdWriteActiveStream}, nil
	case m.State == SessionProtocolBackendLANDegraded && ev == SessionProtocolEventAppSend:
		m.State = SessionProtocolBackendLANDegraded
		return []CmdID{SessionProtocolCmdWriteActiveStream}, nil
	case m.State == SessionProtocolBackendRelayBackoff && ev == SessionProtocolEventAppSend:
		m.State = SessionProtocolBackendRelayBackoff
		return []CmdID{SessionProtocolCmdWriteActiveStream}, nil
	case m.State == SessionProtocolBackendRelayConnected && ev == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolBackendRelayConnected
		return []CmdID{SessionProtocolCmdDeliverRecv}, nil
	case m.State == SessionProtocolBackendLANOffered && ev == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolBackendLANOffered
		return []CmdID{SessionProtocolCmdDeliverRecv}, nil
	case m.State == SessionProtocolBackendLANActive && ev == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolBackendLANActive
		return []CmdID{SessionProtocolCmdDeliverRecv}, nil
	case m.State == SessionProtocolBackendLANDegraded && ev == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolBackendLANDegraded
		return []CmdID{SessionProtocolCmdDeliverRecv}, nil
	case m.State == SessionProtocolBackendRelayBackoff && ev == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolBackendRelayBackoff
		return []CmdID{SessionProtocolCmdDeliverRecv}, nil
	case m.State == SessionProtocolBackendRelayConnected && ev == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolBackendRelayConnected
		return []CmdID{SessionProtocolCmdDeliverRecvError}, nil
	case m.State == SessionProtocolBackendLANOffered && ev == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolBackendLANOffered
		return []CmdID{SessionProtocolCmdDeliverRecvError}, nil
	case m.State == SessionProtocolBackendLANActive && ev == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolBackendLANActive
		return []CmdID{SessionProtocolCmdDeliverRecvError}, nil
	case m.State == SessionProtocolBackendLANDegraded && ev == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolBackendLANDegraded
		return []CmdID{SessionProtocolCmdDeliverRecvError}, nil
	case m.State == SessionProtocolBackendRelayBackoff && ev == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolBackendRelayBackoff
		return []CmdID{SessionProtocolCmdDeliverRecvError}, nil
	case m.State == SessionProtocolBackendRelayConnected && ev == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolBackendRelayConnected
		return []CmdID{SessionProtocolCmdSendActiveDatagram}, nil
	case m.State == SessionProtocolBackendLANOffered && ev == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolBackendLANOffered
		return []CmdID{SessionProtocolCmdSendActiveDatagram}, nil
	case m.State == SessionProtocolBackendLANActive && ev == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolBackendLANActive
		return []CmdID{SessionProtocolCmdSendActiveDatagram}, nil
	case m.State == SessionProtocolBackendLANDegraded && ev == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolBackendLANDegraded
		return []CmdID{SessionProtocolCmdSendActiveDatagram}, nil
	case m.State == SessionProtocolBackendRelayBackoff && ev == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolBackendRelayBackoff
		return []CmdID{SessionProtocolCmdSendActiveDatagram}, nil
	case m.State == SessionProtocolBackendRelayConnected && ev == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolBackendRelayConnected
		return []CmdID{SessionProtocolCmdDeliverRecvDatagram}, nil
	case m.State == SessionProtocolBackendLANOffered && ev == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolBackendLANOffered
		return []CmdID{SessionProtocolCmdDeliverRecvDatagram}, nil
	case m.State == SessionProtocolBackendLANActive && ev == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolBackendLANActive
		return []CmdID{SessionProtocolCmdDeliverRecvDatagram}, nil
	case m.State == SessionProtocolBackendLANDegraded && ev == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolBackendLANDegraded
		return []CmdID{SessionProtocolCmdDeliverRecvDatagram}, nil
	case m.State == SessionProtocolBackendRelayBackoff && ev == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolBackendRelayBackoff
		return []CmdID{SessionProtocolCmdDeliverRecvDatagram}, nil
	case m.State == SessionProtocolBackendLANActive && ev == SessionProtocolEventLanStreamData:
		m.State = SessionProtocolBackendLANActive
		return []CmdID{SessionProtocolCmdDeliverRecv}, nil
	case m.State == SessionProtocolBackendLANDegraded && ev == SessionProtocolEventLanStreamData:
		m.State = SessionProtocolBackendLANDegraded
		return []CmdID{SessionProtocolCmdDeliverRecv}, nil
	case m.State == SessionProtocolBackendLANActive && ev == SessionProtocolEventLanDatagram:
		m.State = SessionProtocolBackendLANActive
		return []CmdID{SessionProtocolCmdDeliverRecvDatagram}, nil
	case m.State == SessionProtocolBackendLANDegraded && ev == SessionProtocolEventLanDatagram:
		m.State = SessionProtocolBackendLANDegraded
		return []CmdID{SessionProtocolCmdDeliverRecvDatagram}, nil
	}
	return nil, nil
}

// SessionProtocolClientMachine is the generated state machine for the client actor.
type SessionProtocolClientMachine struct {
	State State
	ReceivedBackendPub string // pubkey client received in pair_hello_ack
	ClientSharedKey string // ECDH key derived by client
	ClientCode string // code computed by client
	CActivePath string // client active path
	CDispatcherPath string // client datagram dispatcher binding
	LanSignal string // LANReady notification state

	Guards  map[GuardID]func() bool
	Actions map[ActionID]func() error
	OnChange func(varName string)
}

func NewSessionProtocolClientMachine() *SessionProtocolClientMachine {
	return &SessionProtocolClientMachine{
		State: SessionProtocolClientIdle,
		ReceivedBackendPub: "none",
		ClientSharedKey: "",
		ClientCode: "",
		CActivePath: "relay",
		CDispatcherPath: "relay",
		LanSignal: "pending",
		Guards:  make(map[GuardID]func() bool),
		Actions: make(map[ActionID]func() error),
	}
}

func (m *SessionProtocolClientMachine) HandleMessage(msg MsgType) (bool, error) {
	switch {
	case m.State == SessionProtocolClientWaitAck && msg == SessionProtocolMsgPairHelloAck:
		if fn := m.Actions[SessionProtocolActionDeriveSecret]; fn != nil {
			if err := fn(); err != nil { return false, err }
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
			if err := fn(); err != nil { return false, err }
		}
		m.State = SessionProtocolClientPaired
		return true, nil
	case m.State == SessionProtocolClientSendAuth && msg == SessionProtocolMsgAuthOk:
		m.State = SessionProtocolClientSessionActive
		return true, nil
	case m.State == SessionProtocolClientRelayConnected && msg == SessionProtocolMsgLanOffer && m.Guards[SessionProtocolGuardLanEnabled] != nil && m.Guards[SessionProtocolGuardLanEnabled]():
		if fn := m.Actions[SessionProtocolActionDialLan]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.State = SessionProtocolClientLANConnecting
		return true, nil
	case m.State == SessionProtocolClientRelayConnected && msg == SessionProtocolMsgLanOffer && m.Guards[SessionProtocolGuardLanDisabled] != nil && m.Guards[SessionProtocolGuardLanDisabled]():
		m.State = SessionProtocolClientRelayConnected
		return true, nil
	case m.State == SessionProtocolClientLANVerifying && msg == SessionProtocolMsgLanConfirm:
		if fn := m.Actions[SessionProtocolActionActivateLan]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.CActivePath = "lan"
		if m.OnChange != nil { m.OnChange("c_active_path") }
		m.CDispatcherPath = "lan"
		if m.OnChange != nil { m.OnChange("c_dispatcher_path") }
		m.LanSignal = "ready"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = SessionProtocolClientLANActive
		return true, nil
	case m.State == SessionProtocolClientLANActive && msg == SessionProtocolMsgPathPing:
		m.State = SessionProtocolClientLANActive
		return true, nil
	case m.State == SessionProtocolClientLANActive && msg == SessionProtocolMsgLanOffer && m.Guards[SessionProtocolGuardLanEnabled] != nil && m.Guards[SessionProtocolGuardLanEnabled]():
		if fn := m.Actions[SessionProtocolActionDialLan]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.State = SessionProtocolClientLANConnecting
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
			if err := fn(); err != nil { return false, err }
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
	case m.State == SessionProtocolClientLANConnecting && event == SessionProtocolEventLanDialOk:
		m.State = SessionProtocolClientLANVerifying
		return true, nil
	case m.State == SessionProtocolClientLANConnecting && event == SessionProtocolEventLanDialFailed:
		m.State = SessionProtocolClientRelayConnected
		return true, nil
	case m.State == SessionProtocolClientLANVerifying && event == SessionProtocolEventVerifyTimeout:
		m.CDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("c_dispatcher_path") }
		m.State = SessionProtocolClientRelayConnected
		return true, nil
	case m.State == SessionProtocolClientLANActive && event == SessionProtocolEventLanError:
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.CActivePath = "relay"
		if m.OnChange != nil { m.OnChange("c_active_path") }
		m.CDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("c_dispatcher_path") }
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = SessionProtocolClientRelayFallback
		return true, nil
	case m.State == SessionProtocolClientLANActive && event == SessionProtocolEventLanStreamError:
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.CActivePath = "relay"
		if m.OnChange != nil { m.OnChange("c_active_path") }
		m.CDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("c_dispatcher_path") }
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = SessionProtocolClientRelayFallback
		return true, nil
	case m.State == SessionProtocolClientRelayFallback && event == SessionProtocolEventRelayOk:
		m.State = SessionProtocolClientRelayConnected
		return true, nil
	case m.State == SessionProtocolClientLANConnecting && event == SessionProtocolEventAppForceFallback:
		m.State = SessionProtocolClientRelayConnected
		return true, nil
	case m.State == SessionProtocolClientLANVerifying && event == SessionProtocolEventAppForceFallback:
		m.CDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("c_dispatcher_path") }
		m.State = SessionProtocolClientRelayConnected
		return true, nil
	case m.State == SessionProtocolClientLANActive && event == SessionProtocolEventAppForceFallback:
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.CActivePath = "relay"
		if m.OnChange != nil { m.OnChange("c_active_path") }
		m.CDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("c_dispatcher_path") }
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = SessionProtocolClientRelayConnected
		return true, nil
	case m.State == SessionProtocolClientRelayConnected && event == SessionProtocolEventDisconnect:
		m.State = SessionProtocolClientPaired
		return true, nil
	case m.State == SessionProtocolClientRelayConnected && event == SessionProtocolEventAppSend:
		m.State = SessionProtocolClientRelayConnected
		return true, nil
	case m.State == SessionProtocolClientLANConnecting && event == SessionProtocolEventAppSend:
		m.State = SessionProtocolClientLANConnecting
		return true, nil
	case m.State == SessionProtocolClientLANVerifying && event == SessionProtocolEventAppSend:
		m.State = SessionProtocolClientLANVerifying
		return true, nil
	case m.State == SessionProtocolClientLANActive && event == SessionProtocolEventAppSend:
		m.State = SessionProtocolClientLANActive
		return true, nil
	case m.State == SessionProtocolClientRelayFallback && event == SessionProtocolEventAppSend:
		m.State = SessionProtocolClientRelayFallback
		return true, nil
	case m.State == SessionProtocolClientRelayConnected && event == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolClientRelayConnected
		return true, nil
	case m.State == SessionProtocolClientLANConnecting && event == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolClientLANConnecting
		return true, nil
	case m.State == SessionProtocolClientLANVerifying && event == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolClientLANVerifying
		return true, nil
	case m.State == SessionProtocolClientLANActive && event == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolClientLANActive
		return true, nil
	case m.State == SessionProtocolClientRelayFallback && event == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolClientRelayFallback
		return true, nil
	case m.State == SessionProtocolClientRelayConnected && event == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolClientRelayConnected
		return true, nil
	case m.State == SessionProtocolClientLANConnecting && event == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolClientLANConnecting
		return true, nil
	case m.State == SessionProtocolClientLANVerifying && event == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolClientLANVerifying
		return true, nil
	case m.State == SessionProtocolClientLANActive && event == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolClientLANActive
		return true, nil
	case m.State == SessionProtocolClientRelayFallback && event == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolClientRelayFallback
		return true, nil
	case m.State == SessionProtocolClientRelayConnected && event == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolClientRelayConnected
		return true, nil
	case m.State == SessionProtocolClientLANConnecting && event == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolClientLANConnecting
		return true, nil
	case m.State == SessionProtocolClientLANVerifying && event == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolClientLANVerifying
		return true, nil
	case m.State == SessionProtocolClientLANActive && event == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolClientLANActive
		return true, nil
	case m.State == SessionProtocolClientRelayFallback && event == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolClientRelayFallback
		return true, nil
	case m.State == SessionProtocolClientRelayConnected && event == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolClientRelayConnected
		return true, nil
	case m.State == SessionProtocolClientLANConnecting && event == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolClientLANConnecting
		return true, nil
	case m.State == SessionProtocolClientLANVerifying && event == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolClientLANVerifying
		return true, nil
	case m.State == SessionProtocolClientLANActive && event == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolClientLANActive
		return true, nil
	case m.State == SessionProtocolClientRelayFallback && event == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolClientRelayFallback
		return true, nil
	case m.State == SessionProtocolClientLANActive && event == SessionProtocolEventLanStreamData:
		m.State = SessionProtocolClientLANActive
		return true, nil
	case m.State == SessionProtocolClientLANActive && event == SessionProtocolEventLanDatagram:
		m.State = SessionProtocolClientLANActive
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
			if err := fn(); err != nil { return nil, err }
		}
		m.State = SessionProtocolClientWaitAck
		return nil, nil
	case m.State == SessionProtocolClientWaitAck && ev == SessionProtocolEventRecvPairHelloAck:
		if fn := m.Actions[SessionProtocolActionDeriveSecret]; fn != nil {
			if err := fn(); err != nil { return nil, err }
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
			if err := fn(); err != nil { return nil, err }
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
	case m.State == SessionProtocolClientRelayConnected && ev == SessionProtocolEventRecvLanOffer && m.Guards[SessionProtocolGuardLanEnabled] != nil && m.Guards[SessionProtocolGuardLanEnabled]():
		if fn := m.Actions[SessionProtocolActionDialLan]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.State = SessionProtocolClientLANConnecting
		return []CmdID{SessionProtocolCmdDialLan}, nil
	case m.State == SessionProtocolClientRelayConnected && ev == SessionProtocolEventRecvLanOffer && m.Guards[SessionProtocolGuardLanDisabled] != nil && m.Guards[SessionProtocolGuardLanDisabled]():
		m.State = SessionProtocolClientRelayConnected
		return nil, nil
	case m.State == SessionProtocolClientLANConnecting && ev == SessionProtocolEventLanDialOk:
		m.State = SessionProtocolClientLANVerifying
		return []CmdID{SessionProtocolCmdSendLanVerify}, nil
	case m.State == SessionProtocolClientLANConnecting && ev == SessionProtocolEventLanDialFailed:
		m.State = SessionProtocolClientRelayConnected
		return nil, nil
	case m.State == SessionProtocolClientLANVerifying && ev == SessionProtocolEventRecvLanConfirm:
		if fn := m.Actions[SessionProtocolActionActivateLan]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.CActivePath = "lan"
		if m.OnChange != nil { m.OnChange("c_active_path") }
		m.CDispatcherPath = "lan"
		if m.OnChange != nil { m.OnChange("c_dispatcher_path") }
		m.LanSignal = "ready"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = SessionProtocolClientLANActive
		return []CmdID{SessionProtocolCmdStartLanStreamReader, SessionProtocolCmdStartLanDgReader, SessionProtocolCmdSignalLanReady, SessionProtocolCmdSetCryptoDatagram}, nil
	case m.State == SessionProtocolClientLANVerifying && ev == SessionProtocolEventVerifyTimeout:
		m.CDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("c_dispatcher_path") }
		m.State = SessionProtocolClientRelayConnected
		return nil, nil
	case m.State == SessionProtocolClientLANActive && ev == SessionProtocolEventRecvPathPing:
		m.State = SessionProtocolClientLANActive
		return []CmdID{SessionProtocolCmdSendPathPong}, nil
	case m.State == SessionProtocolClientLANActive && ev == SessionProtocolEventLanError:
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.CActivePath = "relay"
		if m.OnChange != nil { m.OnChange("c_active_path") }
		m.CDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("c_dispatcher_path") }
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = SessionProtocolClientRelayFallback
		return []CmdID{SessionProtocolCmdStopLanStreamReader, SessionProtocolCmdStopLanDgReader, SessionProtocolCmdCloseLanPath, SessionProtocolCmdResetLanReady}, nil
	case m.State == SessionProtocolClientLANActive && ev == SessionProtocolEventLanStreamError:
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.CActivePath = "relay"
		if m.OnChange != nil { m.OnChange("c_active_path") }
		m.CDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("c_dispatcher_path") }
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = SessionProtocolClientRelayFallback
		return []CmdID{SessionProtocolCmdStopLanStreamReader, SessionProtocolCmdStopLanDgReader, SessionProtocolCmdCloseLanPath, SessionProtocolCmdResetLanReady}, nil
	case m.State == SessionProtocolClientRelayFallback && ev == SessionProtocolEventRelayOk:
		m.State = SessionProtocolClientRelayConnected
		return nil, nil
	case m.State == SessionProtocolClientLANActive && ev == SessionProtocolEventRecvLanOffer && m.Guards[SessionProtocolGuardLanEnabled] != nil && m.Guards[SessionProtocolGuardLanEnabled]():
		if fn := m.Actions[SessionProtocolActionDialLan]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.State = SessionProtocolClientLANConnecting
		return []CmdID{SessionProtocolCmdStopLanStreamReader, SessionProtocolCmdStopLanDgReader, SessionProtocolCmdCloseLanPath, SessionProtocolCmdDialLan}, nil
	case m.State == SessionProtocolClientLANConnecting && ev == SessionProtocolEventAppForceFallback:
		m.State = SessionProtocolClientRelayConnected
		return nil, nil
	case m.State == SessionProtocolClientLANVerifying && ev == SessionProtocolEventAppForceFallback:
		m.CDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("c_dispatcher_path") }
		m.State = SessionProtocolClientRelayConnected
		return []CmdID{SessionProtocolCmdStopLanStreamReader, SessionProtocolCmdStopLanDgReader, SessionProtocolCmdCloseLanPath}, nil
	case m.State == SessionProtocolClientLANActive && ev == SessionProtocolEventAppForceFallback:
		if fn := m.Actions[SessionProtocolActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.CActivePath = "relay"
		if m.OnChange != nil { m.OnChange("c_active_path") }
		m.CDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("c_dispatcher_path") }
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = SessionProtocolClientRelayConnected
		return []CmdID{SessionProtocolCmdStopLanStreamReader, SessionProtocolCmdStopLanDgReader, SessionProtocolCmdCloseLanPath, SessionProtocolCmdResetLanReady}, nil
	case m.State == SessionProtocolClientRelayConnected && ev == SessionProtocolEventDisconnect:
		m.State = SessionProtocolClientPaired
		return nil, nil
	case m.State == SessionProtocolClientRelayConnected && ev == SessionProtocolEventAppSend:
		m.State = SessionProtocolClientRelayConnected
		return []CmdID{SessionProtocolCmdWriteActiveStream}, nil
	case m.State == SessionProtocolClientLANConnecting && ev == SessionProtocolEventAppSend:
		m.State = SessionProtocolClientLANConnecting
		return []CmdID{SessionProtocolCmdWriteActiveStream}, nil
	case m.State == SessionProtocolClientLANVerifying && ev == SessionProtocolEventAppSend:
		m.State = SessionProtocolClientLANVerifying
		return []CmdID{SessionProtocolCmdWriteActiveStream}, nil
	case m.State == SessionProtocolClientLANActive && ev == SessionProtocolEventAppSend:
		m.State = SessionProtocolClientLANActive
		return []CmdID{SessionProtocolCmdWriteActiveStream}, nil
	case m.State == SessionProtocolClientRelayFallback && ev == SessionProtocolEventAppSend:
		m.State = SessionProtocolClientRelayFallback
		return []CmdID{SessionProtocolCmdWriteActiveStream}, nil
	case m.State == SessionProtocolClientRelayConnected && ev == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolClientRelayConnected
		return []CmdID{SessionProtocolCmdDeliverRecv}, nil
	case m.State == SessionProtocolClientLANConnecting && ev == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolClientLANConnecting
		return []CmdID{SessionProtocolCmdDeliverRecv}, nil
	case m.State == SessionProtocolClientLANVerifying && ev == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolClientLANVerifying
		return []CmdID{SessionProtocolCmdDeliverRecv}, nil
	case m.State == SessionProtocolClientLANActive && ev == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolClientLANActive
		return []CmdID{SessionProtocolCmdDeliverRecv}, nil
	case m.State == SessionProtocolClientRelayFallback && ev == SessionProtocolEventRelayStreamData:
		m.State = SessionProtocolClientRelayFallback
		return []CmdID{SessionProtocolCmdDeliverRecv}, nil
	case m.State == SessionProtocolClientRelayConnected && ev == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolClientRelayConnected
		return []CmdID{SessionProtocolCmdDeliverRecvError}, nil
	case m.State == SessionProtocolClientLANConnecting && ev == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolClientLANConnecting
		return []CmdID{SessionProtocolCmdDeliverRecvError}, nil
	case m.State == SessionProtocolClientLANVerifying && ev == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolClientLANVerifying
		return []CmdID{SessionProtocolCmdDeliverRecvError}, nil
	case m.State == SessionProtocolClientLANActive && ev == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolClientLANActive
		return []CmdID{SessionProtocolCmdDeliverRecvError}, nil
	case m.State == SessionProtocolClientRelayFallback && ev == SessionProtocolEventRelayStreamError:
		m.State = SessionProtocolClientRelayFallback
		return []CmdID{SessionProtocolCmdDeliverRecvError}, nil
	case m.State == SessionProtocolClientRelayConnected && ev == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolClientRelayConnected
		return []CmdID{SessionProtocolCmdSendActiveDatagram}, nil
	case m.State == SessionProtocolClientLANConnecting && ev == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolClientLANConnecting
		return []CmdID{SessionProtocolCmdSendActiveDatagram}, nil
	case m.State == SessionProtocolClientLANVerifying && ev == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolClientLANVerifying
		return []CmdID{SessionProtocolCmdSendActiveDatagram}, nil
	case m.State == SessionProtocolClientLANActive && ev == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolClientLANActive
		return []CmdID{SessionProtocolCmdSendActiveDatagram}, nil
	case m.State == SessionProtocolClientRelayFallback && ev == SessionProtocolEventAppSendDatagram:
		m.State = SessionProtocolClientRelayFallback
		return []CmdID{SessionProtocolCmdSendActiveDatagram}, nil
	case m.State == SessionProtocolClientRelayConnected && ev == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolClientRelayConnected
		return []CmdID{SessionProtocolCmdDeliverRecvDatagram}, nil
	case m.State == SessionProtocolClientLANConnecting && ev == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolClientLANConnecting
		return []CmdID{SessionProtocolCmdDeliverRecvDatagram}, nil
	case m.State == SessionProtocolClientLANVerifying && ev == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolClientLANVerifying
		return []CmdID{SessionProtocolCmdDeliverRecvDatagram}, nil
	case m.State == SessionProtocolClientLANActive && ev == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolClientLANActive
		return []CmdID{SessionProtocolCmdDeliverRecvDatagram}, nil
	case m.State == SessionProtocolClientRelayFallback && ev == SessionProtocolEventRelayDatagram:
		m.State = SessionProtocolClientRelayFallback
		return []CmdID{SessionProtocolCmdDeliverRecvDatagram}, nil
	case m.State == SessionProtocolClientLANActive && ev == SessionProtocolEventLanStreamData:
		m.State = SessionProtocolClientLANActive
		return []CmdID{SessionProtocolCmdDeliverRecv}, nil
	case m.State == SessionProtocolClientLANActive && ev == SessionProtocolEventLanDatagram:
		m.State = SessionProtocolClientLANActive
		return []CmdID{SessionProtocolCmdDeliverRecvDatagram}, nil
	}
	return nil, nil
}

// SessionProtocolRelayMachine is the generated state machine for the relay actor.
type SessionProtocolRelayMachine struct {
	State State
	RelayBridge string // relay bridge state

	Guards  map[GuardID]func() bool
	Actions map[ActionID]func() error
	OnChange func(varName string)
}

func NewSessionProtocolRelayMachine() *SessionProtocolRelayMachine {
	return &SessionProtocolRelayMachine{
		State: SessionProtocolRelayIdle,
		RelayBridge: "idle",
		Guards:  make(map[GuardID]func() bool),
		Actions: make(map[ActionID]func() error),
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
			if err := fn(); err != nil { return false, err }
		}
		m.RelayBridge = "active"
		if m.OnChange != nil { m.OnChange("relay_bridge") }
		m.State = SessionProtocolRelayBridged
		return true, nil
	case m.State == SessionProtocolRelayBridged && event == SessionProtocolEventClientDisconnect:
		if fn := m.Actions[SessionProtocolActionUnbridge]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.RelayBridge = "idle"
		if m.OnChange != nil { m.OnChange("relay_bridge") }
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
			if err := fn(); err != nil { return nil, err }
		}
		m.RelayBridge = "active"
		if m.OnChange != nil { m.OnChange("relay_bridge") }
		m.State = SessionProtocolRelayBridged
		return nil, nil
	case m.State == SessionProtocolRelayBridged && ev == SessionProtocolEventClientDisconnect:
		if fn := m.Actions[SessionProtocolActionUnbridge]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.RelayBridge = "idle"
		if m.OnChange != nil { m.OnChange("relay_bridge") }
		m.State = SessionProtocolRelayBackendRegistered
		return nil, nil
	case m.State == SessionProtocolRelayBackendRegistered && ev == SessionProtocolEventBackendDisconnect:
		m.State = SessionProtocolRelayIdle
		return nil, nil
	}
	return nil, nil
}

