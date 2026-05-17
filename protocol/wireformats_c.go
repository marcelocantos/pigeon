// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"fmt"
	"io"
	"strings"
)

// ExportCHeader writes the C header for every format in the set.
func (s *WireFormatSet) ExportCHeader(w io.Writer) error {
	var b strings.Builder
	b.WriteString("// Copyright 2026 Marcelo Cantos\n")
	b.WriteString("// SPDX-License-Identifier: Apache-2.0\n")
	b.WriteString("//\n")
	b.WriteString("// Code generated from protocol/wireformats.yaml. DO NOT EDIT.\n\n")
	guard := "PIGEON_WIRE_GEN_H"
	fmt.Fprintf(&b, "#ifndef %s\n#define %s\n\n", guard, guard)
	b.WriteString("#include <stdbool.h>\n#include <stddef.h>\n#include <stdint.h>\n\n")
	b.WriteString("#ifdef __cplusplus\nextern \"C\" {\n#endif\n\n")

	// Forward-declare the uvarint primitives the generated bodies rely on.
	b.WriteString("// Provided by pigeon.h / c/src/pigeon.c.\n")
	b.WriteString("int pigeon_uvarint_encode(uint64_t v, uint8_t *buf, size_t buf_len);\n")
	b.WriteString("int pigeon_uvarint_decode(const uint8_t *buf, size_t buf_len, uint64_t *out);\n\n")

	for _, f := range s.Formats {
		switch {
		case f.IsPlain():
			cEmitPlainHeader(&b, f)
		case f.IsVariant():
			cEmitVariantHeader(&b, f)
		case f.IsUnion():
			cEmitUnionHeader(&b, f)
		}
	}
	b.WriteString("\n#ifdef __cplusplus\n}\n#endif\n")
	fmt.Fprintf(&b, "\n#endif // %s\n", guard)
	_, err := io.WriteString(w, b.String())
	return err
}

// ExportCImpl writes the C implementation.
func (s *WireFormatSet) ExportCImpl(w io.Writer) error {
	var b strings.Builder
	b.WriteString("// Copyright 2026 Marcelo Cantos\n")
	b.WriteString("// SPDX-License-Identifier: Apache-2.0\n")
	b.WriteString("//\n")
	b.WriteString("// Code generated from protocol/wireformats.yaml. DO NOT EDIT.\n\n")
	b.WriteString("#include \"pigeon/wire_gen.h\"\n")
	b.WriteString("#include <string.h>\n\n")
	for _, f := range s.Formats {
		switch {
		case f.IsPlain():
			cEmitPlainImpl(&b, f)
		case f.IsVariant():
			cEmitVariantImpl(&b, f)
		case f.IsUnion():
			cEmitUnionImpl(&b, f)
		}
	}
	_, err := io.WriteString(w, b.String())
	return err
}

// --- C helpers ---

func cFn(format, variant string) string {
	if variant == "" {
		return "pigeon_wire_" + format
	}
	return "pigeon_wire_" + format + "_" + variant
}

// Encoder signature builder for plain/variant. Returns (decl, body's
// "out cursor" name = "off").
func cEncoderArgs(fields []WireField, extraLeading string) string {
	var parts []string
	if extraLeading != "" {
		parts = append(parts, extraLeading)
	}
	for _, fl := range fields {
		switch fl.Kind {
		case WFU32BE:
			parts = append(parts, "uint32_t "+fl.Name)
		case WFVarint:
			parts = append(parts, "uint64_t "+fl.Name)
		case WFStringLenV, WFStringToEnd:
			parts = append(parts, "const char *"+fl.Name, "size_t "+fl.Name+"_len")
		case WFBytesToEnd:
			parts = append(parts, "const uint8_t *"+fl.Name, "size_t "+fl.Name+"_len")
		}
	}
	parts = append(parts, "uint8_t *out", "size_t out_len")
	return strings.Join(parts, ", ")
}

func cDecoderArgs(fields []WireField) string {
	var parts []string
	parts = append(parts, "const uint8_t *buf", "size_t buf_len")
	for _, fl := range fields {
		switch fl.Kind {
		case WFU32BE:
			parts = append(parts, "uint32_t *"+fl.Name)
		case WFVarint:
			parts = append(parts, "uint64_t *"+fl.Name)
		case WFStringLenV, WFStringToEnd:
			// NUL-terminated buffer + max length + actual length.
			parts = append(parts, "char *"+fl.Name+"_buf", "size_t "+fl.Name+"_buf_len", "size_t *"+fl.Name+"_len_out")
		case WFBytesToEnd:
			parts = append(parts, "uint8_t *"+fl.Name+"_buf", "size_t "+fl.Name+"_buf_len", "size_t *"+fl.Name+"_len_out")
		}
	}
	return strings.Join(parts, ", ")
}

func cEmitPlainHeader(b *strings.Builder, f WireFormat) {
	fmt.Fprintf(b, "// %s — %s\n", f.Name, f.Desc)
	fmt.Fprintf(b, "int %s(%s);\n",
		cFn(f.Name, "encode"), cEncoderArgs(f.Fields, ""))
	fmt.Fprintf(b, "int %s(%s);\n\n",
		cFn(f.Name, "decode"), cDecoderArgs(f.Fields))
}

func cEmitPlainImpl(b *strings.Builder, f WireFormat) {
	fmt.Fprintf(b, "int %s(%s)\n{\n",
		cFn(f.Name, "encode"), cEncoderArgs(f.Fields, ""))
	b.WriteString("    size_t off = 0;\n")
	for _, fl := range f.Fields {
		cEmitEncodeField(b, fl)
	}
	b.WriteString("    return (int)off;\n}\n\n")

	fmt.Fprintf(b, "int %s(%s)\n{\n",
		cFn(f.Name, "decode"), cDecoderArgs(f.Fields))
	b.WriteString("    size_t off = 0;\n")
	for _, fl := range f.Fields {
		cEmitDecodeField(b, fl)
	}
	b.WriteString("    return (int)off;\n}\n\n")
}

func cEmitVariantHeader(b *strings.Builder, f WireFormat) {
	flag := "is_" + f.Variants[0].Name
	// Union of arguments.
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
	fmt.Fprintf(b, "// %s — %s\n", f.Name, f.Desc)
	fmt.Fprintf(b, "int %s(%s);\n",
		cFn(f.Name, "encode"),
		cEncoderArgs(allFields, "bool "+flag))
	for _, v := range f.Variants {
		fmt.Fprintf(b, "int %s(%s);\n",
			cFn(f.Name, "decode_"+v.Name),
			cDecoderArgs(v.Fields))
	}
	b.WriteString("\n")
}

func cEmitVariantImpl(b *strings.Builder, f WireFormat) {
	flag := "is_" + f.Variants[0].Name
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
	fmt.Fprintf(b, "int %s(%s)\n{\n",
		cFn(f.Name, "encode"),
		cEncoderArgs(allFields, "bool "+flag))
	b.WriteString("    size_t off = 0;\n")
	for i, v := range f.Variants {
		if i == 0 {
			fmt.Fprintf(b, "    if (%s) {\n", flag)
		} else {
			b.WriteString("    } else {\n")
		}
		for _, fl := range v.Fields {
			cEmitEncodeFieldIndent(b, fl, "        ")
		}
	}
	b.WriteString("    }\n")
	b.WriteString("    return (int)off;\n}\n\n")

	// Per-variant decoders.
	for _, v := range f.Variants {
		fmt.Fprintf(b, "int %s(%s)\n{\n",
			cFn(f.Name, "decode_"+v.Name),
			cDecoderArgs(v.Fields))
		b.WriteString("    size_t off = 0;\n")
		for _, fl := range v.Fields {
			cEmitDecodeField(b, fl)
		}
		b.WriteString("    return (int)off;\n}\n\n")
	}
}

func cEmitUnionHeader(b *strings.Builder, f WireFormat) {
	upper := strings.ToUpper(f.Name)
	fmt.Fprintf(b, "// %s — %s\n", f.Name, f.Desc)
	fmt.Fprintf(b, "typedef enum {\n")
	for i, u := range f.Union {
		fmt.Fprintf(b, "    %s_%s = %d,\n", upper, strings.ToUpper(u.Name), i)
	}
	fmt.Fprintf(b, "    %s_UNKNOWN = -1\n", upper)
	b.WriteString("} pigeon_wire_" + f.Name + "_variant;\n\n")

	// Per-variant encoders. For colon-suffix unions, take all suffix
	// fields as nullable char *. For tail-only (e.g. connect: <id>),
	// take a single char* + len pair.
	for _, u := range f.Union {
		fmt.Fprintf(b, "int %s(",
			cFn(f.Name, "encode_"+u.Name))
		var args []string
		for _, fl := range u.Fields {
			args = append(args, "const char *"+fl.Name, "size_t "+fl.Name+"_len")
		}
		for _, fl := range u.ColonSuffix {
			args = append(args, "const char *"+fl.Name)
		}
		args = append(args, "uint8_t *out", "size_t out_len")
		b.WriteString(strings.Join(args, ", "))
		b.WriteString(");\n")
	}
	// Combined decoder writes the variant + (token|instance_id|...)
	// strings into caller-provided buffers. Each suffix field gets a
	// (buf, buf_len, len_out) trio; on a variant that doesn't carry the
	// field, the trio is left untouched (len_out=0 if non-NULL).
	fmt.Fprintf(b, "int %s(",
		cFn(f.Name, "decode"))
	args := []string{"const uint8_t *buf", "size_t buf_len", "pigeon_wire_" + f.Name + "_variant *out_variant"}
	allNames := map[string]bool{}
	for _, u := range f.Union {
		for _, fl := range u.Fields {
			if !allNames[fl.Name] {
				args = append(args, "char *"+fl.Name+"_buf", "size_t "+fl.Name+"_buf_len", "size_t *"+fl.Name+"_len_out")
				allNames[fl.Name] = true
			}
		}
		for _, fl := range u.ColonSuffix {
			if !allNames[fl.Name] {
				args = append(args, "char *"+fl.Name+"_buf", "size_t "+fl.Name+"_buf_len", "size_t *"+fl.Name+"_len_out")
				allNames[fl.Name] = true
			}
		}
	}
	b.WriteString(strings.Join(args, ", "))
	b.WriteString(");\n\n")
}

func cEmitUnionImpl(b *strings.Builder, f WireFormat) {
	for _, u := range f.Union {
		fmt.Fprintf(b, "int %s(",
			cFn(f.Name, "encode_"+u.Name))
		var args []string
		for _, fl := range u.Fields {
			args = append(args, "const char *"+fl.Name, "size_t "+fl.Name+"_len")
		}
		for _, fl := range u.ColonSuffix {
			args = append(args, "const char *"+fl.Name)
		}
		args = append(args, "uint8_t *out", "size_t out_len")
		b.WriteString(strings.Join(args, ", "))
		b.WriteString(")\n{\n")
		// Body — write the literal prefix.
		b.WriteString("    size_t off = 0;\n")
		fmt.Fprintf(b, "    const size_t prefix_len = %d;\n", len(u.Prefix))
		fmt.Fprintf(b, "    if (off + prefix_len > out_len) return -1;\n")
		fmt.Fprintf(b, "    memcpy(out + off, %q, prefix_len);\n", u.Prefix)
		b.WriteString("    off += prefix_len;\n")
		if len(u.Fields) > 0 {
			for _, fl := range u.Fields {
				// Tail string_to_end.
				if fl.Kind == WFStringToEnd {
					fmt.Fprintf(b, "    if (%s && %s_len > 0) {\n", fl.Name, fl.Name)
					fmt.Fprintf(b, "        if (off + %s_len > out_len) return -1;\n", fl.Name)
					fmt.Fprintf(b, "        memcpy(out + off, %s, %s_len);\n", fl.Name, fl.Name)
					fmt.Fprintf(b, "        off += %s_len;\n", fl.Name)
					b.WriteString("    }\n")
				}
			}
		}
		if len(u.ColonSuffix) > 0 {
			// Build the colon-suffix according to the same policy.
			n := len(u.ColonSuffix)
			fmt.Fprintf(b, "    const char *parts[%d];\n", n)
			fmt.Fprintf(b, "    size_t part_lens[%d];\n", n)
			fmt.Fprintf(b, "    bool any_non_empty = false;\n")
			for i, fl := range u.ColonSuffix {
				fmt.Fprintf(b, "    parts[%d] = %s;\n", i, fl.Name)
				fmt.Fprintf(b, "    part_lens[%d] = (%s ? strlen(%s) : 0);\n", i, fl.Name, fl.Name)
				fmt.Fprintf(b, "    if (parts[%d] && part_lens[%d] > 0) any_non_empty = true;\n", i, i)
			}
			b.WriteString("    if (any_non_empty) {\n")
			fmt.Fprintf(b, "        for (int i = 0; i < %d; i++) {\n", n)
			b.WriteString("            if (off + 1 > out_len) return -1;\n")
			b.WriteString("            out[off++] = ':';\n")
			b.WriteString("            if (parts[i] && part_lens[i] > 0) {\n")
			b.WriteString("                if (off + part_lens[i] > out_len) return -1;\n")
			b.WriteString("                memcpy(out + off, parts[i], part_lens[i]);\n")
			b.WriteString("                off += part_lens[i];\n")
			b.WriteString("            }\n")
			b.WriteString("        }\n")
			b.WriteString("    }\n")
		}
		b.WriteString("    return (int)off;\n}\n\n")
	}

	// Decoder.
	fmt.Fprintf(b, "int %s(",
		cFn(f.Name, "decode"))
	args := []string{"const uint8_t *buf", "size_t buf_len", "pigeon_wire_" + f.Name + "_variant *out_variant"}
	allNames := map[string]bool{}
	var allFieldList []WireField
	for _, u := range f.Union {
		for _, fl := range u.Fields {
			if !allNames[fl.Name] {
				args = append(args, "char *"+fl.Name+"_buf", "size_t "+fl.Name+"_buf_len", "size_t *"+fl.Name+"_len_out")
				allNames[fl.Name] = true
				allFieldList = append(allFieldList, fl)
			}
		}
		for _, fl := range u.ColonSuffix {
			if !allNames[fl.Name] {
				args = append(args, "char *"+fl.Name+"_buf", "size_t "+fl.Name+"_buf_len", "size_t *"+fl.Name+"_len_out")
				allNames[fl.Name] = true
				allFieldList = append(allFieldList, fl)
			}
		}
	}
	b.WriteString(strings.Join(args, ", "))
	b.WriteString(")\n{\n")
	// Zero len_out's up front.
	for _, fl := range allFieldList {
		fmt.Fprintf(b, "    if (%s_len_out) *%s_len_out = 0;\n", fl.Name, fl.Name)
	}
	// Order: longest prefix first.
	ordered := append([]WireUnion(nil), f.Union...)
	for i := 0; i < len(ordered); i++ {
		for j := i + 1; j < len(ordered); j++ {
			if len(ordered[j].Prefix) > len(ordered[i].Prefix) {
				ordered[i], ordered[j] = ordered[j], ordered[i]
			}
		}
	}
	upper := strings.ToUpper(f.Name)
	for _, u := range ordered {
		fmt.Fprintf(b, "    if (buf_len >= %d && memcmp(buf, %q, %d) == 0) {\n",
			len(u.Prefix), u.Prefix, len(u.Prefix))
		fmt.Fprintf(b, "        *out_variant = %s_%s;\n", upper, strings.ToUpper(u.Name))
		fmt.Fprintf(b, "        size_t off = %d;\n", len(u.Prefix))
		if len(u.Fields) > 0 {
			// Tail string_to_end.
			for _, fl := range u.Fields {
				if fl.Kind == WFStringToEnd {
					fmt.Fprintf(b, "        size_t %s_n = buf_len - off;\n", fl.Name)
					fmt.Fprintf(b, "        if (%s_n + 1 > %s_buf_len) return -1;\n", fl.Name, fl.Name)
					fmt.Fprintf(b, "        if (%s_n > 0) memcpy(%s_buf, buf + off, %s_n);\n", fl.Name, fl.Name, fl.Name)
					fmt.Fprintf(b, "        %s_buf[%s_n] = '\\0';\n", fl.Name, fl.Name)
					fmt.Fprintf(b, "        if (%s_len_out) *%s_len_out = %s_n;\n", fl.Name, fl.Name, fl.Name)
					b.WriteString("        off = buf_len;\n")
				}
			}
		}
		if len(u.ColonSuffix) > 0 {
			b.WriteString("        if (off < buf_len) {\n")
			b.WriteString("            if (buf[off] != ':') return -1;\n")
			b.WriteString("            off++;\n")
			b.WriteString("            size_t part_start = off;\n")
			fmt.Fprintf(b, "            size_t part_idx = 0;\n")
			fmt.Fprintf(b, "            const size_t expected = %d;\n", len(u.ColonSuffix))
			b.WriteString("            for (; off <= buf_len; off++) {\n")
			b.WriteString("                bool atEnd = (off == buf_len);\n")
			b.WriteString("                if (atEnd || (part_idx + 1 < expected && buf[off] == ':')) {\n")
			b.WriteString("                    size_t part_len = off - part_start;\n")
			b.WriteString("                    switch (part_idx) {\n")
			for i, fl := range u.ColonSuffix {
				fmt.Fprintf(b, "                    case %d:\n", i)
				fmt.Fprintf(b, "                        if (part_len + 1 > %s_buf_len) return -1;\n", fl.Name)
				fmt.Fprintf(b, "                        if (part_len > 0) memcpy(%s_buf, buf + part_start, part_len);\n", fl.Name)
				fmt.Fprintf(b, "                        %s_buf[part_len] = '\\0';\n", fl.Name)
				fmt.Fprintf(b, "                        if (%s_len_out) *%s_len_out = part_len;\n", fl.Name, fl.Name)
				b.WriteString("                        break;\n")
			}
			b.WriteString("                    }\n")
			b.WriteString("                    part_idx++;\n")
			b.WriteString("                    part_start = off + 1;\n")
			b.WriteString("                    if (atEnd) break;\n")
			b.WriteString("                }\n")
			b.WriteString("            }\n")
			b.WriteString("        }\n")
		}
		b.WriteString("        return (int)buf_len;\n")
		b.WriteString("    }\n")
	}
	fmt.Fprintf(b, "    *out_variant = %s_UNKNOWN;\n", upper)
	b.WriteString("    return -1;\n}\n\n")
}

func cEmitEncodeField(b *strings.Builder, fl WireField) {
	cEmitEncodeFieldIndent(b, fl, "    ")
}

func cEmitEncodeFieldIndent(b *strings.Builder, fl WireField, indent string) {
	switch fl.Kind {
	case WFU32BE:
		fmt.Fprintf(b, "%sif (off + 4 > out_len) return -1;\n", indent)
		fmt.Fprintf(b, "%sout[off++] = (uint8_t)(%s >> 24);\n", indent, fl.Name)
		fmt.Fprintf(b, "%sout[off++] = (uint8_t)(%s >> 16);\n", indent, fl.Name)
		fmt.Fprintf(b, "%sout[off++] = (uint8_t)(%s >> 8);\n", indent, fl.Name)
		fmt.Fprintf(b, "%sout[off++] = (uint8_t)(%s);\n", indent, fl.Name)
	case WFVarint:
		fmt.Fprintf(b, "%s{ int _n = pigeon_uvarint_encode(%s, out + off, out_len - off); if (_n < 0) return -1; off += (size_t)_n; }\n", indent, fl.Name)
	case WFStringLenV:
		fmt.Fprintf(b, "%s{ int _n = pigeon_uvarint_encode((uint64_t)%s_len, out + off, out_len - off); if (_n < 0) return -1; off += (size_t)_n; }\n", indent, fl.Name)
		fmt.Fprintf(b, "%sif (off + %s_len > out_len) return -1;\n", indent, fl.Name)
		fmt.Fprintf(b, "%sif (%s_len > 0) { if (!%s) return -1; memcpy(out + off, %s, %s_len); off += %s_len; }\n",
			indent, fl.Name, fl.Name, fl.Name, fl.Name, fl.Name)
	case WFStringToEnd:
		fmt.Fprintf(b, "%sif (off + %s_len > out_len) return -1;\n", indent, fl.Name)
		fmt.Fprintf(b, "%sif (%s_len > 0) { memcpy(out + off, %s, %s_len); off += %s_len; }\n",
			indent, fl.Name, fl.Name, fl.Name, fl.Name)
	case WFBytesToEnd:
		fmt.Fprintf(b, "%sif (off + %s_len > out_len) return -1;\n", indent, fl.Name)
		fmt.Fprintf(b, "%sif (%s_len > 0) { memcpy(out + off, %s, %s_len); off += %s_len; }\n",
			indent, fl.Name, fl.Name, fl.Name, fl.Name)
	}
}

func cEmitDecodeField(b *strings.Builder, fl WireField) {
	switch fl.Kind {
	case WFU32BE:
		fmt.Fprintf(b, "    if (buf_len - off < 4) return -1;\n")
		fmt.Fprintf(b, "    if (%s) *%s = ((uint32_t)buf[off] << 24) | ((uint32_t)buf[off+1] << 16) | ((uint32_t)buf[off+2] << 8) | (uint32_t)buf[off+3];\n",
			fl.Name, fl.Name)
		b.WriteString("    off += 4;\n")
	case WFVarint:
		fmt.Fprintf(b, "    { uint64_t _v = 0; int _n = pigeon_uvarint_decode(buf + off, buf_len - off, &_v); if (_n <= 0) return -1; if (%s) *%s = _v; off += (size_t)_n; }\n",
			fl.Name, fl.Name)
	case WFStringLenV:
		fmt.Fprintf(b, "    { uint64_t _len = 0; int _n = pigeon_uvarint_decode(buf + off, buf_len - off, &_len); if (_n <= 0) return -1; off += (size_t)_n;\n")
		fmt.Fprintf(b, "      if (_len > buf_len - off) return -1;\n")
		fmt.Fprintf(b, "      if (_len + 1 > %s_buf_len) return -1;\n", fl.Name)
		fmt.Fprintf(b, "      if (_len > 0) memcpy(%s_buf, buf + off, (size_t)_len);\n", fl.Name)
		fmt.Fprintf(b, "      %s_buf[_len] = '\\0';\n", fl.Name)
		fmt.Fprintf(b, "      if (%s_len_out) *%s_len_out = (size_t)_len;\n", fl.Name, fl.Name)
		b.WriteString("      off += (size_t)_len; }\n")
	case WFStringToEnd:
		fmt.Fprintf(b, "    { size_t _n = buf_len - off; if (_n + 1 > %s_buf_len) return -1; if (_n > 0) memcpy(%s_buf, buf + off, _n); %s_buf[_n] = '\\0'; if (%s_len_out) *%s_len_out = _n; off = buf_len; }\n",
			fl.Name, fl.Name, fl.Name, fl.Name, fl.Name)
	case WFBytesToEnd:
		fmt.Fprintf(b, "    { size_t _n = buf_len - off; if (_n > %s_buf_len) return -1; if (_n > 0) memcpy(%s_buf, buf + off, _n); if (%s_len_out) *%s_len_out = _n; off = buf_len; }\n",
			fl.Name, fl.Name, fl.Name, fl.Name)
	}
}
