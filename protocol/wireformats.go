// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// One-shot byte-format specifications. Each WireFormat in a WireFormatSet
// describes a flat encoder/decoder pair emitted in every language target
// by cmd/protogen. Three shapes are supported:
//
//   - Plain: a single ordered list of fields (one encoder, one decoder).
//   - Variants: same fields-list shape but two variants distinguished by
//     a role flag. Encoder dispatches on the bool; decoders are split
//     into two functions because the receiver already knows its role.
//     (No live format uses this shape since T45 removed the backend
//     clientTag prefix; the framework keeps it for future role-split
//     formats.)
//   - Union: ASCII colon-delimited prefix dispatch (the relay greeting).
//     Each variant has a literal prefix and an optional ordered list of
//     colon-separated string suffix fields. Encoder is per-variant;
//     decoder dispatches on prefix and returns a tagged result.
//
// The field-kind set is deliberately tiny: u32_be, varint,
// string_with_varint_length (used in the stream header), string_to_end
// (used in connect: relay greeting and union suffixes), bytes_to_end
// (used in the datagram plaintext, where the trailing bytes carry the
// AEAD-encrypted payload bytes directly). Anything more elaborate
// belongs in a state-machine spec, not a one-shot byte format.

// WireFormatSet is a collection of one-shot byte-format specs loaded
// from a single YAML document.
type WireFormatSet struct {
	Name    string
	Formats []WireFormat
}

// WireFormat describes one byte format. Exactly one of (Fields,
// Variants, Union) is populated.
type WireFormat struct {
	Name     string
	Desc     string
	Fields   []WireField   // plain format
	Variants []WireVariant // role-dispatched format
	Union    []WireUnion   // prefix-dispatched format
}

// IsPlain reports whether the format is a plain (no role/variant) shape.
func (f *WireFormat) IsPlain() bool   { return len(f.Variants) == 0 && len(f.Union) == 0 }
func (f *WireFormat) IsVariant() bool { return len(f.Variants) > 0 }
func (f *WireFormat) IsUnion() bool   { return len(f.Union) > 0 }

// WireField describes one field in a plain or variant format body.
type WireField struct {
	Name string
	Kind WireFieldKind
}

// WireFieldKind enumerates supported field types. See package comment
// for the rationale on the small set.
type WireFieldKind string

const (
	WFU32BE       WireFieldKind = "u32_be"
	WFVarint      WireFieldKind = "varint"
	WFStringLenV  WireFieldKind = "string_with_varint_length"
	WFStringToEnd WireFieldKind = "string_to_end"
	WFBytesToEnd  WireFieldKind = "bytes_to_end"
)

// WireVariant is one role-shape of a variant format (today: backend / client).
type WireVariant struct {
	Name   string
	Fields []WireField
}

// WireUnion is one prefix-dispatched alternative in a union format.
type WireUnion struct {
	Name        string      // tag returned by decoder
	Prefix      string      // literal ASCII prefix
	Fields      []WireField // tail fields (no colon separation) — only string_to_end currently
	ColonSuffix []WireField // optional colon-separated string fields after the prefix
}

// --- YAML loader ---

type yamlWireFormatSet struct {
	Name    string               `yaml:"name"`
	Formats []yamlWireFormatSpec `yaml:"formats"`
}

type yamlWireFormatSpec struct {
	Name     string            `yaml:"name"`
	Desc     string            `yaml:"desc"`
	Fields   []yamlWireField   `yaml:"fields"`
	Variants []yamlWireVariant `yaml:"variants"`
	Union    []yamlWireUnion   `yaml:"union"`
}

type yamlWireField struct {
	Name string `yaml:"name"`
	Kind string `yaml:"kind"`
}

type yamlWireVariant struct {
	Name   string          `yaml:"name"`
	Fields []yamlWireField `yaml:"fields"`
}

type yamlWireUnion struct {
	Name        string          `yaml:"name"`
	Prefix      string          `yaml:"prefix"`
	Fields      []yamlWireField `yaml:"fields"`
	ColonSuffix []yamlWireField `yaml:"colon_suffix"`
}

// LoadWireFormatsYAML reads a wire-format set from a YAML file.
func LoadWireFormatsYAML(path string) (*WireFormatSet, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseWireFormatsYAML(data)
}

// ParseWireFormatsYAML parses a wire-format set from YAML bytes.
func ParseWireFormatsYAML(data []byte) (*WireFormatSet, error) {
	var ys yamlWireFormatSet
	if err := yaml.Unmarshal(data, &ys); err != nil {
		return nil, fmt.Errorf("parse YAML: %w", err)
	}
	set := &WireFormatSet{Name: ys.Name}
	for _, ysp := range ys.Formats {
		spec := WireFormat{Name: ysp.Name, Desc: ysp.Desc}
		for _, f := range ysp.Fields {
			spec.Fields = append(spec.Fields, WireField{Name: f.Name, Kind: WireFieldKind(f.Kind)})
		}
		for _, v := range ysp.Variants {
			wv := WireVariant{Name: v.Name}
			for _, f := range v.Fields {
				wv.Fields = append(wv.Fields, WireField{Name: f.Name, Kind: WireFieldKind(f.Kind)})
			}
			spec.Variants = append(spec.Variants, wv)
		}
		for _, u := range ysp.Union {
			wu := WireUnion{Name: u.Name, Prefix: u.Prefix}
			for _, f := range u.Fields {
				wu.Fields = append(wu.Fields, WireField{Name: f.Name, Kind: WireFieldKind(f.Kind)})
			}
			for _, f := range u.ColonSuffix {
				wu.ColonSuffix = append(wu.ColonSuffix, WireField{Name: f.Name, Kind: WireFieldKind(f.Kind)})
			}
			spec.Union = append(spec.Union, wu)
		}
		set.Formats = append(set.Formats, spec)
	}
	return set, nil
}

// Validate sanity-checks the parsed set.
func (s *WireFormatSet) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("wireformat set: name is required")
	}
	if len(s.Formats) == 0 {
		return fmt.Errorf("wireformat set %q: at least one format required", s.Name)
	}
	for _, f := range s.Formats {
		hits := 0
		if len(f.Fields) > 0 {
			hits++
		}
		if len(f.Variants) > 0 {
			hits++
		}
		if len(f.Union) > 0 {
			hits++
		}
		if hits != 1 {
			return fmt.Errorf("format %q: must declare exactly one of fields/variants/union", f.Name)
		}
		// Validate field kinds and field-order constraints. Variable-
		// length fields (string_to_end, bytes_to_end,
		// string_with_varint_length when last) are only well-defined at
		// the tail in greedy parsers; enforce that here.
		check := func(fields []WireField, where string) error {
			for i, fl := range fields {
				switch fl.Kind {
				case WFU32BE, WFVarint, WFStringLenV:
					// OK anywhere.
				case WFStringToEnd, WFBytesToEnd:
					if i != len(fields)-1 {
						return fmt.Errorf("%s: field %q kind %q must be last", where, fl.Name, fl.Kind)
					}
				default:
					return fmt.Errorf("%s: field %q has unknown kind %q", where, fl.Name, fl.Kind)
				}
			}
			return nil
		}
		if err := check(f.Fields, "format "+f.Name); err != nil {
			return err
		}
		for _, v := range f.Variants {
			if err := check(v.Fields, "format "+f.Name+" variant "+v.Name); err != nil {
				return err
			}
		}
		for _, u := range f.Union {
			if u.Prefix == "" {
				return fmt.Errorf("format %q union variant %q: prefix is required", f.Name, u.Name)
			}
			if err := check(u.Fields, "format "+f.Name+" union "+u.Name); err != nil {
				return err
			}
			for _, cs := range u.ColonSuffix {
				if cs.Kind != "string" && cs.Kind != "" {
					return fmt.Errorf("format %q union %q: colon_suffix fields must be string (got %q)", f.Name, u.Name, cs.Kind)
				}
			}
		}
	}
	return nil
}

// --- helpers shared across generators ---

func upperCamel(s string) string {
	parts := strings.Split(s, "_")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, "")
}

func lowerCamel(s string) string {
	out := upperCamel(s)
	if out == "" {
		return out
	}
	return strings.ToLower(out[:1]) + out[1:]
}

// --- Go generator ---

// ExportGo writes a Go source file with encoders/decoders for every
// format in the set. The output package is `pigeon` by default; callers
// can override via pkgName for testing.
func (s *WireFormatSet) ExportGo(w io.Writer, pkgName string) error {
	var b strings.Builder
	b.WriteString("// Copyright 2026 Marcelo Cantos\n")
	b.WriteString("// SPDX-License-Identifier: Apache-2.0\n\n")
	b.WriteString("// Code generated from protocol/wireformats.yaml. DO NOT EDIT.\n\n")
	fmt.Fprintf(&b, "package %s\n\n", pkgName)
	b.WriteString("import (\n")
	b.WriteString("\t\"encoding/binary\"\n")
	b.WriteString("\t\"errors\"\n")
	b.WriteString("\t\"strings\"\n")
	b.WriteString(")\n\n")
	// Suppress unused-import errors if a particular spec doesn't touch one.
	b.WriteString("var (\n")
	b.WriteString("\t_ = binary.BigEndian\n")
	b.WriteString("\t_ = errors.New\n")
	b.WriteString("\t_ = strings.HasPrefix\n")
	b.WriteString(")\n\n")

	for _, f := range s.Formats {
		switch {
		case f.IsPlain():
			goEmitPlain(&b, f)
		case f.IsVariant():
			goEmitVariant(&b, f)
		case f.IsUnion():
			goEmitUnion(&b, f)
		}
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func goEmitPlain(b *strings.Builder, f WireFormat) {
	upper := upperCamel(f.Name)
	fmt.Fprintf(b, "// Encode%s — %s\n", upper, f.Desc)
	fmt.Fprintf(b, "func Encode%s(", upper)
	for i, fl := range f.Fields {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(b, "%s %s", lowerCamel(fl.Name), goFieldType(fl.Kind))
	}
	b.WriteString(") []byte {\n")
	b.WriteString("\tvar buf []byte\n")
	for _, fl := range f.Fields {
		goEmitEncode(b, fl, lowerCamel(fl.Name))
	}
	b.WriteString("\treturn buf\n")
	b.WriteString("}\n\n")

	// Decoder
	fmt.Fprintf(b, "// Decode%s — %s\n", upper, f.Desc)
	fmt.Fprintf(b, "func Decode%s(buf []byte) (", upper)
	for _, fl := range f.Fields {
		fmt.Fprintf(b, "%s %s, ", lowerCamel(fl.Name), goFieldType(fl.Kind))
	}
	b.WriteString("err error) {\n")
	b.WriteString("\toff := 0\n")
	for _, fl := range f.Fields {
		goEmitDecode(b, fl, lowerCamel(fl.Name))
	}
	b.WriteString("\t_ = off\n")
	b.WriteString("\treturn\n")
	b.WriteString("}\n\n")
}

func goEmitVariant(b *strings.Builder, f WireFormat) {
	// Combined encoder that takes is_<variant0name> flag. Two distinct
	// decoders (DecodeXBackend / DecodeXClient), since the receiver
	// always knows its role.
	upper := upperCamel(f.Name)

	// Union of all field names across variants. We assume the variant
	// names are "backend" and "client" — and the contract is that all
	// fields in the second variant appear in the same order in the
	// first. We don't validate that strictly; instead, encoder takes
	// every field that appears anywhere, and the body branches.

	// Encoder: arguments are the union of all fields, in the order they
	// first appear across variants. is_<v0name> bool dispatches.
	type argSpec struct {
		Name string
		Kind WireFieldKind
	}
	var args []argSpec
	seen := map[string]bool{}
	add := func(fields []WireField) {
		for _, fl := range fields {
			if !seen[fl.Name] {
				args = append(args, argSpec{fl.Name, fl.Kind})
				seen[fl.Name] = true
			}
		}
	}
	for _, v := range f.Variants {
		add(v.Fields)
	}
	// is_<v0name> flag.
	flag := "is" + upperCamel(f.Variants[0].Name)

	fmt.Fprintf(b, "// Encode%s — %s\n", upper, f.Desc)
	fmt.Fprintf(b, "func Encode%s(%s bool", upper, flag)
	for _, a := range args {
		fmt.Fprintf(b, ", %s %s", lowerCamel(a.Name), goFieldType(a.Kind))
	}
	b.WriteString(") []byte {\n")
	b.WriteString("\tvar buf []byte\n")
	// Branch.
	for i, v := range f.Variants {
		if i == 0 {
			fmt.Fprintf(b, "\tif %s {\n", flag)
		} else {
			b.WriteString("\t} else {\n")
		}
		for _, fl := range v.Fields {
			b.WriteString("\t")
			goEmitEncode(b, fl, lowerCamel(fl.Name))
		}
	}
	b.WriteString("\t}\n")
	b.WriteString("\treturn buf\n")
	b.WriteString("}\n\n")

	// Per-variant decoders.
	for _, v := range f.Variants {
		fmt.Fprintf(b, "// Decode%s%s decodes the %s-side %s.\n",
			upper, upperCamel(v.Name), v.Name, f.Name)
		fmt.Fprintf(b, "func Decode%s%s(buf []byte) (", upper, upperCamel(v.Name))
		for _, fl := range v.Fields {
			fmt.Fprintf(b, "%s %s, ", lowerCamel(fl.Name), goFieldType(fl.Kind))
		}
		b.WriteString("err error) {\n")
		b.WriteString("\toff := 0\n")
		for _, fl := range v.Fields {
			goEmitDecode(b, fl, lowerCamel(fl.Name))
		}
		b.WriteString("\t_ = off\n")
		b.WriteString("\treturn\n")
		b.WriteString("}\n\n")
	}
}

func goEmitUnion(b *strings.Builder, f WireFormat) {
	upper := upperCamel(f.Name)

	// Result struct: tag + all colon_suffix/Fields fields across variants.
	fmt.Fprintf(b, "// %sVariant identifies which union variant a decoded\n", upper)
	fmt.Fprintf(b, "// %s message belongs to.\n", f.Name)
	fmt.Fprintf(b, "type %sVariant int\n\n", upper)
	b.WriteString("const (\n")
	for i, u := range f.Union {
		fmt.Fprintf(b, "\t%s%s %sVariant = %d\n", upper, upperCamel(u.Name), upper, i)
	}
	b.WriteString(")\n\n")

	// Decoded struct.
	fmt.Fprintf(b, "// %sDecoded is the result of decoding a %s.\n", upper, f.Name)
	fmt.Fprintf(b, "type %sDecoded struct {\n", upper)
	fmt.Fprintf(b, "\tVariant %sVariant\n", upper)
	// Union of all field names across all variants.
	seen := map[string]bool{}
	for _, u := range f.Union {
		for _, fl := range u.Fields {
			if !seen[fl.Name] {
				fmt.Fprintf(b, "\t%s %s\n", upperCamel(fl.Name), goFieldType(fl.Kind))
				seen[fl.Name] = true
			}
		}
		for _, fl := range u.ColonSuffix {
			if !seen[fl.Name] {
				fmt.Fprintf(b, "\t%s string\n", upperCamel(fl.Name))
				seen[fl.Name] = true
			}
		}
	}
	b.WriteString("}\n\n")

	// Per-variant encoder.
	for _, u := range f.Union {
		fmt.Fprintf(b, "// Encode%s%s — %s\n", upper, upperCamel(u.Name), f.Desc)
		fmt.Fprintf(b, "func Encode%s%s(", upper, upperCamel(u.Name))
		first := true
		for _, fl := range u.Fields {
			if !first {
				b.WriteString(", ")
			}
			first = false
			fmt.Fprintf(b, "%s %s", lowerCamel(fl.Name), goFieldType(fl.Kind))
		}
		for _, fl := range u.ColonSuffix {
			if !first {
				b.WriteString(", ")
			}
			first = false
			fmt.Fprintf(b, "%s string", lowerCamel(fl.Name))
		}
		b.WriteString(") []byte {\n")
		// Body: prefix + (colon-joined non-empty suffix segments
		// when applicable). The colon-suffix policy mirrors the
		// hand-rolled register-mux logic exactly: we emit
		//   <prefix>                                 if all empty
		//   <prefix>:<s0>:                           if only s0 set
		//   <prefix>::<s1>                           if only s1 set
		//   <prefix>:<s0>:<s1>                       if both set
		// Generalised: any non-empty colon suffix triggers the colon
		// separator; trailing-only-empty fields get omitted; leading-
		// empty fields are kept (with empty string) to preserve
		// positional meaning.
		if len(u.ColonSuffix) == 0 && len(u.Fields) == 0 {
			fmt.Fprintf(b, "\treturn []byte(%q)\n", u.Prefix)
		} else if len(u.ColonSuffix) == 0 {
			// Tail-only (e.g. connect:<id>).
			fmt.Fprintf(b, "\tvar sb strings.Builder\n")
			fmt.Fprintf(b, "\tsb.WriteString(%q)\n", u.Prefix)
			for _, fl := range u.Fields {
				switch fl.Kind {
				case WFStringToEnd:
					fmt.Fprintf(b, "\tsb.WriteString(%s)\n", lowerCamel(fl.Name))
				default:
					// Shouldn't reach: union tails are string_to_end today.
					fmt.Fprintf(b, "\tsb.Write([]byte(%s))\n", lowerCamel(fl.Name))
				}
			}
			b.WriteString("\treturn []byte(sb.String())\n")
		} else {
			// Colon-suffix policy.
			fmt.Fprintf(b, "\tvar sb strings.Builder\n")
			fmt.Fprintf(b, "\tsb.WriteString(%q)\n", u.Prefix)
			// Determine the maximal non-empty index.
			b.WriteString("\tsuffix := []string{")
			for i, fl := range u.ColonSuffix {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(lowerCamel(fl.Name))
			}
			b.WriteString("}\n")
			b.WriteString("\tanyNonEmpty := false\n")
			b.WriteString("\tfor _, p := range suffix { if p != \"\" { anyNonEmpty = true } }\n")
			b.WriteString("\tif anyNonEmpty {\n")
			b.WriteString("\t\tfor i := range suffix { sb.WriteByte(':'); sb.WriteString(suffix[i]) }\n")
			b.WriteString("\t}\n")
			b.WriteString("\treturn []byte(sb.String())\n")
		}
		b.WriteString("}\n\n")
	}

	// Combined decoder.
	fmt.Fprintf(b, "// Decode%s parses a %s into a tagged union.\n", upper, f.Name)
	fmt.Fprintf(b, "func Decode%s(buf []byte) (%sDecoded, error) {\n", upper, upper)
	fmt.Fprintf(b, "\tvar out %sDecoded\n", upper)
	b.WriteString("\ts := string(buf)\n")
	// Order matters: longer prefixes first. Sort by descending length.
	type prefixed struct {
		idx int
		u   WireUnion
	}
	ordered := make([]prefixed, 0, len(f.Union))
	for i, u := range f.Union {
		ordered = append(ordered, prefixed{i, u})
	}
	for i := 0; i < len(ordered); i++ {
		for j := i + 1; j < len(ordered); j++ {
			if len(ordered[j].u.Prefix) > len(ordered[i].u.Prefix) {
				ordered[i], ordered[j] = ordered[j], ordered[i]
			}
		}
	}
	for _, p := range ordered {
		u := p.u
		fmt.Fprintf(b, "\tif strings.HasPrefix(s, %q) {\n", u.Prefix)
		fmt.Fprintf(b, "\t\tout.Variant = %s%s\n", upper, upperCamel(u.Name))
		fmt.Fprintf(b, "\t\trest := s[len(%q):]\n", u.Prefix)
		if len(u.ColonSuffix) > 0 {
			// Split on ':' into up to len(suffix) parts. Allow bare
			// prefix (rest empty) → all suffix fields empty.
			b.WriteString("\t\tif rest != \"\" {\n")
			// rest must start with ':'.
			b.WriteString("\t\t\tif rest[0] != ':' { return out, errors.New(\"" + f.Name + ": malformed " + u.Name + ": missing ':' after prefix\") }\n")
			b.WriteString("\t\t\trest = rest[1:]\n")
			fmt.Fprintf(b, "\t\t\tparts := strings.SplitN(rest, \":\", %d)\n", len(u.ColonSuffix))
			for i, fl := range u.ColonSuffix {
				fmt.Fprintf(b, "\t\t\tif len(parts) > %d { out.%s = parts[%d] }\n", i, upperCamel(fl.Name), i)
			}
			b.WriteString("\t\t}\n")
		} else if len(u.Fields) > 0 {
			for _, fl := range u.Fields {
				fmt.Fprintf(b, "\t\tout.%s = rest\n", upperCamel(fl.Name))
			}
		}
		b.WriteString("\t\treturn out, nil\n")
		b.WriteString("\t}\n")
	}
	fmt.Fprintf(b, "\treturn out, errors.New(\"%s: unrecognised prefix\")\n", f.Name)
	b.WriteString("}\n\n")
}

func goFieldType(k WireFieldKind) string {
	switch k {
	case WFU32BE:
		return "uint32"
	case WFVarint:
		return "uint64"
	case WFStringLenV, WFStringToEnd:
		return "string"
	case WFBytesToEnd:
		return "[]byte"
	}
	return "unknown"
}

func goEmitEncode(b *strings.Builder, fl WireField, varName string) {
	switch fl.Kind {
	case WFU32BE:
		fmt.Fprintf(b, "\t{ var tmp [4]byte; binary.BigEndian.PutUint32(tmp[:], %s); buf = append(buf, tmp[:]...) }\n", varName)
	case WFVarint:
		fmt.Fprintf(b, "\t{ var tmp [binary.MaxVarintLen64]byte; n := binary.PutUvarint(tmp[:], %s); buf = append(buf, tmp[:n]...) }\n", varName)
	case WFStringLenV:
		fmt.Fprintf(b, "\t{ var tmp [binary.MaxVarintLen64]byte; n := binary.PutUvarint(tmp[:], uint64(len(%s))); buf = append(buf, tmp[:n]...); buf = append(buf, []byte(%s)...) }\n", varName, varName)
	case WFStringToEnd:
		fmt.Fprintf(b, "\tbuf = append(buf, []byte(%s)...)\n", varName)
	case WFBytesToEnd:
		fmt.Fprintf(b, "\tbuf = append(buf, %s...)\n", varName)
	}
}

func goEmitDecode(b *strings.Builder, fl WireField, varName string) {
	switch fl.Kind {
	case WFU32BE:
		fmt.Fprintf(b, "\tif len(buf)-off < 4 { err = errors.New(\"truncated u32_be %s\"); return }\n", fl.Name)
		fmt.Fprintf(b, "\t%s = binary.BigEndian.Uint32(buf[off:off+4])\n", varName)
		b.WriteString("\toff += 4\n")
	case WFVarint:
		fmt.Fprintf(b, "\t{ v, n := binary.Uvarint(buf[off:]); if n <= 0 { err = errors.New(\"bad varint %s\"); return }; %s = v; off += n }\n", fl.Name, varName)
	case WFStringLenV:
		fmt.Fprintf(b, "\t{ ln, n := binary.Uvarint(buf[off:]); if n <= 0 { err = errors.New(\"bad varint length %s\"); return }; off += n; if uint64(len(buf)-off) < ln { err = errors.New(\"truncated %s\"); return }; %s = string(buf[off:off+int(ln)]); off += int(ln) }\n", fl.Name, fl.Name, varName)
	case WFStringToEnd:
		fmt.Fprintf(b, "\t%s = string(buf[off:])\n", varName)
		b.WriteString("\toff = len(buf)\n")
	case WFBytesToEnd:
		fmt.Fprintf(b, "\t%s = append([]byte(nil), buf[off:]...)\n", varName)
		b.WriteString("\toff = len(buf)\n")
	}
}
