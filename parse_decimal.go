package thainum

import (
	"fmt"
	"strings"
)

// fracDigitWords maps each single-digit word (ศูนย์..เก้า) to its value, for
// reading a fractional part digit-by-digit. Kept sorted longest-first so a
// greedy match prefers the longer of any prefix-overlapping words.
var fracDigitWords = buildFracDigitWords()

type fracWord struct {
	word string
	val  int
}

func buildFracDigitWords() []fracWord {
	t := make([]fracWord, 0, 10)
	for d := 0; d < 10; d++ {
		t = append(t, fracWord{word: numberThai[d], val: d})
	}
	for i := 0; i < len(t); i++ {
		for j := i + 1; j < len(t); j++ {
			if len(t[j].word) > len(t[i].word) {
				t[i], t[j] = t[j], t[i]
			}
		}
	}
	return t
}

// ParseDecimal parses Thai decimal words into a canonical decimal string — the
// inverse of SpellDecimal.
//
//	ParseDecimal("สิบสองจุดสามสี่") // "12.34"
//	ParseDecimal("ศูนย์จุดห้า")      // "0.5"
//	ParseDecimal("ลบสามจุดหนึ่งสี่") // "-3.14"
//	ParseDecimal("ยี่สิบเอ็ด")       // "21"  (no จุด → integer string)
//
// The text is split on จุด: the integer part is read with the integer parser;
// the fractional part is a run of single-digit words (ศูนย์..เก้า), each read
// individually and concatenated as digits. A leading ลบ makes the result
// negative. The return value is a canonical decimal string (so precision and
// leading/trailing zeros from the input are preserved exactly).
//
// It returns an error (wrapping ErrParse) if the fractional part contains
// anything that is not a bare single-digit word, or if there is more than one
// จุด.
//
// See ParseInt for the Strict, Lenient and AllowColloquial options (they apply
// to the integer part).
func ParseDecimal(words string, opts ...ParseOption) (string, error) {
	o := applyOptions(opts)
	s := strings.TrimSpace(ToArabicDigits(words))
	if o.lenient {
		s = normalizeLenient(s)
	}
	if s == "" {
		return "", fmt.Errorf("%w: empty input", ErrParse)
	}

	neg := false
	if strings.HasPrefix(s, "ลบ") {
		neg = true
		s = s[len("ลบ"):]
		if s == "" {
			return "", fmt.Errorf("%w: %q is not a number", ErrParse, "ลบ")
		}
	}

	// Forward the colloquial/strict options to the integer-part parser. (lenient
	// was already applied to the whole string above.)
	intOpts := []ParseOption{}
	if o.allowColloquial {
		intOpts = append(intOpts, AllowColloquial())
	}
	if o.strict {
		intOpts = append(intOpts, Strict())
	}

	dotIdx := strings.Index(s, "จุด")
	if dotIdx < 0 {
		// No fractional part: behave like the integer parser.
		v, err := ParseBig(s, intOpts...)
		if err != nil {
			return "", err
		}
		if neg {
			v.Neg(v)
		}
		return v.String(), nil
	}

	intStr := s[:dotIdx]
	fracStr := s[dotIdx+len("จุด"):]
	if strings.Contains(fracStr, "จุด") {
		return "", fmt.Errorf("%w: more than one %q", ErrParse, "จุด")
	}
	if fracStr == "" {
		return "", fmt.Errorf("%w: missing fractional part after %q", ErrParse, "จุด")
	}

	// The integer part may not be empty before จุด.
	if intStr == "" {
		return "", fmt.Errorf("%w: missing integer part before %q", ErrParse, "จุด")
	}
	intVal, err := ParseBig(intStr, intOpts...)
	if err != nil {
		return "", err
	}

	// Read each fractional digit word individually (greedy longest-match).
	var frac strings.Builder
	for fracStr != "" {
		matched := false
		for _, w := range fracDigitWords {
			if strings.HasPrefix(fracStr, w.word) {
				frac.WriteByte(byte('0' + w.val))
				fracStr = fracStr[len(w.word):]
				matched = true
				break
			}
		}
		if !matched {
			r := []rune(fracStr)
			bad := string(r[:min(len(r), 4)])
			return "", fmt.Errorf("%w: invalid fractional digit word %q", ErrParse, bad)
		}
	}

	sign := ""
	if neg {
		sign = "-"
	}
	return sign + intVal.String() + "." + frac.String(), nil
}
