package thainum

import (
	"testing"
	"time"
)

func tm(h, m int) time.Time { return time.Date(2024, 1, 1, h, m, 0, 0, time.UTC) }

func TestFormatTime(t *testing.T) {
	cases := []struct {
		h, m int
		want string
	}{
		{14, 30, "สิบสี่นาฬิกาสามสิบนาที"},
		{14, 0, "สิบสี่นาฬิกา"},
		{0, 0, "ศูนย์นาฬิกา"},
		{9, 5, "เก้านาฬิกาห้านาที"},
	}
	for _, c := range cases {
		if got := FormatTime(tm(c.h, c.m)); got != c.want {
			t.Errorf("FormatTime(%02d:%02d) = %q, want %q", c.h, c.m, got, c.want)
		}
	}
}

func TestFormatClock(t *testing.T) {
	cases := []struct {
		h, m int
		want string
	}{
		{0, 0, "เที่ยงคืน"},
		{1, 0, "ตีหนึ่ง"},
		{5, 0, "ตีห้า"},
		{6, 0, "หกโมงเช้า"},
		{7, 0, "เจ็ดโมงเช้า"},
		{11, 0, "สิบเอ็ดโมงเช้า"},
		{12, 0, "เที่ยง"},
		{13, 0, "บ่ายโมง"},
		{14, 0, "บ่ายสองโมง"},
		{15, 0, "บ่ายสามโมง"},
		{16, 0, "สี่โมงเย็น"},
		{18, 0, "หกโมงเย็น"},
		{19, 0, "หนึ่งทุ่ม"},
		{23, 0, "ห้าทุ่ม"},
		{14, 30, "บ่ายสองโมงครึ่ง"},
		{14, 15, "บ่ายสองโมงสิบห้านาที"},
	}
	for _, c := range cases {
		if got := FormatClock(tm(c.h, c.m)); got != c.want {
			t.Errorf("FormatClock(%02d:%02d) = %q, want %q", c.h, c.m, got, c.want)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want string
	}{
		{90 * time.Minute, "หนึ่งชั่วโมงสามสิบนาที"},
		{45 * time.Second, "สี่สิบห้าวินาที"},
		{2 * time.Hour, "สองชั่วโมง"},
		{25 * time.Hour, "หนึ่งวันหนึ่งชั่วโมง"},
		{0, "ศูนย์วินาที"},
		{-30 * time.Minute, "ลบสามสิบนาที"},
	}
	for _, c := range cases {
		if got := FormatDuration(c.d); got != c.want {
			t.Errorf("FormatDuration(%v) = %q, want %q", c.d, got, c.want)
		}
	}
}
