package thainum

import "testing"

func TestOrdinal(t *testing.T) {
	cases := []struct {
		n    int64
		want string
	}{
		{1, "ที่หนึ่ง"},
		{2, "ที่สอง"},
		{21, "ที่ยี่สิบเอ็ด"},
		{100, "ที่หนึ่งร้อย"},
	}
	for _, c := range cases {
		if got := Ordinal(c.n); got != c.want {
			t.Errorf("Ordinal(%d) = %q, want %q", c.n, got, c.want)
		}
	}
}

func TestFraction(t *testing.T) {
	cases := []struct {
		num, den int64
		want     string
	}{
		{1, 2, "เศษหนึ่งส่วนสอง"},
		{3, 4, "เศษสามส่วนสี่"},
		{21, 100, "เศษยี่สิบเอ็ดส่วนหนึ่งร้อย"},
	}
	for _, c := range cases {
		if got := Fraction(c.num, c.den); got != c.want {
			t.Errorf("Fraction(%d,%d) = %q, want %q", c.num, c.den, got, c.want)
		}
	}
}

func TestYearAndEra(t *testing.T) {
	if got := Year(2566); got != "พุทธศักราชสองพันห้าร้อยหกสิบหก" {
		t.Errorf("Year(2566) = %q", got)
	}
	if got := CEToBE(2023); got != 2566 {
		t.Errorf("CEToBE(2023) = %d, want 2566", got)
	}
	if got := BEToCE(2566); got != 2023 {
		t.Errorf("BEToCE(2566) = %d, want 2023", got)
	}
}

func TestSatangFromFloat(t *testing.T) {
	cases := []struct {
		in   float64
		want int64
	}{
		{21.21, 2121},
		{0.5, 50},
		{100, 10000},
		{0.25, 25},
		{-5, -500},
	}
	for _, c := range cases {
		if got := SatangFromFloat(c.in); got != c.want {
			t.Errorf("SatangFromFloat(%v) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestBahtFromFloat(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{100, "หนึ่งร้อยบาทถ้วน"},
		{21.21, "ยี่สิบเอ็ดบาทยี่สิบเอ็ดสตางค์"},
		{0.25, "ยี่สิบห้าสตางค์"},
		{-5, "ลบห้าบาทถ้วน"},
	}
	for _, c := range cases {
		if got := BahtFromFloat(c.in); got != c.want {
			t.Errorf("BahtFromFloat(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}
