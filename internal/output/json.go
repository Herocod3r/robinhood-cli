package output

import (
	"encoding/json"
	"io"
	"time"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

// JSONWriter renders Envelopes as JSON.
type JSONWriter struct {
	Out    io.Writer
	Indent bool
	// Now is an injectable clock for deterministic tests.
	Now func() time.Time
}

func (w *JSONWriter) now() time.Time {
	if w.Now != nil {
		return w.Now()
	}
	return nowUTC()
}

// WriteSuccess emits a success envelope.
func (w *JSONWriter) WriteSuccess(command string, data any, meta any) error {
	rawData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	var rawMeta json.RawMessage
	if meta != nil {
		rawMeta, err = json.Marshal(meta)
		if err != nil {
			return err
		}
	}
	env := Envelope{
		Schema:      SchemaVersion(),
		Command:     command,
		GeneratedAt: w.now().Format(time.RFC3339),
		Data:        rawData,
		Meta:        rawMeta,
		Error:       nil,
	}
	return w.encode(env)
}

// WriteError emits an error envelope.
func (w *JSONWriter) WriteError(command string, err error) error {
	payload := &ErrorPayload{Code: "unknown", Message: err.Error()}
	if apiErr, ok := err.(*robinhood.APIError); ok {
		payload = &ErrorPayload{
			Code:      string(apiErr.Code),
			Message:   apiErr.Message,
			Hint:      apiErr.Hint,
			Retryable: apiErr.Retryable,
		}
	}
	env := Envelope{
		Schema:      SchemaVersion(),
		Command:     command,
		GeneratedAt: w.now().Format(time.RFC3339),
		Data:        json.RawMessage("null"),
		Error:       payload,
	}
	return w.encode(env)
}

func (w *JSONWriter) encode(env Envelope) error {
	enc := json.NewEncoder(w.Out)
	if w.Indent {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(env)
}
