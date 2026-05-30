package thainum

import (
	"math/big"
	"testing"
)

// Tests for ExtractNumbers, ported from the Dart 0.5.3 oracle
// (thainum.dart/test/extract_test.dart). Expected values, offsets and the
// maximal-munch boundary rule are derived from the oracle, not from the Go
// implementation. Offsets are byte offsets per the issue #11 API contract; the
// invariant text[Start:End] == Matched is asserted in every case.

func bigI(i int64) *big.Int { return big.NewInt(i) }

func TestExtractNumbers_MixedDigitsAndWords(t *testing.T) {
	const text = "ซื้อมา ๓ ชิ้น ราคาห้าร้อยบาท"
	ms := ExtractNumbers(text)
	if len(ms) != 2 {
		t.Fatalf("len = %d, want 2", len(ms))
	}

	// First match: the Thai digit ๓ → 3 (isDigits).
	if got := ms[0].Value; got.Cmp(bigI(3)) != 0 {
		t.Errorf("ms[0].Value = %v, want 3", got)
	}
	if !ms[0].IsDigits {
		t.Errorf("ms[0].IsDigits = false, want true")
	}
	if ms[0].IsWord {
		t.Errorf("ms[0].IsWord = true, want false")
	}
	if ms[0].Matched != "๓" {
		t.Errorf("ms[0].Matched = %q, want %q", ms[0].Matched, "๓")
	}
	if ms[0].Start != 19 || ms[0].End != 22 {
		t.Errorf("ms[0] offsets = %d..%d, want 19..22", ms[0].Start, ms[0].End)
	}
	if sub := text[ms[0].Start:ms[0].End]; sub != ms[0].Matched {
		t.Errorf("text[%d:%d] = %q, want %q", ms[0].Start, ms[0].End, sub, ms[0].Matched)
	}

	// Second match: ห้าร้อย → 500 (isWord).
	if got := ms[1].Value; got.Cmp(bigI(500)) != 0 {
		t.Errorf("ms[1].Value = %v, want 500", got)
	}
	if !ms[1].IsWord {
		t.Errorf("ms[1].IsWord = false, want true")
	}
	if ms[1].IsDigits {
		t.Errorf("ms[1].IsDigits = true, want false")
	}
	if ms[1].Matched != "ห้าร้อย" {
		t.Errorf("ms[1].Matched = %q, want %q", ms[1].Matched, "ห้าร้อย")
	}
	if ms[1].Start != 48 || ms[1].End != 69 {
		t.Errorf("ms[1] offsets = %d..%d, want 48..69", ms[1].Start, ms[1].End)
	}
	if sub := text[ms[1].Start:ms[1].End]; sub != ms[1].Matched {
		t.Errorf("text[%d:%d] = %q, want %q", ms[1].Start, ms[1].End, sub, ms[1].Matched)
	}
}

func TestExtractNumbers_CompoundWordIsOneMatch(t *testing.T) {
	ms := ExtractNumbers("ยี่สิบเอ็ด")
	if len(ms) != 1 {
		t.Fatalf("len = %d, want 1", len(ms))
	}
	if got := ms[0].Value; got.Cmp(bigI(21)) != 0 {
		t.Errorf("Value = %v, want 21", got)
	}
	if ms[0].Matched != "ยี่สิบเอ็ด" {
		t.Errorf("Matched = %q, want %q", ms[0].Matched, "ยี่สิบเอ็ด")
	}
}

func TestExtractNumbers_ArabicAndThaiDigitGroups(t *testing.T) {
	const text = "เลข 101 และ ๒๐๒"
	ms := ExtractNumbers(text)
	if len(ms) != 2 {
		t.Fatalf("len = %d, want 2", len(ms))
	}
	wantVals := []int64{101, 202}
	for i, m := range ms {
		if m.Value.Cmp(bigI(wantVals[i])) != 0 {
			t.Errorf("ms[%d].Value = %v, want %d", i, m.Value, wantVals[i])
		}
		if !m.IsDigits {
			t.Errorf("ms[%d].IsDigits = false, want true", i)
		}
		if sub := text[m.Start:m.End]; sub != m.Matched {
			t.Errorf("ms[%d] text[%d:%d] = %q, want %q", i, m.Start, m.End, sub, m.Matched)
		}
	}
	if ms[0].Start != 10 || ms[0].End != 13 {
		t.Errorf("ms[0] offsets = %d..%d, want 10..13", ms[0].Start, ms[0].End)
	}
	if ms[1].Start != 24 || ms[1].End != 33 {
		t.Errorf("ms[1] offsets = %d..%d, want 24..33", ms[1].Start, ms[1].End)
	}
}

func TestExtractNumbers_NoNumbers(t *testing.T) {
	if ms := ExtractNumbers("ไม่มีเลข"); len(ms) != 0 {
		t.Errorf("ExtractNumbers(non-number) len = %d, want 0", len(ms))
	}
	if ms := ExtractNumbers(""); len(ms) != 0 {
		t.Errorf("ExtractNumbers(empty) len = %d, want 0", len(ms))
	}
}

func TestExtractNumbers_MaximalMunch(t *testing.T) {
	// 'ห้าร้อยสิบสิบ' → 'ห้าร้อยสิบ' (510) + 'สิบ' (10): the second สิบ makes the
	// run invalid (repeated place), so the longest valid prefix is emitted and
	// scanning resumes after it.
	const text = "ห้าร้อยสิบสิบ"
	ms := ExtractNumbers(text)
	if len(ms) != 2 {
		t.Fatalf("len = %d, want 2", len(ms))
	}
	if got := ms[0].Value; got.Cmp(bigI(510)) != 0 {
		t.Errorf("ms[0].Value = %v, want 510", got)
	}
	if ms[0].Matched != "ห้าร้อยสิบ" {
		t.Errorf("ms[0].Matched = %q, want %q", ms[0].Matched, "ห้าร้อยสิบ")
	}
	if ms[0].Start != 0 || ms[0].End != 30 {
		t.Errorf("ms[0] offsets = %d..%d, want 0..30", ms[0].Start, ms[0].End)
	}
	if got := ms[1].Value; got.Cmp(bigI(10)) != 0 {
		t.Errorf("ms[1].Value = %v, want 10", got)
	}
	if ms[1].Matched != "สิบ" {
		t.Errorf("ms[1].Matched = %q, want %q", ms[1].Matched, "สิบ")
	}
	if ms[1].Start != 30 || ms[1].End != 39 {
		t.Errorf("ms[1] offsets = %d..%d, want 30..39", ms[1].Start, ms[1].End)
	}
}

func TestExtractNumbers_Offsets(t *testing.T) {
	const text = "a สิบ b 7 c"
	ms := ExtractNumbers(text)
	if len(ms) != 2 {
		t.Fatalf("len = %d, want 2", len(ms))
	}
	for _, m := range ms {
		if sub := text[m.Start:m.End]; sub != m.Matched {
			t.Errorf("text[%d:%d] = %q, want %q", m.Start, m.End, sub, m.Matched)
		}
	}
	if got := ms[0].Value; got.Cmp(bigI(10)) != 0 {
		t.Errorf("ms[0].Value = %v, want 10", got)
	}
	if got := ms[1].Value; got.Cmp(bigI(7)) != 0 {
		t.Errorf("ms[1].Value = %v, want 7", got)
	}
	if ms[0].Start != 2 || ms[0].End != 11 {
		t.Errorf("ms[0] offsets = %d..%d, want 2..11", ms[0].Start, ms[0].End)
	}
	if ms[1].Start != 14 || ms[1].End != 15 {
		t.Errorf("ms[1] offsets = %d..%d, want 14..15", ms[1].Start, ms[1].End)
	}
}

func TestExtractNumbers_NegConnectorNotPartOfMatch(t *testing.T) {
	// 'ลบ' (minus) is a non-number connector; extracted values are non-negative
	// magnitudes, so 'ลบห้า' yields one match for ห้า (5).
	const text = "ลบห้า"
	ms := ExtractNumbers(text)
	if len(ms) != 1 {
		t.Fatalf("len = %d, want 1", len(ms))
	}
	if got := ms[0].Value; got.Cmp(bigI(5)) != 0 {
		t.Errorf("Value = %v, want 5", got)
	}
	if ms[0].Matched != "ห้า" {
		t.Errorf("Matched = %q, want %q", ms[0].Matched, "ห้า")
	}
	if ms[0].Start != 6 || ms[0].End != 15 {
		t.Errorf("offsets = %d..%d, want 6..15", ms[0].Start, ms[0].End)
	}
	if sub := text[ms[0].Start:ms[0].End]; sub != ms[0].Matched {
		t.Errorf("text[%d:%d] = %q, want %q", ms[0].Start, ms[0].End, sub, ms[0].Matched)
	}
}
