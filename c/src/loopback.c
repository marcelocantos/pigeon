// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

#include "pigeon/loopback.h"

#include <stdlib.h>
#include <string.h>

#define LOOP_MAX_STREAMS 32
#define LOOP_MAX_PENDING 64

typedef struct loop_stream {
    int id;
    uint8_t *msgs[LOOP_MAX_PENDING]; // heap-allocated, sized PIGEON_MAX_MSG
    size_t   msg_lens[LOOP_MAX_PENDING];
    int      msg_head, msg_tail, msg_count;
    bool     in_use;
    bool     accepted;
} loop_stream;

struct pigeon_loopback_endpoint {
    loop_stream  streams[LOOP_MAX_STREAMS];

    // Inbound datagram ringbuffer.
    uint8_t *dgrams[LOOP_MAX_PENDING];   // heap-allocated, sized PIGEON_MAX_MSG + 64
    size_t   dgram_lens[LOOP_MAX_PENDING];
    int      dgram_head, dgram_tail, dgram_count;

    // Stream IDs awaiting accept.
    int accept_queue[LOOP_MAX_STREAMS];
    int accept_head, accept_tail, accept_count;

    struct pigeon_loopback_endpoint *peer;
};

static void free_stream(loop_stream *s)
{
    for (int i = 0; i < LOOP_MAX_PENDING; i++) {
        free(s->msgs[i]);
        s->msgs[i] = NULL;
    }
}

static loop_stream *lb_alloc_stream(pigeon_loopback_endpoint *e)
{
    for (int i = 0; i < LOOP_MAX_STREAMS; i++) {
        if (!e->streams[i].in_use) {
            free_stream(&e->streams[i]);
            memset(&e->streams[i], 0, sizeof(e->streams[i]));
            e->streams[i].in_use = true;
            e->streams[i].id = i;
            return &e->streams[i];
        }
    }
    return NULL;
}

static int lb_open_stream(void *ud, pigeon_stream_handle **out)
{
    pigeon_loopback_endpoint *e = (pigeon_loopback_endpoint *)ud;
    loop_stream *me = lb_alloc_stream(e);
    if (!me) return -1;
    pigeon_loopback_endpoint *p = e->peer;
    if (!p) return -1;
    if (p->streams[me->id].in_use) return -1;
    free_stream(&p->streams[me->id]);
    memset(&p->streams[me->id], 0, sizeof(p->streams[me->id]));
    p->streams[me->id].in_use = true;
    p->streams[me->id].id = me->id;
    p->accept_queue[p->accept_tail] = me->id;
    p->accept_tail = (p->accept_tail + 1) % LOOP_MAX_STREAMS;
    p->accept_count++;
    *out = (pigeon_stream_handle *)me;
    return 0;
}

static int lb_accept_stream(void *ud, pigeon_stream_handle **out)
{
    pigeon_loopback_endpoint *e = (pigeon_loopback_endpoint *)ud;
    if (e->accept_count == 0) return -1;
    int id = e->accept_queue[e->accept_head];
    e->accept_head = (e->accept_head + 1) % LOOP_MAX_STREAMS;
    e->accept_count--;
    if (id < 0 || id >= LOOP_MAX_STREAMS) return -1;
    if (!e->streams[id].in_use) return -1;
    e->streams[id].accepted = true;
    *out = (pigeon_stream_handle *)&e->streams[id];
    return 0;
}

static int lb_send_on_stream(void *ud, pigeon_stream_handle *h,
                             const uint8_t *data, size_t len)
{
    pigeon_loopback_endpoint *e = (pigeon_loopback_endpoint *)ud;
    loop_stream *me = (loop_stream *)h;
    if (!me || !me->in_use) return -1;
    if (!e->peer) return -1;
    loop_stream *peer = &e->peer->streams[me->id];
    if (!peer->in_use) return -1;
    if (peer->msg_count >= LOOP_MAX_PENDING) return -1;
    if (len > PIGEON_MAX_MSG) return -1;
    if (!peer->msgs[peer->msg_tail]) {
        peer->msgs[peer->msg_tail] = (uint8_t *)malloc(PIGEON_MAX_MSG);
        if (!peer->msgs[peer->msg_tail]) return -1;
    }
    memcpy(peer->msgs[peer->msg_tail], data, len);
    peer->msg_lens[peer->msg_tail] = len;
    peer->msg_tail = (peer->msg_tail + 1) % LOOP_MAX_PENDING;
    peer->msg_count++;
    return 0;
}

static int lb_recv_on_stream(void *ud, pigeon_stream_handle *h,
                             uint8_t *buf, size_t buf_len, size_t *out_len)
{
    (void)ud;
    loop_stream *me = (loop_stream *)h;
    if (!me || !me->in_use) return -1;
    if (me->msg_count == 0) return -1;
    size_t n = me->msg_lens[me->msg_head];
    if (n > buf_len) return -1;
    memcpy(buf, me->msgs[me->msg_head], n);
    me->msg_head = (me->msg_head + 1) % LOOP_MAX_PENDING;
    me->msg_count--;
    *out_len = n;
    return 0;
}

static int lb_close_stream(void *ud, pigeon_stream_handle *h)
{
    (void)ud;
    loop_stream *me = (loop_stream *)h;
    if (me) {
        free_stream(me);
        me->in_use = false;
    }
    return 0;
}

static int lb_send_datagram(void *ud, const uint8_t *data, size_t len)
{
    pigeon_loopback_endpoint *e = (pigeon_loopback_endpoint *)ud;
    if (!e->peer) return -1;
    pigeon_loopback_endpoint *p = e->peer;
    if (p->dgram_count >= LOOP_MAX_PENDING) return -1;
    if (len > PIGEON_MAX_MSG + 64) return -1;
    if (!p->dgrams[p->dgram_tail]) {
        p->dgrams[p->dgram_tail] = (uint8_t *)malloc(PIGEON_MAX_MSG + 64);
        if (!p->dgrams[p->dgram_tail]) return -1;
    }
    memcpy(p->dgrams[p->dgram_tail], data, len);
    p->dgram_lens[p->dgram_tail] = len;
    p->dgram_tail = (p->dgram_tail + 1) % LOOP_MAX_PENDING;
    p->dgram_count++;
    return 0;
}

static int lb_recv_datagram(void *ud, uint8_t *buf, size_t buf_len, size_t *out_len)
{
    pigeon_loopback_endpoint *e = (pigeon_loopback_endpoint *)ud;
    if (e->dgram_count == 0) return -1;
    size_t n = e->dgram_lens[e->dgram_head];
    if (n > buf_len) return -1;
    memcpy(buf, e->dgrams[e->dgram_head], n);
    e->dgram_head = (e->dgram_head + 1) % LOOP_MAX_PENDING;
    e->dgram_count--;
    *out_len = n;
    return 0;
}

pigeon_loopback_endpoint *pigeon_loopback_new(void)
{
    return (pigeon_loopback_endpoint *)calloc(1, sizeof(pigeon_loopback_endpoint));
}

void pigeon_loopback_pair(pigeon_loopback_endpoint *a,
                          pigeon_loopback_endpoint *b)
{
    if (a) a->peer = b;
    if (b) b->peer = a;
}

void pigeon_loopback_fill_transport(pigeon_transport *t,
                                    pigeon_loopback_endpoint *e)
{
    memset(t, 0, sizeof(*t));
    t->userdata        = e;
    t->open_stream     = lb_open_stream;
    t->accept_stream   = lb_accept_stream;
    t->send_on_stream  = lb_send_on_stream;
    t->recv_on_stream  = lb_recv_on_stream;
    t->close_stream    = lb_close_stream;
    t->send_datagram   = lb_send_datagram;
    t->recv_datagram   = lb_recv_datagram;
}

int pigeon_loopback_accept_with_header(pigeon_loopback_endpoint *e,
                                       pigeon_stream_handle **out_handle,
                                       uint8_t *hdr, size_t hdr_len,
                                       size_t *hdr_out_len)
{
    pigeon_stream_handle *h = NULL;
    if (lb_accept_stream(e, &h) != 0) return -1;
    size_t n = 0;
    if (lb_recv_on_stream(e, h, hdr, hdr_len, &n) != 0) return -1;
    *out_handle = h;
    if (hdr_out_len) *hdr_out_len = n;
    return 0;
}

void pigeon_loopback_free(pigeon_loopback_endpoint *e)
{
    if (!e) return;
    for (int i = 0; i < LOOP_MAX_STREAMS; i++) {
        free_stream(&e->streams[i]);
    }
    for (int i = 0; i < LOOP_MAX_PENDING; i++) {
        free(e->dgrams[i]);
    }
    free(e);
}
