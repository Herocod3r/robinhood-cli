package robinhood

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/herocod3r/robinhood-cli/internal/config"
	"github.com/herocod3r/robinhood-cli/internal/keychain"
)

// refreshBuffer is how long before ExpiresAt we pre-emptively refresh.
const refreshBuffer = 5 * time.Minute

// keychainService is the service name used in the OS keychain.
const keychainService = "robinhood-cli"

// Session holds tokens and device identity for a single Robinhood account.
type Session struct {
	Username     string
	AccessToken  string
	RefreshToken string
	DeviceToken  string
	ExpiresAt    time.Time

	// Ephemeral is true for sessions loaded from env vars — a refresh
	// against such a session must NOT write back to the keychain
	// (another profile may own the keychain entry). See Fix N.
	Ephemeral bool
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

// sessionBlob is the persisted shape. Kept separate from Session so the
// in-memory struct can evolve without breaking stored blobs. Ephemeral
// is deliberately NOT persisted — by definition a keychain-stored session
// is not ephemeral, and LoadFromKeychain always returns Ephemeral=false.
type sessionBlob struct {
	Username     string `json:"username"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	DeviceToken  string `json:"device_token"`
	ExpiresAt    string `json:"expires_at,omitempty"` // RFC3339; empty => zero
}

// Marshal serializes the session to bytes. ExpiresAt is RFC3339 UTC or empty.
func (s *Session) Marshal() ([]byte, error) {
	blob := sessionBlob{
		Username:     s.Username,
		AccessToken:  s.AccessToken,
		RefreshToken: s.RefreshToken,
		DeviceToken:  s.DeviceToken,
	}
	if !s.ExpiresAt.IsZero() {
		blob.ExpiresAt = s.ExpiresAt.UTC().Format(time.RFC3339)
	}
	return json.Marshal(blob)
}

// UnmarshalSession parses bytes produced by Marshal.
func UnmarshalSession(data []byte) (*Session, error) {
	var blob sessionBlob
	if err := json.Unmarshal(data, &blob); err != nil {
		return nil, err
	}
	s := &Session{
		Username:     blob.Username,
		AccessToken:  blob.AccessToken,
		RefreshToken: blob.RefreshToken,
		DeviceToken:  blob.DeviceToken,
	}
	if blob.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, blob.ExpiresAt)
		if err != nil {
			return nil, err
		}
		s.ExpiresAt = t.UTC()
	}
	return s, nil
}

// SaveToKeychain persists the session under the given profile.
// Refuses to run for ephemeral sessions (env-loaded) per Fix N.
func (s *Session) SaveToKeychain(profile string) error {
	if err := config.ValidProfile(profile); err != nil {
		return err
	}
	if s.Ephemeral {
		return fmt.Errorf("refusing to save ephemeral (env-loaded) session to keychain")
	}
	data, err := s.Marshal()
	if err != nil {
		return err
	}
	store, err := keychain.Open()
	if err != nil {
		return err
	}
	return store.Set(keychainService, profile, data)
}

// LoadFromKeychain returns the session persisted under profile, or an
// APIError{Code:CodeUnauthenticated} if none exists.
func LoadFromKeychain(profile string) (*Session, error) {
	if err := config.ValidProfile(profile); err != nil {
		return nil, err
	}
	store, err := keychain.Open()
	if err != nil {
		return nil, &APIError{Code: CodeUnauthenticated, Message: err.Error(), Hint: "run: rh login"}
	}
	data, err := store.Get(keychainService, profile)
	if err != nil {
		if errors.Is(err, keychain.ErrNotFound) {
			return nil, &APIError{
				Code:    CodeUnauthenticated,
				Message: "no session for profile " + profile,
				Hint:    "run: rh login",
			}
		}
		return nil, &APIError{Code: CodeUnauthenticated, Message: err.Error(), Hint: "run: rh login"}
	}
	return UnmarshalSession(data)
}

// ClearKeychain removes the session blob for the given profile.
// Returns nil if nothing was stored.
func ClearKeychain(profile string) error {
	if err := config.ValidProfile(profile); err != nil {
		return err
	}
	store, err := keychain.Open()
	if err != nil {
		return err
	}
	err = store.Delete(keychainService, profile)
	if err != nil && !errors.Is(err, keychain.ErrNotFound) {
		return err
	}
	return nil
}

// LoadSessionFromEnv builds a Session from env vars.
// Returns CodeUnauthenticated if access+refresh tokens are both missing.
// ExpiresAt is left zero (unknown) — the client will use the access token
// until a 401, then refresh once. Ephemeral is set to true so the client's
// refresh path never writes these env-sourced tokens back to the keychain.
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
		Ephemeral:    true,
		// ExpiresAt left zero (unknown) — see Session.IsExpired / NeedsImmediateRefresh.
	}, nil
}
