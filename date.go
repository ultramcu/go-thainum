package thainum

import (
	"fmt"
	"time"
)

// thaiMonths is 1-indexed (thaiMonths[time.January] == "มกราคม").
var thaiMonths = [...]string{
	"", "มกราคม", "กุมภาพันธ์", "มีนาคม", "เมษายน", "พฤษภาคม", "มิถุนายน",
	"กรกฎาคม", "สิงหาคม", "กันยายน", "ตุลาคม", "พฤศจิกายน", "ธันวาคม",
}

// thaiMonthsAbbr holds the standard abbreviated Thai month names, 1-indexed.
var thaiMonthsAbbr = [...]string{
	"", "ม.ค.", "ก.พ.", "มี.ค.", "เม.ย.", "พ.ค.", "มิ.ย.",
	"ก.ค.", "ส.ค.", "ก.ย.", "ต.ค.", "พ.ย.", "ธ.ค.",
}

// thaiWeekdays is indexed by time.Weekday (0 = Sunday).
var thaiWeekdays = [...]string{"อาทิตย์", "จันทร์", "อังคาร", "พุธ", "พฤหัสบดี", "ศุกร์", "เสาร์"}

var thaiWeekdaysAbbr = [...]string{"อา.", "จ.", "อ.", "พ.", "พฤ.", "ศ.", "ส."}

// MonthTH returns the full Thai month name (มกราคม … ธันวาคม), or "" if m is out
// of range.
func MonthTH(m time.Month) string {
	if m < time.January || m > time.December {
		return ""
	}
	return thaiMonths[m]
}

// MonthAbbrTH returns the abbreviated Thai month name (ม.ค. … ธ.ค.), or "" if m
// is out of range.
func MonthAbbrTH(m time.Month) string {
	if m < time.January || m > time.December {
		return ""
	}
	return thaiMonthsAbbr[m]
}

// WeekdayTH returns the full Thai weekday name with the "วัน" prefix
// (วันอาทิตย์ … วันเสาร์).
func WeekdayTH(d time.Weekday) string {
	if d < time.Sunday || d > time.Saturday {
		return ""
	}
	return "วัน" + thaiWeekdays[d]
}

// WeekdayAbbrTH returns the abbreviated Thai weekday name (อา. จ. อ. พ. พฤ. ศ. ส.).
func WeekdayAbbrTH(d time.Weekday) string {
	if d < time.Sunday || d > time.Saturday {
		return ""
	}
	return thaiWeekdaysAbbr[d]
}

// BuddhistYear returns the Buddhist-Era year of t (Gregorian year + 543).
func BuddhistYear(t time.Time) int { return t.Year() + 543 }

// FormatDate formats t as a Thai date with the Buddhist-Era year and full month:
// "5 มิถุนายน 2567".
func FormatDate(t time.Time) string {
	return fmt.Sprintf("%d %s %d", t.Day(), MonthTH(t.Month()), BuddhistYear(t))
}

// FormatDateAbbr formats t with an abbreviated month: "5 มิ.ย. 2567".
func FormatDateAbbr(t time.Time) string {
	return fmt.Sprintf("%d %s %d", t.Day(), MonthAbbrTH(t.Month()), BuddhistYear(t))
}

// FormatDateFull formats t with the weekday and a "พ.ศ." label:
// "วันพุธที่ 5 มิถุนายน พ.ศ. 2567".
func FormatDateFull(t time.Time) string {
	return fmt.Sprintf("%sที่ %d %s พ.ศ. %d",
		WeekdayTH(t.Weekday()), t.Day(), MonthTH(t.Month()), BuddhistYear(t))
}
