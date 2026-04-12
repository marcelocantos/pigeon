// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Code generated from protocol/*.yaml. DO NOT EDIT.

#ifndef PIGEON_PAIRINGCEREMONY_GEN_H
#define PIGEON_PAIRINGCEREMONY_GEN_H

#include <stdbool.h>
#include <stdint.h>

// PairingCeremony server states.
typedef enum {
	PIGEON_SERVER_IDLE = 0,
	PIGEON_SERVER_GENERATE_TOKEN,
	PIGEON_SERVER_REGISTER_RELAY,
	PIGEON_SERVER_WAITING_FOR_CLIENT,
	PIGEON_SERVER_DERIVE_SECRET,
	PIGEON_SERVER_SEND_ACK,
	PIGEON_SERVER_WAITING_FOR_CODE,
	PIGEON_SERVER_VALIDATE_CODE,
	PIGEON_SERVER_STORE_PAIRED,
	PIGEON_SERVER_PAIRED,
	PIGEON_SERVER_AUTH_CHECK,
	PIGEON_SERVER_SESSION_ACTIVE,
	PIGEON_SERVER_STATE_COUNT
} pigeon_server_state;

// PairingCeremony ios states.
typedef enum {
	PIGEON_APP_IDLE = 0,
	PIGEON_APP_SCAN_QR,
	PIGEON_APP_CONNECT_RELAY,
	PIGEON_APP_GEN_KEY_PAIR,
	PIGEON_APP_WAIT_ACK,
	PIGEON_APP_E2E_READY,
	PIGEON_APP_SHOW_CODE,
	PIGEON_APP_WAIT_PAIR_COMPLETE,
	PIGEON_APP_PAIRED,
	PIGEON_APP_RECONNECT,
	PIGEON_APP_SEND_AUTH,
	PIGEON_APP_SESSION_ACTIVE,
	PIGEON_APP_STATE_COUNT
} pigeon_ios_state;

// PairingCeremony cli states.
typedef enum {
	PIGEON_CLI_IDLE = 0,
	PIGEON_CLI_GET_KEY,
	PIGEON_CLI_BEGIN_PAIR,
	PIGEON_CLI_SHOW_QR,
	PIGEON_CLI_PROMPT_CODE,
	PIGEON_CLI_SUBMIT_CODE,
	PIGEON_CLI_DONE,
	PIGEON_CLI_STATE_COUNT
} pigeon_cli_state;

// PairingCeremony message types.
typedef enum {
	PIGEON_MSG_PAIR_BEGIN = 0,
	PIGEON_MSG_TOKEN_RESPONSE,
	PIGEON_MSG_PAIR_HELLO,
	PIGEON_MSG_PAIR_HELLO_ACK,
	PIGEON_MSG_PAIR_CONFIRM,
	PIGEON_MSG_WAITING_FOR_CODE,
	PIGEON_MSG_CODE_SUBMIT,
	PIGEON_MSG_PAIR_COMPLETE,
	PIGEON_MSG_PAIR_STATUS,
	PIGEON_MSG_AUTH_REQUEST,
	PIGEON_MSG_AUTH_OK,
	PIGEON_MSG_COUNT
} pairing_ceremony_msg_type;

// PairingCeremony guards.
typedef enum {
	PIGEON_GUARD_TOKEN_VALID = 0,
	PIGEON_GUARD_TOKEN_INVALID,
	PIGEON_GUARD_CODE_CORRECT,
	PIGEON_GUARD_CODE_WRONG,
	PIGEON_GUARD_DEVICE_KNOWN,
	PIGEON_GUARD_DEVICE_UNKNOWN,
	PIGEON_GUARD_NONCE_FRESH,
	PIGEON_GUARD_COUNT
} pairing_ceremony_guard_id;

// PairingCeremony actions.
typedef enum {
	PIGEON_ACTION_GENERATE_TOKEN = 0,
	PIGEON_ACTION_REGISTER_RELAY,
	PIGEON_ACTION_DERIVE_SECRET,
	PIGEON_ACTION_STORE_DEVICE,
	PIGEON_ACTION_VERIFY_DEVICE,
	PIGEON_ACTION_SEND_PAIR_HELLO,
	PIGEON_ACTION_STORE_SECRET,
	PIGEON_ACTION_COUNT
} pairing_ceremony_action_id;

// PairingCeremony events.
typedef enum {
	PIGEON_EVENT_TOKEN_CREATED = 0,
	PIGEON_EVENT_RELAY_REGISTERED,
	PIGEON_EVENT_ECDH_COMPLETE,
	PIGEON_EVENT_SIGNAL_CODE_DISPLAY,
	PIGEON_EVENT_CHECK_CODE,
	PIGEON_EVENT_FINALISE,
	PIGEON_EVENT_VERIFY,
	PIGEON_EVENT_DISCONNECT,
	PIGEON_EVENT_USER_SCANS_QR,
	PIGEON_EVENT_QR_PARSED,
	PIGEON_EVENT_RELAY_CONNECTED,
	PIGEON_EVENT_KEY_PAIR_GENERATED,
	PIGEON_EVENT_CODE_DISPLAYED,
	PIGEON_EVENT_APP_LAUNCH,
	PIGEON_EVENT_CLI_INIT,
	PIGEON_EVENT_KEY_STORED,
	PIGEON_EVENT_USER_ENTERS_CODE,
	PIGEON_EVENT_RECV_PAIR_BEGIN,
	PIGEON_EVENT_RECV_PAIR_HELLO,
	PIGEON_EVENT_RECV_CODE_SUBMIT,
	PIGEON_EVENT_RECV_AUTH_REQUEST,
	PIGEON_EVENT_RECV_PAIR_HELLO_ACK,
	PIGEON_EVENT_RECV_PAIR_CONFIRM,
	PIGEON_EVENT_RECV_PAIR_COMPLETE,
	PIGEON_EVENT_RECV_AUTH_OK,
	PIGEON_EVENT_RECV_TOKEN_RESPONSE,
	PIGEON_EVENT_RECV_WAITING_FOR_CODE,
	PIGEON_EVENT_RECV_PAIR_STATUS,
	PIGEON_EVENT_COUNT
} pairing_ceremony_event_id;

// Guard and action callback types.
typedef bool (*pigeon_guard_fn)(void *ctx);
typedef int  (*pigeon_action_fn)(void *ctx);
typedef void (*pigeon_change_fn)(const char *var_name, void *ctx);

// PairingCeremony server state machine.
typedef struct {
	pigeon_server_state state;
	const char * current_token; // pairing token currently in play
	const char * active_tokens; // set of valid (non-revoked) tokens
	const char * used_tokens; // set of revoked tokens
	const char * server_ecdh_pub; // server ECDH public key
	const char * received_client_pub; // pubkey server received in pair_hello (may be adversary's)
	const char * server_shared_key; // ECDH key derived by server (tuple to match DeriveKey output type)
	const char * server_code; // code computed by server from its view of the pubkeys (tuple to match DeriveCode output type)
	const char * received_code; // code received in code_submit (tuple to match DeriveCode output type)
	int code_attempts; // failed code submission attempts
	const char * device_secret; // persistent device secret
	const char * paired_devices; // device IDs that completed pairing
	const char * received_device_id; // device_id from auth_request
	const char * auth_nonces_used; // set of consumed auth nonces
	const char * received_auth_nonce; // nonce from auth_request
	pigeon_guard_fn guards[PIGEON_GUARD_COUNT];
	pigeon_action_fn actions[PIGEON_ACTION_COUNT];
	pigeon_change_fn on_change;
	void *userdata;
} pigeon_server_machine;

void pigeon_server_machine_init(pigeon_server_machine *m);
int  pigeon_server_handle_message(pigeon_server_machine *m, pairing_ceremony_msg_type msg);
int  pigeon_server_step(pigeon_server_machine *m, pairing_ceremony_event_id event);

// PairingCeremony ios state machine.
typedef struct {
	pigeon_ios_state state;
	const char * received_server_pub; // pubkey ios received in pair_hello_ack (may be adversary's)
	const char * client_shared_key; // ECDH key derived by ios (tuple to match DeriveKey output type)
	const char * ios_code; // code computed by ios from its view of the pubkeys (tuple to match DeriveCode output type)
	pigeon_guard_fn guards[PIGEON_GUARD_COUNT];
	pigeon_action_fn actions[PIGEON_ACTION_COUNT];
	pigeon_change_fn on_change;
	void *userdata;
} pigeon_ios_machine;

void pigeon_ios_machine_init(pigeon_ios_machine *m);
int  pigeon_ios_handle_message(pigeon_ios_machine *m, pairing_ceremony_msg_type msg);
int  pigeon_ios_step(pigeon_ios_machine *m, pairing_ceremony_event_id event);

// PairingCeremony cli state machine.
typedef struct {
	pigeon_cli_state state;
	pigeon_guard_fn guards[PIGEON_GUARD_COUNT];
	pigeon_action_fn actions[PIGEON_ACTION_COUNT];
	pigeon_change_fn on_change;
	void *userdata;
} pigeon_cli_machine;

void pigeon_cli_machine_init(pigeon_cli_machine *m);
int  pigeon_cli_handle_message(pigeon_cli_machine *m, pairing_ceremony_msg_type msg);
int  pigeon_cli_step(pigeon_cli_machine *m, pairing_ceremony_event_id event);

#endif // PIGEON_PAIRINGCEREMONY_GEN_H
