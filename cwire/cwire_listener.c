// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// C-side trampoline that bridges pigeon_resolve_device_fn callbacks to
// the //export'd cwireGoResolve in listener.go.
//
// The cwire_go_udata pattern is identical to the one used by
// cwire_confirm_trampoline in cwire_pigeon.c: a small heap-allocated C
// struct carries the cgo.Handle bit-pattern so the Go-side //export can
// recover it with a plain C-pointer read (no unsafe.Pointer ↔ uintptr
// round-trip that go vet would flag).

#include <stddef.h>
#include <stdint.h>

// cwire_go_udata is defined and allocated in cwire_pigeon.c; extern
// here so we can use it without re-declaring it in a header.
typedef struct cwire_go_udata { uintptr_t handle; } cwire_go_udata;

// Forward-declare the //export'd Go trampoline.
extern int cwireGoResolve(void *udata, const char *device_id, void *out_record);

int cwire_resolve_trampoline(void *udata, const char *device_id, void *out_record)
{
    return cwireGoResolve(udata, device_id, out_record);
}
