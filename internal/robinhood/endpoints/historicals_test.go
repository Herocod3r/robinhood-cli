package endpoints

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

func TestHistoricals_Get_Happy(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/quotes/historicals/AAPL/") {
			t.Fatalf("path %s", r.URL.Path)
		}
		if r.URL.Query().Get("interval") != "day" || r.URL.Query().Get("span") != "month" {
			t.Fatalf("params %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"historicals":[{"begins_at":"2026-01-01","open_price":"100","close_price":"110","high_price":"112","low_price":"99","volume":12345,"session":"reg"}]}`))
	}))
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})
	h, err := NewHistoricals(c).Get(context.Background(), "aapl", "day", "month")
	if err != nil {
		t.Fatal(err)
	}
	if h.Symbol != "AAPL" || len(h.Bars) != 1 || string(h.Bars[0].ClosePrice) != "110" {
		t.Fatalf("unexpected: %+v", h)
	}
}

func TestHistoricals_BadInterval(t *testing.T) {
	c := robinhood.NewClient()
	if _, err := NewHistoricals(c).Get(context.Background(), "AAPL", "nope", "year"); err == nil {
		t.Fatal("expected validation error")
	}
}
