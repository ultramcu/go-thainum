package thainum

import (
	"errors"
	"fmt"
)

// ErrNotRoundMagnitude is the sentinel error returned by ThaiMoreThan when its
// argument is not a positive round magnitude (10, 20, …, 100, 1000, …). It is
// returned wrapped, with the offending value, so callers can test it with
// errors.Is(err, ErrNotRoundMagnitude).
var ErrNotRoundMagnitude = errors.New("thainum: not a round magnitude")

// QualifierKind selects which Thai qualifier word ToThaiQualified applies. It
// mirrors the qualifier words produced by ThaiApprox, ThaiNearly and
// ThaiMoreThan.
type QualifierKind int

const (
	// QualifierApprox prefixes "ประมาณ" — approximately (see ThaiApprox).
	QualifierApprox QualifierKind = iota
	// QualifierNearly prefixes "เกือบ" — nearly, just short of (see ThaiNearly).
	QualifierNearly
	// QualifierMoreThan appends "กว่า" — round-magnitude "-something" (see
	// ThaiMoreThan); only valid on a round magnitude (see IsRoundMagnitude).
	QualifierMoreThan
)

// ThaiApprox prefixes the spelled number with "ประมาณ" ("approximately"):
//
//	ThaiApprox(100) // "ประมาณหนึ่งร้อย"
//	ThaiApprox(21)  // "ประมาณยี่สิบเอ็ด"
//	ThaiApprox(-5)  // "ประมาณลบห้า"
//
// Works for any integer (the underlying Spell handles sign and zero).
func ThaiApprox(n int64) string { return "ประมาณ" + Spell(n) }

// ThaiNearly prefixes the spelled number with "เกือบ" ("nearly", just short of):
//
//	ThaiNearly(100) // "เกือบหนึ่งร้อย"
//	ThaiNearly(20)  // "เกือบยี่สิบ"
//
// Works for any integer.
func ThaiNearly(n int64) string { return "เกือบ" + Spell(n) }

// ThaiRange reads an inclusive range a to b as Spell(a) + "ถึง" + Spell(b):
//
//	ThaiRange(10, 20) // "สิบถึงยี่สิบ"
//	ThaiRange(1, 5)   // "หนึ่งถึงห้า"
//
// The numbers are read exactly as Spell renders them; a and b may be in any
// order and may be negative (ThaiRange(-5, 5) -> "ลบห้าถึงห้า"). No ordering or
// equality constraint is imposed — the words are simply joined by "ถึง".
func ThaiRange(a, b int64) string { return Spell(a) + "ถึง" + Spell(b) }

// ThaiMoreThan reads n as "…กว่า" — "n-something", a value a little more than a
// round magnitude:
//
//	ThaiMoreThan(10)      // "สิบกว่า"      (ten-something, i.e. 10–19)
//	ThaiMoreThan(20)      // "ยี่สิบกว่า"   (twenty-something)
//	ThaiMoreThan(100)     // "หนึ่งร้อยกว่า"
//	ThaiMoreThan(500)     // "ห้าร้อยกว่า"
//	ThaiMoreThan(1000000) // "หนึ่งล้านกว่า"
//
// Linguistic limit: in Thai, กว่า attaches only to a round magnitude — a number
// whose decimal form is a single non-zero leading digit followed by zeros (10,
// 20, …, 90, 100, 200, …, 1000, 1000000, and so on). สิบกว่า means
// "ten-something" (10–19); there is no well-formed สิบเอ็ดกว่า. Applying กว่า to
// an arbitrary non-round number is not idiomatic, so a value that is not a
// positive round magnitude (at least 10) returns an error wrapping
// ErrNotRoundMagnitude. On error the returned string is empty.
//
//	ThaiMoreThan(11)  // "", error (11 is not a round magnitude)
//	ThaiMoreThan(150) // "", error
//	ThaiMoreThan(5)   // "", error (single-digit, below 10)
//	ThaiMoreThan(0)   // "", error
//	ThaiMoreThan(-10) // "", error (must be positive)
func ThaiMoreThan(n int64) (string, error) {
	if !IsRoundMagnitude(n) {
		return "", fmt.Errorf(
			"%w: thaiMoreThan expects a positive round magnitude "+
				"(10, 20, …, 100, 1000, …), got %d (no well-formed กว่า)",
			ErrNotRoundMagnitude, n)
	}
	return Spell(n) + "กว่า", nil
}

// IsRoundMagnitude reports whether n is a positive round magnitude: at least 10
// and, in decimal, a single non-zero leading digit followed only by zeros (10,
// 20, …, 90, 100, 200, …, 1000, 10000, 1000000, …). This is exactly the domain
// on which the Thai กว่า qualifier (see ThaiMoreThan) is well-formed.
func IsRoundMagnitude(n int64) bool {
	if n < 10 {
		return false
	}
	// Strip trailing zeros; what remains must be a single digit 1..9.
	m := n
	for m%10 == 0 {
		m /= 10
	}
	return m >= 1 && m <= 9
}

// Qualified applies the chosen kind qualifier to n, a convenience dispatcher
// over ThaiApprox, ThaiNearly and ThaiMoreThan. For QualifierApprox and
// QualifierNearly the error is always nil; for QualifierMoreThan it propagates
// the ThaiMoreThan round-magnitude error. An unknown kind returns an error.
func Qualified(n int64, kind QualifierKind) (string, error) {
	switch kind {
	case QualifierApprox:
		return ThaiApprox(n), nil
	case QualifierNearly:
		return ThaiNearly(n), nil
	case QualifierMoreThan:
		return ThaiMoreThan(n)
	default:
		return "", fmt.Errorf("thainum: unknown QualifierKind %d", int(kind))
	}
}
