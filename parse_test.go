package thainum

import (
	"errors"
	"math/big"
	"math/rand"
	"testing"
)

// ---- explicit golden cases ----------------------------------------------

func TestParseIntGolden(t *testing.T) {
	cases := []struct {
		in   string
		want int64
	}{
		{"ศูนย์", 0},
		{"หนึ่ง", 1},
		{"สอง", 2},
		{"สิบ", 10},
		{"สิบเอ็ด", 11},
		{"สิบห้า", 15},
		{"ยี่สิบ", 20},
		{"ยี่สิบเอ็ด", 21},
		{"ยี่สิบห้า", 25},
		{"สามสิบ", 30},
		{"หนึ่งร้อย", 100},
		{"หนึ่งร้อยเอ็ด", 101},  // EtAlways form
		{"หนึ่งร้อยหนึ่ง", 101}, // EtTensOnly form must also parse
		{"ร้อยเอ็ด", 101},       // bare ร้อย == 1 hundred
		{"ร้อยหนึ่ง", 101},
		{"หนึ่งร้อยสิบเอ็ด", 111},
		{"หนึ่งพัน", 1000},
		{"หนึ่งพันเอ็ด", 1001},
		{"พันเอ็ด", 1001},
		{"พันหนึ่ง", 1001},
		{"หนึ่งพันหนึ่งร้อย", 1100},
		{"หนึ่งหมื่นหนึ่งพันหนึ่งร้อยสิบเอ็ด", 11111},
		{"หนึ่งล้าน", 1000000},
		{"หนึ่งล้านเอ็ด", 1000001},
		{"สิบล้าน", 10000000},
		{"สิบเอ็ดล้าน", 11000000},
		{"ยี่สิบเอ็ดล้าน", 21000000},
		{"หนึ่งร้อยล้าน", 100000000},
		{"หนึ่งล้านล้าน", 1000000000000},
		{"หนึ่งล้านสองแสนสามหมื่นสี่พันห้าร้อยหกสิบเจ็ด", 1234567},
		{"ลบห้า", -5},
		{"ลบยี่สิบเอ็ด", -21},
		// Thai / Arabic digit shortcuts
		{"๒๑", 21},
		{"21", 21},
		{"-5", -5},
		{"๐", 0},
		{"  ยี่สิบเอ็ด  ", 21}, // surrounding spaces
	}
	for _, c := range cases {
		got, err := ParseInt(c.in)
		if err != nil {
			t.Errorf("ParseInt(%q) error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("ParseInt(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestParseBigGolden(t *testing.T) {
	mustBig := func(s string) *big.Int {
		v, _ := new(big.Int).SetString(s, 10)
		return v
	}
	cases := []struct {
		in   string
		want *big.Int
	}{
		{"หนึ่งล้านล้าน", mustBig("1000000000000")},           // 10^12
		{"หนึ่งล้านล้านล้าน", mustBig("1000000000000000000")}, // 10^18
		{"ศูนย์", big.NewInt(0)},
		{"ลบหนึ่งล้านล้าน", mustBig("-1000000000000")},
	}
	for _, c := range cases {
		got, err := ParseBig(c.in)
		if err != nil {
			t.Errorf("ParseBig(%q) error: %v", c.in, err)
			continue
		}
		if got.Cmp(c.want) != 0 {
			t.Errorf("ParseBig(%q) = %s, want %s", c.in, got, c.want)
		}
	}
}

func TestParseIntOverflow(t *testing.T) {
	// 10^18 fits int64; 10^24 (ล้าน^4) does not.
	_, err := ParseInt("หนึ่งล้านล้านล้านล้าน") // 10^24
	if err == nil {
		t.Fatal("expected overflow error for 10^24, got nil")
	}
	if !errors.Is(err, ErrParse) {
		t.Errorf("overflow error should wrap ErrParse, got %v", err)
	}
}

func TestParseErrors(t *testing.T) {
	bad := []string{
		"",
		"   ",
		"abc",
		"สิบสิบ",     // ascending/repeat place
		"ร้อยร้อย",   // repeat place
		"สิบร้อย",    // ascending place
		"สองสาม",     // two digits in a row with no place
		"ลบ",         // sign with no number
		"หนึ่งศูนย์", // misplaced zero
		"ยี่",        // ยี่ without สิบ
		"ยี่ร้อย",    // ยี่ before non-สิบ
	}
	for _, in := range bad {
		if _, err := ParseInt(in); err == nil {
			t.Errorf("ParseInt(%q) expected error, got nil", in)
		} else if !errors.Is(err, ErrParse) {
			t.Errorf("ParseInt(%q) error %v should wrap ErrParse", in, err)
		}
	}
}

// ---- baht golden cases ---------------------------------------------------

func TestParseBahtGolden(t *testing.T) {
	cases := []struct {
		in   string
		want int64
	}{
		{"ยี่สิบเอ็ดบาทยี่สิบเอ็ดสตางค์", 2121},
		{"หนึ่งร้อยบาทถ้วน", 10000},
		{"ศูนย์บาทถ้วน", 0},
		{"ยี่สิบห้าสตางค์", 25},
		{"ลบหนึ่งบาทหนึ่งสตางค์", -101},
		{"หนึ่งบาทถ้วน", 100},
		{"หนึ่งร้อยบาทห้าสิบสตางค์", 10050},
		{"หนึ่งล้านบาทเจ็ดสิบห้าสตางค์", 100000075},
		{"หนึ่งสตางค์", 1},
		{"ห้าสิบสตางค์", 50},
		{"ลบห้าบาทถ้วน", -500},
		{"ลบยี่สิบห้าสตางค์", -25},
		{"หนึ่งร้อยเอ็ดบาทถ้วน", 10100},
	}
	for _, c := range cases {
		got, err := ParseBaht(c.in)
		if err != nil {
			t.Errorf("ParseBaht(%q) error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("ParseBaht(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

// ---- round-trip: the correctness gate -----------------------------------

func TestRoundTripInt(t *testing.T) {
	fixed := []int64{
		0, 1, 2, 10, 11, 20, 21, 25, 100, 101, 111,
		1000, 1001, 1100, 11111, 1000000, 1000001, 10000000,
		11000000, 21000000, 100000000, 1234567, 1000000000000,
		-1, -5, -21, -101, -1000000, -1234567, -1000000000000,
		999999, 999999999999, -999999999999,
	}
	spEt := Speller{Et: EtTensOnly}

	check := func(n int64) {
		// default EtAlways form
		words := Spell(n)
		got, err := ParseInt(words)
		if err != nil {
			t.Errorf("ParseInt(Spell(%d)=%q) error: %v", n, words, err)
			return
		}
		if got != n {
			t.Errorf("round-trip default: ParseInt(%q) = %d, want %d", words, got, n)
		}
		// EtTensOnly form
		words2 := spEt.Int(n)
		got2, err := ParseInt(words2)
		if err != nil {
			t.Errorf("ParseInt(EtTensOnly Int(%d)=%q) error: %v", n, words2, err)
			return
		}
		if got2 != n {
			t.Errorf("round-trip EtTensOnly: ParseInt(%q) = %d, want %d", words2, got2, n)
		}
	}

	for _, n := range fixed {
		check(n)
	}

	// A few thousand pseudo-random int64 in [-10^15, 10^15].
	rng := rand.New(rand.NewSource(20260525))
	const span = int64(1_000_000_000_000_000) // 10^15
	for i := 0; i < 5000; i++ {
		n := rng.Int63n(2*span+1) - span
		check(n)
	}
}

func TestRoundTripBig(t *testing.T) {
	mustBig := func(s string) *big.Int {
		v, _ := new(big.Int).SetString(s, 10)
		return v
	}
	cases := []*big.Int{
		big.NewInt(0),
		mustBig("1000000000000"),           // 10^12
		mustBig("1000000000000000000"),     // 10^18
		mustBig("999999999999999999999"),   // 21 nines
		mustBig("-1000000000000000000000"), // -10^21
		mustBig("123456789012345678901234567890"),
	}
	for _, n := range cases {
		words := SpellBig(n)
		got, err := ParseBig(words)
		if err != nil {
			t.Errorf("ParseBig(SpellBig(%s)=%q) error: %v", n, words, err)
			continue
		}
		if got.Cmp(n) != 0 {
			t.Errorf("round-trip big: ParseBig(%q) = %s, want %s", words, got, n)
		}
	}
}

func TestRoundTripBaht(t *testing.T) {
	fixed := []int64{
		0, 1, 25, 50, 100, 2121, 10100, 100000075,
		-1, -25, -101, -100000075, 99, 199, 10000000000,
	}
	check := func(s int64) {
		text := BahtSatang(s)
		got, err := ParseBaht(text)
		if err != nil {
			t.Errorf("ParseBaht(BahtSatang(%d)=%q) error: %v", s, text, err)
			return
		}
		if got != s {
			t.Errorf("round-trip baht: ParseBaht(%q) = %d, want %d", text, got, s)
		}
	}
	for _, s := range fixed {
		check(s)
	}
	rng := rand.New(rand.NewSource(525))
	const span = int64(100_000_000_000) // up to ~1e9 baht
	for i := 0; i < 3000; i++ {
		s := rng.Int63n(2*span+1) - span
		check(s)
	}
}
