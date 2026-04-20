package endpoints

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

func TestEarnings_Get(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/marketdata/earnings/" {
			t.Fatalf("path %s", r.URL.Path)
		}
		if r.URL.Query().Get("symbol") != "AAPL" {
			t.Fatalf("sym %s", r.URL.Query().Get("symbol"))
		}
		_, _ = w.Write([]byte(`{"results":[{"year":2025,"quarter":4,"report_at":"2026-01-30","eps":{"estimate":"1.50","actual":"1.62"}}]}`))
	}))
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})
	items, err := NewEarnings(c).Get(context.Background(), "aapl")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Year != 2025 || string(items[0].EPS.Actual) != "1.62" {
		t.Fatalf("rows %+v", items)
	}
}

func TestEarnings_Empty(t *testing.T) {
	c := robinhood.NewClient()
	if _, err := NewEarnings(c).Get(context.Background(), ""); err == nil {
		t.Fatal("expected validation error")
	}
}
