package thainum

import (
	"errors"
	"strings"
	"testing"
)

// All expected Thai strings are taken from the Dart oracle
// (thainum.dart/test/qualifiers_test.dart and lib/src/qualifiers.dart),
// not from running this Go code.

func TestThaiApprox(t *testing.T) {
	cases := []struct {
		n    int64
		want string
	}{
		{100, "ประมาณหนึ่งร้อย"},
		{21, "ประมาณยี่สิบเอ็ด"},
		{0, "ประมาณศูนย์"},
		{-5, "ประมาณลบห้า"},
	}
	for _, c := range cases {
		if got := ThaiApprox(c.n); got != c.want {
			t.Errorf("ThaiApprox(%d) = %q, want %q", c.n, got, c.want)
		}
	}
}

func TestThaiNearly(t *testing.T) {
	cases := []struct {
		n    int64
		want string
	}{
		{100, "เกือบหนึ่งร้อย"},
		{20, "เกือบยี่สิบ"},
		{-5, "เกือบลบห้า"},
	}
	for _, c := range cases {
		if got := ThaiNearly(c.n); got != c.want {
			t.Errorf("ThaiNearly(%d) = %q, want %q", c.n, got, c.want)
		}
	}
}

func TestThaiRange(t *testing.T) {
	cases := []struct {
		a, b int64
		want string
	}{
		{10, 20, "สิบถึงยี่สิบ"},
		{1, 5, "หนึ่งถึงห้า"},
		{-5, 5, "ลบห้าถึงห้า"},
		// no ordering imposed
		{20, 10, "ยี่สิบถึงสิบ"},
		{5, 5, "ห้าถึงห้า"},
	}
	for _, c := range cases {
		if got := ThaiRange(c.a, c.b); got != c.want {
			t.Errorf("ThaiRange(%d, %d) = %q, want %q", c.a, c.b, got, c.want)
		}
	}
}

func TestThaiMoreThanValid(t *testing.T) {
	cases := []struct {
		n    int64
		want string
	}{
		{10, "สิบกว่า"},
		{20, "ยี่สิบกว่า"},
		{90, "เก้าสิบกว่า"},
		{100, "หนึ่งร้อยกว่า"},
		{500, "ห้าร้อยกว่า"},
		{1000, "หนึ่งพันกว่า"},
		{10000, "หนึ่งหมื่นกว่า"},
		{100000, "หนึ่งแสนกว่า"},
		{1000000, "หนึ่งล้านกว่า"},
		{3000000, "สามล้านกว่า"},
	}
	for _, c := range cases {
		got, err := ThaiMoreThan(c.n)
		if err != nil {
			t.Errorf("ThaiMoreThan(%d) unexpected error: %v", c.n, err)
			continue
		}
		if got != c.want {
			t.Errorf("ThaiMoreThan(%d) = %q, want %q", c.n, got, c.want)
		}
	}
}

func TestThaiMoreThanRejects(t *testing.T) {
	// non-round numbers + single digit / zero / negative all reject.
	bad := []int64{11, 15, 19, 21, 150, 999, 1234, 0, 1, 5, 9, -10, -100}
	for _, n := range bad {
		got, err := ThaiMoreThan(n)
		if err == nil {
			t.Errorf("ThaiMoreThan(%d) = %q, want error", n, got)
			continue
		}
		if got != "" {
			t.Errorf("ThaiMoreThan(%d) returned %q on error, want empty", n, got)
		}
		if !errors.Is(err, ErrNotRoundMagnitude) {
			t.Errorf("ThaiMoreThan(%d) error = %v, want wraps ErrNotRoundMagnitude", n, err)
		}
	}
}

func TestThaiMoreThanErrorMentionsKwa(t *testing.T) {
	// Dart oracle: notRoundMagnitude message contains 'กว่า'.
	_, err := ThaiMoreThan(11)
	if err == nil {
		t.Fatal("ThaiMoreThan(11) want error")
	}
	if !strings.Contains(err.Error(), "กว่า") {
		t.Errorf("ThaiMoreThan(11) error %q does not mention กว่า", err.Error())
	}
}

func TestIsRoundMagnitude(t *testing.T) {
	round := []int64{10, 20, 90, 100, 500, 1000, 1000000, 7000000}
	for _, n := range round {
		if !IsRoundMagnitude(n) {
			t.Errorf("IsRoundMagnitude(%d) = false, want true", n)
		}
	}
	notRound := []int64{0, 1, 9, 11, 15, 99, 101, 150, 1234, -10}
	for _, n := range notRound {
		if IsRoundMagnitude(n) {
			t.Errorf("IsRoundMagnitude(%d) = true, want false", n)
		}
	}
}

func TestQualified(t *testing.T) {
	// approx / nearly / moreThan dispatch (Dart toThaiQualified oracle).
	if got, err := Qualified(100, QualifierApprox); err != nil || got != "ประมาณหนึ่งร้อย" {
		t.Errorf("Qualified(100, approx) = %q, %v; want %q, nil", got, err, "ประมาณหนึ่งร้อย")
	}
	if got, err := Qualified(20, QualifierNearly); err != nil || got != "เกือบยี่สิบ" {
		t.Errorf("Qualified(20, nearly) = %q, %v; want %q, nil", got, err, "เกือบยี่สิบ")
	}
	if got, err := Qualified(20, QualifierMoreThan); err != nil || got != "ยี่สิบกว่า" {
		t.Errorf("Qualified(20, moreThan) = %q, %v; want %q, nil", got, err, "ยี่สิบกว่า")
	}
	// moreThan on a non-round value propagates the error.
	if _, err := Qualified(11, QualifierMoreThan); !errors.Is(err, ErrNotRoundMagnitude) {
		t.Errorf("Qualified(11, moreThan) error = %v, want wraps ErrNotRoundMagnitude", err)
	}
	// unknown kind errors.
	if _, err := Qualified(10, QualifierKind(99)); err == nil {
		t.Error("Qualified(10, unknown kind) want error")
	}
}
