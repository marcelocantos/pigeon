// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// ngtcp2_transport — QUIC transport for the pigeon C client library.
//
// Implements the pigeon_transport vtable using ngtcp2 (QUIC) + quictls
// (OpenSSL with QUIC extensions) as TLS backend.
//
// Protocol
// --------
// The pigeon relay server speaks raw QUIC with ALPN "pigeon".  There is
// no HTTP/3 layer — clients use a simple bespoke handshake:
//
//   1. Open a QUIC connection to the relay.
//   2. Open one bidirectional QUIC stream (stream 0).
//   3. Send a length-prefixed message: "connect:<instance-id>"
//      where length is a 4-byte big-endian uint32.
//   4. Use that stream for all subsequent pigeon_transport stream calls.
//   5. Send/receive QUIC datagrams for pigeon_transport datagram calls.
//
// Memory model
// ------------
// pigeon_ngtcp2_transport contains all state.  Callers may allocate it
// on the stack, as a static, or embedded in a larger struct — no heap
// allocation is required by pigeon's transport layer itself.
//
// ngtcp2 and OpenSSL will internally allocate memory as needed; that is
// outside pigeon's scope and is documented here as the accepted boundary.
//
// Thread safety
// -------------
// Not thread-safe.  Use from a single thread or add external locking.
//
// TLS certificate validation
// --------------------------
// By default, server certificate verification is disabled so that the
// development relay's self-signed certificate is accepted.  Pass
// verify_peer=true and set ca_cert_pem_path to a CA bundle to enable
// verification in production.

#ifndef PIGEON_NGTCP2_TRANSPORT_H
#define PIGEON_NGTCP2_TRANSPORT_H

#include "pigeon.h"

#include <stddef.h>
#include <stdint.h>
#include <sys/socket.h>

// Opaque pointer to the TLS handle (SSL*).  Avoids exposing OpenSSL types
// in the public header when building without vendored OpenSSL on the
// include path.
typedef void pigeon_ssl_handle;
typedef void pigeon_ssl_ctx_handle;
typedef void pigeon_ngtcp2_conn_handle;

// Maximum receive buffer per packet.
#define PIGEON_NGTCP2_MAX_PKT 65536

// Maximum pending send data in the stream send buffer.
#define PIGEON_NGTCP2_STREAM_BUF 65536

// Maximum pending received stream data.
#define PIGEON_NGTCP2_RECV_BUF 65536

// Configuration passed to pigeon_ngtcp2_transport_init.
typedef struct {
    const char *host;           // relay hostname or IP (required)
    const char *port;           // relay UDP port, e.g. "4433" (required)
    const char *instance_id;    // pigeon instance ID to connect to (required)
    int         verify_peer;    // 1 = verify server cert, 0 = skip (default 0)
    const char *ca_cert_file;   // path to CA bundle PEM (NULL = system default)
    int         timeout_ms;     // overall connect+handshake timeout in ms (0 = 10 000)
} pigeon_ngtcp2_config;

// Internal ring-buffer for stream receive data.
typedef struct {
    uint8_t  buf[PIGEON_NGTCP2_RECV_BUF];
    size_t   head;   // read position
    size_t   tail;   // write position
    size_t   used;   // bytes available
} pigeon_stream_ringbuf;

// All transport state.  Zero-initialize before calling pigeon_ngtcp2_transport_init.
typedef struct pigeon_ngtcp2_transport {
    // ---- vtable-compatible header (must be first) ----
    // The pigeon_transport struct is embedded at offset 0 so that a
    // pointer to pigeon_ngtcp2_transport can be cast to pigeon_transport*.
    pigeon_transport transport;     // filled in by _init

    // ---- connection state ----
    int                       fd;             // UDP socket
    struct sockaddr_storage   local_addr;
    socklen_t                 local_addrlen;
    struct sockaddr_storage   remote_addr;
    socklen_t                 remote_addrlen;

    pigeon_ssl_ctx_handle    *ssl_ctx;        // SSL_CTX*
    pigeon_ssl_handle        *ssl;            // SSL*
    pigeon_ngtcp2_conn_handle *conn;          // ngtcp2_conn*

    // ngtcp2 connection-id ref (used as app-data for TLS callbacks)
    // Declared as char[] to avoid pulling in ngtcp2 headers here.
    char                      conn_ref[64];   // ngtcp2_crypto_conn_ref

    // ---- stream state ----
    int64_t                   stream_id;      // bidi stream ID (-1 = not yet open)
    int                       stream_opened;  // 1 after handshake stream created
    int                       handshake_done; // 1 after pigeon protocol handshake sent

    // Buffered received stream data (fills from QUIC callbacks).
    pigeon_stream_ringbuf     recv_buf;

    // ---- packet I/O buffers (stack-allocated in functions, not here) ----
    // ngtcp2 write functions are called with temporary stack buffers.

    // ---- datagram receive queue ----
    // Simple fixed-depth ring of pointers to malloc'd datagrams.
    // ngtcp2/OpenSSL alloc these; pigeon frees them after delivery.
    uint8_t  *dgram_data[16];
    size_t    dgram_len[16];
    int       dgram_head;
    int       dgram_tail;
    int       dgram_count;

    // ---- error/close state ----
    int       closed;
    char      last_error[128];  // human-readable error string

    // ---- configuration (retained for reconnect / diagnostics) ----
    char      host[256];
    char      port[16];
    char      instance_id[64];
    int       verify_peer;
} pigeon_ngtcp2_transport;

// Initialize the transport and establish a QUIC connection to the relay.
//
// On success, t->transport is wired up and ready to hand to pigeon_init().
// Returns 0 on success, -1 on failure (check t->last_error for a description).
//
// Memory: all state lives in *t.  No heap allocation by pigeon code.
int pigeon_ngtcp2_transport_init(pigeon_ngtcp2_transport *t,
                                 const pigeon_ngtcp2_config *cfg);

// Close the QUIC connection and free OpenSSL/ngtcp2 objects.
// After this call, the transport must not be used.
void pigeon_ngtcp2_transport_close(pigeon_ngtcp2_transport *t);

// Convenience: return &t->transport for passing to pigeon_init().
static inline pigeon_transport *
pigeon_ngtcp2_as_transport(pigeon_ngtcp2_transport *t)
{
    return &t->transport;
}

#endif // PIGEON_NGTCP2_TRANSPORT_H
