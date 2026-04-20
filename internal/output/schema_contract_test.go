//go:build contract

package output

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// noDataCommands is the explicit allowlist of commands that do not
// emit a typed `data` payload. Every other heading in
// docs/JSON_SCHEMA.md MUST have a golden in testdata/contract/.
//
// Keep this in sync with the "Meta commands" section in
// docs/JSON_SCHEMA.md.
var noDataCommands = map[string]bool{
	"login":    true, // no JSON data
	"logout":   true, // no JSON data
	"skill":    true, // no JSON data (subcommand prints progress)
	"version":  true, // trivial {version, commit, schema} — covered by envelope
	"commands": true, // self-describing meta payload
	"schema":   true, // echoes the envelope schema itself
}

// TestContract_EnvelopeSchema extracts each JSON schema from
// docs/JSON_SCHEMA.md and validates the matching golden file under
// testdata/contract/ parses against it.
//
// Parser grammar (Fix CI):
//   - Headings must have the exact form `### `rh <cmd>`` — single
//     backtick-quoted command, no arguments, no angle brackets.
//   - The schema fence must be the FIRST ```json block after the
//     heading (NOT ```jsonc — those are examples).
func TestContract_EnvelopeSchema(t *testing.T) {
	docPath := filepath.Join("..", "..", "docs", "JSON_SCHEMA.md")
	doc, err := os.ReadFile(docPath)
	if err != nil {
		t.Fatalf("read %s: %v", docPath, err)
	}
	schemas := extractSchemas(string(doc))
	if len(schemas) == 0 {
		t.Fatalf("no schemas extracted from %s", docPath)
	}

	goldens := loadGoldens(t)

	// Assert every schema has a golden, unless on the noDataCommands allowlist.
	for name := range schemas {
		if noDataCommands[name] {
			continue
		}
		if _, ok := goldens[name]; !ok {
			t.Errorf("schema for %q has no golden in testdata/contract/", name)
		}
	}

	for command, rawSchema := range schemas {
		golden, ok := goldens[command]
		if !ok {
			continue
		}
		s, err := jsonschema.CompileString(command+".json", rawSchema)
		if err != nil {
			t.Errorf("%s: schema compile: %v", command, err)
			continue
		}
		var v any
		if err := json.Unmarshal(golden, &v); err != nil {
			t.Errorf("%s: golden parse: %v", command, err)
			continue
		}
		if err := s.Validate(v); err != nil {
			t.Errorf("%s: golden does not match schema: %v", command, err)
		}
	}
}

// extractSchemas parses docs/JSON_SCHEMA.md and returns
// map[command] -> schema JSON string.
//
// It only recognizes headings of the exact form:
//
//	### `rh <command>`
//
// The command token must not contain spaces, angle brackets, or
// arguments — subcommand schemas should be documented as sub-headings
// or folded into the parent. The first ```json fence after each
// heading is the schema; later ```jsonc fences are examples and are
// ignored.
func extractSchemas(doc string) map[string]string {
	out := map[string]string{}
	lines := strings.Split(doc, "\n")
	for i := 0; i < len(lines); i++ {
		name, ok := parseRhHeading(lines[i])
		if !ok {
			continue
		}
		// Find the next ```json fence (not ```jsonc).
		for j := i + 1; j < len(lines); j++ {
			// Stop if we hit another ### heading before the fence.
			if strings.HasPrefix(lines[j], "### ") {
				break
			}
			line := lines[j]
			if !isJSONFenceOpen(line) {
				continue
			}
			// Collect until closing fence.
			var buf []string
			closed := false
			for k := j + 1; k < len(lines); k++ {
				if strings.HasPrefix(strings.TrimSpace(lines[k]), "```") {
					out[name] = strings.Join(buf, "\n")
					i = k
					closed = true
					break
				}
				buf = append(buf, lines[k])
			}
			if !closed {
				// Malformed; skip.
			}
			break
		}
	}
	return out
}

// parseRhHeading returns the command name if `line` is exactly
// "### `rh <command>`" with no extra tokens.
func parseRhHeading(line string) (string, bool) {
	const prefix = "### `rh "
	const suffix = "`"
	if !strings.HasPrefix(line, prefix) || !strings.HasSuffix(line, suffix) {
		return "", false
	}
	inner := strings.TrimSuffix(strings.TrimPrefix(line, prefix), suffix)
	inner = strings.TrimSpace(inner)
	if inner == "" {
		return "", false
	}
	// Reject arguments, angle brackets, flag markers.
	if strings.ContainsAny(inner, " <>[]") {
		return "", false
	}
	return inner, true
}

// isJSONFenceOpen returns true for the exact fence `” ```json ”`
// (possibly with trailing whitespace). It rejects `” ```jsonc ”`.
func isJSONFenceOpen(line string) bool {
	t := strings.TrimSpace(line)
	return t == "```json"
}

// loadGoldens reads every testdata/contract/*.json and returns
// map[basename] -> bytes.
func loadGoldens(t *testing.T) map[string][]byte {
	t.Helper()
	dir := filepath.Join("..", "..", "testdata", "contract")
	out := map[string][]byte{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".json")
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		out[name] = data
	}
	return out
}
