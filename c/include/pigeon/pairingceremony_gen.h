// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Code generated from protocol/*.yaml. DO NOT EDIT.

#ifndef PIGEON_PAIRINGCEREMONY_GEN_H
#define PIGEON_PAIRINGCEREMONY_GEN_H

#include <stdbool.h>
#include <stdint.h>

// PairingCeremony acceptor states.
typedef enum {
	PIGEON_ACCEPTOR_IDLE = 0,
	PIGEON_ACCEPTOR_GENERATING_EPHEMERAL,
	PIGEON_ACCEPTOR_REGISTERING_RELAY,
	PIGEON_ACCEPTOR_WAITING_FOR_HELLO,
	PIGEON_ACCEPTOR_DERIVING_CODE,
	PIGEON_ACCEPTOR_AWAITING_USER_CONFIRM,
	PIGEON_ACCEPTOR_AWAITING_PEER_CONFIRM,
	PIGEON_ACCEPTOR_PAIRED,
	PIGEON_ACCEPTOR_ABORTED,
	PIGEON_ACCEPTOR_STATE_COUNT
} pigeon_acceptor_state;

// PairingCeremony initiator states.
typedef enum {
	PIGEON_INITIATOR_IDLE = 0,
	PIGEON_INITIATOR_DECODING_TOKEN,
	PIGEON_INITIATOR_GENERATING_EPHEMERAL,
	PIGEON_INITIATOR_CONNECTING_RELAY,
	PIGEON_INITIATOR_AWAITING_WELCOME,
	PIGEON_INITIATOR_DERIVING_CODE,
	PIGEON_INITIATOR_AWAITING_USER_CONFIRM,
	PIGEON_INITIATOR_AWAITING_PEER_CONFIRM,
	PIGEON_INITIATOR_PAIRED,
	PIGEON_INITIATOR_ABORTED,
	PIGEON_INITIATOR_STATE_COUNT
} pigeon_initiator_state;

// PairingCeremony message types.
typedef enum {
	PIGEON_MSG_HELLO = 0,
	PIGEON_MSG_WELCOME,
	PIGEON_MSG_CONFIRM_TO_INITIATOR,
	PIGEON_MSG_CONFIRM_TO_ACCEPTOR,
	PIGEON_MSG_COUNT
} pairing_ceremony_msg_type;

// PairingCeremony actions.
typedef enum {
	PIGEON_ACTION_GEN_EPHEMERAL = 0,
	PIGEON_ACTION_REGISTER_RELAY,
	PIGEON_ACTION_EMIT_TOKEN,
	PIGEON_ACTION_DERIVE_CODE,
	PIGEON_ACTION_STORE_RECORD,
	PIGEON_ACTION_DECODE_TOKEN,
	PIGEON_ACTION_DIAL_RELAY,
	PIGEON_ACTION_COUNT
} pairing_ceremony_action_id;

// PairingCeremony events.
typedef enum {
	PIGEON_EVENT_PAIR_BEGIN = 0,
	PIGEON_EVENT_EPHEMERAL_READY,
	PIGEON_EVENT_RELAY_REGISTERED,
	PIGEON_EVENT_CODE_READY,
	PIGEON_EVENT_USER_CONFIRM,
	PIGEON_EVENT_USER_CANCEL,
	PIGEON_EVENT_TOKEN_RECEIVED,
	PIGEON_EVENT_TOKEN_DECODED,
	PIGEON_EVENT_RELAY_CONNECTED,
	PIGEON_EVENT_RECV_HELLO,
	PIGEON_EVENT_RECV_CONFIRM_TO_ACCEPTOR,
	PIGEON_EVENT_RECV_WELCOME,
	PIGEON_EVENT_RECV_CONFIRM_TO_INITIATOR,
	PIGEON_EVENT_COUNT
} pairing_ceremony_event_id;

// Guard and action callback types.
typedef bool (*pigeon_guard_fn)(void *ctx);
typedef int  (*pigeon_action_fn)(void *ctx);
typedef void (*pigeon_change_fn)(const char *var_name, void *ctx);

// PairingCeremony acceptor state machine.
typedef struct {
	pigeon_acceptor_state state;
	const char * acceptor_eph_pub; // acceptor's ephemeral X25519 public key
	const char * acceptor_received_eph_pub; // ephemeral pubkey acceptor saw in hello (may be adversary's)
	const char * acceptor_received_identity; // identity pubkey acceptor saw in hello
	const char * acceptor_received_instance; // instance ID acceptor saw in hello
	const char * acceptor_code; // confirmation code acceptor derived from its (ephA, ephB) view
	const char * acceptor_user_confirmed; // has the acceptor's local human pressed y?
	const char * acceptor_received_confirm; // has the acceptor received initiator's confirm message?
	pigeon_action_fn actions[PIGEON_ACTION_COUNT];
	pigeon_change_fn on_change;
	void *userdata;
} pigeon_acceptor_machine;

void pigeon_acceptor_machine_init(pigeon_acceptor_machine *m);
int  pigeon_acceptor_handle_message(pigeon_acceptor_machine *m, pairing_ceremony_msg_type msg);
int  pigeon_acceptor_step(pigeon_acceptor_machine *m, pairing_ceremony_event_id event);

// PairingCeremony initiator state machine.
typedef struct {
	pigeon_initiator_state state;
	const char * initiator_eph_pub; // initiator's ephemeral X25519 public key
	const char * received_acceptor_eph_pub; // acceptor ephemeral pubkey from token (trusted, out-of-band)
	const char * received_acceptor_identity; // acceptor identity pubkey from token
	const char * received_acceptor_instance; // acceptor instance ID from token
	const char * initiator_received_eph_pub; // ephemeral pubkey initiator saw in welcome (may be adversary's)
	const char * initiator_received_identity; // identity pubkey initiator saw in welcome
	const char * initiator_received_instance; // instance ID initiator saw in welcome
	const char * initiator_code; // confirmation code initiator derived from its (ephA, ephB) view
	const char * initiator_user_confirmed; // has the initiator's local human pressed y?
	const char * initiator_received_confirm; // has the initiator received acceptor's confirm message?
	pigeon_action_fn actions[PIGEON_ACTION_COUNT];
	pigeon_change_fn on_change;
	void *userdata;
} pigeon_initiator_machine;

void pigeon_initiator_machine_init(pigeon_initiator_machine *m);
int  pigeon_initiator_handle_message(pigeon_initiator_machine *m, pairing_ceremony_msg_type msg);
int  pigeon_initiator_step(pigeon_initiator_machine *m, pairing_ceremony_event_id event);

#endif // PIGEON_PAIRINGCEREMONY_GEN_H
