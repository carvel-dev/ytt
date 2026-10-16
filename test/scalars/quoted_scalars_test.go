// Copyright 2024 The Carvel Authors.
// SPDX-License-Identifier: Apache-2.0

package scalars_test

import (
	"strings"
	"testing"

	"carvel.dev/ytt/pkg/orderedmap"
	"carvel.dev/ytt/pkg/yamlmeta"
)

// TestQuotedScalarsRoundTrip pins the exact emitted style of scalar strings.
//
// A string holding text that reads back as a different value must be emitted
// quoted; text that round-trips as itself may keep its current style. Rows
// grouped under "out-of-range" currently lose their quotes.
func TestQuotedScalarsRoundTrip(t *testing.T) {
	type row struct {
		name string
		in   string // full YAML line
		out  string // expected emitted YAML line
		// readBack asserts the emitted document parses back to the exact
		// original string
		readBack string
	}

	quoted := func(name, s string) row {
		return row{name: name, in: "value: \"" + s + "\"", out: "value: \"" + s + "\"", readBack: s}
	}
	quotedIn := func(name, in, out string) row {
		return row{name: name, in: "value: \"" + in + "\"", out: "value: " + out}
	}
	plain := func(name, s string) row {
		return row{name: name, in: "value: " + s, out: "value: " + s}
	}
	plainTo := func(name, in, out string) row {
		return row{name: name, in: "value: " + in, out: "value: " + out}
	}

	rows := []row{
		// ─── quoted, value within 64-bit range: already emitted quoted today ───
		quoted("int64 max", "9223372036854775807"),
		quoted("int64 min", "-9223372036854775808"),
		quoted("int64 max plus one, unsigned parse", "9223372036854775808"),
		quoted("uint64 max", "18446744073709551615"),
		quoted("signed hex within 64 bits", "0x5aAeb6053F3E94C9"),
		quoted("binary within 64 bits", "0b101010101010101010101010101010101010101010101010101"),
		quoted("decimal beyond uint64 within float", "123456789012345678901234567890"),
		quoted("underscored decimal within float", "1_000_000_000_000_000_000_000_000_000_000"),
		quoted("float underflow", "1e-999"),
		quoted("float underflow with leading dot", ".5e-999"),

		// ─── quoted, value beyond 64-bit range: currently emitted unquoted, ───
		// ─── silently changing the value ───
		quoted("hex address beyond 64 bits", "0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed"),
		quoted("hex upper-case beyond 64 bits", "0X5AAEB6053F3E94C9B9A09F33669435E7EF1BEAED"),
		quoted("hex plus sign beyond 64 bits", "+0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed"),
		quoted("hex negative beyond 64 bits", "-0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed"),
		quoted("hex underscored beyond 64 bits", "0x5a_Aeb6053F3E94C9b9A09f33669435E7Ef1BeAed"),
		quoted("octal beyond 64 bits", "0o777777777777777777777777777777777777777777777"),
		quoted("octal negative beyond 64 bits", "-0o777777777777777777777777777777777777777777777"),
		quoted("octal upper-case prefix beyond 64 bits", "0O777777777777777777777777777777777777777777777"),
		quoted("binary beyond 64 bits", "0b10101010101010101010101010101010101010101010101010101010101010101"),
		quoted("float overflow", "1e999"),
		quoted("float overflow plus sign", "+1e999"),
		quoted("float overflow negative", "-1e999"),
		quoted("float overflow with fraction", "1.8e308"),
		quoted("float overflow with leading dot", ".5e999"),

		// ─── exact numbers: keep emitting plain ───
		plain("plain int", "123"),
		plain("plain negative int", "-49"),
		plain("plain float", "123.123"),
		plainTo("plain hex within 64 bits", "0x5aAeb6053F3E94C9", "6534360243013326025"),
		plainTo("plain binary within 64 bits", "0b1111111111111111111111", "4194303"),
		plainTo("plain float notation", "1e5", "100000"),
		plainTo("plain decimal beyond uint64", "18446744073709551616", "1.8446744073709552e+19"),

		// ─── non-numeric strings: keep emitting plain ───
		quotedIn("hex-shaped text", "0xZZZ", "0xZZZ"),
		quotedIn("binary prefix only", "0b2", "0b2"),
		quotedIn("hex prefix only", "0x", "0x"),
		quotedIn("hex with trailing words", "0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed extra", "0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed extra"),
		quotedIn("words", "not a number", "not a number"),

		// ─── long non-numeric strings are unaffected ───
		quotedIn("long words", strings.Repeat("abcdef", 20), strings.Repeat("abcdef", 20)),
	}

	for _, r := range rows {
		r := r
		t.Run(r.name, func(t *testing.T) {
			docSet, err := yamlmeta.NewParser(yamlmeta.ParserOpts{}).ParseBytes([]byte(r.in+"\n"), "test.yml")
			if err != nil {
				t.Fatalf("Expected parse to succeed: %s", err)
			}

			bs, err := docSet.AsBytes()
			if err != nil {
				t.Fatalf("Expected emit to succeed: %s", err)
			}
			out := strings.TrimSuffix(string(bs), "\n")

			if out != r.out {
				t.Errorf("Expected %q to be emitted as %q, got %q", r.in, r.out, out)
			}

			if r.readBack == "" {
				return
			}
			// the emitted document must read back as the original string
			reSet, err := yamlmeta.NewParser(yamlmeta.ParserOpts{}).ParseBytes([]byte(out+"\n"), "test.yml")
			if err != nil {
				t.Fatalf("Expected re-parse of emitted output to succeed: %s", err)
			}
			m, ok := yamlmeta.NewGoFromAST(reSet.Items[0].Value).(*orderedmap.Map)
			if !ok {
				t.Fatalf("Expected document value to be a map, got %T", reSet.Items[0].Value)
			}
			value, found := m.Get("value")
			if !found {
				t.Fatalf("Expected key 'value' to be present")
			}
			if s, ok := value.(string); !ok || s != r.readBack {
				t.Errorf("Expected round trip to preserve %q, got %#v (%T)", r.readBack, value, value)
			}
		})
	}
}

// TestQuotedScalarsRoundTripInArrays pins the same contract for array items.
func TestQuotedScalarsRoundTripInArrays(t *testing.T) {
	doc := "- \"0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed\"\n- \"1e999\"\n"
	expected := "- \"0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed\"\n- \"1e999\"\n"

	docSet, err := yamlmeta.NewParser(yamlmeta.ParserOpts{}).ParseBytes([]byte(doc), "test.yml")
	if err != nil {
		t.Fatalf("Expected parse to succeed: %s", err)
	}

	bs, err := docSet.AsBytes()
	if err != nil {
		t.Fatalf("Expected emit to succeed: %s", err)
	}

	if string(bs) != expected {
		t.Errorf("Expected array items to be emitted quoted:\n  in:  %q\n  out: %q", expected, string(bs))
	}
}
