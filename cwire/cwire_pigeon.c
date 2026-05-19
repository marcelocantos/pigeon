// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// Wraps the amalgamated C library (dist/pigeon.c) into the cgo build
// of package cwire. We compile the amalgamation here rather than
// linking a static lib because libpigeon hasn't been packaged yet —
// once T32 step 3 finishes the multi-stream port, this file can be
// replaced with a `#cgo LDFLAGS: -lpigeon` and a binary distribution.
//
// dist/pigeon.c must be regenerated whenever the C sources change:
//   make amalgamate

#include "pigeon.c"
#include "loopback.h"

#include <string.h>
#include <stdlib.h>

// Heap-allocate a zero-initialised pigeon_session / pigeon_channel.
// Defined here (non-static) so multiple cgo .go files can share the
// helper — cgo's per-file preamble can't define static helpers more
// than once across files in the same package, and stdlib calloc() +
// sizeof(struct) sometimes confuses cgo's parser.
void *cwire_calloc_session(void)
{
    return calloc(1, sizeof(pigeon_session));
}
void *cwire_calloc_channel(void)
{
    return calloc(1, sizeof(pigeon_channel));
}

// User-data box for the Go transport bridge. We can't pass a
// cgo.Handle bit-pattern through `void *userdata` without
// unsafe.Pointer ↔ uintptr conversions that go vet (correctly)
// flags as risky. Instead we malloc a tiny C struct that carries the
// handle, hand its pointer to the C library as userdata, and the
// //export'd Go trampolines recover the handle by reading
// `udata->handle` — a normal C pointer dereference, no vet flag.
typedef struct cwire_go_udata {
    uintptr_t handle;
} cwire_go_udata;

cwire_go_udata *cwire_alloc_go_udata(uintptr_t handle)
{
    cwire_go_udata *u = (cwire_go_udata *)calloc(1, sizeof(*u));
    if (u) u->handle = handle;
    return u;
}

void cwire_free_go_udata(cwire_go_udata *u)
{
    free(u);
}

// Forward-declare the //export'd Go trampolines from gotransport.go.
// cgo emits their definitions into _cgo_export.c so they're linked
// into the same shared object. We avoid #including _cgo_export.h
// here because that header pulls in the cgo prelude (which makes
// builds slower and creates duplicate-symbol risk).
extern int cwireGoOpenStream(void *udata, pigeon_stream_handle **out);
extern int cwireGoAcceptStream(void *udata, pigeon_stream_handle **out);
extern int cwireGoSendOnStream(void *udata, pigeon_stream_handle *h,
                               const uint8_t *data, size_t len);
extern int cwireGoRecvOnStream(void *udata, pigeon_stream_handle *h,
                               uint8_t *buf, size_t buf_len, size_t *out_len);
extern int cwireGoCloseStream(void *udata, pigeon_stream_handle *h);
extern int cwireGoSendDatagram(void *udata, const uint8_t *data, size_t len);
extern int cwireGoRecvDatagram(void *udata, uint8_t *buf, size_t buf_len,
                               size_t *out_len);

// Populate a pigeon_transport vtable with the //export'd Go
// callbacks. `udata` is the cgo.Handle (uintptr cast to void*) for
// the Go transport instance. cwire/gotransport.go calls this from
// NewGoSession.
void cwire_make_go_transport(void *udata, pigeon_transport *out)
{
    memset(out, 0, sizeof(*out));
    out->userdata        = udata;
    out->open_stream     = cwireGoOpenStream;
    out->accept_stream   = cwireGoAcceptStream;
    out->send_on_stream  = cwireGoSendOnStream;
    out->recv_on_stream  = cwireGoRecvOnStream;
    out->close_stream    = cwireGoCloseStream;
    out->send_datagram   = cwireGoSendDatagram;
    out->recv_datagram   = cwireGoRecvDatagram;
}

// --- Pairing confirm trampoline ---
//
// pigeon_pair_acceptor / pigeon_pair_initiator call confirm_fn(userdata, code)
// synchronously when the confirmation code is ready. We bridge this to Go
// via a //export'd callback; userdata carries a cwire_go_udata* (same box
// pattern used by the transport bridge) holding a cgo.Handle for the Go
// closure.

extern int cwireGoConfirm(void *udata, const char *code);

int cwire_confirm_trampoline(void *udata, const char *code)
{
    return cwireGoConfirm(udata, code);
}

// cwire_resolve_trampoline (pigeon_resolve_device_fn bridge) lives in
// cwire_listener.c — both the C trampoline and the //export'd Go side
// (cwireGoResolve in listener.go) are owned by the Listener slice; the
// Connect-side RunBackendActivation helper reuses the same bridge.
