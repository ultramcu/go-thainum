package thainum

// IsDigits reports whether s is non-empty and consists solely of ASCII digits
// 0-9.
//
//	IsDigits("123")  // true
//	IsDigits("")     // false
//	IsDigits("12a")  // false
//	IsDigits("๑๒๓")  // false (Thai digits; convert with ToArabicDigits first)
//
// Mirrors the Dart sibling's isDigits. Note that the lower-level, package-
// internal isDigits used by the parsers treats the empty string as vacuously
// all-digits; IsDigits is the public, non-empty-required form.
func IsDigits(s string) bool {
	if s == "" {
		return false
	}
	return isDigits(s)
}
