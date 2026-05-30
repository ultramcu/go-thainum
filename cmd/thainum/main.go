// Command thainum is a tiny command-line front-end for the go-thainum library.
//
// It is a thin layer over the public library API and adds NO dependency: the
// argument parsing is hand-rolled (the standard flag package's positional model
// does not fit a "verb then args then flags" CLI cleanly), so the package graph
// stays on the standard library only. run returns an exit code and writes to an
// io.Writer so it can be driven directly from a fast, deterministic test;
// main just forwards to it.
//
// It mirrors the Dart bin/thainum.dart CLI: the same spell/baht/parse/digits/date
// subcommands and the same --et / --json / --full (plus --abbr) flags.
package main

import (
	"fmt"
	"io"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	thainum "github.com/ultramcu/go-thainum"
)

const usage = `thainum — Thai number toolkit CLI

Usage: thainum <command> [args] [flags]

Commands:
  spell <int>            Spell an integer in Thai words.
  baht <amount>          Render an amount (int or decimal string) as baht text.
  parse <thai-words>     Parse Thai number words back into a number.
  digits <int>           Convert ASCII digits to Thai numerals.
  date <YYYY-MM-DD>      Format a date as a Thai (Buddhist-Era) date.

Flags:
  --et=always|tensOnly   Trailing-one convention for ` + "`spell`" + ` (default always).
  --full                 ` + "`date`" + `: full form (weekday + พ.ศ.).
  --abbr                 ` + "`date`" + `: abbreviated month form.
  --json                 Emit a small JSON object instead of bare text.
  --help, -h             Show this help.

Examples:
  thainum spell 101
  thainum spell 101 --et=tensOnly
  thainum baht 21.21
  thainum parse ยี่สิบเอ็ด --json
  thainum digits 2566
  thainum date 2024-06-05 --full`

// cliError is an internal CLI usage/argument error carrying its own exit code.
type cliError struct {
	message  string
	exitCode int
}

func (e *cliError) Error() string { return e.message }

// run parses args, dispatches the subcommand, writes the result (with a
// trailing newline) to stdout, and returns the process exit code. main calls
// it; tests can call it directly.
func run(args []string, stdout io.Writer) int {
	// Hand-rolled flag split: pull --flag / --flag=value out, keep the rest as
	// positional arguments.
	var positional []string
	flags := map[string]string{}
	for _, a := range args {
		switch {
		case a == "--help" || a == "-h":
			flags["help"] = ""
		case strings.HasPrefix(a, "--"):
			body := a[2:]
			if eq := strings.IndexByte(body, '='); eq >= 0 {
				flags[body[:eq]] = body[eq+1:]
			} else {
				flags[body] = ""
			}
		default:
			positional = append(positional, a)
		}
	}

	if _, ok := flags["help"]; ok {
		fmt.Fprintln(stdout, usage)
		return 0
	}
	if len(positional) == 0 {
		fmt.Fprintln(stdout, usage)
		return 64
	}

	cmd := positional[0]
	rest := positional[1:]
	_, wantJSON := flags["json"]

	label, value, err := dispatch(cmd, rest, flags)
	if err != nil {
		code := 1
		var ce *cliError
		if asCliError(err, &ce) {
			code = ce.exitCode
		}
		fmt.Fprintln(stdout, formatError(err.Error(), wantJSON))
		return code
	}
	fmt.Fprintln(stdout, formatValue(label, value, wantJSON))
	return 0
}

// asCliError reports whether err is a *cliError, assigning it to *target.
func asCliError(err error, target **cliError) bool {
	if ce, ok := err.(*cliError); ok {
		*target = ce
		return true
	}
	return false
}

// dispatch runs one subcommand and returns a (label, value) pair: label is the
// JSON field name used by --json. It mirrors Dart's _dispatch.
func dispatch(cmd string, rest []string, flags map[string]string) (label, value string, err error) {
	switch cmd {
	case "spell":
		n, err := requireBigInt(rest, "spell <int>")
		if err != nil {
			return "", "", err
		}
		et, err := etMode(flags)
		if err != nil {
			return "", "", err
		}
		return "words", thainum.Speller{Et: et}.Big(n), nil
	case "baht":
		amount, err := requireOne(rest, "baht <amount>")
		if err != nil {
			return "", "", err
		}
		// A bare integer is whole baht; anything with a '.' is a decimal string.
		if strings.Contains(amount, ".") {
			text, err := thainum.BahtFromString(amount)
			if err != nil {
				return "", "", err
			}
			return "bahtText", text, nil
		}
		b, ok := new(big.Int).SetString(strings.ReplaceAll(amount, ",", ""), 10)
		if !ok {
			return "", "", &cliError{fmt.Sprintf("expected an integer, got %q", amount), 64}
		}
		return "bahtText", thainum.BahtBig(b), nil
	case "parse":
		words, err := requireOne(rest, "parse <thai-words>")
		if err != nil {
			return "", "", err
		}
		// Lenient so multiple shell tokens joined with spaces (e.g.
		// "parse ยี่สิบ เอ็ด") parse as one number. ParseDecimal mirrors
		// Dart's parseDecimal: it also reads a จุด-separated fractional part
		// (e.g. "สิบสองจุดสามสี่" -> "12.34") and returns a decimal string.
		v, perr := thainum.ParseDecimal(words, thainum.Lenient())
		if perr != nil {
			return "", "", perr
		}
		return "value", v, nil
	case "digits":
		raw, err := requireOne(rest, "digits <int>")
		if err != nil {
			return "", "", err
		}
		return "thaiDigits", thainum.ToThaiDigits(raw), nil
	case "date":
		raw, err := requireOne(rest, "date <YYYY-MM-DD>")
		if err != nil {
			return "", "", err
		}
		d, err := parseISODate(raw)
		if err != nil {
			return "", "", err
		}
		if _, ok := flags["full"]; ok {
			return "date", thainum.FormatDateFull(d), nil
		}
		if _, ok := flags["abbr"]; ok {
			return "date", thainum.FormatDateAbbr(d), nil
		}
		return "date", thainum.FormatDate(d), nil
	default:
		return "", "", &cliError{fmt.Sprintf("unknown command: %s\n\n%s", cmd, usage), 64}
	}
}

// etMode reads the --et flag, mirroring Dart's _etMode.
func etMode(flags map[string]string) (thainum.EtMode, error) {
	switch flags["et"] {
	case "", "always":
		return thainum.EtAlways, nil
	case "tensOnly":
		return thainum.EtTensOnly, nil
	default:
		return thainum.EtAlways, &cliError{
			fmt.Sprintf("--et must be \"always\" or \"tensOnly\" (got %q)", flags["et"]), 64,
		}
	}
}

// requireOne joins all positional remainder tokens with spaces, erroring if
// there are none. Mirrors Dart's _requireOne.
func requireOne(rest []string, form string) (string, error) {
	if len(rest) == 0 {
		return "", &cliError{"missing argument: " + form, 64}
	}
	return strings.Join(rest, " "), nil
}

// requireBigInt joins the remainder, strips thousands commas, and parses it as
// an arbitrary-precision integer. Mirrors Dart's _requireInt (widened to big).
func requireBigInt(rest []string, form string) (*big.Int, error) {
	s, err := requireOne(rest, form)
	if err != nil {
		return nil, err
	}
	n, ok := new(big.Int).SetString(strings.ReplaceAll(s, ",", ""), 10)
	if !ok {
		return nil, &cliError{fmt.Sprintf("expected an integer, got %q", s), 64}
	}
	return n, nil
}

// parseISODate parses a strict YYYY-MM-DD date, rejecting non-calendar dates
// (e.g. 2024-02-31). Mirrors Dart's _parseIsoDate.
func parseISODate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	// Shape gate first (mirrors Dart's YYYY-MM-DD regex): a malformed shape is a
	// "format" error, separate from a well-formed but non-calendar date.
	if !isISODateShape(s) {
		return time.Time{}, &cliError{fmt.Sprintf("expected a date as YYYY-MM-DD, got %q", s), 64}
	}
	y, _ := strconv.Atoi(s[0:4])
	m, _ := strconv.Atoi(s[5:7])
	dd, _ := strconv.Atoi(s[8:10])
	d := time.Date(y, time.Month(m), dd, 0, 0, 0, 0, time.UTC)
	// time.Date normalizes out-of-range fields (e.g. Feb 31 -> Mar 2); if any
	// component changed, the input was not a real calendar date.
	if d.Year() != y || int(d.Month()) != m || d.Day() != dd {
		return time.Time{}, &cliError{fmt.Sprintf("not a valid calendar date: %q", s), 64}
	}
	return d, nil
}

// isISODateShape reports whether s is exactly "DDDD-DD-DD" (ASCII digits).
func isISODateShape(s string) bool {
	if len(s) != 10 || s[4] != '-' || s[7] != '-' {
		return false
	}
	for i := 0; i < len(s); i++ {
		if i == 4 || i == 7 {
			continue
		}
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// formatValue renders the (label, value) result either as bare text or as a
// one-field JSON object. Mirrors Dart's _format.
func formatValue(label, value string, asJSON bool) string {
	if !asJSON {
		return value
	}
	return fmt.Sprintf("{%q: %s}", label, jsonString(value))
}

// formatError renders an error either as "error: ..." text or as a one-field
// JSON object. Mirrors Dart's _error.
func formatError(message string, asJSON bool) string {
	if !asJSON {
		return "error: " + message
	}
	return fmt.Sprintf("{\"error\": %s}", jsonString(message))
}

// jsonString encodes s as a JSON string literal, escaping the characters that
// can occur in our outputs/messages. It mirrors Dart's _jsonString (a manual
// encoder kept so the output byte-matches the Dart CLI).
func jsonString(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\b':
			b.WriteString(`\b`)
		case '\t':
			b.WriteString(`\t`)
		case '\n':
			b.WriteString(`\n`)
		case '\f':
			b.WriteString(`\f`)
		case '\r':
			b.WriteString(`\r`)
		default:
			if r < 0x20 {
				fmt.Fprintf(&b, `\u%04x`, r)
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
	return b.String()
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout))
}
