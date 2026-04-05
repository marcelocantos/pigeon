---- MODULE PairingCeremony ----
\* Auto-generated from protocol YAML. Do not edit.

EXTENDS Integers, Sequences, FiniteSets, TLC

\* States for server
server_Idle == "server_Idle"
server_GenerateToken == "server_GenerateToken"
server_RegisterRelay == "server_RegisterRelay"
server_WaitingForClient == "server_WaitingForClient"
server_DeriveSecret == "server_DeriveSecret"
server_SendAck == "server_SendAck"
server_WaitingForCode == "server_WaitingForCode"
server_ValidateCode == "server_ValidateCode"
server_StorePaired == "server_StorePaired"
server_Paired == "server_Paired"
server_AuthCheck == "server_AuthCheck"
server_SessionActive == "server_SessionActive"

\* States for ios
ios_Idle == "ios_Idle"
ios_ScanQR == "ios_ScanQR"
ios_ConnectRelay == "ios_ConnectRelay"
ios_GenKeyPair == "ios_GenKeyPair"
ios_WaitAck == "ios_WaitAck"
ios_E2EReady == "ios_E2EReady"
ios_ShowCode == "ios_ShowCode"
ios_WaitPairComplete == "ios_WaitPairComplete"
ios_Paired == "ios_Paired"
ios_Reconnect == "ios_Reconnect"
ios_SendAuth == "ios_SendAuth"
ios_SessionActive == "ios_SessionActive"

\* States for cli
cli_Idle == "cli_Idle"
cli_GetKey == "cli_GetKey"
cli_BeginPair == "cli_BeginPair"
cli_ShowQR == "cli_ShowQR"
cli_PromptCode == "cli_PromptCode"
cli_SubmitCode == "cli_SubmitCode"
cli_Done == "cli_Done"

\* Message types
MSG_pair_begin == "pair_begin"
MSG_token_response == "token_response"
MSG_pair_hello == "pair_hello"
MSG_pair_hello_ack == "pair_hello_ack"
MSG_pair_confirm == "pair_confirm"
MSG_waiting_for_code == "waiting_for_code"
MSG_code_submit == "code_submit"
MSG_pair_complete == "pair_complete"
MSG_pair_status == "pair_status"
MSG_auth_request == "auth_request"
MSG_auth_ok == "auth_ok"

\* Assign numeric rank to pubkey names for deterministic ordering
KeyRank(k) == CASE k = "adv_pub" -> 0 [] k = "client_pub" -> 1 [] k = "server_pub" -> 2 [] OTHER -> 3
\* Symbolic ECDH: deterministic key from two public keys (order-independent)
DeriveKey(a, b) == IF KeyRank(a) <= KeyRank(b) THEN <<"ecdh", a, b>> ELSE <<"ecdh", b, a>>
\* Key-bound confirmation code: deterministic from both pubkeys (order-independent)
DeriveCode(a, b) == IF KeyRank(a) <= KeyRank(b) THEN <<"code", a, b>> ELSE <<"code", b, a>>



CONSTANTS adversary_keys, adv_ecdh_pub, adv_saved_client_pub, adv_saved_server_pub

VARIABLES
    server_state,
    ios_state,
    cli_state,
    current_token,
    active_tokens,
    used_tokens,
    server_ecdh_pub,
    received_client_pub,
    received_server_pub,
    server_shared_key,
    client_shared_key,
    server_code,
    ios_code,
    received_code,
    code_attempts,
    device_secret,
    paired_devices,
    received_device_id,
    auth_nonces_used,
    received_auth_nonce,
    received_pair_begin,
    received_pair_hello,
    received_code_submit,
    received_auth_request,
    received_pair_hello_ack,
    received_pair_confirm,
    received_pair_complete,
    received_auth_ok,
    received_token_response,
    received_waiting_for_code,
    received_pair_status

vars == <<server_state, ios_state, cli_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok, received_token_response, received_waiting_for_code, received_pair_status>>

Init ==
    /\ server_state = server_Idle
    /\ ios_state = ios_Idle
    /\ cli_state = cli_Idle
    /\ current_token = "none"
    /\ active_tokens = {}
    /\ used_tokens = {}
    /\ server_ecdh_pub = "none"
    /\ received_client_pub = "none"
    /\ received_server_pub = "none"
    /\ server_shared_key = <<"none">>
    /\ client_shared_key = <<"none">>
    /\ server_code = <<"none">>
    /\ ios_code = <<"none">>
    /\ received_code = <<"none">>
    /\ code_attempts = 0
    /\ device_secret = "none"
    /\ paired_devices = {}
    /\ received_device_id = "none"
    /\ auth_nonces_used = {}
    /\ received_auth_nonce = "none"
    /\ received_pair_begin = [type |-> "none"]
    /\ received_pair_hello = [type |-> "none"]
    /\ received_code_submit = [type |-> "none"]
    /\ received_auth_request = [type |-> "none"]
    /\ received_pair_hello_ack = [type |-> "none"]
    /\ received_pair_confirm = [type |-> "none"]
    /\ received_pair_complete = [type |-> "none"]
    /\ received_auth_ok = [type |-> "none"]
    /\ received_token_response = [type |-> "none"]
    /\ received_waiting_for_code = [type |-> "none"]
    /\ received_pair_status = [type |-> "none"]

\* server: Idle -> GenerateToken on recv pair_begin
server_Idle_to_GenerateToken_on_pair_begin ==
    /\ server_state = server_Idle
    /\ received_pair_begin.type = MSG_pair_begin
    /\ received_pair_begin' = [type |-> "none"]
    /\ server_state' = server_GenerateToken
    /\ current_token' = "tok_1"
    /\ active_tokens' = active_tokens \union {"tok_1"}
    /\ UNCHANGED <<ios_state, cli_state, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok, received_token_response, received_waiting_for_code, received_pair_status>>

\* server: GenerateToken -> RegisterRelay (token created)
server_GenerateToken_to_RegisterRelay_token_created ==
    /\ server_state = server_GenerateToken
    /\ server_state' = server_RegisterRelay
    /\ UNCHANGED <<ios_state, cli_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok, received_token_response, received_waiting_for_code, received_pair_status>>

\* server: RegisterRelay -> WaitingForClient (relay registered)
server_RegisterRelay_to_WaitingForClient_relay_registered ==
    /\ server_state = server_RegisterRelay
    /\ received_token_response' = [type |-> MSG_token_response, instance_id |-> "inst_1", token |-> current_token]
    /\ server_state' = server_WaitingForClient
    /\ UNCHANGED <<ios_state, cli_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok, received_waiting_for_code, received_pair_status>>

\* server: WaitingForClient -> DeriveSecret on recv pair_hello [token_valid]
server_WaitingForClient_to_DeriveSecret_on_pair_hello_token_valid ==
    /\ server_state = server_WaitingForClient
    /\ received_pair_hello.type = MSG_pair_hello
    /\ received_pair_hello.token \in active_tokens
    /\ received_pair_hello' = [type |-> "none"]
    /\ server_state' = server_DeriveSecret
    /\ received_client_pub' = received_pair_hello.pubkey
    /\ server_ecdh_pub' = "server_pub"
    /\ server_shared_key' = DeriveKey("server_pub", received_pair_hello.pubkey)
    /\ server_code' = DeriveCode("server_pub", received_pair_hello.pubkey)
    /\ UNCHANGED <<ios_state, cli_state, current_token, active_tokens, used_tokens, received_server_pub, client_shared_key, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_begin, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok, received_token_response, received_waiting_for_code, received_pair_status>>

\* server: WaitingForClient -> Idle on recv pair_hello [token_invalid]
server_WaitingForClient_to_Idle_on_pair_hello_token_invalid ==
    /\ server_state = server_WaitingForClient
    /\ received_pair_hello.type = MSG_pair_hello
    /\ received_pair_hello.token \notin active_tokens
    /\ received_pair_hello' = [type |-> "none"]
    /\ server_state' = server_Idle
    /\ UNCHANGED <<ios_state, cli_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_begin, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok, received_token_response, received_waiting_for_code, received_pair_status>>

\* server: DeriveSecret -> SendAck (ECDH complete)
server_DeriveSecret_to_SendAck_ECDH_complete ==
    /\ server_state = server_DeriveSecret
    /\ received_pair_hello_ack' = [type |-> MSG_pair_hello_ack, pubkey |-> server_ecdh_pub]
    /\ server_state' = server_SendAck
    /\ UNCHANGED <<ios_state, cli_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_confirm, received_pair_complete, received_auth_ok, received_token_response, received_waiting_for_code, received_pair_status>>

\* server: SendAck -> WaitingForCode (signal code display)
server_SendAck_to_WaitingForCode_signal_code_display ==
    /\ server_state = server_SendAck
    /\ received_pair_confirm' = [type |-> MSG_pair_confirm]
    /\ received_waiting_for_code' = [type |-> MSG_waiting_for_code]
    /\ server_state' = server_WaitingForCode
    /\ UNCHANGED <<ios_state, cli_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_complete, received_auth_ok, received_token_response, received_pair_status>>

\* server: WaitingForCode -> ValidateCode on recv code_submit
server_WaitingForCode_to_ValidateCode_on_code_submit ==
    /\ server_state = server_WaitingForCode
    /\ received_code_submit.type = MSG_code_submit
    /\ received_code_submit' = [type |-> "none"]
    /\ server_state' = server_ValidateCode
    /\ received_code' = received_code_submit.code
    /\ UNCHANGED <<ios_state, cli_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_begin, received_pair_hello, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok, received_token_response, received_waiting_for_code, received_pair_status>>

\* server: ValidateCode -> StorePaired (check code) [code_correct]
server_ValidateCode_to_StorePaired_check_code_code_correct ==
    /\ server_state = server_ValidateCode
    /\ received_code = server_code
    /\ server_state' = server_StorePaired
    /\ UNCHANGED <<ios_state, cli_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok, received_token_response, received_waiting_for_code, received_pair_status>>

\* server: ValidateCode -> Idle (check code) [code_wrong]
server_ValidateCode_to_Idle_check_code_code_wrong ==
    /\ server_state = server_ValidateCode
    /\ received_code /= server_code
    /\ server_state' = server_Idle
    /\ code_attempts' = code_attempts + 1
    /\ UNCHANGED <<ios_state, cli_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok, received_token_response, received_waiting_for_code, received_pair_status>>

\* server: StorePaired -> Paired (finalise)
server_StorePaired_to_Paired_finalise ==
    /\ server_state = server_StorePaired
    /\ received_pair_complete' = [type |-> MSG_pair_complete, key |-> server_shared_key, secret |-> "dev_secret_1"]
    /\ received_pair_status' = [type |-> MSG_pair_status, status |-> "paired"]
    /\ server_state' = server_Paired
    /\ device_secret' = "dev_secret_1"
    /\ paired_devices' = paired_devices \union {"device_1"}
    /\ active_tokens' = active_tokens \ {current_token}
    /\ used_tokens' = used_tokens \union {current_token}
    /\ UNCHANGED <<ios_state, cli_state, current_token, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_auth_ok, received_token_response, received_waiting_for_code>>

\* server: Paired -> AuthCheck on recv auth_request
server_Paired_to_AuthCheck_on_auth_request ==
    /\ server_state = server_Paired
    /\ received_auth_request.type = MSG_auth_request
    /\ received_auth_request' = [type |-> "none"]
    /\ server_state' = server_AuthCheck
    /\ received_device_id' = received_auth_request.device_id
    /\ received_auth_nonce' = received_auth_request.nonce
    /\ UNCHANGED <<ios_state, cli_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, auth_nonces_used, received_pair_begin, received_pair_hello, received_code_submit, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok, received_token_response, received_waiting_for_code, received_pair_status>>

\* server: AuthCheck -> SessionActive (verify) [device_known]
server_AuthCheck_to_SessionActive_verify_device_known ==
    /\ server_state = server_AuthCheck
    /\ received_device_id \in paired_devices
    /\ received_auth_ok' = [type |-> MSG_auth_ok]
    /\ server_state' = server_SessionActive
    /\ auth_nonces_used' = auth_nonces_used \union {received_auth_nonce}
    /\ UNCHANGED <<ios_state, cli_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, received_auth_nonce, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_token_response, received_waiting_for_code, received_pair_status>>

\* server: AuthCheck -> Idle (verify) [device_unknown]
server_AuthCheck_to_Idle_verify_device_unknown ==
    /\ server_state = server_AuthCheck
    /\ received_device_id \notin paired_devices
    /\ server_state' = server_Idle
    /\ UNCHANGED <<ios_state, cli_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok, received_token_response, received_waiting_for_code, received_pair_status>>

\* server: SessionActive -> Paired (disconnect)
server_SessionActive_to_Paired_disconnect ==
    /\ server_state = server_SessionActive
    /\ server_state' = server_Paired
    /\ UNCHANGED <<ios_state, cli_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok, received_token_response, received_waiting_for_code, received_pair_status>>


\* ios: Idle -> ScanQR (user scans QR)
ios_Idle_to_ScanQR_user_scans_QR ==
    /\ ios_state = ios_Idle
    /\ ios_state' = ios_ScanQR
    /\ UNCHANGED <<server_state, cli_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok, received_token_response, received_waiting_for_code, received_pair_status>>

\* ios: ScanQR -> ConnectRelay (QR parsed)
ios_ScanQR_to_ConnectRelay_QR_parsed ==
    /\ ios_state = ios_ScanQR
    /\ ios_state' = ios_ConnectRelay
    /\ UNCHANGED <<server_state, cli_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok, received_token_response, received_waiting_for_code, received_pair_status>>

\* ios: ConnectRelay -> GenKeyPair (relay connected)
ios_ConnectRelay_to_GenKeyPair_relay_connected ==
    /\ ios_state = ios_ConnectRelay
    /\ ios_state' = ios_GenKeyPair
    /\ UNCHANGED <<server_state, cli_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok, received_token_response, received_waiting_for_code, received_pair_status>>

\* ios: GenKeyPair -> WaitAck (key pair generated)
ios_GenKeyPair_to_WaitAck_key_pair_generated ==
    /\ ios_state = ios_GenKeyPair
    /\ received_pair_hello' = [type |-> MSG_pair_hello, pubkey |-> "client_pub", token |-> current_token]
    /\ ios_state' = ios_WaitAck
    /\ UNCHANGED <<server_state, cli_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_begin, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok, received_token_response, received_waiting_for_code, received_pair_status>>

\* ios: WaitAck -> E2EReady on recv pair_hello_ack
ios_WaitAck_to_E2EReady_on_pair_hello_ack ==
    /\ ios_state = ios_WaitAck
    /\ received_pair_hello_ack.type = MSG_pair_hello_ack
    /\ received_pair_hello_ack' = [type |-> "none"]
    /\ ios_state' = ios_E2EReady
    /\ received_server_pub' = received_pair_hello_ack.pubkey
    /\ client_shared_key' = DeriveKey("client_pub", received_pair_hello_ack.pubkey)
    /\ UNCHANGED <<server_state, cli_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, server_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_confirm, received_pair_complete, received_auth_ok, received_token_response, received_waiting_for_code, received_pair_status>>

\* ios: E2EReady -> ShowCode on recv pair_confirm
ios_E2EReady_to_ShowCode_on_pair_confirm ==
    /\ ios_state = ios_E2EReady
    /\ received_pair_confirm.type = MSG_pair_confirm
    /\ received_pair_confirm' = [type |-> "none"]
    /\ ios_state' = ios_ShowCode
    /\ ios_code' = DeriveCode(received_server_pub, "client_pub")
    /\ UNCHANGED <<server_state, cli_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_complete, received_auth_ok, received_token_response, received_waiting_for_code, received_pair_status>>

\* ios: ShowCode -> WaitPairComplete (code displayed)
ios_ShowCode_to_WaitPairComplete_code_displayed ==
    /\ ios_state = ios_ShowCode
    /\ ios_state' = ios_WaitPairComplete
    /\ UNCHANGED <<server_state, cli_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok, received_token_response, received_waiting_for_code, received_pair_status>>

\* ios: WaitPairComplete -> Paired on recv pair_complete
ios_WaitPairComplete_to_Paired_on_pair_complete ==
    /\ ios_state = ios_WaitPairComplete
    /\ received_pair_complete.type = MSG_pair_complete
    /\ received_pair_complete' = [type |-> "none"]
    /\ ios_state' = ios_Paired
    /\ UNCHANGED <<server_state, cli_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_auth_ok, received_token_response, received_waiting_for_code, received_pair_status>>

\* ios: Paired -> Reconnect (app launch)
ios_Paired_to_Reconnect_app_launch ==
    /\ ios_state = ios_Paired
    /\ ios_state' = ios_Reconnect
    /\ UNCHANGED <<server_state, cli_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok, received_token_response, received_waiting_for_code, received_pair_status>>

\* ios: Reconnect -> SendAuth (relay connected)
ios_Reconnect_to_SendAuth_relay_connected ==
    /\ ios_state = ios_Reconnect
    /\ received_auth_request' = [type |-> MSG_auth_request, device_id |-> "device_1", key |-> client_shared_key, nonce |-> "nonce_1", secret |-> device_secret]
    /\ ios_state' = ios_SendAuth
    /\ UNCHANGED <<server_state, cli_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_begin, received_pair_hello, received_code_submit, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok, received_token_response, received_waiting_for_code, received_pair_status>>

\* ios: SendAuth -> SessionActive on recv auth_ok
ios_SendAuth_to_SessionActive_on_auth_ok ==
    /\ ios_state = ios_SendAuth
    /\ received_auth_ok.type = MSG_auth_ok
    /\ received_auth_ok' = [type |-> "none"]
    /\ ios_state' = ios_SessionActive
    /\ UNCHANGED <<server_state, cli_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_token_response, received_waiting_for_code, received_pair_status>>

\* ios: SessionActive -> Paired (disconnect)
ios_SessionActive_to_Paired_disconnect ==
    /\ ios_state = ios_SessionActive
    /\ ios_state' = ios_Paired
    /\ UNCHANGED <<server_state, cli_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok, received_token_response, received_waiting_for_code, received_pair_status>>


\* cli: Idle -> GetKey (cli --init)
cli_Idle_to_GetKey_cli___init ==
    /\ cli_state = cli_Idle
    /\ cli_state' = cli_GetKey
    /\ UNCHANGED <<server_state, ios_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok, received_token_response, received_waiting_for_code, received_pair_status>>

\* cli: GetKey -> BeginPair (key stored)
cli_GetKey_to_BeginPair_key_stored ==
    /\ cli_state = cli_GetKey
    /\ received_pair_begin' = [type |-> MSG_pair_begin]
    /\ cli_state' = cli_BeginPair
    /\ UNCHANGED <<server_state, ios_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok, received_token_response, received_waiting_for_code, received_pair_status>>

\* cli: BeginPair -> ShowQR on recv token_response
cli_BeginPair_to_ShowQR_on_token_response ==
    /\ cli_state = cli_BeginPair
    /\ received_token_response.type = MSG_token_response
    /\ received_token_response' = [type |-> "none"]
    /\ cli_state' = cli_ShowQR
    /\ UNCHANGED <<server_state, ios_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok, received_waiting_for_code, received_pair_status>>

\* cli: ShowQR -> PromptCode on recv waiting_for_code
cli_ShowQR_to_PromptCode_on_waiting_for_code ==
    /\ cli_state = cli_ShowQR
    /\ received_waiting_for_code.type = MSG_waiting_for_code
    /\ received_waiting_for_code' = [type |-> "none"]
    /\ cli_state' = cli_PromptCode
    /\ UNCHANGED <<server_state, ios_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok, received_token_response, received_pair_status>>

\* cli: PromptCode -> SubmitCode (user enters code)
cli_PromptCode_to_SubmitCode_user_enters_code ==
    /\ cli_state = cli_PromptCode
    /\ received_code_submit' = [type |-> MSG_code_submit, code |-> ios_code]
    /\ cli_state' = cli_SubmitCode
    /\ UNCHANGED <<server_state, ios_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_begin, received_pair_hello, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok, received_token_response, received_waiting_for_code, received_pair_status>>

\* cli: SubmitCode -> Done on recv pair_status
cli_SubmitCode_to_Done_on_pair_status ==
    /\ cli_state = cli_SubmitCode
    /\ received_pair_status.type = MSG_pair_status
    /\ received_pair_status' = [type |-> "none"]
    /\ cli_state' = cli_Done
    /\ UNCHANGED <<server_state, ios_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok, received_token_response, received_waiting_for_code>>


Next ==
    \/ server_Idle_to_GenerateToken_on_pair_begin
    \/ server_GenerateToken_to_RegisterRelay_token_created
    \/ server_RegisterRelay_to_WaitingForClient_relay_registered
    \/ server_WaitingForClient_to_DeriveSecret_on_pair_hello_token_valid
    \/ server_WaitingForClient_to_Idle_on_pair_hello_token_invalid
    \/ server_DeriveSecret_to_SendAck_ECDH_complete
    \/ server_SendAck_to_WaitingForCode_signal_code_display
    \/ server_WaitingForCode_to_ValidateCode_on_code_submit
    \/ server_ValidateCode_to_StorePaired_check_code_code_correct
    \/ server_ValidateCode_to_Idle_check_code_code_wrong
    \/ server_StorePaired_to_Paired_finalise
    \/ server_Paired_to_AuthCheck_on_auth_request
    \/ server_AuthCheck_to_SessionActive_verify_device_known
    \/ server_AuthCheck_to_Idle_verify_device_unknown
    \/ server_SessionActive_to_Paired_disconnect
    \/ ios_Idle_to_ScanQR_user_scans_QR
    \/ ios_ScanQR_to_ConnectRelay_QR_parsed
    \/ ios_ConnectRelay_to_GenKeyPair_relay_connected
    \/ ios_GenKeyPair_to_WaitAck_key_pair_generated
    \/ ios_WaitAck_to_E2EReady_on_pair_hello_ack
    \/ ios_E2EReady_to_ShowCode_on_pair_confirm
    \/ ios_ShowCode_to_WaitPairComplete_code_displayed
    \/ ios_WaitPairComplete_to_Paired_on_pair_complete
    \/ ios_Paired_to_Reconnect_app_launch
    \/ ios_Reconnect_to_SendAuth_relay_connected
    \/ ios_SendAuth_to_SessionActive_on_auth_ok
    \/ ios_SessionActive_to_Paired_disconnect
    \/ cli_Idle_to_GetKey_cli___init
    \/ cli_GetKey_to_BeginPair_key_stored
    \/ cli_BeginPair_to_ShowQR_on_token_response
    \/ cli_ShowQR_to_PromptCode_on_waiting_for_code
    \/ cli_PromptCode_to_SubmitCode_user_enters_code
    \/ cli_SubmitCode_to_Done_on_pair_status

Spec == Init /\ [][Next]_vars /\ WF_vars(Next)

\* ================================================================
\* Invariants and properties
\* ================================================================

\* A revoked pairing token is never accepted again
NoTokenReuse == used_tokens \intersect active_tokens = {}
\* If the current session's shared key is compromised and both sides computed codes, the codes differ
MitMDetectedByCodeMismatch == (server_shared_key \in adversary_keys /\ server_code /= <<"none">> /\ ios_code /= <<"none">>) => server_code /= ios_code
\* If the current session's key is compromised, pairing never completes
MitMPrevented == server_shared_key \in adversary_keys => server_state \notin {server_StorePaired, server_Paired, server_AuthCheck, server_SessionActive}
\* A session is only active for a device that completed pairing
AuthRequiresCompletedPairing == server_state = server_SessionActive => received_device_id \in paired_devices
\* Each auth nonce is accepted at most once
NoNonceReuse == server_state = server_SessionActive => received_auth_nonce \notin (auth_nonces_used \ {received_auth_nonce})
\* Pairing only completes with the correct confirmation code
WrongCodeDoesNotPair == (server_state = server_StorePaired \/ server_state = server_Paired) => received_code = server_code \/ received_code = <<"none">>
\* Adversary never learns the device secret in plaintext
DeviceSecretSecrecy == \A m \in adversary_knowledge : "type" \in DOMAIN m => m.type /= "plaintext_secret"
\* If all actors cooperate honestly (no MitM), pairing eventually completes
HonestPairingCompletes == <>(cli_state = cli_Done /\ ios_state = ios_Paired)

====
