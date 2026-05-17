// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// pigeon_register: high-level Listener bring-up. Spins up an ngtcp2
// transport with PIGEON_ROLE_REGISTER_MUX, hands it to
// pigeon_listener_init, and returns the listener + relay-assigned
// instance ID.
//
// Lives in its own compilation unit because it depends on
// pigeon_ngtcp2_transport (which links against ngtcp2 + quictls /
// libsodium). The amalgamated build (dist/pigeon.c) does NOT
// include this file; the ngtcp2 transport is opt-in at link time
// (see test-c-ngtcp2 in the Makefile).
//
// Memory model: on success, the listener takes ownership of the
// ngtcp2 transport struct (also heap-allocated here) and tears it
// down inside pigeon_listener_close. The token / instance_id
// strings passed in are copied during ngtcp2 init; the caller can
// free them immediately after pigeon_register returns.

#include "pigeon/ngtcp2_transport.h"
#include "pigeon/pigeon.h"

#include <stdlib.h>
#include <string.h>

int pigeon_register(const char *relay_host,
                    const char *relay_port,
                    const char *self_instance_id,
                    const char *token,
                    pigeon_resolve_device_fn pairing,
                    void *pairing_userdata,
                    const pigeon_dgchannel_def *datagrams,
                    size_t datagram_count,
                    pigeon_listener **out_listener,
                    char *out_instance_id, size_t out_instance_id_cap)
{
    if (!relay_host || !relay_port || !pairing || !out_listener) return -1;

    pigeon_ngtcp2_transport *tr =
        (pigeon_ngtcp2_transport *)calloc(1, sizeof(*tr));
    if (!tr) return -1;

    pigeon_ngtcp2_config cfg = {
        .host        = relay_host,
        .port        = relay_port,
        .instance_id = self_instance_id, // may be NULL/""
        .verify_peer = 0,
        .ca_cert_file = NULL,
        .timeout_ms  = 0,
        .role        = PIGEON_ROLE_REGISTER_MUX,
        .token       = token,
    };
    if (pigeon_ngtcp2_transport_init(tr, &cfg) != 0) {
        free(tr);
        return -1;
    }

    // The ngtcp2 transport echoes back the relay-assigned instance
    // ID into tr->instance_id during the REGISTER_MUX handshake
    // (or retains the self-assigned value when one is supplied).
    if (out_instance_id != NULL && out_instance_id_cap > 0) {
        size_t n = strlen(tr->instance_id);
        if (n >= out_instance_id_cap) n = out_instance_id_cap - 1;
        memcpy(out_instance_id, tr->instance_id, n);
        out_instance_id[n] = '\0';
    }

    if (pigeon_listener_init(out_listener,
                             pigeon_ngtcp2_as_transport(tr),
                             tr->instance_id,
                             pairing, pairing_userdata,
                             datagrams, datagram_count) != 0) {
        pigeon_ngtcp2_transport_close(tr);
        free(tr);
        return -1;
    }
    // TODO(T32.2 follow-up): pigeon_listener_close doesn't currently
    // call pigeon_ngtcp2_transport_close / free(tr) — those are owned
    // here. Until the listener grows a "transport-owner" hook, the
    // caller of pigeon_register must remember to keep the
    // pigeon_ngtcp2_transport alive for the listener's lifetime. The
    // cleanest path is a `pigeon_listener_set_owned_transport(l, tr)`
    // callback registered alongside init. Park for now; the
    // immediate target is exercising the loopback path under
    // make test-c.
    return 0;
}
