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
//   2. A JVM-callback transport (T36): each pigeon_transport vtable
//      entry thunks through JNI back into a Kotlin object that
//      implements com.marcelocantos.pigeon.peer.JniQuicTransport. This
//      lets the Kotlin side own the underlying QUIC stack (Kwik on
//      desktop JVM, Cronet / native ngtcp2 on Android) while libpigeon
//      drives all AEAD / wire framing. See sessionInitWithJniTransport
//      and the jni_transport_* helpers below.
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
// JavaVM* cache (T36).
//
// Every JVM-callback transport thunk needs a JNIEnv* on the current
// thread. Per the JNI spec the JNIEnv* is per-thread, so we cache the
// JavaVM* once (it's process-global) and recover the JNIEnv* on each
// call via GetEnv / AttachCurrentThread. For our use cases the calls
// always happen on the calling thread (libpigeon is single-threaded;
// the caller is already in JNI), so GetEnv almost always succeeds and
// we don't need to detach.
// ----------------------------------------------------------------------------

static JavaVM *g_jvm = NULL;

JNIEXPORT jint JNICALL JNI_OnLoad(JavaVM *vm, void *reserved)
{
    (void)reserved;
    g_jvm = vm;
    return JNI_VERSION_1_6;
}

// Fetch the JNIEnv* for the current thread, attaching to the JVM if
// the thread isn't already attached. On success *out_attached signals
// whether the caller is responsible for DetachCurrentThread.
//
// In practice libpigeon's JVM-callback transport is always invoked on
// the same JVM thread that called into the JNI surface, so GetEnv
// succeeds and *out_attached stays false. We still handle the attach
// case for robustness (e.g. future native callback dispatch threads).
static JNIEnv *jni_env_for_thread(bool *out_attached)
{
    if (out_attached) *out_attached = false;
    if (!g_jvm) return NULL;
    JNIEnv *env = NULL;
    jint rc = (*g_jvm)->GetEnv(g_jvm, (void **)&env, JNI_VERSION_1_6);
    if (rc == JNI_OK) return env;
    if (rc == JNI_EDETACHED) {
        if ((*g_jvm)->AttachCurrentThread(g_jvm, (void **)&env, NULL) != JNI_OK) {
            return NULL;
        }
        if (out_attached) *out_attached = true;
        return env;
    }
    return NULL;
}

static void jni_env_release(bool attached)
{
    if (attached && g_jvm) {
        (*g_jvm)->DetachCurrentThread(g_jvm);
    }
}

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
    JNIEnv *env, jclass cls, jstring jname)
{
    (void)cls;
    const char *name = NULL;
    size_t name_len = 0;
    if (jname) {
        name = (*env)->GetStringUTFChars(env, jname, NULL);
        name_len = strlen(name);
    }
    uint8_t out[PIGEON_MAX_STREAM_HEADER];
    int n = pigeon_wire_stream_header_encode(name, name_len, out, sizeof(out));
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

// Forward declaration — defined in the JVM-callback transport section
// below. session_holder needs the type so it can own a pointer.
typedef struct jni_transport_udata jni_transport_udata;
static void jni_transport_udata_free(jni_transport_udata *u);

typedef struct session_holder {
    pigeon_session  session;  // owned
    // For loopback-pair sessions: shared loopback transport.
    loopback_pair  *pair;     // shared with the peer holder, or NULL
    bool            is_a;     // which endpoint we used (loopback only)
    // For JNI-callback sessions: heap-allocated bridge to the Kotlin
    // transport object. NULL for loopback sessions. Owned by this
    // holder — released by sessionFree.
    jni_transport_udata *jni_transport;
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
                            nchans ? chans : NULL, nchans) != 0) {
        free(ha); free(hb); free(lp);
        throw_runtime(env, "pigeon_session_init A failed");
        return NULL;
    }
    if (pigeon_session_init(&hb->session, &tb, chb,
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
    if (h->jni_transport) {
        jni_transport_udata_free(h->jni_transport);
        h->jni_transport = NULL;
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
// if no stream is pending. Reads the [varint name-len][name] header off
// it, decodes the name, and returns a pigeon_stream* already bound to
// the session.
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

    // Under T45 the header is just [varint name-len][name] — both peers
    // are symmetric, no clientTag.
    char name[PIGEON_MAX_NAME_LEN] = {0};
    size_t name_len = 0;
    if (pigeon_wire_stream_header_decode(hdr, hn,
            name, sizeof(name), &name_len) < 0) {
        throw_runtime(env, "decode_stream_header failed");
        return 0;
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

// ----------------------------------------------------------------------------
// JVM-callback transport (T36).
//
// A pigeon_transport whose vtable entries thunk through JNI back into a
// Kotlin object that implements com.marcelocantos.pigeon.peer.JniQuicTransport.
// libpigeon drives all AEAD / wire framing; the Kotlin side owns the
// underlying QUIC stack (Kwik / Cronet / ngtcp2).
//
// Memory model: one heap-allocated `jni_transport_udata` per session.
// The udata box owns one global ref to the Kotlin transport object and
// the cached jmethodIDs for each vtable entry. The C transport vtable's
// `userdata` pointer is the udata pointer; cwire uses the same shape
// (cwire_go_udata) so this stays familiar for cross-language reviewers.
//
// Threading: libpigeon's stream/datagram entry points are called on
// the JVM thread that drove into the JNI surface, so GetEnv succeeds
// and no AttachCurrentThread round-trip happens on the hot path.
// jni_env_for_thread handles the rare attach case for robustness.
// ----------------------------------------------------------------------------

struct jni_transport_udata {
    // Global ref to the Kotlin transport object (implements
    // JniQuicTransport). NULL after jni_transport_udata_free.
    jobject  obj;

    // Cached jmethodIDs (resolved once at construction). jmethodIDs
    // remain valid as long as the class is loaded; the global ref on
    // `obj` keeps the class alive transitively.
    jmethodID m_open_stream;     // ()J
    jmethodID m_accept_stream;   // ()J
    jmethodID m_send_on_stream;  // (J[B)V
    jmethodID m_recv_on_stream;  // (J)[B
    jmethodID m_close_stream;    // (J)V
    jmethodID m_send_datagram;   // ([B)V
    jmethodID m_recv_datagram;   // ()[B
};

// Free the udata. Releases the global ref via the calling thread's
// JNIEnv (recovered via jni_env_for_thread). Idempotent on NULL.
static void jni_transport_udata_free(jni_transport_udata *u)
{
    if (!u) return;
    if (u->obj) {
        bool attached = false;
        JNIEnv *env = jni_env_for_thread(&attached);
        if (env) {
            (*env)->DeleteGlobalRef(env, u->obj);
            jni_env_release(attached);
        }
        u->obj = NULL;
    }
    free(u);
}

// Check for a pending Java exception. If one is set, clear it (so the
// C library doesn't re-enter Java with the exception still set) and
// return -1. Otherwise return 0. The original exception is lost; for
// our use case the transport is expected to throw only on hard errors
// where -1 propagation is the correct response.
//
// (A future enhancement could log the exception via spdlog; for now
// the Kotlin layer surfaces failures as RuntimeException at the API
// boundary, which is good enough for the loopback tests.)
static int jni_clear_pending(JNIEnv *env)
{
    if ((*env)->ExceptionCheck(env)) {
        (*env)->ExceptionDescribe(env);
        (*env)->ExceptionClear(env);
        return -1;
    }
    return 0;
}

// --- vtable thunks ---
//
// Each one recovers the JNIEnv*, invokes the Kotlin method, checks for
// a pending exception, and translates errors into a -1 return.

static int jni_tr_open_stream(void *ud, pigeon_stream_handle **out)
{
    jni_transport_udata *u = (jni_transport_udata *)ud;
    bool attached = false;
    JNIEnv *env = jni_env_for_thread(&attached);
    if (!env || !u || !u->obj) return -1;
    jlong h = (*env)->CallLongMethod(env, u->obj, u->m_open_stream);
    int rc = jni_clear_pending(env);
    jni_env_release(attached);
    if (rc != 0) return -1;
    if (h == 0) return -1;
    *out = (pigeon_stream_handle *)(intptr_t)h;
    return 0;
}

static int jni_tr_accept_stream(void *ud, pigeon_stream_handle **out)
{
    jni_transport_udata *u = (jni_transport_udata *)ud;
    bool attached = false;
    JNIEnv *env = jni_env_for_thread(&attached);
    if (!env || !u || !u->obj) return -1;
    jlong h = (*env)->CallLongMethod(env, u->obj, u->m_accept_stream);
    int rc = jni_clear_pending(env);
    jni_env_release(attached);
    if (rc != 0) return -1;
    if (h == 0) return -1;
    *out = (pigeon_stream_handle *)(intptr_t)h;
    return 0;
}

static int jni_tr_send_on_stream(void *ud, pigeon_stream_handle *h,
                                 const uint8_t *data, size_t len)
{
    jni_transport_udata *u = (jni_transport_udata *)ud;
    bool attached = false;
    JNIEnv *env = jni_env_for_thread(&attached);
    if (!env || !u || !u->obj) return -1;
    jbyteArray arr = (*env)->NewByteArray(env, (jsize)len);
    if (!arr) { jni_env_release(attached); return -1; }
    if (len > 0) {
        (*env)->SetByteArrayRegion(env, arr, 0, (jsize)len, (const jbyte *)data);
    }
    (*env)->CallVoidMethod(env, u->obj, u->m_send_on_stream,
                           (jlong)(intptr_t)h, arr);
    int rc = jni_clear_pending(env);
    (*env)->DeleteLocalRef(env, arr);
    jni_env_release(attached);
    return rc;
}

static int jni_tr_recv_on_stream(void *ud, pigeon_stream_handle *h,
                                 uint8_t *buf, size_t buf_len, size_t *out_len)
{
    jni_transport_udata *u = (jni_transport_udata *)ud;
    bool attached = false;
    JNIEnv *env = jni_env_for_thread(&attached);
    if (!env || !u || !u->obj) return -1;
    jobject jres = (*env)->CallObjectMethod(env, u->obj, u->m_recv_on_stream,
                                            (jlong)(intptr_t)h);
    if (jni_clear_pending(env) != 0) {
        if (jres) (*env)->DeleteLocalRef(env, jres);
        jni_env_release(attached);
        return -1;
    }
    if (!jres) {
        jni_env_release(attached);
        return -1;
    }
    jsize n = (*env)->GetArrayLength(env, (jbyteArray)jres);
    if ((size_t)n > buf_len) {
        (*env)->DeleteLocalRef(env, jres);
        jni_env_release(attached);
        return -1;
    }
    if (n > 0) {
        (*env)->GetByteArrayRegion(env, (jbyteArray)jres, 0, n, (jbyte *)buf);
    }
    *out_len = (size_t)n;
    (*env)->DeleteLocalRef(env, jres);
    jni_env_release(attached);
    return 0;
}

static int jni_tr_close_stream(void *ud, pigeon_stream_handle *h)
{
    jni_transport_udata *u = (jni_transport_udata *)ud;
    bool attached = false;
    JNIEnv *env = jni_env_for_thread(&attached);
    if (!env || !u || !u->obj) return -1;
    (*env)->CallVoidMethod(env, u->obj, u->m_close_stream, (jlong)(intptr_t)h);
    int rc = jni_clear_pending(env);
    jni_env_release(attached);
    return rc;
}

static int jni_tr_send_datagram(void *ud, const uint8_t *data, size_t len)
{
    jni_transport_udata *u = (jni_transport_udata *)ud;
    bool attached = false;
    JNIEnv *env = jni_env_for_thread(&attached);
    if (!env || !u || !u->obj) return -1;
    jbyteArray arr = (*env)->NewByteArray(env, (jsize)len);
    if (!arr) { jni_env_release(attached); return -1; }
    if (len > 0) {
        (*env)->SetByteArrayRegion(env, arr, 0, (jsize)len, (const jbyte *)data);
    }
    (*env)->CallVoidMethod(env, u->obj, u->m_send_datagram, arr);
    int rc = jni_clear_pending(env);
    (*env)->DeleteLocalRef(env, arr);
    jni_env_release(attached);
    return rc;
}

static int jni_tr_recv_datagram(void *ud, uint8_t *buf, size_t buf_len, size_t *out_len)
{
    jni_transport_udata *u = (jni_transport_udata *)ud;
    bool attached = false;
    JNIEnv *env = jni_env_for_thread(&attached);
    if (!env || !u || !u->obj) return -1;
    jobject jres = (*env)->CallObjectMethod(env, u->obj, u->m_recv_datagram);
    if (jni_clear_pending(env) != 0) {
        if (jres) (*env)->DeleteLocalRef(env, jres);
        jni_env_release(attached);
        return -1;
    }
    if (!jres) {
        jni_env_release(attached);
        return -1;
    }
    jsize n = (*env)->GetArrayLength(env, (jbyteArray)jres);
    if ((size_t)n > buf_len) {
        (*env)->DeleteLocalRef(env, jres);
        jni_env_release(attached);
        return -1;
    }
    if (n > 0) {
        (*env)->GetByteArrayRegion(env, (jbyteArray)jres, 0, n, (jbyte *)buf);
    }
    *out_len = (size_t)n;
    (*env)->DeleteLocalRef(env, jres);
    jni_env_release(attached);
    return 0;
}

// Build a jni_transport_udata around the supplied Kotlin object. The
// Kotlin object must implement
// com.marcelocantos.pigeon.peer.JniQuicTransport. On success returns a
// heap-allocated box; the caller (sessionInitWithJniTransport) takes
// ownership. Returns NULL and throws a Java RuntimeException on
// failure.
static jni_transport_udata *jni_transport_udata_new(JNIEnv *env, jobject transport)
{
    jclass cls = (*env)->GetObjectClass(env, transport);
    if (!cls) { throw_runtime(env, "GetObjectClass(transport) failed"); return NULL; }

    jni_transport_udata *u = calloc(1, sizeof(*u));
    if (!u) { throw_runtime(env, "alloc failed"); return NULL; }

    u->m_open_stream    = (*env)->GetMethodID(env, cls, "openStream",    "()J");
    u->m_accept_stream  = (*env)->GetMethodID(env, cls, "acceptStream",  "()J");
    u->m_send_on_stream = (*env)->GetMethodID(env, cls, "sendOnStream",  "(J[B)V");
    u->m_recv_on_stream = (*env)->GetMethodID(env, cls, "recvOnStream",  "(J)[B");
    u->m_close_stream   = (*env)->GetMethodID(env, cls, "closeStream",   "(J)V");
    u->m_send_datagram  = (*env)->GetMethodID(env, cls, "sendDatagram",  "([B)V");
    u->m_recv_datagram  = (*env)->GetMethodID(env, cls, "recvDatagram",  "()[B");

    // GetMethodID throws NoSuchMethodError on miss — check once.
    if ((*env)->ExceptionCheck(env)) {
        free(u);
        return NULL;  // exception propagates to Java caller
    }
    if (!u->m_open_stream || !u->m_accept_stream ||
        !u->m_send_on_stream || !u->m_recv_on_stream ||
        !u->m_close_stream || !u->m_send_datagram ||
        !u->m_recv_datagram) {
        free(u);
        throw_runtime(env, "JniQuicTransport method lookup failed");
        return NULL;
    }

    u->obj = (*env)->NewGlobalRef(env, transport);
    if (!u->obj) {
        free(u);
        throw_runtime(env, "NewGlobalRef(transport) failed");
        return NULL;
    }
    return u;
}

static void jni_make_transport(pigeon_transport *t, jni_transport_udata *u)
{
    memset(t, 0, sizeof(*t));
    t->userdata        = u;
    t->open_stream     = jni_tr_open_stream;
    t->accept_stream   = jni_tr_accept_stream;
    t->send_on_stream  = jni_tr_send_on_stream;
    t->recv_on_stream  = jni_tr_recv_on_stream;
    t->close_stream    = jni_tr_close_stream;
    t->send_datagram   = jni_tr_send_datagram;
    t->recv_datagram   = jni_tr_recv_datagram;
}

// Construct a pigeon_session over a JVM-callback transport. The
// session takes ownership of the JNI transport bridge (a global ref
// to `transport` plus cached jmethodIDs); sessionFree releases both
// when the Kotlin Session is closed.
//
// The caller still owns the `channel` handle and the `transport`
// Kotlin reference. The session adds its own global ref to keep the
// transport alive for the session's lifetime — Kotlin code can drop
// its local reference to `transport` as soon as this call returns
// without the C side dangling.
JNIEXPORT jlong JNICALL
Java_com_marcelocantos_pigeon_jni_PigeonNative_sessionInitWithJniTransport(
    JNIEnv *env, jclass cls,
    jlong channelHandle, jobject transport,
    jobjectArray dgnames, jlongArray dgids)
{
    (void)cls;
    pigeon_channel *ch = (pigeon_channel *)(intptr_t)channelHandle;
    if (!ch || !transport) { throw_runtime(env, "null channel or transport"); return 0; }

    pigeon_dgchannel_def chans[PIGEON_MAX_DATAGRAM_CHANNELS];
    size_t nchans = 0;
    if (dgnames && dgids) {
        jsize n = (*env)->GetArrayLength(env, dgnames);
        if ((*env)->GetArrayLength(env, dgids) != n) {
            throw_runtime(env, "dgnames / dgids length mismatch");
            return 0;
        }
        if (n > PIGEON_MAX_DATAGRAM_CHANNELS) {
            throw_runtime(env, "too many datagram channels");
            return 0;
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

    jni_transport_udata *u = jni_transport_udata_new(env, transport);
    if (!u) return 0;  // exception already pending

    session_holder *h = calloc(1, sizeof(session_holder));
    if (!h) {
        jni_transport_udata_free(u);
        throw_runtime(env, "alloc failed");
        return 0;
    }
    h->jni_transport = u;

    pigeon_transport t;
    jni_make_transport(&t, u);

    if (pigeon_session_init(&h->session, &t, ch,
                            nchans ? chans : NULL, nchans) != 0) {
        jni_transport_udata_free(u);
        h->jni_transport = NULL;
        free(h);
        throw_runtime(env, "pigeon_session_init failed");
        return 0;
    }

    return (jlong)(intptr_t)h;
}

// Accept the next inbound stream by driving the session's transport
// vtable (which thunks through JNI back to the Kotlin transport).
// Reads the first framed message off the new stream — the unencrypted
// name-binding header — decodes it according to the session's role,
// and returns a fully-attached pigeon_stream*.
//
// Returns 0 (no exception) if accept_stream signals "no stream
// available" — the Kotlin layer translates this to acceptStream() ==
// null. Returns nonzero on header-decode error or transport hard
// failure (Java exception thrown).
JNIEXPORT jlong JNICALL
Java_com_marcelocantos_pigeon_jni_PigeonNative_sessionAcceptStreamGeneric(
    JNIEnv *env, jclass cls, jlong sessionHandle, jobjectArray nameOut)
{
    (void)cls;
    session_holder *h = (session_holder *)(intptr_t)sessionHandle;
    if (!h) { throw_runtime(env, "null session"); return 0; }

    pigeon_transport *tr = &h->session.transport;
    if (!tr->accept_stream || !tr->recv_on_stream) {
        throw_runtime(env, "transport lacks accept_stream / recv_on_stream");
        return 0;
    }

    pigeon_stream_handle *sh = NULL;
    if (tr->accept_stream(tr->userdata, &sh) != 0) {
        // No stream pending (or transport hiccup the Kotlin side already
        // surfaced as an exception). If an exception is pending leave it
        // for the JVM to observe; otherwise this is the "null" path.
        return 0;
    }

    uint8_t hdr[PIGEON_MAX_STREAM_HEADER];
    size_t hn = 0;
    if (tr->recv_on_stream(tr->userdata, sh, hdr, sizeof(hdr), &hn) != 0) {
        throw_runtime(env, "transport recv_on_stream(header) failed");
        return 0;
    }

    // Under T45 the header is just [varint name-len][name] — both peers
    // are symmetric, no clientTag.
    char name[PIGEON_MAX_NAME_LEN] = {0};
    size_t name_len = 0;
    if (pigeon_wire_stream_header_decode(hdr, hn,
            name, sizeof(name), &name_len) < 0) {
        throw_runtime(env, "decode_stream_header failed");
        return 0;
    }

    pigeon_stream *ps = calloc(1, sizeof(pigeon_stream));
    if (!ps) { throw_runtime(env, "alloc failed"); return 0; }
    ps->session = &h->session;
    ps->handle  = sh;
    memcpy(ps->name, name, name_len);

    if (nameOut && (*env)->GetArrayLength(env, nameOut) > 0) {
        (*env)->SetObjectArrayElement(env, nameOut, 0,
            (*env)->NewStringUTF(env, name));
    }

    return (jlong)(intptr_t)ps;
}
