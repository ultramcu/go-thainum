package thainum

import (
	"errors"
	"testing"
)

// Expected values below are taken from the Dart oracle:
//   thainum.dart/lib/src/speak.dart and
//   thainum.dart/test/speak_digits_test.dart
// They are NOT derived from running this Go implementation.

func TestSpeakDigits(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		// reads each digit individually, not as a quantity
		{"year-digits", "2566", "สอง ห้า หก หก"},
		{"single-zero", "0", "ศูนย์"},
		{"single-nine", "9", "เก้า"},

		// accepts Thai numerals as well as Arabic
		{"thai-digits", "๒๕๖๖", "สอง ห้า หก หก"},
		{"mixed-thai-arabic", "๒566", "สอง ห้า หก หก"},

		// non-digits collapse to a single separator, no leading/trailing
		{"phone-dashes", "081-234-5678", "ศูนย์ แปด หนึ่ง สอง สาม สี่ ห้า หก เจ็ด แปด"},
		{"phone-plain", "0812345678", "ศูนย์ แปด หนึ่ง สอง สาม สี่ ห้า หก เจ็ด แปด"},
		{"phone-spaces-padded", "  081 234 5678  ", "ศูนย์ แปด หนึ่ง สอง สาม สี่ ห้า หก เจ็ด แปด"},

		// empty / no-digit input returns empty string
		{"empty", "", ""},
		{"letters", "abc", ""},
		{"only-delims", "---", ""},

		// default leaves 2 as สอง
		{"single-two-default", "2", "สอง"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := SpeakDigits(tc.in)
			if err != nil {
				t.Fatalf("SpeakDigits(%q) unexpected error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("SpeakDigits(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestSpeakDigitsCustomSeparator(t *testing.T) {
	cases := []struct {
		name string
		in   string
		sep  string
		want string
	}{
		{"dash-separator", "2566", "-", "สอง-ห้า-หก-หก"},
		{"empty-separator", "12", "", "หนึ่งสอง"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// SeparatorSet=true so an explicit empty separator is honoured.
			got, err := DigitSpeaker{Separator: tc.sep, SeparatorSet: true}.Speak(tc.in)
			if err != nil {
				t.Fatalf("Speak(%q) unexpected error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("Speak(%q) sep=%q = %q, want %q", tc.in, tc.sep, got, tc.want)
			}
		})
	}
}

func TestSpeakDigitsColloquialTwo(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"year-colloquial", "2566", "โท ห้า หก หก"},
		{"two-twos", "22", "โท โท"},
		{"other-digits-unchanged", "123", "หนึ่ง โท สาม"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DigitSpeaker{ColloquialTwo: true}.Speak(tc.in)
			if err != nil {
				t.Fatalf("Speak(%q) unexpected error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("Speak(%q) colloquial = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestDigitSpeakerZeroValueMatchesDefault(t *testing.T) {
	// The zero-value DigitSpeaker reproduces the default SpeakDigits reading.
	got, err := DigitSpeaker{}.Speak("2566")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := "สอง ห้า หก หก"; got != want {
		t.Errorf("zero-value Speak = %q, want %q", got, want)
	}
}

func TestSpeakDigitsStrictRejectsNonDigit(t *testing.T) {
	// Strict mode rejects any non-digit rune (unlike the default Dart-parity
	// behaviour, where non-digits are delimiters).
	if _, err := SpeakDigitsStrict("081-234-5678"); !errors.Is(err, ErrParse) {
		t.Errorf("SpeakDigitsStrict(phone) error = %v, want wrapping ErrParse", err)
	}
	if _, err := SpeakDigitsStrict("abc"); !errors.Is(err, ErrParse) {
		t.Errorf("SpeakDigitsStrict(%q) error = %v, want wrapping ErrParse", "abc", err)
	}
}

func TestSpeakDigitsStrictAcceptsDigits(t *testing.T) {
	// Pure digit input (Arabic or Thai) is accepted and matches SpeakDigits.
	cases := []struct {
		in   string
		want string
	}{
		{"2566", "สอง ห้า หก หก"},
		{"๒๕๖๖", "สอง ห้า หก หก"},
		{"", ""},
	}
	for _, tc := range cases {
		got, err := SpeakDigitsStrict(tc.in)
		if err != nil {
			t.Fatalf("SpeakDigitsStrict(%q) unexpected error: %v", tc.in, err)
		}
		if got != tc.want {
			t.Errorf("SpeakDigitsStrict(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
