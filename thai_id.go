package thainum

import (
	"fmt"
	"strings"
)

// ThaiIDKind is the category a Thai National ID encodes in its leading digit,
// per the Department of Provincial Administration (DOPA, กรมการปกครอง).
type ThaiIDKind int

const (
	// ThaiBornRegisteredOnTime (leading 1): Thai national whose birth was
	// registered within the deadline, born on/after 1 Jan 1984 (พ.ศ. 2527).
	ThaiBornRegisteredOnTime ThaiIDKind = iota
	// ThaiBornRegisteredLate (leading 2): Thai national whose birth was
	// registered late, born on/after 1 Jan 1984.
	ThaiBornRegisteredLate
	// ThaiInRegistryBefore1984 (leading 3): Thai national who was already in
	// the household registry before 1 Jan 1984.
	ThaiInRegistryBefore1984
	// ThaiBornBefore1984NotRegistered (leading 4): Thai national born before
	// 1 Jan 1984 but not yet in the registry at the 1984 census.
	ThaiBornBefore1984NotRegistered
	// ThaiAddedLater (leading 5): Thai national added to the registry later
	// (missed the census, or special cases).
	ThaiAddedLater
	// ForeignerTemporary (leading 6): foreign national living in Thailand
	// temporarily, or who entered unlawfully.
	ForeignerTemporary
	// ChildOfForeignerTemporary (leading 7): child of a category-6 person,
	// born in Thailand.
	ChildOfForeignerTemporary
	// NaturalisedOrPermanentResident (leading 8): naturalised Thai citizen or
	// a foreigner with permanent residence.
	NaturalisedOrPermanentResident
	// ThaiIDUnknown is returned for leading digit 0 or 9 (not assigned by DOPA
	// to a person category), or any otherwise unrecognised id. Returned
	// conservatively rather than guessing.
	ThaiIDUnknown
)

// String returns a short stable identifier for the kind (useful in tests and
// logs); it is not a user-facing Thai label.
func (k ThaiIDKind) String() string {
	switch k {
	case ThaiBornRegisteredOnTime:
		return "ThaiBornRegisteredOnTime"
	case ThaiBornRegisteredLate:
		return "ThaiBornRegisteredLate"
	case ThaiInRegistryBefore1984:
		return "ThaiInRegistryBefore1984"
	case ThaiBornBefore1984NotRegistered:
		return "ThaiBornBefore1984NotRegistered"
	case ThaiAddedLater:
		return "ThaiAddedLater"
	case ForeignerTemporary:
		return "ForeignerTemporary"
	case ChildOfForeignerTemporary:
		return "ChildOfForeignerTemporary"
	case NaturalisedOrPermanentResident:
		return "NaturalisedOrPermanentResident"
	default:
		return "ThaiIDUnknown"
	}
}

// ParseThaiID strips dashes/spaces and converts Thai numerals to ASCII,
// returning the bare 13 ASCII digits of a Thai National / personal Tax ID.
//
//	ParseThaiID("1-2345-67890-12-1") // "1234567890121", nil
//	ParseThaiID("๑๑๐๑๗๐๐๒๓๐๗๐๕")     // "1101700230705", nil
//
// It returns an error wrapping ErrParse (test with errors.Is(err, ErrParse))
// if the input contains a character other than a digit, dash or space, or if
// the result is not exactly 13 digits.
func ParseThaiID(id string) (string, error) {
	arabic := ToArabicDigits(id)
	var b strings.Builder
	for _, r := range arabic {
		switch {
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == ' ':
			// dash or space — allowed separators
			continue
		default:
			return "", fmt.Errorf("%w: invalid character %q in Thai ID", ErrParse, string(r))
		}
	}
	digits := b.String()
	if len(digits) != 13 {
		return "", fmt.Errorf("%w: Thai ID must be exactly 13 digits (got %d)", ErrParse, len(digits))
	}
	return digits, nil
}

// thaiIDCheckDigit computes the MOD-11 check digit for the first 12 digits of
// digits13.
//
// For digits d0..d12, sum = Σ d[i]*(13-i) for i=0..11 (weights 13,12,…,2); the
// check digit is (11 - (sum % 11)) % 10.
//
// Note: the third digit (index 2) carries weight 11, and d*11 mod 11 == 0 for
// every digit, so a change to that single digit is invisible to the checksum —
// an inherent blind spot of the MOD-11 weighting, not a defect.
func thaiIDCheckDigit(digits13 string) int {
	sum := 0
	for i := 0; i < 12; i++ {
		d := int(digits13[i] - '0')
		sum += d * (13 - i)
	}
	return (11 - (sum % 11)) % 10
}

// IsValidThaiID reports whether id is a syntactically valid Thai National ID:
// exactly 13 digits (after stripping dashes/spaces and converting Thai
// numerals) with a correct MOD-11 checksum.
//
//	IsValidThaiID("1101700230708") // true (synthetic, checksum-valid)
func IsValidThaiID(id string) bool {
	digits, err := ParseThaiID(id)
	if err != nil {
		return false
	}
	d12 := int(digits[12] - '0')
	return d12 == thaiIDCheckDigit(digits)
}

// IsValidThaiTaxID reports whether id is a valid Thai personal Tax ID. For
// individuals the Tax ID equals the National ID, so this delegates to
// IsValidThaiID.
func IsValidThaiTaxID(id string) bool { return IsValidThaiID(id) }

// FormatThaiID formats a Thai National ID into the canonical
// X-XXXX-XXXXX-XX-X shape.
//
//	FormatThaiID("1234567890121") // "1-2345-67890-12-1", nil
//
// It accepts already-grouped or Thai-numeral input (it re-parses via
// ParseThaiID). It returns an error wrapping ErrParse if not 13 digits.
func FormatThaiID(id string) (string, error) {
	d, err := ParseThaiID(id)
	if err != nil {
		return "", err
	}
	return d[0:1] + "-" + d[1:5] + "-" + d[5:10] + "-" + d[10:12] + "-" + d[12:13], nil
}

// ClassifyThaiID classifies id by its leading digit per DOPA. It returns
// ThaiIDUnknown for leading digit 0/9 or any input that is not 13 digits,
// rather than guessing.
func ClassifyThaiID(id string) ThaiIDKind {
	digits, err := ParseThaiID(id)
	if err != nil {
		return ThaiIDUnknown
	}
	switch digits[0] - '0' {
	case 1:
		return ThaiBornRegisteredOnTime
	case 2:
		return ThaiBornRegisteredLate
	case 3:
		return ThaiInRegistryBefore1984
	case 4:
		return ThaiBornBefore1984NotRegistered
	case 5:
		return ThaiAddedLater
	case 6:
		return ForeignerTemporary
	case 7:
		return ChildOfForeignerTemporary
	case 8:
		return NaturalisedOrPermanentResident
	default:
		return ThaiIDUnknown
	}
}

// SpeakThaiID reads a Thai National ID digit-by-digit as Thai words
// (อ่านเรียงตัว) via SpeakDigits.
//
//	SpeakThaiID("1101700230708")
//	// "หนึ่ง หนึ่ง ศูนย์ หนึ่ง เจ็ด ศูนย์ ศูนย์ สอง สาม ศูนย์ เจ็ด ศูนย์ แปด", nil
//
// It returns an error wrapping ErrParse if id is not 13 digits.
func SpeakThaiID(id string) (string, error) {
	digits, err := ParseThaiID(id)
	if err != nil {
		return "", err
	}
	return SpeakDigits(digits)
}
