package thainum

import (
	"fmt"
	"time"
)

// requireDigits normalises Thai numerals in s to Arabic, then verifies the
// result is composed of exactly expected ASCII digits and nothing else. Any
// non-digit rune (a separator, letter, sign, anything) is rejected, mirroring
// the Dart _requireDigits helper: a lottery number is a fixed-width run of bare
// digits, not a delimited sequence. It returns the digit-only string or an
// error wrapping ErrParse.
func requireDigits(s string, expected int, what string) (string, error) {
	arabic := ToArabicDigits(s)
	digits := make([]rune, 0, expected)
	for _, r := range arabic {
		if r >= '0' && r <= '9' {
			digits = append(digits, r)
			continue
		}
		return "", fmt.Errorf("%w: %s must be exactly %d digits", ErrParse, what, expected)
	}
	if len(digits) != expected {
		return "", fmt.Errorf("%w: %s must be exactly %d digits (got %d)", ErrParse, what, expected, len(digits))
	}
	return string(digits), nil
}

// SpeakLotteryNumber reads a six-digit Thai lottery prize number digit-by-digit
// in Thai words (อ่านเรียงตัว) — the way a lottery number is read aloud, never
// as a quantity.
//
//	SpeakLotteryNumber("123456") // "หนึ่ง สอง สาม สี่ ห้า หก", nil
//	SpeakLotteryNumber("๐๑๒๓๔๕") // "ศูนย์ หนึ่ง สอง สาม สี่ ห้า", nil
//
// Both Arabic (0-9) and Thai (๐-๙) digits are accepted. The reading uses a
// single space separator and the literal สอง for the digit 2; use
// SpeakLotteryNumberWith for a custom separator or the colloquial โท. It returns
// an error wrapping ErrParse (test with errors.Is(err, ErrParse)) when the input
// is not exactly six digits — any non-digit character is rejected.
func SpeakLotteryNumber(sixDigits string) (string, error) {
	return SpeakLotteryNumberWith(sixDigits, DigitSpeaker{})
}

// SpeakLotteryNumberWith is SpeakLotteryNumber with an explicit DigitSpeaker so
// callers can choose the separator and colloquial-two reading.
//
//	SpeakLotteryNumberWith("222222", DigitSpeaker{ColloquialTwo: true})
//	// "โท โท โท โท โท โท", nil
//	SpeakLotteryNumberWith("123456", DigitSpeaker{Separator: "-", SeparatorSet: true})
//	// "หนึ่ง-สอง-สาม-สี่-ห้า-หก", nil
func SpeakLotteryNumberWith(sixDigits string, ds DigitSpeaker) (string, error) {
	d, err := requireDigits(sixDigits, 6, "a lottery number")
	if err != nil {
		return "", err
	}
	return ds.Speak(d)
}

// SpeakTwoDigit reads a two-digit running number (เลขท้าย 2 ตัว) digit-by-digit.
//
//	SpeakTwoDigit("07") // "ศูนย์ เจ็ด", nil
//
// See SpeakLotteryNumber for the digit handling; it returns an error wrapping
// ErrParse when the input is not exactly two digits.
func SpeakTwoDigit(twoDigits string) (string, error) {
	return SpeakTwoDigitWith(twoDigits, DigitSpeaker{})
}

// SpeakTwoDigitWith is SpeakTwoDigit with an explicit DigitSpeaker.
//
//	SpeakTwoDigitWith("20", DigitSpeaker{ColloquialTwo: true}) // "โท ศูนย์", nil
func SpeakTwoDigitWith(twoDigits string, ds DigitSpeaker) (string, error) {
	d, err := requireDigits(twoDigits, 2, "a two-digit number")
	if err != nil {
		return "", err
	}
	return ds.Speak(d)
}

// SpeakThreeDigit reads a three-digit running number (เลขท้าย/เลขหน้า 3 ตัว)
// digit-by-digit.
//
//	SpeakThreeDigit("507") // "ห้า ศูนย์ เจ็ด", nil
//
// See SpeakLotteryNumber for the digit handling; it returns an error wrapping
// ErrParse when the input is not exactly three digits.
func SpeakThreeDigit(threeDigits string) (string, error) {
	return SpeakThreeDigitWith(threeDigits, DigitSpeaker{})
}

// SpeakThreeDigitWith is SpeakThreeDigit with an explicit DigitSpeaker.
//
//	SpeakThreeDigitWith("234", DigitSpeaker{ColloquialTwo: true}) // "โท สาม สี่", nil
func SpeakThreeDigitWith(threeDigits string, ds DigitSpeaker) (string, error) {
	d, err := requireDigits(threeDigits, 3, "a three-digit number")
	if err != nil {
		return "", err
	}
	return ds.Speak(d)
}

// IsLotteryDrawDate reports whether t falls on a Thai Government Lottery draw
// date.
//
// The Thai Government Lottery (สลากกินแบ่งรัฐบาล) is drawn twice a month, on the
// 1st and the 16th. This checks the calendar day only — it does not know about
// the rare official schedule shifts (e.g. when a draw is moved off a public
// holiday), so treat it as a convenience predicate, not an authority.
//
//	IsLotteryDrawDate(time.Date(2024, 6, 1, 0, 0, 0, 0, time.Local))  // true
//	IsLotteryDrawDate(time.Date(2024, 6, 16, 0, 0, 0, 0, time.Local)) // true
//	IsLotteryDrawDate(time.Date(2024, 6, 17, 0, 0, 0, 0, time.Local)) // false
func IsLotteryDrawDate(t time.Time) bool {
	d := t.Day()
	return d == 1 || d == 16
}

// LotteryDrawDates returns the two regular Thai Government Lottery draw dates in
// the given month of year — the 1st and the 16th — as time.Time values at
// midnight local time (like time.Date with hour/min/sec/nsec all zero).
//
//	LotteryDrawDates(2024, time.June)
//	// [2024-06-01 00:00:00, 2024-06-16 00:00:00], nil
//
// As with IsLotteryDrawDate this is the regular schedule and does not encode
// official holiday shifts. It returns an error wrapping ErrParse when month is
// not in the range 1..12 (mirroring the Dart helper, which would otherwise be
// silently normalised by Go's calendar arithmetic).
func LotteryDrawDates(year int, month time.Month) ([]time.Time, error) {
	if month < time.January || month > time.December {
		return nil, fmt.Errorf("%w: month must be 1..12 (got %d)", ErrParse, int(month))
	}
	return []time.Time{
		time.Date(year, month, 1, 0, 0, 0, 0, time.Local),
		time.Date(year, month, 16, 0, 0, 0, 0, time.Local),
	}, nil
}
