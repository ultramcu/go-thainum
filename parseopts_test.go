package thainum

import (
	"errors"
	"testing"
)

// Tests for the ParseOption flags (Strict, Lenient, AllowColloquial), ported
// from the Dart 0.5.3 oracle (thainum.dart/test/parse_decimal_test.dart,
// groups 'allowColloquial' and 'lenient'). Options are applied via the
// variadic ParseInt / ParseDecimal signatures.

func TestAllowColloquial_NuengReadsAsOne(t *testing.T) {
	cases := []struct {
		in   string
		want int64
	}{
		{"ร้อยนึง", 101},
		{"นึง", 1},
	}
	for _, c := range cases {
		got, err := ParseInt(c.in, AllowColloquial())
		if err != nil {
			t.Errorf("ParseInt(%q, AllowColloquial) error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("ParseInt(%q, AllowColloquial) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestAllowColloquial_DefaultRejectsNueng(t *testing.T) {
	// Without AllowColloquial, นึง is not a recognized word.
	_, err := ParseInt("ร้อยนึง")
	if err == nil {
		t.Fatalf("ParseInt(%q) = nil error, want error", "ร้อยนึง")
	}
	if !errors.Is(err, ErrParse) {
		t.Errorf("ParseInt(%q) error %v should wrap ErrParse", "ร้อยนึง", err)
	}
}

func TestLenient_StripsInternalSpaces(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want int64
	}{
		{"ascii space", "ยี่สิบ เอ็ด", 21},
		{"ascii space 2", "หนึ่ง ร้อย", 100},
		{"nbsp", "ยี่สิบ เอ็ด", 21}, // U+00A0 non-breaking space
		{"zwsp", "หนึ่ง​พัน", 1000}, // U+200B zero-width space
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ParseInt(c.in, Lenient())
			if err != nil {
				t.Fatalf("ParseInt(%q, Lenient) error: %v", c.in, err)
			}
			if got != c.want {
				t.Errorf("ParseInt(%q, Lenient) = %d, want %d", c.in, got, c.want)
			}
		})
	}
}

func TestLenient_DefaultRejectsSpacedInput(t *testing.T) {
	_, err := ParseInt("ยี่สิบ เอ็ด")
	if err == nil {
		t.Fatalf("ParseInt(%q) = nil error, want error", "ยี่สิบ เอ็ด")
	}
	if !errors.Is(err, ErrParse) {
		t.Errorf("ParseInt(%q) error %v should wrap ErrParse", "ยี่สิบ เอ็ด", err)
	}
}

func TestLenient_AppliesToParseDecimal(t *testing.T) {
	got, err := ParseDecimal("สิบสอง จุด สามสี่", Lenient())
	if err != nil {
		t.Fatalf("ParseDecimal(Lenient) error: %v", err)
	}
	if got != "12.34" {
		t.Errorf("ParseDecimal(Lenient) = %q, want %q", got, "12.34")
	}
}

func TestStrict_RejectsNonStandardTensForms(t *testing.T) {
	// Per issue #11, Strict() rejects a tens place fed by หนึ่ง (หนึ่งสิบ = 10)
	// or by a plain สอง written instead of ยี่ (สองสิบ = 20). These parse fine
	// under the default (lenient) reader.
	bad := []string{"สองสิบ", "หนึ่งสิบ"}
	for _, in := range bad {
		t.Run(in, func(t *testing.T) {
			_, err := ParseInt(in, Strict())
			if err == nil {
				t.Fatalf("ParseInt(%q, Strict) = nil error, want error", in)
			}
			if !errors.Is(err, ErrParse) {
				t.Errorf("ParseInt(%q, Strict) error %v should wrap ErrParse", in, err)
			}
		})
	}
}

func TestStrict_AcceptsCanonicalTensForms(t *testing.T) {
	// The canonical forms สิบ, ยี่สิบ, ยี่สิบเอ็ด still parse under Strict.
	cases := []struct {
		in   string
		want int64
	}{
		{"สิบ", 10},
		{"ยี่สิบ", 20},
		{"ยี่สิบเอ็ด", 21},
	}
	for _, c := range cases {
		got, err := ParseInt(c.in, Strict())
		if err != nil {
			t.Errorf("ParseInt(%q, Strict) error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("ParseInt(%q, Strict) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestDefault_AcceptsNonStandardTensForms(t *testing.T) {
	// Without Strict, สองสิบ = 20 and หนึ่งสิบ = 10 are tolerated (lenient
	// default behaviour in the oracle).
	cases := []struct {
		in   string
		want int64
	}{
		{"สองสิบ", 20},
		{"หนึ่งสิบ", 10},
	}
	for _, c := range cases {
		got, err := ParseInt(c.in)
		if err != nil {
			t.Errorf("ParseInt(%q) error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("ParseInt(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}
