// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// High-level ngtcp2 bring-up wrappers for the multi-channel API:
//
//   * pigeon_register: spins up PIGEON_ROLE_REGISTER_MUX and hands the
//     transport to pigeon_listener_init.
//   * pigeon_connect:  spins up PIGEON_ROLE_CONNECT, binds the primary
//     QUIC stream as a multi-channel slot, and hands it to
//     pigeon_connect_on_transport.
//
// Lives in its own compilation unit because both wrappers depend on
// pigeon_ngtcp2_transport (which links against ngtcp2 + quictls /
// libsodium). The amalgamated build (dist/pigeon.c) does NOT include
// this file; the ngtcp2 transport is opt-in at link time (see
// test-c-ngtcp2 in the Makefile).
//
// Memory model: on success, the higher-level handle takes ownership of
// the ngtcp2 transport struct (also heap-allocated here) and tears it
// down inside the corresponding *_close. The token / instance_id
// strings passed in are copied during ngtcp2 init; the caller can
// free them immediately after the call returns.

#include "pigeon/ngtcp2_transport.h"
#include "pigeon/pigeon.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

// Forward-declare the internal hook in listener.c (intentionally not
// in the public header — only transport-owning wrappers in this
// library need it).
extern void pigeon_listener_set_owned_transport(
    pigeon_listener *l,
    void *transport,
    void (*close_fn)(void *),
    void (*free_fn)(void *));

// Trampolines that adapt pigeon_ngtcp2_transport_close / free for the
// generic void* signature used by the owned-transport hook.
static void ngtcp2_close_trampoline(void *t)
{
    pigeon_ngtcp2_transport_close((pigeon_ngtcp2_transport *)t);
}

static void ngtcp2_free_trampoline(void *t)
{
    free(t);
}

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
    // Hand ownership of the ngtcp2 transport off to the listener so
    // pigeon_listener_close tears it down.
    pigeon_listener_set_owned_transport(*out_listener, tr,
                                        ngtcp2_close_trampoline,
                                        ngtcp2_free_trampoline);
    return 0;
}

// --- pigeon_connect (T32.4) ---
//
// Fused client-side bring-up: spin up PIGEON_ROLE_CONNECT, bind the
// primary stream as a multi-channel slot, and run the empty-name
// primary header + activation handshake via
// pigeon_connect_on_transport. The returned pigeon_connection owns
// the underlying ngtcp2 transport; pigeon_connect_close tears
// everything down.
//
// The session struct returned via _connect_session() shares lifetime
// with the connection — the connection's _close frees both.

struct pigeon_connection {
    pigeon_ngtcp2_transport *transport;
    pigeon_session           session;
};

int pigeon_connect(const char *relay_host,
                   const char *relay_port,
                   const char *peer_instance_id,
                   const char *device_id,
                   const char *token,
                   const pigeon_pairing_record *record,
                   const pigeon_dgchannel_def *datagrams,
                   size_t datagram_count,
                   pigeon_connection **out_conn)
{
    if (!relay_host || !relay_port || !peer_instance_id || !out_conn) return -1;

    pigeon_connection *c = (pigeon_connection *)calloc(1, sizeof(*c));
    if (!c) return -1;

    c->transport = (pigeon_ngtcp2_transport *)calloc(1, sizeof(*c->transport));
    if (!c->transport) { free(c); return -1; }

    pigeon_ngtcp2_config cfg = {
        .host         = relay_host,
        .port         = relay_port,
        .instance_id  = peer_instance_id,
        .verify_peer  = 0,
        .ca_cert_file = NULL,
        .timeout_ms   = 0,
        .role         = PIGEON_ROLE_CONNECT,
        .token        = token,
    };
    if (pigeon_ngtcp2_transport_init(c->transport, &cfg) != 0) {
        free(c->transport);
        free(c);
        return -1;
    }

    // Promote the primary QUIC stream into the multi-channel slot
    // table so pigeon_connect_on_transport can write the empty-name
    // header + run client activation against it via send_on_stream /
    // recv_on_stream.
    pigeon_stream_handle *primary =
        pigeon_ngtcp2_transport_primary_handle(c->transport);
    if (primary == NULL) {
        pigeon_ngtcp2_transport_close(c->transport);
        free(c->transport);
        free(c);
        return -1;
    }

    if (pigeon_connect_on_transport(pigeon_ngtcp2_as_transport(c->transport),
                                    primary,
                                    peer_instance_id, device_id,
                                    record,
                                    datagrams, datagram_count,
                                    &c->session) != 0) {
        pigeon_ngtcp2_transport_close(c->transport);
        free(c->transport);
        free(c);
        return -1;
    }

    *out_conn = c;
    return 0;
}

pigeon_session *pigeon_connect_session(pigeon_connection *c)
{
    if (!c) return NULL;
    return &c->session;
}

void pigeon_connect_close(pigeon_connection *c)
{
    if (!c) return;
    pigeon_session_close(&c->session);
    if (c->transport) {
        pigeon_ngtcp2_transport_close(c->transport);
        free(c->transport);
        c->transport = NULL;
    }
    free(c);
}
