package thainum

import (
	"fmt"
	"strings"
	"unicode"
)

// PhoneKind is the category a Thai phone number falls into, inferred from its
// prefix and length. It is reported conservatively: anything that does not
// clearly match a known pattern is PhoneUnknown rather than guessed.
type PhoneKind int

const (
	// PhoneMobile is a 10-digit mobile number: 0 + 6/8/9 + 8 more digits
	// (e.g. 081-234-5678).
	PhoneMobile PhoneKind = iota

	// PhoneLandline is a fixed-line (landline) number: 0 + an area code + more
	// digits, 9 digits total (e.g. Bangkok 02-123-4567, provincial 0X-XXX-XXXX).
	PhoneLandline

	// PhoneTollFree is a toll-free / special commercial number beginning 1800.
	PhoneTollFree

	// PhoneShortCode is a short service code (3-4 digits, often 1xxx,
	// e.g. 1669, 191).
	PhoneShortCode

	// PhoneUnknown does not match any recognised Thai pattern.
	PhoneUnknown
)

// String returns the lowerCamel name of the kind (matching the Dart
// ThaiPhoneKind enum value names: "mobile", "landline", "tollFree",
// "shortCode", "unknown").
func (k PhoneKind) String() string {
	switch k {
	case PhoneMobile:
		return "mobile"
	case PhoneLandline:
		return "landline"
	case PhoneTollFree:
		return "tollFree"
	case PhoneShortCode:
		return "shortCode"
	case PhoneUnknown:
		return "unknown"
	default:
		return fmt.Sprintf("PhoneKind(%d)", int(k))
	}
}

// phoneDigitsOnly keeps only the ASCII digit characters of s, converting Thai
// numerals to ASCII first and dropping every separator (spaces, dashes, dots,
// parentheses, a leading +, …). The result is a bare ASCII digit string.
func phoneDigitsOnly(s string) string {
	arabic := ToArabicDigits(s)
	var b strings.Builder
	for _, r := range arabic {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// phoneToNational strips a leading Thailand country code (66, optionally
// introduced by a +) from a digit string and restores the national leading 0,
// so both "+66812345678" and "0812345678" reduce to the same national-format
// "0812345678". A string that does not start with 66 is returned unchanged.
//
// A leading 66 is only treated as a country code when the original input
// carried a + or the 66-form length matches a national number with the 0
// restored (11 digits).
func phoneToNational(raw, digits string) string {
	// Match Dart String.trimLeft(), whose whitespace set includes U+FEFF
	// (ZWNBSP/BOM) in addition to the Unicode White_Space chars covered by
	// unicode.IsSpace; otherwise a BOM directly before '+' hides the country
	// code and diverges from the Dart port.
	hadPlus := strings.HasPrefix(strings.TrimLeftFunc(raw, isDartTrimSpace), "+")
	if strings.HasPrefix(digits, "66") && (hadPlus || len(digits) == 11) {
		return "0" + digits[2:]
	}
	return digits
}

// isDartTrimSpace reports whether r is trimmed by Dart's String.trimLeft/trim,
// which is Unicode White_Space plus U+FEFF (Go's unicode.IsSpace omits U+FEFF).
func isDartTrimSpace(r rune) bool {
	return unicode.IsSpace(r) || r == '\uFEFF'
}

// ThaiPhoneKind classifies a Thai phone number by prefix and length.
//
//	ThaiPhoneKind("0812345678") // PhoneMobile
//	ThaiPhoneKind("021234567")  // PhoneLandline (Bangkok)
//	ThaiPhoneKind("1800000000") // PhoneTollFree
//	ThaiPhoneKind("1669")       // PhoneShortCode
//	ThaiPhoneKind("12345")      // PhoneUnknown
//
// Separators, spaces, Thai numerals and a +66 country code are accepted (the
// +66 form is normalised to the national 0… form first). The rules are:
//
//   - PhoneMobile — 10 digits, 0 followed by 6, 8 or 9.
//   - PhoneTollFree — begins with 1800 (8-10 digits).
//   - PhoneLandline — 9 digits beginning with 0 that is not a mobile prefix
//     (Bangkok 02…, provincial 03…/04…/05…/07…, etc.).
//   - PhoneShortCode — 3 or 4 digits not otherwise matched (e.g. 191, 1669).
//   - PhoneUnknown — everything else.
//
// Distinguishing landline area codes from one another is best-effort: this only
// reports the broad PhoneLandline kind, not the province, because Thai area
// codes are variable-length and overlap the mobile space at the prefix level.
func ThaiPhoneKind(s string) PhoneKind {
	digits := phoneToNational(s, phoneDigitsOnly(s))
	if digits == "" {
		return PhoneUnknown
	}

	// Toll-free 1800… (checked before short codes so 1800xxxxxx isn't a code).
	if strings.HasPrefix(digits, "1800") && len(digits) >= 8 && len(digits) <= 10 {
		return PhoneTollFree
	}

	if strings.HasPrefix(digits, "0") && len(digits) == 10 {
		second := digits[1]
		if second == '6' || second == '8' || second == '9' {
			return PhoneMobile
		}
		return PhoneUnknown
	}

	if strings.HasPrefix(digits, "0") && len(digits) == 9 {
		return PhoneLandline
	}

	// Short service codes: 3-4 digits, not starting with 0.
	if (len(digits) == 3 || len(digits) == 4) && !strings.HasPrefix(digits, "0") {
		return PhoneShortCode
	}

	return PhoneUnknown
}

// FormatThaiPhone formats a Thai phone number with conventional grouping.
//
//	FormatThaiPhone("0812345678") // "081-234-5678", nil
//	FormatThaiPhone("๐๘๑๒๓๔๕๖๗๘") // "081-234-5678", nil
//	FormatThaiPhone("021234567")  // "02-123-4567", nil
//
// Existing separators, spaces, Thai numerals and a leading +66 country code are
// accepted (the +66 form is normalised to 0… first). The number is first
// classified with ThaiPhoneKind and grouped only when the kind warrants it, so
// a number that merely happens to be 10 digits long (e.g. a 1800… toll-free) is
// never mis-grouped as a mobile number:
//
//   - PhoneMobile (10-digit 0 + 6/8/9) → 3-3-4 ("081-234-5678").
//   - PhoneLandline (9-digit) → 2-3-4 ("02-123-4567"). This is the common
//     Bangkok shape and a reasonable default for provincial numbers; precise
//     provincial area-code lengths vary, so landline grouping is best-effort.
//   - PhoneTollFree (1800…), PhoneShortCode (191, 1669, …) and anything of an
//     PhoneUnknown kind are returned as their bare digit string (no separators
//     inserted) rather than mis-grouped.
//
// If the input contains no digits at all it is returned unchanged (a formatter
// does not silently emit an empty string). The error return is always nil; it
// is present for signature symmetry with the rest of the package.
func FormatThaiPhone(s string) (string, error) {
	digits := phoneToNational(s, phoneDigitsOnly(s))
	if digits == "" {
		return s, nil
	}
	switch ThaiPhoneKind(s) {
	case PhoneMobile:
		return digits[0:3] + "-" + digits[3:6] + "-" + digits[6:], nil
	case PhoneLandline:
		return digits[0:2] + "-" + digits[2:5] + "-" + digits[5:], nil
	default: // PhoneTollFree, PhoneShortCode, PhoneUnknown
		return digits, nil
	}
}

// NormalizeThaiPhone normalises a Thai national number to E.164 international
// form (+66…).
//
//	NormalizeThaiPhone("0812345678")  // "+66812345678", nil
//	NormalizeThaiPhone("02-123-4567") // "+6621234567", nil
//	NormalizeThaiPhone("+66812345678") // "+66812345678" (idempotent), nil
//
// Assumptions: the input is a Thailand number, so a single leading national
// trunk 0 is dropped and +66 is prefixed. A number already in +66 / 66… form is
// returned in canonical +66… form. Separators, spaces and Thai numerals are
// accepted. Returns an error wrapping ErrParse if no digits remain.
func NormalizeThaiPhone(s string) (string, error) {
	national := phoneToNational(s, phoneDigitsOnly(s))
	if national == "" {
		return "", fmt.Errorf("%w: no digits in phone number", ErrParse)
	}
	// Drop a single national trunk 0, then prefix +66.
	body := national
	if strings.HasPrefix(body, "0") {
		body = body[1:]
	}
	return "+66" + body, nil
}

// SpeakThaiPhone reads a Thai phone number digit-by-digit in Thai words
// (อ่านเรียงตัว) via SpeakDigits.
//
//	SpeakThaiPhone("0812345678")
//	// "ศูนย์ แปด หนึ่ง สอง สาม สี่ ห้า หก เจ็ด แปด", nil
//
// Separators, spaces, Thai numerals and a leading +66 are accepted; the number
// is normalised to its national 0… form first, so "+66812345678" reads the same
// as "0812345678". For the spoken-distinct form where the digit 2 reads as โท,
// use SpeakThaiPhoneColloquial. The error return is always nil.
func SpeakThaiPhone(s string) (string, error) {
	national := phoneToNational(s, phoneDigitsOnly(s))
	return DigitSpeaker{}.Speak(national)
}

// SpeakThaiPhoneColloquial reads a Thai phone number digit-by-digit like
// SpeakThaiPhone, but reads the digit 2 as โท — the spoken-distinct form used
// when reading digits aloud (e.g. over the phone) so it is not confused with
// สาม. It mirrors Dart's speakThaiPhone(..., colloquialTwo: true).
//
//	SpeakThaiPhoneColloquial("0812345678")
//	// "ศูนย์ แปด หนึ่ง โท สาม สี่ ห้า หก เจ็ด แปด", nil
//
// Separators, spaces, Thai numerals and a leading +66 are accepted. The error
// return is always nil.
func SpeakThaiPhoneColloquial(s string) (string, error) {
	national := phoneToNational(s, phoneDigitsOnly(s))
	return DigitSpeaker{ColloquialTwo: true}.Speak(national)
}
