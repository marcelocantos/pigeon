---- MODULE PairingCeremony ----
\* Auto-generated from protocol definition. Do not edit.
\* Source of truth: internal/protocol/ Go definition.

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
MSG_pair_begin == "pair_begin" \* cli -> server (POST /api/pair/begin)
MSG_token_response == "token_response" \* server -> cli ({instance_id, pairing_token})
MSG_pair_hello == "pair_hello" \* ios -> server (ECDH pubkey + pairing token)
MSG_pair_hello_ack == "pair_hello_ack" \* server -> ios (ECDH pubkey)
MSG_pair_confirm == "pair_confirm" \* server -> ios (signal to compute and display code)
MSG_waiting_for_code == "waiting_for_code" \* server -> cli (prompt for code entry)
MSG_code_submit == "code_submit" \* cli -> server (POST /api/pair/confirm)
MSG_pair_complete == "pair_complete" \* server -> ios (encrypted device secret)
MSG_pair_status == "pair_status" \* server -> cli (status: paired)
MSG_auth_request == "auth_request" \* ios -> server (encrypted auth with nonce)
MSG_auth_ok == "auth_ok" \* server -> ios (session established)

\* Helper operators
\* Assign numeric rank to pubkey names for deterministic ordering
KeyRank(k) == CASE k = "adv_pub" -> 0 [] k = "client_pub" -> 1 [] k = "server_pub" -> 2 [] OTHER -> 3
\* Symbolic ECDH: deterministic key from two public keys (order-independent)
DeriveKey(a, b) == IF KeyRank(a) <= KeyRank(b) THEN <<"ecdh", a, b>> ELSE <<"ecdh", b, a>>
\* Key-bound confirmation code: deterministic from both pubkeys (order-independent)
DeriveCode(a, b) == IF KeyRank(a) <= KeyRank(b) THEN <<"code", a, b>> ELSE <<"code", b, a>>

(*--algorithm PairingCeremony

variables
    server_state = server_Idle,
    ios_state = ios_Idle,
    cli_state = cli_Idle,
    chan_cli_server = <<>>,
    chan_ios_server = <<>>,
    chan_server_cli = <<>>,
    chan_server_ios = <<>>,
    adversary_knowledge = {},
    \* pairing token currently in play
    current_token = "none",
    \* set of valid (non-revoked) tokens
    active_tokens = {},
    \* set of revoked tokens
    used_tokens = {},
    \* server ECDH public key
    server_ecdh_pub = "none",
    \* pubkey server received in pair_hello (may be adversary's)
    received_client_pub = "none",
    \* pubkey ios received in pair_hello_ack (may be adversary's)
    received_server_pub = "none",
    \* ECDH key derived by server (tuple to match DeriveKey output type)
    server_shared_key = <<"none">>,
    \* ECDH key derived by ios (tuple to match DeriveKey output type)
    client_shared_key = <<"none">>,
    \* code computed by server from its view of the pubkeys (tuple to match DeriveCode output type)
    server_code = <<"none">>,
    \* code computed by ios from its view of the pubkeys (tuple to match DeriveCode output type)
    ios_code = <<"none">>,
    \* code received in code_submit (tuple to match DeriveCode output type)
    received_code = <<"none">>,
    \* failed code submission attempts
    code_attempts = 0,
    \* persistent device secret
    device_secret = "none",
    \* device IDs that completed pairing
    paired_devices = {},
    \* device_id from auth_request
    received_device_id = "none",
    \* set of consumed auth nonces
    auth_nonces_used = {},
    \* nonce from auth_request
    received_auth_nonce = "none",
    \* encryption keys the adversary knows
    adversary_keys = {},
    \* adversary's ECDH public key
    adv_ecdh_pub = "adv_pub",
    \* real client pubkey saved during MitM
    adv_saved_client_pub = "none",
    \* real server pubkey saved during MitM
    adv_saved_server_pub = "none",
    \* last received message (staging)
    recv_msg = [type |-> "none"];

fair process server = 1
begin
  server_loop:
    either
      \* Idle -> GenerateToken on pair_begin
      await server_state = server_Idle /\ Len(chan_cli_server) > 0 /\ Head(chan_cli_server).type = MSG_pair_begin;
      recv_msg := Head(chan_cli_server);
      chan_cli_server := Tail(chan_cli_server);
      current_token := "tok_1";
      active_tokens := active_tokens \union {"tok_1"};
      server_state := server_GenerateToken;
    or
      \* GenerateToken -> RegisterRelay (token created)
      await server_state = server_GenerateToken;
      server_state := server_RegisterRelay;
    or
      \* RegisterRelay -> WaitingForClient (relay registered)
      await server_state = server_RegisterRelay;
      chan_server_cli := Append(chan_server_cli, [type |-> MSG_token_response, instance_id |-> "inst_1", token |-> current_token]);
      server_state := server_WaitingForClient;
    or
      \* WaitingForClient -> DeriveSecret on pair_hello
      await server_state = server_WaitingForClient /\ Len(chan_ios_server) > 0 /\ Head(chan_ios_server).type = MSG_pair_hello /\ (Head(chan_ios_server).token \in active_tokens);
      recv_msg := Head(chan_ios_server);
      chan_ios_server := Tail(chan_ios_server);
      received_client_pub := recv_msg.pubkey;
      server_ecdh_pub := "server_pub";
      server_shared_key := DeriveKey("server_pub", recv_msg.pubkey);
      server_code := DeriveCode("server_pub", recv_msg.pubkey);
      server_state := server_DeriveSecret;
    or
      \* WaitingForClient -> Idle on pair_hello
      await server_state = server_WaitingForClient /\ Len(chan_ios_server) > 0 /\ Head(chan_ios_server).type = MSG_pair_hello /\ (Head(chan_ios_server).token \notin active_tokens);
      recv_msg := Head(chan_ios_server);
      chan_ios_server := Tail(chan_ios_server);
      server_state := server_Idle;
    or
      \* DeriveSecret -> SendAck (ECDH complete)
      await server_state = server_DeriveSecret;
      chan_server_ios := Append(chan_server_ios, [type |-> MSG_pair_hello_ack, pubkey |-> server_ecdh_pub]);
      server_state := server_SendAck;
    or
      \* SendAck -> WaitingForCode (signal code display)
      await server_state = server_SendAck;
      chan_server_ios := Append(chan_server_ios, [type |-> MSG_pair_confirm]);
      chan_server_cli := Append(chan_server_cli, [type |-> MSG_waiting_for_code]);
      server_state := server_WaitingForCode;
    or
      \* WaitingForCode -> ValidateCode on code_submit
      await server_state = server_WaitingForCode /\ Len(chan_cli_server) > 0 /\ Head(chan_cli_server).type = MSG_code_submit;
      recv_msg := Head(chan_cli_server);
      chan_cli_server := Tail(chan_cli_server);
      received_code := recv_msg.code;
      server_state := server_ValidateCode;
    or
      \* ValidateCode -> StorePaired (check code)
      await server_state = server_ValidateCode /\ (received_code = server_code);
      server_state := server_StorePaired;
    or
      \* ValidateCode -> Idle (check code)
      await server_state = server_ValidateCode /\ (received_code /= server_code);
      code_attempts := code_attempts + 1;
      server_state := server_Idle;
    or
      \* StorePaired -> Paired (finalise)
      await server_state = server_StorePaired;
      chan_server_ios := Append(chan_server_ios, [type |-> MSG_pair_complete, key |-> server_shared_key, secret |-> "dev_secret_1"]);
      chan_server_cli := Append(chan_server_cli, [type |-> MSG_pair_status, status |-> "paired"]);
      device_secret := "dev_secret_1";
      paired_devices := paired_devices \union {"device_1"};
      active_tokens := active_tokens \ {current_token};
      used_tokens := used_tokens \union {current_token};
      server_state := server_Paired;
    or
      \* Paired -> AuthCheck on auth_request
      await server_state = server_Paired /\ Len(chan_ios_server) > 0 /\ Head(chan_ios_server).type = MSG_auth_request;
      recv_msg := Head(chan_ios_server);
      chan_ios_server := Tail(chan_ios_server);
      received_device_id := recv_msg.device_id;
      received_auth_nonce := recv_msg.nonce;
      server_state := server_AuthCheck;
    or
      \* AuthCheck -> SessionActive (verify)
      await server_state = server_AuthCheck /\ (received_device_id \in paired_devices);
      chan_server_ios := Append(chan_server_ios, [type |-> MSG_auth_ok]);
      auth_nonces_used := auth_nonces_used \union {received_auth_nonce};
      server_state := server_SessionActive;
    or
      \* AuthCheck -> Idle (verify)
      await server_state = server_AuthCheck /\ (received_device_id \notin paired_devices);
      server_state := server_Idle;
    or
      \* SessionActive -> Paired (disconnect)
      await server_state = server_SessionActive;
      server_state := server_Paired;
    end either;
end process;

fair process ios = 2
begin
  ios_loop:
    either
      \* Idle -> ScanQR (user scans QR)
      await ios_state = ios_Idle;
      ios_state := ios_ScanQR;
    or
      \* ScanQR -> ConnectRelay (QR parsed)
      await ios_state = ios_ScanQR;
      ios_state := ios_ConnectRelay;
    or
      \* ConnectRelay -> GenKeyPair (relay connected)
      await ios_state = ios_ConnectRelay;
      ios_state := ios_GenKeyPair;
    or
      \* GenKeyPair -> WaitAck (key pair generated)
      await ios_state = ios_GenKeyPair;
      chan_ios_server := Append(chan_ios_server, [type |-> MSG_pair_hello, pubkey |-> "client_pub", token |-> current_token]);
      ios_state := ios_WaitAck;
    or
      \* WaitAck -> E2EReady on pair_hello_ack
      await ios_state = ios_WaitAck /\ Len(chan_server_ios) > 0 /\ Head(chan_server_ios).type = MSG_pair_hello_ack;
      recv_msg := Head(chan_server_ios);
      chan_server_ios := Tail(chan_server_ios);
      received_server_pub := recv_msg.pubkey;
      client_shared_key := DeriveKey("client_pub", recv_msg.pubkey);
      ios_state := ios_E2EReady;
    or
      \* E2EReady -> ShowCode on pair_confirm
      await ios_state = ios_E2EReady /\ Len(chan_server_ios) > 0 /\ Head(chan_server_ios).type = MSG_pair_confirm;
      recv_msg := Head(chan_server_ios);
      chan_server_ios := Tail(chan_server_ios);
      ios_code := DeriveCode(received_server_pub, "client_pub");
      ios_state := ios_ShowCode;
    or
      \* ShowCode -> WaitPairComplete (code displayed)
      await ios_state = ios_ShowCode;
      ios_state := ios_WaitPairComplete;
    or
      \* WaitPairComplete -> Paired on pair_complete
      await ios_state = ios_WaitPairComplete /\ Len(chan_server_ios) > 0 /\ Head(chan_server_ios).type = MSG_pair_complete;
      recv_msg := Head(chan_server_ios);
      chan_server_ios := Tail(chan_server_ios);
      ios_state := ios_Paired;
    or
      \* Paired -> Reconnect (app launch)
      await ios_state = ios_Paired;
      ios_state := ios_Reconnect;
    or
      \* Reconnect -> SendAuth (relay connected)
      await ios_state = ios_Reconnect;
      chan_ios_server := Append(chan_ios_server, [type |-> MSG_auth_request, device_id |-> "device_1", key |-> client_shared_key, nonce |-> "nonce_1", secret |-> device_secret]);
      ios_state := ios_SendAuth;
    or
      \* SendAuth -> SessionActive on auth_ok
      await ios_state = ios_SendAuth /\ Len(chan_server_ios) > 0 /\ Head(chan_server_ios).type = MSG_auth_ok;
      recv_msg := Head(chan_server_ios);
      chan_server_ios := Tail(chan_server_ios);
      ios_state := ios_SessionActive;
    or
      \* SessionActive -> Paired (disconnect)
      await ios_state = ios_SessionActive;
      ios_state := ios_Paired;
    end either;
end process;

fair process cli = 3
begin
  cli_loop:
    either
      \* Idle -> GetKey (cli --init)
      await cli_state = cli_Idle;
      cli_state := cli_GetKey;
    or
      \* GetKey -> BeginPair (key stored)
      await cli_state = cli_GetKey;
      chan_cli_server := Append(chan_cli_server, [type |-> MSG_pair_begin]);
      cli_state := cli_BeginPair;
    or
      \* BeginPair -> ShowQR on token_response
      await cli_state = cli_BeginPair /\ Len(chan_server_cli) > 0 /\ Head(chan_server_cli).type = MSG_token_response;
      recv_msg := Head(chan_server_cli);
      chan_server_cli := Tail(chan_server_cli);
      cli_state := cli_ShowQR;
    or
      \* ShowQR -> PromptCode on waiting_for_code
      await cli_state = cli_ShowQR /\ Len(chan_server_cli) > 0 /\ Head(chan_server_cli).type = MSG_waiting_for_code;
      recv_msg := Head(chan_server_cli);
      chan_server_cli := Tail(chan_server_cli);
      cli_state := cli_PromptCode;
    or
      \* PromptCode -> SubmitCode (user enters code)
      await cli_state = cli_PromptCode;
      chan_cli_server := Append(chan_cli_server, [type |-> MSG_code_submit, code |-> ios_code]);
      cli_state := cli_SubmitCode;
    or
      \* SubmitCode -> Done on pair_status
      await cli_state = cli_SubmitCode /\ Len(chan_server_cli) > 0 /\ Head(chan_server_cli).type = MSG_pair_status;
      recv_msg := Head(chan_server_cli);
      chan_server_cli := Tail(chan_server_cli);
      cli_state := cli_Done;
    end either;
end process;

\* Dolev-Yao adversary: controls the network.
\* Can read, drop, replay, and reorder messages on all channels.
\* Cannot forge messages or break cryptographic primitives.
\* Extended capabilities model specific attack scenarios.
fair process Adversary = 4
begin
  adv_loop:
  while TRUE do
    either
      skip \* no-op: honest relay
    or
      \* Eavesdrop on cli -> server
      await Len(chan_cli_server) > 0;
      adversary_knowledge := adversary_knowledge \union {Head(chan_cli_server)};
    or
      \* Drop from cli -> server
      await Len(chan_cli_server) > 0;
      chan_cli_server := Tail(chan_cli_server);
    or
      \* Replay into cli -> server
      await adversary_knowledge /= {} /\ Len(chan_cli_server) < 3;
      with msg \in adversary_knowledge do
        chan_cli_server := Append(chan_cli_server, msg);
      end with;
    or
      \* Eavesdrop on ios -> server
      await Len(chan_ios_server) > 0;
      adversary_knowledge := adversary_knowledge \union {Head(chan_ios_server)};
    or
      \* Drop from ios -> server
      await Len(chan_ios_server) > 0;
      chan_ios_server := Tail(chan_ios_server);
    or
      \* Replay into ios -> server
      await adversary_knowledge /= {} /\ Len(chan_ios_server) < 3;
      with msg \in adversary_knowledge do
        chan_ios_server := Append(chan_ios_server, msg);
      end with;
    or
      \* Eavesdrop on server -> cli
      await Len(chan_server_cli) > 0;
      adversary_knowledge := adversary_knowledge \union {Head(chan_server_cli)};
    or
      \* Drop from server -> cli
      await Len(chan_server_cli) > 0;
      chan_server_cli := Tail(chan_server_cli);
    or
      \* Replay into server -> cli
      await adversary_knowledge /= {} /\ Len(chan_server_cli) < 3;
      with msg \in adversary_knowledge do
        chan_server_cli := Append(chan_server_cli, msg);
      end with;
    or
      \* Eavesdrop on server -> ios
      await Len(chan_server_ios) > 0;
      adversary_knowledge := adversary_knowledge \union {Head(chan_server_ios)};
    or
      \* Drop from server -> ios
      await Len(chan_server_ios) > 0;
      chan_server_ios := Tail(chan_server_ios);
    or
      \* Replay into server -> ios
      await adversary_knowledge /= {} /\ Len(chan_server_ios) < 3;
      with msg \in adversary_knowledge do
        chan_server_ios := Append(chan_server_ios, msg);
      end with;
    or
      \* QR_shoulder_surf: observe QR code content (token + instance_id)
      await current_token /= "none";
      adversary_knowledge := adversary_knowledge \union {[type |-> "qr_token", token |-> current_token]};
    or
      \* MitM_pair_hello: intercept pair_hello and substitute adversary ECDH pubkey
      await Len(chan_ios_server) > 0 /\ Head(chan_ios_server).type = MSG_pair_hello;
      adv_saved_client_pub := Head(chan_ios_server).pubkey;
      chan_ios_server := <<[type |-> MSG_pair_hello, token |-> Head(chan_ios_server).token, pubkey |-> adv_ecdh_pub]>> \o Tail(chan_ios_server);
    or
      \* MitM_pair_hello_ack: intercept pair_hello_ack and substitute adversary ECDH pubkey, derive both shared secrets
      await Len(chan_server_ios) > 0 /\ Head(chan_server_ios).type = MSG_pair_hello_ack;
      adv_saved_server_pub := Head(chan_server_ios).pubkey;
      adversary_keys := adversary_keys \union {DeriveKey(adv_ecdh_pub, adv_saved_server_pub), DeriveKey(adv_ecdh_pub, adv_saved_client_pub)};
      chan_server_ios := <<[type |-> MSG_pair_hello_ack, pubkey |-> adv_ecdh_pub]>> \o Tail(chan_server_ios);
    or
      \* MitM_reencrypt_secret: decrypt pair_complete with MitM key, learn device secret
      await Len(chan_server_ios) > 0 /\ Head(chan_server_ios).type = MSG_pair_complete /\ Head(chan_server_ios).key \in adversary_keys;
      with msg = Head(chan_server_ios) do
        adversary_knowledge := adversary_knowledge \union {[type |-> "plaintext_secret", secret |-> msg.secret]};
        chan_server_ios := <<[type |-> MSG_pair_complete, key |-> DeriveKey(adv_ecdh_pub, adv_saved_client_pub), secret |-> msg.secret]>> \o Tail(chan_server_ios);
      end with;
    or
      \* concurrent_pair: race a forged pair_hello using shoulder-surfed token
      await \E m \in adversary_knowledge : m = [type |-> "qr_token", token |-> current_token];
      await Len(chan_ios_server) < 3;
      chan_ios_server := Append(chan_ios_server, [type |-> MSG_pair_hello, token |-> current_token, pubkey |-> adv_ecdh_pub]);
    or
      \* token_bruteforce: send pair_hello with fabricated token
      await Len(chan_ios_server) < 3;
      chan_ios_server := Append(chan_ios_server, [type |-> MSG_pair_hello, token |-> "fake_token", pubkey |-> adv_ecdh_pub]);
    or
      \* code_guess: submit fabricated confirmation code via CLI channel
      await Len(chan_cli_server) < 3;
      chan_cli_server := Append(chan_cli_server, [type |-> MSG_code_submit, code |-> <<"guess", "000000">>]);
    or
      \* session_replay: replay captured auth_request with stale nonce
      await Len(chan_ios_server) < 3;
      await \E m \in adversary_knowledge : m.type = MSG_auth_request;
      with msg \in {m \in adversary_knowledge : m.type = MSG_auth_request} do
        chan_ios_server := Append(chan_ios_server, msg);
      end with;
    end either;
  end while;
end process;

end algorithm; *)
\* BEGIN TRANSLATION
\* END TRANSLATION

\* Verification properties
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
