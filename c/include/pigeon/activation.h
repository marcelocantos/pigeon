// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
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

#ifndef PIGEON_ACTIVATION_H
#define PIGEON_ACTIVATION_H

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

// We deliberately do NOT #include "pigeon.h" or "session_gen.h" here.
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

// --- Wire encoders / decoders ---

// pigeon_encode_auth_request builds the binary payload for the
// client's auth_request message: [tag][uvarint device-id len][device-id].
// Returns the encoded length on success, or -1 if buf_len is too
// small.
int pigeon_encode_auth_request(const char *device_id,
                               uint8_t *buf, size_t buf_len);

// pigeon_decode_auth_request parses an auth_request payload and
// copies the device ID (NUL-terminated) into out_device_id.
// Returns 0 on success, -1 on malformed payload or oversized device
// id.
int pigeon_decode_auth_request(const uint8_t *payload, size_t payload_len,
                               char *out_device_id, size_t out_cap);

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
int pigeon_run_backend_activation(const void *transport,
                                  void *stream,
                                  pigeon_resolve_device_fn resolve,
                                  void *resolve_userdata,
                                  void *out_machine,
                                  char *out_device_id, size_t out_device_id_cap,
                                  void *out_record);

// pigeon_run_client_activation drives the client's per-client
// SessionMachine through Reconnect -> SendAuth -> SessionActive.
// Writes auth_request{device_id} to `stream`, reads the auth_ok
// reply, and on acceptance populates `out_machine`. On rejection
// returns -1 and (if out_reason is non-NULL) copies the backend's
// reason into out_reason.
//
// Returns 0 on success, -1 on transport / decode / rejection.
int pigeon_run_client_activation(const void *transport,
                                 void *stream,
                                 const char *device_id,
                                 void *out_machine,
                                 char *out_reason, size_t out_reason_cap);

#ifdef __cplusplus
}
#endif

#endif // PIGEON_ACTIVATION_H
