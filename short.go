package thainum

import (
	"fmt"
	"math/big"
	"strings"
)

// Abbreviated large-number reading (the สากล พัน/ล้าน convention): render
// values whose magnitude is >= 10^6 with a single "scale unit" word spaced a
// factor of 1000 apart. The coefficient is computed exactly from the decimal
// digit string (never via floating-point division), mirroring the integer and
// string approach used by FormatSatang/Spell.
//
// Scale units (powers of ten, 1000 apart):
//
//	10^6  = ล้าน          10^15 = พันล้านล้าน
//	10^9  = พันล้าน        10^18 = ล้านล้านล้าน
//	10^12 = ล้านล้าน       ...
//
// General rule for the unit at 10^(6+3k), k >= 0: write 'ล้าน' repeated
// (k/2 + 1) times, prefixed with 'พัน' when k is odd.

var bigOneMillion = big.NewInt(1000000)

// shortUnitWord builds the Thai scale-unit word for the unit at 10^(6+3k).
//
// k == 0 -> "ล้าน", 1 -> "พันล้าน", 2 -> "ล้านล้าน", 3 -> "พันล้านล้าน",
// 4 -> "ล้านล้านล้าน", …. Computed generatively so it is correct at arbitrary
// big.Int scale.
func shortUnitWord(k int) string {
	repeats := k/2 + 1
	var b strings.Builder
	if k%2 == 1 {
		b.WriteString("พัน")
	}
	for i := 0; i < repeats; i++ {
		b.WriteString("ล้าน")
	}
	return b.String()
}

// shortCoeff is a coefficient resolved against a scale unit: its integer part
// (a digit string holding 1..999), its trimmed fractional digit string (may be
// empty), and the chosen unit index k (the unit is 10^(6+3k)).
type shortCoeff struct {
	intPart string
	frac    string
	k       int
}

// resolveShort rounds the magnitude digit string mag (no sign, already known to
// be >= 10^6) against the largest scale unit, half-away-from-zero to decimals
// places, handling the carry into the next unit. Float-free: all arithmetic is
// on the digit string.
func resolveShort(mag string, decimals int) shortCoeff {
	// Largest k with 10^(6+3k) <= value, i.e. value has at least 6+3k+1 digits.
	n := len(mag) // n >= 7 here
	k := (n - 1 - 6) / 3
	d := 6 + 3*k // number of fractional (low) digits for this unit

	// Split into integer part (high) and fractional part (low d digits).
	intDigits := mag[:n-d]
	fracDigits := mag[n-d:]

	// Round the fractional part to `decimals` places, half-away-from-zero.
	// Everything below stays as integer digit-string arithmetic.
	if decimals < len(fracDigits) {
		keep := fracDigits[:decimals]
		roundUp := fracDigits[decimals]-'0' >= 5
		combined := intDigits + keep // integer view of intPart.keep
		if roundUp {
			combined = incDecimalString(combined)
		}
		// Re-split: the last `decimals` chars are the fractional part.
		if decimals == 0 {
			intDigits = combined
			fracDigits = ""
		} else {
			// combined length is at least decimals (intDigits was >= 1 char).
			intDigits = combined[:len(combined)-decimals]
			fracDigits = combined[len(combined)-decimals:]
		}
	}
	// else: decimals >= available fraction digits -> no rounding needed.

	// Carry promotion: a round-up can push the integer part to >= 1000, which
	// means the value reached the next scale unit (coefficient becomes 1.000…).
	// Promote one unit at a time; after promotion the fraction is all zeros.
	for len(intDigits) > 3 {
		k++
		d += 3
		intDigits = intDigits[:len(intDigits)-3]
		fracDigits = "" // promoted coefficient is whole (1.000…)
	}

	// Trim trailing zeros from the fractional part.
	end := len(fracDigits)
	for end > 0 && fracDigits[end-1] == '0' {
		end--
	}
	fracDigits = fracDigits[:end]

	return shortCoeff{intPart: intDigits, frac: fracDigits, k: k}
}

// incDecimalString adds 1 to the lowest place of a non-negative decimal digit
// string, propagating the carry; grows the string by one digit on overflow.
func incDecimalString(s string) string {
	units := []byte(s)
	i := len(units) - 1
	for i >= 0 {
		if units[i] == '9' {
			units[i] = '0'
			i--
		} else {
			units[i]++
			return string(units)
		}
	}
	return "1" + string(units)
}

func checkDecimals(decimals int) error {
	if decimals < 0 {
		return fmt.Errorf("thainum: decimals must be non-negative, got %d", decimals)
	}
	return nil
}

// SpellShort spells value in Thai words using an abbreviated scale unit
// (ล้าน / พันล้าน / ล้านล้าน …) when |value| >= 10^6.
//
// For |value| < 10^6 it falls back to Spell (the full reading, no unit word).
// The fractional part of the coefficient is read digit-by-digit after "จุด",
// exactly like SpellDecimal. decimals is the number of fractional places kept,
// rounded half-away-from-zero with trailing zeros trimmed; a whole coefficient
// drops the "จุด" entirely. Pass decimals = 2 for the conventional default.
//
//	SpellShort(1500000, 2)    // "หนึ่งจุดห้าล้าน"
//	SpellShort(15000000, 2)   // "สิบห้าล้าน"
//	SpellShort(2300000000, 2) // "สองจุดสามพันล้าน"
//	SpellShort(-1500000, 2)   // "ลบหนึ่งจุดห้าล้าน"
//
// It panics if decimals is negative; use SpellShortErr for a checked variant.
func SpellShort(value int64, decimals int) string {
	return SpellShortBig(new(big.Int).SetInt64(value), decimals)
}

// SpellShortBig is the big.Int form of SpellShort; it works to arbitrary scale.
//
// It panics if decimals is negative; use SpellShortBigErr for a checked variant.
func SpellShortBig(value *big.Int, decimals int) string {
	s, err := SpellShortBigErr(value, decimals)
	if err != nil {
		panic(err)
	}
	return s
}

// SpellShortErr is SpellShort returning an error instead of panicking when
// decimals is negative.
func SpellShortErr(value int64, decimals int) (string, error) {
	return SpellShortBigErr(new(big.Int).SetInt64(value), decimals)
}

// SpellShortBigErr is the big.Int form of SpellShortErr.
func SpellShortBigErr(value *big.Int, decimals int) (string, error) {
	if err := checkDecimals(decimals); err != nil {
		return "", err
	}
	if value == nil {
		value = new(big.Int)
	}
	neg := value.Sign() < 0
	mag := new(big.Int).Abs(value)
	if mag.Cmp(bigOneMillion) < 0 {
		return SpellBig(value), nil // full reading, carries its own ลบ
	}
	c := resolveShort(mag.String(), decimals)
	unit := shortUnitWord(c.k)
	intPart, _ := new(big.Int).SetString(c.intPart, 10)
	intWords := SpellBig(intPart)
	var b strings.Builder
	if neg {
		b.WriteString("ลบ")
	}
	b.WriteString(intWords)
	if c.frac != "" {
		b.WriteString("จุด")
		for i := 0; i < len(c.frac); i++ {
			b.WriteString(numberThai[c.frac[i]-'0'])
		}
	}
	b.WriteString(unit)
	return b.String(), nil
}

// FormatShort formats value with an abbreviated scale unit and an Arabic-numeral
// coefficient when |value| >= 10^6: "1.5 ล้าน", "2.3 พันล้าน", "50 พันล้าน".
// For |value| < 10^6 it falls back to FormatInt (the grouped full number, no
// unit word).
//
// decimals is the number of fractional places kept, rounded half-away-from-zero
// with trailing zeros trimmed. With thaiDigits true only the ASCII digits of
// the coefficient become Thai numerals; the space, the sign, and the unit word
// are unchanged.
//
//	FormatShort(1500000, 2, false)  // "1.5 ล้าน"
//	FormatShort(15000000, 2, false) // "15 ล้าน"
//	FormatShort(1234567, 2, false)  // "1.23 ล้าน"
//	FormatShort(2000000, 2, false)  // "2 ล้าน"
//	FormatShort(1500000, 2, true)   // "๑.๕ ล้าน"
//
// It panics if decimals is negative; use FormatShortErr for a checked variant.
func FormatShort(value int64, decimals int, thaiDigits bool) string {
	return FormatShortBig(new(big.Int).SetInt64(value), decimals, thaiDigits)
}

// FormatShortBig is the big.Int form of FormatShort; it works to arbitrary scale.
//
// It panics if decimals is negative; use FormatShortBigErr for a checked variant.
func FormatShortBig(value *big.Int, decimals int, thaiDigits bool) string {
	s, err := FormatShortBigErr(value, decimals, thaiDigits)
	if err != nil {
		panic(err)
	}
	return s
}

// FormatShortErr is FormatShort returning an error instead of panicking when
// decimals is negative.
func FormatShortErr(value int64, decimals int, thaiDigits bool) (string, error) {
	return FormatShortBigErr(new(big.Int).SetInt64(value), decimals, thaiDigits)
}

// FormatShortBigErr is the big.Int form of FormatShortErr.
func FormatShortBigErr(value *big.Int, decimals int, thaiDigits bool) (string, error) {
	if err := checkDecimals(decimals); err != nil {
		return "", err
	}
	if value == nil {
		value = new(big.Int)
	}
	neg := value.Sign() < 0
	mag := new(big.Int).Abs(value)
	if mag.Cmp(bigOneMillion) < 0 {
		// Grouped full number, no unit word. FormatInt has no thaiDigits
		// parameter, so apply the digit mapping here. ToThaiDigits maps only
		// ASCII digits, leaving the commas and the leading '-' as ASCII — which
		// matches the Dart formatInt(thaiDigits:) contract exactly.
		out := FormatInt(value.Int64())
		if thaiDigits {
			out = ToThaiDigits(out)
		}
		return out, nil
	}
	c := resolveShort(mag.String(), decimals)
	unit := shortUnitWord(c.k)
	coeff := c.intPart
	if c.frac != "" {
		coeff = coeff + "." + c.frac
	}
	if thaiDigits {
		coeff = ToThaiDigits(coeff)
	}
	sign := ""
	if neg {
		sign = "-"
	}
	return sign + coeff + " " + unit, nil
}
