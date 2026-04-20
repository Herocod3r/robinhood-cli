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

func TestOrders_List_Happy(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/orders/":
			_, _ = w.Write([]byte(`{"results":[
				{"id":"o1","side":"buy","type":"market","state":"filled","quantity":"10","price":"150","average_price":"150","fees":"0","time_in_force":"gfd","created_at":"2026-04-01T10:00:00Z","updated_at":"2026-04-01T10:00:01Z","instrument":"https://api.example.com/instruments/aapl-uuid/"},
				{"id":"o2","side":"sell","type":"limit","state":"cancelled","quantity":"5","price":"200","average_price":"0","fees":"0","time_in_force":"gtc","created_at":"2026-01-15T09:00:00Z","updated_at":"2026-01-16T09:00:00Z","instrument":"https://api.example.com/instruments/nvda-uuid/"}
			],"next":""}`))
		case "/instruments/":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{
					{"id": "aapl-uuid", "url": "https://api.example.com/instruments/aapl-uuid/", "symbol": "AAPL"},
					{"id": "nvda-uuid", "url": "https://api.example.com/instruments/nvda-uuid/", "symbol": "NVDA"},
				},
			})
		default:
			t.Fatalf("unexpected %s", r.URL)
		}
	}))
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})

	os, err := NewOrders(c).List(context.Background(), OrderOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if len(os) != 2 {
		t.Fatalf("len=%d", len(os))
	}
	if os[0].Symbol != "AAPL" || os[0].State != "filled" {
		t.Fatalf("o1=%+v", os[0])
	}
}

func TestOrders_List_SinceAndSymbol(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/orders/":
			_, _ = w.Write([]byte(`{"results":[
				{"id":"o1","side":"buy","type":"market","state":"filled","quantity":"10","price":"150","average_price":"150","fees":"0","time_in_force":"gfd","created_at":"2026-04-01T10:00:00Z","updated_at":"2026-04-01T10:00:01Z","instrument":"https://api.example.com/instruments/aapl-uuid/"},
				{"id":"o2","side":"sell","type":"limit","state":"cancelled","quantity":"5","price":"200","average_price":"0","fees":"0","time_in_force":"gtc","created_at":"2026-01-15T09:00:00Z","updated_at":"2026-01-16T09:00:00Z","instrument":"https://api.example.com/instruments/nvda-uuid/"}
			],"next":""}`))
		case "/instruments/":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{
					{"id": "aapl-uuid", "url": "https://api.example.com/instruments/aapl-uuid/", "symbol": "AAPL"},
					{"id": "nvda-uuid", "url": "https://api.example.com/instruments/nvda-uuid/", "symbol": "NVDA"},
				},
			})
		}
	}))
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})

	since := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	os, err := NewOrders(c).List(context.Background(), OrderOpts{Since: since})
	if err != nil {
		t.Fatal(err)
	}
	if len(os) != 1 || os[0].Symbol != "AAPL" {
		t.Fatalf("since filter %+v", os)
	}
	os2, err := NewOrders(c).List(context.Background(), OrderOpts{Symbol: "NVDA"})
	if err != nil {
		t.Fatal(err)
	}
	if len(os2) != 1 || os2[0].Symbol != "NVDA" {
		t.Fatalf("symbol filter %+v", os2)
	}
}

func TestOrders_List_Limit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/orders/":
			_, _ = w.Write([]byte(`{"results":[
				{"id":"o1","side":"buy","type":"market","state":"filled","quantity":"10","created_at":"2026-04-01T10:00:00Z","instrument":"https://api.example.com/instruments/aapl-uuid/"},
				{"id":"o2","side":"sell","type":"limit","state":"filled","quantity":"5","created_at":"2026-03-01T10:00:00Z","instrument":"https://api.example.com/instruments/nvda-uuid/"}
			],"next":""}`))
		case "/instruments/":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": []map[string]any{
					{"id": "aapl-uuid", "url": "https://api.example.com/instruments/aapl-uuid/", "symbol": "AAPL"},
					{"id": "nvda-uuid", "url": "https://api.example.com/instruments/nvda-uuid/", "symbol": "NVDA"},
				},
			})
		}
	}))
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})
	os, err := NewOrders(c).List(context.Background(), OrderOpts{Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(os) != 1 {
		t.Fatalf("limit=%d", len(os))
	}
}

func TestOrders_List_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) }))
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})
	if _, err := NewOrders(c).List(context.Background(), OrderOpts{}); err == nil {
		t.Fatal("expected error")
	}
}
