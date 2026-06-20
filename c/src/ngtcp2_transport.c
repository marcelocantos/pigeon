// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// ngtcp2_transport.c — QUIC transport implementation for pigeon C client.
//
// Linking:
//   Requires libngtcp2, libngtcp2_crypto_quictls, libssl, libcrypto from
//   the vendored build (vendor/build/).  See vendor/build.sh and the
//   Makefile target build-vendor-deps.
//
// Allocation boundary:
//   pigeon code makes NO malloc/calloc calls.  All transport state lives
//   in the pigeon_ngtcp2_transport struct (caller-allocated).  ngtcp2 and
//   OpenSSL perform internal heap allocations; those are outside pigeon's
//   scope.  The datagram receive queue keeps malloc'd pointers from ngtcp2
//   callbacks until they are consumed and freed.

#include "pigeon/ngtcp2_transport.h"

#include <assert.h>
#include <errno.h>
#include <fcntl.h>
#include <netdb.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <unistd.h>

#include <arpa/inet.h>
#include <sys/select.h>
#include <sys/socket.h>
#include <sys/time.h>

// ngtcp2 headers (built from vendor/build/).
#include <ngtcp2/ngtcp2.h>
#include <ngtcp2/ngtcp2_crypto.h>
#include <ngtcp2/ngtcp2_crypto_quictls.h>

// OpenSSL headers (from vendored quictls build).
#include <openssl/err.h>
#include <openssl/rand.h>
#include <openssl/ssl.h>

// ---- Internal helpers ----

// ALPN token for the pigeon raw-QUIC protocol.
// Format: 1-byte length + ASCII bytes (no null terminator).
static const uint8_t PIGEON_ALPN[] = "\x06pigeon";

static uint64_t now_ns(void)
{
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (uint64_t)ts.tv_sec * NGTCP2_SECONDS + (uint64_t)ts.tv_nsec;
}

static void set_error(pigeon_ngtcp2_transport *t, const char *msg)
{
    snprintf(t->last_error, sizeof(t->last_error), "%s", msg);
}

// ---- Ring buffer ----

static size_t ringbuf_space(const pigeon_stream_ringbuf *rb)
{
    return sizeof(rb->buf) - rb->used;
}

static void ringbuf_push(pigeon_stream_ringbuf *rb,
                          const uint8_t *data, size_t len)
{
    size_t space = ringbuf_space(rb);
    if (len > space) {
        len = space;  // drop excess — caller should size buf adequately
    }
    for (size_t i = 0; i < len; i++) {
        rb->buf[rb->tail] = data[i];
        rb->tail = (rb->tail + 1) % sizeof(rb->buf);
    }
    rb->used += len;
}

static size_t ringbuf_pop(pigeon_stream_ringbuf *rb,
                           uint8_t *out, size_t want)
{
    size_t got = want < rb->used ? want : rb->used;
    for (size_t i = 0; i < got; i++) {
        out[i] = rb->buf[rb->head];
        rb->head = (rb->head + 1) % sizeof(rb->buf);
    }
    rb->used -= got;
    return got;
}

// ---- Datagram queue ----

static int dgram_enqueue(pigeon_ngtcp2_transport *t,
                          const uint8_t *data, size_t len)
{
    if (t->dgram_count >= 16) {
        // Queue full — drop oldest.
        free(t->dgram_data[t->dgram_head]);
        t->dgram_data[t->dgram_head] = NULL;
        t->dgram_head = (t->dgram_head + 1) % 16;
        t->dgram_count--;
    }
    uint8_t *copy = malloc(len);
    if (!copy) return -1;
    memcpy(copy, data, len);
    t->dgram_data[t->dgram_tail] = copy;
    t->dgram_len[t->dgram_tail]  = len;
    t->dgram_tail  = (t->dgram_tail + 1) % 16;
    t->dgram_count++;
    return 0;
}

static size_t dgram_dequeue(pigeon_ngtcp2_transport *t,
                             uint8_t *out, size_t out_len)
{
    if (t->dgram_count == 0) return 0;
    size_t len = t->dgram_len[t->dgram_head];
    if (len > out_len) len = out_len;
    memcpy(out, t->dgram_data[t->dgram_head], len);
    free(t->dgram_data[t->dgram_head]);
    t->dgram_data[t->dgram_head] = NULL;
    t->dgram_head = (t->dgram_head + 1) % 16;
    t->dgram_count--;
    return len;
}

// ---- ngtcp2 callback: get_conn (for crypto conn_ref) ----

static ngtcp2_conn *get_conn_cb(ngtcp2_crypto_conn_ref *ref)
{
    pigeon_ngtcp2_transport *t = ref->user_data;
    return (ngtcp2_conn *)t->conn;
}

// Find an extra-stream slot by ngtcp2 stream ID. Returns -1 if not present.
static int find_extra_slot(pigeon_ngtcp2_transport *t, int64_t stream_id)
{
    for (int i = 0; i < PIGEON_NGTCP2_MAX_EXTRA_STREAMS; i++) {
        if (t->extra_streams[i].in_use && t->extra_streams[i].stream_id == stream_id) {
            return i;
        }
    }
    return -1;
}

// Allocate an unused extra-stream slot. Returns -1 if all slots are in use.
static int alloc_extra_slot(pigeon_ngtcp2_transport *t, int64_t stream_id, bool peer_opened)
{
    for (int i = 0; i < PIGEON_NGTCP2_MAX_EXTRA_STREAMS; i++) {
        if (!t->extra_streams[i].in_use) {
            memset(&t->extra_streams[i], 0, sizeof(t->extra_streams[i]));
            t->extra_streams[i].stream_id   = stream_id;
            t->extra_streams[i].in_use      = true;
            t->extra_streams[i].peer_opened = peer_opened;
            return i;
        }
    }
    return -1;
}

// ---- ngtcp2 callback: stream receive ----

static int recv_stream_data_cb(ngtcp2_conn *conn, uint32_t flags,
                                int64_t stream_id, uint64_t offset,
                                const uint8_t *data, size_t datalen,
                                void *user_data, void *stream_user_data)
{
    (void)conn;
    (void)stream_user_data;
    (void)offset;

    pigeon_ngtcp2_transport *t = user_data;
    bool fin = (flags & NGTCP2_STREAM_DATA_FLAG_FIN) != 0;

    if (stream_id == t->stream_id) {
        if (t->primary_slot_idx >= 0) {
            // Primary has been bound as an extra slot for the post-T22
            // pigeon_session API — route there.
            pigeon_ngtcp2_stream_slot *s =
                &t->extra_streams[t->primary_slot_idx];
            ringbuf_push(&s->recv_buf, data, datalen);
            if (fin) s->fin_received = true;
        } else {
            // Legacy single-channel buffer.
            ringbuf_push(&t->recv_buf, data, datalen);
        }
    } else {
        int slot = find_extra_slot(t, stream_id);
        if (slot < 0) {
            // First time we see this peer-opened stream; allocate a slot
            // and queue it for accept_stream.
            slot = alloc_extra_slot(t, stream_id, true);
            if (slot >= 0 && t->accept_count < PIGEON_NGTCP2_MAX_EXTRA_STREAMS) {
                t->accept_queue[t->accept_tail] = slot;
                t->accept_tail = (t->accept_tail + 1) % PIGEON_NGTCP2_MAX_EXTRA_STREAMS;
                t->accept_count++;
            }
        }
        if (slot >= 0) {
            ringbuf_push(&t->extra_streams[slot].recv_buf, data, datalen);
            if (fin) t->extra_streams[slot].fin_received = true;
        }
    }

    // Notify ngtcp2 we consumed the data.
    ngtcp2_conn_extend_max_stream_offset((ngtcp2_conn *)t->conn,
                                          stream_id, datalen);
    ngtcp2_conn_extend_max_offset((ngtcp2_conn *)t->conn, datalen);
    return 0;
}

// ---- ngtcp2 callback: stream open (peer initiated) ----

static int stream_open_cb(ngtcp2_conn *conn, int64_t stream_id, void *user_data)
{
    (void)conn;
    pigeon_ngtcp2_transport *t = user_data;
    if (stream_id == t->stream_id) return 0; // primary, already tracked
    if (find_extra_slot(t, stream_id) >= 0) return 0; // already known
    int slot = alloc_extra_slot(t, stream_id, true);
    if (slot < 0) return 0; // no room — silently ignore
    if (t->accept_count < PIGEON_NGTCP2_MAX_EXTRA_STREAMS) {
        t->accept_queue[t->accept_tail] = slot;
        t->accept_tail = (t->accept_tail + 1) % PIGEON_NGTCP2_MAX_EXTRA_STREAMS;
        t->accept_count++;
    }
    return 0;
}

// ---- ngtcp2 callback: stream close ----

static int stream_close_cb(ngtcp2_conn *conn, uint32_t flags, int64_t stream_id,
                           uint64_t app_error_code, void *user_data,
                           void *stream_user_data)
{
    (void)conn; (void)flags; (void)app_error_code; (void)stream_user_data;
    pigeon_ngtcp2_transport *t = user_data;
    int slot = find_extra_slot(t, stream_id);
    if (slot >= 0) {
        t->extra_streams[slot].in_use = false;
    }
    return 0;
}

// ---- ngtcp2 callback: datagram receive ----

static int recv_datagram_cb(ngtcp2_conn *conn, uint32_t flags,
                             const uint8_t *data, size_t datalen,
                             void *user_data)
{
    (void)conn;
    (void)flags;
    pigeon_ngtcp2_transport *t = user_data;
    dgram_enqueue(t, data, datalen);
    return 0;
}

// ---- ngtcp2 callback: new bidi stream available ----

static int extend_max_local_streams_bidi_cb(ngtcp2_conn *conn,
                                             uint64_t max_streams,
                                             void *user_data)
{
    (void)max_streams;
    pigeon_ngtcp2_transport *t = user_data;
    if (t->stream_id != -1) return 0;   // already have one

    int rv = ngtcp2_conn_open_bidi_stream(conn, &t->stream_id, NULL);
    if (rv != 0 && rv != NGTCP2_ERR_STREAM_ID_BLOCKED) {
        return NGTCP2_ERR_CALLBACK_FAILURE;
    }
    return 0;
}

// ---- ngtcp2 callback: rand ----

static void rand_cb(uint8_t *dest, size_t destlen,
                    const ngtcp2_rand_ctx *rand_ctx)
{
    (void)rand_ctx;
    RAND_bytes(dest, (int)destlen);
}

// ---- ngtcp2 callback: get_new_connection_id ----

static int get_new_connection_id_cb(ngtcp2_conn *conn, ngtcp2_cid *cid,
                                     ngtcp2_stateless_reset_token *token,
                                     size_t cidlen, void *user_data)
{
    (void)conn;
    (void)user_data;
    RAND_bytes(cid->data, (int)cidlen);
    cid->datalen = cidlen;
    RAND_bytes(token->data, sizeof(token->data));
    return 0;
}

// ---- Packet I/O ----

static int send_packet(pigeon_ngtcp2_transport *t,
                       const uint8_t *data, size_t len)
{
    ssize_t n;
    do {
        n = send(t->fd, data, len, 0);
    } while (n == -1 && errno == EINTR);
    return (n == (ssize_t)len) ? 0 : -1;
}

// Flush outgoing QUIC packets (stream data + misc).
static int write_quic(pigeon_ngtcp2_transport *t)
{
    uint8_t buf[1452];
    ngtcp2_path_storage ps;
    ngtcp2_pkt_info pi;
    ngtcp2_ssize nwrite;

    ngtcp2_path_storage_zero(&ps);

    for (;;) {
        nwrite = ngtcp2_conn_write_pkt((ngtcp2_conn *)t->conn,
                                        &ps.path, &pi,
                                        buf, sizeof(buf), now_ns());
        if (nwrite < 0) {
            char msg[128];
            snprintf(msg, sizeof(msg),
                     "ngtcp2_conn_write_pkt: %s", ngtcp2_strerror((int)nwrite));
            set_error(t, msg);
            return -1;
        }
        if (nwrite == 0) break;
        if (send_packet(t, buf, (size_t)nwrite) != 0) {
            set_error(t, "send_packet failed");
            return -1;
        }
    }
    return 0;
}

// Send stream data through ngtcp2. Writes all `datalen` bytes,
// flushing complete packets as ngtcp2 produces them. Without the
// MORE flag each writev_stream call finalises whatever packet is
// in flight, so the bytes don't sit in ngtcp2's queue waiting for
// a later flush — a common gotcha that caused 🎯T32.4's connect
// path to silently drop the auth_request because the previous
// write_quic only flushed coalesced control frames, not the
// queued stream data.
static int write_stream(pigeon_ngtcp2_transport *t,
                         int64_t stream_id,
                         const uint8_t *data, size_t datalen,
                         int fin)
{
    uint8_t buf[1452];
    ngtcp2_path_storage ps;
    ngtcp2_pkt_info pi;
    ngtcp2_ssize nwrite;
    ngtcp2_ssize wdatalen;
    ngtcp2_vec datav = { .base = (uint8_t *)data, .len = datalen };
    size_t nwritten = 0;

    ngtcp2_path_storage_zero(&ps);

    while (nwritten < datalen) {
        datav.base = (uint8_t *)data + nwritten;
        datav.len  = datalen - nwritten;

        uint32_t flags = 0;
        if (fin && nwritten + datav.len >= datalen) {
            flags |= NGTCP2_WRITE_STREAM_FLAG_FIN;
        }

        nwrite = ngtcp2_conn_writev_stream((ngtcp2_conn *)t->conn,
                                            &ps.path, &pi,
                                            buf, sizeof(buf),
                                            &wdatalen,
                                            flags,
                                            stream_id,
                                            &datav, 1,
                                            now_ns());
        if (nwrite < 0) {
            char msg[128];
            snprintf(msg, sizeof(msg),
                     "ngtcp2_conn_writev_stream: %s", ngtcp2_strerror((int)nwrite));
            set_error(t, msg);
            return -1;
        }
        if (wdatalen > 0) nwritten += (size_t)wdatalen;
        if (nwrite > 0) {
            if (send_packet(t, buf, (size_t)nwrite) != 0) {
                set_error(t, "send_packet (stream) failed");
                return -1;
            }
        }
        if (wdatalen <= 0 && nwrite == 0) {
            // ngtcp2 is congestion- or flow-control-blocked. Bail
            // rather than spinning; subsequent write_quic / recv
            // cycles will retry once credit returns.
            set_error(t, "writev_stream: blocked (no progress)");
            return -1;
        }
    }
    return 0;
}

// Read and process incoming QUIC packets.  Non-blocking (drains what's
// available, then returns).
static int read_quic(pigeon_ngtcp2_transport *t)
{
    uint8_t buf[PIGEON_NGTCP2_MAX_PKT];
    struct sockaddr_storage addr;
    socklen_t addrlen;
    ssize_t nread;
    ngtcp2_path path;
    ngtcp2_pkt_info pi = {0};
    int rv;

    for (;;) {
        addrlen = sizeof(addr);
        nread = recvfrom(t->fd, buf, sizeof(buf), MSG_DONTWAIT,
                         (struct sockaddr *)&addr, &addrlen);
        if (nread < 0) {
            if (errno == EAGAIN || errno == EWOULDBLOCK) break;
            if (errno == EINTR) continue;
            set_error(t, "recvfrom failed");
            return -1;
        }

        path.local.addr    = (struct sockaddr *)&t->local_addr;
        path.local.addrlen = t->local_addrlen;
        path.remote.addr   = (struct sockaddr *)&addr;
        path.remote.addrlen = addrlen;

        rv = ngtcp2_conn_read_pkt((ngtcp2_conn *)t->conn,
                                   &path, &pi, buf, (size_t)nread, now_ns());
        if (rv != 0) {
            char msg[128];
            snprintf(msg, sizeof(msg),
                     "ngtcp2_conn_read_pkt: %s", ngtcp2_strerror(rv));
            set_error(t, msg);
            return -1;
        }
    }
    return 0;
}

// Block until data arrives on the UDP socket or timeout_ms elapses.
// Returns 1 if socket is readable, 0 on timeout, -1 on error.
static int wait_readable(int fd, int timeout_ms)
{
    fd_set rfds;
    struct timeval tv;

    FD_ZERO(&rfds);
    FD_SET(fd, &rfds);

    if (timeout_ms <= 0) timeout_ms = 10000;
    tv.tv_sec  = timeout_ms / 1000;
    tv.tv_usec = (timeout_ms % 1000) * 1000;

    int r;
    do {
        r = select(fd + 1, &rfds, NULL, NULL, &tv);
    } while (r < 0 && errno == EINTR);
    return r;
}

// Run the QUIC event loop until the QUIC handshake completes.
static int run_handshake(pigeon_ngtcp2_transport *t, int timeout_ms)
{
    uint64_t deadline = now_ns() + (uint64_t)(timeout_ms > 0 ? timeout_ms : 10000)
                        * 1000000ULL;

    while (!ngtcp2_conn_get_handshake_completed((ngtcp2_conn *)t->conn)) {
        if (write_quic(t) != 0) return -1;

        uint64_t expiry = ngtcp2_conn_get_expiry2((ngtcp2_conn *)t->conn);
        uint64_t ts     = now_ns();
        if (ts >= deadline) {
            set_error(t, "QUIC handshake timeout");
            return -1;
        }

        int wait_ms;
        if (expiry <= ts) {
            wait_ms = 1;
        } else {
            uint64_t diff = (expiry - ts) / 1000000ULL;
            uint64_t left = (deadline - ts) / 1000000ULL;
            wait_ms = (int)(diff < left ? diff : left);
            if (wait_ms <= 0) wait_ms = 1;
        }

        int r = wait_readable(t->fd, wait_ms);
        if (r < 0) {
            set_error(t, "select() failed during handshake");
            return -1;
        }
        if (r > 0 && read_quic(t) != 0) return -1;

        // Handle timer expiry.
        if (ngtcp2_conn_get_expiry2((ngtcp2_conn *)t->conn) <= now_ns()) {
            ngtcp2_conn_handle_expiry((ngtcp2_conn *)t->conn, now_ns());
        }
    }
    return 0;
}

// Wait until t->stream_id has been assigned (callback fires when server
// grants stream credit) or timeout.
static int wait_stream_open(pigeon_ngtcp2_transport *t, int timeout_ms)
{
    uint64_t deadline = now_ns() + (uint64_t)(timeout_ms > 0 ? timeout_ms : 5000)
                        * 1000000ULL;

    while (t->stream_id == -1) {
        if (write_quic(t) != 0) return -1;

        uint64_t ts = now_ns();
        if (ts >= deadline) {
            set_error(t, "timeout waiting for bidi stream");
            return -1;
        }

        uint64_t expiry  = ngtcp2_conn_get_expiry2((ngtcp2_conn *)t->conn);
        int wait_ms;
        if (expiry <= ts) {
            wait_ms = 1;
        } else {
            uint64_t diff = (expiry - ts) / 1000000ULL;
            uint64_t left = (deadline - ts) / 1000000ULL;
            wait_ms = (int)(diff < left ? diff : left);
            if (wait_ms <= 0) wait_ms = 1;
        }

        int r = wait_readable(t->fd, wait_ms);
        if (r < 0) { set_error(t, "select() failed"); return -1; }
        if (r > 0 && read_quic(t) != 0) return -1;

        if (ngtcp2_conn_get_expiry2((ngtcp2_conn *)t->conn) <= now_ns()) {
            ngtcp2_conn_handle_expiry((ngtcp2_conn *)t->conn, now_ns());
        }

        // Retry opening the stream.
        if (t->stream_id == -1) {
            int rv = ngtcp2_conn_open_bidi_stream((ngtcp2_conn *)t->conn,
                                                   &t->stream_id, NULL);
            if (rv != 0 && rv != NGTCP2_ERR_STREAM_ID_BLOCKED) {
                char msg[128];
                snprintf(msg, sizeof(msg),
                         "ngtcp2_conn_open_bidi_stream: %s", ngtcp2_strerror(rv));
                set_error(t, msg);
                return -1;
            }
        }
    }
    return 0;
}

// ---- pigeon_transport vtable functions ----

static int transport_send_stream(void *userdata,
                                  const uint8_t *data, size_t len)
{
    pigeon_ngtcp2_transport *t = userdata;
    if (t->closed || t->stream_id == -1) return -1;
    return write_stream(t, t->stream_id, data, len, 0);
}

// Read exactly `buf_len` bytes from the stream, blocking as needed.
static int transport_recv_stream(void *userdata,
                                  uint8_t *buf, size_t buf_len,
                                  size_t *out_len)
{
    pigeon_ngtcp2_transport *t = userdata;
    if (t->closed || t->stream_id == -1) return -1;

    size_t total = 0;
    while (total < buf_len) {
        size_t got = ringbuf_pop(&t->recv_buf, buf + total, buf_len - total);
        total += got;
        if (total >= buf_len) break;

        // Flush any pending writes then wait for more data.
        if (write_quic(t) != 0) return -1;

        uint64_t expiry = ngtcp2_conn_get_expiry2((ngtcp2_conn *)t->conn);
        uint64_t ts     = now_ns();
        int wait_ms = 5000;
        if (expiry > ts) {
            uint64_t diff = (expiry - ts) / 1000000ULL;
            wait_ms = (int)(diff < 5000 ? diff : 5000);
        }

        int r = wait_readable(t->fd, wait_ms);
        if (r < 0) { set_error(t, "select() in recv"); return -1; }
        if (r > 0 && read_quic(t) != 0) return -1;

        if (ngtcp2_conn_get_expiry2((ngtcp2_conn *)t->conn) <= now_ns()) {
            ngtcp2_conn_handle_expiry((ngtcp2_conn *)t->conn, now_ns());
        }
    }
    *out_len = total;
    return 0;
}

static int transport_send_datagram(void *userdata,
                                    const uint8_t *data, size_t len)
{
    pigeon_ngtcp2_transport *t = userdata;
    if (t->closed) return -1;

    // Ask ngtcp2 to wrap the datagram in a QUIC packet and send it.
    uint8_t buf[1452];
    ngtcp2_path_storage ps;
    ngtcp2_pkt_info pi;
    ngtcp2_ssize nwrite;

    ngtcp2_path_storage_zero(&ps);

    // Copy to ensure alignment.
    ngtcp2_vec datav = { .base = (uint8_t *)data, .len = len };

    nwrite = ngtcp2_conn_writev_datagram((ngtcp2_conn *)t->conn,
                                          &ps.path, &pi,
                                          buf, sizeof(buf),
                                          NULL,
                                          NGTCP2_WRITE_DATAGRAM_FLAG_NONE,
                                          0,
                                          &datav, 1,
                                          now_ns());
    if (nwrite < 0) {
        char msg[128];
        snprintf(msg, sizeof(msg),
                 "ngtcp2_conn_writev_datagram: %s", ngtcp2_strerror((int)nwrite));
        set_error(t, msg);
        return -1;
    }
    if (nwrite > 0) {
        return send_packet(t, buf, (size_t)nwrite);
    }
    return 0;
}

static int transport_recv_datagram(void *userdata,
                                    uint8_t *buf, size_t buf_len,
                                    size_t *out_len)
{
    pigeon_ngtcp2_transport *t = userdata;
    if (t->closed) return -1;

    // Drain a buffered datagram immediately if one's queued.
    if (t->dgram_count > 0) {
        *out_len = dgram_dequeue(t, buf, buf_len);
        return 0;
    }

    // Otherwise pump QUIC I/O until a datagram lands in the queue
    // or the overall deadline expires. Mirrors slot_read_n's
    // blocking-pump shape for streams — the higher-level
    // pigeon_datagram_recv treats a zero-byte read as failure, so a
    // single wait cycle that wakes spuriously (timer expiry only,
    // no datagram) must not leak through.
    uint64_t deadline = now_ns() + 5000ULL * 1000000ULL;
    while (t->dgram_count == 0) {
        if (write_quic(t) != 0) return -1;
        uint64_t ts = now_ns();
        if (ts >= deadline) {
            *out_len = 0;
            return 0; // no datagram within budget — caller retries
        }
        uint64_t expiry = ngtcp2_conn_get_expiry2((ngtcp2_conn *)t->conn);
        int wait_ms;
        if (expiry <= ts) wait_ms = 1;
        else {
            uint64_t diff = (expiry - ts) / 1000000ULL;
            uint64_t left = (deadline - ts) / 1000000ULL;
            wait_ms = (int)(diff < left ? diff : left);
            if (wait_ms <= 0) wait_ms = 1;
        }
        int r = wait_readable(t->fd, wait_ms);
        if (r < 0) { set_error(t, "select() in recv_datagram"); return -1; }
        if (r > 0 && read_quic(t) != 0) return -1;
        if (ngtcp2_conn_get_expiry2((ngtcp2_conn *)t->conn) <= now_ns()) {
            ngtcp2_conn_handle_expiry((ngtcp2_conn *)t->conn, now_ns());
        }
    }

    *out_len = dgram_dequeue(t, buf, buf_len);
    return 0;
}

// ---- Multi-stream vtable callbacks (post-T22) ----
//
// These implement the optional open_stream / accept_stream /
// send_on_stream / recv_on_stream / close_stream callbacks that the
// pigeon_session API uses for the multi-channel wire. Slot indices
// are stable for a given pigeon_ngtcp2_stream_slot pointer; the
// pigeon_stream_handle the API hands out is &t->extra_streams[i].

static int run_io_until(pigeon_ngtcp2_transport *t,
                        bool (*ready)(pigeon_ngtcp2_transport *),
                        int timeout_ms)
{
    uint64_t deadline = now_ns() + (uint64_t)(timeout_ms > 0 ? timeout_ms : 5000) * 1000000ULL;
    while (!ready(t)) {
        if (write_quic(t) != 0) return -1;
        uint64_t ts = now_ns();
        if (ts >= deadline) { set_error(t, "timeout"); return -1; }
        uint64_t expiry = ngtcp2_conn_get_expiry2((ngtcp2_conn *)t->conn);
        int wait_ms;
        if (expiry <= ts) wait_ms = 1;
        else {
            uint64_t diff = (expiry - ts) / 1000000ULL;
            uint64_t left = (deadline - ts) / 1000000ULL;
            wait_ms = (int)(diff < left ? diff : left);
            if (wait_ms <= 0) wait_ms = 1;
        }
        int r = wait_readable(t->fd, wait_ms);
        if (r < 0) { set_error(t, "select"); return -1; }
        if (r > 0 && read_quic(t) != 0) return -1;
        if (ngtcp2_conn_get_expiry2((ngtcp2_conn *)t->conn) <= now_ns()) {
            ngtcp2_conn_handle_expiry((ngtcp2_conn *)t->conn, now_ns());
        }
    }
    return 0;
}

static bool any_accept_pending(pigeon_ngtcp2_transport *t) { return t->accept_count > 0; }

static int transport_open_stream(void *userdata, pigeon_stream_handle **out_handle)
{
    pigeon_ngtcp2_transport *t = userdata;
    if (t->closed) return -1;
    int64_t sid = -1;
    int rv = ngtcp2_conn_open_bidi_stream((ngtcp2_conn *)t->conn, &sid, NULL);
    if (rv != 0) {
        // Wait for stream credit if blocked.
        if (rv == NGTCP2_ERR_STREAM_ID_BLOCKED) {
            uint64_t deadline = now_ns() + 5000ULL * 1000000ULL;
            while (rv == NGTCP2_ERR_STREAM_ID_BLOCKED && now_ns() < deadline) {
                if (write_quic(t) != 0) return -1;
                wait_readable(t->fd, 50);
                read_quic(t);
                rv = ngtcp2_conn_open_bidi_stream((ngtcp2_conn *)t->conn, &sid, NULL);
            }
        }
        if (rv != 0) { set_error(t, "open_bidi_stream failed"); return -1; }
    }
    int slot = alloc_extra_slot(t, sid, false);
    if (slot < 0) { set_error(t, "no free stream slot"); return -1; }
    *out_handle = (pigeon_stream_handle *)&t->extra_streams[slot];
    return 0;
}

static int transport_accept_stream(void *userdata, pigeon_stream_handle **out_handle)
{
    pigeon_ngtcp2_transport *t = userdata;
    if (t->closed) return -1;
    if (run_io_until(t, any_accept_pending, 0) != 0) return -1;
    int slot = t->accept_queue[t->accept_head];
    t->accept_head = (t->accept_head + 1) % PIGEON_NGTCP2_MAX_EXTRA_STREAMS;
    t->accept_count--;
    if (slot < 0 || slot >= PIGEON_NGTCP2_MAX_EXTRA_STREAMS) return -1;
    *out_handle = (pigeon_stream_handle *)&t->extra_streams[slot];
    return 0;
}

// Drain exactly want bytes from the given slot's recv ringbuf, pumping
// QUIC I/O until enough bytes arrive or the stream is closed. Used by
// the framed send/recv on-stream callbacks.
static int slot_read_n(pigeon_ngtcp2_transport *t,
                       pigeon_ngtcp2_stream_slot *s,
                       uint8_t *buf, size_t want)
{
    size_t total = 0;
    while (total < want) {
        size_t got = ringbuf_pop(&s->recv_buf, buf + total, want - total);
        total += got;
        if (total >= want) break;
        if (write_quic(t) != 0) return -1;
        uint64_t expiry = ngtcp2_conn_get_expiry2((ngtcp2_conn *)t->conn);
        uint64_t ts = now_ns();
        int wait_ms = 5000;
        if (expiry > ts) {
            uint64_t diff = (expiry - ts) / 1000000ULL;
            wait_ms = (int)(diff < 5000 ? diff : 5000);
        }
        int r = wait_readable(t->fd, wait_ms);
        if (r < 0) { set_error(t, "select on stream"); return -1; }
        if (r > 0 && read_quic(t) != 0) return -1;
        if (ngtcp2_conn_get_expiry2((ngtcp2_conn *)t->conn) <= now_ns()) {
            ngtcp2_conn_handle_expiry((ngtcp2_conn *)t->conn, now_ns());
        }
        if (s->fin_received && s->recv_buf.used == 0 && total < want) {
            // Stream closed before our read could fill the buffer.
            return -1;
        }
    }
    return 0;
}

// transport_send_on_stream wraps the payload in a 4-byte big-endian
// length prefix before writing. The matching transport_recv_on_stream
// peels the same prefix. Mirrors the message-oriented semantics of the
// loopback transport's send_on_stream/recv_on_stream and the Go
// relay's writeMessage/readMessage. The pigeon_session API expects one
// whole message per call; layering framing here keeps the higher
// layers identical between loopback and ngtcp2.
static int transport_send_on_stream(void *userdata, pigeon_stream_handle *h,
                                     const uint8_t *data, size_t len)
{
    pigeon_ngtcp2_transport *t = userdata;
    pigeon_ngtcp2_stream_slot *s = (pigeon_ngtcp2_stream_slot *)h;
    if (!s || !s->in_use || t->closed) return -1;
    if (len > PIGEON_MAX_MSG) return -1;

    // 4-byte BE length header + payload as one contiguous write —
    // matches the loopback's message-oriented send_on_stream and
    // the Go relay's writeMessage.
    size_t total = 4 + len;
    uint8_t *buf = (uint8_t *)malloc(total);
    if (!buf) return -1;
    buf[0] = (uint8_t)(len >> 24);
    buf[1] = (uint8_t)(len >> 16);
    buf[2] = (uint8_t)(len >> 8);
    buf[3] = (uint8_t)(len);
    if (len > 0) memcpy(buf + 4, data, len);
    int rv = write_stream(t, s->stream_id, buf, total, 0);
    free(buf);
    return rv;
}

static int transport_recv_on_stream(void *userdata, pigeon_stream_handle *h,
                                     uint8_t *buf, size_t buf_len, size_t *out_len)
{
    pigeon_ngtcp2_transport *t = userdata;
    pigeon_ngtcp2_stream_slot *s = (pigeon_ngtcp2_stream_slot *)h;
    if (!s || !s->in_use || t->closed) return -1;

    // Peel the 4-byte BE length prefix written by transport_send_on_stream
    // (and by the Go relay's writeMessage), then read exactly that many
    // payload bytes. Returns one whole message per call.
    uint8_t hdr[4];
    if (slot_read_n(t, s, hdr, 4) != 0) return -1;
    uint32_t payload_len = ((uint32_t)hdr[0] << 24)
                         | ((uint32_t)hdr[1] << 16)
                         | ((uint32_t)hdr[2] <<  8)
                         |  (uint32_t)hdr[3];
    if (payload_len > PIGEON_MAX_MSG) return -1;
    if ((size_t)payload_len > buf_len) return -1;
    if (payload_len > 0) {
        if (slot_read_n(t, s, buf, payload_len) != 0) return -1;
    }
    *out_len = payload_len;
    return 0;
}

static int transport_close_stream(void *userdata, pigeon_stream_handle *h)
{
    pigeon_ngtcp2_transport *t = userdata;
    pigeon_ngtcp2_stream_slot *s = (pigeon_ngtcp2_stream_slot *)h;
    if (!s || !s->in_use) return -1;
    ngtcp2_conn_shutdown_stream((ngtcp2_conn *)t->conn, 0, s->stream_id, 0);
    write_quic(t);
    s->in_use = false;
    return 0;
}

// ---- Init ----

static int make_udp_socket(pigeon_ngtcp2_transport *t,
                            const char *host, const char *port)
{
    struct addrinfo hints = {0};
    struct addrinfo *res = NULL, *rp;
    int fd = -1;

    hints.ai_family   = AF_UNSPEC;
    hints.ai_socktype = SOCK_DGRAM;

    if (getaddrinfo(host, port, &hints, &res) != 0) {
        set_error(t, "getaddrinfo failed");
        return -1;
    }

    for (rp = res; rp; rp = rp->ai_next) {
        fd = socket(rp->ai_family, rp->ai_socktype, rp->ai_protocol);
        if (fd == -1) continue;

        if (connect(fd, rp->ai_addr, rp->ai_addrlen) == 0) {
            t->remote_addrlen = rp->ai_addrlen;
            memcpy(&t->remote_addr, rp->ai_addr, rp->ai_addrlen);
            break;
        }
        close(fd);
        fd = -1;
    }
    freeaddrinfo(res);

    if (fd == -1) {
        set_error(t, "could not connect UDP socket to relay");
        return -1;
    }

    // Get local address.
    t->local_addrlen = sizeof(t->local_addr);
    if (getsockname(fd, (struct sockaddr *)&t->local_addr,
                    &t->local_addrlen) != 0) {
        close(fd);
        set_error(t, "getsockname failed");
        return -1;
    }

    t->fd = fd;
    return 0;
}

static int init_ssl(pigeon_ngtcp2_transport *t, const pigeon_ngtcp2_config *cfg)
{
    SSL_CTX *ssl_ctx = SSL_CTX_new(TLS_client_method());
    if (!ssl_ctx) {
        set_error(t, "SSL_CTX_new failed");
        return -1;
    }

    if (ngtcp2_crypto_quictls_configure_client_context(ssl_ctx) != 0) {
        SSL_CTX_free(ssl_ctx);
        set_error(t, "ngtcp2_crypto_quictls_configure_client_context failed");
        return -1;
    }

    if (!cfg->verify_peer) {
        SSL_CTX_set_verify(ssl_ctx, SSL_VERIFY_NONE, NULL);
    } else if (cfg->ca_cert_file) {
        if (SSL_CTX_load_verify_locations(ssl_ctx, cfg->ca_cert_file, NULL) != 1) {
            SSL_CTX_free(ssl_ctx);
            set_error(t, "SSL_CTX_load_verify_locations failed");
            return -1;
        }
    }

    SSL *ssl = SSL_new(ssl_ctx);
    if (!ssl) {
        SSL_CTX_free(ssl_ctx);
        set_error(t, "SSL_new failed");
        return -1;
    }

    // Wire up the ngtcp2 conn_ref so TLS callbacks can reach our conn.
    ngtcp2_crypto_conn_ref *ref = (ngtcp2_crypto_conn_ref *)t->conn_ref;
    ref->get_conn  = get_conn_cb;
    ref->user_data = t;
    SSL_set_app_data(ssl, ref);

    SSL_set_connect_state(ssl);

    // Set ALPN: 1-byte length prefix + "pigeon".
    SSL_set_alpn_protos(ssl, PIGEON_ALPN, sizeof(PIGEON_ALPN) - 1);

    // SNI hostname (required for TLS even with IP addresses in non-strict mode).
    // Only set for non-numeric hosts.
    if (cfg->host &&
        inet_addr(cfg->host) == (in_addr_t)-1) {  // crude non-IP check
        SSL_set_tlsext_host_name(ssl, cfg->host);
    }

    t->ssl_ctx = ssl_ctx;
    t->ssl     = ssl;
    return 0;
}

static int init_quic(pigeon_ngtcp2_transport *t)
{
    ngtcp2_path path = {
        .local = {
            .addr    = (struct sockaddr *)&t->local_addr,
            .addrlen = t->local_addrlen,
        },
        .remote = {
            .addr    = (struct sockaddr *)&t->remote_addr,
            .addrlen = t->remote_addrlen,
        },
    };

    ngtcp2_callbacks callbacks = {
        .client_initial              = ngtcp2_crypto_client_initial_cb,
        .recv_crypto_data            = ngtcp2_crypto_recv_crypto_data_cb,
        .encrypt                     = ngtcp2_crypto_encrypt_cb,
        .decrypt                     = ngtcp2_crypto_decrypt_cb,
        .hp_mask                     = ngtcp2_crypto_hp_mask_cb,
        .recv_retry                  = ngtcp2_crypto_recv_retry_cb,
        .rand                        = rand_cb,
        .get_new_connection_id2      = get_new_connection_id_cb,
        .update_key                  = ngtcp2_crypto_update_key_cb,
        .delete_crypto_aead_ctx      = ngtcp2_crypto_delete_crypto_aead_ctx_cb,
        .delete_crypto_cipher_ctx    = ngtcp2_crypto_delete_crypto_cipher_ctx_cb,
        .version_negotiation         = ngtcp2_crypto_version_negotiation_cb,
        .get_path_challenge_data2    = ngtcp2_crypto_get_path_challenge_data2_cb,
        .recv_stream_data            = recv_stream_data_cb,
        .recv_datagram               = recv_datagram_cb,
        .extend_max_local_streams_bidi = extend_max_local_streams_bidi_cb,
        .stream_open                 = stream_open_cb,
        .stream_close                = stream_close_cb,
    };

    ngtcp2_cid dcid, scid;
    dcid.datalen = NGTCP2_MIN_INITIAL_DCIDLEN;
    RAND_bytes(dcid.data, (int)dcid.datalen);
    scid.datalen = 8;
    RAND_bytes(scid.data, (int)scid.datalen);

    ngtcp2_settings settings;
    ngtcp2_settings_default(&settings);
    settings.initial_ts     = now_ns();
    settings.max_tx_udp_payload_size = 1452;
    // Uncomment to enable verbose protocol logging:
    // settings.log_printf = ...;

    ngtcp2_transport_params params;
    ngtcp2_transport_params_default(&params);
    // initial_max_streams_bidi tells the PEER how many bidi streams
    // it may open toward us. The relay needs to open one backend-
    // side bidi stream per connecting client (the per-client primary
    // it bridges) plus one per named sub-stream the client opens.
    // 64 covers a comfortable headroom over PIGEON_NGTCP2_MAX_EXTRA_STREAMS=16.
    params.initial_max_streams_bidi      = 64;
    params.initial_max_streams_uni       = 3;
    params.initial_max_stream_data_bidi_local  = 256 * 1024;
    params.initial_max_stream_data_bidi_remote = 256 * 1024;
    params.initial_max_data              = 1024 * 1024;
    params.max_datagram_frame_size       = 1200;   // enable datagrams

    ngtcp2_conn *conn = NULL;
    int rv = ngtcp2_conn_client_new(&conn, &dcid, &scid, &path,
                                     NGTCP2_PROTO_VER_V1,
                                     &callbacks, &settings, &params,
                                     NULL, t);
    if (rv != 0) {
        char msg[128];
        snprintf(msg, sizeof(msg),
                 "ngtcp2_conn_client_new: %s", ngtcp2_strerror(rv));
        set_error(t, msg);
        return -1;
    }

    ngtcp2_conn_set_tls_native_handle(conn, t->ssl);
    t->conn = conn;
    return 0;
}

// Build the pigeon greeting for this transport's role using the
// protogen-generated relay_greeting encoders (T45 remote-Listen L1).
//
// CONNECT:   "connect:<instance_id>"
// REGISTER:  "register[:<token>[:<instance_id>]]"
// LISTEN:    "listen[:<token>[:<instance_id>]]"
//
// The register/listen encoders collapse the optional token/instance_id
// suffix exactly as the Go EncodeRelayGreeting{Register,Listen} do.
// Returns the message length on success, or -1 on overflow / bad role.
static int build_handshake(const pigeon_ngtcp2_transport *t,
                           char *out, size_t out_len)
{
    const char *token = (t->token[0] != '\0') ? t->token : NULL;
    const char *id    = (t->instance_id[0] != '\0') ? t->instance_id : NULL;
    int n;

    switch (t->role) {
    case PIGEON_ROLE_CONNECT:
        n = pigeon_wire_relay_greeting_encode_connect(
                t->instance_id, strlen(t->instance_id),
                (uint8_t *)out, out_len);
        break;
    case PIGEON_ROLE_REGISTER:
        n = pigeon_wire_relay_greeting_encode_register(
                token, id, (uint8_t *)out, out_len);
        break;
    case PIGEON_ROLE_LISTEN:
        n = pigeon_wire_relay_greeting_encode_listen(
                token, id, (uint8_t *)out, out_len);
        break;
    default:
        return -1;
    }

    if (n < 0 || (size_t)n >= out_len) return -1;
    return n;
}

// Send the pigeon handshake as a length-prefixed message on the established
// bidi stream. Frame: 4-byte BE length + payload (matches Go writeMessage).
static int send_pigeon_handshake(pigeon_ngtcp2_transport *t)
{
    char msg[256];
    int msg_len = build_handshake(t, msg, sizeof(msg));
    if (msg_len < 0) {
        set_error(t, "handshake too long");
        return -1;
    }

    uint8_t frame[4 + sizeof(msg)];
    size_t payload_len = (size_t)msg_len;
    frame[0] = (uint8_t)(payload_len >> 24);
    frame[1] = (uint8_t)(payload_len >> 16);
    frame[2] = (uint8_t)(payload_len >>  8);
    frame[3] = (uint8_t)(payload_len);
    memcpy(frame + 4, msg, payload_len);

    return write_stream(t, t->stream_id, frame, 4 + payload_len, 0);
}

// Read the relay's framed greeting ack from the primary stream (still
// the legacy t->recv_buf — must run before the primary is promoted to
// a multi-channel slot). Under T45 every role gets a reply:
//   REGISTER → the assigned instance ID (stored into t->instance_id),
//   LISTEN   → the instance-ID ack (consumed),
//   CONNECT  → an "ok" ack confirming the end-to-end bridge (consumed).
// Returns 0 on success, -1 on overflow / transport error.
static int read_greeting_ack(pigeon_ngtcp2_transport *t)
{
    uint8_t hdr[4];
    size_t got = 0;
    if (transport_recv_stream(t, hdr, 4, &got) != 0 || got != 4) {
        set_error(t, "read greeting ack length");
        return -1;
    }
    uint32_t plen = ((uint32_t)hdr[0] << 24) | ((uint32_t)hdr[1] << 16)
                  | ((uint32_t)hdr[2] << 8)  |  (uint32_t)hdr[3];
    if (plen >= sizeof(t->instance_id)) {
        set_error(t, "greeting ack too long");
        return -1;
    }
    uint8_t payload[64];
    got = 0;
    if (plen > 0 &&
        (transport_recv_stream(t, payload, plen, &got) != 0 || got != plen)) {
        set_error(t, "read greeting ack payload");
        return -1;
    }
    if (t->role == PIGEON_ROLE_REGISTER) {
        // The relay assigns (or echoes) the instance ID; adopt it so
        // the listener and listen dials use the canonical value.
        memcpy(t->instance_id, payload, plen);
        t->instance_id[plen] = '\0';
    }
    // LISTEN / CONNECT acks are consumed for their gating effect only.
    return 0;
}

// ---- Public API ----

int pigeon_ngtcp2_transport_init(pigeon_ngtcp2_transport *t,
                                  const pigeon_ngtcp2_config *cfg)
{
    if (!t || !cfg || !cfg->host || !cfg->port) {
        if (t) set_error(t, "invalid arguments");
        return -1;
    }
    // CONNECT and LISTEN require an instance_id (the peer / backend to
    // reach); REGISTER accepts NULL/"" (the relay assigns one).
    if ((cfg->role == PIGEON_ROLE_CONNECT || cfg->role == PIGEON_ROLE_LISTEN) &&
        (!cfg->instance_id || cfg->instance_id[0] == '\0')) {
        if (t) set_error(t, "instance_id required for CONNECT/LISTEN role");
        return -1;
    }

    memset(t, 0, sizeof(*t));
    t->fd               = -1;
    t->stream_id        = -1;
    t->primary_slot_idx = -1;

    // Retain config strings.
    snprintf(t->host, sizeof(t->host), "%s", cfg->host);
    snprintf(t->port, sizeof(t->port), "%s", cfg->port);
    if (cfg->instance_id) {
        snprintf(t->instance_id, sizeof(t->instance_id), "%s", cfg->instance_id);
    }
    if (cfg->token) {
        snprintf(t->token, sizeof(t->token), "%s", cfg->token);
    }
    t->verify_peer = cfg->verify_peer;
    t->role        = cfg->role;

    // 1. UDP socket.
    if (make_udp_socket(t, cfg->host, cfg->port) != 0) return -1;

    // 2. TLS (SSL_CTX + SSL).
    if (init_ssl(t, cfg) != 0) {
        close(t->fd); t->fd = -1;
        return -1;
    }

    // 3. QUIC connection object.
    if (init_quic(t) != 0) {
        SSL_free((SSL *)t->ssl);
        SSL_CTX_free((SSL_CTX *)t->ssl_ctx);
        close(t->fd); t->fd = -1;
        return -1;
    }

    // 4. Run QUIC handshake.
    int timeout_ms = cfg->timeout_ms > 0 ? cfg->timeout_ms : 10000;
    if (run_handshake(t, timeout_ms) != 0) {
        pigeon_ngtcp2_transport_close(t);
        return -1;
    }

    // 5. Open bidirectional stream (may already be open via callback).
    if (t->stream_id == -1) {
        int rv = ngtcp2_conn_open_bidi_stream((ngtcp2_conn *)t->conn,
                                               &t->stream_id, NULL);
        if (rv != 0 && rv != NGTCP2_ERR_STREAM_ID_BLOCKED) {
            char msg[128];
            snprintf(msg, sizeof(msg),
                     "ngtcp2_conn_open_bidi_stream: %s", ngtcp2_strerror(rv));
            set_error(t, msg);
            pigeon_ngtcp2_transport_close(t);
            return -1;
        }
        if (rv == NGTCP2_ERR_STREAM_ID_BLOCKED) {
            if (wait_stream_open(t, 5000) != 0) {
                pigeon_ngtcp2_transport_close(t);
                return -1;
            }
        }
    }

    // 6. Send pigeon protocol greeting.
    if (send_pigeon_handshake(t) != 0) {
        pigeon_ngtcp2_transport_close(t);
        return -1;
    }
    if (write_quic(t) != 0) {
        pigeon_ngtcp2_transport_close(t);
        return -1;
    }
    t->handshake_done = 1;

    // 6b. Read the relay's framed greeting ack on the primary (T45):
    //     REGISTER → assigned instance ID, LISTEN/CONNECT → ack. Must
    //     happen before the primary is promoted to a multi-channel slot
    //     (it reads the legacy t->recv_buf).
    if (read_greeting_ack(t) != 0) {
        pigeon_ngtcp2_transport_close(t);
        return -1;
    }

    // 7. Wire up the vtable.
    t->transport.userdata       = t;
    t->transport.send_stream    = transport_send_stream;
    t->transport.recv_stream    = transport_recv_stream;
    t->transport.send_datagram  = transport_send_datagram;
    t->transport.recv_datagram  = transport_recv_datagram;
    t->transport.open_stream    = transport_open_stream;
    t->transport.accept_stream  = transport_accept_stream;
    t->transport.send_on_stream = transport_send_on_stream;
    t->transport.recv_on_stream = transport_recv_on_stream;
    t->transport.close_stream   = transport_close_stream;

    // Mark all extra-stream slots as free.
    for (int i = 0; i < PIGEON_NGTCP2_MAX_EXTRA_STREAMS; i++) {
        t->extra_streams[i].stream_id = -1;
        t->extra_streams[i].in_use    = false;
    }

    return 0;
}

void pigeon_ngtcp2_transport_close(pigeon_ngtcp2_transport *t)
{
    if (!t || t->closed) return;
    t->closed = 1;

    if (t->conn) {
        // Send a CONNECTION_CLOSE packet.
        uint8_t buf[1280];
        ngtcp2_path_storage ps;
        ngtcp2_pkt_info pi;
        ngtcp2_ccerr ccerr;
        ngtcp2_ccerr_default(&ccerr);
        ngtcp2_path_storage_zero(&ps);
        ngtcp2_ssize n = ngtcp2_conn_write_connection_close(
            (ngtcp2_conn *)t->conn, &ps.path, &pi,
            buf, sizeof(buf), &ccerr, now_ns());
        if (n > 0 && t->fd >= 0) {
            send(t->fd, buf, (size_t)n, 0);
        }
        ngtcp2_conn_del((ngtcp2_conn *)t->conn);
        t->conn = NULL;
    }

    if (t->ssl) {
        SSL_free((SSL *)t->ssl);
        t->ssl = NULL;
    }
    if (t->ssl_ctx) {
        SSL_CTX_free((SSL_CTX *)t->ssl_ctx);
        t->ssl_ctx = NULL;
    }
    if (t->fd >= 0) {
        close(t->fd);
        t->fd = -1;
    }

    // Free any buffered datagrams.
    for (int i = 0; i < t->dgram_count; i++) {
        int idx = (t->dgram_head + i) % 16;
        free(t->dgram_data[idx]);
        t->dgram_data[idx] = NULL;
    }
    t->dgram_count = 0;
}

pigeon_stream_handle *
pigeon_ngtcp2_transport_primary_handle(pigeon_ngtcp2_transport *t)
{
    if (!t || t->stream_id < 0) return NULL;
    // Idempotent — return the previously-allocated slot.
    if (t->primary_slot_idx >= 0) {
        return (pigeon_stream_handle *)
               &t->extra_streams[t->primary_slot_idx];
    }
    // Bind the primary into an extra-streams slot so the post-T22
    // session API can treat it uniformly with peer-opened sub-streams.
    // peer_opened=false because we (the QUIC client) opened it.
    int slot = alloc_extra_slot(t, t->stream_id, false);
    if (slot < 0) return NULL;
    t->primary_slot_idx = slot;
    // Migrate any bytes already buffered on the legacy primary
    // recv_buf into the new slot — those bytes arrived before the
    // primary was promoted to a multi-channel slot and would
    // otherwise be lost on the first recv_on_stream call.
    while (t->recv_buf.used > 0) {
        uint8_t scratch[256];
        size_t want = sizeof(scratch);
        if (want > t->recv_buf.used) want = t->recv_buf.used;
        size_t got = ringbuf_pop(&t->recv_buf, scratch, want);
        ringbuf_push(&t->extra_streams[slot].recv_buf, scratch, got);
    }
    return (pigeon_stream_handle *)&t->extra_streams[slot];
}
