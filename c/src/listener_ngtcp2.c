// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// High-level ngtcp2 bring-up wrappers for the multi-channel API
// (T45 remote-Listen L1):
//
//   * pigeon_register: dials the register control connection
//     (PIGEON_ROLE_REGISTER) and wires an ngtcp2 listen-dialer
//     (PIGEON_ROLE_LISTEN) into pigeon_listener_init. Each accepted
//     client rides its own bridged listen connection.
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
// Memory model: the listener owns the register control transport and
// tears it down inside pigeon_listener_close. Each accepted Session
// owns its own listen transport (produced by the listen-dialer) and
// closes it inside pigeon_session_close via the owner hooks.

#include "pigeon/ngtcp2_transport.h"
#include "pigeon/pigeon.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

// Forward-declare the internal hook in listener.c (intentionally not
// in the public header — only control-transport-owning wrappers in
// this library need it).
extern void pigeon_listener_set_owned_control(
    pigeon_listener *l,
    void *transport,
    void (*close_fn)(void *),
    void (*free_fn)(void *));

extern void pigeon_listener_set_dial_userdata_free(
    pigeon_listener *l,
    void (*free_fn)(void *));

// Trampolines that adapt pigeon_ngtcp2_transport_close / free for the
// generic void* signatures used by the owned-control and listen-owner
// hooks.
static void ngtcp2_close_trampoline(void *t)
{
    pigeon_ngtcp2_transport_close((pigeon_ngtcp2_transport *)t);
}

static void ngtcp2_free_trampoline(void *t)
{
    free(t);
}

// --- Listen-dialer ---
//
// Parked as the listener's dial_userdata. Holds the relay coordinates
// needed to dial a fresh PIGEON_ROLE_LISTEN connection per accept.

typedef struct {
    char host[256];
    char port[16];
    char instance_id[64];
    char token[128];
    bool have_token;
} listen_dial_ctx;

// pigeon_listen_dialer implementation: dial one fresh listen
// connection, wait for the relay's instance-ID ack, and hand back the
// transport vtable + primary handle + owner cookie (the heap
// transport, torn down via the owner hooks).
static int ngtcp2_dial_listen(void *userdata,
                              pigeon_transport *out_transport,
                              pigeon_stream_handle **out_primary,
                              void **out_owner)
{
    listen_dial_ctx *ctx = (listen_dial_ctx *)userdata;

    pigeon_ngtcp2_transport *tr =
        (pigeon_ngtcp2_transport *)calloc(1, sizeof(*tr));
    if (!tr) return -1;

    pigeon_ngtcp2_config cfg = {
        .host         = ctx->host,
        .port         = ctx->port,
        .instance_id  = ctx->instance_id,
        .verify_peer  = 0,
        .ca_cert_file = NULL,
        .timeout_ms   = 0,
        .role         = PIGEON_ROLE_LISTEN,
        .token        = ctx->have_token ? ctx->token : NULL,
    };
    if (pigeon_ngtcp2_transport_init(tr, &cfg) != 0) {
        free(tr);
        return -1;
    }

    pigeon_stream_handle *primary =
        pigeon_ngtcp2_transport_primary_handle(tr);
    if (primary == NULL) {
        pigeon_ngtcp2_transport_close(tr);
        free(tr);
        return -1;
    }

    *out_transport = *pigeon_ngtcp2_as_transport(tr);
    *out_primary   = primary;
    *out_owner     = tr;
    return 0;
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

    // 1. Dial the register control connection. The relay echoes back
    //    the assigned instance ID (into ctrl->instance_id) and holds
    //    the connection open for the instance's lifetime.
    pigeon_ngtcp2_transport *ctrl =
        (pigeon_ngtcp2_transport *)calloc(1, sizeof(*ctrl));
    if (!ctrl) return -1;

    pigeon_ngtcp2_config ctrl_cfg = {
        .host        = relay_host,
        .port        = relay_port,
        .instance_id = self_instance_id, // may be NULL/""
        .verify_peer = 0,
        .ca_cert_file = NULL,
        .timeout_ms  = 0,
        .role        = PIGEON_ROLE_REGISTER,
        .token       = token,
    };
    if (pigeon_ngtcp2_transport_init(ctrl, &ctrl_cfg) != 0) {
        free(ctrl);
        return -1;
    }

    if (out_instance_id != NULL && out_instance_id_cap > 0) {
        size_t n = strlen(ctrl->instance_id);
        if (n >= out_instance_id_cap) n = out_instance_id_cap - 1;
        memcpy(out_instance_id, ctrl->instance_id, n);
        out_instance_id[n] = '\0';
    }

    // 2. Build the listen-dialer context. Listen dials target the
    //    relay-assigned instance ID, not the (possibly empty) requested
    //    one.
    listen_dial_ctx *dctx = (listen_dial_ctx *)calloc(1, sizeof(*dctx));
    if (!dctx) {
        pigeon_ngtcp2_transport_close(ctrl);
        free(ctrl);
        return -1;
    }
    snprintf(dctx->host, sizeof(dctx->host), "%s", relay_host);
    snprintf(dctx->port, sizeof(dctx->port), "%s", relay_port);
    snprintf(dctx->instance_id, sizeof(dctx->instance_id), "%s",
             ctrl->instance_id);
    if (token && token[0] != '\0') {
        snprintf(dctx->token, sizeof(dctx->token), "%s", token);
        dctx->have_token = true;
    }

    // 3. Build the listener around the control transport + listen-dialer.
    if (pigeon_listener_init(out_listener,
                             pigeon_ngtcp2_as_transport(ctrl),
                             ctrl->instance_id,
                             ngtcp2_dial_listen, dctx,
                             ngtcp2_close_trampoline,
                             ngtcp2_free_trampoline,
                             pairing, pairing_userdata,
                             datagrams, datagram_count) != 0) {
        free(dctx);
        pigeon_ngtcp2_transport_close(ctrl);
        free(ctrl);
        return -1;
    }

    // Hand ownership of the register control transport and the listen-
    // dialer context to the listener so pigeon_listener_close tears
    // both down.
    pigeon_listener_set_owned_control(*out_listener, ctrl,
                                      ngtcp2_close_trampoline,
                                      ngtcp2_free_trampoline);
    pigeon_listener_set_dial_userdata_free(*out_listener, free);
    return 0;
}

// --- pigeon_connect (T32.4) ---
//
// Fused client-side bring-up: spin up PIGEON_ROLE_CONNECT, bind the
// primary stream as a multi-channel slot, and run the activation
// handshake (or, in pairing mode, the empty arrival marker) via
// pigeon_connect_on_transport. The returned pigeon_connection owns the
// underlying ngtcp2 transport; pigeon_connect_close tears everything
// down.
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
    // table so pigeon_connect_on_transport can run client activation
    // (or send the pairing arrival marker) against it via
    // send_on_stream / recv_on_stream.
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
