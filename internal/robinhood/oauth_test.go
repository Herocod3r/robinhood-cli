package robinhood

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
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

func TestPasswordGrant_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth2/token/" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if got := r.Form.Get("grant_type"); got != "password" {
			t.Fatalf("grant_type = %s", got)
		}
		if r.Form.Get("device_token") == "" {
			t.Fatalf("missing device_token")
		}
		if r.Header.Get("X-Robinhood-API-Version") == "" {
			t.Fatalf("missing X-Robinhood-API-Version header")
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"a","refresh_token":"r","expires_in":3600,"token_type":"Bearer"}`)
	}))
	defer ts.Close()

	o := &oauth{baseURL: ts.URL, httpClient: ts.Client()}
	sess, err := o.PasswordGrant(context.Background(), "u", "p", "dev-token", "")
	if err != nil {
		t.Fatalf("PasswordGrant: %v", err)
	}
	if sess.AccessToken != "a" {
		t.Fatalf("AccessToken = %q", sess.AccessToken)
	}
	if sess.DeviceToken != "dev-token" {
		t.Fatalf("DeviceToken = %q", sess.DeviceToken)
	}
	if sess.ExpiresAt.IsZero() {
		t.Fatalf("ExpiresAt should be set")
	}
}

func TestPasswordGrant_SheriffRequired(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"verification_workflow":{"id":"wf-123","workflow_status":"internal_pending"}}`)
	}))
	defer ts.Close()
	o := &oauth{baseURL: ts.URL, httpClient: ts.Client()}
	_, err := o.PasswordGrant(context.Background(), "u", "p", "dev", "")
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("err = %v, want APIError", err)
	}
	if apiErr.Code != CodeSheriffRequired {
		t.Fatalf("code = %q, want %q", apiErr.Code, CodeSheriffRequired)
	}
	if apiErr.WorkflowID != "wf-123" {
		t.Fatalf("workflow_id = %q, want %q", apiErr.WorkflowID, "wf-123")
	}
}

// Codex Fix E: robin_stocks checks `data['verification_workflow']['id']`.
// Some rollout cohorts instead return the id at the top level. Our
// classifier must accept either shape.
func TestPasswordGrant_SheriffRequired_TopLevelID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"id":"wf-top","detail":"verification required"}`)
	}))
	defer ts.Close()
	o := &oauth{baseURL: ts.URL, httpClient: ts.Client()}
	_, err := o.PasswordGrant(context.Background(), "u", "p", "dev", "")
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.Code != CodeSheriffRequired {
		t.Fatalf("err = %v, want CodeSheriffRequired", err)
	}
	if apiErr.WorkflowID != "wf-top" {
		t.Fatalf("workflow_id = %q, want %q", apiErr.WorkflowID, "wf-top")
	}
}

func TestPasswordGrant_MFARequired(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"mfa_required":true,"mfa_type":"app"}`)
	}))
	defer ts.Close()
	o := &oauth{baseURL: ts.URL, httpClient: ts.Client()}
	_, err := o.PasswordGrant(context.Background(), "u", "p", "dev", "")
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.Code != CodeMFARequired {
		t.Fatalf("err = %v, want CodeMFARequired", err)
	}
}

func TestPasswordGrant_WithMFACode(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.Form.Get("mfa_code") == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"mfa_required":true}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"a","refresh_token":"r","expires_in":3600}`)
	}))
	defer ts.Close()
	o := &oauth{baseURL: ts.URL, httpClient: ts.Client()}
	_, err := o.PasswordGrant(context.Background(), "u", "p", "dev", "123456")
	if err != nil {
		t.Fatalf("PasswordGrant with MFA code: %v", err)
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

func TestTOTPCode_Deterministic(t *testing.T) {
	// Known test vector: a valid base32 secret should produce a 6-digit code.
	got, err := TOTPCodeAt("JBSWY3DPEHPK3PXP", time.Unix(59, 0).UTC())
	if err != nil {
		t.Fatalf("TOTPCodeAt: %v", err)
	}
	if len(got) != 6 {
		t.Fatalf("len = %d, want 6", len(got))
	}
}

func TestTOTPCode_RejectsBadSecret(t *testing.T) {
	if _, err := TOTPCodeAt("not-base32!!!", time.Now()); err == nil {
		t.Fatalf("expected error for invalid secret")
	}
}
