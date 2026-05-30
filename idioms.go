package thainum

import (
	"fmt"
	"strings"
)

// QuantityUnit is one of the small, fixed set of Thai counting idioms this
// package recognises. The set is deliberately closed to ครึ่ง / คู่ / โหล / กุรุส
// and is not a general system of Thai classifiers (ลักษณนาม).
type QuantityUnit int

const (
	// UnitHalf is ครึ่ง — one half (value 0.5).
	UnitHalf QuantityUnit = iota
	// UnitPair is คู่ — a pair (value 2).
	UnitPair
	// UnitDozen is โหล — a dozen (value 12).
	UnitDozen
	// UnitGross is กุรุส — a gross, i.e. twelve dozen (value 144).
	UnitGross
)

// QuantityWord returns the Thai word for unit:
//
//	QuantityWord(UnitDozen) // "โหล"
//
// It returns "" for an unrecognised QuantityUnit.
func QuantityWord(unit QuantityUnit) string {
	switch unit {
	case UnitHalf:
		return "ครึ่ง"
	case UnitPair:
		return "คู่"
	case UnitDozen:
		return "โหล"
	case UnitGross:
		return "กุรุส"
	default:
		return ""
	}
}

// QuantityValue returns the numeric value of unit, split into an integer part
// and a half flag so the result is exact and float-free.
//
//	whole, half := QuantityValue(UnitGross) // 144, false
//	whole, half := QuantityValue(UnitHalf)  // 0, true   (i.e. 0.5)
//
// The numeric value is whole + 0.5 when half is true. UnitHalf is the only unit
// with a fractional value. An unrecognised QuantityUnit returns (0, false).
func QuantityValue(unit QuantityUnit) (whole int64, half bool) {
	switch unit {
	case UnitHalf:
		return 0, true
	case UnitPair:
		return 2, false
	case UnitDozen:
		return 12, false
	case UnitGross:
		return 144, false
	default:
		return 0, false
	}
}

// ParseQuantity parses a narrow, opt-in set of Thai counting idioms into a
// numeric value, returned as an integer part plus a half flag (the value is
// whole + 0.5 when half is true). This keeps fractional results exact and
// avoids float64. It is a focused convenience that sits beside the integer
// grammar — it does not extend ParseInt and does not recognise general
// classifiers.
//
// Exactly the following inputs are supported:
//
//	ParseQuantity("ครึ่ง")        // 0, true   (0.5)
//	ParseQuantity("คู่")          // 2, false
//	ParseQuantity("โหล")          // 12, false
//	ParseQuantity("กุรุส")        // 144, false
//	ParseQuantity("สองครึ่ง")     // 2, true   (2.5)
//	ParseQuantity("สิบครึ่ง")     // 10, true  (10.5)
//	ParseQuantity("สอง")          // 2, false
//	ParseQuantity("๑๒")           // 12, false (Thai digits accepted, like ParseInt)
//
// Rules and limits:
//   - A bare ครึ่ง / คู่ / โหล / กุรุส reads as its unit value.
//   - A Thai integer reading optionally followed by a single trailing ครึ่ง
//     adds 0.5. ครึ่ง may only appear once and only as the suffix.
//   - คู่ / โหล / กุรุส are standalone only: they may not be combined with a
//     number or with ครึ่ง here (e.g. "สองโหล" for 24 is intentionally not
//     supported — it needs a classifier grammar this package does not provide).
//   - The integer part is parsed by ParseInt.
//
// It returns an error wrapping ErrParse on anything outside the supported set.
func ParseQuantity(words string) (whole int64, half bool, err error) {
	s := strings.TrimSpace(ToArabicDigits(words))
	if s == "" {
		return 0, false, fmt.Errorf("%w: empty input", ErrParse)
	}

	// Standalone single-unit words.
	switch s {
	case "คู่":
		return 2, false, nil
	case "โหล":
		return 12, false, nil
	case "กุรุส":
		return 144, false, nil
	case "ครึ่ง":
		return 0, true, nil
	}

	const krueng = "ครึ่ง"

	// <integer> + trailing 'ครึ่ง' (e.g. 'สองครึ่ง' -> 2.5).
	if strings.HasSuffix(s, krueng) {
		head := s[:len(s)-len(krueng)]
		if head == "" {
			// Should have matched the bare 'ครึ่ง' above; defensive.
			return 0, true, nil
		}
		if strings.Contains(head, krueng) {
			return 0, false, fmt.Errorf("%w: %q may appear only once", ErrParse, krueng)
		}
		base, perr := ParseInt(head)
		if perr != nil {
			return 0, false, perr
		}
		return base, true, nil
	}

	// Otherwise it must be a plain Thai integer reading.
	base, perr := ParseInt(s)
	if perr != nil {
		return 0, false, perr
	}
	return base, false, nil
}

// ParseHalfBaht parses the idiom "ครึ่งบาท" ("half a baht") into 50 satang,
// matching the integer-satang convention used by ParseBaht and BahtSatang.
//
//	ParseHalfBaht("ครึ่งบาท")     // 50, nil
//	ParseHalfBaht("  ครึ่งบาท  ") // 50, nil
//
// This is intentionally a tiny, single-purpose helper. Anything other than
// exactly "ครึ่งบาท" (after digit-normalisation and trimming) returns an error
// wrapping ErrParse; use ParseBaht for general baht text.
func ParseHalfBaht(text string) (int64, error) {
	s := strings.TrimSpace(ToArabicDigits(text))
	if s == "ครึ่งบาท" {
		return 50, nil
	}
	return 0, fmt.Errorf("%w: ParseHalfBaht expects exactly %q", ErrParse, "ครึ่งบาท")
}
