package robinhood

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

type pingResponse struct {
	OK bool `json:"ok"`
}

func TestClient_GetJSON_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer acc-tok" {
			t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
		}
		_ = json.NewEncoder(w).Encode(pingResponse{OK: true})
	}))
	defer srv.Close()

	c := NewClientWithHosts(srv.URL, srv.URL, srv.URL, srv.Client())
	c.SetSession(&Session{AccessToken: "acc-tok", ExpiresAt: time.Now().Add(time.Hour)})

	var out pingResponse
	if err := c.getJSON(APIHost, "/ping", &out); err != nil {
		t.Fatal(err)
	}
	if !out.OK {
		t.Error("expected OK true")
	}
}

func TestClient_GetJSON_RefreshesOn401(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ping":
			n := atomic.AddInt32(&calls, 1)
			if n == 1 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if r.Header.Get("Authorization") != "Bearer new-access" {
				t.Errorf("authorization after refresh = %q", r.Header.Get("Authorization"))
			}
			_ = json.NewEncoder(w).Encode(pingResponse{OK: true})
		case "/oauth2/token/":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "new-access",
				"refresh_token": "new-refresh",
				"expires_in":    86400,
			})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := NewClientWithHosts(srv.URL, srv.URL, srv.URL, srv.Client())
	c.SetSession(&Session{AccessToken: "old", RefreshToken: "old-ref", ExpiresAt: time.Now().Add(time.Hour)})

	var out pingResponse
	if err := c.getJSON(APIHost, "/ping", &out); err != nil {
		t.Fatal(err)
	}
	if !out.OK {
		t.Error("expected OK after refresh+retry")
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("calls = %d, want 2", got)
	}
}

func TestClient_GetJSON_RefreshesThenPermanent401(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ping":
			atomic.AddInt32(&calls, 1)
			w.WriteHeader(http.StatusUnauthorized)
		case "/oauth2/token/":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "new-access",
				"refresh_token": "new-refresh",
				"expires_in":    86400,
			})
		}
	}))
	defer srv.Close()

	c := NewClientWithHosts(srv.URL, srv.URL, srv.URL, srv.Client())
	c.SetSession(&Session{AccessToken: "old", RefreshToken: "r", ExpiresAt: time.Now().Add(time.Hour)})

	var out pingResponse
	err := c.getJSON(APIHost, "/ping", &out)
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.Code != CodeUnauthenticated {
		t.Errorf("got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("calls = %d, want exactly 2 (original + one retry)", got)
	}
}

func TestClient_GetJSON_PreEmptiveRefresh(t *testing.T) {
	var oauthCalls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ping":
			if r.Header.Get("Authorization") != "Bearer new-access" {
				t.Errorf("expected refreshed token, got %q", r.Header.Get("Authorization"))
			}
			_ = json.NewEncoder(w).Encode(pingResponse{OK: true})
		case "/oauth2/token/":
			atomic.AddInt32(&oauthCalls, 1)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "new-access",
				"refresh_token": "new-refresh",
				"expires_in":    86400,
			})
		}
	}))
	defer srv.Close()

	c := NewClientWithHosts(srv.URL, srv.URL, srv.URL, srv.Client())
	// Expired token forces pre-emptive refresh.
	c.SetSession(&Session{AccessToken: "old", RefreshToken: "r", ExpiresAt: time.Now().Add(-1 * time.Minute)})

	var out pingResponse
	if err := c.getJSON(APIHost, "/ping", &out); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&oauthCalls); got != 1 {
		t.Errorf("oauth calls = %d, want 1", got)
	}
}

// Env-loaded sessions have unknown expiry. The client MUST use the access
// token and only refresh on a 401 — otherwise one broken refresh token kills
// every request even when the access token is fine. Regression guard for
// Codex adversarial finding #2.
func TestClient_UnknownExpiry_UsesAccessTokenWithoutRefresh(t *testing.T) {
	var pingCalls, oauthCalls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ping":
			atomic.AddInt32(&pingCalls, 1)
			if r.Header.Get("Authorization") != "Bearer valid-access" {
				t.Errorf("expected access token to be sent directly, got %q", r.Header.Get("Authorization"))
			}
			_ = json.NewEncoder(w).Encode(pingResponse{OK: true})
		case "/oauth2/token/":
			atomic.AddInt32(&oauthCalls, 1)
			// Simulate a broken refresh token — should never be called here.
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
		}
	}))
	defer srv.Close()

	c := NewClientWithHosts(srv.URL, srv.URL, srv.URL, srv.Client())
	c.SetSession(&Session{AccessToken: "valid-access", RefreshToken: "broken"}) // ExpiresAt left zero

	var out pingResponse
	if err := c.getJSON(APIHost, "/ping", &out); err != nil {
		t.Fatalf("unknown-expiry should trust access token, got %v", err)
	}
	if pingCalls != 1 {
		t.Errorf("pingCalls = %d, want 1", pingCalls)
	}
	if oauthCalls != 0 {
		t.Errorf("oauthCalls = %d, want 0 (refresh must NOT fire on unknown expiry)", oauthCalls)
	}
}

// A second invocation with the same env-loaded session must also not burn
// the refresh token — each CLI invocation is a fresh process with unknown
// expiry and a valid access token.
func TestClient_UnknownExpiry_RepeatedInvocationsDoNotRefresh(t *testing.T) {
	var oauthCalls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ping":
			_ = json.NewEncoder(w).Encode(pingResponse{OK: true})
		case "/oauth2/token/":
			atomic.AddInt32(&oauthCalls, 1)
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer srv.Close()

	for i := 0; i < 3; i++ {
		c := NewClientWithHosts(srv.URL, srv.URL, srv.URL, srv.Client())
		c.SetSession(&Session{AccessToken: "valid", RefreshToken: "r"})
		var out pingResponse
		if err := c.getJSON(APIHost, "/ping", &out); err != nil {
			t.Fatalf("iteration %d: %v", i, err)
		}
	}
	if oauthCalls != 0 {
		t.Errorf("oauthCalls after 3 invocations = %d, want 0", oauthCalls)
	}
}

func TestClient_GetJSON_RateLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	c := NewClientWithHosts(srv.URL, srv.URL, srv.URL, srv.Client())
	c.SetSession(&Session{AccessToken: "a", ExpiresAt: time.Now().Add(time.Hour)})

	var out pingResponse
	err := c.getJSON(APIHost, "/ping", &out)
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.Code != CodeRateLimited {
		t.Errorf("got %v", err)
	}
	if apiErr.Hint == "" {
		t.Error("rate limit Hint should include retry-after")
	}
}

func TestClient_GetJSON_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewClientWithHosts(srv.URL, srv.URL, srv.URL, srv.Client())
	c.SetSession(&Session{AccessToken: "a", ExpiresAt: time.Now().Add(time.Hour)})

	var out pingResponse
	err := c.getJSON(APIHost, "/missing", &out)
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.Code != CodeNotFound {
		t.Errorf("got %v", err)
	}
}
