package endpoints

import (
	"github.com/herocod3r/robinhood-cli/internal/robinhood"
	"github.com/shopspring/decimal"
)

// toDec converts a robinhood.Money string to a decimal.Decimal.
// Invalid / empty strings parse as zero.
func toDec(m robinhood.Money) decimal.Decimal {
	d, err := decimal.NewFromString(string(m))
	if err != nil {
		return decimal.Zero
	}
	return d
}

// mulMoney returns a * b as a fixed-4-decimal Money string.
func mulMoney(a, b robinhood.Money) robinhood.Money {
	return robinhood.Money(toDec(a).Mul(toDec(b)).StringFixed(4))
}

// subMoney returns a - b as a fixed-4-decimal Money string.
func subMoney(a, b robinhood.Money) robinhood.Money {
	return robinhood.Money(toDec(a).Sub(toDec(b)).StringFixed(4))
}

// divMoney returns a / b as a fixed-4-decimal Money string; b=0 yields "0".
func divMoney(a, b robinhood.Money) robinhood.Money {
	bd := toDec(b)
	if bd.IsZero() {
		return robinhood.Money("0")
	}
	return robinhood.Money(toDec(a).Div(bd).StringFixed(4))
}

// moneyLess reports whether a < b.
func moneyLess(a, b robinhood.Money) bool {
	return toDec(a).LessThan(toDec(b))
}
