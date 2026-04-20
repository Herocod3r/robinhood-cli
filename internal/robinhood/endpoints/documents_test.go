package endpoints

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

func TestDocuments_List_FilterType(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/documents/" {
			t.Fatalf("path %s", r.URL.Path)
		}
		if r.URL.Query().Get("type") != "1099" {
			t.Fatalf("type %s", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"results":[{"id":"d1","type":"1099","date":"2025-02-15","name":"TAX 2024","download_url":"https://x/d1"}],"next":""}`))
	}))
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})
	rows, err := NewDocuments(c).List(context.Background(), DocumentOpts{Type: "1099"})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].ID != "d1" {
		t.Fatalf("rows %+v", rows)
	}
}

func TestDocuments_List_SinceFilter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"results":[{"id":"a","type":"statement","date":"2024-06-01","name":"old"},{"id":"b","type":"statement","date":"2026-01-01","name":"new"}]}`))
	}))
	defer ts.Close()
	c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(&robinhood.Session{AccessToken: "t"})
	since := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	rows, err := NewDocuments(c).List(context.Background(), DocumentOpts{Since: since})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].ID != "b" {
		t.Fatalf("rows %+v", rows)
	}
}

func TestDocuments_Download_SanitizeAndPerms(t *testing.T) {
	// Serve the actual file bytes.
	fileServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("PDFBYTES"))
	}))
	defer fileServer.Close()

	c := robinhood.NewClient()
	c.SetSession(&robinhood.Session{AccessToken: "t"})
	docs := []Document{
		{ID: "d1", Date: "2025-02-15", Name: "../../etc/passwd", DownloadURL: fileServer.URL + "/sanitize"},
	}
	dir := t.TempDir()
	results, err := NewDocuments(c).Download(context.Background(), docs, DownloadOpts{Dir: dir}, fileServer.Client())
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("results %+v", results)
	}
	wantName := "2025-02-15-passwd"
	want := filepath.Join(dir, wantName)
	if results[0].Path != want && results[0].Path != want+".pdf" {
		t.Fatalf("path=%q want prefix %q", results[0].Path, want)
	}
	info, err := os.Stat(results[0].Path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("perm=%v", info.Mode().Perm())
	}
	b, _ := os.ReadFile(results[0].Path)
	if string(b) != "PDFBYTES" {
		t.Fatalf("bytes=%q", b)
	}
}

func TestDocuments_Download_ForceOverwrite(t *testing.T) {
	fileServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("v2"))
	}))
	defer fileServer.Close()

	c := robinhood.NewClient()
	c.SetSession(&robinhood.Session{AccessToken: "t"})
	dir := t.TempDir()
	existing := filepath.Join(dir, "2025-01-01-file.pdf")
	_ = os.WriteFile(existing, []byte("v1"), 0600)

	docs := []Document{{ID: "d1", Date: "2025-01-01", Name: "file.pdf", DownloadURL: fileServer.URL + "/f"}}
	// Without force: skipped.
	res, err := NewDocuments(c).Download(context.Background(), docs, DownloadOpts{Dir: dir}, fileServer.Client())
	if err != nil {
		t.Fatal(err)
	}
	if !res[0].Skipped {
		t.Fatal("expected skip")
	}
	// With force: overwrite.
	res, err = NewDocuments(c).Download(context.Background(), docs, DownloadOpts{Dir: dir, Force: true}, fileServer.Client())
	if err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(res[0].Path)
	if string(b) != "v2" {
		t.Fatalf("bytes %q", b)
	}
}
