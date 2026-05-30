package thainum

// Parity tests for selectable satang rounding on BahtFromString.
// Oracle: thainum.dart/lib/src/baht.dart (SatangRounding) + baht_rounding_test.dart.
import "testing"

func expectedText(satang int64) string { return BahtSatang(satang) }
func negText(mag int64) string         { return "ลบ" + BahtSatang(mag) }

func TestBahtFromStringDefaultEqualsExplicitHalfAway(t *testing.T) {
	inputs := []string{
		"0", "0.00", "21.21", "100", "100.00", "1000.50",
		"0.5", "0.05", "0.005", "0.004", "0.006", "0.015", "0.025",
		"12.345", "12.344", "-21.21", "-0.005", "-0.004",
		"1234567.89", "0.999", "99.995",
	}
	for _, in := range inputs {
		noArg, err := BahtFromString(in)
		if err != nil {
			t.Fatalf("BahtFromString(%q) no-arg error: %v", in, err)
		}
		explicit, err := BahtFromString(in, RoundHalfAwayFromZero)
		if err != nil {
			t.Fatalf("BahtFromString(%q, halfAway) error: %v", in, err)
		}
		if noArg != explicit {
			t.Errorf("BahtFromString(%q): no-arg=%q explicit=%q (must match)", in, noArg, explicit)
		}
	}
}

func TestBahtFromStringHistoricalPinned(t *testing.T) {
	cases := []struct{ in, want string }{
		{"0.00", "ศูนย์บาทถ้วน"},
		{"1.00", "หนึ่งบาทถ้วน"},
		{"21.21", "ยี่สิบเอ็ดบาทยี่สิบเอ็ดสตางค์"},
		{"100.50", "หนึ่งร้อยบาทห้าสิบสตางค์"},
		{"1.005", "หนึ่งบาทหนึ่งสตางค์"},
		{"-5", "ลบห้าบาทถ้วน"},
	}
	for _, c := range cases {
		got, err := BahtFromString(c.in)
		if err != nil {
			t.Errorf("BahtFromString(%q) error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("BahtFromString(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

type roundCase struct {
	in   string
	mode SatangRounding
	want int64
	neg  bool
	cmt  string
}

func runRoundCases(t *testing.T, cases []roundCase) {
	t.Helper()
	for _, c := range cases {
		got, err := BahtFromString(c.in, c.mode)
		if err != nil {
			t.Errorf("BahtFromString(%q, %d) error: %v", c.in, c.mode, err)
			continue
		}
		want := expectedText(c.want)
		if c.neg && c.want != 0 {
			want = negText(c.want)
		}
		if got != want {
			t.Errorf("BahtFromString(%q, mode=%d) = %q, want %q (%s)", c.in, c.mode, got, want, c.cmt)
		}
	}
}

func TestRoundHalfAwayFromZero(t *testing.T) {
	runRoundCases(t, []roundCase{
		{"0.005", RoundHalfAwayFromZero, 1, false, "half -> up"},
		{"0.004", RoundHalfAwayFromZero, 0, false, "below half -> down"},
		{"0.015", RoundHalfAwayFromZero, 2, false, "half 2nd satang -> up"},
		{"0.0051", RoundHalfAwayFromZero, 1, false, "above half -> up"},
		{"-0.005", RoundHalfAwayFromZero, 1, true, "neg half magnitude up"},
	})
}

func TestRoundHalfEven(t *testing.T) {
	runRoundCases(t, []roundCase{
		{"0.005", RoundHalfEven, 0, false, "tie -> even 0"},
		{"0.015", RoundHalfEven, 2, false, "tie 1->2 even"},
		{"0.025", RoundHalfEven, 2, false, "tie 2->2 even"},
		{"0.0151", RoundHalfEven, 2, false, "above half -> up"},
		{"0.0049", RoundHalfEven, 0, false, "below half -> down"},
	})
}

func TestRoundTruncate(t *testing.T) {
	runRoundCases(t, []roundCase{
		{"0.009", RoundTruncate, 0, false, "drop tail"},
		{"0.019", RoundTruncate, 1, false, "drop tail keep 1"},
		{"-0.009", RoundTruncate, 0, false, "toward zero -> 0"},
		{"-0.019", RoundTruncate, 1, true, "magnitude 1"},
	})
}

func TestRoundCeil(t *testing.T) {
	runRoundCases(t, []roundCase{
		{"0.001", RoundCeil, 1, false, "+ tiny tail -> up"},
		{"0.010", RoundCeil, 1, false, "no dropped tail"},
		{"-0.009", RoundCeil, 0, false, "- magnitude truncates -> 0"},
		{"-0.019", RoundCeil, 1, true, "- magnitude truncates -> 1"},
	})
}

func TestRoundFloor(t *testing.T) {
	runRoundCases(t, []roundCase{
		{"0.009", RoundFloor, 0, false, "+ truncates -> 0"},
		{"0.019", RoundFloor, 1, false, "+ truncates -> 1"},
		{"-0.001", RoundFloor, 1, true, "- tiny tail magnitude up"},
		{"-0.010", RoundFloor, 1, true, "- no tail magnitude 1"},
	})
}

func TestRoundExactInputsAreModeInvariant(t *testing.T) {
	inputs := []string{"12.34", "0.50", "100.00", "-7.07", "0.5", "0.05"}
	modes := []SatangRounding{RoundHalfAwayFromZero, RoundHalfEven, RoundTruncate, RoundCeil, RoundFloor}
	for _, in := range inputs {
		base, err := BahtFromString(in, RoundTruncate)
		if err != nil {
			t.Fatalf("BahtFromString(%q, trunc) error: %v", in, err)
		}
		for _, m := range modes {
			got, err := BahtFromString(in, m)
			if err != nil {
				t.Errorf("BahtFromString(%q, %d) error: %v", in, m, err)
				continue
			}
			if got != base {
				t.Errorf("mode=%d changed exact %q: got %q want %q", m, in, got, base)
			}
		}
	}
}

func TestRoundNoFloatPathLongFraction(t *testing.T) {
	above := "0.00500000000000000000000000001"
	if got, err := BahtFromString(above, RoundHalfEven); err != nil || got != expectedText(1) {
		t.Errorf("BahtFromString(%q, halfEven) = (%q,%v) want (%q,nil)", above, got, err, expectedText(1))
	}
	exact := "0.005000000000000000000000000000"
	if got, err := BahtFromString(exact, RoundHalfEven); err != nil || got != expectedText(0) {
		t.Errorf("BahtFromString(%q, halfEven) = (%q,%v) want (%q,nil)", exact, got, err, expectedText(0))
	}
}

func TestSatangRoundingModesDistinct(t *testing.T) {
	modes := []SatangRounding{RoundHalfAwayFromZero, RoundHalfEven, RoundTruncate, RoundCeil, RoundFloor}
	seen := map[SatangRounding]bool{}
	for _, m := range modes {
		if seen[m] {
			t.Errorf("duplicate mode value %d", m)
		}
		seen[m] = true
	}
	if len(seen) != 5 {
		t.Errorf("want 5 distinct modes, got %d", len(seen))
	}
	if RoundHalfAwayFromZero != 0 {
		t.Errorf("RoundHalfAwayFromZero = %d, want 0 (default)", RoundHalfAwayFromZero)
	}
}
