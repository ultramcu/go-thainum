package decimaladapter

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestBaht(t *testing.T) {
	cases := []struct{ in, want string }{
		{"21.21", "ยี่สิบเอ็ดบาทยี่สิบเอ็ดสตางค์"},
		{"100", "หนึ่งร้อยบาทถ้วน"},
		{"0", "ศูนย์บาทถ้วน"},
		{"0.25", "ยี่สิบห้าสตางค์"},
	}
	for _, c := range cases {
		d, _ := decimal.NewFromString(c.in)
		if got := Baht(d); got != c.want {
			t.Errorf("Baht(%s) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSpell(t *testing.T) {
	d, _ := decimal.NewFromString("12.34")
	if got := Spell(d); got != "สิบสองจุดสามสี่" {
		t.Errorf("Spell(12.34) = %q", got)
	}
}
