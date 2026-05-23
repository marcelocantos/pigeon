---- MODULE SessionMachine ----
\* Auto-generated from protocol YAML. Do not edit.
\* Phase: Transport

EXTENDS Integers, Sequences, FiniteSets, TLC

\* States for backend
backend_Paired == "backend_Paired"
backend_SessionActive == "backend_SessionActive"
backend_RelayConnected == "backend_RelayConnected"
backend_CandidatesAdvertised == "backend_CandidatesAdvertised"
backend_AltActive == "backend_AltActive"
backend_RelayBackoff == "backend_RelayBackoff"
backend_AltDegraded == "backend_AltDegraded"

\* States for client
client_Paired == "client_Paired"
client_SessionActive == "client_SessionActive"
client_RelayConnected == "client_RelayConnected"
client_PairDialing == "client_PairDialing"
client_PairChecking == "client_PairChecking"
client_AltActive == "client_AltActive"
client_RelayFallback == "client_RelayFallback"

\* Message types
MSG_candidates == "candidates"
MSG_pair_check == "pair_check"
MSG_pair_check_ack == "pair_check_ack"
MSG_path_ping == "path_ping"
MSG_path_pong == "path_pong"

\* Event types
EVT_alt_datagram == "alt_datagram"
EVT_alt_error == "alt_error"
EVT_alt_stream_data == "alt_stream_data"
EVT_alt_stream_error == "alt_stream_error"
EVT_app_force_fallback == "app_force_fallback"
EVT_app_send == "app_send"
EVT_app_send_datagram == "app_send_datagram"
EVT_backoff_expired == "backoff_expired"
EVT_candidates_changed == "candidates_changed"
EVT_candidates_gathered == "candidates_gathered"
EVT_candidates_refresh_tick == "candidates_refresh_tick"
EVT_candidates_timeout == "candidates_timeout"
EVT_dial_failed == "dial_failed"
EVT_dial_ok == "dial_ok"
EVT_ping_tick == "ping_tick"
EVT_ping_timeout == "ping_timeout"
EVT_recv_candidates == "recv_candidates"
EVT_recv_pair_check == "recv_pair_check"
EVT_recv_pair_check_ack == "recv_pair_check_ack"
EVT_recv_path_ping == "recv_path_ping"
EVT_recv_path_pong == "recv_path_pong"
EVT_relay_datagram == "relay_datagram"
EVT_relay_ok == "relay_ok"
EVT_relay_stream_data == "relay_stream_data"
EVT_relay_stream_error == "relay_stream_error"
EVT_verify_timeout == "verify_timeout"

\* Command types
CMD_cancel_pong_timeout == "cancel_pong_timeout"
CMD_close_alt_path == "close_alt_path"
CMD_deliver_recv == "deliver_recv"
CMD_deliver_recv_datagram == "deliver_recv_datagram"
CMD_deliver_recv_error == "deliver_recv_error"
CMD_dial_candidate == "dial_candidate"
CMD_reset_alt_ready == "reset_alt_ready"
CMD_send_active_datagram == "send_active_datagram"
CMD_send_candidates == "send_candidates"
CMD_send_pair_check == "send_pair_check"
CMD_send_pair_check_ack == "send_pair_check_ack"
CMD_send_path_ping == "send_path_ping"
CMD_send_path_pong == "send_path_pong"
CMD_set_crypto_datagram == "set_crypto_datagram"
CMD_signal_alt_ready == "signal_alt_ready"
CMD_start_alt_dg_reader == "start_alt_dg_reader"
CMD_start_alt_stream_reader == "start_alt_stream_reader"
CMD_start_backoff_timer == "start_backoff_timer"
CMD_start_monitor == "start_monitor"
CMD_start_pong_timeout == "start_pong_timeout"
CMD_stop_alt_dg_reader == "stop_alt_dg_reader"
CMD_stop_alt_stream_reader == "stop_alt_stream_reader"
CMD_stop_monitor == "stop_monitor"
CMD_write_active_stream == "write_active_stream"

\* deterministic ordering for ECDH
KeyRank(k) == CASE k = "adv_pub" -> 0 [] k = "client_pub" -> 1 [] k = "backend_pub" -> 2 [] OTHER -> 3
\* symbolic ECDH
DeriveKey(a, b) == IF KeyRank(a) <= KeyRank(b) THEN <<"ecdh", a, b>> ELSE <<"ecdh", b, a>>
\* confirmation code from pubkeys
DeriveCode(a, b) == IF KeyRank(a) <= KeyRank(b) THEN <<"code", a, b>> ELSE <<"code", b, a>>
\* minimum of two values
Min(a, b) == IF a < b THEN a ELSE b
\* Interface contract with PairingCeremony.tla: a PairingRecord is well-formed iff it carries a non-default peer instance ID and ephemeral pubkey. SessionMachine treats this as an axiom over its initial state — the session begins with a record that PairingCeremony.tla's PairingProducesValidRecord postcondition has already established.
ValidPairingRecord(r) == r.peer_instance_id /= "none" /\ r.peer_eph_pub /= "none"



CONSTANTS local_pair_check_challenge, received_pair_check_challenge, instance_id, max_ping_failures, max_backoff_level, local_listen_addr, local_candidates, remote_candidates, pair_check_acks, active_pair_id

VARIABLES
    backend_state,
    client_state,
    ping_failures,
    backoff_level,
    b_active_path,
    c_active_path,
    b_dispatcher_path,
    c_dispatcher_path,
    monitor_target,
    alt_signal,
    received_pair_check,
    received_path_pong,
    received_candidates,
    received_pair_check_ack,
    received_path_ping

vars == <<backend_state, client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Init ==
    /\ backend_state = backend_RelayConnected
    /\ client_state = client_RelayConnected
    /\ ping_failures = 0
    /\ backoff_level = 0
    /\ b_active_path = "relay"
    /\ c_active_path = "relay"
    /\ b_dispatcher_path = "relay"
    /\ c_dispatcher_path = "relay"
    /\ monitor_target = "none"
    /\ alt_signal = "pending"
    /\ received_pair_check = [type |-> "none"]
    /\ received_path_pong = [type |-> "none"]
    /\ received_candidates = [type |-> "none"]
    /\ received_pair_check_ack = [type |-> "none"]
    /\ received_path_ping = [type |-> "none"]

\* backend: RelayConnected -> CandidatesAdvertised (candidates_gathered)
backend_RelayConnected_to_CandidatesAdvertised_candidates_gathered ==
    /\ backend_state = backend_RelayConnected
    /\ received_candidates' = [type |-> MSG_candidates, challenge |-> local_pair_check_challenge, set |-> local_candidates]
    /\ backend_state' = backend_CandidatesAdvertised
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_pair_check_ack, received_path_ping>>

Cmds_backend_RelayConnected_to_CandidatesAdvertised_candidates_gathered == {CMD_send_candidates}

\* backend: CandidatesAdvertised -> AltActive on recv pair_check [challenge_valid]
backend_CandidatesAdvertised_to_AltActive_on_pair_check_challenge_valid ==
    /\ backend_state = backend_CandidatesAdvertised
    /\ received_pair_check.type = MSG_pair_check
    /\ received_pair_check_challenge = local_pair_check_challenge
    /\ received_pair_check' = [type |-> "none"]
    /\ received_pair_check_ack' = [type |-> MSG_pair_check_ack]
    /\ backend_state' = backend_AltActive
    /\ ping_failures' = 0
    /\ backoff_level' = 0
    /\ b_active_path' = "alt"
    /\ b_dispatcher_path' = "alt"
    /\ monitor_target' = "alt"
    /\ alt_signal' = "ready"
    /\ UNCHANGED <<client_state, c_active_path, c_dispatcher_path, received_path_pong, received_candidates, received_path_ping>>

Cmds_backend_CandidatesAdvertised_to_AltActive_on_pair_check_challenge_valid == {CMD_send_pair_check_ack, CMD_start_alt_stream_reader, CMD_start_alt_dg_reader, CMD_start_monitor, CMD_signal_alt_ready, CMD_set_crypto_datagram}

\* backend: CandidatesAdvertised -> RelayConnected on recv pair_check [challenge_invalid]
backend_CandidatesAdvertised_to_RelayConnected_on_pair_check_challenge_invalid ==
    /\ backend_state = backend_CandidatesAdvertised
    /\ received_pair_check.type = MSG_pair_check
    /\ received_pair_check_challenge /= local_pair_check_challenge
    /\ received_pair_check' = [type |-> "none"]
    /\ backend_state' = backend_RelayConnected
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

\* backend: CandidatesAdvertised -> RelayBackoff (candidates_timeout)
backend_CandidatesAdvertised_to_RelayBackoff_candidates_timeout ==
    /\ backend_state = backend_CandidatesAdvertised
    /\ backend_state' = backend_RelayBackoff
    /\ backoff_level' = Min(backoff_level + 1, max_backoff_level)
    /\ alt_signal' = "pending"
    /\ UNCHANGED <<client_state, ping_failures, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_CandidatesAdvertised_to_RelayBackoff_candidates_timeout == {CMD_reset_alt_ready, CMD_start_backoff_timer}

\* backend: AltActive -> AltActive (ping_tick)
backend_AltActive_to_AltActive_ping_tick ==
    /\ backend_state = backend_AltActive
    /\ received_path_ping' = [type |-> MSG_path_ping]
    /\ backend_state' = backend_AltActive
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack>>

Cmds_backend_AltActive_to_AltActive_ping_tick == {CMD_send_path_ping, CMD_start_pong_timeout}

\* backend: AltActive -> AltDegraded (ping_timeout)
backend_AltActive_to_AltDegraded_ping_timeout ==
    /\ backend_state = backend_AltActive
    /\ backend_state' = backend_AltDegraded
    /\ ping_failures' = 1
    /\ UNCHANGED <<client_state, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

\* backend: AltDegraded -> AltDegraded (ping_tick)
backend_AltDegraded_to_AltDegraded_ping_tick ==
    /\ backend_state = backend_AltDegraded
    /\ received_path_ping' = [type |-> MSG_path_ping]
    /\ backend_state' = backend_AltDegraded
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack>>

Cmds_backend_AltDegraded_to_AltDegraded_ping_tick == {CMD_send_path_ping, CMD_start_pong_timeout}

\* backend: AltActive -> RelayBackoff (alt_stream_error)
backend_AltActive_to_RelayBackoff_alt_stream_error ==
    /\ backend_state = backend_AltActive
    /\ backend_state' = backend_RelayBackoff
    /\ backoff_level' = Min(backoff_level + 1, max_backoff_level)
    /\ b_active_path' = "relay"
    /\ b_dispatcher_path' = "relay"
    /\ monitor_target' = "none"
    /\ alt_signal' = "pending"
    /\ ping_failures' = 0
    /\ UNCHANGED <<client_state, c_active_path, c_dispatcher_path, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_AltActive_to_RelayBackoff_alt_stream_error == {CMD_stop_monitor, CMD_stop_alt_stream_reader, CMD_stop_alt_dg_reader, CMD_close_alt_path, CMD_reset_alt_ready, CMD_start_backoff_timer}

\* backend: AltDegraded -> RelayBackoff (alt_stream_error)
backend_AltDegraded_to_RelayBackoff_alt_stream_error ==
    /\ backend_state = backend_AltDegraded
    /\ backend_state' = backend_RelayBackoff
    /\ backoff_level' = Min(backoff_level + 1, max_backoff_level)
    /\ b_active_path' = "relay"
    /\ b_dispatcher_path' = "relay"
    /\ monitor_target' = "none"
    /\ alt_signal' = "pending"
    /\ ping_failures' = 0
    /\ UNCHANGED <<client_state, c_active_path, c_dispatcher_path, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_AltDegraded_to_RelayBackoff_alt_stream_error == {CMD_stop_monitor, CMD_stop_alt_stream_reader, CMD_stop_alt_dg_reader, CMD_close_alt_path, CMD_reset_alt_ready, CMD_start_backoff_timer}

\* backend: AltDegraded -> AltActive on recv path_pong
backend_AltDegraded_to_AltActive_on_path_pong ==
    /\ backend_state = backend_AltDegraded
    /\ received_path_pong.type = MSG_path_pong
    /\ received_path_pong' = [type |-> "none"]
    /\ backend_state' = backend_AltActive
    /\ ping_failures' = 0
    /\ UNCHANGED <<client_state, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_AltDegraded_to_AltActive_on_path_pong == {CMD_cancel_pong_timeout}

\* backend: AltDegraded -> AltDegraded (ping_timeout) [under_max_failures]
backend_AltDegraded_to_AltDegraded_ping_timeout_under_max_failures ==
    /\ backend_state = backend_AltDegraded
    /\ ping_failures + 1 < max_ping_failures
    /\ backend_state' = backend_AltDegraded
    /\ ping_failures' = ping_failures + 1
    /\ UNCHANGED <<client_state, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

\* backend: AltDegraded -> RelayBackoff (ping_timeout) [at_max_failures]
backend_AltDegraded_to_RelayBackoff_ping_timeout_at_max_failures ==
    /\ backend_state = backend_AltDegraded
    /\ ping_failures + 1 >= max_ping_failures
    /\ backend_state' = backend_RelayBackoff
    /\ backoff_level' = Min(backoff_level + 1, max_backoff_level)
    /\ b_active_path' = "relay"
    /\ b_dispatcher_path' = "relay"
    /\ monitor_target' = "none"
    /\ alt_signal' = "pending"
    /\ ping_failures' = 0
    /\ UNCHANGED <<client_state, c_active_path, c_dispatcher_path, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_AltDegraded_to_RelayBackoff_ping_timeout_at_max_failures == {CMD_stop_monitor, CMD_stop_alt_stream_reader, CMD_stop_alt_dg_reader, CMD_close_alt_path, CMD_reset_alt_ready, CMD_start_backoff_timer}

\* backend: RelayBackoff -> CandidatesAdvertised (backoff_expired)
backend_RelayBackoff_to_CandidatesAdvertised_backoff_expired ==
    /\ backend_state = backend_RelayBackoff
    /\ received_candidates' = [type |-> MSG_candidates, challenge |-> local_pair_check_challenge, set |-> local_candidates]
    /\ backend_state' = backend_CandidatesAdvertised
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_pair_check_ack, received_path_ping>>

Cmds_backend_RelayBackoff_to_CandidatesAdvertised_backoff_expired == {CMD_send_candidates}

\* backend: RelayBackoff -> CandidatesAdvertised (candidates_changed)
backend_RelayBackoff_to_CandidatesAdvertised_candidates_changed ==
    /\ backend_state = backend_RelayBackoff
    /\ received_candidates' = [type |-> MSG_candidates, challenge |-> local_pair_check_challenge, set |-> local_candidates]
    /\ backend_state' = backend_CandidatesAdvertised
    /\ backoff_level' = 0
    /\ UNCHANGED <<client_state, ping_failures, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_pair_check_ack, received_path_ping>>

Cmds_backend_RelayBackoff_to_CandidatesAdvertised_candidates_changed == {CMD_send_candidates}

\* backend: RelayConnected -> CandidatesAdvertised (candidates_refresh_tick) [local_candidates_available]
backend_RelayConnected_to_CandidatesAdvertised_candidates_refresh_tick_local_candidates_available ==
    /\ backend_state = backend_RelayConnected
    /\ local_listen_addr /= "none"
    /\ received_candidates' = [type |-> MSG_candidates, challenge |-> local_pair_check_challenge, set |-> local_candidates]
    /\ backend_state' = backend_CandidatesAdvertised
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_pair_check_ack, received_path_ping>>

Cmds_backend_RelayConnected_to_CandidatesAdvertised_candidates_refresh_tick_local_candidates_available == {CMD_send_candidates}

\* backend: CandidatesAdvertised -> RelayConnected (app_force_fallback)
backend_CandidatesAdvertised_to_RelayConnected_app_force_fallback ==
    /\ backend_state = backend_CandidatesAdvertised
    /\ backend_state' = backend_RelayConnected
    /\ alt_signal' = "pending"
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_CandidatesAdvertised_to_RelayConnected_app_force_fallback == {CMD_reset_alt_ready}

\* backend: AltActive -> RelayBackoff (app_force_fallback)
backend_AltActive_to_RelayBackoff_app_force_fallback ==
    /\ backend_state = backend_AltActive
    /\ backend_state' = backend_RelayBackoff
    /\ backoff_level' = Min(backoff_level + 1, max_backoff_level)
    /\ b_active_path' = "relay"
    /\ b_dispatcher_path' = "relay"
    /\ monitor_target' = "none"
    /\ alt_signal' = "pending"
    /\ ping_failures' = 0
    /\ UNCHANGED <<client_state, c_active_path, c_dispatcher_path, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_AltActive_to_RelayBackoff_app_force_fallback == {CMD_stop_monitor, CMD_cancel_pong_timeout, CMD_stop_alt_stream_reader, CMD_stop_alt_dg_reader, CMD_close_alt_path, CMD_reset_alt_ready, CMD_start_backoff_timer}

\* backend: AltDegraded -> RelayBackoff (app_force_fallback)
backend_AltDegraded_to_RelayBackoff_app_force_fallback ==
    /\ backend_state = backend_AltDegraded
    /\ backend_state' = backend_RelayBackoff
    /\ backoff_level' = Min(backoff_level + 1, max_backoff_level)
    /\ b_active_path' = "relay"
    /\ b_dispatcher_path' = "relay"
    /\ monitor_target' = "none"
    /\ alt_signal' = "pending"
    /\ ping_failures' = 0
    /\ UNCHANGED <<client_state, c_active_path, c_dispatcher_path, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_AltDegraded_to_RelayBackoff_app_force_fallback == {CMD_stop_monitor, CMD_cancel_pong_timeout, CMD_stop_alt_stream_reader, CMD_stop_alt_dg_reader, CMD_close_alt_path, CMD_reset_alt_ready, CMD_start_backoff_timer}

\* backend: RelayConnected -> RelayConnected (app_send)
backend_RelayConnected_to_RelayConnected_app_send ==
    /\ backend_state = backend_RelayConnected
    /\ backend_state' = backend_RelayConnected
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_RelayConnected_to_RelayConnected_app_send == {CMD_write_active_stream}

\* backend: CandidatesAdvertised -> CandidatesAdvertised (app_send)
backend_CandidatesAdvertised_to_CandidatesAdvertised_app_send ==
    /\ backend_state = backend_CandidatesAdvertised
    /\ backend_state' = backend_CandidatesAdvertised
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_CandidatesAdvertised_to_CandidatesAdvertised_app_send == {CMD_write_active_stream}

\* backend: AltActive -> AltActive (app_send)
backend_AltActive_to_AltActive_app_send ==
    /\ backend_state = backend_AltActive
    /\ backend_state' = backend_AltActive
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_AltActive_to_AltActive_app_send == {CMD_write_active_stream}

\* backend: AltDegraded -> AltDegraded (app_send)
backend_AltDegraded_to_AltDegraded_app_send ==
    /\ backend_state = backend_AltDegraded
    /\ backend_state' = backend_AltDegraded
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_AltDegraded_to_AltDegraded_app_send == {CMD_write_active_stream}

\* backend: RelayBackoff -> RelayBackoff (app_send)
backend_RelayBackoff_to_RelayBackoff_app_send ==
    /\ backend_state = backend_RelayBackoff
    /\ backend_state' = backend_RelayBackoff
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_RelayBackoff_to_RelayBackoff_app_send == {CMD_write_active_stream}

\* backend: RelayConnected -> RelayConnected (relay_stream_data)
backend_RelayConnected_to_RelayConnected_relay_stream_data ==
    /\ backend_state = backend_RelayConnected
    /\ backend_state' = backend_RelayConnected
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_RelayConnected_to_RelayConnected_relay_stream_data == {CMD_deliver_recv}

\* backend: CandidatesAdvertised -> CandidatesAdvertised (relay_stream_data)
backend_CandidatesAdvertised_to_CandidatesAdvertised_relay_stream_data ==
    /\ backend_state = backend_CandidatesAdvertised
    /\ backend_state' = backend_CandidatesAdvertised
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_CandidatesAdvertised_to_CandidatesAdvertised_relay_stream_data == {CMD_deliver_recv}

\* backend: AltActive -> AltActive (relay_stream_data)
backend_AltActive_to_AltActive_relay_stream_data ==
    /\ backend_state = backend_AltActive
    /\ backend_state' = backend_AltActive
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_AltActive_to_AltActive_relay_stream_data == {CMD_deliver_recv}

\* backend: AltDegraded -> AltDegraded (relay_stream_data)
backend_AltDegraded_to_AltDegraded_relay_stream_data ==
    /\ backend_state = backend_AltDegraded
    /\ backend_state' = backend_AltDegraded
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_AltDegraded_to_AltDegraded_relay_stream_data == {CMD_deliver_recv}

\* backend: RelayBackoff -> RelayBackoff (relay_stream_data)
backend_RelayBackoff_to_RelayBackoff_relay_stream_data ==
    /\ backend_state = backend_RelayBackoff
    /\ backend_state' = backend_RelayBackoff
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_RelayBackoff_to_RelayBackoff_relay_stream_data == {CMD_deliver_recv}

\* backend: RelayConnected -> RelayConnected (relay_stream_error)
backend_RelayConnected_to_RelayConnected_relay_stream_error ==
    /\ backend_state = backend_RelayConnected
    /\ backend_state' = backend_RelayConnected
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_RelayConnected_to_RelayConnected_relay_stream_error == {CMD_deliver_recv_error}

\* backend: CandidatesAdvertised -> CandidatesAdvertised (relay_stream_error)
backend_CandidatesAdvertised_to_CandidatesAdvertised_relay_stream_error ==
    /\ backend_state = backend_CandidatesAdvertised
    /\ backend_state' = backend_CandidatesAdvertised
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_CandidatesAdvertised_to_CandidatesAdvertised_relay_stream_error == {CMD_deliver_recv_error}

\* backend: AltActive -> AltActive (relay_stream_error)
backend_AltActive_to_AltActive_relay_stream_error ==
    /\ backend_state = backend_AltActive
    /\ backend_state' = backend_AltActive
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_AltActive_to_AltActive_relay_stream_error == {CMD_deliver_recv_error}

\* backend: AltDegraded -> AltDegraded (relay_stream_error)
backend_AltDegraded_to_AltDegraded_relay_stream_error ==
    /\ backend_state = backend_AltDegraded
    /\ backend_state' = backend_AltDegraded
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_AltDegraded_to_AltDegraded_relay_stream_error == {CMD_deliver_recv_error}

\* backend: RelayBackoff -> RelayBackoff (relay_stream_error)
backend_RelayBackoff_to_RelayBackoff_relay_stream_error ==
    /\ backend_state = backend_RelayBackoff
    /\ backend_state' = backend_RelayBackoff
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_RelayBackoff_to_RelayBackoff_relay_stream_error == {CMD_deliver_recv_error}

\* backend: RelayConnected -> RelayConnected (app_send_datagram)
backend_RelayConnected_to_RelayConnected_app_send_datagram ==
    /\ backend_state = backend_RelayConnected
    /\ backend_state' = backend_RelayConnected
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_RelayConnected_to_RelayConnected_app_send_datagram == {CMD_send_active_datagram}

\* backend: CandidatesAdvertised -> CandidatesAdvertised (app_send_datagram)
backend_CandidatesAdvertised_to_CandidatesAdvertised_app_send_datagram ==
    /\ backend_state = backend_CandidatesAdvertised
    /\ backend_state' = backend_CandidatesAdvertised
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_CandidatesAdvertised_to_CandidatesAdvertised_app_send_datagram == {CMD_send_active_datagram}

\* backend: AltActive -> AltActive (app_send_datagram)
backend_AltActive_to_AltActive_app_send_datagram ==
    /\ backend_state = backend_AltActive
    /\ backend_state' = backend_AltActive
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_AltActive_to_AltActive_app_send_datagram == {CMD_send_active_datagram}

\* backend: AltDegraded -> AltDegraded (app_send_datagram)
backend_AltDegraded_to_AltDegraded_app_send_datagram ==
    /\ backend_state = backend_AltDegraded
    /\ backend_state' = backend_AltDegraded
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_AltDegraded_to_AltDegraded_app_send_datagram == {CMD_send_active_datagram}

\* backend: RelayBackoff -> RelayBackoff (app_send_datagram)
backend_RelayBackoff_to_RelayBackoff_app_send_datagram ==
    /\ backend_state = backend_RelayBackoff
    /\ backend_state' = backend_RelayBackoff
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_RelayBackoff_to_RelayBackoff_app_send_datagram == {CMD_send_active_datagram}

\* backend: RelayConnected -> RelayConnected (relay_datagram)
backend_RelayConnected_to_RelayConnected_relay_datagram ==
    /\ backend_state = backend_RelayConnected
    /\ backend_state' = backend_RelayConnected
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_RelayConnected_to_RelayConnected_relay_datagram == {CMD_deliver_recv_datagram}

\* backend: CandidatesAdvertised -> CandidatesAdvertised (relay_datagram)
backend_CandidatesAdvertised_to_CandidatesAdvertised_relay_datagram ==
    /\ backend_state = backend_CandidatesAdvertised
    /\ backend_state' = backend_CandidatesAdvertised
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_CandidatesAdvertised_to_CandidatesAdvertised_relay_datagram == {CMD_deliver_recv_datagram}

\* backend: AltActive -> AltActive (relay_datagram)
backend_AltActive_to_AltActive_relay_datagram ==
    /\ backend_state = backend_AltActive
    /\ backend_state' = backend_AltActive
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_AltActive_to_AltActive_relay_datagram == {CMD_deliver_recv_datagram}

\* backend: AltDegraded -> AltDegraded (relay_datagram)
backend_AltDegraded_to_AltDegraded_relay_datagram ==
    /\ backend_state = backend_AltDegraded
    /\ backend_state' = backend_AltDegraded
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_AltDegraded_to_AltDegraded_relay_datagram == {CMD_deliver_recv_datagram}

\* backend: RelayBackoff -> RelayBackoff (relay_datagram)
backend_RelayBackoff_to_RelayBackoff_relay_datagram ==
    /\ backend_state = backend_RelayBackoff
    /\ backend_state' = backend_RelayBackoff
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_RelayBackoff_to_RelayBackoff_relay_datagram == {CMD_deliver_recv_datagram}

\* backend: AltActive -> AltActive (alt_stream_data)
backend_AltActive_to_AltActive_alt_stream_data ==
    /\ backend_state = backend_AltActive
    /\ backend_state' = backend_AltActive
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_AltActive_to_AltActive_alt_stream_data == {CMD_deliver_recv}

\* backend: AltDegraded -> AltDegraded (alt_stream_data)
backend_AltDegraded_to_AltDegraded_alt_stream_data ==
    /\ backend_state = backend_AltDegraded
    /\ backend_state' = backend_AltDegraded
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_AltDegraded_to_AltDegraded_alt_stream_data == {CMD_deliver_recv}

\* backend: AltActive -> AltActive (alt_datagram)
backend_AltActive_to_AltActive_alt_datagram ==
    /\ backend_state = backend_AltActive
    /\ backend_state' = backend_AltActive
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_AltActive_to_AltActive_alt_datagram == {CMD_deliver_recv_datagram}

\* backend: AltDegraded -> AltDegraded (alt_datagram)
backend_AltDegraded_to_AltDegraded_alt_datagram ==
    /\ backend_state = backend_AltDegraded
    /\ backend_state' = backend_AltDegraded
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_backend_AltDegraded_to_AltDegraded_alt_datagram == {CMD_deliver_recv_datagram}


\* client: RelayConnected -> PairDialing on recv candidates [alt_enabled]
client_RelayConnected_to_PairDialing_on_candidates_alt_enabled ==
    /\ client_state = client_RelayConnected
    /\ received_candidates.type = MSG_candidates
    /\ TRUE
    /\ received_candidates' = [type |-> "none"]
    /\ client_state' = client_PairDialing
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_pair_check_ack, received_path_ping>>

Cmds_client_RelayConnected_to_PairDialing_on_candidates_alt_enabled == {CMD_dial_candidate}

\* client: RelayConnected -> RelayConnected on recv candidates [alt_disabled]
client_RelayConnected_to_RelayConnected_on_candidates_alt_disabled ==
    /\ client_state = client_RelayConnected
    /\ received_candidates.type = MSG_candidates
    /\ FALSE
    /\ received_candidates' = [type |-> "none"]
    /\ client_state' = client_RelayConnected
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_pair_check_ack, received_path_ping>>

\* client: PairDialing -> PairChecking (dial_ok)
client_PairDialing_to_PairChecking_dial_ok ==
    /\ client_state = client_PairDialing
    /\ received_pair_check' = [type |-> MSG_pair_check, challenge |-> received_pair_check_challenge, instance_id |-> instance_id]
    /\ client_state' = client_PairChecking
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_PairDialing_to_PairChecking_dial_ok == {CMD_send_pair_check}

\* client: PairDialing -> RelayConnected (dial_failed)
client_PairDialing_to_RelayConnected_dial_failed ==
    /\ client_state = client_PairDialing
    /\ client_state' = client_RelayConnected
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

\* client: PairChecking -> AltActive on recv pair_check_ack
client_PairChecking_to_AltActive_on_pair_check_ack ==
    /\ client_state = client_PairChecking
    /\ received_pair_check_ack.type = MSG_pair_check_ack
    /\ received_pair_check_ack' = [type |-> "none"]
    /\ client_state' = client_AltActive
    /\ c_active_path' = "alt"
    /\ c_dispatcher_path' = "alt"
    /\ alt_signal' = "ready"
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, b_dispatcher_path, monitor_target, received_pair_check, received_path_pong, received_candidates, received_path_ping>>

Cmds_client_PairChecking_to_AltActive_on_pair_check_ack == {CMD_start_alt_stream_reader, CMD_start_alt_dg_reader, CMD_signal_alt_ready, CMD_set_crypto_datagram}

\* client: PairChecking -> RelayConnected (verify_timeout)
client_PairChecking_to_RelayConnected_verify_timeout ==
    /\ client_state = client_PairChecking
    /\ client_state' = client_RelayConnected
    /\ c_dispatcher_path' = "relay"
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

\* client: AltActive -> AltActive on recv path_ping
client_AltActive_to_AltActive_on_path_ping ==
    /\ client_state = client_AltActive
    /\ received_path_ping.type = MSG_path_ping
    /\ received_path_ping' = [type |-> "none"]
    /\ received_path_pong' = [type |-> MSG_path_pong]
    /\ client_state' = client_AltActive
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_candidates, received_pair_check_ack>>

Cmds_client_AltActive_to_AltActive_on_path_ping == {CMD_send_path_pong}

\* client: AltActive -> RelayFallback (alt_error)
client_AltActive_to_RelayFallback_alt_error ==
    /\ client_state = client_AltActive
    /\ client_state' = client_RelayFallback
    /\ c_active_path' = "relay"
    /\ c_dispatcher_path' = "relay"
    /\ alt_signal' = "pending"
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, b_dispatcher_path, monitor_target, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_AltActive_to_RelayFallback_alt_error == {CMD_stop_alt_stream_reader, CMD_stop_alt_dg_reader, CMD_close_alt_path, CMD_reset_alt_ready}

\* client: AltActive -> RelayFallback (alt_stream_error)
client_AltActive_to_RelayFallback_alt_stream_error ==
    /\ client_state = client_AltActive
    /\ client_state' = client_RelayFallback
    /\ c_active_path' = "relay"
    /\ c_dispatcher_path' = "relay"
    /\ alt_signal' = "pending"
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, b_dispatcher_path, monitor_target, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_AltActive_to_RelayFallback_alt_stream_error == {CMD_stop_alt_stream_reader, CMD_stop_alt_dg_reader, CMD_close_alt_path, CMD_reset_alt_ready}

\* client: RelayFallback -> RelayConnected (relay_ok)
client_RelayFallback_to_RelayConnected_relay_ok ==
    /\ client_state = client_RelayFallback
    /\ client_state' = client_RelayConnected
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

\* client: AltActive -> PairDialing on recv candidates [alt_enabled]
client_AltActive_to_PairDialing_on_candidates_alt_enabled ==
    /\ client_state = client_AltActive
    /\ received_candidates.type = MSG_candidates
    /\ TRUE
    /\ received_candidates' = [type |-> "none"]
    /\ client_state' = client_PairDialing
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_pair_check_ack, received_path_ping>>

Cmds_client_AltActive_to_PairDialing_on_candidates_alt_enabled == {CMD_stop_alt_stream_reader, CMD_stop_alt_dg_reader, CMD_close_alt_path, CMD_dial_candidate}

\* client: PairDialing -> RelayConnected (app_force_fallback)
client_PairDialing_to_RelayConnected_app_force_fallback ==
    /\ client_state = client_PairDialing
    /\ client_state' = client_RelayConnected
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

\* client: PairChecking -> RelayConnected (app_force_fallback)
client_PairChecking_to_RelayConnected_app_force_fallback ==
    /\ client_state = client_PairChecking
    /\ client_state' = client_RelayConnected
    /\ c_dispatcher_path' = "relay"
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_PairChecking_to_RelayConnected_app_force_fallback == {CMD_stop_alt_stream_reader, CMD_stop_alt_dg_reader, CMD_close_alt_path}

\* client: AltActive -> RelayConnected (app_force_fallback)
client_AltActive_to_RelayConnected_app_force_fallback ==
    /\ client_state = client_AltActive
    /\ client_state' = client_RelayConnected
    /\ c_active_path' = "relay"
    /\ c_dispatcher_path' = "relay"
    /\ alt_signal' = "pending"
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, b_dispatcher_path, monitor_target, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_AltActive_to_RelayConnected_app_force_fallback == {CMD_stop_alt_stream_reader, CMD_stop_alt_dg_reader, CMD_close_alt_path, CMD_reset_alt_ready}

\* client: RelayConnected -> RelayConnected (app_send)
client_RelayConnected_to_RelayConnected_app_send ==
    /\ client_state = client_RelayConnected
    /\ client_state' = client_RelayConnected
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_RelayConnected_to_RelayConnected_app_send == {CMD_write_active_stream}

\* client: PairDialing -> PairDialing (app_send)
client_PairDialing_to_PairDialing_app_send ==
    /\ client_state = client_PairDialing
    /\ client_state' = client_PairDialing
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_PairDialing_to_PairDialing_app_send == {CMD_write_active_stream}

\* client: PairChecking -> PairChecking (app_send)
client_PairChecking_to_PairChecking_app_send ==
    /\ client_state = client_PairChecking
    /\ client_state' = client_PairChecking
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_PairChecking_to_PairChecking_app_send == {CMD_write_active_stream}

\* client: AltActive -> AltActive (app_send)
client_AltActive_to_AltActive_app_send ==
    /\ client_state = client_AltActive
    /\ client_state' = client_AltActive
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_AltActive_to_AltActive_app_send == {CMD_write_active_stream}

\* client: RelayFallback -> RelayFallback (app_send)
client_RelayFallback_to_RelayFallback_app_send ==
    /\ client_state = client_RelayFallback
    /\ client_state' = client_RelayFallback
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_RelayFallback_to_RelayFallback_app_send == {CMD_write_active_stream}

\* client: RelayConnected -> RelayConnected (relay_stream_data)
client_RelayConnected_to_RelayConnected_relay_stream_data ==
    /\ client_state = client_RelayConnected
    /\ client_state' = client_RelayConnected
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_RelayConnected_to_RelayConnected_relay_stream_data == {CMD_deliver_recv}

\* client: PairDialing -> PairDialing (relay_stream_data)
client_PairDialing_to_PairDialing_relay_stream_data ==
    /\ client_state = client_PairDialing
    /\ client_state' = client_PairDialing
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_PairDialing_to_PairDialing_relay_stream_data == {CMD_deliver_recv}

\* client: PairChecking -> PairChecking (relay_stream_data)
client_PairChecking_to_PairChecking_relay_stream_data ==
    /\ client_state = client_PairChecking
    /\ client_state' = client_PairChecking
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_PairChecking_to_PairChecking_relay_stream_data == {CMD_deliver_recv}

\* client: AltActive -> AltActive (relay_stream_data)
client_AltActive_to_AltActive_relay_stream_data ==
    /\ client_state = client_AltActive
    /\ client_state' = client_AltActive
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_AltActive_to_AltActive_relay_stream_data == {CMD_deliver_recv}

\* client: RelayFallback -> RelayFallback (relay_stream_data)
client_RelayFallback_to_RelayFallback_relay_stream_data ==
    /\ client_state = client_RelayFallback
    /\ client_state' = client_RelayFallback
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_RelayFallback_to_RelayFallback_relay_stream_data == {CMD_deliver_recv}

\* client: RelayConnected -> RelayConnected (relay_stream_error)
client_RelayConnected_to_RelayConnected_relay_stream_error ==
    /\ client_state = client_RelayConnected
    /\ client_state' = client_RelayConnected
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_RelayConnected_to_RelayConnected_relay_stream_error == {CMD_deliver_recv_error}

\* client: PairDialing -> PairDialing (relay_stream_error)
client_PairDialing_to_PairDialing_relay_stream_error ==
    /\ client_state = client_PairDialing
    /\ client_state' = client_PairDialing
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_PairDialing_to_PairDialing_relay_stream_error == {CMD_deliver_recv_error}

\* client: PairChecking -> PairChecking (relay_stream_error)
client_PairChecking_to_PairChecking_relay_stream_error ==
    /\ client_state = client_PairChecking
    /\ client_state' = client_PairChecking
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_PairChecking_to_PairChecking_relay_stream_error == {CMD_deliver_recv_error}

\* client: AltActive -> AltActive (relay_stream_error)
client_AltActive_to_AltActive_relay_stream_error ==
    /\ client_state = client_AltActive
    /\ client_state' = client_AltActive
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_AltActive_to_AltActive_relay_stream_error == {CMD_deliver_recv_error}

\* client: RelayFallback -> RelayFallback (relay_stream_error)
client_RelayFallback_to_RelayFallback_relay_stream_error ==
    /\ client_state = client_RelayFallback
    /\ client_state' = client_RelayFallback
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_RelayFallback_to_RelayFallback_relay_stream_error == {CMD_deliver_recv_error}

\* client: RelayConnected -> RelayConnected (app_send_datagram)
client_RelayConnected_to_RelayConnected_app_send_datagram ==
    /\ client_state = client_RelayConnected
    /\ client_state' = client_RelayConnected
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_RelayConnected_to_RelayConnected_app_send_datagram == {CMD_send_active_datagram}

\* client: PairDialing -> PairDialing (app_send_datagram)
client_PairDialing_to_PairDialing_app_send_datagram ==
    /\ client_state = client_PairDialing
    /\ client_state' = client_PairDialing
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_PairDialing_to_PairDialing_app_send_datagram == {CMD_send_active_datagram}

\* client: PairChecking -> PairChecking (app_send_datagram)
client_PairChecking_to_PairChecking_app_send_datagram ==
    /\ client_state = client_PairChecking
    /\ client_state' = client_PairChecking
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_PairChecking_to_PairChecking_app_send_datagram == {CMD_send_active_datagram}

\* client: AltActive -> AltActive (app_send_datagram)
client_AltActive_to_AltActive_app_send_datagram ==
    /\ client_state = client_AltActive
    /\ client_state' = client_AltActive
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_AltActive_to_AltActive_app_send_datagram == {CMD_send_active_datagram}

\* client: RelayFallback -> RelayFallback (app_send_datagram)
client_RelayFallback_to_RelayFallback_app_send_datagram ==
    /\ client_state = client_RelayFallback
    /\ client_state' = client_RelayFallback
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_RelayFallback_to_RelayFallback_app_send_datagram == {CMD_send_active_datagram}

\* client: RelayConnected -> RelayConnected (relay_datagram)
client_RelayConnected_to_RelayConnected_relay_datagram ==
    /\ client_state = client_RelayConnected
    /\ client_state' = client_RelayConnected
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_RelayConnected_to_RelayConnected_relay_datagram == {CMD_deliver_recv_datagram}

\* client: PairDialing -> PairDialing (relay_datagram)
client_PairDialing_to_PairDialing_relay_datagram ==
    /\ client_state = client_PairDialing
    /\ client_state' = client_PairDialing
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_PairDialing_to_PairDialing_relay_datagram == {CMD_deliver_recv_datagram}

\* client: PairChecking -> PairChecking (relay_datagram)
client_PairChecking_to_PairChecking_relay_datagram ==
    /\ client_state = client_PairChecking
    /\ client_state' = client_PairChecking
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_PairChecking_to_PairChecking_relay_datagram == {CMD_deliver_recv_datagram}

\* client: AltActive -> AltActive (relay_datagram)
client_AltActive_to_AltActive_relay_datagram ==
    /\ client_state = client_AltActive
    /\ client_state' = client_AltActive
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_AltActive_to_AltActive_relay_datagram == {CMD_deliver_recv_datagram}

\* client: RelayFallback -> RelayFallback (relay_datagram)
client_RelayFallback_to_RelayFallback_relay_datagram ==
    /\ client_state = client_RelayFallback
    /\ client_state' = client_RelayFallback
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_RelayFallback_to_RelayFallback_relay_datagram == {CMD_deliver_recv_datagram}

\* client: AltActive -> AltActive (alt_stream_data)
client_AltActive_to_AltActive_alt_stream_data ==
    /\ client_state = client_AltActive
    /\ client_state' = client_AltActive
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_AltActive_to_AltActive_alt_stream_data == {CMD_deliver_recv}

\* client: AltActive -> AltActive (alt_datagram)
client_AltActive_to_AltActive_alt_datagram ==
    /\ client_state = client_AltActive
    /\ client_state' = client_AltActive
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, alt_signal, received_pair_check, received_path_pong, received_candidates, received_pair_check_ack, received_path_ping>>

Cmds_client_AltActive_to_AltActive_alt_datagram == {CMD_deliver_recv_datagram}


Next ==
    \/ backend_RelayConnected_to_CandidatesAdvertised_candidates_gathered
    \/ backend_CandidatesAdvertised_to_AltActive_on_pair_check_challenge_valid
    \/ backend_CandidatesAdvertised_to_RelayConnected_on_pair_check_challenge_invalid
    \/ backend_CandidatesAdvertised_to_RelayBackoff_candidates_timeout
    \/ backend_AltActive_to_AltActive_ping_tick
    \/ backend_AltActive_to_AltDegraded_ping_timeout
    \/ backend_AltDegraded_to_AltDegraded_ping_tick
    \/ backend_AltActive_to_RelayBackoff_alt_stream_error
    \/ backend_AltDegraded_to_RelayBackoff_alt_stream_error
    \/ backend_AltDegraded_to_AltActive_on_path_pong
    \/ backend_AltDegraded_to_AltDegraded_ping_timeout_under_max_failures
    \/ backend_AltDegraded_to_RelayBackoff_ping_timeout_at_max_failures
    \/ backend_RelayBackoff_to_CandidatesAdvertised_backoff_expired
    \/ backend_RelayBackoff_to_CandidatesAdvertised_candidates_changed
    \/ backend_RelayConnected_to_CandidatesAdvertised_candidates_refresh_tick_local_candidates_available
    \/ backend_CandidatesAdvertised_to_RelayConnected_app_force_fallback
    \/ backend_AltActive_to_RelayBackoff_app_force_fallback
    \/ backend_AltDegraded_to_RelayBackoff_app_force_fallback
    \/ backend_RelayConnected_to_RelayConnected_app_send
    \/ backend_CandidatesAdvertised_to_CandidatesAdvertised_app_send
    \/ backend_AltActive_to_AltActive_app_send
    \/ backend_AltDegraded_to_AltDegraded_app_send
    \/ backend_RelayBackoff_to_RelayBackoff_app_send
    \/ backend_RelayConnected_to_RelayConnected_relay_stream_data
    \/ backend_CandidatesAdvertised_to_CandidatesAdvertised_relay_stream_data
    \/ backend_AltActive_to_AltActive_relay_stream_data
    \/ backend_AltDegraded_to_AltDegraded_relay_stream_data
    \/ backend_RelayBackoff_to_RelayBackoff_relay_stream_data
    \/ backend_RelayConnected_to_RelayConnected_relay_stream_error
    \/ backend_CandidatesAdvertised_to_CandidatesAdvertised_relay_stream_error
    \/ backend_AltActive_to_AltActive_relay_stream_error
    \/ backend_AltDegraded_to_AltDegraded_relay_stream_error
    \/ backend_RelayBackoff_to_RelayBackoff_relay_stream_error
    \/ backend_RelayConnected_to_RelayConnected_app_send_datagram
    \/ backend_CandidatesAdvertised_to_CandidatesAdvertised_app_send_datagram
    \/ backend_AltActive_to_AltActive_app_send_datagram
    \/ backend_AltDegraded_to_AltDegraded_app_send_datagram
    \/ backend_RelayBackoff_to_RelayBackoff_app_send_datagram
    \/ backend_RelayConnected_to_RelayConnected_relay_datagram
    \/ backend_CandidatesAdvertised_to_CandidatesAdvertised_relay_datagram
    \/ backend_AltActive_to_AltActive_relay_datagram
    \/ backend_AltDegraded_to_AltDegraded_relay_datagram
    \/ backend_RelayBackoff_to_RelayBackoff_relay_datagram
    \/ backend_AltActive_to_AltActive_alt_stream_data
    \/ backend_AltDegraded_to_AltDegraded_alt_stream_data
    \/ backend_AltActive_to_AltActive_alt_datagram
    \/ backend_AltDegraded_to_AltDegraded_alt_datagram
    \/ client_RelayConnected_to_PairDialing_on_candidates_alt_enabled
    \/ client_RelayConnected_to_RelayConnected_on_candidates_alt_disabled
    \/ client_PairDialing_to_PairChecking_dial_ok
    \/ client_PairDialing_to_RelayConnected_dial_failed
    \/ client_PairChecking_to_AltActive_on_pair_check_ack
    \/ client_PairChecking_to_RelayConnected_verify_timeout
    \/ client_AltActive_to_AltActive_on_path_ping
    \/ client_AltActive_to_RelayFallback_alt_error
    \/ client_AltActive_to_RelayFallback_alt_stream_error
    \/ client_RelayFallback_to_RelayConnected_relay_ok
    \/ client_AltActive_to_PairDialing_on_candidates_alt_enabled
    \/ client_PairDialing_to_RelayConnected_app_force_fallback
    \/ client_PairChecking_to_RelayConnected_app_force_fallback
    \/ client_AltActive_to_RelayConnected_app_force_fallback
    \/ client_RelayConnected_to_RelayConnected_app_send
    \/ client_PairDialing_to_PairDialing_app_send
    \/ client_PairChecking_to_PairChecking_app_send
    \/ client_AltActive_to_AltActive_app_send
    \/ client_RelayFallback_to_RelayFallback_app_send
    \/ client_RelayConnected_to_RelayConnected_relay_stream_data
    \/ client_PairDialing_to_PairDialing_relay_stream_data
    \/ client_PairChecking_to_PairChecking_relay_stream_data
    \/ client_AltActive_to_AltActive_relay_stream_data
    \/ client_RelayFallback_to_RelayFallback_relay_stream_data
    \/ client_RelayConnected_to_RelayConnected_relay_stream_error
    \/ client_PairDialing_to_PairDialing_relay_stream_error
    \/ client_PairChecking_to_PairChecking_relay_stream_error
    \/ client_AltActive_to_AltActive_relay_stream_error
    \/ client_RelayFallback_to_RelayFallback_relay_stream_error
    \/ client_RelayConnected_to_RelayConnected_app_send_datagram
    \/ client_PairDialing_to_PairDialing_app_send_datagram
    \/ client_PairChecking_to_PairChecking_app_send_datagram
    \/ client_AltActive_to_AltActive_app_send_datagram
    \/ client_RelayFallback_to_RelayFallback_app_send_datagram
    \/ client_RelayConnected_to_RelayConnected_relay_datagram
    \/ client_PairDialing_to_PairDialing_relay_datagram
    \/ client_PairChecking_to_PairChecking_relay_datagram
    \/ client_AltActive_to_AltActive_relay_datagram
    \/ client_RelayFallback_to_RelayFallback_relay_datagram
    \/ client_AltActive_to_AltActive_alt_stream_data
    \/ client_AltActive_to_AltActive_alt_datagram

Spec == Init /\ [][Next]_vars /\ WF_vars(Next)

\* ================================================================
\* Invariants and properties
\* ================================================================

\* Paths are always valid
PathConsistency == b_active_path \in {"relay", "alt"} /\ c_active_path \in {"relay", "alt"}
\* Backoff never exceeds cap
BackoffBounded == backoff_level <= max_backoff_level
\* alt-pair success resets backoff
BackoffResetsOnSuccess == backend_state = backend_AltActive => backoff_level = 0
\* Dispatchers always bound to valid path
DispatcherAlwaysBound == b_dispatcher_path \in {"relay", "alt"} /\ c_dispatcher_path \in {"relay", "alt"}
\* Backend dispatcher on alt-pair when alt-pair active
BackendDispatcherMatchesActive == backend_state = backend_AltActive => b_dispatcher_path = "alt"
\* Client dispatcher on alt-pair when alt-pair active
ClientDispatcherMatchesActive == client_state = client_AltActive => c_dispatcher_path = "alt"
\* Monitor only pings when alt-pair is active or degraded
MonitorOnlyWhenAlt == monitor_target = "alt" => backend_state \in {backend_AltActive, backend_AltDegraded}
\* After fallback, backend eventually re-advertises candidates
FallbackLeadsToReadvertise == (backend_state = backend_RelayBackoff) ~> (backend_state = backend_CandidatesAdvertised)
\* Degraded state eventually resolves (recovery or fallback)
DegradedLeadsToResolutionOrFallback == (backend_state = backend_AltDegraded) ~> (backend_state \in {backend_AltActive, backend_RelayBackoff})

\* ================================================================
\* Command-consistency: state after transition matches emitted commands
\* These are verified by construction (the same YAML defines both
\* the variable updates and the command list), but documenting
\* them as TLA+ operators makes the relationship explicit.
\* ================================================================

\* backend_RelayConnected_to_CandidatesAdvertised_candidates_gathered emits: CMD_send_candidates
\* backend_CandidatesAdvertised_to_AltActive_on_pair_check_challenge_valid emits: CMD_send_pair_check_ack, CMD_start_alt_stream_reader, CMD_start_alt_dg_reader, CMD_start_monitor, CMD_signal_alt_ready, CMD_set_crypto_datagram
\* backend_CandidatesAdvertised_to_RelayBackoff_candidates_timeout emits: CMD_reset_alt_ready, CMD_start_backoff_timer
\* backend_AltActive_to_AltActive_ping_tick emits: CMD_send_path_ping, CMD_start_pong_timeout
\* backend_AltDegraded_to_AltDegraded_ping_tick emits: CMD_send_path_ping, CMD_start_pong_timeout
\* backend_AltActive_to_RelayBackoff_alt_stream_error emits: CMD_stop_monitor, CMD_stop_alt_stream_reader, CMD_stop_alt_dg_reader, CMD_close_alt_path, CMD_reset_alt_ready, CMD_start_backoff_timer
\* backend_AltDegraded_to_RelayBackoff_alt_stream_error emits: CMD_stop_monitor, CMD_stop_alt_stream_reader, CMD_stop_alt_dg_reader, CMD_close_alt_path, CMD_reset_alt_ready, CMD_start_backoff_timer
\* backend_AltDegraded_to_AltActive_on_path_pong emits: CMD_cancel_pong_timeout
\* backend_AltDegraded_to_RelayBackoff_ping_timeout_at_max_failures emits: CMD_stop_monitor, CMD_stop_alt_stream_reader, CMD_stop_alt_dg_reader, CMD_close_alt_path, CMD_reset_alt_ready, CMD_start_backoff_timer
\* backend_RelayBackoff_to_CandidatesAdvertised_backoff_expired emits: CMD_send_candidates
\* backend_RelayBackoff_to_CandidatesAdvertised_candidates_changed emits: CMD_send_candidates
\* backend_RelayConnected_to_CandidatesAdvertised_candidates_refresh_tick_local_candidates_available emits: CMD_send_candidates
\* backend_CandidatesAdvertised_to_RelayConnected_app_force_fallback emits: CMD_reset_alt_ready
\* backend_AltActive_to_RelayBackoff_app_force_fallback emits: CMD_stop_monitor, CMD_cancel_pong_timeout, CMD_stop_alt_stream_reader, CMD_stop_alt_dg_reader, CMD_close_alt_path, CMD_reset_alt_ready, CMD_start_backoff_timer
\* backend_AltDegraded_to_RelayBackoff_app_force_fallback emits: CMD_stop_monitor, CMD_cancel_pong_timeout, CMD_stop_alt_stream_reader, CMD_stop_alt_dg_reader, CMD_close_alt_path, CMD_reset_alt_ready, CMD_start_backoff_timer
\* backend_RelayConnected_to_RelayConnected_app_send emits: CMD_write_active_stream
\* backend_CandidatesAdvertised_to_CandidatesAdvertised_app_send emits: CMD_write_active_stream
\* backend_AltActive_to_AltActive_app_send emits: CMD_write_active_stream
\* backend_AltDegraded_to_AltDegraded_app_send emits: CMD_write_active_stream
\* backend_RelayBackoff_to_RelayBackoff_app_send emits: CMD_write_active_stream
\* backend_RelayConnected_to_RelayConnected_relay_stream_data emits: CMD_deliver_recv
\* backend_CandidatesAdvertised_to_CandidatesAdvertised_relay_stream_data emits: CMD_deliver_recv
\* backend_AltActive_to_AltActive_relay_stream_data emits: CMD_deliver_recv
\* backend_AltDegraded_to_AltDegraded_relay_stream_data emits: CMD_deliver_recv
\* backend_RelayBackoff_to_RelayBackoff_relay_stream_data emits: CMD_deliver_recv
\* backend_RelayConnected_to_RelayConnected_relay_stream_error emits: CMD_deliver_recv_error
\* backend_CandidatesAdvertised_to_CandidatesAdvertised_relay_stream_error emits: CMD_deliver_recv_error
\* backend_AltActive_to_AltActive_relay_stream_error emits: CMD_deliver_recv_error
\* backend_AltDegraded_to_AltDegraded_relay_stream_error emits: CMD_deliver_recv_error
\* backend_RelayBackoff_to_RelayBackoff_relay_stream_error emits: CMD_deliver_recv_error
\* backend_RelayConnected_to_RelayConnected_app_send_datagram emits: CMD_send_active_datagram
\* backend_CandidatesAdvertised_to_CandidatesAdvertised_app_send_datagram emits: CMD_send_active_datagram
\* backend_AltActive_to_AltActive_app_send_datagram emits: CMD_send_active_datagram
\* backend_AltDegraded_to_AltDegraded_app_send_datagram emits: CMD_send_active_datagram
\* backend_RelayBackoff_to_RelayBackoff_app_send_datagram emits: CMD_send_active_datagram
\* backend_RelayConnected_to_RelayConnected_relay_datagram emits: CMD_deliver_recv_datagram
\* backend_CandidatesAdvertised_to_CandidatesAdvertised_relay_datagram emits: CMD_deliver_recv_datagram
\* backend_AltActive_to_AltActive_relay_datagram emits: CMD_deliver_recv_datagram
\* backend_AltDegraded_to_AltDegraded_relay_datagram emits: CMD_deliver_recv_datagram
\* backend_RelayBackoff_to_RelayBackoff_relay_datagram emits: CMD_deliver_recv_datagram
\* backend_AltActive_to_AltActive_alt_stream_data emits: CMD_deliver_recv
\* backend_AltDegraded_to_AltDegraded_alt_stream_data emits: CMD_deliver_recv
\* backend_AltActive_to_AltActive_alt_datagram emits: CMD_deliver_recv_datagram
\* backend_AltDegraded_to_AltDegraded_alt_datagram emits: CMD_deliver_recv_datagram
\* client_RelayConnected_to_PairDialing_on_candidates_alt_enabled emits: CMD_dial_candidate
\* client_PairDialing_to_PairChecking_dial_ok emits: CMD_send_pair_check
\* client_PairChecking_to_AltActive_on_pair_check_ack emits: CMD_start_alt_stream_reader, CMD_start_alt_dg_reader, CMD_signal_alt_ready, CMD_set_crypto_datagram
\* client_AltActive_to_AltActive_on_path_ping emits: CMD_send_path_pong
\* client_AltActive_to_RelayFallback_alt_error emits: CMD_stop_alt_stream_reader, CMD_stop_alt_dg_reader, CMD_close_alt_path, CMD_reset_alt_ready
\* client_AltActive_to_RelayFallback_alt_stream_error emits: CMD_stop_alt_stream_reader, CMD_stop_alt_dg_reader, CMD_close_alt_path, CMD_reset_alt_ready
\* client_AltActive_to_PairDialing_on_candidates_alt_enabled emits: CMD_stop_alt_stream_reader, CMD_stop_alt_dg_reader, CMD_close_alt_path, CMD_dial_candidate
\* client_PairChecking_to_RelayConnected_app_force_fallback emits: CMD_stop_alt_stream_reader, CMD_stop_alt_dg_reader, CMD_close_alt_path
\* client_AltActive_to_RelayConnected_app_force_fallback emits: CMD_stop_alt_stream_reader, CMD_stop_alt_dg_reader, CMD_close_alt_path, CMD_reset_alt_ready
\* client_RelayConnected_to_RelayConnected_app_send emits: CMD_write_active_stream
\* client_PairDialing_to_PairDialing_app_send emits: CMD_write_active_stream
\* client_PairChecking_to_PairChecking_app_send emits: CMD_write_active_stream
\* client_AltActive_to_AltActive_app_send emits: CMD_write_active_stream
\* client_RelayFallback_to_RelayFallback_app_send emits: CMD_write_active_stream
\* client_RelayConnected_to_RelayConnected_relay_stream_data emits: CMD_deliver_recv
\* client_PairDialing_to_PairDialing_relay_stream_data emits: CMD_deliver_recv
\* client_PairChecking_to_PairChecking_relay_stream_data emits: CMD_deliver_recv
\* client_AltActive_to_AltActive_relay_stream_data emits: CMD_deliver_recv
\* client_RelayFallback_to_RelayFallback_relay_stream_data emits: CMD_deliver_recv
\* client_RelayConnected_to_RelayConnected_relay_stream_error emits: CMD_deliver_recv_error
\* client_PairDialing_to_PairDialing_relay_stream_error emits: CMD_deliver_recv_error
\* client_PairChecking_to_PairChecking_relay_stream_error emits: CMD_deliver_recv_error
\* client_AltActive_to_AltActive_relay_stream_error emits: CMD_deliver_recv_error
\* client_RelayFallback_to_RelayFallback_relay_stream_error emits: CMD_deliver_recv_error
\* client_RelayConnected_to_RelayConnected_app_send_datagram emits: CMD_send_active_datagram
\* client_PairDialing_to_PairDialing_app_send_datagram emits: CMD_send_active_datagram
\* client_PairChecking_to_PairChecking_app_send_datagram emits: CMD_send_active_datagram
\* client_AltActive_to_AltActive_app_send_datagram emits: CMD_send_active_datagram
\* client_RelayFallback_to_RelayFallback_app_send_datagram emits: CMD_send_active_datagram
\* client_RelayConnected_to_RelayConnected_relay_datagram emits: CMD_deliver_recv_datagram
\* client_PairDialing_to_PairDialing_relay_datagram emits: CMD_deliver_recv_datagram
\* client_PairChecking_to_PairChecking_relay_datagram emits: CMD_deliver_recv_datagram
\* client_AltActive_to_AltActive_relay_datagram emits: CMD_deliver_recv_datagram
\* client_RelayFallback_to_RelayFallback_relay_datagram emits: CMD_deliver_recv_datagram
\* client_AltActive_to_AltActive_alt_stream_data emits: CMD_deliver_recv
\* client_AltActive_to_AltActive_alt_datagram emits: CMD_deliver_recv_datagram

====
