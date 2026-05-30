package thainum

import (
	"errors"
	"testing"
)

// Oracle: all expected values come from thainum.dart/lib/src/idioms.dart and
// thainum.dart/test/idioms_test.dart, not from running this Go code.

func TestQuantityWord(t *testing.T) {
	cases := []struct {
		unit QuantityUnit
		want string
	}{
		{UnitHalf, "ครึ่ง"},
		{UnitPair, "คู่"},
		{UnitDozen, "โหล"},
		{UnitGross, "กุรุส"},
	}
	for _, c := range cases {
		if got := QuantityWord(c.unit); got != c.want {
			t.Errorf("QuantityWord(%d) = %q, want %q", c.unit, got, c.want)
		}
	}
	if got := QuantityWord(QuantityUnit(99)); got != "" {
		t.Errorf("QuantityWord(unknown) = %q, want empty", got)
	}
}

func TestQuantityValue(t *testing.T) {
	cases := []struct {
		unit      QuantityUnit
		wantWhole int64
		wantHalf  bool
	}{
		{UnitHalf, 0, true}, // 0.5
		{UnitPair, 2, false},
		{UnitDozen, 12, false},
		{UnitGross, 144, false},
	}
	for _, c := range cases {
		whole, half := QuantityValue(c.unit)
		if whole != c.wantWhole || half != c.wantHalf {
			t.Errorf("QuantityValue(%d) = (%d, %v), want (%d, %v)",
				c.unit, whole, half, c.wantWhole, c.wantHalf)
		}
	}
	if whole, half := QuantityValue(QuantityUnit(99)); whole != 0 || half {
		t.Errorf("QuantityValue(unknown) = (%d, %v), want (0, false)", whole, half)
	}
}

func TestParseQuantity(t *testing.T) {
	cases := []struct {
		in        string
		wantWhole int64
		wantHalf  bool
	}{
		// standalone unit words
		{"ครึ่ง", 0, true},    // 0.5
		{"คู่", 2, false},     // 2
		{"โหล", 12, false},    // 12
		{"กุรุส", 144, false}, // 144
		// integer + trailing ครึ่ง
		{"สองครึ่ง", 2, true},         // 2.5
		{"สิบครึ่ง", 10, true},        // 10.5
		{"หนึ่งร้อยครึ่ง", 100, true}, // 100.5
		// plain integer reading
		{"สอง", 2, false},         // 2
		{"ยี่สิบเอ็ด", 21, false}, // 21
		{"๑๒", 12, false},         // 12 (Thai digits)
	}
	for _, c := range cases {
		whole, half, err := ParseQuantity(c.in)
		if err != nil {
			t.Errorf("ParseQuantity(%q) unexpected error: %v", c.in, err)
			continue
		}
		if whole != c.wantWhole || half != c.wantHalf {
			t.Errorf("ParseQuantity(%q) = (%d, %v), want (%d, %v)",
				c.in, whole, half, c.wantWhole, c.wantHalf)
		}
	}
}

func TestParseQuantityErrors(t *testing.T) {
	bad := []string{
		"",           // empty input
		"ไม่ใช่เลข",  // not a number
		"ครึ่งครึ่ง", // double ครึ่ง rejected
	}
	for _, in := range bad {
		_, _, err := ParseQuantity(in)
		if err == nil {
			t.Errorf("ParseQuantity(%q) expected error, got nil", in)
			continue
		}
		if !errors.Is(err, ErrParse) {
			t.Errorf("ParseQuantity(%q) error = %v, want wrap ErrParse", in, err)
		}
	}
}

func TestParseHalfBaht(t *testing.T) {
	good := []string{"ครึ่งบาท", "  ครึ่งบาท  "}
	for _, in := range good {
		got, err := ParseHalfBaht(in)
		if err != nil {
			t.Errorf("ParseHalfBaht(%q) unexpected error: %v", in, err)
			continue
		}
		if got != 50 {
			t.Errorf("ParseHalfBaht(%q) = %d, want 50", in, got)
		}
	}

	bad := []string{"หนึ่งบาท", "ครึ่ง", ""}
	for _, in := range bad {
		_, err := ParseHalfBaht(in)
		if err == nil {
			t.Errorf("ParseHalfBaht(%q) expected error, got nil", in)
			continue
		}
		if !errors.Is(err, ErrParse) {
			t.Errorf("ParseHalfBaht(%q) error = %v, want wrap ErrParse", in, err)
		}
	}
}
