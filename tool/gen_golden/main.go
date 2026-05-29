// Command gen_golden emits a language-neutral conformance vector file
// (vectors.json) that pins the observable behaviour of go-thainum's shared,
// cross-port surface. The Dart port (thainum.dart) replays these vectors in its
// own test suite to prove the two libraries cannot silently drift.
//
// The input set is FIXED and DETERMINISTIC: no time.Now, no host locale, no
// host timezone. All dates/times are built in UTC from explicit components.
// Map iteration is never used for record ordering, so the output is stable.
//
// Usage (writes to the Dart repo, which is checked in next to this one):
//
//	cd go-thainum && go run ./tool/gen_golden > ../thainum.dart/test/conformance/vectors.json
//
// Only the SHARED surface is emitted. Dart-only additions (tryParse*, the
// thaiDigits flag, percent, short, speak, extractNumbers, parseDecimal, …) and
// Go-only float helpers are intentionally excluded — see the README next to the
// generated file for the full included/excluded lists.
package main

import (
	"encoding/json"
	"math/big"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
	"time"

	thainum "github.com/ultramcu/go-thainum"
)

// record is one cross-port assertion: call fn(group) with args, expect out.
// For EtMode-sensitive functions, et is "always" or "tensOnly"; otherwise "".
type record struct {
	Group string        `json:"group"`
	Fn    string        `json:"fn"`
	Args  []interface{} `json:"args"`
	Et    string        `json:"et,omitempty"`
	Out   string        `json:"out"`
}

type file struct {
	Source              string   `json:"source"`
	Version             string   `json:"version"`
	GeneratedInputsOnly bool     `json:"generated_inputs_only"`
	GeneratedAtNote     string   `json:"generated_at_note"`
	Records             []record `json:"records"`
}

// recs accumulates records in deterministic append order.
var recs []record

func add(group, fn string, args []interface{}, out string) {
	recs = append(recs, record{Group: group, Fn: fn, Args: args, Out: out})
}

func addEt(group, fn string, args []interface{}, et string, out string) {
	recs = append(recs, record{Group: group, Fn: fn, Args: args, Et: et, Out: out})
}

func main() {
	genNumerals()
	genSpell()
	genBaht()
	genFormat()
	genParse()
	genExtras()
	genDate()
	genClock()

	out := file{
		Source:              "go-thainum",
		Version:             version(),
		GeneratedInputsOnly: true,
		GeneratedAtNote:     "Inputs are fixed/deterministic; outputs are the Go reference. Regenerate with: cd go-thainum && go run ./tool/gen_golden > ../thainum.dart/test/conformance/vectors.json",
		Records:             recs,
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false) // keep Thai/UTF-8 literal, not \uXXXX
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		panic(err)
	}
}

// version returns "<module path>@<git describe>" so the pinned reference is
// recorded in the file header. The git rev is read by shelling out to git
// (works under `go run`); falls back to the build-info VCS setting, then to the
// module version.
func version() string {
	path := "github.com/ultramcu/go-thainum"
	modVer := ""
	if bi, ok := debug.ReadBuildInfo(); ok {
		if bi.Main.Path != "" {
			path = bi.Main.Path
		}
		modVer = bi.Main.Version
		for _, s := range bi.Settings {
			if s.Key == "vcs.revision" && len(s.Value) >= 7 {
				modVer = s.Value[:7]
			}
		}
	}
	if d := gitDescribe(); d != "" {
		return path + "@" + d
	}
	if modVer != "" && modVer != "(devel)" {
		return path + "@" + modVer
	}
	return path
}

func gitDescribe() string {
	cmd := exec.Command("git", "describe", "--tags", "--always", "--dirty")
	if wd, err := os.Getwd(); err == nil {
		cmd.Dir = wd
	}
	b, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// ---- input sets -----------------------------------------------------------

// intBoundaries are the shared spell/parse integer boundaries.
var intBoundaries = []int64{
	0, 1, 5, 9, 10, 11, 19, 20, 21, 25, 99, 100, 101, 111, 200, 999,
	1000, 1001, 1111, 9999, 10000, 100000, 999999,
	1000000, 1000001, 1000101, 9999999,
	1000000000000, // 10^12
	-1, -5, -11, -21, -101, -1000001,
}

// bigBoundaries are decimal strings for arbitrary-precision spell/parse.
var bigBoundaries = []string{
	"0", "1", "1000000", "1000000000000",
	"1000000000000000000",                   // 10^18
	"123456789012345678901234567890",        // 30 digits
	"-1000000000000000000",                  // negative big
	"1000000000000000000000000000000000001", // 10^36 + 1
}

// satangBoundaries are shared baht-satang boundaries.
var satangBoundaries = []int64{
	0, 1, 25, 50, 99, 100, 101, 2121, 100000, 1000000, 99999999,
	-1, -25, -2121, -100000,
}

// bahtWholeBoundaries are shared whole-baht boundaries.
var bahtWholeBoundaries = []int64{0, 1, 5, 21, 100, 101, 1000000, -5, -101}

// bahtStrings are decimal baht strings (no float).
var bahtStrings = []string{
	"0", "0.25", "21.21", "100", "100.00", "1000.50",
	"0.005", "0.004", "21.215", "-5", "-21.21",
}

func genNumerals() {
	const g = "numerals"
	samples := []string{
		"2566", "Room 101", "0", "1234567890",
		"๒๕๖๖", "ห้อง ๑๐๑", "-42", "ราคา 1,234.50 บาท",
		"", "abc", "๙๘๗๖๕๔๓๒๑๐",
	}
	for _, s := range samples {
		add(g, "toThaiDigits", []interface{}{s}, thainum.ToThaiDigits(s))
		add(g, "toArabicDigits", []interface{}{s}, thainum.ToArabicDigits(s))
	}
	// Explicit round-trip samples (ASCII -> Thai -> ASCII).
	for _, s := range []string{"0", "1", "10", "101", "1000001", "2566"} {
		thai := thainum.ToThaiDigits(s)
		add(g, "toThaiDigits", []interface{}{s}, thai)
		add(g, "toArabicDigits", []interface{}{thai}, thainum.ToArabicDigits(thai))
	}
}

func genSpell() {
	const g = "spell"
	always := thainum.Speller{Et: thainum.EtAlways}
	tens := thainum.Speller{Et: thainum.EtTensOnly}

	for _, n := range intBoundaries {
		arg := []interface{}{n}
		addEt(g, "spell", arg, "always", always.Int(n))
		addEt(g, "spell", arg, "tensOnly", tens.Int(n))
	}
	for _, s := range bigBoundaries {
		bi, _ := new(big.Int).SetString(s, 10)
		arg := []interface{}{s} // big args passed as decimal strings
		addEt(g, "spellBigInt", arg, "always", always.Big(bi))
		addEt(g, "spellBigInt", arg, "tensOnly", tens.Big(bi))
	}
	decimals := []string{
		"0", "0.5", "12.34", "100.00", "100.0", "0.05",
		"-12.34", "1000000.01", "5.0", "99.99",
	}
	for _, s := range decimals {
		a, _ := always.Decimal(s)
		t, _ := tens.Decimal(s)
		arg := []interface{}{s}
		addEt(g, "spellDecimal", arg, "always", a)
		addEt(g, "spellDecimal", arg, "tensOnly", t)
	}
}

func genBaht() {
	const g = "baht"
	for _, b := range bahtWholeBoundaries {
		add(g, "baht", []interface{}{b}, thainum.Baht(b))
	}
	for _, s := range satangBoundaries {
		add(g, "bahtSatang", []interface{}{s}, thainum.BahtSatang(s))
	}
	for _, s := range bahtStrings {
		out, err := thainum.BahtFromString(s)
		if err != nil {
			continue // skip inputs the reference rejects; Dart isn't asked about them
		}
		add(g, "bahtFromString", []interface{}{s}, out)
	}
}

func genFormat() {
	const g = "format"
	intInputs := []int64{0, 1, 12, 123, 1234, 1234567, 1000000, -1, -1234, -1234567}
	for _, n := range intInputs {
		add(g, "formatInt", []interface{}{n}, thainum.FormatInt(n))
	}
	for _, s := range satangBoundaries {
		add(g, "formatSatang", []interface{}{s}, thainum.FormatSatang(s))
		add(g, "formatThb", []interface{}{s}, thainum.FormatTHB(s))
	}
}

func genParse() {
	const g = "parse"
	// Words to parse: reuse spelled outputs (default EtMode) so we exercise the
	// reverse direction on known-good Thai text, plus a few literal forms.
	def := thainum.Speller{}
	for _, n := range intBoundaries {
		words := def.Int(n)
		got, err := thainum.ParseInt(words)
		if err != nil {
			continue
		}
		add(g, "parseInt", []interface{}{words}, itoa(got))
	}
	// Thai-digit and Arabic-digit numeric strings also accepted by parseInt.
	for _, s := range []string{"21", "๒๑", "-5", "ลบห้า", "ศูนย์", "หนึ่งล้าน"} {
		got, err := thainum.ParseInt(s)
		if err != nil {
			continue
		}
		add(g, "parseInt", []interface{}{s}, itoa(got))
	}
	// Big parse.
	for _, s := range bigBoundaries {
		bi, _ := new(big.Int).SetString(s, 10)
		words := def.Big(bi)
		got, err := thainum.ParseBig(words)
		if err != nil {
			continue
		}
		add(g, "parseBigInt", []interface{}{words}, got.String())
	}
	// Baht parse: reverse of bahtSatang.
	for _, s := range satangBoundaries {
		text := thainum.BahtSatang(s)
		got, err := thainum.ParseBaht(text)
		if err != nil {
			continue
		}
		add(g, "parseBaht", []interface{}{text}, itoa(got))
	}
}

func genExtras() {
	const g = "extras"
	for _, n := range []int64{0, 1, 5, 11, 21, 100, 101, -1} {
		add(g, "ordinal", []interface{}{n}, thainum.Ordinal(n))
	}
	fracs := [][2]int64{{1, 2}, {3, 4}, {1, 100}, {21, 1000}, {0, 5}}
	for _, f := range fracs {
		add(g, "fraction", []interface{}{f[0], f[1]}, thainum.Fraction(f[0], f[1]))
	}
	for _, be := range []int64{2566, 2500, 1, 543, 9999} {
		add(g, "year", []interface{}{be}, thainum.Year(be))
	}
	for _, ce := range []int64{2024, 1, 0, -543, 2566} {
		add(g, "ceToBe", []interface{}{ce}, itoa(thainum.CEToBE(ce)))
	}
	for _, be := range []int64{2567, 543, 0, 2566, 1} {
		add(g, "beToCe", []interface{}{be}, itoa(thainum.BEToCE(be)))
	}
}

func genDate() {
	const g = "date"
	// Dates spanning all 7 weekdays + all 12 months + a leap day. Args are
	// [year, month, day] (CE), built as UTC midnight on the Go side.
	dates := [][3]int{
		// A run covering all 7 weekdays (2024-06-02 Sun .. 2024-06-08 Sat).
		{2024, 6, 2}, {2024, 6, 3}, {2024, 6, 4}, {2024, 6, 5},
		{2024, 6, 6}, {2024, 6, 7}, {2024, 6, 8},
		// All 12 months (2023, day 15).
		{2023, 1, 15}, {2023, 2, 15}, {2023, 3, 15}, {2023, 4, 15},
		{2023, 5, 15}, {2023, 6, 15}, {2023, 7, 15}, {2023, 8, 15},
		{2023, 9, 15}, {2023, 10, 15}, {2023, 11, 15}, {2023, 12, 15},
		// Leap day + boundaries.
		{2024, 2, 29}, {2000, 1, 1}, {1999, 12, 31}, {2567 - 543, 4, 13},
	}
	for _, d := range dates {
		arg := []interface{}{d[0], d[1], d[2]}
		t := time.Date(d[0], time.Month(d[1]), d[2], 0, 0, 0, 0, time.UTC)
		add(g, "monthTh", []interface{}{d[1]}, thainum.MonthTH(time.Month(d[1])))
		add(g, "monthAbbrTh", []interface{}{d[1]}, thainum.MonthAbbrTH(time.Month(d[1])))
		add(g, "weekdayTh", arg, thainum.WeekdayTH(t.Weekday()))
		add(g, "weekdayAbbrTh", arg, thainum.WeekdayAbbrTH(t.Weekday()))
		add(g, "buddhistYear", arg, itoa(int64(thainum.BuddhistYear(t))))
		add(g, "formatDate", arg, thainum.FormatDate(t))
		add(g, "formatDateAbbr", arg, thainum.FormatDateAbbr(t))
		add(g, "formatDateFull", arg, thainum.FormatDateFull(t))
		// parseDate round-trips each rendered form back to [y,m,d].
		for _, s := range []string{
			thainum.FormatDate(t), thainum.FormatDateAbbr(t), thainum.FormatDateFull(t),
		} {
			pt, err := thainum.ParseDate(s)
			if err != nil {
				continue
			}
			add(g, "parseDate", []interface{}{s}, ymd(pt))
		}
	}
	// Standalone month-index coverage 1..12 (and out-of-range -> "").
	for m := 0; m <= 13; m++ {
		add(g, "monthTh", []interface{}{m}, thainum.MonthTH(time.Month(m)))
		add(g, "monthAbbrTh", []interface{}{m}, thainum.MonthAbbrTH(time.Month(m)))
	}
}

func genClock() {
	const g = "clock"
	hours := []int{0, 1, 6, 11, 12, 13, 15, 18, 19, 23}
	minutes := []int{0, 30, 1, 7, 45}
	for _, h := range hours {
		for _, m := range minutes {
			// Time-of-day built in UTC; only hour/minute matter to these fns.
			t := time.Date(2024, 1, 1, h, m, 0, 0, time.UTC)
			arg := []interface{}{h, m}
			add(g, "formatTime", arg, thainum.FormatTime(t))
			add(g, "formatClock", arg, thainum.FormatClock(t))
		}
	}
	// Durations: arg is total whole seconds.
	durSecs := []int64{0, 1, 45, 60, 90 * 60, 3600, 86400, 90061, -90061, 2 * 86400}
	for _, s := range durSecs {
		d := time.Duration(s) * time.Second
		add(g, "formatDuration", []interface{}{s}, thainum.FormatDuration(d))
	}
}

// ---- helpers --------------------------------------------------------------

func itoa(n int64) string { return new(big.Int).SetInt64(n).String() }

// ymd renders a time as the canonical "y,m,d" comparison string used for
// parseDate vectors (so the Dart side compares the same shape).
func ymd(t time.Time) string {
	return itoa(int64(t.Year())) + "," + itoa(int64(int(t.Month()))) + "," + itoa(int64(t.Day()))
}
