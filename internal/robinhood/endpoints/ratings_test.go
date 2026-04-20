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

func TestRatings_Get(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/instruments/" && r.URL.Query().Get("query") == "AAPL":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{{"id": "abc-123", "symbol": "AAPL"}},
			})
		case strings.HasPrefix(r.URL.Path, "/midlands/ratings/abc-123"):
			_, _ = w.Write([]byte(`{"summary":{"num_buy_ratings":5,"num_hold_ratings":2,"num_sell_ratings":1},"ratings":[{"type":"buy","text":"Strong","published_at":"2026-01-01"}]}`))
		default:
			t.Fatalf("unexpected %s", r.URL)
		}
	}))
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})
	rt, err := NewRatings(c).Get(context.Background(), "aapl")
	if err != nil {
		t.Fatal(err)
	}
	if rt.Summary.NumBuy != 5 || len(rt.Ratings) != 1 {
		t.Fatalf("rating %+v", rt)
	}
}

func TestRatings_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"results":[]}`))
	}))
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})
	if _, err := NewRatings(c).Get(context.Background(), "NOPE"); err == nil {
		t.Fatal("expected not-found")
	}
}
