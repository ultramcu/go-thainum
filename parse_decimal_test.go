package thainum

import (
	"errors"
	"testing"
)

// Tests for ParseDecimal, ported from the Dart 0.5.3 oracle
// (thainum.dart/test/parse_decimal_test.dart). The fractional part is read
// digit-by-digit; expected canonical decimal strings come from the oracle.

func TestParseDecimal_IntegerFraction(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"สิบสองจุดสามสี่", "12.34"},
		{"ศูนย์จุดห้า", "0.5"},
		{"หนึ่งร้อยจุดศูนย์ห้า", "100.05"},
	}
	for _, c := range cases {
		got, err := ParseDecimal(c.in)
		if err != nil {
			t.Errorf("ParseDecimal(%q) error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("ParseDecimal(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestParseDecimal_Negative(t *testing.T) {
	got, err := ParseDecimal("ลบสามจุดหนึ่งสี่")
	if err != nil {
		t.Fatalf("ParseDecimal error: %v", err)
	}
	if got != "-3.14" {
		t.Errorf("ParseDecimal = %q, want %q", got, "-3.14")
	}
}

func TestParseDecimal_NoDotBehavesLikeInteger(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"ยี่สิบเอ็ด", "21"},
		{"ลบห้า", "-5"},
	}
	for _, c := range cases {
		got, err := ParseDecimal(c.in)
		if err != nil {
			t.Errorf("ParseDecimal(%q) error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("ParseDecimal(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestParseDecimal_AcceptsThaiNumeralsInIntegerPart(t *testing.T) {
	got, err := ParseDecimal("๑๒จุดสามสี่")
	if err != nil {
		t.Fatalf("ParseDecimal error: %v", err)
	}
	if got != "12.34" {
		t.Errorf("ParseDecimal = %q, want %q", got, "12.34")
	}
}

func TestParseDecimal_RejectsMalformed(t *testing.T) {
	bad := []struct {
		name string
		in   string
	}{
		{"empty fractional run", "สิบสองจุด"},
		{"place word in fractional part", "สิบสองจุดสิบ"},
		{"missing integer part", "จุดห้า"},
		{"two จุด", "หนึ่งจุดสองจุดสาม"},
		{"empty", ""},
	}
	for _, c := range bad {
		t.Run(c.name, func(t *testing.T) {
			_, err := ParseDecimal(c.in)
			if err == nil {
				t.Fatalf("ParseDecimal(%q) = nil error, want error", c.in)
			}
			if !errors.Is(err, ErrParse) {
				t.Errorf("ParseDecimal(%q) error %v should wrap ErrParse", c.in, err)
			}
		})
	}
}
