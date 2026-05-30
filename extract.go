package thainum

import (
	"math/big"
	"unicode/utf8"
)

// NumberMatch is a single number found embedded in free text by ExtractNumbers.
type NumberMatch struct {
	// Start is the byte offset (inclusive) into the input string where the match
	// begins. End is the byte offset (exclusive) where it ends, so
	// text[Start:End] == Matched.
	//
	// NOTE: the Dart sibling reports UTF-16 code-unit offsets; this Go port uses
	// BYTE offsets into the input string.
	Start, End int
	// Matched is the exact matched substring of the input.
	Matched string
	// Value is the numeric value of the match, as a non-negative magnitude
	// (ลบ/minus is treated as a connector, never part of a match).
	Value *big.Int
	// IsWord is true when the match came from Thai number words.
	IsWord bool
	// IsDigits is true when the match came from Arabic/Thai digit characters.
	IsDigits bool
}

// matchWordAt returns the longest entry in table whose word is a prefix of
// s[pos:], or ok=false if none match. Slicing a Go string is O(1), so this
// stays allocation-free.
func matchWordAt(s string, pos int, table []wordEntry) (wordEntry, bool) {
	rest := s[pos:]
	for _, e := range table {
		if len(e.word) <= len(rest) && rest[:len(e.word)] == e.word {
			return e, true
		}
	}
	return wordEntry{}, false
}

// isDigitRune reports whether r is an ASCII digit (0-9) or a Thai digit (๐-๙).
func isDigitRune(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= thaiZero && r <= thaiNine)
}

// ExtractNumbers finds every Thai number embedded in free text, left to right.
//
//	ExtractNumbers("ซื้อมา ๓ ชิ้น ราคาห้าร้อยบาท")
//	// [ ๓ → 3 (IsDigits), ห้าร้อย → 500 (IsWord) ]
//
// At each position it consumes the maximal valid number starting there: either
// a run of digit characters (ASCII 0-9 and/or Thai ๐-๙, emitted as one
// digits-match) or a maximal run of contiguous number words that the integer
// grammar accepts as a single number.
//
// Maximal-munch rule for word runs (deterministic): starting at a number word,
// greedily append the next contiguous number-word token while the whole
// accumulated token sequence still parses as a valid integer. Emit the longest
// valid prefix as a single match, then resume scanning immediately after it. As
// a consequence, "ยี่สิบเอ็ด" yields one match (21), and "ห้าร้อยสิบสิบ" yields
// "ห้าร้อยสิบ" (510) followed by "สิบ" (10).
//
// ลบ (minus) is treated as a non-number connector and is never part of a match,
// so extracted values are non-negative magnitudes.
//
// Byte offsets, not code points, are reported in NumberMatch.Start/End.
func ExtractNumbers(text string) []NumberMatch {
	var matches []NumberMatch
	i := 0
	n := len(text)
	for i < n {
		// Decode the rune at i to classify it and to know how far to advance on
		// a non-match.
		r, size := utf8.DecodeRuneInString(text[i:])

		// 1) Digit run (ASCII and/or Thai digits).
		if isDigitRune(r) {
			j := i
			var buf []byte
			for j < n {
				dr, dsize := utf8.DecodeRuneInString(text[j:])
				if !isDigitRune(dr) {
					break
				}
				var d rune
				if dr >= thaiZero {
					d = dr - thaiZero
				} else {
					d = dr - '0'
				}
				buf = append(buf, byte('0'+d))
				j += dsize
			}
			v, _ := new(big.Int).SetString(string(buf), 10)
			matches = append(matches, NumberMatch{
				Start:    i,
				End:      j,
				Matched:  text[i:j],
				Value:    v,
				IsWord:   false,
				IsDigits: true,
			})
			i = j
			continue
		}

		// 2) Word run: try to match a number word at i. Use the strict
		// (non-colloquial) table, mirroring the Dart extractor.
		first, ok := matchWordAt(text, i, numberWords)
		if !ok || first.tok().kind == tkNeg {
			// ลบ alone is not a number start here; advance one rune.
			i += size
			continue
		}

		// Greedily collect word tokens, tracking the longest prefix that parses.
		scan := i
		bestEnd := -1 // exclusive end of the longest valid prefix
		var bestVal *big.Int
		var toks []token
		for scan < n {
			e, ok := matchWordAt(text, scan, numberWords)
			if !ok {
				break
			}
			tk := e.tok()
			if tk.kind == tkNeg {
				break // a sign cannot extend a number
			}
			toks = append(toks, tk)
			scan += len(e.word)
			// Try to evaluate the accumulated token run as a complete number.
			if v, err := evalTokens(toks, false); err == nil {
				bestEnd = scan
				bestVal = v
			}
		}

		if bestEnd > i {
			matches = append(matches, NumberMatch{
				Start:    i,
				End:      bestEnd,
				Matched:  text[i:bestEnd],
				Value:    bestVal,
				IsWord:   true,
				IsDigits: false,
			})
			i = bestEnd
		} else {
			i += size
		}
	}
	return matches
}
