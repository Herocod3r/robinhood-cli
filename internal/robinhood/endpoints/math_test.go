package endpoints

import (
	"testing"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

func TestMulMoney(t *testing.T) {
	got := mulMoney(robinhood.Money("10.0000"), robinhood.Money("150.0000"))
	if string(got) != "1500.0000" {
		t.Fatalf("mulMoney = %q; want 1500.0000", got)
	}
}

func TestSubMoney(t *testing.T) {
	got := subMoney(robinhood.Money("2000.0000"), robinhood.Money("1500.0000"))
	if string(got) != "500.0000" {
		t.Fatalf("subMoney = %q; want 500.0000", got)
	}
}

func TestDivMoney(t *testing.T) {
	got := divMoney(robinhood.Money("500.0000"), robinhood.Money("1500.0000"))
	// 0.3333...
	if string(got) != "0.3333" {
		t.Fatalf("divMoney = %q; want 0.3333", got)
	}
	// divide-by-zero returns "0"
	got = divMoney(robinhood.Money("100"), robinhood.Money("0"))
	if string(got) != "0" {
		t.Fatalf("divMoney 0 denom = %q; want 0", got)
	}
}

func TestMoneyLess(t *testing.T) {
	if !moneyLess("10.00", "20.00") {
		t.Fatal("expected 10 < 20")
	}
	if moneyLess("30.00", "20.00") {
		t.Fatal("expected 30 not < 20")
	}
}

func TestToDec_InvalidIsZero(t *testing.T) {
	if !toDec("not-a-number").IsZero() {
		t.Fatal("invalid Money should parse as zero")
	}
}
