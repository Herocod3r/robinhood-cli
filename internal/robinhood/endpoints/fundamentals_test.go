package endpoints

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

func TestFundamentals_Batch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("symbols") != "AAPL,NVDA" {
			t.Fatalf("symbols=%s", r.URL.Query().Get("symbols"))
		}
		_, _ = w.Write([]byte(`{"results":[{"symbol":"AAPL","open":"180.0","high":"205.0","low":"179.0","market_cap":"3000000000.0","pe_ratio":"30"},{"symbol":"NVDA","market_cap":"2000000000.0"}]}`))
	}))
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})
	f, err := NewFundamentals(c).Batch(context.Background(), []string{"aapl", "NVDA"})
	if err != nil {
		t.Fatal(err)
	}
	if len(f) != 2 {
		t.Fatalf("len=%d", len(f))
	}
	if f[0].Symbol != "AAPL" || string(f[0].High) != "205.0" {
		t.Fatalf("row0=%+v", f[0])
	}
}

func TestFundamentals_TooMany(t *testing.T) {
	c := robinhood.NewClient()
	syms := make([]string, 60)
	for i := range syms {
		syms[i] = "X"
	}
	if _, err := NewFundamentals(c).Batch(context.Background(), syms); err == nil {
		t.Fatal("want error")
	}
}
