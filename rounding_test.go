package thainum

// Blind parity tests for BahtFromString's selectable satang rounding (issue #11
// Stream B). Every expected value here is derived from the Dart 0.5.3 oracle —
// thainum.dart/lib/src/baht.dart (SatangRounding) and
// thainum.dart/test/baht_rounding_test.dart — NOT from running the Go code.
//
// Expected text for a given satang magnitude is produced via BahtSatang, which
// is the same primitive the Dart oracle's `expectedText(int satang)` helper
// uses (bahtSatang). For negative magnitudes the oracle prefixes 'ลบ' only when
// the magnitude is non-zero.
//
// Function and helper names are kept distinct from parity_rounding_test.go so
// both files can live in package thainum without collision.

import "testing"

// roundText spells a non-negative satang magnitude (oracle: bahtSatang).
func roundText(satang int64) string { return BahtSatang(satang) }

// roundNegText spells a negative satang magnitude with the leading 'ลบ' the
// oracle applies when the magnitude is non-zero.
func roundNegText(mag int64) string { return "ลบ" + BahtSatang(mag) }

// historicalBahtFromString is the pre-#11 algorithm copied verbatim from the
// Dart oracle (baht_rounding_test.dart) as an independent reference, so the
// default-mode regression guard does not compare the implementation to itself.
//
// It mirrors: int part * 100 + two truncated fractional digits, then +1 satang
// iff the third fractional digit >= '5'. 'ลบ' prefix only when the result is
// non-zero.
func historicalBahtFromString(amount string) string {
	s := amount
	// trim leading/trailing ASCII spaces (Dart's String.trim is wider, but the
	// regression inputs contain none).
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t' || s[0] == '\n' || s[0] == '\r') {
		s = s[1:]
	}
	for len(s) > 0 {
		c := s[len(s)-1]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			s = s[:len(s)-1]
			continue
		}
		break
	}
	neg := false
	if len(s) > 0 && s[0] == '-' {
		neg = true
		s = s[1:]
	} else if len(s) > 0 && s[0] == '+' {
		s = s[1:]
	}
	intPart := s
	frac := ""
	if dot := indexByte(s, '.'); dot >= 0 {
		intPart = s[:dot]
		frac = s[dot+1:]
	}
	if intPart == "" {
		intPart = "0"
	}
	// satang = int(intPart) * 100, computed digit-by-digit to avoid any float.
	var sat int64
	for i := 0; i < len(intPart); i++ {
		sat = sat*10 + int64(intPart[i]-'0')
	}
	sat *= 100
	cents := int64(0)
	if len(frac) >= 1 {
		cents += int64(frac[0]-'0') * 10
	}
	if len(frac) >= 2 {
		cents += int64(frac[1] - '0')
	}
	if len(frac) >= 3 && frac[2] >= '5' {
		cents++
	}
	sat += cents
	if neg && sat != 0 {
		return "ลบ" + BahtSatang(sat)
	}
	return BahtSatang(sat)
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

// Port of: "default rounding is byte-identical to historical behaviour".
func TestBlindDefaultRoundingMatchesHistorical(t *testing.T) {
	regressionInputs := []string{
		"0", "0.00", "21.21", "100", "100.00", "1000.50", "0.5", "0.05",
		"0.005", "0.004", "0.006", "0.015", "0.025", "12.345", "12.344",
		"-21.21", "-0.005", "-0.004", "1234567.89", "0.999", "99.995",
	}
	for _, in := range regressionInputs {
		got, err := BahtFromString(in)
		if err != nil {
			t.Fatalf("BahtFromString(%q) error: %v", in, err)
		}
		want := historicalBahtFromString(in)
		if got != want {
			t.Errorf("BahtFromString(%q) = %q, want %q (historical oracle)", in, got, want)
		}
		// explicit-default == implicit-default
		explicit, err := BahtFromString(in, RoundHalfAwayFromZero)
		if err != nil {
			t.Fatalf("BahtFromString(%q, RoundHalfAwayFromZero) error: %v", in, err)
		}
		if explicit != got {
			t.Errorf("BahtFromString(%q): implicit=%q explicit-half-away=%q (must match)", in, got, explicit)
		}
	}
}

// blindRoundCase is one (input, mode, expected-satang, negative?) tuple from
// the Dart oracle.
type blindRoundCase struct {
	in   string
	mode SatangRounding
	sat  int64
	neg  bool
	note string
}

func runBlindRoundCases(t *testing.T, cases []blindRoundCase) {
	t.Helper()
	for _, c := range cases {
		got, err := BahtFromString(c.in, c.mode)
		if err != nil {
			t.Errorf("BahtFromString(%q, mode=%d) error: %v", c.in, c.mode, err)
			continue
		}
		want := roundText(c.sat)
		if c.neg {
			want = roundNegText(c.sat)
		}
		if got != want {
			t.Errorf("BahtFromString(%q, mode=%d) = %q, want %q (%s)", c.in, c.mode, got, want, c.note)
		}
	}
}

// Port of group 'halfAwayFromZero'.
func TestBlindRoundHalfAwayFromZero(t *testing.T) {
	runBlindRoundCases(t, []blindRoundCase{
		{"0.005", RoundHalfAwayFromZero, 1, false, "0.005 -> 1 satang"},
		{"0.004", RoundHalfAwayFromZero, 0, false, "0.004 -> 0 satang"},
		{"0.015", RoundHalfAwayFromZero, 2, false, "0.015 -> 2 satang"},
		{"0.0051", RoundHalfAwayFromZero, 1, false, "0.0051 above half -> 1"},
		{"-0.005", RoundHalfAwayFromZero, 1, true, "-0.005 -> -1 satang magnitude"},
	})
}

// Port of group "halfEven (banker's rounding)".
func TestBlindRoundHalfEven(t *testing.T) {
	runBlindRoundCases(t, []blindRoundCase{
		{"0.005", RoundHalfEven, 0, false, "0.005 tie -> even 0"},
		{"0.015", RoundHalfEven, 2, false, "0.015 tie -> even 2"},
		{"0.025", RoundHalfEven, 2, false, "0.025 tie -> even 2"},
		{"0.0151", RoundHalfEven, 2, false, "0.0151 above half -> 2"},
		{"0.0049", RoundHalfEven, 0, false, "0.0049 below half -> 0"},
		{"-0.025", RoundHalfEven, 2, true, "-0.025 -> -2 satang magnitude (even)"},
	})
}

// Port of group 'truncate (toward zero)'.
func TestBlindRoundTruncate(t *testing.T) {
	runBlindRoundCases(t, []blindRoundCase{
		{"0.009", RoundTruncate, 0, false, "0.009 -> 0 satang"},
		{"0.019", RoundTruncate, 1, false, "0.019 -> 1 satang"},
		{"-0.009", RoundTruncate, 0, false, "-0.009 -> 0 satang"},
		{"-0.019", RoundTruncate, 1, true, "-0.019 -> -1 satang magnitude"},
	})
}

// Port of group 'ceil (toward +inf)'.
func TestBlindRoundCeil(t *testing.T) {
	runBlindRoundCases(t, []blindRoundCase{
		{"0.001", RoundCeil, 1, false, "0.001 -> 1 satang"},
		{"0.010", RoundCeil, 1, false, "0.010 no tail, no bump -> 1"},
		{"-0.009", RoundCeil, 0, false, "-0.009 magnitude truncates -> 0"},
		{"-0.019", RoundCeil, 1, true, "-0.019 -> -1 satang magnitude"},
	})
}

// Port of group 'floor (toward -inf)'.
func TestBlindRoundFloor(t *testing.T) {
	runBlindRoundCases(t, []blindRoundCase{
		{"0.009", RoundFloor, 0, false, "0.009 -> 0 satang"},
		{"0.019", RoundFloor, 1, false, "0.019 -> 1 satang"},
		{"-0.001", RoundFloor, 1, true, "-0.001 -> -1 satang magnitude"},
		{"-0.010", RoundFloor, 1, true, "-0.010 no tail -> -1 satang magnitude"},
	})
}

// Port of 'exactly-on-satang and zero-tail inputs are mode-invariant'.
// An input with no discarded fractional tail must produce the same text under
// every mode (compared against RoundTruncate as the baseline).
func TestBlindRoundExactInputsAreModeInvariant(t *testing.T) {
	inputs := []string{"12.34", "0.50", "100.00", "-7.07", "0.5", "0.05"}
	modes := []SatangRounding{
		RoundHalfAwayFromZero, RoundHalfEven, RoundTruncate, RoundCeil, RoundFloor,
	}
	for _, in := range inputs {
		base, err := BahtFromString(in, RoundTruncate)
		if err != nil {
			t.Fatalf("BahtFromString(%q, RoundTruncate) error: %v", in, err)
		}
		for _, m := range modes {
			got, err := BahtFromString(in, m)
			if err != nil {
				t.Fatalf("BahtFromString(%q, mode=%d) error: %v", in, m, err)
			}
			if got != base {
				t.Errorf("BahtFromString(%q, mode=%d) = %q, want %q (exact input must be mode-invariant)", in, m, got, base)
			}
		}
	}
}

// Port of 'no float path: many-digit fraction does not lose precision'.
func TestBlindRoundNoFloatPathLongFraction(t *testing.T) {
	// Strictly above half -> round up despite being even.
	above := "0.00500000000000000000000000001"
	if got, err := BahtFromString(above, RoundHalfEven); err != nil || got != roundText(1) {
		t.Errorf("BahtFromString(%q, RoundHalfEven) = %q, err=%v, want %q", above, got, err, roundText(1))
	}
	// Exactly half -> round to even (0).
	exact := "0.005000000000000000000000000000"
	if got, err := BahtFromString(exact, RoundHalfEven); err != nil || got != roundText(0) {
		t.Errorf("BahtFromString(%q, RoundHalfEven) = %q, err=%v, want %q", exact, got, err, roundText(0))
	}
}
