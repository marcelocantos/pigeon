// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Code generated from protocol/*.yaml. DO NOT EDIT.

#include "session_gen.h"
#include <string.h>

void pigeon_backend_machine_init(pigeon_backend_machine *m)
{
	memset(m, 0, sizeof(*m));
	m->state = PIGEON_BACKEND_IDLE;
	m->current_token = "none";
	m->backend_ecdh_pub = "none";
	m->received_client_pub = "none";
	m->code_attempts = 0;
	m->device_secret = "none";
	m->received_device_id = "none";
	m->received_auth_nonce = "none";
	m->secret_published = false;
	m->ping_failures = 0;
	m->backoff_level = 0;
	m->b_active_path = "relay";
	m->b_dispatcher_path = "relay";
	m->monitor_target = "none";
	m->lan_signal = "pending";
}

int pigeon_backend_handle_message(pigeon_backend_machine *m, session_msg_type msg)
{
	if (m->state == PIGEON_BACKEND_WAITING_FOR_CLIENT && msg == PIGEON_MSG_PAIR_HELLO && m->guards[PIGEON_GUARD_TOKEN_VALID] && m->guards[PIGEON_GUARD_TOKEN_VALID](m->userdata)) {
		if (m->actions[PIGEON_ACTION_DERIVE_SECRET]) {
			int err = m->actions[PIGEON_ACTION_DERIVE_SECRET](m->userdata);
			if (err) return -err;
		}
		// received_client_pub: recv_msg.pubkey (set by action)
		m->backend_ecdh_pub = "backend_pub";
		if (m->on_change) m->on_change("backend_ecdh_pub", m->userdata);
		// backend_shared_key: DeriveKey("backend_pub", recv_msg.pubkey) (set by action)
		// backend_code: DeriveCode("backend_pub", recv_msg.pubkey) (set by action)
		m->state = PIGEON_BACKEND_DERIVE_SECRET;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_WAITING_FOR_CLIENT && msg == PIGEON_MSG_PAIR_HELLO && m->guards[PIGEON_GUARD_TOKEN_INVALID] && m->guards[PIGEON_GUARD_TOKEN_INVALID](m->userdata)) {
		m->state = PIGEON_BACKEND_IDLE;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_PAIRED && msg == PIGEON_MSG_AUTH_REQUEST) {
		// received_device_id: recv_msg.device_id (set by action)
		// received_auth_nonce: recv_msg.nonce (set by action)
		m->state = PIGEON_BACKEND_AUTH_CHECK;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_OFFERED && msg == PIGEON_MSG_LAN_VERIFY && m->guards[PIGEON_GUARD_CHALLENGE_VALID] && m->guards[PIGEON_GUARD_CHALLENGE_VALID](m->userdata)) {
		if (m->actions[PIGEON_ACTION_ACTIVATE_LAN]) {
			int err = m->actions[PIGEON_ACTION_ACTIVATE_LAN](m->userdata);
			if (err) return -err;
		}
		m->ping_failures = 0;
		if (m->on_change) m->on_change("ping_failures", m->userdata);
		m->backoff_level = 0;
		if (m->on_change) m->on_change("backoff_level", m->userdata);
		m->b_active_path = "lan";
		if (m->on_change) m->on_change("b_active_path", m->userdata);
		m->b_dispatcher_path = "lan";
		if (m->on_change) m->on_change("b_dispatcher_path", m->userdata);
		m->monitor_target = "lan";
		if (m->on_change) m->on_change("monitor_target", m->userdata);
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

int pigeon_backend_step(pigeon_backend_machine *m, session_event_id event)
{
	if (m->state == PIGEON_BACKEND_IDLE && event == PIGEON_EVENT_CLI_INIT_PAIR) {
		if (m->actions[PIGEON_ACTION_GENERATE_TOKEN]) {
			int err = m->actions[PIGEON_ACTION_GENERATE_TOKEN](m->userdata);
			if (err) return -err;
		}
		m->current_token = "tok_1";
		if (m->on_change) m->on_change("current_token", m->userdata);
		// active_tokens: active_tokens \union {"tok_1"} (set by action)
		m->state = PIGEON_BACKEND_GENERATE_TOKEN;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_GENERATE_TOKEN && event == PIGEON_EVENT_TOKEN_CREATED) {
		if (m->actions[PIGEON_ACTION_REGISTER_RELAY]) {
			int err = m->actions[PIGEON_ACTION_REGISTER_RELAY](m->userdata);
			if (err) return -err;
		}
		m->state = PIGEON_BACKEND_REGISTER_RELAY;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_REGISTER_RELAY && event == PIGEON_EVENT_RELAY_REGISTERED) {
		m->secret_published = true;
		if (m->on_change) m->on_change("secret_published", m->userdata);
		m->state = PIGEON_BACKEND_WAITING_FOR_CLIENT;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_DERIVE_SECRET && event == PIGEON_EVENT_ECDH_COMPLETE) {
		m->state = PIGEON_BACKEND_SEND_ACK;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_SEND_ACK && event == PIGEON_EVENT_SIGNAL_CODE_DISPLAY) {
		m->state = PIGEON_BACKEND_WAITING_FOR_CODE;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_WAITING_FOR_CODE && event == PIGEON_EVENT_CLI_CODE_ENTERED) {
		// received_code: cli_entered_code (set by action)
		m->state = PIGEON_BACKEND_VALIDATE_CODE;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_VALIDATE_CODE && event == PIGEON_EVENT_CHECK_CODE && m->guards[PIGEON_GUARD_CODE_CORRECT] && m->guards[PIGEON_GUARD_CODE_CORRECT](m->userdata)) {
		m->state = PIGEON_BACKEND_STORE_PAIRED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_VALIDATE_CODE && event == PIGEON_EVENT_CHECK_CODE && m->guards[PIGEON_GUARD_CODE_WRONG] && m->guards[PIGEON_GUARD_CODE_WRONG](m->userdata)) {
		m->code_attempts = m->code_attempts + 1;
		if (m->on_change) m->on_change("code_attempts", m->userdata);
		m->state = PIGEON_BACKEND_IDLE;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_STORE_PAIRED && event == PIGEON_EVENT_FINALISE) {
		if (m->actions[PIGEON_ACTION_STORE_DEVICE]) {
			int err = m->actions[PIGEON_ACTION_STORE_DEVICE](m->userdata);
			if (err) return -err;
		}
		m->device_secret = "dev_secret_1";
		if (m->on_change) m->on_change("device_secret", m->userdata);
		// paired_devices: paired_devices \union {"device_1"} (set by action)
		// active_tokens: active_tokens \ {current_token} (set by action)
		// used_tokens: used_tokens \union {current_token} (set by action)
		m->state = PIGEON_BACKEND_PAIRED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_AUTH_CHECK && event == PIGEON_EVENT_VERIFY && m->guards[PIGEON_GUARD_DEVICE_KNOWN] && m->guards[PIGEON_GUARD_DEVICE_KNOWN](m->userdata)) {
		if (m->actions[PIGEON_ACTION_VERIFY_DEVICE]) {
			int err = m->actions[PIGEON_ACTION_VERIFY_DEVICE](m->userdata);
			if (err) return -err;
		}
		// auth_nonces_used: auth_nonces_used \union {received_auth_nonce} (set by action)
		m->state = PIGEON_BACKEND_SESSION_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_AUTH_CHECK && event == PIGEON_EVENT_VERIFY && m->guards[PIGEON_GUARD_DEVICE_UNKNOWN] && m->guards[PIGEON_GUARD_DEVICE_UNKNOWN](m->userdata)) {
		m->state = PIGEON_BACKEND_IDLE;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_SESSION_ACTIVE && event == PIGEON_EVENT_SESSION_ESTABLISHED) {
		m->state = PIGEON_BACKEND_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_RELAY_CONNECTED && event == PIGEON_EVENT_LAN_SERVER_READY) {
		m->state = PIGEON_BACKEND_LAN_OFFERED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_OFFERED && event == PIGEON_EVENT_OFFER_TIMEOUT) {
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m->lan_signal = "pending";
		if (m->on_change) m->on_change("lan_signal", m->userdata);
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
	if (m->state == PIGEON_BACKEND_LAN_ACTIVE && event == PIGEON_EVENT_LAN_STREAM_ERROR) {
		if (m->actions[PIGEON_ACTION_FALLBACK_TO_RELAY]) {
			int err = m->actions[PIGEON_ACTION_FALLBACK_TO_RELAY](m->userdata);
			if (err) return -err;
		}
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m->b_active_path = "relay";
		if (m->on_change) m->on_change("b_active_path", m->userdata);
		m->b_dispatcher_path = "relay";
		if (m->on_change) m->on_change("b_dispatcher_path", m->userdata);
		m->monitor_target = "none";
		if (m->on_change) m->on_change("monitor_target", m->userdata);
		m->lan_signal = "pending";
		if (m->on_change) m->on_change("lan_signal", m->userdata);
		m->ping_failures = 0;
		if (m->on_change) m->on_change("ping_failures", m->userdata);
		m->state = PIGEON_BACKEND_RELAY_BACKOFF;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_DEGRADED && event == PIGEON_EVENT_LAN_STREAM_ERROR) {
		if (m->actions[PIGEON_ACTION_FALLBACK_TO_RELAY]) {
			int err = m->actions[PIGEON_ACTION_FALLBACK_TO_RELAY](m->userdata);
			if (err) return -err;
		}
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m->b_active_path = "relay";
		if (m->on_change) m->on_change("b_active_path", m->userdata);
		m->b_dispatcher_path = "relay";
		if (m->on_change) m->on_change("b_dispatcher_path", m->userdata);
		m->monitor_target = "none";
		if (m->on_change) m->on_change("monitor_target", m->userdata);
		m->lan_signal = "pending";
		if (m->on_change) m->on_change("lan_signal", m->userdata);
		m->ping_failures = 0;
		if (m->on_change) m->on_change("ping_failures", m->userdata);
		m->state = PIGEON_BACKEND_RELAY_BACKOFF;
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
		m->b_active_path = "relay";
		if (m->on_change) m->on_change("b_active_path", m->userdata);
		m->b_dispatcher_path = "relay";
		if (m->on_change) m->on_change("b_dispatcher_path", m->userdata);
		m->monitor_target = "none";
		if (m->on_change) m->on_change("monitor_target", m->userdata);
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
	if (m->state == PIGEON_BACKEND_LAN_OFFERED && event == PIGEON_EVENT_APP_FORCE_FALLBACK) {
		m->lan_signal = "pending";
		if (m->on_change) m->on_change("lan_signal", m->userdata);
		m->state = PIGEON_BACKEND_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_ACTIVE && event == PIGEON_EVENT_APP_FORCE_FALLBACK) {
		if (m->actions[PIGEON_ACTION_FALLBACK_TO_RELAY]) {
			int err = m->actions[PIGEON_ACTION_FALLBACK_TO_RELAY](m->userdata);
			if (err) return -err;
		}
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m->b_active_path = "relay";
		if (m->on_change) m->on_change("b_active_path", m->userdata);
		m->b_dispatcher_path = "relay";
		if (m->on_change) m->on_change("b_dispatcher_path", m->userdata);
		m->monitor_target = "none";
		if (m->on_change) m->on_change("monitor_target", m->userdata);
		m->lan_signal = "pending";
		if (m->on_change) m->on_change("lan_signal", m->userdata);
		m->ping_failures = 0;
		if (m->on_change) m->on_change("ping_failures", m->userdata);
		m->state = PIGEON_BACKEND_RELAY_BACKOFF;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_DEGRADED && event == PIGEON_EVENT_APP_FORCE_FALLBACK) {
		if (m->actions[PIGEON_ACTION_FALLBACK_TO_RELAY]) {
			int err = m->actions[PIGEON_ACTION_FALLBACK_TO_RELAY](m->userdata);
			if (err) return -err;
		}
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m->b_active_path = "relay";
		if (m->on_change) m->on_change("b_active_path", m->userdata);
		m->b_dispatcher_path = "relay";
		if (m->on_change) m->on_change("b_dispatcher_path", m->userdata);
		m->monitor_target = "none";
		if (m->on_change) m->on_change("monitor_target", m->userdata);
		m->lan_signal = "pending";
		if (m->on_change) m->on_change("lan_signal", m->userdata);
		m->ping_failures = 0;
		if (m->on_change) m->on_change("ping_failures", m->userdata);
		m->state = PIGEON_BACKEND_RELAY_BACKOFF;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_RELAY_CONNECTED && event == PIGEON_EVENT_DISCONNECT) {
		m->state = PIGEON_BACKEND_PAIRED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_RELAY_CONNECTED && event == PIGEON_EVENT_APP_SEND) {
		m->state = PIGEON_BACKEND_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_OFFERED && event == PIGEON_EVENT_APP_SEND) {
		m->state = PIGEON_BACKEND_LAN_OFFERED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_ACTIVE && event == PIGEON_EVENT_APP_SEND) {
		m->state = PIGEON_BACKEND_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_DEGRADED && event == PIGEON_EVENT_APP_SEND) {
		m->state = PIGEON_BACKEND_LAN_DEGRADED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_RELAY_BACKOFF && event == PIGEON_EVENT_APP_SEND) {
		m->state = PIGEON_BACKEND_RELAY_BACKOFF;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_RELAY_CONNECTED && event == PIGEON_EVENT_RELAY_STREAM_DATA) {
		m->state = PIGEON_BACKEND_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_OFFERED && event == PIGEON_EVENT_RELAY_STREAM_DATA) {
		m->state = PIGEON_BACKEND_LAN_OFFERED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_ACTIVE && event == PIGEON_EVENT_RELAY_STREAM_DATA) {
		m->state = PIGEON_BACKEND_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_DEGRADED && event == PIGEON_EVENT_RELAY_STREAM_DATA) {
		m->state = PIGEON_BACKEND_LAN_DEGRADED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_RELAY_BACKOFF && event == PIGEON_EVENT_RELAY_STREAM_DATA) {
		m->state = PIGEON_BACKEND_RELAY_BACKOFF;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_RELAY_CONNECTED && event == PIGEON_EVENT_RELAY_STREAM_ERROR) {
		m->state = PIGEON_BACKEND_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_OFFERED && event == PIGEON_EVENT_RELAY_STREAM_ERROR) {
		m->state = PIGEON_BACKEND_LAN_OFFERED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_ACTIVE && event == PIGEON_EVENT_RELAY_STREAM_ERROR) {
		m->state = PIGEON_BACKEND_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_DEGRADED && event == PIGEON_EVENT_RELAY_STREAM_ERROR) {
		m->state = PIGEON_BACKEND_LAN_DEGRADED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_RELAY_BACKOFF && event == PIGEON_EVENT_RELAY_STREAM_ERROR) {
		m->state = PIGEON_BACKEND_RELAY_BACKOFF;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_RELAY_CONNECTED && event == PIGEON_EVENT_APP_SEND_DATAGRAM) {
		m->state = PIGEON_BACKEND_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_OFFERED && event == PIGEON_EVENT_APP_SEND_DATAGRAM) {
		m->state = PIGEON_BACKEND_LAN_OFFERED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_ACTIVE && event == PIGEON_EVENT_APP_SEND_DATAGRAM) {
		m->state = PIGEON_BACKEND_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_DEGRADED && event == PIGEON_EVENT_APP_SEND_DATAGRAM) {
		m->state = PIGEON_BACKEND_LAN_DEGRADED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_RELAY_BACKOFF && event == PIGEON_EVENT_APP_SEND_DATAGRAM) {
		m->state = PIGEON_BACKEND_RELAY_BACKOFF;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_RELAY_CONNECTED && event == PIGEON_EVENT_RELAY_DATAGRAM) {
		m->state = PIGEON_BACKEND_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_OFFERED && event == PIGEON_EVENT_RELAY_DATAGRAM) {
		m->state = PIGEON_BACKEND_LAN_OFFERED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_ACTIVE && event == PIGEON_EVENT_RELAY_DATAGRAM) {
		m->state = PIGEON_BACKEND_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_DEGRADED && event == PIGEON_EVENT_RELAY_DATAGRAM) {
		m->state = PIGEON_BACKEND_LAN_DEGRADED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_RELAY_BACKOFF && event == PIGEON_EVENT_RELAY_DATAGRAM) {
		m->state = PIGEON_BACKEND_RELAY_BACKOFF;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_ACTIVE && event == PIGEON_EVENT_LAN_STREAM_DATA) {
		m->state = PIGEON_BACKEND_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_DEGRADED && event == PIGEON_EVENT_LAN_STREAM_DATA) {
		m->state = PIGEON_BACKEND_LAN_DEGRADED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_ACTIVE && event == PIGEON_EVENT_LAN_DATAGRAM) {
		m->state = PIGEON_BACKEND_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_DEGRADED && event == PIGEON_EVENT_LAN_DATAGRAM) {
		m->state = PIGEON_BACKEND_LAN_DEGRADED;
		return 1;
	}
	return 0;
}

void pigeon_client_machine_init(pigeon_client_machine *m)
{
	memset(m, 0, sizeof(*m));
	m->state = PIGEON_CLIENT_IDLE;
	m->received_backend_pub = "none";
	m->c_active_path = "relay";
	m->c_dispatcher_path = "relay";
	m->lan_signal = "pending";
}

int pigeon_client_handle_message(pigeon_client_machine *m, session_msg_type msg)
{
	if (m->state == PIGEON_CLIENT_WAIT_ACK && msg == PIGEON_MSG_PAIR_HELLO_ACK) {
		if (m->actions[PIGEON_ACTION_DERIVE_SECRET]) {
			int err = m->actions[PIGEON_ACTION_DERIVE_SECRET](m->userdata);
			if (err) return -err;
		}
		// received_backend_pub: recv_msg.pubkey (set by action)
		// client_shared_key: DeriveKey("client_pub", recv_msg.pubkey) (set by action)
		m->state = PIGEON_CLIENT_E2E_READY;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_E2E_READY && msg == PIGEON_MSG_PAIR_CONFIRM) {
		// client_code: DeriveCode(received_backend_pub, "client_pub") (set by action)
		m->state = PIGEON_CLIENT_SHOW_CODE;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_WAIT_PAIR_COMPLETE && msg == PIGEON_MSG_PAIR_COMPLETE) {
		if (m->actions[PIGEON_ACTION_STORE_SECRET]) {
			int err = m->actions[PIGEON_ACTION_STORE_SECRET](m->userdata);
			if (err) return -err;
		}
		m->state = PIGEON_CLIENT_PAIRED;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_SEND_AUTH && msg == PIGEON_MSG_AUTH_OK) {
		m->state = PIGEON_CLIENT_SESSION_ACTIVE;
		return 1;
	}
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
		m->c_active_path = "lan";
		if (m->on_change) m->on_change("c_active_path", m->userdata);
		m->c_dispatcher_path = "lan";
		if (m->on_change) m->on_change("c_dispatcher_path", m->userdata);
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

int pigeon_client_step(pigeon_client_machine *m, session_event_id event)
{
	if (m->state == PIGEON_CLIENT_IDLE && event == PIGEON_EVENT_BACKCHANNEL_RECEIVED) {
		m->state = PIGEON_CLIENT_OBTAIN_BACKCHANNEL_SECRET;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_OBTAIN_BACKCHANNEL_SECRET && event == PIGEON_EVENT_SECRET_PARSED) {
		m->state = PIGEON_CLIENT_CONNECT_RELAY;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_CONNECT_RELAY && event == PIGEON_EVENT_RELAY_CONNECTED) {
		m->state = PIGEON_CLIENT_GEN_KEY_PAIR;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_GEN_KEY_PAIR && event == PIGEON_EVENT_KEY_PAIR_GENERATED) {
		if (m->actions[PIGEON_ACTION_SEND_PAIR_HELLO]) {
			int err = m->actions[PIGEON_ACTION_SEND_PAIR_HELLO](m->userdata);
			if (err) return -err;
		}
		m->state = PIGEON_CLIENT_WAIT_ACK;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_SHOW_CODE && event == PIGEON_EVENT_CODE_DISPLAYED) {
		m->state = PIGEON_CLIENT_WAIT_PAIR_COMPLETE;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_PAIRED && event == PIGEON_EVENT_APP_LAUNCH) {
		m->state = PIGEON_CLIENT_RECONNECT;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_RECONNECT && event == PIGEON_EVENT_RELAY_CONNECTED) {
		m->state = PIGEON_CLIENT_SEND_AUTH;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_SESSION_ACTIVE && event == PIGEON_EVENT_SESSION_ESTABLISHED) {
		m->state = PIGEON_CLIENT_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_CONNECTING && event == PIGEON_EVENT_LAN_DIAL_OK) {
		m->state = PIGEON_CLIENT_LAN_VERIFYING;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_CONNECTING && event == PIGEON_EVENT_LAN_DIAL_FAILED) {
		m->state = PIGEON_CLIENT_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_VERIFYING && event == PIGEON_EVENT_VERIFY_TIMEOUT) {
		m->c_dispatcher_path = "relay";
		if (m->on_change) m->on_change("c_dispatcher_path", m->userdata);
		m->state = PIGEON_CLIENT_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_ACTIVE && event == PIGEON_EVENT_LAN_ERROR) {
		if (m->actions[PIGEON_ACTION_FALLBACK_TO_RELAY]) {
			int err = m->actions[PIGEON_ACTION_FALLBACK_TO_RELAY](m->userdata);
			if (err) return -err;
		}
		m->c_active_path = "relay";
		if (m->on_change) m->on_change("c_active_path", m->userdata);
		m->c_dispatcher_path = "relay";
		if (m->on_change) m->on_change("c_dispatcher_path", m->userdata);
		m->lan_signal = "pending";
		if (m->on_change) m->on_change("lan_signal", m->userdata);
		m->state = PIGEON_CLIENT_RELAY_FALLBACK;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_ACTIVE && event == PIGEON_EVENT_LAN_STREAM_ERROR) {
		if (m->actions[PIGEON_ACTION_FALLBACK_TO_RELAY]) {
			int err = m->actions[PIGEON_ACTION_FALLBACK_TO_RELAY](m->userdata);
			if (err) return -err;
		}
		m->c_active_path = "relay";
		if (m->on_change) m->on_change("c_active_path", m->userdata);
		m->c_dispatcher_path = "relay";
		if (m->on_change) m->on_change("c_dispatcher_path", m->userdata);
		m->lan_signal = "pending";
		if (m->on_change) m->on_change("lan_signal", m->userdata);
		m->state = PIGEON_CLIENT_RELAY_FALLBACK;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_RELAY_FALLBACK && event == PIGEON_EVENT_RELAY_OK) {
		m->state = PIGEON_CLIENT_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_CONNECTING && event == PIGEON_EVENT_APP_FORCE_FALLBACK) {
		m->state = PIGEON_CLIENT_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_VERIFYING && event == PIGEON_EVENT_APP_FORCE_FALLBACK) {
		m->c_dispatcher_path = "relay";
		if (m->on_change) m->on_change("c_dispatcher_path", m->userdata);
		m->state = PIGEON_CLIENT_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_ACTIVE && event == PIGEON_EVENT_APP_FORCE_FALLBACK) {
		if (m->actions[PIGEON_ACTION_FALLBACK_TO_RELAY]) {
			int err = m->actions[PIGEON_ACTION_FALLBACK_TO_RELAY](m->userdata);
			if (err) return -err;
		}
		m->c_active_path = "relay";
		if (m->on_change) m->on_change("c_active_path", m->userdata);
		m->c_dispatcher_path = "relay";
		if (m->on_change) m->on_change("c_dispatcher_path", m->userdata);
		m->lan_signal = "pending";
		if (m->on_change) m->on_change("lan_signal", m->userdata);
		m->state = PIGEON_CLIENT_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_RELAY_CONNECTED && event == PIGEON_EVENT_DISCONNECT) {
		m->state = PIGEON_CLIENT_PAIRED;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_RELAY_CONNECTED && event == PIGEON_EVENT_APP_SEND) {
		m->state = PIGEON_CLIENT_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_CONNECTING && event == PIGEON_EVENT_APP_SEND) {
		m->state = PIGEON_CLIENT_LAN_CONNECTING;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_VERIFYING && event == PIGEON_EVENT_APP_SEND) {
		m->state = PIGEON_CLIENT_LAN_VERIFYING;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_ACTIVE && event == PIGEON_EVENT_APP_SEND) {
		m->state = PIGEON_CLIENT_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_RELAY_FALLBACK && event == PIGEON_EVENT_APP_SEND) {
		m->state = PIGEON_CLIENT_RELAY_FALLBACK;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_RELAY_CONNECTED && event == PIGEON_EVENT_RELAY_STREAM_DATA) {
		m->state = PIGEON_CLIENT_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_CONNECTING && event == PIGEON_EVENT_RELAY_STREAM_DATA) {
		m->state = PIGEON_CLIENT_LAN_CONNECTING;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_VERIFYING && event == PIGEON_EVENT_RELAY_STREAM_DATA) {
		m->state = PIGEON_CLIENT_LAN_VERIFYING;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_ACTIVE && event == PIGEON_EVENT_RELAY_STREAM_DATA) {
		m->state = PIGEON_CLIENT_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_RELAY_FALLBACK && event == PIGEON_EVENT_RELAY_STREAM_DATA) {
		m->state = PIGEON_CLIENT_RELAY_FALLBACK;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_RELAY_CONNECTED && event == PIGEON_EVENT_RELAY_STREAM_ERROR) {
		m->state = PIGEON_CLIENT_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_CONNECTING && event == PIGEON_EVENT_RELAY_STREAM_ERROR) {
		m->state = PIGEON_CLIENT_LAN_CONNECTING;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_VERIFYING && event == PIGEON_EVENT_RELAY_STREAM_ERROR) {
		m->state = PIGEON_CLIENT_LAN_VERIFYING;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_ACTIVE && event == PIGEON_EVENT_RELAY_STREAM_ERROR) {
		m->state = PIGEON_CLIENT_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_RELAY_FALLBACK && event == PIGEON_EVENT_RELAY_STREAM_ERROR) {
		m->state = PIGEON_CLIENT_RELAY_FALLBACK;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_RELAY_CONNECTED && event == PIGEON_EVENT_APP_SEND_DATAGRAM) {
		m->state = PIGEON_CLIENT_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_CONNECTING && event == PIGEON_EVENT_APP_SEND_DATAGRAM) {
		m->state = PIGEON_CLIENT_LAN_CONNECTING;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_VERIFYING && event == PIGEON_EVENT_APP_SEND_DATAGRAM) {
		m->state = PIGEON_CLIENT_LAN_VERIFYING;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_ACTIVE && event == PIGEON_EVENT_APP_SEND_DATAGRAM) {
		m->state = PIGEON_CLIENT_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_RELAY_FALLBACK && event == PIGEON_EVENT_APP_SEND_DATAGRAM) {
		m->state = PIGEON_CLIENT_RELAY_FALLBACK;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_RELAY_CONNECTED && event == PIGEON_EVENT_RELAY_DATAGRAM) {
		m->state = PIGEON_CLIENT_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_CONNECTING && event == PIGEON_EVENT_RELAY_DATAGRAM) {
		m->state = PIGEON_CLIENT_LAN_CONNECTING;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_VERIFYING && event == PIGEON_EVENT_RELAY_DATAGRAM) {
		m->state = PIGEON_CLIENT_LAN_VERIFYING;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_ACTIVE && event == PIGEON_EVENT_RELAY_DATAGRAM) {
		m->state = PIGEON_CLIENT_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_RELAY_FALLBACK && event == PIGEON_EVENT_RELAY_DATAGRAM) {
		m->state = PIGEON_CLIENT_RELAY_FALLBACK;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_ACTIVE && event == PIGEON_EVENT_LAN_STREAM_DATA) {
		m->state = PIGEON_CLIENT_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_ACTIVE && event == PIGEON_EVENT_LAN_DATAGRAM) {
		m->state = PIGEON_CLIENT_LAN_ACTIVE;
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

int pigeon_relay_handle_message(pigeon_relay_machine *m, session_msg_type msg)
{
	return 0;
}

int pigeon_relay_step(pigeon_relay_machine *m, session_event_id event)
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

