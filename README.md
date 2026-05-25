# go-thainum

**ชุดเครื่องมือจัดการตัวเลขภาษาไทยแบบครบวงจรสำหรับภาษา Go — เลขไทย, อ่านเป็นคำ, บาทตัวอักษร, จัดรูปแบบ และแปลงคำกลับเป็นตัวเลข**

**A comprehensive Thai number toolkit for Go — Thai numerals, number-to-words, baht text, formatting, and (uniquely) reverse parsing of Thai words back into numbers.**

[![Go Reference](https://pkg.go.dev/badge/github.com/ultramcu/go-thainum.svg)](https://pkg.go.dev/github.com/ultramcu/go-thainum)
[![Release](https://img.shields.io/github/v/release/ultramcu/go-thainum?sort=semver)](https://github.com/ultramcu/go-thainum/releases)
[![CI](https://github.com/ultramcu/go-thainum/actions/workflows/ci.yml/badge.svg)](https://github.com/ultramcu/go-thainum/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ultramcu/go-thainum)](https://goreportcard.com/report/github.com/ultramcu/go-thainum)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

```sh
go get github.com/ultramcu/go-thainum
```

Pure Go, MIT-licensed, **zero external dependencies** in the core (stdlib only), and requires **Go 1.21+**.

## Features

- **Thai numerals** — convert between Arabic and Thai digits (`101` ⇄ `๑๐๑`).
- **Spell numbers as Thai words** — `int64`, `*big.Int`, and decimals, correct to ล้านล้าน (10¹²) **and beyond**.
- **Baht text (บาทตัวอักษร)** — render currency amounts as the formal Thai spelling used on cheques and invoices.
- **Formatting** — thousands separators, satang-to-decimal, and a `฿` Thai Baht display format.
- **Reverse parsing (the flagship feature)** — turn Thai words *back* into numbers and satang. Effectively nothing else in the Go ecosystem does this.
- **Money is exact** — amounts are handled in integer **satang** (1 baht = 100 satang) or `math/big`, **never `float64`**, so there are no rounding surprises. A clearly-labelled lossy float entry point exists for convenience.
- **EtMode** — choose between the Royal-Institute-recommended `เอ็ด` form and the plain `หนึ่ง` form for trailing ones.
- **Optional `decimaladapter` subpackage** — adds [`shopspring/decimal`](https://github.com/shopspring/decimal) support. Only callers who import the subpackage pull that dependency; the core stays dependency-free.

## Why this exists

There are a handful of Thai "bahttext" libraries for Go, but they are **baht-only**, and most are **unlicensed and untested**. None of them cover the full surface that real Thai applications need:

> numerals + spell + baht + format + **reverse parsing (words → number)**

`go-thainum` is the only Go library that does all of it in one place. The reverse parser in particular — taking `"ยี่สิบเอ็ดบาทยี่สิบเอ็ดสตางค์"` and giving you back `2121` satang — is something essentially no other Go library provides. The whole library is decimal/integer-based (no float bugs), correct well past a trillion, MIT-licensed, and thoroughly tested.

## Quick Start

### Thai numerals

```go
package main

import (
	"fmt"

	thainum "github.com/ultramcu/go-thainum"
)

func main() {
	fmt.Println(thainum.ToThaiDigits("101"))   // ๑๐๑
	fmt.Println(thainum.ToArabicDigits("๑๐๑")) // 101
}
```

### Spell numbers as Thai words

```go
fmt.Println(thainum.Spell(21))  // ยี่สิบเอ็ด
fmt.Println(thainum.Spell(101)) // หนึ่งร้อยเอ็ด

// Arbitrary precision via *big.Int
big1T := new(big.Int).SetUint64(1_000_000_000_000)
fmt.Println(thainum.SpellBig(big1T)) // หนึ่งล้านล้าน
```

### Spell a decimal

```go
s, err := thainum.SpellDecimal("12.34")
if err != nil {
	log.Fatal(err)
}
fmt.Println(s) // สิบสองจุดสามสี่
```

### Baht text (บาทตัวอักษร)

`Baht` takes an amount in **baht** (the usual unit); use `BahtSatang` for
sub-baht precision (1 baht = 100 satang), or `BahtFromString` for a decimal
string:

```go
fmt.Println(thainum.Baht(100))           // หนึ่งร้อยบาทถ้วน   (100 baht)
fmt.Println(thainum.Baht(0))             // ศูนย์บาทถ้วน
fmt.Println(thainum.BahtSatang(2121))    // ยี่สิบเอ็ดบาทยี่สิบเอ็ดสตางค์ (21.21 baht)
fmt.Println(thainum.BahtSatang(25))      // ยี่สิบห้าสตางค์
s, _ := thainum.BahtFromString("21.21")
fmt.Println(s)                           // ยี่สิบเอ็ดบาทยี่สิบเอ็ดสตางค์
```

### Baht text from a string amount (no float)

`BahtFromString` parses a decimal-string amount in **baht** and is exact:

```go
text, err := thainum.BahtFromString("21.21")
if err != nil {
	log.Fatal(err)
}
fmt.Println(text) // ยี่สิบเอ็ดบาทยี่สิบเอ็ดสตางค์
```

There is also `BahtFromFloat(float64) string` for convenience, but it is **lossy** — prefer satang or strings for anything that must be exact.

### Formatting

```go
fmt.Println(thainum.FormatInt(1234567)) // 1,234,567
fmt.Println(thainum.FormatSatang(2121)) // 21.21
fmt.Println(thainum.FormatTHB(2121))    // ฿21.21
```

### Reverse parsing — words back into numbers

```go
n, err := thainum.ParseInt("ยี่สิบเอ็ด")
if err != nil {
	log.Fatal(err)
}
fmt.Println(n) // 21

satang, err := thainum.ParseBaht("ยี่สิบเอ็ดบาทยี่สิบเอ็ดสตางค์")
if err != nil {
	log.Fatal(err)
}
fmt.Println(satang) // 2121
```

`ParseBig(words string) (*big.Int, error)` handles arbitrarily large values. All parse errors wrap the sentinel `ErrParse`, so you can match them with `errors.Is(err, thainum.ErrParse)`.

### Ordinals, fractions, and Buddhist-Era years

```go
fmt.Println(thainum.Ordinal(21))       // ที่ยี่สิบเอ็ด
fmt.Println(thainum.Fraction(3, 4))    // เศษสามส่วนสี่
fmt.Println(thainum.Year(2566))        // พุทธศักราชสองพันห้าร้อยหกสิบหก
fmt.Println(thainum.CEToBE(2023))      // 2566
```

### Thai dates (เดือนไทย / วันไทย / ปี พ.ศ.)

```go
d := time.Date(2024, time.June, 5, 0, 0, 0, 0, time.UTC)
fmt.Println(thainum.FormatDate(d))     // 5 มิถุนายน 2567
fmt.Println(thainum.FormatDateAbbr(d)) // 5 มิ.ย. 2567
fmt.Println(thainum.FormatDateFull(d)) // วันพุธที่ 5 มิถุนายน พ.ศ. 2567
fmt.Println(thainum.MonthTH(time.June))     // มิถุนายน
fmt.Println(thainum.WeekdayTH(time.Sunday)) // วันอาทิตย์
fmt.Println(thainum.BuddhistYear(d))        // 2567
```

Dates use the Buddhist-Era year (Gregorian + 543). Wrap the result with
`ToThaiDigits` if you want Thai numerals (e.g. `๕ มิถุนายน ๒๕๖๗`).

Parse a Thai date string back into a `time.Time` (accepts any `FormatDate*` form,
Arabic or Thai digits, Buddhist-Era year):

```go
d, err := thainum.ParseDate("วันพุธที่ 5 มิถุนายน พ.ศ. 2567")
// time.Date(2024, time.June, 5, ...)
```

### Time of day and durations

```go
t := time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC)
fmt.Println(thainum.FormatTime(t))  // สิบสี่นาฬิกาสามสิบนาที (formal)
fmt.Println(thainum.FormatClock(t)) // บ่ายสองโมงครึ่ง (colloquial)
fmt.Println(thainum.FormatDuration(90 * time.Minute)) // หนึ่งชั่วโมงสามสิบนาที
```

### Money from a float

`BahtFromFloat` (and the `SatangFromFloat` helper) convert a float baht amount to
satang and reuse `BahtSatang`. Float money is lossy — prefer `BahtSatang`
(satang) or `BahtFromString` for exact input.

```go
fmt.Println(thainum.SatangFromFloat(21.21)) // 2121
fmt.Println(thainum.BahtFromFloat(21.21))   // ยี่สิบเอ็ดบาทยี่สิบเอ็ดสตางค์
```

### EtMode — `เอ็ด` vs `หนึ่ง`

By default the library uses `EtAlways`, the Royal-Institute-recommended form where a trailing one is read `เอ็ด`. Use `EtTensOnly` if you want a trailing one to read `หนึ่ง` except in the tens place:

```go
defaultSpeller := thainum.Speller{Et: thainum.EtAlways}
plainSpeller := thainum.Speller{Et: thainum.EtTensOnly}

fmt.Println(defaultSpeller.Int(101)) // หนึ่งร้อยเอ็ด
fmt.Println(plainSpeller.Int(101))   // หนึ่งร้อยหนึ่ง
```

`Speller` exposes `.Int`, `.Big`, `.Decimal`, and `.Baht` methods so you can pick the `EtMode` once and reuse it.

## Optional: `decimaladapter`

If your code already works in [`shopspring/decimal`](https://github.com/shopspring/decimal), import the adapter subpackage:

```go
import "github.com/ultramcu/go-thainum/decimaladapter"
```

Only importers of `decimaladapter` pull in `shopspring/decimal`; the core `go-thainum` package remains dependency-free.

## A note on money and precision

Money in this library is handled in **integer satang** (1 baht = 100 satang) or `math/big`, never `float64`. This means there are no binary-floating-point rounding surprises in your baht text. The `*FromFloat` entry points are provided only as a convenience and are documented as lossy — reach for the satang/string/`big.Int` APIs whenever correctness matters.

## License

[MIT](LICENSE) © 2026 MaIII (ultramcu)
