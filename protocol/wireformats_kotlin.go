// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"fmt"
	"io"
	"strings"
)

// ExportKotlin writes a Kotlin source file with native encoders/decoders
// for every format in the set. Generated as a top-level object in the
// requested package.
func (s *WireFormatSet) ExportKotlin(w io.Writer, pkg string) error {
	var b strings.Builder
	b.WriteString("// Copyright 2026 Marcelo Cantos\n")
	b.WriteString("// SPDX-License-Identifier: Apache-2.0\n")
	b.WriteString("//\n")
	b.WriteString("// Code generated from protocol/wireformats.yaml. DO NOT EDIT.\n\n")
	fmt.Fprintf(&b, "package %s\n\n", pkg)
	b.WriteString("class PigeonWireException(msg: String) : RuntimeException(msg)\n\n")
	b.WriteString("object PigeonWire {\n")

	// Varint helpers.
	b.WriteString("    fun encodeUvarint(value: ULong): ByteArray {\n")
	b.WriteString("        val out = ArrayList<Byte>(10)\n")
	b.WriteString("        var v = value\n")
	b.WriteString("        while (v >= 0x80UL) {\n")
	b.WriteString("            out.add(((v and 0x7fUL) or 0x80UL).toByte())\n")
	b.WriteString("            v = v shr 7\n        }\n")
	b.WriteString("        out.add(v.toByte())\n        return out.toByteArray()\n    }\n\n")
	b.WriteString("    fun decodeUvarint(buf: ByteArray, offset: Int = 0): Pair<ULong, Int> {\n")
	b.WriteString("        var v: ULong = 0UL\n        var shift = 0\n")
	b.WriteString("        var i = offset\n")
	b.WriteString("        while (i < buf.size) {\n")
	b.WriteString("            if (i - offset >= 10) throw PigeonWireException(\"uvarint too long\")\n")
	b.WriteString("            val b = buf[i].toUByte().toInt()\n")
	b.WriteString("            if ((b and 0x80) == 0) { v = v or ((b.toULong()) shl shift); return Pair(v, i - offset + 1) }\n")
	b.WriteString("            v = v or ((b and 0x7f).toULong() shl shift)\n")
	b.WriteString("            shift += 7\n            i += 1\n        }\n")
	b.WriteString("        throw PigeonWireException(\"truncated uvarint\")\n    }\n\n")

	for _, f := range s.Formats {
		switch {
		case f.IsPlain():
			kotlinEmitPlain(&b, f)
		case f.IsVariant():
			kotlinEmitVariant(&b, f)
		case f.IsUnion():
			kotlinEmitUnion(&b, f)
		}
	}
	b.WriteString("}\n")
	_, err := io.WriteString(w, b.String())
	return err
}

func wireKotlinType(k WireFieldKind) string {
	switch k {
	case WFU32BE:
		return "UInt"
	case WFVarint:
		return "ULong"
	case WFStringLenV, WFStringToEnd:
		return "String"
	case WFBytesToEnd:
		return "ByteArray"
	}
	return "Any"
}

func kotlinEmitPlain(b *strings.Builder, f WireFormat) {
	upper := upperCamel(f.Name)
	// Encoder.
	fmt.Fprintf(b, "    fun encode%s(", upper)
	for i, fl := range f.Fields {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(b, "%s: %s", lowerCamel(fl.Name), wireKotlinType(fl.Kind))
	}
	b.WriteString("): ByteArray {\n")
	b.WriteString("        val out = ArrayList<Byte>()\n")
	for _, fl := range f.Fields {
		kotlinEmitEncodeField(b, fl, "        ")
	}
	b.WriteString("        return out.toByteArray()\n    }\n\n")

	// Decoder result class.
	fmt.Fprintf(b, "    data class %sDecoded(", upper)
	for i, fl := range f.Fields {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(b, "val %s: %s", lowerCamel(fl.Name), wireKotlinType(fl.Kind))
	}
	b.WriteString(", val consumed: Int)\n\n")

	fmt.Fprintf(b, "    fun decode%s(buf: ByteArray): %sDecoded {\n", upper, upper)
	b.WriteString("        var off = 0\n")
	for _, fl := range f.Fields {
		kotlinEmitDecodeField(b, fl)
	}
	fmt.Fprintf(b, "        return %sDecoded(", upper)
	for i, fl := range f.Fields {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(lowerCamel(fl.Name))
	}
	b.WriteString(", off)\n    }\n\n")
}

func kotlinEmitVariant(b *strings.Builder, f WireFormat) {
	upper := upperCamel(f.Name)
	flag := "is" + upperCamel(f.Variants[0].Name)
	seen := map[string]bool{}
	var allFields []WireField
	for _, v := range f.Variants {
		for _, fl := range v.Fields {
			if !seen[fl.Name] {
				allFields = append(allFields, fl)
				seen[fl.Name] = true
			}
		}
	}
	fmt.Fprintf(b, "    fun encode%s(%s: Boolean", upper, flag)
	for _, fl := range allFields {
		fmt.Fprintf(b, ", %s: %s", lowerCamel(fl.Name), wireKotlinType(fl.Kind))
	}
	b.WriteString("): ByteArray {\n")
	b.WriteString("        val out = ArrayList<Byte>()\n")
	for i, v := range f.Variants {
		if i == 0 {
			fmt.Fprintf(b, "        if (%s) {\n", flag)
		} else {
			b.WriteString("        } else {\n")
		}
		for _, fl := range v.Fields {
			kotlinEmitEncodeField(b, fl, "            ")
		}
	}
	b.WriteString("        }\n")
	b.WriteString("        return out.toByteArray()\n    }\n\n")

	// Per-variant decoders.
	for _, v := range f.Variants {
		fmt.Fprintf(b, "    data class %s%sDecoded(", upper, upperCamel(v.Name))
		for i, fl := range v.Fields {
			if i > 0 {
				b.WriteString(", ")
			}
			fmt.Fprintf(b, "val %s: %s", lowerCamel(fl.Name), wireKotlinType(fl.Kind))
		}
		b.WriteString(", val consumed: Int)\n\n")

		fmt.Fprintf(b, "    fun decode%s%s(buf: ByteArray): %s%sDecoded {\n",
			upper, upperCamel(v.Name), upper, upperCamel(v.Name))
		b.WriteString("        var off = 0\n")
		for _, fl := range v.Fields {
			kotlinEmitDecodeField(b, fl)
		}
		fmt.Fprintf(b, "        return %s%sDecoded(", upper, upperCamel(v.Name))
		for i, fl := range v.Fields {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(lowerCamel(fl.Name))
		}
		b.WriteString(", off)\n    }\n\n")
	}
}

func kotlinEmitUnion(b *strings.Builder, f WireFormat) {
	upper := upperCamel(f.Name)
	fmt.Fprintf(b, "    enum class %sVariant {", upper)
	for i, u := range f.Union {
		if i > 0 {
			b.WriteString(",")
		}
		fmt.Fprintf(b, " %s", strings.ToUpper(u.Name))
	}
	b.WriteString(" }\n\n")

	// Data class.
	fmt.Fprintf(b, "    data class %sDecoded(\n", upper)
	fmt.Fprintf(b, "        val variant: %sVariant,\n", upper)
	allNames := map[string]bool{}
	allFields := []WireField{}
	for _, u := range f.Union {
		for _, fl := range u.Fields {
			if !allNames[fl.Name] {
				fmt.Fprintf(b, "        val %s: %s = \"\",\n", lowerCamel(fl.Name), wireKotlinType(fl.Kind))
				allNames[fl.Name] = true
				allFields = append(allFields, fl)
			}
		}
		for _, fl := range u.ColonSuffix {
			if !allNames[fl.Name] {
				fmt.Fprintf(b, "        val %s: String = \"\",\n", lowerCamel(fl.Name))
				allNames[fl.Name] = true
				allFields = append(allFields, WireField{Name: fl.Name, Kind: WFStringToEnd})
			}
		}
	}
	b.WriteString("    )\n\n")

	// Encoders per variant.
	for _, u := range f.Union {
		fmt.Fprintf(b, "    fun encode%s%s(", upper, upperCamel(u.Name))
		first := true
		for _, fl := range u.Fields {
			if !first {
				b.WriteString(", ")
			}
			first = false
			fmt.Fprintf(b, "%s: String", lowerCamel(fl.Name))
		}
		for _, fl := range u.ColonSuffix {
			if !first {
				b.WriteString(", ")
			}
			first = false
			fmt.Fprintf(b, "%s: String", lowerCamel(fl.Name))
		}
		b.WriteString("): ByteArray {\n")
		fmt.Fprintf(b, "        val sb = StringBuilder(%q)\n", u.Prefix)
		if len(u.Fields) > 0 {
			for _, fl := range u.Fields {
				fmt.Fprintf(b, "        sb.append(%s)\n", lowerCamel(fl.Name))
			}
		}
		if len(u.ColonSuffix) > 0 {
			b.WriteString("        val parts = listOf(")
			for i, fl := range u.ColonSuffix {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(lowerCamel(fl.Name))
			}
			b.WriteString(")\n")
			b.WriteString("        val anyNonEmpty = parts.any { it.isNotEmpty() }\n")
			b.WriteString("        if (anyNonEmpty) {\n")
			b.WriteString("            for (p in parts) { sb.append(':'); sb.append(p) }\n")
			b.WriteString("        }\n")
		}
		b.WriteString("        return sb.toString().toByteArray(Charsets.UTF_8)\n    }\n\n")
	}

	// Decoder.
	fmt.Fprintf(b, "    fun decode%s(buf: ByteArray): %sDecoded {\n", upper, upper)
	b.WriteString("        val s = String(buf, Charsets.UTF_8)\n")
	// Sort by descending prefix length.
	ordered := append([]WireUnion(nil), f.Union...)
	for i := 0; i < len(ordered); i++ {
		for j := i + 1; j < len(ordered); j++ {
			if len(ordered[j].Prefix) > len(ordered[i].Prefix) {
				ordered[i], ordered[j] = ordered[j], ordered[i]
			}
		}
	}
	for _, u := range ordered {
		fmt.Fprintf(b, "        if (s.startsWith(%q)) {\n", u.Prefix)
		fmt.Fprintf(b, "            val rest = s.substring(%d)\n", len(u.Prefix))
		defaults := map[string]string{}
		for _, fl := range allFields {
			switch fl.Kind {
			case WFU32BE, WFVarint:
				defaults[fl.Name] = "0U"
			default:
				defaults[fl.Name] = "\"\""
			}
		}
		assigns := map[string]string{}
		for k, v := range defaults {
			assigns[k] = v
		}
		if len(u.Fields) > 0 {
			for _, fl := range u.Fields {
				assigns[fl.Name] = "rest"
			}
		}
		if len(u.ColonSuffix) > 0 {
			b.WriteString("            val suffix: List<String> = if (rest.isEmpty()) ")
			fmt.Fprintf(b, "List(%d) { \"\" }", len(u.ColonSuffix))
			b.WriteString(" else {\n")
			b.WriteString("                if (!rest.startsWith(':')) throw PigeonWireException(\"malformed\")\n")
			fmt.Fprintf(b, "                val parts = rest.substring(1).split(':', limit = %d)\n", len(u.ColonSuffix))
			fmt.Fprintf(b, "                parts + List(%d - parts.size) { \"\" }\n", len(u.ColonSuffix))
			b.WriteString("            }\n")
			for i, fl := range u.ColonSuffix {
				assigns[fl.Name] = fmt.Sprintf("suffix[%d]", i)
			}
		}
		fmt.Fprintf(b, "            return %sDecoded(\n", upper)
		fmt.Fprintf(b, "                variant = %sVariant.%s", upper, strings.ToUpper(u.Name))
		for _, fl := range allFields {
			val, ok := assigns[fl.Name]
			if !ok {
				val = defaults[fl.Name]
			}
			fmt.Fprintf(b, ",\n                %s = %s", lowerCamel(fl.Name), val)
		}
		b.WriteString("\n            )\n")
		b.WriteString("        }\n")
	}
	b.WriteString("        throw PigeonWireException(\"unrecognised prefix\")\n")
	b.WriteString("    }\n\n")
}

func kotlinEmitEncodeField(b *strings.Builder, fl WireField, indent string) {
	v := lowerCamel(fl.Name)
	switch fl.Kind {
	case WFU32BE:
		fmt.Fprintf(b, "%sout.add(((%s.toInt() shr 24) and 0xff).toByte())\n", indent, v)
		fmt.Fprintf(b, "%sout.add(((%s.toInt() shr 16) and 0xff).toByte())\n", indent, v)
		fmt.Fprintf(b, "%sout.add(((%s.toInt() shr 8) and 0xff).toByte())\n", indent, v)
		fmt.Fprintf(b, "%sout.add((%s.toInt() and 0xff).toByte())\n", indent, v)
	case WFVarint:
		fmt.Fprintf(b, "%sfor (b in encodeUvarint(%s)) out.add(b)\n", indent, v)
	case WFStringLenV:
		fmt.Fprintf(b, "%srun {\n", indent)
		fmt.Fprintf(b, "%s    val bytes = %s.toByteArray(Charsets.UTF_8)\n", indent, v)
		fmt.Fprintf(b, "%s    for (b in encodeUvarint(bytes.size.toULong())) out.add(b)\n", indent)
		fmt.Fprintf(b, "%s    for (b in bytes) out.add(b)\n", indent)
		fmt.Fprintf(b, "%s}\n", indent)
	case WFStringToEnd:
		fmt.Fprintf(b, "%sfor (b in %s.toByteArray(Charsets.UTF_8)) out.add(b)\n", indent, v)
	case WFBytesToEnd:
		fmt.Fprintf(b, "%sfor (b in %s) out.add(b)\n", indent, v)
	}
}

func kotlinEmitDecodeField(b *strings.Builder, fl WireField) {
	v := lowerCamel(fl.Name)
	switch fl.Kind {
	case WFU32BE:
		fmt.Fprintf(b, "        if (buf.size - off < 4) throw PigeonWireException(\"truncated u32 %s\")\n", fl.Name)
		fmt.Fprintf(b, "        val %s: UInt = ((buf[off].toUByte().toUInt() shl 24)\n", v)
		b.WriteString("            or (buf[off + 1].toUByte().toUInt() shl 16)\n")
		b.WriteString("            or (buf[off + 2].toUByte().toUInt() shl 8)\n")
		b.WriteString("            or buf[off + 3].toUByte().toUInt())\n")
		b.WriteString("        off += 4\n")
	case WFVarint:
		fmt.Fprintf(b, "        val (%s, _vn) = decodeUvarint(buf, off)\n", v)
		b.WriteString("        off += _vn\n")
	case WFStringLenV:
		b.WriteString("        val (_lenU, _ln) = decodeUvarint(buf, off)\n")
		b.WriteString("        off += _ln\n")
		b.WriteString("        val _len = _lenU.toInt()\n")
		b.WriteString("        if (buf.size - off < _len) throw PigeonWireException(\"truncated string\")\n")
		fmt.Fprintf(b, "        val %s = String(buf, off, _len, Charsets.UTF_8)\n", v)
		b.WriteString("        off += _len\n")
	case WFStringToEnd:
		fmt.Fprintf(b, "        val %s = String(buf, off, buf.size - off, Charsets.UTF_8)\n", v)
		b.WriteString("        off = buf.size\n")
	case WFBytesToEnd:
		fmt.Fprintf(b, "        val %s = buf.copyOfRange(off, buf.size)\n", v)
		b.WriteString("        off = buf.size\n")
	}
}
