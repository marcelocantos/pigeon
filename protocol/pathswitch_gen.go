// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Code generated from protocol/*.yaml. DO NOT EDIT.

package protocol

// PathSwitch backend states.
const (
	PathSwitchBackendRelayConnected State = "RelayConnected"
	PathSwitchBackendLANOffered State = "LANOffered"
	PathSwitchBackendLANActive State = "LANActive"
	PathSwitchBackendRelayBackoff State = "RelayBackoff"
	PathSwitchBackendLANDegraded State = "LANDegraded"
)

// PathSwitch client states.
const (
	PathSwitchClientRelayConnected State = "RelayConnected"
	PathSwitchClientLANConnecting State = "LANConnecting"
	PathSwitchClientLANVerifying State = "LANVerifying"
	PathSwitchClientLANActive State = "LANActive"
	PathSwitchClientRelayFallback State = "RelayFallback"
)

// PathSwitch relay states.
const (
	PathSwitchRelayIdle State = "Idle"
	PathSwitchRelayBackendRegistered State = "BackendRegistered"
	PathSwitchRelayBridged State = "Bridged"
)

// PathSwitch message types.
const (
	PathSwitchMsgLanOffer MsgType = "lan_offer"
	PathSwitchMsgLanVerify MsgType = "lan_verify"
	PathSwitchMsgLanConfirm MsgType = "lan_confirm"
	PathSwitchMsgPathPing MsgType = "path_ping"
	PathSwitchMsgPathPong MsgType = "path_pong"
	PathSwitchMsgRelayResume MsgType = "relay_resume"
	PathSwitchMsgRelayResumed MsgType = "relay_resumed"
)

// PathSwitch guards.
const (
	PathSwitchGuardChallengeValid GuardID = "challenge_valid"
	PathSwitchGuardChallengeInvalid GuardID = "challenge_invalid"
	PathSwitchGuardLanEnabled GuardID = "lan_enabled"
	PathSwitchGuardLanDisabled GuardID = "lan_disabled"
	PathSwitchGuardLanServerAvailable GuardID = "lan_server_available"
	PathSwitchGuardUnderMaxFailures GuardID = "under_max_failures"
	PathSwitchGuardAtMaxFailures GuardID = "at_max_failures"
)

// PathSwitch actions.
const (
	PathSwitchActionActivateLan ActionID = "activate_lan"
	PathSwitchActionBridgeStreams ActionID = "bridge_streams"
	PathSwitchActionDialLan ActionID = "dial_lan"
	PathSwitchActionFallbackToRelay ActionID = "fallback_to_relay"
	PathSwitchActionRebridgeStreams ActionID = "rebridge_streams"
	PathSwitchActionResetFailures ActionID = "reset_failures"
	PathSwitchActionUnbridge ActionID = "unbridge"
)

// PathSwitch events.
const (
	PathSwitchEventBackendDisconnect EventID = "backend_disconnect"
	PathSwitchEventBackendRegister EventID = "backend_register"
	PathSwitchEventBackoffExpired EventID = "backoff_expired"
	PathSwitchEventClientConnect EventID = "client_connect"
	PathSwitchEventClientDisconnect EventID = "client_disconnect"
	PathSwitchEventLanDialFailed EventID = "lan_dial_failed"
	PathSwitchEventLanDialOk EventID = "lan_dial_ok"
	PathSwitchEventLanError EventID = "lan_error"
	PathSwitchEventLanServerChanged EventID = "lan_server_changed"
	PathSwitchEventLanServerReady EventID = "lan_server_ready"
	PathSwitchEventOfferTimeout EventID = "offer_timeout"
	PathSwitchEventPingTick EventID = "ping_tick"
	PathSwitchEventPingTimeout EventID = "ping_timeout"
	PathSwitchEventReadvertiseTick EventID = "readvertise_tick"
	PathSwitchEventRecvLanConfirm EventID = "recv_lan_confirm"
	PathSwitchEventRecvLanOffer EventID = "recv_lan_offer"
	PathSwitchEventRecvLanVerify EventID = "recv_lan_verify"
	PathSwitchEventRecvPathPing EventID = "recv_path_ping"
	PathSwitchEventRecvPathPong EventID = "recv_path_pong"
	PathSwitchEventRecvRelayResume EventID = "recv_relay_resume"
	PathSwitchEventRelayOk EventID = "relay_ok"
	PathSwitchEventVerifyTimeout EventID = "verify_timeout"
)

func PathSwitch() *Protocol {
	return &Protocol{
		Name: "PathSwitch",
		Actors: []Actor{
			{Name: "backend", Initial: "RelayConnected", Transitions: []Transition{
				{From: "RelayConnected", To: "LANOffered", On: Internal("lan_server_ready"), Sends: []Send{{To: "client", Msg: "lan_offer", Fields: map[string]string{"addr": "lan_addr", "challenge": "challenge_bytes", }}, }},
				{From: "LANOffered", To: "LANActive", On: Recv("lan_verify"), Guard: "challenge_valid", Do: "activate_lan", Sends: []Send{{To: "client", Msg: "lan_confirm"}, }, Updates: []VarUpdate{{Var: "ping_failures", Expr: "0"}, {Var: "backoff_level", Expr: "0"}, {Var: "active_path", Expr: "\"lan\""}, {Var: "monitor_target", Expr: "\"lan\""}, {Var: "dispatcher_path", Expr: "\"lan\""}, {Var: "lan_signal", Expr: "\"ready\""}, }},
				{From: "LANOffered", To: "RelayConnected", On: Recv("lan_verify"), Guard: "challenge_invalid"},
				{From: "LANOffered", To: "RelayBackoff", On: Internal("offer_timeout"), Updates: []VarUpdate{{Var: "backoff_level", Expr: "Min(backoff_level + 1, max_backoff_level)"}, }},
				{From: "LANActive", To: "LANActive", On: Internal("ping_tick"), Sends: []Send{{To: "client", Msg: "path_ping"}, }},
				{From: "LANActive", To: "LANDegraded", On: Internal("ping_timeout"), Updates: []VarUpdate{{Var: "ping_failures", Expr: "1"}, }},
				{From: "LANDegraded", To: "LANDegraded", On: Internal("ping_tick"), Sends: []Send{{To: "client", Msg: "path_ping"}, }},
				{From: "LANDegraded", To: "LANActive", On: Recv("path_pong"), Do: "reset_failures", Updates: []VarUpdate{{Var: "ping_failures", Expr: "0"}, }},
				{From: "LANDegraded", To: "LANDegraded", On: Internal("ping_timeout"), Guard: "under_max_failures", Updates: []VarUpdate{{Var: "ping_failures", Expr: "ping_failures + 1"}, }},
				{From: "LANDegraded", To: "RelayBackoff", On: Internal("ping_timeout"), Guard: "at_max_failures", Do: "fallback_to_relay", Updates: []VarUpdate{{Var: "backoff_level", Expr: "Min(backoff_level + 1, max_backoff_level)"}, {Var: "active_path", Expr: "\"relay\""}, {Var: "monitor_target", Expr: "\"none\""}, {Var: "dispatcher_path", Expr: "\"relay\""}, {Var: "lan_signal", Expr: "\"pending\""}, {Var: "ping_failures", Expr: "0"}, }},
				{From: "RelayBackoff", To: "LANOffered", On: Internal("backoff_expired"), Sends: []Send{{To: "client", Msg: "lan_offer", Fields: map[string]string{"addr": "lan_addr", "challenge": "challenge_bytes", }}, }},
				{From: "RelayBackoff", To: "LANOffered", On: Internal("lan_server_changed"), Sends: []Send{{To: "client", Msg: "lan_offer", Fields: map[string]string{"addr": "lan_addr", "challenge": "challenge_bytes", }}, }, Updates: []VarUpdate{{Var: "backoff_level", Expr: "0"}, }},
				{From: "RelayConnected", To: "LANOffered", On: Internal("readvertise_tick"), Guard: "lan_server_available", Sends: []Send{{To: "client", Msg: "lan_offer", Fields: map[string]string{"addr": "lan_addr", "challenge": "challenge_bytes", }}, }},
			}},
			{Name: "client", Initial: "RelayConnected", Transitions: []Transition{
				{From: "RelayConnected", To: "LANConnecting", On: Recv("lan_offer"), Guard: "lan_enabled", Do: "dial_lan"},
				{From: "RelayConnected", To: "RelayConnected", On: Recv("lan_offer"), Guard: "lan_disabled"},
				{From: "LANConnecting", To: "LANVerifying", On: Internal("lan_dial_ok"), Sends: []Send{{To: "backend", Msg: "lan_verify", Fields: map[string]string{"challenge": "offer_challenge", "instance_id": "instance_id", }}, }},
				{From: "LANConnecting", To: "RelayConnected", On: Internal("lan_dial_failed")},
				{From: "LANVerifying", To: "LANActive", On: Recv("lan_confirm"), Do: "activate_lan", Updates: []VarUpdate{{Var: "active_path", Expr: "\"lan\""}, {Var: "dispatcher_path", Expr: "\"lan\""}, {Var: "lan_signal", Expr: "\"ready\""}, }},
				{From: "LANVerifying", To: "RelayConnected", On: Internal("verify_timeout"), Updates: []VarUpdate{{Var: "dispatcher_path", Expr: "\"relay\""}, }},
				{From: "LANActive", To: "LANActive", On: Recv("path_ping"), Sends: []Send{{To: "backend", Msg: "path_pong"}, }},
				{From: "LANActive", To: "RelayFallback", On: Internal("lan_error"), Do: "fallback_to_relay", Updates: []VarUpdate{{Var: "active_path", Expr: "\"relay\""}, {Var: "dispatcher_path", Expr: "\"relay\""}, {Var: "lan_signal", Expr: "\"pending\""}, }},
				{From: "RelayFallback", To: "RelayConnected", On: Internal("relay_ok")},
				{From: "LANActive", To: "LANConnecting", On: Recv("lan_offer"), Guard: "lan_enabled", Do: "dial_lan"},
			}},
			{Name: "relay", Initial: "Idle", Transitions: []Transition{
				{From: "Idle", To: "BackendRegistered", On: Internal("backend_register")},
				{From: "BackendRegistered", To: "Bridged", On: Internal("client_connect"), Do: "bridge_streams", Updates: []VarUpdate{{Var: "relay_bridge", Expr: "\"active\""}, }},
				{From: "Bridged", To: "BackendRegistered", On: Internal("client_disconnect"), Do: "unbridge", Updates: []VarUpdate{{Var: "relay_bridge", Expr: "\"idle\""}, }},
				{From: "Bridged", To: "Bridged", On: Recv("relay_resume"), Do: "rebridge_streams", Sends: []Send{{To: "client", Msg: "relay_resumed"}, }},
				{From: "BackendRegistered", To: "Idle", On: Internal("backend_disconnect")},
			}},
		},
		Messages: []Message{
			{Type: "lan_offer", From: "backend", To: "client", Desc: "LAN address + challenge (sent via relay)"},
			{Type: "lan_verify", From: "client", To: "backend", Desc: "challenge response + instance ID (sent via LAN)"},
			{Type: "lan_confirm", From: "backend", To: "client", Desc: "LAN verified, path is live (sent via LAN)"},
			{Type: "path_ping", From: "backend", To: "client", Desc: "health check on active direct path"},
			{Type: "path_pong", From: "client", To: "backend", Desc: "health check response"},
			{Type: "relay_resume", From: "client", To: "relay", Desc: "client reconnects relay stream after direct path failure"},
			{Type: "relay_resumed", From: "relay", To: "client", Desc: "relay confirms stream is re-bridged"},
		},
		Vars: []VarDef{
			{Name: "lan_addr", Initial: "\"none\"", Desc: "LAN server address (host:port)"},
			{Name: "challenge_bytes", Initial: "\"none\"", Desc: "32-byte random challenge for LAN verification"},
			{Name: "offer_challenge", Initial: "\"none\"", Desc: "challenge from the most recent LAN offer"},
			{Name: "instance_id", Initial: "\"none\"", Desc: "relay instance ID of this peer"},
			{Name: "ping_failures", Initial: "0", Desc: "consecutive failed pings on the direct path"},
			{Name: "max_ping_failures", Initial: "3", Desc: "threshold before fallback"},
			{Name: "backoff_level", Initial: "0", Desc: "current exponential backoff level (0 = no backoff)"},
			{Name: "max_backoff_level", Initial: "5", Desc: "cap on backoff level (2^5 * base = 32x base interval)"},
			{Name: "active_path", Initial: "\"relay\"", Desc: "\"relay\" or \"lan\" — which path carries application traffic"},
			{Name: "dispatcher_path", Initial: "\"relay\"", Desc: "which path the datagram dispatcher reads from (\"relay\", \"lan\", \"none\")"},
			{Name: "monitor_target", Initial: "\"none\"", Desc: "which path the health monitor pings (\"lan\", \"none\")"},
			{Name: "lan_signal", Initial: "\"pending\"", Desc: "LANReady notification state (\"pending\" = not yet, \"ready\" = closed/signalled)"},
			{Name: "relay_bridge", Initial: "\"idle\"", Desc: "relay bridge state (\"active\" = bridging, \"idle\" = backend registered but no client)"},
			{Name: "lan_server_addr", Initial: "\"none\"", Desc: "LAN server listen address (backend only)"},
		},
		Guards: []GuardDef{
			{ID: "challenge_valid", Expr: "offer_challenge = challenge_bytes"},
			{ID: "challenge_invalid", Expr: "offer_challenge /= challenge_bytes"},
			{ID: "lan_enabled", Expr: "TRUE"},
			{ID: "lan_disabled", Expr: "FALSE"},
			{ID: "lan_server_available", Expr: "lan_server_addr /= \"none\""},
			{ID: "under_max_failures", Expr: "ping_failures + 1 < max_ping_failures"},
			{ID: "at_max_failures", Expr: "ping_failures + 1 >= max_ping_failures"},
		},
		Operators: []Operator{
		},
		AdvActions: []AdvAction{
		},
		Properties: []Property{
			{Name: "RelayAlwaysAvailable", Kind: Invariant, Expr: "relay_state \\in {relay_BackendRegistered, relay_Bridged}", Desc: "The relay registration is never lost while the session is active"},
			{Name: "PathConsistency", Kind: Invariant, Expr: "active_path \\in {\"relay\", \"lan\"}", Desc: "Traffic flows through exactly one valid path"},
			{Name: "LANRequiresVerification", Kind: Invariant, Expr: "(backend_state = backend_LANActive /\\ client_state = client_LANActive) => challenge_bytes = offer_challenge", Desc: "LAN path is only active after successful challenge verification"},
			{Name: "BackoffBounded", Kind: Invariant, Expr: "backoff_level <= max_backoff_level", Desc: "Backoff level never exceeds the cap"},
			{Name: "BackoffResetsOnSuccess", Kind: Invariant, Expr: "backend_state = backend_LANActive => backoff_level = 0", Desc: "Successful LAN establishment resets the backoff level"},
			{Name: "DispatcherAlwaysBound", Kind: Invariant, Expr: "dispatcher_path \\in {\"relay\", \"lan\"}", Desc: "The datagram dispatcher is always reading from a valid path"},
			{Name: "DispatcherMatchesActivePath", Kind: Invariant, Expr: "(backend_state = backend_LANActive \\/ client_state = client_LANActive) => dispatcher_path = \"lan\"", Desc: "When LAN is active, the dispatcher reads from LAN"},
			{Name: "DispatcherRelayOnFallback", Kind: Invariant, Expr: "(backend_state = backend_RelayBackoff \\/ client_state = client_RelayFallback) => dispatcher_path = \"relay\"", Desc: "After fallback, the dispatcher reads from relay"},
			{Name: "MonitorOnlyWhenLANActive", Kind: Invariant, Expr: "monitor_target = \"lan\" => backend_state \\in {backend_LANActive, backend_LANDegraded}", Desc: "Health monitor only pings when LAN is active or degraded"},
			{Name: "MonitorOffOnFallback", Kind: Invariant, Expr: "backend_state = backend_RelayBackoff => monitor_target = \"none\"", Desc: "Health monitor stops on fallback"},
			{Name: "LANSignalReady", Kind: Invariant, Expr: "(backend_state = backend_LANActive /\\ lan_signal = \"ready\") => active_path = \"lan\"", Desc: "LANReady signal is only \"ready\" when LAN is the active path"},
			{Name: "LANSignalPendingOnFallback", Kind: Invariant, Expr: "backend_state = backend_RelayBackoff => lan_signal = \"pending\"", Desc: "LANReady resets to pending on fallback"},
		},
		ChannelBound: 3,
		OneShot: false,
	}
}

// PathSwitchBackendMachine is the generated state machine for the backend actor.
type PathSwitchBackendMachine struct {
	State State
	PingFailures int // consecutive failed pings on the direct path
	BackoffLevel int // current exponential backoff level (0 = no backoff)
	ActivePath string // "relay" or "lan" — which path carries application traffic
	DispatcherPath string // which path the datagram dispatcher reads from ("relay", "lan", "none")
	MonitorTarget string // which path the health monitor pings ("lan", "none")
	LanSignal string // LANReady notification state ("pending" = not yet, "ready" = closed/signalled)

	Guards  map[GuardID]func() bool
	Actions map[ActionID]func() error
	OnChange func(varName string)
}

func NewPathSwitchBackendMachine() *PathSwitchBackendMachine {
	return &PathSwitchBackendMachine{
		State: PathSwitchBackendRelayConnected,
		PingFailures: 0,
		BackoffLevel: 0,
		ActivePath: "relay",
		DispatcherPath: "relay",
		MonitorTarget: "none",
		LanSignal: "pending",
		Guards:  make(map[GuardID]func() bool),
		Actions: make(map[ActionID]func() error),
	}
}

func (m *PathSwitchBackendMachine) HandleMessage(msg MsgType) (bool, error) {
	switch {
	case m.State == PathSwitchBackendLANOffered && msg == PathSwitchMsgLanVerify && m.Guards[PathSwitchGuardChallengeValid] != nil && m.Guards[PathSwitchGuardChallengeValid]():
		if fn := m.Actions[PathSwitchActionActivateLan]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.PingFailures = 0
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.BackoffLevel = 0
		if m.OnChange != nil { m.OnChange("backoff_level") }
		m.ActivePath = "lan"
		if m.OnChange != nil { m.OnChange("active_path") }
		m.MonitorTarget = "lan"
		if m.OnChange != nil { m.OnChange("monitor_target") }
		m.DispatcherPath = "lan"
		if m.OnChange != nil { m.OnChange("dispatcher_path") }
		m.LanSignal = "ready"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = PathSwitchBackendLANActive
		return true, nil
	case m.State == PathSwitchBackendLANOffered && msg == PathSwitchMsgLanVerify && m.Guards[PathSwitchGuardChallengeInvalid] != nil && m.Guards[PathSwitchGuardChallengeInvalid]():
		m.State = PathSwitchBackendRelayConnected
		return true, nil
	case m.State == PathSwitchBackendLANDegraded && msg == PathSwitchMsgPathPong:
		if fn := m.Actions[PathSwitchActionResetFailures]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.PingFailures = 0
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = PathSwitchBackendLANActive
		return true, nil
	}
	return false, nil
}

func (m *PathSwitchBackendMachine) Step(event EventID) (bool, error) {
	switch {
	case m.State == PathSwitchBackendRelayConnected && event == PathSwitchEventLanServerReady:
		m.State = PathSwitchBackendLANOffered
		return true, nil
	case m.State == PathSwitchBackendLANOffered && event == PathSwitchEventOfferTimeout:
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.State = PathSwitchBackendRelayBackoff
		return true, nil
	case m.State == PathSwitchBackendLANActive && event == PathSwitchEventPingTick:
		m.State = PathSwitchBackendLANActive
		return true, nil
	case m.State == PathSwitchBackendLANActive && event == PathSwitchEventPingTimeout:
		m.PingFailures = 1
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = PathSwitchBackendLANDegraded
		return true, nil
	case m.State == PathSwitchBackendLANDegraded && event == PathSwitchEventPingTick:
		m.State = PathSwitchBackendLANDegraded
		return true, nil
	case m.State == PathSwitchBackendLANDegraded && event == PathSwitchEventPingTimeout && m.Guards[PathSwitchGuardUnderMaxFailures] != nil && m.Guards[PathSwitchGuardUnderMaxFailures]():
		m.PingFailures = m.PingFailures + 1
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = PathSwitchBackendLANDegraded
		return true, nil
	case m.State == PathSwitchBackendLANDegraded && event == PathSwitchEventPingTimeout && m.Guards[PathSwitchGuardAtMaxFailures] != nil && m.Guards[PathSwitchGuardAtMaxFailures]():
		if fn := m.Actions[PathSwitchActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.ActivePath = "relay"
		if m.OnChange != nil { m.OnChange("active_path") }
		m.MonitorTarget = "none"
		if m.OnChange != nil { m.OnChange("monitor_target") }
		m.DispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("dispatcher_path") }
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.PingFailures = 0
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = PathSwitchBackendRelayBackoff
		return true, nil
	case m.State == PathSwitchBackendRelayBackoff && event == PathSwitchEventBackoffExpired:
		m.State = PathSwitchBackendLANOffered
		return true, nil
	case m.State == PathSwitchBackendRelayBackoff && event == PathSwitchEventLanServerChanged:
		m.BackoffLevel = 0
		if m.OnChange != nil { m.OnChange("backoff_level") }
		m.State = PathSwitchBackendLANOffered
		return true, nil
	case m.State == PathSwitchBackendRelayConnected && event == PathSwitchEventReadvertiseTick && m.Guards[PathSwitchGuardLanServerAvailable] != nil && m.Guards[PathSwitchGuardLanServerAvailable]():
		m.State = PathSwitchBackendLANOffered
		return true, nil
	}
	return false, nil
}

func (m *PathSwitchBackendMachine) HandleEvent(ev EventID) ([]CmdID, error) {
	switch {
	case m.State == PathSwitchBackendRelayConnected && ev == PathSwitchEventLanServerReady:
		m.State = PathSwitchBackendLANOffered
		return nil, nil
	case m.State == PathSwitchBackendLANOffered && ev == PathSwitchEventRecvLanVerify && m.Guards[PathSwitchGuardChallengeValid] != nil && m.Guards[PathSwitchGuardChallengeValid]():
		if fn := m.Actions[PathSwitchActionActivateLan]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.PingFailures = 0
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.BackoffLevel = 0
		if m.OnChange != nil { m.OnChange("backoff_level") }
		m.ActivePath = "lan"
		if m.OnChange != nil { m.OnChange("active_path") }
		m.MonitorTarget = "lan"
		if m.OnChange != nil { m.OnChange("monitor_target") }
		m.DispatcherPath = "lan"
		if m.OnChange != nil { m.OnChange("dispatcher_path") }
		m.LanSignal = "ready"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = PathSwitchBackendLANActive
		return nil, nil
	case m.State == PathSwitchBackendLANOffered && ev == PathSwitchEventRecvLanVerify && m.Guards[PathSwitchGuardChallengeInvalid] != nil && m.Guards[PathSwitchGuardChallengeInvalid]():
		m.State = PathSwitchBackendRelayConnected
		return nil, nil
	case m.State == PathSwitchBackendLANOffered && ev == PathSwitchEventOfferTimeout:
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.State = PathSwitchBackendRelayBackoff
		return nil, nil
	case m.State == PathSwitchBackendLANActive && ev == PathSwitchEventPingTick:
		m.State = PathSwitchBackendLANActive
		return nil, nil
	case m.State == PathSwitchBackendLANActive && ev == PathSwitchEventPingTimeout:
		m.PingFailures = 1
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = PathSwitchBackendLANDegraded
		return nil, nil
	case m.State == PathSwitchBackendLANDegraded && ev == PathSwitchEventPingTick:
		m.State = PathSwitchBackendLANDegraded
		return nil, nil
	case m.State == PathSwitchBackendLANDegraded && ev == PathSwitchEventRecvPathPong:
		if fn := m.Actions[PathSwitchActionResetFailures]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.PingFailures = 0
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = PathSwitchBackendLANActive
		return nil, nil
	case m.State == PathSwitchBackendLANDegraded && ev == PathSwitchEventPingTimeout && m.Guards[PathSwitchGuardUnderMaxFailures] != nil && m.Guards[PathSwitchGuardUnderMaxFailures]():
		m.PingFailures = m.PingFailures + 1
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = PathSwitchBackendLANDegraded
		return nil, nil
	case m.State == PathSwitchBackendLANDegraded && ev == PathSwitchEventPingTimeout && m.Guards[PathSwitchGuardAtMaxFailures] != nil && m.Guards[PathSwitchGuardAtMaxFailures]():
		if fn := m.Actions[PathSwitchActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m.ActivePath = "relay"
		if m.OnChange != nil { m.OnChange("active_path") }
		m.MonitorTarget = "none"
		if m.OnChange != nil { m.OnChange("monitor_target") }
		m.DispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("dispatcher_path") }
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.PingFailures = 0
		if m.OnChange != nil { m.OnChange("ping_failures") }
		m.State = PathSwitchBackendRelayBackoff
		return nil, nil
	case m.State == PathSwitchBackendRelayBackoff && ev == PathSwitchEventBackoffExpired:
		m.State = PathSwitchBackendLANOffered
		return nil, nil
	case m.State == PathSwitchBackendRelayBackoff && ev == PathSwitchEventLanServerChanged:
		m.BackoffLevel = 0
		if m.OnChange != nil { m.OnChange("backoff_level") }
		m.State = PathSwitchBackendLANOffered
		return nil, nil
	case m.State == PathSwitchBackendRelayConnected && ev == PathSwitchEventReadvertiseTick && m.Guards[PathSwitchGuardLanServerAvailable] != nil && m.Guards[PathSwitchGuardLanServerAvailable]():
		m.State = PathSwitchBackendLANOffered
		return nil, nil
	}
	return nil, nil
}

// PathSwitchClientMachine is the generated state machine for the client actor.
type PathSwitchClientMachine struct {
	State State
	ActivePath string // "relay" or "lan" — which path carries application traffic
	DispatcherPath string // which path the datagram dispatcher reads from ("relay", "lan", "none")
	LanSignal string // LANReady notification state ("pending" = not yet, "ready" = closed/signalled)

	Guards  map[GuardID]func() bool
	Actions map[ActionID]func() error
	OnChange func(varName string)
}

func NewPathSwitchClientMachine() *PathSwitchClientMachine {
	return &PathSwitchClientMachine{
		State: PathSwitchClientRelayConnected,
		ActivePath: "relay",
		DispatcherPath: "relay",
		LanSignal: "pending",
		Guards:  make(map[GuardID]func() bool),
		Actions: make(map[ActionID]func() error),
	}
}

func (m *PathSwitchClientMachine) HandleMessage(msg MsgType) (bool, error) {
	switch {
	case m.State == PathSwitchClientRelayConnected && msg == PathSwitchMsgLanOffer && m.Guards[PathSwitchGuardLanEnabled] != nil && m.Guards[PathSwitchGuardLanEnabled]():
		if fn := m.Actions[PathSwitchActionDialLan]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.State = PathSwitchClientLANConnecting
		return true, nil
	case m.State == PathSwitchClientRelayConnected && msg == PathSwitchMsgLanOffer && m.Guards[PathSwitchGuardLanDisabled] != nil && m.Guards[PathSwitchGuardLanDisabled]():
		m.State = PathSwitchClientRelayConnected
		return true, nil
	case m.State == PathSwitchClientLANVerifying && msg == PathSwitchMsgLanConfirm:
		if fn := m.Actions[PathSwitchActionActivateLan]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.ActivePath = "lan"
		if m.OnChange != nil { m.OnChange("active_path") }
		m.DispatcherPath = "lan"
		if m.OnChange != nil { m.OnChange("dispatcher_path") }
		m.LanSignal = "ready"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = PathSwitchClientLANActive
		return true, nil
	case m.State == PathSwitchClientLANActive && msg == PathSwitchMsgPathPing:
		m.State = PathSwitchClientLANActive
		return true, nil
	case m.State == PathSwitchClientLANActive && msg == PathSwitchMsgLanOffer && m.Guards[PathSwitchGuardLanEnabled] != nil && m.Guards[PathSwitchGuardLanEnabled]():
		if fn := m.Actions[PathSwitchActionDialLan]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.State = PathSwitchClientLANConnecting
		return true, nil
	}
	return false, nil
}

func (m *PathSwitchClientMachine) Step(event EventID) (bool, error) {
	switch {
	case m.State == PathSwitchClientLANConnecting && event == PathSwitchEventLanDialOk:
		m.State = PathSwitchClientLANVerifying
		return true, nil
	case m.State == PathSwitchClientLANConnecting && event == PathSwitchEventLanDialFailed:
		m.State = PathSwitchClientRelayConnected
		return true, nil
	case m.State == PathSwitchClientLANVerifying && event == PathSwitchEventVerifyTimeout:
		m.DispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("dispatcher_path") }
		m.State = PathSwitchClientRelayConnected
		return true, nil
	case m.State == PathSwitchClientLANActive && event == PathSwitchEventLanError:
		if fn := m.Actions[PathSwitchActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.ActivePath = "relay"
		if m.OnChange != nil { m.OnChange("active_path") }
		m.DispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("dispatcher_path") }
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = PathSwitchClientRelayFallback
		return true, nil
	case m.State == PathSwitchClientRelayFallback && event == PathSwitchEventRelayOk:
		m.State = PathSwitchClientRelayConnected
		return true, nil
	}
	return false, nil
}

func (m *PathSwitchClientMachine) HandleEvent(ev EventID) ([]CmdID, error) {
	switch {
	case m.State == PathSwitchClientRelayConnected && ev == PathSwitchEventRecvLanOffer && m.Guards[PathSwitchGuardLanEnabled] != nil && m.Guards[PathSwitchGuardLanEnabled]():
		if fn := m.Actions[PathSwitchActionDialLan]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.State = PathSwitchClientLANConnecting
		return nil, nil
	case m.State == PathSwitchClientRelayConnected && ev == PathSwitchEventRecvLanOffer && m.Guards[PathSwitchGuardLanDisabled] != nil && m.Guards[PathSwitchGuardLanDisabled]():
		m.State = PathSwitchClientRelayConnected
		return nil, nil
	case m.State == PathSwitchClientLANConnecting && ev == PathSwitchEventLanDialOk:
		m.State = PathSwitchClientLANVerifying
		return nil, nil
	case m.State == PathSwitchClientLANConnecting && ev == PathSwitchEventLanDialFailed:
		m.State = PathSwitchClientRelayConnected
		return nil, nil
	case m.State == PathSwitchClientLANVerifying && ev == PathSwitchEventRecvLanConfirm:
		if fn := m.Actions[PathSwitchActionActivateLan]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.ActivePath = "lan"
		if m.OnChange != nil { m.OnChange("active_path") }
		m.DispatcherPath = "lan"
		if m.OnChange != nil { m.OnChange("dispatcher_path") }
		m.LanSignal = "ready"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = PathSwitchClientLANActive
		return nil, nil
	case m.State == PathSwitchClientLANVerifying && ev == PathSwitchEventVerifyTimeout:
		m.DispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("dispatcher_path") }
		m.State = PathSwitchClientRelayConnected
		return nil, nil
	case m.State == PathSwitchClientLANActive && ev == PathSwitchEventRecvPathPing:
		m.State = PathSwitchClientLANActive
		return nil, nil
	case m.State == PathSwitchClientLANActive && ev == PathSwitchEventLanError:
		if fn := m.Actions[PathSwitchActionFallbackToRelay]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.ActivePath = "relay"
		if m.OnChange != nil { m.OnChange("active_path") }
		m.DispatcherPath = "relay"
		if m.OnChange != nil { m.OnChange("dispatcher_path") }
		m.LanSignal = "pending"
		if m.OnChange != nil { m.OnChange("lan_signal") }
		m.State = PathSwitchClientRelayFallback
		return nil, nil
	case m.State == PathSwitchClientRelayFallback && ev == PathSwitchEventRelayOk:
		m.State = PathSwitchClientRelayConnected
		return nil, nil
	case m.State == PathSwitchClientLANActive && ev == PathSwitchEventRecvLanOffer && m.Guards[PathSwitchGuardLanEnabled] != nil && m.Guards[PathSwitchGuardLanEnabled]():
		if fn := m.Actions[PathSwitchActionDialLan]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.State = PathSwitchClientLANConnecting
		return nil, nil
	}
	return nil, nil
}

// PathSwitchRelayMachine is the generated state machine for the relay actor.
type PathSwitchRelayMachine struct {
	State State
	RelayBridge string // relay bridge state ("active" = bridging, "idle" = backend registered but no client)

	Guards  map[GuardID]func() bool
	Actions map[ActionID]func() error
	OnChange func(varName string)
}

func NewPathSwitchRelayMachine() *PathSwitchRelayMachine {
	return &PathSwitchRelayMachine{
		State: PathSwitchRelayIdle,
		RelayBridge: "idle",
		Guards:  make(map[GuardID]func() bool),
		Actions: make(map[ActionID]func() error),
	}
}

func (m *PathSwitchRelayMachine) HandleMessage(msg MsgType) (bool, error) {
	switch {
	case m.State == PathSwitchRelayBridged && msg == PathSwitchMsgRelayResume:
		if fn := m.Actions[PathSwitchActionRebridgeStreams]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.State = PathSwitchRelayBridged
		return true, nil
	}
	return false, nil
}

func (m *PathSwitchRelayMachine) Step(event EventID) (bool, error) {
	switch {
	case m.State == PathSwitchRelayIdle && event == PathSwitchEventBackendRegister:
		m.State = PathSwitchRelayBackendRegistered
		return true, nil
	case m.State == PathSwitchRelayBackendRegistered && event == PathSwitchEventClientConnect:
		if fn := m.Actions[PathSwitchActionBridgeStreams]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.RelayBridge = "active"
		if m.OnChange != nil { m.OnChange("relay_bridge") }
		m.State = PathSwitchRelayBridged
		return true, nil
	case m.State == PathSwitchRelayBridged && event == PathSwitchEventClientDisconnect:
		if fn := m.Actions[PathSwitchActionUnbridge]; fn != nil {
			if err := fn(); err != nil { return false, err }
		}
		m.RelayBridge = "idle"
		if m.OnChange != nil { m.OnChange("relay_bridge") }
		m.State = PathSwitchRelayBackendRegistered
		return true, nil
	case m.State == PathSwitchRelayBackendRegistered && event == PathSwitchEventBackendDisconnect:
		m.State = PathSwitchRelayIdle
		return true, nil
	}
	return false, nil
}

func (m *PathSwitchRelayMachine) HandleEvent(ev EventID) ([]CmdID, error) {
	switch {
	case m.State == PathSwitchRelayIdle && ev == PathSwitchEventBackendRegister:
		m.State = PathSwitchRelayBackendRegistered
		return nil, nil
	case m.State == PathSwitchRelayBackendRegistered && ev == PathSwitchEventClientConnect:
		if fn := m.Actions[PathSwitchActionBridgeStreams]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.RelayBridge = "active"
		if m.OnChange != nil { m.OnChange("relay_bridge") }
		m.State = PathSwitchRelayBridged
		return nil, nil
	case m.State == PathSwitchRelayBridged && ev == PathSwitchEventClientDisconnect:
		if fn := m.Actions[PathSwitchActionUnbridge]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.RelayBridge = "idle"
		if m.OnChange != nil { m.OnChange("relay_bridge") }
		m.State = PathSwitchRelayBackendRegistered
		return nil, nil
	case m.State == PathSwitchRelayBridged && ev == PathSwitchEventRecvRelayResume:
		if fn := m.Actions[PathSwitchActionRebridgeStreams]; fn != nil {
			if err := fn(); err != nil { return nil, err }
		}
		m.State = PathSwitchRelayBridged
		return nil, nil
	case m.State == PathSwitchRelayBackendRegistered && ev == PathSwitchEventBackendDisconnect:
		m.State = PathSwitchRelayIdle
		return nil, nil
	}
	return nil, nil
}

