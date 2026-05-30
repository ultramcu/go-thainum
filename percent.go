package thainum

import "strings"

// PercentStyle selects how a percentage value is read aloud in Thai.
type PercentStyle int

const (
	// RoyalRoiLa is the Royal-Institute / legal form: the prefix "ร้อยละ"
	// followed by the spelled number ("ร้อยละยี่สิบห้า"). This is the form used
	// in laws, contracts and official documents. It is the default (zero value).
	RoyalRoiLa PercentStyle = iota
	// ColloquialPercent is the everyday/colloquial form: the spelled number
	// followed by the loanword "เปอร์เซ็นต์" ("ยี่สิบห้าเปอร์เซ็นต์").
	ColloquialPercent
)

const percentSuffix = "เปอร์เซ็นต์"

// Percent reads a percentage value in Thai words. The value is given as a
// decimal numeric string (e.g. "25", "25.5", "-5") so that exact decimals are
// preserved with no float rounding surprises, consistent with the rest of this
// package. A whole number (including a value such as "25.0") is spelled as an
// integer; a fractional value is spelled per-digit after "จุด".
//
//	Percent("25", RoyalRoiLa)           // "ร้อยละยี่สิบห้า"
//	Percent("25.5", RoyalRoiLa)         // "ร้อยละยี่สิบห้าจุดห้า"
//	Percent("25", ColloquialPercent)    // "ยี่สิบห้าเปอร์เซ็นต์"
//
// A malformed value returns an error.
func Percent(value string, style PercentStyle) (string, error) {
	spelled, err := SpellDecimal(value)
	if err != nil {
		return "", err
	}
	switch style {
	case ColloquialPercent:
		return spelled + percentSuffix, nil
	default: // RoyalRoiLa
		return "ร้อยละ" + spelled, nil
	}
}

// PercentInt reads an integer percentage in Thai words, the exact-int
// convenience form of Percent.
//
//	PercentInt(25, RoyalRoiLa)        // "ร้อยละยี่สิบห้า"
//	PercentInt(-5, RoyalRoiLa)        // "ร้อยละลบห้า"
//	PercentInt(25, ColloquialPercent) // "ยี่สิบห้าเปอร์เซ็นต์"
func PercentInt(value int64, style PercentStyle) string {
	spelled := Spell(value)
	switch style {
	case ColloquialPercent:
		return spelled + percentSuffix
	default: // RoyalRoiLa
		return "ร้อยละ" + spelled
	}
}

// FormatPercent gives the numeric display form of a percentage: an integer
// reads as "25%", a fractional value as a trimmed decimal + "%" with trailing
// zeros (and any dangling dot) removed. The value is a decimal numeric string.
//
//	FormatPercent("25")    // "25%"
//	FormatPercent("25.5")  // "25.5%"
//	FormatPercent("25.0")  // "25%"
//	FormatPercent("25.50") // "25.5%"
//
// A malformed value returns an error.
func FormatPercent(value string) (string, error) {
	neg, intPart, frac, err := splitDecimal(value)
	if err != nil {
		return "", err
	}
	// Trim trailing zeros from the fractional part.
	frac = strings.TrimRight(frac, "0")

	var b strings.Builder
	// Normalize the integer part: drop leading zeros but keep a single "0".
	ip := strings.TrimLeft(intPart, "0")
	if ip == "" {
		ip = "0"
	}
	// Suppress a negative sign for an effective zero (matches Dart's int read).
	isZero := ip == "0" && frac == ""
	if neg && !isZero {
		b.WriteByte('-')
	}
	b.WriteString(ip)
	if frac != "" {
		b.WriteByte('.')
		b.WriteString(frac)
	}
	b.WriteByte('%')
	return b.String(), nil
}
