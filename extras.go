package thainum

// Ordinal renders n as a Thai ordinal ("ที่" + the number).
//
//	Ordinal(1)  // "ที่หนึ่ง"
//	Ordinal(21) // "ที่ยี่สิบเอ็ด"
func Ordinal(n int64) string { return Speller{}.Ordinal(n) }

// Fraction renders num/den as a Thai fraction ("เศษ" num "ส่วน" den).
// den is expected to be non-zero.
//
//	Fraction(1, 2) // "เศษหนึ่งส่วนสอง"
//	Fraction(3, 4) // "เศษสามส่วนสี่"
func Fraction(num, den int64) string { return Speller{}.Fraction(num, den) }

// Year renders a Buddhist-Era year as Thai words prefixed with "พุทธศักราช".
//
//	Year(2566) // "พุทธศักราชสองพันห้าร้อยหกสิบหก"
func Year(be int64) string { return Speller{}.Year(be) }

// CEToBE converts a Common-Era (Gregorian) year to the Buddhist Era (+543).
func CEToBE(ce int64) int64 { return ce + 543 }

// BEToCE converts a Buddhist-Era year to the Common Era (−543).
func BEToCE(be int64) int64 { return be - 543 }

// Ordinal renders n as a Thai ordinal.
func (sp Speller) Ordinal(n int64) string { return "ที่" + sp.Int(n) }

// Fraction renders num/den as a Thai fraction.
func (sp Speller) Fraction(num, den int64) string {
	return "เศษ" + sp.Int(num) + "ส่วน" + sp.Int(den)
}

// Year renders a Buddhist-Era year as "พุทธศักราช" + the spelled number.
func (sp Speller) Year(be int64) string { return "พุทธศักราช" + sp.Int(be) }
