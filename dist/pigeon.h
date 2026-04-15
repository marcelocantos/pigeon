// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

#ifndef PIGEON_H
#define PIGEON_H

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

// PairingCeremony server/pairing states.
typedef enum {
	PIGEON_SERVER_PAIRING_IDLE = 0,
	PIGEON_SERVER_PAIRING_GENERATE_TOKEN,
	PIGEON_SERVER_PAIRING_REGISTER_RELAY,
	PIGEON_SERVER_PAIRING_WAITING_FOR_CLIENT,
	PIGEON_SERVER_PAIRING_DERIVE_SECRET,
	PIGEON_SERVER_PAIRING_SEND_ACK,
	PIGEON_SERVER_PAIRING_WAITING_FOR_CODE,
	PIGEON_SERVER_PAIRING_VALIDATE_CODE,
	PIGEON_SERVER_PAIRING_STORE_PAIRED,
	PIGEON_SERVER_PAIRING_PAIRING_COMPLETE,
	PIGEON_SERVER_PAIRING_STATE_COUNT
} pigeon_server_pairing_state;

// PairingCeremony server/auth states.
typedef enum {
	PIGEON_SERVER_AUTH_IDLE = 0,
	PIGEON_SERVER_AUTH_PAIRED,
	PIGEON_SERVER_AUTH_AUTH_CHECK,
	PIGEON_SERVER_AUTH_SESSION_ACTIVE,
	PIGEON_SERVER_AUTH_STATE_COUNT
} pigeon_server_auth_state;

// PairingCeremony ios/pairing states.
typedef enum {
	PIGEON_APP_PAIRING_IDLE = 0,
	PIGEON_APP_PAIRING_SCAN_QR,
	PIGEON_APP_PAIRING_CONNECT_RELAY,
	PIGEON_APP_PAIRING_GEN_KEY_PAIR,
	PIGEON_APP_PAIRING_WAIT_ACK,
	PIGEON_APP_PAIRING_E2E_READY,
	PIGEON_APP_PAIRING_SHOW_CODE,
	PIGEON_APP_PAIRING_WAIT_PAIR_COMPLETE,
	PIGEON_APP_PAIRING_PAIRING_COMPLETE,
	PIGEON_APP_PAIRING_STATE_COUNT
} pigeon_ios_pairing_state;

// PairingCeremony ios/auth states.
typedef enum {
	PIGEON_APP_AUTH_IDLE = 0,
	PIGEON_APP_AUTH_PAIRED,
	PIGEON_APP_AUTH_RECONNECT,
	PIGEON_APP_AUTH_SEND_AUTH,
	PIGEON_APP_AUTH_SESSION_ACTIVE,
	PIGEON_APP_AUTH_STATE_COUNT
} pigeon_ios_auth_state;

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
	PIGEON_EVENT_CREDENTIAL_READY,
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
	PIGEON_EVENT_PAIRED,
	PIGEON_EVENT_COUNT
} pairing_ceremony_event_id;

// Guard and action callback types.
typedef bool (*pigeon_guard_fn)(void *ctx);
typedef int  (*pigeon_action_fn)(void *ctx);
typedef void (*pigeon_change_fn)(const char *var_name, void *ctx);

// PairingCeremony server/pairing state machine.
typedef struct {
	pigeon_server_pairing_state state;
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
	pigeon_guard_fn guards[PIGEON_GUARD_COUNT];
	pigeon_action_fn actions[PIGEON_ACTION_COUNT];
	pigeon_change_fn on_change;
	void *userdata;
} pigeon_server_pairing_machine;

void pigeon_server_pairing_machine_init(pigeon_server_pairing_machine *m);
int  pigeon_server_pairing_handle_message(pigeon_server_pairing_machine *m, pairing_ceremony_msg_type msg);
int  pigeon_server_pairing_step(pigeon_server_pairing_machine *m, pairing_ceremony_event_id event);

// PairingCeremony server/auth state machine.
typedef struct {
	pigeon_server_auth_state state;
	const char * received_device_id; // device_id from auth_request
	const char * auth_nonces_used; // set of consumed auth nonces
	const char * received_auth_nonce; // nonce from auth_request
	pigeon_guard_fn guards[PIGEON_GUARD_COUNT];
	pigeon_action_fn actions[PIGEON_ACTION_COUNT];
	pigeon_change_fn on_change;
	void *userdata;
} pigeon_server_auth_machine;

void pigeon_server_auth_machine_init(pigeon_server_auth_machine *m);
int  pigeon_server_auth_handle_message(pigeon_server_auth_machine *m, pairing_ceremony_msg_type msg);
int  pigeon_server_auth_step(pigeon_server_auth_machine *m, pairing_ceremony_event_id event);

// PairingCeremony server composite actor.
typedef struct {
	pigeon_server_pairing_machine pairing;
	pigeon_server_auth_machine auth;
	pigeon_guard_fn route_guards[PIGEON_GUARD_COUNT];
	void *userdata;
} pigeon_server_composite;

void pigeon_server_composite_init(pigeon_server_composite *c);
int  pigeon_server_route(pigeon_server_composite *c, const char *from, pairing_ceremony_event_id event);

// PairingCeremony ios/pairing state machine.
typedef struct {
	pigeon_ios_pairing_state state;
	const char * received_server_pub; // pubkey ios received in pair_hello_ack (may be adversary's)
	const char * client_shared_key; // ECDH key derived by ios (tuple to match DeriveKey output type)
	const char * ios_code; // code computed by ios from its view of the pubkeys (tuple to match DeriveCode output type)
	pigeon_guard_fn guards[PIGEON_GUARD_COUNT];
	pigeon_action_fn actions[PIGEON_ACTION_COUNT];
	pigeon_change_fn on_change;
	void *userdata;
} pigeon_ios_pairing_machine;

void pigeon_ios_pairing_machine_init(pigeon_ios_pairing_machine *m);
int  pigeon_ios_pairing_handle_message(pigeon_ios_pairing_machine *m, pairing_ceremony_msg_type msg);
int  pigeon_ios_pairing_step(pigeon_ios_pairing_machine *m, pairing_ceremony_event_id event);

// PairingCeremony ios/auth state machine.
typedef struct {
	pigeon_ios_auth_state state;
	pigeon_guard_fn guards[PIGEON_GUARD_COUNT];
	pigeon_action_fn actions[PIGEON_ACTION_COUNT];
	pigeon_change_fn on_change;
	void *userdata;
} pigeon_ios_auth_machine;

void pigeon_ios_auth_machine_init(pigeon_ios_auth_machine *m);
int  pigeon_ios_auth_handle_message(pigeon_ios_auth_machine *m, pairing_ceremony_msg_type msg);
int  pigeon_ios_auth_step(pigeon_ios_auth_machine *m, pairing_ceremony_event_id event);

// PairingCeremony ios composite actor.
typedef struct {
	pigeon_ios_pairing_machine pairing;
	pigeon_ios_auth_machine auth;
	pigeon_guard_fn route_guards[PIGEON_GUARD_COUNT];
	void *userdata;
} pigeon_ios_composite;

void pigeon_ios_composite_init(pigeon_ios_composite *c);
int  pigeon_ios_route(pigeon_ios_composite *c, const char *from, pairing_ceremony_event_id event);

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


#ifndef PIGEON_MAX_MSG
#define PIGEON_MAX_MSG 1048576
#endif

// --- Crypto types ---

typedef struct {
    uint8_t private_key[32];
    uint8_t public_key[32];
} pigeon_keypair;

typedef enum {
    PIGEON_MODE_STRICT = 0,   // streams: reject gaps
    PIGEON_MODE_DATAGRAMS = 1 // datagrams: allow gaps, reject replays
} pigeon_channel_mode;

typedef struct {
    uint8_t send_key[32];
    uint8_t recv_key[32];
    uint64_t send_seq;
    uint64_t recv_seq;
    pigeon_channel_mode mode;
    bool established; // true after pigeon_channel_init / _init_symmetric
} pigeon_channel;

typedef struct {
    char peer_instance_id[64];
    char relay_url[256];
    uint8_t local_private_key[32];
    uint8_t local_public_key[32];
    uint8_t peer_public_key[32];
} pigeon_pairing_record;

// --- Transport abstraction ---
// User provides these callbacks to connect pigeon to their QUIC stack.

typedef struct {
    void *userdata;
    // Send length-prefixed message on the bidirectional stream.
    int (*send_stream)(void *userdata, const uint8_t *data, size_t len);
    // Receive a length-prefixed message. Returns message length in *out_len.
    int (*recv_stream)(void *userdata, uint8_t *buf, size_t buf_len, size_t *out_len);
    // Send a raw datagram.
    int (*send_datagram)(void *userdata, const uint8_t *data, size_t len);
    // Receive a raw datagram. Returns datagram length in *out_len.
    int (*recv_datagram)(void *userdata, uint8_t *buf, size_t buf_len, size_t *out_len);
} pigeon_transport;

// --- Client context ---
// All library state. Allocate however you want: stack, static, embedded.

typedef struct {
    // Crypto state
    pigeon_keypair keypair;
    uint8_t peer_pubkey[32];
    pigeon_channel stream_channel;
    pigeon_channel datagram_channel;
    uint8_t hkdf_scratch[96];

    // Pairing
    pigeon_pairing_record record;
    pigeon_ios_composite pairing;

    // Transport
    pigeon_transport transport;

    // Message buffers
    uint8_t read_buf[PIGEON_MAX_MSG];
    uint8_t write_buf[PIGEON_MAX_MSG];
} pigeon_ctx;

// --- API ---

// Initialise context. Infallible for memory — just zeroes and sets defaults.
void pigeon_init(pigeon_ctx *ctx, const pigeon_transport *transport);

// --- Crypto ---

// Generate an X25519 key pair. Returns 0 on success, -1 on error.
int pigeon_generate_keypair(pigeon_keypair *kp);

// Derive a 32-byte session key from local private key + peer public key.
// info/info_len provide HKDF context. Output written to out_key (32 bytes).
int pigeon_derive_session_key(const uint8_t *private_key,
                              const uint8_t *peer_public_key,
                              const uint8_t *info, size_t info_len,
                              uint8_t *out_key);

// Derive a 6-digit confirmation code from two public keys. Order-independent.
// Writes null-terminated 7-byte string to out_code.
int pigeon_derive_confirmation_code(const uint8_t *pub_a,
                                    const uint8_t *pub_b,
                                    char *out_code);

// Initialise a channel with separate send/recv keys.
void pigeon_channel_init(pigeon_channel *ch,
                         const uint8_t *send_key,
                         const uint8_t *recv_key,
                         pigeon_channel_mode mode);

// Initialise a symmetric channel (both directions from one master key).
// is_server flips the send/recv direction labels.
int pigeon_channel_init_symmetric(pigeon_channel *ch,
                                  const uint8_t *master_key,
                                  bool is_server);

// Encrypt plaintext. Writes [8-byte seq LE][ciphertext+tag] to out.
// Returns total output length, or -1 on error.
int pigeon_channel_encrypt(pigeon_channel *ch,
                           const uint8_t *plaintext, size_t plaintext_len,
                           uint8_t *out, size_t out_len);

// Decrypt [8-byte seq LE][ciphertext+tag]. Writes plaintext to out.
// Returns plaintext length, or -1 on error.
int pigeon_channel_decrypt(pigeon_channel *ch,
                           const uint8_t *data, size_t data_len,
                           uint8_t *out, size_t out_len);

// --- Connection ---

// Send a message (length-prefixed) through the transport, optionally encrypted.
int pigeon_send(pigeon_ctx *ctx, const uint8_t *data, size_t len);

// Receive a message. Writes to ctx->read_buf. Returns message length, or -1.
int pigeon_recv(pigeon_ctx *ctx, uint8_t *out, size_t out_len);

// Send a datagram, optionally encrypted.
int pigeon_send_datagram(pigeon_ctx *ctx, const uint8_t *data, size_t len);

// Receive a datagram. Returns datagram length, or -1.
int pigeon_recv_datagram(pigeon_ctx *ctx, uint8_t *out, size_t out_len);

// --- Wire framing ---

// Write a 4-byte big-endian length prefix + payload to buf.
// Returns total written length (4 + len), or -1 if buf_len is insufficient.
int pigeon_frame_message(const uint8_t *payload, size_t len,
                         uint8_t *buf, size_t buf_len);

// Read the length prefix from a 4-byte buffer. Returns the payload length.
uint32_t pigeon_read_frame_length(const uint8_t *buf);

// --- PairingRecord serialisation ---
// Fixed-schema, zero-alloc format:
//   [0]     magic byte 0x50 ('P')
//   [1]     magic byte 0x47 ('G')
//   [2]     magic byte 0x52 ('R')
//   [3]     version byte (1)
//   [4..67]   peer_instance_id (64 bytes, null-padded)
//   [68..323] relay_url (256 bytes, null-padded)
//   [324..355] local_private_key (32 bytes)
//   [356..387] local_public_key (32 bytes)
//   [388..419] peer_public_key (32 bytes)
// Total: 420 bytes.

#define PIGEON_PAIRING_RECORD_SIZE 420

// Serialise rec into buf. Returns PIGEON_PAIRING_RECORD_SIZE on success,
// or -1 if buf_len < PIGEON_PAIRING_RECORD_SIZE.
int pigeon_pairing_record_serialize(const pigeon_pairing_record *rec,
                                    uint8_t *buf, size_t buf_len);

// Deserialise rec from buf. Returns PIGEON_PAIRING_RECORD_SIZE on success,
// or -1 if buf_len < PIGEON_PAIRING_RECORD_SIZE or the header is malformed.
int pigeon_pairing_record_deserialize(pigeon_pairing_record *rec,
                                      const uint8_t *buf, size_t buf_len);

#endif // PIGEON_H
