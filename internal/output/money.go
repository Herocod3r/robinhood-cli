package output

import (
	"github.com/herocod3r/robinhood-cli/internal/robinhood"
	"github.com/shopspring/decimal"
)

// toDec parses a Money string to a decimal.Decimal; invalid → zero.
func toDec(m robinhood.Money) decimal.Decimal {
	d, err := decimal.NewFromString(string(m))
	if err != nil {
		return decimal.Zero
	}
	return d
}

// subDecimal returns (a - b) as a 4-place decimal string for display.
func subDecimal(a, b robinhood.Money) string {
	return toDec(a).Sub(toDec(b)).StringFixed(4)
}

// divPercent returns 100 * (a-b) / b as a 2-place decimal; "—" on zero denom.
func divPercent(a, b robinhood.Money) string {
	bd := toDec(b)
	if bd.IsZero() {
		return "—"
	}
	pct := toDec(a).Sub(bd).Div(bd).Mul(decimal.NewFromInt(100))
	return pct.StringFixed(2) + "%"
}

// truncate shortens s to n runes with an ellipsis if longer.
func truncate(s string, n int) string {
	if n <= 1 {
		return s
	}
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n-1]) + "…"
}
