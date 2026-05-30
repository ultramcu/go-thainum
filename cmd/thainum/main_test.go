package main

// Blind end-to-end tests for the thainum CLI (issue #11 Stream B). The CLI is
// exercised as a black box: this file compiles the command once and drives the
// resulting binary via os/exec, asserting stdout and exit code. Expected output
// is derived from the Dart 0.5.3 oracle — thainum.dart/bin/thainum.dart and
// thainum.dart/test/cli_test.dart — NOT from running the Go code.
//
// Oracle behaviour being pinned:
//   - On success: the bare value (or JSON object with --json) on stdout, exit 0.
//   - On error: an "error: ..." line (or {"error": ...} with --json), non-zero
//     exit. Usage/argument errors use a non-zero exit; the Dart oracle uses 64,
//     so this file only requires "non-zero" where the Dart test did, and exact
//     codes only where the Dart test pinned them (ThaiNumException -> 1).
//   - Output carries a trailing newline (Dart writeln / Go Println); the test
//     trims a single trailing "\n" before comparing.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var cliBin string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "thainum-cli-test")
	if err != nil {
		panic(err)
	}
	bin := filepath.Join(dir, "thainum")
	// Build the CLI in the current package directory.
	build := exec.Command("go", "build", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		os.Stderr.Write(out)
		os.RemoveAll(dir)
		panic("go build of cmd/thainum failed: " + err.Error())
	}
	cliBin = bin
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// run executes the built CLI with args and returns trimmed combined-relevant
// output and the exit code. stdout (success) and stderr (error) are merged
// because the Dart oracle's runCli returns a single output string regardless of
// stream.
func runCLI(t *testing.T, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(cliBin, args...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		} else {
			t.Fatalf("running %v: %v", args, err)
		}
	}
	out := stdout.String()
	if out == "" {
		out = stderr.String()
	}
	out = strings.TrimSuffix(out, "\n")
	return out, code
}

// --- spell -----------------------------------------------------------------

func TestCLISpell(t *testing.T) {
	if out, code := runCLI(t, "spell", "101"); code != 0 || out != "หนึ่งร้อยเอ็ด" {
		t.Errorf("spell 101 = %q (code %d), want %q (0)", out, code, "หนึ่งร้อยเอ็ด")
	}
}

func TestCLISpellEtTensOnly(t *testing.T) {
	if out, _ := runCLI(t, "spell", "101"); out != "หนึ่งร้อยเอ็ด" {
		t.Errorf("spell 101 = %q, want %q", out, "หนึ่งร้อยเอ็ด")
	}
	if out, _ := runCLI(t, "spell", "101", "--et=tensOnly"); out != "หนึ่งร้อยหนึ่ง" {
		t.Errorf("spell 101 --et=tensOnly = %q, want %q", out, "หนึ่งร้อยหนึ่ง")
	}
}

func TestCLISpellRejectsBadEt(t *testing.T) {
	out, code := runCLI(t, "spell", "1", "--et=nope")
	if code == 0 {
		t.Errorf("spell 1 --et=nope should exit non-zero, got 0")
	}
	if !strings.Contains(out, "--et") {
		t.Errorf("spell 1 --et=nope output = %q, want it to mention --et", out)
	}
}

func TestCLISpellCommaGrouped(t *testing.T) {
	if out, _ := runCLI(t, "spell", "1,000"); out != "หนึ่งพัน" {
		t.Errorf("spell 1,000 = %q, want %q", out, "หนึ่งพัน")
	}
}

// --- baht ------------------------------------------------------------------

func TestCLIBahtWhole(t *testing.T) {
	if out, _ := runCLI(t, "baht", "100"); out != "หนึ่งร้อยบาทถ้วน" {
		t.Errorf("baht 100 = %q, want %q", out, "หนึ่งร้อยบาทถ้วน")
	}
}

func TestCLIBahtDecimal(t *testing.T) {
	want := "ยี่สิบเอ็ดบาทยี่สิบเอ็ดสตางค์"
	if out, _ := runCLI(t, "baht", "21.21"); out != want {
		t.Errorf("baht 21.21 = %q, want %q", out, want)
	}
}

// --- parse -----------------------------------------------------------------

func TestCLIParse(t *testing.T) {
	if out, _ := runCLI(t, "parse", "ยี่สิบเอ็ด"); out != "21" {
		t.Errorf("parse ยี่สิบเอ็ด = %q, want %q", out, "21")
	}
}

func TestCLIParseJoinsTokens(t *testing.T) {
	if out, _ := runCLI(t, "parse", "ยี่สิบ", "เอ็ด"); out != "21" {
		t.Errorf("parse ยี่สิบ เอ็ด = %q, want %q", out, "21")
	}
}

func TestCLIParseDecimalWords(t *testing.T) {
	if out, _ := runCLI(t, "parse", "สิบสองจุดสามสี่"); out != "12.34" {
		t.Errorf("parse สิบสองจุดสามสี่ = %q, want %q", out, "12.34")
	}
}

// --- digits ----------------------------------------------------------------

func TestCLIDigits(t *testing.T) {
	if out, _ := runCLI(t, "digits", "2566"); out != "๒๕๖๖" {
		t.Errorf("digits 2566 = %q, want %q", out, "๒๕๖๖")
	}
}

// --- date ------------------------------------------------------------------

func TestCLIDateShort(t *testing.T) {
	if out, _ := runCLI(t, "date", "2024-06-05"); out != "5 มิถุนายน 2567" {
		t.Errorf("date 2024-06-05 = %q, want %q", out, "5 มิถุนายน 2567")
	}
}

func TestCLIDateAbbr(t *testing.T) {
	if out, _ := runCLI(t, "date", "2024-06-05", "--abbr"); out != "5 มิ.ย. 2567" {
		t.Errorf("date 2024-06-05 --abbr = %q, want %q", out, "5 มิ.ย. 2567")
	}
}

func TestCLIDateFull(t *testing.T) {
	want := "วันพุธที่ 5 มิถุนายน พ.ศ. 2567"
	if out, _ := runCLI(t, "date", "2024-06-05", "--full"); out != want {
		t.Errorf("date 2024-06-05 --full = %q, want %q", out, want)
	}
}

func TestCLIDateRejectsMalformed(t *testing.T) {
	if _, code := runCLI(t, "date", "not-a-date"); code == 0 {
		t.Error("date not-a-date should exit non-zero, got 0")
	}
}

// --- --json ----------------------------------------------------------------

func TestCLIJSONObject(t *testing.T) {
	out, code := runCLI(t, "spell", "101", "--json")
	if code != 0 {
		t.Errorf("spell 101 --json exit = %d, want 0", code)
	}
	if out != `{"words": "หนึ่งร้อยเอ็ด"}` {
		t.Errorf("spell 101 --json = %q, want %q", out, `{"words": "หนึ่งร้อยเอ็ด"}`)
	}
}

func TestCLIJSONError(t *testing.T) {
	out, code := runCLI(t, "baht", "xyz", "--json")
	if code == 0 {
		t.Error("baht xyz --json should exit non-zero, got 0")
	}
	if !strings.HasPrefix(out, `{"error":`) {
		t.Errorf("baht xyz --json = %q, want it to start with {\"error\":", out)
	}
}

// --- help & errors ---------------------------------------------------------

func TestCLIHelp(t *testing.T) {
	out, code := runCLI(t, "--help")
	if code != 0 {
		t.Errorf("--help exit = %d, want 0", code)
	}
	if !strings.Contains(out, "Usage:") {
		t.Errorf("--help output missing 'Usage:': %q", out)
	}
}

func TestCLINoArgs(t *testing.T) {
	out, code := runCLI(t)
	if code == 0 {
		t.Error("no args should exit non-zero, got 0")
	}
	if !strings.Contains(out, "Usage:") {
		t.Errorf("no-args output missing 'Usage:': %q", out)
	}
}

func TestCLIUnknownCommand(t *testing.T) {
	out, code := runCLI(t, "frobnicate")
	if code == 0 {
		t.Error("unknown command should exit non-zero, got 0")
	}
	if !strings.Contains(out, "unknown command") {
		t.Errorf("unknown command output = %q, want it to contain 'unknown command'", out)
	}
}

func TestCLIThaiNumExceptionCleanMessage(t *testing.T) {
	out, code := runCLI(t, "parse", "ไม่ใช่เลข")
	if code != 1 {
		t.Errorf("parse ไม่ใช่เลข exit = %d, want 1", code)
	}
	if !strings.HasPrefix(out, "error:") {
		t.Errorf("parse ไม่ใช่เลข output = %q, want it to start with 'error:'", out)
	}
}
