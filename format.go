package thainum

import (
	"fmt"
	"math/big"
	"strings"
)

func errInvalid(s string) error { return fmt.Errorf("thainum: invalid number %q", s) }

// FormatInt formats an integer with thousands separators: 1234567 -> "1,234,567".
func FormatInt(n int64) string {
	neg := n < 0
	s := new(big.Int).Abs(big.NewInt(n)).String()
	s = groupThousands(s)
	if neg {
		return "-" + s
	}
	return s
}

// FormatSatang formats a satang amount as grouped baht with two decimals:
// 2121 -> "21.21", 100000 -> "1,000.00".
func FormatSatang(satang int64) string {
	neg := satang < 0
	mag := new(big.Int).Abs(big.NewInt(satang))
	baht := new(big.Int)
	sat := new(big.Int)
	baht.DivMod(mag, hundred, sat)
	out := groupThousands(baht.String()) + fmt.Sprintf(".%02d", sat.Int64())
	if neg {
		return "-" + out
	}
	return out
}

// FormatTHB formats a satang amount as Thai baht with the ฿ symbol:
// 2121 -> "฿21.21".
func FormatTHB(satang int64) string {
	if satang < 0 {
		return "-฿" + FormatSatang(-satang)
	}
	return "฿" + FormatSatang(satang)
}

// groupThousands inserts commas every three digits into a plain digit string.
func groupThousands(s string) string {
	n := len(s)
	if n <= 3 {
		return s
	}
	var b strings.Builder
	pre := n % 3
	if pre > 0 {
		b.WriteString(s[:pre])
	}
	for i := pre; i < n; i += 3 {
		if b.Len() > 0 {
			b.WriteByte(',')
		}
		b.WriteString(s[i : i+3])
	}
	return b.String()
}
