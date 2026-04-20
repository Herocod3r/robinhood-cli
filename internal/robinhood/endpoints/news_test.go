package endpoints

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

func TestNews_Get_Happy(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/midlands/news/AAPL/" {
			t.Fatalf("path %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"results":[{"title":"A","source":"X","url":"https://a","published_at":"t1"},{"title":"B","source":"Y","url":"https://b","published_at":"t2"}]}`))
	}))
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})
	items, err := NewNews(c).Get(context.Background(), "aapl", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Title != "A" {
		t.Fatalf("items=%+v", items)
	}
}

func TestNews_Get_Empty(t *testing.T) {
	c := robinhood.NewClient()
	if _, err := NewNews(c).Get(context.Background(), "", 0); err == nil {
		t.Fatal("expected validation error")
	}
}
