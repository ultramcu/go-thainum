package thainum

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
)

// ErrParse is the sentinel error returned (wrapped, with the offending token as
// context) when Thai number words cannot be parsed. Test for it with
// errors.Is(err, ErrParse).
var ErrParse = errors.New("thainum: cannot parse")

// token kinds produced by the tokenizer.
type tokKind int

const (
	tkDigit   tokKind = iota // a units digit 1..9 (ศูนย์..เก้า, ยี่, เอ็ด)
	tkZero                   // ศูนย์
	tkPlace                  // สิบ ร้อย พัน หมื่น แสน
	tkMillion                // ล้าน
	tkNeg                    // ลบ
)

type token struct {
	kind tokKind
	text string // the matched source word (for error context)
	val  int64  // digit value (tkDigit) or place power-of-ten (tkPlace)
	yi   bool   // true if this digit came from ยี่ (only valid before สิบ)
	et   bool   // true if this digit came from เอ็ด (a trailing one)
}

// wordEntry describes one recognizable Thai number word. The tokenizer matches
// longest-first, so longer words shadow any shorter word that is a prefix.
type wordEntry struct {
	word string
	tok  func() token
}

// numberWords is the dictionary of recognizable words, kept sorted by
// descending rune/byte length so a greedy longest-match tokenizer is correct.
var numberWords = buildWordTable()

func buildWordTable() []wordEntry {
	mk := func(w string, k tokKind, v int64) wordEntry {
		return wordEntry{word: w, tok: func() token { return token{kind: k, text: w, val: v} }}
	}
	t := []wordEntry{
		// sign
		{word: "ลบ", tok: func() token { return token{kind: tkNeg, text: "ลบ"} }},
		// special digits
		{word: "ยี่", tok: func() token { return token{kind: tkDigit, text: "ยี่", val: 2, yi: true} }},
		{word: "เอ็ด", tok: func() token { return token{kind: tkDigit, text: "เอ็ด", val: 1, et: true} }},
		// place / multiplier words
		mk("ล้าน", tkMillion, 1000000),
		mk("แสน", tkPlace, 100000),
		mk("หมื่น", tkPlace, 10000),
		mk("พัน", tkPlace, 1000),
		mk("ร้อย", tkPlace, 100),
		mk("สิบ", tkPlace, 10),
	}
	// digit words 0..9
	for d := 0; d < 10; d++ {
		dv := int64(d)
		w := numberThai[d]
		if d == 0 {
			t = append(t, wordEntry{word: w, tok: func() token { return token{kind: tkZero, text: w, val: 0} }})
		} else {
			t = append(t, wordEntry{word: w, tok: func() token { return token{kind: tkDigit, text: w, val: dv} }})
		}
	}
	// sort by descending word length for greedy longest-match
	for i := 0; i < len(t); i++ {
		for j := i + 1; j < len(t); j++ {
			if len(t[j].word) > len(t[i].word) {
				t[i], t[j] = t[j], t[i]
			}
		}
	}
	return t
}

// ParseInt parses Thai number words into an int64.
//
//	ParseInt("ยี่สิบเอ็ด") // 21
//	ParseInt("หนึ่งล้าน")  // 1000000
//	ParseInt("ลบห้า")      // -5
//	ParseInt("ศูนย์")      // 0
//
// Thai or Arabic digit characters are also accepted (ParseInt("๒๑") == 21).
// Overflow of int64 returns an error suggesting ParseBig.
//
// Trailing ParseOption values (Strict, Lenient, AllowColloquial) tune the
// behaviour; with no options the result is identical to the historic parser.
func ParseInt(words string, opts ...ParseOption) (int64, error) {
	b, err := ParseBig(words, opts...)
	if err != nil {
		return 0, err
	}
	if !b.IsInt64() {
		return 0, fmt.Errorf("%w: value %s overflows int64 (use ParseBig)", ErrParse, b.String())
	}
	return b.Int64(), nil
}

// ParseBig parses Thai number words into an arbitrary-precision integer.
//
//	ParseBig("หนึ่งล้านล้าน")      // 10^12
//	ParseBig("หนึ่งล้านล้านล้าน")  // 10^18
//
// See ParseInt for the Strict, Lenient and AllowColloquial options.
func ParseBig(words string, opts ...ParseOption) (*big.Int, error) {
	o := applyOptions(opts)
	s := strings.TrimSpace(ToArabicDigits(words))
	if o.lenient {
		s = normalizeLenient(s)
	}
	if s == "" {
		return nil, fmt.Errorf("%w: empty input", ErrParse)
	}

	// Allow a plain numeric string (already converted from Thai digits above):
	// an optional sign followed by ASCII digits.
	if v, ok := parsePlainDigits(s); ok {
		return v, nil
	}

	neg := false
	if strings.HasPrefix(s, "ลบ") {
		neg = true
		s = s[len("ลบ"):]
		if s == "" {
			return nil, fmt.Errorf("%w: %q is not a number", ErrParse, "ลบ")
		}
	}

	toks, err := tokenize(s, wordTable(o.allowColloquial))
	if err != nil {
		return nil, err
	}
	v, err := evalTokens(toks, o.strict)
	if err != nil {
		return nil, err
	}
	if neg {
		v.Neg(v)
	}
	return v, nil
}

// parsePlainDigits handles a purely-numeric input (optionally signed). It
// returns ok=false when the string is not purely numeric so the word parser can
// take over.
func parsePlainDigits(s string) (*big.Int, bool) {
	t := s
	switch {
	case strings.HasPrefix(t, "+"):
		t = t[1:]
	case strings.HasPrefix(t, "-"):
		t = t[1:]
	}
	if t == "" || !isDigits(t) {
		return nil, false
	}
	v, ok := new(big.Int).SetString(s, 10)
	return v, ok
}

// tokenize splits a sign-free Thai number string into recognized tokens using a
// greedy longest-match scan. Any unrecognized run yields an ErrParse with the
// offending substring as context.
func tokenize(s string, table []wordEntry) ([]token, error) {
	var toks []token
	for len(s) > 0 {
		matched := false
		for _, e := range table {
			if strings.HasPrefix(s, e.word) {
				toks = append(toks, e.tok())
				s = s[len(e.word):]
				matched = true
				break
			}
		}
		if !matched {
			// Report the next rune as the unknown token for context.
			r := []rune(s)
			bad := string(r[0])
			if len(r) > 1 {
				bad = string(r[:min(len(r), 4)])
			}
			return nil, fmt.Errorf("%w: unknown token %q", ErrParse, bad)
		}
	}
	return toks, nil
}

// evalTokens runs the accumulator state machine over the tokens.
//
// Thai writes large numbers in six-digit groups separated by ล้าน (10^6), and
// the higher groups stack the word — ล้านล้าน is 10^12, ล้านล้านล้าน is 10^18,
// and so on. So a group followed by k consecutive ล้าน contributes
// group * 10^(6k). We build each group's value in current, and when a run of
// ล้าน is seen we add current * 10^(6k) to result. The million-run lengths must
// strictly decrease from one group to the next (the natural, unambiguous
// ordering), otherwise the input is malformed (e.g. "สองล้านสามล้านล้าน").
//
// Within a group, pending holds a units digit (1..9) seen but not yet attached
// to a place; a place word multiplies the pending digit (default 1) into current
// at that place. At the very end the final (scale 10^0) group is added.
func evalTokens(toks []token, strict bool) (*big.Int, error) {
	result := big.NewInt(0)
	current := big.NewInt(0)

	var pending int64 = -1 // -1 means "no pending digit"
	pendingYi := false     // pending digit came from ยี่

	tmp := new(big.Int)
	scale := new(big.Int)
	million := big.NewInt(1000000)

	// maxPlaceInGroup rejects repeated/ascending places within a group
	// ("สิบสิบ", "สิบร้อย", "ร้อยร้อย").
	var maxPlaceInGroup int64 // highest place value seen since the last flush
	curHasContent := false    // current group has received any value

	// lastMillionPower is the ล้าน-run length of the previously flushed group;
	// runs must strictly decrease. -1 means no group has been flushed yet.
	lastMillionPower := -1

	flushPendingAsUnits := func() error {
		if pending < 0 {
			return nil
		}
		if pendingYi {
			// ยี่ is only valid immediately before สิบ.
			return fmt.Errorf("%w: %q must be followed by สิบ", ErrParse, "ยี่")
		}
		// A trailing units digit lands in the ones place.
		current.Add(current, big.NewInt(pending))
		curHasContent = true
		pending = -1
		return nil
	}

	for i := 0; i < len(toks); i++ {
		tk := toks[i]
		switch tk.kind {
		case tkNeg:
			return nil, fmt.Errorf("%w: misplaced %q", ErrParse, tk.text)

		case tkZero:
			// ศูนย์ is only meaningful as the entire value (handled by the
			// plain/zero shortcut). Inside a compound it is malformed.
			if len(toks) == 1 {
				return big.NewInt(0), nil
			}
			return nil, fmt.Errorf("%w: misplaced %q", ErrParse, tk.text)

		case tkDigit:
			if pending >= 0 {
				return nil, fmt.Errorf("%w: two digits in a row near %q", ErrParse, tk.text)
			}
			pending = tk.val
			pendingYi = tk.yi
			if tk.et {
				// เอ็ด is a trailing one; it must terminate a group (it cannot
				// take a place word). Add it now and require it be last or
				// followed only by ล้าน.
				if i+1 < len(toks) {
					nxt := toks[i+1]
					if nxt.kind == tkPlace {
						return nil, fmt.Errorf("%w: %q cannot precede a place word", ErrParse, tk.text)
					}
				}
				current.Add(current, big.NewInt(1))
				curHasContent = true
				pending = -1
				pendingYi = false
			}

		case tkPlace:
			place := tk.val
			digit := int64(1)
			isYi := false
			hadPending := false
			if pending >= 0 {
				digit = pending
				isYi = pendingYi
				hadPending = true
				pending = -1
				pendingYi = false
			}
			// ยี่ is only valid directly before สิบ (ยี่สิบ = 20).
			if isYi && place != 10 {
				return nil, fmt.Errorf("%w: %q must be followed by สิบ", ErrParse, "ยี่")
			}
			// Strict mode rejects a non-standard tens form: หนึ่งสิบ (digit-1 over
			// สิบ) or สองสิบ (a plain สอง instead of ยี่). A bare สิบ (no pending
			// digit, digit defaults to 1) is the canonical form and is allowed.
			if strict && place == 10 && hadPending && (digit == 1 || (digit == 2 && !isYi)) {
				return nil, fmt.Errorf("%w: non-standard tens form %q", ErrParse, tk.text)
			}
			// Enforce strictly descending places within a group so that
			// repeats/ascents like "สิบสิบ" or "สิบร้อย" are rejected.
			if curHasContent && place >= maxPlaceInGroup {
				return nil, fmt.Errorf("%w: misplaced place word %q", ErrParse, tk.text)
			}
			tmp.Mul(big.NewInt(digit), big.NewInt(place))
			current.Add(current, tmp)
			curHasContent = true
			maxPlaceInGroup = place

		case tkMillion:
			// Flush any pending units digit into the current group first.
			if err := flushPendingAsUnits(); err != nil {
				return nil, err
			}
			if !curHasContent {
				// A bare ล้าน (no digits before it) reads as 1,000,000, e.g.
				// "ล้าน" or "ล้านล้าน" -> 10^6, 10^12.
				current.SetInt64(1)
				curHasContent = true
			}
			// Count this and any directly-following ล้าน tokens as one run.
			power := 1
			for i+1 < len(toks) && toks[i+1].kind == tkMillion {
				power++
				i++
			}
			if lastMillionPower >= 0 && power >= lastMillionPower {
				return nil, fmt.Errorf("%w: ล้าน groups out of order", ErrParse)
			}
			lastMillionPower = power
			// result += current * 10^(6*power)
			scale.Exp(million, big.NewInt(int64(power)), nil)
			tmp.Mul(current, scale)
			result.Add(result, tmp)
			current.SetInt64(0)
			curHasContent = false
			maxPlaceInGroup = 0
		}
	}

	if err := flushPendingAsUnits(); err != nil {
		return nil, err
	}
	result.Add(result, current)
	return result, nil
}

// ParseBaht parses Thai Baht text into an integer satang amount.
//
//	ParseBaht("ยี่สิบเอ็ดบาทยี่สิบเอ็ดสตางค์") // 2121
//	ParseBaht("หนึ่งร้อยบาทถ้วน")             // 10000
//	ParseBaht("ศูนย์บาทถ้วน")                 // 0
//	ParseBaht("ยี่สิบห้าสตางค์")              // 25
//	ParseBaht("ลบหนึ่งบาทหนึ่งสตางค์")        // -101
//
// See ParseInt for the Strict, Lenient and AllowColloquial options.
func ParseBaht(text string, opts ...ParseOption) (satang int64, err error) {
	o := applyOptions(opts)
	s := strings.TrimSpace(ToArabicDigits(text))
	// lenient normalizes the whole string here (before the บาท/สตางค์ split), so
	// the baht/satang parts are already space-free — the inner ParseInt calls
	// below only need to forward allowColloquial and strict, not lenient.
	if o.lenient {
		s = normalizeLenient(s)
	}
	if s == "" {
		return 0, fmt.Errorf("%w: empty input", ErrParse)
	}

	neg := false
	if strings.HasPrefix(s, "ลบ") {
		neg = true
		s = s[len("ลบ"):]
	}

	// Drop a trailing ถ้วน ("exactly", no satang).
	s = strings.TrimSuffix(s, "ถ้วน")

	bahtPart := ""
	satPart := ""

	if idx := strings.Index(s, "บาท"); idx >= 0 {
		bahtPart = s[:idx]
		rest := s[idx+len("บาท"):]
		satPart = strings.TrimSuffix(rest, "สตางค์")
	} else if strings.HasSuffix(s, "สตางค์") {
		// Satang-only amount, e.g. "ยี่สิบห้าสตางค์".
		satPart = strings.TrimSuffix(s, "สตางค์")
	} else {
		// No บาท and no สตางค์: only valid as a bare ศูนย์ (e.g. "ศูนย์" after
		// trimming ถ้วน from "ศูนย์บาทถ้วน" would still contain บาท, so this
		// path means a genuinely odd input).
		bahtPart = s
	}

	innerOpts := []ParseOption{}
	if o.allowColloquial {
		innerOpts = append(innerOpts, AllowColloquial())
	}
	if o.strict {
		innerOpts = append(innerOpts, Strict())
	}

	var bahtVal int64
	if strings.TrimSpace(bahtPart) != "" {
		bv, err := ParseInt(bahtPart, innerOpts...)
		if err != nil {
			return 0, fmt.Errorf("%w (baht part)", err)
		}
		bahtVal = bv
	}

	var satVal int64
	if strings.TrimSpace(satPart) != "" {
		sv, err := ParseInt(satPart, innerOpts...)
		if err != nil {
			return 0, fmt.Errorf("%w (satang part)", err)
		}
		if sv < 0 || sv > 99 {
			return 0, fmt.Errorf("%w: satang %d out of range 0..99", ErrParse, sv)
		}
		satVal = sv
	}

	total := bahtVal*100 + satVal
	if neg {
		total = -total
	}
	return total, nil
}
