package money

import (
	"encoding/json"
	"math/big"
	"testing"
)

// All expected strings are taken from the Dart thainum source/tests
// (thainum.dart/test/money_test.dart and money_json_test.dart), used as an
// independent oracle, not from running this Go code.

func TestBahtText(t *testing.T) {
	// money_test.dart: Baht(100).toBahtText() == 'หนึ่งร้อยบาทถ้วน'
	if got := Baht(100).Text(); got != "หนึ่งร้อยบาทถ้วน" {
		t.Errorf("Baht(100).Text() = %q, want %q", got, "หนึ่งร้อยบาทถ้วน")
	}
	// Baht(21).toBahtText() ends with 'บาทถ้วน'.
	if got := Baht(21).Text(); got != "ยี่สิบเอ็ดบาทถ้วน" {
		t.Errorf("Baht(21).Text() = %q, want %q", got, "ยี่สิบเอ็ดบาทถ้วน")
	}
}

func TestSatangText(t *testing.T) {
	// money_test.dart: Satang(2121).toBahtText() == 'ยี่สิบเอ็ดบาทยี่สิบเอ็ดสตางค์'
	if got := Satang(2121).Text(); got != "ยี่สิบเอ็ดบาทยี่สิบเอ็ดสตางค์" {
		t.Errorf("Satang(2121).Text() = %q, want %q", got, "ยี่สิบเอ็ดบาทยี่สิบเอ็ดสตางค์")
	}
	// Satang(25).toBahtText() == 'ยี่สิบห้าสตางค์'
	if got := Satang(25).Text(); got != "ยี่สิบห้าสตางค์" {
		t.Errorf("Satang(25).Text() = %q, want %q", got, "ยี่สิบห้าสตางค์")
	}
}

func TestSatangDecimal(t *testing.T) {
	// money_test.dart: Satang(2121).toDecimal() == '21.21'
	if got := Satang(2121).Decimal(false); got != "21.21" {
		t.Errorf("Satang(2121).Decimal(false) = %q, want %q", got, "21.21")
	}
	// format.dart doc: formatSatang(2121, thaiDigits: true) == '๒๑.๒๑'
	if got := Satang(2121).Decimal(true); got != "๒๑.๒๑" {
		t.Errorf("Satang(2121).Decimal(true) = %q, want %q", got, "๒๑.๒๑")
	}
	// 100000 satang -> '1,000.00' (format.go/format.dart docs).
	if got := Satang(100000).Decimal(false); got != "1,000.00" {
		t.Errorf("Satang(100000).Decimal(false) = %q, want %q", got, "1,000.00")
	}
}

func TestSatangTHB(t *testing.T) {
	// money_test.dart: Satang(2121).toThb() == '฿21.21'
	if got := Satang(2121).THB(false); got != "฿21.21" {
		t.Errorf("Satang(2121).THB(false) = %q, want %q", got, "฿21.21")
	}
	// format.dart doc: thaiDigits keeps ฿ ASCII -> '฿๒๑.๒๑'
	if got := Satang(2121).THB(true); got != "฿๒๑.๒๑" {
		t.Errorf("Satang(2121).THB(true) = %q, want %q", got, "฿๒๑.๒๑")
	}
}

func TestBahtBigIntText(t *testing.T) {
	// money_test.dart forwards to bahtBigInt(10^13). 10^13 baht.
	v := new(big.Int).Exp(big.NewInt(10), big.NewInt(13), nil)
	// Oracle value computed from the spelling rules: 10,000,000,000,000 baht
	// = สิบล้านล้านบาทถ้วน.
	want := "สิบล้านล้านบาทถ้วน"
	if got := (BahtBigInt{V: v}).Text(); got != want {
		t.Errorf("BahtBigInt(10^13).Text() = %q, want %q", got, want)
	}
}

func TestSatangBigIntText(t *testing.T) {
	// money_test.dart forwards to bahtSatangBigInt(10^11) = 1 billion baht in
	// satang -> หนึ่งพันล้านบาทถ้วน.
	v := new(big.Int).Exp(big.NewInt(10), big.NewInt(11), nil)
	want := "หนึ่งพันล้านบาทถ้วน"
	if got := (SatangBigInt{V: v}).Text(); got != want {
		t.Errorf("SatangBigInt(10^11).Text() = %q, want %q", got, want)
	}
}

func TestCmp(t *testing.T) {
	// money_json_test.dart Comparable group.
	if Satang(100).Cmp(Satang(200)) >= 0 {
		t.Error("Satang(100).Cmp(Satang(200)) should be < 0")
	}
	if Satang(200).Cmp(Satang(100)) <= 0 {
		t.Error("Satang(200).Cmp(Satang(100)) should be > 0")
	}
	if Satang(100).Cmp(Satang(100)) != 0 {
		t.Error("Satang(100).Cmp(Satang(100)) should be 0")
	}
	if Baht(1).Cmp(Baht(2)) >= 0 {
		t.Error("Baht(1).Cmp(Baht(2)) should be < 0")
	}
	if (BahtBigInt{V: big.NewInt(5)}).Cmp(BahtBigInt{V: big.NewInt(9)}) >= 0 {
		t.Error("BahtBigInt(5).Cmp(9) should be < 0")
	}
	if (SatangBigInt{V: big.NewInt(9)}).Cmp(SatangBigInt{V: big.NewInt(5)}) <= 0 {
		t.Error("SatangBigInt(9).Cmp(5) should be > 0")
	}
	// nil V treated as zero.
	if (BahtBigInt{}).Cmp(BahtBigInt{V: big.NewInt(1)}) >= 0 {
		t.Error("BahtBigInt(nil).Cmp(1) should be < 0")
	}
}

func TestStringForm(t *testing.T) {
	// money_json_test.dart toString contract.
	cases := []struct {
		got, want string
	}{
		{Baht(100).String(), "Baht(100)"},
		{Satang(2121).String(), "Satang(2121)"},
		{(BahtBigInt{V: big.NewInt(5)}).String(), "BahtBigInt(5)"},
		{(SatangBigInt{V: big.NewInt(9)}).String(), "SatangBigInt(9)"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("String() = %q, want %q", c.got, c.want)
		}
	}
}

func TestBahtSatangJSONRoundTrip(t *testing.T) {
	// money_json_test.dart: Baht/Satang wire form is an int.
	for _, v := range []Baht{100, -5, 0} {
		data, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("Marshal(Baht(%d)): %v", v, err)
		}
		var back Baht
		if err := json.Unmarshal(data, &back); err != nil {
			t.Fatalf("Unmarshal(%s): %v", data, err)
		}
		if back != v {
			t.Errorf("Baht round-trip: got %d, want %d (wire %s)", back, v, data)
		}
	}
	// Baht(100) marshals to exactly 100.
	if data, _ := json.Marshal(Baht(100)); string(data) != "100" {
		t.Errorf("Marshal(Baht(100)) = %s, want 100", data)
	}

	for _, v := range []Satang{2121, 25, 0, -50} {
		data, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("Marshal(Satang(%d)): %v", v, err)
		}
		var back Satang
		if err := json.Unmarshal(data, &back); err != nil {
			t.Fatalf("Unmarshal(%s): %v", data, err)
		}
		if back != v {
			t.Errorf("Satang round-trip: got %d, want %d", back, v)
		}
	}
	if data, _ := json.Marshal(Satang(2121)); string(data) != "2121" {
		t.Errorf("Marshal(Satang(2121)) = %s, want 2121", data)
	}
}

func TestIntWrapperAcceptsIntegerValuedNumber(t *testing.T) {
	// money_json_test.dart: Satang.fromJson(21.0) == Satang(21).
	var s Satang
	if err := json.Unmarshal([]byte("21.0"), &s); err != nil {
		t.Fatalf("Unmarshal(21.0) into Satang: %v", err)
	}
	if s != 21 {
		t.Errorf("Unmarshal(21.0) = %d, want 21", s)
	}
	var b Baht
	if err := json.Unmarshal([]byte("100.0"), &b); err != nil {
		t.Fatalf("Unmarshal(100.0) into Baht: %v", err)
	}
	if b != 100 {
		t.Errorf("Unmarshal(100.0) = %d, want 100", b)
	}
}

func TestIntWrapperRejectsBadShape(t *testing.T) {
	// money_json_test.dart: reject fractional number, string, null, bool, array.
	bad := []string{"2.5", `"100"`, "null", "true", "[1]"}
	for _, in := range bad {
		var s Satang
		if err := json.Unmarshal([]byte(in), &s); err == nil {
			t.Errorf("Unmarshal(%s) into Satang: expected error", in)
		}
	}
}

func TestBigIntJSONWireForm(t *testing.T) {
	// money_json_test.dart: BahtBigInt wire form is a decimal String.
	big1, _ := new(big.Int).SetString("123456789012345678901234567890", 10)
	data, err := json.Marshal(BahtBigInt{V: big1})
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `"123456789012345678901234567890"` {
		t.Errorf("Marshal(BahtBigInt) = %s, want quoted decimal", data)
	}
	var back BahtBigInt
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatal(err)
	}
	if back.V.Cmp(big1) != 0 {
		t.Errorf("BahtBigInt round-trip mismatch: got %s", back.V)
	}

	// SatangBigInt negative wire form.
	neg, _ := new(big.Int).SetString("-999999999999999999999", 10)
	data2, _ := json.Marshal(SatangBigInt{V: neg})
	if string(data2) != `"-999999999999999999999"` {
		t.Errorf("Marshal(SatangBigInt) = %s, want quoted decimal", data2)
	}
	var back2 SatangBigInt
	if err := json.Unmarshal(data2, &back2); err != nil {
		t.Fatal(err)
	}
	if back2.V.Cmp(neg) != 0 {
		t.Errorf("SatangBigInt round-trip mismatch: got %s", back2.V)
	}
}

func TestBigIntAcceptsBareInt(t *testing.T) {
	// money_json_test.dart: BigInt wrappers accept a bare int for convenience.
	var b BahtBigInt
	if err := json.Unmarshal([]byte("5"), &b); err != nil {
		t.Fatal(err)
	}
	if b.V.Cmp(big.NewInt(5)) != 0 {
		t.Errorf("Unmarshal(5) into BahtBigInt = %s, want 5", b.V)
	}
	var s SatangBigInt
	if err := json.Unmarshal([]byte("-7"), &s); err != nil {
		t.Fatal(err)
	}
	if s.V.Cmp(big.NewInt(-7)) != 0 {
		t.Errorf("Unmarshal(-7) into SatangBigInt = %s, want -7", s.V)
	}
}

func TestBigIntRejectsBadShape(t *testing.T) {
	// money_json_test.dart: reject malformed string, fractional string, double, null.
	bad := []string{`"not-a-number"`, `"12.5"`, "1.5", "null"}
	for _, in := range bad {
		var b BahtBigInt
		if err := json.Unmarshal([]byte(in), &b); err == nil {
			t.Errorf("Unmarshal(%s) into BahtBigInt: expected error", in)
		}
	}
}

func TestNilBigIntMarshalsAsZero(t *testing.T) {
	data, _ := json.Marshal(BahtBigInt{})
	if string(data) != `"0"` {
		t.Errorf("Marshal(BahtBigInt{}) = %s, want \"0\"", data)
	}
	if got := (BahtBigInt{}).Text(); got != "ศูนย์บาทถ้วน" {
		t.Errorf("BahtBigInt{}.Text() = %q, want ศูนย์บาทถ้วน", got)
	}
}

func TestRealJSONEncodeDecodePass(t *testing.T) {
	// money_json_test.dart: survives a real encode/decode of a struct.
	type wallet struct {
		Baht      Baht         `json:"baht"`
		Satang    Satang       `json:"satang"`
		BigBaht   BahtBigInt   `json:"bigBaht"`
		BigSatang SatangBigInt `json:"bigSatang"`
	}
	bb, _ := new(big.Int).SetString("10000000000000000000", 10)
	bs, _ := new(big.Int).SetString("20000000000000000001", 10)
	in := wallet{
		Baht:      100,
		Satang:    2121,
		BigBaht:   BahtBigInt{V: bb},
		BigSatang: SatangBigInt{V: bs},
	}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out wallet
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	if out.Baht != 100 || out.Satang != 2121 ||
		out.BigBaht.V.Cmp(bb) != 0 || out.BigSatang.V.Cmp(bs) != 0 {
		t.Errorf("wallet round-trip mismatch: %+v", out)
	}
}
