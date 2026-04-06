---- MODULE Session_Transport ----
\* Auto-generated from protocol YAML. Do not edit.
\* Phase: Transport

EXTENDS Integers, Sequences, FiniteSets, TLC

\* States for backend
backend_Paired == "backend_Paired"
backend_SessionActive == "backend_SessionActive"
backend_RelayConnected == "backend_RelayConnected"
backend_LANOffered == "backend_LANOffered"
backend_LANActive == "backend_LANActive"
backend_RelayBackoff == "backend_RelayBackoff"
backend_LANDegraded == "backend_LANDegraded"

\* States for client
client_Paired == "client_Paired"
client_SessionActive == "client_SessionActive"
client_RelayConnected == "client_RelayConnected"
client_LANConnecting == "client_LANConnecting"
client_LANVerifying == "client_LANVerifying"
client_LANActive == "client_LANActive"
client_RelayFallback == "client_RelayFallback"

\* Message types
MSG_lan_offer == "lan_offer"
MSG_lan_verify == "lan_verify"
MSG_lan_confirm == "lan_confirm"
MSG_path_ping == "path_ping"
MSG_path_pong == "path_pong"

\* Event types
EVT_app_force_fallback == "app_force_fallback"
EVT_app_send == "app_send"
EVT_app_send_datagram == "app_send_datagram"
EVT_backoff_expired == "backoff_expired"
EVT_lan_datagram == "lan_datagram"
EVT_lan_dial_failed == "lan_dial_failed"
EVT_lan_dial_ok == "lan_dial_ok"
EVT_lan_error == "lan_error"
EVT_lan_server_changed == "lan_server_changed"
EVT_lan_server_ready == "lan_server_ready"
EVT_lan_stream_data == "lan_stream_data"
EVT_lan_stream_error == "lan_stream_error"
EVT_offer_timeout == "offer_timeout"
EVT_ping_tick == "ping_tick"
EVT_ping_timeout == "ping_timeout"
EVT_readvertise_tick == "readvertise_tick"
EVT_recv_lan_confirm == "recv_lan_confirm"
EVT_recv_lan_offer == "recv_lan_offer"
EVT_recv_lan_verify == "recv_lan_verify"
EVT_recv_path_ping == "recv_path_ping"
EVT_recv_path_pong == "recv_path_pong"
EVT_relay_datagram == "relay_datagram"
EVT_relay_ok == "relay_ok"
EVT_relay_stream_data == "relay_stream_data"
EVT_relay_stream_error == "relay_stream_error"
EVT_verify_timeout == "verify_timeout"

\* Command types
CMD_cancel_pong_timeout == "cancel_pong_timeout"
CMD_close_lan_path == "close_lan_path"
CMD_deliver_recv == "deliver_recv"
CMD_deliver_recv_datagram == "deliver_recv_datagram"
CMD_deliver_recv_error == "deliver_recv_error"
CMD_dial_lan == "dial_lan"
CMD_reset_lan_ready == "reset_lan_ready"
CMD_send_active_datagram == "send_active_datagram"
CMD_send_lan_confirm == "send_lan_confirm"
CMD_send_lan_offer == "send_lan_offer"
CMD_send_lan_verify == "send_lan_verify"
CMD_send_path_ping == "send_path_ping"
CMD_send_path_pong == "send_path_pong"
CMD_set_crypto_datagram == "set_crypto_datagram"
CMD_signal_lan_ready == "signal_lan_ready"
CMD_start_backoff_timer == "start_backoff_timer"
CMD_start_lan_dg_reader == "start_lan_dg_reader"
CMD_start_lan_stream_reader == "start_lan_stream_reader"
CMD_start_monitor == "start_monitor"
CMD_start_pong_timeout == "start_pong_timeout"
CMD_stop_lan_dg_reader == "stop_lan_dg_reader"
CMD_stop_lan_stream_reader == "stop_lan_stream_reader"
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



CONSTANTS lan_addr, challenge_bytes, offer_challenge, instance_id, max_ping_failures, max_backoff_level, lan_server_addr

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
    lan_signal,
    received_lan_verify,
    received_path_pong,
    received_lan_offer,
    received_lan_confirm,
    received_path_ping

vars == <<backend_state, client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

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
    /\ lan_signal = "pending"
    /\ received_lan_verify = [type |-> "none"]
    /\ received_path_pong = [type |-> "none"]
    /\ received_lan_offer = [type |-> "none"]
    /\ received_lan_confirm = [type |-> "none"]
    /\ received_path_ping = [type |-> "none"]

\* backend: RelayConnected -> LANOffered (lan_server_ready)
backend_RelayConnected_to_LANOffered_lan_server_ready ==
    /\ backend_state = backend_RelayConnected
    /\ received_lan_offer' = [type |-> MSG_lan_offer, addr |-> lan_addr, challenge |-> challenge_bytes]
    /\ backend_state' = backend_LANOffered
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_confirm, received_path_ping>>

Cmds_backend_RelayConnected_to_LANOffered_lan_server_ready == {CMD_send_lan_offer}

\* backend: LANOffered -> LANActive on recv lan_verify [challenge_valid]
backend_LANOffered_to_LANActive_on_lan_verify_challenge_valid ==
    /\ backend_state = backend_LANOffered
    /\ received_lan_verify.type = MSG_lan_verify
    /\ offer_challenge = challenge_bytes
    /\ received_lan_verify' = [type |-> "none"]
    /\ received_lan_confirm' = [type |-> MSG_lan_confirm]
    /\ backend_state' = backend_LANActive
    /\ ping_failures' = 0
    /\ backoff_level' = 0
    /\ b_active_path' = "lan"
    /\ b_dispatcher_path' = "lan"
    /\ monitor_target' = "lan"
    /\ lan_signal' = "ready"
    /\ UNCHANGED <<client_state, c_active_path, c_dispatcher_path, received_path_pong, received_lan_offer, received_path_ping>>

Cmds_backend_LANOffered_to_LANActive_on_lan_verify_challenge_valid == {CMD_send_lan_confirm, CMD_start_lan_stream_reader, CMD_start_lan_dg_reader, CMD_start_monitor, CMD_signal_lan_ready, CMD_set_crypto_datagram}

\* backend: LANOffered -> RelayConnected on recv lan_verify [challenge_invalid]
backend_LANOffered_to_RelayConnected_on_lan_verify_challenge_invalid ==
    /\ backend_state = backend_LANOffered
    /\ received_lan_verify.type = MSG_lan_verify
    /\ offer_challenge /= challenge_bytes
    /\ received_lan_verify' = [type |-> "none"]
    /\ backend_state' = backend_RelayConnected
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

\* backend: LANOffered -> RelayBackoff (offer_timeout)
backend_LANOffered_to_RelayBackoff_offer_timeout ==
    /\ backend_state = backend_LANOffered
    /\ backend_state' = backend_RelayBackoff
    /\ backoff_level' = Min(backoff_level + 1, max_backoff_level)
    /\ lan_signal' = "pending"
    /\ UNCHANGED <<client_state, ping_failures, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_LANOffered_to_RelayBackoff_offer_timeout == {CMD_reset_lan_ready, CMD_start_backoff_timer}

\* backend: LANActive -> LANActive (ping_tick)
backend_LANActive_to_LANActive_ping_tick ==
    /\ backend_state = backend_LANActive
    /\ received_path_ping' = [type |-> MSG_path_ping]
    /\ backend_state' = backend_LANActive
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm>>

Cmds_backend_LANActive_to_LANActive_ping_tick == {CMD_send_path_ping, CMD_start_pong_timeout}

\* backend: LANActive -> LANDegraded (ping_timeout)
backend_LANActive_to_LANDegraded_ping_timeout ==
    /\ backend_state = backend_LANActive
    /\ backend_state' = backend_LANDegraded
    /\ ping_failures' = 1
    /\ UNCHANGED <<client_state, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

\* backend: LANDegraded -> LANDegraded (ping_tick)
backend_LANDegraded_to_LANDegraded_ping_tick ==
    /\ backend_state = backend_LANDegraded
    /\ received_path_ping' = [type |-> MSG_path_ping]
    /\ backend_state' = backend_LANDegraded
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm>>

Cmds_backend_LANDegraded_to_LANDegraded_ping_tick == {CMD_send_path_ping, CMD_start_pong_timeout}

\* backend: LANActive -> RelayBackoff (lan_stream_error)
backend_LANActive_to_RelayBackoff_lan_stream_error ==
    /\ backend_state = backend_LANActive
    /\ backend_state' = backend_RelayBackoff
    /\ backoff_level' = Min(backoff_level + 1, max_backoff_level)
    /\ b_active_path' = "relay"
    /\ b_dispatcher_path' = "relay"
    /\ monitor_target' = "none"
    /\ lan_signal' = "pending"
    /\ ping_failures' = 0
    /\ UNCHANGED <<client_state, c_active_path, c_dispatcher_path, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_LANActive_to_RelayBackoff_lan_stream_error == {CMD_stop_monitor, CMD_stop_lan_stream_reader, CMD_stop_lan_dg_reader, CMD_close_lan_path, CMD_reset_lan_ready, CMD_start_backoff_timer}

\* backend: LANDegraded -> RelayBackoff (lan_stream_error)
backend_LANDegraded_to_RelayBackoff_lan_stream_error ==
    /\ backend_state = backend_LANDegraded
    /\ backend_state' = backend_RelayBackoff
    /\ backoff_level' = Min(backoff_level + 1, max_backoff_level)
    /\ b_active_path' = "relay"
    /\ b_dispatcher_path' = "relay"
    /\ monitor_target' = "none"
    /\ lan_signal' = "pending"
    /\ ping_failures' = 0
    /\ UNCHANGED <<client_state, c_active_path, c_dispatcher_path, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_LANDegraded_to_RelayBackoff_lan_stream_error == {CMD_stop_monitor, CMD_stop_lan_stream_reader, CMD_stop_lan_dg_reader, CMD_close_lan_path, CMD_reset_lan_ready, CMD_start_backoff_timer}

\* backend: LANDegraded -> LANActive on recv path_pong
backend_LANDegraded_to_LANActive_on_path_pong ==
    /\ backend_state = backend_LANDegraded
    /\ received_path_pong.type = MSG_path_pong
    /\ received_path_pong' = [type |-> "none"]
    /\ backend_state' = backend_LANActive
    /\ ping_failures' = 0
    /\ UNCHANGED <<client_state, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_LANDegraded_to_LANActive_on_path_pong == {CMD_cancel_pong_timeout}

\* backend: LANDegraded -> LANDegraded (ping_timeout) [under_max_failures]
backend_LANDegraded_to_LANDegraded_ping_timeout_under_max_failures ==
    /\ backend_state = backend_LANDegraded
    /\ ping_failures + 1 < max_ping_failures
    /\ backend_state' = backend_LANDegraded
    /\ ping_failures' = ping_failures + 1
    /\ UNCHANGED <<client_state, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

\* backend: LANDegraded -> RelayBackoff (ping_timeout) [at_max_failures]
backend_LANDegraded_to_RelayBackoff_ping_timeout_at_max_failures ==
    /\ backend_state = backend_LANDegraded
    /\ ping_failures + 1 >= max_ping_failures
    /\ backend_state' = backend_RelayBackoff
    /\ backoff_level' = Min(backoff_level + 1, max_backoff_level)
    /\ b_active_path' = "relay"
    /\ b_dispatcher_path' = "relay"
    /\ monitor_target' = "none"
    /\ lan_signal' = "pending"
    /\ ping_failures' = 0
    /\ UNCHANGED <<client_state, c_active_path, c_dispatcher_path, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_LANDegraded_to_RelayBackoff_ping_timeout_at_max_failures == {CMD_stop_monitor, CMD_stop_lan_stream_reader, CMD_stop_lan_dg_reader, CMD_close_lan_path, CMD_reset_lan_ready, CMD_start_backoff_timer}

\* backend: RelayBackoff -> LANOffered (backoff_expired)
backend_RelayBackoff_to_LANOffered_backoff_expired ==
    /\ backend_state = backend_RelayBackoff
    /\ received_lan_offer' = [type |-> MSG_lan_offer, addr |-> lan_addr, challenge |-> challenge_bytes]
    /\ backend_state' = backend_LANOffered
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_confirm, received_path_ping>>

Cmds_backend_RelayBackoff_to_LANOffered_backoff_expired == {CMD_send_lan_offer}

\* backend: RelayBackoff -> LANOffered (lan_server_changed)
backend_RelayBackoff_to_LANOffered_lan_server_changed ==
    /\ backend_state = backend_RelayBackoff
    /\ received_lan_offer' = [type |-> MSG_lan_offer, addr |-> lan_addr, challenge |-> challenge_bytes]
    /\ backend_state' = backend_LANOffered
    /\ backoff_level' = 0
    /\ UNCHANGED <<client_state, ping_failures, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_confirm, received_path_ping>>

Cmds_backend_RelayBackoff_to_LANOffered_lan_server_changed == {CMD_send_lan_offer}

\* backend: RelayConnected -> LANOffered (readvertise_tick) [lan_server_available]
backend_RelayConnected_to_LANOffered_readvertise_tick_lan_server_available ==
    /\ backend_state = backend_RelayConnected
    /\ lan_server_addr /= "none"
    /\ received_lan_offer' = [type |-> MSG_lan_offer, addr |-> lan_addr, challenge |-> challenge_bytes]
    /\ backend_state' = backend_LANOffered
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_confirm, received_path_ping>>

Cmds_backend_RelayConnected_to_LANOffered_readvertise_tick_lan_server_available == {CMD_send_lan_offer}

\* backend: LANOffered -> RelayConnected (app_force_fallback)
backend_LANOffered_to_RelayConnected_app_force_fallback ==
    /\ backend_state = backend_LANOffered
    /\ backend_state' = backend_RelayConnected
    /\ lan_signal' = "pending"
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_LANOffered_to_RelayConnected_app_force_fallback == {CMD_reset_lan_ready}

\* backend: LANActive -> RelayBackoff (app_force_fallback)
backend_LANActive_to_RelayBackoff_app_force_fallback ==
    /\ backend_state = backend_LANActive
    /\ backend_state' = backend_RelayBackoff
    /\ backoff_level' = Min(backoff_level + 1, max_backoff_level)
    /\ b_active_path' = "relay"
    /\ b_dispatcher_path' = "relay"
    /\ monitor_target' = "none"
    /\ lan_signal' = "pending"
    /\ ping_failures' = 0
    /\ UNCHANGED <<client_state, c_active_path, c_dispatcher_path, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_LANActive_to_RelayBackoff_app_force_fallback == {CMD_stop_monitor, CMD_cancel_pong_timeout, CMD_stop_lan_stream_reader, CMD_stop_lan_dg_reader, CMD_close_lan_path, CMD_reset_lan_ready, CMD_start_backoff_timer}

\* backend: LANDegraded -> RelayBackoff (app_force_fallback)
backend_LANDegraded_to_RelayBackoff_app_force_fallback ==
    /\ backend_state = backend_LANDegraded
    /\ backend_state' = backend_RelayBackoff
    /\ backoff_level' = Min(backoff_level + 1, max_backoff_level)
    /\ b_active_path' = "relay"
    /\ b_dispatcher_path' = "relay"
    /\ monitor_target' = "none"
    /\ lan_signal' = "pending"
    /\ ping_failures' = 0
    /\ UNCHANGED <<client_state, c_active_path, c_dispatcher_path, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_LANDegraded_to_RelayBackoff_app_force_fallback == {CMD_stop_monitor, CMD_cancel_pong_timeout, CMD_stop_lan_stream_reader, CMD_stop_lan_dg_reader, CMD_close_lan_path, CMD_reset_lan_ready, CMD_start_backoff_timer}

\* backend: RelayConnected -> RelayConnected (app_send)
backend_RelayConnected_to_RelayConnected_app_send ==
    /\ backend_state = backend_RelayConnected
    /\ backend_state' = backend_RelayConnected
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_RelayConnected_to_RelayConnected_app_send == {CMD_write_active_stream}

\* backend: LANOffered -> LANOffered (app_send)
backend_LANOffered_to_LANOffered_app_send ==
    /\ backend_state = backend_LANOffered
    /\ backend_state' = backend_LANOffered
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_LANOffered_to_LANOffered_app_send == {CMD_write_active_stream}

\* backend: LANActive -> LANActive (app_send)
backend_LANActive_to_LANActive_app_send ==
    /\ backend_state = backend_LANActive
    /\ backend_state' = backend_LANActive
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_LANActive_to_LANActive_app_send == {CMD_write_active_stream}

\* backend: LANDegraded -> LANDegraded (app_send)
backend_LANDegraded_to_LANDegraded_app_send ==
    /\ backend_state = backend_LANDegraded
    /\ backend_state' = backend_LANDegraded
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_LANDegraded_to_LANDegraded_app_send == {CMD_write_active_stream}

\* backend: RelayBackoff -> RelayBackoff (app_send)
backend_RelayBackoff_to_RelayBackoff_app_send ==
    /\ backend_state = backend_RelayBackoff
    /\ backend_state' = backend_RelayBackoff
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_RelayBackoff_to_RelayBackoff_app_send == {CMD_write_active_stream}

\* backend: RelayConnected -> RelayConnected (relay_stream_data)
backend_RelayConnected_to_RelayConnected_relay_stream_data ==
    /\ backend_state = backend_RelayConnected
    /\ backend_state' = backend_RelayConnected
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_RelayConnected_to_RelayConnected_relay_stream_data == {CMD_deliver_recv}

\* backend: LANOffered -> LANOffered (relay_stream_data)
backend_LANOffered_to_LANOffered_relay_stream_data ==
    /\ backend_state = backend_LANOffered
    /\ backend_state' = backend_LANOffered
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_LANOffered_to_LANOffered_relay_stream_data == {CMD_deliver_recv}

\* backend: LANActive -> LANActive (relay_stream_data)
backend_LANActive_to_LANActive_relay_stream_data ==
    /\ backend_state = backend_LANActive
    /\ backend_state' = backend_LANActive
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_LANActive_to_LANActive_relay_stream_data == {CMD_deliver_recv}

\* backend: LANDegraded -> LANDegraded (relay_stream_data)
backend_LANDegraded_to_LANDegraded_relay_stream_data ==
    /\ backend_state = backend_LANDegraded
    /\ backend_state' = backend_LANDegraded
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_LANDegraded_to_LANDegraded_relay_stream_data == {CMD_deliver_recv}

\* backend: RelayBackoff -> RelayBackoff (relay_stream_data)
backend_RelayBackoff_to_RelayBackoff_relay_stream_data ==
    /\ backend_state = backend_RelayBackoff
    /\ backend_state' = backend_RelayBackoff
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_RelayBackoff_to_RelayBackoff_relay_stream_data == {CMD_deliver_recv}

\* backend: RelayConnected -> RelayConnected (relay_stream_error)
backend_RelayConnected_to_RelayConnected_relay_stream_error ==
    /\ backend_state = backend_RelayConnected
    /\ backend_state' = backend_RelayConnected
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_RelayConnected_to_RelayConnected_relay_stream_error == {CMD_deliver_recv_error}

\* backend: LANOffered -> LANOffered (relay_stream_error)
backend_LANOffered_to_LANOffered_relay_stream_error ==
    /\ backend_state = backend_LANOffered
    /\ backend_state' = backend_LANOffered
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_LANOffered_to_LANOffered_relay_stream_error == {CMD_deliver_recv_error}

\* backend: LANActive -> LANActive (relay_stream_error)
backend_LANActive_to_LANActive_relay_stream_error ==
    /\ backend_state = backend_LANActive
    /\ backend_state' = backend_LANActive
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_LANActive_to_LANActive_relay_stream_error == {CMD_deliver_recv_error}

\* backend: LANDegraded -> LANDegraded (relay_stream_error)
backend_LANDegraded_to_LANDegraded_relay_stream_error ==
    /\ backend_state = backend_LANDegraded
    /\ backend_state' = backend_LANDegraded
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_LANDegraded_to_LANDegraded_relay_stream_error == {CMD_deliver_recv_error}

\* backend: RelayBackoff -> RelayBackoff (relay_stream_error)
backend_RelayBackoff_to_RelayBackoff_relay_stream_error ==
    /\ backend_state = backend_RelayBackoff
    /\ backend_state' = backend_RelayBackoff
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_RelayBackoff_to_RelayBackoff_relay_stream_error == {CMD_deliver_recv_error}

\* backend: RelayConnected -> RelayConnected (app_send_datagram)
backend_RelayConnected_to_RelayConnected_app_send_datagram ==
    /\ backend_state = backend_RelayConnected
    /\ backend_state' = backend_RelayConnected
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_RelayConnected_to_RelayConnected_app_send_datagram == {CMD_send_active_datagram}

\* backend: LANOffered -> LANOffered (app_send_datagram)
backend_LANOffered_to_LANOffered_app_send_datagram ==
    /\ backend_state = backend_LANOffered
    /\ backend_state' = backend_LANOffered
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_LANOffered_to_LANOffered_app_send_datagram == {CMD_send_active_datagram}

\* backend: LANActive -> LANActive (app_send_datagram)
backend_LANActive_to_LANActive_app_send_datagram ==
    /\ backend_state = backend_LANActive
    /\ backend_state' = backend_LANActive
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_LANActive_to_LANActive_app_send_datagram == {CMD_send_active_datagram}

\* backend: LANDegraded -> LANDegraded (app_send_datagram)
backend_LANDegraded_to_LANDegraded_app_send_datagram ==
    /\ backend_state = backend_LANDegraded
    /\ backend_state' = backend_LANDegraded
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_LANDegraded_to_LANDegraded_app_send_datagram == {CMD_send_active_datagram}

\* backend: RelayBackoff -> RelayBackoff (app_send_datagram)
backend_RelayBackoff_to_RelayBackoff_app_send_datagram ==
    /\ backend_state = backend_RelayBackoff
    /\ backend_state' = backend_RelayBackoff
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_RelayBackoff_to_RelayBackoff_app_send_datagram == {CMD_send_active_datagram}

\* backend: RelayConnected -> RelayConnected (relay_datagram)
backend_RelayConnected_to_RelayConnected_relay_datagram ==
    /\ backend_state = backend_RelayConnected
    /\ backend_state' = backend_RelayConnected
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_RelayConnected_to_RelayConnected_relay_datagram == {CMD_deliver_recv_datagram}

\* backend: LANOffered -> LANOffered (relay_datagram)
backend_LANOffered_to_LANOffered_relay_datagram ==
    /\ backend_state = backend_LANOffered
    /\ backend_state' = backend_LANOffered
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_LANOffered_to_LANOffered_relay_datagram == {CMD_deliver_recv_datagram}

\* backend: LANActive -> LANActive (relay_datagram)
backend_LANActive_to_LANActive_relay_datagram ==
    /\ backend_state = backend_LANActive
    /\ backend_state' = backend_LANActive
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_LANActive_to_LANActive_relay_datagram == {CMD_deliver_recv_datagram}

\* backend: LANDegraded -> LANDegraded (relay_datagram)
backend_LANDegraded_to_LANDegraded_relay_datagram ==
    /\ backend_state = backend_LANDegraded
    /\ backend_state' = backend_LANDegraded
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_LANDegraded_to_LANDegraded_relay_datagram == {CMD_deliver_recv_datagram}

\* backend: RelayBackoff -> RelayBackoff (relay_datagram)
backend_RelayBackoff_to_RelayBackoff_relay_datagram ==
    /\ backend_state = backend_RelayBackoff
    /\ backend_state' = backend_RelayBackoff
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_RelayBackoff_to_RelayBackoff_relay_datagram == {CMD_deliver_recv_datagram}

\* backend: LANActive -> LANActive (lan_stream_data)
backend_LANActive_to_LANActive_lan_stream_data ==
    /\ backend_state = backend_LANActive
    /\ backend_state' = backend_LANActive
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_LANActive_to_LANActive_lan_stream_data == {CMD_deliver_recv}

\* backend: LANDegraded -> LANDegraded (lan_stream_data)
backend_LANDegraded_to_LANDegraded_lan_stream_data ==
    /\ backend_state = backend_LANDegraded
    /\ backend_state' = backend_LANDegraded
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_LANDegraded_to_LANDegraded_lan_stream_data == {CMD_deliver_recv}

\* backend: LANActive -> LANActive (lan_datagram)
backend_LANActive_to_LANActive_lan_datagram ==
    /\ backend_state = backend_LANActive
    /\ backend_state' = backend_LANActive
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_LANActive_to_LANActive_lan_datagram == {CMD_deliver_recv_datagram}

\* backend: LANDegraded -> LANDegraded (lan_datagram)
backend_LANDegraded_to_LANDegraded_lan_datagram ==
    /\ backend_state = backend_LANDegraded
    /\ backend_state' = backend_LANDegraded
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_backend_LANDegraded_to_LANDegraded_lan_datagram == {CMD_deliver_recv_datagram}


\* client: RelayConnected -> LANConnecting on recv lan_offer [lan_enabled]
client_RelayConnected_to_LANConnecting_on_lan_offer_lan_enabled ==
    /\ client_state = client_RelayConnected
    /\ received_lan_offer.type = MSG_lan_offer
    /\ TRUE
    /\ received_lan_offer' = [type |-> "none"]
    /\ client_state' = client_LANConnecting
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_confirm, received_path_ping>>

Cmds_client_RelayConnected_to_LANConnecting_on_lan_offer_lan_enabled == {CMD_dial_lan}

\* client: RelayConnected -> RelayConnected on recv lan_offer [lan_disabled]
client_RelayConnected_to_RelayConnected_on_lan_offer_lan_disabled ==
    /\ client_state = client_RelayConnected
    /\ received_lan_offer.type = MSG_lan_offer
    /\ FALSE
    /\ received_lan_offer' = [type |-> "none"]
    /\ client_state' = client_RelayConnected
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_confirm, received_path_ping>>

\* client: LANConnecting -> LANVerifying (lan_dial_ok)
client_LANConnecting_to_LANVerifying_lan_dial_ok ==
    /\ client_state = client_LANConnecting
    /\ received_lan_verify' = [type |-> MSG_lan_verify, challenge |-> offer_challenge, instance_id |-> instance_id]
    /\ client_state' = client_LANVerifying
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_LANConnecting_to_LANVerifying_lan_dial_ok == {CMD_send_lan_verify}

\* client: LANConnecting -> RelayConnected (lan_dial_failed)
client_LANConnecting_to_RelayConnected_lan_dial_failed ==
    /\ client_state = client_LANConnecting
    /\ client_state' = client_RelayConnected
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

\* client: LANVerifying -> LANActive on recv lan_confirm
client_LANVerifying_to_LANActive_on_lan_confirm ==
    /\ client_state = client_LANVerifying
    /\ received_lan_confirm.type = MSG_lan_confirm
    /\ received_lan_confirm' = [type |-> "none"]
    /\ client_state' = client_LANActive
    /\ c_active_path' = "lan"
    /\ c_dispatcher_path' = "lan"
    /\ lan_signal' = "ready"
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, b_dispatcher_path, monitor_target, received_lan_verify, received_path_pong, received_lan_offer, received_path_ping>>

Cmds_client_LANVerifying_to_LANActive_on_lan_confirm == {CMD_start_lan_stream_reader, CMD_start_lan_dg_reader, CMD_signal_lan_ready, CMD_set_crypto_datagram}

\* client: LANVerifying -> RelayConnected (verify_timeout)
client_LANVerifying_to_RelayConnected_verify_timeout ==
    /\ client_state = client_LANVerifying
    /\ client_state' = client_RelayConnected
    /\ c_dispatcher_path' = "relay"
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

\* client: LANActive -> LANActive on recv path_ping
client_LANActive_to_LANActive_on_path_ping ==
    /\ client_state = client_LANActive
    /\ received_path_ping.type = MSG_path_ping
    /\ received_path_ping' = [type |-> "none"]
    /\ received_path_pong' = [type |-> MSG_path_pong]
    /\ client_state' = client_LANActive
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_lan_offer, received_lan_confirm>>

Cmds_client_LANActive_to_LANActive_on_path_ping == {CMD_send_path_pong}

\* client: LANActive -> RelayFallback (lan_error)
client_LANActive_to_RelayFallback_lan_error ==
    /\ client_state = client_LANActive
    /\ client_state' = client_RelayFallback
    /\ c_active_path' = "relay"
    /\ c_dispatcher_path' = "relay"
    /\ lan_signal' = "pending"
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, b_dispatcher_path, monitor_target, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_LANActive_to_RelayFallback_lan_error == {CMD_stop_lan_stream_reader, CMD_stop_lan_dg_reader, CMD_close_lan_path, CMD_reset_lan_ready}

\* client: LANActive -> RelayFallback (lan_stream_error)
client_LANActive_to_RelayFallback_lan_stream_error ==
    /\ client_state = client_LANActive
    /\ client_state' = client_RelayFallback
    /\ c_active_path' = "relay"
    /\ c_dispatcher_path' = "relay"
    /\ lan_signal' = "pending"
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, b_dispatcher_path, monitor_target, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_LANActive_to_RelayFallback_lan_stream_error == {CMD_stop_lan_stream_reader, CMD_stop_lan_dg_reader, CMD_close_lan_path, CMD_reset_lan_ready}

\* client: RelayFallback -> RelayConnected (relay_ok)
client_RelayFallback_to_RelayConnected_relay_ok ==
    /\ client_state = client_RelayFallback
    /\ client_state' = client_RelayConnected
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

\* client: LANActive -> LANConnecting on recv lan_offer [lan_enabled]
client_LANActive_to_LANConnecting_on_lan_offer_lan_enabled ==
    /\ client_state = client_LANActive
    /\ received_lan_offer.type = MSG_lan_offer
    /\ TRUE
    /\ received_lan_offer' = [type |-> "none"]
    /\ client_state' = client_LANConnecting
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_confirm, received_path_ping>>

Cmds_client_LANActive_to_LANConnecting_on_lan_offer_lan_enabled == {CMD_stop_lan_stream_reader, CMD_stop_lan_dg_reader, CMD_close_lan_path, CMD_dial_lan}

\* client: LANConnecting -> RelayConnected (app_force_fallback)
client_LANConnecting_to_RelayConnected_app_force_fallback ==
    /\ client_state = client_LANConnecting
    /\ client_state' = client_RelayConnected
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

\* client: LANVerifying -> RelayConnected (app_force_fallback)
client_LANVerifying_to_RelayConnected_app_force_fallback ==
    /\ client_state = client_LANVerifying
    /\ client_state' = client_RelayConnected
    /\ c_dispatcher_path' = "relay"
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_LANVerifying_to_RelayConnected_app_force_fallback == {CMD_stop_lan_stream_reader, CMD_stop_lan_dg_reader, CMD_close_lan_path}

\* client: LANActive -> RelayConnected (app_force_fallback)
client_LANActive_to_RelayConnected_app_force_fallback ==
    /\ client_state = client_LANActive
    /\ client_state' = client_RelayConnected
    /\ c_active_path' = "relay"
    /\ c_dispatcher_path' = "relay"
    /\ lan_signal' = "pending"
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, b_dispatcher_path, monitor_target, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_LANActive_to_RelayConnected_app_force_fallback == {CMD_stop_lan_stream_reader, CMD_stop_lan_dg_reader, CMD_close_lan_path, CMD_reset_lan_ready}

\* client: RelayConnected -> RelayConnected (app_send)
client_RelayConnected_to_RelayConnected_app_send ==
    /\ client_state = client_RelayConnected
    /\ client_state' = client_RelayConnected
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_RelayConnected_to_RelayConnected_app_send == {CMD_write_active_stream}

\* client: LANConnecting -> LANConnecting (app_send)
client_LANConnecting_to_LANConnecting_app_send ==
    /\ client_state = client_LANConnecting
    /\ client_state' = client_LANConnecting
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_LANConnecting_to_LANConnecting_app_send == {CMD_write_active_stream}

\* client: LANVerifying -> LANVerifying (app_send)
client_LANVerifying_to_LANVerifying_app_send ==
    /\ client_state = client_LANVerifying
    /\ client_state' = client_LANVerifying
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_LANVerifying_to_LANVerifying_app_send == {CMD_write_active_stream}

\* client: LANActive -> LANActive (app_send)
client_LANActive_to_LANActive_app_send ==
    /\ client_state = client_LANActive
    /\ client_state' = client_LANActive
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_LANActive_to_LANActive_app_send == {CMD_write_active_stream}

\* client: RelayFallback -> RelayFallback (app_send)
client_RelayFallback_to_RelayFallback_app_send ==
    /\ client_state = client_RelayFallback
    /\ client_state' = client_RelayFallback
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_RelayFallback_to_RelayFallback_app_send == {CMD_write_active_stream}

\* client: RelayConnected -> RelayConnected (relay_stream_data)
client_RelayConnected_to_RelayConnected_relay_stream_data ==
    /\ client_state = client_RelayConnected
    /\ client_state' = client_RelayConnected
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_RelayConnected_to_RelayConnected_relay_stream_data == {CMD_deliver_recv}

\* client: LANConnecting -> LANConnecting (relay_stream_data)
client_LANConnecting_to_LANConnecting_relay_stream_data ==
    /\ client_state = client_LANConnecting
    /\ client_state' = client_LANConnecting
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_LANConnecting_to_LANConnecting_relay_stream_data == {CMD_deliver_recv}

\* client: LANVerifying -> LANVerifying (relay_stream_data)
client_LANVerifying_to_LANVerifying_relay_stream_data ==
    /\ client_state = client_LANVerifying
    /\ client_state' = client_LANVerifying
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_LANVerifying_to_LANVerifying_relay_stream_data == {CMD_deliver_recv}

\* client: LANActive -> LANActive (relay_stream_data)
client_LANActive_to_LANActive_relay_stream_data ==
    /\ client_state = client_LANActive
    /\ client_state' = client_LANActive
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_LANActive_to_LANActive_relay_stream_data == {CMD_deliver_recv}

\* client: RelayFallback -> RelayFallback (relay_stream_data)
client_RelayFallback_to_RelayFallback_relay_stream_data ==
    /\ client_state = client_RelayFallback
    /\ client_state' = client_RelayFallback
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_RelayFallback_to_RelayFallback_relay_stream_data == {CMD_deliver_recv}

\* client: RelayConnected -> RelayConnected (relay_stream_error)
client_RelayConnected_to_RelayConnected_relay_stream_error ==
    /\ client_state = client_RelayConnected
    /\ client_state' = client_RelayConnected
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_RelayConnected_to_RelayConnected_relay_stream_error == {CMD_deliver_recv_error}

\* client: LANConnecting -> LANConnecting (relay_stream_error)
client_LANConnecting_to_LANConnecting_relay_stream_error ==
    /\ client_state = client_LANConnecting
    /\ client_state' = client_LANConnecting
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_LANConnecting_to_LANConnecting_relay_stream_error == {CMD_deliver_recv_error}

\* client: LANVerifying -> LANVerifying (relay_stream_error)
client_LANVerifying_to_LANVerifying_relay_stream_error ==
    /\ client_state = client_LANVerifying
    /\ client_state' = client_LANVerifying
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_LANVerifying_to_LANVerifying_relay_stream_error == {CMD_deliver_recv_error}

\* client: LANActive -> LANActive (relay_stream_error)
client_LANActive_to_LANActive_relay_stream_error ==
    /\ client_state = client_LANActive
    /\ client_state' = client_LANActive
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_LANActive_to_LANActive_relay_stream_error == {CMD_deliver_recv_error}

\* client: RelayFallback -> RelayFallback (relay_stream_error)
client_RelayFallback_to_RelayFallback_relay_stream_error ==
    /\ client_state = client_RelayFallback
    /\ client_state' = client_RelayFallback
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_RelayFallback_to_RelayFallback_relay_stream_error == {CMD_deliver_recv_error}

\* client: RelayConnected -> RelayConnected (app_send_datagram)
client_RelayConnected_to_RelayConnected_app_send_datagram ==
    /\ client_state = client_RelayConnected
    /\ client_state' = client_RelayConnected
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_RelayConnected_to_RelayConnected_app_send_datagram == {CMD_send_active_datagram}

\* client: LANConnecting -> LANConnecting (app_send_datagram)
client_LANConnecting_to_LANConnecting_app_send_datagram ==
    /\ client_state = client_LANConnecting
    /\ client_state' = client_LANConnecting
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_LANConnecting_to_LANConnecting_app_send_datagram == {CMD_send_active_datagram}

\* client: LANVerifying -> LANVerifying (app_send_datagram)
client_LANVerifying_to_LANVerifying_app_send_datagram ==
    /\ client_state = client_LANVerifying
    /\ client_state' = client_LANVerifying
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_LANVerifying_to_LANVerifying_app_send_datagram == {CMD_send_active_datagram}

\* client: LANActive -> LANActive (app_send_datagram)
client_LANActive_to_LANActive_app_send_datagram ==
    /\ client_state = client_LANActive
    /\ client_state' = client_LANActive
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_LANActive_to_LANActive_app_send_datagram == {CMD_send_active_datagram}

\* client: RelayFallback -> RelayFallback (app_send_datagram)
client_RelayFallback_to_RelayFallback_app_send_datagram ==
    /\ client_state = client_RelayFallback
    /\ client_state' = client_RelayFallback
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_RelayFallback_to_RelayFallback_app_send_datagram == {CMD_send_active_datagram}

\* client: RelayConnected -> RelayConnected (relay_datagram)
client_RelayConnected_to_RelayConnected_relay_datagram ==
    /\ client_state = client_RelayConnected
    /\ client_state' = client_RelayConnected
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_RelayConnected_to_RelayConnected_relay_datagram == {CMD_deliver_recv_datagram}

\* client: LANConnecting -> LANConnecting (relay_datagram)
client_LANConnecting_to_LANConnecting_relay_datagram ==
    /\ client_state = client_LANConnecting
    /\ client_state' = client_LANConnecting
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_LANConnecting_to_LANConnecting_relay_datagram == {CMD_deliver_recv_datagram}

\* client: LANVerifying -> LANVerifying (relay_datagram)
client_LANVerifying_to_LANVerifying_relay_datagram ==
    /\ client_state = client_LANVerifying
    /\ client_state' = client_LANVerifying
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_LANVerifying_to_LANVerifying_relay_datagram == {CMD_deliver_recv_datagram}

\* client: LANActive -> LANActive (relay_datagram)
client_LANActive_to_LANActive_relay_datagram ==
    /\ client_state = client_LANActive
    /\ client_state' = client_LANActive
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_LANActive_to_LANActive_relay_datagram == {CMD_deliver_recv_datagram}

\* client: RelayFallback -> RelayFallback (relay_datagram)
client_RelayFallback_to_RelayFallback_relay_datagram ==
    /\ client_state = client_RelayFallback
    /\ client_state' = client_RelayFallback
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_RelayFallback_to_RelayFallback_relay_datagram == {CMD_deliver_recv_datagram}

\* client: LANActive -> LANActive (lan_stream_data)
client_LANActive_to_LANActive_lan_stream_data ==
    /\ client_state = client_LANActive
    /\ client_state' = client_LANActive
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_LANActive_to_LANActive_lan_stream_data == {CMD_deliver_recv}

\* client: LANActive -> LANActive (lan_datagram)
client_LANActive_to_LANActive_lan_datagram ==
    /\ client_state = client_LANActive
    /\ client_state' = client_LANActive
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

Cmds_client_LANActive_to_LANActive_lan_datagram == {CMD_deliver_recv_datagram}


Next ==
    \/ backend_RelayConnected_to_LANOffered_lan_server_ready
    \/ backend_LANOffered_to_LANActive_on_lan_verify_challenge_valid
    \/ backend_LANOffered_to_RelayConnected_on_lan_verify_challenge_invalid
    \/ backend_LANOffered_to_RelayBackoff_offer_timeout
    \/ backend_LANActive_to_LANActive_ping_tick
    \/ backend_LANActive_to_LANDegraded_ping_timeout
    \/ backend_LANDegraded_to_LANDegraded_ping_tick
    \/ backend_LANActive_to_RelayBackoff_lan_stream_error
    \/ backend_LANDegraded_to_RelayBackoff_lan_stream_error
    \/ backend_LANDegraded_to_LANActive_on_path_pong
    \/ backend_LANDegraded_to_LANDegraded_ping_timeout_under_max_failures
    \/ backend_LANDegraded_to_RelayBackoff_ping_timeout_at_max_failures
    \/ backend_RelayBackoff_to_LANOffered_backoff_expired
    \/ backend_RelayBackoff_to_LANOffered_lan_server_changed
    \/ backend_RelayConnected_to_LANOffered_readvertise_tick_lan_server_available
    \/ backend_LANOffered_to_RelayConnected_app_force_fallback
    \/ backend_LANActive_to_RelayBackoff_app_force_fallback
    \/ backend_LANDegraded_to_RelayBackoff_app_force_fallback
    \/ backend_RelayConnected_to_RelayConnected_app_send
    \/ backend_LANOffered_to_LANOffered_app_send
    \/ backend_LANActive_to_LANActive_app_send
    \/ backend_LANDegraded_to_LANDegraded_app_send
    \/ backend_RelayBackoff_to_RelayBackoff_app_send
    \/ backend_RelayConnected_to_RelayConnected_relay_stream_data
    \/ backend_LANOffered_to_LANOffered_relay_stream_data
    \/ backend_LANActive_to_LANActive_relay_stream_data
    \/ backend_LANDegraded_to_LANDegraded_relay_stream_data
    \/ backend_RelayBackoff_to_RelayBackoff_relay_stream_data
    \/ backend_RelayConnected_to_RelayConnected_relay_stream_error
    \/ backend_LANOffered_to_LANOffered_relay_stream_error
    \/ backend_LANActive_to_LANActive_relay_stream_error
    \/ backend_LANDegraded_to_LANDegraded_relay_stream_error
    \/ backend_RelayBackoff_to_RelayBackoff_relay_stream_error
    \/ backend_RelayConnected_to_RelayConnected_app_send_datagram
    \/ backend_LANOffered_to_LANOffered_app_send_datagram
    \/ backend_LANActive_to_LANActive_app_send_datagram
    \/ backend_LANDegraded_to_LANDegraded_app_send_datagram
    \/ backend_RelayBackoff_to_RelayBackoff_app_send_datagram
    \/ backend_RelayConnected_to_RelayConnected_relay_datagram
    \/ backend_LANOffered_to_LANOffered_relay_datagram
    \/ backend_LANActive_to_LANActive_relay_datagram
    \/ backend_LANDegraded_to_LANDegraded_relay_datagram
    \/ backend_RelayBackoff_to_RelayBackoff_relay_datagram
    \/ backend_LANActive_to_LANActive_lan_stream_data
    \/ backend_LANDegraded_to_LANDegraded_lan_stream_data
    \/ backend_LANActive_to_LANActive_lan_datagram
    \/ backend_LANDegraded_to_LANDegraded_lan_datagram
    \/ client_RelayConnected_to_LANConnecting_on_lan_offer_lan_enabled
    \/ client_RelayConnected_to_RelayConnected_on_lan_offer_lan_disabled
    \/ client_LANConnecting_to_LANVerifying_lan_dial_ok
    \/ client_LANConnecting_to_RelayConnected_lan_dial_failed
    \/ client_LANVerifying_to_LANActive_on_lan_confirm
    \/ client_LANVerifying_to_RelayConnected_verify_timeout
    \/ client_LANActive_to_LANActive_on_path_ping
    \/ client_LANActive_to_RelayFallback_lan_error
    \/ client_LANActive_to_RelayFallback_lan_stream_error
    \/ client_RelayFallback_to_RelayConnected_relay_ok
    \/ client_LANActive_to_LANConnecting_on_lan_offer_lan_enabled
    \/ client_LANConnecting_to_RelayConnected_app_force_fallback
    \/ client_LANVerifying_to_RelayConnected_app_force_fallback
    \/ client_LANActive_to_RelayConnected_app_force_fallback
    \/ client_RelayConnected_to_RelayConnected_app_send
    \/ client_LANConnecting_to_LANConnecting_app_send
    \/ client_LANVerifying_to_LANVerifying_app_send
    \/ client_LANActive_to_LANActive_app_send
    \/ client_RelayFallback_to_RelayFallback_app_send
    \/ client_RelayConnected_to_RelayConnected_relay_stream_data
    \/ client_LANConnecting_to_LANConnecting_relay_stream_data
    \/ client_LANVerifying_to_LANVerifying_relay_stream_data
    \/ client_LANActive_to_LANActive_relay_stream_data
    \/ client_RelayFallback_to_RelayFallback_relay_stream_data
    \/ client_RelayConnected_to_RelayConnected_relay_stream_error
    \/ client_LANConnecting_to_LANConnecting_relay_stream_error
    \/ client_LANVerifying_to_LANVerifying_relay_stream_error
    \/ client_LANActive_to_LANActive_relay_stream_error
    \/ client_RelayFallback_to_RelayFallback_relay_stream_error
    \/ client_RelayConnected_to_RelayConnected_app_send_datagram
    \/ client_LANConnecting_to_LANConnecting_app_send_datagram
    \/ client_LANVerifying_to_LANVerifying_app_send_datagram
    \/ client_LANActive_to_LANActive_app_send_datagram
    \/ client_RelayFallback_to_RelayFallback_app_send_datagram
    \/ client_RelayConnected_to_RelayConnected_relay_datagram
    \/ client_LANConnecting_to_LANConnecting_relay_datagram
    \/ client_LANVerifying_to_LANVerifying_relay_datagram
    \/ client_LANActive_to_LANActive_relay_datagram
    \/ client_RelayFallback_to_RelayFallback_relay_datagram
    \/ client_LANActive_to_LANActive_lan_stream_data
    \/ client_LANActive_to_LANActive_lan_datagram

Spec == Init /\ [][Next]_vars /\ WF_vars(Next)

\* ================================================================
\* Invariants and properties
\* ================================================================

\* Paths are always valid
PathConsistency == b_active_path \in {"relay", "lan"} /\ c_active_path \in {"relay", "lan"}
\* Backoff never exceeds cap
BackoffBounded == backoff_level <= max_backoff_level
\* LAN success resets backoff
BackoffResetsOnSuccess == backend_state = backend_LANActive => backoff_level = 0
\* Dispatchers always bound to valid path
DispatcherAlwaysBound == b_dispatcher_path \in {"relay", "lan"} /\ c_dispatcher_path \in {"relay", "lan"}
\* Backend dispatcher on LAN when LAN active
BackendDispatcherMatchesActive == backend_state = backend_LANActive => b_dispatcher_path = "lan"
\* Client dispatcher on LAN when LAN active
ClientDispatcherMatchesActive == client_state = client_LANActive => c_dispatcher_path = "lan"
\* Monitor only pings when LAN is active/degraded
MonitorOnlyWhenLAN == monitor_target = "lan" => backend_state \in {backend_LANActive, backend_LANDegraded}
\* After fallback, backend eventually re-advertises LAN
FallbackLeadsToReadvertise == (backend_state = backend_RelayBackoff) ~> (backend_state = backend_LANOffered)
\* Degraded state eventually resolves (recovery or fallback)
DegradedLeadsToResolutionOrFallback == (backend_state = backend_LANDegraded) ~> (backend_state \in {backend_LANActive, backend_RelayBackoff})

\* ================================================================
\* Command-consistency: state after transition matches emitted commands
\* These are verified by construction (the same YAML defines both
\* the variable updates and the command list), but documenting
\* them as TLA+ operators makes the relationship explicit.
\* ================================================================

\* backend_RelayConnected_to_LANOffered_lan_server_ready emits: CMD_send_lan_offer
\* backend_LANOffered_to_LANActive_on_lan_verify_challenge_valid emits: CMD_send_lan_confirm, CMD_start_lan_stream_reader, CMD_start_lan_dg_reader, CMD_start_monitor, CMD_signal_lan_ready, CMD_set_crypto_datagram
\* backend_LANOffered_to_RelayBackoff_offer_timeout emits: CMD_reset_lan_ready, CMD_start_backoff_timer
\* backend_LANActive_to_LANActive_ping_tick emits: CMD_send_path_ping, CMD_start_pong_timeout
\* backend_LANDegraded_to_LANDegraded_ping_tick emits: CMD_send_path_ping, CMD_start_pong_timeout
\* backend_LANActive_to_RelayBackoff_lan_stream_error emits: CMD_stop_monitor, CMD_stop_lan_stream_reader, CMD_stop_lan_dg_reader, CMD_close_lan_path, CMD_reset_lan_ready, CMD_start_backoff_timer
\* backend_LANDegraded_to_RelayBackoff_lan_stream_error emits: CMD_stop_monitor, CMD_stop_lan_stream_reader, CMD_stop_lan_dg_reader, CMD_close_lan_path, CMD_reset_lan_ready, CMD_start_backoff_timer
\* backend_LANDegraded_to_LANActive_on_path_pong emits: CMD_cancel_pong_timeout
\* backend_LANDegraded_to_RelayBackoff_ping_timeout_at_max_failures emits: CMD_stop_monitor, CMD_stop_lan_stream_reader, CMD_stop_lan_dg_reader, CMD_close_lan_path, CMD_reset_lan_ready, CMD_start_backoff_timer
\* backend_RelayBackoff_to_LANOffered_backoff_expired emits: CMD_send_lan_offer
\* backend_RelayBackoff_to_LANOffered_lan_server_changed emits: CMD_send_lan_offer
\* backend_RelayConnected_to_LANOffered_readvertise_tick_lan_server_available emits: CMD_send_lan_offer
\* backend_LANOffered_to_RelayConnected_app_force_fallback emits: CMD_reset_lan_ready
\* backend_LANActive_to_RelayBackoff_app_force_fallback emits: CMD_stop_monitor, CMD_cancel_pong_timeout, CMD_stop_lan_stream_reader, CMD_stop_lan_dg_reader, CMD_close_lan_path, CMD_reset_lan_ready, CMD_start_backoff_timer
\* backend_LANDegraded_to_RelayBackoff_app_force_fallback emits: CMD_stop_monitor, CMD_cancel_pong_timeout, CMD_stop_lan_stream_reader, CMD_stop_lan_dg_reader, CMD_close_lan_path, CMD_reset_lan_ready, CMD_start_backoff_timer
\* backend_RelayConnected_to_RelayConnected_app_send emits: CMD_write_active_stream
\* backend_LANOffered_to_LANOffered_app_send emits: CMD_write_active_stream
\* backend_LANActive_to_LANActive_app_send emits: CMD_write_active_stream
\* backend_LANDegraded_to_LANDegraded_app_send emits: CMD_write_active_stream
\* backend_RelayBackoff_to_RelayBackoff_app_send emits: CMD_write_active_stream
\* backend_RelayConnected_to_RelayConnected_relay_stream_data emits: CMD_deliver_recv
\* backend_LANOffered_to_LANOffered_relay_stream_data emits: CMD_deliver_recv
\* backend_LANActive_to_LANActive_relay_stream_data emits: CMD_deliver_recv
\* backend_LANDegraded_to_LANDegraded_relay_stream_data emits: CMD_deliver_recv
\* backend_RelayBackoff_to_RelayBackoff_relay_stream_data emits: CMD_deliver_recv
\* backend_RelayConnected_to_RelayConnected_relay_stream_error emits: CMD_deliver_recv_error
\* backend_LANOffered_to_LANOffered_relay_stream_error emits: CMD_deliver_recv_error
\* backend_LANActive_to_LANActive_relay_stream_error emits: CMD_deliver_recv_error
\* backend_LANDegraded_to_LANDegraded_relay_stream_error emits: CMD_deliver_recv_error
\* backend_RelayBackoff_to_RelayBackoff_relay_stream_error emits: CMD_deliver_recv_error
\* backend_RelayConnected_to_RelayConnected_app_send_datagram emits: CMD_send_active_datagram
\* backend_LANOffered_to_LANOffered_app_send_datagram emits: CMD_send_active_datagram
\* backend_LANActive_to_LANActive_app_send_datagram emits: CMD_send_active_datagram
\* backend_LANDegraded_to_LANDegraded_app_send_datagram emits: CMD_send_active_datagram
\* backend_RelayBackoff_to_RelayBackoff_app_send_datagram emits: CMD_send_active_datagram
\* backend_RelayConnected_to_RelayConnected_relay_datagram emits: CMD_deliver_recv_datagram
\* backend_LANOffered_to_LANOffered_relay_datagram emits: CMD_deliver_recv_datagram
\* backend_LANActive_to_LANActive_relay_datagram emits: CMD_deliver_recv_datagram
\* backend_LANDegraded_to_LANDegraded_relay_datagram emits: CMD_deliver_recv_datagram
\* backend_RelayBackoff_to_RelayBackoff_relay_datagram emits: CMD_deliver_recv_datagram
\* backend_LANActive_to_LANActive_lan_stream_data emits: CMD_deliver_recv
\* backend_LANDegraded_to_LANDegraded_lan_stream_data emits: CMD_deliver_recv
\* backend_LANActive_to_LANActive_lan_datagram emits: CMD_deliver_recv_datagram
\* backend_LANDegraded_to_LANDegraded_lan_datagram emits: CMD_deliver_recv_datagram
\* client_RelayConnected_to_LANConnecting_on_lan_offer_lan_enabled emits: CMD_dial_lan
\* client_LANConnecting_to_LANVerifying_lan_dial_ok emits: CMD_send_lan_verify
\* client_LANVerifying_to_LANActive_on_lan_confirm emits: CMD_start_lan_stream_reader, CMD_start_lan_dg_reader, CMD_signal_lan_ready, CMD_set_crypto_datagram
\* client_LANActive_to_LANActive_on_path_ping emits: CMD_send_path_pong
\* client_LANActive_to_RelayFallback_lan_error emits: CMD_stop_lan_stream_reader, CMD_stop_lan_dg_reader, CMD_close_lan_path, CMD_reset_lan_ready
\* client_LANActive_to_RelayFallback_lan_stream_error emits: CMD_stop_lan_stream_reader, CMD_stop_lan_dg_reader, CMD_close_lan_path, CMD_reset_lan_ready
\* client_LANActive_to_LANConnecting_on_lan_offer_lan_enabled emits: CMD_stop_lan_stream_reader, CMD_stop_lan_dg_reader, CMD_close_lan_path, CMD_dial_lan
\* client_LANVerifying_to_RelayConnected_app_force_fallback emits: CMD_stop_lan_stream_reader, CMD_stop_lan_dg_reader, CMD_close_lan_path
\* client_LANActive_to_RelayConnected_app_force_fallback emits: CMD_stop_lan_stream_reader, CMD_stop_lan_dg_reader, CMD_close_lan_path, CMD_reset_lan_ready
\* client_RelayConnected_to_RelayConnected_app_send emits: CMD_write_active_stream
\* client_LANConnecting_to_LANConnecting_app_send emits: CMD_write_active_stream
\* client_LANVerifying_to_LANVerifying_app_send emits: CMD_write_active_stream
\* client_LANActive_to_LANActive_app_send emits: CMD_write_active_stream
\* client_RelayFallback_to_RelayFallback_app_send emits: CMD_write_active_stream
\* client_RelayConnected_to_RelayConnected_relay_stream_data emits: CMD_deliver_recv
\* client_LANConnecting_to_LANConnecting_relay_stream_data emits: CMD_deliver_recv
\* client_LANVerifying_to_LANVerifying_relay_stream_data emits: CMD_deliver_recv
\* client_LANActive_to_LANActive_relay_stream_data emits: CMD_deliver_recv
\* client_RelayFallback_to_RelayFallback_relay_stream_data emits: CMD_deliver_recv
\* client_RelayConnected_to_RelayConnected_relay_stream_error emits: CMD_deliver_recv_error
\* client_LANConnecting_to_LANConnecting_relay_stream_error emits: CMD_deliver_recv_error
\* client_LANVerifying_to_LANVerifying_relay_stream_error emits: CMD_deliver_recv_error
\* client_LANActive_to_LANActive_relay_stream_error emits: CMD_deliver_recv_error
\* client_RelayFallback_to_RelayFallback_relay_stream_error emits: CMD_deliver_recv_error
\* client_RelayConnected_to_RelayConnected_app_send_datagram emits: CMD_send_active_datagram
\* client_LANConnecting_to_LANConnecting_app_send_datagram emits: CMD_send_active_datagram
\* client_LANVerifying_to_LANVerifying_app_send_datagram emits: CMD_send_active_datagram
\* client_LANActive_to_LANActive_app_send_datagram emits: CMD_send_active_datagram
\* client_RelayFallback_to_RelayFallback_app_send_datagram emits: CMD_send_active_datagram
\* client_RelayConnected_to_RelayConnected_relay_datagram emits: CMD_deliver_recv_datagram
\* client_LANConnecting_to_LANConnecting_relay_datagram emits: CMD_deliver_recv_datagram
\* client_LANVerifying_to_LANVerifying_relay_datagram emits: CMD_deliver_recv_datagram
\* client_LANActive_to_LANActive_relay_datagram emits: CMD_deliver_recv_datagram
\* client_RelayFallback_to_RelayFallback_relay_datagram emits: CMD_deliver_recv_datagram
\* client_LANActive_to_LANActive_lan_stream_data emits: CMD_deliver_recv
\* client_LANActive_to_LANActive_lan_datagram emits: CMD_deliver_recv_datagram

====
