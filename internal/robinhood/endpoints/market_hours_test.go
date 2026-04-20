package endpoints

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

func TestMarkets_List(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/markets/":
			_, _ = w.Write([]byte(`{"results":[{"mic":"XNYS","name":"NYSE","acronym":"NYSE"},{"mic":"XNAS","name":"NASDAQ","acronym":"NAS"}]}`))
		case strings.HasPrefix(r.URL.Path, "/markets/XNYS/hours/"):
			_, _ = w.Write([]byte(`{"date":"2026-04-20","is_open":true,"opens_at":"2026-04-20T13:30:00Z","closes_at":"2026-04-20T20:00:00Z"}`))
		case strings.HasPrefix(r.URL.Path, "/markets/XNAS/hours/"):
			_, _ = w.Write([]byte(`{"date":"2026-04-20","is_open":true,"opens_at":"2026-04-20T13:30:00Z","closes_at":"2026-04-20T20:00:00Z"}`))
		default:
			t.Fatalf("unexpected %s", r.URL)
		}
	}))
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})
	rows, err := NewMarkets(c).List(context.Background(), "2026-04-20")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("len=%d", len(rows))
	}
	if rows[0].MIC != "XNYS" || !rows[0].IsOpen {
		t.Fatalf("rows[0]=%+v", rows[0])
	}
}

func TestMarkets_List_IndexError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) }))
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})
	if _, err := NewMarkets(c).List(context.Background(), "2026-04-20"); err == nil {
		t.Fatal("expected error")
	}
}
