package thainum

import "testing"

func TestBahtWholeUnit(t *testing.T) {
	cases := []struct {
		baht int64
		want string
	}{
		{0, "ศูนย์บาทถ้วน"},
		{1, "หนึ่งบาทถ้วน"},
		{21, "ยี่สิบเอ็ดบาทถ้วน"},
		{100, "หนึ่งร้อยบาทถ้วน"},
		{1000000, "หนึ่งล้านบาทถ้วน"},
		{-5, "ลบห้าบาทถ้วน"},
	}
	for _, c := range cases {
		if got := Baht(c.baht); got != c.want {
			t.Errorf("Baht(%d) = %q, want %q", c.baht, got, c.want)
		}
	}
}

func TestBahtSatang(t *testing.T) {
	cases := []struct {
		satang int64
		want   string
	}{
		{0, "ศูนย์บาทถ้วน"},
		{100, "หนึ่งบาทถ้วน"},
		{200, "สองบาทถ้วน"},
		{1000, "สิบบาทถ้วน"},
		{1100, "สิบเอ็ดบาทถ้วน"},
		{2100, "ยี่สิบเอ็ดบาทถ้วน"},
		{2500, "ยี่สิบห้าบาทถ้วน"},
		{10000, "หนึ่งร้อยบาทถ้วน"},
		{10050, "หนึ่งร้อยบาทห้าสิบสตางค์"},
		{10100, "หนึ่งร้อยเอ็ดบาทถ้วน"},
		{2121, "ยี่สิบเอ็ดบาทยี่สิบเอ็ดสตางค์"},
		{100000, "หนึ่งพันบาทถ้วน"},
		{100000000, "หนึ่งล้านบาทถ้วน"},
		{100000075, "หนึ่งล้านบาทเจ็ดสิบห้าสตางค์"},
		{1000000000, "สิบล้านบาทถ้วน"},
		{10000000000, "หนึ่งร้อยล้านบาทถ้วน"},
		{1, "หนึ่งสตางค์"},
		{25, "ยี่สิบห้าสตางค์"},
		{50, "ห้าสิบสตางค์"},
		{-500, "ลบห้าบาทถ้วน"},
		{-25, "ลบยี่สิบห้าสตางค์"},
		{-101, "ลบหนึ่งบาทหนึ่งสตางค์"},
	}
	for _, c := range cases {
		if got := BahtSatang(c.satang); got != c.want {
			t.Errorf("BahtSatang(%d) = %q, want %q", c.satang, got, c.want)
		}
	}
}

func TestBahtFromString(t *testing.T) {
	cases := []struct{ in, want string }{
		{"0.00", "ศูนย์บาทถ้วน"},
		{"1.00", "หนึ่งบาทถ้วน"},
		{"21.21", "ยี่สิบเอ็ดบาทยี่สิบเอ็ดสตางค์"},
		{"100.50", "หนึ่งร้อยบาทห้าสิบสตางค์"},
		{"1000000.75", "หนึ่งล้านบาทเจ็ดสิบห้าสตางค์"},
		{"0.01", "หนึ่งสตางค์"},
		{"0.25", "ยี่สิบห้าสตางค์"},
		{"1.005", "หนึ่งบาทหนึ่งสตางค์"}, // away-from-zero rounding of the 3rd digit
		{"-5", "ลบห้าบาทถ้วน"},
	}
	for _, c := range cases {
		got, err := BahtFromString(c.in)
		if err != nil {
			t.Errorf("BahtFromString(%q) error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("BahtFromString(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFormat(t *testing.T) {
	if got := FormatInt(1234567); got != "1,234,567" {
		t.Errorf("FormatInt = %q", got)
	}
	if got := FormatSatang(100000); got != "1,000.00" {
		t.Errorf("FormatSatang = %q", got)
	}
	if got := FormatTHB(2121); got != "฿21.21" {
		t.Errorf("FormatTHB = %q", got)
	}
	if got := FormatInt(-50); got != "-50" {
		t.Errorf("FormatInt neg = %q", got)
	}
}
