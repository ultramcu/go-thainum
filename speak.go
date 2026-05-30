package thainum

import (
	"fmt"
	"strings"
)

// DigitSpeaker reads a digit string digit-by-digit as Thai words (อ่านเรียงตัว)
// — a per-digit reading rather than a quantity, the form used for phone
// numbers, account numbers, PINs and years-read-as-digits.
//
// The zero value reproduces the default reading: a single space separator and
// the literal สอง for the digit 2. Set the fields to customise.
type DigitSpeaker struct {
	// Separator is written between consecutive digit-words. The zero value ""
	// is treated as a single space (" ") so the zero-value DigitSpeaker matches
	// the default SpeakDigits reading; use SeparatorSet with an explicit empty
	// string to join the words with no separator at all.
	Separator string

	// SeparatorSet must be true to honour an explicitly empty Separator. When
	// false (the zero value) an empty Separator means "use the default space".
	SeparatorSet bool

	// ColloquialTwo, when true, reads the digit 2 as โท — the spoken-distinct
	// form used when reading digits aloud (e.g. over the phone) so it is not
	// confused with สาม. All other digits are unchanged.
	ColloquialTwo bool
}

// sep returns the effective separator for this speaker.
func (ds DigitSpeaker) sep() string {
	if ds.Separator == "" && !ds.SeparatorSet {
		return " "
	}
	return ds.Separator
}

// SpeakDigits reads each digit in s individually as a Thai word
// (ศูนย์ หนึ่ง สอง …), joined by a single space. It is a digit-by-digit reading
// (อ่านเรียงตัว) rather than a quantity, the form used for phone numbers,
// account numbers, PINs and years-read-as-digits.
//
//	SpeakDigits("2566")        // "สอง ห้า หก หก", nil
//	SpeakDigits("081-234-5678") // "ศูนย์ แปด หนึ่ง สอง สาม สี่ ห้า หก เจ็ด แปด", nil
//
// Both Arabic (0-9) and Thai (๐-๙) digits are accepted. Every non-digit rune
// acts as a delimiter: a run of one or more non-digits between two emitted
// digit-words collapses to exactly one separator, and there is never a leading
// or trailing separator. Input with no digits returns "". The error return is
// always nil for this function; it is present for signature symmetry with the
// rest of the package and with SpeakDigitsStrict.
func SpeakDigits(s string) (string, error) {
	return DigitSpeaker{}.Speak(s)
}

// Speak reads each digit in s individually as a Thai word using the speaker's
// Separator and ColloquialTwo settings. Non-digit runes are delimiters
// (collapsing to a single separator, never leading/trailing); input with no
// digits returns "". The error return is always nil; use SpeakStrict to reject
// non-digit input.
func (ds DigitSpeaker) Speak(s string) (string, error) {
	return ds.speak(s, false)
}

// SpeakDigitsStrict reads each digit in s individually as a Thai word like
// SpeakDigits, but treats any non-digit rune as an error rather than a
// delimiter. It returns an error wrapping ErrParse (test with
// errors.Is(err, ErrParse)) on the first non-digit rune. An empty input returns
// "", nil.
//
//	SpeakDigitsStrict("2566")        // "สอง ห้า หก หก", nil
//	SpeakDigitsStrict("081-234")     // "", error wrapping ErrParse
func SpeakDigitsStrict(s string) (string, error) {
	return DigitSpeaker{}.SpeakStrict(s)
}

// SpeakStrict behaves like Speak but rejects any non-digit rune with an error
// wrapping ErrParse instead of treating it as a delimiter.
func (ds DigitSpeaker) SpeakStrict(s string) (string, error) {
	return ds.speak(s, true)
}

// speak is the shared implementation. When strict is true any non-digit rune
// (after Thai→Arabic normalisation) yields an ErrParse-wrapped error; when
// false non-digit runes are delimiters that collapse to a single separator.
func (ds DigitSpeaker) speak(s string, strict bool) (string, error) {
	// Normalise Thai numerals to ASCII so we only deal with one range.
	norm := ToArabicDigits(s)
	sep := ds.sep()
	var b strings.Builder
	emitted := false
	for _, r := range norm {
		if r >= '0' && r <= '9' {
			if emitted {
				b.WriteString(sep)
			}
			d := r - '0'
			if d == 2 && ds.ColloquialTwo {
				b.WriteString("โท")
			} else {
				b.WriteString(numberThai[d])
			}
			emitted = true
			continue
		}
		if strict {
			return "", fmt.Errorf("%w: non-digit %q", ErrParse, string(r))
		}
		// Non-strict: non-digits are skipped here; the emitted flag inserts
		// exactly one separator before the next emitted digit, collapsing any
		// run of delimiters and avoiding leading/trailing separators.
	}
	return b.String(), nil
}
