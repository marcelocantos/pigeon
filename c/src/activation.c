// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

#include "pigeon/activation.h"

// pigeon.h gives us pigeon_transport / pigeon_stream_handle /
// pigeon_pairing_record. session_gen.h gives us the SessionMachine
// state / event constants and the per-side machine struct
// definitions. Post-T32.1's protogen rename of the COUNT sentinels,
// pairingceremony_gen.h (transitively via pigeon.h) and
// session_gen.h coexist in one TU without enumerator collisions.

#include <string.h>

#include "pigeon/pigeon.h"
#include "pigeon/session_gen.h"

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

    // Drive recv_auth_request → AuthCheck.
    if (pigeon_backend_step(machine, PIGEON_SESSION_EVENT_RECV_AUTH_REQUEST) <= 0) {
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

    if (pigeon_client_step(machine, PIGEON_SESSION_EVENT_RECV_AUTH_OK) <= 0) {
        return -1;
    }
    if (machine->state != PIGEON_CLIENT_SESSION_ACTIVE) return -1;
    return 0;
}
