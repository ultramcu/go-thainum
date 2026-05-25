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
