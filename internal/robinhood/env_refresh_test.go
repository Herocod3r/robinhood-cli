package robinhood

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

// TestEnv_SessionRefresh_DoesNotClobberKeychain is the Fix N regression guard:
// an env-loaded (Ephemeral=true) session hitting a 401 must refresh in memory
// ONLY, leaving the on-disk keychain entry for the same profile untouched.
//
// Setup:
//   - Env tokens present (Ephemeral=true session).
//   - A DIFFERENT session written to the file-backend keychain for "default".
//   - Fake Robinhood server that 401s the first GET, accepts the refresh, and
//     then returns 200 for the retry.
//   - Assert: request succeeds with rotated env token; keychain file unchanged.
func TestEnv_SessionRefresh_DoesNotClobberKeychain(t *testing.T) {
	cfgRoot := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", cfgRoot)
	t.Setenv("ROBINHOOD_KEYCHAIN_BACKEND", "file")

	// Seed keychain with a DIFFERENT session than the env one.
	kcSess := &Session{
		Username:     "kc-user",
		AccessToken:  "kc-access",
		RefreshToken: "kc-refresh",
		DeviceToken:  "kc-device",
		ExpiresAt:    time.Now().Add(time.Hour),
	}
	if err := kcSess.SaveToKeychain("default"); err != nil {
		t.Fatalf("seed keychain: %v", err)
	}
	// Snapshot the keychain file to diff at the end. The file-backend writes
	// under $XDG_CONFIG_HOME/robinhood-cli/ with a layout we find by walking
	// to stay robust against layout tweaks.
	var keychainPath string
	_ = filepath.Walk(filepath.Join(cfgRoot, "robinhood-cli"), func(p string, info os.FileInfo, _ error) error {
		if info == nil || info.IsDir() {
			return nil
		}
		// Expect a name containing "default" (profile) — the probe's config file
		// is config.yaml and gets filtered out.
		if filepath.Base(p) == "config.yaml" {
			return nil
		}
		keychainPath = p
		return nil
	})
	if keychainPath == "" {
		t.Fatalf("could not locate seeded keychain file under %s", cfgRoot)
	}
	kcBefore, err := os.ReadFile(keychainPath)
	if err != nil {
		t.Fatalf("read keychain file: %v", err)
	}

	// Fake Robinhood:
	//  - First GET /positions/ returns 401
	//  - POST /oauth2/token/ (grant_type=refresh_token) returns rotated tokens
	//  - Retry GET /positions/ returns 200 with a minimal body
	var refreshCalls atomic.Int32
	var retryServedWithNewToken atomic.Bool
	firstGet := atomic.Bool{}
	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token/":
			refreshCalls.Add(1)
			// Rotated tokens.
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"access_token":"env-access-rotated","refresh_token":"env-refresh-rotated","expires_in":3600}`)
		case "/positions/":
			if !firstGet.Swap(true) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if r.Header.Get("Authorization") == "Bearer env-access-rotated" {
				retryServedWithNewToken.Store(true)
			}
			fmt.Fprint(w, `{"results":[]}`)
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	// Set env tokens so LoadSessionFromEnv returns an Ephemeral session with
	// a DIFFERENT access token than the keychain.
	t.Setenv("ROBINHOOD_ACCESS_TOKEN", "env-access-initial")
	t.Setenv("ROBINHOOD_REFRESH_TOKEN", "env-refresh-initial")

	envSess, err := LoadSessionFromEnv()
	if err != nil {
		t.Fatalf("LoadSessionFromEnv: %v", err)
	}
	if !envSess.Ephemeral {
		t.Fatalf("expected Ephemeral session")
	}

	c := NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
	c.SetSession(envSess)
	c.SetProfile("default")

	// Trigger the 401 -> refresh -> retry path.
	var resp struct {
		Results []any `json:"results"`
	}
	if err := c.GetJSONCtx(context.Background(), APIHost, "/positions/", &resp); err != nil {
		t.Fatalf("GetJSONCtx: %v", err)
	}

	if refreshCalls.Load() != 1 {
		t.Fatalf("refresh called %d times, want 1", refreshCalls.Load())
	}
	if !retryServedWithNewToken.Load() {
		t.Fatalf("retry did not present rotated env token")
	}
	// In-memory session should hold the rotated env token.
	if got := c.Session().AccessToken; got != "env-access-rotated" {
		t.Fatalf("in-memory access token = %q, want env-access-rotated", got)
	}
	if !c.Session().Ephemeral {
		t.Fatalf("Ephemeral flag lost across refresh")
	}

	// Keychain file must be byte-for-byte unchanged.
	kcAfter, err := os.ReadFile(keychainPath)
	if err != nil {
		t.Fatalf("reread keychain: %v", err)
	}
	if string(kcAfter) != string(kcBefore) {
		t.Fatalf("keychain file clobbered by env-session refresh\nbefore: %s\nafter:  %s", string(kcBefore), string(kcAfter))
	}

	// Sanity: the on-disk session still has the original keychain token.
	reloaded, err := LoadFromKeychain("default")
	if err != nil {
		t.Fatalf("LoadFromKeychain after test: %v", err)
	}
	if reloaded.AccessToken != "kc-access" {
		t.Fatalf("keychain access token = %q, want kc-access", reloaded.AccessToken)
	}
}
