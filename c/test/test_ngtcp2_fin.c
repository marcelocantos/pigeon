// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
// Deterministic callback-to-framed-reader regression: no network or timer.
// Include the transport to exercise its actual private ngtcp2 callback and
// framed read path, rather than a duplicate model of their state transitions.
#include "../src/ngtcp2_transport.c"

static void check_frame(const uint8_t *frame, size_t size, bool fin,
                        uint32_t close_flags, bool expect_reply)
{
    pigeon_ngtcp2_transport *t = calloc(1, sizeof(*t));
    assert(t);
    t->stream_id = -1;
    pthread_mutex_init(&t->io_mu, NULL);
    pthread_cond_init(&t->io_idle, NULL);
    int idx = alloc_extra_slot(t, 4, false);
    assert(idx >= 0);
    pigeon_ngtcp2_stream_slot *s = &t->extra_streams[idx];
    ringbuf_push(&s->recv_buf, frame, size); // queued by recv_stream_data_cb
    s->fin_received = fin;                   // same pump as stream_close_cb
    assert(stream_close_cb(NULL, close_flags, 4, 0, t, NULL) == 0);
    assert(s->in_use); // a callback never disposes an application handle
    uint8_t reply[16] = {0};
    size_t n = 99;
    int rc = transport_recv_on_stream(t, (pigeon_stream_handle *)s,
                                reply, sizeof(reply), &n);
    if (expect_reply) {
        assert(rc == 0 && n == 3 && memcmp(reply, "yes", 3) == 0);
        assert(transport_recv_on_stream(t, (pigeon_stream_handle *)s,
                                  reply, sizeof(reply), &n) == -1); // EOF
    } else {
        assert(rc == -1); // partial header/payload or reset, never success
    }
    assert(!s->in_use);
    pthread_cond_destroy(&t->io_idle);
    pthread_mutex_destroy(&t->io_mu);
    free(t);
}

int main(void)
{
    const uint8_t full[] = {0, 0, 0, 3, 'y', 'e', 's'};
    check_frame(full, sizeof(full), true, 0, true);
    check_frame(full, 2, true, 0, false); // truncated prefix
    check_frame(full, 6, true, 0, false); // truncated payload
    check_frame(full, sizeof(full), true,
                NGTCP2_STREAM_CLOSE_FLAG_APP_ERROR_CODE_SET, false); // reset
    puts("ngtcp2 buffered FIN/close and reset regressions: OK");
    return 0;
}
