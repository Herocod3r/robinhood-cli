package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
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

func TestExitCodeForAPIError(t *testing.T) {
	cases := []struct {
		code robinhood.Code
		want int
	}{
		{robinhood.CodeUnauthenticated, 2},
		{robinhood.CodeSessionExpired, 2},
		{robinhood.CodeSheriffRequired, 3},
		{robinhood.CodeRateLimited, 4},
		{robinhood.CodeValidation, 5},
		{robinhood.CodeRobinhoodUnavailable, 1},
		{robinhood.CodeNotFound, 1},
	}
	for _, c := range cases {
		t.Run(string(c.code), func(t *testing.T) {
			err := &robinhood.APIError{Code: c.code}
			if err.ExitCode() != c.want {
				t.Errorf("%s: got %d, want %d", c.code, err.ExitCode(), c.want)
			}
		})
	}
}

func TestRootCommand_Registers(t *testing.T) {
	root := newRootCommand()
	names := map[string]bool{}
	for _, c := range root.Commands() {
		names[c.Name()] = true
	}
	for _, want := range []string{"version", "commands", "schema", "portfolio"} {
		if !names[want] {
			t.Errorf("expected %q subcommand registered", want)
		}
	}
}

// TestE2E_UnauthenticatedPortfolioIsOneEnvelope: running `rh portfolio --json`
// with no credentials MUST produce exactly one JSON error envelope on stdout,
// nothing on stderr, and exit code 2. Regression guard for Codex finding #1.
func TestE2E_UnauthenticatedPortfolioIsOneEnvelope(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e build in -short mode")
	}
	bin := buildRH(t)

	cmd := exec.Command(bin, "portfolio", "--json")
	// Scrub any ambient tokens and point the keychain at an empty temp dir so
	// we deterministically hit the unauth path even on dev machines where the
	// user has a real keychain entry. (Task 14 added keychain fallback; the
	// subprocess inherits os.Environ by default.)
	cmd.Env = append(os.Environ(),
		"ROBINHOOD_ACCESS_TOKEN=",
		"ROBINHOOD_REFRESH_TOKEN=",
		"ROBINHOOD_DEVICE_TOKEN=",
		"ROBINHOOD_KEYCHAIN_BACKEND=file",
		"XDG_CONFIG_HOME="+t.TempDir(),
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected ExitError, got %v (stdout=%q stderr=%q)", err, stdout.String(), stderr.String())
	}
	if got := exitErr.ExitCode(); got != 2 {
		t.Errorf("exit code = %d, want 2", got)
	}
	if strings.TrimSpace(stderr.String()) != "" {
		t.Errorf("stderr should be empty in JSON mode, got: %q", stderr.String())
	}

	env := assertSingleJSONDoc(t, stdout.Bytes())
	if env["command"] != "portfolio" {
		t.Errorf(`command = %v, want "portfolio"`, env["command"])
	}
	errObj, _ := env["error"].(map[string]any)
	if errObj == nil {
		t.Fatalf("error envelope missing .error")
	}
	if errObj["code"] != "unauthenticated" {
		t.Errorf(`error.code = %v, want "unauthenticated"`, errObj["code"])
	}
}

// TestE2E_UnknownCommandIsOneEnvelope: `rh commands nope --json` must also
// produce exactly one JSON envelope (not plaintext). Regression guard for
// the "runCommands returned error without an envelope" slice of Codex #1.
func TestE2E_UnknownCommandIsOneEnvelope(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e build in -short mode")
	}
	bin := buildRH(t)

	cmd := exec.Command(bin, "commands", "definitely-not-a-command", "--json")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected ExitError, got %v", err)
	}
	if got := exitErr.ExitCode(); got != 5 {
		t.Errorf("exit code = %d, want 5 (validation)", got)
	}
	if strings.TrimSpace(stderr.String()) != "" {
		t.Errorf("stderr should be empty in JSON mode, got: %q", stderr.String())
	}

	env := assertSingleJSONDoc(t, stdout.Bytes())
	errObj, _ := env["error"].(map[string]any)
	if errObj == nil || errObj["code"] != "validation" {
		t.Errorf("expected validation error envelope, got %v", env)
	}
}

// TestE2E_SuccessfulVersionIsOneEnvelope: sanity check on the success path.
func TestE2E_SuccessfulVersionIsOneEnvelope(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e build in -short mode")
	}
	bin := buildRH(t)
	out, err := exec.Command(bin, "version", "--json").Output()
	if err != nil {
		t.Fatal(err)
	}
	env := assertSingleJSONDoc(t, out)
	if env["error"] != nil {
		t.Errorf("success envelope should have null error, got %v", env["error"])
	}
	if env["schema"] != "robinhood-cli/v1" {
		t.Errorf("schema = %v", env["schema"])
	}
}
