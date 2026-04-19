// Package output renders command results as JSON or pretty tables.
package output

import (
	"encoding/json"
	"time"

	"github.com/herocod3r/robinhood-cli/internal/buildinfo"
)

// Envelope is the stable JSON shape emitted by every command.
type Envelope struct {
	Schema      string          `json:"schema"`
	Command     string          `json:"command"`
	GeneratedAt string          `json:"generated_at"`
	Data        json.RawMessage `json:"data"`
	Meta        json.RawMessage `json:"meta,omitempty"`
	Error       *ErrorPayload   `json:"error"`
}

// ErrorPayload is the JSON-facing shape of a robinhood.APIError.
type ErrorPayload struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Hint      string `json:"hint,omitempty"`
	Retryable bool   `json:"retryable"`
}

// Meta holds optional metadata such as counts.
type Meta struct {
	Count   int    `json:"count,omitempty"`
	Profile string `json:"profile,omitempty"`
}

// SchemaVersion returns the canonical schema identifier.
func SchemaVersion() string {
	return buildinfo.SchemaVersion
}

// nowUTC defaults time provider used by writers; tests override.
func nowUTC() time.Time {
	return time.Now().UTC()
}
