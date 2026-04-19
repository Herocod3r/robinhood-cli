package robinhood

import (
	"os"
	"time"

	"github.com/herocod3r/robinhood-cli/internal/config"
)

// refreshBuffer is how long before ExpiresAt we pre-emptively refresh.
const refreshBuffer = 5 * time.Minute

// Session holds tokens and device identity for a single Robinhood account.
type Session struct {
	Username     string
	AccessToken  string
	RefreshToken string
	DeviceToken  string
	ExpiresAt    time.Time
}

// IsExpired returns true ONLY when ExpiresAt is known and at/within the
// refresh buffer. Zero ExpiresAt means "unknown" — callers must NOT treat
// unknown as expired. The client only proactively refreshes on known-expired;
// unknown tokens go to the endpoint first and refresh on 401.
func (s *Session) IsExpired() bool {
	if s.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().Add(refreshBuffer).After(s.ExpiresAt)
}

// NeedsImmediateRefresh reports whether a request cannot go out with the
// current tokens — i.e., no access token (but refresh available), or a
// known-expired access token. Unknown expiry with a present access token
// returns false (trust the access token; let 401 drive refresh).
func (s *Session) NeedsImmediateRefresh() bool {
	if s.AccessToken == "" {
		return s.RefreshToken != ""
	}
	return s.IsExpired()
}

// LoadSessionFromEnv builds a Session from env vars.
// Returns CodeUnauthenticated if access+refresh tokens are both missing.
// ExpiresAt is left zero (unknown) — the client will use the access token
// until a 401, then refresh once.
func LoadSessionFromEnv() (*Session, error) {
	access := os.Getenv(config.EnvAccessToken)
	refresh := os.Getenv(config.EnvRefreshToken)
	if access == "" && refresh == "" {
		return nil, &APIError{
			Code:    CodeUnauthenticated,
			Message: "no access/refresh tokens found",
			Hint:    "set ROBINHOOD_ACCESS_TOKEN and ROBINHOOD_REFRESH_TOKEN, or run: rh login (Plan B)",
		}
	}
	return &Session{
		Username:     os.Getenv(config.EnvUsername),
		AccessToken:  access,
		RefreshToken: refresh,
		DeviceToken:  os.Getenv(config.EnvDeviceToken),
		// ExpiresAt left zero (unknown) — see Session.IsExpired / NeedsImmediateRefresh.
	}, nil
}
