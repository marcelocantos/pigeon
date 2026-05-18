// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"fmt"
	"io"
	"strings"
)

// ExportTypeScript writes a TypeScript module with native encoders/decoders.
func (s *WireFormatSet) ExportTypeScript(w io.Writer) error {
	var b strings.Builder
	b.WriteString("// Copyright 2026 Marcelo Cantos\n")
	b.WriteString("// SPDX-License-Identifier: Apache-2.0\n")
	b.WriteString("//\n")
	b.WriteString("// Code generated from protocol/wireformats.yaml. DO NOT EDIT.\n\n")

	// Varint helpers.
	b.WriteString("export function encodeUvarint(value: bigint): Uint8Array {\n")
	b.WriteString("  if (value < 0n) throw new Error(\"uvarint: negative value\");\n")
	b.WriteString("  const out: number[] = [];\n  let v = value;\n")
	b.WriteString("  while (v >= 0x80n) { out.push(Number((v & 0x7fn) | 0x80n)); v >>= 7n; }\n")
	b.WriteString("  out.push(Number(v));\n  return new Uint8Array(out);\n}\n\n")
	b.WriteString("export function decodeUvarint(buf: Uint8Array, offset = 0): [bigint, number] {\n")
	b.WriteString("  let v = 0n; let shift = 0n;\n")
	b.WriteString("  for (let i = 0; offset + i < buf.length; i++) {\n")
	b.WriteString("    if (i >= 10) throw new Error(\"uvarint: too many bytes\");\n")
	b.WriteString("    const b = buf[offset + i];\n")
	b.WriteString("    if ((b & 0x80) === 0) { v |= BigInt(b) << shift; return [v, i + 1]; }\n")
	b.WriteString("    v |= BigInt(b & 0x7f) << shift; shift += 7n;\n  }\n")
	b.WriteString("  throw new Error(\"uvarint: truncated\");\n}\n\n")

	for _, f := range s.Formats {
		switch {
		case f.IsPlain():
			tsEmitPlain(&b, f)
		case f.IsVariant():
			tsEmitVariant(&b, f)
		case f.IsUnion():
			tsEmitUnion(&b, f)
		}
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func tsType(k WireFieldKind) string {
	switch k {
	case WFU32BE:
		return "number"
	case WFVarint:
		return "bigint"
	case WFStringLenV, WFStringToEnd:
		return "string"
	case WFBytesToEnd:
		return "Uint8Array"
	}
	return "unknown"
}

func tsEmitPlain(b *strings.Builder, f WireFormat) {
	upper := upperCamel(f.Name)
	fmt.Fprintf(b, "export function encode%s(", upper)
	for i, fl := range f.Fields {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(b, "%s: %s", lowerCamel(fl.Name), tsType(fl.Kind))
	}
	b.WriteString("): Uint8Array {\n")
	b.WriteString("  const parts: Uint8Array[] = [];\n")
	for _, fl := range f.Fields {
		tsEmitEncodeField(b, fl)
	}
	b.WriteString("  let total = 0; for (const p of parts) total += p.length;\n")
	b.WriteString("  const out = new Uint8Array(total); let off = 0;\n")
	b.WriteString("  for (const p of parts) { out.set(p, off); off += p.length; }\n")
	b.WriteString("  return out;\n}\n\n")

	// Decoder result.
	fmt.Fprintf(b, "export interface %sDecoded {\n", upper)
	for _, fl := range f.Fields {
		fmt.Fprintf(b, "  %s: %s;\n", lowerCamel(fl.Name), tsType(fl.Kind))
	}
	b.WriteString("  consumed: number;\n}\n\n")

	fmt.Fprintf(b, "export function decode%s(buf: Uint8Array): %sDecoded {\n", upper, upper)
	b.WriteString("  let off = 0;\n")
	for _, fl := range f.Fields {
		tsEmitDecodeField(b, fl)
	}
	b.WriteString("  return { ")
	for i, fl := range f.Fields {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(b, "%s", lowerCamel(fl.Name))
	}
	b.WriteString(", consumed: off };\n}\n\n")
}

func tsEmitVariant(b *strings.Builder, f WireFormat) {
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
	fmt.Fprintf(b, "export function encode%s(%s: boolean", upper, flag)
	for _, fl := range allFields {
		fmt.Fprintf(b, ", %s: %s", lowerCamel(fl.Name), tsType(fl.Kind))
	}
	b.WriteString("): Uint8Array {\n")
	b.WriteString("  const parts: Uint8Array[] = [];\n")
	for i, v := range f.Variants {
		if i == 0 {
			fmt.Fprintf(b, "  if (%s) {\n", flag)
		} else {
			b.WriteString("  } else {\n")
		}
		for _, fl := range v.Fields {
			tsEmitEncodeFieldIndent(b, fl, "    ")
		}
	}
	b.WriteString("  }\n")
	b.WriteString("  let total = 0; for (const p of parts) total += p.length;\n")
	b.WriteString("  const out = new Uint8Array(total); let off = 0;\n")
	b.WriteString("  for (const p of parts) { out.set(p, off); off += p.length; }\n")
	b.WriteString("  return out;\n}\n\n")

	// Per-variant decoders.
	for _, v := range f.Variants {
		fmt.Fprintf(b, "export interface %s%sDecoded {\n", upper, upperCamel(v.Name))
		for _, fl := range v.Fields {
			fmt.Fprintf(b, "  %s: %s;\n", lowerCamel(fl.Name), tsType(fl.Kind))
		}
		b.WriteString("  consumed: number;\n}\n\n")
		fmt.Fprintf(b, "export function decode%s%s(buf: Uint8Array): %s%sDecoded {\n",
			upper, upperCamel(v.Name), upper, upperCamel(v.Name))
		b.WriteString("  let off = 0;\n")
		for _, fl := range v.Fields {
			tsEmitDecodeField(b, fl)
		}
		b.WriteString("  return { ")
		for i, fl := range v.Fields {
			if i > 0 {
				b.WriteString(", ")
			}
			fmt.Fprintf(b, "%s", lowerCamel(fl.Name))
		}
		b.WriteString(", consumed: off };\n}\n\n")
	}
}

func tsEmitUnion(b *strings.Builder, f WireFormat) {
	upper := upperCamel(f.Name)
	fmt.Fprintf(b, "export const enum %sVariant {", upper)
	for i, u := range f.Union {
		if i > 0 {
			b.WriteString(",")
		}
		fmt.Fprintf(b, " %s = %q", upperCamel(u.Name), u.Name)
	}
	b.WriteString(" }\n\n")

	fmt.Fprintf(b, "export interface %sDecoded {\n", upper)
	fmt.Fprintf(b, "  variant: %sVariant;\n", upper)
	allNames := map[string]bool{}
	allFields := []WireField{}
	for _, u := range f.Union {
		for _, fl := range u.Fields {
			if !allNames[fl.Name] {
				fmt.Fprintf(b, "  %s: %s;\n", lowerCamel(fl.Name), tsType(fl.Kind))
				allNames[fl.Name] = true
				allFields = append(allFields, fl)
			}
		}
		for _, fl := range u.ColonSuffix {
			if !allNames[fl.Name] {
				fmt.Fprintf(b, "  %s: string;\n", lowerCamel(fl.Name))
				allNames[fl.Name] = true
				allFields = append(allFields, WireField{Name: fl.Name, Kind: WFStringToEnd})
			}
		}
	}
	b.WriteString("}\n\n")

	// Per-variant encoders.
	for _, u := range f.Union {
		fmt.Fprintf(b, "export function encode%s%s(", upper, upperCamel(u.Name))
		first := true
		for _, fl := range u.Fields {
			if !first {
				b.WriteString(", ")
			}
			first = false
			fmt.Fprintf(b, "%s: string", lowerCamel(fl.Name))
		}
		for _, fl := range u.ColonSuffix {
			if !first {
				b.WriteString(", ")
			}
			first = false
			fmt.Fprintf(b, "%s: string", lowerCamel(fl.Name))
		}
		b.WriteString("): Uint8Array {\n")
		fmt.Fprintf(b, "  let s = %q;\n", u.Prefix)
		if len(u.Fields) > 0 {
			for _, fl := range u.Fields {
				fmt.Fprintf(b, "  s += %s;\n", lowerCamel(fl.Name))
			}
		}
		if len(u.ColonSuffix) > 0 {
			b.WriteString("  const parts = [")
			for i, fl := range u.ColonSuffix {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(lowerCamel(fl.Name))
			}
			b.WriteString("];\n")
			b.WriteString("  const anyNonEmpty = parts.some((p) => p !== \"\");\n")
			b.WriteString("  if (anyNonEmpty) { for (const p of parts) { s += \":\"; s += p; } }\n")
		}
		b.WriteString("  return new TextEncoder().encode(s);\n}\n\n")
	}

	// Combined decoder.
	fmt.Fprintf(b, "export function decode%s(buf: Uint8Array): %sDecoded {\n", upper, upper)
	b.WriteString("  const s = new TextDecoder().decode(buf);\n")
	ordered := append([]WireUnion(nil), f.Union...)
	for i := 0; i < len(ordered); i++ {
		for j := i + 1; j < len(ordered); j++ {
			if len(ordered[j].Prefix) > len(ordered[i].Prefix) {
				ordered[i], ordered[j] = ordered[j], ordered[i]
			}
		}
	}
	for _, u := range ordered {
		fmt.Fprintf(b, "  if (s.startsWith(%q)) {\n", u.Prefix)
		fmt.Fprintf(b, "    const rest = s.slice(%d);\n", len(u.Prefix))
		defaults := map[string]string{}
		for _, fl := range allFields {
			switch fl.Kind {
			case WFU32BE:
				defaults[fl.Name] = "0"
			case WFVarint:
				defaults[fl.Name] = "0n"
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
			b.WriteString("    let suffix: string[];\n")
			b.WriteString("    if (rest === \"\") {\n")
			fmt.Fprintf(b, "      suffix = new Array(%d).fill(\"\");\n", len(u.ColonSuffix))
			b.WriteString("    } else {\n")
			b.WriteString("      if (!rest.startsWith(\":\")) throw new Error(\"" + f.Name + ": malformed\");\n")
			b.WriteString("      const body = rest.slice(1);\n")
			fmt.Fprintf(b, "      const parts = body.split(\":\");\n")
			fmt.Fprintf(b, "      if (parts.length > %d) {\n", len(u.ColonSuffix))
			fmt.Fprintf(b, "        const merged = parts.slice(%d).join(\":\");\n", len(u.ColonSuffix)-1)
			fmt.Fprintf(b, "        suffix = parts.slice(0, %d);\n", len(u.ColonSuffix)-1)
			b.WriteString("        suffix.push(merged);\n")
			b.WriteString("      } else {\n")
			b.WriteString("        suffix = parts.slice();\n")
			fmt.Fprintf(b, "        while (suffix.length < %d) suffix.push(\"\");\n", len(u.ColonSuffix))
			b.WriteString("      }\n")
			b.WriteString("    }\n")
			for i, fl := range u.ColonSuffix {
				assigns[fl.Name] = fmt.Sprintf("suffix[%d]", i)
			}
		}
		fmt.Fprintf(b, "    return { variant: %sVariant.%s", upper, upperCamel(u.Name))
		for _, fl := range allFields {
			val, ok := assigns[fl.Name]
			if !ok {
				val = defaults[fl.Name]
			}
			fmt.Fprintf(b, ", %s: %s", lowerCamel(fl.Name), val)
		}
		b.WriteString(" };\n")
		b.WriteString("  }\n")
	}
	fmt.Fprintf(b, "  throw new Error(\"%s: unrecognised prefix\");\n", f.Name)
	b.WriteString("}\n\n")
}

func tsEmitEncodeField(b *strings.Builder, fl WireField) {
	tsEmitEncodeFieldIndent(b, fl, "  ")
}

func tsEmitEncodeFieldIndent(b *strings.Builder, fl WireField, indent string) {
	v := lowerCamel(fl.Name)
	switch fl.Kind {
	case WFU32BE:
		fmt.Fprintf(b, "%s{\n", indent)
		fmt.Fprintf(b, "%s  const t = new Uint8Array(4);\n", indent)
		fmt.Fprintf(b, "%s  new DataView(t.buffer).setUint32(0, %s, false);\n", indent, v)
		fmt.Fprintf(b, "%s  parts.push(t);\n", indent)
		fmt.Fprintf(b, "%s}\n", indent)
	case WFVarint:
		fmt.Fprintf(b, "%sparts.push(encodeUvarint(%s));\n", indent, v)
	case WFStringLenV:
		fmt.Fprintf(b, "%s{\n", indent)
		fmt.Fprintf(b, "%s  const bytes = new TextEncoder().encode(%s);\n", indent, v)
		fmt.Fprintf(b, "%s  parts.push(encodeUvarint(BigInt(bytes.length)));\n", indent)
		fmt.Fprintf(b, "%s  parts.push(bytes);\n", indent)
		fmt.Fprintf(b, "%s}\n", indent)
	case WFStringToEnd:
		fmt.Fprintf(b, "%sparts.push(new TextEncoder().encode(%s));\n", indent, v)
	case WFBytesToEnd:
		fmt.Fprintf(b, "%sparts.push(%s);\n", indent, v)
	}
}

func tsEmitDecodeField(b *strings.Builder, fl WireField) {
	v := lowerCamel(fl.Name)
	switch fl.Kind {
	case WFU32BE:
		fmt.Fprintf(b, "  if (buf.length - off < 4) throw new Error(\"truncated u32 %s\");\n", fl.Name)
		fmt.Fprintf(b, "  const %s = new DataView(buf.buffer, buf.byteOffset + off, 4).getUint32(0, false);\n", v)
		b.WriteString("  off += 4;\n")
	case WFVarint:
		fmt.Fprintf(b, "  const [%s, _vn] = decodeUvarint(buf, off);\n", v)
		b.WriteString("  off += _vn;\n")
	case WFStringLenV:
		b.WriteString("  const [_lenU, _ln] = decodeUvarint(buf, off);\n")
		b.WriteString("  off += _ln;\n")
		b.WriteString("  const _len = Number(_lenU);\n")
		b.WriteString("  if (buf.length - off < _len) throw new Error(\"truncated string\");\n")
		fmt.Fprintf(b, "  const %s = new TextDecoder().decode(buf.subarray(off, off + _len));\n", v)
		b.WriteString("  off += _len;\n")
	case WFStringToEnd:
		fmt.Fprintf(b, "  const %s = new TextDecoder().decode(buf.subarray(off));\n", v)
		b.WriteString("  off = buf.length;\n")
	case WFBytesToEnd:
		fmt.Fprintf(b, "  const %s = buf.slice(off);\n", v)
		b.WriteString("  off = buf.length;\n")
	}
}
