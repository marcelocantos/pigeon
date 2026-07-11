// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Code generated from protocol/*.yaml. DO NOT EDIT.

#include "pairingceremony_gen.h"
#include <string.h>

void pigeon_acceptor_machine_init(pigeon_acceptor_machine *m)
{
	memset(m, 0, sizeof(*m));
	m->state = PIGEON_ACCEPTOR_IDLE;
	m->acceptor_eph_pub = "none";
	m->acceptor_received_commit = "none";
	m->acceptor_received_eph_pub = "none";
	m->acceptor_received_identity = "none";
	m->acceptor_received_instance = "none";
	m->acceptor_commit_ok = "false";
	m->acceptor_user_confirmed = "false";
	m->acceptor_received_confirm = "false";
}

int pigeon_acceptor_handle_message(pigeon_acceptor_machine *m, pairing_ceremony_msg_type msg)
{
	if (m->state == PIGEON_ACCEPTOR_WAITING_FOR_HELLO && msg == PIGEON_PAIRINGCEREMONY_MSG_HELLO) {
		if (m->actions[PIGEON_PAIRINGCEREMONY_ACTION_STORE_COMMIT]) {
			int err = m->actions[PIGEON_PAIRINGCEREMONY_ACTION_STORE_COMMIT](m->userdata);
			if (err) return -err;
		}
		// acceptor_received_commit: recv_msg.commit (set by action)
		// acceptor_received_identity: recv_msg.identity_pub (set by action)
		// acceptor_received_instance: recv_msg.instance_id (set by action)
		m->state = PIGEON_ACCEPTOR_WAITING_FOR_REVEAL;
		return 1;
	}
	if (m->state == PIGEON_ACCEPTOR_WAITING_FOR_REVEAL && msg == PIGEON_PAIRINGCEREMONY_MSG_REVEAL) {
		if (m->actions[PIGEON_PAIRINGCEREMONY_ACTION_VERIFY_COMMIT_AND_DERIVE]) {
			int err = m->actions[PIGEON_PAIRINGCEREMONY_ACTION_VERIFY_COMMIT_AND_DERIVE](m->userdata);
			if (err) return -err;
		}
		// acceptor_received_eph_pub: recv_msg.eph_pub (set by action)
		// acceptor_commit_ok: IF CommitMatches(acceptor_received_commit, recv_msg.eph_pub) THEN "true" ELSE "false" (set by action)
		// acceptor_code: IF CommitMatches(acceptor_received_commit, recv_msg.eph_pub) THEN DeriveCode(acceptor_eph_pub, recv_msg.eph_pub) ELSE <<"none">> (set by action)
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
	if (m->state == PIGEON_ACCEPTOR_WAITING_FOR_REVEAL && event == PIGEON_PAIRINGCEREMONY_EVENT_COMMIT_FAIL) {
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
	m->initiator_commit = "none";
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
		if (m->actions[PIGEON_PAIRINGCEREMONY_ACTION_SEND_REVEAL]) {
			int err = m->actions[PIGEON_PAIRINGCEREMONY_ACTION_SEND_REVEAL](m->userdata);
			if (err) return -err;
		}
		// initiator_received_eph_pub: recv_msg.eph_pub (set by action)
		// initiator_received_identity: recv_msg.identity_pub (set by action)
		// initiator_received_instance: recv_msg.instance_id (set by action)
		m->state = PIGEON_INITIATOR_REVEALING;
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
		m->initiator_commit = "initiator_commit";
		if (m->on_change) m->on_change("initiator_commit", m->userdata);
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
	if (m->state == PIGEON_INITIATOR_REVEALING && event == PIGEON_PAIRINGCEREMONY_EVENT_REVEAL_SENT) {
		if (m->actions[PIGEON_PAIRINGCEREMONY_ACTION_DERIVE_CODE]) {
			int err = m->actions[PIGEON_PAIRINGCEREMONY_ACTION_DERIVE_CODE](m->userdata);
			if (err) return -err;
		}
		// initiator_code: DeriveCode(initiator_eph_pub, initiator_received_eph_pub) (set by action)
		m->state = PIGEON_INITIATOR_DERIVING_CODE;
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

