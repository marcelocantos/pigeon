// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// JNI shim for the Pigeon C peer library (T30).
//
// This shim exposes the post-T22 pigeon_session / pigeon_stream /
// pigeon_datagram API to Kotlin/Java via JNI. Native objects are
// allocated on the C heap and referenced from the JVM via opaque jlong
// handles; the Kotlin wrappers (Session, Stream, Datagram) own them
// and free via close().
//
// Two transport flavours are supported:
//
//   1. An in-process C loopback transport, mirroring
//      c/test/test_pigeon.c::loopback_make_transport. Used by the JNI
//      unit tests to exercise the full multi-stream wire end-to-end
//      without a live ngtcp2 stack. See pigeonNewLoopbackPair below.
//
//   2. (Future) A Java-callback transport that invokes Kotlin
//      QuicTransport methods from C. Not yet implemented; this is the
//      gap that gates real Android NDK use. See T30 commit message.
//
// Build:
//   clang -shared -fPIC \
//     -I${JNI_HEADERS} -Idist -Ic/vendor/build/include \
//     -DPIGEON_CRYPTO_LIBSODIUM \
//     dist/pigeon.c c/jni/pigeon_jni.c \
//     c/vendor/build/lib/libsodium.a \
//     -o libpigeon-jni.dylib
//
// All JNI methods follow the canonical naming convention for the
// class com.marcelocantos.pigeon.jni.PigeonNative, registered in
// PigeonNative.kt with `external fun` declarations.

#include <jni.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

#include "pigeon.h"

// ----------------------------------------------------------------------------
// Loopback transport (mirrors c/test/test_pigeon.c).
// ----------------------------------------------------------------------------

#define LOOP_MAX_STREAMS 8
#define LOOP_MAX_PENDING 32

typedef struct loop_stream {
    int      id;
    uint8_t  msgs[LOOP_MAX_PENDING][PIGEON_MAX_MSG];
    size_t   msg_lens[LOOP_MAX_PENDING];
    int      msg_head, msg_tail, msg_count;
    bool     in_use;
    bool     accepted;
} loop_stream;

typedef struct loop_endpoint {
    loop_stream streams[LOOP_MAX_STREAMS];

    uint8_t dgrams[LOOP_MAX_PENDING][PIGEON_MAX_MSG + 64];
    size_t  dgram_lens[LOOP_MAX_PENDING];
    int     dgram_head, dgram_tail, dgram_count;

    int accept_queue[LOOP_MAX_STREAMS];
    int accept_head, accept_tail, accept_count;

    struct loop_endpoint *peer;
} loop_endpoint;

static loop_stream *loop_alloc_stream(loop_endpoint *e)
{
    for (int i = 0; i < LOOP_MAX_STREAMS; i++) {
        if (!e->streams[i].in_use) {
            memset(&e->streams[i], 0, sizeof(e->streams[i]));
            e->streams[i].in_use = true;
            e->streams[i].id = i;
            return &e->streams[i];
        }
    }
    return NULL;
}

static int loop_open_stream(void *ud, pigeon_stream_handle **out)
{
    loop_endpoint *e = (loop_endpoint *)ud;
    loop_stream *me = loop_alloc_stream(e);
    if (!me) return -1;
    loop_endpoint *p = e->peer;
    if (p->streams[me->id].in_use) return -1;
    memset(&p->streams[me->id], 0, sizeof(p->streams[me->id]));
    p->streams[me->id].in_use = true;
    p->streams[me->id].id = me->id;
    p->accept_queue[p->accept_tail] = me->id;
    p->accept_tail = (p->accept_tail + 1) % LOOP_MAX_STREAMS;
    p->accept_count++;
    *out = (pigeon_stream_handle *)me;
    return 0;
}

static int loop_accept_stream(void *ud, pigeon_stream_handle **out)
{
    loop_endpoint *e = (loop_endpoint *)ud;
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

static int loop_send_on_stream(void *ud, pigeon_stream_handle *h,
                               const uint8_t *data, size_t len)
{
    loop_endpoint *e = (loop_endpoint *)ud;
    loop_stream *me = (loop_stream *)h;
    if (!me->in_use) return -1;
    loop_stream *peer = &e->peer->streams[me->id];
    if (!peer->in_use) return -1;
    if (peer->msg_count >= LOOP_MAX_PENDING) return -1;
    if (len > PIGEON_MAX_MSG) return -1;
    memcpy(peer->msgs[peer->msg_tail], data, len);
    peer->msg_lens[peer->msg_tail] = len;
    peer->msg_tail = (peer->msg_tail + 1) % LOOP_MAX_PENDING;
    peer->msg_count++;
    return 0;
}

static int loop_recv_on_stream(void *ud, pigeon_stream_handle *h,
                               uint8_t *buf, size_t buf_len, size_t *out_len)
{
    (void)ud;
    loop_stream *me = (loop_stream *)h;
    if (!me->in_use) return -1;
    if (me->msg_count == 0) return -1;
    size_t n = me->msg_lens[me->msg_head];
    if (n > buf_len) return -1;
    memcpy(buf, me->msgs[me->msg_head], n);
    me->msg_head = (me->msg_head + 1) % LOOP_MAX_PENDING;
    me->msg_count--;
    *out_len = n;
    return 0;
}

static int loop_close_stream(void *ud, pigeon_stream_handle *h)
{
    (void)ud;
    loop_stream *me = (loop_stream *)h;
    me->in_use = false;
    return 0;
}

static int loop_send_datagram(void *ud, const uint8_t *data, size_t len)
{
    loop_endpoint *e = (loop_endpoint *)ud;
    loop_endpoint *p = e->peer;
    if (p->dgram_count >= LOOP_MAX_PENDING) return -1;
    if (len > sizeof(p->dgrams[0])) return -1;
    memcpy(p->dgrams[p->dgram_tail], data, len);
    p->dgram_lens[p->dgram_tail] = len;
    p->dgram_tail = (p->dgram_tail + 1) % LOOP_MAX_PENDING;
    p->dgram_count++;
    return 0;
}

static int loop_recv_datagram(void *ud, uint8_t *buf, size_t buf_len, size_t *out_len)
{
    loop_endpoint *e = (loop_endpoint *)ud;
    if (e->dgram_count == 0) return -1;
    size_t n = e->dgram_lens[e->dgram_head];
    if (n > buf_len) return -1;
    memcpy(buf, e->dgrams[e->dgram_head], n);
    e->dgram_head = (e->dgram_head + 1) % LOOP_MAX_PENDING;
    e->dgram_count--;
    *out_len = n;
    return 0;
}

static void loop_make_transport(pigeon_transport *t, loop_endpoint *e)
{
    memset(t, 0, sizeof(*t));
    t->userdata        = e;
    t->open_stream     = loop_open_stream;
    t->accept_stream   = loop_accept_stream;
    t->send_on_stream  = loop_send_on_stream;
    t->recv_on_stream  = loop_recv_on_stream;
    t->close_stream    = loop_close_stream;
    t->send_datagram   = loop_send_datagram;
    t->recv_datagram   = loop_recv_datagram;
}

// ----------------------------------------------------------------------------
// JNI utilities.
// ----------------------------------------------------------------------------

static void throw_runtime(JNIEnv *env, const char *msg)
{
    jclass cls = (*env)->FindClass(env, "java/lang/RuntimeException");
    if (cls) (*env)->ThrowNew(env, cls, msg);
}

static jbyteArray bytes_to_jba(JNIEnv *env, const uint8_t *buf, size_t len)
{
    jbyteArray arr = (*env)->NewByteArray(env, (jsize)len);
    if (!arr) return NULL;
    (*env)->SetByteArrayRegion(env, arr, 0, (jsize)len, (const jbyte *)buf);
    return arr;
}

// ----------------------------------------------------------------------------
// Crypto: keypair, derive session key, derive confirmation code.
// ----------------------------------------------------------------------------

JNIEXPORT jbyteArray JNICALL
Java_com_marcelocantos_pigeon_jni_PigeonNative_keypair(JNIEnv *env, jclass cls)
{
    (void)cls;
    pigeon_keypair kp;
    if (pigeon_generate_keypair(&kp) != 0) {
        throw_runtime(env, "pigeon_generate_keypair failed");
        return NULL;
    }
    uint8_t out[64];
    memcpy(out, kp.private_key, 32);
    memcpy(out + 32, kp.public_key, 32);
    return bytes_to_jba(env, out, 64);
}

JNIEXPORT jbyteArray JNICALL
Java_com_marcelocantos_pigeon_jni_PigeonNative_deriveSessionKey(
    JNIEnv *env, jclass cls,
    jbyteArray priv, jbyteArray peerPub, jbyteArray info)
{
    (void)cls;
    if ((*env)->GetArrayLength(env, priv) != 32 ||
        (*env)->GetArrayLength(env, peerPub) != 32) {
        throw_runtime(env, "key length must be 32");
        return NULL;
    }
    jbyte *p_priv = (*env)->GetByteArrayElements(env, priv, NULL);
    jbyte *p_pub  = (*env)->GetByteArrayElements(env, peerPub, NULL);
    jsize info_len = info ? (*env)->GetArrayLength(env, info) : 0;
    jbyte *p_info = info ? (*env)->GetByteArrayElements(env, info, NULL) : NULL;

    uint8_t key[32];
    int rc = pigeon_derive_session_key(
        (const uint8_t *)p_priv,
        (const uint8_t *)p_pub,
        (const uint8_t *)p_info, (size_t)info_len,
        key);

    (*env)->ReleaseByteArrayElements(env, priv, p_priv, JNI_ABORT);
    (*env)->ReleaseByteArrayElements(env, peerPub, p_pub, JNI_ABORT);
    if (info) (*env)->ReleaseByteArrayElements(env, info, p_info, JNI_ABORT);

    if (rc != 0) {
        throw_runtime(env, "pigeon_derive_session_key failed");
        return NULL;
    }
    return bytes_to_jba(env, key, 32);
}

JNIEXPORT jstring JNICALL
Java_com_marcelocantos_pigeon_jni_PigeonNative_deriveConfirmationCode(
    JNIEnv *env, jclass cls,
    jbyteArray pubA, jbyteArray pubB)
{
    (void)cls;
    if ((*env)->GetArrayLength(env, pubA) != 32 ||
        (*env)->GetArrayLength(env, pubB) != 32) {
        throw_runtime(env, "pubkey length must be 32");
        return NULL;
    }
    jbyte *a = (*env)->GetByteArrayElements(env, pubA, NULL);
    jbyte *b = (*env)->GetByteArrayElements(env, pubB, NULL);
    char code[7] = {0};
    int rc = pigeon_derive_confirmation_code(
        (const uint8_t *)a, (const uint8_t *)b, code);
    (*env)->ReleaseByteArrayElements(env, pubA, a, JNI_ABORT);
    (*env)->ReleaseByteArrayElements(env, pubB, b, JNI_ABORT);
    if (rc != 0) {
        throw_runtime(env, "pigeon_derive_confirmation_code failed");
        return NULL;
    }
    return (*env)->NewStringUTF(env, code);
}

// ----------------------------------------------------------------------------
// Wire helpers (varint, stream-header).
// ----------------------------------------------------------------------------

JNIEXPORT jbyteArray JNICALL
Java_com_marcelocantos_pigeon_jni_PigeonNative_uvarintEncode(
    JNIEnv *env, jclass cls, jlong v)
{
    (void)cls;
    uint8_t buf[PIGEON_MAX_VARINT_LEN];
    int n = pigeon_uvarint_encode((uint64_t)v, buf, sizeof(buf));
    if (n < 0) {
        throw_runtime(env, "pigeon_uvarint_encode failed");
        return NULL;
    }
    return bytes_to_jba(env, buf, (size_t)n);
}

JNIEXPORT jbyteArray JNICALL
Java_com_marcelocantos_pigeon_jni_PigeonNative_encodeStreamHeader(
    JNIEnv *env, jclass cls,
    jboolean isBackend, jint clientTag, jstring jname)
{
    (void)cls;
    const char *name = NULL;
    size_t name_len = 0;
    if (jname) {
        name = (*env)->GetStringUTFChars(env, jname, NULL);
        name_len = strlen(name);
    }
    uint8_t out[PIGEON_MAX_STREAM_HEADER];
    int n = pigeon_wire_stream_header_encode(
        isBackend ? true : false,
        (uint32_t)clientTag,
        name, name_len,
        out, sizeof(out));
    if (jname) (*env)->ReleaseStringUTFChars(env, jname, name);
    if (n < 0) {
        throw_runtime(env, "pigeon_wire_stream_header_encode failed");
        return NULL;
    }
    return bytes_to_jba(env, out, (size_t)n);
}

// ----------------------------------------------------------------------------
// Channel.
// ----------------------------------------------------------------------------

JNIEXPORT jlong JNICALL
Java_com_marcelocantos_pigeon_jni_PigeonNative_channelInit(
    JNIEnv *env, jclass cls,
    jbyteArray sendKey, jbyteArray recvKey, jint mode)
{
    (void)cls;
    if ((*env)->GetArrayLength(env, sendKey) != 32 ||
        (*env)->GetArrayLength(env, recvKey) != 32) {
        throw_runtime(env, "key length must be 32");
        return 0;
    }
    pigeon_channel *ch = calloc(1, sizeof(pigeon_channel));
    if (!ch) {
        throw_runtime(env, "alloc failed");
        return 0;
    }
    jbyte *s = (*env)->GetByteArrayElements(env, sendKey, NULL);
    jbyte *r = (*env)->GetByteArrayElements(env, recvKey, NULL);
    pigeon_channel_init(ch,
        (const uint8_t *)s, (const uint8_t *)r,
        (pigeon_channel_mode)mode);
    (*env)->ReleaseByteArrayElements(env, sendKey, s, JNI_ABORT);
    (*env)->ReleaseByteArrayElements(env, recvKey, r, JNI_ABORT);
    return (jlong)(intptr_t)ch;
}

JNIEXPORT void JNICALL
Java_com_marcelocantos_pigeon_jni_PigeonNative_channelFree(
    JNIEnv *env, jclass cls, jlong handle)
{
    (void)env; (void)cls;
    if (handle) free((void *)(intptr_t)handle);
}

JNIEXPORT jbyteArray JNICALL
Java_com_marcelocantos_pigeon_jni_PigeonNative_channelEncrypt(
    JNIEnv *env, jclass cls, jlong handle, jbyteArray plaintext)
{
    (void)cls;
    pigeon_channel *ch = (pigeon_channel *)(intptr_t)handle;
    if (!ch) { throw_runtime(env, "null channel"); return NULL; }
    jsize pn = (*env)->GetArrayLength(env, plaintext);
    jbyte *p = (*env)->GetByteArrayElements(env, plaintext, NULL);
    size_t out_cap = (size_t)pn + 8 + 16; // seq + GCM tag
    uint8_t *out = malloc(out_cap);
    if (!out) {
        (*env)->ReleaseByteArrayElements(env, plaintext, p, JNI_ABORT);
        throw_runtime(env, "alloc failed");
        return NULL;
    }
    int n = pigeon_channel_encrypt(ch,
        (const uint8_t *)p, (size_t)pn,
        out, out_cap);
    (*env)->ReleaseByteArrayElements(env, plaintext, p, JNI_ABORT);
    if (n < 0) {
        free(out);
        throw_runtime(env, "pigeon_channel_encrypt failed");
        return NULL;
    }
    jbyteArray res = bytes_to_jba(env, out, (size_t)n);
    free(out);
    return res;
}

JNIEXPORT jbyteArray JNICALL
Java_com_marcelocantos_pigeon_jni_PigeonNative_channelDecrypt(
    JNIEnv *env, jclass cls, jlong handle, jbyteArray ciphertext)
{
    (void)cls;
    pigeon_channel *ch = (pigeon_channel *)(intptr_t)handle;
    if (!ch) { throw_runtime(env, "null channel"); return NULL; }
    jsize cn = (*env)->GetArrayLength(env, ciphertext);
    jbyte *c = (*env)->GetByteArrayElements(env, ciphertext, NULL);
    uint8_t *out = malloc((size_t)cn);
    if (!out) {
        (*env)->ReleaseByteArrayElements(env, ciphertext, c, JNI_ABORT);
        throw_runtime(env, "alloc failed");
        return NULL;
    }
    int n = pigeon_channel_decrypt(ch,
        (const uint8_t *)c, (size_t)cn,
        out, (size_t)cn);
    (*env)->ReleaseByteArrayElements(env, ciphertext, c, JNI_ABORT);
    if (n < 0) {
        free(out);
        throw_runtime(env, "pigeon_channel_decrypt failed");
        return NULL;
    }
    jbyteArray res = bytes_to_jba(env, out, (size_t)n);
    free(out);
    return res;
}

// ----------------------------------------------------------------------------
// Loopback session pair.
// ----------------------------------------------------------------------------
//
// The loopback transports own a heap-allocated pair of endpoints. The
// session itself is *also* heap-allocated; the JNI wrapper owns the
// session pointer and frees it on close. Both endpoints (a, b) live
// alongside the A-side session — freeing A frees both endpoints; B
// must be freed first because it references A's endpoints' peer.

typedef struct loopback_pair {
    loop_endpoint a;
    loop_endpoint b;
    int ref_count;  // sessions that reference this pair (0..2)
} loopback_pair;

typedef struct session_holder {
    pigeon_session  session;  // owned
    loopback_pair  *pair;     // shared with the peer holder
    bool            is_a;     // which endpoint we used
} session_holder;

static loopback_pair *new_loopback_pair(void)
{
    loopback_pair *lp = calloc(1, sizeof(loopback_pair));
    if (!lp) return NULL;
    lp->a.peer = &lp->b;
    lp->b.peer = &lp->a;
    return lp;
}

static jobject new_long_array_pair(JNIEnv *env, jlong a, jlong b)
{
    jlongArray arr = (*env)->NewLongArray(env, 2);
    if (!arr) return NULL;
    jlong vals[2] = { a, b };
    (*env)->SetLongArrayRegion(env, arr, 0, 2, vals);
    return arr;
}

JNIEXPORT jlongArray JNICALL
Java_com_marcelocantos_pigeon_jni_PigeonNative_newLoopbackPair(
    JNIEnv *env, jclass cls,
    jlong channelA, jlong channelB,
    jboolean aIsBackend, jint aTag, jboolean bIsBackend, jint bTag,
    jobjectArray dgnames, jlongArray dgids)
{
    (void)cls;
    pigeon_channel *cha = (pigeon_channel *)(intptr_t)channelA;
    pigeon_channel *chb = (pigeon_channel *)(intptr_t)channelB;
    if (!cha || !chb) { throw_runtime(env, "null channel"); return NULL; }

    pigeon_dgchannel_def chans[PIGEON_MAX_DATAGRAM_CHANNELS];
    size_t nchans = 0;
    if (dgnames && dgids) {
        jsize n = (*env)->GetArrayLength(env, dgnames);
        if ((*env)->GetArrayLength(env, dgids) != n) {
            throw_runtime(env, "dgnames / dgids length mismatch");
            return NULL;
        }
        if (n > PIGEON_MAX_DATAGRAM_CHANNELS) {
            throw_runtime(env, "too many datagram channels");
            return NULL;
        }
        nchans = (size_t)n;
        jlong *ids = (*env)->GetLongArrayElements(env, dgids, NULL);
        for (jsize i = 0; i < n; i++) {
            jstring jn = (jstring)(*env)->GetObjectArrayElement(env, dgnames, i);
            const char *cn = (*env)->GetStringUTFChars(env, jn, NULL);
            memset(&chans[i], 0, sizeof(chans[i]));
            strncpy(chans[i].name, cn, PIGEON_MAX_NAME_LEN - 1);
            chans[i].channel_id = (uint64_t)ids[i];
            (*env)->ReleaseStringUTFChars(env, jn, cn);
            (*env)->DeleteLocalRef(env, jn);
        }
        (*env)->ReleaseLongArrayElements(env, dgids, ids, JNI_ABORT);
    }

    loopback_pair *lp = new_loopback_pair();
    if (!lp) { throw_runtime(env, "alloc failed"); return NULL; }

    session_holder *ha = calloc(1, sizeof(session_holder));
    session_holder *hb = calloc(1, sizeof(session_holder));
    if (!ha || !hb) {
        free(ha); free(hb); free(lp);
        throw_runtime(env, "alloc failed");
        return NULL;
    }
    ha->pair = lp; ha->is_a = true;
    hb->pair = lp; hb->is_a = false;
    lp->ref_count = 2;

    pigeon_transport ta, tb;
    loop_make_transport(&ta, &lp->a);
    loop_make_transport(&tb, &lp->b);

    if (pigeon_session_init(&ha->session, &ta, cha,
                            aIsBackend ? true : false, (uint32_t)aTag,
                            nchans ? chans : NULL, nchans) != 0) {
        free(ha); free(hb); free(lp);
        throw_runtime(env, "pigeon_session_init A failed");
        return NULL;
    }
    if (pigeon_session_init(&hb->session, &tb, chb,
                            bIsBackend ? true : false, (uint32_t)bTag,
                            nchans ? chans : NULL, nchans) != 0) {
        free(ha); free(hb); free(lp);
        throw_runtime(env, "pigeon_session_init B failed");
        return NULL;
    }

    return new_long_array_pair(env,
        (jlong)(intptr_t)ha, (jlong)(intptr_t)hb);
}

JNIEXPORT void JNICALL
Java_com_marcelocantos_pigeon_jni_PigeonNative_sessionFree(
    JNIEnv *env, jclass cls, jlong handle)
{
    (void)env; (void)cls;
    session_holder *h = (session_holder *)(intptr_t)handle;
    if (!h) return;
    if (h->pair && --h->pair->ref_count == 0) {
        free(h->pair);
    }
    free(h);
}

// ----------------------------------------------------------------------------
// Stream.
// ----------------------------------------------------------------------------
//
// pigeon_stream is a value type; we heap-allocate one and hand out the
// pointer.

JNIEXPORT jlong JNICALL
Java_com_marcelocantos_pigeon_jni_PigeonNative_sessionOpenStream(
    JNIEnv *env, jclass cls, jlong sessionHandle, jstring jname)
{
    (void)cls;
    session_holder *h = (session_holder *)(intptr_t)sessionHandle;
    if (!h) { throw_runtime(env, "null session"); return 0; }
    const char *name = (*env)->GetStringUTFChars(env, jname, NULL);
    pigeon_stream *s = calloc(1, sizeof(pigeon_stream));
    if (!s) {
        (*env)->ReleaseStringUTFChars(env, jname, name);
        throw_runtime(env, "alloc failed");
        return 0;
    }
    int rc = pigeon_session_open_stream(&h->session, name, s);
    (*env)->ReleaseStringUTFChars(env, jname, name);
    if (rc != 0) {
        free(s);
        throw_runtime(env, "pigeon_session_open_stream failed");
        return 0;
    }
    return (jlong)(intptr_t)s;
}

// Accept the next inbound stream on the loopback transport. Returns 0
// if no stream is pending. Reads the header off it, decodes the name
// (backend side strips the 4-byte tag), and returns a pigeon_stream*
// already bound to the session.
//
// out_name is filled in via a String[] passed by the caller (size 1).

JNIEXPORT jlong JNICALL
Java_com_marcelocantos_pigeon_jni_PigeonNative_sessionAcceptStream(
    JNIEnv *env, jclass cls, jlong sessionHandle, jobjectArray nameOut)
{
    (void)cls;
    session_holder *h = (session_holder *)(intptr_t)sessionHandle;
    if (!h || !h->pair) { throw_runtime(env, "null session"); return 0; }

    loop_endpoint *ep = h->is_a ? &h->pair->a : &h->pair->b;
    pigeon_stream_handle *sh = NULL;
    if (loop_accept_stream(ep, &sh) != 0) {
        return 0;
    }

    uint8_t hdr[PIGEON_MAX_STREAM_HEADER];
    size_t hn = 0;
    if (loop_recv_on_stream(ep, sh, hdr, sizeof(hdr), &hn) != 0) {
        throw_runtime(env, "loopback recv header failed");
        return 0;
    }

    // Inbound header was written by the *peer*. If we're the client
    // (is_backend == false), the peer is the backend and writes the
    // 4-byte tag prefix. If we're the backend, the peer is the client
    // and writes the no-tag form.
    char name[PIGEON_MAX_NAME_LEN] = {0};
    size_t name_len = 0;
    if (!h->session.is_backend) {
        uint32_t tag = 0;
        if (pigeon_wire_stream_header_decode_backend(hdr, hn, &tag,
                name, sizeof(name), &name_len) < 0) {
            throw_runtime(env, "decode_backend_stream_header failed");
            return 0;
        }
    } else {
        if (pigeon_wire_stream_header_decode_client(hdr, hn,
                name, sizeof(name), &name_len) < 0) {
            throw_runtime(env, "decode_client_stream_header failed");
            return 0;
        }
    }

    pigeon_stream *ps = calloc(1, sizeof(pigeon_stream));
    if (!ps) { throw_runtime(env, "alloc failed"); return 0; }
    ps->session = &h->session;
    ps->handle = sh;
    memcpy(ps->name, name, name_len);

    if (nameOut && (*env)->GetArrayLength(env, nameOut) > 0) {
        (*env)->SetObjectArrayElement(env, nameOut, 0,
            (*env)->NewStringUTF(env, name));
    }

    return (jlong)(intptr_t)ps;
}

JNIEXPORT void JNICALL
Java_com_marcelocantos_pigeon_jni_PigeonNative_streamFree(
    JNIEnv *env, jclass cls, jlong handle)
{
    (void)env; (void)cls;
    if (handle) free((void *)(intptr_t)handle);
}

JNIEXPORT void JNICALL
Java_com_marcelocantos_pigeon_jni_PigeonNative_streamSend(
    JNIEnv *env, jclass cls, jlong handle, jbyteArray msg)
{
    (void)cls;
    pigeon_stream *s = (pigeon_stream *)(intptr_t)handle;
    if (!s) { throw_runtime(env, "null stream"); return; }
    jsize n = (*env)->GetArrayLength(env, msg);
    jbyte *p = (*env)->GetByteArrayElements(env, msg, NULL);
    int rc = pigeon_stream_send(s, (const uint8_t *)p, (size_t)n);
    (*env)->ReleaseByteArrayElements(env, msg, p, JNI_ABORT);
    if (rc != 0) throw_runtime(env, "pigeon_stream_send failed");
}

JNIEXPORT jbyteArray JNICALL
Java_com_marcelocantos_pigeon_jni_PigeonNative_streamRecv(
    JNIEnv *env, jclass cls, jlong handle)
{
    (void)cls;
    pigeon_stream *s = (pigeon_stream *)(intptr_t)handle;
    if (!s) { throw_runtime(env, "null stream"); return NULL; }
    uint8_t *buf = malloc(PIGEON_MAX_MSG);
    if (!buf) { throw_runtime(env, "alloc failed"); return NULL; }
    int n = pigeon_stream_recv(s, buf, PIGEON_MAX_MSG);
    if (n < 0) {
        free(buf);
        throw_runtime(env, "pigeon_stream_recv failed");
        return NULL;
    }
    jbyteArray res = bytes_to_jba(env, buf, (size_t)n);
    free(buf);
    return res;
}

JNIEXPORT void JNICALL
Java_com_marcelocantos_pigeon_jni_PigeonNative_streamClose(
    JNIEnv *env, jclass cls, jlong handle)
{
    (void)env; (void)cls;
    pigeon_stream *s = (pigeon_stream *)(intptr_t)handle;
    if (s) pigeon_stream_close(s);
}

// ----------------------------------------------------------------------------
// Datagram.
// ----------------------------------------------------------------------------

JNIEXPORT jlong JNICALL
Java_com_marcelocantos_pigeon_jni_PigeonNative_sessionGetDatagram(
    JNIEnv *env, jclass cls, jlong sessionHandle, jstring jname)
{
    (void)cls;
    session_holder *h = (session_holder *)(intptr_t)sessionHandle;
    if (!h) { throw_runtime(env, "null session"); return 0; }
    const char *name = (*env)->GetStringUTFChars(env, jname, NULL);
    pigeon_datagram *d = calloc(1, sizeof(pigeon_datagram));
    if (!d) {
        (*env)->ReleaseStringUTFChars(env, jname, name);
        throw_runtime(env, "alloc failed");
        return 0;
    }
    int rc = pigeon_session_get_datagram(&h->session, name, d);
    (*env)->ReleaseStringUTFChars(env, jname, name);
    if (rc != 0) {
        free(d);
        throw_runtime(env, "pigeon_session_get_datagram failed");
        return 0;
    }
    return (jlong)(intptr_t)d;
}

JNIEXPORT void JNICALL
Java_com_marcelocantos_pigeon_jni_PigeonNative_datagramFree(
    JNIEnv *env, jclass cls, jlong handle)
{
    (void)env; (void)cls;
    if (handle) free((void *)(intptr_t)handle);
}

JNIEXPORT void JNICALL
Java_com_marcelocantos_pigeon_jni_PigeonNative_datagramSend(
    JNIEnv *env, jclass cls, jlong handle, jbyteArray payload)
{
    (void)cls;
    pigeon_datagram *d = (pigeon_datagram *)(intptr_t)handle;
    if (!d) { throw_runtime(env, "null datagram"); return; }
    jsize n = (*env)->GetArrayLength(env, payload);
    jbyte *p = (*env)->GetByteArrayElements(env, payload, NULL);
    int rc = pigeon_datagram_send(d, (const uint8_t *)p, (size_t)n);
    (*env)->ReleaseByteArrayElements(env, payload, p, JNI_ABORT);
    if (rc != 0) throw_runtime(env, "pigeon_datagram_send failed");
}

JNIEXPORT jbyteArray JNICALL
Java_com_marcelocantos_pigeon_jni_PigeonNative_datagramRecv(
    JNIEnv *env, jclass cls, jlong handle)
{
    (void)cls;
    pigeon_datagram *d = (pigeon_datagram *)(intptr_t)handle;
    if (!d) { throw_runtime(env, "null datagram"); return NULL; }
    uint8_t *buf = malloc(PIGEON_MAX_MSG);
    if (!buf) { throw_runtime(env, "alloc failed"); return NULL; }
    int n = pigeon_datagram_recv(d, buf, PIGEON_MAX_MSG);
    if (n < 0) {
        free(buf);
        throw_runtime(env, "pigeon_datagram_recv failed");
        return NULL;
    }
    // n == 0 is a legitimate "wrong channel; redispatch" signal.
    jbyteArray res = bytes_to_jba(env, buf, (size_t)n);
    free(buf);
    return res;
}
