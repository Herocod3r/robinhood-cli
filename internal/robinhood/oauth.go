package robinhood

import (
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
