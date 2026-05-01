// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Code generated from protocol/*.yaml. DO NOT EDIT.

#include "pathswitch_gen.h"
#include <string.h>

void pigeon_backend_machine_init(pigeon_backend_machine *m)
{
	memset(m, 0, sizeof(*m));
	m->state = PIGEON_BACKEND_RELAY_CONNECTED;
	m->ping_failures = 0;
	m->backoff_level = 0;
	m->active_path = "relay";
	m->dispatcher_path = "relay";
	m->monitor_target = "none";
	m->lan_signal = "pending";
}

int pigeon_backend_handle_message(pigeon_backend_machine *m, path_switch_msg_type msg)
{
	if (m->state == PIGEON_BACKEND_LAN_OFFERED && msg == PIGEON_MSG_LAN_VERIFY && m->guards[PIGEON_GUARD_CHALLENGE_VALID] && m->guards[PIGEON_GUARD_CHALLENGE_VALID](m->userdata)) {
		if (m->actions[PIGEON_ACTION_ACTIVATE_LAN]) {
			int err = m->actions[PIGEON_ACTION_ACTIVATE_LAN](m->userdata);
			if (err) return -err;
		}
		m->ping_failures = 0;
		if (m->on_change) m->on_change("ping_failures", m->userdata);
		m->backoff_level = 0;
		if (m->on_change) m->on_change("backoff_level", m->userdata);
		m->active_path = "lan";
		if (m->on_change) m->on_change("active_path", m->userdata);
		m->monitor_target = "lan";
		if (m->on_change) m->on_change("monitor_target", m->userdata);
		m->dispatcher_path = "lan";
		if (m->on_change) m->on_change("dispatcher_path", m->userdata);
		m->lan_signal = "ready";
		if (m->on_change) m->on_change("lan_signal", m->userdata);
		m->state = PIGEON_BACKEND_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_OFFERED && msg == PIGEON_MSG_LAN_VERIFY && m->guards[PIGEON_GUARD_CHALLENGE_INVALID] && m->guards[PIGEON_GUARD_CHALLENGE_INVALID](m->userdata)) {
		m->state = PIGEON_BACKEND_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_DEGRADED && msg == PIGEON_MSG_PATH_PONG) {
		if (m->actions[PIGEON_ACTION_RESET_FAILURES]) {
			int err = m->actions[PIGEON_ACTION_RESET_FAILURES](m->userdata);
			if (err) return -err;
		}
		m->ping_failures = 0;
		if (m->on_change) m->on_change("ping_failures", m->userdata);
		m->state = PIGEON_BACKEND_LAN_ACTIVE;
		return 1;
	}
	return 0;
}

int pigeon_backend_step(pigeon_backend_machine *m, path_switch_event_id event)
{
	if (m->state == PIGEON_BACKEND_RELAY_CONNECTED && event == PIGEON_EVENT_LAN_SERVER_READY) {
		m->state = PIGEON_BACKEND_LAN_OFFERED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_OFFERED && event == PIGEON_EVENT_OFFER_TIMEOUT) {
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m->state = PIGEON_BACKEND_RELAY_BACKOFF;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_ACTIVE && event == PIGEON_EVENT_PING_TICK) {
		m->state = PIGEON_BACKEND_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_ACTIVE && event == PIGEON_EVENT_PING_TIMEOUT) {
		m->ping_failures = 1;
		if (m->on_change) m->on_change("ping_failures", m->userdata);
		m->state = PIGEON_BACKEND_LAN_DEGRADED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_DEGRADED && event == PIGEON_EVENT_PING_TICK) {
		m->state = PIGEON_BACKEND_LAN_DEGRADED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_DEGRADED && event == PIGEON_EVENT_PING_TIMEOUT && m->guards[PIGEON_GUARD_UNDER_MAX_FAILURES] && m->guards[PIGEON_GUARD_UNDER_MAX_FAILURES](m->userdata)) {
		m->ping_failures = m->ping_failures + 1;
		if (m->on_change) m->on_change("ping_failures", m->userdata);
		m->state = PIGEON_BACKEND_LAN_DEGRADED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_DEGRADED && event == PIGEON_EVENT_PING_TIMEOUT && m->guards[PIGEON_GUARD_AT_MAX_FAILURES] && m->guards[PIGEON_GUARD_AT_MAX_FAILURES](m->userdata)) {
		if (m->actions[PIGEON_ACTION_FALLBACK_TO_RELAY]) {
			int err = m->actions[PIGEON_ACTION_FALLBACK_TO_RELAY](m->userdata);
			if (err) return -err;
		}
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m->active_path = "relay";
		if (m->on_change) m->on_change("active_path", m->userdata);
		m->monitor_target = "none";
		if (m->on_change) m->on_change("monitor_target", m->userdata);
		m->dispatcher_path = "relay";
		if (m->on_change) m->on_change("dispatcher_path", m->userdata);
		m->lan_signal = "pending";
		if (m->on_change) m->on_change("lan_signal", m->userdata);
		m->ping_failures = 0;
		if (m->on_change) m->on_change("ping_failures", m->userdata);
		m->state = PIGEON_BACKEND_RELAY_BACKOFF;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_RELAY_BACKOFF && event == PIGEON_EVENT_BACKOFF_EXPIRED) {
		m->state = PIGEON_BACKEND_LAN_OFFERED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_RELAY_BACKOFF && event == PIGEON_EVENT_LAN_SERVER_CHANGED) {
		m->backoff_level = 0;
		if (m->on_change) m->on_change("backoff_level", m->userdata);
		m->state = PIGEON_BACKEND_LAN_OFFERED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_RELAY_CONNECTED && event == PIGEON_EVENT_READVERTISE_TICK && m->guards[PIGEON_GUARD_LAN_SERVER_AVAILABLE] && m->guards[PIGEON_GUARD_LAN_SERVER_AVAILABLE](m->userdata)) {
		m->state = PIGEON_BACKEND_LAN_OFFERED;
		return 1;
	}
	return 0;
}

void pigeon_client_machine_init(pigeon_client_machine *m)
{
	memset(m, 0, sizeof(*m));
	m->state = PIGEON_CLIENT_RELAY_CONNECTED;
	m->active_path = "relay";
	m->dispatcher_path = "relay";
	m->lan_signal = "pending";
}

int pigeon_client_handle_message(pigeon_client_machine *m, path_switch_msg_type msg)
{
	if (m->state == PIGEON_CLIENT_RELAY_CONNECTED && msg == PIGEON_MSG_LAN_OFFER && m->guards[PIGEON_GUARD_LAN_ENABLED] && m->guards[PIGEON_GUARD_LAN_ENABLED](m->userdata)) {
		if (m->actions[PIGEON_ACTION_DIAL_LAN]) {
			int err = m->actions[PIGEON_ACTION_DIAL_LAN](m->userdata);
			if (err) return -err;
		}
		m->state = PIGEON_CLIENT_LAN_CONNECTING;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_RELAY_CONNECTED && msg == PIGEON_MSG_LAN_OFFER && m->guards[PIGEON_GUARD_LAN_DISABLED] && m->guards[PIGEON_GUARD_LAN_DISABLED](m->userdata)) {
		m->state = PIGEON_CLIENT_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_VERIFYING && msg == PIGEON_MSG_LAN_CONFIRM) {
		if (m->actions[PIGEON_ACTION_ACTIVATE_LAN]) {
			int err = m->actions[PIGEON_ACTION_ACTIVATE_LAN](m->userdata);
			if (err) return -err;
		}
		m->active_path = "lan";
		if (m->on_change) m->on_change("active_path", m->userdata);
		m->dispatcher_path = "lan";
		if (m->on_change) m->on_change("dispatcher_path", m->userdata);
		m->lan_signal = "ready";
		if (m->on_change) m->on_change("lan_signal", m->userdata);
		m->state = PIGEON_CLIENT_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_ACTIVE && msg == PIGEON_MSG_PATH_PING) {
		m->state = PIGEON_CLIENT_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_ACTIVE && msg == PIGEON_MSG_LAN_OFFER && m->guards[PIGEON_GUARD_LAN_ENABLED] && m->guards[PIGEON_GUARD_LAN_ENABLED](m->userdata)) {
		if (m->actions[PIGEON_ACTION_DIAL_LAN]) {
			int err = m->actions[PIGEON_ACTION_DIAL_LAN](m->userdata);
			if (err) return -err;
		}
		m->state = PIGEON_CLIENT_LAN_CONNECTING;
		return 1;
	}
	return 0;
}

int pigeon_client_step(pigeon_client_machine *m, path_switch_event_id event)
{
	if (m->state == PIGEON_CLIENT_LAN_CONNECTING && event == PIGEON_EVENT_LAN_DIAL_OK) {
		m->state = PIGEON_CLIENT_LAN_VERIFYING;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_CONNECTING && event == PIGEON_EVENT_LAN_DIAL_FAILED) {
		m->state = PIGEON_CLIENT_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_VERIFYING && event == PIGEON_EVENT_VERIFY_TIMEOUT) {
		m->dispatcher_path = "relay";
		if (m->on_change) m->on_change("dispatcher_path", m->userdata);
		m->state = PIGEON_CLIENT_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_ACTIVE && event == PIGEON_EVENT_LAN_ERROR) {
		if (m->actions[PIGEON_ACTION_FALLBACK_TO_RELAY]) {
			int err = m->actions[PIGEON_ACTION_FALLBACK_TO_RELAY](m->userdata);
			if (err) return -err;
		}
		m->active_path = "relay";
		if (m->on_change) m->on_change("active_path", m->userdata);
		m->dispatcher_path = "relay";
		if (m->on_change) m->on_change("dispatcher_path", m->userdata);
		m->lan_signal = "pending";
		if (m->on_change) m->on_change("lan_signal", m->userdata);
		m->state = PIGEON_CLIENT_RELAY_FALLBACK;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_RELAY_FALLBACK && event == PIGEON_EVENT_RELAY_OK) {
		m->state = PIGEON_CLIENT_RELAY_CONNECTED;
		return 1;
	}
	return 0;
}

void pigeon_relay_machine_init(pigeon_relay_machine *m)
{
	memset(m, 0, sizeof(*m));
	m->state = PIGEON_RELAY_IDLE;
	m->relay_bridge = "idle";
}

int pigeon_relay_handle_message(pigeon_relay_machine *m, path_switch_msg_type msg)
{
	if (m->state == PIGEON_RELAY_BRIDGED && msg == PIGEON_MSG_RELAY_RESUME) {
		if (m->actions[PIGEON_ACTION_REBRIDGE_STREAMS]) {
			int err = m->actions[PIGEON_ACTION_REBRIDGE_STREAMS](m->userdata);
			if (err) return -err;
		}
		m->state = PIGEON_RELAY_BRIDGED;
		return 1;
	}
	return 0;
}

int pigeon_relay_step(pigeon_relay_machine *m, path_switch_event_id event)
{
	if (m->state == PIGEON_RELAY_IDLE && event == PIGEON_EVENT_BACKEND_REGISTER) {
		m->state = PIGEON_RELAY_BACKEND_REGISTERED;
		return 1;
	}
	if (m->state == PIGEON_RELAY_BACKEND_REGISTERED && event == PIGEON_EVENT_CLIENT_CONNECT) {
		if (m->actions[PIGEON_ACTION_BRIDGE_STREAMS]) {
			int err = m->actions[PIGEON_ACTION_BRIDGE_STREAMS](m->userdata);
			if (err) return -err;
		}
		m->relay_bridge = "active";
		if (m->on_change) m->on_change("relay_bridge", m->userdata);
		m->state = PIGEON_RELAY_BRIDGED;
		return 1;
	}
	if (m->state == PIGEON_RELAY_BRIDGED && event == PIGEON_EVENT_CLIENT_DISCONNECT) {
		if (m->actions[PIGEON_ACTION_UNBRIDGE]) {
			int err = m->actions[PIGEON_ACTION_UNBRIDGE](m->userdata);
			if (err) return -err;
		}
		m->relay_bridge = "idle";
		if (m->on_change) m->on_change("relay_bridge", m->userdata);
		m->state = PIGEON_RELAY_BACKEND_REGISTERED;
		return 1;
	}
	if (m->state == PIGEON_RELAY_BACKEND_REGISTERED && event == PIGEON_EVENT_BACKEND_DISCONNECT) {
		m->state = PIGEON_RELAY_IDLE;
		return 1;
	}
	return 0;
}

