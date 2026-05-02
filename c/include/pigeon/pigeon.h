// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

#ifndef PIGEON_H
#define PIGEON_H

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

// Include the generated protocol header.
#include "pairingceremony_gen.h"

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

// Maximum size of an unencrypted stream-header (backend side):
// 4 bytes (clientTag, big-endian uint32) + varint name-len + 256 bytes name.
#define PIGEON_MAX_STREAM_HEADER (4 + PIGEON_MAX_VARINT_LEN + 256)

// Encode an unsigned varint (Go's binary.PutUvarint format) into buf.
// Returns number of bytes written, or -1 if buf_len is insufficient.
// Each byte stores 7 bits of value; the high bit signals continuation.
int pigeon_uvarint_encode(uint64_t v, uint8_t *buf, size_t buf_len);

// Decode an unsigned varint from buf. On success, writes the decoded
// value to *out and returns the number of bytes consumed (1..10).
// Returns 0 if buf is too short, -1 if the encoding overflows uint64.
int pigeon_uvarint_decode(const uint8_t *buf, size_t buf_len, uint64_t *out);

// Encode the per-stream first-message header in the post-T22 wire:
//   backend side: [4-byte clientTag-BE][varint name-len][name]
//   client  side:                       [varint name-len][name]
//
// Returns the number of bytes written, or -1 if out_len is insufficient.
// `name` may be NULL iff name_len == 0 (legitimate for the client primary
// stream, which uses an empty name).
int pigeon_encode_stream_header(bool is_backend, uint32_t client_tag,
                                const char *name, size_t name_len,
                                uint8_t *out, size_t out_len);

// Decode a backend-side stream header: extracts the 4-byte clientTag and
// then the varint-prefixed name. The name is copied into `name_buf` (NUL-
// terminated; truncated and an error returned if name_buf_len is too
// small). Writes the decoded name length to *name_len_out (excluding NUL).
// Returns the total number of bytes consumed, or -1 on error.
int pigeon_decode_backend_stream_header(const uint8_t *buf, size_t buf_len,
                                        uint32_t *client_tag,
                                        char *name_buf, size_t name_buf_len,
                                        size_t *name_len_out);

// Decode a client-side stream header (no clientTag).
// Returns the total number of bytes consumed, or -1 on error.
int pigeon_decode_client_stream_header(const uint8_t *buf, size_t buf_len,
                                       char *name_buf, size_t name_buf_len,
                                       size_t *name_len_out);

// Compose a datagram payload for a named channel:
//   plain    = [varint channel-id][payload]
//   wire     = AEAD(plain)                    on the client side
//   wire     = [4-byte tag-BE][AEAD(plain)]   on the backend side
// `out` must be sized for the final wire bytes. The pigeon_channel
// supplied performs the AEAD encryption in place after the optional tag
// prefix. Returns total wire bytes written, or -1 on error.
int pigeon_encode_datagram(pigeon_channel *ch,
                           bool is_backend, uint32_t client_tag,
                           uint64_t channel_id,
                           const uint8_t *payload, size_t payload_len,
                           uint8_t *out, size_t out_len);

// Decode a datagram in the post-T22 framing. On the backend side strips
// the 4-byte tag prefix first and returns it via *client_tag. AEAD-
// decrypts using the supplied channel. On success writes the channel id
// to *channel_id and the application payload to `payload_buf`; returns
// the payload length. -1 on AEAD failure or any framing error.
int pigeon_decode_datagram(pigeon_channel *ch,
                           bool is_backend,
                           const uint8_t *wire, size_t wire_len,
                           uint32_t *client_tag,
                           uint64_t *channel_id,
                           uint8_t *payload_buf, size_t payload_buf_len);

// --- Multi-channel session API (post-T22) ---
//
// A pigeon_session is a single peer-to-peer association: backend ↔
// one paired client (or vice-versa from the client side). It owns the
// AEAD channel derived from the PairingRecord, the role-discriminator
// (backend/client) and the relay-assigned clientTag (backend side
// only), plus a small fixed-size table of pre-declared datagram
// channels (name → varint id, agreed by both peers up-front). Streams
// are opened on demand and live as long as the underlying transport
// stream.
//
// This API does NOT include a pigeon_listener (multi-client demux on
// the backend side); that's a separate concern that needs the
// transport's accept_stream loop to dispatch by clientTag. For now,
// pigeon_session is constructed by application code that knows whether
// it's the backend or client side.

#define PIGEON_MAX_DATAGRAM_CHANNELS 16
#define PIGEON_MAX_NAME_LEN 64

typedef struct {
    char     name[PIGEON_MAX_NAME_LEN];
    uint64_t channel_id;
} pigeon_dgchannel_def;

typedef struct {
    // PairingRecord-derived AEAD channel. Encrypts every stream message
    // (after the unencrypted name-binding header) and every datagram
    // payload (before the optional clientTag prefix).
    pigeon_channel channel;

    // Role discriminator. is_backend == true means this Session is the
    // server-side half of the pair; outbound stream/datagram framing
    // includes the 4-byte clientTag prefix the relay routes by.
    bool          is_backend;
    uint32_t      client_tag;

    // Transport vtable + opaque userdata.
    pigeon_transport transport;

    // Pre-declared datagram channels.
    pigeon_dgchannel_def datagrams[PIGEON_MAX_DATAGRAM_CHANNELS];
    size_t               datagram_count;

    // Heap-allocated scratch buffers used by the per-call I/O wrappers
    // (pigeon_stream_send/recv, pigeon_datagram_send/recv). Sized at
    // PIGEON_MAX_MSG + AEAD overhead (max 64 bytes for tag + nonce
    // prefix + 4-byte clientTag prefix on datagrams). Allocated lazily
    // on first send/recv; freed by pigeon_session_close.
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
                        bool is_backend,
                        uint32_t client_tag,
                        const pigeon_dgchannel_def *datagrams,
                        size_t datagram_count);

// Open a fresh named stream toward the peer. Writes the
// [optional 4-byte tag][varint name-len][name] header as the first
// message on the new stream. Returns 0 on success.
int pigeon_session_open_stream(pigeon_session *s,
                               const char *name,
                               pigeon_stream *out_stream);

// Look up a pre-declared datagram channel by name. Returns 0 on
// success and populates *out, -1 if the name isn't in the session's
// datagram-channel list.
int pigeon_session_get_datagram(pigeon_session *s,
                                const char *name,
                                pigeon_datagram *out);

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

// Send one datagram on the named channel. Wraps with the post-T22
// framing (AEAD([varint id][payload]) + optional 4-byte tag prefix on
// backend side). Returns 0 on success.
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

#endif // PIGEON_H
