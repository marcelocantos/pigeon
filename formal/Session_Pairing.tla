---- MODULE Session_Pairing ----
\* Auto-generated from protocol YAML. Do not edit.
\* Phase: Pairing

EXTENDS Integers, Sequences, FiniteSets, TLC

\* States for backend
backend_Idle == "backend_Idle"
backend_GenerateToken == "backend_GenerateToken"
backend_RegisterRelay == "backend_RegisterRelay"
backend_WaitingForClient == "backend_WaitingForClient"
backend_DeriveSecret == "backend_DeriveSecret"
backend_SendAck == "backend_SendAck"
backend_WaitingForCode == "backend_WaitingForCode"
backend_ValidateCode == "backend_ValidateCode"
backend_StorePaired == "backend_StorePaired"
backend_Paired == "backend_Paired"
backend_AuthCheck == "backend_AuthCheck"
backend_SessionActive == "backend_SessionActive"
backend_RelayConnected == "backend_RelayConnected"

\* States for client
client_Idle == "client_Idle"
client_ObtainBackchannelSecret == "client_ObtainBackchannelSecret"
client_ConnectRelay == "client_ConnectRelay"
client_GenKeyPair == "client_GenKeyPair"
client_WaitAck == "client_WaitAck"
client_E2EReady == "client_E2EReady"
client_ShowCode == "client_ShowCode"
client_WaitPairComplete == "client_WaitPairComplete"
client_Paired == "client_Paired"
client_Reconnect == "client_Reconnect"
client_SendAuth == "client_SendAuth"
client_SessionActive == "client_SessionActive"
client_RelayConnected == "client_RelayConnected"

\* States for relay
relay_Idle == "relay_Idle"
relay_BackendRegistered == "relay_BackendRegistered"

\* Message types
MSG_pair_hello == "pair_hello"
MSG_pair_hello_ack == "pair_hello_ack"
MSG_pair_confirm == "pair_confirm"
MSG_pair_complete == "pair_complete"
MSG_auth_request == "auth_request"
MSG_auth_ok == "auth_ok"

\* Event types
EVT_app_launch == "app_launch"
EVT_backchannel_received == "backchannel_received"
EVT_check_code == "check_code"
EVT_cli_code_entered == "cli_code_entered"
EVT_cli_init_pair == "cli_init_pair"
EVT_code_displayed == "code_displayed"
EVT_ecdh_complete == "ecdh_complete"
EVT_finalise == "finalise"
EVT_key_pair_generated == "key_pair_generated"
EVT_recv_auth_ok == "recv_auth_ok"
EVT_recv_auth_request == "recv_auth_request"
EVT_recv_pair_complete == "recv_pair_complete"
EVT_recv_pair_confirm == "recv_pair_confirm"
EVT_recv_pair_hello == "recv_pair_hello"
EVT_recv_pair_hello_ack == "recv_pair_hello_ack"
EVT_relay_connected == "relay_connected"
EVT_relay_registered == "relay_registered"
EVT_secret_parsed == "secret_parsed"
EVT_signal_code_display == "signal_code_display"
EVT_token_created == "token_created"
EVT_verify == "verify"

\* deterministic ordering for ECDH
KeyRank(k) == CASE k = "adv_pub" -> 0 [] k = "client_pub" -> 1 [] k = "backend_pub" -> 2 [] OTHER -> 3
\* symbolic ECDH
DeriveKey(a, b) == IF KeyRank(a) <= KeyRank(b) THEN <<"ecdh", a, b>> ELSE <<"ecdh", b, a>>
\* confirmation code from pubkeys
DeriveCode(a, b) == IF KeyRank(a) <= KeyRank(b) THEN <<"code", a, b>> ELSE <<"code", b, a>>
\* minimum of two values
Min(a, b) == IF a < b THEN a ELSE b



CONSTANTS cli_entered_code, adversary_keys, adv_ecdh_pub, adv_saved_client_pub, adv_saved_server_pub

VARIABLES
    backend_state,
    client_state,
    current_token,
    active_tokens,
    used_tokens,
    backend_ecdh_pub,
    received_client_pub,
    received_backend_pub,
    backend_shared_key,
    client_shared_key,
    backend_code,
    client_code,
    received_code,
    code_attempts,
    device_secret,
    paired_devices,
    received_device_id,
    auth_nonces_used,
    received_auth_nonce,
    secret_published,
    received_pair_hello,
    received_auth_request,
    received_pair_hello_ack,
    received_pair_confirm,
    received_pair_complete,
    received_auth_ok

vars == <<backend_state, client_state, current_token, active_tokens, used_tokens, backend_ecdh_pub, received_client_pub, received_backend_pub, backend_shared_key, client_shared_key, backend_code, client_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, secret_published, received_pair_hello, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

Init ==
    /\ backend_state = backend_Idle
    /\ client_state = client_Idle
    /\ current_token = "none"
    /\ active_tokens = {}
    /\ used_tokens = {}
    /\ backend_ecdh_pub = "none"
    /\ received_client_pub = "none"
    /\ received_backend_pub = "none"
    /\ backend_shared_key = <<"none">>
    /\ client_shared_key = <<"none">>
    /\ backend_code = <<"none">>
    /\ client_code = <<"none">>
    /\ received_code = <<"none">>
    /\ code_attempts = 0
    /\ device_secret = "none"
    /\ paired_devices = {}
    /\ received_device_id = "none"
    /\ auth_nonces_used = {}
    /\ received_auth_nonce = "none"
    /\ secret_published = FALSE
    /\ received_pair_hello = [type |-> "none"]
    /\ received_auth_request = [type |-> "none"]
    /\ received_pair_hello_ack = [type |-> "none"]
    /\ received_pair_confirm = [type |-> "none"]
    /\ received_pair_complete = [type |-> "none"]
    /\ received_auth_ok = [type |-> "none"]

\* backend: Idle -> GenerateToken (cli_init_pair)
backend_Idle_to_GenerateToken_cli_init_pair ==
    /\ backend_state = backend_Idle
    /\ backend_state' = backend_GenerateToken
    /\ current_token' = "tok_1"
    /\ active_tokens' = active_tokens \union {"tok_1"}
    /\ UNCHANGED <<client_state, used_tokens, backend_ecdh_pub, received_client_pub, received_backend_pub, backend_shared_key, client_shared_key, backend_code, client_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, secret_published, received_pair_hello, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* backend: GenerateToken -> RegisterRelay (token_created)
backend_GenerateToken_to_RegisterRelay_token_created ==
    /\ backend_state = backend_GenerateToken
    /\ backend_state' = backend_RegisterRelay
    /\ UNCHANGED <<client_state, current_token, active_tokens, used_tokens, backend_ecdh_pub, received_client_pub, received_backend_pub, backend_shared_key, client_shared_key, backend_code, client_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, secret_published, received_pair_hello, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* backend: RegisterRelay -> WaitingForClient (relay_registered)
backend_RegisterRelay_to_WaitingForClient_relay_registered ==
    /\ backend_state = backend_RegisterRelay
    /\ backend_state' = backend_WaitingForClient
    /\ secret_published' = TRUE
    /\ UNCHANGED <<client_state, current_token, active_tokens, used_tokens, backend_ecdh_pub, received_client_pub, received_backend_pub, backend_shared_key, client_shared_key, backend_code, client_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, received_pair_hello, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* backend: WaitingForClient -> DeriveSecret on recv pair_hello [token_valid]
backend_WaitingForClient_to_DeriveSecret_on_pair_hello_token_valid ==
    /\ backend_state = backend_WaitingForClient
    /\ received_pair_hello.type = MSG_pair_hello
    /\ received_pair_hello.token \in active_tokens
    /\ received_pair_hello' = [type |-> "none"]
    /\ backend_state' = backend_DeriveSecret
    /\ received_client_pub' = received_pair_hello.pubkey
    /\ backend_ecdh_pub' = "backend_pub"
    /\ backend_shared_key' = DeriveKey("backend_pub", received_pair_hello.pubkey)
    /\ backend_code' = DeriveCode("backend_pub", received_pair_hello.pubkey)
    /\ UNCHANGED <<client_state, current_token, active_tokens, used_tokens, received_backend_pub, client_shared_key, client_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, secret_published, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* backend: WaitingForClient -> Idle on recv pair_hello [token_invalid]
backend_WaitingForClient_to_Idle_on_pair_hello_token_invalid ==
    /\ backend_state = backend_WaitingForClient
    /\ received_pair_hello.type = MSG_pair_hello
    /\ received_pair_hello.token \notin active_tokens
    /\ received_pair_hello' = [type |-> "none"]
    /\ backend_state' = backend_Idle
    /\ UNCHANGED <<client_state, current_token, active_tokens, used_tokens, backend_ecdh_pub, received_client_pub, received_backend_pub, backend_shared_key, client_shared_key, backend_code, client_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, secret_published, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* backend: DeriveSecret -> SendAck (ecdh_complete)
backend_DeriveSecret_to_SendAck_ecdh_complete ==
    /\ backend_state = backend_DeriveSecret
    /\ received_pair_hello_ack' = [type |-> MSG_pair_hello_ack, pubkey |-> backend_ecdh_pub]
    /\ backend_state' = backend_SendAck
    /\ UNCHANGED <<client_state, current_token, active_tokens, used_tokens, backend_ecdh_pub, received_client_pub, received_backend_pub, backend_shared_key, client_shared_key, backend_code, client_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, secret_published, received_pair_hello, received_auth_request, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* backend: SendAck -> WaitingForCode (signal_code_display)
backend_SendAck_to_WaitingForCode_signal_code_display ==
    /\ backend_state = backend_SendAck
    /\ received_pair_confirm' = [type |-> MSG_pair_confirm]
    /\ backend_state' = backend_WaitingForCode
    /\ UNCHANGED <<client_state, current_token, active_tokens, used_tokens, backend_ecdh_pub, received_client_pub, received_backend_pub, backend_shared_key, client_shared_key, backend_code, client_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, secret_published, received_pair_hello, received_auth_request, received_pair_hello_ack, received_pair_complete, received_auth_ok>>

\* backend: WaitingForCode -> ValidateCode (cli_code_entered)
backend_WaitingForCode_to_ValidateCode_cli_code_entered ==
    /\ backend_state = backend_WaitingForCode
    /\ backend_state' = backend_ValidateCode
    /\ received_code' = cli_entered_code
    /\ UNCHANGED <<client_state, current_token, active_tokens, used_tokens, backend_ecdh_pub, received_client_pub, received_backend_pub, backend_shared_key, client_shared_key, backend_code, client_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, secret_published, received_pair_hello, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* backend: ValidateCode -> StorePaired (check_code) [code_correct]
backend_ValidateCode_to_StorePaired_check_code_code_correct ==
    /\ backend_state = backend_ValidateCode
    /\ received_code = backend_code
    /\ backend_state' = backend_StorePaired
    /\ UNCHANGED <<client_state, current_token, active_tokens, used_tokens, backend_ecdh_pub, received_client_pub, received_backend_pub, backend_shared_key, client_shared_key, backend_code, client_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, secret_published, received_pair_hello, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* backend: ValidateCode -> Idle (check_code) [code_wrong]
backend_ValidateCode_to_Idle_check_code_code_wrong ==
    /\ backend_state = backend_ValidateCode
    /\ received_code /= backend_code
    /\ backend_state' = backend_Idle
    /\ code_attempts' = code_attempts + 1
    /\ UNCHANGED <<client_state, current_token, active_tokens, used_tokens, backend_ecdh_pub, received_client_pub, received_backend_pub, backend_shared_key, client_shared_key, backend_code, client_code, received_code, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, secret_published, received_pair_hello, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* backend: StorePaired -> Paired (finalise)
backend_StorePaired_to_Paired_finalise ==
    /\ backend_state = backend_StorePaired
    /\ received_pair_complete' = [type |-> MSG_pair_complete, key |-> backend_shared_key, secret |-> "dev_secret_1"]
    /\ backend_state' = backend_Paired
    /\ device_secret' = "dev_secret_1"
    /\ paired_devices' = paired_devices \union {"device_1"}
    /\ active_tokens' = active_tokens \ {current_token}
    /\ used_tokens' = used_tokens \union {current_token}
    /\ UNCHANGED <<client_state, current_token, backend_ecdh_pub, received_client_pub, received_backend_pub, backend_shared_key, client_shared_key, backend_code, client_code, received_code, code_attempts, received_device_id, auth_nonces_used, received_auth_nonce, secret_published, received_pair_hello, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_auth_ok>>

\* backend: Paired -> AuthCheck on recv auth_request
backend_Paired_to_AuthCheck_on_auth_request ==
    /\ backend_state = backend_Paired
    /\ received_auth_request.type = MSG_auth_request
    /\ received_auth_request' = [type |-> "none"]
    /\ backend_state' = backend_AuthCheck
    /\ received_device_id' = received_auth_request.device_id
    /\ received_auth_nonce' = received_auth_request.nonce
    /\ UNCHANGED <<client_state, current_token, active_tokens, used_tokens, backend_ecdh_pub, received_client_pub, received_backend_pub, backend_shared_key, client_shared_key, backend_code, client_code, received_code, code_attempts, device_secret, paired_devices, auth_nonces_used, secret_published, received_pair_hello, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* backend: AuthCheck -> SessionActive (verify) [device_known]
backend_AuthCheck_to_SessionActive_verify_device_known ==
    /\ backend_state = backend_AuthCheck
    /\ received_device_id \in paired_devices
    /\ received_auth_ok' = [type |-> MSG_auth_ok]
    /\ backend_state' = backend_SessionActive
    /\ auth_nonces_used' = auth_nonces_used \union {received_auth_nonce}
    /\ UNCHANGED <<client_state, current_token, active_tokens, used_tokens, backend_ecdh_pub, received_client_pub, received_backend_pub, backend_shared_key, client_shared_key, backend_code, client_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, received_auth_nonce, secret_published, received_pair_hello, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete>>

\* backend: AuthCheck -> Idle (verify) [device_unknown]
backend_AuthCheck_to_Idle_verify_device_unknown ==
    /\ backend_state = backend_AuthCheck
    /\ received_device_id \notin paired_devices
    /\ backend_state' = backend_Idle
    /\ UNCHANGED <<client_state, current_token, active_tokens, used_tokens, backend_ecdh_pub, received_client_pub, received_backend_pub, backend_shared_key, client_shared_key, backend_code, client_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, secret_published, received_pair_hello, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>


\* client: Idle -> ObtainBackchannelSecret (backchannel_received)
client_Idle_to_ObtainBackchannelSecret_backchannel_received ==
    /\ client_state = client_Idle
    /\ client_state' = client_ObtainBackchannelSecret
    /\ UNCHANGED <<backend_state, current_token, active_tokens, used_tokens, backend_ecdh_pub, received_client_pub, received_backend_pub, backend_shared_key, client_shared_key, backend_code, client_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, secret_published, received_pair_hello, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* client: ObtainBackchannelSecret -> ConnectRelay (secret_parsed)
client_ObtainBackchannelSecret_to_ConnectRelay_secret_parsed ==
    /\ client_state = client_ObtainBackchannelSecret
    /\ client_state' = client_ConnectRelay
    /\ UNCHANGED <<backend_state, current_token, active_tokens, used_tokens, backend_ecdh_pub, received_client_pub, received_backend_pub, backend_shared_key, client_shared_key, backend_code, client_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, secret_published, received_pair_hello, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* client: ConnectRelay -> GenKeyPair (relay_connected)
client_ConnectRelay_to_GenKeyPair_relay_connected ==
    /\ client_state = client_ConnectRelay
    /\ client_state' = client_GenKeyPair
    /\ UNCHANGED <<backend_state, current_token, active_tokens, used_tokens, backend_ecdh_pub, received_client_pub, received_backend_pub, backend_shared_key, client_shared_key, backend_code, client_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, secret_published, received_pair_hello, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* client: GenKeyPair -> WaitAck (key_pair_generated)
client_GenKeyPair_to_WaitAck_key_pair_generated ==
    /\ client_state = client_GenKeyPair
    /\ received_pair_hello' = [type |-> MSG_pair_hello, pubkey |-> "client_pub", token |-> current_token]
    /\ client_state' = client_WaitAck
    /\ UNCHANGED <<backend_state, current_token, active_tokens, used_tokens, backend_ecdh_pub, received_client_pub, received_backend_pub, backend_shared_key, client_shared_key, backend_code, client_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, secret_published, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* client: WaitAck -> E2EReady on recv pair_hello_ack
client_WaitAck_to_E2EReady_on_pair_hello_ack ==
    /\ client_state = client_WaitAck
    /\ received_pair_hello_ack.type = MSG_pair_hello_ack
    /\ received_pair_hello_ack' = [type |-> "none"]
    /\ client_state' = client_E2EReady
    /\ received_backend_pub' = received_pair_hello_ack.pubkey
    /\ client_shared_key' = DeriveKey("client_pub", received_pair_hello_ack.pubkey)
    /\ UNCHANGED <<backend_state, current_token, active_tokens, used_tokens, backend_ecdh_pub, received_client_pub, backend_shared_key, backend_code, client_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, secret_published, received_pair_hello, received_auth_request, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* client: E2EReady -> ShowCode on recv pair_confirm
client_E2EReady_to_ShowCode_on_pair_confirm ==
    /\ client_state = client_E2EReady
    /\ received_pair_confirm.type = MSG_pair_confirm
    /\ received_pair_confirm' = [type |-> "none"]
    /\ client_state' = client_ShowCode
    /\ client_code' = DeriveCode(received_backend_pub, "client_pub")
    /\ UNCHANGED <<backend_state, current_token, active_tokens, used_tokens, backend_ecdh_pub, received_client_pub, received_backend_pub, backend_shared_key, client_shared_key, backend_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, secret_published, received_pair_hello, received_auth_request, received_pair_hello_ack, received_pair_complete, received_auth_ok>>

\* client: ShowCode -> WaitPairComplete (code_displayed)
client_ShowCode_to_WaitPairComplete_code_displayed ==
    /\ client_state = client_ShowCode
    /\ client_state' = client_WaitPairComplete
    /\ UNCHANGED <<backend_state, current_token, active_tokens, used_tokens, backend_ecdh_pub, received_client_pub, received_backend_pub, backend_shared_key, client_shared_key, backend_code, client_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, secret_published, received_pair_hello, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* client: WaitPairComplete -> Paired on recv pair_complete
client_WaitPairComplete_to_Paired_on_pair_complete ==
    /\ client_state = client_WaitPairComplete
    /\ received_pair_complete.type = MSG_pair_complete
    /\ received_pair_complete' = [type |-> "none"]
    /\ client_state' = client_Paired
    /\ UNCHANGED <<backend_state, current_token, active_tokens, used_tokens, backend_ecdh_pub, received_client_pub, received_backend_pub, backend_shared_key, client_shared_key, backend_code, client_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, secret_published, received_pair_hello, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_auth_ok>>

\* client: Paired -> Reconnect (app_launch)
client_Paired_to_Reconnect_app_launch ==
    /\ client_state = client_Paired
    /\ client_state' = client_Reconnect
    /\ UNCHANGED <<backend_state, current_token, active_tokens, used_tokens, backend_ecdh_pub, received_client_pub, received_backend_pub, backend_shared_key, client_shared_key, backend_code, client_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, secret_published, received_pair_hello, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* client: Reconnect -> SendAuth (relay_connected)
client_Reconnect_to_SendAuth_relay_connected ==
    /\ client_state = client_Reconnect
    /\ received_auth_request' = [type |-> MSG_auth_request, device_id |-> "device_1", key |-> client_shared_key, nonce |-> "nonce_1", secret |-> device_secret]
    /\ client_state' = client_SendAuth
    /\ UNCHANGED <<backend_state, current_token, active_tokens, used_tokens, backend_ecdh_pub, received_client_pub, received_backend_pub, backend_shared_key, client_shared_key, backend_code, client_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, secret_published, received_pair_hello, received_pair_hello_ack, received_pair_confirm, received_pair_complete, received_auth_ok>>

\* client: SendAuth -> SessionActive on recv auth_ok
client_SendAuth_to_SessionActive_on_auth_ok ==
    /\ client_state = client_SendAuth
    /\ received_auth_ok.type = MSG_auth_ok
    /\ received_auth_ok' = [type |-> "none"]
    /\ client_state' = client_SessionActive
    /\ UNCHANGED <<backend_state, current_token, active_tokens, used_tokens, backend_ecdh_pub, received_client_pub, received_backend_pub, backend_shared_key, client_shared_key, backend_code, client_code, received_code, code_attempts, device_secret, paired_devices, received_device_id, auth_nonces_used, received_auth_nonce, secret_published, received_pair_hello, received_auth_request, received_pair_hello_ack, received_pair_confirm, received_pair_complete>>


Next ==
    \/ backend_Idle_to_GenerateToken_cli_init_pair
    \/ backend_GenerateToken_to_RegisterRelay_token_created
    \/ backend_RegisterRelay_to_WaitingForClient_relay_registered
    \/ backend_WaitingForClient_to_DeriveSecret_on_pair_hello_token_valid
    \/ backend_WaitingForClient_to_Idle_on_pair_hello_token_invalid
    \/ backend_DeriveSecret_to_SendAck_ecdh_complete
    \/ backend_SendAck_to_WaitingForCode_signal_code_display
    \/ backend_WaitingForCode_to_ValidateCode_cli_code_entered
    \/ backend_ValidateCode_to_StorePaired_check_code_code_correct
    \/ backend_ValidateCode_to_Idle_check_code_code_wrong
    \/ backend_StorePaired_to_Paired_finalise
    \/ backend_Paired_to_AuthCheck_on_auth_request
    \/ backend_AuthCheck_to_SessionActive_verify_device_known
    \/ backend_AuthCheck_to_Idle_verify_device_unknown
    \/ client_Idle_to_ObtainBackchannelSecret_backchannel_received
    \/ client_ObtainBackchannelSecret_to_ConnectRelay_secret_parsed
    \/ client_ConnectRelay_to_GenKeyPair_relay_connected
    \/ client_GenKeyPair_to_WaitAck_key_pair_generated
    \/ client_WaitAck_to_E2EReady_on_pair_hello_ack
    \/ client_E2EReady_to_ShowCode_on_pair_confirm
    \/ client_ShowCode_to_WaitPairComplete_code_displayed
    \/ client_WaitPairComplete_to_Paired_on_pair_complete
    \/ client_Paired_to_Reconnect_app_launch
    \/ client_Reconnect_to_SendAuth_relay_connected
    \/ client_SendAuth_to_SessionActive_on_auth_ok

Spec == Init /\ [][Next]_vars /\ WF_vars(Next)

\* ================================================================
\* Invariants and properties
\* ================================================================

\* A revoked pairing token is never accepted again
NoTokenReuse == used_tokens \intersect active_tokens = {}
\* MitM produces mismatched codes
MitMDetectedByCodeMismatch == (backend_shared_key \in adversary_keys /\ backend_code /= <<"none">> /\ client_code /= <<"none">>) => backend_code /= client_code
\* Compromised key prevents pairing completion
MitMPrevented == backend_shared_key \in adversary_keys => backend_state \notin {backend_StorePaired, backend_Paired, backend_AuthCheck, backend_SessionActive}
\* Session requires completed pairing
AuthRequiresCompletedPairing == backend_state = backend_SessionActive => received_device_id \in paired_devices
\* Each auth nonce accepted at most once
NoNonceReuse == backend_state = backend_SessionActive => received_auth_nonce \notin (auth_nonces_used \ {received_auth_nonce})
\* Adversary never learns device secret
DeviceSecretSecrecy == \A m \in adversary_knowledge : "type" \in DOMAIN m => m.type /= "plaintext_secret"
\* After fallback, backend eventually re-advertises LAN
FallbackLeadsToReadvertise == (backend_state = backend_RelayBackoff) ~> (backend_state = backend_LANOffered)
\* Degraded state eventually resolves (recovery or fallback)
DegradedLeadsToResolutionOrFallback == (backend_state = backend_LANDegraded) ~> (backend_state \in {backend_LANActive, backend_RelayBackoff})

====
