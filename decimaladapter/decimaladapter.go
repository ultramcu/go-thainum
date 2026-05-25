// Package decimaladapter renders github.com/shopspring/decimal values as Thai
// words and Thai Baht text. It is a thin bridge over go-thainum for callers who
// already hold decimal.Decimal money values; importing it is the only way to
// pull shopspring/decimal into your build — the core go-thainum package stays
// dependency-free.
package decimaladapter

import (
	"github.com/shopspring/decimal"
	thainum "github.com/ultramcu/go-thainum"
)

// Baht renders a decimal baht amount as Thai Baht text, rounding to two decimal
// places (e.g. decimal "21.21" -> "ยี่สิบเอ็ดบาทยี่สิบเอ็ดสตางค์").
func Baht(d decimal.Decimal) string {
	s, _ := thainum.BahtFromString(d.StringFixed(2))
	return s
}

// Spell renders a decimal number as Thai words (integer part read normally, then
// "จุด" and each fractional digit; e.g. "12.34" -> "สิบสองจุดสามสี่").
func Spell(d decimal.Decimal) string {
	s, _ := thainum.SpellDecimal(d.String())
	return s
}
