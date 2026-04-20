package endpoints

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

func TestOptionsPositions_List(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/options/aggregate_positions/":
			_, _ = w.Write([]byte(`{"results":[{"chain_symbol":"AAPL","strategy":"long_call","quantity":"1","average_open_price":"2.50","legs":[{"option":"https://api.example.com/options/instruments/opt-1/","position_type":"long"}]}],"next":""}`))
		case strings.HasPrefix(r.URL.Path, "/options/instruments/opt-1"):
			_, _ = w.Write([]byte(`{"id":"opt-1","chain_symbol":"AAPL","strike_price":"200.00","type":"call","expiration_date":"2026-06-19"}`))
		default:
			t.Fatalf("unexpected %s", r.URL)
		}
	}))
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})
	ps, err := NewOptionsPositions(c).List(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(ps) != 1 || ps[0].Symbol != "AAPL" {
		t.Fatalf("ps=%+v", ps)
	}
	if len(ps[0].Legs) != 1 || ps[0].Legs[0].Type != "call" || string(ps[0].Legs[0].StrikePrice) != "200.00" {
		t.Fatalf("leg=%+v", ps[0].Legs)
	}
}

func TestOptionsPositions_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})
	if _, err := NewOptionsPositions(c).List(context.Background(), ""); err == nil {
		t.Fatal("expected error")
	}
}
