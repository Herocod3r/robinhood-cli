package endpoints

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

func TestAccount_Get_Happy(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/accounts/":
			_, _ = w.Write([]byte(`{"results":[{"account_number":"1AA","buying_power":"500.00","cash":"100.00","sweep_enabled":true,"margin_balances":{"day_trade_count":2,"margin_balance":"50.00"}}]}`))
		case "/accounts/unified":
			_, _ = w.Write([]byte(`{"buying_power":"600.00","cash":"100.00","pattern_day_trader":false,"day_trade_count":0}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})
	a, err := NewAccount(c).Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if a.AccountNumber != "1AA" {
		t.Fatalf("acct=%s", a.AccountNumber)
	}
	// Unified overrode buying_power.
	if string(a.BuyingPower) != "600.00" {
		t.Fatalf("bp=%s", a.BuyingPower)
	}
	if a.DayTradeCount != 2 {
		t.Fatalf("dtc=%d", a.DayTradeCount)
	}
}

func TestAccount_Get_Empty(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"results":[]}`))
	}))
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})
	if _, err := NewAccount(c).Get(context.Background()); err == nil {
		t.Fatal("expected not-found")
	}
}
