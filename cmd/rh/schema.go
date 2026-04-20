package main

import (
	"os"

	"github.com/herocod3r/robinhood-cli/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	c := &cobra.Command{
		Use:   "schema [command]",
		Short: "Emit JSON schema information (envelope and per-command data shapes)",
		Args:  cobra.MaximumNArgs(1),
		Annotations: map[string]string{
			"category":  "meta",
			"stability": "green",
		},
		RunE: runSchema,
	}
	register(c)
}

// envelopeSchema is a hand-written description of the stable envelope shape.
// Expanded into per-command data shapes in Plan B as endpoints land.
var envelopeSchema = map[string]any{
	"envelope": map[string]any{
		"type": "object",
		"fields": map[string]any{
			"schema":       map[string]any{"type": "string", "const": "robinhood-cli/v1"},
			"command":      map[string]any{"type": "string"},
			"generated_at": map[string]any{"type": "string", "format": "date-time"},
			"data":         map[string]any{"type": "any", "nullable": true},
			"meta":         map[string]any{"type": "object", "nullable": true},
			"error":        map[string]any{"type": "object", "nullable": true},
		},
	},
	"error": map[string]any{
		"type": "object",
		"fields": map[string]any{
			"code":      map[string]any{"type": "string", "enum": []string{"unauthenticated", "session_expired", "sheriff_required", "rate_limited", "robinhood_unavailable", "not_found", "validation"}},
			"message":   map[string]any{"type": "string"},
			"hint":      map[string]any{"type": "string"},
			"retryable": map[string]any{"type": "bool"},
		},
	},
}

func runSchema(cmd *cobra.Command, args []string) error {
	w := &output.JSONWriter{Out: os.Stdout}

	if len(args) == 1 {
		name := args[0]
		if name == "envelope" {
			return w.WriteSuccess("schema", envelopeSchema["envelope"], map[string]string{"kind": "envelope"})
		}
		// Plan A only exposes the envelope; data schemas come in Plan B.
		return w.WriteSuccess("schema", map[string]string{
			"note": "per-command data schemas land in Plan B",
		}, map[string]string{"kind": "unknown"})
	}
	return w.WriteSuccess("schema", envelopeSchema, nil)
}
