// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Code generated from protocol/*.yaml. DO NOT EDIT.

package protocol

import "github.com/arr-ai/frozen"

var _ frozen.Set[string] // suppress unused import

// Session backend states.
const (
	SessionBackendIdle State = "Idle"
	SessionBackendGenerateToken State = "GenerateToken"
	SessionBackendRegisterRelay State = "RegisterRelay"
	SessionBackendWaitingForClient State = "WaitingForClient"
	SessionBackendDeriveSecret State = "DeriveSecret"
	SessionBackendSendAck State = "SendAck"
	SessionBackendWaitingForCode State = "WaitingForCode"
	SessionBackendValidateCode State = "ValidateCode"
	SessionBackendStorePaired State = "StorePaired"
	SessionBackendPaired State = "Paired"
	SessionBackendAuthCheck State = "AuthCheck"
	SessionBackendSessionActive State = "SessionActive"
	SessionBackendRelayConnected State = "RelayConnected"
	SessionBackendLANOffered State = "LANOffered"
	SessionBackendLANActive State = "LANActive"
	SessionBackendRelayBackoff State = "RelayBackoff"
	SessionBackendLANDegraded State = "LANDegraded"
)

// Session client states.
const (
	SessionClientIdle State = "Idle"
	SessionClientObtainBackchannelSecret State = "ObtainBackchannelSecret"
	SessionClientConnectRelay State = "ConnectRelay"
	SessionClientGenKeyPair State = "GenKeyPair"
	SessionClientWaitAck State = "WaitAck"
	SessionClientE2EReady State = "E2EReady"
	SessionClientShowCode State = "ShowCode"
	SessionClientWaitPairComplete State = "WaitPairComplete"
	SessionClientPaired State = "Paired"
	SessionClientReconnect State = "Reconnect"
	SessionClientSendAuth State = "SendAuth"
	SessionClientSessionActive State = "SessionActive"
	SessionClientRelayConnected State = "RelayConnected"
	SessionClientLANConnecting State = "LANConnecting"
	SessionClientLANVerifying State = "LANVerifying"
	SessionClientLANActive State = "LANActive"
	SessionClientRelayFallback State = "RelayFallback"
)

// Session relay states.
const (
	SessionRelayIdle State = "Idle"
	SessionRelayBackendRegistered State = "BackendRegistered"
	SessionRelayBridged State = "Bridged"
)

// Session message types.
const (
	SessionMsgPairHello MsgType = "pair_hello"
	SessionMsgPairHelloAck MsgType = "pair_hello_ack"
	SessionMsgPairConfirm MsgType = "pair_confirm"
	SessionMsgPairComplete MsgType = "pair_complete"
	SessionMsgAuthRequest MsgType = "auth_request"
	SessionMsgAuthOk MsgType = "auth_ok"
	SessionMsgLanOffer MsgType = "lan_offer"
	SessionMsgLanVerify MsgType = "lan_verify"
	SessionMsgLanConfirm MsgType = "lan_confirm"
	SessionMsgPathPing MsgType = "path_ping"
	SessionMsgPathPong MsgType = "path_pong"
)

// Session guards.
const (
	SessionGuardTokenValid GuardID = "token_valid"
	SessionGuardTokenInvalid GuardID = "token_invalid"
	SessionGuardCodeCorrect GuardID = "code_correct"
	SessionGuardCodeWrong GuardID = "code_wrong"
	SessionGuardDeviceKnown GuardID = "device_known"
	SessionGuardDeviceUnknown GuardID = "device_unknown"
	SessionGuardNonceFresh GuardID = "nonce_fresh"
	SessionGuardChallengeValid GuardID = "challenge_valid"
	SessionGuardChallengeInvalid GuardID = "challenge_invalid"
	SessionGuardLanEnabled GuardID = "lan_enabled"
	SessionGuardLanDisabled GuardID = "lan_disabled"
	SessionGuardLanServerAvailable GuardID = "lan_server_available"
	SessionGuardUnderMaxFailures GuardID = "under_max_failures"
	SessionGuardAtMaxFailures GuardID = "at_max_failures"
)

// Session actions.
const (
	SessionActionActivateLan ActionID = "activate_lan"
	SessionActionBridgeStreams ActionID = "bridge_streams"
	SessionActionDeriveSecret ActionID = "derive_secret"
	SessionActionDialLan ActionID = "dial_lan"
	SessionActionFallbackToRelay ActionID = "fallback_to_relay"
	SessionActionGenerateToken ActionID = "generate_token"
	SessionActionRegisterRelay ActionID = "register_relay"
	SessionActionResetFailures ActionID = "reset_failures"
	SessionActionSendPairHello ActionID = "send_pair_hello"
	SessionActionStoreDevice ActionID = "store_device"
	SessionActionStoreSecret ActionID = "store_secret"
	SessionActionUnbridge ActionID = "unbridge"
	SessionActionVerifyDevice ActionID = "verify_device"
)

// Session events.
const (
	SessionEventAppClose EventID = "app_close"
	SessionEventAppForceFallback EventID = "app_force_fallback"
	SessionEventAppLaunch EventID = "app_launch"
	SessionEventAppRecv EventID = "app_recv"
	SessionEventAppRecvDatagram EventID = "app_recv_datagram"
	SessionEventAppSend EventID = "app_send"
	SessionEventAppSendDatagram EventID = "app_send_datagram"
	SessionEventBackchannelReceived EventID = "backchannel_received"
	SessionEventBackendDisconnect EventID = "backend_disconnect"
	SessionEventBackendRegister EventID = "backend_register"
	SessionEventBackoffExpired EventID = "backoff_expired"
	SessionEventCheckCode EventID = "check_code"
	SessionEventCliCodeEntered EventID = "cli_code_entered"
	SessionEventCliInitPair EventID = "cli_init_pair"
	SessionEventClientConnect EventID = "client_connect"
	SessionEventClientDisconnect EventID = "client_disconnect"
	SessionEventCodeDisplayed EventID = "code_displayed"
	SessionEventDisconnect EventID = "disconnect"
	SessionEventEcdhComplete EventID = "ecdh_complete"
	SessionEventFinalise EventID = "finalise"
	SessionEventKeyPairGenerated EventID = "key_pair_generated"
	SessionEventLanDatagram EventID = "lan_datagram"
	SessionEventLanDialFailed EventID = "lan_dial_failed"
	SessionEventLanDialOk EventID = "lan_dial_ok"
	SessionEventLanError EventID = "lan_error"
	SessionEventLanServerChanged EventID = "lan_server_changed"
	SessionEventLanServerReady EventID = "lan_server_ready"
	SessionEventLanStreamData EventID = "lan_stream_data"
	SessionEventLanStreamError EventID = "lan_stream_error"
	SessionEventLanVerifyOk EventID = "lan_verify_ok"
	SessionEventOfferTimeout EventID = "offer_timeout"
	SessionEventPingTick EventID = "ping_tick"
	SessionEventPingTimeout EventID = "ping_timeout"
	SessionEventReadvertiseTick EventID = "readvertise_tick"
	SessionEventRecvAuthOk EventID = "recv_auth_ok"
	SessionEventRecvAuthRequest EventID = "recv_auth_request"
	SessionEventRecvLanConfirm EventID = "recv_lan_confirm"
	SessionEventRecvLanOffer EventID = "recv_lan_offer"
	SessionEventRecvLanVerify EventID = "recv_lan_verify"
	SessionEventRecvPairComplete EventID = "recv_pair_complete"
	SessionEventRecvPairConfirm EventID = "recv_pair_confirm"
	SessionEventRecvPairHello EventID = "recv_pair_hello"
	SessionEventRecvPairHelloAck EventID = "recv_pair_hello_ack"
	SessionEventRecvPathPing EventID = "recv_path_ping"
	SessionEventRecvPathPong EventID = "recv_path_pong"
	SessionEventRelayConnected EventID = "relay_connected"
	SessionEventRelayDatagram EventID = "relay_datagram"
	SessionEventRelayOk EventID = "relay_ok"
	SessionEventRelayRegistered EventID = "relay_registered"
	SessionEventRelayStreamData EventID = "relay_stream_data"
	SessionEventRelayStreamError EventID = "relay_stream_error"
	SessionEventSecretParsed EventID = "secret_parsed"
	SessionEventSessionEstablished EventID = "session_established"
	SessionEventSignalCodeDisplay EventID = "signal_code_display"
	SessionEventTokenCreated EventID = "token_created"
	SessionEventVerify EventID = "verify"
	SessionEventVerifyTimeout EventID = "verify_timeout"
)

// Session commands.
const (
	SessionCmdWriteActiveStream CmdID = "write_active_stream"
	SessionCmdSendActiveDatagram CmdID = "send_active_datagram"
	SessionCmdSendPathPing CmdID = "send_path_ping"
	SessionCmdSendPathPong CmdID = "send_path_pong"
	SessionCmdSendLanOffer CmdID = "send_lan_offer"
	SessionCmdSendLanVerify CmdID = "send_lan_verify"
	SessionCmdSendLanConfirm CmdID = "send_lan_confirm"
	SessionCmdDialLan CmdID = "dial_lan"
	SessionCmdDeliverRecv CmdID = "deliver_recv"
	SessionCmdDeliverRecvError CmdID = "deliver_recv_error"
	SessionCmdDeliverRecvDatagram CmdID = "deliver_recv_datagram"
	SessionCmdStartLanStreamReader CmdID = "start_lan_stream_reader"
	SessionCmdStopLanStreamReader CmdID = "stop_lan_stream_reader"
	SessionCmdStartLanDgReader CmdID = "start_lan_dg_reader"
	SessionCmdStopLanDgReader CmdID = "stop_lan_dg_reader"
	SessionCmdStartMonitor CmdID = "start_monitor"
	SessionCmdStopMonitor CmdID = "stop_monitor"
	SessionCmdStartPongTimeout CmdID = "start_pong_timeout"
	SessionCmdCancelPongTimeout CmdID = "cancel_pong_timeout"
	SessionCmdStartBackoffTimer CmdID = "start_backoff_timer"
	SessionCmdCloseLanPath CmdID = "close_lan_path"
	SessionCmdSignalLanReady CmdID = "signal_lan_ready"
	SessionCmdResetLanReady CmdID = "reset_lan_ready"
	SessionCmdSetCryptoDatagram CmdID = "set_crypto_datagram"
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

func Session() *Protocol {
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

// SessionBackendMachine is the generated state machine for the backend actor.
type SessionBackendMachine struct {
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

func NewSessionBackendMachine() *SessionBackendMachine {
	return &SessionBackendMachine{
		State: SessionBackendIdle,
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

func (m *SessionBackendMachine) HandleMessage(msg MsgType) (bool, error) {
	switch {
	case m.State == SessionBackendWaitingForClient && msg == SessionMsgPairHello && m.Guards[SessionGuardTokenValid] != nil && m.Guards[SessionGuardTokenValid]():
		if fn := m.Actions[SessionActionDeriveSecret]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		// received_client_pub: recv_msg.pubkey (set by action)
		m.BackendEcdhPub = "backend_pub"
		if m.OnChange != nil { m.OnChange("backend_ecdh_pub") }
		// backend_shared_key: DeriveKey("backend_pub", recv_msg.pubkey) (set by action)
		// backend_code: DeriveCode("backend_pub", recv_msg.pubkey) (set by action)
		m.State = SessionBackendDeriveSecret
		return true, nil
	case m.State == SessionBackendWaitingForClient && msg == SessionMsgPairHello && m.Guards[SessionGuardTokenInvalid] != nil && m.Guards[SessionGuardTokenInvalid]():
		m.State = SessionBackendIdle
		return true, nil
	case m.State == SessionBackendPaired && msg == SessionMsgAuthRequest:
		// received_device_id: recv_msg.device_id (set by action)
		// received_auth_nonce: recv_msg.nonce (set by action)
		m.State = SessionBackendAuthCheck
		return true, nil
	case m.State == SessionBackendLANOffered && msg == SessionMsgLanVerify && m.Guards[SessionGuardChallengeValid] != nil && m.Guards[SessionGuardChallengeValid]():
		if fn := m.Actions[SessionActionActivateLan]; fn != nil {
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
		m.State = SessionBackendLANActive
		return true, nil
	case m.State == SessionBackendLANOffered && msg == SessionMsgLanVerify && m.Guards[SessionGuardChallengeInvalid] != nil && m.Guards[SessionGuardChallengeInvalid]():
		m.State = SessionBackendRelayConnected
		return true, nil
	case m.State == SessionBackendLANDegraded && msg == SessionMsgPathPong:
		if fn := m.Actions[SessionActionResetFailures]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.PingFailures = 0
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = SessionBackendLANActive
		return true, nil
	}
	return false, nil
}

func (m *SessionBackendMachine) Step(event EventID) (bool, error) {
	switch {
	case m.State == SessionBackendIdle && event == SessionEventCliInitPair:
		if fn := m.Actions[SessionActionGenerateToken]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.CurrentToken = "tok_1"
		if m.OnChange != nil { m.OnChange("current_token") }
		// active_tokens: active_tokens \union {"tok_1"} (set by action)
		m.State = SessionBackendGenerateToken
		return true, nil
	case m.State == SessionBackendGenerateToken && event == SessionEventTokenCreated:
		if fn := m.Actions[SessionActionRegisterRelay]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.State = SessionBackendRegisterRelay
		return true, nil
	case m.State == SessionBackendRegisterRelay && event == SessionEventRelayRegistered:
		m.SecretPublished = true
		if m.OnChange != nil { m.OnChange("secret_published") }
		m.State = SessionBackendWaitingForClient
		return true, nil
	case m.State == SessionBackendDeriveSecret && event == SessionEventEcdhComplete:
		m.State = SessionBackendSendAck
		return true, nil
	case m.State == SessionBackendSendAck && event == SessionEventSignalCodeDisplay:
		m.State = SessionBackendWaitingForCode
		return true, nil
	case m.State == SessionBackendWaitingForCode && event == SessionEventCliCodeEntered:
		// received_code: cli_entered_code (set by action)
		m.State = SessionBackendValidateCode
		return true, nil
	case m.State == SessionBackendValidateCode && event == SessionEventCheckCode && m.Guards[SessionGuardCodeCorrect] != nil && m.Guards[SessionGuardCodeCorrect]():
		m.State = SessionBackendStorePaired
		return true, nil
	case m.State == SessionBackendValidateCode && event == SessionEventCheckCode && m.Guards[SessionGuardCodeWrong] != nil && m.Guards[SessionGuardCodeWrong]():
		m.CodeAttempts = m.CodeAttempts + 1
		if m.OnChange != nil { m.OnChange("code_attempts") }
		m.State = SessionBackendIdle
		return true, nil
	case m.State == SessionBackendStorePaired && event == SessionEventFinalise:
		if fn := m.Actions[SessionActionStoreDevice]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.DeviceSecret = "dev_secret_1"
		if m.OnChange != nil { m.OnChange("device_secret") }
		// paired_devices: paired_devices \union {"device_1"} (set by action)
		// active_tokens: active_tokens \ {current_token} (set by action)
		// used_tokens: used_tokens \union {current_token} (set by action)
		m.State = SessionBackendPaired
		return true, nil
	case m.State == SessionBackendAuthCheck && event == SessionEventVerify && m.Guards[SessionGuardDeviceKnown] != nil && m.Guards[SessionGuardDeviceKnown]():
		if fn := m.Actions[SessionActionVerifyDevice]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		// auth_nonces_used: auth_nonces_used \union {received_auth_nonce} (set by action)
		m.State = SessionBackendSessionActive
		return true, nil
	case m.State == SessionBackendAuthCheck && event == SessionEventVerify && m.Guards[SessionGuardDeviceUnknown] != nil && m.Guards[SessionGuardDeviceUnknown]():
		m.State = SessionBackendIdle
		return true, nil
	case m.State == SessionBackendSessionActive && event == SessionEventSessionEstablished:
		m.State = SessionBackendRelayConnected
		return true, nil
	case m.State == SessionBackendRelayConnected && event == SessionEventLanServerReady:
		m.State = SessionBackendLANOffered
		return true, nil
	case m.State == SessionBackendLANOffered && event == SessionEventOfferTimeout:
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = SessionBackendRelayBackoff
		return true, nil
	case m.State == SessionBackendLANActive && event == SessionEventPingTick:
		m.State = SessionBackendLANActive
		return true, nil
	case m.State == SessionBackendLANActive && event == SessionEventPingTimeout:
		m.PingFailures = 1
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = SessionBackendLANDegraded
		return true, nil
	case m.State == SessionBackendLANDegraded && event == SessionEventPingTick:
		m.State = SessionBackendLANDegraded
		return true, nil
	case m.State == SessionBackendLANActive && event == SessionEventLanStreamError:
		if fn := m.Actions[SessionActionFallbackToRelay]; fn != nil {
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
		m.State = SessionBackendRelayBackoff
		return true, nil
	case m.State == SessionBackendLANDegraded && event == SessionEventLanStreamError:
		if fn := m.Actions[SessionActionFallbackToRelay]; fn != nil {
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
		m.State = SessionBackendRelayBackoff
		return true, nil
	case m.State == SessionBackendLANDegraded && event == SessionEventPingTimeout && m.Guards[SessionGuardUnderMaxFailures] != nil && m.Guards[SessionGuardUnderMaxFailures]():
		m.PingFailures = m.PingFailures + 1
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = SessionBackendLANDegraded
		return true, nil
	case m.State == SessionBackendLANDegraded && event == SessionEventPingTimeout && m.Guards[SessionGuardAtMaxFailures] != nil && m.Guards[SessionGuardAtMaxFailures]():
		if fn := m.Actions[SessionActionFallbackToRelay]; fn != nil {
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
		m.State = SessionBackendRelayBackoff
		return true, nil
	case m.State == SessionBackendRelayBackoff && event == SessionEventBackoffExpired:
		m.State = SessionBackendLANOffered
		return true, nil
	case m.State == SessionBackendRelayBackoff && event == SessionEventLanServerChanged:
		m.BackoffLevel = 0
		if m.OnChange != nil { m.OnChange("backoff_level") }
		m.State = SessionBackendLANOffered
		return true, nil
	case m.State == SessionBackendRelayConnected && event == SessionEventReadvertiseTick && m.Guards[SessionGuardLanServerAvailable] != nil && m.Guards[SessionGuardLanServerAvailable]():
		m.State = SessionBackendLANOffered
		return true, nil
	case m.State == SessionBackendLANOffered && event == SessionEventAppForceFallback:
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = SessionBackendRelayConnected
		return true, nil
	case m.State == SessionBackendLANActive && event == SessionEventAppForceFallback:
		if fn := m.Actions[SessionActionFallbackToRelay]; fn != nil {
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
		m.State = SessionBackendRelayBackoff
		return true, nil
	case m.State == SessionBackendLANDegraded && event == SessionEventAppForceFallback:
		if fn := m.Actions[SessionActionFallbackToRelay]; fn != nil {
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
		m.State = SessionBackendRelayBackoff
		return true, nil
	case m.State == SessionBackendRelayConnected && event == SessionEventDisconnect:
		m.State = SessionBackendPaired
		return true, nil
	case m.State == SessionBackendRelayConnected && event == SessionEventAppSend:
		m.State = SessionBackendRelayConnected
		return true, nil
	case m.State == SessionBackendLANOffered && event == SessionEventAppSend:
		m.State = SessionBackendLANOffered
		return true, nil
	case m.State == SessionBackendLANActive && event == SessionEventAppSend:
		m.State = SessionBackendLANActive
		return true, nil
	case m.State == SessionBackendLANDegraded && event == SessionEventAppSend:
		m.State = SessionBackendLANDegraded
		return true, nil
	case m.State == SessionBackendRelayBackoff && event == SessionEventAppSend:
		m.State = SessionBackendRelayBackoff
		return true, nil
	case m.State == SessionBackendRelayConnected && event == SessionEventRelayStreamData:
		m.State = SessionBackendRelayConnected
		return true, nil
	case m.State == SessionBackendLANOffered && event == SessionEventRelayStreamData:
		m.State = SessionBackendLANOffered
		return true, nil
	case m.State == SessionBackendLANActive && event == SessionEventRelayStreamData:
		m.State = SessionBackendLANActive
		return true, nil
	case m.State == SessionBackendLANDegraded && event == SessionEventRelayStreamData:
		m.State = SessionBackendLANDegraded
		return true, nil
	case m.State == SessionBackendRelayBackoff && event == SessionEventRelayStreamData:
		m.State = SessionBackendRelayBackoff
		return true, nil
	case m.State == SessionBackendRelayConnected && event == SessionEventRelayStreamError:
		m.State = SessionBackendRelayConnected
		return true, nil
	case m.State == SessionBackendLANOffered && event == SessionEventRelayStreamError:
		m.State = SessionBackendLANOffered
		return true, nil
	case m.State == SessionBackendLANActive && event == SessionEventRelayStreamError:
		m.State = SessionBackendLANActive
		return true, nil
	case m.State == SessionBackendLANDegraded && event == SessionEventRelayStreamError:
		m.State = SessionBackendLANDegraded
		return true, nil
	case m.State == SessionBackendRelayBackoff && event == SessionEventRelayStreamError:
		m.State = SessionBackendRelayBackoff
		return true, nil
	case m.State == SessionBackendRelayConnected && event == SessionEventAppSendDatagram:
		m.State = SessionBackendRelayConnected
		return true, nil
	case m.State == SessionBackendLANOffered && event == SessionEventAppSendDatagram:
		m.State = SessionBackendLANOffered
		return true, nil
	case m.State == SessionBackendLANActive && event == SessionEventAppSendDatagram:
		m.State = SessionBackendLANActive
		return true, nil
	case m.State == SessionBackendLANDegraded && event == SessionEventAppSendDatagram:
		m.State = SessionBackendLANDegraded
		return true, nil
	case m.State == SessionBackendRelayBackoff && event == SessionEventAppSendDatagram:
		m.State = SessionBackendRelayBackoff
		return true, nil
	case m.State == SessionBackendRelayConnected && event == SessionEventRelayDatagram:
		m.State = SessionBackendRelayConnected
		return true, nil
	case m.State == SessionBackendLANOffered && event == SessionEventRelayDatagram:
		m.State = SessionBackendLANOffered
		return true, nil
	case m.State == SessionBackendLANActive && event == SessionEventRelayDatagram:
		m.State = SessionBackendLANActive
		return true, nil
	case m.State == SessionBackendLANDegraded && event == SessionEventRelayDatagram:
		m.State = SessionBackendLANDegraded
		return true, nil
	case m.State == SessionBackendRelayBackoff && event == SessionEventRelayDatagram:
		m.State = SessionBackendRelayBackoff
		return true, nil
	case m.State == SessionBackendLANActive && event == SessionEventLanStreamData:
		m.State = SessionBackendLANActive
		return true, nil
	case m.State == SessionBackendLANDegraded && event == SessionEventLanStreamData:
		m.State = SessionBackendLANDegraded
		return true, nil
	case m.State == SessionBackendLANActive && event == SessionEventLanDatagram:
		m.State = SessionBackendLANActive
		return true, nil
	case m.State == SessionBackendLANDegraded && event == SessionEventLanDatagram:
		m.State = SessionBackendLANDegraded
		return true, nil
	}
	return false, nil
}

func (m *SessionBackendMachine) HandleEvent(ev EventID) ([]CmdID, error) {
	switch {
	case m.State == SessionBackendIdle && ev == SessionEventCliInitPair:
		if fn := m.Actions[SessionActionGenerateToken]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.CurrentToken = "tok_1"
		if m.OnChange != nil { m.OnChange("current_token") }
		// active_tokens: active_tokens \union {"tok_1"} (set by action)
		m.State = SessionBackendGenerateToken
		return nil, nil
	case m.State == SessionBackendGenerateToken && ev == SessionEventTokenCreated:
		if fn := m.Actions[SessionActionRegisterRelay]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.State = SessionBackendRegisterRelay
		return nil, nil
	case m.State == SessionBackendRegisterRelay && ev == SessionEventRelayRegistered:
		m.SecretPublished = true
		if m.OnChange != nil { m.OnChange("secret_published") }
		m.State = SessionBackendWaitingForClient
		return nil, nil
	case m.State == SessionBackendWaitingForClient && ev == SessionEventRecvPairHello && m.Guards[SessionGuardTokenValid] != nil && m.Guards[SessionGuardTokenValid]():
		if fn := m.Actions[SessionActionDeriveSecret]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		// received_client_pub: recv_msg.pubkey (set by action)
		m.BackendEcdhPub = "backend_pub"
		if m.OnChange != nil { m.OnChange("backend_ecdh_pub") }
		// backend_shared_key: DeriveKey("backend_pub", recv_msg.pubkey) (set by action)
		// backend_code: DeriveCode("backend_pub", recv_msg.pubkey) (set by action)
		m.State = SessionBackendDeriveSecret
		return nil, nil
	case m.State == SessionBackendWaitingForClient && ev == SessionEventRecvPairHello && m.Guards[SessionGuardTokenInvalid] != nil && m.Guards[SessionGuardTokenInvalid]():
		m.State = SessionBackendIdle
		return nil, nil
	case m.State == SessionBackendDeriveSecret && ev == SessionEventEcdhComplete:
		m.State = SessionBackendSendAck
		return nil, nil
	case m.State == SessionBackendSendAck && ev == SessionEventSignalCodeDisplay:
		m.State = SessionBackendWaitingForCode
		return nil, nil
	case m.State == SessionBackendWaitingForCode && ev == SessionEventCliCodeEntered:
		// received_code: cli_entered_code (set by action)
		m.State = SessionBackendValidateCode
		return nil, nil
	case m.State == SessionBackendValidateCode && ev == SessionEventCheckCode && m.Guards[SessionGuardCodeCorrect] != nil && m.Guards[SessionGuardCodeCorrect]():
		m.State = SessionBackendStorePaired
		return nil, nil
	case m.State == SessionBackendValidateCode && ev == SessionEventCheckCode && m.Guards[SessionGuardCodeWrong] != nil && m.Guards[SessionGuardCodeWrong]():
		m.CodeAttempts = m.CodeAttempts + 1
		if m.OnChange != nil { m.OnChange("code_attempts") }
		m.State = SessionBackendIdle
		return nil, nil
	case m.State == SessionBackendStorePaired && ev == SessionEventFinalise:
		if fn := m.Actions[SessionActionStoreDevice]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.DeviceSecret = "dev_secret_1"
		if m.OnChange != nil { m.OnChange("device_secret") }
		// paired_devices: paired_devices \union {"device_1"} (set by action)
		// active_tokens: active_tokens \ {current_token} (set by action)
		// used_tokens: used_tokens \union {current_token} (set by action)
		m.State = SessionBackendPaired
		return nil, nil
	case m.State == SessionBackendPaired && ev == SessionEventRecvAuthRequest:
		// received_device_id: recv_msg.device_id (set by action)
		// received_auth_nonce: recv_msg.nonce (set by action)
		m.State = SessionBackendAuthCheck
		return nil, nil
	case m.State == SessionBackendAuthCheck && ev == SessionEventVerify && m.Guards[SessionGuardDeviceKnown] != nil && m.Guards[SessionGuardDeviceKnown]():
		if fn := m.Actions[SessionActionVerifyDevice]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		// auth_nonces_used: auth_nonces_used \union {received_auth_nonce} (set by action)
		m.State = SessionBackendSessionActive
		return nil, nil
	case m.State == SessionBackendAuthCheck && ev == SessionEventVerify && m.Guards[SessionGuardDeviceUnknown] != nil && m.Guards[SessionGuardDeviceUnknown]():
		m.State = SessionBackendIdle
		return nil, nil
	case m.State == SessionBackendSessionActive && ev == SessionEventSessionEstablished:
		m.State = SessionBackendRelayConnected
		return nil, nil
	case m.State == SessionBackendRelayConnected && ev == SessionEventLanServerReady:
		m.State = SessionBackendLANOffered
		return []CmdID{SessionCmdSendLanOffer}, nil
	case m.State == SessionBackendLANOffered && ev == SessionEventRecvLanVerify && m.Guards[SessionGuardChallengeValid] != nil && m.Guards[SessionGuardChallengeValid]():
		if fn := m.Actions[SessionActionActivateLan]; fn != nil {
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
		m.State = SessionBackendLANActive
		return []CmdID{SessionCmdSendLanConfirm, SessionCmdStartLanStreamReader, SessionCmdStartLanDgReader, SessionCmdStartMonitor, SessionCmdSignalLanReady, SessionCmdSetCryptoDatagram}, nil
	case m.State == SessionBackendLANOffered && ev == SessionEventRecvLanVerify && m.Guards[SessionGuardChallengeInvalid] != nil && m.Guards[SessionGuardChallengeInvalid]():
		m.State = SessionBackendRelayConnected
		return nil, nil
	case m.State == SessionBackendLANOffered && ev == SessionEventOfferTimeout:
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = SessionBackendRelayBackoff
		return []CmdID{SessionCmdResetLanReady, SessionCmdStartBackoffTimer}, nil
	case m.State == SessionBackendLANActive && ev == SessionEventPingTick:
		m.State = SessionBackendLANActive
		return []CmdID{SessionCmdSendPathPing, SessionCmdStartPongTimeout}, nil
	case m.State == SessionBackendLANActive && ev == SessionEventPingTimeout:
		m.PingFailures = 1
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = SessionBackendLANDegraded
		return nil, nil
	case m.State == SessionBackendLANDegraded && ev == SessionEventPingTick:
		m.State = SessionBackendLANDegraded
		return []CmdID{SessionCmdSendPathPing, SessionCmdStartPongTimeout}, nil
	case m.State == SessionBackendLANActive && ev == SessionEventLanStreamError:
		if fn := m.Actions[SessionActionFallbackToRelay]; fn != nil {
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
		m.State = SessionBackendRelayBackoff
		return []CmdID{SessionCmdStopMonitor, SessionCmdStopLanStreamReader, SessionCmdStopLanDgReader, SessionCmdCloseLanPath, SessionCmdResetLanReady, SessionCmdStartBackoffTimer}, nil
	case m.State == SessionBackendLANDegraded && ev == SessionEventLanStreamError:
		if fn := m.Actions[SessionActionFallbackToRelay]; fn != nil {
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
		m.State = SessionBackendRelayBackoff
		return []CmdID{SessionCmdStopMonitor, SessionCmdStopLanStreamReader, SessionCmdStopLanDgReader, SessionCmdCloseLanPath, SessionCmdResetLanReady, SessionCmdStartBackoffTimer}, nil
	case m.State == SessionBackendLANDegraded && ev == SessionEventRecvPathPong:
		if fn := m.Actions[SessionActionResetFailures]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.PingFailures = 0
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = SessionBackendLANActive
		return []CmdID{SessionCmdCancelPongTimeout}, nil
	case m.State == SessionBackendLANDegraded && ev == SessionEventPingTimeout && m.Guards[SessionGuardUnderMaxFailures] != nil && m.Guards[SessionGuardUnderMaxFailures]():
		m.PingFailures = m.PingFailures + 1
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = SessionBackendLANDegraded
		return nil, nil
	case m.State == SessionBackendLANDegraded && ev == SessionEventPingTimeout && m.Guards[SessionGuardAtMaxFailures] != nil && m.Guards[SessionGuardAtMaxFailures]():
		if fn := m.Actions[SessionActionFallbackToRelay]; fn != nil {
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
		m.State = SessionBackendRelayBackoff
		return []CmdID{SessionCmdStopMonitor, SessionCmdStopLanStreamReader, SessionCmdStopLanDgReader, SessionCmdCloseLanPath, SessionCmdResetLanReady, SessionCmdStartBackoffTimer}, nil
	case m.State == SessionBackendRelayBackoff && ev == SessionEventBackoffExpired:
		m.State = SessionBackendLANOffered
		return []CmdID{SessionCmdSendLanOffer}, nil
	case m.State == SessionBackendRelayBackoff && ev == SessionEventLanServerChanged:
		m.BackoffLevel = 0
		if m.OnChange != nil { m.OnChange("backoff_level") }
		m.State = SessionBackendLANOffered
		return []CmdID{SessionCmdSendLanOffer}, nil
	case m.State == SessionBackendRelayConnected && ev == SessionEventReadvertiseTick && m.Guards[SessionGuardLanServerAvailable] != nil && m.Guards[SessionGuardLanServerAvailable]():
		m.State = SessionBackendLANOffered
		return []CmdID{SessionCmdSendLanOffer}, nil
	case m.State == SessionBackendLANOffered && ev == SessionEventAppForceFallback:
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = SessionBackendRelayConnected
		return []CmdID{SessionCmdResetLanReady}, nil
	case m.State == SessionBackendLANActive && ev == SessionEventAppForceFallback:
		if fn := m.Actions[SessionActionFallbackToRelay]; fn != nil {
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
		m.State = SessionBackendRelayBackoff
		return []CmdID{SessionCmdStopMonitor, SessionCmdCancelPongTimeout, SessionCmdStopLanStreamReader, SessionCmdStopLanDgReader, SessionCmdCloseLanPath, SessionCmdResetLanReady, SessionCmdStartBackoffTimer}, nil
	case m.State == SessionBackendLANDegraded && ev == SessionEventAppForceFallback:
		if fn := m.Actions[SessionActionFallbackToRelay]; fn != nil {
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
		m.State = SessionBackendRelayBackoff
		return []CmdID{SessionCmdStopMonitor, SessionCmdCancelPongTimeout, SessionCmdStopLanStreamReader, SessionCmdStopLanDgReader, SessionCmdCloseLanPath, SessionCmdResetLanReady, SessionCmdStartBackoffTimer}, nil
	case m.State == SessionBackendRelayConnected && ev == SessionEventDisconnect:
		m.State = SessionBackendPaired
		return nil, nil
	case m.State == SessionBackendRelayConnected && ev == SessionEventAppSend:
		m.State = SessionBackendRelayConnected
		return []CmdID{SessionCmdWriteActiveStream}, nil
	case m.State == SessionBackendLANOffered && ev == SessionEventAppSend:
		m.State = SessionBackendLANOffered
		return []CmdID{SessionCmdWriteActiveStream}, nil
	case m.State == SessionBackendLANActive && ev == SessionEventAppSend:
		m.State = SessionBackendLANActive
		return []CmdID{SessionCmdWriteActiveStream}, nil
	case m.State == SessionBackendLANDegraded && ev == SessionEventAppSend:
		m.State = SessionBackendLANDegraded
		return []CmdID{SessionCmdWriteActiveStream}, nil
	case m.State == SessionBackendRelayBackoff && ev == SessionEventAppSend:
		m.State = SessionBackendRelayBackoff
		return []CmdID{SessionCmdWriteActiveStream}, nil
	case m.State == SessionBackendRelayConnected && ev == SessionEventRelayStreamData:
		m.State = SessionBackendRelayConnected
		return []CmdID{SessionCmdDeliverRecv}, nil
	case m.State == SessionBackendLANOffered && ev == SessionEventRelayStreamData:
		m.State = SessionBackendLANOffered
		return []CmdID{SessionCmdDeliverRecv}, nil
	case m.State == SessionBackendLANActive && ev == SessionEventRelayStreamData:
		m.State = SessionBackendLANActive
		return []CmdID{SessionCmdDeliverRecv}, nil
	case m.State == SessionBackendLANDegraded && ev == SessionEventRelayStreamData:
		m.State = SessionBackendLANDegraded
		return []CmdID{SessionCmdDeliverRecv}, nil
	case m.State == SessionBackendRelayBackoff && ev == SessionEventRelayStreamData:
		m.State = SessionBackendRelayBackoff
		return []CmdID{SessionCmdDeliverRecv}, nil
	case m.State == SessionBackendRelayConnected && ev == SessionEventRelayStreamError:
		m.State = SessionBackendRelayConnected
		return []CmdID{SessionCmdDeliverRecvError}, nil
	case m.State == SessionBackendLANOffered && ev == SessionEventRelayStreamError:
		m.State = SessionBackendLANOffered
		return []CmdID{SessionCmdDeliverRecvError}, nil
	case m.State == SessionBackendLANActive && ev == SessionEventRelayStreamError:
		m.State = SessionBackendLANActive
		return []CmdID{SessionCmdDeliverRecvError}, nil
	case m.State == SessionBackendLANDegraded && ev == SessionEventRelayStreamError:
		m.State = SessionBackendLANDegraded
		return []CmdID{SessionCmdDeliverRecvError}, nil
	case m.State == SessionBackendRelayBackoff && ev == SessionEventRelayStreamError:
		m.State = SessionBackendRelayBackoff
		return []CmdID{SessionCmdDeliverRecvError}, nil
	case m.State == SessionBackendRelayConnected && ev == SessionEventAppSendDatagram:
		m.State = SessionBackendRelayConnected
		return []CmdID{SessionCmdSendActiveDatagram}, nil
	case m.State == SessionBackendLANOffered && ev == SessionEventAppSendDatagram:
		m.State = SessionBackendLANOffered
		return []CmdID{SessionCmdSendActiveDatagram}, nil
	case m.State == SessionBackendLANActive && ev == SessionEventAppSendDatagram:
		m.State = SessionBackendLANActive
		return []CmdID{SessionCmdSendActiveDatagram}, nil
	case m.State == SessionBackendLANDegraded && ev == SessionEventAppSendDatagram:
		m.State = SessionBackendLANDegraded
		return []CmdID{SessionCmdSendActiveDatagram}, nil
	case m.State == SessionBackendRelayBackoff && ev == SessionEventAppSendDatagram:
		m.State = SessionBackendRelayBackoff
		return []CmdID{SessionCmdSendActiveDatagram}, nil
	case m.State == SessionBackendRelayConnected && ev == SessionEventRelayDatagram:
		m.State = SessionBackendRelayConnected
		return []CmdID{SessionCmdDeliverRecvDatagram}, nil
	case m.State == SessionBackendLANOffered && ev == SessionEventRelayDatagram:
		m.State = SessionBackendLANOffered
		return []CmdID{SessionCmdDeliverRecvDatagram}, nil
	case m.State == SessionBackendLANActive && ev == SessionEventRelayDatagram:
		m.State = SessionBackendLANActive
		return []CmdID{SessionCmdDeliverRecvDatagram}, nil
	case m.State == SessionBackendLANDegraded && ev == SessionEventRelayDatagram:
		m.State = SessionBackendLANDegraded
		return []CmdID{SessionCmdDeliverRecvDatagram}, nil
	case m.State == SessionBackendRelayBackoff && ev == SessionEventRelayDatagram:
		m.State = SessionBackendRelayBackoff
		return []CmdID{SessionCmdDeliverRecvDatagram}, nil
	case m.State == SessionBackendLANActive && ev == SessionEventLanStreamData:
		m.State = SessionBackendLANActive
		return []CmdID{SessionCmdDeliverRecv}, nil
	case m.State == SessionBackendLANDegraded && ev == SessionEventLanStreamData:
		m.State = SessionBackendLANDegraded
		return []CmdID{SessionCmdDeliverRecv}, nil
	case m.State == SessionBackendLANActive && ev == SessionEventLanDatagram:
		m.State = SessionBackendLANActive
		return []CmdID{SessionCmdDeliverRecvDatagram}, nil
	case m.State == SessionBackendLANDegraded && ev == SessionEventLanDatagram:
		m.State = SessionBackendLANDegraded
		return []CmdID{SessionCmdDeliverRecvDatagram}, nil
	}
	return nil, nil
}

// SessionClientMachine is the generated state machine for the client actor.
type SessionClientMachine struct {
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

func NewSessionClientMachine() *SessionClientMachine {
	return &SessionClientMachine{
		State: SessionClientIdle,
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

func (m *SessionClientMachine) HandleMessage(msg MsgType) (bool, error) {
	switch {
	case m.State == SessionClientWaitAck && msg == SessionMsgPairHelloAck:
		if fn := m.Actions[SessionActionDeriveSecret]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		// received_backend_pub: recv_msg.pubkey (set by action)
		// client_shared_key: DeriveKey("client_pub", recv_msg.pubkey) (set by action)
		m.State = SessionClientE2EReady
		return true, nil
	case m.State == SessionClientE2EReady && msg == SessionMsgPairConfirm:
		// client_code: DeriveCode(received_backend_pub, "client_pub") (set by action)
		m.State = SessionClientShowCode
		return true, nil
	case m.State == SessionClientWaitPairComplete && msg == SessionMsgPairComplete:
		if fn := m.Actions[SessionActionStoreSecret]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.State = SessionClientPaired
		return true, nil
	case m.State == SessionClientSendAuth && msg == SessionMsgAuthOk:
		m.State = SessionClientSessionActive
		return true, nil
	case m.State == SessionClientRelayConnected && msg == SessionMsgLanOffer && m.Guards[SessionGuardLanEnabled] != nil && m.Guards[SessionGuardLanEnabled]():
		if fn := m.Actions[SessionActionDialLan]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.State = SessionClientLANConnecting
		return true, nil
	case m.State == SessionClientRelayConnected && msg == SessionMsgLanOffer && m.Guards[SessionGuardLanDisabled] != nil && m.Guards[SessionGuardLanDisabled]():
		m.State = SessionClientRelayConnected
		return true, nil
	case m.State == SessionClientLANVerifying && msg == SessionMsgLanConfirm:
		if fn := m.Actions[SessionActionActivateLan]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.CActivePath = "lan"
		if m.OnChange != nil { m.OnChange("c_active_path") }
		m.CDispatcherPath = "lan"
		if m.OnChange != nil { m.OnChange("c_dispatcher_path") }
		m.LanSignal = "ready"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = SessionClientLANActive
		return true, nil
	case m.State == SessionClientLANActive && msg == SessionMsgPathPing:
		m.State = SessionClientLANActive
		return true, nil
	case m.State == SessionClientLANActive && msg == SessionMsgLanOffer && m.Guards[SessionGuardLanEnabled] != nil && m.Guards[SessionGuardLanEnabled]():
		if fn := m.Actions[SessionActionDialLan]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.State = SessionClientLANConnecting
		return true, nil
	}
	return false, nil
}

func (m *SessionClientMachine) Step(event EventID) (bool, error) {
	switch {
	case m.State == SessionClientIdle && event == SessionEventBackchannelReceived:
		m.State = SessionClientObtainBackchannelSecret
		return true, nil
	case m.State == SessionClientObtainBackchannelSecret && event == SessionEventSecretParsed:
		m.State = SessionClientConnectRelay
		return true, nil
	case m.State == SessionClientConnectRelay && event == SessionEventRelayConnected:
		m.State = SessionClientGenKeyPair
		return true, nil
	case m.State == SessionClientGenKeyPair && event == SessionEventKeyPairGenerated:
		if fn := m.Actions[SessionActionSendPairHello]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.State = SessionClientWaitAck
		return true, nil
	case m.State == SessionClientShowCode && event == SessionEventCodeDisplayed:
		m.State = SessionClientWaitPairComplete
		return true, nil
	case m.State == SessionClientPaired && event == SessionEventAppLaunch:
		m.State = SessionClientReconnect
		return true, nil
	case m.State == SessionClientReconnect && event == SessionEventRelayConnected:
		m.State = SessionClientSendAuth
		return true, nil
	case m.State == SessionClientSessionActive && event == SessionEventSessionEstablished:
		m.State = SessionClientRelayConnected
		return true, nil
	case m.State == SessionClientLANConnecting && event == SessionEventLanDialOk:
		m.State = SessionClientLANVerifying
		return true, nil
	case m.State == SessionClientLANConnecting && event == SessionEventLanDialFailed:
		m.State = SessionClientRelayConnected
		return true, nil
	case m.State == SessionClientLANVerifying && event == SessionEventVerifyTimeout:
		m.CDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("c_dispatcher_path") }
		m.State = SessionClientRelayConnected
		return true, nil
	case m.State == SessionClientLANActive && event == SessionEventLanError:
		if fn := m.Actions[SessionActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.CActivePath = "relay"
		if m.OnChange != nil { m.OnChange("c_active_path") }
		m.CDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("c_dispatcher_path") }
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = SessionClientRelayFallback
		return true, nil
	case m.State == SessionClientLANActive && event == SessionEventLanStreamError:
		if fn := m.Actions[SessionActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.CActivePath = "relay"
		if m.OnChange != nil { m.OnChange("c_active_path") }
		m.CDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("c_dispatcher_path") }
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = SessionClientRelayFallback
		return true, nil
	case m.State == SessionClientRelayFallback && event == SessionEventRelayOk:
		m.State = SessionClientRelayConnected
		return true, nil
	case m.State == SessionClientLANConnecting && event == SessionEventAppForceFallback:
		m.State = SessionClientRelayConnected
		return true, nil
	case m.State == SessionClientLANVerifying && event == SessionEventAppForceFallback:
		m.CDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("c_dispatcher_path") }
		m.State = SessionClientRelayConnected
		return true, nil
	case m.State == SessionClientLANActive && event == SessionEventAppForceFallback:
		if fn := m.Actions[SessionActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.CActivePath = "relay"
		if m.OnChange != nil { m.OnChange("c_active_path") }
		m.CDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("c_dispatcher_path") }
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = SessionClientRelayConnected
		return true, nil
	case m.State == SessionClientRelayConnected && event == SessionEventDisconnect:
		m.State = SessionClientPaired
		return true, nil
	case m.State == SessionClientRelayConnected && event == SessionEventAppSend:
		m.State = SessionClientRelayConnected
		return true, nil
	case m.State == SessionClientLANConnecting && event == SessionEventAppSend:
		m.State = SessionClientLANConnecting
		return true, nil
	case m.State == SessionClientLANVerifying && event == SessionEventAppSend:
		m.State = SessionClientLANVerifying
		return true, nil
	case m.State == SessionClientLANActive && event == SessionEventAppSend:
		m.State = SessionClientLANActive
		return true, nil
	case m.State == SessionClientRelayFallback && event == SessionEventAppSend:
		m.State = SessionClientRelayFallback
		return true, nil
	case m.State == SessionClientRelayConnected && event == SessionEventRelayStreamData:
		m.State = SessionClientRelayConnected
		return true, nil
	case m.State == SessionClientLANConnecting && event == SessionEventRelayStreamData:
		m.State = SessionClientLANConnecting
		return true, nil
	case m.State == SessionClientLANVerifying && event == SessionEventRelayStreamData:
		m.State = SessionClientLANVerifying
		return true, nil
	case m.State == SessionClientLANActive && event == SessionEventRelayStreamData:
		m.State = SessionClientLANActive
		return true, nil
	case m.State == SessionClientRelayFallback && event == SessionEventRelayStreamData:
		m.State = SessionClientRelayFallback
		return true, nil
	case m.State == SessionClientRelayConnected && event == SessionEventRelayStreamError:
		m.State = SessionClientRelayConnected
		return true, nil
	case m.State == SessionClientLANConnecting && event == SessionEventRelayStreamError:
		m.State = SessionClientLANConnecting
		return true, nil
	case m.State == SessionClientLANVerifying && event == SessionEventRelayStreamError:
		m.State = SessionClientLANVerifying
		return true, nil
	case m.State == SessionClientLANActive && event == SessionEventRelayStreamError:
		m.State = SessionClientLANActive
		return true, nil
	case m.State == SessionClientRelayFallback && event == SessionEventRelayStreamError:
		m.State = SessionClientRelayFallback
		return true, nil
	case m.State == SessionClientRelayConnected && event == SessionEventAppSendDatagram:
		m.State = SessionClientRelayConnected
		return true, nil
	case m.State == SessionClientLANConnecting && event == SessionEventAppSendDatagram:
		m.State = SessionClientLANConnecting
		return true, nil
	case m.State == SessionClientLANVerifying && event == SessionEventAppSendDatagram:
		m.State = SessionClientLANVerifying
		return true, nil
	case m.State == SessionClientLANActive && event == SessionEventAppSendDatagram:
		m.State = SessionClientLANActive
		return true, nil
	case m.State == SessionClientRelayFallback && event == SessionEventAppSendDatagram:
		m.State = SessionClientRelayFallback
		return true, nil
	case m.State == SessionClientRelayConnected && event == SessionEventRelayDatagram:
		m.State = SessionClientRelayConnected
		return true, nil
	case m.State == SessionClientLANConnecting && event == SessionEventRelayDatagram:
		m.State = SessionClientLANConnecting
		return true, nil
	case m.State == SessionClientLANVerifying && event == SessionEventRelayDatagram:
		m.State = SessionClientLANVerifying
		return true, nil
	case m.State == SessionClientLANActive && event == SessionEventRelayDatagram:
		m.State = SessionClientLANActive
		return true, nil
	case m.State == SessionClientRelayFallback && event == SessionEventRelayDatagram:
		m.State = SessionClientRelayFallback
		return true, nil
	case m.State == SessionClientLANActive && event == SessionEventLanStreamData:
		m.State = SessionClientLANActive
		return true, nil
	case m.State == SessionClientLANActive && event == SessionEventLanDatagram:
		m.State = SessionClientLANActive
		return true, nil
	}
	return false, nil
}

func (m *SessionClientMachine) HandleEvent(ev EventID) ([]CmdID, error) {
	switch {
	case m.State == SessionClientIdle && ev == SessionEventBackchannelReceived:
		m.State = SessionClientObtainBackchannelSecret
		return nil, nil
	case m.State == SessionClientObtainBackchannelSecret && ev == SessionEventSecretParsed:
		m.State = SessionClientConnectRelay
		return nil, nil
	case m.State == SessionClientConnectRelay && ev == SessionEventRelayConnected:
		m.State = SessionClientGenKeyPair
		return nil, nil
	case m.State == SessionClientGenKeyPair && ev == SessionEventKeyPairGenerated:
		if fn := m.Actions[SessionActionSendPairHello]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.State = SessionClientWaitAck
		return nil, nil
	case m.State == SessionClientWaitAck && ev == SessionEventRecvPairHelloAck:
		if fn := m.Actions[SessionActionDeriveSecret]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		// received_backend_pub: recv_msg.pubkey (set by action)
		// client_shared_key: DeriveKey("client_pub", recv_msg.pubkey) (set by action)
		m.State = SessionClientE2EReady
		return nil, nil
	case m.State == SessionClientE2EReady && ev == SessionEventRecvPairConfirm:
		// client_code: DeriveCode(received_backend_pub, "client_pub") (set by action)
		m.State = SessionClientShowCode
		return nil, nil
	case m.State == SessionClientShowCode && ev == SessionEventCodeDisplayed:
		m.State = SessionClientWaitPairComplete
		return nil, nil
	case m.State == SessionClientWaitPairComplete && ev == SessionEventRecvPairComplete:
		if fn := m.Actions[SessionActionStoreSecret]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.State = SessionClientPaired
		return nil, nil
	case m.State == SessionClientPaired && ev == SessionEventAppLaunch:
		m.State = SessionClientReconnect
		return nil, nil
	case m.State == SessionClientReconnect && ev == SessionEventRelayConnected:
		m.State = SessionClientSendAuth
		return nil, nil
	case m.State == SessionClientSendAuth && ev == SessionEventRecvAuthOk:
		m.State = SessionClientSessionActive
		return nil, nil
	case m.State == SessionClientSessionActive && ev == SessionEventSessionEstablished:
		m.State = SessionClientRelayConnected
		return nil, nil
	case m.State == SessionClientRelayConnected && ev == SessionEventRecvLanOffer && m.Guards[SessionGuardLanEnabled] != nil && m.Guards[SessionGuardLanEnabled]():
		if fn := m.Actions[SessionActionDialLan]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.State = SessionClientLANConnecting
		return []CmdID{SessionCmdDialLan}, nil
	case m.State == SessionClientRelayConnected && ev == SessionEventRecvLanOffer && m.Guards[SessionGuardLanDisabled] != nil && m.Guards[SessionGuardLanDisabled]():
		m.State = SessionClientRelayConnected
		return nil, nil
	case m.State == SessionClientLANConnecting && ev == SessionEventLanDialOk:
		m.State = SessionClientLANVerifying
		return []CmdID{SessionCmdSendLanVerify}, nil
	case m.State == SessionClientLANConnecting && ev == SessionEventLanDialFailed:
		m.State = SessionClientRelayConnected
		return nil, nil
	case m.State == SessionClientLANVerifying && ev == SessionEventRecvLanConfirm:
		if fn := m.Actions[SessionActionActivateLan]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.CActivePath = "lan"
		if m.OnChange != nil { m.OnChange("c_active_path") }
		m.CDispatcherPath = "lan"
		if m.OnChange != nil { m.OnChange("c_dispatcher_path") }
		m.LanSignal = "ready"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = SessionClientLANActive
		return []CmdID{SessionCmdStartLanStreamReader, SessionCmdStartLanDgReader, SessionCmdSignalLanReady, SessionCmdSetCryptoDatagram}, nil
	case m.State == SessionClientLANVerifying && ev == SessionEventVerifyTimeout:
		m.CDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("c_dispatcher_path") }
		m.State = SessionClientRelayConnected
		return nil, nil
	case m.State == SessionClientLANActive && ev == SessionEventRecvPathPing:
		m.State = SessionClientLANActive
		return []CmdID{SessionCmdSendPathPong}, nil
	case m.State == SessionClientLANActive && ev == SessionEventLanError:
		if fn := m.Actions[SessionActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.CActivePath = "relay"
		if m.OnChange != nil { m.OnChange("c_active_path") }
		m.CDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("c_dispatcher_path") }
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = SessionClientRelayFallback
		return []CmdID{SessionCmdStopLanStreamReader, SessionCmdStopLanDgReader, SessionCmdCloseLanPath, SessionCmdResetLanReady}, nil
	case m.State == SessionClientLANActive && ev == SessionEventLanStreamError:
		if fn := m.Actions[SessionActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.CActivePath = "relay"
		if m.OnChange != nil { m.OnChange("c_active_path") }
		m.CDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("c_dispatcher_path") }
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = SessionClientRelayFallback
		return []CmdID{SessionCmdStopLanStreamReader, SessionCmdStopLanDgReader, SessionCmdCloseLanPath, SessionCmdResetLanReady}, nil
	case m.State == SessionClientRelayFallback && ev == SessionEventRelayOk:
		m.State = SessionClientRelayConnected
		return nil, nil
	case m.State == SessionClientLANActive && ev == SessionEventRecvLanOffer && m.Guards[SessionGuardLanEnabled] != nil && m.Guards[SessionGuardLanEnabled]():
		if fn := m.Actions[SessionActionDialLan]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.State = SessionClientLANConnecting
		return []CmdID{SessionCmdStopLanStreamReader, SessionCmdStopLanDgReader, SessionCmdCloseLanPath, SessionCmdDialLan}, nil
	case m.State == SessionClientLANConnecting && ev == SessionEventAppForceFallback:
		m.State = SessionClientRelayConnected
		return nil, nil
	case m.State == SessionClientLANVerifying && ev == SessionEventAppForceFallback:
		m.CDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("c_dispatcher_path") }
		m.State = SessionClientRelayConnected
		return []CmdID{SessionCmdStopLanStreamReader, SessionCmdStopLanDgReader, SessionCmdCloseLanPath}, nil
	case m.State == SessionClientLANActive && ev == SessionEventAppForceFallback:
		if fn := m.Actions[SessionActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.CActivePath = "relay"
		if m.OnChange != nil { m.OnChange("c_active_path") }
		m.CDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("c_dispatcher_path") }
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = SessionClientRelayConnected
		return []CmdID{SessionCmdStopLanStreamReader, SessionCmdStopLanDgReader, SessionCmdCloseLanPath, SessionCmdResetLanReady}, nil
	case m.State == SessionClientRelayConnected && ev == SessionEventDisconnect:
		m.State = SessionClientPaired
		return nil, nil
	case m.State == SessionClientRelayConnected && ev == SessionEventAppSend:
		m.State = SessionClientRelayConnected
		return []CmdID{SessionCmdWriteActiveStream}, nil
	case m.State == SessionClientLANConnecting && ev == SessionEventAppSend:
		m.State = SessionClientLANConnecting
		return []CmdID{SessionCmdWriteActiveStream}, nil
	case m.State == SessionClientLANVerifying && ev == SessionEventAppSend:
		m.State = SessionClientLANVerifying
		return []CmdID{SessionCmdWriteActiveStream}, nil
	case m.State == SessionClientLANActive && ev == SessionEventAppSend:
		m.State = SessionClientLANActive
		return []CmdID{SessionCmdWriteActiveStream}, nil
	case m.State == SessionClientRelayFallback && ev == SessionEventAppSend:
		m.State = SessionClientRelayFallback
		return []CmdID{SessionCmdWriteActiveStream}, nil
	case m.State == SessionClientRelayConnected && ev == SessionEventRelayStreamData:
		m.State = SessionClientRelayConnected
		return []CmdID{SessionCmdDeliverRecv}, nil
	case m.State == SessionClientLANConnecting && ev == SessionEventRelayStreamData:
		m.State = SessionClientLANConnecting
		return []CmdID{SessionCmdDeliverRecv}, nil
	case m.State == SessionClientLANVerifying && ev == SessionEventRelayStreamData:
		m.State = SessionClientLANVerifying
		return []CmdID{SessionCmdDeliverRecv}, nil
	case m.State == SessionClientLANActive && ev == SessionEventRelayStreamData:
		m.State = SessionClientLANActive
		return []CmdID{SessionCmdDeliverRecv}, nil
	case m.State == SessionClientRelayFallback && ev == SessionEventRelayStreamData:
		m.State = SessionClientRelayFallback
		return []CmdID{SessionCmdDeliverRecv}, nil
	case m.State == SessionClientRelayConnected && ev == SessionEventRelayStreamError:
		m.State = SessionClientRelayConnected
		return []CmdID{SessionCmdDeliverRecvError}, nil
	case m.State == SessionClientLANConnecting && ev == SessionEventRelayStreamError:
		m.State = SessionClientLANConnecting
		return []CmdID{SessionCmdDeliverRecvError}, nil
	case m.State == SessionClientLANVerifying && ev == SessionEventRelayStreamError:
		m.State = SessionClientLANVerifying
		return []CmdID{SessionCmdDeliverRecvError}, nil
	case m.State == SessionClientLANActive && ev == SessionEventRelayStreamError:
		m.State = SessionClientLANActive
		return []CmdID{SessionCmdDeliverRecvError}, nil
	case m.State == SessionClientRelayFallback && ev == SessionEventRelayStreamError:
		m.State = SessionClientRelayFallback
		return []CmdID{SessionCmdDeliverRecvError}, nil
	case m.State == SessionClientRelayConnected && ev == SessionEventAppSendDatagram:
		m.State = SessionClientRelayConnected
		return []CmdID{SessionCmdSendActiveDatagram}, nil
	case m.State == SessionClientLANConnecting && ev == SessionEventAppSendDatagram:
		m.State = SessionClientLANConnecting
		return []CmdID{SessionCmdSendActiveDatagram}, nil
	case m.State == SessionClientLANVerifying && ev == SessionEventAppSendDatagram:
		m.State = SessionClientLANVerifying
		return []CmdID{SessionCmdSendActiveDatagram}, nil
	case m.State == SessionClientLANActive && ev == SessionEventAppSendDatagram:
		m.State = SessionClientLANActive
		return []CmdID{SessionCmdSendActiveDatagram}, nil
	case m.State == SessionClientRelayFallback && ev == SessionEventAppSendDatagram:
		m.State = SessionClientRelayFallback
		return []CmdID{SessionCmdSendActiveDatagram}, nil
	case m.State == SessionClientRelayConnected && ev == SessionEventRelayDatagram:
		m.State = SessionClientRelayConnected
		return []CmdID{SessionCmdDeliverRecvDatagram}, nil
	case m.State == SessionClientLANConnecting && ev == SessionEventRelayDatagram:
		m.State = SessionClientLANConnecting
		return []CmdID{SessionCmdDeliverRecvDatagram}, nil
	case m.State == SessionClientLANVerifying && ev == SessionEventRelayDatagram:
		m.State = SessionClientLANVerifying
		return []CmdID{SessionCmdDeliverRecvDatagram}, nil
	case m.State == SessionClientLANActive && ev == SessionEventRelayDatagram:
		m.State = SessionClientLANActive
		return []CmdID{SessionCmdDeliverRecvDatagram}, nil
	case m.State == SessionClientRelayFallback && ev == SessionEventRelayDatagram:
		m.State = SessionClientRelayFallback
		return []CmdID{SessionCmdDeliverRecvDatagram}, nil
	case m.State == SessionClientLANActive && ev == SessionEventLanStreamData:
		m.State = SessionClientLANActive
		return []CmdID{SessionCmdDeliverRecv}, nil
	case m.State == SessionClientLANActive && ev == SessionEventLanDatagram:
		m.State = SessionClientLANActive
		return []CmdID{SessionCmdDeliverRecvDatagram}, nil
	}
	return nil, nil
}

// SessionRelayMachine is the generated state machine for the relay actor.
type SessionRelayMachine struct {
	State State
	RelayBridge string // relay bridge state

	Guards  map[GuardID]func() bool
	Actions map[ActionID]func() error
	OnChange func(varName string)
}

func NewSessionRelayMachine() *SessionRelayMachine {
	return &SessionRelayMachine{
		State: SessionRelayIdle,
		RelayBridge: "idle",
		Guards:  make(map[GuardID]func() bool),
		Actions: make(map[ActionID]func() error),
	}
}

func (m *SessionRelayMachine) HandleMessage(msg MsgType) (bool, error) {
	switch {
	}
	return false, nil
}

func (m *SessionRelayMachine) Step(event EventID) (bool, error) {
	switch {
	case m.State == SessionRelayIdle && event == SessionEventBackendRegister:
		m.State = SessionRelayBackendRegistered
		return true, nil
	case m.State == SessionRelayBackendRegistered && event == SessionEventClientConnect:
		if fn := m.Actions[SessionActionBridgeStreams]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.RelayBridge = "active"
		if m.OnChange != nil { m.OnChange("relay_bridge") }
		m.State = SessionRelayBridged
		return true, nil
	case m.State == SessionRelayBridged && event == SessionEventClientDisconnect:
		if fn := m.Actions[SessionActionUnbridge]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.RelayBridge = "idle"
		if m.OnChange != nil { m.OnChange("relay_bridge") }
		m.State = SessionRelayBackendRegistered
		return true, nil
	case m.State == SessionRelayBackendRegistered && event == SessionEventBackendDisconnect:
		m.State = SessionRelayIdle
		return true, nil
	}
	return false, nil
}

func (m *SessionRelayMachine) HandleEvent(ev EventID) ([]CmdID, error) {
	switch {
	case m.State == SessionRelayIdle && ev == SessionEventBackendRegister:
		m.State = SessionRelayBackendRegistered
		return nil, nil
	case m.State == SessionRelayBackendRegistered && ev == SessionEventClientConnect:
		if fn := m.Actions[SessionActionBridgeStreams]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.RelayBridge = "active"
		if m.OnChange != nil { m.OnChange("relay_bridge") }
		m.State = SessionRelayBridged
		return nil, nil
	case m.State == SessionRelayBridged && ev == SessionEventClientDisconnect:
		if fn := m.Actions[SessionActionUnbridge]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.RelayBridge = "idle"
		if m.OnChange != nil { m.OnChange("relay_bridge") }
		m.State = SessionRelayBackendRegistered
		return nil, nil
	case m.State == SessionRelayBackendRegistered && ev == SessionEventBackendDisconnect:
		m.State = SessionRelayIdle
		return nil, nil
	}
	return nil, nil
}

