package thainum

import (
	"testing"
	"time"
)

func TestThaiDate(t *testing.T) {
	d := time.Date(2024, time.June, 5, 0, 0, 0, 0, time.UTC) // a Wednesday, BE 2567

	if got := MonthTH(time.June); got != "มิถุนายน" {
		t.Errorf("MonthTH(June) = %q", got)
	}
	if got := MonthAbbrTH(time.June); got != "มิ.ย." {
		t.Errorf("MonthAbbrTH(June) = %q", got)
	}
	if got := WeekdayTH(time.Wednesday); got != "วันพุธ" {
		t.Errorf("WeekdayTH(Wed) = %q", got)
	}
	if got := WeekdayAbbrTH(time.Wednesday); got != "พ." {
		t.Errorf("WeekdayAbbrTH(Wed) = %q", got)
	}
	if got := BuddhistYear(d); got != 2567 {
		t.Errorf("BuddhistYear = %d, want 2567", got)
	}
	if got := FormatDate(d); got != "5 มิถุนายน 2567" {
		t.Errorf("FormatDate = %q", got)
	}
	if got := FormatDateAbbr(d); got != "5 มิ.ย. 2567" {
		t.Errorf("FormatDateAbbr = %q", got)
	}
	if got := FormatDateFull(d); got != "วันพุธที่ 5 มิถุนายน พ.ศ. 2567" {
		t.Errorf("FormatDateFull = %q", got)
	}
}

func TestParseDate(t *testing.T) {
	want := time.Date(2024, time.June, 5, 0, 0, 0, 0, time.UTC)
	inputs := []string{
		"5 มิถุนายน 2567",
		"5 มิ.ย. 2567",
		"วันพุธที่ 5 มิถุนายน พ.ศ. 2567",
		"๕ มิถุนายน ๒๕๖๗", // Thai digits
	}
	for _, in := range inputs {
		got, err := ParseDate(in)
		if err != nil {
			t.Errorf("ParseDate(%q) error: %v", in, err)
			continue
		}
		if !got.Equal(want) {
			t.Errorf("ParseDate(%q) = %v, want %v", in, got, want)
		}
	}

	if _, err := ParseDate("ไม่ใช่วันที่"); err == nil {
		t.Error("expected error for non-date input")
	}
	if _, err := ParseDate("31 กุมภาพันธ์ 2567"); err == nil {
		t.Error("expected error for invalid 31 February")
	}
}

func TestDateRoundTrip(t *testing.T) {
	for _, d := range []time.Time{
		time.Date(2024, time.June, 5, 0, 0, 0, 0, time.UTC),
		time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2030, time.December, 31, 0, 0, 0, 0, time.UTC),
	} {
		for _, s := range []string{FormatDate(d), FormatDateAbbr(d), FormatDateFull(d)} {
			got, err := ParseDate(s)
			if err != nil {
				t.Errorf("round-trip ParseDate(%q) error: %v", s, err)
				continue
			}
			if !got.Equal(d) {
				t.Errorf("round-trip %q = %v, want %v", s, got, d)
			}
		}
	}
}

func TestThaiDateAllMonthsWeekdays(t *testing.T) {
	if MonthTH(time.January) != "มกราคม" || MonthTH(time.December) != "ธันวาคม" {
		t.Error("month bounds wrong")
	}
	if WeekdayTH(time.Sunday) != "วันอาทิตย์" || WeekdayTH(time.Saturday) != "วันเสาร์" {
		t.Error("weekday bounds wrong")
	}
	if MonthTH(time.Month(0)) != "" || MonthTH(time.Month(13)) != "" {
		t.Error("out-of-range month should be empty")
	}
}
