---- MODULE PairingCeremony ----
\* Auto-generated from protocol YAML. Do not edit.

EXTENDS Integers, Sequences, FiniteSets, TLC

\* States for server/pairing
server_pairing_Idle == "server_pairing_Idle"
server_pairing_GenerateToken == "server_pairing_GenerateToken"
server_pairing_RegisterRelay == "server_pairing_RegisterRelay"
server_pairing_WaitingForClient == "server_pairing_WaitingForClient"
server_pairing_DeriveSecret == "server_pairing_DeriveSecret"
server_pairing_SendAck == "server_pairing_SendAck"
server_pairing_WaitingForCode == "server_pairing_WaitingForCode"
server_pairing_ValidateCode == "server_pairing_ValidateCode"
server_pairing_StorePaired == "server_pairing_StorePaired"
server_pairing_PairingComplete == "server_pairing_PairingComplete"

\* States for server/auth
server_auth_Idle == "server_auth_Idle"
server_auth_Paired == "server_auth_Paired"
server_auth_AuthCheck == "server_auth_AuthCheck"
server_auth_SessionActive == "server_auth_SessionActive"

\* States for ios/pairing
ios_pairing_Idle == "ios_pairing_Idle"
ios_pairing_ScanQR == "ios_pairing_ScanQR"
ios_pairing_ConnectRelay == "ios_pairing_ConnectRelay"
ios_pairing_GenKeyPair == "ios_pairing_GenKeyPair"
ios_pairing_WaitAck == "ios_pairing_WaitAck"
ios_pairing_E2EReady == "ios_pairing_E2EReady"
ios_pairing_ShowCode == "ios_pairing_ShowCode"
ios_pairing_WaitPairComplete == "ios_pairing_WaitPairComplete"
ios_pairing_PairingComplete == "ios_pairing_PairingComplete"

\* States for ios/auth
ios_auth_Idle == "ios_auth_Idle"
ios_auth_Paired == "ios_auth_Paired"
ios_auth_Reconnect == "ios_auth_Reconnect"
ios_auth_SendAuth == "ios_auth_SendAuth"
ios_auth_SessionActive == "ios_auth_SessionActive"

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

\* Event types
EVT_ECDH_complete == "ECDH complete"
EVT_QR_parsed == "QR parsed"
EVT_app_launch == "app launch"
EVT_check_code == "check code"
EVT_cli___init == "cli --init"
EVT_code_displayed == "code displayed"
EVT_credential_ready == "credential_ready"
EVT_disconnect == "disconnect"
EVT_finalise == "finalise"
EVT_key_pair_generated == "key pair generated"
EVT_key_stored == "key stored"
EVT_recv_auth_ok == "recv_auth_ok"
EVT_recv_auth_request == "recv_auth_request"
EVT_recv_code_submit == "recv_code_submit"
EVT_recv_pair_begin == "recv_pair_begin"
EVT_recv_pair_complete == "recv_pair_complete"
EVT_recv_pair_confirm == "recv_pair_confirm"
EVT_recv_pair_hello == "recv_pair_hello"
EVT_recv_pair_hello_ack == "recv_pair_hello_ack"
EVT_recv_pair_status == "recv_pair_status"
EVT_recv_token_response == "recv_token_response"
EVT_recv_waiting_for_code == "recv_waiting_for_code"
EVT_relay_connected == "relay connected"
EVT_relay_registered == "relay registered"
EVT_signal_code_display == "signal code display"
EVT_token_created == "token created"
EVT_user_enters_code == "user enters code"
EVT_user_scans_QR == "user scans QR"
EVT_verify == "verify"

\* Assign numeric rank to pubkey names for deterministic ordering
KeyRank(k) == CASE k = "adv_pub" -> 0 [] k = "client_pub" -> 1 [] k = "server_pub" -> 2 [] OTHER -> 3
\* Symbolic ECDH: deterministic key from two public keys (order-independent)
DeriveKey(a, b) == IF KeyRank(a) <= KeyRank(b) THEN <<"ecdh", a, b>> ELSE <<"ecdh", b, a>>
\* Key-bound confirmation code: deterministic from both pubkeys (order-independent)
DeriveCode(a, b) == IF KeyRank(a) <= KeyRank(b) THEN <<"code", a, b>> ELSE <<"code", b, a>>



CONSTANTS adversary_keys, adv_ecdh_pub, adv_saved_client_pub, adv_saved_server_pub

VARIABLES
    cli_state,
    server_pairing_state,
    server_auth_state,
    ios_pairing_state,
    ios_auth_state,
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
    received_token_response,
    received_waiting_for_code,
    received_pair_status,
    received_pair_begin,
    received_pair_hello,
    received_code_submit,
    received_auth_request,
    received_pair_hello_ack,
    received_pair_confirm,
    received_pair_complete,
    received_auth_ok

vars == <<cli_state, server_pairing_state, server_auth_state, ios_pairing_state, ios_auth_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_status, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

Init ==
    /\ cli_state = cli_Idle
    /\ server_pairing_state = server_pairing_Idle
    /\ server_auth_state = server_auth_Idle
    /\ ios_pairing_state = ios_pairing_Idle
    /\ ios_auth_state = ios_auth_Idle
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
    /\ received_token_response = [type |-> "none"]
    /\ received_waiting_for_code = [type |-> "none"]
    /\ received_pair_status = [type |-> "none"]
    /\ received_pair_begin = [type |-> "none"]
    /\ received_pair_hello = [type |-> "none"]
    /\ received_code_submit = [type |-> "none"]
    /\ received_auth_request = [type |-> "none"]
    /\ received_pair_hello_ack = [type |-> "none"]
    /\ received_pair_confirm = [type |-> "none"]
    /\ received_pair_complete = [type |-> "none"]
    /\ received_auth_ok = [type |-> "none"]

\* cli: Idle -> GetKey (cli --init)
cli_Idle_to_GetKey_cli___init ==
    /\ cli_state = cli_Idle
    /\ cli_state' = cli_GetKey
    /\ UNCHANGED <<server_pairing_state, server_auth_state, ios_pairing_state, ios_auth_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_status, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* cli: GetKey -> BeginPair (key stored)
cli_GetKey_to_BeginPair_key_stored ==
    /\ cli_state = cli_GetKey
    /\ received_pair_begin' = [type |-> MSG_pair_begin]
    /\ cli_state' = cli_BeginPair
    /\ UNCHANGED <<server_pairing_state, server_auth_state, ios_pairing_state, ios_auth_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_status, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* cli: BeginPair -> ShowQR on recv token_response
cli_BeginPair_to_ShowQR_on_token_response ==
    /\ cli_state = cli_BeginPair
    /\ received_token_response.type = MSG_token_response
    /\ received_token_response' = [type |-> "none"]
    /\ cli_state' = cli_ShowQR
    /\ UNCHANGED <<server_pairing_state, server_auth_state, ios_pairing_state, ios_auth_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_waiting_for_code, received_pair_status, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* cli: ShowQR -> PromptCode on recv waiting_for_code
cli_ShowQR_to_PromptCode_on_waiting_for_code ==
    /\ cli_state = cli_ShowQR
    /\ received_waiting_for_code.type = MSG_waiting_for_code
    /\ received_waiting_for_code' = [type |-> "none"]
    /\ cli_state' = cli_PromptCode
    /\ UNCHANGED <<server_pairing_state, server_auth_state, ios_pairing_state, ios_auth_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_pair_status, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* cli: PromptCode -> SubmitCode (user enters code)
cli_PromptCode_to_SubmitCode_user_enters_code ==
    /\ cli_state = cli_PromptCode
    /\ received_code_submit' = [type |-> MSG_code_submit, code |-> ios_code]
    /\ cli_state' = cli_SubmitCode
    /\ UNCHANGED <<server_pairing_state, server_auth_state, ios_pairing_state, ios_auth_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_status, received_pair_begin, received_pair_hello, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* cli: SubmitCode -> Done on recv pair_status
cli_SubmitCode_to_Done_on_pair_status ==
    /\ cli_state = cli_SubmitCode
    /\ received_pair_status.type = MSG_pair_status
    /\ received_pair_status' = [type |-> "none"]
    /\ cli_state' = cli_Done
    /\ UNCHANGED <<server_pairing_state, server_auth_state, ios_pairing_state, ios_auth_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>


\* server/pairing: Idle -> GenerateToken on recv pair_begin
server_pairing_Idle_to_GenerateToken_on_pair_begin ==
    /\ server_pairing_state = server_pairing_Idle
    /\ received_pair_begin.type = MSG_pair_begin
    /\ received_pair_begin' = [type |-> "none"]
    /\ server_pairing_state' = server_pairing_GenerateToken
    /\ current_token' = "tok_1"
    /\ active_tokens' = active_tokens \union {"tok_1"}
    /\ UNCHANGED <<cli_state, server_auth_state, ios_pairing_state, ios_auth_state, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_status, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* server/pairing: GenerateToken -> RegisterRelay (token created)
server_pairing_GenerateToken_to_RegisterRelay_token_created ==
    /\ server_pairing_state = server_pairing_GenerateToken
    /\ server_pairing_state' = server_pairing_RegisterRelay
    /\ UNCHANGED <<cli_state, server_auth_state, ios_pairing_state, ios_auth_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_status, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* server/pairing: RegisterRelay -> WaitingForClient (relay registered)
server_pairing_RegisterRelay_to_WaitingForClient_relay_registered ==
    /\ server_pairing_state = server_pairing_RegisterRelay
    /\ received_token_response' = [type |-> MSG_token_response, instance_id |-> "inst_1", token |-> current_token]
    /\ server_pairing_state' = server_pairing_WaitingForClient
    /\ UNCHANGED <<cli_state, server_auth_state, ios_pairing_state, ios_auth_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_waiting_for_code, received_pair_status, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* server/pairing: WaitingForClient -> DeriveSecret on recv pair_hello [token_valid]
server_pairing_WaitingForClient_to_DeriveSecret_on_pair_hello_token_valid ==
    /\ server_pairing_state = server_pairing_WaitingForClient
    /\ received_pair_hello.type = MSG_pair_hello
    /\ received_pair_hello.token \in active_tokens
    /\ received_pair_hello' = [type |-> "none"]
    /\ server_pairing_state' = server_pairing_DeriveSecret
    /\ received_client_pub' = received_pair_hello.pubkey
    /\ server_ecdh_pub' = "server_pub"
    /\ server_shared_key' = DeriveKey("server_pub", received_pair_hello.pubkey)
    /\ server_code' = DeriveCode("server_pub", received_pair_hello.pubkey)
    /\ active_tokens' = active_tokens \ {current_token}
    /\ used_tokens' = used_tokens \union {current_token}
    /\ UNCHANGED <<cli_state, server_auth_state, ios_pairing_state, ios_auth_state, current_token, received_server_pub, client_shared_key, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_status, received_pair_begin, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* server/pairing: WaitingForClient -> Idle on recv pair_hello [token_invalid]
server_pairing_WaitingForClient_to_Idle_on_pair_hello_token_invalid ==
    /\ server_pairing_state = server_pairing_WaitingForClient
    /\ received_pair_hello.type = MSG_pair_hello
    /\ received_pair_hello.token \notin active_tokens
    /\ received_pair_hello' = [type |-> "none"]
    /\ server_pairing_state' = server_pairing_Idle
    /\ UNCHANGED <<cli_state, server_auth_state, ios_pairing_state, ios_auth_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_status, received_pair_begin, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* server/pairing: DeriveSecret -> SendAck (ECDH complete)
server_pairing_DeriveSecret_to_SendAck_ECDH_complete ==
    /\ server_pairing_state = server_pairing_DeriveSecret
    /\ received_pair_hello_ack' = [type |-> MSG_pair_hello_ack, pubkey |-> server_ecdh_pub]
    /\ server_pairing_state' = server_pairing_SendAck
    /\ UNCHANGED <<cli_state, server_auth_state, ios_pairing_state, ios_auth_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_status, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* server/pairing: SendAck -> WaitingForCode (signal code display)
server_pairing_SendAck_to_WaitingForCode_signal_code_display ==
    /\ server_pairing_state = server_pairing_SendAck
    /\ received_pair_confirm' = [type |-> MSG_pair_confirm]
    /\ received_waiting_for_code' = [type |-> MSG_waiting_for_code]
    /\ server_pairing_state' = server_pairing_WaitingForCode
    /\ UNCHANGED <<cli_state, server_auth_state, ios_pairing_state, ios_auth_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_pair_status, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_complete, received_auth_ok>>

\* server/pairing: WaitingForCode -> ValidateCode on recv code_submit
server_pairing_WaitingForCode_to_ValidateCode_on_code_submit ==
    /\ server_pairing_state = server_pairing_WaitingForCode
    /\ received_code_submit.type = MSG_code_submit
    /\ received_code_submit' = [type |-> "none"]
    /\ server_pairing_state' = server_pairing_ValidateCode
    /\ received_code' = received_code_submit.code
    /\ UNCHANGED <<cli_state, server_auth_state, ios_pairing_state, ios_auth_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_status, received_pair_begin, received_pair_hello, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* server/pairing: ValidateCode -> StorePaired (check code) [code_correct]
server_pairing_ValidateCode_to_StorePaired_check_code_code_correct ==
    /\ server_pairing_state = server_pairing_ValidateCode
    /\ received_code = server_code
    /\ server_pairing_state' = server_pairing_StorePaired
    /\ UNCHANGED <<cli_state, server_auth_state, ios_pairing_state, ios_auth_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_status, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* server/pairing: ValidateCode -> Idle (check code) [code_wrong]
server_pairing_ValidateCode_to_Idle_check_code_code_wrong ==
    /\ server_pairing_state = server_pairing_ValidateCode
    /\ received_code /= server_code
    /\ server_pairing_state' = server_pairing_Idle
    /\ code_attempts' = code_attempts + 1
    /\ UNCHANGED <<cli_state, server_auth_state, ios_pairing_state, ios_auth_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_status, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* server/pairing: StorePaired -> PairingComplete (finalise)
server_pairing_StorePaired_to_PairingComplete_finalise ==
    /\ server_pairing_state = server_pairing_StorePaired
    /\ received_pair_complete' = [type |-> MSG_pair_complete, key |-> server_shared_key, secret |-> "dev_secret_1"]
    /\ received_pair_status' = [type |-> MSG_pair_status, status |-> "paired"]
    /\ server_pairing_state' = server_pairing_PairingComplete
    /\ device_secret' = "dev_secret_1"
    /\ paired_devices' = paired_devices \union {"device_1"}
    /\ UNCHANGED <<cli_state, server_auth_state, ios_pairing_state, ios_auth_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_auth_ok>>


\* server/auth: Idle -> Paired (credential_ready)
server_auth_Idle_to_Paired_credential_ready ==
    /\ server_auth_state = server_auth_Idle
    /\ server_auth_state' = server_auth_Paired
    /\ UNCHANGED <<cli_state, server_pairing_state, ios_pairing_state, ios_auth_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_status, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* server/auth: Paired -> AuthCheck on recv auth_request
server_auth_Paired_to_AuthCheck_on_auth_request ==
    /\ server_auth_state = server_auth_Paired
    /\ received_auth_request.type = MSG_auth_request
    /\ received_auth_request' = [type |-> "none"]
    /\ server_auth_state' = server_auth_AuthCheck
    /\ received_device_id' = received_auth_request.device_id
    /\ received_auth_nonce' = received_auth_request.nonce
    /\ UNCHANGED <<cli_state, server_pairing_state, ios_pairing_state, ios_auth_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, auth_nonces_used, received_token_response, received_waiting_for_code, received_pair_status, received_pair_begin, received_pair_hello, received_code_submit, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* server/auth: AuthCheck -> SessionActive (verify) [device_known]
server_auth_AuthCheck_to_SessionActive_verify_device_known ==
    /\ server_auth_state = server_auth_AuthCheck
    /\ received_device_id \in paired_devices
    /\ received_auth_ok' = [type |-> MSG_auth_ok]
    /\ server_auth_state' = server_auth_SessionActive
    /\ auth_nonces_used' = auth_nonces_used \union {received_auth_nonce}
    /\ UNCHANGED <<cli_state, server_pairing_state, ios_pairing_state, ios_auth_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_status, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete>>

\* server/auth: AuthCheck -> Idle (verify) [device_unknown]
server_auth_AuthCheck_to_Idle_verify_device_unknown ==
    /\ server_auth_state = server_auth_AuthCheck
    /\ received_device_id \notin paired_devices
    /\ server_auth_state' = server_auth_Idle
    /\ UNCHANGED <<cli_state, server_pairing_state, ios_pairing_state, ios_auth_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_status, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* server/auth: SessionActive -> Paired (disconnect)
server_auth_SessionActive_to_Paired_disconnect ==
    /\ server_auth_state = server_auth_SessionActive
    /\ server_auth_state' = server_auth_Paired
    /\ UNCHANGED <<cli_state, server_pairing_state, ios_pairing_state, ios_auth_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_status, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>


\* Route: pairing reports paired -> delivers to targets
server_route_pairing_paired ==
    /\ TRUE  \* route guard: no matching reporting state found
    /\ server_auth_state' = server_auth_Paired
    /\ UNCHANGED <<cli_state, server_pairing_state, ios_pairing_state, ios_auth_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_status, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* ios/pairing: Idle -> ScanQR (user scans QR)
ios_pairing_Idle_to_ScanQR_user_scans_QR ==
    /\ ios_pairing_state = ios_pairing_Idle
    /\ ios_pairing_state' = ios_pairing_ScanQR
    /\ UNCHANGED <<cli_state, server_pairing_state, server_auth_state, ios_auth_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_status, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* ios/pairing: ScanQR -> ConnectRelay (QR parsed)
ios_pairing_ScanQR_to_ConnectRelay_QR_parsed ==
    /\ ios_pairing_state = ios_pairing_ScanQR
    /\ ios_pairing_state' = ios_pairing_ConnectRelay
    /\ UNCHANGED <<cli_state, server_pairing_state, server_auth_state, ios_auth_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_status, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* ios/pairing: ConnectRelay -> GenKeyPair (relay connected)
ios_pairing_ConnectRelay_to_GenKeyPair_relay_connected ==
    /\ ios_pairing_state = ios_pairing_ConnectRelay
    /\ ios_pairing_state' = ios_pairing_GenKeyPair
    /\ UNCHANGED <<cli_state, server_pairing_state, server_auth_state, ios_auth_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_status, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* ios/pairing: GenKeyPair -> WaitAck (key pair generated)
ios_pairing_GenKeyPair_to_WaitAck_key_pair_generated ==
    /\ ios_pairing_state = ios_pairing_GenKeyPair
    /\ received_pair_hello' = [type |-> MSG_pair_hello, pubkey |-> "client_pub", token |-> current_token]
    /\ ios_pairing_state' = ios_pairing_WaitAck
    /\ UNCHANGED <<cli_state, server_pairing_state, server_auth_state, ios_auth_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_status, received_pair_begin, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* ios/pairing: WaitAck -> E2EReady on recv pair_hello_ack
ios_pairing_WaitAck_to_E2EReady_on_pair_hello_ack ==
    /\ ios_pairing_state = ios_pairing_WaitAck
    /\ received_pair_hello_ack.type = MSG_pair_hello_ack
    /\ received_pair_hello_ack' = [type |-> "none"]
    /\ ios_pairing_state' = ios_pairing_E2EReady
    /\ received_server_pub' = received_pair_hello_ack.pubkey
    /\ client_shared_key' = DeriveKey("client_pub", received_pair_hello_ack.pubkey)
    /\ UNCHANGED <<cli_state, server_pairing_state, server_auth_state, ios_auth_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, server_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_status, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* ios/pairing: E2EReady -> ShowCode on recv pair_confirm
ios_pairing_E2EReady_to_ShowCode_on_pair_confirm ==
    /\ ios_pairing_state = ios_pairing_E2EReady
    /\ received_pair_confirm.type = MSG_pair_confirm
    /\ received_pair_confirm' = [type |-> "none"]
    /\ ios_pairing_state' = ios_pairing_ShowCode
    /\ ios_code' = DeriveCode(received_server_pub, "client_pub")
    /\ UNCHANGED <<cli_state, server_pairing_state, server_auth_state, ios_auth_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_status, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_complete, received_auth_ok>>

\* ios/pairing: ShowCode -> WaitPairComplete (code displayed)
ios_pairing_ShowCode_to_WaitPairComplete_code_displayed ==
    /\ ios_pairing_state = ios_pairing_ShowCode
    /\ ios_pairing_state' = ios_pairing_WaitPairComplete
    /\ UNCHANGED <<cli_state, server_pairing_state, server_auth_state, ios_auth_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_status, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* ios/pairing: WaitPairComplete -> PairingComplete on recv pair_complete
ios_pairing_WaitPairComplete_to_PairingComplete_on_pair_complete ==
    /\ ios_pairing_state = ios_pairing_WaitPairComplete
    /\ received_pair_complete.type = MSG_pair_complete
    /\ received_pair_complete' = [type |-> "none"]
    /\ ios_pairing_state' = ios_pairing_PairingComplete
    /\ UNCHANGED <<cli_state, server_pairing_state, server_auth_state, ios_auth_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_status, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_auth_ok>>


\* ios/auth: Idle -> Paired (credential_ready)
ios_auth_Idle_to_Paired_credential_ready ==
    /\ ios_auth_state = ios_auth_Idle
    /\ ios_auth_state' = ios_auth_Paired
    /\ UNCHANGED <<cli_state, server_pairing_state, server_auth_state, ios_pairing_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_status, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* ios/auth: Paired -> Reconnect (app launch)
ios_auth_Paired_to_Reconnect_app_launch ==
    /\ ios_auth_state = ios_auth_Paired
    /\ ios_auth_state' = ios_auth_Reconnect
    /\ UNCHANGED <<cli_state, server_pairing_state, server_auth_state, ios_pairing_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_status, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* ios/auth: Reconnect -> SendAuth (relay connected)
ios_auth_Reconnect_to_SendAuth_relay_connected ==
    /\ ios_auth_state = ios_auth_Reconnect
    /\ received_auth_request' = [type |-> MSG_auth_request, device_id |-> "device_1", key |-> client_shared_key, nonce |-> "nonce_1", secret |-> device_secret]
    /\ ios_auth_state' = ios_auth_SendAuth
    /\ UNCHANGED <<cli_state, server_pairing_state, server_auth_state, ios_pairing_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_status, received_pair_begin, received_pair_hello, received_code_submit, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* ios/auth: SendAuth -> SessionActive on recv auth_ok
ios_auth_SendAuth_to_SessionActive_on_auth_ok ==
    /\ ios_auth_state = ios_auth_SendAuth
    /\ received_auth_ok.type = MSG_auth_ok
    /\ received_auth_ok' = [type |-> "none"]
    /\ ios_auth_state' = ios_auth_SessionActive
    /\ UNCHANGED <<cli_state, server_pairing_state, server_auth_state, ios_pairing_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_status, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete>>

\* ios/auth: SessionActive -> Paired (disconnect)
ios_auth_SessionActive_to_Paired_disconnect ==
    /\ ios_auth_state = ios_auth_SessionActive
    /\ ios_auth_state' = ios_auth_Paired
    /\ UNCHANGED <<cli_state, server_pairing_state, server_auth_state, ios_pairing_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_status, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>


\* Route: pairing reports paired -> delivers to targets
ios_route_pairing_paired ==
    /\ TRUE  \* route guard: no matching reporting state found
    /\ ios_auth_state' = ios_auth_Paired
    /\ UNCHANGED <<cli_state, server_pairing_state, server_auth_state, ios_pairing_state, current_token, active_tokens, used_tokens, server_ecdh_pub, received_client_pub, received_server_pub, server_shared_key, client_shared_key, server_code, ios_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_token_response, received_waiting_for_code, received_pair_status, received_pair_begin, received_pair_hello, received_code_submit, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

Next ==
    \/ cli_Idle_to_GetKey_cli___init
    \/ cli_GetKey_to_BeginPair_key_stored
    \/ cli_BeginPair_to_ShowQR_on_token_response
    \/ cli_ShowQR_to_PromptCode_on_waiting_for_code
    \/ cli_PromptCode_to_SubmitCode_user_enters_code
    \/ cli_SubmitCode_to_Done_on_pair_status
    \/ server_pairing_Idle_to_GenerateToken_on_pair_begin
    \/ server_pairing_GenerateToken_to_RegisterRelay_token_created
    \/ server_pairing_RegisterRelay_to_WaitingForClient_relay_registered
    \/ server_pairing_WaitingForClient_to_DeriveSecret_on_pair_hello_token_valid
    \/ server_pairing_WaitingForClient_to_Idle_on_pair_hello_token_invalid
    \/ server_pairing_DeriveSecret_to_SendAck_ECDH_complete
    \/ server_pairing_SendAck_to_WaitingForCode_signal_code_display
    \/ server_pairing_WaitingForCode_to_ValidateCode_on_code_submit
    \/ server_pairing_ValidateCode_to_StorePaired_check_code_code_correct
    \/ server_pairing_ValidateCode_to_Idle_check_code_code_wrong
    \/ server_pairing_StorePaired_to_PairingComplete_finalise
    \/ server_auth_Idle_to_Paired_credential_ready
    \/ server_auth_Paired_to_AuthCheck_on_auth_request
    \/ server_auth_AuthCheck_to_SessionActive_verify_device_known
    \/ server_auth_AuthCheck_to_Idle_verify_device_unknown
    \/ server_auth_SessionActive_to_Paired_disconnect
    \/ server_route_pairing_paired
    \/ ios_pairing_Idle_to_ScanQR_user_scans_QR
    \/ ios_pairing_ScanQR_to_ConnectRelay_QR_parsed
    \/ ios_pairing_ConnectRelay_to_GenKeyPair_relay_connected
    \/ ios_pairing_GenKeyPair_to_WaitAck_key_pair_generated
    \/ ios_pairing_WaitAck_to_E2EReady_on_pair_hello_ack
    \/ ios_pairing_E2EReady_to_ShowCode_on_pair_confirm
    \/ ios_pairing_ShowCode_to_WaitPairComplete_code_displayed
    \/ ios_pairing_WaitPairComplete_to_PairingComplete_on_pair_complete
    \/ ios_auth_Idle_to_Paired_credential_ready
    \/ ios_auth_Paired_to_Reconnect_app_launch
    \/ ios_auth_Reconnect_to_SendAuth_relay_connected
    \/ ios_auth_SendAuth_to_SessionActive_on_auth_ok
    \/ ios_auth_SessionActive_to_Paired_disconnect
    \/ ios_route_pairing_paired

Spec == Init /\ [][Next]_vars /\ WF_vars(Next)

\* ================================================================
\* Invariants and properties
\* ================================================================

\* A revoked pairing token is never accepted again
NoTokenReuse == used_tokens \intersect active_tokens = {}
\* If the current session's shared key is compromised and both sides computed codes, the codes differ
MitMDetectedByCodeMismatch == (server_shared_key \in adversary_keys /\ server_code /= <<"none">> /\ ios_code /= <<"none">>) => server_code /= ios_code
\* If the current session's key is compromised, pairing never completes
MitMPrevented == server_shared_key \in adversary_keys => server_pairing_state \notin {"StorePaired", "PairingComplete"}
\* A session is only active for a device that completed pairing
AuthRequiresCompletedPairing == server_auth_state = "SessionActive" => received_device_id \in paired_devices
\* Each auth nonce is accepted at most once
NoNonceReuse == server_auth_state = "SessionActive" => received_auth_nonce \notin (auth_nonces_used \ {received_auth_nonce})
\* Pairing only completes with the correct confirmation code
WrongCodeDoesNotPair == (server_pairing_state = "StorePaired" \/ server_pairing_state = "PairingComplete") => received_code = server_code \/ received_code = <<"none">>
\* Adversary never learns the device secret in plaintext
DeviceSecretSecrecy == \A m \in adversary_knowledge : "type" \in DOMAIN m => m.type /= "plaintext_secret"
\* If all actors cooperate honestly (no MitM), pairing eventually completes
HonestPairingCompletes == <>(cli_state = "Done" /\ ios_pairing_state = "PairingComplete")

====
