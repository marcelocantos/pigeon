// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// Code generated from protocol/wireformats.yaml. DO NOT EDIT.

package com.marcelocantos.pigeon.crypto

class PigeonWireException(msg: String) : RuntimeException(msg)

object PigeonWire {
    fun encodeUvarint(value: ULong): ByteArray {
        val out = ArrayList<Byte>(10)
        var v = value
        while (v >= 0x80UL) {
            out.add(((v and 0x7fUL) or 0x80UL).toByte())
            v = v shr 7
        }
        out.add(v.toByte())
        return out.toByteArray()
    }

    fun decodeUvarint(buf: ByteArray, offset: Int = 0): Pair<ULong, Int> {
        var v: ULong = 0UL
        var shift = 0
        var i = offset
        while (i < buf.size) {
            if (i - offset >= 10) throw PigeonWireException("uvarint too long")
            val b = buf[i].toUByte().toInt()
            if ((b and 0x80) == 0) { v = v or ((b.toULong()) shl shift); return Pair(v, i - offset + 1) }
            v = v or ((b and 0x7f).toULong() shl shift)
            shift += 7
            i += 1
        }
        throw PigeonWireException("truncated uvarint")
    }

    fun encodeStreamHeader(name: String): ByteArray {
        val out = ArrayList<Byte>()
        run {
            val bytes = name.toByteArray(Charsets.UTF_8)
            for (b in encodeUvarint(bytes.size.toULong())) out.add(b)
            for (b in bytes) out.add(b)
        }
        return out.toByteArray()
    }

    data class StreamHeaderDecoded(val name: String, val consumed: Int)

    fun decodeStreamHeader(buf: ByteArray): StreamHeaderDecoded {
        var off = 0
        val (_lenU, _ln) = decodeUvarint(buf, off)
        off += _ln
        val _len = _lenU.toInt()
        if (buf.size - off < _len) throw PigeonWireException("truncated string")
        val name = String(buf, off, _len, Charsets.UTF_8)
        off += _len
        return StreamHeaderDecoded(name, off)
    }

    fun encodeDatagramPlaintext(channelId: ULong, payload: ByteArray): ByteArray {
        val out = ArrayList<Byte>()
        for (b in encodeUvarint(channelId)) out.add(b)
        for (b in payload) out.add(b)
        return out.toByteArray()
    }

    data class DatagramPlaintextDecoded(val channelId: ULong, val payload: ByteArray, val consumed: Int)

    fun decodeDatagramPlaintext(buf: ByteArray): DatagramPlaintextDecoded {
        var off = 0
        val (channelId, _vn) = decodeUvarint(buf, off)
        off += _vn
        val payload = buf.copyOfRange(off, buf.size)
        off = buf.size
        return DatagramPlaintextDecoded(channelId, payload, off)
    }

    enum class RelayGreetingVariant { CONNECT, REGISTER, LISTEN }

    data class RelayGreetingDecoded(
        val variant: RelayGreetingVariant,
        val instanceId: String = "",
        val token: String = "",
    )

    fun encodeRelayGreetingConnect(instanceId: String): ByteArray {
        val sb = StringBuilder("connect:")
        sb.append(instanceId)
        return sb.toString().toByteArray(Charsets.UTF_8)
    }

    fun encodeRelayGreetingRegister(token: String, instanceId: String): ByteArray {
        val sb = StringBuilder("register")
        val parts = listOf(token, instanceId)
        val anyNonEmpty = parts.any { it.isNotEmpty() }
        if (anyNonEmpty) {
            for (p in parts) { sb.append(':'); sb.append(p) }
        }
        return sb.toString().toByteArray(Charsets.UTF_8)
    }

    fun encodeRelayGreetingListen(token: String, instanceId: String): ByteArray {
        val sb = StringBuilder("listen")
        val parts = listOf(token, instanceId)
        val anyNonEmpty = parts.any { it.isNotEmpty() }
        if (anyNonEmpty) {
            for (p in parts) { sb.append(':'); sb.append(p) }
        }
        return sb.toString().toByteArray(Charsets.UTF_8)
    }

    fun decodeRelayGreeting(buf: ByteArray): RelayGreetingDecoded {
        val s = String(buf, Charsets.UTF_8)
        if (s.startsWith("connect:")) {
            val rest = s.substring(8)
            return RelayGreetingDecoded(
                variant = RelayGreetingVariant.CONNECT,
                instanceId = rest,
                token = ""
            )
        }
        if (s.startsWith("register")) {
            val rest = s.substring(8)
            val suffix: List<String> = if (rest.isEmpty()) List(2) { "" } else {
                if (!rest.startsWith(':')) throw PigeonWireException("malformed")
                val parts = rest.substring(1).split(':', limit = 2)
                parts + List(2 - parts.size) { "" }
            }
            return RelayGreetingDecoded(
                variant = RelayGreetingVariant.REGISTER,
                instanceId = suffix[1],
                token = suffix[0]
            )
        }
        if (s.startsWith("listen")) {
            val rest = s.substring(6)
            val suffix: List<String> = if (rest.isEmpty()) List(2) { "" } else {
                if (!rest.startsWith(':')) throw PigeonWireException("malformed")
                val parts = rest.substring(1).split(':', limit = 2)
                parts + List(2 - parts.size) { "" }
            }
            return RelayGreetingDecoded(
                variant = RelayGreetingVariant.LISTEN,
                instanceId = suffix[1],
                token = suffix[0]
            )
        }
        throw PigeonWireException("unrecognised prefix")
    }

}
