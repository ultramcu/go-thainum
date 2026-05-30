package money

// Blind parity tests for the #11 money-type ordering + JSON wire contract.
// Expected values are derived from the Dart 0.5.3 oracle —
// thainum.dart/lib/src/money.dart and thainum.dart/test/money_json_test.dart —
// NOT from running the Go code.
//
// Go API (from issue #11):
//   - Baht, Satang are named integer types: money.Baht(100), money.Satang(2121).
//   - BahtBigInt, SatangBigInt are struct{ V *big.Int }.
//   - Ordering is the Cmp(other) int method (Dart's compareTo): <0 / 0 / >0.
//   - JSON: Baht/Satang marshal to a bare JSON int; BigInt variants marshal to
//     a decimal JSON string. UnmarshalJSON is the inverse.
//
// Function names are kept distinct from money_test.go / parity_money_test.go so
// the three files coexist in package money.

import (
	"encoding/json"
	"math/big"
	"sort"
	"testing"
)

// Port of money_json_test.dart group 'Comparable' (Satang/Baht compareTo).
func TestOrderingCmpSign(t *testing.T) {
	// Satang.compareTo orders by value.
	if Satang(100).Cmp(Satang(200)) >= 0 {
		t.Error("Satang(100).Cmp(Satang(200)) should be < 0")
	}
	if Satang(200).Cmp(Satang(100)) <= 0 {
		t.Error("Satang(200).Cmp(Satang(100)) should be > 0")
	}
	if Satang(100).Cmp(Satang(100)) != 0 {
		t.Error("Satang(100).Cmp(Satang(100)) should be 0")
	}
	// Baht.compareTo orders by value.
	if Baht(1).Cmp(Baht(2)) >= 0 {
		t.Error("Baht(1).Cmp(Baht(2)) should be < 0")
	}
	if Baht(2).Cmp(Baht(2)) != 0 {
		t.Error("Baht(2).Cmp(Baht(2)) should be 0")
	}
	// BigInt wrappers order by value.
	if (BahtBigInt{V: big.NewInt(5)}).Cmp(BahtBigInt{V: big.NewInt(9)}) >= 0 {
		t.Error("BahtBigInt(5).Cmp(9) should be < 0")
	}
	if (SatangBigInt{V: big.NewInt(9)}).Cmp(SatangBigInt{V: big.NewInt(5)}) <= 0 {
		t.Error("SatangBigInt(9).Cmp(5) should be > 0")
	}
}

// Port of 'a list of Satang sorts ascending' — Cmp must drive sort.Slice into
// the same order as the Dart oracle (-50, 0, 100, 300).
func TestOrderingSatangSortAscending(t *testing.T) {
	list := []Satang{Satang(300), Satang(-50), Satang(100), Satang(0)}
	sort.Slice(list, func(i, j int) bool { return list[i].Cmp(list[j]) < 0 })
	want := []Satang{Satang(-50), Satang(0), Satang(100), Satang(300)}
	for i := range want {
		if list[i] != want[i] {
			t.Fatalf("sorted Satang = %v, want %v", list, want)
		}
	}
}

// Port of 'a list of BahtBigInt sorts ascending' — including a value far beyond
// int64 range (-7, 0, 1000000000000000000000).
func TestOrderingBahtBigIntSortAscending(t *testing.T) {
	huge, _ := new(big.Int).SetString("1000000000000000000000", 10)
	list := []BahtBigInt{
		{V: huge},
		{V: big.NewInt(-7)},
		{V: big.NewInt(0)},
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Cmp(list[j]) < 0 })
	want := []string{"-7", "0", "1000000000000000000000"}
	for i, w := range want {
		if list[i].V.String() != w {
			t.Fatalf("sorted BahtBigInt[%d] = %s, want %s", i, list[i].V.String(), w)
		}
	}
}

// --- JSON wire form (money_json_test.dart 'toJson / fromJson round-trip') ----

// Baht / Satang wire form is a bare JSON int that round-trips.
func TestOrderingIntWireFormRoundTrip(t *testing.T) {
	for _, v := range []Baht{Baht(100), Baht(-5), Baht(0)} {
		data, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("Marshal Baht(%d): %v", int64(v), err)
		}
		var back Baht
		if err := json.Unmarshal(data, &back); err != nil {
			t.Fatalf("Unmarshal %s into Baht: %v", data, err)
		}
		if back != v {
			t.Errorf("Baht round-trip: got %d want %d", int64(back), int64(v))
		}
	}
	if data, _ := json.Marshal(Baht(100)); string(data) != "100" {
		t.Errorf("Marshal Baht(100) = %s, want 100 (bare int)", data)
	}
	if data, _ := json.Marshal(Satang(2121)); string(data) != "2121" {
		t.Errorf("Marshal Satang(2121) = %s, want 2121 (bare int)", data)
	}
}

// BigInt wire form is a decimal JSON string, round-tripping values past int64.
func TestOrderingBigIntWireFormRoundTrip(t *testing.T) {
	big1, _ := new(big.Int).SetString("123456789012345678901234567890", 10)
	data, err := json.Marshal(BahtBigInt{V: big1})
	if err != nil {
		t.Fatalf("Marshal BahtBigInt: %v", err)
	}
	if string(data) != `"123456789012345678901234567890"` {
		t.Errorf("Marshal BahtBigInt = %s, want quoted decimal string", data)
	}
	var back BahtBigInt
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("Unmarshal BahtBigInt: %v", err)
	}
	if back.V.Cmp(big1) != 0 {
		t.Errorf("BahtBigInt round-trip mismatch: got %s", back.V)
	}

	neg, _ := new(big.Int).SetString("-999999999999999999999", 10)
	data2, err := json.Marshal(SatangBigInt{V: neg})
	if err != nil {
		t.Fatalf("Marshal SatangBigInt: %v", err)
	}
	if string(data2) != `"-999999999999999999999"` {
		t.Errorf("Marshal SatangBigInt = %s, want quoted decimal string", data2)
	}
	var back2 SatangBigInt
	if err := json.Unmarshal(data2, &back2); err != nil {
		t.Fatalf("Unmarshal SatangBigInt: %v", err)
	}
	if back2.V.Cmp(neg) != 0 {
		t.Errorf("SatangBigInt round-trip mismatch: got %s", back2.V)
	}
}

// Port of 'int wrappers accept an integer-valued JSON number (21.0)'.
func TestOrderingIntWrapperAcceptsIntegerValuedNumber(t *testing.T) {
	var s Satang
	if err := json.Unmarshal([]byte("21.0"), &s); err != nil {
		t.Fatalf("Unmarshal 21.0 into Satang: %v", err)
	}
	if s != Satang(21) {
		t.Errorf("Unmarshal 21.0 = Satang(%d), want Satang(21)", int64(s))
	}
	var b Baht
	if err := json.Unmarshal([]byte("100.0"), &b); err != nil {
		t.Fatalf("Unmarshal 100.0 into Baht: %v", err)
	}
	if b != Baht(100) {
		t.Errorf("Unmarshal 100.0 = Baht(%d), want Baht(100)", int64(b))
	}
}

// Port of 'fromJson throws on bad shape' for the int wrappers: a fractional
// number and a string must be rejected.
func TestOrderingIntWrapperRejectsBadShape(t *testing.T) {
	var s Satang
	if err := json.Unmarshal([]byte("2.5"), &s); err == nil {
		t.Error("Unmarshal 2.5 into Satang should error (fractional number)")
	}
	var b Baht
	if err := json.Unmarshal([]byte("1.1"), &b); err == nil {
		t.Error("Unmarshal 1.1 into Baht should error (fractional number)")
	}
	var s2 Satang
	if err := json.Unmarshal([]byte(`"100"`), &s2); err == nil {
		t.Error(`Unmarshal "100" (string) into Satang should error`)
	}
}

// Port of 'BigInt wrappers accept a bare int for convenience' + 'reject a
// malformed string / double'.
func TestOrderingBigIntWrapperShapes(t *testing.T) {
	var b BahtBigInt
	if err := json.Unmarshal([]byte("5"), &b); err != nil {
		t.Fatalf("Unmarshal bare int 5 into BahtBigInt: %v", err)
	}
	if b.V.Cmp(big.NewInt(5)) != 0 {
		t.Errorf("Unmarshal 5 into BahtBigInt = %s, want 5", b.V)
	}
	var s SatangBigInt
	if err := json.Unmarshal([]byte("-7"), &s); err != nil {
		t.Fatalf("Unmarshal bare int -7 into SatangBigInt: %v", err)
	}
	if s.V.Cmp(big.NewInt(-7)) != 0 {
		t.Errorf("Unmarshal -7 into SatangBigInt = %s, want -7", s.V)
	}
	// Malformed / fractional string must be rejected.
	for _, in := range []string{`"not-a-number"`, `"12.5"`} {
		var bad BahtBigInt
		if err := json.Unmarshal([]byte(in), &bad); err == nil {
			t.Errorf("Unmarshal %s into BahtBigInt should error", in)
		}
	}
}
