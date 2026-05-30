package thainum

// SatangRounding selects how the fractional tail beyond two decimal places is
// resolved when a decimal baht string is converted to integer satang.
//
// A baht amount has at most two fractional digits of precision (one satang =
// 0.01 baht). When the input string carries more digits than that ("12.345",
// "0.005"), the digits beyond the second must be folded into the satang value
// somehow. The mode decides how.
//
// The conversion is done entirely on the digit string with integer arithmetic
// — there is no float64 anywhere on the path — so none of these modes can
// introduce binary-floating-point error.
//
// Which mode to use is an accounting/legal question, not a numeric one. This
// mirrors the Dart thainum SatangRounding enum.
type SatangRounding int

const (
	// RoundHalfAwayFromZero rounds half away from zero. This is the default
	// (the zero value) and the historical behaviour: "0.005" -> 1 satang,
	// "-0.005" -> -1 satang. It matches everyday "round the receipt"
	// intuition and Go's math.Round. Mirrors Dart SatangRounding.halfAwayFromZero.
	RoundHalfAwayFromZero SatangRounding = iota

	// RoundHalfEven is banker's rounding (round-half-to-even): "0.005" -> 0
	// satang, "0.015" -> 2 satang. It reduces cumulative bias over many
	// roundings and is the IEEE-754 default. Mirrors Dart SatangRounding.halfEven.
	RoundHalfEven

	// RoundTruncate drops the extra digits (rounds toward zero): "0.009" -> 0
	// satang, "-0.009" -> 0 satang. Used where the law forbids rounding up
	// against the payer. Mirrors Dart SatangRounding.truncate.
	RoundTruncate

	// RoundCeil rounds toward positive infinity: "0.001" -> 1 satang,
	// "-0.009" -> 0 satang. Mirrors Dart SatangRounding.ceil.
	RoundCeil

	// RoundFloor rounds toward negative infinity: "0.009" -> 0 satang,
	// "-0.001" -> -1 satang. Mirrors Dart SatangRounding.floor.
	RoundFloor
)

// roundUpDecision decides whether the truncated two-decimal satang magnitude
// (cents, 0..99) should be incremented by one, given the rounding mode, the
// sign of the amount (neg), and the discarded fractional tail of frac (its
// digits at index 2 and beyond).
//
// It mirrors the switch in Dart's _parseToSatang: cmp classifies the tail
// relative to one half-satang (0 exact-zero, -1 below half, 1 above half,
// 2 exactly half), and a single +1 increment is decided from that.
func roundUpDecision(r SatangRounding, neg bool, frac string) bool {
	cmp := compareTailToHalf(frac)
	switch r {
	case RoundHalfAwayFromZero:
		// Round up on >= half (the historical behaviour: third digit >= '5').
		return cmp == 1 || cmp == 2
	case RoundHalfEven:
		// Round up on > half; on exactly half, only if it makes cents even.
		// cents is even iff incrementing it makes it odd, so we round up at
		// exactly-half only when the truncated cents value is odd.
		return cmp == 1 || (cmp == 2 && centsIsOdd(frac))
	case RoundTruncate:
		return false
	case RoundCeil:
		// Toward +inf: a positive amount with any remainder rounds its
		// magnitude up; a negative amount truncates its magnitude.
		return !neg && cmp != 0
	case RoundFloor:
		// Toward -inf: a negative amount with any remainder rounds its
		// magnitude up; a positive amount truncates its magnitude.
		return neg && cmp != 0
	default:
		return cmp == 1 || cmp == 2
	}
}

// centsIsOdd reports whether the truncated two-decimal satang magnitude (the
// first two fractional digits read as a number 0..99) is odd. Only the second
// fractional digit determines parity. Mirrors Dart's cents.isOdd test.
func centsIsOdd(frac string) bool {
	if len(frac) >= 2 {
		return (frac[1]-'0')%2 == 1
	}
	return false
}

// compareTailToHalf compares the discarded fractional tail (digits at index 2
// and beyond of frac) to one half-satang. It returns 0 for an exactly-zero
// remainder, -1 for below half, 1 for above half, and 2 for exactly half.
//
// Mirrors Dart's _compareTailToHalf, operating on the digit bytes directly so
// no float is involved.
func compareTailToHalf(frac string) int {
	if len(frac) < 3 {
		return 0
	}
	first := frac[2]
	if first < '5' {
		// First dropped digit < '5': below half, unless every dropped digit
		// is zero (then the remainder is exactly zero).
		for i := 2; i < len(frac); i++ {
			if frac[i] != '0' {
				return -1
			}
		}
		return 0
	}
	if first > '5' {
		return 1 // first dropped digit > '5' -> above half
	}
	// First dropped digit == '5': exactly half unless a later digit is non-zero.
	for i := 3; i < len(frac); i++ {
		if frac[i] != '0' {
			return 1
		}
	}
	return 2
}
