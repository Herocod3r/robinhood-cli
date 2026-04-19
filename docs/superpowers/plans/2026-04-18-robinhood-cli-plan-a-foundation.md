# robinhood-cli — Plan A: Foundation + First Vertical Slice

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Stand up the Go module, build the HTTP client foundation with env-var auth + automatic token refresh, emit the stable JSON envelope, and ship one end-to-end command (`rh portfolio`) plus discovery commands (`rh version`, `rh commands`, `rh schema`). A user with valid Robinhood tokens should be able to run `rh portfolio --json` and get real data.

**Architecture:** Three layers — `cmd/rh/` (Cobra commands, one file each), `internal/output/` (TTY-aware JSON/table renderer), `internal/robinhood/` (HTTP client + typed endpoints). Auth in Plan A accepts only pre-obtained `ROBINHOOD_ACCESS_TOKEN` + `ROBINHOOD_REFRESH_TOKEN` env vars; the client auto-refreshes on 401. Interactive `rh login` and Sheriff flow ship in Plan B.

**Tech Stack:** Go 1.22+, Cobra (CLI), `stretchr/testify` (assertions), `httptest` (net/http stdlib, fixture servers), `dnaeon/go-vcr/v3` (later plans), `olekukonko/tablewriter` (tables), `mattn/go-isatty` (TTY detect), `golangci-lint` (lint), `gotestsum` (nicer CI output).

**Spec reference:** `docs/superpowers/specs/2026-04-18-robinhood-cli-design.md`

**Out of scope for Plan A (shipped in B/C):** Sheriff workflow, interactive `rh login`, OS keychain, remaining 16 data commands, skill, GoReleaser, Homebrew.

---

## File Structure

Created in this plan:

```
robinhood-cli/
├── go.mod
├── go.sum
├── Makefile
├── .gitignore
├── .golangci.yml
├── LICENSE                               # MIT
├── README.md                             # minimal; full version ships in Plan C
├── CONTRIBUTING.md                       # minimal
├── .github/
│   └── workflows/
│       └── ci.yaml                       # go test + vet + lint
├── cmd/
│   └── rh/
│       ├── main.go                       # root Cobra cmd, global flags
│       ├── root.go                       # root command setup
│       ├── version.go                    # rh version
│       ├── commands.go                   # rh commands (discovery)
│       ├── schema.go                     # rh schema (discovery)
│       └── portfolio.go                  # rh portfolio
├── internal/
│   ├── buildinfo/
│   │   └── buildinfo.go                  # ldflags-populated version vars
│   ├── config/
│   │   └── config.go                     # XDG paths, env-var names
│   ├── output/
│   │   ├── output.go                     # Writer interface + TTY detection
│   │   ├── envelope.go                   # Envelope struct + error codes
│   │   ├── json.go                       # JSON renderer
│   │   ├── table.go                      # table renderer + per-command tables
│   │   └── output_test.go
│   └── robinhood/
│       ├── client.go                     # HTTP Client + Do() + 401 refresh retry
│       ├── client_test.go
│       ├── errors.go                     # typed error codes
│       ├── errors_test.go
│       ├── session.go                    # Session struct + env loader
│       ├── session_test.go
│       ├── oauth.go                      # refresh-token-grant; password-grant stubbed (Plan B)
│       ├── oauth_test.go
│       ├── types.go                      # shared helpers (Money string)
│       └── endpoints/
│           ├── portfolio.go
│           └── portfolio_test.go
```

**Responsibility contract:**

- `cmd/rh/*.go` — no business logic; only flag parsing, one client call, one output-writer call
- `internal/robinhood/client.go` — owns HTTP, auth header injection, 401-refresh-retry, rate-limit mapping
- `internal/robinhood/oauth.go` — owns grant types, token refresh, never touches transport retry logic
- `internal/robinhood/session.go` — owns token storage + env-var loading; never makes HTTP calls
- `internal/output/` — owns all rendering; client returns structs, CLI hands them off here
- `internal/robinhood/endpoints/` — one file per domain, each exposes typed methods on `*Client`

**Not creating in Plan A** (files that will appear in B/C): `cmd/rh/login.go`, `cmd/rh/logout.go`, `cmd/rh/skill.go`, `internal/keychain/`, `internal/robinhood/sheriff.go`, `skills/`, `.goreleaser.yaml`.

---

## Task 1: Bootstrap the Go module

**Files:**
- Create: `go.mod`
- Create: `.gitignore`
- Create: `LICENSE`
- Create: `README.md`
- Create: `CONTRIBUTING.md`
- Create: `Makefile`
- Create: `.golangci.yml`

- [ ] **Step 1: Initialize Go module**

Run:
```bash
cd /Users/jethro/Documents/Projects/robinhood-cli/.claude/worktrees/bootsrap
go mod init github.com/herocod3r/robinhood-cli
```

Expected: `go.mod` created with `module github.com/herocod3r/robinhood-cli` and a `go 1.22` (or whatever's installed) line. If Go isn't 1.22+, update the directive manually: `go 1.22`.

- [ ] **Step 2: Write `.gitignore`**

Create `.gitignore`:

```gitignore
# Binaries
/rh
/bin/
/dist/

# Test
coverage.out
coverage.html
*.test

# Editors
.idea/
.vscode/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Local env
.env
.env.local
```

- [ ] **Step 3: Write `LICENSE` (MIT)**

Create `LICENSE`:

```
MIT License

Copyright (c) 2026 herocod3r

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

- [ ] **Step 4: Write minimal `README.md`**

Create `README.md`:

```markdown
# rh — Robinhood CLI

Open-source, read-only Robinhood CLI for portfolio research and trade analysis.

> **⚠️ Research tool only.** Read-only. Uses an unofficial API; may break without notice.

Full docs ship with v1 release. See `docs/superpowers/specs/2026-04-18-robinhood-cli-design.md`
for the current design.

## License

MIT
```

- [ ] **Step 5: Write minimal `CONTRIBUTING.md`**

Create `CONTRIBUTING.md`:

```markdown
# Contributing

1. Fork, branch, PR to `main`.
2. `make test` and `make lint` must pass.
3. Follow the design spec at `docs/superpowers/specs/2026-04-18-robinhood-cli-design.md`.
4. Commits: conventional commits style (`feat:`, `fix:`, `chore:`, `docs:`, `test:`).
```

- [ ] **Step 6: Write `Makefile`**

Create `Makefile`:

```makefile
.PHONY: build test test-race lint fmt vet tidy clean run help

GO ?= go
BINARY := rh
PKG := github.com/herocod3r/robinhood-cli
VERSION ?= dev
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
LDFLAGS := -s -w -X $(PKG)/internal/buildinfo.Version=$(VERSION) -X $(PKG)/internal/buildinfo.Commit=$(COMMIT)

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-12s %s\n", $$1, $$2}'

build: ## Build the rh binary
	$(GO) build -ldflags="$(LDFLAGS)" -o $(BINARY) ./cmd/rh

test: ## Run unit tests
	$(GO) test ./...

test-race: ## Run tests with -race
	$(GO) test -race ./...

lint: ## Run golangci-lint
	golangci-lint run ./...

fmt: ## Format
	$(GO) fmt ./...

vet: ## go vet
	$(GO) vet ./...

tidy: ## go mod tidy
	$(GO) mod tidy

clean: ## Remove build artifacts
	rm -f $(BINARY) coverage.out coverage.html
	rm -rf dist/

run: build ## Build and run
	./$(BINARY) $(ARGS)
```

- [ ] **Step 7: Write `.golangci.yml`**

Create `.golangci.yml`:

```yaml
run:
  timeout: 3m
  tests: true

linters:
  disable-all: true
  enable:
    - errcheck
    - govet
    - ineffassign
    - staticcheck
    - unused
    - gofmt
    - goimports
    - gosec
    - revive
    - misspell

linters-settings:
  gosec:
    excludes:
      - G104   # unhandled errors checked by errcheck
  revive:
    severity: warning
    rules:
      - name: var-naming
      - name: exported
      - name: unused-parameter
      - name: blank-imports
      - name: context-as-argument

issues:
  max-same-issues: 0
  exclude-rules:
    - path: _test\.go
      linters: [gosec, errcheck]
```

- [ ] **Step 8: Verify module compiles (no code yet, but `go build` shouldn't error on empty package)**

Run:
```bash
go build ./...
```

Expected: no output (no packages yet). If `go build` complains about "no Go files," that's fine — we'll add packages in later tasks.

- [ ] **Step 9: Commit**

```bash
git add go.mod .gitignore LICENSE README.md CONTRIBUTING.md Makefile .golangci.yml
git commit -m "chore: bootstrap Go module, Makefile, license"
```

---

## Task 2: Build-info package (version/commit vars for ldflags)

**Files:**
- Create: `internal/buildinfo/buildinfo.go`
- Create: `internal/buildinfo/buildinfo_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/buildinfo/buildinfo_test.go`:

```go
package buildinfo

import "testing"

func TestDefaults(t *testing.T) {
	if Version == "" {
		t.Error("Version should have a default")
	}
	if Commit == "" {
		t.Error("Commit should have a default")
	}
	if SchemaVersion != "robinhood-cli/v1" {
		t.Errorf("SchemaVersion = %q, want %q", SchemaVersion, "robinhood-cli/v1")
	}
}
```

- [ ] **Step 2: Run the test (should fail — package doesn't exist)**

Run:
```bash
go test ./internal/buildinfo/...
```

Expected: build failure, `no Go files` or package-not-found error.

- [ ] **Step 3: Implement the package**

Create `internal/buildinfo/buildinfo.go`:

```go
// Package buildinfo exposes version metadata set at build time via -ldflags.
package buildinfo

// Version is the semver version of the CLI, set at build time.
var Version = "dev"

// Commit is the git short SHA, set at build time.
var Commit = "none"

// SchemaVersion is the stable JSON envelope version. Never changes within v1.
const SchemaVersion = "robinhood-cli/v1"
```

- [ ] **Step 4: Run the test (should pass)**

Run:
```bash
go test ./internal/buildinfo/...
```

Expected: `ok github.com/herocod3r/robinhood-cli/internal/buildinfo`.

- [ ] **Step 5: Commit**

```bash
git add internal/buildinfo/
git commit -m "feat(buildinfo): ldflags-injected version vars + schema version constant"
```

---

## Task 3: Error taxonomy

**Files:**
- Create: `internal/robinhood/errors.go`
- Create: `internal/robinhood/errors_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/robinhood/errors_test.go`:

```go
package robinhood

import (
	"errors"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	e := &APIError{Code: CodeNotFound, Message: "no such symbol"}
	if e.Error() != "not_found: no such symbol" {
		t.Errorf("got %q", e.Error())
	}
}

func TestAPIError_ExitCode(t *testing.T) {
	cases := []struct {
		code Code
		want int
	}{
		{CodeUnauthenticated, 2},
		{CodeSessionExpired, 2},
		{CodeSheriffRequired, 3},
		{CodeRateLimited, 4},
		{CodeValidation, 5},
		{CodeRobinhoodUnavailable, 1},
		{CodeNotFound, 1},
	}
	for _, c := range cases {
		e := &APIError{Code: c.code}
		if e.ExitCode() != c.want {
			t.Errorf("%s: ExitCode() = %d, want %d", c.code, e.ExitCode(), c.want)
		}
	}
}

func TestAPIError_Is(t *testing.T) {
	e := &APIError{Code: CodeNotFound}
	if !errors.Is(e, &APIError{Code: CodeNotFound}) {
		t.Error("errors.Is should match on Code")
	}
	if errors.Is(e, &APIError{Code: CodeRateLimited}) {
		t.Error("errors.Is should not match different Code")
	}
}
```

- [ ] **Step 2: Run the test (should fail)**

Run:
```bash
go test ./internal/robinhood/...
```

Expected: build failure, `APIError undefined`.

- [ ] **Step 3: Implement the package**

Create `internal/robinhood/errors.go`:

```go
// Package robinhood is the internal Robinhood HTTP client and typed endpoints.
package robinhood

import "fmt"

// Code is the stable error taxonomy exposed in the JSON envelope and via exit codes.
type Code string

const (
	CodeUnauthenticated      Code = "unauthenticated"
	CodeSessionExpired       Code = "session_expired"
	CodeSheriffRequired      Code = "sheriff_required"
	CodeRateLimited          Code = "rate_limited"
	CodeRobinhoodUnavailable Code = "robinhood_unavailable"
	CodeNotFound             Code = "not_found"
	CodeValidation           Code = "validation"
)

// APIError is the structured error returned from all Robinhood client operations.
type APIError struct {
	Code      Code
	Message   string
	Hint      string
	Retryable bool
	// HTTPStatus is the underlying HTTP status code, 0 if not applicable.
	HTTPStatus int
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Message == "" {
		return string(e.Code)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Is supports errors.Is matching on Code only.
func (e *APIError) Is(target error) bool {
	t, ok := target.(*APIError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// ExitCode maps the error code to the CLI exit code per spec section 6.5.
func (e *APIError) ExitCode() int {
	switch e.Code {
	case CodeUnauthenticated, CodeSessionExpired:
		return 2
	case CodeSheriffRequired:
		return 3
	case CodeRateLimited:
		return 4
	case CodeValidation:
		return 5
	default:
		return 1
	}
}
```

- [ ] **Step 4: Run the test (should pass)**

Run:
```bash
go test ./internal/robinhood/...
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/robinhood/errors.go internal/robinhood/errors_test.go
git commit -m "feat(robinhood): error taxonomy + exit code mapping"
```

---

## Task 4: Config package (env-var names + XDG paths)

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/config/config_test.go`:

```go
package config

import (
	"path/filepath"
	"testing"
)

func TestConfigDir_UsesHome(t *testing.T) {
	t.Setenv("HOME", "/tmp/testhome")
	t.Setenv("XDG_CONFIG_HOME", "")
	dir, err := ConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("/tmp/testhome", ".config", "robinhood-cli")
	if dir != want {
		t.Errorf("got %q, want %q", dir, want)
	}
}

func TestConfigDir_UsesXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg")
	dir, err := ConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("/tmp/xdg", "robinhood-cli")
	if dir != want {
		t.Errorf("got %q, want %q", dir, want)
	}
}

func TestEnvNamesAreStable(t *testing.T) {
	// These names are a public contract; changing them breaks users.
	if EnvAccessToken != "ROBINHOOD_ACCESS_TOKEN" {
		t.Error("EnvAccessToken changed")
	}
	if EnvRefreshToken != "ROBINHOOD_REFRESH_TOKEN" {
		t.Error("EnvRefreshToken changed")
	}
	if EnvDeviceToken != "ROBINHOOD_DEVICE_TOKEN" {
		t.Error("EnvDeviceToken changed")
	}
}
```

- [ ] **Step 2: Run the test (should fail)**

Run:
```bash
go test ./internal/config/...
```

Expected: build failure.

- [ ] **Step 3: Implement the package**

Create `internal/config/config.go`:

```go
// Package config holds env-var names and config-path resolution.
package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Env-var names — public contract with users and the skill.
const (
	EnvUsername     = "ROBINHOOD_USERNAME"
	EnvPassword     = "ROBINHOOD_PASSWORD"
	EnvTOTPSecret   = "ROBINHOOD_TOTP_SECRET"
	EnvDeviceToken  = "ROBINHOOD_DEVICE_TOKEN"
	EnvAccessToken  = "ROBINHOOD_ACCESS_TOKEN"
	EnvRefreshToken = "ROBINHOOD_REFRESH_TOKEN"
	EnvProfile      = "ROBINHOOD_PROFILE"
)

// AppName is the application directory name under the config root.
const AppName = "robinhood-cli"

// ConfigDir returns $XDG_CONFIG_HOME/robinhood-cli, falling back to $HOME/.config/robinhood-cli.
func ConfigDir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, AppName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".config", AppName), nil
}

// EnsureConfigDir creates the config dir with 0700 perms if missing.
func EnsureConfigDir() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("create %s: %w", dir, err)
	}
	return dir, nil
}
```

- [ ] **Step 4: Run the test (should pass)**

Run:
```bash
go test ./internal/config/...
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat(config): env-var constants + XDG config dir resolver"
```

---

## Task 5: Session struct with env-var loading

**Files:**
- Create: `internal/robinhood/session.go`
- Create: `internal/robinhood/session_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/robinhood/session_test.go`:

```go
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

func TestSession_IsExpired(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name   string
		expiry time.Time
		want   bool
	}{
		{"zero time (unknown)", time.Time{}, true},
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
```

- [ ] **Step 2: Run the test (should fail)**

Run:
```bash
go test ./internal/robinhood/...
```

Expected: build failure, `LoadSessionFromEnv undefined`.

- [ ] **Step 3: Implement the session file**

Create `internal/robinhood/session.go`:

```go
package robinhood

import (
	"os"
	"time"

	"github.com/herocod3r/robinhood-cli/internal/config"
)

// refreshBuffer is how long before ExpiresAt we pre-emptively refresh.
const refreshBuffer = 5 * time.Minute

// Session holds tokens and device identity for a single Robinhood account.
type Session struct {
	Username     string
	AccessToken  string
	RefreshToken string
	DeviceToken  string
	ExpiresAt    time.Time
}

// IsExpired reports whether the access token is at or within the refresh buffer.
// A zero ExpiresAt is treated as unknown-expired (safer default).
func (s *Session) IsExpired() bool {
	if s.ExpiresAt.IsZero() {
		return true
	}
	return time.Now().Add(refreshBuffer).After(s.ExpiresAt)
}

// LoadSessionFromEnv builds a Session from env vars.
// Returns CodeUnauthenticated if access+refresh tokens are both missing.
func LoadSessionFromEnv() (*Session, error) {
	access := os.Getenv(config.EnvAccessToken)
	refresh := os.Getenv(config.EnvRefreshToken)
	if access == "" && refresh == "" {
		return nil, &APIError{
			Code:    CodeUnauthenticated,
			Message: "no access/refresh tokens found",
			Hint:    "set ROBINHOOD_ACCESS_TOKEN and ROBINHOOD_REFRESH_TOKEN, or run: rh login (Plan B)",
		}
	}
	return &Session{
		Username:     os.Getenv(config.EnvUsername),
		AccessToken:  access,
		RefreshToken: refresh,
		DeviceToken:  os.Getenv(config.EnvDeviceToken),
		// ExpiresAt unknown from env; IsExpired() returns true, forcing refresh on first call.
	}, nil
}
```

- [ ] **Step 4: Run the test (should pass)**

Run:
```bash
go test ./internal/robinhood/...
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/robinhood/session.go internal/robinhood/session_test.go
git commit -m "feat(robinhood): Session struct + env-var loader + IsExpired"
```

---

## Task 6: OAuth refresh-token grant

**Files:**
- Create: `internal/robinhood/oauth.go`
- Create: `internal/robinhood/oauth_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/robinhood/oauth_test.go`:

```go
package robinhood

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestRefreshAccessToken_Success(t *testing.T) {
	var gotBody url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth2/token/" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		gotBody, _ = url.ParseQuery(string(body))

		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "new-access",
			"refresh_token": "new-refresh",
			"expires_in":    86400,
			"token_type":    "Bearer",
		})
	}))
	defer srv.Close()

	o := &oauth{baseURL: srv.URL, httpClient: srv.Client()}
	sess := &Session{RefreshToken: "old-refresh", DeviceToken: "dev-123"}

	if err := o.Refresh(sess); err != nil {
		t.Fatal(err)
	}
	if sess.AccessToken != "new-access" {
		t.Errorf("AccessToken = %q", sess.AccessToken)
	}
	if sess.RefreshToken != "new-refresh" {
		t.Errorf("RefreshToken = %q", sess.RefreshToken)
	}
	if sess.ExpiresAt.IsZero() {
		t.Error("ExpiresAt should be set")
	}
	if gotBody.Get("grant_type") != "refresh_token" {
		t.Errorf("grant_type = %q", gotBody.Get("grant_type"))
	}
	if gotBody.Get("refresh_token") != "old-refresh" {
		t.Errorf("refresh_token = %q", gotBody.Get("refresh_token"))
	}
	if gotBody.Get("device_token") != "dev-123" {
		t.Errorf("device_token = %q", gotBody.Get("device_token"))
	}
}

func TestRefreshAccessToken_InvalidGrant(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"error":"invalid_grant","error_description":"refresh token is invalid"}`)
	}))
	defer srv.Close()

	o := &oauth{baseURL: srv.URL, httpClient: srv.Client()}
	sess := &Session{RefreshToken: "bad"}

	err := o.Refresh(sess)
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("got %T", err)
	}
	if apiErr.Code != CodeSessionExpired {
		t.Errorf("Code = %s, want %s", apiErr.Code, CodeSessionExpired)
	}
}

func TestRefreshAccessToken_NoRefreshToken(t *testing.T) {
	o := &oauth{baseURL: "http://unused", httpClient: http.DefaultClient}
	sess := &Session{}
	err := o.Refresh(sess)
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.Code != CodeUnauthenticated {
		t.Errorf("got %v", err)
	}
}

func TestRefreshAccessToken_UnknownError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, "oops")
	}))
	defer srv.Close()

	o := &oauth{baseURL: srv.URL, httpClient: srv.Client()}
	sess := &Session{RefreshToken: "x"}

	err := o.Refresh(sess)
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("got %T", err)
	}
	if apiErr.Code != CodeRobinhoodUnavailable {
		t.Errorf("Code = %s", apiErr.Code)
	}
	if !strings.Contains(apiErr.Message, "500") {
		t.Errorf("Message = %q", apiErr.Message)
	}
}
```

- [ ] **Step 2: Run the test (should fail)**

Run:
```bash
go test ./internal/robinhood/...
```

Expected: build failure, `oauth undefined`.

- [ ] **Step 3: Implement OAuth refresh**

Create `internal/robinhood/oauth.go`:

```go
package robinhood

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// robinhoodOAuthClientID is the same client_id the mobile/web app uses, copied from
// robin_stocks. Unofficial; may be rotated by Robinhood.
const robinhoodOAuthClientID = "c82SH0WZOsabOXGP2sxqcj34FxkvfnWRZBKlBjFS"

// oauth owns grant types and token refresh.
type oauth struct {
	baseURL    string
	httpClient *http.Client
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type oauthErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// Refresh exchanges the session's RefreshToken for new tokens, mutating the session in place.
func (o *oauth) Refresh(s *Session) error {
	if s.RefreshToken == "" {
		return &APIError{
			Code:    CodeUnauthenticated,
			Message: "no refresh token available",
			Hint:    "run: rh login",
		}
	}
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", s.RefreshToken)
	form.Set("client_id", robinhoodOAuthClientID)
	form.Set("expires_in", "86400")
	form.Set("scope", "internal")
	if s.DeviceToken != "" {
		form.Set("device_token", s.DeviceToken)
	}

	req, err := http.NewRequest(http.MethodPost, o.baseURL+"/oauth2/token/", strings.NewReader(form.Encode()))
	if err != nil {
		return &APIError{Code: CodeRobinhoodUnavailable, Message: err.Error()}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return &APIError{Code: CodeRobinhoodUnavailable, Message: err.Error()}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	switch {
	case resp.StatusCode == http.StatusOK:
		var tr tokenResponse
		if err := json.Unmarshal(body, &tr); err != nil {
			return &APIError{Code: CodeRobinhoodUnavailable, Message: "malformed token response"}
		}
		s.AccessToken = tr.AccessToken
		if tr.RefreshToken != "" {
			s.RefreshToken = tr.RefreshToken
		}
		if tr.ExpiresIn > 0 {
			s.ExpiresAt = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
		}
		return nil

	case resp.StatusCode == http.StatusBadRequest, resp.StatusCode == http.StatusUnauthorized:
		var er oauthErrorResponse
		_ = json.Unmarshal(body, &er)
		return &APIError{
			Code:       CodeSessionExpired,
			Message:    firstNonEmpty(er.ErrorDescription, er.Error, "refresh rejected"),
			Hint:       "run: rh login",
			HTTPStatus: resp.StatusCode,
		}

	default:
		return &APIError{
			Code:       CodeRobinhoodUnavailable,
			Message:    fmt.Sprintf("oauth refresh HTTP %d", resp.StatusCode),
			HTTPStatus: resp.StatusCode,
		}
	}
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}
```

- [ ] **Step 4: Run the test (should pass)**

Run:
```bash
go test ./internal/robinhood/...
```

Expected: PASS (all four oauth subtests).

- [ ] **Step 5: Commit**

```bash
git add internal/robinhood/oauth.go internal/robinhood/oauth_test.go
git commit -m "feat(robinhood): oauth refresh-token grant with typed errors"
```

---

## Task 7: HTTP Client with 401-refresh-retry

**Files:**
- Create: `internal/robinhood/types.go`
- Create: `internal/robinhood/client.go`
- Create: `internal/robinhood/client_test.go`

- [ ] **Step 1: Write types file (no tests — just a type alias)**

Create `internal/robinhood/types.go`:

```go
package robinhood

// Money is a decimal-string representation of monetary values.
// We preserve Robinhood's string form; conversion to a numeric type is a
// caller concern (and should use a decimal library, not float64).
type Money string
```

- [ ] **Step 2: Write the failing Client test**

Create `internal/robinhood/client_test.go`:

```go
package robinhood

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

type pingResponse struct {
	OK bool `json:"ok"`
}

func TestClient_GetJSON_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer acc-tok" {
			t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
		}
		_ = json.NewEncoder(w).Encode(pingResponse{OK: true})
	}))
	defer srv.Close()

	c := NewClientWithHosts(srv.URL, srv.URL, srv.URL, srv.Client())
	c.SetSession(&Session{AccessToken: "acc-tok", ExpiresAt: time.Now().Add(time.Hour)})

	var out pingResponse
	if err := c.getJSON(APIHost, "/ping", &out); err != nil {
		t.Fatal(err)
	}
	if !out.OK {
		t.Error("expected OK true")
	}
}

func TestClient_GetJSON_RefreshesOn401(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ping":
			n := atomic.AddInt32(&calls, 1)
			if n == 1 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if r.Header.Get("Authorization") != "Bearer new-access" {
				t.Errorf("authorization after refresh = %q", r.Header.Get("Authorization"))
			}
			_ = json.NewEncoder(w).Encode(pingResponse{OK: true})
		case "/oauth2/token/":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "new-access",
				"refresh_token": "new-refresh",
				"expires_in":    86400,
			})
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := NewClientWithHosts(srv.URL, srv.URL, srv.URL, srv.Client())
	c.SetSession(&Session{AccessToken: "old", RefreshToken: "old-ref", ExpiresAt: time.Now().Add(time.Hour)})

	var out pingResponse
	if err := c.getJSON(APIHost, "/ping", &out); err != nil {
		t.Fatal(err)
	}
	if !out.OK {
		t.Error("expected OK after refresh+retry")
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("calls = %d, want 2", got)
	}
}

func TestClient_GetJSON_RefreshesThenPermanent401(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ping":
			atomic.AddInt32(&calls, 1)
			w.WriteHeader(http.StatusUnauthorized)
		case "/oauth2/token/":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "new-access",
				"refresh_token": "new-refresh",
				"expires_in":    86400,
			})
		}
	}))
	defer srv.Close()

	c := NewClientWithHosts(srv.URL, srv.URL, srv.URL, srv.Client())
	c.SetSession(&Session{AccessToken: "old", RefreshToken: "r", ExpiresAt: time.Now().Add(time.Hour)})

	var out pingResponse
	err := c.getJSON(APIHost, "/ping", &out)
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.Code != CodeUnauthenticated {
		t.Errorf("got %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("calls = %d, want exactly 2 (original + one retry)", got)
	}
}

func TestClient_GetJSON_PreEmptiveRefresh(t *testing.T) {
	var oauthCalls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ping":
			if r.Header.Get("Authorization") != "Bearer new-access" {
				t.Errorf("expected refreshed token, got %q", r.Header.Get("Authorization"))
			}
			_ = json.NewEncoder(w).Encode(pingResponse{OK: true})
		case "/oauth2/token/":
			atomic.AddInt32(&oauthCalls, 1)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "new-access",
				"refresh_token": "new-refresh",
				"expires_in":    86400,
			})
		}
	}))
	defer srv.Close()

	c := NewClientWithHosts(srv.URL, srv.URL, srv.URL, srv.Client())
	// Expired token forces pre-emptive refresh.
	c.SetSession(&Session{AccessToken: "old", RefreshToken: "r", ExpiresAt: time.Now().Add(-1 * time.Minute)})

	var out pingResponse
	if err := c.getJSON(APIHost, "/ping", &out); err != nil {
		t.Fatal(err)
	}
	if got := atomic.LoadInt32(&oauthCalls); got != 1 {
		t.Errorf("oauth calls = %d, want 1", got)
	}
}

func TestClient_GetJSON_RateLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	c := NewClientWithHosts(srv.URL, srv.URL, srv.URL, srv.Client())
	c.SetSession(&Session{AccessToken: "a", ExpiresAt: time.Now().Add(time.Hour)})

	var out pingResponse
	err := c.getJSON(APIHost, "/ping", &out)
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.Code != CodeRateLimited {
		t.Errorf("got %v", err)
	}
	if apiErr.Hint == "" {
		t.Error("rate limit Hint should include retry-after")
	}
}

func TestClient_GetJSON_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewClientWithHosts(srv.URL, srv.URL, srv.URL, srv.Client())
	c.SetSession(&Session{AccessToken: "a", ExpiresAt: time.Now().Add(time.Hour)})

	var out pingResponse
	err := c.getJSON(APIHost, "/missing", &out)
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.Code != CodeNotFound {
		t.Errorf("got %v", err)
	}
}
```

- [ ] **Step 3: Run the test (should fail — build error)**

Run:
```bash
go test ./internal/robinhood/...
```

Expected: `NewClientWithHosts undefined`, etc.

- [ ] **Step 4: Implement the Client**

Create `internal/robinhood/client.go`:

```go
package robinhood

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Host enumerates the Robinhood API hosts we talk to.
type Host int

const (
	APIHost     Host = iota // api.robinhood.com
	NummusHost              // nummus.robinhood.com  (crypto)
	PhoenixHost             // phoenix.robinhood.com (unified account)
)

// Default production hosts.
const (
	defaultAPIBase     = "https://api.robinhood.com"
	defaultNummusBase  = "https://nummus.robinhood.com"
	defaultPhoenixBase = "https://phoenix.robinhood.com"
)

// Client is the HTTP client for Robinhood's unofficial API.
type Client struct {
	apiBase     string
	nummusBase  string
	phoenixBase string
	http        *http.Client
	oauth       *oauth

	mu      sync.Mutex
	session *Session
}

// NewClient returns a client pointed at production hosts.
func NewClient() *Client {
	return NewClientWithHosts(defaultAPIBase, defaultNummusBase, defaultPhoenixBase, &http.Client{Timeout: 30 * time.Second})
}

// NewClientWithHosts lets tests override hosts and the underlying http.Client.
func NewClientWithHosts(apiBase, nummusBase, phoenixBase string, h *http.Client) *Client {
	return &Client{
		apiBase:     apiBase,
		nummusBase:  nummusBase,
		phoenixBase: phoenixBase,
		http:        h,
		oauth:       &oauth{baseURL: apiBase, httpClient: h},
	}
}

// SetSession installs the active session.
func (c *Client) SetSession(s *Session) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.session = s
}

// Session returns the active session (may be nil).
func (c *Client) Session() *Session {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.session
}

// baseFor returns the base URL for the given host.
func (c *Client) baseFor(h Host) string {
	switch h {
	case NummusHost:
		return c.nummusBase
	case PhoenixHost:
		return c.phoenixBase
	default:
		return c.apiBase
	}
}

// ensureFresh refreshes pre-emptively if the session is expired.
func (c *Client) ensureFresh() error {
	c.mu.Lock()
	s := c.session
	c.mu.Unlock()
	if s == nil {
		return &APIError{Code: CodeUnauthenticated, Message: "no session", Hint: "run: rh login"}
	}
	if s.IsExpired() && s.RefreshToken != "" {
		return c.oauth.Refresh(s)
	}
	return nil
}

// getJSON does a GET, auto-refreshes once on 401, and decodes into out.
// It is the private workhorse used by all endpoint helpers.
func (c *Client) getJSON(host Host, path string, out any) error {
	if err := c.ensureFresh(); err != nil {
		return err
	}
	resp, err := c.do(host, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		// One retry after refresh.
		if rerr := c.oauth.Refresh(c.Session()); rerr != nil {
			return rerr
		}
		resp2, err := c.do(host, http.MethodGet, path, nil)
		if err != nil {
			return err
		}
		defer resp2.Body.Close()
		if resp2.StatusCode == http.StatusUnauthorized {
			return &APIError{Code: CodeUnauthenticated, Message: "401 after refresh", Hint: "run: rh login"}
		}
		return decodeOrMap(resp2, out)
	}

	return decodeOrMap(resp, out)
}

// do builds and sends a single request with the Authorization header.
func (c *Client) do(host Host, method, path string, body io.Reader) (*http.Response, error) {
	s := c.Session()
	req, err := http.NewRequest(method, c.baseFor(host)+path, body)
	if err != nil {
		return nil, &APIError{Code: CodeValidation, Message: err.Error()}
	}
	if s != nil && s.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.AccessToken)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "robinhood-cli (+https://github.com/herocod3r/robinhood-cli)")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, &APIError{Code: CodeRobinhoodUnavailable, Message: err.Error()}
	}
	return resp, nil
}

// decodeOrMap converts non-2xx responses into typed APIErrors, and decodes 2xx into out.
func decodeOrMap(resp *http.Response, out any) error {
	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		if out == nil {
			return nil
		}
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return &APIError{Code: CodeRobinhoodUnavailable, Message: fmt.Sprintf("decode: %v", err)}
		}
		return nil
	case resp.StatusCode == http.StatusNotFound:
		return &APIError{Code: CodeNotFound, Message: "not found", HTTPStatus: resp.StatusCode}
	case resp.StatusCode == http.StatusTooManyRequests:
		retry := resp.Header.Get("Retry-After")
		msg := "rate limited"
		hint := "retry in 30s"
		if retry != "" {
			hint = "retry in " + retry + "s"
		}
		return &APIError{Code: CodeRateLimited, Message: msg, Hint: hint, Retryable: true, HTTPStatus: resp.StatusCode}
	case resp.StatusCode >= 500:
		return &APIError{Code: CodeRobinhoodUnavailable, Message: fmt.Sprintf("HTTP %d", resp.StatusCode), HTTPStatus: resp.StatusCode}
	default:
		buf, _ := io.ReadAll(resp.Body)
		return &APIError{Code: CodeValidation, Message: strings.TrimSpace(string(buf)), HTTPStatus: resp.StatusCode}
	}
}
```

- [ ] **Step 5: Run the tests (should pass)**

Run:
```bash
go test ./internal/robinhood/... -race
```

Expected: PASS, including the race detector.

- [ ] **Step 6: Commit**

```bash
git add internal/robinhood/client.go internal/robinhood/client_test.go internal/robinhood/types.go
git commit -m "feat(robinhood): HTTP client with 401-refresh-retry and typed errors"
```

---

## Task 8: Portfolio endpoint

**Files:**
- Create: `internal/robinhood/endpoints/portfolio.go`
- Create: `internal/robinhood/endpoints/portfolio_test.go`

- [ ] **Step 1: Write the failing endpoint test**

Create `internal/robinhood/endpoints/portfolio_test.go`:

```go
package endpoints

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

const portfolioFixture = `{
  "results": [{
    "account": "https://api.robinhood.com/accounts/5ABC/",
    "equity": "98234.5000",
    "extended_hours_equity": "98240.1200",
    "market_value": "72450.1000",
    "last_core_equity": "97892.3500"
  }]
}`

const unifiedFixture = `{
  "total_equity": "98234.5000",
  "cash": "25784.4000",
  "buying_power": "25784.4000"
}`

func TestPortfolio_Get(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/portfolios/":
			_, _ = w.Write([]byte(portfolioFixture))
		case "/accounts/unified":
			_, _ = w.Write([]byte(unifiedFixture))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c := robinhood.NewClientWithHosts(srv.URL, srv.URL, srv.URL, srv.Client())
	c.SetSession(&robinhood.Session{AccessToken: "tok", ExpiresAt: time.Now().Add(time.Hour)})

	p := NewPortfolio(c)
	got, err := p.Get()
	if err != nil {
		t.Fatal(err)
	}
	if got.Equity != "98234.5000" {
		t.Errorf("Equity = %q", got.Equity)
	}
	if got.ExtendedHoursEquity != "98240.1200" {
		t.Errorf("ExtendedHoursEquity = %q", got.ExtendedHoursEquity)
	}
	if got.Cash != "25784.4000" {
		t.Errorf("Cash = %q", got.Cash)
	}
	if got.BuyingPower != "25784.4000" {
		t.Errorf("BuyingPower = %q", got.BuyingPower)
	}
}

func TestPortfolio_Get_EmptyResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/portfolios/":
			_ = json.NewEncoder(w).Encode(map[string]any{"results": []any{}})
		case "/accounts/unified":
			_, _ = w.Write([]byte(unifiedFixture))
		}
	}))
	defer srv.Close()

	c := robinhood.NewClientWithHosts(srv.URL, srv.URL, srv.URL, srv.Client())
	c.SetSession(&robinhood.Session{AccessToken: "tok", ExpiresAt: time.Now().Add(time.Hour)})

	p := NewPortfolio(c)
	_, err := p.Get()
	apiErr, ok := err.(*robinhood.APIError)
	if !ok || apiErr.Code != robinhood.CodeNotFound {
		t.Errorf("got %v", err)
	}
}
```

- [ ] **Step 2: Run the test (should fail — build error)**

Run:
```bash
go test ./internal/robinhood/endpoints/...
```

Expected: `NewPortfolio undefined`.

- [ ] **Step 3: Expose `GetJSON` on the Client for endpoint packages**

The Portfolio endpoint lives in a sibling package and needs access to the request helper. Edit `internal/robinhood/client.go` — at the end of the file (after `decodeOrMap`), append the exported method:

```go
// GetJSON is the exported host-scoped GET for endpoint packages.
// Endpoint subpackages use this rather than the internal getJSON directly.
func (c *Client) GetJSON(host Host, path string, out any) error {
	return c.getJSON(host, path, out)
}
```

- [ ] **Step 4: Implement the portfolio endpoint**

Create `internal/robinhood/endpoints/portfolio.go`:

```go
// Package endpoints holds typed Robinhood REST endpoint wrappers.
package endpoints

import (
	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

// Portfolio wraps portfolio-related endpoints.
type Portfolio struct {
	c *robinhood.Client
}

// NewPortfolio constructs a Portfolio endpoint bound to the given client.
func NewPortfolio(c *robinhood.Client) *Portfolio {
	return &Portfolio{c: c}
}

// PortfolioSummary is the merged view presented by `rh portfolio`.
type PortfolioSummary struct {
	Equity              robinhood.Money `json:"equity"`
	ExtendedHoursEquity robinhood.Money `json:"extended_hours_equity"`
	MarketValue         robinhood.Money `json:"market_value"`
	Cash                robinhood.Money `json:"cash"`
	BuyingPower         robinhood.Money `json:"buying_power"`
	LastCoreEquity      robinhood.Money `json:"last_core_equity,omitempty"`
}

type portfolioResultsResp struct {
	Results []struct {
		Account             string          `json:"account"`
		Equity              robinhood.Money `json:"equity"`
		ExtendedHoursEquity robinhood.Money `json:"extended_hours_equity"`
		MarketValue         robinhood.Money `json:"market_value"`
		LastCoreEquity      robinhood.Money `json:"last_core_equity"`
	} `json:"results"`
}

type unifiedAccountResp struct {
	TotalEquity robinhood.Money `json:"total_equity"`
	Cash        robinhood.Money `json:"cash"`
	BuyingPower robinhood.Money `json:"buying_power"`
}

// Get returns a merged PortfolioSummary combining /portfolios/ and /accounts/unified.
func (p *Portfolio) Get() (*PortfolioSummary, error) {
	var pr portfolioResultsResp
	if err := p.c.GetJSON(robinhood.APIHost, "/portfolios/", &pr); err != nil {
		return nil, err
	}
	if len(pr.Results) == 0 {
		return nil, &robinhood.APIError{
			Code:    robinhood.CodeNotFound,
			Message: "no portfolios returned for this account",
		}
	}
	first := pr.Results[0]

	var uni unifiedAccountResp
	if err := p.c.GetJSON(robinhood.PhoenixHost, "/accounts/unified", &uni); err != nil {
		return nil, err
	}

	return &PortfolioSummary{
		Equity:              first.Equity,
		ExtendedHoursEquity: first.ExtendedHoursEquity,
		MarketValue:         first.MarketValue,
		LastCoreEquity:      first.LastCoreEquity,
		Cash:                uni.Cash,
		BuyingPower:         uni.BuyingPower,
	}, nil
}
```

- [ ] **Step 5: Run the tests (should pass)**

Run:
```bash
go test ./internal/robinhood/...
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/robinhood/client.go internal/robinhood/endpoints/
git commit -m "feat(endpoints): Portfolio endpoint merging /portfolios/ + unified"
```

---

## Task 9: Output envelope + JSON renderer

**Files:**
- Create: `internal/output/envelope.go`
- Create: `internal/output/json.go`
- Create: `internal/output/envelope_test.go`

- [ ] **Step 1: Write the failing envelope test**

Create `internal/output/envelope_test.go`:

```go
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
	if env.Data != nil {
		t.Error("Data should be nil on error")
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
```

- [ ] **Step 2: Run the test (should fail)**

Run:
```bash
go test ./internal/output/...
```

Expected: build failure.

- [ ] **Step 3: Implement envelope + JSON writer**

Create `internal/output/envelope.go`:

```go
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
```

Create `internal/output/json.go`:

```go
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
```

- [ ] **Step 4: Run the test (should pass)**

Run:
```bash
go mod tidy
go test ./internal/output/...
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/output/envelope.go internal/output/envelope_test.go internal/output/json.go
git commit -m "feat(output): JSON envelope + WriteSuccess/WriteError"
```

---

## Task 10: Output table renderer for Portfolio

**Files:**
- Create: `internal/output/table.go`
- Create: `internal/output/table_test.go`

- [ ] **Step 1: Add tablewriter dependency**

Run:
```bash
go get github.com/olekukonko/tablewriter@v0.0.5
go mod tidy
```

Expected: `go.sum` updated, no errors.

- [ ] **Step 2: Write the failing table test**

Create `internal/output/table_test.go`:

```go
package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/herocod3r/robinhood-cli/internal/robinhood/endpoints"
)

func TestTable_Portfolio(t *testing.T) {
	var buf bytes.Buffer
	w := &TableWriter{Out: &buf}

	if err := w.WritePortfolio(&endpoints.PortfolioSummary{
		Equity:              "98234.50",
		ExtendedHoursEquity: "98240.12",
		MarketValue:         "72450.10",
		Cash:                "25784.40",
		BuyingPower:         "25784.40",
	}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"Equity", "Cash", "98234.50", "25784.40"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output, got:\n%s", want, out)
		}
	}
}

func TestTable_Error(t *testing.T) {
	var buf bytes.Buffer
	w := &TableWriter{Out: &buf}
	err := w.WriteError("portfolio", &testErr{msg: "boom", hint: "run: rh login"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "boom") {
		t.Errorf("expected 'boom' in error output: %s", buf.String())
	}
}

type testErr struct{ msg, hint string }

func (e *testErr) Error() string { return e.msg }
```

- [ ] **Step 3: Run the test (should fail)**

Run:
```bash
go test ./internal/output/...
```

Expected: build failure.

- [ ] **Step 4: Implement table writer**

Create `internal/output/table.go`:

```go
package output

import (
	"fmt"
	"io"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
	"github.com/herocod3r/robinhood-cli/internal/robinhood/endpoints"
	"github.com/olekukonko/tablewriter"
)

// TableWriter renders structured data as ASCII tables suitable for a terminal.
type TableWriter struct {
	Out io.Writer
}

// WritePortfolio renders a portfolio summary as a key/value table.
func (w *TableWriter) WritePortfolio(p *endpoints.PortfolioSummary) error {
	t := tablewriter.NewWriter(w.Out)
	t.SetHeader([]string{"Field", "Value"})
	t.SetBorder(false)
	t.Append([]string{"Equity", string(p.Equity)})
	t.Append([]string{"Extended hours equity", string(p.ExtendedHoursEquity)})
	t.Append([]string{"Market value", string(p.MarketValue)})
	t.Append([]string{"Cash", string(p.Cash)})
	t.Append([]string{"Buying power", string(p.BuyingPower)})
	t.Render()
	return nil
}

// WriteError renders a CLI-friendly error line. Always writes to Out even if that
// is stderr — caller decides.
func (w *TableWriter) WriteError(command string, err error) error {
	if apiErr, ok := err.(*robinhood.APIError); ok {
		_, perr := fmt.Fprintf(w.Out, "error (%s): %s\n", apiErr.Code, apiErr.Message)
		if perr != nil {
			return perr
		}
		if apiErr.Hint != "" {
			_, _ = fmt.Fprintf(w.Out, "hint: %s\n", apiErr.Hint)
		}
		return nil
	}
	_, perr := fmt.Fprintf(w.Out, "error (%s): %v\n", command, err)
	return perr
}
```

- [ ] **Step 5: Run the tests (should pass)**

Run:
```bash
go test ./internal/output/...
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add go.mod go.sum internal/output/table.go internal/output/table_test.go
git commit -m "feat(output): TableWriter with WritePortfolio and WriteError"
```

---

## Task 11: Writer interface + TTY detection

**Files:**
- Create: `internal/output/output.go`
- Create: `internal/output/output_test.go`

- [ ] **Step 1: Add isatty dependency**

Run:
```bash
go get github.com/mattn/go-isatty@v0.0.20
go mod tidy
```

Expected: `go.sum` updated.

- [ ] **Step 2: Write the failing test**

Create `internal/output/output_test.go`:

```go
package output

import (
	"bytes"
	"testing"
)

func TestChoose_ExplicitJSON(t *testing.T) {
	if ChooseMode(ModeJSON, false) != ModeJSON {
		t.Error("explicit JSON must stick")
	}
}

func TestChoose_ExplicitPretty(t *testing.T) {
	if ChooseMode(ModePretty, true) != ModePretty {
		t.Error("explicit Pretty must stick")
	}
}

func TestChoose_AutoTTY(t *testing.T) {
	if ChooseMode(ModeAuto, true) != ModePretty {
		t.Error("auto + tty should yield Pretty")
	}
	if ChooseMode(ModeAuto, false) != ModeJSON {
		t.Error("auto + non-tty should yield JSON")
	}
}

func TestNewWriter_JSON(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(ModeJSON, &buf)
	_, ok := w.(*JSONWriter)
	if !ok {
		t.Errorf("ModeJSON should produce *JSONWriter, got %T", w)
	}
}

func TestNewWriter_Pretty(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(ModePretty, &buf)
	_, ok := w.(*TableWriter)
	if !ok {
		t.Errorf("ModePretty should produce *TableWriter, got %T", w)
	}
}
```

- [ ] **Step 3: Run the test (should fail)**

Run:
```bash
go test ./internal/output/...
```

Expected: build failure.

- [ ] **Step 4: Implement output.go**

Create `internal/output/output.go`:

```go
package output

import (
	"io"
	"os"

	"github.com/mattn/go-isatty"
)

// Mode selects the output renderer.
type Mode int

const (
	ModeAuto   Mode = iota // choose based on TTY
	ModeJSON               // force JSON
	ModePretty             // force pretty table
)

// Writer is the minimal contract every renderer satisfies.
// Concrete writers (*JSONWriter, *TableWriter) also expose
// per-command methods (e.g., WritePortfolio); callers typecheck
// when they need those.
type Writer interface {
	WriteError(command string, err error) error
}

// ChooseMode resolves ModeAuto based on whether stdout is a TTY.
// `isTTY` is injected rather than probed so tests can exercise both paths.
func ChooseMode(m Mode, isTTY bool) Mode {
	if m != ModeAuto {
		return m
	}
	if isTTY {
		return ModePretty
	}
	return ModeJSON
}

// IsStdoutTTY reports whether os.Stdout is attached to a terminal.
func IsStdoutTTY() bool {
	return isatty.IsTerminal(os.Stdout.Fd())
}

// NewWriter returns a writer for the resolved mode.
// Callers usually call NewWriter(ChooseMode(flag, IsStdoutTTY()), os.Stdout).
func NewWriter(m Mode, out io.Writer) Writer {
	switch m {
	case ModeJSON:
		return &JSONWriter{Out: out}
	case ModePretty:
		return &TableWriter{Out: out}
	default:
		// Defensive: default to JSON if mode is unresolved.
		return &JSONWriter{Out: out}
	}
}
```

- [ ] **Step 5: Run tests (should pass)**

Run:
```bash
go test ./internal/output/...
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add go.mod go.sum internal/output/output.go internal/output/output_test.go
git commit -m "feat(output): Mode resolution + NewWriter dispatch"
```

---

## Task 12: Root Cobra command with global flags

**Files:**
- Create: `cmd/rh/main.go`
- Create: `cmd/rh/root.go`

- [ ] **Step 1: Add Cobra dependency**

Run:
```bash
go get github.com/spf13/cobra@v1.8.1
go mod tidy
```

Expected: `go.sum` updated.

- [ ] **Step 2: Create `main.go`**

Create `cmd/rh/main.go`:

```go
package main

import (
	"fmt"
	"os"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

func main() {
	root := newRootCommand()
	if err := root.Execute(); err != nil {
		exit := 1
		if apiErr, ok := err.(*robinhood.APIError); ok {
			exit = apiErr.ExitCode()
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exit)
	}
}
```

- [ ] **Step 3: Create `root.go`**

Create `cmd/rh/root.go`:

```go
package main

import (
	"github.com/herocod3r/robinhood-cli/internal/output"
	"github.com/spf13/cobra"
)

// globalFlags holds flags applied to the root command and inherited by children.
type globalFlags struct {
	JSON    bool
	Pretty  bool
	Profile string
	NoColor bool
	Verbose bool
}

var gflags globalFlags

// newRootCommand builds the root Cobra tree. Subcommands register themselves
// by calling rootCmd.AddCommand in their file's init().
func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "rh",
		Short:        "Read-only Robinhood CLI for portfolio research",
		SilenceUsage: true,
		// SilenceErrors: main() prints errors itself so it can choose the exit code.
		SilenceErrors: true,
	}
	cmd.PersistentFlags().BoolVar(&gflags.JSON, "json", false, "force JSON output")
	cmd.PersistentFlags().BoolVar(&gflags.Pretty, "pretty", false, "force pretty/table output")
	cmd.PersistentFlags().StringVar(&gflags.Profile, "profile", "default", "account profile")
	cmd.PersistentFlags().BoolVar(&gflags.NoColor, "no-color", false, "disable ANSI color")
	cmd.PersistentFlags().BoolVarP(&gflags.Verbose, "verbose", "v", false, "debug logging to stderr")

	for _, sub := range registeredCommands {
		cmd.AddCommand(sub)
	}
	return cmd
}

// registeredCommands is populated by subcommand files via init().
var registeredCommands []*cobra.Command

// register appends a subcommand to be wired at root construction.
func register(c *cobra.Command) {
	registeredCommands = append(registeredCommands, c)
}

// resolveOutputMode picks the output mode based on global flags and TTY state.
func resolveOutputMode() output.Mode {
	switch {
	case gflags.JSON:
		return output.ModeJSON
	case gflags.Pretty:
		return output.ModePretty
	default:
		return output.ChooseMode(output.ModeAuto, output.IsStdoutTTY())
	}
}
```

- [ ] **Step 4: Verify build**

Run:
```bash
go build ./cmd/rh
./rh --help
```

Expected: usage text printed. The binary is at repo root.

- [ ] **Step 5: Commit**

```bash
git add go.mod go.sum cmd/rh/main.go cmd/rh/root.go
git commit -m "feat(cmd): Cobra root command with global flags"
```

---

## Task 13: `rh version`

**Files:**
- Create: `cmd/rh/version.go`

- [ ] **Step 1: Implement `rh version`**

Create `cmd/rh/version.go`:

```go
package main

import (
	"fmt"
	"os"

	"github.com/herocod3r/robinhood-cli/internal/buildinfo"
	"github.com/herocod3r/robinhood-cli/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	c := &cobra.Command{
		Use:   "version",
		Short: "Print CLI version and schema version",
		RunE:  runVersion,
	}
	register(c)
}

type versionPayload struct {
	Version       string `json:"version"`
	Commit        string `json:"commit"`
	SchemaVersion string `json:"schema_version"`
}

func runVersion(cmd *cobra.Command, args []string) error {
	payload := versionPayload{
		Version:       buildinfo.Version,
		Commit:        buildinfo.Commit,
		SchemaVersion: buildinfo.SchemaVersion,
	}

	mode := resolveOutputMode()
	switch mode {
	case output.ModeJSON:
		w := &output.JSONWriter{Out: os.Stdout}
		return w.WriteSuccess("version", payload, nil)
	default:
		_, err := fmt.Fprintf(os.Stdout, "rh %s (%s) — schema %s\n",
			payload.Version, payload.Commit, payload.SchemaVersion)
		return err
	}
}
```

- [ ] **Step 2: Build and smoke-test**

Run:
```bash
go build -ldflags="-X github.com/herocod3r/robinhood-cli/internal/buildinfo.Version=0.0.1 -X github.com/herocod3r/robinhood-cli/internal/buildinfo.Commit=deadbeef" -o rh ./cmd/rh
./rh version --json
./rh version
```

Expected — for `--json` (first call), roughly:

```json
{"schema":"robinhood-cli/v1","command":"version","generated_at":"2026-…","data":{"version":"0.0.1","commit":"deadbeef","schema_version":"robinhood-cli/v1"},"error":null}
```

For the second call: `rh 0.0.1 (deadbeef) — schema robinhood-cli/v1`.

- [ ] **Step 3: Commit**

```bash
git add cmd/rh/version.go
git commit -m "feat(cmd): rh version with JSON and TTY modes"
```

---

## Task 14: Shared helper — build a ready client from env

**Files:**
- Create: `cmd/rh/client.go`

Every data command needs a client initialized from env vars. Build the helper once.

- [ ] **Step 1: Implement the helper**

Create `cmd/rh/client.go`:

```go
package main

import (
	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

// newAuthedClient returns a Robinhood client with a Session loaded from env vars.
// Returns the raw *robinhood.APIError on failure so main() can map to exit codes.
func newAuthedClient() (*robinhood.Client, error) {
	sess, err := robinhood.LoadSessionFromEnv()
	if err != nil {
		return nil, err
	}
	c := robinhood.NewClient()
	c.SetSession(sess)
	return c, nil
}
```

- [ ] **Step 2: Verify build**

Run:
```bash
go build ./cmd/rh
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add cmd/rh/client.go
git commit -m "feat(cmd): newAuthedClient helper reading env vars"
```

---

## Task 15: `rh portfolio`

**Files:**
- Create: `cmd/rh/portfolio.go`

- [ ] **Step 1: Implement `rh portfolio`**

Create `cmd/rh/portfolio.go`:

```go
package main

import (
	"os"

	"github.com/herocod3r/robinhood-cli/internal/output"
	"github.com/herocod3r/robinhood-cli/internal/robinhood/endpoints"
	"github.com/spf13/cobra"
)

func init() {
	c := &cobra.Command{
		Use:     "portfolio",
		Aliases: []string{"pf"},
		Short:   "Show account portfolio value, equity, cash, buying power",
		RunE:    runPortfolio,
	}
	register(c)
}

func runPortfolio(cmd *cobra.Command, args []string) error {
	client, err := newAuthedClient()
	if err != nil {
		writeErr(os.Stderr, "portfolio", err)
		return err
	}

	p := endpoints.NewPortfolio(client)
	summary, err := p.Get()
	if err != nil {
		writeErr(os.Stderr, "portfolio", err)
		return err
	}

	mode := resolveOutputMode()
	switch mode {
	case output.ModeJSON:
		w := &output.JSONWriter{Out: os.Stdout}
		return w.WriteSuccess("portfolio", summary, nil)
	default:
		w := &output.TableWriter{Out: os.Stdout}
		return w.WritePortfolio(summary)
	}
}

// writeErr emits the error in the same mode as the selected output, to the
// given writer (usually stderr). Keeps error UX consistent across commands.
func writeErr(w *os.File, command string, err error) {
	mode := resolveOutputMode()
	switch mode {
	case output.ModeJSON:
		jw := &output.JSONWriter{Out: w}
		_ = jw.WriteError(command, err)
	default:
		tw := &output.TableWriter{Out: w}
		_ = tw.WriteError(command, err)
	}
}
```

- [ ] **Step 2: Build and dry-run against a local fake**

Run:
```bash
go build ./cmd/rh
./rh portfolio --json 2>&1 || true
```

Expected: JSON envelope with `error.code = "unauthenticated"` (no env vars set), exit code 2.

- [ ] **Step 3: Commit**

```bash
git add cmd/rh/portfolio.go
git commit -m "feat(cmd): rh portfolio with TTY-aware rendering and env-scoped error surface"
```

---

## Task 16: `rh commands` discovery

**Files:**
- Create: `cmd/rh/commands.go`
- Create: `cmd/rh/commands_test.go`

Rather than hand-maintain a JSON table, reflect over Cobra's tree. Tests live as a package-external test to avoid import cycles.

- [ ] **Step 1: Write the failing test (package-external, uses `os/exec`)**

Create `cmd/rh/commands_test.go`:

```go
package main

import (
	"encoding/json"
	"testing"
)

func TestCommandsOutput_HasPortfolio(t *testing.T) {
	root := newRootCommand()
	data := buildCommandsPayload(root)

	var found bool
	for _, c := range data {
		if c.Name == "portfolio" {
			found = true
			if len(c.Aliases) == 0 || c.Aliases[0] != "pf" {
				t.Errorf("portfolio aliases = %v", c.Aliases)
			}
		}
	}
	if !found {
		t.Error("portfolio should appear in commands payload")
	}
}

func TestCommandsOutput_JSONRoundTrip(t *testing.T) {
	root := newRootCommand()
	data := buildCommandsPayload(root)
	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	var out []CommandMeta
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if len(out) != len(data) {
		t.Errorf("round-trip len mismatch")
	}
}
```

- [ ] **Step 2: Run test (should fail)**

Run:
```bash
go test ./cmd/rh/...
```

Expected: `buildCommandsPayload undefined`.

- [ ] **Step 3: Implement `rh commands`**

Create `cmd/rh/commands.go`:

```go
package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/herocod3r/robinhood-cli/internal/output"
	"github.com/herocod3r/robinhood-cli/internal/robinhood"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() {
	c := &cobra.Command{
		Use:   "commands [name]",
		Short: "List commands (for scripts/skills); pass a name for one command's detail",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runCommands,
	}
	register(c)
}

// CommandMeta is the stable discovery payload.
type CommandMeta struct {
	Name     string     `json:"name"`
	Aliases  []string   `json:"aliases,omitempty"`
	Summary  string     `json:"summary"`
	Category string     `json:"category,omitempty"`
	Args     []string   `json:"args,omitempty"`
	Flags    []FlagMeta `json:"flags,omitempty"`
}

// FlagMeta is a single flag in the discovery payload.
type FlagMeta struct {
	Name        string `json:"name"`
	Shorthand   string `json:"shorthand,omitempty"`
	Type        string `json:"type"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description,omitempty"`
}

// buildCommandsPayload walks the Cobra tree and produces the stable discovery list.
func buildCommandsPayload(root *cobra.Command) []CommandMeta {
	var out []CommandMeta
	for _, c := range root.Commands() {
		if c.Hidden {
			continue
		}
		out = append(out, CommandMeta{
			Name:    c.Name(),
			Aliases: c.Aliases,
			Summary: c.Short,
			Flags:   flagsFor(c),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// flagsFor returns the local (non-inherited) flags of a command.
func flagsFor(c *cobra.Command) []FlagMeta {
	var out []FlagMeta
	c.LocalFlags().VisitAll(func(f *pflag.Flag) {
		out = append(out, FlagMeta{
			Name:        f.Name,
			Shorthand:   f.Shorthand,
			Type:        f.Value.Type(),
			Default:     f.DefValue,
			Description: f.Usage,
		})
	})
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func runCommands(cmd *cobra.Command, args []string) error {
	root := cmd.Root()
	payload := buildCommandsPayload(root)

	if len(args) == 1 {
		name := args[0]
		for _, c := range payload {
			if c.Name == name {
				return emitSingle(c)
			}
		}
		return &robinhood.APIError{
			Code:    robinhood.CodeValidation,
			Message: fmt.Sprintf("unknown command: %s", name),
			Hint:    "run: rh commands --json",
		}
	}

	mode := resolveOutputMode()
	switch mode {
	case output.ModeJSON:
		w := &output.JSONWriter{Out: os.Stdout}
		return w.WriteSuccess("commands", payload, map[string]int{"count": len(payload)})
	default:
		return printCommandsTable(os.Stdout, payload)
	}
}

func emitSingle(c CommandMeta) error {
	w := &output.JSONWriter{Out: os.Stdout}
	return w.WriteSuccess("commands", c, nil)
}

func printCommandsTable(out io.Writer, payload []CommandMeta) error {
	for _, c := range payload {
		aliases := ""
		if len(c.Aliases) > 0 {
			aliases = " (" + strings.Join(c.Aliases, ", ") + ")"
		}
		if _, err := fmt.Fprintf(out, "%-20s%s  %s\n", c.Name+aliases, "", c.Summary); err != nil {
			return err
		}
	}
	return nil
}
```

- [ ] **Step 4: Run tests (should pass)**

Run:
```bash
go test ./cmd/rh/...
```

Expected: PASS for `TestCommandsOutput_HasPortfolio` and `TestCommandsOutput_JSONRoundTrip`.

- [ ] **Step 5: Smoke-test**

Run:
```bash
go build ./cmd/rh
./rh commands --json | head -40
./rh commands
./rh commands portfolio --json
```

Expected: JSON list, plain-text list, single-command JSON respectively.

- [ ] **Step 6: Commit**

```bash
git add cmd/rh/commands.go cmd/rh/commands_test.go
git commit -m "feat(cmd): rh commands discovery from Cobra tree"
```

---

## Task 17: `rh schema` (stub returning the envelope schema doc)

**Files:**
- Create: `cmd/rh/schema.go`

For Plan A, `rh schema` returns the stable envelope schema only. Per-command data schemas arrive in Plan B once more endpoints exist.

- [ ] **Step 1: Implement `rh schema`**

Create `cmd/rh/schema.go`:

```go
package main

import (
	"os"

	"github.com/herocod3r/robinhood-cli/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	c := &cobra.Command{
		Use:   "schema [command]",
		Short: "Emit JSON schema information (Plan A: envelope only)",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runSchema,
	}
	register(c)
}

// envelopeSchema is a hand-written description of the stable envelope shape.
// Expanded into per-command data shapes in Plan B as endpoints land.
var envelopeSchema = map[string]any{
	"envelope": map[string]any{
		"type":  "object",
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
```

- [ ] **Step 2: Smoke-test**

Run:
```bash
go build ./cmd/rh
./rh schema --json | head -40
./rh schema envelope --json | head -40
```

Expected: JSON payloads with the envelope schema.

- [ ] **Step 3: Commit**

```bash
git add cmd/rh/schema.go
git commit -m "feat(cmd): rh schema emitting stable envelope shape (Plan A scope)"
```

---

## Task 18: CI workflow (test + vet + lint)

**Files:**
- Create: `.github/workflows/ci.yaml`

- [ ] **Step 1: Write the workflow**

Create `.github/workflows/ci.yaml`:

```yaml
name: ci

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
          check-latest: true
          cache: true
      - name: Tidy check
        run: |
          go mod tidy
          git diff --exit-code go.mod go.sum
      - name: Vet
        run: go vet ./...
      - name: Test
        run: go test -race -coverprofile=coverage.out ./...
      - name: Coverage summary
        run: go tool cover -func=coverage.out | tail -n 1

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
          cache: true
      - uses: golangci/golangci-lint-action@v6
        with:
          version: v1.61.0
          args: --timeout=3m

  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
          cache: true
      - name: Build
        run: go build -o rh ./cmd/rh
      - name: Smoke tests
        run: |
          ./rh version --json
          ./rh commands --json | head -5
          ./rh schema --json | head -10
```

- [ ] **Step 2: Verify locally (what CI will run)**

Run:
```bash
go mod tidy
go vet ./...
go test -race ./...
go build -o rh ./cmd/rh
./rh version --json
./rh commands --json | head -5
./rh schema --json | head -10
```

Expected: every command succeeds; JSON envelopes printed. If `go mod tidy` changes go.sum/go.mod, include those changes in the commit.

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/ci.yaml go.mod go.sum
git commit -m "chore(ci): GitHub Actions running vet, test, lint, build smoke"
```

---

## Task 19: Exit-code and error-envelope integration test

A high-signal integration test that binds the error taxonomy to the CLI exit code — catches silent regressions where a refactor drops a branch.

**Files:**
- Create: `cmd/rh/exit_test.go`

- [ ] **Step 1: Write the failing test**

Create `cmd/rh/exit_test.go`:

```go
package main

import (
	"testing"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

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
```

- [ ] **Step 2: Run tests (should pass — logic already implemented in earlier tasks)**

Run:
```bash
go test ./cmd/rh/...
```

Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git add cmd/rh/exit_test.go
git commit -m "test(cmd): lock exit-code mapping and subcommand registration"
```

---

## Task 20: Plan A README update + wrap-up smoke

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Update README with Plan A usage snippet**

Replace `README.md` with:

```markdown
# rh — Robinhood CLI

Open-source, read-only Robinhood CLI for portfolio research and trade analysis.

> **⚠️ Research tool only.** Read-only. Uses an unofficial API; may break without notice.

## Status

**Plan A (foundation) implemented.** The binary builds, emits the stable JSON
envelope, and has one working data command (`rh portfolio`). Interactive
authentication (Sheriff flow, `rh login`) ships in Plan B.

## Build

```bash
make build
./rh --help
./rh version --json
./rh commands --json
./rh schema --json
```

## Using `rh portfolio` (Plan A auth — env vars)

Plan A does not yet implement interactive login. Supply pre-obtained tokens
(for example, extracted from a `robin_stocks` session, or from Plan B once it
ships) via environment variables:

```bash
export ROBINHOOD_ACCESS_TOKEN=...
export ROBINHOOD_REFRESH_TOKEN=...
export ROBINHOOD_DEVICE_TOKEN=...          # optional; reduces Sheriff prompts
./rh portfolio --json
```

Exit codes:

| Code | Meaning |
|---|---|
| 0 | OK |
| 1 | Runtime or network error |
| 2 | Not authenticated / session expired |
| 3 | Sheriff interactive step required |
| 4 | Rate limited by Robinhood |
| 5 | Bad flags or args |

## Development

```bash
make test          # unit tests
make test-race     # with race detector
make lint          # golangci-lint
make build         # local binary
```

Design spec: `docs/superpowers/specs/2026-04-18-robinhood-cli-design.md`
Plan A (this plan): `docs/superpowers/plans/2026-04-18-robinhood-cli-plan-a-foundation.md`

## License

MIT
```

- [ ] **Step 2: Run the full local suite one more time**

Run:
```bash
go mod tidy
go vet ./...
go test -race ./...
make build
./rh version
./rh commands
./rh schema --json | head -20
./rh portfolio --json 2>&1 || true   # expect exit 2 with unauthenticated envelope
```

Expected: every step clean; `rh portfolio --json` emits a JSON envelope with `error.code = "unauthenticated"` and exits 2.

- [ ] **Step 3: Commit**

```bash
git add README.md
git commit -m "docs: update README for Plan A status and usage"
```

---

## Plan A Done — Acceptance Checklist

Run through this list before calling Plan A shipped:

- [ ] `make test` is green
- [ ] `make test-race` is green
- [ ] `make lint` is green
- [ ] `make build` produces a `./rh` binary
- [ ] `./rh version --json` prints a valid envelope with `schema: "robinhood-cli/v1"`
- [ ] `./rh commands --json` lists at least `portfolio`, `version`, `commands`, `schema`
- [ ] `./rh schema --json` prints the envelope schema
- [ ] `./rh portfolio --json` with no env vars returns `error.code = "unauthenticated"` and exits 2
- [ ] `./rh portfolio --json` with valid tokens returns a success envelope with `data.equity` populated (manual check with real tokens — outside CI)
- [ ] CI workflow passes on the branch

**Tasks in Plan B (next):** Interactive `rh login` with Sheriff workflow, OS keychain storage, remaining 16 data commands (`positions`, `position`, `account`, `quote`, `fundamentals`, `historicals`, `news`, `earnings`, `ratings`, `dividends`, `options-positions`, `orders`, `watchlist`, `search`, `market-hours`, `documents`), `rh logout`.

**Tasks in Plan C (after B):** SKILL.md authoring + `rh skill install` with `embed.FS`, GoReleaser + Homebrew tap, release workflow, full documentation (`docs/JSON_SCHEMA.md`, `docs/ROBINHOOD_API.md`, `docs/AUTH.md`, `docs/E2E.md`).
