// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Code generated from protocol/*.yaml. DO NOT EDIT.

package tern

import (
	"github.com/marcelocantos/tern/protocol"
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

// backend states.
const (
	BackendIdle State = "Idle"
	BackendGenerateToken State = "GenerateToken"
	BackendRegisterRelay State = "RegisterRelay"
	BackendWaitingForClient State = "WaitingForClient"
	BackendDeriveSecret State = "DeriveSecret"
	BackendSendAck State = "SendAck"
	BackendWaitingForCode State = "WaitingForCode"
	BackendValidateCode State = "ValidateCode"
	BackendStorePaired State = "StorePaired"
	BackendPaired State = "Paired"
	BackendAuthCheck State = "AuthCheck"
	BackendSessionActive State = "SessionActive"
	BackendRelayConnected State = "RelayConnected"
	BackendLANOffered State = "LANOffered"
	BackendLANActive State = "LANActive"
	BackendLANDegraded State = "LANDegraded"
	BackendRelayBackoff State = "RelayBackoff"
)

// client states.
const (
	ClientIdle State = "Idle"
	ClientObtainBackchannelSecret State = "ObtainBackchannelSecret"
	ClientConnectRelay State = "ConnectRelay"
	ClientGenKeyPair State = "GenKeyPair"
	ClientWaitAck State = "WaitAck"
	ClientE2EReady State = "E2EReady"
	ClientShowCode State = "ShowCode"
	ClientWaitPairComplete State = "WaitPairComplete"
	ClientPaired State = "Paired"
	ClientReconnect State = "Reconnect"
	ClientSendAuth State = "SendAuth"
	ClientSessionActive State = "SessionActive"
	ClientRelayConnected State = "RelayConnected"
	ClientLANConnecting State = "LANConnecting"
	ClientLANVerifying State = "LANVerifying"
	ClientLANActive State = "LANActive"
	ClientRelayFallback State = "RelayFallback"
)

// relay states.
const (
	RelayIdle State = "Idle"
	RelayBackendRegistered State = "BackendRegistered"
	RelayBridged State = "Bridged"
)

// Message types.
const (
	MsgPairHello MsgType = "pair_hello"
	MsgPairHelloAck MsgType = "pair_hello_ack"
	MsgPairConfirm MsgType = "pair_confirm"
	MsgPairComplete MsgType = "pair_complete"
	MsgAuthRequest MsgType = "auth_request"
	MsgAuthOk MsgType = "auth_ok"
	MsgLanOffer MsgType = "lan_offer"
	MsgLanVerify MsgType = "lan_verify"
	MsgLanConfirm MsgType = "lan_confirm"
	MsgPathPing MsgType = "path_ping"
	MsgPathPong MsgType = "path_pong"
)

// Guards.
const (
	GuardTokenValid GuardID = "token_valid"
	GuardTokenInvalid GuardID = "token_invalid"
	GuardCodeCorrect GuardID = "code_correct"
	GuardCodeWrong GuardID = "code_wrong"
	GuardDeviceKnown GuardID = "device_known"
	GuardDeviceUnknown GuardID = "device_unknown"
	GuardNonceFresh GuardID = "nonce_fresh"
	GuardChallengeValid GuardID = "challenge_valid"
	GuardChallengeInvalid GuardID = "challenge_invalid"
	GuardLanEnabled GuardID = "lan_enabled"
	GuardLanDisabled GuardID = "lan_disabled"
	GuardLanServerAvailable GuardID = "lan_server_available"
	GuardUnderMaxFailures GuardID = "under_max_failures"
	GuardAtMaxFailures GuardID = "at_max_failures"
)

// Actions.
const (
	ActionActivateLan ActionID = "activate_lan"
	ActionBridgeStreams ActionID = "bridge_streams"
	ActionDeriveSecret ActionID = "derive_secret"
	ActionDialLan ActionID = "dial_lan"
	ActionFallbackToRelay ActionID = "fallback_to_relay"
	ActionGenerateToken ActionID = "generate_token"
	ActionRegisterRelay ActionID = "register_relay"
	ActionResetFailures ActionID = "reset_failures"
	ActionSendPairHello ActionID = "send_pair_hello"
	ActionStoreDevice ActionID = "store_device"
	ActionStoreSecret ActionID = "store_secret"
	ActionUnbridge ActionID = "unbridge"
	ActionVerifyDevice ActionID = "verify_device"
)

// Events.
const (
	EventAppClose EventID = "app_close"
	EventAppLaunch EventID = "app_launch"
	EventAppRecv EventID = "app_recv"
	EventAppRecvDatagram EventID = "app_recv_datagram"
	EventAppSend EventID = "app_send"
	EventAppSendDatagram EventID = "app_send_datagram"
	EventBackchannelReceived EventID = "backchannel_received"
	EventBackendDisconnect EventID = "backend_disconnect"
	EventBackendRegister EventID = "backend_register"
	EventBackoffExpired EventID = "backoff_expired"
	EventCheckCode EventID = "check_code"
	EventCliCodeEntered EventID = "cli_code_entered"
	EventCliInitPair EventID = "cli_init_pair"
	EventClientConnect EventID = "client_connect"
	EventClientDisconnect EventID = "client_disconnect"
	EventCodeDisplayed EventID = "code_displayed"
	EventDisconnect EventID = "disconnect"
	EventEcdhComplete EventID = "ecdh_complete"
	EventFinalise EventID = "finalise"
	EventKeyPairGenerated EventID = "key_pair_generated"
	EventLanDatagram EventID = "lan_datagram"
	EventLanDialFailed EventID = "lan_dial_failed"
	EventLanDialOk EventID = "lan_dial_ok"
	EventLanError EventID = "lan_error"
	EventLanServerChanged EventID = "lan_server_changed"
	EventLanServerReady EventID = "lan_server_ready"
	EventLanStreamData EventID = "lan_stream_data"
	EventLanStreamError EventID = "lan_stream_error"
	EventLanVerifyOk EventID = "lan_verify_ok"
	EventOfferTimeout EventID = "offer_timeout"
	EventPingTick EventID = "ping_tick"
	EventPingTimeout EventID = "ping_timeout"
	EventReadvertiseTick EventID = "readvertise_tick"
	EventRecvAuthOk EventID = "recv_auth_ok"
	EventRecvAuthRequest EventID = "recv_auth_request"
	EventRecvLanConfirm EventID = "recv_lan_confirm"
	EventRecvLanOffer EventID = "recv_lan_offer"
	EventRecvLanVerify EventID = "recv_lan_verify"
	EventRecvPairComplete EventID = "recv_pair_complete"
	EventRecvPairConfirm EventID = "recv_pair_confirm"
	EventRecvPairHello EventID = "recv_pair_hello"
	EventRecvPairHelloAck EventID = "recv_pair_hello_ack"
	EventRecvPathPing EventID = "recv_path_ping"
	EventRecvPathPong EventID = "recv_path_pong"
	EventRelayConnected EventID = "relay_connected"
	EventRelayDatagram EventID = "relay_datagram"
	EventRelayOk EventID = "relay_ok"
	EventRelayRegistered EventID = "relay_registered"
	EventRelayStreamData EventID = "relay_stream_data"
	EventRelayStreamError EventID = "relay_stream_error"
	EventSecretParsed EventID = "secret_parsed"
	EventSessionEstablished EventID = "session_established"
	EventSignalCodeDisplay EventID = "signal_code_display"
	EventTokenCreated EventID = "token_created"
	EventVerify EventID = "verify"
	EventVerifyTimeout EventID = "verify_timeout"
)

// Commands.
const (
	CmdWriteActiveStream CmdID = "write_active_stream"
	CmdSendActiveDatagram CmdID = "send_active_datagram"
	CmdSendPathPing CmdID = "send_path_ping"
	CmdSendPathPong CmdID = "send_path_pong"
	CmdSendLanOffer CmdID = "send_lan_offer"
	CmdSendLanVerify CmdID = "send_lan_verify"
	CmdSendLanConfirm CmdID = "send_lan_confirm"
	CmdDialLan CmdID = "dial_lan"
	CmdDeliverRecv CmdID = "deliver_recv"
	CmdDeliverRecvDatagram CmdID = "deliver_recv_datagram"
	CmdStartLanStreamReader CmdID = "start_lan_stream_reader"
	CmdStopLanStreamReader CmdID = "stop_lan_stream_reader"
	CmdStartLanDgReader CmdID = "start_lan_dg_reader"
	CmdStopLanDgReader CmdID = "stop_lan_dg_reader"
	CmdStartMonitor CmdID = "start_monitor"
	CmdStopMonitor CmdID = "stop_monitor"
	CmdStartBackoffTimer CmdID = "start_backoff_timer"
	CmdCloseLanPath CmdID = "close_lan_path"
	CmdSignalLanReady CmdID = "signal_lan_ready"
	CmdResetLanReady CmdID = "reset_lan_ready"
	CmdSetCryptoDatagram CmdID = "set_crypto_datagram"
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
				{From: "RelayConnected", To: "RelayConnected", On: Internal("app_send")},
				{From: "RelayConnected", To: "RelayConnected", On: Internal("relay_stream_data")},
				{From: "LANOffered", To: "LANOffered", On: Internal("app_send")},
				{From: "LANOffered", To: "LANOffered", On: Internal("relay_stream_data")},
				{From: "LANActive", To: "LANActive", On: Internal("app_send")},
				{From: "LANActive", To: "LANActive", On: Internal("lan_stream_data")},
				{From: "LANActive", To: "LANActive", On: Internal("relay_stream_data")},
				{From: "LANDegraded", To: "LANDegraded", On: Internal("app_send")},
				{From: "LANDegraded", To: "LANDegraded", On: Internal("lan_stream_data")},
				{From: "LANDegraded", To: "LANDegraded", On: Internal("relay_stream_data")},
				{From: "RelayBackoff", To: "RelayBackoff", On: Internal("app_send")},
				{From: "RelayBackoff", To: "RelayBackoff", On: Internal("relay_stream_data")},
				{From: "RelayConnected", To: "RelayConnected", On: Internal("app_send_datagram")},
				{From: "RelayConnected", To: "RelayConnected", On: Internal("relay_datagram")},
				{From: "LANOffered", To: "LANOffered", On: Internal("app_send_datagram")},
				{From: "LANOffered", To: "LANOffered", On: Internal("relay_datagram")},
				{From: "LANActive", To: "LANActive", On: Internal("app_send_datagram")},
				{From: "LANActive", To: "LANActive", On: Internal("lan_datagram")},
				{From: "LANActive", To: "LANActive", On: Internal("relay_datagram")},
				{From: "LANDegraded", To: "LANDegraded", On: Internal("app_send_datagram")},
				{From: "LANDegraded", To: "LANDegraded", On: Internal("lan_datagram")},
				{From: "LANDegraded", To: "LANDegraded", On: Internal("relay_datagram")},
				{From: "RelayBackoff", To: "RelayBackoff", On: Internal("app_send_datagram")},
				{From: "RelayBackoff", To: "RelayBackoff", On: Internal("relay_datagram")},
				{From: "RelayConnected", To: "LANOffered", On: Internal("lan_server_ready"), Sends: []Send{{To: "client", Msg: "lan_offer", Fields: map[string]string{"addr": "lan_addr", "challenge": "challenge_bytes", }}, }},
				{From: "LANOffered", To: "LANActive", On: Recv("lan_verify"), Guard: "challenge_valid", Do: "activate_lan", Sends: []Send{{To: "client", Msg: "lan_confirm"}, }, Updates: []VarUpdate{{Var: "ping_failures", Expr: "0"}, {Var: "backoff_level", Expr: "0"}, {Var: "b_active_path", Expr: "\"lan\""}, {Var: "b_dispatcher_path", Expr: "\"lan\""}, {Var: "monitor_target", Expr: "\"lan\""}, {Var: "lan_signal", Expr: "\"ready\""}, }},
				{From: "LANOffered", To: "RelayConnected", On: Recv("lan_verify"), Guard: "challenge_invalid"},
				{From: "LANOffered", To: "RelayBackoff", On: Internal("offer_timeout"), Updates: []VarUpdate{{Var: "backoff_level", Expr: "Min(backoff_level + 1, max_backoff_level)"}, {Var: "lan_signal", Expr: "\"pending\""}, }},
				{From: "LANActive", To: "LANActive", On: Internal("ping_tick"), Sends: []Send{{To: "client", Msg: "path_ping"}, }},
				{From: "LANActive", To: "LANDegraded", On: Internal("ping_timeout"), Updates: []VarUpdate{{Var: "ping_failures", Expr: "1"}, }},
				{From: "LANDegraded", To: "LANDegraded", On: Internal("ping_tick"), Sends: []Send{{To: "client", Msg: "path_ping"}, }},
				{From: "LANDegraded", To: "LANActive", On: Recv("path_pong"), Do: "reset_failures", Updates: []VarUpdate{{Var: "ping_failures", Expr: "0"}, }},
				{From: "LANDegraded", To: "LANDegraded", On: Internal("ping_timeout"), Guard: "under_max_failures", Updates: []VarUpdate{{Var: "ping_failures", Expr: "ping_failures + 1"}, }},
				{From: "LANDegraded", To: "RelayBackoff", On: Internal("ping_timeout"), Guard: "at_max_failures", Do: "fallback_to_relay", Updates: []VarUpdate{{Var: "backoff_level", Expr: "Min(backoff_level + 1, max_backoff_level)"}, {Var: "b_active_path", Expr: "\"relay\""}, {Var: "b_dispatcher_path", Expr: "\"relay\""}, {Var: "monitor_target", Expr: "\"none\""}, {Var: "lan_signal", Expr: "\"pending\""}, {Var: "ping_failures", Expr: "0"}, }},
				{From: "RelayBackoff", To: "LANOffered", On: Internal("backoff_expired"), Sends: []Send{{To: "client", Msg: "lan_offer", Fields: map[string]string{"addr": "lan_addr", "challenge": "challenge_bytes", }}, }},
				{From: "RelayBackoff", To: "LANOffered", On: Internal("lan_server_changed"), Sends: []Send{{To: "client", Msg: "lan_offer", Fields: map[string]string{"challenge": "challenge_bytes", "addr": "lan_addr", }}, }, Updates: []VarUpdate{{Var: "backoff_level", Expr: "0"}, }},
				{From: "RelayConnected", To: "LANOffered", On: Internal("readvertise_tick"), Guard: "lan_server_available", Sends: []Send{{To: "client", Msg: "lan_offer", Fields: map[string]string{"addr": "lan_addr", "challenge": "challenge_bytes", }}, }},
				{From: "RelayConnected", To: "Paired", On: Internal("disconnect")},
			}},
			{Name: "client", Initial: "Idle", Transitions: []Transition{
				{From: "Idle", To: "ObtainBackchannelSecret", On: Internal("backchannel_received")},
				{From: "ObtainBackchannelSecret", To: "ConnectRelay", On: Internal("secret_parsed")},
				{From: "ConnectRelay", To: "GenKeyPair", On: Internal("relay_connected")},
				{From: "GenKeyPair", To: "WaitAck", On: Internal("key_pair_generated"), Do: "send_pair_hello", Sends: []Send{{To: "backend", Msg: "pair_hello", Fields: map[string]string{"token": "current_token", "pubkey": "\"client_pub\"", }}, }},
				{From: "WaitAck", To: "E2EReady", On: Recv("pair_hello_ack"), Do: "derive_secret", Updates: []VarUpdate{{Var: "received_backend_pub", Expr: "recv_msg.pubkey"}, {Var: "client_shared_key", Expr: "DeriveKey(\"client_pub\", recv_msg.pubkey)"}, }},
				{From: "E2EReady", To: "ShowCode", On: Recv("pair_confirm"), Updates: []VarUpdate{{Var: "client_code", Expr: "DeriveCode(received_backend_pub, \"client_pub\")"}, }},
				{From: "ShowCode", To: "WaitPairComplete", On: Internal("code_displayed")},
				{From: "WaitPairComplete", To: "Paired", On: Recv("pair_complete"), Do: "store_secret"},
				{From: "Paired", To: "Reconnect", On: Internal("app_launch")},
				{From: "Reconnect", To: "SendAuth", On: Internal("relay_connected"), Sends: []Send{{To: "backend", Msg: "auth_request", Fields: map[string]string{"device_id": "\"device_1\"", "secret": "device_secret", "nonce": "\"nonce_1\"", "key": "client_shared_key", }}, }},
				{From: "SendAuth", To: "SessionActive", On: Recv("auth_ok")},
				{From: "SessionActive", To: "RelayConnected", On: Internal("session_established")},
				{From: "RelayConnected", To: "RelayConnected", On: Internal("app_send")},
				{From: "RelayConnected", To: "RelayConnected", On: Internal("relay_stream_data")},
				{From: "LANConnecting", To: "LANConnecting", On: Internal("app_send")},
				{From: "LANConnecting", To: "LANConnecting", On: Internal("relay_stream_data")},
				{From: "LANVerifying", To: "LANVerifying", On: Internal("app_send")},
				{From: "LANVerifying", To: "LANVerifying", On: Internal("relay_stream_data")},
				{From: "LANActive", To: "LANActive", On: Internal("app_send")},
				{From: "LANActive", To: "LANActive", On: Internal("lan_stream_data")},
				{From: "LANActive", To: "LANActive", On: Internal("relay_stream_data")},
				{From: "RelayFallback", To: "RelayFallback", On: Internal("app_send")},
				{From: "RelayFallback", To: "RelayFallback", On: Internal("relay_stream_data")},
				{From: "RelayConnected", To: "RelayConnected", On: Internal("app_send_datagram")},
				{From: "RelayConnected", To: "RelayConnected", On: Internal("relay_datagram")},
				{From: "LANConnecting", To: "LANConnecting", On: Internal("app_send_datagram")},
				{From: "LANConnecting", To: "LANConnecting", On: Internal("relay_datagram")},
				{From: "LANVerifying", To: "LANVerifying", On: Internal("app_send_datagram")},
				{From: "LANVerifying", To: "LANVerifying", On: Internal("relay_datagram")},
				{From: "LANActive", To: "LANActive", On: Internal("app_send_datagram")},
				{From: "LANActive", To: "LANActive", On: Internal("lan_datagram")},
				{From: "LANActive", To: "LANActive", On: Internal("relay_datagram")},
				{From: "RelayFallback", To: "RelayFallback", On: Internal("app_send_datagram")},
				{From: "RelayFallback", To: "RelayFallback", On: Internal("relay_datagram")},
				{From: "RelayConnected", To: "LANConnecting", On: Recv("lan_offer"), Guard: "lan_enabled", Do: "dial_lan"},
				{From: "RelayConnected", To: "RelayConnected", On: Recv("lan_offer"), Guard: "lan_disabled"},
				{From: "LANConnecting", To: "LANVerifying", On: Internal("lan_dial_ok"), Sends: []Send{{To: "backend", Msg: "lan_verify", Fields: map[string]string{"challenge": "offer_challenge", "instance_id": "instance_id", }}, }},
				{From: "LANConnecting", To: "RelayConnected", On: Internal("lan_dial_failed")},
				{From: "LANVerifying", To: "LANActive", On: Recv("lan_confirm"), Do: "activate_lan", Updates: []VarUpdate{{Var: "c_active_path", Expr: "\"lan\""}, {Var: "c_dispatcher_path", Expr: "\"lan\""}, {Var: "lan_signal", Expr: "\"ready\""}, }},
				{From: "LANVerifying", To: "RelayConnected", On: Internal("verify_timeout"), Updates: []VarUpdate{{Var: "c_dispatcher_path", Expr: "\"relay\""}, }},
				{From: "LANActive", To: "LANActive", On: Recv("path_ping"), Sends: []Send{{To: "backend", Msg: "path_pong"}, }},
				{From: "LANActive", To: "RelayFallback", On: Internal("lan_error"), Do: "fallback_to_relay", Updates: []VarUpdate{{Var: "c_active_path", Expr: "\"relay\""}, {Var: "c_dispatcher_path", Expr: "\"relay\""}, {Var: "lan_signal", Expr: "\"pending\""}, }},
				{From: "RelayFallback", To: "RelayConnected", On: Internal("relay_ok")},
				{From: "LANActive", To: "LANConnecting", On: Recv("lan_offer"), Guard: "lan_enabled", Do: "dial_lan"},
				{From: "RelayConnected", To: "Paired", On: Internal("disconnect")},
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

// BackendMachine is the generated state machine for the backend actor.
type BackendMachine struct {
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

func NewBackendMachine() *BackendMachine {
	return &BackendMachine{
		State: BackendIdle,
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

func (m *BackendMachine) HandleMessage(msg MsgType) (bool, error) {
	switch {
	case m.State == BackendWaitingForClient && msg == MsgPairHello && m.Guards[GuardTokenValid] != nil && m.Guards[GuardTokenValid]():
		if fn := m.Actions[ActionDeriveSecret]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		// received_client_pub: recv_msg.pubkey (set by action)
		m.BackendEcdhPub = "backend_pub"
		if m.OnChange != nil { m.OnChange("backend_ecdh_pub") }
		// backend_shared_key: DeriveKey("backend_pub", recv_msg.pubkey) (set by action)
		// backend_code: DeriveCode("backend_pub", recv_msg.pubkey) (set by action)
		m.State = BackendDeriveSecret
		return true, nil
	case m.State == BackendWaitingForClient && msg == MsgPairHello && m.Guards[GuardTokenInvalid] != nil && m.Guards[GuardTokenInvalid]():
		m.State = BackendIdle
		return true, nil
	case m.State == BackendPaired && msg == MsgAuthRequest:
		// received_device_id: recv_msg.device_id (set by action)
		// received_auth_nonce: recv_msg.nonce (set by action)
		m.State = BackendAuthCheck
		return true, nil
	case m.State == BackendLANOffered && msg == MsgLanVerify && m.Guards[GuardChallengeValid] != nil && m.Guards[GuardChallengeValid]():
		if fn := m.Actions[ActionActivateLan]; fn != nil {
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
		m.State = BackendLANActive
		return true, nil
	case m.State == BackendLANOffered && msg == MsgLanVerify && m.Guards[GuardChallengeInvalid] != nil && m.Guards[GuardChallengeInvalid]():
		m.State = BackendRelayConnected
		return true, nil
	case m.State == BackendLANDegraded && msg == MsgPathPong:
		if fn := m.Actions[ActionResetFailures]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.PingFailures = 0
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = BackendLANActive
		return true, nil
	}
	return false, nil
}

func (m *BackendMachine) Step(event EventID) (bool, error) {
	switch {
	case m.State == BackendIdle && event == EventCliInitPair:
		if fn := m.Actions[ActionGenerateToken]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.CurrentToken = "tok_1"
		if m.OnChange != nil { m.OnChange("current_token") }
		// active_tokens: active_tokens \union {"tok_1"} (set by action)
		m.State = BackendGenerateToken
		return true, nil
	case m.State == BackendGenerateToken && event == EventTokenCreated:
		if fn := m.Actions[ActionRegisterRelay]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.State = BackendRegisterRelay
		return true, nil
	case m.State == BackendRegisterRelay && event == EventRelayRegistered:
		m.SecretPublished = true
		if m.OnChange != nil { m.OnChange("secret_published") }
		m.State = BackendWaitingForClient
		return true, nil
	case m.State == BackendDeriveSecret && event == EventEcdhComplete:
		m.State = BackendSendAck
		return true, nil
	case m.State == BackendSendAck && event == EventSignalCodeDisplay:
		m.State = BackendWaitingForCode
		return true, nil
	case m.State == BackendWaitingForCode && event == EventCliCodeEntered:
		// received_code: cli_entered_code (set by action)
		m.State = BackendValidateCode
		return true, nil
	case m.State == BackendValidateCode && event == EventCheckCode && m.Guards[GuardCodeCorrect] != nil && m.Guards[GuardCodeCorrect]():
		m.State = BackendStorePaired
		return true, nil
	case m.State == BackendValidateCode && event == EventCheckCode && m.Guards[GuardCodeWrong] != nil && m.Guards[GuardCodeWrong]():
		// code_attempts: code_attempts + 1 (set by action)
		m.State = BackendIdle
		return true, nil
	case m.State == BackendStorePaired && event == EventFinalise:
		if fn := m.Actions[ActionStoreDevice]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.DeviceSecret = "dev_secret_1"
		if m.OnChange != nil { m.OnChange("device_secret") }
		// paired_devices: paired_devices \union {"device_1"} (set by action)
		// active_tokens: active_tokens \ {current_token} (set by action)
		// used_tokens: used_tokens \union {current_token} (set by action)
		m.State = BackendPaired
		return true, nil
	case m.State == BackendAuthCheck && event == EventVerify && m.Guards[GuardDeviceKnown] != nil && m.Guards[GuardDeviceKnown]():
		if fn := m.Actions[ActionVerifyDevice]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		// auth_nonces_used: auth_nonces_used \union {received_auth_nonce} (set by action)
		m.State = BackendSessionActive
		return true, nil
	case m.State == BackendAuthCheck && event == EventVerify && m.Guards[GuardDeviceUnknown] != nil && m.Guards[GuardDeviceUnknown]():
		m.State = BackendIdle
		return true, nil
	case m.State == BackendSessionActive && event == EventSessionEstablished:
		m.State = BackendRelayConnected
		return true, nil
	case m.State == BackendRelayConnected && event == EventAppSend:
		m.State = BackendRelayConnected
		return true, nil
	case m.State == BackendRelayConnected && event == EventRelayStreamData:
		m.State = BackendRelayConnected
		return true, nil
	case m.State == BackendLANOffered && event == EventAppSend:
		m.State = BackendLANOffered
		return true, nil
	case m.State == BackendLANOffered && event == EventRelayStreamData:
		m.State = BackendLANOffered
		return true, nil
	case m.State == BackendLANActive && event == EventAppSend:
		m.State = BackendLANActive
		return true, nil
	case m.State == BackendLANActive && event == EventLanStreamData:
		m.State = BackendLANActive
		return true, nil
	case m.State == BackendLANActive && event == EventRelayStreamData:
		m.State = BackendLANActive
		return true, nil
	case m.State == BackendLANDegraded && event == EventAppSend:
		m.State = BackendLANDegraded
		return true, nil
	case m.State == BackendLANDegraded && event == EventLanStreamData:
		m.State = BackendLANDegraded
		return true, nil
	case m.State == BackendLANDegraded && event == EventRelayStreamData:
		m.State = BackendLANDegraded
		return true, nil
	case m.State == BackendRelayBackoff && event == EventAppSend:
		m.State = BackendRelayBackoff
		return true, nil
	case m.State == BackendRelayBackoff && event == EventRelayStreamData:
		m.State = BackendRelayBackoff
		return true, nil
	case m.State == BackendRelayConnected && event == EventAppSendDatagram:
		m.State = BackendRelayConnected
		return true, nil
	case m.State == BackendRelayConnected && event == EventRelayDatagram:
		m.State = BackendRelayConnected
		return true, nil
	case m.State == BackendLANOffered && event == EventAppSendDatagram:
		m.State = BackendLANOffered
		return true, nil
	case m.State == BackendLANOffered && event == EventRelayDatagram:
		m.State = BackendLANOffered
		return true, nil
	case m.State == BackendLANActive && event == EventAppSendDatagram:
		m.State = BackendLANActive
		return true, nil
	case m.State == BackendLANActive && event == EventLanDatagram:
		m.State = BackendLANActive
		return true, nil
	case m.State == BackendLANActive && event == EventRelayDatagram:
		m.State = BackendLANActive
		return true, nil
	case m.State == BackendLANDegraded && event == EventAppSendDatagram:
		m.State = BackendLANDegraded
		return true, nil
	case m.State == BackendLANDegraded && event == EventLanDatagram:
		m.State = BackendLANDegraded
		return true, nil
	case m.State == BackendLANDegraded && event == EventRelayDatagram:
		m.State = BackendLANDegraded
		return true, nil
	case m.State == BackendRelayBackoff && event == EventAppSendDatagram:
		m.State = BackendRelayBackoff
		return true, nil
	case m.State == BackendRelayBackoff && event == EventRelayDatagram:
		m.State = BackendRelayBackoff
		return true, nil
	case m.State == BackendRelayConnected && event == EventLanServerReady:
		m.State = BackendLANOffered
		return true, nil
	case m.State == BackendLANOffered && event == EventOfferTimeout:
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = BackendRelayBackoff
		return true, nil
	case m.State == BackendLANActive && event == EventPingTick:
		m.State = BackendLANActive
		return true, nil
	case m.State == BackendLANActive && event == EventPingTimeout:
		m.PingFailures = 1
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = BackendLANDegraded
		return true, nil
	case m.State == BackendLANDegraded && event == EventPingTick:
		m.State = BackendLANDegraded
		return true, nil
	case m.State == BackendLANDegraded && event == EventPingTimeout && m.Guards[GuardUnderMaxFailures] != nil && m.Guards[GuardUnderMaxFailures]():
		// ping_failures: ping_failures + 1 (set by action)
		m.State = BackendLANDegraded
		return true, nil
	case m.State == BackendLANDegraded && event == EventPingTimeout && m.Guards[GuardAtMaxFailures] != nil && m.Guards[GuardAtMaxFailures]():
		if fn := m.Actions[ActionFallbackToRelay]; fn != nil {
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
		m.State = BackendRelayBackoff
		return true, nil
	case m.State == BackendRelayBackoff && event == EventBackoffExpired:
		m.State = BackendLANOffered
		return true, nil
	case m.State == BackendRelayBackoff && event == EventLanServerChanged:
		m.BackoffLevel = 0
		if m.OnChange != nil { m.OnChange("backoff_level") }
		m.State = BackendLANOffered
		return true, nil
	case m.State == BackendRelayConnected && event == EventReadvertiseTick && m.Guards[GuardLanServerAvailable] != nil && m.Guards[GuardLanServerAvailable]():
		m.State = BackendLANOffered
		return true, nil
	case m.State == BackendRelayConnected && event == EventDisconnect:
		m.State = BackendPaired
		return true, nil
	}
	return false, nil
}

func (m *BackendMachine) HandleEvent(ev EventID) ([]CmdID, error) {
	switch {
	case m.State == BackendIdle && ev == EventCliInitPair:
		if fn := m.Actions[ActionGenerateToken]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.CurrentToken = "tok_1"
		if m.OnChange != nil { m.OnChange("current_token") }
		// active_tokens: active_tokens \union {"tok_1"} (set by action)
		m.State = BackendGenerateToken
		return nil, nil
	case m.State == BackendGenerateToken && ev == EventTokenCreated:
		if fn := m.Actions[ActionRegisterRelay]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.State = BackendRegisterRelay
		return nil, nil
	case m.State == BackendRegisterRelay && ev == EventRelayRegistered:
		m.SecretPublished = true
		if m.OnChange != nil { m.OnChange("secret_published") }
		m.State = BackendWaitingForClient
		return nil, nil
	case m.State == BackendWaitingForClient && ev == EventRecvPairHello && m.Guards[GuardTokenValid] != nil && m.Guards[GuardTokenValid]():
		if fn := m.Actions[ActionDeriveSecret]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		// received_client_pub: recv_msg.pubkey (set by action)
		m.BackendEcdhPub = "backend_pub"
		if m.OnChange != nil { m.OnChange("backend_ecdh_pub") }
		// backend_shared_key: DeriveKey("backend_pub", recv_msg.pubkey) (set by action)
		// backend_code: DeriveCode("backend_pub", recv_msg.pubkey) (set by action)
		m.State = BackendDeriveSecret
		return nil, nil
	case m.State == BackendWaitingForClient && ev == EventRecvPairHello && m.Guards[GuardTokenInvalid] != nil && m.Guards[GuardTokenInvalid]():
		m.State = BackendIdle
		return nil, nil
	case m.State == BackendDeriveSecret && ev == EventEcdhComplete:
		m.State = BackendSendAck
		return nil, nil
	case m.State == BackendSendAck && ev == EventSignalCodeDisplay:
		m.State = BackendWaitingForCode
		return nil, nil
	case m.State == BackendWaitingForCode && ev == EventCliCodeEntered:
		// received_code: cli_entered_code (set by action)
		m.State = BackendValidateCode
		return nil, nil
	case m.State == BackendValidateCode && ev == EventCheckCode && m.Guards[GuardCodeCorrect] != nil && m.Guards[GuardCodeCorrect]():
		m.State = BackendStorePaired
		return nil, nil
	case m.State == BackendValidateCode && ev == EventCheckCode && m.Guards[GuardCodeWrong] != nil && m.Guards[GuardCodeWrong]():
		// code_attempts: code_attempts + 1 (set by action)
		m.State = BackendIdle
		return nil, nil
	case m.State == BackendStorePaired && ev == EventFinalise:
		if fn := m.Actions[ActionStoreDevice]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.DeviceSecret = "dev_secret_1"
		if m.OnChange != nil { m.OnChange("device_secret") }
		// paired_devices: paired_devices \union {"device_1"} (set by action)
		// active_tokens: active_tokens \ {current_token} (set by action)
		// used_tokens: used_tokens \union {current_token} (set by action)
		m.State = BackendPaired
		return nil, nil
	case m.State == BackendPaired && ev == EventRecvAuthRequest:
		// received_device_id: recv_msg.device_id (set by action)
		// received_auth_nonce: recv_msg.nonce (set by action)
		m.State = BackendAuthCheck
		return nil, nil
	case m.State == BackendAuthCheck && ev == EventVerify && m.Guards[GuardDeviceKnown] != nil && m.Guards[GuardDeviceKnown]():
		if fn := m.Actions[ActionVerifyDevice]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		// auth_nonces_used: auth_nonces_used \union {received_auth_nonce} (set by action)
		m.State = BackendSessionActive
		return nil, nil
	case m.State == BackendAuthCheck && ev == EventVerify && m.Guards[GuardDeviceUnknown] != nil && m.Guards[GuardDeviceUnknown]():
		m.State = BackendIdle
		return nil, nil
	case m.State == BackendSessionActive && ev == EventSessionEstablished:
		m.State = BackendRelayConnected
		return nil, nil
	case m.State == BackendRelayConnected && ev == EventAppSend:
		m.State = BackendRelayConnected
		return []CmdID{CmdWriteActiveStream}, nil
	case m.State == BackendRelayConnected && ev == EventRelayStreamData:
		m.State = BackendRelayConnected
		return []CmdID{CmdDeliverRecv}, nil
	case m.State == BackendLANOffered && ev == EventAppSend:
		m.State = BackendLANOffered
		return []CmdID{CmdWriteActiveStream}, nil
	case m.State == BackendLANOffered && ev == EventRelayStreamData:
		m.State = BackendLANOffered
		return []CmdID{CmdDeliverRecv}, nil
	case m.State == BackendLANActive && ev == EventAppSend:
		m.State = BackendLANActive
		return []CmdID{CmdWriteActiveStream}, nil
	case m.State == BackendLANActive && ev == EventLanStreamData:
		m.State = BackendLANActive
		return []CmdID{CmdDeliverRecv}, nil
	case m.State == BackendLANActive && ev == EventRelayStreamData:
		m.State = BackendLANActive
		return []CmdID{CmdDeliverRecv}, nil
	case m.State == BackendLANDegraded && ev == EventAppSend:
		m.State = BackendLANDegraded
		return []CmdID{CmdWriteActiveStream}, nil
	case m.State == BackendLANDegraded && ev == EventLanStreamData:
		m.State = BackendLANDegraded
		return []CmdID{CmdDeliverRecv}, nil
	case m.State == BackendLANDegraded && ev == EventRelayStreamData:
		m.State = BackendLANDegraded
		return []CmdID{CmdDeliverRecv}, nil
	case m.State == BackendRelayBackoff && ev == EventAppSend:
		m.State = BackendRelayBackoff
		return []CmdID{CmdWriteActiveStream}, nil
	case m.State == BackendRelayBackoff && ev == EventRelayStreamData:
		m.State = BackendRelayBackoff
		return []CmdID{CmdDeliverRecv}, nil
	case m.State == BackendRelayConnected && ev == EventAppSendDatagram:
		m.State = BackendRelayConnected
		return []CmdID{CmdSendActiveDatagram}, nil
	case m.State == BackendRelayConnected && ev == EventRelayDatagram:
		m.State = BackendRelayConnected
		return []CmdID{CmdDeliverRecvDatagram}, nil
	case m.State == BackendLANOffered && ev == EventAppSendDatagram:
		m.State = BackendLANOffered
		return []CmdID{CmdSendActiveDatagram}, nil
	case m.State == BackendLANOffered && ev == EventRelayDatagram:
		m.State = BackendLANOffered
		return []CmdID{CmdDeliverRecvDatagram}, nil
	case m.State == BackendLANActive && ev == EventAppSendDatagram:
		m.State = BackendLANActive
		return []CmdID{CmdSendActiveDatagram}, nil
	case m.State == BackendLANActive && ev == EventLanDatagram:
		m.State = BackendLANActive
		return []CmdID{CmdDeliverRecvDatagram}, nil
	case m.State == BackendLANActive && ev == EventRelayDatagram:
		m.State = BackendLANActive
		return []CmdID{CmdDeliverRecvDatagram}, nil
	case m.State == BackendLANDegraded && ev == EventAppSendDatagram:
		m.State = BackendLANDegraded
		return []CmdID{CmdSendActiveDatagram}, nil
	case m.State == BackendLANDegraded && ev == EventLanDatagram:
		m.State = BackendLANDegraded
		return []CmdID{CmdDeliverRecvDatagram}, nil
	case m.State == BackendLANDegraded && ev == EventRelayDatagram:
		m.State = BackendLANDegraded
		return []CmdID{CmdDeliverRecvDatagram}, nil
	case m.State == BackendRelayBackoff && ev == EventAppSendDatagram:
		m.State = BackendRelayBackoff
		return []CmdID{CmdSendActiveDatagram}, nil
	case m.State == BackendRelayBackoff && ev == EventRelayDatagram:
		m.State = BackendRelayBackoff
		return []CmdID{CmdDeliverRecvDatagram}, nil
	case m.State == BackendRelayConnected && ev == EventLanServerReady:
		m.State = BackendLANOffered
		return []CmdID{CmdSendLanOffer}, nil
	case m.State == BackendLANOffered && ev == EventRecvLanVerify && m.Guards[GuardChallengeValid] != nil && m.Guards[GuardChallengeValid]():
		if fn := m.Actions[ActionActivateLan]; fn != nil {
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
		m.State = BackendLANActive
		return []CmdID{CmdSendLanConfirm, CmdStartLanStreamReader, CmdStartLanDgReader, CmdStartMonitor, CmdSignalLanReady, CmdSetCryptoDatagram}, nil
	case m.State == BackendLANOffered && ev == EventRecvLanVerify && m.Guards[GuardChallengeInvalid] != nil && m.Guards[GuardChallengeInvalid]():
		m.State = BackendRelayConnected
		return nil, nil
	case m.State == BackendLANOffered && ev == EventOfferTimeout:
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = BackendRelayBackoff
		return []CmdID{CmdResetLanReady, CmdStartBackoffTimer}, nil
	case m.State == BackendLANActive && ev == EventPingTick:
		m.State = BackendLANActive
		return []CmdID{CmdSendPathPing}, nil
	case m.State == BackendLANActive && ev == EventPingTimeout:
		m.PingFailures = 1
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = BackendLANDegraded
		return nil, nil
	case m.State == BackendLANDegraded && ev == EventPingTick:
		m.State = BackendLANDegraded
		return []CmdID{CmdSendPathPing}, nil
	case m.State == BackendLANDegraded && ev == EventRecvPathPong:
		if fn := m.Actions[ActionResetFailures]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.PingFailures = 0
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = BackendLANActive
		return nil, nil
	case m.State == BackendLANDegraded && ev == EventPingTimeout && m.Guards[GuardUnderMaxFailures] != nil && m.Guards[GuardUnderMaxFailures]():
		// ping_failures: ping_failures + 1 (set by action)
		m.State = BackendLANDegraded
		return nil, nil
	case m.State == BackendLANDegraded && ev == EventPingTimeout && m.Guards[GuardAtMaxFailures] != nil && m.Guards[GuardAtMaxFailures]():
		if fn := m.Actions[ActionFallbackToRelay]; fn != nil {
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
		m.State = BackendRelayBackoff
		return []CmdID{CmdStopMonitor, CmdStopLanStreamReader, CmdStopLanDgReader, CmdCloseLanPath, CmdResetLanReady, CmdStartBackoffTimer}, nil
	case m.State == BackendRelayBackoff && ev == EventBackoffExpired:
		m.State = BackendLANOffered
		return []CmdID{CmdSendLanOffer}, nil
	case m.State == BackendRelayBackoff && ev == EventLanServerChanged:
		m.BackoffLevel = 0
		if m.OnChange != nil { m.OnChange("backoff_level") }
		m.State = BackendLANOffered
		return []CmdID{CmdSendLanOffer}, nil
	case m.State == BackendRelayConnected && ev == EventReadvertiseTick && m.Guards[GuardLanServerAvailable] != nil && m.Guards[GuardLanServerAvailable]():
		m.State = BackendLANOffered
		return []CmdID{CmdSendLanOffer}, nil
	case m.State == BackendRelayConnected && ev == EventDisconnect:
		m.State = BackendPaired
		return nil, nil
	}
	return nil, nil
}

// ClientMachine is the generated state machine for the client actor.
type ClientMachine struct {
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

func NewClientMachine() *ClientMachine {
	return &ClientMachine{
		State: ClientIdle,
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

func (m *ClientMachine) HandleMessage(msg MsgType) (bool, error) {
	switch {
	case m.State == ClientWaitAck && msg == MsgPairHelloAck:
		if fn := m.Actions[ActionDeriveSecret]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		// received_backend_pub: recv_msg.pubkey (set by action)
		// client_shared_key: DeriveKey("client_pub", recv_msg.pubkey) (set by action)
		m.State = ClientE2EReady
		return true, nil
	case m.State == ClientE2EReady && msg == MsgPairConfirm:
		// client_code: DeriveCode(received_backend_pub, "client_pub") (set by action)
		m.State = ClientShowCode
		return true, nil
	case m.State == ClientWaitPairComplete && msg == MsgPairComplete:
		if fn := m.Actions[ActionStoreSecret]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.State = ClientPaired
		return true, nil
	case m.State == ClientSendAuth && msg == MsgAuthOk:
		m.State = ClientSessionActive
		return true, nil
	case m.State == ClientRelayConnected && msg == MsgLanOffer && m.Guards[GuardLanEnabled] != nil && m.Guards[GuardLanEnabled]():
		if fn := m.Actions[ActionDialLan]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.State = ClientLANConnecting
		return true, nil
	case m.State == ClientRelayConnected && msg == MsgLanOffer && m.Guards[GuardLanDisabled] != nil && m.Guards[GuardLanDisabled]():
		m.State = ClientRelayConnected
		return true, nil
	case m.State == ClientLANVerifying && msg == MsgLanConfirm:
		if fn := m.Actions[ActionActivateLan]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.CActivePath = "lan"
		if m.OnChange != nil { m.OnChange("c_active_path") }
		m.CDispatcherPath = "lan"
		if m.OnChange != nil { m.OnChange("c_dispatcher_path") }
		m.LanSignal = "ready"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = ClientLANActive
		return true, nil
	case m.State == ClientLANActive && msg == MsgPathPing:
		m.State = ClientLANActive
		return true, nil
	case m.State == ClientLANActive && msg == MsgLanOffer && m.Guards[GuardLanEnabled] != nil && m.Guards[GuardLanEnabled]():
		if fn := m.Actions[ActionDialLan]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.State = ClientLANConnecting
		return true, nil
	}
	return false, nil
}

func (m *ClientMachine) Step(event EventID) (bool, error) {
	switch {
	case m.State == ClientIdle && event == EventBackchannelReceived:
		m.State = ClientObtainBackchannelSecret
		return true, nil
	case m.State == ClientObtainBackchannelSecret && event == EventSecretParsed:
		m.State = ClientConnectRelay
		return true, nil
	case m.State == ClientConnectRelay && event == EventRelayConnected:
		m.State = ClientGenKeyPair
		return true, nil
	case m.State == ClientGenKeyPair && event == EventKeyPairGenerated:
		if fn := m.Actions[ActionSendPairHello]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.State = ClientWaitAck
		return true, nil
	case m.State == ClientShowCode && event == EventCodeDisplayed:
		m.State = ClientWaitPairComplete
		return true, nil
	case m.State == ClientPaired && event == EventAppLaunch:
		m.State = ClientReconnect
		return true, nil
	case m.State == ClientReconnect && event == EventRelayConnected:
		m.State = ClientSendAuth
		return true, nil
	case m.State == ClientSessionActive && event == EventSessionEstablished:
		m.State = ClientRelayConnected
		return true, nil
	case m.State == ClientRelayConnected && event == EventAppSend:
		m.State = ClientRelayConnected
		return true, nil
	case m.State == ClientRelayConnected && event == EventRelayStreamData:
		m.State = ClientRelayConnected
		return true, nil
	case m.State == ClientLANConnecting && event == EventAppSend:
		m.State = ClientLANConnecting
		return true, nil
	case m.State == ClientLANConnecting && event == EventRelayStreamData:
		m.State = ClientLANConnecting
		return true, nil
	case m.State == ClientLANVerifying && event == EventAppSend:
		m.State = ClientLANVerifying
		return true, nil
	case m.State == ClientLANVerifying && event == EventRelayStreamData:
		m.State = ClientLANVerifying
		return true, nil
	case m.State == ClientLANActive && event == EventAppSend:
		m.State = ClientLANActive
		return true, nil
	case m.State == ClientLANActive && event == EventLanStreamData:
		m.State = ClientLANActive
		return true, nil
	case m.State == ClientLANActive && event == EventRelayStreamData:
		m.State = ClientLANActive
		return true, nil
	case m.State == ClientRelayFallback && event == EventAppSend:
		m.State = ClientRelayFallback
		return true, nil
	case m.State == ClientRelayFallback && event == EventRelayStreamData:
		m.State = ClientRelayFallback
		return true, nil
	case m.State == ClientRelayConnected && event == EventAppSendDatagram:
		m.State = ClientRelayConnected
		return true, nil
	case m.State == ClientRelayConnected && event == EventRelayDatagram:
		m.State = ClientRelayConnected
		return true, nil
	case m.State == ClientLANConnecting && event == EventAppSendDatagram:
		m.State = ClientLANConnecting
		return true, nil
	case m.State == ClientLANConnecting && event == EventRelayDatagram:
		m.State = ClientLANConnecting
		return true, nil
	case m.State == ClientLANVerifying && event == EventAppSendDatagram:
		m.State = ClientLANVerifying
		return true, nil
	case m.State == ClientLANVerifying && event == EventRelayDatagram:
		m.State = ClientLANVerifying
		return true, nil
	case m.State == ClientLANActive && event == EventAppSendDatagram:
		m.State = ClientLANActive
		return true, nil
	case m.State == ClientLANActive && event == EventLanDatagram:
		m.State = ClientLANActive
		return true, nil
	case m.State == ClientLANActive && event == EventRelayDatagram:
		m.State = ClientLANActive
		return true, nil
	case m.State == ClientRelayFallback && event == EventAppSendDatagram:
		m.State = ClientRelayFallback
		return true, nil
	case m.State == ClientRelayFallback && event == EventRelayDatagram:
		m.State = ClientRelayFallback
		return true, nil
	case m.State == ClientLANConnecting && event == EventLanDialOk:
		m.State = ClientLANVerifying
		return true, nil
	case m.State == ClientLANConnecting && event == EventLanDialFailed:
		m.State = ClientRelayConnected
		return true, nil
	case m.State == ClientLANVerifying && event == EventVerifyTimeout:
		m.CDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("c_dispatcher_path") }
		m.State = ClientRelayConnected
		return true, nil
	case m.State == ClientLANActive && event == EventLanError:
		if fn := m.Actions[ActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.CActivePath = "relay"
		if m.OnChange != nil { m.OnChange("c_active_path") }
		m.CDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("c_dispatcher_path") }
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = ClientRelayFallback
		return true, nil
	case m.State == ClientRelayFallback && event == EventRelayOk:
		m.State = ClientRelayConnected
		return true, nil
	case m.State == ClientRelayConnected && event == EventDisconnect:
		m.State = ClientPaired
		return true, nil
	}
	return false, nil
}

func (m *ClientMachine) HandleEvent(ev EventID) ([]CmdID, error) {
	switch {
	case m.State == ClientIdle && ev == EventBackchannelReceived:
		m.State = ClientObtainBackchannelSecret
		return nil, nil
	case m.State == ClientObtainBackchannelSecret && ev == EventSecretParsed:
		m.State = ClientConnectRelay
		return nil, nil
	case m.State == ClientConnectRelay && ev == EventRelayConnected:
		m.State = ClientGenKeyPair
		return nil, nil
	case m.State == ClientGenKeyPair && ev == EventKeyPairGenerated:
		if fn := m.Actions[ActionSendPairHello]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.State = ClientWaitAck
		return nil, nil
	case m.State == ClientWaitAck && ev == EventRecvPairHelloAck:
		if fn := m.Actions[ActionDeriveSecret]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		// received_backend_pub: recv_msg.pubkey (set by action)
		// client_shared_key: DeriveKey("client_pub", recv_msg.pubkey) (set by action)
		m.State = ClientE2EReady
		return nil, nil
	case m.State == ClientE2EReady && ev == EventRecvPairConfirm:
		// client_code: DeriveCode(received_backend_pub, "client_pub") (set by action)
		m.State = ClientShowCode
		return nil, nil
	case m.State == ClientShowCode && ev == EventCodeDisplayed:
		m.State = ClientWaitPairComplete
		return nil, nil
	case m.State == ClientWaitPairComplete && ev == EventRecvPairComplete:
		if fn := m.Actions[ActionStoreSecret]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.State = ClientPaired
		return nil, nil
	case m.State == ClientPaired && ev == EventAppLaunch:
		m.State = ClientReconnect
		return nil, nil
	case m.State == ClientReconnect && ev == EventRelayConnected:
		m.State = ClientSendAuth
		return nil, nil
	case m.State == ClientSendAuth && ev == EventRecvAuthOk:
		m.State = ClientSessionActive
		return nil, nil
	case m.State == ClientSessionActive && ev == EventSessionEstablished:
		m.State = ClientRelayConnected
		return nil, nil
	case m.State == ClientRelayConnected && ev == EventAppSend:
		m.State = ClientRelayConnected
		return []CmdID{CmdWriteActiveStream}, nil
	case m.State == ClientRelayConnected && ev == EventRelayStreamData:
		m.State = ClientRelayConnected
		return []CmdID{CmdDeliverRecv}, nil
	case m.State == ClientLANConnecting && ev == EventAppSend:
		m.State = ClientLANConnecting
		return []CmdID{CmdWriteActiveStream}, nil
	case m.State == ClientLANConnecting && ev == EventRelayStreamData:
		m.State = ClientLANConnecting
		return []CmdID{CmdDeliverRecv}, nil
	case m.State == ClientLANVerifying && ev == EventAppSend:
		m.State = ClientLANVerifying
		return []CmdID{CmdWriteActiveStream}, nil
	case m.State == ClientLANVerifying && ev == EventRelayStreamData:
		m.State = ClientLANVerifying
		return []CmdID{CmdDeliverRecv}, nil
	case m.State == ClientLANActive && ev == EventAppSend:
		m.State = ClientLANActive
		return []CmdID{CmdWriteActiveStream}, nil
	case m.State == ClientLANActive && ev == EventLanStreamData:
		m.State = ClientLANActive
		return []CmdID{CmdDeliverRecv}, nil
	case m.State == ClientLANActive && ev == EventRelayStreamData:
		m.State = ClientLANActive
		return []CmdID{CmdDeliverRecv}, nil
	case m.State == ClientRelayFallback && ev == EventAppSend:
		m.State = ClientRelayFallback
		return []CmdID{CmdWriteActiveStream}, nil
	case m.State == ClientRelayFallback && ev == EventRelayStreamData:
		m.State = ClientRelayFallback
		return []CmdID{CmdDeliverRecv}, nil
	case m.State == ClientRelayConnected && ev == EventAppSendDatagram:
		m.State = ClientRelayConnected
		return []CmdID{CmdSendActiveDatagram}, nil
	case m.State == ClientRelayConnected && ev == EventRelayDatagram:
		m.State = ClientRelayConnected
		return []CmdID{CmdDeliverRecvDatagram}, nil
	case m.State == ClientLANConnecting && ev == EventAppSendDatagram:
		m.State = ClientLANConnecting
		return []CmdID{CmdSendActiveDatagram}, nil
	case m.State == ClientLANConnecting && ev == EventRelayDatagram:
		m.State = ClientLANConnecting
		return []CmdID{CmdDeliverRecvDatagram}, nil
	case m.State == ClientLANVerifying && ev == EventAppSendDatagram:
		m.State = ClientLANVerifying
		return []CmdID{CmdSendActiveDatagram}, nil
	case m.State == ClientLANVerifying && ev == EventRelayDatagram:
		m.State = ClientLANVerifying
		return []CmdID{CmdDeliverRecvDatagram}, nil
	case m.State == ClientLANActive && ev == EventAppSendDatagram:
		m.State = ClientLANActive
		return []CmdID{CmdSendActiveDatagram}, nil
	case m.State == ClientLANActive && ev == EventLanDatagram:
		m.State = ClientLANActive
		return []CmdID{CmdDeliverRecvDatagram}, nil
	case m.State == ClientLANActive && ev == EventRelayDatagram:
		m.State = ClientLANActive
		return []CmdID{CmdDeliverRecvDatagram}, nil
	case m.State == ClientRelayFallback && ev == EventAppSendDatagram:
		m.State = ClientRelayFallback
		return []CmdID{CmdSendActiveDatagram}, nil
	case m.State == ClientRelayFallback && ev == EventRelayDatagram:
		m.State = ClientRelayFallback
		return []CmdID{CmdDeliverRecvDatagram}, nil
	case m.State == ClientRelayConnected && ev == EventRecvLanOffer && m.Guards[GuardLanEnabled] != nil && m.Guards[GuardLanEnabled]():
		if fn := m.Actions[ActionDialLan]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.State = ClientLANConnecting
		return []CmdID{CmdDialLan}, nil
	case m.State == ClientRelayConnected && ev == EventRecvLanOffer && m.Guards[GuardLanDisabled] != nil && m.Guards[GuardLanDisabled]():
		m.State = ClientRelayConnected
		return nil, nil
	case m.State == ClientLANConnecting && ev == EventLanDialOk:
		m.State = ClientLANVerifying
		return []CmdID{CmdSendLanVerify}, nil
	case m.State == ClientLANConnecting && ev == EventLanDialFailed:
		m.State = ClientRelayConnected
		return nil, nil
	case m.State == ClientLANVerifying && ev == EventRecvLanConfirm:
		if fn := m.Actions[ActionActivateLan]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.CActivePath = "lan"
		if m.OnChange != nil { m.OnChange("c_active_path") }
		m.CDispatcherPath = "lan"
		if m.OnChange != nil { m.OnChange("c_dispatcher_path") }
		m.LanSignal = "ready"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = ClientLANActive
		return []CmdID{CmdStartLanStreamReader, CmdStartLanDgReader, CmdSignalLanReady, CmdSetCryptoDatagram}, nil
	case m.State == ClientLANVerifying && ev == EventVerifyTimeout:
		m.CDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("c_dispatcher_path") }
		m.State = ClientRelayConnected
		return nil, nil
	case m.State == ClientLANActive && ev == EventRecvPathPing:
		m.State = ClientLANActive
		return []CmdID{CmdSendPathPong}, nil
	case m.State == ClientLANActive && ev == EventLanError:
		if fn := m.Actions[ActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.CActivePath = "relay"
		if m.OnChange != nil { m.OnChange("c_active_path") }
		m.CDispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("c_dispatcher_path") }
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = ClientRelayFallback
		return []CmdID{CmdStopLanStreamReader, CmdStopLanDgReader, CmdCloseLanPath, CmdResetLanReady}, nil
	case m.State == ClientRelayFallback && ev == EventRelayOk:
		m.State = ClientRelayConnected
		return nil, nil
	case m.State == ClientLANActive && ev == EventRecvLanOffer && m.Guards[GuardLanEnabled] != nil && m.Guards[GuardLanEnabled]():
		if fn := m.Actions[ActionDialLan]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.State = ClientLANConnecting
		return []CmdID{CmdStopLanStreamReader, CmdStopLanDgReader, CmdCloseLanPath, CmdDialLan}, nil
	case m.State == ClientRelayConnected && ev == EventDisconnect:
		m.State = ClientPaired
		return nil, nil
	}
	return nil, nil
}

// RelayMachine is the generated state machine for the relay actor.
type RelayMachine struct {
	State State
	RelayBridge string // relay bridge state

	Guards  map[GuardID]func() bool
	Actions map[ActionID]func() error
	OnChange func(varName string)
}

func NewRelayMachine() *RelayMachine {
	return &RelayMachine{
		State: RelayIdle,
		RelayBridge: "idle",
		Guards:  make(map[GuardID]func() bool),
		Actions: make(map[ActionID]func() error),
	}
}

func (m *RelayMachine) HandleMessage(msg MsgType) (bool, error) {
	switch {
	}
	return false, nil
}

func (m *RelayMachine) Step(event EventID) (bool, error) {
	switch {
	case m.State == RelayIdle && event == EventBackendRegister:
		m.State = RelayBackendRegistered
		return true, nil
	case m.State == RelayBackendRegistered && event == EventClientConnect:
		if fn := m.Actions[ActionBridgeStreams]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.RelayBridge = "active"
		if m.OnChange != nil { m.OnChange("relay_bridge") }
		m.State = RelayBridged
		return true, nil
	case m.State == RelayBridged && event == EventClientDisconnect:
		if fn := m.Actions[ActionUnbridge]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.RelayBridge = "idle"
		if m.OnChange != nil { m.OnChange("relay_bridge") }
		m.State = RelayBackendRegistered
		return true, nil
	case m.State == RelayBackendRegistered && event == EventBackendDisconnect:
		m.State = RelayIdle
		return true, nil
	}
	return false, nil
}

func (m *RelayMachine) HandleEvent(ev EventID) ([]CmdID, error) {
	switch {
	case m.State == RelayIdle && ev == EventBackendRegister:
		m.State = RelayBackendRegistered
		return nil, nil
	case m.State == RelayBackendRegistered && ev == EventClientConnect:
		if fn := m.Actions[ActionBridgeStreams]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.RelayBridge = "active"
		if m.OnChange != nil { m.OnChange("relay_bridge") }
		m.State = RelayBridged
		return nil, nil
	case m.State == RelayBridged && ev == EventClientDisconnect:
		if fn := m.Actions[ActionUnbridge]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.RelayBridge = "idle"
		if m.OnChange != nil { m.OnChange("relay_bridge") }
		m.State = RelayBackendRegistered
		return nil, nil
	case m.State == RelayBackendRegistered && ev == EventBackendDisconnect:
		m.State = RelayIdle
		return nil, nil
	}
	return nil, nil
}

