// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"fmt"
	"io"
	"strings"
)

// ExportSwift writes a Swift source file with native encoders/decoders
// for every format in the set. The generated code is dependency-free
// (no CPigeon import); it implements the varint primitives inline.
func (s *WireFormatSet) ExportSwift(w io.Writer) error {
	var b strings.Builder
	b.WriteString("// Copyright 2026 Marcelo Cantos\n")
	b.WriteString("// SPDX-License-Identifier: Apache-2.0\n")
	b.WriteString("//\n")
	b.WriteString("// Code generated from protocol/wireformats.yaml. DO NOT EDIT.\n\n")
	b.WriteString("import Foundation\n\n")

	b.WriteString("public enum PigeonWireError: Error {\n")
	b.WriteString("    case truncated\n    case malformed\n    case bufferTooSmall\n}\n\n")

	b.WriteString("public enum PigeonWire {\n\n")
	// Inline varint helpers.
	b.WriteString("    public static func encodeUvarint(_ v: UInt64) -> Data {\n")
	b.WriteString("        var out = Data()\n        var x = v\n")
	b.WriteString("        while x >= 0x80 {\n")
	b.WriteString("            out.append(UInt8((x & 0x7f) | 0x80))\n")
	b.WriteString("            x >>= 7\n        }\n")
	b.WriteString("        out.append(UInt8(x))\n        return out\n    }\n\n")
	b.WriteString("    public static func decodeUvarint(_ data: Data) throws -> (value: UInt64, consumed: Int) {\n")
	b.WriteString("        var v: UInt64 = 0\n        var shift: UInt64 = 0\n")
	b.WriteString("        for (i, b) in data.enumerated() {\n")
	b.WriteString("            if i >= 10 { throw PigeonWireError.malformed }\n")
	b.WriteString("            if b & 0x80 == 0 { v |= UInt64(b) << shift; return (v, i + 1) }\n")
	b.WriteString("            v |= UInt64(b & 0x7f) << shift\n            shift += 7\n        }\n")
	b.WriteString("        throw PigeonWireError.truncated\n    }\n\n")

	for _, f := range s.Formats {
		switch {
		case f.IsPlain():
			swiftEmitPlain(&b, f)
		case f.IsVariant():
			swiftEmitVariant(&b, f)
		case f.IsUnion():
			swiftEmitUnion(&b, f)
		}
	}
	b.WriteString("}\n")
	_, err := io.WriteString(w, b.String())
	return err
}

func wireSwiftType(k WireFieldKind) string {
	switch k {
	case WFU32BE:
		return "UInt32"
	case WFVarint:
		return "UInt64"
	case WFStringLenV, WFStringToEnd:
		return "String"
	case WFBytesToEnd:
		return "Data"
	}
	return "Any"
}

func swiftEmitPlain(b *strings.Builder, f WireFormat) {
	fn := lowerCamel(f.Name)
	swiftEmitEncodeFn(b, "encode"+upperCamel(f.Name), f.Fields, false, "")
	swiftEmitDecodeFn(b, "decode"+upperCamel(f.Name), f.Fields)
	_ = fn
}

func swiftEmitVariant(b *strings.Builder, f WireFormat) {
	// Combined encoder with is_<v0name> bool first parameter.
	flag := "is" + upperCamel(f.Variants[0].Name)
	upper := upperCamel(f.Name)
	// Union of fields across variants.
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
	// Build signature
	fmt.Fprintf(b, "    public static func encode%s(%s: Bool", upper, flag)
	for _, fl := range allFields {
		fmt.Fprintf(b, ", %s: %s", lowerCamel(fl.Name), wireSwiftType(fl.Kind))
	}
	b.WriteString(") -> Data {\n")
	b.WriteString("        var out = Data()\n")
	for i, v := range f.Variants {
		if i == 0 {
			fmt.Fprintf(b, "        if %s {\n", flag)
		} else {
			b.WriteString("        } else {\n")
		}
		for _, fl := range v.Fields {
			swiftEmitEncodeField(b, fl, "            ")
		}
	}
	b.WriteString("        }\n        return out\n    }\n\n")

	for _, v := range f.Variants {
		swiftEmitDecodeFn(b, "decode"+upper+upperCamel(v.Name), v.Fields)
	}
}

func swiftEmitEncodeFn(b *strings.Builder, fn string, fields []WireField, takesFlag bool, flagName string) {
	fmt.Fprintf(b, "    public static func %s(", fn)
	first := true
	if takesFlag {
		fmt.Fprintf(b, "%s: Bool", flagName)
		first = false
	}
	for _, fl := range fields {
		if !first {
			b.WriteString(", ")
		}
		first = false
		fmt.Fprintf(b, "%s: %s", lowerCamel(fl.Name), wireSwiftType(fl.Kind))
	}
	b.WriteString(") -> Data {\n")
	b.WriteString("        var out = Data()\n")
	for _, fl := range fields {
		swiftEmitEncodeField(b, fl, "        ")
	}
	b.WriteString("        return out\n    }\n\n")
}

func swiftEmitDecodeFn(b *strings.Builder, fn string, fields []WireField) {
	// Return tuple shape (...) or struct? Use tuple for plain, throws on error.
	fmt.Fprintf(b, "    public static func %s(_ data: Data) throws -> (", fn)
	for i, fl := range fields {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(b, "%s: %s", lowerCamel(fl.Name), wireSwiftType(fl.Kind))
	}
	b.WriteString(", consumed: Int) {\n")
	b.WriteString("        var off = 0\n")
	for _, fl := range fields {
		swiftEmitDecodeField(b, fl)
	}
	// Return tuple.
	b.WriteString("        return (")
	for i, fl := range fields {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(lowerCamel(fl.Name))
	}
	b.WriteString(", off)\n")
	b.WriteString("    }\n\n")
}

func swiftEmitEncodeField(b *strings.Builder, fl WireField, indent string) {
	v := lowerCamel(fl.Name)
	switch fl.Kind {
	case WFU32BE:
		fmt.Fprintf(b, "%sout.append(contentsOf: [UInt8(truncatingIfNeeded: %s >> 24), UInt8(truncatingIfNeeded: %s >> 16), UInt8(truncatingIfNeeded: %s >> 8), UInt8(truncatingIfNeeded: %s)])\n",
			indent, v, v, v, v)
	case WFVarint:
		fmt.Fprintf(b, "%sout.append(PigeonWire.encodeUvarint(%s))\n", indent, v)
	case WFStringLenV:
		fmt.Fprintf(b, "%sdo {\n", indent)
		fmt.Fprintf(b, "%s    let bytes = Array(%s.utf8)\n", indent, v)
		fmt.Fprintf(b, "%s    out.append(PigeonWire.encodeUvarint(UInt64(bytes.count)))\n", indent)
		fmt.Fprintf(b, "%s    out.append(contentsOf: bytes)\n", indent)
		fmt.Fprintf(b, "%s}\n", indent)
	case WFStringToEnd:
		fmt.Fprintf(b, "%sout.append(contentsOf: Array(%s.utf8))\n", indent, v)
	case WFBytesToEnd:
		fmt.Fprintf(b, "%sout.append(%s)\n", indent, v)
	}
}

func swiftEmitDecodeField(b *strings.Builder, fl WireField) {
	v := lowerCamel(fl.Name)
	switch fl.Kind {
	case WFU32BE:
		fmt.Fprintf(b, "        if data.count - off < 4 { throw PigeonWireError.truncated }\n")
		fmt.Fprintf(b, "        let %s: UInt32 = (UInt32(data[data.startIndex + off]) << 24)\n", v)
		fmt.Fprintf(b, "            | (UInt32(data[data.startIndex + off + 1]) << 16)\n")
		fmt.Fprintf(b, "            | (UInt32(data[data.startIndex + off + 2]) << 8)\n")
		fmt.Fprintf(b, "            | UInt32(data[data.startIndex + off + 3])\n")
		b.WriteString("        off += 4\n")
	case WFVarint:
		fmt.Fprintf(b, "        let (%s, _vn) = try PigeonWire.decodeUvarint(data.subdata(in: (data.startIndex + off)..<data.endIndex))\n", v)
		b.WriteString("        off += _vn\n")
	case WFStringLenV:
		fmt.Fprintf(b, "        let (_lenU, _ln) = try PigeonWire.decodeUvarint(data.subdata(in: (data.startIndex + off)..<data.endIndex))\n")
		b.WriteString("        off += _ln\n")
		b.WriteString("        let _len = Int(_lenU)\n")
		b.WriteString("        if data.count - off < _len { throw PigeonWireError.truncated }\n")
		fmt.Fprintf(b, "        let %s = String(data: data.subdata(in: (data.startIndex + off)..<(data.startIndex + off + _len)), encoding: .utf8) ?? \"\"\n", v)
		b.WriteString("        off += _len\n")
	case WFStringToEnd:
		fmt.Fprintf(b, "        let %s = String(data: data.subdata(in: (data.startIndex + off)..<data.endIndex), encoding: .utf8) ?? \"\"\n", v)
		b.WriteString("        off = data.count\n")
	case WFBytesToEnd:
		fmt.Fprintf(b, "        let %s = data.subdata(in: (data.startIndex + off)..<data.endIndex)\n", v)
		b.WriteString("        off = data.count\n")
	}
}

func swiftEmitUnion(b *strings.Builder, f WireFormat) {
	upper := upperCamel(f.Name)
	// Enum for variants.
	fmt.Fprintf(b, "    public enum %sVariant {\n", upper)
	for _, u := range f.Union {
		fmt.Fprintf(b, "        case %s\n", lowerCamel(u.Name))
	}
	b.WriteString("    }\n\n")

	// Decoded result struct: include variant + every union field.
	fmt.Fprintf(b, "    public struct %sDecoded {\n", upper)
	fmt.Fprintf(b, "        public let variant: %sVariant\n", upper)
	seen := map[string]bool{}
	allFields := []WireField{}
	for _, u := range f.Union {
		for _, fl := range u.Fields {
			if !seen[fl.Name] {
				fmt.Fprintf(b, "        public let %s: %s\n", lowerCamel(fl.Name), wireSwiftType(fl.Kind))
				seen[fl.Name] = true
				allFields = append(allFields, fl)
			}
		}
		for _, fl := range u.ColonSuffix {
			if !seen[fl.Name] {
				fmt.Fprintf(b, "        public let %s: String\n", lowerCamel(fl.Name))
				seen[fl.Name] = true
				allFields = append(allFields, WireField{Name: fl.Name, Kind: WFStringToEnd})
			}
		}
	}
	b.WriteString("    }\n\n")

	for _, u := range f.Union {
		fmt.Fprintf(b, "    public static func encode%s%s(", upper, upperCamel(u.Name))
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
		b.WriteString(") -> Data {\n")
		fmt.Fprintf(b, "        var s = %q\n", u.Prefix)
		if len(u.Fields) > 0 {
			for _, fl := range u.Fields {
				fmt.Fprintf(b, "        s += %s\n", lowerCamel(fl.Name))
			}
		}
		if len(u.ColonSuffix) > 0 {
			b.WriteString("        let parts: [String] = [")
			for i, fl := range u.ColonSuffix {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(lowerCamel(fl.Name))
			}
			b.WriteString("]\n")
			b.WriteString("        var anyNonEmpty = false\n")
			b.WriteString("        for p in parts { if !p.isEmpty { anyNonEmpty = true } }\n")
			b.WriteString("        if anyNonEmpty {\n")
			b.WriteString("            for p in parts { s += \":\"; s += p }\n")
			b.WriteString("        }\n")
		}
		b.WriteString("        return Data(s.utf8)\n    }\n\n")
	}

	// Decoder.
	fmt.Fprintf(b, "    public static func decode%s(_ data: Data) throws -> %sDecoded {\n", upper, upper)
	b.WriteString("        let s = String(data: data, encoding: .utf8) ?? \"\"\n")
	// Sort union variants by descending prefix length.
	ordered := append([]WireUnion(nil), f.Union...)
	for i := 0; i < len(ordered); i++ {
		for j := i + 1; j < len(ordered); j++ {
			if len(ordered[j].Prefix) > len(ordered[i].Prefix) {
				ordered[i], ordered[j] = ordered[j], ordered[i]
			}
		}
	}
	for _, u := range ordered {
		fmt.Fprintf(b, "        if s.hasPrefix(%q) {\n", u.Prefix)
		fmt.Fprintf(b, "            let rest = String(s.dropFirst(%d))\n", len(u.Prefix))
		// Initialise defaults for every field.
		defaults := map[string]string{}
		for _, fl := range allFields {
			switch fl.Kind {
			case WFU32BE:
				defaults[fl.Name] = "0"
			case WFVarint:
				defaults[fl.Name] = "0"
			default:
				defaults[fl.Name] = "\"\""
			}
		}
		// Per-variant assignments.
		assigns := map[string]string{}
		for k, v := range defaults {
			assigns[k] = v
		}
		if len(u.Fields) > 0 {
			for _, fl := range u.Fields {
				assigns[fl.Name] = "rest"
			}
		}
		// For colon-suffix: declare named variables.
		if len(u.ColonSuffix) > 0 {
			b.WriteString("            var _suffix: [String] = []\n")
			b.WriteString("            if !rest.isEmpty {\n")
			b.WriteString("                if !rest.hasPrefix(\":\") { throw PigeonWireError.malformed }\n")
			b.WriteString("                let body = String(rest.dropFirst())\n")
			fmt.Fprintf(b, "                _suffix = body.components(separatedBy: \":\")\n")
			fmt.Fprintf(b, "                while _suffix.count < %d { _suffix.append(\"\") }\n", len(u.ColonSuffix))
			fmt.Fprintf(b, "                if _suffix.count > %d {\n", len(u.ColonSuffix))
			// Merge any extras into the final field.
			fmt.Fprintf(b, "                    _suffix[%d] = _suffix[%d...].joined(separator: \":\")\n", len(u.ColonSuffix)-1, len(u.ColonSuffix)-1)
			fmt.Fprintf(b, "                    _suffix = Array(_suffix.prefix(%d))\n", len(u.ColonSuffix))
			b.WriteString("                }\n")
			b.WriteString("            } else {\n")
			fmt.Fprintf(b, "                _suffix = Array(repeating: \"\", count: %d)\n", len(u.ColonSuffix))
			b.WriteString("            }\n")
			for i, fl := range u.ColonSuffix {
				assigns[fl.Name] = fmt.Sprintf("_suffix[%d]", i)
			}
		}
		// Build the decoded result.
		fmt.Fprintf(b, "            return %sDecoded(\n", upper)
		fmt.Fprintf(b, "                variant: .%s", lowerCamel(u.Name))
		for _, fl := range allFields {
			val, ok := assigns[fl.Name]
			if !ok {
				val = defaults[fl.Name]
			}
			fmt.Fprintf(b, ",\n                %s: %s", lowerCamel(fl.Name), val)
		}
		b.WriteString("\n            )\n")
		b.WriteString("        }\n")
	}
	b.WriteString("        throw PigeonWireError.malformed\n")
	b.WriteString("    }\n\n")
}
