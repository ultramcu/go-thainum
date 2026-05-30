package money_test

// Parity tests for the money wrapper types, reconciled to the real Go API:
//   money.Baht, money.Satang        — named int64 (construct money.Baht(100))
//   money.BahtBigInt, money.SatangBigInt — struct{ V *big.Int }
//   ordering via Cmp(other) int; MarshalJSON/UnmarshalJSON.
// Oracle: thainum.dart/lib/src/money.dart + money_json_test.dart (wire form:
// Baht/Satang -> JSON int; BigInt variants -> decimal String).
import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ultramcu/go-thainum/money"
)

func TestBahtSatangValueEquality(t *testing.T) {
	if money.Baht(100) != money.Baht(100) {
		t.Error("Baht(100) should == Baht(100)")
	}
	if money.Baht(100) == money.Baht(101) {
		t.Error("Baht(100) should != Baht(101)")
	}
	if money.Satang(2121) != money.Satang(2121) {
		t.Error("Satang(2121) should == Satang(2121)")
	}
}

func TestCmpOrdering(t *testing.T) {
	if money.Satang(100).Cmp(money.Satang(200)) >= 0 {
		t.Error("Satang(100).Cmp(200) should be < 0")
	}
	if money.Satang(200).Cmp(money.Satang(100)) <= 0 {
		t.Error("Satang(200).Cmp(100) should be > 0")
	}
	if money.Satang(100).Cmp(money.Satang(100)) != 0 {
		t.Error("Satang(100).Cmp(100) should be 0")
	}
	if money.Baht(1).Cmp(money.Baht(2)) >= 0 {
		t.Error("Baht(1).Cmp(2) should be < 0")
	}
	a, b, c := money.Satang(-50), money.Satang(0), money.Satang(300)
	if !(a.Cmp(b) < 0 && b.Cmp(c) < 0 && a.Cmp(c) < 0) {
		t.Error("expected -50 < 0 < 300")
	}
}

func TestBigIntValueEqualityAndCmp(t *testing.T) {
	a := money.BahtBigInt{V: big.NewInt(123)}
	b := money.BahtBigInt{V: big.NewInt(123)}
	if a.Cmp(b) != 0 {
		t.Errorf("BahtBigInt(123).Cmp(123) = %d, want 0", a.Cmp(b))
	}
	if a.Cmp(money.BahtBigInt{V: big.NewInt(124)}) == 0 {
		t.Error("BahtBigInt(123) should != 124")
	}
	if (money.SatangBigInt{V: big.NewInt(9)}).Cmp(money.SatangBigInt{V: big.NewInt(5)}) <= 0 {
		t.Error("SatangBigInt(9).Cmp(5) should be > 0")
	}
}

func TestBahtSatangJSONWireForm(t *testing.T) {
	cases := []struct {
		v    interface{}
		want string
	}{
		{money.Baht(100), "100"},
		{money.Baht(-5), "-5"},
		{money.Satang(2121), "2121"},
		{money.Satang(0), "0"},
	}
	for _, c := range cases {
		b, err := json.Marshal(c.v)
		if err != nil {
			t.Errorf("Marshal(%v) error: %v", c.v, err)
			continue
		}
		if string(b) != c.want {
			t.Errorf("Marshal(%v) = %q, want %q (bare int)", c.v, string(b), c.want)
		}
	}
}

func TestBigIntJSONWireForm(t *testing.T) {
	bigVal := "123456789012345678901234567890"
	bi, _ := new(big.Int).SetString(bigVal, 10)
	b, err := json.Marshal(money.BahtBigInt{V: bi})
	if err != nil {
		t.Fatalf("BahtBigInt Marshal error: %v", err)
	}
	if want := `"` + bigVal + `"`; string(b) != want {
		t.Errorf("BahtBigInt Marshal = %q, want %q (decimal String)", string(b), want)
	}
	negVal := "-999999999999999999999"
	nbi, _ := new(big.Int).SetString(negVal, 10)
	b, err = json.Marshal(money.SatangBigInt{V: nbi})
	if err != nil {
		t.Fatalf("SatangBigInt Marshal error: %v", err)
	}
	if want := `"` + negVal + `"`; string(b) != want {
		t.Errorf("SatangBigInt Marshal = %q, want %q (decimal String)", string(b), want)
	}
}

func TestJSONRoundTrip(t *testing.T) {
	{
		orig := money.Baht(100)
		b, _ := json.Marshal(orig)
		var got money.Baht
		if err := json.Unmarshal(b, &got); err != nil {
			t.Fatalf("Baht Unmarshal: %v", err)
		}
		if got != orig {
			t.Errorf("Baht round-trip = %v want %v", got, orig)
		}
	}
	{
		orig := money.Satang(2121)
		b, _ := json.Marshal(orig)
		var got money.Satang
		if err := json.Unmarshal(b, &got); err != nil {
			t.Fatalf("Satang Unmarshal: %v", err)
		}
		if got != orig {
			t.Errorf("Satang round-trip = %v want %v", got, orig)
		}
	}
	{
		bi, _ := new(big.Int).SetString("10000000000000000000", 10)
		orig := money.BahtBigInt{V: bi}
		b, _ := json.Marshal(orig)
		var got money.BahtBigInt
		if err := json.Unmarshal(b, &got); err != nil {
			t.Fatalf("BahtBigInt Unmarshal: %v", err)
		}
		if orig.Cmp(got) != 0 {
			t.Errorf("BahtBigInt round-trip Cmp != 0: got %v want %v", got.V, orig.V)
		}
	}
}

func TestBigIntUnmarshalRejectsMalformed(t *testing.T) {
	var bb money.BahtBigInt
	if err := json.Unmarshal([]byte(`"not-a-number"`), &bb); err == nil {
		t.Error("BahtBigInt.Unmarshal of garbage should error")
	}
	var sb money.SatangBigInt
	if err := json.Unmarshal([]byte(`"12.5"`), &sb); err == nil {
		t.Error("SatangBigInt.Unmarshal of 12.5 should error")
	}
}
