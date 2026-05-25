package thainum

import (
	"math/big"
	"strconv"
)

// Baht renders a whole-baht amount as Thai Baht text, using the default EtMode.
// The argument is in BAHT (the unit people normally use); for sub-baht precision
// use BahtSatang (satang) or BahtFromString ("21.21").
//
//	Baht(100) // "หนึ่งร้อยบาทถ้วน"
//	Baht(21)  // "ยี่สิบเอ็ดบาทถ้วน"
//	Baht(-5)  // "ลบห้าบาทถ้วน"
func Baht(baht int64) string { return Speller{}.Baht(baht) }

// BahtSatang renders an amount given in satang (1 baht = 100 satang) as Thai
// Baht text. Use this when you need sub-baht precision.
//
//	BahtSatang(2121) // "ยี่สิบเอ็ดบาทยี่สิบเอ็ดสตางค์"
//	BahtSatang(25)   // "ยี่สิบห้าสตางค์"
//	BahtSatang(0)    // "ศูนย์บาทถ้วน"
func BahtSatang(satang int64) string { return Speller{}.BahtSatang(satang) }

// BahtBig renders an arbitrarily large whole-baht amount as Thai Baht text.
func BahtBig(baht *big.Int) string { return Speller{}.BahtBig(baht) }

// BahtSatangBig renders an arbitrarily large satang amount as Thai Baht text.
func BahtSatangBig(satang *big.Int) string { return Speller{}.BahtSatangBig(satang) }

// BahtFromString renders a decimal baht amount written as a string (e.g.
// "21.21") as Thai Baht text. The fractional part is rounded to two decimal
// places, away from zero. Because the input is a string, no float rounding is
// involved.
func BahtFromString(amount string) (string, error) { return Speller{}.BahtFromString(amount) }

// BahtFromFloat is a convenience wrapper that renders a float baht amount. It is
// lossy: the float is formatted to two decimals first. Prefer Baht (whole baht),
// BahtSatang (satang) or BahtFromString for money-grade input.
func BahtFromFloat(amount float64) string { return Speller{}.BahtFromFloat(amount) }

// Baht renders a whole-baht amount as Thai Baht text. It is a thin wrapper over
// BahtSatang (baht × 100).
func (sp Speller) Baht(baht int64) string {
	return sp.BahtSatangBig(new(big.Int).Mul(big.NewInt(baht), hundred))
}

// BahtBig renders a whole-baht amount (big) as Thai Baht text, wrapping
// BahtSatangBig.
func (sp Speller) BahtBig(baht *big.Int) string {
	if baht == nil {
		baht = big.NewInt(0)
	}
	return sp.BahtSatangBig(new(big.Int).Mul(baht, hundred))
}

// BahtSatang renders a satang amount as Thai Baht text.
func (sp Speller) BahtSatang(satang int64) string {
	neg := satang < 0
	mag := new(big.Int).Abs(big.NewInt(satang))
	out := sp.bahtFromBig(mag)
	if neg {
		return "ลบ" + out
	}
	return out
}

// BahtSatangBig renders an arbitrarily large satang amount as Thai Baht text.
func (sp Speller) BahtSatangBig(satang *big.Int) string {
	if satang == nil {
		return sp.bahtParts(big.NewInt(0), 0)
	}
	neg := satang.Sign() < 0
	out := sp.bahtFromBig(new(big.Int).Abs(satang))
	if neg {
		return "ลบ" + out
	}
	return out
}

var hundred = big.NewInt(100)

// bahtFromBig renders a non-negative satang magnitude.
func (sp Speller) bahtFromBig(mag *big.Int) string {
	baht := new(big.Int)
	sat := new(big.Int)
	baht.DivMod(mag, hundred, sat)
	return sp.bahtParts(baht, int(sat.Int64()))
}

// bahtParts assembles the baht text from a non-negative baht count and a
// satang value in 0..99.
func (sp Speller) bahtParts(baht *big.Int, sat int) string {
	bahtZero := baht.Sign() == 0
	if bahtZero && sat == 0 {
		return numberThai[0] + "บาทถ้วน" // ศูนย์บาทถ้วน
	}
	var out string
	if !bahtZero {
		out = sp.Big(baht) + "บาท"
	}
	if sat == 0 {
		return out + "ถ้วน"
	}
	return out + sp.spellGroup(sat) + "สตางค์"
}

// BahtFromString renders a decimal baht amount string as Thai Baht text.
func (sp Speller) BahtFromString(amount string) (string, error) {
	neg, satang, err := parseToSatang(amount)
	if err != nil {
		return "", err
	}
	out := sp.bahtFromBig(satang)
	if neg && satang.Sign() != 0 {
		return "ลบ" + out, nil
	}
	return out, nil
}

// BahtFromFloat renders a float baht amount (lossy; see the package-level doc).
func (sp Speller) BahtFromFloat(amount float64) string {
	out, _ := sp.BahtFromString(strconv.FormatFloat(amount, 'f', 2, 64))
	return out
}

// parseToSatang converts a decimal baht string to a non-negative satang
// magnitude plus a sign, rounding the fraction to two places away from zero.
func parseToSatang(amount string) (neg bool, satang *big.Int, err error) {
	neg, intPart, frac, err := splitDecimal(amount)
	if err != nil {
		return false, nil, err
	}
	baht, ok := new(big.Int).SetString(zeroIfEmpty(intPart), 10)
	if !ok {
		return false, nil, errInvalid(amount)
	}
	sat := baht.Mul(baht, hundred)

	// Two-decimal satang with away-from-zero rounding on the third digit.
	cents := 0
	if len(frac) >= 1 {
		cents += int(frac[0]-'0') * 10
	}
	if len(frac) >= 2 {
		cents += int(frac[1] - '0')
	}
	if len(frac) >= 3 && frac[2] >= '5' {
		cents++ // round up
	}
	sat.Add(sat, big.NewInt(int64(cents)))
	return neg, sat, nil
}

func zeroIfEmpty(s string) string {
	if s == "" {
		return "0"
	}
	return s
}
