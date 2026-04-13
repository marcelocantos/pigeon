// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Code generated from protocol/*.yaml. DO NOT EDIT.

package protocol

// PairingCeremony server/pairing states.
const (
	PairingCeremonyServerPairingIdle State = "Idle"
	PairingCeremonyServerPairingGenerateToken State = "GenerateToken"
	PairingCeremonyServerPairingRegisterRelay State = "RegisterRelay"
	PairingCeremonyServerPairingWaitingForClient State = "WaitingForClient"
	PairingCeremonyServerPairingDeriveSecret State = "DeriveSecret"
	PairingCeremonyServerPairingSendAck State = "SendAck"
	PairingCeremonyServerPairingWaitingForCode State = "WaitingForCode"
	PairingCeremonyServerPairingValidateCode State = "ValidateCode"
	PairingCeremonyServerPairingStorePaired State = "StorePaired"
	PairingCeremonyServerPairingPairingComplete State = "PairingComplete"
)

// PairingCeremony server/auth states.
const (
	PairingCeremonyServerAuthIdle State = "Idle"
	PairingCeremonyServerAuthPaired State = "Paired"
	PairingCeremonyServerAuthAuthCheck State = "AuthCheck"
	PairingCeremonyServerAuthSessionActive State = "SessionActive"
)

// PairingCeremony ios/pairing states.
const (
	PairingCeremonyAppPairingIdle State = "Idle"
	PairingCeremonyAppPairingScanQR State = "ScanQR"
	PairingCeremonyAppPairingConnectRelay State = "ConnectRelay"
	PairingCeremonyAppPairingGenKeyPair State = "GenKeyPair"
	PairingCeremonyAppPairingWaitAck State = "WaitAck"
	PairingCeremonyAppPairingE2EReady State = "E2EReady"
	PairingCeremonyAppPairingShowCode State = "ShowCode"
	PairingCeremonyAppPairingWaitPairComplete State = "WaitPairComplete"
	PairingCeremonyAppPairingPairingComplete State = "PairingComplete"
)

// PairingCeremony ios/auth states.
const (
	PairingCeremonyAppAuthIdle State = "Idle"
	PairingCeremonyAppAuthPaired State = "Paired"
	PairingCeremonyAppAuthReconnect State = "Reconnect"
	PairingCeremonyAppAuthSendAuth State = "SendAuth"
	PairingCeremonyAppAuthSessionActive State = "SessionActive"
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
	PairingCeremonyEventCredentialReady EventID = "credential_ready"
	PairingCeremonyEventDisconnect EventID = "disconnect"
	PairingCeremonyEventFinalise EventID = "finalise"
	PairingCeremonyEventKeyPairGenerated EventID = "key pair generated"
	PairingCeremonyEventKeyStored EventID = "key stored"
	PairingCeremonyEventPaired EventID = "paired"
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
			{Name: "server", Machines: []SubMachine{
				{Name: "pairing", Initial: "Idle", Transitions: []Transition{
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
					{From: "StorePaired", To: "PairingComplete", On: Internal("finalise"), Do: "store_device", Sends: []Send{{To: "ios", Msg: "pair_complete", Fields: map[string]string{"key": "server_shared_key", "secret": "\"dev_secret_1\"", }}, {To: "cli", Msg: "pair_status", Fields: map[string]string{"status": "\"paired\"", }}, }, Updates: []VarUpdate{{Var: "device_secret", Expr: "\"dev_secret_1\""}, {Var: "paired_devices", Expr: "paired_devices \\union {\"device_1\"}"}, {Var: "active_tokens", Expr: "active_tokens \\ {current_token}"}, {Var: "used_tokens", Expr: "used_tokens \\union {current_token}"}, }},
				}},
				{Name: "auth", Initial: "Idle", Transitions: []Transition{
					{From: "Idle", To: "Paired", On: Internal("credential_ready")},
					{From: "Paired", To: "AuthCheck", On: Recv("auth_request"), Updates: []VarUpdate{{Var: "received_device_id", Expr: "recv_msg.device_id"}, {Var: "received_auth_nonce", Expr: "recv_msg.nonce"}, }},
					{From: "AuthCheck", To: "SessionActive", On: Internal("verify"), Guard: "device_known", Do: "verify_device", Sends: []Send{{To: "ios", Msg: "auth_ok"}, }, Updates: []VarUpdate{{Var: "auth_nonces_used", Expr: "auth_nonces_used \\union {received_auth_nonce}"}, }},
					{From: "AuthCheck", To: "Idle", On: Internal("verify"), Guard: "device_unknown"},
					{From: "SessionActive", To: "Paired", On: Internal("disconnect")},
				}},
				},
				Routes: []Route{
					{On: "paired", From: "pairing", Sends: []RouteSend{{To: "auth", Event: "credential_ready"}, }},
				},
			},
			{Name: "ios", Machines: []SubMachine{
				{Name: "pairing", Initial: "Idle", Transitions: []Transition{
					{From: "Idle", To: "ScanQR", On: Internal("user scans QR")},
					{From: "ScanQR", To: "ConnectRelay", On: Internal("QR parsed")},
					{From: "ConnectRelay", To: "GenKeyPair", On: Internal("relay connected")},
					{From: "GenKeyPair", To: "WaitAck", On: Internal("key pair generated"), Do: "send_pair_hello", Sends: []Send{{To: "server", Msg: "pair_hello", Fields: map[string]string{"pubkey": "\"client_pub\"", "token": "current_token", }}, }},
					{From: "WaitAck", To: "E2EReady", On: Recv("pair_hello_ack"), Do: "derive_secret", Updates: []VarUpdate{{Var: "received_server_pub", Expr: "recv_msg.pubkey"}, {Var: "client_shared_key", Expr: "DeriveKey(\"client_pub\", recv_msg.pubkey)"}, }},
					{From: "E2EReady", To: "ShowCode", On: Recv("pair_confirm"), Updates: []VarUpdate{{Var: "ios_code", Expr: "DeriveCode(received_server_pub, \"client_pub\")"}, }},
					{From: "ShowCode", To: "WaitPairComplete", On: Internal("code displayed")},
					{From: "WaitPairComplete", To: "PairingComplete", On: Recv("pair_complete"), Do: "store_secret"},
				}},
				{Name: "auth", Initial: "Idle", Transitions: []Transition{
					{From: "Idle", To: "Paired", On: Internal("credential_ready")},
					{From: "Paired", To: "Reconnect", On: Internal("app launch")},
					{From: "Reconnect", To: "SendAuth", On: Internal("relay connected"), Sends: []Send{{To: "server", Msg: "auth_request", Fields: map[string]string{"device_id": "\"device_1\"", "key": "client_shared_key", "nonce": "\"nonce_1\"", "secret": "device_secret", }}, }},
					{From: "SendAuth", To: "SessionActive", On: Recv("auth_ok")},
					{From: "SessionActive", To: "Paired", On: Internal("disconnect")},
				}},
				},
				Routes: []Route{
					{On: "paired", From: "pairing", Sends: []RouteSend{{To: "auth", Event: "credential_ready"}, }},
				},
			},
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
			{Name: "MitMPrevented", Kind: Invariant, Expr: "server_shared_key \\in adversary_keys => server_pairing_state \\notin {\"StorePaired\", \"PairingComplete\"}", Desc: "If the current session's key is compromised, pairing never completes"},
			{Name: "AuthRequiresCompletedPairing", Kind: Invariant, Expr: "server_auth_state = \"SessionActive\" => received_device_id \\in paired_devices", Desc: "A session is only active for a device that completed pairing"},
			{Name: "NoNonceReuse", Kind: Invariant, Expr: "server_auth_state = \"SessionActive\" => received_auth_nonce \\notin (auth_nonces_used \\ {received_auth_nonce})", Desc: "Each auth nonce is accepted at most once"},
			{Name: "WrongCodeDoesNotPair", Kind: Invariant, Expr: "(server_pairing_state = \"StorePaired\" \\/ server_pairing_state = \"PairingComplete\") => received_code = server_code \\/ received_code = <<\"none\">>", Desc: "Pairing only completes with the correct confirmation code"},
			{Name: "DeviceSecretSecrecy", Kind: Invariant, Expr: "\\A m \\in adversary_knowledge : \"type\" \\in DOMAIN m => m.type /= \"plaintext_secret\"", Desc: "Adversary never learns the device secret in plaintext"},
			{Name: "HonestPairingCompletes", Kind: Liveness, Expr: "cli_state = \"Done\" /\\ ios_pairing_state = \"PairingComplete\"", Desc: "If all actors cooperate honestly (no MitM), pairing eventually completes"},
		},
		ChannelBound: 3,
		OneShot: true,
	}
}

// PairingCeremonyServerPairingMachine is the generated state machine for server/pairing.
type PairingCeremonyServerPairingMachine struct {
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

	Guards  map[GuardID]func() bool
	Actions map[ActionID]func() error
	OnChange func(varName string)
}

func NewPairingCeremonyServerPairingMachine() *PairingCeremonyServerPairingMachine {
	return &PairingCeremonyServerPairingMachine{
		State: PairingCeremonyServerPairingIdle,
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
		Guards:  make(map[GuardID]func() bool),
		Actions: make(map[ActionID]func() error),
	}
}

func (m *PairingCeremonyServerPairingMachine) Step(event EventID) (bool, error) {
	switch {
	case m.State == PairingCeremonyServerPairingGenerateToken && event == PairingCeremonyEventTokenCreated:
		if fn := m.Actions[PairingCeremonyActionRegisterRelay]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.State = PairingCeremonyServerPairingRegisterRelay
		return true, nil
	case m.State == PairingCeremonyServerPairingRegisterRelay && event == PairingCeremonyEventRelayRegistered:
		m.State = PairingCeremonyServerPairingWaitingForClient
		return true, nil
	case m.State == PairingCeremonyServerPairingDeriveSecret && event == PairingCeremonyEventECDHComplete:
		m.State = PairingCeremonyServerPairingSendAck
		return true, nil
	case m.State == PairingCeremonyServerPairingSendAck && event == PairingCeremonyEventSignalCodeDisplay:
		m.State = PairingCeremonyServerPairingWaitingForCode
		return true, nil
	case m.State == PairingCeremonyServerPairingValidateCode && event == PairingCeremonyEventCheckCode && m.Guards[PairingCeremonyGuardCodeCorrect] != nil && m.Guards[PairingCeremonyGuardCodeCorrect]():
		m.State = PairingCeremonyServerPairingStorePaired
		return true, nil
	case m.State == PairingCeremonyServerPairingValidateCode && event == PairingCeremonyEventCheckCode && m.Guards[PairingCeremonyGuardCodeWrong] != nil && m.Guards[PairingCeremonyGuardCodeWrong]():
		m.CodeAttempts = m.CodeAttempts + 1
		if m.OnChange != nil { m.OnChange("code_attempts") }
		m.State = PairingCeremonyServerPairingIdle
		return true, nil
	case m.State == PairingCeremonyServerPairingStorePaired && event == PairingCeremonyEventFinalise:
		if fn := m.Actions[PairingCeremonyActionStoreDevice]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.DeviceSecret = "dev_secret_1"
		if m.OnChange != nil { m.OnChange("device_secret") }
		// paired_devices: paired_devices \union {"device_1"} (set by action)
		// active_tokens: active_tokens \ {current_token} (set by action)
		// used_tokens: used_tokens \union {current_token} (set by action)
		m.State = PairingCeremonyServerPairingPairingComplete
		return true, nil
	}
	return false, nil
}

func (m *PairingCeremonyServerPairingMachine) HandleMessage(msg MsgType) (bool, error) {
	switch {
	case m.State == PairingCeremonyServerPairingIdle && msg == PairingCeremonyMsgPairBegin:
		if fn := m.Actions[PairingCeremonyActionGenerateToken]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.CurrentToken = "tok_1"
		if m.OnChange != nil { m.OnChange("current_token") }
		// active_tokens: active_tokens \union {"tok_1"} (set by action)
		m.State = PairingCeremonyServerPairingGenerateToken
		return true, nil
	case m.State == PairingCeremonyServerPairingWaitingForClient && msg == PairingCeremonyMsgPairHello && m.Guards[PairingCeremonyGuardTokenValid] != nil && m.Guards[PairingCeremonyGuardTokenValid]():
		if fn := m.Actions[PairingCeremonyActionDeriveSecret]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		// received_client_pub: recv_msg.pubkey (set by action)
		m.ServerEcdhPub = "server_pub"
		if m.OnChange != nil { m.OnChange("server_ecdh_pub") }
		// server_shared_key: DeriveKey("server_pub", recv_msg.pubkey) (set by action)
		// server_code: DeriveCode("server_pub", recv_msg.pubkey) (set by action)
		m.State = PairingCeremonyServerPairingDeriveSecret
		return true, nil
	case m.State == PairingCeremonyServerPairingWaitingForClient && msg == PairingCeremonyMsgPairHello && m.Guards[PairingCeremonyGuardTokenInvalid] != nil && m.Guards[PairingCeremonyGuardTokenInvalid]():
		m.State = PairingCeremonyServerPairingIdle
		return true, nil
	case m.State == PairingCeremonyServerPairingWaitingForCode && msg == PairingCeremonyMsgCodeSubmit:
		// received_code: recv_msg.code (set by action)
		m.State = PairingCeremonyServerPairingValidateCode
		return true, nil
	}
	return false, nil
}

// PairingCeremonyServerAuthMachine is the generated state machine for server/auth.
type PairingCeremonyServerAuthMachine struct {
	State State
	ReceivedDeviceId string // device_id from auth_request
	AuthNoncesUsed string // set of consumed auth nonces
	ReceivedAuthNonce string // nonce from auth_request

	Guards  map[GuardID]func() bool
	Actions map[ActionID]func() error
	OnChange func(varName string)
}

func NewPairingCeremonyServerAuthMachine() *PairingCeremonyServerAuthMachine {
	return &PairingCeremonyServerAuthMachine{
		State: PairingCeremonyServerAuthIdle,
		ReceivedDeviceId: "none",
		AuthNoncesUsed: "",
		ReceivedAuthNonce: "none",
		Guards:  make(map[GuardID]func() bool),
		Actions: make(map[ActionID]func() error),
	}
}

func (m *PairingCeremonyServerAuthMachine) Step(event EventID) (bool, error) {
	switch {
	case m.State == PairingCeremonyServerAuthIdle && event == PairingCeremonyEventCredentialReady:
		m.State = PairingCeremonyServerAuthPaired
		return true, nil
	case m.State == PairingCeremonyServerAuthAuthCheck && event == PairingCeremonyEventVerify && m.Guards[PairingCeremonyGuardDeviceKnown] != nil && m.Guards[PairingCeremonyGuardDeviceKnown]():
		if fn := m.Actions[PairingCeremonyActionVerifyDevice]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		// auth_nonces_used: auth_nonces_used \union {received_auth_nonce} (set by action)
		m.State = PairingCeremonyServerAuthSessionActive
		return true, nil
	case m.State == PairingCeremonyServerAuthAuthCheck && event == PairingCeremonyEventVerify && m.Guards[PairingCeremonyGuardDeviceUnknown] != nil && m.Guards[PairingCeremonyGuardDeviceUnknown]():
		m.State = PairingCeremonyServerAuthIdle
		return true, nil
	case m.State == PairingCeremonyServerAuthSessionActive && event == PairingCeremonyEventDisconnect:
		m.State = PairingCeremonyServerAuthPaired
		return true, nil
	}
	return false, nil
}

func (m *PairingCeremonyServerAuthMachine) HandleMessage(msg MsgType) (bool, error) {
	switch {
	case m.State == PairingCeremonyServerAuthPaired && msg == PairingCeremonyMsgAuthRequest:
		// received_device_id: recv_msg.device_id (set by action)
		// received_auth_nonce: recv_msg.nonce (set by action)
		m.State = PairingCeremonyServerAuthAuthCheck
		return true, nil
	}
	return false, nil
}

// PairingCeremonyServerComposite holds all sub-machines for the server actor.
type PairingCeremonyServerComposite struct {
	Pairing PairingCeremonyServerPairingMachine
	Auth PairingCeremonyServerAuthMachine

	RouteGuards map[GuardID]func() bool
}

func NewPairingCeremonyServerComposite() *PairingCeremonyServerComposite {
	c := &PairingCeremonyServerComposite{
		RouteGuards: make(map[GuardID]func() bool),
	}
	initPairing := NewPairingCeremonyServerPairingMachine()
	c.Pairing = *initPairing
	initAuth := NewPairingCeremonyServerAuthMachine()
	c.Auth = *initAuth
	return c
}

// Route dispatches inter-machine events according to the routing table.
// Call this after a sub-machine reports an event.
func (c *PairingCeremonyServerComposite) Route(from string, event EventID) error {
	switch {
	case from == "pairing" && event == PairingCeremonyEventPaired:
		if _, err := c.Auth.Step(PairingCeremonyEventCredentialReady); err != nil {
			return err
		}
	}
	return nil
}

// PairingCeremonyAppPairingMachine is the generated state machine for ios/pairing.
type PairingCeremonyAppPairingMachine struct {
	State State
	ReceivedServerPub string // pubkey ios received in pair_hello_ack (may be adversary's)
	ClientSharedKey string // ECDH key derived by ios (tuple to match DeriveKey output type)
	IosCode string // code computed by ios from its view of the pubkeys (tuple to match DeriveCode output type)

	Guards  map[GuardID]func() bool
	Actions map[ActionID]func() error
	OnChange func(varName string)
}

func NewPairingCeremonyAppPairingMachine() *PairingCeremonyAppPairingMachine {
	return &PairingCeremonyAppPairingMachine{
		State: PairingCeremonyAppPairingIdle,
		ReceivedServerPub: "none",
		ClientSharedKey: "",
		IosCode: "",
		Guards:  make(map[GuardID]func() bool),
		Actions: make(map[ActionID]func() error),
	}
}

func (m *PairingCeremonyAppPairingMachine) Step(event EventID) (bool, error) {
	switch {
	case m.State == PairingCeremonyAppPairingIdle && event == PairingCeremonyEventUserScansQR:
		m.State = PairingCeremonyAppPairingScanQR
		return true, nil
	case m.State == PairingCeremonyAppPairingScanQR && event == PairingCeremonyEventQRParsed:
		m.State = PairingCeremonyAppPairingConnectRelay
		return true, nil
	case m.State == PairingCeremonyAppPairingConnectRelay && event == PairingCeremonyEventRelayConnected:
		m.State = PairingCeremonyAppPairingGenKeyPair
		return true, nil
	case m.State == PairingCeremonyAppPairingGenKeyPair && event == PairingCeremonyEventKeyPairGenerated:
		if fn := m.Actions[PairingCeremonyActionSendPairHello]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.State = PairingCeremonyAppPairingWaitAck
		return true, nil
	case m.State == PairingCeremonyAppPairingShowCode && event == PairingCeremonyEventCodeDisplayed:
		m.State = PairingCeremonyAppPairingWaitPairComplete
		return true, nil
	}
	return false, nil
}

func (m *PairingCeremonyAppPairingMachine) HandleMessage(msg MsgType) (bool, error) {
	switch {
	case m.State == PairingCeremonyAppPairingWaitAck && msg == PairingCeremonyMsgPairHelloAck:
		if fn := m.Actions[PairingCeremonyActionDeriveSecret]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		// received_server_pub: recv_msg.pubkey (set by action)
		// client_shared_key: DeriveKey("client_pub", recv_msg.pubkey) (set by action)
		m.State = PairingCeremonyAppPairingE2EReady
		return true, nil
	case m.State == PairingCeremonyAppPairingE2EReady && msg == PairingCeremonyMsgPairConfirm:
		// ios_code: DeriveCode(received_server_pub, "client_pub") (set by action)
		m.State = PairingCeremonyAppPairingShowCode
		return true, nil
	case m.State == PairingCeremonyAppPairingWaitPairComplete && msg == PairingCeremonyMsgPairComplete:
		if fn := m.Actions[PairingCeremonyActionStoreSecret]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.State = PairingCeremonyAppPairingPairingComplete
		return true, nil
	}
	return false, nil
}

// PairingCeremonyAppAuthMachine is the generated state machine for ios/auth.
type PairingCeremonyAppAuthMachine struct {
	State State

	Guards  map[GuardID]func() bool
	Actions map[ActionID]func() error
	OnChange func(varName string)
}

func NewPairingCeremonyAppAuthMachine() *PairingCeremonyAppAuthMachine {
	return &PairingCeremonyAppAuthMachine{
		State: PairingCeremonyAppAuthIdle,
		Guards:  make(map[GuardID]func() bool),
		Actions: make(map[ActionID]func() error),
	}
}

func (m *PairingCeremonyAppAuthMachine) Step(event EventID) (bool, error) {
	switch {
	case m.State == PairingCeremonyAppAuthIdle && event == PairingCeremonyEventCredentialReady:
		m.State = PairingCeremonyAppAuthPaired
		return true, nil
	case m.State == PairingCeremonyAppAuthPaired && event == PairingCeremonyEventAppLaunch:
		m.State = PairingCeremonyAppAuthReconnect
		return true, nil
	case m.State == PairingCeremonyAppAuthReconnect && event == PairingCeremonyEventRelayConnected:
		m.State = PairingCeremonyAppAuthSendAuth
		return true, nil
	case m.State == PairingCeremonyAppAuthSessionActive && event == PairingCeremonyEventDisconnect:
		m.State = PairingCeremonyAppAuthPaired
		return true, nil
	}
	return false, nil
}

func (m *PairingCeremonyAppAuthMachine) HandleMessage(msg MsgType) (bool, error) {
	switch {
	case m.State == PairingCeremonyAppAuthSendAuth && msg == PairingCeremonyMsgAuthOk:
		m.State = PairingCeremonyAppAuthSessionActive
		return true, nil
	}
	return false, nil
}

// PairingCeremonyAppComposite holds all sub-machines for the ios actor.
type PairingCeremonyAppComposite struct {
	Pairing PairingCeremonyAppPairingMachine
	Auth PairingCeremonyAppAuthMachine

	RouteGuards map[GuardID]func() bool
}

func NewPairingCeremonyAppComposite() *PairingCeremonyAppComposite {
	c := &PairingCeremonyAppComposite{
		RouteGuards: make(map[GuardID]func() bool),
	}
	initPairing := NewPairingCeremonyAppPairingMachine()
	c.Pairing = *initPairing
	initAuth := NewPairingCeremonyAppAuthMachine()
	c.Auth = *initAuth
	return c
}

// Route dispatches inter-machine events according to the routing table.
// Call this after a sub-machine reports an event.
func (c *PairingCeremonyAppComposite) Route(from string, event EventID) error {
	switch {
	case from == "pairing" && event == PairingCeremonyEventPaired:
		if _, err := c.Auth.Step(PairingCeremonyEventCredentialReady); err != nil {
			return err
		}
	}
	return nil
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

