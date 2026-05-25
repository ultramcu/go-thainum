package thainum

import "strings"

// Thai digits ๐..๙ occupy the contiguous Unicode range U+0E50..U+0E59.
const (
	thaiZero = '๐' // U+0E50
	thaiNine = '๙' // U+0E59
)

// ToThaiDigits replaces every ASCII digit 0-9 in s with its Thai numeral
// (๐-๙), leaving all other runes untouched.
//
//	ToThaiDigits("Room 101")  // "Room ๑๐๑"
func ToThaiDigits(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return thaiZero + (r - '0')
		}
		return r
	}, s)
}

// ToArabicDigits replaces every Thai numeral ๐-๙ in s with its ASCII digit
// 0-9, leaving all other runes untouched.
//
//	ToArabicDigits("ห้อง ๑๐๑")  // "ห้อง 101"
func ToArabicDigits(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= thaiZero && r <= thaiNine {
			return '0' + (r - thaiZero)
		}
		return r
	}, s)
}
