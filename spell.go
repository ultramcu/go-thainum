package thainum

import (
	"fmt"
	"math/big"
	"strings"
)

// EtMode selects how a units digit 1 is read when it sits above a zero tens
// digit (e.g. 101, 1001, 1,000,001). This is the one genuine convention split
// in Thai number reading.
type EtMode int

const (
	// EtAlways reads such a 1 as "เอ็ด" (101 -> หนึ่งร้อยเอ็ด). This is the
	// Royal Institute's recommended form and the default (zero value).
	EtAlways EtMode = iota
	// EtTensOnly reads such a 1 as "หนึ่ง" (101 -> หนึ่งร้อยหนึ่ง), using "เอ็ด"
	// only when the tens digit is non-zero (11, 21, …).
	EtTensOnly
)

var numberThai = [10]string{"ศูนย์", "หนึ่ง", "สอง", "สาม", "สี่", "ห้า", "หก", "เจ็ด", "แปด", "เก้า"}

// placeWord[p] is the Thai word for position p within a six-digit group:
// 0=units, 1=tens, 2=hundreds, 3=thousands, 4=ten-thousands, 5=hundred-thousands.
var placeWord = [6]string{"", "สิบ", "ร้อย", "พัน", "หมื่น", "แสน"}

// Speller spells numbers as Thai words with a chosen EtMode.
type Speller struct {
	Et EtMode
}

// Spell renders the integer n as Thai words using the default EtMode (EtAlways).
//
//	Spell(21)       // "ยี่สิบเอ็ด"
//	Spell(1000000)  // "หนึ่งล้าน"
//	Spell(-5)       // "ลบห้า"
func Spell(n int64) string { return Speller{}.Int(n) }

// SpellBig renders an arbitrarily large integer as Thai words (default EtMode).
func SpellBig(n *big.Int) string { return Speller{}.Big(n) }

// SpellDecimal renders a decimal number written as a string (e.g. "12.34")
// as Thai words: the integer part is read normally, then "จุด", then each
// fractional digit individually (12.34 -> "สิบสองจุดสามสี่"). A fractional part
// that is empty or all zeros is dropped (100.00 -> "หนึ่งร้อย"). Uses the
// default EtMode.
func SpellDecimal(s string) (string, error) { return Speller{}.Decimal(s) }

// Int renders the integer n as Thai words.
func (sp Speller) Int(n int64) string {
	if n == 0 {
		return numberThai[0]
	}
	neg := n < 0
	// Use big to format the magnitude safely (handles math.MinInt64).
	mag := new(big.Int).Abs(big.NewInt(n))
	out := sp.spellDigits(mag.String())
	if neg {
		return "ลบ" + out
	}
	return out
}

// Big renders an arbitrarily large integer as Thai words.
func (sp Speller) Big(n *big.Int) string {
	if n == nil || n.Sign() == 0 {
		return numberThai[0]
	}
	mag := new(big.Int).Abs(n)
	out := sp.spellDigits(mag.String())
	if n.Sign() < 0 {
		return "ลบ" + out
	}
	return out
}

// Decimal renders a decimal numeric string as Thai words.
func (sp Speller) Decimal(s string) (string, error) {
	neg, intPart, frac, err := splitDecimal(s)
	if err != nil {
		return "", err
	}
	var out string
	if intPart == "" || isAllZero(intPart) {
		out = numberThai[0] // ศูนย์
	} else {
		out = sp.spellDigits(intPart)
	}
	if frac != "" && !isAllZero(frac) {
		var b strings.Builder
		b.WriteString(out)
		b.WriteString("จุด")
		for _, r := range frac {
			b.WriteString(numberThai[r-'0'])
		}
		out = b.String()
	}
	if neg {
		out = "ลบ" + out
	}
	return out, nil
}

// spellDigits spells a non-empty, non-negative integer digit string (no sign,
// leading zeros tolerated) as Thai words.
func (sp Speller) spellDigits(s string) string {
	s = strings.TrimLeft(s, "0")
	if s == "" {
		return numberThai[0]
	}
	n := len(s)
	numGroups := (n + 5) / 6

	var b strings.Builder
	higherSeen := false
	for gi := numGroups - 1; gi >= 0; gi-- {
		end := n - gi*6
		start := end - 6
		if start < 0 {
			start = 0
		}
		grp := s[start:end]
		gv := atoiSmall(grp)
		if gv == 0 {
			continue
		}
		if gi == 0 && gv == 1 && higherSeen {
			// Trailing lone 1 after higher groups: cross-group เอ็ด rule.
			// Tens is zero here, so EtTensOnly yields หนึ่ง.
			if sp.Et == EtAlways {
				b.WriteString("เอ็ด")
			} else {
				b.WriteString(numberThai[1])
			}
		} else {
			b.WriteString(sp.spellGroup(gv))
		}
		for r := 0; r < gi; r++ {
			b.WriteString("ล้าน")
		}
		higherSeen = true
	}
	return b.String()
}

// spellGroup spells a value 1..999999 (a single six-digit group) as Thai words.
func (sp Speller) spellGroup(n int) string {
	var b strings.Builder
	tens := (n / 10) % 10
	for p := 5; p >= 0; p-- {
		d := (n / pow10[p]) % 10
		if d == 0 {
			continue
		}
		switch p {
		case 1: // tens
			switch d {
			case 1:
				b.WriteString("สิบ")
			case 2:
				b.WriteString("ยี่สิบ")
			default:
				b.WriteString(numberThai[d] + "สิบ")
			}
		case 0: // units
			if d == 1 && n >= 10 {
				if sp.Et == EtAlways || tens != 0 {
					b.WriteString("เอ็ด")
				} else {
					b.WriteString(numberThai[1])
				}
			} else {
				b.WriteString(numberThai[d])
			}
		default: // hundreds..hundred-thousands
			b.WriteString(numberThai[d] + placeWord[p])
		}
	}
	return b.String()
}

var pow10 = [6]int{1, 10, 100, 1000, 10000, 100000}

// atoiSmall parses a short digit string (<=6 chars) into an int.
func atoiSmall(s string) int {
	v := 0
	for i := 0; i < len(s); i++ {
		v = v*10 + int(s[i]-'0')
	}
	return v
}

func isAllZero(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] != '0' {
			return false
		}
	}
	return true
}

// splitDecimal parses an optionally-signed decimal numeric string into its sign,
// integer-digit part and fractional-digit part. Surrounding spaces, a leading
// '+'/'-', and Thai digits are accepted.
func splitDecimal(s string) (neg bool, intPart, frac string, err error) {
	s = strings.TrimSpace(ToArabicDigits(s))
	if s == "" {
		return false, "", "", fmt.Errorf("thainum: empty number")
	}
	switch s[0] {
	case '-':
		neg, s = true, s[1:]
	case '+':
		s = s[1:]
	}
	intPart = s
	if i := strings.IndexByte(s, '.'); i >= 0 {
		intPart, frac = s[:i], s[i+1:]
	}
	if intPart == "" && frac == "" {
		return false, "", "", fmt.Errorf("thainum: invalid number %q", s)
	}
	if !isDigits(intPart) || !isDigits(frac) {
		return false, "", "", fmt.Errorf("thainum: invalid number %q", s)
	}
	return neg, intPart, frac, nil
}

func isDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
