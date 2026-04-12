// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Code generated from protocol/*.yaml. DO NOT EDIT.

#include "pairingceremony_gen.h"
#include <string.h>

void pigeon_server_machine_init(pigeon_server_machine *m)
{
	memset(m, 0, sizeof(*m));
	m->state = PIGEON_SERVER_IDLE;
	m->current_token = "none";
	m->server_ecdh_pub = "none";
	m->received_client_pub = "none";
	m->code_attempts = 0;
	m->device_secret = "none";
	m->received_device_id = "none";
	m->received_auth_nonce = "none";
}

int pigeon_server_handle_message(pigeon_server_machine *m, pairing_ceremony_msg_type msg)
{
	if (m->state == PIGEON_SERVER_IDLE && msg == PIGEON_MSG_PAIR_BEGIN) {
		if (m->actions[PIGEON_ACTION_GENERATE_TOKEN]) {
			int err = m->actions[PIGEON_ACTION_GENERATE_TOKEN](m->userdata);
			if (err) return -err;
		}
		m->current_token = "tok_1";
		if (m->on_change) m->on_change("current_token", m->userdata);
		// active_tokens: active_tokens \union {"tok_1"} (set by action)
		m->state = PIGEON_SERVER_GENERATE_TOKEN;
		return 1;
	}
	if (m->state == PIGEON_SERVER_WAITING_FOR_CLIENT && msg == PIGEON_MSG_PAIR_HELLO && m->guards[PIGEON_GUARD_TOKEN_VALID] && m->guards[PIGEON_GUARD_TOKEN_VALID](m->userdata)) {
		if (m->actions[PIGEON_ACTION_DERIVE_SECRET]) {
			int err = m->actions[PIGEON_ACTION_DERIVE_SECRET](m->userdata);
			if (err) return -err;
		}
		// received_client_pub: recv_msg.pubkey (set by action)
		m->server_ecdh_pub = "server_pub";
		if (m->on_change) m->on_change("server_ecdh_pub", m->userdata);
		// server_shared_key: DeriveKey("server_pub", recv_msg.pubkey) (set by action)
		// server_code: DeriveCode("server_pub", recv_msg.pubkey) (set by action)
		m->state = PIGEON_SERVER_DERIVE_SECRET;
		return 1;
	}
	if (m->state == PIGEON_SERVER_WAITING_FOR_CLIENT && msg == PIGEON_MSG_PAIR_HELLO && m->guards[PIGEON_GUARD_TOKEN_INVALID] && m->guards[PIGEON_GUARD_TOKEN_INVALID](m->userdata)) {
		m->state = PIGEON_SERVER_IDLE;
		return 1;
	}
	if (m->state == PIGEON_SERVER_WAITING_FOR_CODE && msg == PIGEON_MSG_CODE_SUBMIT) {
		// received_code: recv_msg.code (set by action)
		m->state = PIGEON_SERVER_VALIDATE_CODE;
		return 1;
	}
	if (m->state == PIGEON_SERVER_PAIRED && msg == PIGEON_MSG_AUTH_REQUEST) {
		// received_device_id: recv_msg.device_id (set by action)
		// received_auth_nonce: recv_msg.nonce (set by action)
		m->state = PIGEON_SERVER_AUTH_CHECK;
		return 1;
	}
	return 0;
}

int pigeon_server_step(pigeon_server_machine *m, pairing_ceremony_event_id event)
{
	if (m->state == PIGEON_SERVER_GENERATE_TOKEN && event == PIGEON_EVENT_TOKEN_CREATED) {
		if (m->actions[PIGEON_ACTION_REGISTER_RELAY]) {
			int err = m->actions[PIGEON_ACTION_REGISTER_RELAY](m->userdata);
			if (err) return -err;
		}
		m->state = PIGEON_SERVER_REGISTER_RELAY;
		return 1;
	}
	if (m->state == PIGEON_SERVER_REGISTER_RELAY && event == PIGEON_EVENT_RELAY_REGISTERED) {
		m->state = PIGEON_SERVER_WAITING_FOR_CLIENT;
		return 1;
	}
	if (m->state == PIGEON_SERVER_DERIVE_SECRET && event == PIGEON_EVENT_ECDH_COMPLETE) {
		m->state = PIGEON_SERVER_SEND_ACK;
		return 1;
	}
	if (m->state == PIGEON_SERVER_SEND_ACK && event == PIGEON_EVENT_SIGNAL_CODE_DISPLAY) {
		m->state = PIGEON_SERVER_WAITING_FOR_CODE;
		return 1;
	}
	if (m->state == PIGEON_SERVER_VALIDATE_CODE && event == PIGEON_EVENT_CHECK_CODE && m->guards[PIGEON_GUARD_CODE_CORRECT] && m->guards[PIGEON_GUARD_CODE_CORRECT](m->userdata)) {
		m->state = PIGEON_SERVER_STORE_PAIRED;
		return 1;
	}
	if (m->state == PIGEON_SERVER_VALIDATE_CODE && event == PIGEON_EVENT_CHECK_CODE && m->guards[PIGEON_GUARD_CODE_WRONG] && m->guards[PIGEON_GUARD_CODE_WRONG](m->userdata)) {
		m->code_attempts = m->code_attempts + 1;
		if (m->on_change) m->on_change("code_attempts", m->userdata);
		m->state = PIGEON_SERVER_IDLE;
		return 1;
	}
	if (m->state == PIGEON_SERVER_STORE_PAIRED && event == PIGEON_EVENT_FINALISE) {
		if (m->actions[PIGEON_ACTION_STORE_DEVICE]) {
			int err = m->actions[PIGEON_ACTION_STORE_DEVICE](m->userdata);
			if (err) return -err;
		}
		m->device_secret = "dev_secret_1";
		if (m->on_change) m->on_change("device_secret", m->userdata);
		// paired_devices: paired_devices \union {"device_1"} (set by action)
		// active_tokens: active_tokens \ {current_token} (set by action)
		// used_tokens: used_tokens \union {current_token} (set by action)
		m->state = PIGEON_SERVER_PAIRED;
		return 1;
	}
	if (m->state == PIGEON_SERVER_AUTH_CHECK && event == PIGEON_EVENT_VERIFY && m->guards[PIGEON_GUARD_DEVICE_KNOWN] && m->guards[PIGEON_GUARD_DEVICE_KNOWN](m->userdata)) {
		if (m->actions[PIGEON_ACTION_VERIFY_DEVICE]) {
			int err = m->actions[PIGEON_ACTION_VERIFY_DEVICE](m->userdata);
			if (err) return -err;
		}
		// auth_nonces_used: auth_nonces_used \union {received_auth_nonce} (set by action)
		m->state = PIGEON_SERVER_SESSION_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_SERVER_AUTH_CHECK && event == PIGEON_EVENT_VERIFY && m->guards[PIGEON_GUARD_DEVICE_UNKNOWN] && m->guards[PIGEON_GUARD_DEVICE_UNKNOWN](m->userdata)) {
		m->state = PIGEON_SERVER_IDLE;
		return 1;
	}
	if (m->state == PIGEON_SERVER_SESSION_ACTIVE && event == PIGEON_EVENT_DISCONNECT) {
		m->state = PIGEON_SERVER_PAIRED;
		return 1;
	}
	return 0;
}

void pigeon_ios_machine_init(pigeon_ios_machine *m)
{
	memset(m, 0, sizeof(*m));
	m->state = PIGEON_APP_IDLE;
	m->received_server_pub = "none";
}

int pigeon_ios_handle_message(pigeon_ios_machine *m, pairing_ceremony_msg_type msg)
{
	if (m->state == PIGEON_APP_WAIT_ACK && msg == PIGEON_MSG_PAIR_HELLO_ACK) {
		if (m->actions[PIGEON_ACTION_DERIVE_SECRET]) {
			int err = m->actions[PIGEON_ACTION_DERIVE_SECRET](m->userdata);
			if (err) return -err;
		}
		// received_server_pub: recv_msg.pubkey (set by action)
		// client_shared_key: DeriveKey("client_pub", recv_msg.pubkey) (set by action)
		m->state = PIGEON_APP_E2E_READY;
		return 1;
	}
	if (m->state == PIGEON_APP_E2E_READY && msg == PIGEON_MSG_PAIR_CONFIRM) {
		// ios_code: DeriveCode(received_server_pub, "client_pub") (set by action)
		m->state = PIGEON_APP_SHOW_CODE;
		return 1;
	}
	if (m->state == PIGEON_APP_WAIT_PAIR_COMPLETE && msg == PIGEON_MSG_PAIR_COMPLETE) {
		if (m->actions[PIGEON_ACTION_STORE_SECRET]) {
			int err = m->actions[PIGEON_ACTION_STORE_SECRET](m->userdata);
			if (err) return -err;
		}
		m->state = PIGEON_APP_PAIRED;
		return 1;
	}
	if (m->state == PIGEON_APP_SEND_AUTH && msg == PIGEON_MSG_AUTH_OK) {
		m->state = PIGEON_APP_SESSION_ACTIVE;
		return 1;
	}
	return 0;
}

int pigeon_ios_step(pigeon_ios_machine *m, pairing_ceremony_event_id event)
{
	if (m->state == PIGEON_APP_IDLE && event == PIGEON_EVENT_USER_SCANS_QR) {
		m->state = PIGEON_APP_SCAN_QR;
		return 1;
	}
	if (m->state == PIGEON_APP_SCAN_QR && event == PIGEON_EVENT_QR_PARSED) {
		m->state = PIGEON_APP_CONNECT_RELAY;
		return 1;
	}
	if (m->state == PIGEON_APP_CONNECT_RELAY && event == PIGEON_EVENT_RELAY_CONNECTED) {
		m->state = PIGEON_APP_GEN_KEY_PAIR;
		return 1;
	}
	if (m->state == PIGEON_APP_GEN_KEY_PAIR && event == PIGEON_EVENT_KEY_PAIR_GENERATED) {
		if (m->actions[PIGEON_ACTION_SEND_PAIR_HELLO]) {
			int err = m->actions[PIGEON_ACTION_SEND_PAIR_HELLO](m->userdata);
			if (err) return -err;
		}
		m->state = PIGEON_APP_WAIT_ACK;
		return 1;
	}
	if (m->state == PIGEON_APP_SHOW_CODE && event == PIGEON_EVENT_CODE_DISPLAYED) {
		m->state = PIGEON_APP_WAIT_PAIR_COMPLETE;
		return 1;
	}
	if (m->state == PIGEON_APP_PAIRED && event == PIGEON_EVENT_APP_LAUNCH) {
		m->state = PIGEON_APP_RECONNECT;
		return 1;
	}
	if (m->state == PIGEON_APP_RECONNECT && event == PIGEON_EVENT_RELAY_CONNECTED) {
		m->state = PIGEON_APP_SEND_AUTH;
		return 1;
	}
	if (m->state == PIGEON_APP_SESSION_ACTIVE && event == PIGEON_EVENT_DISCONNECT) {
		m->state = PIGEON_APP_PAIRED;
		return 1;
	}
	return 0;
}

void pigeon_cli_machine_init(pigeon_cli_machine *m)
{
	memset(m, 0, sizeof(*m));
	m->state = PIGEON_CLI_IDLE;
}

int pigeon_cli_handle_message(pigeon_cli_machine *m, pairing_ceremony_msg_type msg)
{
	if (m->state == PIGEON_CLI_BEGIN_PAIR && msg == PIGEON_MSG_TOKEN_RESPONSE) {
		m->state = PIGEON_CLI_SHOW_QR;
		return 1;
	}
	if (m->state == PIGEON_CLI_SHOW_QR && msg == PIGEON_MSG_WAITING_FOR_CODE) {
		m->state = PIGEON_CLI_PROMPT_CODE;
		return 1;
	}
	if (m->state == PIGEON_CLI_SUBMIT_CODE && msg == PIGEON_MSG_PAIR_STATUS) {
		m->state = PIGEON_CLI_DONE;
		return 1;
	}
	return 0;
}

int pigeon_cli_step(pigeon_cli_machine *m, pairing_ceremony_event_id event)
{
	if (m->state == PIGEON_CLI_IDLE && event == PIGEON_EVENT_CLI_INIT) {
		m->state = PIGEON_CLI_GET_KEY;
		return 1;
	}
	if (m->state == PIGEON_CLI_GET_KEY && event == PIGEON_EVENT_KEY_STORED) {
		m->state = PIGEON_CLI_BEGIN_PAIR;
		return 1;
	}
	if (m->state == PIGEON_CLI_PROMPT_CODE && event == PIGEON_EVENT_USER_ENTERS_CODE) {
		m->state = PIGEON_CLI_SUBMIT_CODE;
		return 1;
	}
	return 0;
}

