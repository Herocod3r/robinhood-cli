//go:build fixtures

package robinhood

import (
	"context"
	"testing"

	rhvcr "github.com/herocod3r/robinhood-cli/internal/robinhood/vcr"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

// TestCassette_Portfolio replays a recorded /portfolios/ interaction.
//
// The cassette ships under testdata/cassettes/portfolio.yaml. To re-record
// it (maintainer-only, requires a real Robinhood account), temporarily
// flip the recorder Mode below to ModeRecordOnly, run the test, then
// inspect the sanitized output before committing.
func TestCassette_Portfolio(t *testing.T) {
	r, err := recorder.NewWithOptions(&recorder.Options{
		CassetteName: "../../testdata/cassettes/portfolio",
		Mode:         recorder.ModeReplayOnly,
	})
	if err != nil {
		t.Fatal(err)
	}
	// SanitizeHook is a no-op in replay mode but we wire it anyway so a
	// maintainer who flips Mode to ModeRecordOnly gets sanitization for free.
	r.AddHook(rhvcr.SanitizeHook, recorder.BeforeSaveHook)
	defer func() { _ = r.Stop() }()

	// Sanity — cassette should load and have at least one interaction.
	c, err := cassette.Load("../../testdata/cassettes/portfolio")
	if err != nil {
		t.Fatalf("cassette.Load: %v", err)
	}
	if len(c.Interactions) == 0 {
		t.Fatalf("cassette has no interactions")
	}

	httpClient := r.GetDefaultClient()
	rc := NewClientWithHosts(
		"https://api.robinhood.com",
		"https://nummus.robinhood.com",
		"https://phoenix.robinhood.com",
		httpClient,
	)
	rc.SetSession(&Session{AccessToken: "recorded-token"})

	var out map[string]any
	if err := rc.GetJSONCtx(context.Background(), APIHost, "/portfolios/", &out); err != nil {
		t.Fatalf("GetJSONCtx: %v", err)
	}
	results, ok := out["results"].([]any)
	if !ok || len(results) == 0 {
		t.Fatalf("expected non-empty results array, got: %#v", out)
	}
}
