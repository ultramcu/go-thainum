package thainum

import "strings"

// parseOptions holds the (unexported) tunable behaviour for the parse family.
// The zero value is the default, round-trip-safe behaviour, byte-identical to
// the historic no-option parsers.
type parseOptions struct {
	// allowColloquial accepts spoken contractions such as นึง (= 1).
	allowColloquial bool
	// lenient strips interior ASCII spaces, NBSP (U+00A0) and zero-width
	// spaces (U+200B) before parsing.
	lenient bool
	// strict rejects non-canonical tens forms (สองสิบ, หนึ่งสิบ) that the
	// default reader otherwise tolerates.
	strict bool
}

// ParseOption configures the parse family (ParseInt, ParseBig, ParseBaht,
// ParseDecimal). Pass any combination as trailing arguments; with no options
// the behaviour is identical to the historic parsers.
//
// Mirrors the named parameters (allowColloquial, lenient, strict) of the Dart
// sibling's parse functions.
type ParseOption func(*parseOptions)

// Strict rejects non-standard tens forms that the default (lenient) reader
// tolerates: a tens place fed by หนึ่ง (หนึ่งสิบ = 10) or by a plain สอง written
// instead of ยี่ (สองสิบ = 20). Under Strict these return an error; the
// canonical forms สิบ, ยี่สิบ, ยี่สิบเอ็ด still parse.
//
//	ParseInt("สองสิบ", Strict()) // error (use ยี่สิบ)
func Strict() ParseOption {
	return func(o *parseOptions) { o.strict = true }
}

// Lenient strips interior ASCII spaces, NBSP (U+00A0) and zero-width spaces
// (U+200B) before parsing, so spaced-out input still tokenizes.
//
//	ParseInt("ยี่สิบ เอ็ด", Lenient()) // 21
func Lenient() ParseOption {
	return func(o *parseOptions) { o.lenient = true }
}

// AllowColloquial also accepts นึง (the spoken contraction of หนึ่ง) as 1, so
// ร้อยนึง parses as 101.
//
//	ParseInt("นึง", AllowColloquial()) // 1
func AllowColloquial() ParseOption {
	return func(o *parseOptions) { o.allowColloquial = true }
}

// applyOptions folds the variadic options into a parseOptions value.
func applyOptions(opts []ParseOption) parseOptions {
	var o parseOptions
	for _, fn := range opts {
		if fn != nil {
			fn(&o)
		}
	}
	return o
}

const (
	// nbsp is the non-breaking space U+00A0.
	nbsp = " "
	// zwsp is the zero-width space U+200B.
	zwsp = "​"
)

// normalizeLenient removes interior ASCII spaces, NBSP and zero-width spaces
// from s. Used by the lenient parse path so spaced-out input still tokenizes.
func normalizeLenient(s string) string {
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, nbsp, "")
	s = strings.ReplaceAll(s, zwsp, "")
	return s
}

// numberWordsColloquial is the strict dictionary extended with colloquial words
// (currently นึง for 1), kept sorted longest-first like numberWords.
var numberWordsColloquial = buildWordTableColloquial()

func buildWordTableColloquial() []wordEntry {
	t := buildWordTable()
	// นึง is the spoken contraction of หนึ่ง (one). Read like a plain one so
	// ร้อยนึง parses as 101.
	t = append(t, wordEntry{word: "นึง", tok: func() token { return token{kind: tkDigit, text: "นึง", val: 1} }})
	// re-sort by descending word length for greedy longest-match
	for i := 0; i < len(t); i++ {
		for j := i + 1; j < len(t); j++ {
			if len(t[j].word) > len(t[i].word) {
				t[i], t[j] = t[j], t[i]
			}
		}
	}
	return t
}

// wordTable returns the dictionary to use for the given allowColloquial flag.
func wordTable(allowColloquial bool) []wordEntry {
	if allowColloquial {
		return numberWordsColloquial
	}
	return numberWords
}
