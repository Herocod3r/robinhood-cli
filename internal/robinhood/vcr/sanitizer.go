// Package vcr provides shared helpers for cassette-based HTTP playback.
//
// The sanitizer hook is the PRIMARY defense against committing secrets
// into cassette YAML. It runs at record time (BeforeSaveHook) and
// redacts:
//
//   - the Authorization header (whatever the caller sent)
//   - JSON fields access_token, refresh_token, device_token in bodies
//
// Callers add the hook to a recorder with:
//
//	r.AddHook(vcr.SanitizeHook, recorder.BeforeSaveHook)
//
// There is no standalone external sanitizer script — the hook is the
// single source of truth. A CI gitleaks scan runs as a second-pass
// check over testdata/cassettes to fail closed if an unredacted token
// ever slips in.
package vcr

import (
	"regexp"

	"gopkg.in/dnaeon/go-vcr.v3/cassette"
)

// accessTokenRe matches JSON fields `access_token`, `refresh_token`,
// `device_token` with string values of any length. We redact the VALUE
// while preserving the KEY so shape-based consumers still parse.
var tokenFieldRe = regexp.MustCompile(`("(?:access_token|refresh_token|device_token)"\s*:\s*")[^"]*(")`)

// bearerRe matches `Bearer <token>` sequences anywhere in a body
// (covers cases where the server echoes back the header).
var bearerRe = regexp.MustCompile(`(?i)(bearer\s+)[A-Za-z0-9_\-\.]+`)

// redactString applies token redaction to an arbitrary string.
func redactString(s string) string {
	s = tokenFieldRe.ReplaceAllString(s, `${1}REDACTED${2}`)
	s = bearerRe.ReplaceAllString(s, `${1}REDACTED`)
	return s
}

// SanitizeHook is a go-vcr BeforeSaveHook that redacts secrets from an
// Interaction before it is persisted to disk.
//
// It is safe to add multiple times, and safe to run on already-sanitized
// data (idempotent).
func SanitizeHook(i *cassette.Interaction) error {
	// Redact the Authorization request header unconditionally.
	if i.Request.Headers != nil {
		if _, ok := i.Request.Headers["Authorization"]; ok {
			i.Request.Headers.Set("Authorization", "Bearer REDACTED")
		}
	}
	// Response headers can also echo Authorization in some flows.
	if i.Response.Headers != nil {
		if _, ok := i.Response.Headers["Authorization"]; ok {
			i.Response.Headers.Set("Authorization", "Bearer REDACTED")
		}
	}

	// Bodies: request body may carry `access_token`/`refresh_token` (OAuth
	// grants echo the input), response body carries them on login/refresh.
	i.Request.Body = redactString(i.Request.Body)
	i.Response.Body = redactString(i.Response.Body)
	return nil
}
