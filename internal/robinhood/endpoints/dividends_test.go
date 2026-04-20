package endpoints

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

func TestDividends_List_Filters(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/dividends/":
			_, _ = w.Write([]byte(`{"results":[
				{"amount":"1.00","rate":"0.5","position":"2","paid_at":"2025-06-15","state":"paid","instrument":"https://ex/instruments/aapl-uuid/"},
				{"amount":"2.00","rate":"1.0","position":"2","paid_at":"2026-03-15","state":"paid","instrument":"https://ex/instruments/nvda-uuid/"}
			],"next":""}`))
		case r.URL.Path == "/instruments/":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{
					{"id": "aapl-uuid", "url": "https://ex/instruments/aapl-uuid/", "symbol": "AAPL"},
					{"id": "nvda-uuid", "url": "https://ex/instruments/nvda-uuid/", "symbol": "NVDA"},
				},
			})
		default:
			t.Fatalf("unexpected %s", r.URL)
		}
	}))
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})

	all, err := NewDividends(c).List(context.Background(), DividendOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("all=%d", len(all))
	}
	yr, err := NewDividends(c).List(context.Background(), DividendOpts{Year: 2025})
	if err != nil {
		t.Fatal(err)
	}
	if len(yr) != 1 || yr[0].Symbol != "AAPL" {
		t.Fatalf("year filter %+v", yr)
	}
	sym, err := NewDividends(c).List(context.Background(), DividendOpts{Symbol: "nvda"})
	if err != nil {
		t.Fatal(err)
	}
	if len(sym) != 1 || sym[0].Symbol != "NVDA" {
		t.Fatalf("sym filter %+v", sym)
	}
}

func TestDividends_List_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})
	if _, err := NewDividends(c).List(context.Background(), DividendOpts{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestParseYearFromDate(t *testing.T) {
	if parseYearFromDate("2025-06-15") != 2025 {
		t.Fatal("YYYY-MM-DD")
	}
	if parseYearFromDate("2026-01-02T03:04:05Z") != 2026 {
		t.Fatal("rfc3339")
	}
	if parseYearFromDate("") != 0 {
		t.Fatal("empty")
	}
}
