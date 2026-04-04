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

\* States for relay
relay_Idle == "relay_Idle"
relay_BackendRegistered == "relay_BackendRegistered"
relay_Bridged == "relay_Bridged"

\* Message types
MSG_pair_hello == "pair_hello"
MSG_pair_hello_ack == "pair_hello_ack"
MSG_pair_confirm == "pair_confirm"
MSG_pair_complete == "pair_complete"
MSG_auth_request == "auth_request"
MSG_auth_ok == "auth_ok"
MSG_lan_offer == "lan_offer"
MSG_lan_verify == "lan_verify"
MSG_lan_confirm == "lan_confirm"
MSG_path_ping == "path_ping"
MSG_path_pong == "path_pong"

\* deterministic ordering for ECDH
KeyRank(k) == CASE k = "adv_pub" -> 0 [] k = "client_pub" -> 1 [] k = "backend_pub" -> 2 [] OTHER -> 3
\* symbolic ECDH
DeriveKey(a, b) == IF KeyRank(a) <= KeyRank(b) THEN <<"ecdh", a, b>> ELSE <<"ecdh", b, a>>
\* confirmation code from pubkeys
DeriveCode(a, b) == IF KeyRank(a) <= KeyRank(b) THEN <<"code", a, b>> ELSE <<"code", b, a>>
\* minimum of two values
Min(a, b) == IF a < b THEN a ELSE b

CONSTANTS MaxChanLen

VARIABLES
    backend_state,
    client_state,
    relay_state,
    chan_client_backend,
    chan_backend_client,
    recv_msg,
    lan_addr,
    challenge_bytes,
    offer_challenge,
    instance_id,
    ping_failures,
    max_ping_failures,
    backoff_level,
    max_backoff_level,
    lan_server_addr,
    b_active_path,
    c_active_path,
    b_dispatcher_path,
    c_dispatcher_path,
    monitor_target,
    lan_signal,
    relay_bridge

vars == <<backend_state, client_state, relay_state, chan_client_backend, chan_backend_client, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, ping_failures, max_ping_failures, backoff_level, max_backoff_level, lan_server_addr, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, relay_bridge>>

CanSend(ch) == Len(ch) < MaxChanLen

Init ==
    /\ backend_state = backend_RelayConnected
    /\ client_state = client_RelayConnected
    /\ relay_state = relay_BackendRegistered
    /\ chan_client_backend = <<>>
    /\ chan_backend_client = <<>>
    /\ recv_msg = [type |-> "none"]
    /\ lan_addr = "none"
    /\ challenge_bytes = "none"
    /\ offer_challenge = "none"
    /\ instance_id = "none"
    /\ ping_failures = 0
    /\ max_ping_failures = 3
    /\ backoff_level = 0
    /\ max_backoff_level = 5
    /\ lan_server_addr = "none"
    /\ b_active_path = "relay"
    /\ c_active_path = "relay"
    /\ b_dispatcher_path = "relay"
    /\ c_dispatcher_path = "relay"
    /\ monitor_target = "none"
    /\ lan_signal = "pending"
    /\ relay_bridge = "idle"

\* ================================================================
\* backend actions
\* ================================================================

\* backend: SessionActive -> RelayConnected (session_established)
backend_SessionActive_to_RelayConnected_on_session_established ==
    /\ backend_state = backend_SessionActive
    /\ backend_state' = backend_RelayConnected
    /\ UNCHANGED <<client_state, relay_state, chan_client_backend, chan_backend_client, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, ping_failures, max_ping_failures, backoff_level, max_backoff_level, lan_server_addr, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, relay_bridge>>

\* backend: RelayConnected -> LANOffered (lan_server_ready)
backend_RelayConnected_to_LANOffered_on_lan_server_ready ==
    /\ backend_state = backend_RelayConnected
    /\ chan_backend_client' = Append(chan_backend_client, [type |-> MSG_lan_offer, addr |-> lan_addr, challenge |-> challenge_bytes])
    /\ backend_state' = backend_LANOffered
    /\ UNCHANGED <<client_state, relay_state, chan_client_backend, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, ping_failures, max_ping_failures, backoff_level, max_backoff_level, lan_server_addr, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, relay_bridge>>

\* backend: LANOffered -> LANActive on recv lan_verify [challenge_valid]
backend_LANOffered_to_LANActive_on_lan_verify_challenge_valid ==
    /\ backend_state = backend_LANOffered
    /\ Len(chan_client_backend) > 0
    /\ Head(chan_client_backend).type = MSG_lan_verify
    /\ offer_challenge = challenge_bytes
    /\ chan_client_backend' = Tail(chan_client_backend)
    /\ chan_backend_client' = Append(chan_backend_client, [type |-> MSG_lan_confirm])
    /\ backend_state' = backend_LANActive
    /\ ping_failures' = 0
    /\ backoff_level' = 0
    /\ b_active_path' = "lan"
    /\ b_dispatcher_path' = "lan"
    /\ monitor_target' = "lan"
    /\ lan_signal' = "ready"
    /\ UNCHANGED <<client_state, relay_state, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, max_ping_failures, max_backoff_level, lan_server_addr, c_active_path, c_dispatcher_path, relay_bridge>>

\* backend: LANOffered -> RelayConnected on recv lan_verify [challenge_invalid]
backend_LANOffered_to_RelayConnected_on_lan_verify_challenge_invalid ==
    /\ backend_state = backend_LANOffered
    /\ Len(chan_client_backend) > 0
    /\ Head(chan_client_backend).type = MSG_lan_verify
    /\ offer_challenge /= challenge_bytes
    /\ chan_client_backend' = Tail(chan_client_backend)
    /\ backend_state' = backend_RelayConnected
    /\ UNCHANGED <<client_state, relay_state, chan_backend_client, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, ping_failures, max_ping_failures, backoff_level, max_backoff_level, lan_server_addr, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, relay_bridge>>

\* backend: LANOffered -> RelayBackoff (offer_timeout)
backend_LANOffered_to_RelayBackoff_on_offer_timeout ==
    /\ backend_state = backend_LANOffered
    /\ backend_state' = backend_RelayBackoff
    /\ backoff_level' = Min(backoff_level + 1, max_backoff_level)
    /\ lan_signal' = "pending"
    /\ UNCHANGED <<client_state, relay_state, chan_client_backend, chan_backend_client, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, ping_failures, max_ping_failures, max_backoff_level, lan_server_addr, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, relay_bridge>>

\* backend: LANActive -> LANActive (ping_tick)
backend_LANActive_to_LANActive_on_ping_tick ==
    /\ backend_state = backend_LANActive
    /\ chan_backend_client' = Append(chan_backend_client, [type |-> MSG_path_ping])
    /\ backend_state' = backend_LANActive
    /\ UNCHANGED <<client_state, relay_state, chan_client_backend, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, ping_failures, max_ping_failures, backoff_level, max_backoff_level, lan_server_addr, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, relay_bridge>>

\* backend: LANActive -> LANDegraded (ping_timeout)
backend_LANActive_to_LANDegraded_on_ping_timeout ==
    /\ backend_state = backend_LANActive
    /\ backend_state' = backend_LANDegraded
    /\ ping_failures' = 1
    /\ UNCHANGED <<client_state, relay_state, chan_client_backend, chan_backend_client, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, max_ping_failures, backoff_level, max_backoff_level, lan_server_addr, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, relay_bridge>>

\* backend: LANDegraded -> LANDegraded (ping_tick)
backend_LANDegraded_to_LANDegraded_on_ping_tick ==
    /\ backend_state = backend_LANDegraded
    /\ chan_backend_client' = Append(chan_backend_client, [type |-> MSG_path_ping])
    /\ backend_state' = backend_LANDegraded
    /\ UNCHANGED <<client_state, relay_state, chan_client_backend, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, ping_failures, max_ping_failures, backoff_level, max_backoff_level, lan_server_addr, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, relay_bridge>>

\* backend: LANDegraded -> LANActive on recv path_pong
backend_LANDegraded_to_LANActive_on_path_pong ==
    /\ backend_state = backend_LANDegraded
    /\ Len(chan_client_backend) > 0
    /\ Head(chan_client_backend).type = MSG_path_pong
    /\ chan_client_backend' = Tail(chan_client_backend)
    /\ backend_state' = backend_LANActive
    /\ ping_failures' = 0
    /\ UNCHANGED <<client_state, relay_state, chan_backend_client, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, max_ping_failures, backoff_level, max_backoff_level, lan_server_addr, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, relay_bridge>>

\* backend: LANDegraded -> LANDegraded (ping_timeout) [under_max_failures]
backend_LANDegraded_to_LANDegraded_on_ping_timeout_under_max_failures ==
    /\ backend_state = backend_LANDegraded
    /\ ping_failures + 1 < max_ping_failures
    /\ backend_state' = backend_LANDegraded
    /\ ping_failures' = ping_failures + 1
    /\ UNCHANGED <<client_state, relay_state, chan_client_backend, chan_backend_client, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, max_ping_failures, backoff_level, max_backoff_level, lan_server_addr, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, relay_bridge>>

\* backend: LANDegraded -> RelayBackoff (ping_timeout) [at_max_failures]
backend_LANDegraded_to_RelayBackoff_on_ping_timeout_at_max_failures ==
    /\ backend_state = backend_LANDegraded
    /\ ping_failures + 1 >= max_ping_failures
    /\ backend_state' = backend_RelayBackoff
    /\ backoff_level' = Min(backoff_level + 1, max_backoff_level)
    /\ b_active_path' = "relay"
    /\ b_dispatcher_path' = "relay"
    /\ monitor_target' = "none"
    /\ lan_signal' = "pending"
    /\ ping_failures' = 0
    /\ UNCHANGED <<client_state, relay_state, chan_client_backend, chan_backend_client, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, max_ping_failures, max_backoff_level, lan_server_addr, c_active_path, c_dispatcher_path, relay_bridge>>

\* backend: RelayBackoff -> LANOffered (backoff_expired)
backend_RelayBackoff_to_LANOffered_on_backoff_expired ==
    /\ backend_state = backend_RelayBackoff
    /\ chan_backend_client' = Append(chan_backend_client, [type |-> MSG_lan_offer, addr |-> lan_addr, challenge |-> challenge_bytes])
    /\ backend_state' = backend_LANOffered
    /\ UNCHANGED <<client_state, relay_state, chan_client_backend, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, ping_failures, max_ping_failures, backoff_level, max_backoff_level, lan_server_addr, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, relay_bridge>>

\* backend: RelayBackoff -> LANOffered (lan_server_changed)
backend_RelayBackoff_to_LANOffered_on_lan_server_changed ==
    /\ backend_state = backend_RelayBackoff
    /\ chan_backend_client' = Append(chan_backend_client, [type |-> MSG_lan_offer, addr |-> lan_addr, challenge |-> challenge_bytes])
    /\ backend_state' = backend_LANOffered
    /\ backoff_level' = 0
    /\ UNCHANGED <<client_state, relay_state, chan_client_backend, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, ping_failures, max_ping_failures, max_backoff_level, lan_server_addr, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, relay_bridge>>

\* backend: RelayConnected -> LANOffered (readvertise_tick) [lan_server_available]
backend_RelayConnected_to_LANOffered_on_readvertise_tick_lan_server_available ==
    /\ backend_state = backend_RelayConnected
    /\ lan_server_addr /= "none"
    /\ chan_backend_client' = Append(chan_backend_client, [type |-> MSG_lan_offer, addr |-> lan_addr, challenge |-> challenge_bytes])
    /\ backend_state' = backend_LANOffered
    /\ UNCHANGED <<client_state, relay_state, chan_client_backend, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, ping_failures, max_ping_failures, backoff_level, max_backoff_level, lan_server_addr, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, relay_bridge>>

\* backend: RelayConnected -> Paired (disconnect)
backend_RelayConnected_to_Paired_on_disconnect ==
    /\ backend_state = backend_RelayConnected
    /\ backend_state' = backend_Paired
    /\ UNCHANGED <<client_state, relay_state, chan_client_backend, chan_backend_client, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, ping_failures, max_ping_failures, backoff_level, max_backoff_level, lan_server_addr, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, relay_bridge>>

\* ================================================================
\* client actions
\* ================================================================

\* client: SessionActive -> RelayConnected (session_established)
client_SessionActive_to_RelayConnected_on_session_established ==
    /\ client_state = client_SessionActive
    /\ client_state' = client_RelayConnected
    /\ UNCHANGED <<backend_state, relay_state, chan_client_backend, chan_backend_client, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, ping_failures, max_ping_failures, backoff_level, max_backoff_level, lan_server_addr, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, relay_bridge>>

\* client: RelayConnected -> LANConnecting on recv lan_offer [lan_enabled]
client_RelayConnected_to_LANConnecting_on_lan_offer_lan_enabled ==
    /\ client_state = client_RelayConnected
    /\ Len(chan_backend_client) > 0
    /\ Head(chan_backend_client).type = MSG_lan_offer
    /\ TRUE
    /\ chan_backend_client' = Tail(chan_backend_client)
    /\ client_state' = client_LANConnecting
    /\ UNCHANGED <<backend_state, relay_state, chan_client_backend, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, ping_failures, max_ping_failures, backoff_level, max_backoff_level, lan_server_addr, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, relay_bridge>>

\* client: RelayConnected -> RelayConnected on recv lan_offer [lan_disabled]
client_RelayConnected_to_RelayConnected_on_lan_offer_lan_disabled ==
    /\ client_state = client_RelayConnected
    /\ Len(chan_backend_client) > 0
    /\ Head(chan_backend_client).type = MSG_lan_offer
    /\ FALSE
    /\ chan_backend_client' = Tail(chan_backend_client)
    /\ client_state' = client_RelayConnected
    /\ UNCHANGED <<backend_state, relay_state, chan_client_backend, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, ping_failures, max_ping_failures, backoff_level, max_backoff_level, lan_server_addr, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, relay_bridge>>

\* client: LANConnecting -> LANVerifying (lan_dial_ok)
client_LANConnecting_to_LANVerifying_on_lan_dial_ok ==
    /\ client_state = client_LANConnecting
    /\ chan_client_backend' = Append(chan_client_backend, [type |-> MSG_lan_verify, challenge |-> offer_challenge, instance_id |-> instance_id])
    /\ client_state' = client_LANVerifying
    /\ UNCHANGED <<backend_state, relay_state, chan_backend_client, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, ping_failures, max_ping_failures, backoff_level, max_backoff_level, lan_server_addr, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, relay_bridge>>

\* client: LANConnecting -> RelayConnected (lan_dial_failed)
client_LANConnecting_to_RelayConnected_on_lan_dial_failed ==
    /\ client_state = client_LANConnecting
    /\ client_state' = client_RelayConnected
    /\ UNCHANGED <<backend_state, relay_state, chan_client_backend, chan_backend_client, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, ping_failures, max_ping_failures, backoff_level, max_backoff_level, lan_server_addr, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, relay_bridge>>

\* client: LANVerifying -> LANActive on recv lan_confirm
client_LANVerifying_to_LANActive_on_lan_confirm ==
    /\ client_state = client_LANVerifying
    /\ Len(chan_backend_client) > 0
    /\ Head(chan_backend_client).type = MSG_lan_confirm
    /\ chan_backend_client' = Tail(chan_backend_client)
    /\ client_state' = client_LANActive
    /\ c_active_path' = "lan"
    /\ c_dispatcher_path' = "lan"
    /\ lan_signal' = "ready"
    /\ UNCHANGED <<backend_state, relay_state, chan_client_backend, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, ping_failures, max_ping_failures, backoff_level, max_backoff_level, lan_server_addr, b_active_path, b_dispatcher_path, monitor_target, relay_bridge>>

\* client: LANVerifying -> RelayConnected (verify_timeout)
client_LANVerifying_to_RelayConnected_on_verify_timeout ==
    /\ client_state = client_LANVerifying
    /\ client_state' = client_RelayConnected
    /\ c_dispatcher_path' = "relay"
    /\ UNCHANGED <<backend_state, relay_state, chan_client_backend, chan_backend_client, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, ping_failures, max_ping_failures, backoff_level, max_backoff_level, lan_server_addr, b_active_path, c_active_path, b_dispatcher_path, monitor_target, lan_signal, relay_bridge>>

\* client: LANActive -> LANActive on recv path_ping
client_LANActive_to_LANActive_on_path_ping ==
    /\ client_state = client_LANActive
    /\ Len(chan_backend_client) > 0
    /\ Head(chan_backend_client).type = MSG_path_ping
    /\ chan_backend_client' = Tail(chan_backend_client)
    /\ chan_client_backend' = Append(chan_client_backend, [type |-> MSG_path_pong])
    /\ client_state' = client_LANActive
    /\ UNCHANGED <<backend_state, relay_state, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, ping_failures, max_ping_failures, backoff_level, max_backoff_level, lan_server_addr, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, relay_bridge>>

\* client: LANActive -> RelayFallback (lan_error)
client_LANActive_to_RelayFallback_on_lan_error ==
    /\ client_state = client_LANActive
    /\ client_state' = client_RelayFallback
    /\ c_active_path' = "relay"
    /\ c_dispatcher_path' = "relay"
    /\ lan_signal' = "pending"
    /\ UNCHANGED <<backend_state, relay_state, chan_client_backend, chan_backend_client, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, ping_failures, max_ping_failures, backoff_level, max_backoff_level, lan_server_addr, b_active_path, b_dispatcher_path, monitor_target, relay_bridge>>

\* client: RelayFallback -> RelayConnected (relay_ok)
client_RelayFallback_to_RelayConnected_on_relay_ok ==
    /\ client_state = client_RelayFallback
    /\ client_state' = client_RelayConnected
    /\ UNCHANGED <<backend_state, relay_state, chan_client_backend, chan_backend_client, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, ping_failures, max_ping_failures, backoff_level, max_backoff_level, lan_server_addr, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, relay_bridge>>

\* client: LANActive -> LANConnecting on recv lan_offer [lan_enabled]
client_LANActive_to_LANConnecting_on_lan_offer_lan_enabled ==
    /\ client_state = client_LANActive
    /\ Len(chan_backend_client) > 0
    /\ Head(chan_backend_client).type = MSG_lan_offer
    /\ TRUE
    /\ chan_backend_client' = Tail(chan_backend_client)
    /\ client_state' = client_LANConnecting
    /\ UNCHANGED <<backend_state, relay_state, chan_client_backend, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, ping_failures, max_ping_failures, backoff_level, max_backoff_level, lan_server_addr, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, relay_bridge>>

\* client: RelayConnected -> Paired (disconnect)
client_RelayConnected_to_Paired_on_disconnect ==
    /\ client_state = client_RelayConnected
    /\ client_state' = client_Paired
    /\ UNCHANGED <<backend_state, relay_state, chan_client_backend, chan_backend_client, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, ping_failures, max_ping_failures, backoff_level, max_backoff_level, lan_server_addr, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, relay_bridge>>

\* ================================================================
\* relay actions
\* ================================================================

\* relay: Idle -> BackendRegistered (backend_register)
relay_Idle_to_BackendRegistered_on_backend_register ==
    /\ relay_state = relay_Idle
    /\ relay_state' = relay_BackendRegistered
    /\ UNCHANGED <<backend_state, client_state, chan_client_backend, chan_backend_client, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, ping_failures, max_ping_failures, backoff_level, max_backoff_level, lan_server_addr, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, relay_bridge>>

\* relay: BackendRegistered -> Bridged (client_connect)
relay_BackendRegistered_to_Bridged_on_client_connect ==
    /\ relay_state = relay_BackendRegistered
    /\ relay_state' = relay_Bridged
    /\ relay_bridge' = "active"
    /\ UNCHANGED <<backend_state, client_state, chan_client_backend, chan_backend_client, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, ping_failures, max_ping_failures, backoff_level, max_backoff_level, lan_server_addr, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal>>

\* relay: Bridged -> BackendRegistered (client_disconnect)
relay_Bridged_to_BackendRegistered_on_client_disconnect ==
    /\ relay_state = relay_Bridged
    /\ relay_state' = relay_BackendRegistered
    /\ relay_bridge' = "idle"
    /\ UNCHANGED <<backend_state, client_state, chan_client_backend, chan_backend_client, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, ping_failures, max_ping_failures, backoff_level, max_backoff_level, lan_server_addr, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal>>

\* relay: BackendRegistered -> Idle (backend_disconnect)
relay_BackendRegistered_to_Idle_on_backend_disconnect ==
    /\ relay_state = relay_BackendRegistered
    /\ relay_state' = relay_Idle
    /\ UNCHANGED <<backend_state, client_state, chan_client_backend, chan_backend_client, recv_msg, lan_addr, challenge_bytes, offer_challenge, instance_id, ping_failures, max_ping_failures, backoff_level, max_backoff_level, lan_server_addr, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, relay_bridge>>

Next ==
    \/ backend_SessionActive_to_RelayConnected_on_session_established
    \/ backend_RelayConnected_to_LANOffered_on_lan_server_ready
    \/ backend_LANOffered_to_LANActive_on_lan_verify_challenge_valid
    \/ backend_LANOffered_to_RelayConnected_on_lan_verify_challenge_invalid
    \/ backend_LANOffered_to_RelayBackoff_on_offer_timeout
    \/ backend_LANActive_to_LANActive_on_ping_tick
    \/ backend_LANActive_to_LANDegraded_on_ping_timeout
    \/ backend_LANDegraded_to_LANDegraded_on_ping_tick
    \/ backend_LANDegraded_to_LANActive_on_path_pong
    \/ backend_LANDegraded_to_LANDegraded_on_ping_timeout_under_max_failures
    \/ backend_LANDegraded_to_RelayBackoff_on_ping_timeout_at_max_failures
    \/ backend_RelayBackoff_to_LANOffered_on_backoff_expired
    \/ backend_RelayBackoff_to_LANOffered_on_lan_server_changed
    \/ backend_RelayConnected_to_LANOffered_on_readvertise_tick_lan_server_available
    \/ backend_RelayConnected_to_Paired_on_disconnect
    \/ client_SessionActive_to_RelayConnected_on_session_established
    \/ client_RelayConnected_to_LANConnecting_on_lan_offer_lan_enabled
    \/ client_RelayConnected_to_RelayConnected_on_lan_offer_lan_disabled
    \/ client_LANConnecting_to_LANVerifying_on_lan_dial_ok
    \/ client_LANConnecting_to_RelayConnected_on_lan_dial_failed
    \/ client_LANVerifying_to_LANActive_on_lan_confirm
    \/ client_LANVerifying_to_RelayConnected_on_verify_timeout
    \/ client_LANActive_to_LANActive_on_path_ping
    \/ client_LANActive_to_RelayFallback_on_lan_error
    \/ client_RelayFallback_to_RelayConnected_on_relay_ok
    \/ client_LANActive_to_LANConnecting_on_lan_offer_lan_enabled
    \/ client_RelayConnected_to_Paired_on_disconnect
    \/ relay_Idle_to_BackendRegistered_on_backend_register
    \/ relay_BackendRegistered_to_Bridged_on_client_connect
    \/ relay_Bridged_to_BackendRegistered_on_client_disconnect
    \/ relay_BackendRegistered_to_Idle_on_backend_disconnect

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

====
