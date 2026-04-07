---- MODULE PathSwitch ----
\* Auto-generated from protocol YAML. Do not edit.

EXTENDS Integers, Sequences, FiniteSets, TLC

\* States for backend
backend_RelayConnected == "backend_RelayConnected"
backend_LANOffered == "backend_LANOffered"
backend_LANActive == "backend_LANActive"
backend_RelayBackoff == "backend_RelayBackoff"
backend_LANDegraded == "backend_LANDegraded"

\* States for client
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
MSG_lan_offer == "lan_offer"
MSG_lan_verify == "lan_verify"
MSG_lan_confirm == "lan_confirm"
MSG_path_ping == "path_ping"
MSG_path_pong == "path_pong"
MSG_relay_resume == "relay_resume"
MSG_relay_resumed == "relay_resumed"

\* Event types
EVT_backend_disconnect == "backend_disconnect"
EVT_backend_register == "backend_register"
EVT_backoff_expired == "backoff_expired"
EVT_client_connect == "client_connect"
EVT_client_disconnect == "client_disconnect"
EVT_lan_dial_failed == "lan_dial_failed"
EVT_lan_dial_ok == "lan_dial_ok"
EVT_lan_error == "lan_error"
EVT_lan_server_changed == "lan_server_changed"
EVT_lan_server_ready == "lan_server_ready"
EVT_offer_timeout == "offer_timeout"
EVT_ping_tick == "ping_tick"
EVT_ping_timeout == "ping_timeout"
EVT_readvertise_tick == "readvertise_tick"
EVT_recv_lan_confirm == "recv_lan_confirm"
EVT_recv_lan_offer == "recv_lan_offer"
EVT_recv_lan_verify == "recv_lan_verify"
EVT_recv_path_ping == "recv_path_ping"
EVT_recv_path_pong == "recv_path_pong"
EVT_recv_relay_resume == "recv_relay_resume"
EVT_relay_ok == "relay_ok"
EVT_verify_timeout == "verify_timeout"



CONSTANTS lan_addr, challenge_bytes, offer_challenge, instance_id, max_ping_failures, max_backoff_level, lan_server_addr

VARIABLES
    backend_state,
    client_state,
    relay_state,
    ping_failures,
    backoff_level,
    active_path,
    dispatcher_path,
    monitor_target,
    lan_signal,
    relay_bridge,
    received_lan_verify,
    received_path_pong,
    received_lan_offer,
    received_lan_confirm,
    received_path_ping,
    received_relay_resume

vars == <<backend_state, client_state, relay_state, ping_failures, backoff_level, active_path, dispatcher_path, monitor_target, lan_signal, relay_bridge, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping, received_relay_resume>>

Init ==
    /\ backend_state = backend_RelayConnected
    /\ client_state = client_RelayConnected
    /\ relay_state = relay_Idle
    /\ ping_failures = 0
    /\ backoff_level = 0
    /\ active_path = "relay"
    /\ dispatcher_path = "relay"
    /\ monitor_target = "none"
    /\ lan_signal = "pending"
    /\ relay_bridge = "idle"
    /\ received_lan_verify = [type |-> "none"]
    /\ received_path_pong = [type |-> "none"]
    /\ received_lan_offer = [type |-> "none"]
    /\ received_lan_confirm = [type |-> "none"]
    /\ received_path_ping = [type |-> "none"]
    /\ received_relay_resume = [type |-> "none"]

\* backend: RelayConnected -> LANOffered (lan_server_ready)
backend_RelayConnected_to_LANOffered_lan_server_ready ==
    /\ backend_state = backend_RelayConnected
    /\ received_lan_offer' = [type |-> MSG_lan_offer, addr |-> lan_addr, challenge |-> challenge_bytes]
    /\ backend_state' = backend_LANOffered
    /\ UNCHANGED <<client_state, relay_state, ping_failures, backoff_level, active_path, dispatcher_path, monitor_target, lan_signal, relay_bridge, received_lan_verify, received_path_pong, received_lan_confirm, received_path_ping, received_relay_resume>>

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
    /\ active_path' = "lan"
    /\ monitor_target' = "lan"
    /\ dispatcher_path' = "lan"
    /\ lan_signal' = "ready"
    /\ UNCHANGED <<client_state, relay_state, relay_bridge, received_path_pong, received_lan_offer, received_path_ping, received_relay_resume>>

\* backend: LANOffered -> RelayConnected on recv lan_verify [challenge_invalid]
backend_LANOffered_to_RelayConnected_on_lan_verify_challenge_invalid ==
    /\ backend_state = backend_LANOffered
    /\ received_lan_verify.type = MSG_lan_verify
    /\ offer_challenge /= challenge_bytes
    /\ received_lan_verify' = [type |-> "none"]
    /\ backend_state' = backend_RelayConnected
    /\ UNCHANGED <<client_state, relay_state, ping_failures, backoff_level, active_path, dispatcher_path, monitor_target, lan_signal, relay_bridge, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping, received_relay_resume>>

\* backend: LANOffered -> RelayBackoff (offer_timeout)
backend_LANOffered_to_RelayBackoff_offer_timeout ==
    /\ backend_state = backend_LANOffered
    /\ backend_state' = backend_RelayBackoff
    /\ backoff_level' = Min(backoff_level + 1, max_backoff_level)
    /\ UNCHANGED <<client_state, relay_state, ping_failures, active_path, dispatcher_path, monitor_target, lan_signal, relay_bridge, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping, received_relay_resume>>

\* backend: LANActive -> LANActive (ping_tick)
backend_LANActive_to_LANActive_ping_tick ==
    /\ backend_state = backend_LANActive
    /\ received_path_ping' = [type |-> MSG_path_ping]
    /\ backend_state' = backend_LANActive
    /\ UNCHANGED <<client_state, relay_state, ping_failures, backoff_level, active_path, dispatcher_path, monitor_target, lan_signal, relay_bridge, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_relay_resume>>

\* backend: LANActive -> LANDegraded (ping_timeout)
backend_LANActive_to_LANDegraded_ping_timeout ==
    /\ backend_state = backend_LANActive
    /\ backend_state' = backend_LANDegraded
    /\ ping_failures' = 1
    /\ UNCHANGED <<client_state, relay_state, backoff_level, active_path, dispatcher_path, monitor_target, lan_signal, relay_bridge, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping, received_relay_resume>>

\* backend: LANDegraded -> LANDegraded (ping_tick)
backend_LANDegraded_to_LANDegraded_ping_tick ==
    /\ backend_state = backend_LANDegraded
    /\ received_path_ping' = [type |-> MSG_path_ping]
    /\ backend_state' = backend_LANDegraded
    /\ UNCHANGED <<client_state, relay_state, ping_failures, backoff_level, active_path, dispatcher_path, monitor_target, lan_signal, relay_bridge, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_relay_resume>>

\* backend: LANDegraded -> LANActive on recv path_pong
backend_LANDegraded_to_LANActive_on_path_pong ==
    /\ backend_state = backend_LANDegraded
    /\ received_path_pong.type = MSG_path_pong
    /\ received_path_pong' = [type |-> "none"]
    /\ backend_state' = backend_LANActive
    /\ ping_failures' = 0
    /\ UNCHANGED <<client_state, relay_state, backoff_level, active_path, dispatcher_path, monitor_target, lan_signal, relay_bridge, received_lan_verify, received_lan_offer, received_lan_confirm, received_path_ping, received_relay_resume>>

\* backend: LANDegraded -> LANDegraded (ping_timeout) [under_max_failures]
backend_LANDegraded_to_LANDegraded_ping_timeout_under_max_failures ==
    /\ backend_state = backend_LANDegraded
    /\ ping_failures + 1 < max_ping_failures
    /\ backend_state' = backend_LANDegraded
    /\ ping_failures' = ping_failures + 1
    /\ UNCHANGED <<client_state, relay_state, backoff_level, active_path, dispatcher_path, monitor_target, lan_signal, relay_bridge, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping, received_relay_resume>>

\* backend: LANDegraded -> RelayBackoff (ping_timeout) [at_max_failures]
backend_LANDegraded_to_RelayBackoff_ping_timeout_at_max_failures ==
    /\ backend_state = backend_LANDegraded
    /\ ping_failures + 1 >= max_ping_failures
    /\ backend_state' = backend_RelayBackoff
    /\ backoff_level' = Min(backoff_level + 1, max_backoff_level)
    /\ active_path' = "relay"
    /\ monitor_target' = "none"
    /\ dispatcher_path' = "relay"
    /\ lan_signal' = "pending"
    /\ ping_failures' = 0
    /\ UNCHANGED <<client_state, relay_state, relay_bridge, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping, received_relay_resume>>

\* backend: RelayBackoff -> LANOffered (backoff_expired)
backend_RelayBackoff_to_LANOffered_backoff_expired ==
    /\ backend_state = backend_RelayBackoff
    /\ received_lan_offer' = [type |-> MSG_lan_offer, addr |-> lan_addr, challenge |-> challenge_bytes]
    /\ backend_state' = backend_LANOffered
    /\ UNCHANGED <<client_state, relay_state, ping_failures, backoff_level, active_path, dispatcher_path, monitor_target, lan_signal, relay_bridge, received_lan_verify, received_path_pong, received_lan_confirm, received_path_ping, received_relay_resume>>

\* backend: RelayBackoff -> LANOffered (lan_server_changed)
backend_RelayBackoff_to_LANOffered_lan_server_changed ==
    /\ backend_state = backend_RelayBackoff
    /\ received_lan_offer' = [type |-> MSG_lan_offer, addr |-> lan_addr, challenge |-> challenge_bytes]
    /\ backend_state' = backend_LANOffered
    /\ backoff_level' = 0
    /\ UNCHANGED <<client_state, relay_state, ping_failures, active_path, dispatcher_path, monitor_target, lan_signal, relay_bridge, received_lan_verify, received_path_pong, received_lan_confirm, received_path_ping, received_relay_resume>>

\* backend: RelayConnected -> LANOffered (readvertise_tick) [lan_server_available]
backend_RelayConnected_to_LANOffered_readvertise_tick_lan_server_available ==
    /\ backend_state = backend_RelayConnected
    /\ lan_server_addr /= "none"
    /\ received_lan_offer' = [type |-> MSG_lan_offer, addr |-> lan_addr, challenge |-> challenge_bytes]
    /\ backend_state' = backend_LANOffered
    /\ UNCHANGED <<client_state, relay_state, ping_failures, backoff_level, active_path, dispatcher_path, monitor_target, lan_signal, relay_bridge, received_lan_verify, received_path_pong, received_lan_confirm, received_path_ping, received_relay_resume>>


\* client: RelayConnected -> LANConnecting on recv lan_offer [lan_enabled]
client_RelayConnected_to_LANConnecting_on_lan_offer_lan_enabled ==
    /\ client_state = client_RelayConnected
    /\ received_lan_offer.type = MSG_lan_offer
    /\ TRUE
    /\ received_lan_offer' = [type |-> "none"]
    /\ client_state' = client_LANConnecting
    /\ UNCHANGED <<backend_state, relay_state, ping_failures, backoff_level, active_path, dispatcher_path, monitor_target, lan_signal, relay_bridge, received_lan_verify, received_path_pong, received_lan_confirm, received_path_ping, received_relay_resume>>

\* client: RelayConnected -> RelayConnected on recv lan_offer [lan_disabled]
client_RelayConnected_to_RelayConnected_on_lan_offer_lan_disabled ==
    /\ client_state = client_RelayConnected
    /\ received_lan_offer.type = MSG_lan_offer
    /\ FALSE
    /\ received_lan_offer' = [type |-> "none"]
    /\ client_state' = client_RelayConnected
    /\ UNCHANGED <<backend_state, relay_state, ping_failures, backoff_level, active_path, dispatcher_path, monitor_target, lan_signal, relay_bridge, received_lan_verify, received_path_pong, received_lan_confirm, received_path_ping, received_relay_resume>>

\* client: LANConnecting -> LANVerifying (lan_dial_ok)
client_LANConnecting_to_LANVerifying_lan_dial_ok ==
    /\ client_state = client_LANConnecting
    /\ received_lan_verify' = [type |-> MSG_lan_verify, challenge |-> offer_challenge, instance_id |-> instance_id]
    /\ client_state' = client_LANVerifying
    /\ UNCHANGED <<backend_state, relay_state, ping_failures, backoff_level, active_path, dispatcher_path, monitor_target, lan_signal, relay_bridge, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping, received_relay_resume>>

\* client: LANConnecting -> RelayConnected (lan_dial_failed)
client_LANConnecting_to_RelayConnected_lan_dial_failed ==
    /\ client_state = client_LANConnecting
    /\ client_state' = client_RelayConnected
    /\ UNCHANGED <<backend_state, relay_state, ping_failures, backoff_level, active_path, dispatcher_path, monitor_target, lan_signal, relay_bridge, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping, received_relay_resume>>

\* client: LANVerifying -> LANActive on recv lan_confirm
client_LANVerifying_to_LANActive_on_lan_confirm ==
    /\ client_state = client_LANVerifying
    /\ received_lan_confirm.type = MSG_lan_confirm
    /\ received_lan_confirm' = [type |-> "none"]
    /\ client_state' = client_LANActive
    /\ active_path' = "lan"
    /\ dispatcher_path' = "lan"
    /\ lan_signal' = "ready"
    /\ UNCHANGED <<backend_state, relay_state, ping_failures, backoff_level, monitor_target, relay_bridge, received_lan_verify, received_path_pong, received_lan_offer, received_path_ping, received_relay_resume>>

\* client: LANVerifying -> RelayConnected (verify_timeout)
client_LANVerifying_to_RelayConnected_verify_timeout ==
    /\ client_state = client_LANVerifying
    /\ client_state' = client_RelayConnected
    /\ dispatcher_path' = "relay"
    /\ UNCHANGED <<backend_state, relay_state, ping_failures, backoff_level, active_path, monitor_target, lan_signal, relay_bridge, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping, received_relay_resume>>

\* client: LANActive -> LANActive on recv path_ping
client_LANActive_to_LANActive_on_path_ping ==
    /\ client_state = client_LANActive
    /\ received_path_ping.type = MSG_path_ping
    /\ received_path_ping' = [type |-> "none"]
    /\ received_path_pong' = [type |-> MSG_path_pong]
    /\ client_state' = client_LANActive
    /\ UNCHANGED <<backend_state, relay_state, ping_failures, backoff_level, active_path, dispatcher_path, monitor_target, lan_signal, relay_bridge, received_lan_verify, received_lan_offer, received_lan_confirm, received_relay_resume>>

\* client: LANActive -> RelayFallback (lan_error)
client_LANActive_to_RelayFallback_lan_error ==
    /\ client_state = client_LANActive
    /\ client_state' = client_RelayFallback
    /\ active_path' = "relay"
    /\ dispatcher_path' = "relay"
    /\ lan_signal' = "pending"
    /\ UNCHANGED <<backend_state, relay_state, ping_failures, backoff_level, monitor_target, relay_bridge, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping, received_relay_resume>>

\* client: RelayFallback -> RelayConnected (relay_ok)
client_RelayFallback_to_RelayConnected_relay_ok ==
    /\ client_state = client_RelayFallback
    /\ client_state' = client_RelayConnected
    /\ UNCHANGED <<backend_state, relay_state, ping_failures, backoff_level, active_path, dispatcher_path, monitor_target, lan_signal, relay_bridge, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping, received_relay_resume>>

\* client: LANActive -> LANConnecting on recv lan_offer [lan_enabled]
client_LANActive_to_LANConnecting_on_lan_offer_lan_enabled ==
    /\ client_state = client_LANActive
    /\ received_lan_offer.type = MSG_lan_offer
    /\ TRUE
    /\ received_lan_offer' = [type |-> "none"]
    /\ client_state' = client_LANConnecting
    /\ UNCHANGED <<backend_state, relay_state, ping_failures, backoff_level, active_path, dispatcher_path, monitor_target, lan_signal, relay_bridge, received_lan_verify, received_path_pong, received_lan_confirm, received_path_ping, received_relay_resume>>


\* relay: Idle -> BackendRegistered (backend_register)
relay_Idle_to_BackendRegistered_backend_register ==
    /\ relay_state = relay_Idle
    /\ relay_state' = relay_BackendRegistered
    /\ UNCHANGED <<backend_state, client_state, ping_failures, backoff_level, active_path, dispatcher_path, monitor_target, lan_signal, relay_bridge, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping, received_relay_resume>>

\* relay: BackendRegistered -> Bridged (client_connect)
relay_BackendRegistered_to_Bridged_client_connect ==
    /\ relay_state = relay_BackendRegistered
    /\ relay_state' = relay_Bridged
    /\ relay_bridge' = "active"
    /\ UNCHANGED <<backend_state, client_state, ping_failures, backoff_level, active_path, dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping, received_relay_resume>>

\* relay: Bridged -> BackendRegistered (client_disconnect)
relay_Bridged_to_BackendRegistered_client_disconnect ==
    /\ relay_state = relay_Bridged
    /\ relay_state' = relay_BackendRegistered
    /\ relay_bridge' = "idle"
    /\ UNCHANGED <<backend_state, client_state, ping_failures, backoff_level, active_path, dispatcher_path, monitor_target, lan_signal, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping, received_relay_resume>>

\* relay: Bridged -> Bridged on recv relay_resume
relay_Bridged_to_Bridged_on_relay_resume ==
    /\ relay_state = relay_Bridged
    /\ received_relay_resume.type = MSG_relay_resume
    /\ received_relay_resume' = [type |-> "none"]
    /\ received_relay_resumed' = [type |-> MSG_relay_resumed]
    /\ relay_state' = relay_Bridged
    /\ UNCHANGED <<backend_state, client_state, ping_failures, backoff_level, active_path, dispatcher_path, monitor_target, lan_signal, relay_bridge, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping>>

\* relay: BackendRegistered -> Idle (backend_disconnect)
relay_BackendRegistered_to_Idle_backend_disconnect ==
    /\ relay_state = relay_BackendRegistered
    /\ relay_state' = relay_Idle
    /\ UNCHANGED <<backend_state, client_state, ping_failures, backoff_level, active_path, dispatcher_path, monitor_target, lan_signal, relay_bridge, received_lan_verify, received_path_pong, received_lan_offer, received_lan_confirm, received_path_ping, received_relay_resume>>


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
    \/ relay_Idle_to_BackendRegistered_backend_register
    \/ relay_BackendRegistered_to_Bridged_client_connect
    \/ relay_Bridged_to_BackendRegistered_client_disconnect
    \/ relay_Bridged_to_Bridged_on_relay_resume
    \/ relay_BackendRegistered_to_Idle_backend_disconnect

Spec == Init /\ [][Next]_vars /\ WF_vars(Next)

\* ================================================================
\* Invariants and properties
\* ================================================================

\* The relay registration is never lost while the session is active
RelayAlwaysAvailable == relay_state \in {relay_BackendRegistered, relay_Bridged}
\* Traffic flows through exactly one valid path
PathConsistency == active_path \in {"relay", "lan"}
\* LAN path is only active after successful challenge verification
LANRequiresVerification == (backend_state = backend_LANActive /\ client_state = client_LANActive) => challenge_bytes = offer_challenge
\* Backoff level never exceeds the cap
BackoffBounded == backoff_level <= max_backoff_level
\* Successful LAN establishment resets the backoff level
BackoffResetsOnSuccess == backend_state = backend_LANActive => backoff_level = 0
\* The datagram dispatcher is always reading from a valid path
DispatcherAlwaysBound == dispatcher_path \in {"relay", "lan"}
\* When LAN is active, the dispatcher reads from LAN
DispatcherMatchesActivePath == (backend_state = backend_LANActive \/ client_state = client_LANActive) => dispatcher_path = "lan"
\* After fallback, the dispatcher reads from relay
DispatcherRelayOnFallback == (backend_state = backend_RelayBackoff \/ client_state = client_RelayFallback) => dispatcher_path = "relay"
\* Health monitor only pings when LAN is active or degraded
MonitorOnlyWhenLANActive == monitor_target = "lan" => backend_state \in {backend_LANActive, backend_LANDegraded}
\* Health monitor stops on fallback
MonitorOffOnFallback == backend_state = backend_RelayBackoff => monitor_target = "none"
\* LANReady signal is only "ready" when LAN is the active path
LANSignalReady == (backend_state = backend_LANActive /\ lan_signal = "ready") => active_path = "lan"
\* LANReady resets to pending on fallback
LANSignalPendingOnFallback == backend_state = backend_RelayBackoff => lan_signal = "pending"

====
