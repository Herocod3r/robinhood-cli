package robinhood

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestRefreshAccessToken_Success(t *testing.T) {
	var gotBody url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth2/token/" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		gotBody, _ = url.ParseQuery(string(body))

		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "new-access",
			"refresh_token": "new-refresh",
			"expires_in":    86400,
			"token_type":    "Bearer",
		})
	}))
	defer srv.Close()

	o := &oauth{baseURL: srv.URL, httpClient: srv.Client()}
	sess := &Session{RefreshToken: "old-refresh", DeviceToken: "dev-123"}

	if err := o.Refresh(sess); err != nil {
		t.Fatal(err)
	}
	if sess.AccessToken != "new-access" {
		t.Errorf("AccessToken = %q", sess.AccessToken)
	}
	if sess.RefreshToken != "new-refresh" {
		t.Errorf("RefreshToken = %q", sess.RefreshToken)
	}
	if sess.ExpiresAt.IsZero() {
		t.Error("ExpiresAt should be set")
	}
	if gotBody.Get("grant_type") != "refresh_token" {
		t.Errorf("grant_type = %q", gotBody.Get("grant_type"))
	}
	if gotBody.Get("refresh_token") != "old-refresh" {
		t.Errorf("refresh_token = %q", gotBody.Get("refresh_token"))
	}
	if gotBody.Get("device_token") != "dev-123" {
		t.Errorf("device_token = %q", gotBody.Get("device_token"))
	}
}

func TestRefreshAccessToken_InvalidGrant(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"error":"invalid_grant","error_description":"refresh token is invalid"}`)
	}))
	defer srv.Close()

	o := &oauth{baseURL: srv.URL, httpClient: srv.Client()}
	sess := &Session{RefreshToken: "bad"}

	err := o.Refresh(sess)
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("got %T", err)
	}
	if apiErr.Code != CodeSessionExpired {
		t.Errorf("Code = %s, want %s", apiErr.Code, CodeSessionExpired)
	}
}

func TestRefreshAccessToken_NoRefreshToken(t *testing.T) {
	o := &oauth{baseURL: "http://unused", httpClient: http.DefaultClient}
	sess := &Session{}
	err := o.Refresh(sess)
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.Code != CodeUnauthenticated {
		t.Errorf("got %v", err)
	}
}

func TestRefreshAccessToken_UnknownError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, "oops")
	}))
	defer srv.Close()

	o := &oauth{baseURL: srv.URL, httpClient: srv.Client()}
	sess := &Session{RefreshToken: "x"}

	err := o.Refresh(sess)
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("got %T", err)
	}
	if apiErr.Code != CodeRobinhoodUnavailable {
		t.Errorf("Code = %s", apiErr.Code)
	}
	if !strings.Contains(apiErr.Message, "500") {
		t.Errorf("Message = %q", apiErr.Message)
	}
}

// TestRefresh_SendsAPIVersionHeader is a Fix D regression guard: every
// oauth call must include the X-Robinhood-API-Version header that the
// Sheriff rollout cohort checks, plus a User-Agent so our calls are
// identifiable in their logs.
func TestRefresh_SendsAPIVersionHeader(t *testing.T) {
	var gotAPIVer, gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAPIVer = r.Header.Get("X-Robinhood-API-Version")
		gotUA = r.Header.Get("User-Agent")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "a",
			"refresh_token": "r",
			"expires_in":    86400,
		})
	}))
	defer srv.Close()

	o := &oauth{baseURL: srv.URL, httpClient: srv.Client()}
	sess := &Session{RefreshToken: "r"}
	if err := o.Refresh(sess); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if gotAPIVer == "" {
		t.Fatal("X-Robinhood-API-Version header missing on Refresh")
	}
	if gotUA == "" {
		t.Fatal("User-Agent header missing on Refresh")
	}
}
