package robinhood

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// robinhoodOAuthClientID is the same client_id the mobile/web app uses, copied from
// robin_stocks. Unofficial; may be rotated by Robinhood.
const robinhoodOAuthClientID = "c82SH0WZOsabOXGP2sxqcj34FxkvfnWRZBKlBjFS"

// oauth owns grant types and token refresh.
type oauth struct {
	baseURL    string
	httpClient *http.Client
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type oauthErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// Refresh exchanges the session's RefreshToken for new tokens, mutating the session in place.
func (o *oauth) Refresh(s *Session) error {
	if s.RefreshToken == "" {
		return &APIError{
			Code:    CodeUnauthenticated,
			Message: "no refresh token available",
			Hint:    "run: rh login",
		}
	}
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", s.RefreshToken)
	form.Set("client_id", robinhoodOAuthClientID)
	form.Set("expires_in", "86400")
	form.Set("scope", "internal")
	if s.DeviceToken != "" {
		form.Set("device_token", s.DeviceToken)
	}

	req, err := http.NewRequest(http.MethodPost, o.baseURL+"/oauth2/token/", strings.NewReader(form.Encode()))
	if err != nil {
		return &APIError{Code: CodeRobinhoodUnavailable, Message: err.Error()}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	// Sheriff rollout cohorts reject oauth calls without this header (Fix D).
	req.Header.Set("X-Robinhood-API-Version", "1.431.4")
	req.Header.Set("User-Agent", "robinhood-cli (+https://github.com/herocod3r/robinhood-cli)")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return &APIError{Code: CodeRobinhoodUnavailable, Message: err.Error()}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	switch {
	case resp.StatusCode == http.StatusOK:
		var tr tokenResponse
		if err := json.Unmarshal(body, &tr); err != nil {
			return &APIError{Code: CodeRobinhoodUnavailable, Message: "malformed token response"}
		}
		s.AccessToken = tr.AccessToken
		if tr.RefreshToken != "" {
			s.RefreshToken = tr.RefreshToken
		}
		if tr.ExpiresIn > 0 {
			s.ExpiresAt = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
		}
		return nil

	case resp.StatusCode == http.StatusBadRequest, resp.StatusCode == http.StatusUnauthorized:
		var er oauthErrorResponse
		_ = json.Unmarshal(body, &er)
		return &APIError{
			Code:       CodeSessionExpired,
			Message:    firstNonEmpty(er.ErrorDescription, er.Error, "refresh rejected"),
			Hint:       "run: rh login",
			HTTPStatus: resp.StatusCode,
		}

	default:
		return &APIError{
			Code:       CodeRobinhoodUnavailable,
			Message:    fmt.Sprintf("oauth refresh HTTP %d", resp.StatusCode),
			HTTPStatus: resp.StatusCode,
		}
	}
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}

// PasswordGrant exchanges username+password (and optional MFA code) for tokens.
// Returns (*Session, nil) on success. On 400 with verification_workflow,
// returns APIError{Code:CodeSheriffRequired, WorkflowID: ...}. On 400 with
// mfa_required, returns APIError{Code:CodeMFARequired}.
func (o *oauth) PasswordGrant(ctx context.Context, username, password, deviceToken, mfaCode string) (*Session, error) {
	form := baseGrantForm(username, password, deviceToken, mfaCode)
	return o.passwordGrant(ctx, form, "")
}

// PasswordGrantWithWorkflow performs the post-Sheriff re-attempt. The
// X-Robinhood-Challenge-Response-Id header correlates the token POST with
// the workflow the Sheriff state machine just validated. See Task 11 + Fix E.
func (o *oauth) PasswordGrantWithWorkflow(ctx context.Context, username, password, deviceToken, mfaCode, workflowID string) (*Session, error) {
	form := baseGrantForm(username, password, deviceToken, mfaCode)
	return o.passwordGrant(ctx, form, workflowID)
}

// baseGrantForm builds the x-www-form-urlencoded payload shared by the two
// public grant entry points. Fields match robin_stocks login_payload in
// robin_stocks/robinhood/authentication.py (master branch, checked 2026-04-20):
// grant_type=password, client_id, username, password, device_token,
// expires_in=86400, scope=internal, mfa_code (optional).
func baseGrantForm(username, password, deviceToken, mfaCode string) url.Values {
	form := url.Values{}
	form.Set("grant_type", "password")
	form.Set("username", username)
	form.Set("password", password)
	form.Set("client_id", robinhoodOAuthClientID)
	form.Set("device_token", deviceToken)
	form.Set("expires_in", "86400")
	form.Set("scope", "internal")
	if mfaCode != "" {
		form.Set("mfa_code", mfaCode)
	}
	return form
}

// passwordGrant is the single shared workhorse that actually posts to
// /oauth2/token/. When workflowID is non-empty the X-Robinhood-Challenge-
// Response-Id header is added — that's the bridge between Sheriff success
// and the final token exchange (Task 11).
func (o *oauth) passwordGrant(ctx context.Context, form url.Values, workflowID string) (*Session, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL+"/oauth2/token/", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, &APIError{Code: CodeRobinhoodUnavailable, Message: err.Error()}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	// Sheriff rollout cohorts reject oauth calls without this header (Fix D).
	req.Header.Set("X-Robinhood-API-Version", "1.431.4")
	req.Header.Set("User-Agent", "robinhood-cli (+https://github.com/herocod3r/robinhood-cli)")
	if workflowID != "" {
		req.Header.Set("X-Robinhood-Challenge-Response-Id", workflowID)
	}

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, &APIError{Code: CodeRobinhoodUnavailable, Message: err.Error()}
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
		var tr tokenResponse
		if err := json.Unmarshal(body, &tr); err != nil {
			return nil, &APIError{Code: CodeRobinhoodUnavailable, Message: "malformed token response"}
		}
		s := &Session{
			Username:     form.Get("username"),
			AccessToken:  tr.AccessToken,
			RefreshToken: tr.RefreshToken,
			DeviceToken:  form.Get("device_token"),
		}
		if tr.ExpiresIn > 0 {
			s.ExpiresAt = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second).UTC()
		}
		return s, nil

	case http.StatusBadRequest, http.StatusUnauthorized:
		return nil, classifyPasswordErr(body, resp.StatusCode)

	default:
		return nil, &APIError{
			Code:       CodeRobinhoodUnavailable,
			Message:    fmt.Sprintf("password grant HTTP %d", resp.StatusCode),
			HTTPStatus: resp.StatusCode,
		}
	}
}

// passwordErrResp is the union of error body shapes we've seen in the wild.
// Some rollout cohorts nest the verification id under verification_workflow;
// others return a bare top-level "id". TODO: verify against real response —
// robin_stocks master uses only the nested shape.
type passwordErrResp struct {
	MFARequired          bool   `json:"mfa_required"`
	MFAType              string `json:"mfa_type"`
	VerificationWorkflow struct {
		ID             string `json:"id"`
		WorkflowStatus string `json:"workflow_status"`
	} `json:"verification_workflow"`
	// Top-level id — seen when the server inlines the workflow itself
	// rather than wrapping it (observed in some cohorts).
	ID               string `json:"id"`
	Detail           string `json:"detail"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func classifyPasswordErr(body []byte, status int) error {
	var er passwordErrResp
	_ = json.Unmarshal(body, &er)
	// Accept both verification_workflow.id (robin_stocks) and a bare
	// top-level id (observed in some rollout cohorts).
	workflowID := er.VerificationWorkflow.ID
	if workflowID == "" && er.ID != "" {
		workflowID = er.ID
	}
	if workflowID != "" {
		return &APIError{
			Code:       CodeSheriffRequired,
			Message:    "device verification required",
			Hint:       "complete the challenge shown in your Robinhood app or SMS",
			HTTPStatus: status,
			WorkflowID: workflowID,
		}
	}
	if er.MFARequired {
		hint := "enter your TOTP code"
		if er.MFAType == "sms" {
			hint = "enter the SMS code Robinhood just sent"
		}
		return &APIError{
			Code:       CodeMFARequired,
			Message:    "MFA required",
			Hint:       hint,
			HTTPStatus: status,
		}
	}
	msg := firstNonEmpty(er.Detail, er.ErrorDescription, er.Error, "password grant rejected")
	return &APIError{Code: CodeUnauthenticated, Message: msg, HTTPStatus: status}
}
