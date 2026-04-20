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

func TestInstruments_ResolveIDs_Batch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/instruments/") {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		ids := r.URL.Query().Get("ids")
		if ids != "uuid-1,uuid-2" {
			t.Fatalf("unexpected ids: %q", ids)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"results": []map[string]any{
				{"id": "uuid-1", "url": "https://api.example.com/instruments/uuid-1/", "symbol": "AAPL", "name": "Apple"},
				{"id": "uuid-2", "url": "https://api.example.com/instruments/uuid-2/", "symbol": "NVDA", "name": "Nvidia"},
			},
		})
	}))
	defer ts.Close()

	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})
	m, err := NewInstruments(c).ResolveIDs(context.Background(), []string{"uuid-1", "uuid-2", "uuid-1"})
	if err != nil {
		t.Fatal(err)
	}
	if got := m.Get("https://api.example.com/instruments/uuid-1/"); got != "AAPL" {
		t.Fatalf("url lookup = %q; want AAPL", got)
	}
	if got := m.Get("uuid-2"); got != "NVDA" {
		t.Fatalf("id lookup = %q; want NVDA", got)
	}
	syms := m.Symbols()
	if len(syms) != 2 {
		t.Fatalf("symbols = %v", syms)
	}
}

func TestInstruments_ResolveByURL_Pagination(t *testing.T) {
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		switch calls {
		case 1:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{
					{"id": "u1", "symbol": "AAPL"},
				},
				"next": "https://example.com/instruments/?cursor=nxt",
			})
		case 2:
			if r.URL.Query().Get("cursor") != "nxt" {
				t.Fatalf("page 2 missing cursor")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{
					{"id": "u2", "symbol": "NVDA"},
				},
			})
		default:
			t.Fatalf("unexpected extra call %d", calls)
		}
	}))
	defer ts.Close()

	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})
	m, err := NewInstruments(c).ResolveByURL(context.Background(), []string{
		"https://api.example.com/instruments/u1/",
		"https://api.example.com/instruments/u2/",
	})
	if err != nil {
		t.Fatal(err)
	}
	if m.Get("u1") != "AAPL" || m.Get("u2") != "NVDA" {
		t.Fatalf("map wrong: %+v", m)
	}
}

func TestInstrumentIDFromURL(t *testing.T) {
	cases := map[string]string{
		"https://api.robinhood.com/instruments/abc-123/": "abc-123",
		"https://api.robinhood.com/instruments/abc-123":  "abc-123",
		"":        "",
		"/broken": "broken",
	}
	for in, want := range cases {
		if got := instrumentIDFromURL(in); got != want {
			t.Errorf("instrumentIDFromURL(%q)=%q; want %q", in, got, want)
		}
	}
}

func TestInstruments_Search(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("query") != "apple" {
			t.Fatalf("missing query: %s", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"results": []map[string]any{
				{"id": "a", "symbol": "AAPL", "name": "Apple"},
				{"id": "b", "symbol": "APLE", "name": "Apple Hospitality"},
			},
		})
	}))
	defer ts.Close()

	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})
	res, err := NewInstruments(c).Search(context.Background(), "apple", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 1 || res[0].Symbol != "AAPL" {
		t.Fatalf("unexpected: %+v", res)
	}
}

func TestInstruments_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(`{"detail":"server error"}`))
	}))
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})
	if _, err := NewInstruments(c).ResolveIDs(context.Background(), []string{"x"}); err == nil {
		t.Fatal("expected error")
	}
}
