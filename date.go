package thainum

import (
	"fmt"
	"sort"
	"strings"
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

// monthLookup maps every Thai month name (full and abbreviated) to its month,
// ordered by descending length so the longest match wins.
type monthName struct {
	name string
	m    time.Month
}

var monthNames = func() []monthName {
	var out []monthName
	for i := time.January; i <= time.December; i++ {
		out = append(out, monthName{thaiMonths[i], i}, monthName{thaiMonthsAbbr[i], i})
	}
	sort.Slice(out, func(a, b int) bool { return len(out[a].name) > len(out[b].name) })
	return out
}()

// ParseDate parses a Thai date string back into a time.Time (at midnight UTC).
// It accepts the forms produced by FormatDate, FormatDateAbbr and
// FormatDateFull — e.g. "5 มิถุนายน 2567", "5 มิ.ย. 2567" and
// "วันพุธที่ 5 มิถุนายน พ.ศ. 2567" — with Arabic or Thai digits. The year is
// interpreted as a Buddhist-Era year (converted to CE with −543).
func ParseDate(s string) (time.Time, error) {
	norm := ToArabicDigits(s)

	var mon time.Month
	for _, mn := range monthNames {
		if strings.Contains(norm, mn.name) {
			mon = mn.m
			break
		}
	}
	if mon == 0 {
		return time.Time{}, fmt.Errorf("thainum: no Thai month found in %q", s)
	}

	nums := digitGroups(norm)
	if len(nums) < 2 {
		return time.Time{}, fmt.Errorf("thainum: need a day and a year in %q", s)
	}
	day := nums[0]
	be := nums[len(nums)-1]
	ce := be - 543

	res := time.Date(ce, mon, day, 0, 0, 0, 0, time.UTC)
	// time.Date normalizes out-of-range values; reject anything that shifted.
	if res.Year() != ce || res.Month() != mon || res.Day() != day {
		return time.Time{}, fmt.Errorf("thainum: invalid date in %q", s)
	}
	return res, nil
}

// digitGroups returns each run of ASCII digits in s as an int.
func digitGroups(s string) []int {
	var out []int
	i := 0
	for i < len(s) {
		if s[i] < '0' || s[i] > '9' {
			i++
			continue
		}
		n := 0
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			n = n*10 + int(s[i]-'0')
			i++
		}
		out = append(out, n)
	}
	return out
}
