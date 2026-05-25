package thainum

import (
	"strings"
	"time"
)

// FormatTime reads the time of day t in the formal 24-hour Thai style
// (นาฬิกา / นาที): 14:30 -> "สิบสี่นาฬิกาสามสิบนาที", 14:00 -> "สิบสี่นาฬิกา".
func FormatTime(t time.Time) string {
	out := Spell(int64(t.Hour())) + "นาฬิกา"
	if m := t.Minute(); m != 0 {
		out += Spell(int64(m)) + "นาที"
	}
	return out
}

// FormatClock reads the time of day t in the colloquial Thai 6-hour style
// (ตี / โมงเช้า / บ่าย / โมงเย็น / ทุ่ม / เที่ยง / เที่ยงคืน). A half hour reads
// "ครึ่ง"; other minutes read "…นาที".
//
//	08:00 -> "แปดโมงเช้า"   13:00 -> "บ่ายโมง"   19:00 -> "หนึ่งทุ่ม"
//	14:30 -> "บ่ายสองโมงครึ่ง"
func FormatClock(t time.Time) string {
	h := t.Hour()
	var base string
	switch {
	case h == 0:
		base = "เที่ยงคืน"
	case h <= 5: // 1–5
		base = "ตี" + Spell(int64(h))
	case h == 6:
		base = "หกโมงเช้า"
	case h <= 11: // 7–11
		base = Spell(int64(h)) + "โมงเช้า"
	case h == 12:
		base = "เที่ยง"
	case h == 13:
		base = "บ่ายโมง"
	case h <= 15: // 14–15
		base = "บ่าย" + Spell(int64(h-12)) + "โมง"
	case h <= 18: // 16–18
		base = Spell(int64(h-12)) + "โมงเย็น"
	default: // 19–23
		base = Spell(int64(h-18)) + "ทุ่ม"
	}
	switch m := t.Minute(); {
	case m == 30:
		return base + "ครึ่ง"
	case m != 0:
		return base + Spell(int64(m)) + "นาที"
	default:
		return base
	}
}

// FormatDuration reads a duration as Thai words, using วัน / ชั่วโมง / นาที /
// วินาที for the non-zero components (sub-second parts are dropped). A zero
// duration reads "ศูนย์วินาที"; a negative one is prefixed with "ลบ".
//
//	90*time.Minute -> "หนึ่งชั่วโมงสามสิบนาที"
//	45*time.Second -> "สี่สิบห้าวินาที"
func FormatDuration(d time.Duration) string {
	neg := d < 0
	if neg {
		d = -d
	}
	total := int64(d / time.Second)
	days, rem := total/86400, total%86400
	hours, rem := rem/3600, rem%3600
	mins, secs := rem/60, rem%60

	var b strings.Builder
	if days > 0 {
		b.WriteString(Spell(days) + "วัน")
	}
	if hours > 0 {
		b.WriteString(Spell(hours) + "ชั่วโมง")
	}
	if mins > 0 {
		b.WriteString(Spell(mins) + "นาที")
	}
	if secs > 0 {
		b.WriteString(Spell(secs) + "วินาที")
	}
	out := b.String()
	if out == "" {
		out = "ศูนย์วินาที"
	}
	if neg {
		out = "ลบ" + out
	}
	return out
}
