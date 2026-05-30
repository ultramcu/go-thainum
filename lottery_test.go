package thainum

import (
	"errors"
	"testing"
	"time"
)

// All expected values below are taken from the Dart oracle
// (thainum.dart/lib/src/lottery.dart and test/lottery_test.dart), not from
// running the Go implementation.

func TestSpeakLotteryNumber(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"six digits digit-by-digit", "123456", "หนึ่ง สอง สาม สี่ ห้า หก"},
		{"leading zero", "012345", "ศูนย์ หนึ่ง สอง สาม สี่ ห้า"},
		{"thai numerals", "๐๑๒๓๔๕", "ศูนย์ หนึ่ง สอง สาม สี่ ห้า"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := SpeakLotteryNumber(c.in)
			if err != nil {
				t.Fatalf("SpeakLotteryNumber(%q) error: %v", c.in, err)
			}
			if got != c.want {
				t.Errorf("SpeakLotteryNumber(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestSpeakLotteryNumberWith(t *testing.T) {
	t.Run("colloquialTwo reads 2 as โท", func(t *testing.T) {
		got, err := SpeakLotteryNumberWith("222222", DigitSpeaker{ColloquialTwo: true})
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if want := "โท โท โท โท โท โท"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
	t.Run("custom separator", func(t *testing.T) {
		got, err := SpeakLotteryNumberWith("123456", DigitSpeaker{Separator: "-", SeparatorSet: true})
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if want := "หนึ่ง-สอง-สาม-สี่-ห้า-หก"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestSpeakLotteryNumberErrors(t *testing.T) {
	// Dart throws on wrong length and on any non-digit character.
	bad := []string{
		"12345",   // too short
		"1234567", // too long
		"12-456",  // non-digit (separator) — must be rejected, not collapsed
		"abcdef",  // non-digit letters
	}
	for _, in := range bad {
		got, err := SpeakLotteryNumber(in)
		if err == nil {
			t.Errorf("SpeakLotteryNumber(%q) = %q, want error", in, got)
			continue
		}
		if !errors.Is(err, ErrParse) {
			t.Errorf("SpeakLotteryNumber(%q) error %v does not wrap ErrParse", in, err)
		}
	}
}

func TestSpeakTwoDigit(t *testing.T) {
	cases := []struct{ in, want string }{
		{"07", "ศูนย์ เจ็ด"},
		{"๒๕", "สอง ห้า"},
	}
	for _, c := range cases {
		got, err := SpeakTwoDigit(c.in)
		if err != nil {
			t.Fatalf("SpeakTwoDigit(%q) error: %v", c.in, err)
		}
		if got != c.want {
			t.Errorf("SpeakTwoDigit(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSpeakThreeDigit(t *testing.T) {
	got, err := SpeakThreeDigit("507")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if want := "ห้า ศูนย์ เจ็ด"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSpeakDigitColloquialTwo(t *testing.T) {
	got2, err := SpeakTwoDigitWith("20", DigitSpeaker{ColloquialTwo: true})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if want := "โท ศูนย์"; got2 != want {
		t.Errorf("SpeakTwoDigitWith = %q, want %q", got2, want)
	}
	got3, err := SpeakThreeDigitWith("234", DigitSpeaker{ColloquialTwo: true})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if want := "โท สาม สี่"; got3 != want {
		t.Errorf("SpeakThreeDigitWith = %q, want %q", got3, want)
	}
}

func TestSpeakTwoThreeDigitLengthValidation(t *testing.T) {
	if _, err := SpeakTwoDigit("5"); !errors.Is(err, ErrParse) {
		t.Errorf("SpeakTwoDigit(\"5\") err = %v, want ErrParse", err)
	}
	if _, err := SpeakTwoDigit("123"); !errors.Is(err, ErrParse) {
		t.Errorf("SpeakTwoDigit(\"123\") err = %v, want ErrParse", err)
	}
	if _, err := SpeakThreeDigit("12"); !errors.Is(err, ErrParse) {
		t.Errorf("SpeakThreeDigit(\"12\") err = %v, want ErrParse", err)
	}
	if _, err := SpeakThreeDigit("1234"); !errors.Is(err, ErrParse) {
		t.Errorf("SpeakThreeDigit(\"1234\") err = %v, want ErrParse", err)
	}
}

func TestIsLotteryDrawDate(t *testing.T) {
	tt := func(y int, m time.Month, d int) time.Time {
		return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	}
	// true on the 1st and 16th
	for _, d := range []int{1, 16} {
		if !IsLotteryDrawDate(tt(2024, time.June, d)) {
			t.Errorf("IsLotteryDrawDate(2024-06-%02d) = false, want true", d)
		}
	}
	// false otherwise
	for _, d := range []int{2, 15, 30} {
		if IsLotteryDrawDate(tt(2024, time.June, d)) {
			t.Errorf("IsLotteryDrawDate(2024-06-%02d) = true, want false", d)
		}
	}
}

func TestLotteryDrawDates(t *testing.T) {
	got, err := LotteryDrawDates(2024, time.June)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	want := []time.Time{
		time.Date(2024, time.June, 1, 0, 0, 0, 0, time.Local),
		time.Date(2024, time.June, 16, 0, 0, 0, 0, time.Local),
	}
	if len(got) != len(want) {
		t.Fatalf("got %d dates, want %d", len(got), len(want))
	}
	for i := range want {
		if !got[i].Equal(want[i]) {
			t.Errorf("date[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestLotteryDrawDatesOutOfRangeMonth(t *testing.T) {
	// Dart throws for month 0 and 13.
	if _, err := LotteryDrawDates(2024, time.Month(0)); !errors.Is(err, ErrParse) {
		t.Errorf("LotteryDrawDates(2024, 0) err = %v, want ErrParse", err)
	}
	if _, err := LotteryDrawDates(2024, time.Month(13)); !errors.Is(err, ErrParse) {
		t.Errorf("LotteryDrawDates(2024, 13) err = %v, want ErrParse", err)
	}
}
