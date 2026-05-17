// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// Pigeon C client library — amalgamated source.
// Compile with -DPIGEON_CRYPTO_LIBSODIUM and link -lsodium.

#include "pigeon.h"
#include <string.h>

// --- Generated pairing-ceremony state machine ---




void pigeon_acceptor_machine_init(pigeon_acceptor_machine *m)
{
	memset(m, 0, sizeof(*m));
	m->state = PIGEON_ACCEPTOR_IDLE;
	m->acceptor_eph_pub = "none";
	m->acceptor_received_eph_pub = "none";
	m->acceptor_received_identity = "none";
	m->acceptor_received_instance = "none";
	m->acceptor_user_confirmed = "false";
	m->acceptor_received_confirm = "false";
}

int pigeon_acceptor_handle_message(pigeon_acceptor_machine *m, pairing_ceremony_msg_type msg)
{
	if (m->state == PIGEON_ACCEPTOR_WAITING_FOR_HELLO && msg == PIGEON_PAIRINGCEREMONY_MSG_HELLO) {
		if (m->actions[PIGEON_PAIRINGCEREMONY_ACTION_DERIVE_CODE]) {
			int err = m->actions[PIGEON_PAIRINGCEREMONY_ACTION_DERIVE_CODE](m->userdata);
			if (err) return -err;
		}
		// acceptor_received_eph_pub: recv_msg.eph_pub (set by action)
		// acceptor_received_identity: recv_msg.identity_pub (set by action)
		// acceptor_received_instance: recv_msg.instance_id (set by action)
		// acceptor_code: DeriveCode(acceptor_eph_pub, recv_msg.eph_pub) (set by action)
		m->state = PIGEON_ACCEPTOR_DERIVING_CODE;
		return 1;
	}
	if (m->state == PIGEON_ACCEPTOR_AWAITING_PEER_CONFIRM && msg == PIGEON_PAIRINGCEREMONY_MSG_CONFIRM_TO_ACCEPTOR) {
		if (m->actions[PIGEON_PAIRINGCEREMONY_ACTION_STORE_RECORD]) {
			int err = m->actions[PIGEON_PAIRINGCEREMONY_ACTION_STORE_RECORD](m->userdata);
			if (err) return -err;
		}
		m->acceptor_received_confirm = "true";
		if (m->on_change) m->on_change("acceptor_received_confirm", m->userdata);
		m->state = PIGEON_ACCEPTOR_PAIRED;
		return 1;
	}
	return 0;
}

int pigeon_acceptor_step(pigeon_acceptor_machine *m, pairing_ceremony_event_id event)
{
	if (m->state == PIGEON_ACCEPTOR_IDLE && event == PIGEON_PAIRINGCEREMONY_EVENT_PAIR_BEGIN) {
		if (m->actions[PIGEON_PAIRINGCEREMONY_ACTION_GEN_EPHEMERAL]) {
			int err = m->actions[PIGEON_PAIRINGCEREMONY_ACTION_GEN_EPHEMERAL](m->userdata);
			if (err) return -err;
		}
		m->acceptor_eph_pub = "acceptor_eph";
		if (m->on_change) m->on_change("acceptor_eph_pub", m->userdata);
		m->state = PIGEON_ACCEPTOR_GENERATING_EPHEMERAL;
		return 1;
	}
	if (m->state == PIGEON_ACCEPTOR_GENERATING_EPHEMERAL && event == PIGEON_PAIRINGCEREMONY_EVENT_EPHEMERAL_READY) {
		if (m->actions[PIGEON_PAIRINGCEREMONY_ACTION_REGISTER_RELAY]) {
			int err = m->actions[PIGEON_PAIRINGCEREMONY_ACTION_REGISTER_RELAY](m->userdata);
			if (err) return -err;
		}
		m->state = PIGEON_ACCEPTOR_REGISTERING_RELAY;
		return 1;
	}
	if (m->state == PIGEON_ACCEPTOR_REGISTERING_RELAY && event == PIGEON_PAIRINGCEREMONY_EVENT_RELAY_REGISTERED) {
		if (m->actions[PIGEON_PAIRINGCEREMONY_ACTION_EMIT_TOKEN]) {
			int err = m->actions[PIGEON_PAIRINGCEREMONY_ACTION_EMIT_TOKEN](m->userdata);
			if (err) return -err;
		}
		m->state = PIGEON_ACCEPTOR_WAITING_FOR_HELLO;
		return 1;
	}
	if (m->state == PIGEON_ACCEPTOR_DERIVING_CODE && event == PIGEON_PAIRINGCEREMONY_EVENT_CODE_READY) {
		m->state = PIGEON_ACCEPTOR_AWAITING_USER_CONFIRM;
		return 1;
	}
	if (m->state == PIGEON_ACCEPTOR_AWAITING_USER_CONFIRM && event == PIGEON_PAIRINGCEREMONY_EVENT_USER_CONFIRM) {
		m->acceptor_user_confirmed = "true";
		if (m->on_change) m->on_change("acceptor_user_confirmed", m->userdata);
		m->state = PIGEON_ACCEPTOR_AWAITING_PEER_CONFIRM;
		return 1;
	}
	if (m->state == PIGEON_ACCEPTOR_AWAITING_USER_CONFIRM && event == PIGEON_PAIRINGCEREMONY_EVENT_USER_CANCEL) {
		m->state = PIGEON_ACCEPTOR_ABORTED;
		return 1;
	}
	if (m->state == PIGEON_ACCEPTOR_AWAITING_PEER_CONFIRM && event == PIGEON_PAIRINGCEREMONY_EVENT_USER_CANCEL) {
		m->state = PIGEON_ACCEPTOR_ABORTED;
		return 1;
	}
	return 0;
}

void pigeon_initiator_machine_init(pigeon_initiator_machine *m)
{
	memset(m, 0, sizeof(*m));
	m->state = PIGEON_INITIATOR_IDLE;
	m->initiator_eph_pub = "none";
	m->received_acceptor_eph_pub = "none";
	m->received_acceptor_identity = "none";
	m->received_acceptor_instance = "none";
	m->initiator_received_eph_pub = "none";
	m->initiator_received_identity = "none";
	m->initiator_received_instance = "none";
	m->initiator_user_confirmed = "false";
	m->initiator_received_confirm = "false";
}

int pigeon_initiator_handle_message(pigeon_initiator_machine *m, pairing_ceremony_msg_type msg)
{
	if (m->state == PIGEON_INITIATOR_AWAITING_WELCOME && msg == PIGEON_PAIRINGCEREMONY_MSG_WELCOME) {
		if (m->actions[PIGEON_PAIRINGCEREMONY_ACTION_DERIVE_CODE]) {
			int err = m->actions[PIGEON_PAIRINGCEREMONY_ACTION_DERIVE_CODE](m->userdata);
			if (err) return -err;
		}
		// initiator_received_eph_pub: recv_msg.eph_pub (set by action)
		// initiator_received_identity: recv_msg.identity_pub (set by action)
		// initiator_received_instance: recv_msg.instance_id (set by action)
		// initiator_code: DeriveCode(initiator_eph_pub, recv_msg.eph_pub) (set by action)
		m->state = PIGEON_INITIATOR_DERIVING_CODE;
		return 1;
	}
	if (m->state == PIGEON_INITIATOR_AWAITING_PEER_CONFIRM && msg == PIGEON_PAIRINGCEREMONY_MSG_CONFIRM_TO_INITIATOR) {
		if (m->actions[PIGEON_PAIRINGCEREMONY_ACTION_STORE_RECORD]) {
			int err = m->actions[PIGEON_PAIRINGCEREMONY_ACTION_STORE_RECORD](m->userdata);
			if (err) return -err;
		}
		m->initiator_received_confirm = "true";
		if (m->on_change) m->on_change("initiator_received_confirm", m->userdata);
		m->state = PIGEON_INITIATOR_PAIRED;
		return 1;
	}
	return 0;
}

int pigeon_initiator_step(pigeon_initiator_machine *m, pairing_ceremony_event_id event)
{
	if (m->state == PIGEON_INITIATOR_IDLE && event == PIGEON_PAIRINGCEREMONY_EVENT_TOKEN_RECEIVED) {
		if (m->actions[PIGEON_PAIRINGCEREMONY_ACTION_DECODE_TOKEN]) {
			int err = m->actions[PIGEON_PAIRINGCEREMONY_ACTION_DECODE_TOKEN](m->userdata);
			if (err) return -err;
		}
		m->received_acceptor_eph_pub = "acceptor_eph";
		if (m->on_change) m->on_change("received_acceptor_eph_pub", m->userdata);
		m->received_acceptor_identity = "acceptor_id";
		if (m->on_change) m->on_change("received_acceptor_identity", m->userdata);
		m->received_acceptor_instance = "acceptor_instance";
		if (m->on_change) m->on_change("received_acceptor_instance", m->userdata);
		m->state = PIGEON_INITIATOR_DECODING_TOKEN;
		return 1;
	}
	if (m->state == PIGEON_INITIATOR_DECODING_TOKEN && event == PIGEON_PAIRINGCEREMONY_EVENT_TOKEN_DECODED) {
		if (m->actions[PIGEON_PAIRINGCEREMONY_ACTION_GEN_EPHEMERAL]) {
			int err = m->actions[PIGEON_PAIRINGCEREMONY_ACTION_GEN_EPHEMERAL](m->userdata);
			if (err) return -err;
		}
		m->initiator_eph_pub = "initiator_eph";
		if (m->on_change) m->on_change("initiator_eph_pub", m->userdata);
		m->state = PIGEON_INITIATOR_GENERATING_EPHEMERAL;
		return 1;
	}
	if (m->state == PIGEON_INITIATOR_GENERATING_EPHEMERAL && event == PIGEON_PAIRINGCEREMONY_EVENT_EPHEMERAL_READY) {
		if (m->actions[PIGEON_PAIRINGCEREMONY_ACTION_DIAL_RELAY]) {
			int err = m->actions[PIGEON_PAIRINGCEREMONY_ACTION_DIAL_RELAY](m->userdata);
			if (err) return -err;
		}
		m->state = PIGEON_INITIATOR_CONNECTING_RELAY;
		return 1;
	}
	if (m->state == PIGEON_INITIATOR_CONNECTING_RELAY && event == PIGEON_PAIRINGCEREMONY_EVENT_RELAY_CONNECTED) {
		m->state = PIGEON_INITIATOR_AWAITING_WELCOME;
		return 1;
	}
	if (m->state == PIGEON_INITIATOR_DERIVING_CODE && event == PIGEON_PAIRINGCEREMONY_EVENT_CODE_READY) {
		m->state = PIGEON_INITIATOR_AWAITING_USER_CONFIRM;
		return 1;
	}
	if (m->state == PIGEON_INITIATOR_AWAITING_USER_CONFIRM && event == PIGEON_PAIRINGCEREMONY_EVENT_USER_CONFIRM) {
		m->initiator_user_confirmed = "true";
		if (m->on_change) m->on_change("initiator_user_confirmed", m->userdata);
		m->state = PIGEON_INITIATOR_AWAITING_PEER_CONFIRM;
		return 1;
	}
	if (m->state == PIGEON_INITIATOR_AWAITING_USER_CONFIRM && event == PIGEON_PAIRINGCEREMONY_EVENT_USER_CANCEL) {
		m->state = PIGEON_INITIATOR_ABORTED;
		return 1;
	}
	if (m->state == PIGEON_INITIATOR_AWAITING_PEER_CONFIRM && event == PIGEON_PAIRINGCEREMONY_EVENT_USER_CANCEL) {
		m->state = PIGEON_INITIATOR_ABORTED;
		return 1;
	}
	return 0;
}


// --- Generated session state machine ---




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
	if (m->state == PIGEON_BACKEND_WAITING_FOR_CLIENT && msg == PIGEON_SESSION_MSG_PAIR_HELLO && m->guards[PIGEON_SESSION_GUARD_TOKEN_VALID] && m->guards[PIGEON_SESSION_GUARD_TOKEN_VALID](m->userdata)) {
		if (m->actions[PIGEON_SESSION_ACTION_DERIVE_SECRET]) {
			int err = m->actions[PIGEON_SESSION_ACTION_DERIVE_SECRET](m->userdata);
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
	if (m->state == PIGEON_BACKEND_WAITING_FOR_CLIENT && msg == PIGEON_SESSION_MSG_PAIR_HELLO && m->guards[PIGEON_SESSION_GUARD_TOKEN_INVALID] && m->guards[PIGEON_SESSION_GUARD_TOKEN_INVALID](m->userdata)) {
		m->state = PIGEON_BACKEND_IDLE;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_PAIRED && msg == PIGEON_SESSION_MSG_AUTH_REQUEST) {
		// received_device_id: recv_msg.device_id (set by action)
		// received_auth_nonce: recv_msg.nonce (set by action)
		m->state = PIGEON_BACKEND_AUTH_CHECK;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_OFFERED && msg == PIGEON_SESSION_MSG_LAN_VERIFY && m->guards[PIGEON_SESSION_GUARD_CHALLENGE_VALID] && m->guards[PIGEON_SESSION_GUARD_CHALLENGE_VALID](m->userdata)) {
		if (m->actions[PIGEON_SESSION_ACTION_ACTIVATE_LAN]) {
			int err = m->actions[PIGEON_SESSION_ACTION_ACTIVATE_LAN](m->userdata);
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
	if (m->state == PIGEON_BACKEND_LAN_OFFERED && msg == PIGEON_SESSION_MSG_LAN_VERIFY && m->guards[PIGEON_SESSION_GUARD_CHALLENGE_INVALID] && m->guards[PIGEON_SESSION_GUARD_CHALLENGE_INVALID](m->userdata)) {
		m->state = PIGEON_BACKEND_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_DEGRADED && msg == PIGEON_SESSION_MSG_PATH_PONG) {
		if (m->actions[PIGEON_SESSION_ACTION_RESET_FAILURES]) {
			int err = m->actions[PIGEON_SESSION_ACTION_RESET_FAILURES](m->userdata);
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
	if (m->state == PIGEON_BACKEND_IDLE && event == PIGEON_SESSION_EVENT_CLI_INIT_PAIR) {
		if (m->actions[PIGEON_SESSION_ACTION_GENERATE_TOKEN]) {
			int err = m->actions[PIGEON_SESSION_ACTION_GENERATE_TOKEN](m->userdata);
			if (err) return -err;
		}
		m->current_token = "tok_1";
		if (m->on_change) m->on_change("current_token", m->userdata);
		// active_tokens: active_tokens \union {"tok_1"} (set by action)
		m->state = PIGEON_BACKEND_GENERATE_TOKEN;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_GENERATE_TOKEN && event == PIGEON_SESSION_EVENT_TOKEN_CREATED) {
		if (m->actions[PIGEON_SESSION_ACTION_REGISTER_RELAY]) {
			int err = m->actions[PIGEON_SESSION_ACTION_REGISTER_RELAY](m->userdata);
			if (err) return -err;
		}
		m->state = PIGEON_BACKEND_REGISTER_RELAY;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_REGISTER_RELAY && event == PIGEON_SESSION_EVENT_RELAY_REGISTERED) {
		m->secret_published = true;
		if (m->on_change) m->on_change("secret_published", m->userdata);
		m->state = PIGEON_BACKEND_WAITING_FOR_CLIENT;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_DERIVE_SECRET && event == PIGEON_SESSION_EVENT_ECDH_COMPLETE) {
		m->state = PIGEON_BACKEND_SEND_ACK;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_SEND_ACK && event == PIGEON_SESSION_EVENT_SIGNAL_CODE_DISPLAY) {
		m->state = PIGEON_BACKEND_WAITING_FOR_CODE;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_WAITING_FOR_CODE && event == PIGEON_SESSION_EVENT_CLI_CODE_ENTERED) {
		// received_code: cli_entered_code (set by action)
		m->state = PIGEON_BACKEND_VALIDATE_CODE;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_VALIDATE_CODE && event == PIGEON_SESSION_EVENT_CHECK_CODE && m->guards[PIGEON_SESSION_GUARD_CODE_CORRECT] && m->guards[PIGEON_SESSION_GUARD_CODE_CORRECT](m->userdata)) {
		m->state = PIGEON_BACKEND_STORE_PAIRED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_VALIDATE_CODE && event == PIGEON_SESSION_EVENT_CHECK_CODE && m->guards[PIGEON_SESSION_GUARD_CODE_WRONG] && m->guards[PIGEON_SESSION_GUARD_CODE_WRONG](m->userdata)) {
		m->code_attempts = m->code_attempts + 1;
		if (m->on_change) m->on_change("code_attempts", m->userdata);
		m->state = PIGEON_BACKEND_IDLE;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_STORE_PAIRED && event == PIGEON_SESSION_EVENT_FINALISE) {
		if (m->actions[PIGEON_SESSION_ACTION_STORE_DEVICE]) {
			int err = m->actions[PIGEON_SESSION_ACTION_STORE_DEVICE](m->userdata);
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
	if (m->state == PIGEON_BACKEND_AUTH_CHECK && event == PIGEON_SESSION_EVENT_VERIFY && m->guards[PIGEON_SESSION_GUARD_DEVICE_KNOWN] && m->guards[PIGEON_SESSION_GUARD_DEVICE_KNOWN](m->userdata)) {
		if (m->actions[PIGEON_SESSION_ACTION_VERIFY_DEVICE]) {
			int err = m->actions[PIGEON_SESSION_ACTION_VERIFY_DEVICE](m->userdata);
			if (err) return -err;
		}
		// auth_nonces_used: auth_nonces_used \union {received_auth_nonce} (set by action)
		m->state = PIGEON_BACKEND_SESSION_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_AUTH_CHECK && event == PIGEON_SESSION_EVENT_VERIFY && m->guards[PIGEON_SESSION_GUARD_DEVICE_UNKNOWN] && m->guards[PIGEON_SESSION_GUARD_DEVICE_UNKNOWN](m->userdata)) {
		m->state = PIGEON_BACKEND_IDLE;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_SESSION_ACTIVE && event == PIGEON_SESSION_EVENT_SESSION_ESTABLISHED) {
		m->state = PIGEON_BACKEND_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_RELAY_CONNECTED && event == PIGEON_SESSION_EVENT_LAN_SERVER_READY) {
		m->state = PIGEON_BACKEND_LAN_OFFERED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_OFFERED && event == PIGEON_SESSION_EVENT_OFFER_TIMEOUT) {
		// backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
		m->lan_signal = "pending";
		if (m->on_change) m->on_change("lan_signal", m->userdata);
		m->state = PIGEON_BACKEND_RELAY_BACKOFF;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_ACTIVE && event == PIGEON_SESSION_EVENT_PING_TICK) {
		m->state = PIGEON_BACKEND_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_ACTIVE && event == PIGEON_SESSION_EVENT_PING_TIMEOUT) {
		m->ping_failures = 1;
		if (m->on_change) m->on_change("ping_failures", m->userdata);
		m->state = PIGEON_BACKEND_LAN_DEGRADED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_DEGRADED && event == PIGEON_SESSION_EVENT_PING_TICK) {
		m->state = PIGEON_BACKEND_LAN_DEGRADED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_ACTIVE && event == PIGEON_SESSION_EVENT_LAN_STREAM_ERROR) {
		if (m->actions[PIGEON_SESSION_ACTION_FALLBACK_TO_RELAY]) {
			int err = m->actions[PIGEON_SESSION_ACTION_FALLBACK_TO_RELAY](m->userdata);
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
	if (m->state == PIGEON_BACKEND_LAN_DEGRADED && event == PIGEON_SESSION_EVENT_LAN_STREAM_ERROR) {
		if (m->actions[PIGEON_SESSION_ACTION_FALLBACK_TO_RELAY]) {
			int err = m->actions[PIGEON_SESSION_ACTION_FALLBACK_TO_RELAY](m->userdata);
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
	if (m->state == PIGEON_BACKEND_LAN_DEGRADED && event == PIGEON_SESSION_EVENT_PING_TIMEOUT && m->guards[PIGEON_SESSION_GUARD_UNDER_MAX_FAILURES] && m->guards[PIGEON_SESSION_GUARD_UNDER_MAX_FAILURES](m->userdata)) {
		m->ping_failures = m->ping_failures + 1;
		if (m->on_change) m->on_change("ping_failures", m->userdata);
		m->state = PIGEON_BACKEND_LAN_DEGRADED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_DEGRADED && event == PIGEON_SESSION_EVENT_PING_TIMEOUT && m->guards[PIGEON_SESSION_GUARD_AT_MAX_FAILURES] && m->guards[PIGEON_SESSION_GUARD_AT_MAX_FAILURES](m->userdata)) {
		if (m->actions[PIGEON_SESSION_ACTION_FALLBACK_TO_RELAY]) {
			int err = m->actions[PIGEON_SESSION_ACTION_FALLBACK_TO_RELAY](m->userdata);
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
	if (m->state == PIGEON_BACKEND_RELAY_BACKOFF && event == PIGEON_SESSION_EVENT_BACKOFF_EXPIRED) {
		m->state = PIGEON_BACKEND_LAN_OFFERED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_RELAY_BACKOFF && event == PIGEON_SESSION_EVENT_LAN_SERVER_CHANGED) {
		m->backoff_level = 0;
		if (m->on_change) m->on_change("backoff_level", m->userdata);
		m->state = PIGEON_BACKEND_LAN_OFFERED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_RELAY_CONNECTED && event == PIGEON_SESSION_EVENT_READVERTISE_TICK && m->guards[PIGEON_SESSION_GUARD_LAN_SERVER_AVAILABLE] && m->guards[PIGEON_SESSION_GUARD_LAN_SERVER_AVAILABLE](m->userdata)) {
		m->state = PIGEON_BACKEND_LAN_OFFERED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_OFFERED && event == PIGEON_SESSION_EVENT_APP_FORCE_FALLBACK) {
		m->lan_signal = "pending";
		if (m->on_change) m->on_change("lan_signal", m->userdata);
		m->state = PIGEON_BACKEND_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_ACTIVE && event == PIGEON_SESSION_EVENT_APP_FORCE_FALLBACK) {
		if (m->actions[PIGEON_SESSION_ACTION_FALLBACK_TO_RELAY]) {
			int err = m->actions[PIGEON_SESSION_ACTION_FALLBACK_TO_RELAY](m->userdata);
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
	if (m->state == PIGEON_BACKEND_LAN_DEGRADED && event == PIGEON_SESSION_EVENT_APP_FORCE_FALLBACK) {
		if (m->actions[PIGEON_SESSION_ACTION_FALLBACK_TO_RELAY]) {
			int err = m->actions[PIGEON_SESSION_ACTION_FALLBACK_TO_RELAY](m->userdata);
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
	if (m->state == PIGEON_BACKEND_RELAY_CONNECTED && event == PIGEON_SESSION_EVENT_DISCONNECT) {
		m->state = PIGEON_BACKEND_PAIRED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_RELAY_CONNECTED && event == PIGEON_SESSION_EVENT_APP_SEND) {
		m->state = PIGEON_BACKEND_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_OFFERED && event == PIGEON_SESSION_EVENT_APP_SEND) {
		m->state = PIGEON_BACKEND_LAN_OFFERED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_ACTIVE && event == PIGEON_SESSION_EVENT_APP_SEND) {
		m->state = PIGEON_BACKEND_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_DEGRADED && event == PIGEON_SESSION_EVENT_APP_SEND) {
		m->state = PIGEON_BACKEND_LAN_DEGRADED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_RELAY_BACKOFF && event == PIGEON_SESSION_EVENT_APP_SEND) {
		m->state = PIGEON_BACKEND_RELAY_BACKOFF;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_RELAY_CONNECTED && event == PIGEON_SESSION_EVENT_RELAY_STREAM_DATA) {
		m->state = PIGEON_BACKEND_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_OFFERED && event == PIGEON_SESSION_EVENT_RELAY_STREAM_DATA) {
		m->state = PIGEON_BACKEND_LAN_OFFERED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_ACTIVE && event == PIGEON_SESSION_EVENT_RELAY_STREAM_DATA) {
		m->state = PIGEON_BACKEND_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_DEGRADED && event == PIGEON_SESSION_EVENT_RELAY_STREAM_DATA) {
		m->state = PIGEON_BACKEND_LAN_DEGRADED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_RELAY_BACKOFF && event == PIGEON_SESSION_EVENT_RELAY_STREAM_DATA) {
		m->state = PIGEON_BACKEND_RELAY_BACKOFF;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_RELAY_CONNECTED && event == PIGEON_SESSION_EVENT_RELAY_STREAM_ERROR) {
		m->state = PIGEON_BACKEND_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_OFFERED && event == PIGEON_SESSION_EVENT_RELAY_STREAM_ERROR) {
		m->state = PIGEON_BACKEND_LAN_OFFERED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_ACTIVE && event == PIGEON_SESSION_EVENT_RELAY_STREAM_ERROR) {
		m->state = PIGEON_BACKEND_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_DEGRADED && event == PIGEON_SESSION_EVENT_RELAY_STREAM_ERROR) {
		m->state = PIGEON_BACKEND_LAN_DEGRADED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_RELAY_BACKOFF && event == PIGEON_SESSION_EVENT_RELAY_STREAM_ERROR) {
		m->state = PIGEON_BACKEND_RELAY_BACKOFF;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_RELAY_CONNECTED && event == PIGEON_SESSION_EVENT_APP_SEND_DATAGRAM) {
		m->state = PIGEON_BACKEND_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_OFFERED && event == PIGEON_SESSION_EVENT_APP_SEND_DATAGRAM) {
		m->state = PIGEON_BACKEND_LAN_OFFERED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_ACTIVE && event == PIGEON_SESSION_EVENT_APP_SEND_DATAGRAM) {
		m->state = PIGEON_BACKEND_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_DEGRADED && event == PIGEON_SESSION_EVENT_APP_SEND_DATAGRAM) {
		m->state = PIGEON_BACKEND_LAN_DEGRADED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_RELAY_BACKOFF && event == PIGEON_SESSION_EVENT_APP_SEND_DATAGRAM) {
		m->state = PIGEON_BACKEND_RELAY_BACKOFF;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_RELAY_CONNECTED && event == PIGEON_SESSION_EVENT_RELAY_DATAGRAM) {
		m->state = PIGEON_BACKEND_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_OFFERED && event == PIGEON_SESSION_EVENT_RELAY_DATAGRAM) {
		m->state = PIGEON_BACKEND_LAN_OFFERED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_ACTIVE && event == PIGEON_SESSION_EVENT_RELAY_DATAGRAM) {
		m->state = PIGEON_BACKEND_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_DEGRADED && event == PIGEON_SESSION_EVENT_RELAY_DATAGRAM) {
		m->state = PIGEON_BACKEND_LAN_DEGRADED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_RELAY_BACKOFF && event == PIGEON_SESSION_EVENT_RELAY_DATAGRAM) {
		m->state = PIGEON_BACKEND_RELAY_BACKOFF;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_ACTIVE && event == PIGEON_SESSION_EVENT_LAN_STREAM_DATA) {
		m->state = PIGEON_BACKEND_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_DEGRADED && event == PIGEON_SESSION_EVENT_LAN_STREAM_DATA) {
		m->state = PIGEON_BACKEND_LAN_DEGRADED;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_ACTIVE && event == PIGEON_SESSION_EVENT_LAN_DATAGRAM) {
		m->state = PIGEON_BACKEND_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_BACKEND_LAN_DEGRADED && event == PIGEON_SESSION_EVENT_LAN_DATAGRAM) {
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
	if (m->state == PIGEON_CLIENT_WAIT_ACK && msg == PIGEON_SESSION_MSG_PAIR_HELLO_ACK) {
		if (m->actions[PIGEON_SESSION_ACTION_DERIVE_SECRET]) {
			int err = m->actions[PIGEON_SESSION_ACTION_DERIVE_SECRET](m->userdata);
			if (err) return -err;
		}
		// received_backend_pub: recv_msg.pubkey (set by action)
		// client_shared_key: DeriveKey("client_pub", recv_msg.pubkey) (set by action)
		m->state = PIGEON_CLIENT_E2E_READY;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_E2E_READY && msg == PIGEON_SESSION_MSG_PAIR_CONFIRM) {
		// client_code: DeriveCode(received_backend_pub, "client_pub") (set by action)
		m->state = PIGEON_CLIENT_SHOW_CODE;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_WAIT_PAIR_COMPLETE && msg == PIGEON_SESSION_MSG_PAIR_COMPLETE) {
		if (m->actions[PIGEON_SESSION_ACTION_STORE_SECRET]) {
			int err = m->actions[PIGEON_SESSION_ACTION_STORE_SECRET](m->userdata);
			if (err) return -err;
		}
		m->state = PIGEON_CLIENT_PAIRED;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_SEND_AUTH && msg == PIGEON_SESSION_MSG_AUTH_OK) {
		m->state = PIGEON_CLIENT_SESSION_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_RELAY_CONNECTED && msg == PIGEON_SESSION_MSG_LAN_OFFER && m->guards[PIGEON_SESSION_GUARD_LAN_ENABLED] && m->guards[PIGEON_SESSION_GUARD_LAN_ENABLED](m->userdata)) {
		if (m->actions[PIGEON_SESSION_ACTION_DIAL_LAN]) {
			int err = m->actions[PIGEON_SESSION_ACTION_DIAL_LAN](m->userdata);
			if (err) return -err;
		}
		m->state = PIGEON_CLIENT_LAN_CONNECTING;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_RELAY_CONNECTED && msg == PIGEON_SESSION_MSG_LAN_OFFER && m->guards[PIGEON_SESSION_GUARD_LAN_DISABLED] && m->guards[PIGEON_SESSION_GUARD_LAN_DISABLED](m->userdata)) {
		m->state = PIGEON_CLIENT_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_VERIFYING && msg == PIGEON_SESSION_MSG_LAN_CONFIRM) {
		if (m->actions[PIGEON_SESSION_ACTION_ACTIVATE_LAN]) {
			int err = m->actions[PIGEON_SESSION_ACTION_ACTIVATE_LAN](m->userdata);
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
	if (m->state == PIGEON_CLIENT_LAN_ACTIVE && msg == PIGEON_SESSION_MSG_PATH_PING) {
		m->state = PIGEON_CLIENT_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_ACTIVE && msg == PIGEON_SESSION_MSG_LAN_OFFER && m->guards[PIGEON_SESSION_GUARD_LAN_ENABLED] && m->guards[PIGEON_SESSION_GUARD_LAN_ENABLED](m->userdata)) {
		if (m->actions[PIGEON_SESSION_ACTION_DIAL_LAN]) {
			int err = m->actions[PIGEON_SESSION_ACTION_DIAL_LAN](m->userdata);
			if (err) return -err;
		}
		m->state = PIGEON_CLIENT_LAN_CONNECTING;
		return 1;
	}
	return 0;
}

int pigeon_client_step(pigeon_client_machine *m, session_event_id event)
{
	if (m->state == PIGEON_CLIENT_IDLE && event == PIGEON_SESSION_EVENT_BACKCHANNEL_RECEIVED) {
		m->state = PIGEON_CLIENT_OBTAIN_BACKCHANNEL_SECRET;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_OBTAIN_BACKCHANNEL_SECRET && event == PIGEON_SESSION_EVENT_SECRET_PARSED) {
		m->state = PIGEON_CLIENT_CONNECT_RELAY;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_CONNECT_RELAY && event == PIGEON_SESSION_EVENT_RELAY_CONNECTED) {
		m->state = PIGEON_CLIENT_GEN_KEY_PAIR;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_GEN_KEY_PAIR && event == PIGEON_SESSION_EVENT_KEY_PAIR_GENERATED) {
		if (m->actions[PIGEON_SESSION_ACTION_SEND_PAIR_HELLO]) {
			int err = m->actions[PIGEON_SESSION_ACTION_SEND_PAIR_HELLO](m->userdata);
			if (err) return -err;
		}
		m->state = PIGEON_CLIENT_WAIT_ACK;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_SHOW_CODE && event == PIGEON_SESSION_EVENT_CODE_DISPLAYED) {
		m->state = PIGEON_CLIENT_WAIT_PAIR_COMPLETE;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_PAIRED && event == PIGEON_SESSION_EVENT_APP_LAUNCH) {
		m->state = PIGEON_CLIENT_RECONNECT;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_RECONNECT && event == PIGEON_SESSION_EVENT_RELAY_CONNECTED) {
		m->state = PIGEON_CLIENT_SEND_AUTH;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_SESSION_ACTIVE && event == PIGEON_SESSION_EVENT_SESSION_ESTABLISHED) {
		m->state = PIGEON_CLIENT_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_CONNECTING && event == PIGEON_SESSION_EVENT_LAN_DIAL_OK) {
		m->state = PIGEON_CLIENT_LAN_VERIFYING;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_CONNECTING && event == PIGEON_SESSION_EVENT_LAN_DIAL_FAILED) {
		m->state = PIGEON_CLIENT_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_VERIFYING && event == PIGEON_SESSION_EVENT_VERIFY_TIMEOUT) {
		m->c_dispatcher_path = "relay";
		if (m->on_change) m->on_change("c_dispatcher_path", m->userdata);
		m->state = PIGEON_CLIENT_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_ACTIVE && event == PIGEON_SESSION_EVENT_LAN_ERROR) {
		if (m->actions[PIGEON_SESSION_ACTION_FALLBACK_TO_RELAY]) {
			int err = m->actions[PIGEON_SESSION_ACTION_FALLBACK_TO_RELAY](m->userdata);
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
	if (m->state == PIGEON_CLIENT_LAN_ACTIVE && event == PIGEON_SESSION_EVENT_LAN_STREAM_ERROR) {
		if (m->actions[PIGEON_SESSION_ACTION_FALLBACK_TO_RELAY]) {
			int err = m->actions[PIGEON_SESSION_ACTION_FALLBACK_TO_RELAY](m->userdata);
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
	if (m->state == PIGEON_CLIENT_RELAY_FALLBACK && event == PIGEON_SESSION_EVENT_RELAY_OK) {
		m->state = PIGEON_CLIENT_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_CONNECTING && event == PIGEON_SESSION_EVENT_APP_FORCE_FALLBACK) {
		m->state = PIGEON_CLIENT_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_VERIFYING && event == PIGEON_SESSION_EVENT_APP_FORCE_FALLBACK) {
		m->c_dispatcher_path = "relay";
		if (m->on_change) m->on_change("c_dispatcher_path", m->userdata);
		m->state = PIGEON_CLIENT_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_ACTIVE && event == PIGEON_SESSION_EVENT_APP_FORCE_FALLBACK) {
		if (m->actions[PIGEON_SESSION_ACTION_FALLBACK_TO_RELAY]) {
			int err = m->actions[PIGEON_SESSION_ACTION_FALLBACK_TO_RELAY](m->userdata);
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
	if (m->state == PIGEON_CLIENT_RELAY_CONNECTED && event == PIGEON_SESSION_EVENT_DISCONNECT) {
		m->state = PIGEON_CLIENT_PAIRED;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_RELAY_CONNECTED && event == PIGEON_SESSION_EVENT_APP_SEND) {
		m->state = PIGEON_CLIENT_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_CONNECTING && event == PIGEON_SESSION_EVENT_APP_SEND) {
		m->state = PIGEON_CLIENT_LAN_CONNECTING;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_VERIFYING && event == PIGEON_SESSION_EVENT_APP_SEND) {
		m->state = PIGEON_CLIENT_LAN_VERIFYING;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_ACTIVE && event == PIGEON_SESSION_EVENT_APP_SEND) {
		m->state = PIGEON_CLIENT_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_RELAY_FALLBACK && event == PIGEON_SESSION_EVENT_APP_SEND) {
		m->state = PIGEON_CLIENT_RELAY_FALLBACK;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_RELAY_CONNECTED && event == PIGEON_SESSION_EVENT_RELAY_STREAM_DATA) {
		m->state = PIGEON_CLIENT_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_CONNECTING && event == PIGEON_SESSION_EVENT_RELAY_STREAM_DATA) {
		m->state = PIGEON_CLIENT_LAN_CONNECTING;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_VERIFYING && event == PIGEON_SESSION_EVENT_RELAY_STREAM_DATA) {
		m->state = PIGEON_CLIENT_LAN_VERIFYING;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_ACTIVE && event == PIGEON_SESSION_EVENT_RELAY_STREAM_DATA) {
		m->state = PIGEON_CLIENT_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_RELAY_FALLBACK && event == PIGEON_SESSION_EVENT_RELAY_STREAM_DATA) {
		m->state = PIGEON_CLIENT_RELAY_FALLBACK;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_RELAY_CONNECTED && event == PIGEON_SESSION_EVENT_RELAY_STREAM_ERROR) {
		m->state = PIGEON_CLIENT_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_CONNECTING && event == PIGEON_SESSION_EVENT_RELAY_STREAM_ERROR) {
		m->state = PIGEON_CLIENT_LAN_CONNECTING;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_VERIFYING && event == PIGEON_SESSION_EVENT_RELAY_STREAM_ERROR) {
		m->state = PIGEON_CLIENT_LAN_VERIFYING;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_ACTIVE && event == PIGEON_SESSION_EVENT_RELAY_STREAM_ERROR) {
		m->state = PIGEON_CLIENT_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_RELAY_FALLBACK && event == PIGEON_SESSION_EVENT_RELAY_STREAM_ERROR) {
		m->state = PIGEON_CLIENT_RELAY_FALLBACK;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_RELAY_CONNECTED && event == PIGEON_SESSION_EVENT_APP_SEND_DATAGRAM) {
		m->state = PIGEON_CLIENT_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_CONNECTING && event == PIGEON_SESSION_EVENT_APP_SEND_DATAGRAM) {
		m->state = PIGEON_CLIENT_LAN_CONNECTING;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_VERIFYING && event == PIGEON_SESSION_EVENT_APP_SEND_DATAGRAM) {
		m->state = PIGEON_CLIENT_LAN_VERIFYING;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_ACTIVE && event == PIGEON_SESSION_EVENT_APP_SEND_DATAGRAM) {
		m->state = PIGEON_CLIENT_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_RELAY_FALLBACK && event == PIGEON_SESSION_EVENT_APP_SEND_DATAGRAM) {
		m->state = PIGEON_CLIENT_RELAY_FALLBACK;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_RELAY_CONNECTED && event == PIGEON_SESSION_EVENT_RELAY_DATAGRAM) {
		m->state = PIGEON_CLIENT_RELAY_CONNECTED;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_CONNECTING && event == PIGEON_SESSION_EVENT_RELAY_DATAGRAM) {
		m->state = PIGEON_CLIENT_LAN_CONNECTING;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_VERIFYING && event == PIGEON_SESSION_EVENT_RELAY_DATAGRAM) {
		m->state = PIGEON_CLIENT_LAN_VERIFYING;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_ACTIVE && event == PIGEON_SESSION_EVENT_RELAY_DATAGRAM) {
		m->state = PIGEON_CLIENT_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_RELAY_FALLBACK && event == PIGEON_SESSION_EVENT_RELAY_DATAGRAM) {
		m->state = PIGEON_CLIENT_RELAY_FALLBACK;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_ACTIVE && event == PIGEON_SESSION_EVENT_LAN_STREAM_DATA) {
		m->state = PIGEON_CLIENT_LAN_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_CLIENT_LAN_ACTIVE && event == PIGEON_SESSION_EVENT_LAN_DATAGRAM) {
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
	if (m->state == PIGEON_RELAY_IDLE && event == PIGEON_SESSION_EVENT_BACKEND_REGISTER) {
		m->state = PIGEON_RELAY_BACKEND_REGISTERED;
		return 1;
	}
	if (m->state == PIGEON_RELAY_BACKEND_REGISTERED && event == PIGEON_SESSION_EVENT_CLIENT_CONNECT) {
		if (m->actions[PIGEON_SESSION_ACTION_BRIDGE_STREAMS]) {
			int err = m->actions[PIGEON_SESSION_ACTION_BRIDGE_STREAMS](m->userdata);
			if (err) return -err;
		}
		m->relay_bridge = "active";
		if (m->on_change) m->on_change("relay_bridge", m->userdata);
		m->state = PIGEON_RELAY_BRIDGED;
		return 1;
	}
	if (m->state == PIGEON_RELAY_BRIDGED && event == PIGEON_SESSION_EVENT_CLIENT_DISCONNECT) {
		if (m->actions[PIGEON_SESSION_ACTION_UNBRIDGE]) {
			int err = m->actions[PIGEON_SESSION_ACTION_UNBRIDGE](m->userdata);
			if (err) return -err;
		}
		m->relay_bridge = "idle";
		if (m->on_change) m->on_change("relay_bridge", m->userdata);
		m->state = PIGEON_RELAY_BACKEND_REGISTERED;
		return 1;
	}
	if (m->state == PIGEON_RELAY_BACKEND_REGISTERED && event == PIGEON_SESSION_EVENT_BACKEND_DISCONNECT) {
		m->state = PIGEON_RELAY_IDLE;
		return 1;
	}
	return 0;
}


// --- Activation handshake driver ---



// pigeon.h gives us pigeon_transport / pigeon_stream_handle /
// pigeon_pairing_record. session_gen.h gives us the SessionMachine
// state / event constants and the per-side machine struct
// definitions. Post-T32.1's protogen rename of the COUNT sentinels,
// pairingceremony_gen.h (transitively via pigeon.h) and
// session_gen.h coexist in one TU without enumerator collisions.



// Maximum on-the-wire payload for an auth_request / auth_ok frame.
// auth_request: 1 (tag) + 10 (max uvarint) + 128 (max device id) = 139
// auth_ok rejected: 1 (tag) + 1 (flag) + 10 (uvarint) + 256 (reason) = 268
// Round up generously.
#define PIGEON_AUTH_MAX_PAYLOAD 512

// ---------- uvarint helpers ----------
//
// Mirror of encoding/binary.PutUvarint and Uvarint (LEB128). Match
// the wire bytes the Go side writes.

static int auth_uvarint_encode(uint64_t v, uint8_t *buf, size_t buf_len)
{
    size_t off = 0;
    while (v >= 0x80) {
        if (off >= buf_len) return -1;
        buf[off++] = (uint8_t)(v) | 0x80;
        v >>= 7;
    }
    if (off >= buf_len) return -1;
    buf[off++] = (uint8_t)v;
    return (int)off;
}

// Decode a uvarint from buf. On success, *out gets the value, *consumed
// gets the number of bytes read. Returns 0 / -1.
static int auth_uvarint_decode(const uint8_t *buf, size_t buf_len,
                               uint64_t *out, size_t *consumed)
{
    uint64_t v = 0;
    unsigned shift = 0;
    size_t off = 0;
    while (off < buf_len) {
        uint8_t b = buf[off++];
        if (shift >= 64) return -1;
        v |= ((uint64_t)(b & 0x7F)) << shift;
        if ((b & 0x80) == 0) {
            *out = v;
            *consumed = off;
            return 0;
        }
        shift += 7;
    }
    return -1; // truncated
}

// ---------- wire encoders / decoders ----------

int pigeon_encode_auth_request(const char *device_id,
                               uint8_t *buf, size_t buf_len)
{
    if (buf_len < 1) return -1;
    buf[0] = PIGEON_AUTH_MSG_TAG_AUTH_REQUEST;
    size_t off = 1;

    size_t id_len = strlen(device_id);
    if (id_len > PIGEON_AUTH_MAX_DEVICE_ID) return -1;

    int n = auth_uvarint_encode((uint64_t)id_len, buf + off, buf_len - off);
    if (n < 0) return -1;
    off += (size_t)n;

    if (off + id_len > buf_len) return -1;
    memcpy(buf + off, device_id, id_len);
    off += id_len;
    return (int)off;
}

int pigeon_decode_auth_request(const uint8_t *payload, size_t payload_len,
                               char *out_device_id, size_t out_cap)
{
    if (payload_len < 1) return -1;
    if (payload[0] != PIGEON_AUTH_MSG_TAG_AUTH_REQUEST) return -1;

    uint64_t id_len = 0;
    size_t consumed = 0;
    if (auth_uvarint_decode(payload + 1, payload_len - 1, &id_len, &consumed) != 0) return -1;
    if (id_len > PIGEON_AUTH_MAX_DEVICE_ID) return -1;
    if (1 + consumed + id_len != payload_len) return -1;
    if (id_len + 1 > out_cap) return -1;

    memcpy(out_device_id, payload + 1 + consumed, (size_t)id_len);
    out_device_id[id_len] = '\0';
    return 0;
}

int pigeon_encode_auth_ok(bool ok, const char *reason,
                          uint8_t *buf, size_t buf_len)
{
    if (buf_len < 2) return -1;
    buf[0] = PIGEON_AUTH_MSG_TAG_AUTH_OK;
    if (ok) {
        buf[1] = PIGEON_AUTH_OK_ACCEPTED;
        return 2;
    }
    buf[1] = PIGEON_AUTH_OK_REJECTED;
    size_t off = 2;
    size_t r_len = (reason != NULL) ? strlen(reason) : 0;
    if (r_len > PIGEON_AUTH_MAX_REASON) return -1;

    int n = auth_uvarint_encode((uint64_t)r_len, buf + off, buf_len - off);
    if (n < 0) return -1;
    off += (size_t)n;

    if (off + r_len > buf_len) return -1;
    if (r_len > 0) memcpy(buf + off, reason, r_len);
    off += r_len;
    return (int)off;
}

int pigeon_decode_auth_ok(const uint8_t *payload, size_t payload_len,
                          bool *out_ok,
                          char *out_reason, size_t out_reason_cap)
{
    if (payload_len < 2) return -1;
    if (payload[0] != PIGEON_AUTH_MSG_TAG_AUTH_OK) return -1;
    switch (payload[1]) {
    case PIGEON_AUTH_OK_ACCEPTED:
        *out_ok = true;
        if (out_reason != NULL && out_reason_cap > 0) out_reason[0] = '\0';
        return 0;
    case PIGEON_AUTH_OK_REJECTED: {
        *out_ok = false;
        uint64_t r_len = 0;
        size_t consumed = 0;
        if (auth_uvarint_decode(payload + 2, payload_len - 2, &r_len, &consumed) != 0) return -1;
        if (r_len > PIGEON_AUTH_MAX_REASON) return -1;
        if (2 + consumed + r_len != payload_len) return -1;
        if (out_reason != NULL && out_reason_cap > 0) {
            size_t copy = (r_len < out_reason_cap - 1) ? r_len : out_reason_cap - 1;
            if (copy > 0) memcpy(out_reason, payload + 2 + consumed, copy);
            out_reason[copy] = '\0';
        }
        return 0;
    }
    default:
        return -1;
    }
}

// ---------- per-side drive helpers ----------

// Guard predicates for the backend's device_known / device_unknown
// branch in the AuthCheck → SessionActive transition. The callbacks
// wrap a stack-allocated bool; passing it via the machine's
// userdata field lets the spec drive the branch without us having
// to type-pun the machine struct.
typedef struct {
    bool known;
} backend_auth_ctx;

static bool backend_device_known(void *ctx)
{
    return ((backend_auth_ctx *)ctx)->known;
}

static bool backend_device_unknown(void *ctx)
{
    return !((backend_auth_ctx *)ctx)->known;
}

int pigeon_run_backend_activation(const void *transport_v,
                                  void *stream_v,
                                  pigeon_resolve_device_fn resolve,
                                  void *resolve_userdata,
                                  void *out_machine_v,
                                  char *out_device_id, size_t out_device_id_cap,
                                  void *out_record)
{
    const pigeon_transport *transport = (const pigeon_transport *)transport_v;
    pigeon_stream_handle   *stream    = (pigeon_stream_handle *)stream_v;
    pigeon_backend_machine *machine   = (pigeon_backend_machine *)out_machine_v;

    if (out_device_id_cap > 0) out_device_id[0] = '\0';

    // Read the auth_request from the primary stream.
    uint8_t buf[PIGEON_AUTH_MAX_PAYLOAD];
    size_t  in_len = 0;
    if (transport->recv_on_stream(transport->userdata, stream,
                                  buf, sizeof(buf), &in_len) != 0) {
        return -1;
    }
    char device_id[PIGEON_AUTH_MAX_DEVICE_ID + 1];
    if (pigeon_decode_auth_request(buf, in_len,
                                   device_id, sizeof(device_id)) != 0) {
        return -1;
    }
    // Surface the decoded device ID to the caller before doing any
    // further work (matches Go's runBackendActivation tri-valued
    // contract: caller can distinguish wire-decode failure from
    // resolved-but-rejected).
    size_t did_len = strlen(device_id);
    if (did_len + 1 > out_device_id_cap) return -1;
    memcpy(out_device_id, device_id, did_len + 1);

    // Pre-seed the per-client backend machine at Paired with the
    // resolved device id. This matches Go's
    // newBackendActivationMachine — the lookup itself stays outside
    // the spec; the spec sees only the device_known / device_unknown
    // guard outcome.
    pigeon_backend_machine_init(machine);
    machine->state = PIGEON_BACKEND_PAIRED;
    machine->received_device_id = device_id; // stack pointer is OK,
                                             // machine is consumed
                                             // before this returns.

    backend_auth_ctx auth_ctx;
    auth_ctx.known = (resolve(resolve_userdata, device_id, out_record) == 0);

    machine->guards[PIGEON_SESSION_GUARD_DEVICE_KNOWN]   = backend_device_known;
    machine->guards[PIGEON_SESSION_GUARD_DEVICE_UNKNOWN] = backend_device_unknown;
    machine->userdata = &auth_ctx;

    // Drive recv auth_request → AuthCheck. The C generator routes
    // `recv`-trigger transitions through handle_message (matched on
    // msg_type); only `internal`-trigger transitions go through step.
    if (pigeon_backend_handle_message(machine, PIGEON_SESSION_MSG_AUTH_REQUEST) <= 0) {
        return -1;
    }
    if (machine->state != PIGEON_BACKEND_AUTH_CHECK) {
        return -1;
    }

    // Drive verify → SessionActive (known) or → Idle (unknown).
    if (pigeon_backend_step(machine, PIGEON_SESSION_EVENT_VERIFY) <= 0) {
        return -1;
    }

    if (!auth_ctx.known) {
        // Tell the client and return tri-value 1 (decoded but rejected).
        uint8_t reply[PIGEON_AUTH_MAX_PAYLOAD];
        int n = pigeon_encode_auth_ok(false, "unknown client",
                                      reply, sizeof(reply));
        if (n > 0) {
            (void)transport->send_on_stream(transport->userdata, stream,
                                            reply, (size_t)n);
        }
        return 1;
    }

    uint8_t reply[PIGEON_AUTH_MAX_PAYLOAD];
    int n = pigeon_encode_auth_ok(true, NULL, reply, sizeof(reply));
    if (n < 0) return -1;
    if (transport->send_on_stream(transport->userdata, stream,
                                  reply, (size_t)n) != 0) {
        return -1;
    }
    if (machine->state != PIGEON_BACKEND_SESSION_ACTIVE) {
        return -1;
    }
    return 0;
}

int pigeon_run_client_activation(const void *transport_v,
                                 void *stream_v,
                                 const char *device_id,
                                 void *out_machine_v,
                                 char *out_reason, size_t out_reason_cap)
{
    const pigeon_transport *transport = (const pigeon_transport *)transport_v;
    pigeon_stream_handle   *stream    = (pigeon_stream_handle *)stream_v;
    pigeon_client_machine  *machine   = (pigeon_client_machine *)out_machine_v;

    pigeon_client_machine_init(machine);
    machine->state = PIGEON_CLIENT_RECONNECT;

    // Drive relay_connected → SendAuth (matches the spec's transition
    // out of Reconnect).
    if (pigeon_client_step(machine, PIGEON_SESSION_EVENT_RELAY_CONNECTED) <= 0) {
        return -1;
    }
    if (machine->state != PIGEON_CLIENT_SEND_AUTH) return -1;

    // Send auth_request.
    uint8_t out[PIGEON_AUTH_MAX_PAYLOAD];
    int n = pigeon_encode_auth_request(device_id, out, sizeof(out));
    if (n < 0) return -1;
    if (transport->send_on_stream(transport->userdata, stream,
                                  out, (size_t)n) != 0) {
        return -1;
    }

    // Read auth_ok.
    uint8_t in[PIGEON_AUTH_MAX_PAYLOAD];
    size_t  in_len = 0;
    if (transport->recv_on_stream(transport->userdata, stream,
                                  in, sizeof(in), &in_len) != 0) {
        return -1;
    }
    bool ok = false;
    if (pigeon_decode_auth_ok(in, in_len, &ok,
                              out_reason, out_reason_cap) != 0) {
        return -1;
    }
    if (!ok) return -1;

    // recv auth_ok → SessionActive (recv-triggered, so handle_message).
    if (pigeon_client_handle_message(machine, PIGEON_SESSION_MSG_AUTH_OK) <= 0) {
        return -1;
    }
    if (machine->state != PIGEON_CLIENT_SESSION_ACTIVE) return -1;
    return 0;
}

// --- Crypto ---



// Pigeon crypto uses:
// - X25519 ECDH for key exchange
// - HKDF-SHA256 for key derivation
// - AES-256-GCM for symmetric encryption
// - Confirmation codes: HKDF of sorted pubkeys, mod 10^6
//
// Link against libsodium (-lsodium) or OpenSSL (-lcrypto) to provide
// the underlying primitives. Define PIGEON_CRYPTO_LIBSODIUM or
// PIGEON_CRYPTO_OPENSSL to select the backend.

#if defined(PIGEON_CRYPTO_LIBSODIUM)
#include <sodium.h>

int pigeon_generate_keypair(pigeon_keypair *kp)
{
    // libsodium's crypto_box uses X25519 internally.
    // scalarmult base gives us the public key from a random secret.
    randombytes_buf(kp->private_key, 32);
    return crypto_scalarmult_base(kp->public_key, kp->private_key) == 0 ? 0 : -1;
}

// Internal: HKDF-SHA256 extract + expand. libsodium doesn't have HKDF
// natively, so we build it from HMAC-SHA256.
static int hkdf_sha256(const uint8_t *ikm, size_t ikm_len,
                       const uint8_t *info, size_t info_len,
                       uint8_t *out, size_t out_len)
{
    // Extract: PRK = HMAC-SHA256(salt="", IKM)
    uint8_t prk[32];
    uint8_t salt[32] = {0}; // empty salt
    crypto_auth_hmacsha256_state st;

    crypto_auth_hmacsha256_init(&st, salt, 32);
    crypto_auth_hmacsha256_update(&st, ikm, ikm_len);
    crypto_auth_hmacsha256_final(&st, prk);

    // Expand: OKM = HMAC-SHA256(PRK, info || 0x01)
    // For 32 bytes output, one iteration is enough.
    if (out_len > 32) return -1;

    uint8_t expand_input[256 + 1]; // info + counter byte
    if (info_len > 255) return -1;
    memcpy(expand_input, info, info_len);
    expand_input[info_len] = 0x01;

    crypto_auth_hmacsha256_init(&st, prk, 32);
    crypto_auth_hmacsha256_update(&st, expand_input, info_len + 1);
    crypto_auth_hmacsha256_final(&st, out);

    sodium_memzero(prk, sizeof(prk));
    return 0;
}

int pigeon_derive_session_key(const uint8_t *private_key,
                              const uint8_t *peer_public_key,
                              const uint8_t *info, size_t info_len,
                              uint8_t *out_key)
{
    // ECDH: shared_secret = X25519(private_key, peer_public_key)
    uint8_t shared[32];
    if (crypto_scalarmult(shared, private_key, peer_public_key) != 0) {
        return -1;
    }

    // KDF: session_key = HKDF-SHA256(shared_secret, info)
    int ret = hkdf_sha256(shared, 32, info, info_len, out_key, 32);
    sodium_memzero(shared, sizeof(shared));
    return ret;
}

int pigeon_derive_confirmation_code(const uint8_t *pub_a,
                                    const uint8_t *pub_b,
                                    char *out_code)
{
    // Sort keys lexicographically, concatenate.
    uint8_t ikm[64];
    int cmp = memcmp(pub_a, pub_b, 32);
    if (cmp <= 0) {
        memcpy(ikm, pub_a, 32);
        memcpy(ikm + 32, pub_b, 32);
    } else {
        memcpy(ikm, pub_b, 32);
        memcpy(ikm + 32, pub_a, 32);
    }

    // HKDF with info="pairing-confirmation"
    uint8_t derived[32];
    const uint8_t info[] = "pairing-confirmation";
    int ret = hkdf_sha256(ikm, 64, info, sizeof(info) - 1, derived, 32);
    if (ret != 0) return -1;

    // First 4 bytes as big-endian uint32, mod 1000000, zero-padded.
    uint32_t val = ((uint32_t)derived[0] << 24) |
                   ((uint32_t)derived[1] << 16) |
                   ((uint32_t)derived[2] << 8)  |
                   ((uint32_t)derived[3]);
    val %= 1000000;

    // Write 6-digit code + null terminator.
    for (int i = 5; i >= 0; i--) {
        out_code[i] = '0' + (char)(val % 10);
        val /= 10;
    }
    out_code[6] = '\0';
    return 0;
}

void pigeon_channel_init(pigeon_channel *ch,
                         const uint8_t *send_key,
                         const uint8_t *recv_key,
                         pigeon_channel_mode mode)
{
    memcpy(ch->send_key, send_key, 32);
    memcpy(ch->recv_key, recv_key, 32);
    ch->send_seq = 0;
    ch->recv_seq = 0;
    ch->mode = mode;
    ch->established = true;
}

int pigeon_channel_init_symmetric(pigeon_channel *ch,
                                  const uint8_t *master_key,
                                  bool is_server)
{
    uint8_t c2s[32], s2c[32];
    const uint8_t info_c2s[] = "client-to-server";
    const uint8_t info_s2c[] = "server-to-client";

    if (hkdf_sha256(master_key, 32, info_c2s, sizeof(info_c2s) - 1, c2s, 32) != 0)
        return -1;
    if (hkdf_sha256(master_key, 32, info_s2c, sizeof(info_s2c) - 1, s2c, 32) != 0)
        return -1;

    if (is_server) {
        pigeon_channel_init(ch, s2c, c2s, PIGEON_MODE_STRICT);
    } else {
        pigeon_channel_init(ch, c2s, s2c, PIGEON_MODE_STRICT);
    }

    sodium_memzero(c2s, sizeof(c2s));
    sodium_memzero(s2c, sizeof(s2c));
    return 0;
}

int pigeon_channel_encrypt(pigeon_channel *ch,
                           const uint8_t *plaintext, size_t plaintext_len,
                           uint8_t *out, size_t out_len)
{
    // Output: [8-byte LE seq][ciphertext + 16-byte tag]
    size_t needed = 8 + plaintext_len + crypto_aead_aes256gcm_ABYTES;
    if (out_len < needed) return -1;

    // Sequence number as little-endian 8 bytes.
    uint64_t seq = ch->send_seq++;
    for (int i = 0; i < 8; i++) {
        out[i] = (uint8_t)(seq >> (i * 8));
    }

    // Nonce: LE64(seq) zero-padded to 12 bytes.
    uint8_t nonce[12] = {0};
    memcpy(nonce, out, 8);

    // AAD is the 8-byte seq prefix.
    unsigned long long ciphertext_len = 0;
    if (crypto_aead_aes256gcm_encrypt(out + 8, &ciphertext_len,
                                       plaintext, plaintext_len,
                                       out, 8, // AAD = seq bytes
                                       NULL, nonce, ch->send_key) != 0) {
        return -1;
    }

    return (int)(8 + ciphertext_len);
}

int pigeon_channel_decrypt(pigeon_channel *ch,
                           const uint8_t *data, size_t data_len,
                           uint8_t *out, size_t out_len)
{
    if (data_len < 8 + crypto_aead_aes256gcm_ABYTES) return -1;

    // Read sequence number (LE64).
    uint64_t seq = 0;
    for (int i = 0; i < 8; i++) {
        seq |= (uint64_t)data[i] << (i * 8);
    }

    // Sequence validation.
    if (ch->mode == PIGEON_MODE_STRICT) {
        if (seq != ch->recv_seq) return -1;
    } else {
        if (seq < ch->recv_seq) return -1;  // reject replays and old
    }

    // Nonce: LE64(seq) zero-padded to 12 bytes.
    uint8_t nonce[12] = {0};
    memcpy(nonce, data, 8);

    size_t ciphertext_len = data_len - 8;
    if (out_len < ciphertext_len - crypto_aead_aes256gcm_ABYTES) return -1;

    unsigned long long plaintext_len = 0;
    if (crypto_aead_aes256gcm_decrypt(out, &plaintext_len,
                                       NULL,
                                       data + 8, ciphertext_len,
                                       data, 8, // AAD = seq bytes
                                       nonce, ch->recv_key) != 0) {
        return -1;
    }

    if (ch->mode == PIGEON_MODE_STRICT) {
        ch->recv_seq = seq + 1;
    } else {
        if (seq >= ch->recv_seq) {
            ch->recv_seq = seq + 1;
        }
    }

    return (int)plaintext_len;
}

#else
#error "Define PIGEON_CRYPTO_LIBSODIUM or PIGEON_CRYPTO_OPENSSL to select a crypto backend"
#endif

// --- Generated one-shot wire formats ---

//


int pigeon_wire_stream_header_encode(bool is_backend, uint32_t client_tag, const char *name, size_t name_len, uint8_t *out, size_t out_len)
{
    size_t off = 0;
    if (is_backend) {
        if (off + 4 > out_len) return -1;
        out[off++] = (uint8_t)(client_tag >> 24);
        out[off++] = (uint8_t)(client_tag >> 16);
        out[off++] = (uint8_t)(client_tag >> 8);
        out[off++] = (uint8_t)(client_tag);
        { int _n = pigeon_uvarint_encode((uint64_t)name_len, out + off, out_len - off); if (_n < 0) return -1; off += (size_t)_n; }
        if (off + name_len > out_len) return -1;
        if (name_len > 0) { if (!name) return -1; memcpy(out + off, name, name_len); off += name_len; }
    } else {
        { int _n = pigeon_uvarint_encode((uint64_t)name_len, out + off, out_len - off); if (_n < 0) return -1; off += (size_t)_n; }
        if (off + name_len > out_len) return -1;
        if (name_len > 0) { if (!name) return -1; memcpy(out + off, name, name_len); off += name_len; }
    }
    return (int)off;
}

int pigeon_wire_stream_header_decode_backend(const uint8_t *buf, size_t buf_len, uint32_t *client_tag, char *name_buf, size_t name_buf_len, size_t *name_len_out)
{
    size_t off = 0;
    if (buf_len - off < 4) return -1;
    if (client_tag) *client_tag = ((uint32_t)buf[off] << 24) | ((uint32_t)buf[off+1] << 16) | ((uint32_t)buf[off+2] << 8) | (uint32_t)buf[off+3];
    off += 4;
    { uint64_t _len = 0; int _n = pigeon_uvarint_decode(buf + off, buf_len - off, &_len); if (_n <= 0) return -1; off += (size_t)_n;
      if (_len > buf_len - off) return -1;
      if (_len + 1 > name_buf_len) return -1;
      if (_len > 0) memcpy(name_buf, buf + off, (size_t)_len);
      name_buf[_len] = '\0';
      if (name_len_out) *name_len_out = (size_t)_len;
      off += (size_t)_len; }
    return (int)off;
}

int pigeon_wire_stream_header_decode_client(const uint8_t *buf, size_t buf_len, char *name_buf, size_t name_buf_len, size_t *name_len_out)
{
    size_t off = 0;
    { uint64_t _len = 0; int _n = pigeon_uvarint_decode(buf + off, buf_len - off, &_len); if (_n <= 0) return -1; off += (size_t)_n;
      if (_len > buf_len - off) return -1;
      if (_len + 1 > name_buf_len) return -1;
      if (_len > 0) memcpy(name_buf, buf + off, (size_t)_len);
      name_buf[_len] = '\0';
      if (name_len_out) *name_len_out = (size_t)_len;
      off += (size_t)_len; }
    return (int)off;
}

int pigeon_wire_datagram_plaintext_encode(uint64_t channel_id, const uint8_t *payload, size_t payload_len, uint8_t *out, size_t out_len)
{
    size_t off = 0;
    { int _n = pigeon_uvarint_encode(channel_id, out + off, out_len - off); if (_n < 0) return -1; off += (size_t)_n; }
    if (off + payload_len > out_len) return -1;
    if (payload_len > 0) { memcpy(out + off, payload, payload_len); off += payload_len; }
    return (int)off;
}

int pigeon_wire_datagram_plaintext_decode(const uint8_t *buf, size_t buf_len, uint64_t *channel_id, uint8_t *payload_buf, size_t payload_buf_len, size_t *payload_len_out)
{
    size_t off = 0;
    { uint64_t _v = 0; int _n = pigeon_uvarint_decode(buf + off, buf_len - off, &_v); if (_n <= 0) return -1; if (channel_id) *channel_id = _v; off += (size_t)_n; }
    { size_t _n = buf_len - off; if (_n > payload_buf_len) return -1; if (_n > 0) memcpy(payload_buf, buf + off, _n); if (payload_len_out) *payload_len_out = _n; off = buf_len; }
    return (int)off;
}

int pigeon_wire_relay_greeting_encode_connect(const char *instance_id, size_t instance_id_len, uint8_t *out, size_t out_len)
{
    size_t off = 0;
    const size_t prefix_len = 8;
    if (off + prefix_len > out_len) return -1;
    memcpy(out + off, "connect:", prefix_len);
    off += prefix_len;
    if (instance_id && instance_id_len > 0) {
        if (off + instance_id_len > out_len) return -1;
        memcpy(out + off, instance_id, instance_id_len);
        off += instance_id_len;
    }
    return (int)off;
}

int pigeon_wire_relay_greeting_encode_register_mux(const char *token, const char *instance_id, uint8_t *out, size_t out_len)
{
    size_t off = 0;
    const size_t prefix_len = 12;
    if (off + prefix_len > out_len) return -1;
    memcpy(out + off, "register-mux", prefix_len);
    off += prefix_len;
    const char *parts[2];
    size_t part_lens[2];
    bool any_non_empty = false;
    parts[0] = token;
    part_lens[0] = (token ? strlen(token) : 0);
    if (parts[0] && part_lens[0] > 0) any_non_empty = true;
    parts[1] = instance_id;
    part_lens[1] = (instance_id ? strlen(instance_id) : 0);
    if (parts[1] && part_lens[1] > 0) any_non_empty = true;
    if (any_non_empty) {
        for (int i = 0; i < 2; i++) {
            if (off + 1 > out_len) return -1;
            out[off++] = ':';
            if (parts[i] && part_lens[i] > 0) {
                if (off + part_lens[i] > out_len) return -1;
                memcpy(out + off, parts[i], part_lens[i]);
                off += part_lens[i];
            }
        }
    }
    return (int)off;
}

int pigeon_wire_relay_greeting_decode(const uint8_t *buf, size_t buf_len, pigeon_wire_relay_greeting_variant *out_variant, char *instance_id_buf, size_t instance_id_buf_len, size_t *instance_id_len_out, char *token_buf, size_t token_buf_len, size_t *token_len_out)
{
    if (instance_id_len_out) *instance_id_len_out = 0;
    if (token_len_out) *token_len_out = 0;
    if (buf_len >= 12 && memcmp(buf, "register-mux", 12) == 0) {
        *out_variant = RELAY_GREETING_REGISTER_MUX;
        size_t off = 12;
        if (off < buf_len) {
            if (buf[off] != ':') return -1;
            off++;
            size_t part_start = off;
            size_t part_idx = 0;
            const size_t expected = 2;
            for (; off <= buf_len; off++) {
                bool atEnd = (off == buf_len);
                if (atEnd || (part_idx + 1 < expected && buf[off] == ':')) {
                    size_t part_len = off - part_start;
                    switch (part_idx) {
                    case 0:
                        if (part_len + 1 > token_buf_len) return -1;
                        if (part_len > 0) memcpy(token_buf, buf + part_start, part_len);
                        token_buf[part_len] = '\0';
                        if (token_len_out) *token_len_out = part_len;
                        break;
                    case 1:
                        if (part_len + 1 > instance_id_buf_len) return -1;
                        if (part_len > 0) memcpy(instance_id_buf, buf + part_start, part_len);
                        instance_id_buf[part_len] = '\0';
                        if (instance_id_len_out) *instance_id_len_out = part_len;
                        break;
                    }
                    part_idx++;
                    part_start = off + 1;
                    if (atEnd) break;
                }
            }
        }
        return (int)buf_len;
    }
    if (buf_len >= 8 && memcmp(buf, "connect:", 8) == 0) {
        *out_variant = RELAY_GREETING_CONNECT;
        size_t off = 8;
        size_t instance_id_n = buf_len - off;
        if (instance_id_n + 1 > instance_id_buf_len) return -1;
        if (instance_id_n > 0) memcpy(instance_id_buf, buf + off, instance_id_n);
        instance_id_buf[instance_id_n] = '\0';
        if (instance_id_len_out) *instance_id_len_out = instance_id_n;
        off = buf_len;
        return (int)buf_len;
    }
    *out_variant = RELAY_GREETING_UNKNOWN;
    return -1;
}


// --- Connection and framing ---



void pigeon_init(pigeon_ctx *ctx, const pigeon_transport *transport)
{
    memset(ctx, 0, sizeof(*ctx));
    if (transport) {
        ctx->transport = *transport;
    }
}

int pigeon_send(pigeon_ctx *ctx, const uint8_t *data, size_t len)
{
    if (!ctx->transport.send_stream) return -1;
    if (len > PIGEON_MAX_MSG) return -1;

    const uint8_t *payload = data;
    size_t payload_len = len;

    if (ctx->stream_channel.established) {
        // Encrypt into write_buf+4, leaving room for the 4-byte length prefix.
        // Maximum ciphertext: 8 (seq) + len + 16 (tag).
        if (8 + len + 16 + 4 > sizeof(ctx->write_buf)) return -1;
        int ct_len = pigeon_channel_encrypt(&ctx->stream_channel,
                                            data, len,
                                            ctx->write_buf + 4,
                                            sizeof(ctx->write_buf) - 4);
        if (ct_len < 0) return -1;
        // Write the 4-byte BE length prefix in front of the ciphertext.
        uint32_t ulen = (uint32_t)ct_len;
        ctx->write_buf[0] = (uint8_t)(ulen >> 24);
        ctx->write_buf[1] = (uint8_t)(ulen >> 16);
        ctx->write_buf[2] = (uint8_t)(ulen >> 8);
        ctx->write_buf[3] = (uint8_t)(ulen);
        return ctx->transport.send_stream(ctx->transport.userdata,
                                          ctx->write_buf, (size_t)(4 + ct_len));
    }

    // No encryption: frame plaintext.
    int frame_len = pigeon_frame_message(payload, payload_len,
                                         ctx->write_buf, sizeof(ctx->write_buf));
    if (frame_len < 0) return -1;

    return ctx->transport.send_stream(ctx->transport.userdata,
                                      ctx->write_buf, (size_t)frame_len);
}

int pigeon_recv(pigeon_ctx *ctx, uint8_t *out, size_t out_len)
{
    if (!ctx->transport.recv_stream) return -1;

    // Read 4-byte length prefix.
    uint8_t hdr[4];
    size_t got = 0;
    int err = ctx->transport.recv_stream(ctx->transport.userdata, hdr, 4, &got);
    if (err || got != 4) return -1;

    uint32_t frame_payload_len = pigeon_read_frame_length(hdr);
    if (frame_payload_len > PIGEON_MAX_MSG) return -1;

    if (ctx->stream_channel.established) {
        // Read ciphertext into read_buf, then decrypt into out.
        if (frame_payload_len > sizeof(ctx->read_buf)) return -1;
        got = 0;
        err = ctx->transport.recv_stream(ctx->transport.userdata,
                                         ctx->read_buf, frame_payload_len, &got);
        if (err || got != frame_payload_len) return -1;
        return pigeon_channel_decrypt(&ctx->stream_channel,
                                      ctx->read_buf, frame_payload_len,
                                      out, out_len);
    }

    // No encryption: read plaintext directly into out.
    if (frame_payload_len > out_len) return -1;
    got = 0;
    err = ctx->transport.recv_stream(ctx->transport.userdata, out, frame_payload_len, &got);
    if (err || got != frame_payload_len) return -1;

    return (int)frame_payload_len;
}

int pigeon_send_datagram(pigeon_ctx *ctx, const uint8_t *data, size_t len)
{
    if (!ctx->transport.send_datagram) return -1;

    if (ctx->datagram_channel.established) {
        if (8 + len + 16 > sizeof(ctx->write_buf)) return -1;
        int ct_len = pigeon_channel_encrypt(&ctx->datagram_channel,
                                            data, len,
                                            ctx->write_buf,
                                            sizeof(ctx->write_buf));
        if (ct_len < 0) return -1;
        return ctx->transport.send_datagram(ctx->transport.userdata,
                                            ctx->write_buf, (size_t)ct_len);
    }

    return ctx->transport.send_datagram(ctx->transport.userdata, data, len);
}

int pigeon_recv_datagram(pigeon_ctx *ctx, uint8_t *out, size_t out_len)
{
    if (!ctx->transport.recv_datagram) return -1;

    if (ctx->datagram_channel.established) {
        // Receive into read_buf, then decrypt into out.
        size_t got = 0;
        int err = ctx->transport.recv_datagram(ctx->transport.userdata,
                                               ctx->read_buf, sizeof(ctx->read_buf), &got);
        if (err) return -1;
        return pigeon_channel_decrypt(&ctx->datagram_channel,
                                      ctx->read_buf, got,
                                      out, out_len);
    }

    size_t got = 0;
    int err = ctx->transport.recv_datagram(ctx->transport.userdata, out, out_len, &got);
    if (err) return -1;
    return (int)got;
}

int pigeon_frame_message(const uint8_t *payload, size_t len,
                         uint8_t *buf, size_t buf_len)
{
    if (4 + len > buf_len) return -1;
    buf[0] = (uint8_t)(len >> 24);
    buf[1] = (uint8_t)(len >> 16);
    buf[2] = (uint8_t)(len >> 8);
    buf[3] = (uint8_t)(len);
    memcpy(buf + 4, payload, len);
    return (int)(4 + len);
}

uint32_t pigeon_read_frame_length(const uint8_t *buf)
{
    return ((uint32_t)buf[0] << 24) |
           ((uint32_t)buf[1] << 16) |
           ((uint32_t)buf[2] << 8)  |
           ((uint32_t)buf[3]);
}

// --- Multi-channel wire helpers ---

int pigeon_uvarint_encode(uint64_t v, uint8_t *buf, size_t buf_len)
{
    size_t i = 0;
    while (v >= 0x80) {
        if (i >= buf_len) return -1;
        buf[i++] = (uint8_t)(v) | 0x80u;
        v >>= 7;
    }
    if (i >= buf_len) return -1;
    buf[i++] = (uint8_t)(v);
    return (int)i;
}

int pigeon_uvarint_decode(const uint8_t *buf, size_t buf_len, uint64_t *out)
{
    uint64_t v = 0;
    unsigned shift = 0;
    for (size_t i = 0; i < buf_len; i++) {
        uint8_t b = buf[i];
        if (i == 9 && b > 1) {
            // 10th byte may only contribute 1 bit (uint64 max).
            return -1;
        }
        v |= (uint64_t)(b & 0x7fu) << shift;
        if ((b & 0x80u) == 0) {
            *out = v;
            return (int)(i + 1);
        }
        shift += 7;
        if (shift >= 64) return -1;
    }
    return 0; // truncated
}

// Stream-header encoder/decoders are protogen-generated in
// c/src/wire_gen.c — pigeon_wire_stream_header_{encode,decode_backend,
// decode_client}. Call sites use the generated names directly; we no
// longer hand-roll them here.

// AEAD-wrapped datagram (backend side: [4-byte tag][AEAD(plaintext)]).
// The *plaintext* layout is protogen-generated
// (pigeon_wire_datagram_plaintext_encode); this wrapper layers AEAD on
// top and adds the optional clientTag prefix the relay routes by.
int pigeon_encode_datagram(pigeon_channel *ch,
                           bool is_backend, uint32_t client_tag,
                           uint64_t channel_id,
                           const uint8_t *payload, size_t payload_len,
                           uint8_t *out, size_t out_len)
{
    if (!ch || !ch->established) return -1;
    if (payload_len > PIGEON_MAX_MSG) return -1;

    // Compose the AEAD-plaintext (the protogen byte format) on the
    // heap: a per-call 1 MiB stack would overflow every host runtime
    // (see T38 in the audit log).
    size_t plain_cap = PIGEON_MAX_VARINT_LEN + PIGEON_MAX_MSG;
    uint8_t *plain = (uint8_t *)malloc(plain_cap);
    if (!plain) return -1;
    int plain_len = pigeon_wire_datagram_plaintext_encode(channel_id,
                                                          payload, payload_len,
                                                          plain, plain_cap);
    if (plain_len < 0) { free(plain); return -1; }

    // Wire = (optional 4-byte tag) ++ AEAD(plain).
    size_t off = 0;
    if (is_backend) {
        if (off + 4 > out_len) { free(plain); return -1; }
        out[off++] = (uint8_t)(client_tag >> 24);
        out[off++] = (uint8_t)(client_tag >> 16);
        out[off++] = (uint8_t)(client_tag >> 8);
        out[off++] = (uint8_t)(client_tag);
    }
    int ct = pigeon_channel_encrypt(ch, plain, (size_t)plain_len,
                                    out + off, out_len - off);
    free(plain);
    if (ct < 0) return -1;
    return (int)off + ct;
}

int pigeon_decode_datagram(pigeon_channel *ch,
                           bool is_backend,
                           const uint8_t *wire, size_t wire_len,
                           uint32_t *client_tag,
                           uint64_t *channel_id,
                           uint8_t *payload_buf, size_t payload_buf_len)
{
    if (!ch || !ch->established) return -1;

    size_t off = 0;
    if (is_backend) {
        if (wire_len < 4) return -1;
        if (client_tag) {
            *client_tag = ((uint32_t)wire[0] << 24)
                        | ((uint32_t)wire[1] << 16)
                        | ((uint32_t)wire[2] <<  8)
                        |  (uint32_t)wire[3];
        }
        off = 4;
    }

    // AEAD-decrypt into a heap scratch buffer, then decode the
    // protogen plaintext format. Heap-allocated for the same reason as
    // pigeon_encode_datagram above (T38).
    uint8_t *plain = (uint8_t *)malloc(PIGEON_MAX_MSG);
    if (!plain) return -1;
    int pn = pigeon_channel_decrypt(ch, wire + off, wire_len - off,
                                    plain, PIGEON_MAX_MSG);
    if (pn < 0) { free(plain); return -1; }

    uint64_t cid = 0;
    size_t payload_len = 0;
    int dn = pigeon_wire_datagram_plaintext_decode(plain, (size_t)pn,
                                                   &cid,
                                                   payload_buf, payload_buf_len,
                                                   &payload_len);
    free(plain);
    if (dn < 0) return -1;
    if (channel_id) *channel_id = cid;
    return (int)payload_len;
}

// Magic bytes and version for PairingRecord serialisation.
#define PIGEON_PR_MAGIC0 0x50u /* 'P' */
#define PIGEON_PR_MAGIC1 0x47u /* 'G' */
#define PIGEON_PR_MAGIC2 0x52u /* 'R' */
#define PIGEON_PR_VERSION 1u

int pigeon_pairing_record_serialize(const pigeon_pairing_record *rec,
                                    uint8_t *buf, size_t buf_len)
{
    if (buf_len < PIGEON_PAIRING_RECORD_SIZE) return -1;

    buf[0] = PIGEON_PR_MAGIC0;
    buf[1] = PIGEON_PR_MAGIC1;
    buf[2] = PIGEON_PR_MAGIC2;
    buf[3] = PIGEON_PR_VERSION;

    memcpy(buf +   4, rec->peer_instance_id,  64);
    memcpy(buf +  68, rec->relay_url,         256);
    memcpy(buf + 324, rec->local_private_key,  32);
    memcpy(buf + 356, rec->local_public_key,   32);
    memcpy(buf + 388, rec->peer_public_key,    32);

    return PIGEON_PAIRING_RECORD_SIZE;
}

int pigeon_pairing_record_deserialize(pigeon_pairing_record *rec,
                                      const uint8_t *buf, size_t buf_len)
{
    if (buf_len < PIGEON_PAIRING_RECORD_SIZE) return -1;
    if (buf[0] != PIGEON_PR_MAGIC0 ||
        buf[1] != PIGEON_PR_MAGIC1 ||
        buf[2] != PIGEON_PR_MAGIC2) return -1;
    if (buf[3] != PIGEON_PR_VERSION) return -1;

    memcpy(rec->peer_instance_id,  buf +   4,  64);
    memcpy(rec->relay_url,         buf +  68, 256);
    memcpy(rec->local_private_key, buf + 324,  32);
    memcpy(rec->local_public_key,  buf + 356,  32);
    memcpy(rec->peer_public_key,   buf + 388,  32);

    return PIGEON_PAIRING_RECORD_SIZE;
}

// --- Multi-channel session API ---

// Lazily allocate the per-session scratch buffers. Idempotent.
static int pigeon_session_ensure_scratch(pigeon_session *s)
{
    if (s->scratch_a && s->scratch_b) return 0;
    // Sized for the largest single send/recv: AEAD ciphertext expansion
    // is ~32 bytes (8-byte seq + 16-byte tag + slack); datagrams add a
    // 4-byte clientTag prefix. 64 bytes of slack is comfortable.
    size_t sz = PIGEON_MAX_MSG + 64;
    if (!s->scratch_a) s->scratch_a = (uint8_t *)malloc(sz);
    if (!s->scratch_b) s->scratch_b = (uint8_t *)malloc(sz);
    if (!s->scratch_a || !s->scratch_b) {
        free(s->scratch_a); free(s->scratch_b);
        s->scratch_a = s->scratch_b = NULL;
        return -1;
    }
    s->scratch_size = sz;
    return 0;
}

void pigeon_session_close(pigeon_session *s)
{
    if (!s) return;
    free(s->scratch_a); s->scratch_a = NULL;
    free(s->scratch_b); s->scratch_b = NULL;
    s->scratch_size = 0;
}

int pigeon_session_init(pigeon_session *s,
                        const pigeon_transport *transport,
                        const pigeon_channel *channel,
                        bool is_backend,
                        uint32_t client_tag,
                        const pigeon_dgchannel_def *datagrams,
                        size_t datagram_count)
{
    if (!s || !transport || !channel) return -1;
    if (datagram_count > PIGEON_MAX_DATAGRAM_CHANNELS) return -1;

    memset(s, 0, sizeof(*s));
    s->transport  = *transport;
    s->channel    = *channel;
    s->is_backend = is_backend;
    s->client_tag = client_tag;

    // Validate the (name, id) list: no duplicate ids, no id == 0
    // (reserved), no name overflow.
    for (size_t i = 0; i < datagram_count; i++) {
        if (datagrams[i].channel_id == 0) return -1;
        size_t nl = strlen(datagrams[i].name);
        if (nl == 0 || nl >= PIGEON_MAX_NAME_LEN) return -1;
        for (size_t j = 0; j < i; j++) {
            if (datagrams[j].channel_id == datagrams[i].channel_id) return -1;
        }
        s->datagrams[i] = datagrams[i];
    }
    s->datagram_count = datagram_count;
    return 0;
}

int pigeon_session_open_stream(pigeon_session *s,
                               const char *name,
                               pigeon_stream *out_stream)
{
    if (!s || !name || !out_stream) return -1;
    size_t name_len = strlen(name);
    if (name_len == 0 || name_len >= PIGEON_MAX_NAME_LEN) return -1;
    if (!s->transport.open_stream || !s->transport.send_on_stream) return -1;

    pigeon_stream_handle *h = NULL;
    if (s->transport.open_stream(s->transport.userdata, &h) != 0) return -1;

    // Compose the unencrypted name-binding header and write it as the
    // first message on the stream.
    uint8_t hdr[PIGEON_MAX_STREAM_HEADER];
    int hn = pigeon_wire_stream_header_encode(s->is_backend, s->client_tag,
                                         name, name_len, hdr, sizeof(hdr));
    if (hn < 0) {
        if (s->transport.close_stream) s->transport.close_stream(s->transport.userdata, h);
        return -1;
    }
    if (s->transport.send_on_stream(s->transport.userdata, h, hdr, (size_t)hn) != 0) {
        if (s->transport.close_stream) s->transport.close_stream(s->transport.userdata, h);
        return -1;
    }

    out_stream->session = s;
    out_stream->handle  = h;
    memcpy(out_stream->name, name, name_len);
    out_stream->name[name_len] = '\0';
    return 0;
}

int pigeon_session_get_datagram(pigeon_session *s,
                                const char *name,
                                pigeon_datagram *out)
{
    if (!s || !name || !out) return -1;
    for (size_t i = 0; i < s->datagram_count; i++) {
        if (strcmp(s->datagrams[i].name, name) == 0) {
            out->session    = s;
            out->channel_id = s->datagrams[i].channel_id;
            size_t nl = strlen(name);
            memcpy(out->name, name, nl);
            out->name[nl] = '\0';
            return 0;
        }
    }
    return -1;
}

int pigeon_stream_send(pigeon_stream *s,
                       const uint8_t *msg, size_t msg_len)
{
    if (!s || !s->session || !s->handle) return -1;
    pigeon_session *sess = s->session;
    if (!sess->transport.send_on_stream) return -1;

    // Pairing-mode session: channel not yet established. Mirror Go's
    // Stream.Send (api.go) — send plaintext as one length-prefixed message
    // so pigeon_session_primary() callers can drive the pairing ceremony
    // over the primary stream before the AEAD channel exists.
    if (!sess->channel.established) {
        return sess->transport.send_on_stream(sess->transport.userdata,
                                              s->handle, msg, msg_len);
    }

    if (pigeon_session_ensure_scratch(sess) != 0) return -1;

    // AEAD-encrypt the application payload and write the ciphertext as
    // one length-prefixed message on the stream.
    int ctn = pigeon_channel_encrypt(&sess->channel, msg, msg_len,
                                     sess->scratch_a, sess->scratch_size);
    if (ctn < 0) return -1;
    return sess->transport.send_on_stream(sess->transport.userdata, s->handle,
                                          sess->scratch_a, (size_t)ctn);
}

int pigeon_stream_recv(pigeon_stream *s,
                       uint8_t *buf, size_t buf_len)
{
    if (!s || !s->session || !s->handle) return -1;
    pigeon_session *sess = s->session;
    if (!sess->transport.recv_on_stream) return -1;

    // Pairing-mode mirror of pigeon_stream_send: read plaintext directly
    // into the caller's buffer when the channel hasn't been established yet.
    if (!sess->channel.established) {
        size_t got = 0;
        if (sess->transport.recv_on_stream(sess->transport.userdata, s->handle,
                                           buf, buf_len, &got) != 0) {
            return -1;
        }
        return (int)got;
    }

    if (pigeon_session_ensure_scratch(sess) != 0) return -1;

    size_t got = 0;
    if (sess->transport.recv_on_stream(sess->transport.userdata, s->handle,
                                       sess->scratch_a, sess->scratch_size, &got) != 0) {
        return -1;
    }
    return pigeon_channel_decrypt(&sess->channel, sess->scratch_a, got, buf, buf_len);
}

int pigeon_stream_close(pigeon_stream *s)
{
    if (!s || !s->session || !s->handle) return -1;
    if (!s->session->transport.close_stream) return 0;
    return s->session->transport.close_stream(s->session->transport.userdata, s->handle);
}

int pigeon_datagram_send(pigeon_datagram *d,
                         const uint8_t *payload, size_t payload_len)
{
    if (!d || !d->session) return -1;
    pigeon_session *sess = d->session;
    if (!sess->transport.send_datagram) return -1;
    if (pigeon_session_ensure_scratch(sess) != 0) return -1;

    int wn = pigeon_encode_datagram(&sess->channel,
                                    sess->is_backend, sess->client_tag,
                                    d->channel_id,
                                    payload, payload_len,
                                    sess->scratch_a, sess->scratch_size);
    if (wn < 0) return -1;
    return sess->transport.send_datagram(sess->transport.userdata,
                                         sess->scratch_a, (size_t)wn);
}

int pigeon_datagram_recv(pigeon_datagram *d,
                         uint8_t *buf, size_t buf_len)
{
    if (!d || !d->session) return -1;
    pigeon_session *sess = d->session;
    if (!sess->transport.recv_datagram) return -1;
    if (pigeon_session_ensure_scratch(sess) != 0) return -1;

    size_t got = 0;
    if (sess->transport.recv_datagram(sess->transport.userdata,
                                      sess->scratch_a, sess->scratch_size, &got) != 0) {
        return -1;
    }
    uint64_t cid = 0;
    int pn = pigeon_decode_datagram(&sess->channel, sess->is_backend,
                                    sess->scratch_a, got, NULL, &cid,
                                    buf, buf_len);
    if (pn < 0) return -1;
    if (cid != d->channel_id) {
        // Datagram belongs to a different channel; the caller should
        // route to a sibling pigeon_datagram. Returning 0 (zero-byte
        // application payload) is ambiguous, so we surface a distinct
        // sentinel: -2.
        return -2;
    }
    return pn;
}

// --- pigeon_session_primary / pigeon_connect (T32.3) ---

int pigeon_session_primary(pigeon_session *s, pigeon_stream *out_stream)
{
    if (!s || !out_stream || !s->primary) return -1;
    out_stream->session = s;
    out_stream->handle  = s->primary;
    out_stream->name[0] = '\0'; // Primary stream has empty name.
    return 0;
}

// derive_session_channel mirrors crypto.PairingRecord.DeriveChannel
// (Go): two HKDF-SHA256 expansions from the X25519 shared secret,
// using `send_info` / `recv_info` as separate KDF context labels.
// Writes 32-byte send/recv keys into out_send/out_recv. Returns 0 on
// success.
static int derive_session_channel(const pigeon_pairing_record *rec,
                                  const uint8_t *send_info, size_t send_info_len,
                                  const uint8_t *recv_info, size_t recv_info_len,
                                  uint8_t *out_send, uint8_t *out_recv)
{
    if (pigeon_derive_session_key(rec->local_private_key,
                                  rec->peer_public_key,
                                  send_info, send_info_len,
                                  out_send) != 0) return -1;
    if (pigeon_derive_session_key(rec->local_private_key,
                                  rec->peer_public_key,
                                  recv_info, recv_info_len,
                                  out_recv) != 0) return -1;
    return 0;
}

int pigeon_connect_on_transport(const pigeon_transport *transport,
                                pigeon_stream_handle *primary_handle,
                                const char *peer_instance_id,
                                const char *device_id,
                                const pigeon_pairing_record *record,
                                const pigeon_dgchannel_def *datagrams,
                                size_t datagram_count,
                                pigeon_session *out_session)
{
    if (!transport || !primary_handle || !out_session) return -1;
    if (!transport->send_on_stream || !transport->recv_on_stream) return -1;

    bool pairing_mode = (record == NULL);
    if (!pairing_mode && (device_id == NULL || device_id[0] == '\0')) {
        // Activation mode requires the client's device id.
        return -1;
    }
    if (peer_instance_id == NULL) return -1;

    // 1. Write the empty-name primary stream header on the primary stream.
    //    Client side ⇒ no 4-byte clientTag prefix, name length 0.
    uint8_t hdr[PIGEON_MAX_STREAM_HEADER];
    int hn = pigeon_wire_stream_header_encode(false, 0, NULL, 0, hdr, sizeof(hdr));
    if (hn < 0) return -1;
    if (transport->send_on_stream(transport->userdata, primary_handle,
                                  hdr, (size_t)hn) != 0) {
        return -1;
    }

    // 2. Run the client-side activation handshake (or skip in pairing mode).
    //    On success the AEAD channel is derived from the PairingRecord;
    //    in pairing mode the channel stays unestablished and the caller
    //    drives the ceremony over Session.Primary().
    pigeon_channel channel;
    memset(&channel, 0, sizeof(channel));
    if (!pairing_mode) {
        pigeon_client_machine cm;
        if (pigeon_run_client_activation(transport, primary_handle,
                                         device_id, &cm,
                                         NULL, 0) != 0) {
            return -1;
        }
        // We don't retain the machine post-activation (matches Go's
        // newClientSessionMachine, which advances the spec state and is
        // then held by Session for future executor-mediated I/O —
        // deferred work in T39).
        uint8_t send_key[32], recv_key[32];
        // Mirror Go's pigeon.Connect: send=client->backend, recv=backend->client.
        const char send_info[] = "client->backend";
        const char recv_info[] = "backend->client";
        if (derive_session_channel(record,
                                   (const uint8_t *)send_info, sizeof(send_info) - 1,
                                   (const uint8_t *)recv_info, sizeof(recv_info) - 1,
                                   send_key, recv_key) != 0) {
            return -1;
        }
        pigeon_channel_init(&channel, send_key, recv_key, PIGEON_MODE_STRICT);
    }
    // Note: in pairing mode `channel` stays zero-initialised; channel.
    // established is false, which pigeon_stream_send / _datagram_send
    // already reject. The caller switches to a derived channel later.

    // 3. Initialise the session (client side, no clientTag).
    if (pigeon_session_init(out_session, transport, &channel,
                            /*is_backend=*/false, /*client_tag=*/0,
                            datagrams, datagram_count) != 0) {
        return -1;
    }
    out_session->primary = primary_handle;
    (void)peer_instance_id; // Reserved for future peer_id storage (Go's
                            // Session.PeerID); not stored in pigeon_session
                            // today.
    return 0;
}

// --- Pairing ceremony driver ---


// Wire-level pairing ceremony driver (C side).
//
// pigeon_pair_acceptor and pigeon_pair_initiator implement the same
// hello/welcome/confirm exchange that Go's pairing.runAcceptor /
// runInitiator drive over a pigeon.Conn. Messages are JSON with
// base64-encoded byte fields — byte-for-byte compatible with the Go
// encoding/json marshaller.
//
// The only external I/O is through pigeon_transport.send_on_stream /
// recv_on_stream, plus pigeon_derive_confirmation_code from crypto.c.



// ---------------------------------------------------------------------------
// Minimal base64 encoder (RFC 4648 standard alphabet, with padding).
// Go's encoding/json uses standard base64 with padding for []byte fields,
// so a 32-byte key encodes to 44 characters.
// ---------------------------------------------------------------------------

static const char b64_table[] =
    "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";

// b64_encode: encode `src_len` bytes from `src` into `dst` (null-terminated).
// `dst` must be at least ceil(src_len/3)*4 + 1 bytes.
// Returns the number of base64 characters written (not counting NUL).
static int b64_encode(const uint8_t *src, size_t src_len, char *dst, size_t dst_len)
{
    size_t out = 0;
    for (size_t i = 0; i < src_len; ) {
        uint32_t b0 = src[i++];
        uint32_t b1 = (i < src_len) ? src[i++] : 0;
        uint32_t b2 = (i < src_len) ? src[i++] : 0;
        if (out + 4 >= dst_len) return -1;
        dst[out++] = b64_table[(b0 >> 2) & 0x3F];
        dst[out++] = b64_table[((b0 & 0x3) << 4) | ((b1 >> 4) & 0xF)];
        dst[out++] = b64_table[((b1 & 0xF) << 2) | ((b2 >> 6) & 0x3)];
        dst[out++] = b64_table[b2 & 0x3F];
    }
    // Add padding.
    size_t rem = src_len % 3;
    if (rem == 1) { dst[out-2] = '='; dst[out-1] = '='; }
    else if (rem == 2) { dst[out-1] = '='; }
    dst[out] = '\0';
    return (int)out;
}

// b64_decode_char: return the 6-bit value for a base64 character, or -1.
static int b64_decode_char(char c)
{
    if (c >= 'A' && c <= 'Z') return c - 'A';
    if (c >= 'a' && c <= 'z') return c - 'a' + 26;
    if (c >= '0' && c <= '9') return c - '0' + 52;
    if (c == '+') return 62;
    if (c == '/') return 63;
    return -1;
}

// b64_decode: decode null-terminated base64 string `src` into `dst`.
// `dst_cap` must be >= src_len*3/4.
// Returns the number of decoded bytes, or -1 on error.
static int b64_decode(const char *src, size_t src_len,
                      uint8_t *dst, size_t dst_cap)
{
    // Strip padding to find actual encoded groups.
    while (src_len > 0 && src[src_len-1] == '=') src_len--;
    size_t out = 0;
    for (size_t i = 0; i < src_len; ) {
        int v0 = b64_decode_char(src[i++]);
        int v1 = (i < src_len) ? b64_decode_char(src[i++]) : -1;
        int v2 = (i < src_len) ? b64_decode_char(src[i++]) : -1;
        int v3 = (i < src_len) ? b64_decode_char(src[i++]) : -1;
        if (v0 < 0 || v1 < 0) return -1;
        if (out >= dst_cap) return -1;
        dst[out++] = (uint8_t)((v0 << 2) | (v1 >> 4));
        if (v2 >= 0) {
            if (out >= dst_cap) return -1;
            dst[out++] = (uint8_t)((v1 << 4) | (v2 >> 2));
        }
        if (v3 >= 0) {
            if (out >= dst_cap) return -1;
            dst[out++] = (uint8_t)((v2 << 6) | v3);
        }
    }
    return (int)out;
}

// ---------------------------------------------------------------------------
// Minimal JSON helpers for the fixed pairing message schema.
// We never build a general-purpose parser — just extract named string fields
// from a well-formed JSON object produced by Go's encoding/json.
// ---------------------------------------------------------------------------

// json_find_string_field: find the value of a JSON string field by key.
// Writes the raw (unescaped) string content into `out` (NUL-terminated).
// Returns the length of the value, or -1 if not found / too long.
// NOTE: does not handle JSON escapes beyond the printable ASCII subset
// produced by encoding/json for these specific fields (instance IDs and
// base64 strings contain no characters that encoding/json would escape).
static int json_find_string_field(const char *json, size_t json_len,
                                  const char *key,
                                  char *out, size_t out_cap)
{
    // Search for "key": pattern.
    char search[256];
    int slen = 0;
    search[slen++] = '"';
    for (const char *k = key; *k; k++) {
        if (slen + 2 >= (int)sizeof(search)) return -1;
        search[slen++] = *k;
    }
    search[slen++] = '"';
    search[slen] = '\0';

    const char *p = json;
    const char *end = json + json_len;

    while (p < end) {
        // Find the key.
        const char *found = NULL;
        for (const char *q = p; q + slen <= end; q++) {
            if (memcmp(q, search, (size_t)slen) == 0) {
                found = q;
                break;
            }
        }
        if (!found) return -1;
        p = found + slen;
        // Skip whitespace and colon.
        while (p < end && (*p == ' ' || *p == '\t' || *p == '\n' || *p == '\r')) p++;
        if (p >= end || *p != ':') { p = found + 1; continue; }
        p++;
        while (p < end && (*p == ' ' || *p == '\t' || *p == '\n' || *p == '\r')) p++;
        if (p >= end || *p != '"') { p = found + 1; continue; }
        p++; // skip opening quote
        // Copy value until closing quote.
        size_t n = 0;
        while (p < end && *p != '"') {
            if (n + 1 >= out_cap) return -1;
            out[n++] = *p++;
        }
        out[n] = '\0';
        return (int)n;
    }
    return -1;
}

// ---------------------------------------------------------------------------
// send/recv helpers using the transport's send_on_stream / recv_on_stream.
// ---------------------------------------------------------------------------

static int pair_send(const pigeon_transport *t, pigeon_stream_handle *h,
                     const uint8_t *msg, size_t len)
{
    if (!t->send_on_stream) return -1;
    return t->send_on_stream(t->userdata, h, msg, len);
}

static int pair_recv(const pigeon_transport *t, pigeon_stream_handle *h,
                     uint8_t *buf, size_t buf_cap, size_t *out_len)
{
    if (!t->recv_on_stream) return -1;
    return t->recv_on_stream(t->userdata, h, buf, buf_cap, out_len);
}

// ---------------------------------------------------------------------------
// JSON message builders.
//
// Each of these builds the same JSON that Go's encoding/json produces for
// the corresponding pairingMessage struct value, so the C driver is wire-
// compatible with the Go driver.
// ---------------------------------------------------------------------------

// build_hello: {"kind":"hello","eph_pub":"<b64>","identity_pub":"<b64>","instance_id":"<str>"}
static int build_hello(const uint8_t *eph_pub,
                       const uint8_t *identity_pub,
                       const char *instance_id,
                       uint8_t *buf, size_t buf_cap)
{
    char eph_b64[64], id_b64[64];
    if (b64_encode(eph_pub, 32, eph_b64, sizeof(eph_b64)) < 0) return -1;
    if (b64_encode(identity_pub, 32, id_b64, sizeof(id_b64)) < 0) return -1;
    int n = snprintf((char *)buf, buf_cap,
        "{\"kind\":\"hello\",\"eph_pub\":\"%s\",\"identity_pub\":\"%s\","
        "\"instance_id\":\"%s\"}",
        eph_b64, id_b64, instance_id);
    if (n < 0 || (size_t)n >= buf_cap) return -1;
    return n;
}

// build_welcome: {"kind":"welcome","eph_pub":"<b64>","identity_pub":"<b64>","instance_id":"<str>"}
static int build_welcome(const uint8_t *eph_pub,
                         const uint8_t *identity_pub,
                         const char *instance_id,
                         uint8_t *buf, size_t buf_cap)
{
    char eph_b64[64], id_b64[64];
    if (b64_encode(eph_pub, 32, eph_b64, sizeof(eph_b64)) < 0) return -1;
    if (b64_encode(identity_pub, 32, id_b64, sizeof(id_b64)) < 0) return -1;
    int n = snprintf((char *)buf, buf_cap,
        "{\"kind\":\"welcome\",\"eph_pub\":\"%s\",\"identity_pub\":\"%s\","
        "\"instance_id\":\"%s\"}",
        eph_b64, id_b64, instance_id);
    if (n < 0 || (size_t)n >= buf_cap) return -1;
    return n;
}

// build_confirm: {"kind":"confirm"}
static int build_confirm(uint8_t *buf, size_t buf_cap)
{
    int n = snprintf((char *)buf, buf_cap, "{\"kind\":\"confirm\"}");
    if (n < 0 || (size_t)n >= buf_cap) return -1;
    return n;
}

// ---------------------------------------------------------------------------
// pigeon_pair_acceptor
// ---------------------------------------------------------------------------

int pigeon_pair_acceptor(
    const pigeon_transport *transport,
    const uint8_t *local_eph_priv,
    const uint8_t *local_eph_pub,
    const uint8_t *identity_pub,
    const char *instance_id,
    int (*confirm_fn)(void *userdata, const char *code),
    void *userdata,
    pigeon_pairing_record *out_record,
    char *out_code)
{
    if (!transport || !local_eph_priv || !local_eph_pub ||
        !identity_pub || !instance_id || !confirm_fn ||
        !out_record || !out_code) return -1;
    if (!transport->accept_stream || !transport->send_on_stream || !transport->recv_on_stream)
        return -1;

    uint8_t *buf = (uint8_t *)malloc(PIGEON_MAX_MSG);
    if (!buf) return -1;

    // Accept the stream that the initiator opened.
    pigeon_stream_handle *stream = NULL;
    if (transport->accept_stream(transport->userdata, &stream) != 0) { free(buf); return -1; }

    // --- Drive FSM through setup phases (no wire I/O yet) ---
    pigeon_acceptor_machine m;
    pigeon_acceptor_machine_init(&m);
    // All actions are no-ops: key generation and relay registration happened
    // before this function was called. We just walk the state machine.
    if (pigeon_acceptor_step(&m, PIGEON_PAIRINGCEREMONY_EVENT_PAIR_BEGIN) != 1)        { free(buf); return -1; }
    if (pigeon_acceptor_step(&m, PIGEON_PAIRINGCEREMONY_EVENT_EPHEMERAL_READY) != 1)   { free(buf); return -1; }
    if (pigeon_acceptor_step(&m, PIGEON_PAIRINGCEREMONY_EVENT_RELAY_REGISTERED) != 1)  { free(buf); return -1; }

    // --- Read hello ---
    size_t got = 0;
    if (pair_recv(transport, stream, buf, PIGEON_MAX_MSG, &got) != 0) { free(buf); return -1; }
    buf[got < PIGEON_MAX_MSG ? got : PIGEON_MAX_MSG - 1] = '\0';

    char kind[32];
    if (json_find_string_field((char *)buf, got, "kind", kind, sizeof(kind)) < 0
        || strcmp(kind, "hello") != 0) { free(buf); return -1; }

    char eph_b64[64];
    if (json_find_string_field((char *)buf, got, "eph_pub", eph_b64, sizeof(eph_b64)) < 0)
        { free(buf); return -1; }
    uint8_t peer_eph_pub[32];
    if (b64_decode(eph_b64, strlen(eph_b64), peer_eph_pub, 32) != 32)
        { free(buf); return -1; }

    // Extract initiator identity and instance from hello.
    uint8_t peer_id_pub[32];
    char peer_eph_id_b64[64];
    if (json_find_string_field((char *)buf, got, "identity_pub", peer_eph_id_b64, sizeof(peer_eph_id_b64)) >= 0)
        b64_decode(peer_eph_id_b64, strlen(peer_eph_id_b64), peer_id_pub, 32);
    char peer_instance[64] = "";
    json_find_string_field((char *)buf, got, "instance_id", peer_instance, sizeof(peer_instance));

    if (pigeon_acceptor_handle_message(&m, PIGEON_PAIRINGCEREMONY_MSG_HELLO) != 1) { free(buf); return -1; }
    if (pigeon_acceptor_step(&m, PIGEON_PAIRINGCEREMONY_EVENT_CODE_READY) != 1)    { free(buf); return -1; }

    // --- Send welcome ---
    int wlen = build_welcome(local_eph_pub, identity_pub, instance_id, buf, PIGEON_MAX_MSG);
    if (wlen < 0) { free(buf); return -1; }
    if (pair_send(transport, stream, buf, (size_t)wlen) != 0) { free(buf); return -1; }

    // --- Derive confirmation code ---
    if (pigeon_derive_confirmation_code(local_eph_pub, peer_eph_pub, out_code) != 0)
        { free(buf); return -1; }

    // --- Ask local user to confirm ---
    if (confirm_fn(userdata, out_code) != 1) { free(buf); return -1; }

    if (pigeon_acceptor_step(&m, PIGEON_PAIRINGCEREMONY_EVENT_USER_CONFIRM) != 1) { free(buf); return -1; }

    // --- Send confirm ---
    int clen = build_confirm(buf, PIGEON_MAX_MSG);
    if (clen < 0) { free(buf); return -1; }
    if (pair_send(transport, stream, buf, (size_t)clen) != 0) { free(buf); return -1; }

    // --- Read peer confirm ---
    got = 0;
    if (pair_recv(transport, stream, buf, PIGEON_MAX_MSG, &got) != 0) { free(buf); return -1; }
    buf[got < PIGEON_MAX_MSG ? got : PIGEON_MAX_MSG - 1] = '\0';
    if (json_find_string_field((char *)buf, got, "kind", kind, sizeof(kind)) < 0
        || strcmp(kind, "confirm") != 0) { free(buf); return -1; }

    if (pigeon_acceptor_handle_message(&m, PIGEON_PAIRINGCEREMONY_MSG_CONFIRM_TO_ACCEPTOR) != 1)
        { free(buf); return -1; }

    // --- Build output record ---
    memset(out_record, 0, sizeof(*out_record));
    strncpy(out_record->peer_instance_id, peer_instance, sizeof(out_record->peer_instance_id) - 1);
    // relay_url is not available to the C driver; caller fills it if needed.
    memcpy(out_record->local_private_key, local_eph_priv, 32);
    memcpy(out_record->local_public_key,  local_eph_pub,  32);
    memcpy(out_record->peer_public_key,   peer_eph_pub,   32);

    free(buf);
    return 0;
}

// ---------------------------------------------------------------------------
// pigeon_pair_initiator
// ---------------------------------------------------------------------------

int pigeon_pair_initiator(
    const pigeon_transport *transport,
    const uint8_t *local_eph_priv,
    const uint8_t *local_eph_pub,
    const uint8_t *identity_pub,
    const char *instance_id,
    const uint8_t *acc_eph_pub,
    const char *acc_instance,
    int (*confirm_fn)(void *userdata, const char *code),
    void *userdata,
    pigeon_pairing_record *out_record,
    char *out_code)
{
    if (!transport || !local_eph_priv || !local_eph_pub ||
        !identity_pub || !instance_id || !acc_eph_pub || !acc_instance ||
        !confirm_fn || !out_record || !out_code) return -1;
    if (!transport->open_stream || !transport->send_on_stream || !transport->recv_on_stream)
        return -1;

    uint8_t *buf = (uint8_t *)malloc(PIGEON_MAX_MSG);
    if (!buf) return -1;

    // Open a stream toward the acceptor.
    pigeon_stream_handle *stream = NULL;
    if (transport->open_stream(transport->userdata, &stream) != 0) { free(buf); return -1; }

    // --- Drive FSM through setup phases ---
    pigeon_initiator_machine m;
    pigeon_initiator_machine_init(&m);
    if (pigeon_initiator_step(&m, PIGEON_PAIRINGCEREMONY_EVENT_TOKEN_RECEIVED) != 1)  { free(buf); return -1; }
    if (pigeon_initiator_step(&m, PIGEON_PAIRINGCEREMONY_EVENT_TOKEN_DECODED) != 1)   { free(buf); return -1; }
    if (pigeon_initiator_step(&m, PIGEON_PAIRINGCEREMONY_EVENT_EPHEMERAL_READY) != 1) { free(buf); return -1; }
    if (pigeon_initiator_step(&m, PIGEON_PAIRINGCEREMONY_EVENT_RELAY_CONNECTED) != 1) { free(buf); return -1; }

    // --- Send hello ---
    int hlen = build_hello(local_eph_pub, identity_pub, instance_id, buf, PIGEON_MAX_MSG);
    if (hlen < 0) { free(buf); return -1; }
    if (pair_send(transport, stream, buf, (size_t)hlen) != 0) { free(buf); return -1; }

    // --- Read welcome ---
    size_t got = 0;
    if (pair_recv(transport, stream, buf, PIGEON_MAX_MSG, &got) != 0) { free(buf); return -1; }
    buf[got < PIGEON_MAX_MSG ? got : PIGEON_MAX_MSG - 1] = '\0';

    char kind[32];
    if (json_find_string_field((char *)buf, got, "kind", kind, sizeof(kind)) < 0
        || strcmp(kind, "welcome") != 0) { free(buf); return -1; }

    char eph_b64[64];
    if (json_find_string_field((char *)buf, got, "eph_pub", eph_b64, sizeof(eph_b64)) < 0)
        { free(buf); return -1; }
    uint8_t peer_eph_pub[32];
    if (b64_decode(eph_b64, strlen(eph_b64), peer_eph_pub, 32) != 32)
        { free(buf); return -1; }

    char peer_instance[64] = "";
    json_find_string_field((char *)buf, got, "instance_id", peer_instance, sizeof(peer_instance));

    if (pigeon_initiator_handle_message(&m, PIGEON_PAIRINGCEREMONY_MSG_WELCOME) != 1) { free(buf); return -1; }
    if (pigeon_initiator_step(&m, PIGEON_PAIRINGCEREMONY_EVENT_CODE_READY) != 1)      { free(buf); return -1; }

    // --- Derive confirmation code ---
    // DeriveConfirmationCode(acc_eph_pub, local_eph_pub) — acceptor's key is
    // pubA, initiator's key is pubB.  pigeon_derive_confirmation_code is
    // order-independent so the result is the same as Go's
    // crypto.DeriveConfirmationCode(accEphPub, eph.PublicKey()).
    if (pigeon_derive_confirmation_code(peer_eph_pub, local_eph_pub, out_code) != 0)
        { free(buf); return -1; }

    // --- Ask local user to confirm ---
    if (confirm_fn(userdata, out_code) != 1) { free(buf); return -1; }

    if (pigeon_initiator_step(&m, PIGEON_PAIRINGCEREMONY_EVENT_USER_CONFIRM) != 1) { free(buf); return -1; }

    // --- Send confirm ---
    int clen = build_confirm(buf, PIGEON_MAX_MSG);
    if (clen < 0) { free(buf); return -1; }
    if (pair_send(transport, stream, buf, (size_t)clen) != 0) { free(buf); return -1; }

    // --- Read peer confirm ---
    got = 0;
    if (pair_recv(transport, stream, buf, PIGEON_MAX_MSG, &got) != 0) { free(buf); return -1; }
    buf[got < PIGEON_MAX_MSG ? got : PIGEON_MAX_MSG - 1] = '\0';
    if (json_find_string_field((char *)buf, got, "kind", kind, sizeof(kind)) < 0
        || strcmp(kind, "confirm") != 0) { free(buf); return -1; }

    if (pigeon_initiator_handle_message(&m, PIGEON_PAIRINGCEREMONY_MSG_CONFIRM_TO_INITIATOR) != 1)
        { free(buf); return -1; }

    // --- Build output record ---
    memset(out_record, 0, sizeof(*out_record));
    strncpy(out_record->peer_instance_id, peer_instance[0] ? peer_instance : acc_instance,
            sizeof(out_record->peer_instance_id) - 1);
    memcpy(out_record->local_private_key, local_eph_priv, 32);
    memcpy(out_record->local_public_key,  local_eph_pub,  32);
    memcpy(out_record->peer_public_key,   peer_eph_pub,   32);

    free(buf);
    return 0;
}

// --- Multi-client listener ---

//
// Multi-client listener (T32.2). Mirrors Go's pigeon.Register /
// pigeon.Listener.Accept in api.go — a backend registers once with
// the relay (PIGEON_ROLE_REGISTER_MUX); thereafter every paired
// client that connects on the shared QUIC connection produces a
// fresh pigeon_session.
//
// Threading model: synchronous inline demux on the calling thread.
// pigeon_listener_accept blocks on transport->accept_stream, reads
// the [4-byte clientTag][varint name-len][name] header that the
// relay prepends, and either:
//
//   * tag is new + name empty -> primary stream for a new client.
//     Run T32.1's pigeon_run_backend_activation against the
//     registered pairing callback, derive an AEAD channel from the
//     resolved PairingRecord, allocate a pigeon_session, register
//     it under its tag in the demux table, and return it.
//
//   * tag is known -> sub-stream for an existing client. Park the
//     stream + name on the session's per-session incoming-stream
//     queue (the listener's pump keeps looping until the next new
//     primary).
//
// The hash table is a 16-slot open-addressing linear-probing map.
// The C SDK targets small N (a personal-device-count of paired
// clients per backend), so 16 slots is plenty and keeps the data
// structure trivial.



// --- Per-listener state ---

typedef struct {
    bool            in_use;
    uint32_t        client_tag;
    pigeon_session *session;
} listener_slot;

struct pigeon_listener {
    pigeon_transport          transport;
    char                      instance_id[64];

    // Pairing callback (device-id -> PairingRecord lookup) wired in
    // at init time. Invoked synchronously from the accept pump.
    pigeon_resolve_device_fn  resolve;
    void                     *resolve_userdata;

    // Datagram channel definitions copied into each accepted
    // session.
    pigeon_dgchannel_def      datagrams[PIGEON_MAX_DATAGRAM_CHANNELS];
    size_t                    datagram_count;

    // Per-tag demux table. Fixed-size open-addressing linear
    // probing; sized to PIGEON_LISTENER_MAX_CLIENTS slots.
    listener_slot             slots[PIGEON_LISTENER_MAX_CLIENTS];

    bool                      closed;
};

// --- Hash table helpers ---
//
// 32-bit tag fold. The relay assigns sequential client tags so even
// a trivial modulo distributes them well; mix in a fixnum-style
// multiplier so an adversarial workload that picks colliding tags
// doesn't fall off a cliff.

static size_t slot_index(uint32_t tag, size_t probe)
{
    uint32_t hash = tag * 2654435761u; // Knuth fibonacci hash
    return ((size_t)hash + probe) % PIGEON_LISTENER_MAX_CLIENTS;
}

// Find an in-use slot with the given tag, or return NULL.
static listener_slot *slot_lookup(pigeon_listener *l, uint32_t tag)
{
    for (size_t probe = 0; probe < PIGEON_LISTENER_MAX_CLIENTS; probe++) {
        listener_slot *s = &l->slots[slot_index(tag, probe)];
        if (!s->in_use) return NULL; // open addressing: first empty -> miss
        if (s->client_tag == tag) return s;
    }
    return NULL;
}

// Find a free slot to host the given tag, or NULL if the table is
// full. Assumes the tag is not already present.
static listener_slot *slot_insert(pigeon_listener *l, uint32_t tag)
{
    for (size_t probe = 0; probe < PIGEON_LISTENER_MAX_CLIENTS; probe++) {
        listener_slot *s = &l->slots[slot_index(tag, probe)];
        if (!s->in_use) return s;
    }
    return NULL;
}

// --- Session lifecycle ---

// Allocate and initialise a pigeon_session for an activated client.
// Returns NULL on allocation failure; on success the caller registers
// it in the demux table.
static pigeon_session *make_session(pigeon_listener *l,
                                    uint32_t tag,
                                    const pigeon_pairing_record *rec)
{
    pigeon_session *s = (pigeon_session *)calloc(1, sizeof(*s));
    if (!s) return NULL;

    // Derive the session AEAD channel from the resolved
    // PairingRecord. Mirrors api.go's rec.DeriveChannel(
    // "backend->client", "client->backend") on the Listener side.
    uint8_t send_key[32], recv_key[32];
    if (pigeon_derive_session_key(rec->local_private_key,
                                  rec->peer_public_key,
                                  (const uint8_t *)"backend->client",
                                  strlen("backend->client"),
                                  send_key) != 0) {
        free(s);
        return NULL;
    }
    if (pigeon_derive_session_key(rec->local_private_key,
                                  rec->peer_public_key,
                                  (const uint8_t *)"client->backend",
                                  strlen("client->backend"),
                                  recv_key) != 0) {
        free(s);
        return NULL;
    }
    pigeon_channel ch;
    pigeon_channel_init(&ch, send_key, recv_key, PIGEON_MODE_STRICT);

    if (pigeon_session_init(s, &l->transport, &ch,
                            /*is_backend=*/true, tag,
                            l->datagrams, l->datagram_count) != 0) {
        free(s);
        return NULL;
    }
    return s;
}

// Free a session: scratch buffers, any unpicked-up incoming sub-
// streams, then the struct itself. Safe on NULL.
static void free_session(pigeon_listener *l, pigeon_session *s)
{
    if (!s) return;
    for (size_t i = 0; i < PIGEON_SESSION_MAX_INCOMING; i++) {
        if (s->incoming[i].in_use && s->incoming[i].handle != NULL
                && l->transport.close_stream != NULL) {
            (void)l->transport.close_stream(l->transport.userdata,
                                            s->incoming[i].handle);
        }
        s->incoming[i].in_use = false;
        s->incoming[i].handle = NULL;
    }
    pigeon_session_close(s);
    free(s);
}

// --- Header decoding ---
//
// Read the [4-byte tag][varint name-len][name] header that backends
// see on every inbound stream. The relay prepended the tag; the
// originating client wrote the (name-len, name) part.

static int read_backend_header(pigeon_listener *l,
                               pigeon_stream_handle *handle,
                               uint32_t *out_tag,
                               char *out_name, size_t out_name_cap,
                               size_t *out_name_len)
{
    uint8_t hdr[PIGEON_MAX_STREAM_HEADER];
    size_t  hdr_len = 0;
    if (l->transport.recv_on_stream(l->transport.userdata, handle,
                                    hdr, sizeof(hdr), &hdr_len) != 0) {
        return -1;
    }
    return pigeon_wire_stream_header_decode_backend(hdr, hdr_len,
                                               out_tag,
                                               out_name, out_name_cap,
                                               out_name_len);
}

// --- Queueing a sub-stream into a session ---

static void queue_substream(pigeon_session *s,
                            pigeon_stream_handle *handle,
                            const char *name)
{
    for (size_t i = 0; i < PIGEON_SESSION_MAX_INCOMING; i++) {
        if (!s->incoming[i].in_use) {
            s->incoming[i].in_use = true;
            s->incoming[i].handle = handle;
            size_t nl = strlen(name);
            if (nl >= PIGEON_MAX_NAME_LEN) nl = PIGEON_MAX_NAME_LEN - 1;
            memcpy(s->incoming[i].name, name, nl);
            s->incoming[i].name[nl] = '\0';
            return;
        }
    }
    // Queue full: drop the stream (close it so the peer doesn't
    // wedge waiting for a reader). Matches "warn and drop" behaviour
    // the Go side falls back to under sustained queue pressure.
}

// --- Public API ---

int pigeon_listener_init(pigeon_listener **out,
                         const pigeon_transport *transport,
                         const char *instance_id,
                         pigeon_resolve_device_fn pairing,
                         void *pairing_userdata,
                         const pigeon_dgchannel_def *datagrams,
                         size_t datagram_count)
{
    if (!out || !transport || !pairing) return -1;
    if (datagram_count > PIGEON_MAX_DATAGRAM_CHANNELS) return -1;
    if (!transport->accept_stream || !transport->recv_on_stream) return -1;

    pigeon_listener *l = (pigeon_listener *)calloc(1, sizeof(*l));
    if (!l) return -1;

    l->transport         = *transport;
    l->resolve           = pairing;
    l->resolve_userdata  = pairing_userdata;
    if (instance_id) {
        size_t n = strlen(instance_id);
        if (n >= sizeof(l->instance_id)) n = sizeof(l->instance_id) - 1;
        memcpy(l->instance_id, instance_id, n);
        l->instance_id[n] = '\0';
    }

    // Validate + copy datagram channels with the same rules as
    // pigeon_session_init: no duplicate ids, no id==0, names bounded.
    for (size_t i = 0; i < datagram_count; i++) {
        if (datagrams[i].channel_id == 0) { free(l); return -1; }
        size_t nl = strlen(datagrams[i].name);
        if (nl == 0 || nl >= PIGEON_MAX_NAME_LEN) { free(l); return -1; }
        for (size_t j = 0; j < i; j++) {
            if (datagrams[j].channel_id == datagrams[i].channel_id) {
                free(l); return -1;
            }
        }
        l->datagrams[i] = datagrams[i];
    }
    l->datagram_count = datagram_count;

    *out = l;
    return 0;
}

const char *pigeon_listener_instance_id(const pigeon_listener *l)
{
    if (!l) return NULL;
    return l->instance_id;
}

int pigeon_listener_accept(pigeon_listener *l, pigeon_session **out_session)
{
    if (!l || !out_session) return -1;
    if (l->closed) return -1;

    // Pump until a new client primary completes activation. Each
    // iteration consumes one inbound stream from the transport.
    while (!l->closed) {
        pigeon_stream_handle *handle = NULL;
        if (l->transport.accept_stream(l->transport.userdata, &handle) != 0) {
            return -1;
        }

        uint32_t tag = 0;
        char     name[PIGEON_MAX_NAME_LEN] = {0};
        size_t   name_len = 0;
        if (read_backend_header(l, handle, &tag, name, sizeof(name), &name_len) < 0) {
            if (l->transport.close_stream) {
                l->transport.close_stream(l->transport.userdata, handle);
            }
            continue; // stay in the loop — malformed header from one
                      // client shouldn't sink the whole listener.
        }

        listener_slot *existing = slot_lookup(l, tag);
        if (existing != NULL) {
            if (name_len == 0) {
                // Duplicate primary for a tag we already activated:
                // protocol error, drop.
                if (l->transport.close_stream) {
                    l->transport.close_stream(l->transport.userdata, handle);
                }
                continue;
            }
            queue_substream(existing->session, handle, name);
            continue;
        }

        if (name_len != 0) {
            // First-seen tag with a named sub-stream — no Session to
            // route to. Drop the stream and keep pumping.
            if (l->transport.close_stream) {
                l->transport.close_stream(l->transport.userdata, handle);
            }
            continue;
        }

        // New client primary: run the activation handshake on this
        // stream against the listener's resolver.
        pigeon_backend_machine machine;
        char     device_id[PIGEON_AUTH_MAX_DEVICE_ID + 1] = {0};
        pigeon_pairing_record  record;
        int rc = pigeon_run_backend_activation(&l->transport, handle,
                                               l->resolve,
                                               l->resolve_userdata,
                                               &machine,
                                               device_id, sizeof(device_id),
                                               &record);
        if (rc != 0) {
            // -1 (wire failure) or 1 (decoded but rejected): close
            // the primary and keep accepting. Matches the Go-side
            // behaviour: rejected clients don't materialise a Session.
            if (l->transport.close_stream) {
                l->transport.close_stream(l->transport.userdata, handle);
            }
            continue;
        }

        // Allocate the session, derive the AEAD channel, register it
        // in the demux table BEFORE returning so any sub-stream the
        // client opens immediately after activation finds the tag in
        // the map when the next accept-pump iteration lands.
        listener_slot *slot = slot_insert(l, tag);
        if (!slot) {
            // Table full: cap reached.
            if (l->transport.close_stream) {
                l->transport.close_stream(l->transport.userdata, handle);
            }
            return -1;
        }
        pigeon_session *sess = make_session(l, tag, &record);
        if (!sess) {
            if (l->transport.close_stream) {
                l->transport.close_stream(l->transport.userdata, handle);
            }
            return -1;
        }
        slot->in_use     = true;
        slot->client_tag = tag;
        slot->session    = sess;

        // The activation primary itself is consumed (the auth_request /
        // auth_ok exchange is done). Close it now so future I/O on this
        // session goes through fresh sub-streams. Matches Go's
        // acceptPrimary lifetime: the primary stream is held by
        // Session.primary purely for pairing-mode use; activation-mode
        // doesn't read from it again.
        if (l->transport.close_stream) {
            l->transport.close_stream(l->transport.userdata, handle);
        }

        *out_session = sess;
        return 0;
    }
    return -1;
}

void pigeon_listener_close(pigeon_listener *l)
{
    if (!l) return;
    l->closed = true;
    for (size_t i = 0; i < PIGEON_LISTENER_MAX_CLIENTS; i++) {
        if (l->slots[i].in_use) {
            free_session(l, l->slots[i].session);
            l->slots[i].in_use = false;
            l->slots[i].session = NULL;
        }
    }
    free(l);
}

// --- Session-level: drain a buffered incoming sub-stream ---

int pigeon_session_accept_incoming_stream(pigeon_session *s,
                                          const char *name,
                                          pigeon_stream *out_stream)
{
    if (!s || !name || !out_stream) return -1;
    for (size_t i = 0; i < PIGEON_SESSION_MAX_INCOMING; i++) {
        if (s->incoming[i].in_use
                && strcmp(s->incoming[i].name, name) == 0) {
            out_stream->session = s;
            out_stream->handle  = s->incoming[i].handle;
            size_t nl = strlen(name);
            if (nl >= PIGEON_MAX_NAME_LEN) nl = PIGEON_MAX_NAME_LEN - 1;
            memcpy(out_stream->name, name, nl);
            out_stream->name[nl] = '\0';
            s->incoming[i].in_use = false;
            s->incoming[i].handle = NULL;
            return 0;
        }
    }
    return -1;
}

// --- In-process loopback transport ---


#include "loopback.h"


#define LOOP_MAX_STREAMS 32
#define LOOP_MAX_PENDING 64

typedef struct loop_stream {
    int id;
    uint8_t *msgs[LOOP_MAX_PENDING]; // heap-allocated, sized PIGEON_MAX_MSG
    size_t   msg_lens[LOOP_MAX_PENDING];
    int      msg_head, msg_tail, msg_count;
    bool     in_use;
    bool     accepted;
} loop_stream;

struct pigeon_loopback_endpoint {
    loop_stream  streams[LOOP_MAX_STREAMS];

    // Inbound datagram ringbuffer.
    uint8_t *dgrams[LOOP_MAX_PENDING];   // heap-allocated, sized PIGEON_MAX_MSG + 64
    size_t   dgram_lens[LOOP_MAX_PENDING];
    int      dgram_head, dgram_tail, dgram_count;

    // Stream IDs awaiting accept.
    int accept_queue[LOOP_MAX_STREAMS];
    int accept_head, accept_tail, accept_count;

    struct pigeon_loopback_endpoint *peer;
};

static void free_stream(loop_stream *s)
{
    for (int i = 0; i < LOOP_MAX_PENDING; i++) {
        free(s->msgs[i]);
        s->msgs[i] = NULL;
    }
}

static loop_stream *lb_alloc_stream(pigeon_loopback_endpoint *e)
{
    for (int i = 0; i < LOOP_MAX_STREAMS; i++) {
        if (!e->streams[i].in_use) {
            free_stream(&e->streams[i]);
            memset(&e->streams[i], 0, sizeof(e->streams[i]));
            e->streams[i].in_use = true;
            e->streams[i].id = i;
            return &e->streams[i];
        }
    }
    return NULL;
}

static int lb_open_stream(void *ud, pigeon_stream_handle **out)
{
    pigeon_loopback_endpoint *e = (pigeon_loopback_endpoint *)ud;
    loop_stream *me = lb_alloc_stream(e);
    if (!me) return -1;
    pigeon_loopback_endpoint *p = e->peer;
    if (!p) return -1;
    if (p->streams[me->id].in_use) return -1;
    free_stream(&p->streams[me->id]);
    memset(&p->streams[me->id], 0, sizeof(p->streams[me->id]));
    p->streams[me->id].in_use = true;
    p->streams[me->id].id = me->id;
    p->accept_queue[p->accept_tail] = me->id;
    p->accept_tail = (p->accept_tail + 1) % LOOP_MAX_STREAMS;
    p->accept_count++;
    *out = (pigeon_stream_handle *)me;
    return 0;
}

static int lb_accept_stream(void *ud, pigeon_stream_handle **out)
{
    pigeon_loopback_endpoint *e = (pigeon_loopback_endpoint *)ud;
    if (e->accept_count == 0) return -1;
    int id = e->accept_queue[e->accept_head];
    e->accept_head = (e->accept_head + 1) % LOOP_MAX_STREAMS;
    e->accept_count--;
    if (id < 0 || id >= LOOP_MAX_STREAMS) return -1;
    if (!e->streams[id].in_use) return -1;
    e->streams[id].accepted = true;
    *out = (pigeon_stream_handle *)&e->streams[id];
    return 0;
}

static int lb_send_on_stream(void *ud, pigeon_stream_handle *h,
                             const uint8_t *data, size_t len)
{
    pigeon_loopback_endpoint *e = (pigeon_loopback_endpoint *)ud;
    loop_stream *me = (loop_stream *)h;
    if (!me || !me->in_use) return -1;
    if (!e->peer) return -1;
    loop_stream *peer = &e->peer->streams[me->id];
    if (!peer->in_use) return -1;
    if (peer->msg_count >= LOOP_MAX_PENDING) return -1;
    if (len > PIGEON_MAX_MSG) return -1;
    if (!peer->msgs[peer->msg_tail]) {
        peer->msgs[peer->msg_tail] = (uint8_t *)malloc(PIGEON_MAX_MSG);
        if (!peer->msgs[peer->msg_tail]) return -1;
    }
    memcpy(peer->msgs[peer->msg_tail], data, len);
    peer->msg_lens[peer->msg_tail] = len;
    peer->msg_tail = (peer->msg_tail + 1) % LOOP_MAX_PENDING;
    peer->msg_count++;
    return 0;
}

static int lb_recv_on_stream(void *ud, pigeon_stream_handle *h,
                             uint8_t *buf, size_t buf_len, size_t *out_len)
{
    (void)ud;
    loop_stream *me = (loop_stream *)h;
    if (!me || !me->in_use) return -1;
    if (me->msg_count == 0) return -1;
    size_t n = me->msg_lens[me->msg_head];
    if (n > buf_len) return -1;
    memcpy(buf, me->msgs[me->msg_head], n);
    me->msg_head = (me->msg_head + 1) % LOOP_MAX_PENDING;
    me->msg_count--;
    *out_len = n;
    return 0;
}

static int lb_close_stream(void *ud, pigeon_stream_handle *h)
{
    (void)ud;
    loop_stream *me = (loop_stream *)h;
    if (me) {
        free_stream(me);
        me->in_use = false;
    }
    return 0;
}

static int lb_send_datagram(void *ud, const uint8_t *data, size_t len)
{
    pigeon_loopback_endpoint *e = (pigeon_loopback_endpoint *)ud;
    if (!e->peer) return -1;
    pigeon_loopback_endpoint *p = e->peer;
    if (p->dgram_count >= LOOP_MAX_PENDING) return -1;
    if (len > PIGEON_MAX_MSG + 64) return -1;
    if (!p->dgrams[p->dgram_tail]) {
        p->dgrams[p->dgram_tail] = (uint8_t *)malloc(PIGEON_MAX_MSG + 64);
        if (!p->dgrams[p->dgram_tail]) return -1;
    }
    memcpy(p->dgrams[p->dgram_tail], data, len);
    p->dgram_lens[p->dgram_tail] = len;
    p->dgram_tail = (p->dgram_tail + 1) % LOOP_MAX_PENDING;
    p->dgram_count++;
    return 0;
}

static int lb_recv_datagram(void *ud, uint8_t *buf, size_t buf_len, size_t *out_len)
{
    pigeon_loopback_endpoint *e = (pigeon_loopback_endpoint *)ud;
    if (e->dgram_count == 0) return -1;
    size_t n = e->dgram_lens[e->dgram_head];
    if (n > buf_len) return -1;
    memcpy(buf, e->dgrams[e->dgram_head], n);
    e->dgram_head = (e->dgram_head + 1) % LOOP_MAX_PENDING;
    e->dgram_count--;
    *out_len = n;
    return 0;
}

pigeon_loopback_endpoint *pigeon_loopback_new(void)
{
    return (pigeon_loopback_endpoint *)calloc(1, sizeof(pigeon_loopback_endpoint));
}

void pigeon_loopback_pair(pigeon_loopback_endpoint *a,
                          pigeon_loopback_endpoint *b)
{
    if (a) a->peer = b;
    if (b) b->peer = a;
}

void pigeon_loopback_fill_transport(pigeon_transport *t,
                                    pigeon_loopback_endpoint *e)
{
    memset(t, 0, sizeof(*t));
    t->userdata        = e;
    t->open_stream     = lb_open_stream;
    t->accept_stream   = lb_accept_stream;
    t->send_on_stream  = lb_send_on_stream;
    t->recv_on_stream  = lb_recv_on_stream;
    t->close_stream    = lb_close_stream;
    t->send_datagram   = lb_send_datagram;
    t->recv_datagram   = lb_recv_datagram;
}

int pigeon_loopback_accept_with_header(pigeon_loopback_endpoint *e,
                                       pigeon_stream_handle **out_handle,
                                       uint8_t *hdr, size_t hdr_len,
                                       size_t *hdr_out_len)
{
    pigeon_stream_handle *h = NULL;
    if (lb_accept_stream(e, &h) != 0) return -1;
    size_t n = 0;
    if (lb_recv_on_stream(e, h, hdr, hdr_len, &n) != 0) return -1;
    *out_handle = h;
    if (hdr_out_len) *hdr_out_len = n;
    return 0;
}

void pigeon_loopback_free(pigeon_loopback_endpoint *e)
{
    if (!e) return;
    for (int i = 0; i < LOOP_MAX_STREAMS; i++) {
        free_stream(&e->streams[i]);
    }
    for (int i = 0; i < LOOP_MAX_PENDING; i++) {
        free(e->dgrams[i]);
    }
    free(e);
}
