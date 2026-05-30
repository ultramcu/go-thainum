package thainum

import "testing"

// Tests for IsDigits. The Dart 0.5.3 oracle (lib/src/spell.dart) defines
// isDigits as: true iff every code unit is ASCII '0'..'9' (0x30..0x39). Thai
// numeral characters (๐-๙) are NOT accepted by isDigits in the oracle — they
// are normalized to Arabic by toArabicDigits before any isDigits check. The
// empty string returns true in the oracle (the loop has nothing to reject).
//
// NOTE (ambiguity, see report): the oracle's isDigits is Arabic-only and
// empty->true. The Go API contract in issue #11 only gives the signature
// `func IsDigits(s string) bool`, so the Thai-numeral and empty-string cases
// are derived from the oracle and may warrant confirmation against the Go impl.

func TestIsDigits_ArabicTrue(t *testing.T) {
	cases := []string{"0", "7", "101", "0123456789", "00"}
	for _, s := range cases {
		if !IsDigits(s) {
			t.Errorf("IsDigits(%q) = false, want true", s)
		}
	}
}

func TestIsDigits_NonDigitFalse(t *testing.T) {
	cases := []string{
		"12a",   // trailing letter
		"a12",   // leading letter
		"1 2",   // embedded space
		"12.34", // decimal point
		"-5",    // sign
		"+5",    // sign
		"สิบ",   // Thai word
		"๑๒",    // Thai numerals (oracle: isDigits is Arabic-only)
		"1๒3",   // mixed Arabic + Thai numeral
	}
	for _, s := range cases {
		if IsDigits(s) {
			t.Errorf("IsDigits(%q) = true, want false", s)
		}
	}
}

func TestIsDigits_Empty(t *testing.T) {
	// Dart's internal isDigits is vacuously true on "", but Dart HIDES isDigits
	// from its public API, so there is no public oracle for the empty case. The
	// exported Go IsDigits deliberately requires a non-empty all-digit string —
	// an empty string is not "digits" — so the empty case is false by design.
	if IsDigits("") {
		t.Errorf("IsDigits(\"\") = true, want false (exported API requires non-empty)")
	}
}
