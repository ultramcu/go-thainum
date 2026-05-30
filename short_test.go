package thainum

import (
	"math/big"
	"testing"
)

// All expected values are taken from the Dart source of truth
// (thainum.dart/lib/src/short.dart and test/short_test.dart) as an independent
// oracle — not from running this Go implementation.

func TestSpellShortWorkedExamples(t *testing.T) {
	cases := []struct {
		n    int64
		want string
	}{
		{1500000, "หนึ่งจุดห้าล้าน"},
		{15000000, "สิบห้าล้าน"},
		{2300000000, "สองจุดสามพันล้าน"},
		{50000000000, "ห้าสิบพันล้าน"},
		{1200000000000, "หนึ่งจุดสองล้านล้าน"},
		{1234567, "หนึ่งจุดสองสามล้าน"},
	}
	for _, c := range cases {
		if got := SpellShort(c.n, 2); got != c.want {
			t.Errorf("SpellShort(%d, 2) = %q, want %q", c.n, got, c.want)
		}
	}
}

func TestFormatShortWorkedExamples(t *testing.T) {
	cases := []struct {
		n    int64
		want string
	}{
		{1500000, "1.5 ล้าน"},
		{15000000, "15 ล้าน"},
		{2300000000, "2.3 พันล้าน"},
		{50000000000, "50 พันล้าน"},
		{1200000000000, "1.2 ล้านล้าน"},
		{1234567, "1.23 ล้าน"},
	}
	for _, c := range cases {
		if got := FormatShort(c.n, 2, false); got != c.want {
			t.Errorf("FormatShort(%d, 2, false) = %q, want %q", c.n, got, c.want)
		}
	}
}

func TestFormatShortUnitNames(t *testing.T) {
	cases := []struct {
		n    int64
		want string
	}{
		{1000000, "1 ล้าน"},
		{1000000000, "1 พันล้าน"},
		{1000000000000, "1 ล้านล้าน"},
		{1000000000000000, "1 พันล้านล้าน"},
		{1000000000000000000, "1 ล้านล้านล้าน"},
	}
	for _, c := range cases {
		if got := FormatShort(c.n, 2, false); got != c.want {
			t.Errorf("FormatShort(%d, 2, false) = %q, want %q", c.n, got, c.want)
		}
	}
}

func TestSpellShortUnitNames(t *testing.T) {
	cases := []struct {
		n    int64
		want string
	}{
		{1000000, "หนึ่งล้าน"},
		{1000000000, "หนึ่งพันล้าน"},
		{1000000000000, "หนึ่งล้านล้าน"},
		{1000000000000000, "หนึ่งพันล้านล้าน"},
		{1000000000000000000, "หนึ่งล้านล้านล้าน"},
	}
	for _, c := range cases {
		if got := SpellShort(c.n, 2); got != c.want {
			t.Errorf("SpellShort(%d, 2) = %q, want %q", c.n, got, c.want)
		}
	}
}

func TestShortBigUnitNames(t *testing.T) {
	cases := []struct {
		s         string
		wantFmt   string
		wantSpell string
	}{
		{"1000000000000000000000", "1 พันล้านล้านล้าน", "หนึ่งพันล้านล้านล้าน"},
		{"1000000000000000000000000", "1 ล้านล้านล้านล้าน", "หนึ่งล้านล้านล้านล้าน"},
	}
	for _, c := range cases {
		v, _ := new(big.Int).SetString(c.s, 10)
		if got := FormatShortBig(v, 2, false); got != c.wantFmt {
			t.Errorf("FormatShortBig(%s, 2, false) = %q, want %q", c.s, got, c.wantFmt)
		}
		if got := SpellShortBig(v, 2); got != c.wantSpell {
			t.Errorf("SpellShortBig(%s, 2) = %q, want %q", c.s, got, c.wantSpell)
		}
	}
}

func TestShortCoefficientBoundaries(t *testing.T) {
	// 999,000,000 < 10^9 -> ล้าน unit, coefficient 999.
	if got := FormatShort(999000000, 2, false); got != "999 ล้าน" {
		t.Errorf("FormatShort(999000000) = %q, want %q", got, "999 ล้าน")
	}
	if got := SpellShort(999000000, 2); got != "เก้าร้อยเก้าสิบเก้าล้าน" {
		t.Errorf("SpellShort(999000000) = %q, want %q", got, "เก้าร้อยเก้าสิบเก้าล้าน")
	}

	// Exact unit -> whole coefficient, no จุด.
	if got := SpellShort(2000000, 2); got != "สองล้าน" {
		t.Errorf("SpellShort(2000000) = %q, want %q", got, "สองล้าน")
	}
	if got := FormatShort(2000000, 2, false); got != "2 ล้าน" {
		t.Errorf("FormatShort(2000000) = %q, want %q", got, "2 ล้าน")
	}

	// Rounding carry: 999,999,999 ~ 1000.00 ล้าน rounds up and promotes.
	if got := FormatShort(999999999, 2, false); got != "1 พันล้าน" {
		t.Errorf("FormatShort(999999999) = %q, want %q", got, "1 พันล้าน")
	}
	if got := SpellShort(999999999, 2); got != "หนึ่งพันล้าน" {
		t.Errorf("SpellShort(999999999) = %q, want %q", got, "หนึ่งพันล้าน")
	}
}

func TestShortDecimalsParameter(t *testing.T) {
	// decimals 0 rounds to a whole coefficient.
	if got := FormatShort(1500000, 0, false); got != "2 ล้าน" {
		t.Errorf("FormatShort(1500000, 0) = %q, want %q", got, "2 ล้าน")
	}
	if got := FormatShort(1234567, 0, false); got != "1 ล้าน" {
		t.Errorf("FormatShort(1234567, 0) = %q, want %q", got, "1 ล้าน")
	}
	if got := SpellShort(1500000, 0); got != "สองล้าน" {
		t.Errorf("SpellShort(1500000, 0) = %q, want %q", got, "สองล้าน")
	}

	// decimals 3 keeps more precision.
	if got := FormatShort(1234567, 3, false); got != "1.235 ล้าน" {
		t.Errorf("FormatShort(1234567, 3) = %q, want %q", got, "1.235 ล้าน")
	}
	if got := SpellShort(1234567, 3); got != "หนึ่งจุดสองสามห้าล้าน" {
		t.Errorf("SpellShort(1234567, 3) = %q, want %q", got, "หนึ่งจุดสองสามห้าล้าน")
	}
}

func TestShortNegativeDecimals(t *testing.T) {
	if _, err := SpellShortErr(1500000, -1); err == nil {
		t.Errorf("SpellShortErr(1500000, -1) expected error, got nil")
	}
	if _, err := FormatShortErr(1500000, -1, false); err == nil {
		t.Errorf("FormatShortErr(1500000, -1, false) expected error, got nil")
	}
	// Panic variants must panic on negative decimals.
	assertPanics(t, "SpellShort", func() { SpellShort(1500000, -1) })
	assertPanics(t, "FormatShort", func() { FormatShort(1500000, -1, false) })
}

func assertPanics(t *testing.T, name string, f func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Errorf("%s with negative decimals did not panic", name)
		}
	}()
	f()
}

func TestShortTrailingZeroTrimming(t *testing.T) {
	if got := SpellShort(2000000, 2); got != "สองล้าน" {
		t.Errorf("SpellShort(2000000) = %q, want %q", got, "สองล้าน")
	}
	if got := FormatShort(2000000, 2, false); got != "2 ล้าน" {
		t.Errorf("FormatShort(2000000) = %q, want %q", got, "2 ล้าน")
	}
	// One significant fractional digit.
	if got := FormatShort(2500000, 2, false); got != "2.5 ล้าน" {
		t.Errorf("FormatShort(2500000) = %q, want %q", got, "2.5 ล้าน")
	}
	if got := SpellShort(2500000, 2); got != "สองจุดห้าล้าน" {
		t.Errorf("SpellShort(2500000) = %q, want %q", got, "สองจุดห้าล้าน")
	}
}

func TestShortNegative(t *testing.T) {
	if got := SpellShort(-1500000, 2); got != "ลบหนึ่งจุดห้าล้าน" {
		t.Errorf("SpellShort(-1500000) = %q, want %q", got, "ลบหนึ่งจุดห้าล้าน")
	}
	if got := SpellShort(-15000000, 2); got != "ลบสิบห้าล้าน" {
		t.Errorf("SpellShort(-15000000) = %q, want %q", got, "ลบสิบห้าล้าน")
	}
	if got := FormatShort(-1500000, 2, false); got != "-1.5 ล้าน" {
		t.Errorf("FormatShort(-1500000) = %q, want %q", got, "-1.5 ล้าน")
	}
	if got := FormatShort(-50000000000, 2, false); got != "-50 พันล้าน" {
		t.Errorf("FormatShort(-50000000000) = %q, want %q", got, "-50 พันล้าน")
	}
}

// The < 10^6 fallback must equal the full reading (Spell / FormatInt), exactly
// as the Dart test asserts spellShort(n) == spell(n) and
// formatShort(n) == formatInt(n).
func TestShortFallbackBelowMillion(t *testing.T) {
	samples := []int64{0, 5, 999, 1000, 12345, 999999, -999999, -1}
	for _, n := range samples {
		if got, want := SpellShort(n, 2), Spell(n); got != want {
			t.Errorf("SpellShort(%d) = %q, want Spell = %q", n, got, want)
		}
		if got, want := FormatShort(n, 2, false), FormatInt(n); got != want {
			t.Errorf("FormatShort(%d) = %q, want FormatInt = %q", n, got, want)
		}
	}
	// 999999 fallback golden (Dart literal).
	if got := FormatShort(999999, 2, false); got != "999,999" {
		t.Errorf("FormatShort(999999) = %q, want %q", got, "999,999")
	}
}

func TestShortThaiDigitsFlag(t *testing.T) {
	cases := []struct {
		n    int64
		want string
	}{
		{1500000, "๑.๕ ล้าน"},
		{15000000, "๑๕ ล้าน"},
		{2300000000, "๒.๓ พันล้าน"},
	}
	for _, c := range cases {
		if got := FormatShort(c.n, 2, true); got != c.want {
			t.Errorf("FormatShort(%d, 2, true) = %q, want %q", c.n, got, c.want)
		}
	}
	// Negative with thaiDigits keeps ASCII sign and point.
	if got := FormatShort(-1500000, 2, true); got != "-๑.๕ ล้าน" {
		t.Errorf("FormatShort(-1500000, 2, true) = %q, want %q", got, "-๑.๕ ล้าน")
	}
}

func TestShortBigScaleGoldens(t *testing.T) {
	// 1.234e18 -> ล้านล้านล้าน unit, coefficient 1.23 (decimals=2).
	v1, _ := new(big.Int).SetString("1234000000000000000", 10)
	if got := SpellShortBig(v1, 2); got != "หนึ่งจุดสองสามล้านล้านล้าน" {
		t.Errorf("SpellShortBig(1234e15) = %q, want %q", got, "หนึ่งจุดสองสามล้านล้านล้าน")
	}
	if got := FormatShortBig(v1, 2, false); got != "1.23 ล้านล้านล้าน" {
		t.Errorf("FormatShortBig(1234e15) = %q, want %q", got, "1.23 ล้านล้านล้าน")
	}

	// 999,999,999,999,999,999: largest unit <= it is 10^15 (พันล้านล้าน),
	// coefficient ~999.9999 rounds up to 1000.00 -> promotes to 10^18.
	v2, _ := new(big.Int).SetString("999999999999999999", 10)
	if got := SpellShortBig(v2, 2); got != "หนึ่งล้านล้านล้าน" {
		t.Errorf("SpellShortBig(1e18-1) = %q, want %q", got, "หนึ่งล้านล้านล้าน")
	}
	if got := FormatShortBig(v2, 2, false); got != "1 ล้านล้านล้าน" {
		t.Errorf("FormatShortBig(1e18-1) = %q, want %q", got, "1 ล้านล้านล้าน")
	}
}

func TestShortIntBigAgree(t *testing.T) {
	samples := []int64{1500000, 15000000, 2300000000, 50000000000, 999999999, 1234567}
	for _, n := range samples {
		big := new(big.Int).SetInt64(n)
		if SpellShort(n, 2) != SpellShortBig(big, 2) {
			t.Errorf("SpellShort(%d) != SpellShortBig form", n)
		}
		if FormatShort(n, 2, false) != FormatShortBig(big, 2, false) {
			t.Errorf("FormatShort(%d) != FormatShortBig form", n)
		}
	}
}
