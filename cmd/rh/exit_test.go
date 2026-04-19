package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func buildRH(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "rh")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("go build: %v", err)
	}
	return bin
}

func assertSingleJSONDoc(t *testing.T, stdout []byte) map[string]any {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(stdout))
	var first map[string]any
	if err := dec.Decode(&first); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\n%s", err, string(stdout))
	}
	var trailing any
	if err := dec.Decode(&trailing); err == nil {
		t.Fatalf("stdout contains >1 JSON document (contract violation):\n%s", string(stdout))
	}
	return first
}

// TestE2E_UnknownFlagIsValidation: `rh --bogus-flag --json` must produce ONE
// JSON envelope with error.code="validation" and exit code 5. Regression guard
// for the cobra-error-taxonomy fix.
func TestE2E_UnknownFlagIsValidation(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	bin := buildRH(t)
	cmd := exec.Command(bin, "--bogus-flag", "--json")
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected ExitError, got %v", err)
	}
	if got := exitErr.ExitCode(); got != 5 {
		t.Errorf("exit = %d, want 5", got)
	}
	if strings.TrimSpace(stderr.String()) != "" {
		t.Errorf("stderr should be empty in JSON mode, got %q", stderr.String())
	}
	env := assertSingleJSONDoc(t, stdout.Bytes())
	errObj, _ := env["error"].(map[string]any)
	if errObj == nil || errObj["code"] != "validation" {
		t.Errorf("expected validation error, got %v", env)
	}
}

// TestE2E_UnknownSubcommandIsValidation: `rh bogussubcmd --json` must also
// map to code="validation" exit=5.
func TestE2E_UnknownSubcommandIsValidation(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	bin := buildRH(t)
	cmd := exec.Command(bin, "bogussubcmd", "--json")
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected ExitError, got %v", err)
	}
	if got := exitErr.ExitCode(); got != 5 {
		t.Errorf("exit = %d, want 5", got)
	}
	if strings.TrimSpace(stderr.String()) != "" {
		t.Errorf("stderr should be empty, got %q", stderr.String())
	}
	env := assertSingleJSONDoc(t, stdout.Bytes())
	errObj, _ := env["error"].(map[string]any)
	if errObj == nil || errObj["code"] != "validation" {
		t.Errorf("expected validation error, got %v", env)
	}
}
