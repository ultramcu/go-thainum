// parity_parse_test.go — BLIND TESTER A (Dev Rabbit).
// Authored from the public API contract + the Dart oracle (parse.dart,
// spell.dart) only; the new Go files were NOT read.
package thainum

import (
	"errors"
	"math/big"
	"testing"
)

func TestExtractNumbers_CompoundIsOneMatch(t *testing.T) {
	ms := ExtractNumbers("ยี่สิบเอ็ด")
	if len(ms) != 1 {
		t.Fatalf("len = %d, want 1; matches = %+v", len(ms), ms)
	}
	if ms[0].Value == nil || ms[0].Value.Cmp(big.NewInt(21)) != 0 {
		t.Errorf("Value = %v, want 21", ms[0].Value)
	}
	if !ms[0].IsWord {
		t.Errorf("IsWord = false, want true")
	}
	if ms[0].IsDigits {
		t.Errorf("IsDigits = true, want false")
	}
	if ms[0].Matched != "ยี่สิบเอ็ด" {
		t.Errorf("Matched = %q, want %q", ms[0].Matched, "ยี่สิบเอ็ด")
	}
}

func TestExtractNumbers_DigitRun(t *testing.T) {
	ms := ExtractNumbers("123")
	if len(ms) != 1 {
		t.Fatalf("len = %d, want 1; matches = %+v", len(ms), ms)
	}
	if ms[0].Value == nil || ms[0].Value.Cmp(big.NewInt(123)) != 0 {
		t.Errorf("Value = %v, want 123", ms[0].Value)
	}
	if !ms[0].IsDigits || ms[0].IsWord {
		t.Errorf("flags IsDigits=%v IsWord=%v want true/false", ms[0].IsDigits, ms[0].IsWord)
	}
	if ms[0].Matched != "123" {
		t.Errorf("Matched = %q, want %q", ms[0].Matched, "123")
	}
}

func TestExtractNumbers_MixedSentence(t *testing.T) {
	text := "บ้าน 21 หลัง ราคา ห้าร้อย บาท"
	ms := ExtractNumbers(text)
	if len(ms) != 2 {
		t.Fatalf("len = %d, want 2; matches = %+v", len(ms), ms)
	}
	if ms[0].Value == nil || ms[0].Value.Cmp(big.NewInt(21)) != 0 || !ms[0].IsDigits {
		t.Errorf("ms[0] = %+v, want 21 digits", ms[0])
	}
	if ms[1].Value == nil || ms[1].Value.Cmp(big.NewInt(500)) != 0 || !ms[1].IsWord {
		t.Errorf("ms[1] = %+v, want 500 word", ms[1])
	}
	for i, m := range ms {
		if m.Start < 0 || m.End > len(text) || m.Start > m.End {
			t.Fatalf("ms[%d] offsets out of range: %d:%d len=%d", i, m.Start, m.End, len(text))
		}
		if got := text[m.Start:m.End]; got != m.Matched {
			t.Errorf("ms[%d]: text[%d:%d]=%q want Matched %q", i, m.Start, m.End, got, m.Matched)
		}
	}
}

func TestExtractNumbers_MaximalMunchSplit(t *testing.T) {
	text := "ห้าร้อยสิบสิบ"
	ms := ExtractNumbers(text)
	if len(ms) != 2 {
		t.Fatalf("len = %d, want 2; matches = %+v", len(ms), ms)
	}
	if ms[0].Value == nil || ms[0].Value.Cmp(big.NewInt(510)) != 0 {
		t.Errorf("ms[0].Value = %v, want 510", ms[0].Value)
	}
	if ms[1].Value == nil || ms[1].Value.Cmp(big.NewInt(10)) != 0 {
		t.Errorf("ms[1].Value = %v, want 10", ms[1].Value)
	}
	for i, m := range ms {
		if got := text[m.Start:m.End]; got != m.Matched {
			t.Errorf("ms[%d]: slice-back %q != %q", i, got, m.Matched)
		}
	}
}

func TestExtractNumbers_NegIsConnector(t *testing.T) {
	ms := ExtractNumbers("ลบห้า")
	if len(ms) != 1 {
		t.Fatalf("len = %d, want 1; matches = %+v", len(ms), ms)
	}
	if ms[0].Value == nil || ms[0].Value.Cmp(big.NewInt(5)) != 0 || ms[0].Value.Sign() < 0 {
		t.Errorf("Value = %v, want non-negative 5", ms[0].Value)
	}
	if ms[0].Matched != "ห้า" {
		t.Errorf("Matched = %q, want %q", ms[0].Matched, "ห้า")
	}
}

func TestExtractNumbers_NoNumbersEmpty(t *testing.T) {
	if ms := ExtractNumbers("ไม่มีเลข"); len(ms) != 0 {
		t.Errorf("got %+v, want empty", ms)
	}
	if ms := ExtractNumbers(""); len(ms) != 0 {
		t.Errorf("empty input got %+v, want empty", ms)
	}
}

func TestIsDigits(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"123", true}, {"", false}, {"12a", false}, {"๑๒", false},
	}
	for _, c := range cases {
		if got := IsDigits(c.in); got != c.want {
			t.Errorf("IsDigits(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestParseDecimal_RoundTripsSpellDecimal(t *testing.T) {
	for _, x := range []string{"1.5", "0.05", "123.45", "12.34", "0.5"} {
		words, err := SpellDecimal(x)
		if err != nil {
			t.Fatalf("SpellDecimal(%q) error: %v", x, err)
		}
		got, err := ParseDecimal(words)
		if err != nil {
			t.Errorf("ParseDecimal(%q) error: %v", words, err)
			continue
		}
		if got != x {
			t.Errorf("round-trip(%q): ParseDecimal(%q) = %q, want %q", x, words, got, x)
		}
	}
}

func TestParseDecimal_MalformedWrapsErrParse(t *testing.T) {
	for _, s := range []string{"จุดห้า", "หนึ่งจุดสองจุดสาม", ""} {
		_, err := ParseDecimal(s)
		if err == nil {
			t.Errorf("ParseDecimal(%q) nil err; want error", s)
			continue
		}
		if !errors.Is(err, ErrParse) {
			t.Errorf("ParseDecimal(%q) err = %v; want wraps ErrParse", s, err)
		}
	}
}

func TestStrict_RejectsNonStandardTens(t *testing.T) {
	if v, err := ParseInt("สองสิบ", Strict()); err == nil {
		t.Errorf("ParseInt(สองสิบ, Strict()) = %d, want error", v)
	}
	if v, err := ParseInt("ยี่สิบ", Strict()); err != nil || v != 20 {
		t.Errorf("ParseInt(ยี่สิบ, Strict()) = (%d,%v) want (20,nil)", v, err)
	}
	if v, err := ParseInt("สิบ", Strict()); err != nil || v != 10 {
		t.Errorf("ParseInt(สิบ, Strict()) = (%d,%v) want (10,nil)", v, err)
	}
}

func TestAllowColloquial(t *testing.T) {
	if v, err := ParseInt("นึง", AllowColloquial()); err != nil || v != 1 {
		t.Errorf("ParseInt(นึง, AllowColloquial()) = (%d,%v) want (1,nil)", v, err)
	}
	if v, err := ParseInt("นึง"); err == nil {
		t.Errorf("ParseInt(นึง) without opt = %d, want error", v)
	}
}

func TestParseInt_BackwardCompatNoOpts(t *testing.T) {
	if v, err := ParseInt("ยี่สิบเอ็ด"); err != nil || v != 21 {
		t.Errorf("ParseInt(ยี่สิบเอ็ด) = (%d,%v) want (21,nil)", v, err)
	}
	if v, err := ParseInt("หนึ่งล้าน"); err != nil || v != 1000000 {
		t.Errorf("ParseInt(หนึ่งล้าน) = (%d,%v) want (1000000,nil)", v, err)
	}
}
