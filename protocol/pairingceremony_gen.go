// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Code generated from protocol/*.yaml. DO NOT EDIT.

package protocol

// PairingCeremony server states.
const (
	PairingCeremonyServerIdle State = "Idle"
	PairingCeremonyServerGenerateToken State = "GenerateToken"
	PairingCeremonyServerRegisterRelay State = "RegisterRelay"
	PairingCeremonyServerWaitingForClient State = "WaitingForClient"
	PairingCeremonyServerDeriveSecret State = "DeriveSecret"
	PairingCeremonyServerSendAck State = "SendAck"
	PairingCeremonyServerWaitingForCode State = "WaitingForCode"
	PairingCeremonyServerValidateCode State = "ValidateCode"
	PairingCeremonyServerStorePaired State = "StorePaired"
	PairingCeremonyServerPaired State = "Paired"
	PairingCeremonyServerAuthCheck State = "AuthCheck"
	PairingCeremonyServerSessionActive State = "SessionActive"
)

// PairingCeremony ios states.
const (
	PairingCeremonyAppIdle State = "Idle"
	PairingCeremonyAppScanQR State = "ScanQR"
	PairingCeremonyAppConnectRelay State = "ConnectRelay"
	PairingCeremonyAppGenKeyPair State = "GenKeyPair"
	PairingCeremonyAppWaitAck State = "WaitAck"
	PairingCeremonyAppE2EReady State = "E2EReady"
	PairingCeremonyAppShowCode State = "ShowCode"
	PairingCeremonyAppWaitPairComplete State = "WaitPairComplete"
	PairingCeremonyAppPaired State = "Paired"
	PairingCeremonyAppReconnect State = "Reconnect"
	PairingCeremonyAppSendAuth State = "SendAuth"
	PairingCeremonyAppSessionActive State = "SessionActive"
)

// PairingCeremony cli states.
const (
	PairingCeremonyCLIIdle State = "Idle"
	PairingCeremonyCLIGetKey State = "GetKey"
	PairingCeremonyCLIBeginPair State = "BeginPair"
	PairingCeremonyCLIShowQR State = "ShowQR"
	PairingCeremonyCLIPromptCode State = "PromptCode"
	PairingCeremonyCLISubmitCode State = "SubmitCode"
	PairingCeremonyCLIDone State = "Done"
)

// PairingCeremony message types.
const (
	PairingCeremonyMsgPairBegin MsgType = "pair_begin"
	PairingCeremonyMsgTokenResponse MsgType = "token_response"
	PairingCeremonyMsgPairHello MsgType = "pair_hello"
	PairingCeremonyMsgPairHelloAck MsgType = "pair_hello_ack"
	PairingCeremonyMsgPairConfirm MsgType = "pair_confirm"
	PairingCeremonyMsgWaitingForCode MsgType = "waiting_for_code"
	PairingCeremonyMsgCodeSubmit MsgType = "code_submit"
	PairingCeremonyMsgPairComplete MsgType = "pair_complete"
	PairingCeremonyMsgPairStatus MsgType = "pair_status"
	PairingCeremonyMsgAuthRequest MsgType = "auth_request"
	PairingCeremonyMsgAuthOk MsgType = "auth_ok"
)

// PairingCeremony guards.
const (
	PairingCeremonyGuardTokenValid GuardID = "token_valid"
	PairingCeremonyGuardTokenInvalid GuardID = "token_invalid"
	PairingCeremonyGuardCodeCorrect GuardID = "code_correct"
	PairingCeremonyGuardCodeWrong GuardID = "code_wrong"
	PairingCeremonyGuardDeviceKnown GuardID = "device_known"
	PairingCeremonyGuardDeviceUnknown GuardID = "device_unknown"
	PairingCeremonyGuardNonceFresh GuardID = "nonce_fresh"
)

// PairingCeremony actions.
const (
	PairingCeremonyActionDeriveSecret ActionID = "derive_secret"
	PairingCeremonyActionGenerateToken ActionID = "generate_token"
	PairingCeremonyActionRegisterRelay ActionID = "register_relay"
	PairingCeremonyActionSendPairHello ActionID = "send_pair_hello"
	PairingCeremonyActionStoreDevice ActionID = "store_device"
	PairingCeremonyActionStoreSecret ActionID = "store_secret"
	PairingCeremonyActionVerifyDevice ActionID = "verify_device"
)

// PairingCeremony events.
const (
	PairingCeremonyEventECDHComplete EventID = "ECDH complete"
	PairingCeremonyEventQRParsed EventID = "QR parsed"
	PairingCeremonyEventAppLaunch EventID = "app launch"
	PairingCeremonyEventCheckCode EventID = "check code"
	PairingCeremonyEventCliInit EventID = "cli --init"
	PairingCeremonyEventCodeDisplayed EventID = "code displayed"
	PairingCeremonyEventDisconnect EventID = "disconnect"
	PairingCeremonyEventFinalise EventID = "finalise"
	PairingCeremonyEventKeyPairGenerated EventID = "key pair generated"
	PairingCeremonyEventKeyStored EventID = "key stored"
	PairingCeremonyEventRecvAuthOk EventID = "recv_auth_ok"
	PairingCeremonyEventRecvAuthRequest EventID = "recv_auth_request"
	PairingCeremonyEventRecvCodeSubmit EventID = "recv_code_submit"
	PairingCeremonyEventRecvPairBegin EventID = "recv_pair_begin"
	PairingCeremonyEventRecvPairComplete EventID = "recv_pair_complete"
	PairingCeremonyEventRecvPairConfirm EventID = "recv_pair_confirm"
	PairingCeremonyEventRecvPairHello EventID = "recv_pair_hello"
	PairingCeremonyEventRecvPairHelloAck EventID = "recv_pair_hello_ack"
	PairingCeremonyEventRecvPairStatus EventID = "recv_pair_status"
	PairingCeremonyEventRecvTokenResponse EventID = "recv_token_response"
	PairingCeremonyEventRecvWaitingForCode EventID = "recv_waiting_for_code"
	PairingCeremonyEventRelayConnected EventID = "relay connected"
	PairingCeremonyEventRelayRegistered EventID = "relay registered"
	PairingCeremonyEventSignalCodeDisplay EventID = "signal code display"
	PairingCeremonyEventTokenCreated EventID = "token created"
	PairingCeremonyEventUserEntersCode EventID = "user enters code"
	PairingCeremonyEventUserScansQR EventID = "user scans QR"
	PairingCeremonyEventVerify EventID = "verify"
)

func PairingCeremony() *Protocol {
	return &Protocol{
		Name: "PairingCeremony",
		Actors: []Actor{
			{Name: "server", Initial: "Idle", Transitions: []Transition{
				{From: "Idle", To: "GenerateToken", On: Recv("pair_begin"), Do: "generate_token", Updates: []VarUpdate{{Var: "current_token", Expr: "\"tok_1\""}, {Var: "active_tokens", Expr: "active_tokens \\union {\"tok_1\"}"}, }},
				{From: "GenerateToken", To: "RegisterRelay", On: Internal("token created"), Do: "register_relay"},
				{From: "RegisterRelay", To: "WaitingForClient", On: Internal("relay registered"), Sends: []Send{{To: "cli", Msg: "token_response", Fields: map[string]string{"instance_id": "\"inst_1\"", "token": "current_token", }}, }},
				{From: "WaitingForClient", To: "DeriveSecret", On: Recv("pair_hello"), Guard: "token_valid", Do: "derive_secret", Updates: []VarUpdate{{Var: "received_client_pub", Expr: "recv_msg.pubkey"}, {Var: "server_ecdh_pub", Expr: "\"server_pub\""}, {Var: "server_shared_key", Expr: "DeriveKey(\"server_pub\", recv_msg.pubkey)"}, {Var: "server_code", Expr: "DeriveCode(\"server_pub\", recv_msg.pubkey)"}, }},
				{From: "WaitingForClient", To: "Idle", On: Recv("pair_hello"), Guard: "token_invalid"},
				{From: "DeriveSecret", To: "SendAck", On: Internal("ECDH complete"), Sends: []Send{{To: "ios", Msg: "pair_hello_ack", Fields: map[string]string{"pubkey": "server_ecdh_pub", }}, }},
				{From: "SendAck", To: "WaitingForCode", On: Internal("signal code display"), Sends: []Send{{To: "ios", Msg: "pair_confirm"}, {To: "cli", Msg: "waiting_for_code"}, }},
				{From: "WaitingForCode", To: "ValidateCode", On: Recv("code_submit"), Updates: []VarUpdate{{Var: "received_code", Expr: "recv_msg.code"}, }},
				{From: "ValidateCode", To: "StorePaired", On: Internal("check code"), Guard: "code_correct"},
				{From: "ValidateCode", To: "Idle", On: Internal("check code"), Guard: "code_wrong", Updates: []VarUpdate{{Var: "code_attempts", Expr: "code_attempts + 1"}, }},
				{From: "StorePaired", To: "Paired", On: Internal("finalise"), Do: "store_device", Sends: []Send{{To: "ios", Msg: "pair_complete", Fields: map[string]string{"key": "server_shared_key", "secret": "\"dev_secret_1\"", }}, {To: "cli", Msg: "pair_status", Fields: map[string]string{"status": "\"paired\"", }}, }, Updates: []VarUpdate{{Var: "device_secret", Expr: "\"dev_secret_1\""}, {Var: "paired_devices", Expr: "paired_devices \\union {\"device_1\"}"}, {Var: "active_tokens", Expr: "active_tokens \\ {current_token}"}, {Var: "used_tokens", Expr: "used_tokens \\union {current_token}"}, }},
				{From: "Paired", To: "AuthCheck", On: Recv("auth_request"), Updates: []VarUpdate{{Var: "received_device_id", Expr: "recv_msg.device_id"}, {Var: "received_auth_nonce", Expr: "recv_msg.nonce"}, }},
				{From: "AuthCheck", To: "SessionActive", On: Internal("verify"), Guard: "device_known", Do: "verify_device", Sends: []Send{{To: "ios", Msg: "auth_ok"}, }, Updates: []VarUpdate{{Var: "auth_nonces_used", Expr: "auth_nonces_used \\union {received_auth_nonce}"}, }},
				{From: "AuthCheck", To: "Idle", On: Internal("verify"), Guard: "device_unknown"},
				{From: "SessionActive", To: "Paired", On: Internal("disconnect")},
			}},
			{Name: "ios", Initial: "Idle", Transitions: []Transition{
				{From: "Idle", To: "ScanQR", On: Internal("user scans QR")},
				{From: "ScanQR", To: "ConnectRelay", On: Internal("QR parsed")},
				{From: "ConnectRelay", To: "GenKeyPair", On: Internal("relay connected")},
				{From: "GenKeyPair", To: "WaitAck", On: Internal("key pair generated"), Do: "send_pair_hello", Sends: []Send{{To: "server", Msg: "pair_hello", Fields: map[string]string{"pubkey": "\"client_pub\"", "token": "current_token", }}, }},
				{From: "WaitAck", To: "E2EReady", On: Recv("pair_hello_ack"), Do: "derive_secret", Updates: []VarUpdate{{Var: "received_server_pub", Expr: "recv_msg.pubkey"}, {Var: "client_shared_key", Expr: "DeriveKey(\"client_pub\", recv_msg.pubkey)"}, }},
				{From: "E2EReady", To: "ShowCode", On: Recv("pair_confirm"), Updates: []VarUpdate{{Var: "ios_code", Expr: "DeriveCode(received_server_pub, \"client_pub\")"}, }},
				{From: "ShowCode", To: "WaitPairComplete", On: Internal("code displayed")},
				{From: "WaitPairComplete", To: "Paired", On: Recv("pair_complete"), Do: "store_secret"},
				{From: "Paired", To: "Reconnect", On: Internal("app launch")},
				{From: "Reconnect", To: "SendAuth", On: Internal("relay connected"), Sends: []Send{{To: "server", Msg: "auth_request", Fields: map[string]string{"device_id": "\"device_1\"", "key": "client_shared_key", "nonce": "\"nonce_1\"", "secret": "device_secret", }}, }},
				{From: "SendAuth", To: "SessionActive", On: Recv("auth_ok")},
				{From: "SessionActive", To: "Paired", On: Internal("disconnect")},
			}},
			{Name: "cli", Initial: "Idle", Transitions: []Transition{
				{From: "Idle", To: "GetKey", On: Internal("cli --init")},
				{From: "GetKey", To: "BeginPair", On: Internal("key stored"), Sends: []Send{{To: "server", Msg: "pair_begin"}, }},
				{From: "BeginPair", To: "ShowQR", On: Recv("token_response")},
				{From: "ShowQR", To: "PromptCode", On: Recv("waiting_for_code")},
				{From: "PromptCode", To: "SubmitCode", On: Internal("user enters code"), Sends: []Send{{To: "server", Msg: "code_submit", Fields: map[string]string{"code": "ios_code", }}, }},
				{From: "SubmitCode", To: "Done", On: Recv("pair_status")},
			}},
		},
		Messages: []Message{
			{Type: "pair_begin", From: "cli", To: "server", Desc: "POST /api/pair/begin"},
			{Type: "token_response", From: "server", To: "cli", Desc: "{instance_id, pairing_token}"},
			{Type: "pair_hello", From: "ios", To: "server", Desc: "ECDH pubkey + pairing token"},
			{Type: "pair_hello_ack", From: "server", To: "ios", Desc: "ECDH pubkey"},
			{Type: "pair_confirm", From: "server", To: "ios", Desc: "signal to compute and display code"},
			{Type: "waiting_for_code", From: "server", To: "cli", Desc: "prompt for code entry"},
			{Type: "code_submit", From: "cli", To: "server", Desc: "POST /api/pair/confirm"},
			{Type: "pair_complete", From: "server", To: "ios", Desc: "encrypted device secret"},
			{Type: "pair_status", From: "server", To: "cli", Desc: "status: paired"},
			{Type: "auth_request", From: "ios", To: "server", Desc: "encrypted auth with nonce"},
			{Type: "auth_ok", From: "server", To: "ios", Desc: "session established"},
		},
		Vars: []VarDef{
			{Name: "current_token", Initial: "\"none\"", Desc: "pairing token currently in play"},
			{Name: "active_tokens", Initial: "{}", Desc: "set of valid (non-revoked) tokens"},
			{Name: "used_tokens", Initial: "{}", Desc: "set of revoked tokens"},
			{Name: "server_ecdh_pub", Initial: "\"none\"", Desc: "server ECDH public key"},
			{Name: "received_client_pub", Initial: "\"none\"", Desc: "pubkey server received in pair_hello (may be adversary's)"},
			{Name: "received_server_pub", Initial: "\"none\"", Desc: "pubkey ios received in pair_hello_ack (may be adversary's)"},
			{Name: "server_shared_key", Initial: "<<\"none\">>", Desc: "ECDH key derived by server (tuple to match DeriveKey output type)"},
			{Name: "client_shared_key", Initial: "<<\"none\">>", Desc: "ECDH key derived by ios (tuple to match DeriveKey output type)"},
			{Name: "server_code", Initial: "<<\"none\">>", Desc: "code computed by server from its view of the pubkeys (tuple to match DeriveCode output type)"},
			{Name: "ios_code", Initial: "<<\"none\">>", Desc: "code computed by ios from its view of the pubkeys (tuple to match DeriveCode output type)"},
			{Name: "received_code", Initial: "<<\"none\">>", Desc: "code received in code_submit (tuple to match DeriveCode output type)"},
			{Name: "code_attempts", Initial: "0", Desc: "failed code submission attempts"},
			{Name: "device_secret", Initial: "\"none\"", Desc: "persistent device secret"},
			{Name: "paired_devices", Initial: "{}", Desc: "device IDs that completed pairing"},
			{Name: "received_device_id", Initial: "\"none\"", Desc: "device_id from auth_request"},
			{Name: "auth_nonces_used", Initial: "{}", Desc: "set of consumed auth nonces"},
			{Name: "received_auth_nonce", Initial: "\"none\"", Desc: "nonce from auth_request"},
			{Name: "adversary_keys", Initial: "{}", Desc: "encryption keys the adversary knows"},
			{Name: "adv_ecdh_pub", Initial: "\"adv_pub\"", Desc: "adversary's ECDH public key"},
			{Name: "adv_saved_client_pub", Initial: "\"none\"", Desc: "real client pubkey saved during MitM"},
			{Name: "adv_saved_server_pub", Initial: "\"none\"", Desc: "real server pubkey saved during MitM"},
			{Name: "recv_msg", Initial: "[type |-> \"none\"]", Desc: "last received message (staging)"},
		},
		Guards: []GuardDef{
			{ID: "token_valid", Expr: "recv_msg.token \\in active_tokens"},
			{ID: "token_invalid", Expr: "recv_msg.token \\notin active_tokens"},
			{ID: "code_correct", Expr: "received_code = server_code"},
			{ID: "code_wrong", Expr: "received_code /= server_code"},
			{ID: "device_known", Expr: "received_device_id \\in paired_devices"},
			{ID: "device_unknown", Expr: "received_device_id \\notin paired_devices"},
			{ID: "nonce_fresh", Expr: "received_auth_nonce \\notin auth_nonces_used"},
		},
		Operators: []Operator{
			{Name: "KeyRank", Params: "k", Expr: "CASE k = \"adv_pub\" -> 0 [] k = \"client_pub\" -> 1 [] k = \"server_pub\" -> 2 [] OTHER -> 3", Desc: "Assign numeric rank to pubkey names for deterministic ordering"},
			{Name: "DeriveKey", Params: "a, b", Expr: "IF KeyRank(a) <= KeyRank(b) THEN <<\"ecdh\", a, b>> ELSE <<\"ecdh\", b, a>>", Desc: "Symbolic ECDH: deterministic key from two public keys (order-independent)"},
			{Name: "DeriveCode", Params: "a, b", Expr: "IF KeyRank(a) <= KeyRank(b) THEN <<\"code\", a, b>> ELSE <<\"code\", b, a>>", Desc: "Key-bound confirmation code: deterministic from both pubkeys (order-independent)"},
		},
		AdvActions: []AdvAction{
			{Name: "QR_shoulder_surf", Desc: "observe QR code content (token + instance_id)", Code: "      await current_token /= \"none\";\n      adversary_knowledge := adversary_knowledge \\union {[type |-> \"qr_token\", token |-> current_token]};"},
			{Name: "MitM_pair_hello", Desc: "intercept pair_hello and substitute adversary ECDH pubkey", Code: "      await Len(chan_ios_server) > 0 /\\ Head(chan_ios_server).type = MSG_pair_hello;\n      adv_saved_client_pub := Head(chan_ios_server).pubkey;\n      chan_ios_server := <<[type |-> MSG_pair_hello, token |-> Head(chan_ios_server).token, pubkey |-> adv_ecdh_pub]>> \\o Tail(chan_ios_server);"},
			{Name: "MitM_pair_hello_ack", Desc: "intercept pair_hello_ack and substitute adversary ECDH pubkey, derive both shared secrets", Code: "      await Len(chan_server_ios) > 0 /\\ Head(chan_server_ios).type = MSG_pair_hello_ack;\n      adv_saved_server_pub := Head(chan_server_ios).pubkey;\n      adversary_keys := adversary_keys \\union {DeriveKey(adv_ecdh_pub, adv_saved_server_pub), DeriveKey(adv_ecdh_pub, adv_saved_client_pub)};\n      chan_server_ios := <<[type |-> MSG_pair_hello_ack, pubkey |-> adv_ecdh_pub]>> \\o Tail(chan_server_ios);"},
			{Name: "MitM_reencrypt_secret", Desc: "decrypt pair_complete with MitM key, learn device secret", Code: "      await Len(chan_server_ios) > 0 /\\ Head(chan_server_ios).type = MSG_pair_complete /\\ Head(chan_server_ios).key \\in adversary_keys;\n      with msg = Head(chan_server_ios) do\n        adversary_knowledge := adversary_knowledge \\union {[type |-> \"plaintext_secret\", secret |-> msg.secret]};\n        chan_server_ios := <<[type |-> MSG_pair_complete, key |-> DeriveKey(adv_ecdh_pub, adv_saved_client_pub), secret |-> msg.secret]>> \\o Tail(chan_server_ios);\n      end with;"},
			{Name: "concurrent_pair", Desc: "race a forged pair_hello using shoulder-surfed token", Code: "      await \\E m \\in adversary_knowledge : m = [type |-> \"qr_token\", token |-> current_token];\n      await Len(chan_ios_server) < 3;\n      chan_ios_server := Append(chan_ios_server, [type |-> MSG_pair_hello, token |-> current_token, pubkey |-> adv_ecdh_pub]);"},
			{Name: "token_bruteforce", Desc: "send pair_hello with fabricated token", Code: "      await Len(chan_ios_server) < 3;\n      chan_ios_server := Append(chan_ios_server, [type |-> MSG_pair_hello, token |-> \"fake_token\", pubkey |-> adv_ecdh_pub]);"},
			{Name: "code_guess", Desc: "submit fabricated confirmation code via CLI channel", Code: "      await Len(chan_cli_server) < 3;\n      chan_cli_server := Append(chan_cli_server, [type |-> MSG_code_submit, code |-> <<\"guess\", \"000000\">>]);"},
			{Name: "session_replay", Desc: "replay captured auth_request with stale nonce", Code: "      await Len(chan_ios_server) < 3;\n      await \\E m \\in adversary_knowledge : m.type = MSG_auth_request;\n      with msg \\in {m \\in adversary_knowledge : m.type = MSG_auth_request} do\n        chan_ios_server := Append(chan_ios_server, msg);\n      end with;"},
		},
		Properties: []Property{
			{Name: "NoTokenReuse", Kind: Invariant, Expr: "used_tokens \\intersect active_tokens = {}", Desc: "A revoked pairing token is never accepted again"},
			{Name: "MitMDetectedByCodeMismatch", Kind: Invariant, Expr: "(server_shared_key \\in adversary_keys /\\ server_code /= <<\"none\">> /\\ ios_code /= <<\"none\">>) => server_code /= ios_code", Desc: "If the current session's shared key is compromised and both sides computed codes, the codes differ"},
			{Name: "MitMPrevented", Kind: Invariant, Expr: "server_shared_key \\in adversary_keys => server_state \\notin {server_StorePaired, server_Paired, server_AuthCheck, server_SessionActive}", Desc: "If the current session's key is compromised, pairing never completes"},
			{Name: "AuthRequiresCompletedPairing", Kind: Invariant, Expr: "server_state = server_SessionActive => received_device_id \\in paired_devices", Desc: "A session is only active for a device that completed pairing"},
			{Name: "NoNonceReuse", Kind: Invariant, Expr: "server_state = server_SessionActive => received_auth_nonce \\notin (auth_nonces_used \\ {received_auth_nonce})", Desc: "Each auth nonce is accepted at most once"},
			{Name: "WrongCodeDoesNotPair", Kind: Invariant, Expr: "(server_state = server_StorePaired \\/ server_state = server_Paired) => received_code = server_code \\/ received_code = <<\"none\">>", Desc: "Pairing only completes with the correct confirmation code"},
			{Name: "DeviceSecretSecrecy", Kind: Invariant, Expr: "\\A m \\in adversary_knowledge : \"type\" \\in DOMAIN m => m.type /= \"plaintext_secret\"", Desc: "Adversary never learns the device secret in plaintext"},
			{Name: "HonestPairingCompletes", Kind: Liveness, Expr: "cli_state = cli_Done /\\ ios_state = ios_Paired", Desc: "If all actors cooperate honestly (no MitM), pairing eventually completes"},
		},
		ChannelBound: 3,
		OneShot: true,
	}
}

// PairingCeremonyServerMachine is the generated state machine for the server actor.
type PairingCeremonyServerMachine struct {
	State State
	CurrentToken string // pairing token currently in play
	ActiveTokens string // set of valid (non-revoked) tokens
	UsedTokens string // set of revoked tokens
	ServerEcdhPub string // server ECDH public key
	ReceivedClientPub string // pubkey server received in pair_hello (may be adversary's)
	ServerSharedKey string // ECDH key derived by server (tuple to match DeriveKey output type)
	ServerCode string // code computed by server from its view of the pubkeys (tuple to match DeriveCode output type)
	ReceivedCode string // code received in code_submit (tuple to match DeriveCode output type)
	CodeAttempts int // failed code submission attempts
	DeviceSecret string // persistent device secret
	PairedDevices string // device IDs that completed pairing
	ReceivedDeviceId string // device_id from auth_request
	AuthNoncesUsed string // set of consumed auth nonces
	ReceivedAuthNonce string // nonce from auth_request

	Guards  map[GuardID]func() bool
	Actions map[ActionID]func() error
	OnChange func(varName string)
}

func NewPairingCeremonyServerMachine() *PairingCeremonyServerMachine {
	return &PairingCeremonyServerMachine{
		State: PairingCeremonyServerIdle,
		CurrentToken: "none",
		ActiveTokens: "",
		UsedTokens: "",
		ServerEcdhPub: "none",
		ReceivedClientPub: "none",
		ServerSharedKey: "",
		ServerCode: "",
		ReceivedCode: "",
		CodeAttempts: 0,
		DeviceSecret: "none",
		PairedDevices: "",
		ReceivedDeviceId: "none",
		AuthNoncesUsed: "",
		ReceivedAuthNonce: "none",
		Guards:  make(map[GuardID]func() bool),
		Actions: make(map[ActionID]func() error),
	}
}

func (m *PairingCeremonyServerMachine) HandleMessage(msg MsgType) (bool, error) {
	switch {
	case m.State == PairingCeremonyServerIdle && msg == PairingCeremonyMsgPairBegin:
		if fn := m.Actions[PairingCeremonyActionGenerateToken]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.CurrentToken = "tok_1"
		if m.OnChange != nil { m.OnChange("current_token") }
		// active_tokens: active_tokens \union {"tok_1"} (set by action)
		m.State = PairingCeremonyServerGenerateToken
		return true, nil
	case m.State == PairingCeremonyServerWaitingForClient && msg == PairingCeremonyMsgPairHello && m.Guards[PairingCeremonyGuardTokenValid] != nil && m.Guards[PairingCeremonyGuardTokenValid]():
		if fn := m.Actions[PairingCeremonyActionDeriveSecret]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		// received_client_pub: recv_msg.pubkey (set by action)
		m.ServerEcdhPub = "server_pub"
		if m.OnChange != nil { m.OnChange("server_ecdh_pub") }
		// server_shared_key: DeriveKey("server_pub", recv_msg.pubkey) (set by action)
		// server_code: DeriveCode("server_pub", recv_msg.pubkey) (set by action)
		m.State = PairingCeremonyServerDeriveSecret
		return true, nil
	case m.State == PairingCeremonyServerWaitingForClient && msg == PairingCeremonyMsgPairHello && m.Guards[PairingCeremonyGuardTokenInvalid] != nil && m.Guards[PairingCeremonyGuardTokenInvalid]():
		m.State = PairingCeremonyServerIdle
		return true, nil
	case m.State == PairingCeremonyServerWaitingForCode && msg == PairingCeremonyMsgCodeSubmit:
		// received_code: recv_msg.code (set by action)
		m.State = PairingCeremonyServerValidateCode
		return true, nil
	case m.State == PairingCeremonyServerPaired && msg == PairingCeremonyMsgAuthRequest:
		// received_device_id: recv_msg.device_id (set by action)
		// received_auth_nonce: recv_msg.nonce (set by action)
		m.State = PairingCeremonyServerAuthCheck
		return true, nil
	}
	return false, nil
}

func (m *PairingCeremonyServerMachine) Step(event EventID) (bool, error) {
	switch {
	case m.State == PairingCeremonyServerGenerateToken && event == PairingCeremonyEventTokenCreated:
		if fn := m.Actions[PairingCeremonyActionRegisterRelay]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.State = PairingCeremonyServerRegisterRelay
		return true, nil
	case m.State == PairingCeremonyServerRegisterRelay && event == PairingCeremonyEventRelayRegistered:
		m.State = PairingCeremonyServerWaitingForClient
		return true, nil
	case m.State == PairingCeremonyServerDeriveSecret && event == PairingCeremonyEventECDHComplete:
		m.State = PairingCeremonyServerSendAck
		return true, nil
	case m.State == PairingCeremonyServerSendAck && event == PairingCeremonyEventSignalCodeDisplay:
		m.State = PairingCeremonyServerWaitingForCode
		return true, nil
	case m.State == PairingCeremonyServerValidateCode && event == PairingCeremonyEventCheckCode && m.Guards[PairingCeremonyGuardCodeCorrect] != nil && m.Guards[PairingCeremonyGuardCodeCorrect]():
		m.State = PairingCeremonyServerStorePaired
		return true, nil
	case m.State == PairingCeremonyServerValidateCode && event == PairingCeremonyEventCheckCode && m.Guards[PairingCeremonyGuardCodeWrong] != nil && m.Guards[PairingCeremonyGuardCodeWrong]():
		m.CodeAttempts = m.CodeAttempts + 1
		if m.OnChange != nil { m.OnChange("code_attempts") }
		m.State = PairingCeremonyServerIdle
		return true, nil
	case m.State == PairingCeremonyServerStorePaired && event == PairingCeremonyEventFinalise:
		if fn := m.Actions[PairingCeremonyActionStoreDevice]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.DeviceSecret = "dev_secret_1"
		if m.OnChange != nil { m.OnChange("device_secret") }
		// paired_devices: paired_devices \union {"device_1"} (set by action)
		// active_tokens: active_tokens \ {current_token} (set by action)
		// used_tokens: used_tokens \union {current_token} (set by action)
		m.State = PairingCeremonyServerPaired
		return true, nil
	case m.State == PairingCeremonyServerAuthCheck && event == PairingCeremonyEventVerify && m.Guards[PairingCeremonyGuardDeviceKnown] != nil && m.Guards[PairingCeremonyGuardDeviceKnown]():
		if fn := m.Actions[PairingCeremonyActionVerifyDevice]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		// auth_nonces_used: auth_nonces_used \union {received_auth_nonce} (set by action)
		m.State = PairingCeremonyServerSessionActive
		return true, nil
	case m.State == PairingCeremonyServerAuthCheck && event == PairingCeremonyEventVerify && m.Guards[PairingCeremonyGuardDeviceUnknown] != nil && m.Guards[PairingCeremonyGuardDeviceUnknown]():
		m.State = PairingCeremonyServerIdle
		return true, nil
	case m.State == PairingCeremonyServerSessionActive && event == PairingCeremonyEventDisconnect:
		m.State = PairingCeremonyServerPaired
		return true, nil
	}
	return false, nil
}

func (m *PairingCeremonyServerMachine) HandleEvent(ev EventID) ([]CmdID, error) {
	switch {
	case m.State == PairingCeremonyServerIdle && ev == PairingCeremonyEventRecvPairBegin:
		if fn := m.Actions[PairingCeremonyActionGenerateToken]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.CurrentToken = "tok_1"
		if m.OnChange != nil { m.OnChange("current_token") }
		// active_tokens: active_tokens \union {"tok_1"} (set by action)
		m.State = PairingCeremonyServerGenerateToken
		return nil, nil
	case m.State == PairingCeremonyServerGenerateToken && ev == PairingCeremonyEventTokenCreated:
		if fn := m.Actions[PairingCeremonyActionRegisterRelay]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.State = PairingCeremonyServerRegisterRelay
		return nil, nil
	case m.State == PairingCeremonyServerRegisterRelay && ev == PairingCeremonyEventRelayRegistered:
		m.State = PairingCeremonyServerWaitingForClient
		return nil, nil
	case m.State == PairingCeremonyServerWaitingForClient && ev == PairingCeremonyEventRecvPairHello && m.Guards[PairingCeremonyGuardTokenValid] != nil && m.Guards[PairingCeremonyGuardTokenValid]():
		if fn := m.Actions[PairingCeremonyActionDeriveSecret]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		// received_client_pub: recv_msg.pubkey (set by action)
		m.ServerEcdhPub = "server_pub"
		if m.OnChange != nil { m.OnChange("server_ecdh_pub") }
		// server_shared_key: DeriveKey("server_pub", recv_msg.pubkey) (set by action)
		// server_code: DeriveCode("server_pub", recv_msg.pubkey) (set by action)
		m.State = PairingCeremonyServerDeriveSecret
		return nil, nil
	case m.State == PairingCeremonyServerWaitingForClient && ev == PairingCeremonyEventRecvPairHello && m.Guards[PairingCeremonyGuardTokenInvalid] != nil && m.Guards[PairingCeremonyGuardTokenInvalid]():
		m.State = PairingCeremonyServerIdle
		return nil, nil
	case m.State == PairingCeremonyServerDeriveSecret && ev == PairingCeremonyEventECDHComplete:
		m.State = PairingCeremonyServerSendAck
		return nil, nil
	case m.State == PairingCeremonyServerSendAck && ev == PairingCeremonyEventSignalCodeDisplay:
		m.State = PairingCeremonyServerWaitingForCode
		return nil, nil
	case m.State == PairingCeremonyServerWaitingForCode && ev == PairingCeremonyEventRecvCodeSubmit:
		// received_code: recv_msg.code (set by action)
		m.State = PairingCeremonyServerValidateCode
		return nil, nil
	case m.State == PairingCeremonyServerValidateCode && ev == PairingCeremonyEventCheckCode && m.Guards[PairingCeremonyGuardCodeCorrect] != nil && m.Guards[PairingCeremonyGuardCodeCorrect]():
		m.State = PairingCeremonyServerStorePaired
		return nil, nil
	case m.State == PairingCeremonyServerValidateCode && ev == PairingCeremonyEventCheckCode && m.Guards[PairingCeremonyGuardCodeWrong] != nil && m.Guards[PairingCeremonyGuardCodeWrong]():
		m.CodeAttempts = m.CodeAttempts + 1
		if m.OnChange != nil { m.OnChange("code_attempts") }
		m.State = PairingCeremonyServerIdle
		return nil, nil
	case m.State == PairingCeremonyServerStorePaired && ev == PairingCeremonyEventFinalise:
		if fn := m.Actions[PairingCeremonyActionStoreDevice]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.DeviceSecret = "dev_secret_1"
		if m.OnChange != nil { m.OnChange("device_secret") }
		// paired_devices: paired_devices \union {"device_1"} (set by action)
		// active_tokens: active_tokens \ {current_token} (set by action)
		// used_tokens: used_tokens \union {current_token} (set by action)
		m.State = PairingCeremonyServerPaired
		return nil, nil
	case m.State == PairingCeremonyServerPaired && ev == PairingCeremonyEventRecvAuthRequest:
		// received_device_id: recv_msg.device_id (set by action)
		// received_auth_nonce: recv_msg.nonce (set by action)
		m.State = PairingCeremonyServerAuthCheck
		return nil, nil
	case m.State == PairingCeremonyServerAuthCheck && ev == PairingCeremonyEventVerify && m.Guards[PairingCeremonyGuardDeviceKnown] != nil && m.Guards[PairingCeremonyGuardDeviceKnown]():
		if fn := m.Actions[PairingCeremonyActionVerifyDevice]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		// auth_nonces_used: auth_nonces_used \union {received_auth_nonce} (set by action)
		m.State = PairingCeremonyServerSessionActive
		return nil, nil
	case m.State == PairingCeremonyServerAuthCheck && ev == PairingCeremonyEventVerify && m.Guards[PairingCeremonyGuardDeviceUnknown] != nil && m.Guards[PairingCeremonyGuardDeviceUnknown]():
		m.State = PairingCeremonyServerIdle
		return nil, nil
	case m.State == PairingCeremonyServerSessionActive && ev == PairingCeremonyEventDisconnect:
		m.State = PairingCeremonyServerPaired
		return nil, nil
	}
	return nil, nil
}

// PairingCeremonyAppMachine is the generated state machine for the ios actor.
type PairingCeremonyAppMachine struct {
	State State
	ReceivedServerPub string // pubkey ios received in pair_hello_ack (may be adversary's)
	ClientSharedKey string // ECDH key derived by ios (tuple to match DeriveKey output type)
	IosCode string // code computed by ios from its view of the pubkeys (tuple to match DeriveCode output type)

	Guards  map[GuardID]func() bool
	Actions map[ActionID]func() error
	OnChange func(varName string)
}

func NewPairingCeremonyAppMachine() *PairingCeremonyAppMachine {
	return &PairingCeremonyAppMachine{
		State: PairingCeremonyAppIdle,
		ReceivedServerPub: "none",
		ClientSharedKey: "",
		IosCode: "",
		Guards:  make(map[GuardID]func() bool),
		Actions: make(map[ActionID]func() error),
	}
}

func (m *PairingCeremonyAppMachine) HandleMessage(msg MsgType) (bool, error) {
	switch {
	case m.State == PairingCeremonyAppWaitAck && msg == PairingCeremonyMsgPairHelloAck:
		if fn := m.Actions[PairingCeremonyActionDeriveSecret]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		// received_server_pub: recv_msg.pubkey (set by action)
		// client_shared_key: DeriveKey("client_pub", recv_msg.pubkey) (set by action)
		m.State = PairingCeremonyAppE2EReady
		return true, nil
	case m.State == PairingCeremonyAppE2EReady && msg == PairingCeremonyMsgPairConfirm:
		// ios_code: DeriveCode(received_server_pub, "client_pub") (set by action)
		m.State = PairingCeremonyAppShowCode
		return true, nil
	case m.State == PairingCeremonyAppWaitPairComplete && msg == PairingCeremonyMsgPairComplete:
		if fn := m.Actions[PairingCeremonyActionStoreSecret]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.State = PairingCeremonyAppPaired
		return true, nil
	case m.State == PairingCeremonyAppSendAuth && msg == PairingCeremonyMsgAuthOk:
		m.State = PairingCeremonyAppSessionActive
		return true, nil
	}
	return false, nil
}

func (m *PairingCeremonyAppMachine) Step(event EventID) (bool, error) {
	switch {
	case m.State == PairingCeremonyAppIdle && event == PairingCeremonyEventUserScansQR:
		m.State = PairingCeremonyAppScanQR
		return true, nil
	case m.State == PairingCeremonyAppScanQR && event == PairingCeremonyEventQRParsed:
		m.State = PairingCeremonyAppConnectRelay
		return true, nil
	case m.State == PairingCeremonyAppConnectRelay && event == PairingCeremonyEventRelayConnected:
		m.State = PairingCeremonyAppGenKeyPair
		return true, nil
	case m.State == PairingCeremonyAppGenKeyPair && event == PairingCeremonyEventKeyPairGenerated:
		if fn := m.Actions[PairingCeremonyActionSendPairHello]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.State = PairingCeremonyAppWaitAck
		return true, nil
	case m.State == PairingCeremonyAppShowCode && event == PairingCeremonyEventCodeDisplayed:
		m.State = PairingCeremonyAppWaitPairComplete
		return true, nil
	case m.State == PairingCeremonyAppPaired && event == PairingCeremonyEventAppLaunch:
		m.State = PairingCeremonyAppReconnect
		return true, nil
	case m.State == PairingCeremonyAppReconnect && event == PairingCeremonyEventRelayConnected:
		m.State = PairingCeremonyAppSendAuth
		return true, nil
	case m.State == PairingCeremonyAppSessionActive && event == PairingCeremonyEventDisconnect:
		m.State = PairingCeremonyAppPaired
		return true, nil
	}
	return false, nil
}

func (m *PairingCeremonyAppMachine) HandleEvent(ev EventID) ([]CmdID, error) {
	switch {
	case m.State == PairingCeremonyAppIdle && ev == PairingCeremonyEventUserScansQR:
		m.State = PairingCeremonyAppScanQR
		return nil, nil
	case m.State == PairingCeremonyAppScanQR && ev == PairingCeremonyEventQRParsed:
		m.State = PairingCeremonyAppConnectRelay
		return nil, nil
	case m.State == PairingCeremonyAppConnectRelay && ev == PairingCeremonyEventRelayConnected:
		m.State = PairingCeremonyAppGenKeyPair
		return nil, nil
	case m.State == PairingCeremonyAppGenKeyPair && ev == PairingCeremonyEventKeyPairGenerated:
		if fn := m.Actions[PairingCeremonyActionSendPairHello]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.State = PairingCeremonyAppWaitAck
		return nil, nil
	case m.State == PairingCeremonyAppWaitAck && ev == PairingCeremonyEventRecvPairHelloAck:
		if fn := m.Actions[PairingCeremonyActionDeriveSecret]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		// received_server_pub: recv_msg.pubkey (set by action)
		// client_shared_key: DeriveKey("client_pub", recv_msg.pubkey) (set by action)
		m.State = PairingCeremonyAppE2EReady
		return nil, nil
	case m.State == PairingCeremonyAppE2EReady && ev == PairingCeremonyEventRecvPairConfirm:
		// ios_code: DeriveCode(received_server_pub, "client_pub") (set by action)
		m.State = PairingCeremonyAppShowCode
		return nil, nil
	case m.State == PairingCeremonyAppShowCode && ev == PairingCeremonyEventCodeDisplayed:
		m.State = PairingCeremonyAppWaitPairComplete
		return nil, nil
	case m.State == PairingCeremonyAppWaitPairComplete && ev == PairingCeremonyEventRecvPairComplete:
		if fn := m.Actions[PairingCeremonyActionStoreSecret]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.State = PairingCeremonyAppPaired
		return nil, nil
	case m.State == PairingCeremonyAppPaired && ev == PairingCeremonyEventAppLaunch:
		m.State = PairingCeremonyAppReconnect
		return nil, nil
	case m.State == PairingCeremonyAppReconnect && ev == PairingCeremonyEventRelayConnected:
		m.State = PairingCeremonyAppSendAuth
		return nil, nil
	case m.State == PairingCeremonyAppSendAuth && ev == PairingCeremonyEventRecvAuthOk:
		m.State = PairingCeremonyAppSessionActive
		return nil, nil
	case m.State == PairingCeremonyAppSessionActive && ev == PairingCeremonyEventDisconnect:
		m.State = PairingCeremonyAppPaired
		return nil, nil
	}
	return nil, nil
}

// PairingCeremonyCLIMachine is the generated state machine for the cli actor.
type PairingCeremonyCLIMachine struct {
	State State

	Guards  map[GuardID]func() bool
	Actions map[ActionID]func() error
	OnChange func(varName string)
}

func NewPairingCeremonyCLIMachine() *PairingCeremonyCLIMachine {
	return &PairingCeremonyCLIMachine{
		State: PairingCeremonyCLIIdle,
		Guards:  make(map[GuardID]func() bool),
		Actions: make(map[ActionID]func() error),
	}
}

func (m *PairingCeremonyCLIMachine) HandleMessage(msg MsgType) (bool, error) {
	switch {
	case m.State == PairingCeremonyCLIBeginPair && msg == PairingCeremonyMsgTokenResponse:
		m.State = PairingCeremonyCLIShowQR
		return true, nil
	case m.State == PairingCeremonyCLIShowQR && msg == PairingCeremonyMsgWaitingForCode:
		m.State = PairingCeremonyCLIPromptCode
		return true, nil
	case m.State == PairingCeremonyCLISubmitCode && msg == PairingCeremonyMsgPairStatus:
		m.State = PairingCeremonyCLIDone
		return true, nil
	}
	return false, nil
}

func (m *PairingCeremonyCLIMachine) Step(event EventID) (bool, error) {
	switch {
	case m.State == PairingCeremonyCLIIdle && event == PairingCeremonyEventCliInit:
		m.State = PairingCeremonyCLIGetKey
		return true, nil
	case m.State == PairingCeremonyCLIGetKey && event == PairingCeremonyEventKeyStored:
		m.State = PairingCeremonyCLIBeginPair
		return true, nil
	case m.State == PairingCeremonyCLIPromptCode && event == PairingCeremonyEventUserEntersCode:
		m.State = PairingCeremonyCLISubmitCode
		return true, nil
	}
	return false, nil
}

func (m *PairingCeremonyCLIMachine) HandleEvent(ev EventID) ([]CmdID, error) {
	switch {
	case m.State == PairingCeremonyCLIIdle && ev == PairingCeremonyEventCliInit:
		m.State = PairingCeremonyCLIGetKey
		return nil, nil
	case m.State == PairingCeremonyCLIGetKey && ev == PairingCeremonyEventKeyStored:
		m.State = PairingCeremonyCLIBeginPair
		return nil, nil
	case m.State == PairingCeremonyCLIBeginPair && ev == PairingCeremonyEventRecvTokenResponse:
		m.State = PairingCeremonyCLIShowQR
		return nil, nil
	case m.State == PairingCeremonyCLIShowQR && ev == PairingCeremonyEventRecvWaitingForCode:
		m.State = PairingCeremonyCLIPromptCode
		return nil, nil
	case m.State == PairingCeremonyCLIPromptCode && ev == PairingCeremonyEventUserEntersCode:
		m.State = PairingCeremonyCLISubmitCode
		return nil, nil
	case m.State == PairingCeremonyCLISubmitCode && ev == PairingCeremonyEventRecvPairStatus:
		m.State = PairingCeremonyCLIDone
		return nil, nil
	}
	return nil, nil
}

