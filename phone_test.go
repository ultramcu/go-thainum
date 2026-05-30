package thainum

import (
	"errors"
	"testing"
)

// All expected values below are taken from the Dart oracle
// (thainum.dart/lib/src/phone.dart + test/phone_test.dart), not from running
// the Go implementation.

func TestFormatThaiPhone(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"mobile 3-3-4 grouping", "0812345678", "081-234-5678"},
		{"strips spaces", "081 234 5678", "081-234-5678"},
		{"strips dashes", "081-234-5678", "081-234-5678"},
		{"strips parens and spaces", "(081) 234-5678", "081-234-5678"},
		{"accepts Thai numerals", "๐๘๑๒๓๔๕๖๗๘", "081-234-5678"},
		{"landline 2-3-4 grouping", "021234567", "02-123-4567"},
		{"accepts +66 country code", "+66812345678", "081-234-5678"},
		{"unrecognised length is bare digits", "1234", "1234"},
		{"toll-free 1800 not mis-grouped", "1800123456", "1800123456"},
		{"short code 1669 bare", "1669", "1669"},
		{"short code 191 bare", "191", "191"},
		{"no-digit input unchanged (abc)", "abc", "abc"},
		{"no-digit input unchanged (empty)", "", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := FormatThaiPhone(c.in)
			if err != nil {
				t.Fatalf("FormatThaiPhone(%q) unexpected error: %v", c.in, err)
			}
			if got != c.want {
				t.Errorf("FormatThaiPhone(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestThaiPhoneKind(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want PhoneKind
	}{
		{"mobile 06", "0612345678", PhoneMobile},
		{"mobile 08", "0812345678", PhoneMobile},
		{"mobile 09", "0912345678", PhoneMobile},
		{"10-digit 0 + non-mobile second digit unknown", "0712345678", PhoneUnknown},
		{"Bangkok landline 02 + 9 digits", "021234567", PhoneLandline},
		{"provincial landline 0X + 9 digits", "053123456", PhoneLandline},
		{"toll-free 1800", "1800123456", PhoneTollFree},
		{"short code 191", "191", PhoneShortCode},
		{"short code 1669", "1669", PhoneShortCode},
		{"+66 mobile classifies as mobile", "+66812345678", PhoneMobile},
		{"unknown 12345", "12345", PhoneUnknown},
		{"unknown empty", "", PhoneUnknown},
		{"unknown abc", "abc", PhoneUnknown},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ThaiPhoneKind(c.in); got != c.want {
				t.Errorf("ThaiPhoneKind(%q) = %v, want %v", c.in, got, c.want)
			}
		})
	}
}

func TestPhoneKindString(t *testing.T) {
	// Names mirror the Dart ThaiPhoneKind enum value names exactly.
	cases := []struct {
		k    PhoneKind
		want string
	}{
		{PhoneMobile, "mobile"},
		{PhoneLandline, "landline"},
		{PhoneTollFree, "tollFree"},
		{PhoneShortCode, "shortCode"},
		{PhoneUnknown, "unknown"},
	}
	for _, c := range cases {
		if got := c.k.String(); got != c.want {
			t.Errorf("PhoneKind(%d).String() = %q, want %q", int(c.k), got, c.want)
		}
	}
}

func TestNormalizeThaiPhone(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"drops leading 0, prefixes +66", "0812345678", "+66812345678"},
		{"landline normalises too", "02-123-4567", "+6621234567"},
		{"idempotent on +66 input", "+66812345678", "+66812345678"},
		{"accepts Thai numerals and separators", "๐๘๑-๒๓๔-๕๖๗๘", "+66812345678"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := NormalizeThaiPhone(c.in)
			if err != nil {
				t.Fatalf("NormalizeThaiPhone(%q) unexpected error: %v", c.in, err)
			}
			if got != c.want {
				t.Errorf("NormalizeThaiPhone(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestNormalizeThaiPhoneNoDigits(t *testing.T) {
	// Dart throws ThaiNumException (a FormatException) when no digits remain;
	// the Go port returns an error wrapping ErrParse.
	for _, in := range []string{"---", "", "abc"} {
		got, err := NormalizeThaiPhone(in)
		if err == nil {
			t.Errorf("NormalizeThaiPhone(%q) = %q, want error", in, got)
			continue
		}
		if !errors.Is(err, ErrParse) {
			t.Errorf("NormalizeThaiPhone(%q) error = %v, want wrapping ErrParse", in, err)
		}
	}
}

func TestSpeakThaiPhone(t *testing.T) {
	const want = "ศูนย์ แปด หนึ่ง สอง สาม สี่ ห้า หก เจ็ด แปด"
	got, err := SpeakThaiPhone("0812345678")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("SpeakThaiPhone = %q, want %q", got, want)
	}
}

func TestSpeakThaiPhoneColloquial(t *testing.T) {
	const want = "ศูนย์ แปด หนึ่ง โท สาม สี่ ห้า หก เจ็ด แปด"
	got, err := SpeakThaiPhoneColloquial("0812345678")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("SpeakThaiPhoneColloquial = %q, want %q", got, want)
	}
}

func TestSpeakThaiPhonePlusSixSixSameAsNational(t *testing.T) {
	intl, err := SpeakThaiPhone("+66812345678")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	natl, err := SpeakThaiPhone("0812345678")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if intl != natl {
		t.Errorf("SpeakThaiPhone(+66...) = %q, want same as national %q", intl, natl)
	}
}

// A leading U+FEFF (BOM/ZWNBSP) before '+' must not hide the country code:
// Dart's String.trimLeft() strips U+FEFF, so "<BOM>+6681234567" reads as a
// landline (08-123-4567 / +6681234567). Go's unicode.IsSpace omits U+FEFF, so
// without isDartTrimSpace this diverged from the Dart port. (Parity regression.)
func TestPhoneBOMBeforePlus(t *testing.T) {
	const in = "\uFEFF+6681234567"
	if got := ThaiPhoneKind(in); got != PhoneLandline {
		t.Errorf("ThaiPhoneKind(%q) = %v, want PhoneLandline", in, got)
	}
	if got, err := FormatThaiPhone(in); err != nil || got != "08-123-4567" {
		t.Errorf("FormatThaiPhone(%q) = %q, %v; want %q, nil", in, got, err, "08-123-4567")
	}
	if got, err := NormalizeThaiPhone(in); err != nil || got != "+6681234567" {
		t.Errorf("NormalizeThaiPhone(%q) = %q, %v; want %q, nil", in, got, err, "+6681234567")
	}
}
