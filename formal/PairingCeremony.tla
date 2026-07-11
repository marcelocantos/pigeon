---- MODULE PairingCeremony ----
\* Auto-generated from protocol YAML. Do not edit.

EXTENDS Integers, Sequences, FiniteSets, TLC

\* States for acceptor
acceptor_Idle == "acceptor_Idle"
acceptor_GeneratingEphemeral == "acceptor_GeneratingEphemeral"
acceptor_RegisteringRelay == "acceptor_RegisteringRelay"
acceptor_WaitingForHello == "acceptor_WaitingForHello"
acceptor_WaitingForReveal == "acceptor_WaitingForReveal"
acceptor_DerivingCode == "acceptor_DerivingCode"
acceptor_AwaitingUserConfirm == "acceptor_AwaitingUserConfirm"
acceptor_AwaitingPeerConfirm == "acceptor_AwaitingPeerConfirm"
acceptor_Paired == "acceptor_Paired"
acceptor_Aborted == "acceptor_Aborted"

\* States for initiator
initiator_Idle == "initiator_Idle"
initiator_DecodingToken == "initiator_DecodingToken"
initiator_GeneratingEphemeral == "initiator_GeneratingEphemeral"
initiator_ConnectingRelay == "initiator_ConnectingRelay"
initiator_AwaitingWelcome == "initiator_AwaitingWelcome"
initiator_Revealing == "initiator_Revealing"
initiator_DerivingCode == "initiator_DerivingCode"
initiator_AwaitingUserConfirm == "initiator_AwaitingUserConfirm"
initiator_AwaitingPeerConfirm == "initiator_AwaitingPeerConfirm"
initiator_Paired == "initiator_Paired"
initiator_Aborted == "initiator_Aborted"

\* Message types
MSG_hello == "hello"
MSG_welcome == "welcome"
MSG_reveal == "reveal"
MSG_confirm_to_initiator == "confirm_to_initiator"
MSG_confirm_to_acceptor == "confirm_to_acceptor"

\* Event types
EVT_code_ready == "code_ready"
EVT_commit_fail == "commit_fail"
EVT_ephemeral_ready == "ephemeral_ready"
EVT_pair_begin == "pair_begin"
EVT_recv_confirm_to_acceptor == "recv_confirm_to_acceptor"
EVT_recv_confirm_to_initiator == "recv_confirm_to_initiator"
EVT_recv_hello == "recv_hello"
EVT_recv_reveal == "recv_reveal"
EVT_recv_welcome == "recv_welcome"
EVT_relay_connected == "relay_connected"
EVT_relay_registered == "relay_registered"
EVT_reveal_sent == "reveal_sent"
EVT_token_decoded == "token_decoded"
EVT_token_received == "token_received"
EVT_user_cancel == "user_cancel"
EVT_user_confirm == "user_confirm"

\* Stable rank for symbolic pubkeys so DeriveCode is order-independent.
KeyRank(k) == CASE k = "adv_eph" -> 0 [] k = "initiator_eph" -> 1 [] k = "acceptor_eph" -> 2 [] OTHER -> 3
\* Confirmation code derived from both ephemeral pubkeys (order-independent).
DeriveCode(a, b) == IF KeyRank(a) <= KeyRank(b) THEN <<"code", a, b>> ELSE <<"code", b, a>>
\* Symbolic commitment: commit = H(eph||blind). Models the binding that SHA256 provides in the implementation.
CommitOf(eph) == CASE eph = "initiator_eph" -> "initiator_commit" [] eph = "adv_eph" -> "adv_commit" [] eph = "acceptor_eph" -> "acceptor_commit" [] OTHER -> "unknown_commit"
\* True iff reveal(eph, blind) opens the previously received commit.
CommitMatches(commit, eph) == commit = CommitOf(eph)
\* Interface contract with SessionMachine: a PairingRecord is well-formed iff it carries a non-default peer instance ID and ephemeral pubkey. Pairing proves it as a postcondition (PairingProducesValidRecord); SessionMachine assumes it over its initial state.
ValidPairingRecord(r) == r.peer_instance_id /= "none" /\ r.peer_eph_pub /= "none"



CONSTANTS acceptor_identity_pub, acceptor_instance_id, initiator_identity_pub, initiator_instance_id, adversary_keys, adv_eph_pub, adv_commit, adv_saved_acceptor_eph, adv_saved_initiator_eph, adv_saved_initiator_commit

VARIABLES
    acceptor_state,
    initiator_state,
    acceptor_eph_pub,
    acceptor_received_commit,
    acceptor_received_eph_pub,
    acceptor_received_identity,
    acceptor_received_instance,
    acceptor_commit_ok,
    acceptor_code,
    acceptor_user_confirmed,
    acceptor_received_confirm,
    initiator_eph_pub,
    initiator_commit,
    received_acceptor_eph_pub,
    received_acceptor_identity,
    received_acceptor_instance,
    initiator_received_eph_pub,
    initiator_received_identity,
    initiator_received_instance,
    initiator_code,
    initiator_user_confirmed,
    initiator_received_confirm,
    received_hello,
    received_reveal,
    received_confirm_to_acceptor,
    received_welcome,
    received_confirm_to_initiator

vars == <<acceptor_state, initiator_state, acceptor_eph_pub, acceptor_received_commit, acceptor_received_eph_pub, acceptor_received_identity, acceptor_received_instance, acceptor_commit_ok, acceptor_code, acceptor_user_confirmed, acceptor_received_confirm, initiator_eph_pub, initiator_commit, received_acceptor_eph_pub, received_acceptor_identity, received_acceptor_instance, initiator_received_eph_pub, initiator_received_identity, initiator_received_instance, initiator_code, initiator_user_confirmed, initiator_received_confirm, received_hello, received_reveal, received_confirm_to_acceptor, received_welcome, received_confirm_to_initiator>>

Init ==
    /\ acceptor_state = acceptor_Idle
    /\ initiator_state = initiator_Idle
    /\ acceptor_eph_pub = "none"
    /\ acceptor_received_commit = "none"
    /\ acceptor_received_eph_pub = "none"
    /\ acceptor_received_identity = "none"
    /\ acceptor_received_instance = "none"
    /\ acceptor_commit_ok = "false"
    /\ acceptor_code = <<"none">>
    /\ acceptor_user_confirmed = "false"
    /\ acceptor_received_confirm = "false"
    /\ initiator_eph_pub = "none"
    /\ initiator_commit = "none"
    /\ received_acceptor_eph_pub = "none"
    /\ received_acceptor_identity = "none"
    /\ received_acceptor_instance = "none"
    /\ initiator_received_eph_pub = "none"
    /\ initiator_received_identity = "none"
    /\ initiator_received_instance = "none"
    /\ initiator_code = <<"none">>
    /\ initiator_user_confirmed = "false"
    /\ initiator_received_confirm = "false"
    /\ received_hello = [type |-> "none"]
    /\ received_reveal = [type |-> "none"]
    /\ received_confirm_to_acceptor = [type |-> "none"]
    /\ received_welcome = [type |-> "none"]
    /\ received_confirm_to_initiator = [type |-> "none"]

\* acceptor: Idle -> GeneratingEphemeral (pair_begin)
acceptor_Idle_to_GeneratingEphemeral_pair_begin ==
    /\ acceptor_state = acceptor_Idle
    /\ acceptor_state' = acceptor_GeneratingEphemeral
    /\ acceptor_eph_pub' = "acceptor_eph"
    /\ UNCHANGED <<initiator_state, acceptor_received_commit, acceptor_received_eph_pub, acceptor_received_identity, acceptor_received_instance, acceptor_commit_ok, acceptor_code, acceptor_user_confirmed, acceptor_received_confirm, initiator_eph_pub, initiator_commit, received_acceptor_eph_pub, received_acceptor_identity, received_acceptor_instance, initiator_received_eph_pub, initiator_received_identity, initiator_received_instance, initiator_code, initiator_user_confirmed, initiator_received_confirm, received_hello, received_reveal, received_confirm_to_acceptor, received_welcome, received_confirm_to_initiator>>

\* acceptor: GeneratingEphemeral -> RegisteringRelay (ephemeral_ready)
acceptor_GeneratingEphemeral_to_RegisteringRelay_ephemeral_ready ==
    /\ acceptor_state = acceptor_GeneratingEphemeral
    /\ acceptor_state' = acceptor_RegisteringRelay
    /\ UNCHANGED <<initiator_state, acceptor_eph_pub, acceptor_received_commit, acceptor_received_eph_pub, acceptor_received_identity, acceptor_received_instance, acceptor_commit_ok, acceptor_code, acceptor_user_confirmed, acceptor_received_confirm, initiator_eph_pub, initiator_commit, received_acceptor_eph_pub, received_acceptor_identity, received_acceptor_instance, initiator_received_eph_pub, initiator_received_identity, initiator_received_instance, initiator_code, initiator_user_confirmed, initiator_received_confirm, received_hello, received_reveal, received_confirm_to_acceptor, received_welcome, received_confirm_to_initiator>>

\* acceptor: RegisteringRelay -> WaitingForHello (relay_registered)
acceptor_RegisteringRelay_to_WaitingForHello_relay_registered ==
    /\ acceptor_state = acceptor_RegisteringRelay
    /\ acceptor_state' = acceptor_WaitingForHello
    /\ UNCHANGED <<initiator_state, acceptor_eph_pub, acceptor_received_commit, acceptor_received_eph_pub, acceptor_received_identity, acceptor_received_instance, acceptor_commit_ok, acceptor_code, acceptor_user_confirmed, acceptor_received_confirm, initiator_eph_pub, initiator_commit, received_acceptor_eph_pub, received_acceptor_identity, received_acceptor_instance, initiator_received_eph_pub, initiator_received_identity, initiator_received_instance, initiator_code, initiator_user_confirmed, initiator_received_confirm, received_hello, received_reveal, received_confirm_to_acceptor, received_welcome, received_confirm_to_initiator>>

\* acceptor: WaitingForHello -> WaitingForReveal on recv hello
acceptor_WaitingForHello_to_WaitingForReveal_on_hello ==
    /\ acceptor_state = acceptor_WaitingForHello
    /\ received_hello.type = MSG_hello
    /\ received_hello' = [type |-> "none"]
    /\ received_welcome' = [type |-> MSG_welcome, eph_pub |-> acceptor_eph_pub, identity_pub |-> acceptor_identity_pub, instance_id |-> acceptor_instance_id]
    /\ acceptor_state' = acceptor_WaitingForReveal
    /\ acceptor_received_commit' = received_hello.commit
    /\ acceptor_received_identity' = received_hello.identity_pub
    /\ acceptor_received_instance' = received_hello.instance_id
    /\ UNCHANGED <<initiator_state, acceptor_eph_pub, acceptor_received_eph_pub, acceptor_commit_ok, acceptor_code, acceptor_user_confirmed, acceptor_received_confirm, initiator_eph_pub, initiator_commit, received_acceptor_eph_pub, received_acceptor_identity, received_acceptor_instance, initiator_received_eph_pub, initiator_received_identity, initiator_received_instance, initiator_code, initiator_user_confirmed, initiator_received_confirm, received_reveal, received_confirm_to_acceptor, received_confirm_to_initiator>>

\* acceptor: WaitingForReveal -> DerivingCode on recv reveal
acceptor_WaitingForReveal_to_DerivingCode_on_reveal ==
    /\ acceptor_state = acceptor_WaitingForReveal
    /\ received_reveal.type = MSG_reveal
    /\ received_reveal' = [type |-> "none"]
    /\ acceptor_state' = acceptor_DerivingCode
    /\ acceptor_received_eph_pub' = received_reveal.eph_pub
    /\ acceptor_commit_ok' = IF CommitMatches(acceptor_received_commit, received_reveal.eph_pub) THEN "true" ELSE "false"
    /\ acceptor_code' = IF CommitMatches(acceptor_received_commit, received_reveal.eph_pub) THEN DeriveCode(acceptor_eph_pub, received_reveal.eph_pub) ELSE <<"none">>
    /\ UNCHANGED <<initiator_state, acceptor_eph_pub, acceptor_received_commit, acceptor_received_identity, acceptor_received_instance, acceptor_user_confirmed, acceptor_received_confirm, initiator_eph_pub, initiator_commit, received_acceptor_eph_pub, received_acceptor_identity, received_acceptor_instance, initiator_received_eph_pub, initiator_received_identity, initiator_received_instance, initiator_code, initiator_user_confirmed, initiator_received_confirm, received_hello, received_confirm_to_acceptor, received_welcome, received_confirm_to_initiator>>

\* acceptor: DerivingCode -> AwaitingUserConfirm (code_ready)
acceptor_DerivingCode_to_AwaitingUserConfirm_code_ready ==
    /\ acceptor_state = acceptor_DerivingCode
    /\ acceptor_state' = acceptor_AwaitingUserConfirm
    /\ UNCHANGED <<initiator_state, acceptor_eph_pub, acceptor_received_commit, acceptor_received_eph_pub, acceptor_received_identity, acceptor_received_instance, acceptor_commit_ok, acceptor_code, acceptor_user_confirmed, acceptor_received_confirm, initiator_eph_pub, initiator_commit, received_acceptor_eph_pub, received_acceptor_identity, received_acceptor_instance, initiator_received_eph_pub, initiator_received_identity, initiator_received_instance, initiator_code, initiator_user_confirmed, initiator_received_confirm, received_hello, received_reveal, received_confirm_to_acceptor, received_welcome, received_confirm_to_initiator>>

\* acceptor: AwaitingUserConfirm -> AwaitingPeerConfirm (user_confirm)
acceptor_AwaitingUserConfirm_to_AwaitingPeerConfirm_user_confirm ==
    /\ acceptor_state = acceptor_AwaitingUserConfirm
    /\ received_confirm_to_initiator' = [type |-> MSG_confirm_to_initiator]
    /\ acceptor_state' = acceptor_AwaitingPeerConfirm
    /\ acceptor_user_confirmed' = "true"
    /\ UNCHANGED <<initiator_state, acceptor_eph_pub, acceptor_received_commit, acceptor_received_eph_pub, acceptor_received_identity, acceptor_received_instance, acceptor_commit_ok, acceptor_code, acceptor_received_confirm, initiator_eph_pub, initiator_commit, received_acceptor_eph_pub, received_acceptor_identity, received_acceptor_instance, initiator_received_eph_pub, initiator_received_identity, initiator_received_instance, initiator_code, initiator_user_confirmed, initiator_received_confirm, received_hello, received_reveal, received_confirm_to_acceptor, received_welcome>>

\* acceptor: AwaitingPeerConfirm -> Paired on recv confirm_to_acceptor
acceptor_AwaitingPeerConfirm_to_Paired_on_confirm_to_acceptor ==
    /\ acceptor_state = acceptor_AwaitingPeerConfirm
    /\ received_confirm_to_acceptor.type = MSG_confirm_to_acceptor
    /\ received_confirm_to_acceptor' = [type |-> "none"]
    /\ acceptor_state' = acceptor_Paired
    /\ acceptor_received_confirm' = "true"
    /\ UNCHANGED <<initiator_state, acceptor_eph_pub, acceptor_received_commit, acceptor_received_eph_pub, acceptor_received_identity, acceptor_received_instance, acceptor_commit_ok, acceptor_code, acceptor_user_confirmed, initiator_eph_pub, initiator_commit, received_acceptor_eph_pub, received_acceptor_identity, received_acceptor_instance, initiator_received_eph_pub, initiator_received_identity, initiator_received_instance, initiator_code, initiator_user_confirmed, initiator_received_confirm, received_hello, received_reveal, received_welcome, received_confirm_to_initiator>>

\* acceptor: AwaitingUserConfirm -> Aborted (user_cancel)
acceptor_AwaitingUserConfirm_to_Aborted_user_cancel ==
    /\ acceptor_state = acceptor_AwaitingUserConfirm
    /\ acceptor_state' = acceptor_Aborted
    /\ UNCHANGED <<initiator_state, acceptor_eph_pub, acceptor_received_commit, acceptor_received_eph_pub, acceptor_received_identity, acceptor_received_instance, acceptor_commit_ok, acceptor_code, acceptor_user_confirmed, acceptor_received_confirm, initiator_eph_pub, initiator_commit, received_acceptor_eph_pub, received_acceptor_identity, received_acceptor_instance, initiator_received_eph_pub, initiator_received_identity, initiator_received_instance, initiator_code, initiator_user_confirmed, initiator_received_confirm, received_hello, received_reveal, received_confirm_to_acceptor, received_welcome, received_confirm_to_initiator>>

\* acceptor: AwaitingPeerConfirm -> Aborted (user_cancel)
acceptor_AwaitingPeerConfirm_to_Aborted_user_cancel ==
    /\ acceptor_state = acceptor_AwaitingPeerConfirm
    /\ acceptor_state' = acceptor_Aborted
    /\ UNCHANGED <<initiator_state, acceptor_eph_pub, acceptor_received_commit, acceptor_received_eph_pub, acceptor_received_identity, acceptor_received_instance, acceptor_commit_ok, acceptor_code, acceptor_user_confirmed, acceptor_received_confirm, initiator_eph_pub, initiator_commit, received_acceptor_eph_pub, received_acceptor_identity, received_acceptor_instance, initiator_received_eph_pub, initiator_received_identity, initiator_received_instance, initiator_code, initiator_user_confirmed, initiator_received_confirm, received_hello, received_reveal, received_confirm_to_acceptor, received_welcome, received_confirm_to_initiator>>

\* acceptor: WaitingForReveal -> Aborted (commit_fail)
acceptor_WaitingForReveal_to_Aborted_commit_fail ==
    /\ acceptor_state = acceptor_WaitingForReveal
    /\ acceptor_state' = acceptor_Aborted
    /\ UNCHANGED <<initiator_state, acceptor_eph_pub, acceptor_received_commit, acceptor_received_eph_pub, acceptor_received_identity, acceptor_received_instance, acceptor_commit_ok, acceptor_code, acceptor_user_confirmed, acceptor_received_confirm, initiator_eph_pub, initiator_commit, received_acceptor_eph_pub, received_acceptor_identity, received_acceptor_instance, initiator_received_eph_pub, initiator_received_identity, initiator_received_instance, initiator_code, initiator_user_confirmed, initiator_received_confirm, received_hello, received_reveal, received_confirm_to_acceptor, received_welcome, received_confirm_to_initiator>>


\* initiator: Idle -> DecodingToken (token_received)
initiator_Idle_to_DecodingToken_token_received ==
    /\ initiator_state = initiator_Idle
    /\ initiator_state' = initiator_DecodingToken
    /\ received_acceptor_eph_pub' = "acceptor_eph"
    /\ received_acceptor_identity' = "acceptor_id"
    /\ received_acceptor_instance' = "acceptor_instance"
    /\ UNCHANGED <<acceptor_state, acceptor_eph_pub, acceptor_received_commit, acceptor_received_eph_pub, acceptor_received_identity, acceptor_received_instance, acceptor_commit_ok, acceptor_code, acceptor_user_confirmed, acceptor_received_confirm, initiator_eph_pub, initiator_commit, initiator_received_eph_pub, initiator_received_identity, initiator_received_instance, initiator_code, initiator_user_confirmed, initiator_received_confirm, received_hello, received_reveal, received_confirm_to_acceptor, received_welcome, received_confirm_to_initiator>>

\* initiator: DecodingToken -> GeneratingEphemeral (token_decoded)
initiator_DecodingToken_to_GeneratingEphemeral_token_decoded ==
    /\ initiator_state = initiator_DecodingToken
    /\ initiator_state' = initiator_GeneratingEphemeral
    /\ initiator_eph_pub' = "initiator_eph"
    /\ initiator_commit' = "initiator_commit"
    /\ UNCHANGED <<acceptor_state, acceptor_eph_pub, acceptor_received_commit, acceptor_received_eph_pub, acceptor_received_identity, acceptor_received_instance, acceptor_commit_ok, acceptor_code, acceptor_user_confirmed, acceptor_received_confirm, received_acceptor_eph_pub, received_acceptor_identity, received_acceptor_instance, initiator_received_eph_pub, initiator_received_identity, initiator_received_instance, initiator_code, initiator_user_confirmed, initiator_received_confirm, received_hello, received_reveal, received_confirm_to_acceptor, received_welcome, received_confirm_to_initiator>>

\* initiator: GeneratingEphemeral -> ConnectingRelay (ephemeral_ready)
initiator_GeneratingEphemeral_to_ConnectingRelay_ephemeral_ready ==
    /\ initiator_state = initiator_GeneratingEphemeral
    /\ initiator_state' = initiator_ConnectingRelay
    /\ UNCHANGED <<acceptor_state, acceptor_eph_pub, acceptor_received_commit, acceptor_received_eph_pub, acceptor_received_identity, acceptor_received_instance, acceptor_commit_ok, acceptor_code, acceptor_user_confirmed, acceptor_received_confirm, initiator_eph_pub, initiator_commit, received_acceptor_eph_pub, received_acceptor_identity, received_acceptor_instance, initiator_received_eph_pub, initiator_received_identity, initiator_received_instance, initiator_code, initiator_user_confirmed, initiator_received_confirm, received_hello, received_reveal, received_confirm_to_acceptor, received_welcome, received_confirm_to_initiator>>

\* initiator: ConnectingRelay -> AwaitingWelcome (relay_connected)
initiator_ConnectingRelay_to_AwaitingWelcome_relay_connected ==
    /\ initiator_state = initiator_ConnectingRelay
    /\ received_hello' = [type |-> MSG_hello, commit |-> initiator_commit, identity_pub |-> initiator_identity_pub, instance_id |-> initiator_instance_id]
    /\ initiator_state' = initiator_AwaitingWelcome
    /\ UNCHANGED <<acceptor_state, acceptor_eph_pub, acceptor_received_commit, acceptor_received_eph_pub, acceptor_received_identity, acceptor_received_instance, acceptor_commit_ok, acceptor_code, acceptor_user_confirmed, acceptor_received_confirm, initiator_eph_pub, initiator_commit, received_acceptor_eph_pub, received_acceptor_identity, received_acceptor_instance, initiator_received_eph_pub, initiator_received_identity, initiator_received_instance, initiator_code, initiator_user_confirmed, initiator_received_confirm, received_reveal, received_confirm_to_acceptor, received_welcome, received_confirm_to_initiator>>

\* initiator: AwaitingWelcome -> Revealing on recv welcome
initiator_AwaitingWelcome_to_Revealing_on_welcome ==
    /\ initiator_state = initiator_AwaitingWelcome
    /\ received_welcome.type = MSG_welcome
    /\ received_welcome' = [type |-> "none"]
    /\ received_reveal' = [type |-> MSG_reveal, blind |-> "initiator_blind", eph_pub |-> initiator_eph_pub]
    /\ initiator_state' = initiator_Revealing
    /\ initiator_received_eph_pub' = received_welcome.eph_pub
    /\ initiator_received_identity' = received_welcome.identity_pub
    /\ initiator_received_instance' = received_welcome.instance_id
    /\ UNCHANGED <<acceptor_state, acceptor_eph_pub, acceptor_received_commit, acceptor_received_eph_pub, acceptor_received_identity, acceptor_received_instance, acceptor_commit_ok, acceptor_code, acceptor_user_confirmed, acceptor_received_confirm, initiator_eph_pub, initiator_commit, received_acceptor_eph_pub, received_acceptor_identity, received_acceptor_instance, initiator_code, initiator_user_confirmed, initiator_received_confirm, received_hello, received_confirm_to_acceptor, received_confirm_to_initiator>>

\* initiator: Revealing -> DerivingCode (reveal_sent)
initiator_Revealing_to_DerivingCode_reveal_sent ==
    /\ initiator_state = initiator_Revealing
    /\ initiator_state' = initiator_DerivingCode
    /\ initiator_code' = DeriveCode(initiator_eph_pub, initiator_received_eph_pub)
    /\ UNCHANGED <<acceptor_state, acceptor_eph_pub, acceptor_received_commit, acceptor_received_eph_pub, acceptor_received_identity, acceptor_received_instance, acceptor_commit_ok, acceptor_code, acceptor_user_confirmed, acceptor_received_confirm, initiator_eph_pub, initiator_commit, received_acceptor_eph_pub, received_acceptor_identity, received_acceptor_instance, initiator_received_eph_pub, initiator_received_identity, initiator_received_instance, initiator_user_confirmed, initiator_received_confirm, received_hello, received_reveal, received_confirm_to_acceptor, received_welcome, received_confirm_to_initiator>>

\* initiator: DerivingCode -> AwaitingUserConfirm (code_ready)
initiator_DerivingCode_to_AwaitingUserConfirm_code_ready ==
    /\ initiator_state = initiator_DerivingCode
    /\ initiator_state' = initiator_AwaitingUserConfirm
    /\ UNCHANGED <<acceptor_state, acceptor_eph_pub, acceptor_received_commit, acceptor_received_eph_pub, acceptor_received_identity, acceptor_received_instance, acceptor_commit_ok, acceptor_code, acceptor_user_confirmed, acceptor_received_confirm, initiator_eph_pub, initiator_commit, received_acceptor_eph_pub, received_acceptor_identity, received_acceptor_instance, initiator_received_eph_pub, initiator_received_identity, initiator_received_instance, initiator_code, initiator_user_confirmed, initiator_received_confirm, received_hello, received_reveal, received_confirm_to_acceptor, received_welcome, received_confirm_to_initiator>>

\* initiator: AwaitingUserConfirm -> AwaitingPeerConfirm (user_confirm)
initiator_AwaitingUserConfirm_to_AwaitingPeerConfirm_user_confirm ==
    /\ initiator_state = initiator_AwaitingUserConfirm
    /\ received_confirm_to_acceptor' = [type |-> MSG_confirm_to_acceptor]
    /\ initiator_state' = initiator_AwaitingPeerConfirm
    /\ initiator_user_confirmed' = "true"
    /\ UNCHANGED <<acceptor_state, acceptor_eph_pub, acceptor_received_commit, acceptor_received_eph_pub, acceptor_received_identity, acceptor_received_instance, acceptor_commit_ok, acceptor_code, acceptor_user_confirmed, acceptor_received_confirm, initiator_eph_pub, initiator_commit, received_acceptor_eph_pub, received_acceptor_identity, received_acceptor_instance, initiator_received_eph_pub, initiator_received_identity, initiator_received_instance, initiator_code, initiator_received_confirm, received_hello, received_reveal, received_welcome, received_confirm_to_initiator>>

\* initiator: AwaitingPeerConfirm -> Paired on recv confirm_to_initiator
initiator_AwaitingPeerConfirm_to_Paired_on_confirm_to_initiator ==
    /\ initiator_state = initiator_AwaitingPeerConfirm
    /\ received_confirm_to_initiator.type = MSG_confirm_to_initiator
    /\ received_confirm_to_initiator' = [type |-> "none"]
    /\ initiator_state' = initiator_Paired
    /\ initiator_received_confirm' = "true"
    /\ UNCHANGED <<acceptor_state, acceptor_eph_pub, acceptor_received_commit, acceptor_received_eph_pub, acceptor_received_identity, acceptor_received_instance, acceptor_commit_ok, acceptor_code, acceptor_user_confirmed, acceptor_received_confirm, initiator_eph_pub, initiator_commit, received_acceptor_eph_pub, received_acceptor_identity, received_acceptor_instance, initiator_received_eph_pub, initiator_received_identity, initiator_received_instance, initiator_code, initiator_user_confirmed, received_hello, received_reveal, received_confirm_to_acceptor, received_welcome>>

\* initiator: AwaitingUserConfirm -> Aborted (user_cancel)
initiator_AwaitingUserConfirm_to_Aborted_user_cancel ==
    /\ initiator_state = initiator_AwaitingUserConfirm
    /\ initiator_state' = initiator_Aborted
    /\ UNCHANGED <<acceptor_state, acceptor_eph_pub, acceptor_received_commit, acceptor_received_eph_pub, acceptor_received_identity, acceptor_received_instance, acceptor_commit_ok, acceptor_code, acceptor_user_confirmed, acceptor_received_confirm, initiator_eph_pub, initiator_commit, received_acceptor_eph_pub, received_acceptor_identity, received_acceptor_instance, initiator_received_eph_pub, initiator_received_identity, initiator_received_instance, initiator_code, initiator_user_confirmed, initiator_received_confirm, received_hello, received_reveal, received_confirm_to_acceptor, received_welcome, received_confirm_to_initiator>>

\* initiator: AwaitingPeerConfirm -> Aborted (user_cancel)
initiator_AwaitingPeerConfirm_to_Aborted_user_cancel ==
    /\ initiator_state = initiator_AwaitingPeerConfirm
    /\ initiator_state' = initiator_Aborted
    /\ UNCHANGED <<acceptor_state, acceptor_eph_pub, acceptor_received_commit, acceptor_received_eph_pub, acceptor_received_identity, acceptor_received_instance, acceptor_commit_ok, acceptor_code, acceptor_user_confirmed, acceptor_received_confirm, initiator_eph_pub, initiator_commit, received_acceptor_eph_pub, received_acceptor_identity, received_acceptor_instance, initiator_received_eph_pub, initiator_received_identity, initiator_received_instance, initiator_code, initiator_user_confirmed, initiator_received_confirm, received_hello, received_reveal, received_confirm_to_acceptor, received_welcome, received_confirm_to_initiator>>


Next ==
    \/ acceptor_Idle_to_GeneratingEphemeral_pair_begin
    \/ acceptor_GeneratingEphemeral_to_RegisteringRelay_ephemeral_ready
    \/ acceptor_RegisteringRelay_to_WaitingForHello_relay_registered
    \/ acceptor_WaitingForHello_to_WaitingForReveal_on_hello
    \/ acceptor_WaitingForReveal_to_DerivingCode_on_reveal
    \/ acceptor_DerivingCode_to_AwaitingUserConfirm_code_ready
    \/ acceptor_AwaitingUserConfirm_to_AwaitingPeerConfirm_user_confirm
    \/ acceptor_AwaitingPeerConfirm_to_Paired_on_confirm_to_acceptor
    \/ acceptor_AwaitingUserConfirm_to_Aborted_user_cancel
    \/ acceptor_AwaitingPeerConfirm_to_Aborted_user_cancel
    \/ acceptor_WaitingForReveal_to_Aborted_commit_fail
    \/ initiator_Idle_to_DecodingToken_token_received
    \/ initiator_DecodingToken_to_GeneratingEphemeral_token_decoded
    \/ initiator_GeneratingEphemeral_to_ConnectingRelay_ephemeral_ready
    \/ initiator_ConnectingRelay_to_AwaitingWelcome_relay_connected
    \/ initiator_AwaitingWelcome_to_Revealing_on_welcome
    \/ initiator_Revealing_to_DerivingCode_reveal_sent
    \/ initiator_DerivingCode_to_AwaitingUserConfirm_code_ready
    \/ initiator_AwaitingUserConfirm_to_AwaitingPeerConfirm_user_confirm
    \/ initiator_AwaitingPeerConfirm_to_Paired_on_confirm_to_initiator
    \/ initiator_AwaitingUserConfirm_to_Aborted_user_cancel
    \/ initiator_AwaitingPeerConfirm_to_Aborted_user_cancel

Spec == Init /\ [][Next]_vars /\ WF_vars(Next)

\* ================================================================
\* Invariants and properties
\* ================================================================

\* If the adversary has injected its ephemeral pubkey AND both sides have computed codes, the codes differ — so the human comparison would catch the MitM.
MitMDetectedByCodeMismatch == (adv_eph_pub \in adversary_keys /\ acceptor_code /= <<"none">> /\ initiator_code /= <<"none">>) => acceptor_code /= initiator_code
\* When codes differ (MitM detected), at least one side never reaches Paired (because the human cancels).
MitMPreventsPairing == (adv_eph_pub \in adversary_keys /\ acceptor_code /= initiator_code) => (acceptor_state /= acceptor_Paired \/ initiator_state /= initiator_Paired)
\* Without adversary interference, both sides derive the same 6-digit code.
HonestPairingMatchesCodes == (adversary_keys = {} /\ acceptor_code /= <<"none">> /\ initiator_code /= <<"none">>) => acceptor_code = initiator_code
\* Without adversary interference and with both users pressing y, both sides eventually reach Paired.
HonestPairingCompletes == <>(acceptor_state = acceptor_Paired /\ initiator_state = initiator_Paired)
\* Postcondition: whenever either side reaches Paired, the implicit PairingRecord built from received-peer fields satisfies the ValidPairingRecord interface contract that SessionMachine assumes.
PairingProducesValidRecord == (acceptor_state = acceptor_Paired => ValidPairingRecord([peer_instance_id |-> acceptor_received_instance, peer_eph_pub |-> acceptor_received_eph_pub])) /\ (initiator_state = initiator_Paired => ValidPairingRecord([peer_instance_id |-> initiator_received_instance, peer_eph_pub |-> initiator_received_eph_pub]))
\* Acceptor never derives a SAS code unless the reveal opened the hello commit — commitment binding (anti-grind).
CodeRequiresCommit == acceptor_code /= <<"none">> => acceptor_commit_ok = "true"
\* When commit verification succeeded, the revealed eph is exactly the one bound by the earlier hello commit; the acceptor cannot have been shown a ground key after the fact.
RevealBoundByCommit == (acceptor_received_eph_pub /= "none" /\ acceptor_commit_ok = "true") => CommitMatches(acceptor_received_commit, acceptor_received_eph_pub)

====
