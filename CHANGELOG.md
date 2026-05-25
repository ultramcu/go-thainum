# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **Thai dates** — `MonthTH`/`MonthAbbrTH` (มกราคม / ม.ค.), `WeekdayTH`/
  `WeekdayAbbrTH` (วันอาทิตย์ / อา.), `BuddhistYear(time.Time)` (+543), and
  `FormatDate`/`FormatDateAbbr`/`FormatDateFull` that render a `time.Time` as a
  Thai date with the Buddhist-Era year (e.g. "วันพุธที่ 5 มิถุนายน พ.ศ. 2567").

## [0.2.0] - 2026-05-25

### Added

- **Ordinals** — `Ordinal(int64)` ("ที่" + the number; e.g. 21 → "ที่ยี่สิบเอ็ด").
- **Fractions** — `Fraction(num, den int64)` ("เศษ" num "ส่วน" den;
  e.g. 3/4 → "เศษสามส่วนสี่").
- **Buddhist-Era years** — `Year(be int64)` ("พุทธศักราช" + the number) plus
  `CEToBE`/`BEToCE` year converters (±543).
- **`SatangFromFloat(float64) int64`** — convert a float baht amount to satang
  (rounded half away from zero), exposing the float→satang→`BahtSatang` path.
- All of the above are also available as `Speller` methods, honouring `EtMode`.

### Changed

- `BahtFromFloat` now converts via `SatangFromFloat` (float → satang int64) and
  defers to `BahtSatang`, instead of formatting through a string.

## [0.1.0] - 2026-05-25

### Added

- **Thai numerals** — `ToThaiDigits` and `ToArabicDigits` for converting between
  Arabic and Thai digit characters (e.g. `"101"` ⇄ `"๑๐๑"`).
- **Number-to-words spelling** — `Spell(int64)`, `SpellBig(*big.Int)`, and
  `SpellDecimal(string)`, correct to ล้านล้าน (10¹²) and beyond.
- **EtMode** — `EtAlways` (default; Royal-Institute-recommended `เอ็ด` form) and
  `EtTensOnly`, configurable via the `Speller{Et EtMode}` type with `.Int`,
  `.Big`, `.Decimal`, and `.Baht` methods.
- **Baht text (บาทตัวอักษร)** — `Baht(baht int64)` (whole-baht unit) wrapping
  `BahtSatang(satang int64)` for sub-baht precision, plus `BahtBig`/
  `BahtSatangBig(*big.Int)`, `BahtFromString(string)`, and the convenience
  (lossy) `BahtFromFloat(float64)`. Amounts are handled in integer baht/satang or
  `math/big`, never `float64`.
- **Formatting** — `FormatInt` (thousands separators), `FormatSatang`
  (satang → decimal), and `FormatTHB` (`฿` Thai Baht display).
- **Reverse parser** — `ParseInt(string)`, `ParseBig(string)`, and
  `ParseBaht(string)` to turn Thai words back into numbers and satang. All parse
  errors wrap the sentinel `ErrParse`.
- **`decimaladapter` subpackage** — optional `shopspring/decimal` support that
  keeps the core package dependency-free; only importers of the subpackage pull
  in the dependency.

[Unreleased]: https://github.com/ultramcu/go-thainum/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/ultramcu/go-thainum/releases/tag/v0.1.0
