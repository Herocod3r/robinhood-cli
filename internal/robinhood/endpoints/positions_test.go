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

// mockRobinhood returns a test server that can play back positions + instruments + quotes.
func mockRobinhood(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/positions/" && r.URL.Query().Get("cursor") == "":
			_, _ = w.Write([]byte(`{"results":[{"quantity":"10.0000","average_buy_price":"150.0000","instrument":"https://api.example.com/instruments/uuid-1/"}],"next":"https://api.example.com/positions/?cursor=p2"}`))
		case r.URL.Path == "/positions/" && r.URL.Query().Get("cursor") == "p2":
			_, _ = w.Write([]byte(`{"results":[{"quantity":"5.0000","average_buy_price":"200.0000","instrument":"https://api.example.com/instruments/uuid-2/"},{"quantity":"0","average_buy_price":"0","instrument":"https://api.example.com/instruments/uuid-3/"}],"next":""}`))
		case r.URL.Path == "/instruments/" && r.URL.Query().Get("ids") != "":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{
					{"id": "uuid-1", "url": "https://api.example.com/instruments/uuid-1/", "symbol": "AAPL", "name": "Apple"},
					{"id": "uuid-2", "url": "https://api.example.com/instruments/uuid-2/", "symbol": "NVDA", "name": "Nvidia"},
					{"id": "uuid-3", "url": "https://api.example.com/instruments/uuid-3/", "symbol": "ZERO", "name": "Zero Co"},
				},
			})
		case r.URL.Path == "/quotes/":
			_, _ = w.Write([]byte(`{"results":[
				{"symbol":"AAPL","last_trade_price":"200.0000","previous_close":"180.0000"},
				{"symbol":"NVDA","last_trade_price":"150.0000","previous_close":"140.0000"},
				{"symbol":"ZERO","last_trade_price":"1.0000","previous_close":"1.0000"}
			]}`))
		default:
			t.Fatalf("unexpected request: %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
		}
	}))
}

func TestPositions_List_Paginates_Enriches_Sorts(t *testing.T) {
	ts := mockRobinhood(t)
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})

	ps, err := NewPositions(c).List(context.Background(), ListOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if len(ps) != 3 {
		t.Fatalf("len=%d", len(ps))
	}
	// Default sort is by market value desc: AAPL 10*200=2000 > NVDA 5*150=750 > ZERO 0.
	if ps[0].Symbol != "AAPL" {
		t.Fatalf("first=%s", ps[0].Symbol)
	}
	if string(ps[0].MarketValue) != "2000.0000" {
		t.Fatalf("AAPL mv=%s", ps[0].MarketValue)
	}
	if string(ps[0].UnrealizedPL) != "500.0000" {
		t.Fatalf("AAPL pl=%s", ps[0].UnrealizedPL)
	}
	if ps[0].Name != "Apple" {
		t.Fatalf("AAPL name=%q", ps[0].Name)
	}
}

func TestPositions_List_NonZero(t *testing.T) {
	ts := mockRobinhood(t)
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})

	ps, err := NewPositions(c).List(context.Background(), ListOpts{NonZero: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(ps) != 2 {
		t.Fatalf("nonzero len=%d; want 2 (no ZERO row)", len(ps))
	}
	for _, p := range ps {
		if p.Symbol == "ZERO" {
			t.Fatalf("nonzero filter kept zero row")
		}
	}
}

func TestPositions_List_SortSymbol(t *testing.T) {
	ts := mockRobinhood(t)
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})

	ps, err := NewPositions(c).List(context.Background(), ListOpts{Sort: "symbol"})
	if err != nil {
		t.Fatal(err)
	}
	if ps[0].Symbol != "AAPL" || ps[1].Symbol != "NVDA" || ps[2].Symbol != "ZERO" {
		t.Fatalf("sort=symbol got %v/%v/%v", ps[0].Symbol, ps[1].Symbol, ps[2].Symbol)
	}
}

func TestPositions_List_Limit(t *testing.T) {
	ts := mockRobinhood(t)
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})

	ps, err := NewPositions(c).List(context.Background(), ListOpts{Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(ps) != 1 {
		t.Fatalf("limit=1 got %d", len(ps))
	}
}

func TestPositions_List_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(`{"detail":"boom"}`))
	}))
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})
	if _, err := NewPositions(c).List(context.Background(), ListOpts{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestPositions_List_SymbolFilter(t *testing.T) {
	ts := mockRobinhood(t)
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})

	ps, err := NewPositions(c).List(context.Background(), ListOpts{Symbol: "NVDA"})
	if err != nil {
		t.Fatal(err)
	}
	if len(ps) != 1 || ps[0].Symbol != "NVDA" {
		t.Fatalf("symbol filter got %+v", ps)
	}
}

// sanity-check mock server returns expected shape
func TestPositions_MockShape(t *testing.T) {
	ts := mockRobinhood(t)
	defer ts.Close()
	resp, err := ts.Client().Get(ts.URL + "/positions/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var body map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if !strings.Contains(body["next"].(string), "cursor=p2") {
		t.Fatalf("mock: %+v", body)
	}
}
