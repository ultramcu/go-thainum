package thainum

import (
	"math/big"
	"testing"
)

func TestSpellCardinal(t *testing.T) {
	cases := []struct {
		n    int64
		want string
	}{
		{0, "ศูนย์"},
		{1, "หนึ่ง"},
		{2, "สอง"},
		{10, "สิบ"},
		{11, "สิบเอ็ด"},
		{12, "สิบสอง"},
		{20, "ยี่สิบ"},
		{21, "ยี่สิบเอ็ด"},
		{25, "ยี่สิบห้า"},
		{100, "หนึ่งร้อย"},
		{101, "หนึ่งร้อยเอ็ด"}, // EtAlways default
		{111, "หนึ่งร้อยสิบเอ็ด"},
		{1000, "หนึ่งพัน"},
		{1001, "หนึ่งพันเอ็ด"},
		{1100, "หนึ่งพันหนึ่งร้อย"},
		{11111, "หนึ่งหมื่นหนึ่งพันหนึ่งร้อยสิบเอ็ด"},
		{1000000, "หนึ่งล้าน"},
		{1000001, "หนึ่งล้านเอ็ด"},
		{10000000, "สิบล้าน"},
		{11000000, "สิบเอ็ดล้าน"},
		{21000000, "ยี่สิบเอ็ดล้าน"},
		{100000000, "หนึ่งร้อยล้าน"},
		{1000000000000, "หนึ่งล้านล้าน"},
		{1234567, "หนึ่งล้านสองแสนสามหมื่นสี่พันห้าร้อยหกสิบเจ็ด"},
		{-5, "ลบห้า"},
	}
	for _, c := range cases {
		if got := Spell(c.n); got != c.want {
			t.Errorf("Spell(%d) = %q, want %q", c.n, got, c.want)
		}
	}
}

func TestSpellEtTensOnly(t *testing.T) {
	sp := Speller{Et: EtTensOnly}
	cases := []struct {
		n    int64
		want string
	}{
		{11, "สิบเอ็ด"},
		{21, "ยี่สิบเอ็ด"},
		{101, "หนึ่งร้อยหนึ่ง"},
		{1001, "หนึ่งพันหนึ่ง"},
		{1000001, "หนึ่งล้านหนึ่ง"},
		{201, "สองร้อยหนึ่ง"},
	}
	for _, c := range cases {
		if got := sp.Int(c.n); got != c.want {
			t.Errorf("EtTensOnly Int(%d) = %q, want %q", c.n, got, c.want)
		}
	}
}

func TestSpellBig(t *testing.T) {
	n, _ := new(big.Int).SetString("1000000000000000000", 10) // 10^18
	if got := SpellBig(n); got != "หนึ่งล้านล้านล้าน" {
		t.Errorf("SpellBig(10^18) = %q, want หนึ่งล้านล้านล้าน", got)
	}
}

func TestSpellDecimal(t *testing.T) {
	cases := []struct{ in, want string }{
		{"0.25", "ศูนย์จุดสองห้า"},
		{"12.34", "สิบสองจุดสามสี่"},
		{"100.00", "หนึ่งร้อย"},
		{"12.05", "สิบสองจุดศูนย์ห้า"},
		{"-3.5", "ลบสามจุดห้า"},
	}
	for _, c := range cases {
		got, err := SpellDecimal(c.in)
		if err != nil {
			t.Errorf("SpellDecimal(%q) error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("SpellDecimal(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNumerals(t *testing.T) {
	if got := ToThaiDigits("Room 101"); got != "Room ๑๐๑" {
		t.Errorf("ToThaiDigits = %q", got)
	}
	if got := ToArabicDigits("ห้อง ๑๐๑"); got != "ห้อง 101" {
		t.Errorf("ToArabicDigits = %q", got)
	}
}
