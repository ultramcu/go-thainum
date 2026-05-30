package thainum

import (
	"errors"
	"testing"
)

// validID is a SYNTHETIC Thai ID: its 13th digit was computed by the MOD-11
// algorithm from the first 12 digits, so it is structurally valid but NOT a
// real person's number. Taken verbatim from the Dart oracle (thai_id_test.dart).
const validID = "1101700230708"

func TestIsValidThaiID_AcceptsSyntheticValid(t *testing.T) {
	if !IsValidThaiID(validID) {
		t.Errorf("IsValidThaiID(%q) = false, want true", validID)
	}
}

func TestIsValidThaiID_MutationInvalidates(t *testing.T) {
	// Position 2 (the third digit) carries MOD-11 weight 11, so changing it
	// alone never alters the checksum (d*11 mod 11 == 0). Skip it here; it is
	// asserted separately as a documented blind spot.
	const blindSpot = 2
	for pos := 0; pos < 13; pos++ {
		if pos == blindSpot {
			continue
		}
		for alt := 0; alt < 10; alt++ {
			if byte(alt)+'0' == validID[pos] {
				continue
			}
			chars := []byte(validID)
			chars[pos] = byte(alt) + '0'
			mutated := string(chars)
			if IsValidThaiID(mutated) {
				t.Errorf("mutating pos %d to %d should be invalid (%s)", pos, alt, mutated)
			}
		}
	}
}

func TestIsValidThaiID_Weight11BlindSpot(t *testing.T) {
	// Documents the inherent MOD-11 property: changing the 3rd digit (index 2)
	// never changes validity.
	for alt := 0; alt < 10; alt++ {
		chars := []byte(validID)
		chars[2] = byte(alt) + '0'
		if !IsValidThaiID(string(chars)) {
			t.Errorf("changing index-2 digit to %d should stay valid", alt)
		}
	}
}

func TestIsValidThaiID_LengthNegatives(t *testing.T) {
	cases := []string{
		"110170023070",   // 12 digits
		"11017002307088", // 14 digits
		"",
	}
	for _, c := range cases {
		if IsValidThaiID(c) {
			t.Errorf("IsValidThaiID(%q) = true, want false", c)
		}
	}
}

func TestIsValidThaiID_NonDigit(t *testing.T) {
	if IsValidThaiID("11017002307AB") {
		t.Errorf("IsValidThaiID with non-digit chars = true, want false")
	}
}

func TestIsValidThaiID_GroupedSpacedThaiForms(t *testing.T) {
	cases := []string{
		"1-1017-00230-70-8",
		"1 1017 00230 70 8",
		ToThaiDigits(validID),
	}
	for _, c := range cases {
		if !IsValidThaiID(c) {
			t.Errorf("IsValidThaiID(%q) = false, want true", c)
		}
	}
}

func TestIsValidThaiTaxID_Delegates(t *testing.T) {
	if IsValidThaiTaxID(validID) != IsValidThaiID(validID) {
		t.Errorf("IsValidThaiTaxID(validID) != IsValidThaiID(validID)")
	}
	if IsValidThaiTaxID("110170023070") {
		t.Errorf("IsValidThaiTaxID(12-digit) = true, want false")
	}
}

func TestParseThaiID_StripsSeparatorsAndThaiNumerals(t *testing.T) {
	cases := []string{
		"1-1017-00230-70-8",
		"1 1017 00230 70 8",
		ToThaiDigits(validID),
	}
	for _, c := range cases {
		got, err := ParseThaiID(c)
		if err != nil {
			t.Errorf("ParseThaiID(%q) unexpected error: %v", c, err)
			continue
		}
		if got != validID {
			t.Errorf("ParseThaiID(%q) = %q, want %q", c, got, validID)
		}
	}
}

func TestParseThaiID_Errors(t *testing.T) {
	cases := []string{
		"110170023070",  // wrong length
		"11017002307AB", // bad characters
	}
	for _, c := range cases {
		_, err := ParseThaiID(c)
		if err == nil {
			t.Errorf("ParseThaiID(%q) = nil error, want error", c)
			continue
		}
		if !errors.Is(err, ErrParse) {
			t.Errorf("ParseThaiID(%q) error %v does not wrap ErrParse", c, err)
		}
	}
}

func TestFormatThaiID(t *testing.T) {
	got, err := FormatThaiID(validID)
	if err != nil {
		t.Fatalf("FormatThaiID(%q) unexpected error: %v", validID, err)
	}
	const want = "1-1017-00230-70-8"
	if got != want {
		t.Errorf("FormatThaiID(%q) = %q, want %q", validID, got, want)
	}
}

func TestFormatThaiID_RoundTrip(t *testing.T) {
	formatted, err := FormatThaiID(validID)
	if err != nil {
		t.Fatalf("FormatThaiID error: %v", err)
	}
	got, err := ParseThaiID(formatted)
	if err != nil {
		t.Fatalf("ParseThaiID error: %v", err)
	}
	if got != validID {
		t.Errorf("round-trip = %q, want %q", got, validID)
	}
}

func TestFormatThaiID_Error(t *testing.T) {
	_, err := FormatThaiID("123")
	if err == nil || !errors.Is(err, ErrParse) {
		t.Errorf("FormatThaiID(%q) = %v, want error wrapping ErrParse", "123", err)
	}
}

func TestClassifyThaiID_LeadingDigits(t *testing.T) {
	cases := map[string]ThaiIDKind{
		"1": ThaiBornRegisteredOnTime,
		"2": ThaiBornRegisteredLate,
		"3": ThaiInRegistryBefore1984,
		"4": ThaiBornBefore1984NotRegistered,
		"5": ThaiAddedLater,
		"6": ForeignerTemporary,
		"7": ChildOfForeignerTemporary,
		"8": NaturalisedOrPermanentResident,
		"0": ThaiIDUnknown,
		"9": ThaiIDUnknown,
	}
	for lead, want := range cases {
		// 13 chars; classification only looks at the leading digit.
		id := lead + "000000000000"
		if got := ClassifyThaiID(id); got != want {
			t.Errorf("ClassifyThaiID(leading %s) = %v, want %v", lead, got, want)
		}
	}
}

func TestClassifyThaiID_NonThirteenIsUnknown(t *testing.T) {
	if got := ClassifyThaiID("123"); got != ThaiIDUnknown {
		t.Errorf("ClassifyThaiID(%q) = %v, want ThaiIDUnknown", "123", got)
	}
}

func TestSpeakThaiID_ReadsDigitByDigit(t *testing.T) {
	const want = "หนึ่ง หนึ่ง ศูนย์ หนึ่ง เจ็ด ศูนย์ ศูนย์ สอง สาม ศูนย์ เจ็ด ศูนย์ แปด"
	got, err := SpeakThaiID(validID)
	if err != nil {
		t.Fatalf("SpeakThaiID(%q) unexpected error: %v", validID, err)
	}
	if got != want {
		t.Errorf("SpeakThaiID(%q) = %q, want %q", validID, got, want)
	}
}

func TestSpeakThaiID_ErrorWhenNotThirteen(t *testing.T) {
	_, err := SpeakThaiID("123")
	if err == nil || !errors.Is(err, ErrParse) {
		t.Errorf("SpeakThaiID(%q) = %v, want error wrapping ErrParse", "123", err)
	}
}
