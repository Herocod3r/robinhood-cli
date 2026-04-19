package robinhood

import (
	"testing"
	"time"
)

func TestSession_LoadFromEnv_Present(t *testing.T) {
	t.Setenv("ROBINHOOD_ACCESS_TOKEN", "acc-tok")
	t.Setenv("ROBINHOOD_REFRESH_TOKEN", "ref-tok")
	t.Setenv("ROBINHOOD_DEVICE_TOKEN", "device-xyz")

	s, err := LoadSessionFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if s.AccessToken != "acc-tok" {
		t.Errorf("AccessToken = %q", s.AccessToken)
	}
	if s.RefreshToken != "ref-tok" {
		t.Errorf("RefreshToken = %q", s.RefreshToken)
	}
	if s.DeviceToken != "device-xyz" {
		t.Errorf("DeviceToken = %q", s.DeviceToken)
	}
}

func TestSession_LoadFromEnv_Missing(t *testing.T) {
	t.Setenv("ROBINHOOD_ACCESS_TOKEN", "")
	t.Setenv("ROBINHOOD_REFRESH_TOKEN", "")

	_, err := LoadSessionFromEnv()
	if err == nil {
		t.Fatal("expected error when tokens missing")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Code != CodeUnauthenticated {
		t.Errorf("Code = %s, want %s", apiErr.Code, CodeUnauthenticated)
	}
}

// IsExpired must return true ONLY when ExpiresAt is known and at/within the
// refresh buffer. Zero ExpiresAt = unknown = false (trust access token,
// refresh only on 401). Guards Codex finding #2: env-loaded sessions must
// not burn their refresh token on every invocation.
func TestSession_IsExpired(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name   string
		expiry time.Time
		want   bool
	}{
		{"zero time (unknown)", time.Time{}, false}, // unknown → do not proactively refresh
		{"past", now.Add(-1 * time.Hour), true},
		{"within buffer", now.Add(1 * time.Minute), true},
		{"well ahead", now.Add(1 * time.Hour), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := &Session{ExpiresAt: c.expiry}
			if got := s.IsExpired(); got != c.want {
				t.Errorf("IsExpired() = %v, want %v", got, c.want)
			}
		})
	}
}

func TestSession_NeedsImmediateRefresh(t *testing.T) {
	cases := []struct {
		name string
		sess *Session
		want bool
	}{
		{"access empty, refresh present", &Session{AccessToken: "", RefreshToken: "r"}, true},
		{"access present, unknown expiry", &Session{AccessToken: "a"}, false},
		{"both empty", &Session{}, false},
		{"access present, past expiry", &Session{AccessToken: "a", ExpiresAt: time.Now().Add(-time.Hour)}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.sess.NeedsImmediateRefresh(); got != c.want {
				t.Errorf("NeedsImmediateRefresh() = %v, want %v", got, c.want)
			}
		})
	}
}
