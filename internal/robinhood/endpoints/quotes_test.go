package endpoints

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

func TestQuotes_Batch_Happy(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/quotes/" {
			t.Fatalf("path %s", r.URL.Path)
		}
		if r.URL.Query().Get("symbols") != "AAPL,NVDA" {
			t.Fatalf("symbols %s", r.URL.Query().Get("symbols"))
		}
		_, _ = w.Write([]byte(`{"results":[
			{"symbol":"AAPL","last_trade_price":"200.0000","bid_price":"199.9","ask_price":"200.1","previous_close":"180.0000"},
			{"symbol":"NVDA","last_trade_price":"150.0000","previous_close":"140.0000"}
		]}`))
	}))
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})
	m, err := NewQuotes(c).Batch(context.Background(), []string{"AAPL", "NVDA"})
	if err != nil {
		t.Fatal(err)
	}
	if string(m["AAPL"].LastTradePrice) != "200.0000" {
		t.Fatalf("AAPL last=%q", m["AAPL"].LastTradePrice)
	}
	if string(m["NVDA"].PreviousClose) != "140.0000" {
		t.Fatalf("NVDA prev=%q", m["NVDA"].PreviousClose)
	}
}

func TestQuotes_Batch_TooMany(t *testing.T) {
	c := robinhood.NewClient()
	c.SetSession(&robinhood.Session{AccessToken: "t"})
	syms := make([]string, 51)
	for i := range syms {
		syms[i] = "S"
	}
	if _, err := NewQuotes(c).Batch(context.Background(), syms); err == nil {
		t.Fatal("want error for >50 syms")
	}
}

func TestQuotes_Single_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"results":[]}`))
	}))
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})
	if _, err := NewQuotes(c).Single(context.Background(), "NOPE"); err == nil {
		t.Fatal("expected not-found error")
	}
}
