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

func TestSession_MarshalRoundTrip(t *testing.T) {
	in := &Session{
		Username:     "alice@example.com",
		AccessToken:  "a.b.c",
		RefreshToken: "r.e.f",
		DeviceToken:  "dt-uuid",
		ExpiresAt:    time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC),
	}
	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	out, err := UnmarshalSession(data)
	if err != nil {
		t.Fatalf("UnmarshalSession: %v", err)
	}
	if out.Username != in.Username || out.AccessToken != in.AccessToken ||
		out.RefreshToken != in.RefreshToken || out.DeviceToken != in.DeviceToken ||
		!out.ExpiresAt.Equal(in.ExpiresAt) {
		t.Fatalf("round-trip mismatch: got %+v want %+v", out, in)
	}
	// Ephemeral is an in-memory flag — keychain-persisted sessions are never ephemeral.
	if out.Ephemeral {
		t.Fatalf("UnmarshalSession should return Ephemeral=false, got true")
	}
}

func TestSaveAndLoadFromKeychain(t *testing.T) {
	t.Setenv("ROBINHOOD_KEYCHAIN_BACKEND", "file")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	in := &Session{
		Username: "x@y", AccessToken: "a", RefreshToken: "r",
		DeviceToken: "d", ExpiresAt: time.Now().Add(time.Hour).UTC().Truncate(time.Second),
	}
	if err := in.SaveToKeychain("default"); err != nil {
		t.Fatalf("SaveToKeychain: %v", err)
	}
	out, err := LoadFromKeychain("default")
	if err != nil {
		t.Fatalf("LoadFromKeychain: %v", err)
	}
	if out.AccessToken != in.AccessToken {
		t.Fatalf("got %q want %q", out.AccessToken, in.AccessToken)
	}
	if out.Ephemeral {
		t.Fatalf("keychain-loaded session should never be ephemeral")
	}
	if err := ClearKeychain("default"); err != nil {
		t.Fatalf("ClearKeychain: %v", err)
	}
	if _, err := LoadFromKeychain("default"); err == nil {
		t.Fatalf("LoadFromKeychain after Clear should fail")
	}
}

func TestSaveToKeychain_RefusesEphemeral(t *testing.T) {
	t.Setenv("ROBINHOOD_KEYCHAIN_BACKEND", "file")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	s := &Session{AccessToken: "a", RefreshToken: "r", Ephemeral: true}
	if err := s.SaveToKeychain("default"); err == nil {
		t.Fatal("SaveToKeychain must refuse ephemeral sessions")
	}
}

func TestLoadSessionFromEnv_IsEphemeral(t *testing.T) {
	t.Setenv("ROBINHOOD_ACCESS_TOKEN", "acc")
	t.Setenv("ROBINHOOD_REFRESH_TOKEN", "ref")
	s, err := LoadSessionFromEnv()
	if err != nil {
		t.Fatalf("LoadSessionFromEnv: %v", err)
	}
	if !s.Ephemeral {
		t.Fatal("env-loaded session must have Ephemeral=true")
	}
}

func TestKeychain_InvalidProfileRejected(t *testing.T) {
	t.Setenv("ROBINHOOD_KEYCHAIN_BACKEND", "file")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	s := &Session{AccessToken: "a", RefreshToken: "r"}
	if err := s.SaveToKeychain("../evil"); err == nil {
		t.Fatal("SaveToKeychain must reject path-traversal profile")
	}
	if _, err := LoadFromKeychain("../evil"); err == nil {
		t.Fatal("LoadFromKeychain must reject path-traversal profile")
	}
	if err := ClearKeychain("../evil"); err == nil {
		t.Fatal("ClearKeychain must reject path-traversal profile")
	}
}
