package endpoints

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

const portfolioFixture = `{
  "results": [{
    "account": "https://api.robinhood.com/accounts/5ABC/",
    "equity": "98234.5000",
    "extended_hours_equity": "98240.1200",
    "market_value": "72450.1000",
    "last_core_equity": "97892.3500"
  }]
}`

const unifiedFixture = `{
  "total_equity": "98234.5000",
  "cash": "25784.4000",
  "buying_power": "25784.4000"
}`

func TestPortfolio_Get(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/portfolios/":
			_, _ = w.Write([]byte(portfolioFixture))
		case "/accounts/unified":
			_, _ = w.Write([]byte(unifiedFixture))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c := robinhood.NewClientWithHosts(srv.URL, srv.URL, srv.URL, srv.Client())
	c.SetSession(&robinhood.Session{AccessToken: "tok", ExpiresAt: time.Now().Add(time.Hour)})

	p := NewPortfolio(c)
	got, err := p.Get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got.Equity != "98234.5000" {
		t.Errorf("Equity = %q", got.Equity)
	}
	if got.ExtendedHoursEquity != "98240.1200" {
		t.Errorf("ExtendedHoursEquity = %q", got.ExtendedHoursEquity)
	}
	if got.Cash != "25784.4000" {
		t.Errorf("Cash = %q", got.Cash)
	}
	if got.BuyingPower != "25784.4000" {
		t.Errorf("BuyingPower = %q", got.BuyingPower)
	}
}

func TestPortfolio_Get_EmptyResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/portfolios/":
			_ = json.NewEncoder(w).Encode(map[string]any{"results": []any{}})
		case "/accounts/unified":
			_, _ = w.Write([]byte(unifiedFixture))
		}
	}))
	defer srv.Close()

	c := robinhood.NewClientWithHosts(srv.URL, srv.URL, srv.URL, srv.Client())
	c.SetSession(&robinhood.Session{AccessToken: "tok", ExpiresAt: time.Now().Add(time.Hour)})

	p := NewPortfolio(c)
	_, err := p.Get(context.Background())
	apiErr, ok := err.(*robinhood.APIError)
	if !ok || apiErr.Code != robinhood.CodeNotFound {
		t.Errorf("got %v", err)
	}
}
