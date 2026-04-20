package endpoints

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

func TestWatchlist_Get(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/watchlists/Default/":
			_, _ = w.Write([]byte(`{"results":[{"instrument":"https://ex/instruments/aapl-uuid/","watchlist":"Default","created_at":"2025-01-01"},{"instrument":"https://ex/instruments/nvda-uuid/","watchlist":"Default","created_at":"2025-02-01"}]}`))
		case r.URL.Path == "/instruments/":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{
					{"id": "aapl-uuid", "url": "https://ex/instruments/aapl-uuid/", "symbol": "AAPL"},
					{"id": "nvda-uuid", "url": "https://ex/instruments/nvda-uuid/", "symbol": "NVDA"},
				},
			})
		case strings.HasPrefix(r.URL.Path, "/quotes/"):
			_, _ = w.Write([]byte(`{"results":[{"symbol":"AAPL","last_trade_price":"200"},{"symbol":"NVDA","last_trade_price":"150"}]}`))
		default:
			t.Fatalf("unexpected %s", r.URL)
		}
	}))
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})
	items, err := NewWatchlist(c).Get(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("len=%d", len(items))
	}
	if items[0].Symbol != "AAPL" || string(items[0].LastPrice) != "200" {
		t.Fatalf("items %+v", items)
	}
}

func TestWatchlist_List(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"results":[{"name":"Default","display_name":"Default"},{"name":"tech","display_name":"Tech"}]}`))
	}))
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})
	names, err := NewWatchlist(c).List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 {
		t.Fatalf("%v", names)
	}
}

func TestWatchlist_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) }))
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})
	if _, err := NewWatchlist(c).Get(context.Background(), "Default"); err == nil {
		t.Fatal("expected error")
	}
}
