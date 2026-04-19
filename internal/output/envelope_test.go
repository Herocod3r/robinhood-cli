package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

func TestJSON_WriteSuccess(t *testing.T) {
	var buf bytes.Buffer
	w := &JSONWriter{Out: &buf, Now: func() time.Time { return time.Date(2026, 4, 18, 18, 42, 3, 0, time.UTC) }}

	data := map[string]string{"equity": "98234.50"}
	if err := w.WriteSuccess("portfolio", data, nil); err != nil {
		t.Fatal(err)
	}

	var env Envelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if env.Schema != "robinhood-cli/v1" {
		t.Errorf("Schema = %q", env.Schema)
	}
	if env.Command != "portfolio" {
		t.Errorf("Command = %q", env.Command)
	}
	if env.GeneratedAt != "2026-04-18T18:42:03Z" {
		t.Errorf("GeneratedAt = %q", env.GeneratedAt)
	}
	if env.Error != nil {
		t.Errorf("Error should be nil, got %+v", env.Error)
	}
	if !strings.Contains(buf.String(), "\"equity\":\"98234.50\"") {
		t.Errorf("data not embedded: %s", buf.String())
	}
}

func TestJSON_WriteError(t *testing.T) {
	var buf bytes.Buffer
	w := &JSONWriter{Out: &buf, Now: func() time.Time { return time.Unix(0, 0).UTC() }}

	apiErr := &robinhood.APIError{
		Code:    robinhood.CodeSessionExpired,
		Message: "refresh rejected",
		Hint:    "run: rh login",
	}
	if err := w.WriteError("portfolio", apiErr); err != nil {
		t.Fatal(err)
	}

	var env Envelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatal(err)
	}

	// The wire shape we promise is `"data": null`. After unmarshalling into
	// json.RawMessage, that is the non-nil byte slice []byte("null") — so
	// assert on the raw bytes, not `== nil`.
	if string(env.Data) != "null" {
		t.Errorf("Data raw = %q, want %q", string(env.Data), "null")
	}
	if env.Error == nil {
		t.Fatal("Error should be set")
	}
	if env.Error.Code != "session_expired" {
		t.Errorf("Error.Code = %q", env.Error.Code)
	}
	if env.Error.Hint != "run: rh login" {
		t.Errorf("Error.Hint = %q", env.Error.Hint)
	}
}

// TestJSON_ErrorEnvelope_WireShape pins the on-the-wire JSON shape directly
// so the contract survives any later refactor of Envelope internals.
func TestJSON_ErrorEnvelope_WireShape(t *testing.T) {
	var buf bytes.Buffer
	w := &JSONWriter{Out: &buf, Now: func() time.Time { return time.Unix(0, 0).UTC() }}
	if err := w.WriteError("x", &robinhood.APIError{Code: robinhood.CodeNotFound, Message: "m"}); err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &raw); err != nil {
		t.Fatal(err)
	}
	if string(raw["data"]) != "null" {
		t.Errorf(`expected "data": null on the wire, got %s`, string(raw["data"]))
	}
	if string(raw["error"]) == "null" {
		t.Error("error should not be null on error envelope")
	}
	if string(raw["schema"]) != `"robinhood-cli/v1"` {
		t.Errorf("schema = %s", string(raw["schema"]))
	}
}

func TestJSON_ErrorEnvelope_IsSingleDocument(t *testing.T) {
	var buf bytes.Buffer
	w := &JSONWriter{Out: &buf, Now: func() time.Time { return time.Unix(0, 0).UTC() }}
	if err := w.WriteError("x", &robinhood.APIError{Code: robinhood.CodeValidation}); err != nil {
		t.Fatal(err)
	}
	dec := json.NewDecoder(&buf)
	var first any
	if err := dec.Decode(&first); err != nil {
		t.Fatal(err)
	}
	// Any trailing bytes after one document = contract violation.
	var trailing any
	if err := dec.Decode(&trailing); err == nil {
		t.Errorf("expected single JSON document, got trailing content %v", trailing)
	}
}
