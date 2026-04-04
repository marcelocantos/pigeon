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

\* backend: LANActive -> LANActive (ping_tick)
backend_LANActive_to_LANActive_ping_tick ==
    /\ backend_state = backend_LANActive
    /\ received_path_ping' = [type |-> MSG_path_ping]
    /\ backend_state' = backend_LANActive
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm>>

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

\* backend: LANDegraded -> LANActive on recv path_pong
backend_LANDegraded_to_LANActive_on_path_pong ==
    /\ backend_state = backend_LANDegraded
    /\ received_path_pong.type = MSG_path_pong
    /\ received_path_pong' = [type |-> "none"]
    /\ backend_state' = backend_LANActive
    /\ ping_failures' = 0
    /\ UNCHANGED <<client_state, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_lan_offer, received_lan_confirm, received_path_ping>>

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

\* backend: RelayBackoff -> LANOffered (backoff_expired)
backend_RelayBackoff_to_LANOffered_backoff_expired ==
    /\ backend_state = backend_RelayBackoff
    /\ received_lan_offer' = [type |-> MSG_lan_offer, addr |-> lan_addr, challenge |-> challenge_bytes]
    /\ backend_state' = backend_LANOffered
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_confirm, received_path_ping>>

\* backend: RelayBackoff -> LANOffered (lan_server_changed)
backend_RelayBackoff_to_LANOffered_lan_server_changed ==
    /\ backend_state = backend_RelayBackoff
    /\ received_lan_offer' = [type |-> MSG_lan_offer, addr |-> lan_addr, challenge |-> challenge_bytes]
    /\ backend_state' = backend_LANOffered
    /\ backoff_level' = 0
    /\ UNCHANGED <<client_state, ping_failures, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_confirm, received_path_ping>>

\* backend: RelayConnected -> LANOffered (readvertise_tick) [lan_server_available]
backend_RelayConnected_to_LANOffered_readvertise_tick_lan_server_available ==
    /\ backend_state = backend_RelayConnected
    /\ lan_server_addr /= "none"
    /\ received_lan_offer' = [type |-> MSG_lan_offer, addr |-> lan_addr, challenge |-> challenge_bytes]
    /\ backend_state' = backend_LANOffered
    /\ UNCHANGED <<client_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_confirm, received_path_ping>>


\* client: RelayConnected -> LANConnecting on recv lan_offer [lan_enabled]
client_RelayConnected_to_LANConnecting_on_lan_offer_lan_enabled ==
    /\ client_state = client_RelayConnected
    /\ received_lan_offer.type = MSG_lan_offer
    /\ TRUE
    /\ received_lan_offer' = [type |-> "none"]
    /\ client_state' = client_LANConnecting
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, c_active_path, b_dispatcher_path, c_dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_confirm, received_path_ping>>

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

\* client: LANActive -> RelayFallback (lan_error)
client_LANActive_to_RelayFallback_lan_error ==
    /\ client_state = client_LANActive
    /\ client_state' = client_RelayFallback
    /\ c_active_path' = "relay"
    /\ c_dispatcher_path' = "relay"
    /\ lan_signal' = "pending"
    /\ UNCHANGED <<backend_state, ping_failures, backoff_level, b_active_path, b_dispatcher_path, monitor_target, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

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


Next ==
    \/ backend_RelayConnected_to_LANOffered_lan_server_ready
    \/ backend_LANOffered_to_LANActive_on_lan_verify_challenge_valid
    \/ backend_LANOffered_to_RelayConnected_on_lan_verify_challenge_invalid
    \/ backend_LANOffered_to_RelayBackoff_offer_timeout
    \/ backend_LANActive_to_LANActive_ping_tick
    \/ backend_LANActive_to_LANDegraded_ping_timeout
    \/ backend_LANDegraded_to_LANDegraded_ping_tick
    \/ backend_LANDegraded_to_LANActive_on_path_pong
    \/ backend_LANDegraded_to_LANDegraded_ping_timeout_under_max_failures
    \/ backend_LANDegraded_to_RelayBackoff_ping_timeout_at_max_failures
    \/ backend_RelayBackoff_to_LANOffered_backoff_expired
    \/ backend_RelayBackoff_to_LANOffered_lan_server_changed
    \/ backend_RelayConnected_to_LANOffered_readvertise_tick_lan_server_available
    \/ client_RelayConnected_to_LANConnecting_on_lan_offer_lan_enabled
    \/ client_RelayConnected_to_RelayConnected_on_lan_offer_lan_disabled
    \/ client_LANConnecting_to_LANVerifying_lan_dial_ok
    \/ client_LANConnecting_to_RelayConnected_lan_dial_failed
    \/ client_LANVerifying_to_LANActive_on_lan_confirm
    \/ client_LANVerifying_to_RelayConnected_verify_timeout
    \/ client_LANActive_to_LANActive_on_path_ping
    \/ client_LANActive_to_RelayFallback_lan_error
    \/ client_RelayFallback_to_RelayConnected_relay_ok
    \/ client_LANActive_to_LANConnecting_on_lan_offer_lan_enabled

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
