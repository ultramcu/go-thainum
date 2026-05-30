// Package money provides thin, type-safe wrappers around an integer (or
// math/big.Int) money amount, on top of the github.com/ultramcu/go-thainum
// core package.
//
// In untyped code it is easy to pass a satang amount somewhere that expects
// baht: the resulting Thai Baht text is silently wrong. Wrapping the amount in
// a distinct named type gives the compiler enough information to catch that at
// the call site:
//
//	money.Baht(100).Text()    // "หนึ่งร้อยบาทถ้วน"        (baht)
//	money.Satang(2121).Text() // "ยี่สิบเอ็ดบาทยี่สิบเอ็ดสตางค์" (satang)
//
// The wrappers are value types (Baht and Satang are named integer types;
// BahtBigInt and SatangBigInt wrap a *big.Int) so they compare and serialize
// by value.
//
// This is a separate package from the core go-thainum because the core
// already exports Baht and BahtSatang as functions, which would collide with
// the type names; importing this package is the only thing that pulls the
// wrappers in.
//
// Adaptations from the original Dart thainum money wrappers:
//
//   - Dart's copyWith is not idiomatic in Go and is dropped: a value of a
//     named integer type is copied by assignment, and BahtBigInt/SatangBigInt
//     expose their V field directly. Construct a new value instead.
//   - Dart's Comparable.compareTo is replaced by Cmp methods (and, for the
//     int-backed types, ordinary < / > comparison also works).
//   - JSON is wired through encoding/json's Marshaler/Unmarshaler instead of
//     Dart's toJson/fromJson. The wire forms match Dart exactly: Baht and
//     Satang marshal as a JSON integer; BahtBigInt and SatangBigInt marshal as
//     a decimal JSON string (a big integer can exceed the safe range of a JSON
//     number).
package money

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"

	thainum "github.com/ultramcu/go-thainum"
)

// Baht is a whole-baht amount. Use it when you want the compiler to enforce
// that the integer you hold is measured in baht (not satang). Use BahtBigInt
// for amounts beyond the int64 range.
type Baht int64

// Text spells this amount as Thai Baht text: Baht(100).Text() ->
// "หนึ่งร้อยบาทถ้วน".
func (b Baht) Text() string { return thainum.Baht(int64(b)) }

// Cmp orders by the underlying whole-baht value: it returns -1, 0 or +1 as b
// is less than, equal to, or greater than other.
func (b Baht) Cmp(other Baht) int {
	switch {
	case b < other:
		return -1
	case b > other:
		return 1
	default:
		return 0
	}
}

// MarshalJSON encodes b as a JSON integer (the canonical wire form). It
// round-trips through UnmarshalJSON.
func (b Baht) MarshalJSON() ([]byte, error) { return json.Marshal(int64(b)) }

// UnmarshalJSON decodes a JSON integer (a number with no fractional part is
// also accepted, so 100.0 round-trips) into b. It returns an error on any
// other shape.
func (b *Baht) UnmarshalJSON(data []byte) error {
	v, err := intFromJSON(data, "Baht")
	if err != nil {
		return err
	}
	*b = Baht(v)
	return nil
}

// String returns the Go debug form, e.g. "Baht(100)".
func (b Baht) String() string { return fmt.Sprintf("Baht(%d)", int64(b)) }

// Satang is an amount in satang (1 baht = 100 satang). All methods preserve
// the exact integer satang value through the conversion — there is no float
// round-trip. Use SatangBigInt for amounts beyond the int64 range.
type Satang int64

// Text spells this amount as Thai Baht text: Satang(2121).Text() ->
// "ยี่สิบเอ็ดบาทยี่สิบเอ็ดสตางค์".
func (s Satang) Text() string { return thainum.BahtSatang(int64(s)) }

// Decimal is the decimal display in baht: Satang(2121).Decimal() -> "21.21".
// With thaiDigits true the digits render as Thai numerals ("๒๑.๒๑"); the
// thousands commas, decimal point and minus sign stay ASCII.
func (s Satang) Decimal(thaiDigits bool) string {
	out := thainum.FormatSatang(int64(s))
	if thaiDigits {
		return thainum.ToThaiDigits(out)
	}
	return out
}

// THB is the ฿-prefixed decimal display: Satang(2121).THB() -> "฿21.21". With
// thaiDigits true the digits render as Thai numerals ("฿๒๑.๒๑"); the ฿ symbol,
// commas, decimal point and minus sign stay ASCII.
func (s Satang) THB(thaiDigits bool) string {
	out := thainum.FormatTHB(int64(s))
	if thaiDigits {
		return thainum.ToThaiDigits(out)
	}
	return out
}

// Cmp orders by the underlying satang value (-1, 0 or +1).
func (s Satang) Cmp(other Satang) int {
	switch {
	case s < other:
		return -1
	case s > other:
		return 1
	default:
		return 0
	}
}

// MarshalJSON encodes s as a JSON integer (the canonical wire form).
func (s Satang) MarshalJSON() ([]byte, error) { return json.Marshal(int64(s)) }

// UnmarshalJSON decodes a JSON integer (a number with no fractional part is
// also accepted) into s. It returns an error on any other shape.
func (s *Satang) UnmarshalJSON(data []byte) error {
	v, err := intFromJSON(data, "Satang")
	if err != nil {
		return err
	}
	*s = Satang(v)
	return nil
}

// String returns the Go debug form, e.g. "Satang(2121)".
func (s Satang) String() string { return fmt.Sprintf("Satang(%d)", int64(s)) }

// BahtBigInt is a whole-baht amount stored as a *big.Int, for amounts that
// exceed the int64 range. The zero value (V == nil) is treated as zero baht.
type BahtBigInt struct {
	V *big.Int
}

// Text spells this amount as Thai Baht text.
func (b BahtBigInt) Text() string { return thainum.BahtBig(b.V) }

// Cmp orders by the underlying value (-1, 0 or +1). A nil V is treated as zero.
func (b BahtBigInt) Cmp(other BahtBigInt) int {
	return bigOrZero(b.V).Cmp(bigOrZero(other.V))
}

// MarshalJSON encodes b as a decimal JSON string, because a big integer can
// exceed the safe-integer range of a JSON number. A nil V encodes as "0".
func (b BahtBigInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(bigOrZero(b.V).String())
}

// UnmarshalJSON decodes a decimal JSON string (the canonical form) or a JSON
// integer into b. It returns an error on any other shape.
func (b *BahtBigInt) UnmarshalJSON(data []byte) error {
	v, err := bigFromJSON(data, "BahtBigInt")
	if err != nil {
		return err
	}
	b.V = v
	return nil
}

// String returns the Go debug form, e.g. "BahtBigInt(5)".
func (b BahtBigInt) String() string { return fmt.Sprintf("BahtBigInt(%s)", bigOrZero(b.V).String()) }

// SatangBigInt is a satang amount stored as a *big.Int, for amounts that
// exceed the int64 range. The zero value (V == nil) is treated as zero satang.
type SatangBigInt struct {
	V *big.Int
}

// Text spells this amount as Thai Baht text.
func (s SatangBigInt) Text() string { return thainum.BahtSatangBig(s.V) }

// Cmp orders by the underlying value (-1, 0 or +1). A nil V is treated as zero.
func (s SatangBigInt) Cmp(other SatangBigInt) int {
	return bigOrZero(s.V).Cmp(bigOrZero(other.V))
}

// MarshalJSON encodes s as a decimal JSON string. A nil V encodes as "0".
func (s SatangBigInt) MarshalJSON() ([]byte, error) {
	return json.Marshal(bigOrZero(s.V).String())
}

// UnmarshalJSON decodes a decimal JSON string (the canonical form) or a JSON
// integer into s. It returns an error on any other shape.
func (s *SatangBigInt) UnmarshalJSON(data []byte) error {
	v, err := bigFromJSON(data, "SatangBigInt")
	if err != nil {
		return err
	}
	s.V = v
	return nil
}

// String returns the Go debug form, e.g. "SatangBigInt(9)".
func (s SatangBigInt) String() string {
	return fmt.Sprintf("SatangBigInt(%s)", bigOrZero(s.V).String())
}

func bigOrZero(v *big.Int) *big.Int {
	if v == nil {
		return big.NewInt(0)
	}
	return v
}

// intFromJSON decodes the JSON wire form of an int-backed money wrapper. It
// accepts a JSON integer, or a number that is mathematically an integer (so
// 21.0 round-trips), and rejects fractional numbers, strings and any other
// shape. typ names the wrapper for the error message.
func intFromJSON(data []byte, typ string) (int64, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var raw interface{}
	if err := dec.Decode(&raw); err != nil {
		return 0, fmt.Errorf("thainum: %s.UnmarshalJSON: %w", typ, err)
	}
	n, ok := raw.(json.Number)
	if !ok {
		return 0, fmt.Errorf("thainum: %s.UnmarshalJSON expects an integer value, got %T", typ, raw)
	}
	if i, err := n.Int64(); err == nil {
		return i, nil
	}
	// A JSON number written with a fractional part: accept it only if it is
	// mathematically an integer (e.g. 21.0), matching the Dart contract.
	f, err := n.Float64()
	if err == nil && f == float64(int64(f)) {
		return int64(f), nil
	}
	return 0, fmt.Errorf("thainum: %s.UnmarshalJSON expects an integer value, got %q", typ, n.String())
}

// bigFromJSON decodes the JSON wire form of a big-int-backed money wrapper. It
// accepts a decimal JSON string (the canonical form) or a JSON integer, and
// rejects fractional numbers, malformed strings and any other shape.
func bigFromJSON(data []byte, typ string) (*big.Int, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var raw interface{}
	if err := dec.Decode(&raw); err != nil {
		return nil, fmt.Errorf("thainum: %s.UnmarshalJSON: %w", typ, err)
	}
	switch v := raw.(type) {
	case string:
		if b, ok := new(big.Int).SetString(trimSpace(v), 10); ok {
			return b, nil
		}
	case json.Number:
		if b, ok := new(big.Int).SetString(v.String(), 10); ok {
			return b, nil
		}
	}
	return nil, fmt.Errorf("thainum: %s.UnmarshalJSON expects a decimal string value, got %T", typ, raw)
}

// trimSpace trims leading and trailing ASCII whitespace from s. It mirrors the
// Dart wrapper's String.trim() before parsing.
func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && isSpace(s[start]) {
		start++
	}
	for end > start && isSpace(s[end-1]) {
		end--
	}
	return s[start:end]
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\v' || c == '\f'
}
