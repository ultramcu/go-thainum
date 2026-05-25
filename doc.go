// Package thainum is a toolkit for Thai numbers: it converts between Arabic and
// Thai digits, spells numbers out as Thai words, renders Thai Baht text
// (บาทไทย), formats money, formats Thai dates and times (Buddhist-Era year with
// Thai month and weekday names; formal and colloquial clock; durations), parses
// Thai dates back into time.Time, and parses Thai number words back into integers.
//
// Everything is pure Go with no dependencies (the optional decimaladapter
// subpackage adds shopspring/decimal for callers who want it).
//
//	thainum.ToThaiDigits("2566")   // "๒๕๖๖"
//	thainum.Spell(21)              // "ยี่สิบเอ็ด"
//	thainum.Baht(100)              // 100 baht -> "หนึ่งร้อยบาทถ้วน"
//	thainum.BahtSatang(2121)       // 2121 satang -> "ยี่สิบเอ็ดบาทยี่สิบเอ็ดสตางค์"
//	thainum.FormatTHB(2121)        // "฿21.21"
//	thainum.ParseInt("ยี่สิบเอ็ด") // 21, nil
//
// Money values are handled in integer satang (1 baht = 100 satang) or via
// math/big, never float64, so there are no rounding surprises. A documented
// float entry point is provided for convenience.
//
// Thai reading rules implemented: สิบ (ten) and ยี่สิบ (twenty); เอ็ด for a
// trailing 1; ล้าน scaling that repeats for ล้านล้าน and beyond; ศูนย์ for zero;
// ลบ for negatives; and per-digit reading after the decimal point (จุด). The one
// genuine convention split — whether a units 1 above a zero tens reads เอ็ด
// (e.g. ร้อยเอ็ด) or หนึ่ง (ร้อยหนึ่ง) — is selectable via EtMode; the default
// (EtAlways) follows the Royal Institute's recommendation.
package thainum
