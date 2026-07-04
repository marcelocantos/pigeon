// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

#ifndef PIGEON_H
#define PIGEON_H

#include <stdbool.h>
#include <stddef.h>
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
	PIGEON_PAIRINGCEREMONY_MSG_HELLO = 0,
	PIGEON_PAIRINGCEREMONY_MSG_WELCOME,
	PIGEON_PAIRINGCEREMONY_MSG_CONFIRM_TO_INITIATOR,
	PIGEON_PAIRINGCEREMONY_MSG_CONFIRM_TO_ACCEPTOR,
	PIGEON_PAIRINGCEREMONY_MSG_COUNT
} pairing_ceremony_msg_type;

// PairingCeremony actions.
typedef enum {
	PIGEON_PAIRINGCEREMONY_ACTION_GEN_EPHEMERAL = 0,
	PIGEON_PAIRINGCEREMONY_ACTION_REGISTER_RELAY,
	PIGEON_PAIRINGCEREMONY_ACTION_EMIT_TOKEN,
	PIGEON_PAIRINGCEREMONY_ACTION_DERIVE_CODE,
	PIGEON_PAIRINGCEREMONY_ACTION_STORE_RECORD,
	PIGEON_PAIRINGCEREMONY_ACTION_DECODE_TOKEN,
	PIGEON_PAIRINGCEREMONY_ACTION_DIAL_RELAY,
	PIGEON_PAIRINGCEREMONY_ACTION_COUNT
} pairing_ceremony_action_id;

// PairingCeremony events.
typedef enum {
	PIGEON_PAIRINGCEREMONY_EVENT_PAIR_BEGIN = 0,
	PIGEON_PAIRINGCEREMONY_EVENT_EPHEMERAL_READY,
	PIGEON_PAIRINGCEREMONY_EVENT_RELAY_REGISTERED,
	PIGEON_PAIRINGCEREMONY_EVENT_CODE_READY,
	PIGEON_PAIRINGCEREMONY_EVENT_USER_CONFIRM,
	PIGEON_PAIRINGCEREMONY_EVENT_USER_CANCEL,
	PIGEON_PAIRINGCEREMONY_EVENT_TOKEN_RECEIVED,
	PIGEON_PAIRINGCEREMONY_EVENT_TOKEN_DECODED,
	PIGEON_PAIRINGCEREMONY_EVENT_RELAY_CONNECTED,
	PIGEON_PAIRINGCEREMONY_EVENT_RECV_HELLO,
	PIGEON_PAIRINGCEREMONY_EVENT_RECV_CONFIRM_TO_ACCEPTOR,
	PIGEON_PAIRINGCEREMONY_EVENT_RECV_WELCOME,
	PIGEON_PAIRINGCEREMONY_EVENT_RECV_CONFIRM_TO_INITIATOR,
	PIGEON_PAIRINGCEREMONY_EVENT_COUNT
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
	pigeon_action_fn actions[PIGEON_PAIRINGCEREMONY_ACTION_COUNT];
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
	pigeon_action_fn actions[PIGEON_PAIRINGCEREMONY_ACTION_COUNT];
	pigeon_change_fn on_change;
	void *userdata;
} pigeon_initiator_machine;

void pigeon_initiator_machine_init(pigeon_initiator_machine *m);
int  pigeon_initiator_handle_message(pigeon_initiator_machine *m, pairing_ceremony_msg_type msg);
int  pigeon_initiator_step(pigeon_initiator_machine *m, pairing_ceremony_event_id event);


//




// Provided by pigeon.h / c/src/pigeon.c.
int pigeon_uvarint_encode(uint64_t v, uint8_t *buf, size_t buf_len);
int pigeon_uvarint_decode(const uint8_t *buf, size_t buf_len, uint64_t *out);

// stream_header — First message on every stream — binds the stream to a name.
int pigeon_wire_stream_header_encode(const char *name, size_t name_len, uint8_t *out, size_t out_len);
int pigeon_wire_stream_header_decode(const uint8_t *buf, size_t buf_len, char *name_buf, size_t name_buf_len, size_t *name_len_out);

// datagram_plaintext — Plaintext layout of a datagram before AEAD encryption.
int pigeon_wire_datagram_plaintext_encode(uint64_t channel_id, const uint8_t *payload, size_t payload_len, uint8_t *out, size_t out_len);
int pigeon_wire_datagram_plaintext_decode(const uint8_t *buf, size_t buf_len, uint64_t *channel_id, uint8_t *payload_buf, size_t payload_buf_len, size_t *payload_len_out);

// relay_greeting — Relay-side greeting sent on the primary stream after dial.
typedef enum {
    RELAY_GREETING_CONNECT = 0,
    RELAY_GREETING_REGISTER = 1,
    RELAY_GREETING_LISTEN = 2,
    RELAY_GREETING_UNKNOWN = -1
} pigeon_wire_relay_greeting_variant;

int pigeon_wire_relay_greeting_encode_connect(const char *instance_id, size_t instance_id_len, uint8_t *out, size_t out_len);
int pigeon_wire_relay_greeting_encode_register(const char *token, const char *instance_id, uint8_t *out, size_t out_len);
int pigeon_wire_relay_greeting_encode_listen(const char *token, const char *instance_id, uint8_t *out, size_t out_len);
int pigeon_wire_relay_greeting_decode(const uint8_t *buf, size_t buf_len, pigeon_wire_relay_greeting_variant *out_variant, char *instance_id_buf, size_t instance_id_buf_len, size_t *instance_id_len_out, char *token_buf, size_t token_buf_len, size_t *token_len_out);




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

// Opaque per-stream handle. The transport defines its own concrete type
// behind this pointer and pigeon never inspects it; pigeon just stores
// it on a pigeon_stream and hands it back on subsequent stream I/O.
typedef struct pigeon_stream_handle pigeon_stream_handle;

typedef struct {
    void *userdata;

    // --- Single-stream + datagram (legacy / pair-mode) ---
    //
    // These callbacks operate on the transport's primary bidirectional
    // stream and the connection-level datagram channel. They are the
    // foundation of the v0.16-era single-channel API and remain
    // available for backwards compatibility (e.g. the cmd/crypto-peer
    // cross-language test).

    // Send length-prefixed message on the bidirectional stream.
    int (*send_stream)(void *userdata, const uint8_t *data, size_t len);
    // Receive a length-prefixed message. Returns message length in *out_len.
    int (*recv_stream)(void *userdata, uint8_t *buf, size_t buf_len, size_t *out_len);
    // Send a raw datagram.
    int (*send_datagram)(void *userdata, const uint8_t *data, size_t len);
    // Receive a raw datagram. Returns datagram length in *out_len.
    int (*recv_datagram)(void *userdata, uint8_t *buf, size_t buf_len, size_t *out_len);

    // --- Multi-stream (post-T22) ---
    //
    // These callbacks support the post-T22 multi-channel wire. A
    // transport that doesn't support multiple QUIC streams can leave
    // them NULL; the higher-level pigeon_session API will refuse to
    // open additional streams and return -1.
    //
    // Stream framing on the wire is length-prefixed messages (the
    // transport handles framing). The first message on every stream
    // is the unencrypted name-binding header produced by
    // pigeon_encode_stream_header(); subsequent messages are AEAD-
    // encrypted by pigeon_session_*_send and decrypted by the matching
    // _recv on the peer.

    // Open a new bidirectional stream. Returns 0 on success and writes
    // the stream handle to *out_handle.
    int (*open_stream)(void *userdata, pigeon_stream_handle **out_handle);
    // Accept the next incoming bidirectional stream from the peer.
    // Blocks until a stream arrives or the transport is closed.
    int (*accept_stream)(void *userdata, pigeon_stream_handle **out_handle);
    // Send a length-prefixed message on a specific stream.
    int (*send_on_stream)(void *userdata, pigeon_stream_handle *handle,
                          const uint8_t *data, size_t len);
    // Receive the next length-prefixed message on a specific stream.
    int (*recv_on_stream)(void *userdata, pigeon_stream_handle *handle,
                          uint8_t *buf, size_t buf_len, size_t *out_len);
    // Close a specific stream.
    int (*close_stream)(void *userdata, pigeon_stream_handle *handle);
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

    // Pairing record (post-ceremony state). Callers drive the
    // pairing ceremony itself via the generated state machines
    // (pigeon_acceptor_machine / pigeon_initiator_machine in
    // pairingceremony_gen.h) — this struct doesn't pre-allocate
    // the FSM.
    pigeon_pairing_record record;

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

// Fill buf with n cryptographically-secure random bytes. Used to mint the
// per-session activation nonce (see PIGEON_AUTH_NONCE_LEN in activation.h).
void pigeon_random_bytes(uint8_t *buf, size_t n);

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

// --- Multi-channel wire helpers (post-T22) ---
//
// Maximum number of bytes a Go-style unsigned varint can take (10 bytes
// covers uint64). Matches encoding/binary.MaxVarintLen64.
#define PIGEON_MAX_VARINT_LEN 10

// Maximum size of an unencrypted stream-header: varint name-len +
// 256 bytes name. Under T45 there is no clientTag prefix — each
// session rides its own end-to-end QUIC pipe.
#define PIGEON_MAX_STREAM_HEADER (PIGEON_MAX_VARINT_LEN + 256)

// Encode an unsigned varint (Go's binary.PutUvarint format) into buf.
// Returns number of bytes written, or -1 if buf_len is insufficient.
// Each byte stores 7 bits of value; the high bit signals continuation.
int pigeon_uvarint_encode(uint64_t v, uint8_t *buf, size_t buf_len);

// Decode an unsigned varint from buf. On success, writes the decoded
// value to *out and returns the number of bytes consumed (1..10).
// Returns 0 if buf is too short, -1 if the encoding overflows uint64.
int pigeon_uvarint_decode(const uint8_t *buf, size_t buf_len, uint64_t *out);

// Stream-header encoder/decoder moved to wire_gen.h (protogen-
// generated from protocol/wireformats.yaml): see
// pigeon_wire_stream_header_encode, pigeon_wire_stream_header_decode.
// Under T45 the header is just [varint name-len][name].

// Compose a datagram payload for a named channel:
//   plain    = [varint channel-id][payload]
//   wire     = AEAD(plain)
// `out` must be sized for the final wire bytes. The pigeon_channel
// supplied performs the AEAD encryption. Under T45 there is no
// clientTag prefix — each session owns its own QUIC pipe. Returns total
// wire bytes written, or -1 on error.
int pigeon_encode_datagram(pigeon_channel *ch,
                           uint64_t channel_id,
                           const uint8_t *payload, size_t payload_len,
                           uint8_t *out, size_t out_len);

// Decode a datagram in the post-T45 framing: AEAD-decrypt using the
// supplied channel, then parse the [varint channel-id][payload]
// plaintext. On success writes the channel id to *channel_id and the
// application payload to `payload_buf`; returns the payload length.
// -1 on AEAD failure or any framing error.
int pigeon_decode_datagram(pigeon_channel *ch,
                           const uint8_t *wire, size_t wire_len,
                           uint64_t *channel_id,
                           uint8_t *payload_buf, size_t payload_buf_len);

// --- Multi-channel session API (post-T22) ---
//
// A pigeon_session is a single peer-to-peer association: backend ↔
// one paired client (or vice-versa from the client side). Under T45's
// remote-Listen L1 model each session owns its own end-to-end QUIC
// pipe (the relay bridges the two connections opaquely), so both sides
// are symmetric — there is no role discriminator and no relay-assigned
// clientTag. The session owns the AEAD channel derived from the
// PairingRecord plus a small fixed-size table of pre-declared datagram
// channels (name → varint id, agreed by both peers up-front). Streams
// are opened on demand and live as long as the underlying transport
// stream.

#define PIGEON_MAX_DATAGRAM_CHANNELS 16
#define PIGEON_MAX_NAME_LEN 64

// Maximum number of pre-arrived sub-streams buffered per session
// while the application hasn't yet called accept_incoming_stream.
// Declared up here (rather than next to the listener API below) so
// the pigeon_session struct can size its incoming-stream array.
#define PIGEON_SESSION_MAX_INCOMING 8

typedef struct {
    char     name[PIGEON_MAX_NAME_LEN];
    uint64_t channel_id;
} pigeon_dgchannel_def;

// pigeon_listen_owner_fn closes / frees the resources behind the
// owner cookie a listen-dialer hands to an adopted Session (see the
// Listener API below). Declared here so the pigeon_session struct can
// hold the teardown hooks.
typedef void (*pigeon_listen_owner_fn)(void *owner);

typedef struct {
    // PairingRecord-derived AEAD channel. Encrypts every stream message
    // (after the unencrypted name-binding header) and every datagram
    // payload.
    //
    // In pairing mode (no PairingRecord; pre-pairing handshake), the
    // channel's `established` flag is false and stream/datagram APIs
    // that rely on AEAD will refuse to operate. Callers drive the
    // pairing ceremony over Session.Primary() directly, then derive
    // a channel and re-init the session for the post-pairing wire.
    pigeon_channel channel;

    // Transport vtable + opaque userdata. Under T45 each session rides
    // its own end-to-end QUIC pipe.
    pigeon_transport transport;

    // Optional owned-transport teardown hooks. When a Session is built
    // by pigeon_listener_accept it ADOPTS the listen transport produced
    // by the listen-dialer: owner is the dialer's opaque cookie, and
    // pigeon_session_close invokes owner_close then owner_free on it.
    // NULL on Sessions whose transport the caller owns separately (the
    // lower-level pigeon_session_init path and the pigeon_connect
    // wrapper, which closes the transport itself).
    void                  *owner;
    pigeon_listen_owner_fn owner_close;
    pigeon_listen_owner_fn owner_free;

    // Primary stream handle, populated by pigeon_connect_on_transport /
    // (T32.2) pigeon_listener_accept. NULL when the session was built
    // via the lower-level pigeon_session_init path and no primary has
    // been bound. pigeon_session_primary() wraps this as a *pigeon_stream*.
    //
    // Lifetime: owned by the transport (the transport's open_stream /
    // accept_stream produced it). The session does not close it on
    // pigeon_session_close — the caller closes the transport, which
    // tears down all streams it owns.
    pigeon_stream_handle *primary;

    // Pre-declared datagram channels.
    pigeon_dgchannel_def datagrams[PIGEON_MAX_DATAGRAM_CHANNELS];
    size_t               datagram_count;

    // Heap-allocated scratch buffers used by the per-call I/O wrappers
    // (pigeon_stream_send/recv, pigeon_datagram_send/recv). Sized at
    // PIGEON_MAX_MSG + AEAD overhead (max 64 bytes for nonce prefix +
    // tag). Allocated lazily on first send/recv; freed by
    // pigeon_session_close.
    //
    // Why heap: the previous stack-allocated [PIGEON_MAX_MSG + 64]
    // arrays (≈ 1 MiB) overflow the default thread-stack size on every
    // host runtime that wraps libpigeon — Swift's cooperative pool
    // (~256 KiB), Apple's GCD workers (~64 KiB), JVM threads (~512 KiB
    // on macOS aarch64). Per-session scratch on the heap keeps the C
    // ABI usable from any host thread without 4 MiB stack workarounds.
    uint8_t      *scratch_a;
    uint8_t      *scratch_b;
    size_t        scratch_size;

    // Sub-streams the peer opened on this session's own QUIC pipe that
    // arrived while the application was waiting for a differently-named
    // stream. Filled by pigeon_session_accept_incoming_stream as it
    // pulls streams off this session's transport; drained by name on a
    // later matching call. The C ABI is single-threaded so no lock is
    // needed.
    struct {
        pigeon_stream_handle *handle;
        char                  name[PIGEON_MAX_NAME_LEN];
        bool                  in_use;
    } incoming[PIGEON_SESSION_MAX_INCOMING];
} pigeon_session;

typedef struct {
    pigeon_session       *session;
    pigeon_stream_handle *handle;
    char                  name[PIGEON_MAX_NAME_LEN];
} pigeon_stream;

typedef struct {
    pigeon_session *session;
    uint64_t        channel_id;
    char            name[PIGEON_MAX_NAME_LEN];
} pigeon_datagram;

// Initialise a session. Copies the transport vtable. Channel must
// already be initialised (e.g. via pigeon_channel_init from
// PairingRecord-derived send/recv keys).
//
// Returns 0 on success, -1 if validation fails (duplicate datagram ids,
// names too long, etc.). The datagram-channel list is the pre-agreed
// (name, id) mapping; both peers must declare an identical list.
int pigeon_session_init(pigeon_session *s,
                        const pigeon_transport *transport,
                        const pigeon_channel *channel,
                        const pigeon_dgchannel_def *datagrams,
                        size_t datagram_count);

// Open a fresh named stream toward the peer. Writes the
// [varint name-len][name] header as the first message on the new
// stream. Returns 0 on success.
int pigeon_session_open_stream(pigeon_session *s,
                               const char *name,
                               pigeon_stream *out_stream);

// Look up a pre-declared datagram channel by name. Returns 0 on
// success and populates *out, -1 if the name isn't in the session's
// datagram-channel list.
int pigeon_session_get_datagram(pigeon_session *s,
                                const char *name,
                                pigeon_datagram *out);

// Wrap the session's primary stream (bound at pigeon_connect /
// pigeon_listener_accept time) as a *pigeon_stream*. Mirrors Go's
// Session.Primary() (T39.6.1).
//
// Meaningful in pairing-mode where the activation handshake is
// skipped and the primary is otherwise unused — pairing.c and the
// cross-language crypto-peer fixtures use this to talk on the
// primary stream without opening a sub-stream (the modern client
// side might not support multi-stream QUIC, e.g. Swift NWConnection).
//
// In activation-mode the primary is consumed by pigeon_run_*_activation
// during session construction; reading from it after activation will
// block. pigeon_session_primary() is safe to call regardless, but only
// useful in pairing-mode.
//
// Returns 0 on success and populates *out_stream. Returns -1 if no
// primary handle is bound on this session (legacy session_init path,
// or session never went through pigeon_connect / pigeon_listener_accept).
int pigeon_session_primary(pigeon_session *s, pigeon_stream *out_stream);

// --- pigeon_connect (T32.3) ---
//
// Bring up a client-side session against a paired backend. Two modes,
// distinguished by whether `record` is supplied (mirrors Go's
// pigeon.Connect from api.go):
//
//   * Activation mode (record != NULL, device_id != NULL): runs the
//     auth_request / auth_ok handshake against the device id; on
//     success the returned session carries an AEAD channel derived
//     from `record`.
//   * Pairing mode (record == NULL, device_id == NULL): activation
//     is skipped and the session is returned with channel.established
//     == false. The caller drives the pairing ceremony over
//     pigeon_session_primary() and re-derives the channel from the
//     resulting PairingRecord (docs/DESIGN.md §3 L2).
//
// pigeon_connect_on_transport is the transport-agnostic core: write
// the empty-name primary header, run client activation (or skip in
// pairing mode), derive the AEAD channel from `record`, init the
// session with the supplied datagram channel set, and bind the
// primary handle. The caller owns `transport` (and `primary_handle`)
// — pigeon_session_close does NOT close them. This factoring lets the
// in-process loopback tests exercise the full handshake against
// pigeon_run_backend_activation without bringing up a live relay.
//
// Returns 0 on success.
int pigeon_connect_on_transport(const pigeon_transport *transport,
                                pigeon_stream_handle *primary_handle,
                                const char *peer_instance_id,
                                const char *device_id,
                                const pigeon_pairing_record *record,
                                const pigeon_dgchannel_def *datagrams,
                                size_t datagram_count,
                                pigeon_session *out_session);

// AEAD-encrypt and send one application-level message on the stream.
// Returns 0 on success, -1 on transport or encryption error.
int pigeon_stream_send(pigeon_stream *s,
                       const uint8_t *msg, size_t msg_len);

// Receive the next application-level message on the stream. AEAD-
// decrypts and copies the plaintext to `buf`. Returns the plaintext
// length, or -1 on transport / decryption error.
int pigeon_stream_recv(pigeon_stream *s,
                       uint8_t *buf, size_t buf_len);

// Close the stream's underlying transport handle.
int pigeon_stream_close(pigeon_stream *s);

// Free the heap scratch buffers allocated lazily on first send/recv.
// Idempotent. Call when the session is no longer needed; the rest of
// the pigeon_session struct can still be re-initialised afterwards.
void pigeon_session_close(pigeon_session *s);

// Send one datagram on the named channel. Wraps with the post-T45
// framing (AEAD([varint id][payload])). Returns 0 on success.
int pigeon_datagram_send(pigeon_datagram *d,
                         const uint8_t *payload, size_t payload_len);

// Receive the next datagram on the connection's datagram channel.
// AEAD-decrypts, parses the channel-id, and — if the channel-id
// matches the requested datagram — copies the application payload to
// `buf` and returns its length. If the channel-id doesn't match,
// returns 0 (the caller should re-dispatch this datagram). Returns -1
// on framing or AEAD error.
//
// Application code that wants per-channel queues should run a single
// pump on the connection's datagram channel and demultiplex into
// per-channel buffers; pigeon_datagram_recv is a single-channel
// convenience for tests and simple client-side patterns.
int pigeon_datagram_recv(pigeon_datagram *d,
                         uint8_t *buf, size_t buf_len);

// --- Pairing ceremony wire driver ---
//
// These two functions implement the acceptor/initiator sides of the
// hello/welcome/confirm wire exchange. They are wire-compatible with Go's
// pairing.runAcceptor / runInitiator: messages are JSON with standard
// base64-encoded byte fields, framed by the transport's send_on_stream /
// recv_on_stream callbacks.
//
// transport: must have open_stream / accept_stream / send_on_stream /
//   recv_on_stream populated. pigeon_pair_acceptor calls accept_stream to
//   wait for the initiator; pigeon_pair_initiator calls open_stream.
// local_eph_priv/pub: 32-byte X25519 key pair generated before the call.
// identity_pub: 32-byte identity public key.
// instance_id: NUL-terminated instance identifier string.
// confirm_fn: called with (userdata, code) after the code is derived.
//   Return 1 to confirm, 0 (or negative) to cancel.
// out_record: filled on success with the new pairing record.
// out_code: 7-byte buffer; receives the 6-digit code + NUL on success.
// Returns 0 on success, -1 on any error.

int pigeon_pair_acceptor(
    const pigeon_transport *transport,
    const uint8_t *local_eph_priv,
    const uint8_t *local_eph_pub,
    const uint8_t *identity_pub,
    const char *instance_id,
    int (*confirm_fn)(void *userdata, const char *code),
    void *userdata,
    pigeon_pairing_record *out_record,
    char *out_code);

// acc_eph_pub: 32-byte acceptor ephemeral public key decoded from the token.
// acc_instance: acceptor's instance ID decoded from the token.
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
    char *out_code);

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

// --- Multi-client Listener (post-T32.2) ---
//
// The Listener mirrors Go's pigeon.Register / Listener.Accept under
// T45's remote-Listen L1 model (docs/DESIGN.md §3). A backend holds
// one register control connection open for the instance's lifetime,
// then dials a fresh `listen` connection per client. The relay matches
// each arriving client to a parked listen and bridges the two QUIC
// connections end-to-end, so every accepted Session rides its OWN
// pipe — there is no shared connection or per-client clientTag demux.
//
// The listener does not own the transports itself; instead it holds a
// `dial_listen` callback that produces a fresh listen transport on
// each accept. pigeon_register wires this to an ngtcp2 dialer
// (PIGEON_ROLE_LISTEN); tests wire it to a loopback dialer. The
// callback owns the heap memory for each transport and the listener
// hands that ownership to the resulting Session.
//
// Threading: per docs/DESIGN.md §5, the C ABI is "single-threaded,
// caller-owns-concurrency". pigeon_listener_accept does its work
// synchronously on the calling thread — it dials a listen, runs the
// activation handshake, and returns one Session bound to that pipe.
//
// Memory: each accepted Session OWNS its own listen transport and
// closes it on pigeon_session_close. The application is responsible
// for closing accepted Sessions (call pigeon_session_close). The
// register control connection is owned by the listener and torn down
// by pigeon_listener_close.

// Opaque listener handle. Definition lives in c/src/listener.c.
typedef struct pigeon_listener pigeon_listener;

// pigeon_resolve_device_fn looks up a connecting client's
// PairingRecord by its device ID. Identical to the typedef of the
// same name in activation.h — repeated here (C11 allows redundant
// typedefs of the same shape) so the listener API doesn't require
// the activation header. Return 0 on success (out_record populated
// as a pigeon_pairing_record*), -1 to reject the client.
typedef int (*pigeon_resolve_device_fn)(void *userdata,
                                        const char *device_id,
                                        void *out_record);

// pigeon_listen_dialer dials one fresh `listen` connection for the
// instance and waits for the relay to bridge a client onto it. On
// success it must:
//   * fill *out_transport with the listen connection's transport vtable,
//   * write the bridged primary stream handle into *out_primary,
//   * write an opaque owner cookie into *out_owner (passed back to the
//     close/free hooks when the Session that adopts this transport is
//     torn down).
// Returns 0 on success, -1 on dial/bridge failure. Invoked
// synchronously from pigeon_listener_accept on the caller's thread.
typedef int (*pigeon_listen_dialer)(void *userdata,
                                    pigeon_transport *out_transport,
                                    pigeon_stream_handle **out_primary,
                                    void **out_owner);

// (pigeon_listen_owner_fn is declared above, next to the
// pigeon_session struct that stores the teardown hooks.) owner_close
// is called first (tear down the QUIC connection), then owner_free
// (release the heap).

// Initialise a listener around a register control transport and a
// listen-dialer. The control transport is kept open for the instance's
// lifetime (no traffic flows on it). `instance_id` is the relay-
// assigned (or self-assigned) instance ID this listener advertises.
//
// `dial_listen` / `dial_userdata` produce a fresh bridged listen
// connection on each accept (see pigeon_listen_dialer). `owner_close`
// / `owner_free` tear down the per-listen transport when its Session
// closes.
//
// `pairing` is the per-client device-id → PairingRecord lookup,
// invoked synchronously on the accept thread when a client's
// auth_request arrives. `datagrams` declares the named datagram
// channels available on each accepted session.
//
// Returns 0 on success and writes the listener handle into *out; the
// listener is heap-allocated and must be released with
// pigeon_listener_close. Returns -1 on validation failure.
int pigeon_listener_init(pigeon_listener **out,
                         const pigeon_transport *control,
                         const char *instance_id,
                         pigeon_listen_dialer dial_listen,
                         void *dial_userdata,
                         pigeon_listen_owner_fn owner_close,
                         pigeon_listen_owner_fn owner_free,
                         pigeon_resolve_device_fn pairing,
                         void *pairing_userdata,
                         const pigeon_dgchannel_def *datagrams,
                         size_t datagram_count);

// Return the listener's relay-assigned instance ID. The pointer
// remains valid for the listener's lifetime.
const char *pigeon_listener_instance_id(const pigeon_listener *l);

// Dial a fresh listen connection, wait for the relay to bridge a
// client onto it, run the auth_request / auth_ok activation handshake
// against the resolver supplied to pigeon_listener_init, and return
// the resulting session bound to that pipe.
//
// Returns 0 on success and writes the heap-allocated session pointer
// into *out_session. The session owns its listen transport; the caller
// owns the session and must release it with pigeon_session_close
// (which closes the underlying transport). Returns -1 on dial /
// activation failure or listener shutdown.
int pigeon_listener_accept(pigeon_listener *l, pigeon_session **out_session);

// Tear down the listener: close the register control transport and
// free the listener struct itself. Idempotent on NULL. Sessions
// already returned by pigeon_listener_accept are independently owned
// and unaffected — close them via pigeon_session_close.
void pigeon_listener_close(pigeon_listener *l);

// Pull the next peer-opened sub-stream with the given name off this
// session's own QUIC pipe. Streams that arrive with a different name
// are buffered for a later matching call. Returns 0 on success and
// populates out_stream, -1 if the transport fails or shuts down.
//
// Because each session owns its connection, this accepts directly
// from the session's transport — there is no listener demux to pump.
int pigeon_session_accept_incoming_stream(pigeon_session *s,
                                          const char *name,
                                          pigeon_stream *out_stream);

// --- pigeon_register convenience (high-level) ---
//
// `pigeon_register` is a thin wrapper that dials the register control
// connection (PIGEON_ROLE_REGISTER) and wires an ngtcp2 listen-dialer
// (PIGEON_ROLE_LISTEN) into pigeon_listener_init. It lives in a
// separate compilation unit (c/src/listener_ngtcp2.c) because it
// depends on ngtcp2 + quictls (vendored under c/vendor/), which the
// amalgamated build does not link. Builds that need this entry point
// link c/src/listener.c + c/src/listener_ngtcp2.c +
// c/src/ngtcp2_transport.c against the vendored libs (see
// test-c-ngtcp2 in the Makefile).
//
// On success, writes the listener handle into *out_listener, the
// relay-assigned instance ID into out_instance_id (NUL-terminated;
// truncated if the buffer is too small), and returns 0. The listener
// owns the register control transport; pigeon_listener_close tears it
// down. Each accepted Session owns its own listen transport.
int pigeon_register(const char *relay_host,
                    const char *relay_port,
                    const char *self_instance_id,  // may be NULL
                    const char *token,             // may be NULL
                    pigeon_resolve_device_fn pairing,
                    void *pairing_userdata,
                    const pigeon_dgchannel_def *datagrams,
                    size_t datagram_count,
                    pigeon_listener **out_listener,
                    char *out_instance_id, size_t out_instance_id_cap);

// --- pigeon_connect convenience (high-level, T32.4) ---
//
// `pigeon_connect` is the client-side mirror of `pigeon_register`:
// brings up an ngtcp2 transport (PIGEON_ROLE_CONNECT), binds the
// primary QUIC stream as a multi-channel slot, and runs the
// auth_request / auth_ok activation handshake (pairing mode: an empty
// arrival marker) via pigeon_connect_on_transport. Lives in the same
// compilation unit as pigeon_register because both depend on the
// ngtcp2 transport.
//
// On success, *out_conn points to a freshly allocated pigeon_connection
// that owns the underlying ngtcp2 transport and the resulting
// pigeon_session. Use pigeon_connect_session() to obtain the session
// pointer (suitable for pigeon_session_open_stream et al), and call
// pigeon_connect_close() when finished — it tears down the session,
// closes the ngtcp2 transport, and frees the connection struct.

// Opaque connection handle. Owns one ngtcp2 transport + one session.
typedef struct pigeon_connection pigeon_connection;

int pigeon_connect(const char *relay_host,
                   const char *relay_port,
                   const char *peer_instance_id,
                   const char *device_id,
                   const char *token, // may be NULL
                   const pigeon_pairing_record *record,
                   const pigeon_dgchannel_def *datagrams,
                   size_t datagram_count,
                   pigeon_connection **out_conn);

// Return the connection's session pointer. The session is owned by
// the connection; do NOT call pigeon_session_close on it. Stays valid
// until pigeon_connect_close().
pigeon_session *pigeon_connect_session(pigeon_connection *c);

// Tear down the connection: close the session, close the underlying
// ngtcp2 transport, free the connection struct. Idempotent on NULL.
void pigeon_connect_close(pigeon_connection *c);

#endif // PIGEON_H

#ifndef PIGEON_H_AMALGAMATED_EXTRAS
#define PIGEON_H_AMALGAMATED_EXTRAS

// --- SessionMachine declarations (from session_gen.h) ---





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


// --- Activation handshake driver (from activation.h) ---

//
// Activation drives the session.yaml pairing-phase auth states
// (Paired -> AuthCheck -> SessionActive on the backend; Reconnect ->
// SendAuth -> SessionActive on the client) on a per-client
// SessionMachine. The machine instance is exclusive to this client
// connection -- no shared state between clients on the backend side.
//
// The exchange is two messages on the relay's primary stream,
// immediately after the stream-name binding header:
//
//	client  -> backend: auth_request{device_id}
//	backend -> client:  auth_ok{ok, reason?}
//
// Both messages are sent via the transport's send_on_stream /
// recv_on_stream callbacks (which handle the 4-byte big-endian
// length prefix). The wire format below is byte-for-byte identical
// to the Go-side activation.go shipped under T39.1.



// Activation's API surface only needs opaque pointers, so keep this
// header dependency-light — implementation files include the right
// generated headers for their TU.

#ifdef __cplusplus
extern "C" {
#endif

// Opaque type aliases. The driver functions take void* for the
// machine arguments so this header doesn't need to include
// session_gen.h. Caller passes a real `pigeon_backend_machine *` /
// `pigeon_client_machine *` (zero-initialised); activation.c casts
// internally. Same shape for the transport / stream / pairing-
// record types defined in pigeon.h.

// --- Wire constants (mirror activation.go) ---

// authMsgTagAuthRequest tags an auth_request payload.
#define PIGEON_AUTH_MSG_TAG_AUTH_REQUEST ((uint8_t)0x01)
// authMsgTagAuthOk tags an auth_ok payload.
#define PIGEON_AUTH_MSG_TAG_AUTH_OK      ((uint8_t)0x02)
// authOkAccepted is the ok flag value for an accepted activation.
#define PIGEON_AUTH_OK_ACCEPTED          ((uint8_t)0x01)
// authOkRejected is the ok flag value for a rejected activation
// (carries a reason string).
#define PIGEON_AUTH_OK_REJECTED          ((uint8_t)0x00)

// PIGEON_AUTH_MAX_DEVICE_ID is a hard cap on the device-id string we
// accept on the wire. Mirrors the Go side's implicit bound (device
// IDs in pigeon are 32-byte hex strings = 64 chars + NUL).
#define PIGEON_AUTH_MAX_DEVICE_ID 128

// PIGEON_AUTH_MAX_REASON is a hard cap on the reason string in a
// rejected auth_ok. Keeps wire decode bounded.
#define PIGEON_AUTH_MAX_REASON    256

// PIGEON_AUTH_NONCE_LEN is the length of the fresh per-session nonce the
// client mints and carries in auth_request. Both sides fold it into the
// session-key HKDF info so that concurrent or reconnecting sessions under
// one PairingRecord derive distinct AEAD keys (🎯T44.1). Client-contributed
// only: auth_ok carries no nonce.
#define PIGEON_AUTH_NONCE_LEN     16

// --- Wire encoders / decoders ---

// pigeon_encode_auth_request builds the binary payload for the client's
// auth_request message:
//   [tag][uvarint device-id len][device-id][PIGEON_AUTH_NONCE_LEN nonce].
// `nonce` must point to PIGEON_AUTH_NONCE_LEN bytes. Returns the encoded
// length on success, or -1 if buf_len is too small.
int pigeon_encode_auth_request(const char *device_id,
                               const uint8_t *nonce,
                               uint8_t *buf, size_t buf_len);

// pigeon_decode_auth_request parses an auth_request payload, copies the
// device ID (NUL-terminated) into out_device_id, and copies the
// PIGEON_AUTH_NONCE_LEN-byte session nonce into out_nonce. Returns 0 on
// success, -1 on malformed payload or oversized device id.
int pigeon_decode_auth_request(const uint8_t *payload, size_t payload_len,
                               char *out_device_id, size_t out_cap,
                               uint8_t *out_nonce);

// pigeon_encode_auth_ok builds the binary payload for the backend's
// auth_ok reply. ok=true: [tag][0x01]; ok=false: [tag][0x00][uvarint
// reason len][reason]. Returns encoded length or -1 on overflow.
int pigeon_encode_auth_ok(bool ok, const char *reason,
                          uint8_t *buf, size_t buf_len);

// pigeon_decode_auth_ok parses an auth_ok payload. Sets *out_ok and
// (on rejection) copies the reason into out_reason. Returns 0 on
// success, -1 on malformed payload.
int pigeon_decode_auth_ok(const uint8_t *payload, size_t payload_len,
                          bool *out_ok,
                          char *out_reason, size_t out_reason_cap);

// --- Per-side drive helpers ---

// pigeon_resolve_device_fn is invoked by pigeon_run_backend_activation
// to look up a connecting client's PairingRecord by its device ID.
// The callback owns the record memory; the activation driver only
// reads it for the duration of the call.
//
// Return values: 0 on success (out_record populated), -1 if the
// device is unknown or the lookup fails. Mirrors Go's
// runBackendActivation's `resolve(deviceID) (*crypto.PairingRecord,
// error)` shape.
// out_record is a `pigeon_pairing_record *` (cast from void *). The
// callback writes the resolved record into *out_record.
typedef int (*pigeon_resolve_device_fn)(void *userdata,
                                        const char *device_id,
                                        void *out_record);

// pigeon_run_backend_activation drives the backend's per-client
// SessionMachine through Paired -> AuthCheck -> SessionActive on a
// valid auth_request, or Paired -> AuthCheck -> Idle on an unknown
// device.
//
// Reads the auth_request from `stream`, looks up the PairingRecord
// via `resolve`, writes the auth_ok reply, populates `out_machine`
// with the post-activation backend machine state, copies the
// resolved device ID into `out_device_id`, and (on success) writes
// the looked-up PairingRecord into `out_record`.
//
// Returns 0 on success. Returns -1 on wire / transition failure
// (out_device_id empty); returns 1 when the device ID was decoded
// but the lookup rejected the client (out_device_id populated, no
// record). Mirrors Go's runBackendActivation's tri-valued return.
// transport / stream / out_machine / out_record are all opaque
// pointer types here — see the comment block above. Pass real
// pigeon_transport* / pigeon_stream_handle* / pigeon_backend_machine*
// / pigeon_pairing_record* values (cast happens internally).
//
// On success (return 0), out_nonce receives the PIGEON_AUTH_NONCE_LEN-byte
// session nonce decoded from the client's auth_request; the caller folds it
// into session-key derivation (🎯T44.1). out_nonce must point to at least
// PIGEON_AUTH_NONCE_LEN bytes; it is untouched on the -1/1 paths.
int pigeon_run_backend_activation(const void *transport,
                                  void *stream,
                                  pigeon_resolve_device_fn resolve,
                                  void *resolve_userdata,
                                  void *out_machine,
                                  char *out_device_id, size_t out_device_id_cap,
                                  void *out_record,
                                  uint8_t *out_nonce);

// pigeon_run_client_activation drives the client's per-client
// SessionMachine through Reconnect -> SendAuth -> SessionActive.
// Mints a fresh PIGEON_AUTH_NONCE_LEN-byte session nonce, writes
// auth_request{device_id, nonce} to `stream`, reads the auth_ok reply,
// and on acceptance populates `out_machine`. On rejection returns -1 and
// (if out_reason is non-NULL) copies the backend's reason into out_reason.
//
// On success (return 0), out_nonce receives the minted session nonce; the
// caller folds it into session-key derivation (🎯T44.1). out_nonce must
// point to at least PIGEON_AUTH_NONCE_LEN bytes.
//
// Returns 0 on success, -1 on transport / decode / rejection.
int pigeon_run_client_activation(const void *transport,
                                 void *stream,
                                 const char *device_id,
                                 void *out_machine,
                                 char *out_reason, size_t out_reason_cap,
                                 uint8_t *out_nonce);

#ifdef __cplusplus
}
#endif


#endif // PIGEON_H_AMALGAMATED_EXTRAS
