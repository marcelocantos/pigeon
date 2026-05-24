// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Code generated from protocol/*.yaml. DO NOT EDIT.

#ifndef PIGEON_SESSION_GEN_H
#define PIGEON_SESSION_GEN_H

#include <stdbool.h>
#include <stdint.h>

// Session backend states.
typedef enum {
	PIGEON_BACKEND_IDLE = 0,
	PIGEON_BACKEND_GENERATE_TOKEN,
	PIGEON_BACKEND_REGISTER_RELAY,
	PIGEON_BACKEND_WAITING_FOR_CLIENT,
	PIGEON_BACKEND_DERIVE_SECRET,
	PIGEON_BACKEND_SEND_ACK,
	PIGEON_BACKEND_WAITING_FOR_CODE,
	PIGEON_BACKEND_VALIDATE_CODE,
	PIGEON_BACKEND_STORE_PAIRED,
	PIGEON_BACKEND_PAIRED,
	PIGEON_BACKEND_AUTH_CHECK,
	PIGEON_BACKEND_SESSION_ACTIVE,
	PIGEON_BACKEND_RELAY_CONNECTED,
	PIGEON_BACKEND_CANDIDATES_ADVERTISED,
	PIGEON_BACKEND_ALT_ACTIVE,
	PIGEON_BACKEND_RELAY_BACKOFF,
	PIGEON_BACKEND_ALT_DEGRADED,
	PIGEON_BACKEND_STATE_COUNT
} pigeon_backend_state;

// Session client states.
typedef enum {
	PIGEON_CLIENT_IDLE = 0,
	PIGEON_CLIENT_OBTAIN_BACKCHANNEL_SECRET,
	PIGEON_CLIENT_CONNECT_RELAY,
	PIGEON_CLIENT_GEN_KEY_PAIR,
	PIGEON_CLIENT_WAIT_ACK,
	PIGEON_CLIENT_E2E_READY,
	PIGEON_CLIENT_SHOW_CODE,
	PIGEON_CLIENT_WAIT_PAIR_COMPLETE,
	PIGEON_CLIENT_PAIRED,
	PIGEON_CLIENT_RECONNECT,
	PIGEON_CLIENT_SEND_AUTH,
	PIGEON_CLIENT_SESSION_ACTIVE,
	PIGEON_CLIENT_RELAY_CONNECTED,
	PIGEON_CLIENT_PAIR_DIALING,
	PIGEON_CLIENT_PAIR_CHECKING,
	PIGEON_CLIENT_ALT_ACTIVE,
	PIGEON_CLIENT_RELAY_FALLBACK,
	PIGEON_CLIENT_STATE_COUNT
} pigeon_client_state;

// Session relay states.
typedef enum {
	PIGEON_RELAY_IDLE = 0,
	PIGEON_RELAY_BACKEND_REGISTERED,
	PIGEON_RELAY_BRIDGED,
	PIGEON_RELAY_STATE_COUNT
} pigeon_relay_state;

// Session message types.
typedef enum {
	PIGEON_SESSION_MSG_PAIR_HELLO = 0,
	PIGEON_SESSION_MSG_PAIR_HELLO_ACK,
	PIGEON_SESSION_MSG_PAIR_CONFIRM,
	PIGEON_SESSION_MSG_PAIR_COMPLETE,
	PIGEON_SESSION_MSG_AUTH_REQUEST,
	PIGEON_SESSION_MSG_AUTH_OK,
	PIGEON_SESSION_MSG_CANDIDATES,
	PIGEON_SESSION_MSG_PAIR_CHECK,
	PIGEON_SESSION_MSG_PAIR_CHECK_ACK,
	PIGEON_SESSION_MSG_PATH_PING,
	PIGEON_SESSION_MSG_PATH_PONG,
	PIGEON_SESSION_MSG_COUNT
} session_msg_type;

// Session guards.
typedef enum {
	PIGEON_SESSION_GUARD_TOKEN_VALID = 0,
	PIGEON_SESSION_GUARD_TOKEN_INVALID,
	PIGEON_SESSION_GUARD_CODE_CORRECT,
	PIGEON_SESSION_GUARD_CODE_WRONG,
	PIGEON_SESSION_GUARD_DEVICE_KNOWN,
	PIGEON_SESSION_GUARD_DEVICE_UNKNOWN,
	PIGEON_SESSION_GUARD_NONCE_FRESH,
	PIGEON_SESSION_GUARD_CHALLENGE_VALID,
	PIGEON_SESSION_GUARD_CHALLENGE_INVALID,
	PIGEON_SESSION_GUARD_ALT_ENABLED,
	PIGEON_SESSION_GUARD_ALT_DISABLED,
	PIGEON_SESSION_GUARD_LOCAL_CANDIDATES_AVAILABLE,
	PIGEON_SESSION_GUARD_UNDER_MAX_FAILURES,
	PIGEON_SESSION_GUARD_AT_MAX_FAILURES,
	PIGEON_SESSION_GUARD_COUNT
} session_guard_id;

// Session actions.
typedef enum {
	PIGEON_SESSION_ACTION_GENERATE_TOKEN = 0,
	PIGEON_SESSION_ACTION_REGISTER_RELAY,
	PIGEON_SESSION_ACTION_DERIVE_SECRET,
	PIGEON_SESSION_ACTION_STORE_DEVICE,
	PIGEON_SESSION_ACTION_VERIFY_DEVICE,
	PIGEON_SESSION_ACTION_ACTIVATE_LAN,
	PIGEON_SESSION_ACTION_FALLBACK_TO_RELAY,
	PIGEON_SESSION_ACTION_RESET_FAILURES,
	PIGEON_SESSION_ACTION_SEND_PAIR_HELLO,
	PIGEON_SESSION_ACTION_STORE_SECRET,
	PIGEON_SESSION_ACTION_DIAL_CANDIDATE,
	PIGEON_SESSION_ACTION_BRIDGE_STREAMS,
	PIGEON_SESSION_ACTION_UNBRIDGE,
	PIGEON_SESSION_ACTION_COUNT
} session_action_id;

// Session events.
typedef enum {
	PIGEON_SESSION_EVENT_APP_SEND = 0,
	PIGEON_SESSION_EVENT_APP_RECV,
	PIGEON_SESSION_EVENT_APP_SEND_DATAGRAM,
	PIGEON_SESSION_EVENT_APP_RECV_DATAGRAM,
	PIGEON_SESSION_EVENT_APP_CLOSE,
	PIGEON_SESSION_EVENT_APP_FORCE_FALLBACK,
	PIGEON_SESSION_EVENT_RELAY_STREAM_DATA,
	PIGEON_SESSION_EVENT_RELAY_STREAM_ERROR,
	PIGEON_SESSION_EVENT_RELAY_DATAGRAM,
	PIGEON_SESSION_EVENT_ALT_STREAM_DATA,
	PIGEON_SESSION_EVENT_ALT_STREAM_ERROR,
	PIGEON_SESSION_EVENT_ALT_DATAGRAM,
	PIGEON_SESSION_EVENT_DIAL_OK,
	PIGEON_SESSION_EVENT_DIAL_FAILED,
	PIGEON_SESSION_EVENT_PAIR_CHECK_OK,
	PIGEON_SESSION_EVENT_PING_TIMEOUT,
	PIGEON_SESSION_EVENT_PING_TICK,
	PIGEON_SESSION_EVENT_BACKOFF_EXPIRED,
	PIGEON_SESSION_EVENT_CANDIDATES_TIMEOUT,
	PIGEON_SESSION_EVENT_CLI_INIT_PAIR,
	PIGEON_SESSION_EVENT_TOKEN_CREATED,
	PIGEON_SESSION_EVENT_RELAY_REGISTERED,
	PIGEON_SESSION_EVENT_ECDH_COMPLETE,
	PIGEON_SESSION_EVENT_SIGNAL_CODE_DISPLAY,
	PIGEON_SESSION_EVENT_CLI_CODE_ENTERED,
	PIGEON_SESSION_EVENT_CHECK_CODE,
	PIGEON_SESSION_EVENT_FINALISE,
	PIGEON_SESSION_EVENT_VERIFY,
	PIGEON_SESSION_EVENT_SESSION_ESTABLISHED,
	PIGEON_SESSION_EVENT_CANDIDATES_GATHERED,
	PIGEON_SESSION_EVENT_CANDIDATES_CHANGED,
	PIGEON_SESSION_EVENT_CANDIDATES_REFRESH_TICK,
	PIGEON_SESSION_EVENT_DISCONNECT,
	PIGEON_SESSION_EVENT_BACKCHANNEL_RECEIVED,
	PIGEON_SESSION_EVENT_SECRET_PARSED,
	PIGEON_SESSION_EVENT_RELAY_CONNECTED,
	PIGEON_SESSION_EVENT_KEY_PAIR_GENERATED,
	PIGEON_SESSION_EVENT_CODE_DISPLAYED,
	PIGEON_SESSION_EVENT_APP_LAUNCH,
	PIGEON_SESSION_EVENT_VERIFY_TIMEOUT,
	PIGEON_SESSION_EVENT_ALT_ERROR,
	PIGEON_SESSION_EVENT_RELAY_OK,
	PIGEON_SESSION_EVENT_BACKEND_REGISTER,
	PIGEON_SESSION_EVENT_CLIENT_CONNECT,
	PIGEON_SESSION_EVENT_CLIENT_DISCONNECT,
	PIGEON_SESSION_EVENT_BACKEND_DISCONNECT,
	PIGEON_SESSION_EVENT_RECV_PAIR_HELLO,
	PIGEON_SESSION_EVENT_RECV_AUTH_REQUEST,
	PIGEON_SESSION_EVENT_RECV_PAIR_CHECK,
	PIGEON_SESSION_EVENT_RECV_PATH_PONG,
	PIGEON_SESSION_EVENT_RECV_PAIR_HELLO_ACK,
	PIGEON_SESSION_EVENT_RECV_PAIR_CONFIRM,
	PIGEON_SESSION_EVENT_RECV_PAIR_COMPLETE,
	PIGEON_SESSION_EVENT_RECV_AUTH_OK,
	PIGEON_SESSION_EVENT_RECV_CANDIDATES,
	PIGEON_SESSION_EVENT_RECV_PAIR_CHECK_ACK,
	PIGEON_SESSION_EVENT_RECV_PATH_PING,
	PIGEON_SESSION_EVENT_COUNT
} session_event_id;

// Session commands.
typedef enum {
	PIGEON_SESSION_CMD_WRITE_ACTIVE_STREAM = 0,
	PIGEON_SESSION_CMD_SEND_ACTIVE_DATAGRAM,
	PIGEON_SESSION_CMD_SEND_PATH_PING,
	PIGEON_SESSION_CMD_SEND_PATH_PONG,
	PIGEON_SESSION_CMD_SEND_CANDIDATES,
	PIGEON_SESSION_CMD_SEND_PAIR_CHECK,
	PIGEON_SESSION_CMD_SEND_PAIR_CHECK_ACK,
	PIGEON_SESSION_CMD_DIAL_CANDIDATE,
	PIGEON_SESSION_CMD_DELIVER_RECV,
	PIGEON_SESSION_CMD_DELIVER_RECV_ERROR,
	PIGEON_SESSION_CMD_DELIVER_RECV_DATAGRAM,
	PIGEON_SESSION_CMD_START_ALT_STREAM_READER,
	PIGEON_SESSION_CMD_STOP_ALT_STREAM_READER,
	PIGEON_SESSION_CMD_START_ALT_DG_READER,
	PIGEON_SESSION_CMD_STOP_ALT_DG_READER,
	PIGEON_SESSION_CMD_START_MONITOR,
	PIGEON_SESSION_CMD_STOP_MONITOR,
	PIGEON_SESSION_CMD_START_PONG_TIMEOUT,
	PIGEON_SESSION_CMD_CANCEL_PONG_TIMEOUT,
	PIGEON_SESSION_CMD_START_BACKOFF_TIMER,
	PIGEON_SESSION_CMD_CLOSE_ALT_PATH,
	PIGEON_SESSION_CMD_SIGNAL_ALT_READY,
	PIGEON_SESSION_CMD_RESET_ALT_READY,
	PIGEON_SESSION_CMD_SET_CRYPTO_DATAGRAM,
	PIGEON_SESSION_CMD_COUNT
} session_cmd_id;

// Wire constants.
#define PIGEON_WIRE_DG_CONN_WHOLE ((uint8_t)0x00) // conn-level single-frame datagram
#define PIGEON_WIRE_DG_PING ((uint8_t)0x10) // health ping on direct path
#define PIGEON_WIRE_DG_PONG ((uint8_t)0x11) // health pong on direct path
#define PIGEON_WIRE_DG_CONN_FRAGMENT ((uint8_t)0x40) // conn-level multi-frame datagram
#define PIGEON_WIRE_DG_CHAN_WHOLE ((uint8_t)0x80) // channel single-frame datagram
#define PIGEON_WIRE_DG_CHAN_FRAGMENT ((uint8_t)0xC0) // channel multi-frame datagram
#define PIGEON_WIRE_FRAG_HEADER_SIZE 8 // fragment header: msgID(4) + fragIdx(2) + totalFrags(2)
#define PIGEON_WIRE_CHAN_ID_SIZE 2 // channel ID prefix size in bytes
#define PIGEON_WIRE_MAX_DATAGRAM_PAYLOAD 1200 // max payload per QUIC datagram (bytes)
#define PIGEON_WIRE_FRAGMENT_TIMEOUT_MS 5000 /* ms */ // fragment reassembly timeout
#define PIGEON_WIRE_FRAME_APP ((uint8_t)0x00) // application data
#define PIGEON_WIRE_FRAME_CANDIDATES ((uint8_t)0x01) // candidate set advertisement (relay control channel)
#define PIGEON_WIRE_FRAME_CUTOVER ((uint8_t)0x02) // transport cutover marker
#define PIGEON_WIRE_FRAME_PAIR_CHECK ((uint8_t)0x03) // candidate-pair connectivity-check challenge (on candidate pipe)
#define PIGEON_WIRE_FRAME_PAIR_CHECK_ACK ((uint8_t)0x04) // candidate-pair connectivity-check ack / implicit nomination (on candidate pipe)
#define PIGEON_WIRE_CAND_HOST "host" // host candidate (local interface address; what LAN-direct uses)
#define PIGEON_WIRE_CAND_SRFLX "srflx" // server-reflexive candidate (STUN-derived public address; Phase 1 STUN target)
#define PIGEON_WIRE_MAX_MESSAGE_SIZE 1048576 // max stream message size (1 MiB)
#define PIGEON_WIRE_LENGTH_PREFIX_SIZE 4 // big-endian length prefix size
#define PIGEON_WIRE_PING_INTERVAL_MS 5000 /* ms */ // health ping interval
#define PIGEON_WIRE_PONG_TIMEOUT_MS 4000 /* ms */ // pong reply timeout
#define PIGEON_WIRE_MAX_PING_FAILURES 3 // consecutive failures before fallback
#define PIGEON_WIRE_MAX_BACKOFF_LEVEL 5 // exponential backoff cap
#define PIGEON_WIRE_STREAM_CHANNEL_OPENER_SUFFIX ":o2a" // HKDF info suffix for opener→acceptor stream key
#define PIGEON_WIRE_STREAM_CHANNEL_ACCEPT_SUFFIX ":a2o" // HKDF info suffix for acceptor→opener stream key
#define PIGEON_WIRE_DG_CHANNEL_SEND_SUFFIX ":dg:send" // HKDF info suffix for datagram send key
#define PIGEON_WIRE_DG_CHANNEL_RECV_SUFFIX ":dg:recv" // HKDF info suffix for datagram recv key
#define PIGEON_WIRE_CHANNEL_ID_HASH_MULTIPLIER 31 // hash multiplier for channel name → uint16 ID

// Guard and action callback types.
typedef bool (*pigeon_guard_fn)(void *ctx);
typedef int  (*pigeon_action_fn)(void *ctx);
typedef void (*pigeon_change_fn)(const char *var_name, void *ctx);

// Session backend state machine.
typedef struct {
	pigeon_backend_state state;
	const char * current_token; // pairing token currently in play
	const char * backend_ecdh_pub; // backend ECDH public key
	const char * received_client_pub; // pubkey backend received in pair_hello
	const char * backend_shared_key; // ECDH key derived by backend
	const char * backend_code; // code computed by backend
	const char * received_code; // code entered via CLI
	int code_attempts; // failed code submission attempts
	const char * device_secret; // persistent device secret
	const char * received_device_id; // device_id from auth_request
	const char * received_auth_nonce; // nonce from auth_request
	bool secret_published; // whether token has been published via backchannel
	int ping_failures; // consecutive failed pings
	int backoff_level; // exponential backoff level
	const char * b_active_path; // backend active path
	const char * b_dispatcher_path; // backend datagram dispatcher binding
	const char * monitor_target; // health monitor target
	const char * alt_signal; // AltReady notification state
	pigeon_guard_fn guards[PIGEON_SESSION_GUARD_COUNT];
	pigeon_action_fn actions[PIGEON_SESSION_ACTION_COUNT];
	pigeon_change_fn on_change;
	void *userdata;
} pigeon_backend_machine;

void pigeon_backend_machine_init(pigeon_backend_machine *m);
int  pigeon_backend_handle_message(pigeon_backend_machine *m, session_msg_type msg);
int  pigeon_backend_step(pigeon_backend_machine *m, session_event_id event);

// Session client state machine.
typedef struct {
	pigeon_client_state state;
	const char * received_backend_pub; // pubkey client received in pair_hello_ack
	const char * client_shared_key; // ECDH key derived by client
	const char * client_code; // code computed by client
	const char * c_active_path; // client active path
	const char * c_dispatcher_path; // client datagram dispatcher binding
	const char * alt_signal; // AltReady notification state
	pigeon_guard_fn guards[PIGEON_SESSION_GUARD_COUNT];
	pigeon_action_fn actions[PIGEON_SESSION_ACTION_COUNT];
	pigeon_change_fn on_change;
	void *userdata;
} pigeon_client_machine;

void pigeon_client_machine_init(pigeon_client_machine *m);
int  pigeon_client_handle_message(pigeon_client_machine *m, session_msg_type msg);
int  pigeon_client_step(pigeon_client_machine *m, session_event_id event);

// Session relay state machine.
typedef struct {
	pigeon_relay_state state;
	const char * relay_bridge; // relay bridge state
	pigeon_guard_fn guards[PIGEON_SESSION_GUARD_COUNT];
	pigeon_action_fn actions[PIGEON_SESSION_ACTION_COUNT];
	pigeon_change_fn on_change;
	void *userdata;
} pigeon_relay_machine;

void pigeon_relay_machine_init(pigeon_relay_machine *m);
int  pigeon_relay_handle_message(pigeon_relay_machine *m, session_msg_type msg);
int  pigeon_relay_step(pigeon_relay_machine *m, session_event_id event);

#endif // PIGEON_SESSION_GEN_H
