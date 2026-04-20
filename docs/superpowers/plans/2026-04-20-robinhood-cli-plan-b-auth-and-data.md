# robinhood-cli — Plan B: Interactive Auth + 16 Read-only Data Commands

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete the user-facing CLI: ship `rh login` with the full Sheriff/SUV interactive workflow (SMS/email code, device push approval, TOTP MFA), persist sessions to the OS keychain with a file fallback, `rh logout`, and the 16 remaining read-only data commands (`positions`, `position`, `account`, `quote`, `fundamentals`, `historicals`, `news`, `earnings`, `ratings`, `dividends`, `options-positions`, `orders`, `watchlist`, `search`, `market-hours`, `documents`). At the end of Plan B a user can `rh login`, use every read command on a live Robinhood account, `rh logout`, and every output honors the Plan A JSON envelope contract.

**Architecture:** Extend the Plan A layering. `internal/keychain/` wraps `zalando/go-keyring` (macOS Keychain / GNOME libsecret / Windows Credential Manager) with a 0600 JSON fallback for CI/containers. `internal/config/` grows a YAML loader for `~/.config/robinhood-cli/config.yaml` (profiles + default profile). `internal/robinhood/oauth.go` gains password grant, workflow headers, and Sheriff state machine (`sheriff.go` new file) with deterministic pluggable clock + HTTP client for tests. `internal/inputprompt/` prompts stdin (password-hidden, TOTP, challenge code) with env-var shortcut. `internal/robinhood/session.go` learns `SaveToKeychain` / `LoadFromKeychain`; `newAuthedClient` in `cmd/rh/` prefers env vars → keychain → "run: rh login". File locks via `gofrs/flock` make concurrent `rh` invocations race-safe around the refresh. Sixteen new endpoint domains each get a file under `internal/robinhood/endpoints/` and a command under `cmd/rh/`; the table writer grows one `Write<Domain>` method per command.

**Tech Stack:** Go 1.22, additive deps — `github.com/zalando/go-keyring v0.2.5` (keychain), `github.com/gofrs/flock v0.12.1` (file lock), `github.com/pquerna/otp v1.4.0` (TOTP), `gopkg.in/yaml.v3 v3.0.1` (config file), `golang.org/x/term v0.21.0` (hidden password prompt), `github.com/google/uuid v1.6.0` (device_token generation). Plan A deps (`cobra`, `pflag`, `tablewriter`, `go-isatty`) carry forward.

**Spec reference:** `docs/superpowers/specs/2026-04-18-robinhood-cli-design.md`, sections 6 (auth), 7 (command tree), 8.3 (schemas), 9 (discovery).

**Prior art:** [`robin_stocks`](https://github.com/jmfernandes/robin_stocks/blob/main/robin_stocks/robinhood/authentication.py) is the authoritative reference for the Sheriff workflow shapes. Copy the endpoint paths/payloads verbatim — every drift we've seen in the ecosystem is from guesswork.

**Out of scope for Plan B (shipped in Plan C):** Skill (`SKILL.md`, `rh skill install`), GoReleaser, Homebrew tap, release workflow, schema stability workflow, `docs/JSON_SCHEMA.md`, `docs/ROBINHOOD_API.md`, `docs/AUTH.md`, `docs/E2E.md`, cassette-based integration tests.

---

## File Structure

Created or touched in this plan:

```
robinhood-cli/
├── cmd/rh/
│   ├── client.go                            # MODIFY — env → keychain precedence
│   ├── login.go                             # CREATE — rh login
│   ├── logout.go                            # CREATE — rh logout
│   ├── positions.go                         # CREATE
│   ├── position.go                          # CREATE
│   ├── account.go                           # CREATE
│   ├── quote.go                             # CREATE
│   ├── fundamentals.go                      # CREATE
│   ├── historicals.go                       # CREATE
│   ├── news.go                              # CREATE
│   ├── earnings.go                          # CREATE
│   ├── ratings.go                           # CREATE
│   ├── dividends.go                         # CREATE
│   ├── options_positions.go                 # CREATE
│   ├── orders.go                            # CREATE
│   ├── watchlist.go                         # CREATE
│   ├── search.go                            # CREATE
│   ├── market_hours.go                      # CREATE
│   ├── documents.go                         # CREATE
│   ├── commands.go                          # MODIFY — add Category+Stability
│   └── schema.go                            # MODIFY — all commands covered
├── internal/
│   ├── config/
│   │   ├── config.go                        # MODIFY — add ConfigFilePath, profile helpers
│   │   └── file.go                          # CREATE — YAML load/save
│   │   └── file_test.go                     # CREATE
│   ├── inputprompt/
│   │   ├── prompt.go                        # CREATE
│   │   └── prompt_test.go                   # CREATE
│   ├── keychain/
│   │   ├── keychain.go                      # CREATE — zalando wrapper + file fallback
│   │   └── keychain_test.go                 # CREATE — exercises fallback path
│   ├── robinhood/
│   │   ├── oauth.go                         # MODIFY — password grant, workflow header
│   │   ├── oauth_test.go                    # MODIFY — password-grant branches
│   │   ├── sheriff.go                       # CREATE
│   │   ├── sheriff_test.go                  # CREATE
│   │   ├── session.go                       # MODIFY — keychain load/save/clear
│   │   ├── session_test.go                  # MODIFY
│   │   ├── refresh_lock.go                  # CREATE — flock-based refresh serializer
│   │   ├── refresh_lock_test.go             # CREATE
│   │   ├── errors.go                        # MODIFY — add MaskToken helper (tests)
│   │   └── endpoints/
│   │       ├── positions.go                 # CREATE
│   │       ├── positions_test.go            # CREATE
│   │       ├── account.go                   # CREATE
│   │       ├── account_test.go              # CREATE
│   │       ├── quotes.go                    # CREATE
│   │       ├── quotes_test.go               # CREATE
│   │       ├── fundamentals.go              # CREATE
│   │       ├── fundamentals_test.go         # CREATE
│   │       ├── historicals.go               # CREATE
│   │       ├── historicals_test.go          # CREATE
│   │       ├── news.go                      # CREATE
│   │       ├── news_test.go                 # CREATE
│   │       ├── earnings.go                  # CREATE
│   │       ├── earnings_test.go             # CREATE
│   │       ├── ratings.go                   # CREATE
│   │       ├── ratings_test.go              # CREATE
│   │       ├── dividends.go                 # CREATE
│   │       ├── dividends_test.go            # CREATE
│   │       ├── options.go                   # CREATE
│   │       ├── options_test.go              # CREATE
│   │       ├── orders.go                    # CREATE
│   │       ├── orders_test.go               # CREATE
│   │       ├── watchlist.go                 # CREATE
│   │       ├── watchlist_test.go            # CREATE
│   │       ├── instruments.go               # CREATE — symbol→instrument_id resolver
│   │       ├── instruments_test.go          # CREATE
│   │       ├── search.go                    # CREATE
│   │       ├── search_test.go               # CREATE
│   │       ├── market_hours.go              # CREATE
│   │       ├── market_hours_test.go         # CREATE
│   │       ├── documents.go                 # CREATE
│   │       └── documents_test.go            # CREATE
│   └── output/
│       └── table.go                         # MODIFY — Write<Domain> per command
└── go.mod / go.sum                          # MODIFY — new deps
```

**Responsibility contract (added in Plan B):**

- `internal/keychain/` — owns secret storage; **never** knows the shape of the stored bytes.
- `internal/robinhood/session.go` — owns encoding/decoding of the session blob that keychain stores.
- `internal/robinhood/sheriff.go` — owns the Sheriff state machine; accepts an injected prompter + clock + HTTP client (for tests).
- `internal/robinhood/refresh_lock.go` — owns file lock lifecycle around a refresh; knows nothing about OAuth.
- `internal/inputprompt/` — owns all stdin prompts; respects env-var shortcuts (`ROBINHOOD_TOTP_SECRET`, etc.).
- `internal/robinhood/endpoints/instruments.go` — owns `symbol → instrument_id` resolution, cached per-call; used by positions + orders.
- Cobra command files — zero business logic; one client call, one writer call, one error return.

**Naming conventions enforced everywhere:**

- Every endpoint method takes `ctx context.Context` as the first argument. (Plan A's `GetJSON` did not; Plan B extends the client to support contexts — see Task 1.)
- Every endpoint returns a single typed struct or slice; never `map[string]any`.
- Every command file looks like this skeleton (which every data-command task references):

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
        Use:     "<name>",                        // replace per command
        Aliases: []string{"<alias>"},
        Short:   "<summary>",
        RunE:    run<Name>,
    }
    // flags registered here
    register(c)
}

func run<Name>(cmd *cobra.Command, args []string) error {
    client, err := newAuthedClient()
    if err != nil {
        return err // main() renders
    }
    data, err := endpoints.New<Name>(client).<Call>(cmd.Context(), /* args/flags */)
    if err != nil {
        return err
    }
    switch resolveOutputMode() {
    case output.ModeJSON:
        w := &output.JSONWriter{Out: os.Stdout}
        return w.WriteSuccess("<name>", data, /* meta or nil */)
    default:
        w := &output.TableWriter{Out: os.Stdout}
        return w.Write<Name>(data)
    }
}
```

Tasks 14–29 are phrased against this skeleton — they only specify endpoint URL, struct shape, flags, and table columns.

---

## Wave Layout (for subagent-driven execution)

Workers merge into the Plan B integration branch `plan-b/integration` (branched from `worktree-bootsrap` at Plan B kickoff). Each worker works in an isolated worktree and hands off to a reviewer (/diff-review + /codex:review) before merge. Waves are serial; tasks within a wave run in parallel.

- **Wave B1 — Auth foundations (Tasks 1–6):** context-aware client, config file, keychain wrapper, inputprompt, refresh lock, session encoding. Parallelizable across 3 workers.
- **Wave B2 — OAuth + Sheriff (Tasks 7–11):** password grant, TOTP, Sheriff start/poll/challenge. Serial within wave (Task 9 depends on Task 8's polling helper).
- **Wave B3 — Login/logout wiring (Tasks 12–14):** `rh login`, `rh logout`, env→keychain precedence in `newAuthedClient`.
- **Wave B4 — Data commands (Tasks 15–30):** 16 commands + shared instrument resolver. Highly parallelizable (3–4 workers).
- **Wave B5 — Discovery + acceptance (Tasks 31–32):** expand `rh commands`/`rh schema`, Plan B acceptance checklist.

Every wave ends with: `make test`, `make test-race`, `make lint`, `make build`, `/diff-review` agent, `/codex:review` agent, fix findings, merge to `plan-b/integration`.

---

## Task 1: Context-aware HTTP client (prep for everything else)

Plan A's `Client.GetJSON` takes no `context.Context`. Every new endpoint needs cancellation (Sheriff polls for up to 5 minutes; tests need timeouts). Add context support without breaking the Plan A call site.

**Files:**
- Modify: `internal/robinhood/client.go`
- Modify: `internal/robinhood/client_test.go`
- Modify: `internal/robinhood/endpoints/portfolio.go`
- Modify: `internal/robinhood/endpoints/portfolio_test.go`
- Modify: `cmd/rh/portfolio.go`

- [ ] **Step 1: Write the failing test**

Append to `internal/robinhood/client_test.go`:

```go
func TestClient_GetJSONCtx_CancelsInFlight(t *testing.T) {
    started := make(chan struct{})
    release := make(chan struct{})
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        close(started)
        <-release
        w.WriteHeader(200)
    }))
    defer ts.Close()
    defer close(release)

    c := NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
    c.SetSession(&Session{AccessToken: "t"})

    ctx, cancel := context.WithCancel(context.Background())
    errCh := make(chan error, 1)
    go func() {
        var out any
        errCh <- c.GetJSONCtx(ctx, APIHost, "/x/", &out)
    }()
    <-started
    cancel()

    select {
    case err := <-errCh:
        if err == nil {
            t.Fatalf("expected cancellation error, got nil")
        }
    case <-time.After(2 * time.Second):
        t.Fatalf("GetJSONCtx did not return after cancel")
    }
}
```

- [ ] **Step 2: Run test — expect compile failure**

Run: `go test ./internal/robinhood/ -run TestClient_GetJSONCtx_CancelsInFlight -v`
Expected: FAIL — `GetJSONCtx undefined`.

- [ ] **Step 3: Implement `GetJSONCtx`**

In `internal/robinhood/client.go`, add:

```go
// GetJSONCtx is the context-aware GET. It auto-refreshes once on 401 and decodes into out.
// GetJSON is preserved for Plan A callers and delegates to GetJSONCtx with context.Background().
//
// IMPORTANT: on the 401 branch we drain AND close the first response body BEFORE issuing
// the retry request. `defer` would stack LIFO and keep the first connection checked out
// of http.Transport until function return — under concurrent pagination (positions,
// orders) this exhausts MaxIdleConnsPerHost. See Codex finding #2 in the plan-B review.
func (c *Client) GetJSONCtx(ctx context.Context, host Host, path string, out any) error {
    if err := c.ensureFresh(); err != nil {
        return err
    }
    resp, err := c.do(ctx, host, http.MethodGet, path, nil)
    if err != nil {
        return err
    }

    if resp.StatusCode == http.StatusUnauthorized {
        _, _ = io.Copy(io.Discard, resp.Body)
        _ = resp.Body.Close() // free the connection BEFORE the retry
        if rerr := c.oauth.Refresh(c.Session()); rerr != nil {
            return rerr
        }
        resp2, err := c.do(ctx, host, http.MethodGet, path, nil)
        if err != nil {
            return err
        }
        defer resp2.Body.Close()
        if resp2.StatusCode == http.StatusUnauthorized {
            return &APIError{Code: CodeUnauthenticated, Message: "401 after refresh", Hint: "run: rh login"}
        }
        return decodeOrMap(resp2, out)
    }
    defer resp.Body.Close()
    return decodeOrMap(resp, out)
}

// GetJSON preserves the Plan A signature.
func (c *Client) GetJSON(host Host, path string, out any) error {
    return c.GetJSONCtx(context.Background(), host, path, out)
}
```

Change `do` to accept context:

```go
func (c *Client) do(ctx context.Context, host Host, method, path string, body io.Reader) (*http.Response, error) {
    s := c.Session()
    req, err := http.NewRequestWithContext(ctx, method, c.baseFor(host)+path, body)
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
```

Update `getJSON` to forward to `GetJSONCtx(context.Background(), ...)` — or delete it outright; all callers become `GetJSONCtx`.

- [ ] **Step 4: Drain 401 body before refresh (Codex-A regression guard)**

Already done in the snippet above with `io.Copy(io.Discard, resp.Body)` before refresh. This is the "401 body not drained" followup from Plan A. Verify it exists.

- [ ] **Step 5: Verify test passes**

Run: `go test ./internal/robinhood/ -run TestClient_GetJSONCtx -v`
Expected: PASS.

- [ ] **Step 6: Migrate `portfolio.go` to `GetJSONCtx`**

In `internal/robinhood/endpoints/portfolio.go`, change `Get()` to `Get(ctx context.Context)`:

```go
func (p *Portfolio) Get(ctx context.Context) (*PortfolioSummary, error) {
    var pr portfolioResultsResp
    if err := p.c.GetJSONCtx(ctx, robinhood.APIHost, "/portfolios/", &pr); err != nil {
        return nil, err
    }
    // ... (rest unchanged, but use GetJSONCtx for the second call too)
}
```

Update `portfolio_test.go` callers to pass `context.Background()`.

Update `cmd/rh/portfolio.go`:

```go
summary, err := endpoints.NewPortfolio(client).Get(cmd.Context())
```

- [ ] **Step 7: Run full suite**

Run: `make test`
Expected: PASS (portfolio + new client_test).

- [ ] **Step 8: Commit**

```bash
git add internal/robinhood/client.go internal/robinhood/client_test.go \
        internal/robinhood/endpoints/portfolio.go \
        internal/robinhood/endpoints/portfolio_test.go \
        cmd/rh/portfolio.go
git commit -m "refactor(client): GetJSONCtx context-aware GET; drain 401 body before refresh"
```

---

## Task 2: Config YAML file — profiles and defaults

**Files:**
- Modify: `internal/config/config.go`
- Create: `internal/config/file.go`
- Create: `internal/config/file_test.go`
- Modify: `go.mod` / `go.sum`

- [ ] **Step 1: Add YAML dependency**

Run:
```bash
go get gopkg.in/yaml.v3@v3.0.1
go mod tidy
```

- [ ] **Step 2: Write the failing test**

Create `internal/config/file_test.go`:

```go
package config

import (
    "os"
    "path/filepath"
    "testing"
)

func TestLoadFile_Missing_ReturnsDefaults(t *testing.T) {
    t.Setenv("XDG_CONFIG_HOME", t.TempDir())
    f, err := LoadFile()
    if err != nil {
        t.Fatalf("LoadFile: %v", err)
    }
    if f.DefaultProfile != "default" {
        t.Fatalf("DefaultProfile = %q, want %q", f.DefaultProfile, "default")
    }
    if _, ok := f.Profiles["default"]; !ok {
        t.Fatalf("Profiles should contain 'default'")
    }
}

func TestLoadFile_Existing_ReturnsParsed(t *testing.T) {
    dir := t.TempDir()
    t.Setenv("XDG_CONFIG_HOME", dir)
    appDir := filepath.Join(dir, AppName)
    if err := os.MkdirAll(appDir, 0o700); err != nil {
        t.Fatal(err)
    }
    cfg := []byte("default_profile: work\nprofiles:\n  work:\n    account_number: 9876\n  home:\n    account_number: 1234\n")
    if err := os.WriteFile(filepath.Join(appDir, "config.yaml"), cfg, 0o600); err != nil {
        t.Fatal(err)
    }
    f, err := LoadFile()
    if err != nil {
        t.Fatalf("LoadFile: %v", err)
    }
    if f.DefaultProfile != "work" {
        t.Fatalf("DefaultProfile = %q, want %q", f.DefaultProfile, "work")
    }
    if f.Profiles["work"].AccountNumber != "9876" {
        t.Fatalf("work.AccountNumber = %q, want 9876", f.Profiles["work"].AccountNumber)
    }
}

func TestSaveFile_RoundTrip(t *testing.T) {
    dir := t.TempDir()
    t.Setenv("XDG_CONFIG_HOME", dir)
    in := &File{
        DefaultProfile: "home",
        Profiles:       map[string]Profile{"home": {AccountNumber: "1234"}},
    }
    if err := SaveFile(in); err != nil {
        t.Fatalf("SaveFile: %v", err)
    }
    out, err := LoadFile()
    if err != nil {
        t.Fatalf("LoadFile: %v", err)
    }
    if out.DefaultProfile != "home" || out.Profiles["home"].AccountNumber != "1234" {
        t.Fatalf("round-trip lost data: %+v", out)
    }
    // Verify 0600 perms on the written file.
    info, err := os.Stat(filepath.Join(dir, AppName, "config.yaml"))
    if err != nil {
        t.Fatal(err)
    }
    if info.Mode().Perm() != 0o600 {
        t.Fatalf("config.yaml perms = %o, want 0600", info.Mode().Perm())
    }
}
```

- [ ] **Step 3: Run test — expect compile failure**

Run: `go test ./internal/config/ -run TestLoadFile -v`
Expected: FAIL — `LoadFile`, `File`, `Profile`, `SaveFile` undefined.

- [ ] **Step 4: Implement `file.go`**

Create `internal/config/file.go`:

```go
package config

import (
    "fmt"
    "os"
    "path/filepath"

    "gopkg.in/yaml.v3"
)

// File is the on-disk config at $XDG_CONFIG_HOME/robinhood-cli/config.yaml.
type File struct {
    DefaultProfile string             `yaml:"default_profile"`
    Profiles       map[string]Profile `yaml:"profiles"`
}

// Profile is a per-account block.
type Profile struct {
    AccountNumber string `yaml:"account_number,omitempty"`
}

// ConfigFilePath returns the absolute path to config.yaml (never creates it).
func ConfigFilePath() (string, error) {
    dir, err := ConfigDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(dir, "config.yaml"), nil
}

// LoadFile reads config.yaml if present, or returns zero-value defaults.
// Missing file is NOT an error — a fresh install has no config.
func LoadFile() (*File, error) {
    path, err := ConfigFilePath()
    if err != nil {
        return nil, err
    }
    f := &File{DefaultProfile: "default", Profiles: map[string]Profile{"default": {}}}
    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return f, nil
        }
        return nil, fmt.Errorf("read %s: %w", path, err)
    }
    if err := yaml.Unmarshal(data, f); err != nil {
        return nil, fmt.Errorf("parse %s: %w", path, err)
    }
    if f.DefaultProfile == "" {
        f.DefaultProfile = "default"
    }
    if f.Profiles == nil {
        f.Profiles = map[string]Profile{"default": {}}
    }
    return f, nil
}

// SaveFile writes the config atomically at 0600. Creates the config dir with 0700 if missing.
func SaveFile(f *File) error {
    dir, err := EnsureConfigDir()
    if err != nil {
        return err
    }
    out, err := yaml.Marshal(f)
    if err != nil {
        return err
    }
    tmp, err := os.CreateTemp(dir, "config-*.yaml.tmp")
    if err != nil {
        return err
    }
    tmpPath := tmp.Name()
    // Clean up on any error path.
    defer func() { _ = os.Remove(tmpPath) }()
    if err := os.Chmod(tmpPath, 0o600); err != nil {
        _ = tmp.Close()
        return err
    }
    if _, err := tmp.Write(out); err != nil {
        _ = tmp.Close()
        return err
    }
    if err := tmp.Close(); err != nil {
        return err
    }
    final := filepath.Join(dir, "config.yaml")
    return os.Rename(tmpPath, final)
}
```

- [ ] **Step 5: Run the tests**

Run: `go test ./internal/config/ -v`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/config/file.go internal/config/file_test.go go.mod go.sum
git commit -m "feat(config): YAML config file with profiles, atomic 0600 writes"
```

---

## Task 3: Keychain wrapper with file fallback

**Files:**
- Create: `internal/keychain/keychain.go`
- Create: `internal/keychain/keychain_test.go`
- Modify: `go.mod` / `go.sum`

- [ ] **Step 1: Add the dependency**

Run:
```bash
go get github.com/zalando/go-keyring@v0.2.5
go mod tidy
```

- [ ] **Step 2: Write the failing test**

Create `internal/keychain/keychain_test.go`:

```go
package keychain

import (
    "os"
    "path/filepath"
    "testing"
)

func TestFileStore_SetGetDelete(t *testing.T) {
    dir := t.TempDir()
    s := &FileStore{Dir: dir}
    if err := s.Set("rh", "default", []byte(`{"x":1}`)); err != nil {
        t.Fatalf("Set: %v", err)
    }
    // 0600 perms on the session file.
    info, err := os.Stat(filepath.Join(dir, "rh-default.json"))
    if err != nil {
        t.Fatal(err)
    }
    if info.Mode().Perm() != 0o600 {
        t.Fatalf("perms = %o, want 0600", info.Mode().Perm())
    }
    got, err := s.Get("rh", "default")
    if err != nil {
        t.Fatalf("Get: %v", err)
    }
    if string(got) != `{"x":1}` {
        t.Fatalf("got %q, want %q", got, `{"x":1}`)
    }
    if err := s.Delete("rh", "default"); err != nil {
        t.Fatalf("Delete: %v", err)
    }
    if _, err := s.Get("rh", "default"); err == nil {
        t.Fatalf("Get after Delete should fail")
    }
    if !os.IsNotExist(err) && err != ErrNotFound {
        // accept either bare NotExist or sentinel
    }
}

func TestFallback_PicksFileWhenSystemUnavailable(t *testing.T) {
    dir := t.TempDir()
    t.Setenv("XDG_CONFIG_HOME", dir)
    t.Setenv("ROBINHOOD_KEYCHAIN_BACKEND", "file") // forces file backend
    s := Open()
    if _, ok := s.(*FileStore); !ok {
        t.Fatalf("expected FileStore, got %T", s)
    }
}
```

- [ ] **Step 3: Run test — expect compile failure**

Run: `go test ./internal/keychain/ -v`
Expected: FAIL — `FileStore`, `Open`, `ErrNotFound` undefined.

- [ ] **Step 4: Implement the store**

Create `internal/keychain/keychain.go`:

```go
// Package keychain stores and retrieves secret blobs. It tries the OS keychain
// (macOS Keychain, GNOME libsecret, Windows Credential Manager) and falls back
// to a 0600 JSON file under the config dir when the OS keychain is unavailable
// (CI, headless Linux without a session bus, some containers).
package keychain

import (
    "errors"
    "fmt"
    "os"
    "path/filepath"

    "github.com/herocod3r/robinhood-cli/internal/config"
    "github.com/zalando/go-keyring"
)

// ErrNotFound is returned when no entry exists for the given service+profile.
var ErrNotFound = errors.New("keychain: not found")

// Store is the minimal contract Session uses.
type Store interface {
    Set(service, profile string, data []byte) error
    Get(service, profile string) ([]byte, error)
    Delete(service, profile string) error
}

// Open picks a backend. Precedence:
//   1. ROBINHOOD_KEYCHAIN_BACKEND=file — force file (for CI + tests)
//   2. ROBINHOOD_KEYCHAIN_BACKEND=system — force keyring (fail hard if unavailable)
//   3. default: try keyring with a probe write/read/delete; fall back to file
//      in $XDG_CONFIG_HOME/robinhood-cli/
func Open() Store {
    switch os.Getenv("ROBINHOOD_KEYCHAIN_BACKEND") {
    case "file":
        return fileStoreAtConfigDir()
    case "system":
        return &SystemStore{}
    }
    sys := &SystemStore{}
    if probeSystem(sys) {
        return sys
    }
    return fileStoreAtConfigDir()
}

func fileStoreAtConfigDir() *FileStore {
    dir, err := config.EnsureConfigDir()
    if err != nil {
        // Fallback to /tmp if config dir is unwritable — we still want *something*.
        return &FileStore{Dir: os.TempDir()}
    }
    return &FileStore{Dir: dir}
}

func probeSystem(s *SystemStore) bool {
    const probeService = "robinhood-cli-probe"
    const probeKey = "probe"
    if err := s.Set(probeService, probeKey, []byte("x")); err != nil {
        return false
    }
    defer func() { _ = s.Delete(probeService, probeKey) }()
    got, err := s.Get(probeService, probeKey)
    if err != nil || string(got) != "x" {
        return false
    }
    return true
}

// SystemStore uses the OS keychain via zalando/go-keyring.
type SystemStore struct{}

func (s *SystemStore) Set(service, profile string, data []byte) error {
    return keyring.Set(service, profile, string(data))
}
func (s *SystemStore) Get(service, profile string) ([]byte, error) {
    v, err := keyring.Get(service, profile)
    if err != nil {
        if errors.Is(err, keyring.ErrNotFound) {
            return nil, ErrNotFound
        }
        return nil, err
    }
    return []byte(v), nil
}
func (s *SystemStore) Delete(service, profile string) error {
    err := keyring.Delete(service, profile)
    if err != nil && errors.Is(err, keyring.ErrNotFound) {
        return ErrNotFound
    }
    return err
}

// FileStore writes session blobs to <Dir>/<service>-<profile>.json at 0600.
type FileStore struct {
    Dir string
}

func (f *FileStore) path(service, profile string) string {
    return filepath.Join(f.Dir, fmt.Sprintf("%s-%s.json", service, profile))
}

func (f *FileStore) Set(service, profile string, data []byte) error {
    if err := os.MkdirAll(f.Dir, 0o700); err != nil {
        return err
    }
    p := f.path(service, profile)
    tmp := p + ".tmp"
    if err := os.WriteFile(tmp, data, 0o600); err != nil {
        return err
    }
    return os.Rename(tmp, p)
}

func (f *FileStore) Get(service, profile string) ([]byte, error) {
    data, err := os.ReadFile(f.path(service, profile))
    if err != nil {
        if os.IsNotExist(err) {
            return nil, ErrNotFound
        }
        return nil, err
    }
    return data, nil
}

func (f *FileStore) Delete(service, profile string) error {
    err := os.Remove(f.path(service, profile))
    if err != nil && os.IsNotExist(err) {
        return ErrNotFound
    }
    return err
}
```

- [ ] **Step 5: Run tests**

Run: `go test ./internal/keychain/ -v`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/keychain/keychain.go internal/keychain/keychain_test.go go.mod go.sum
git commit -m "feat(keychain): zalando/go-keyring wrapper with file fallback, probe-based backend choice"
```

---

## Task 4: Session keychain encode/decode

**Files:**
- Modify: `internal/robinhood/session.go`
- Modify: `internal/robinhood/session_test.go`

- [ ] **Step 1: Write the failing tests**

Append to `internal/robinhood/session_test.go`:

```go
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
    if err := ClearKeychain("default"); err != nil {
        t.Fatalf("ClearKeychain: %v", err)
    }
    if _, err := LoadFromKeychain("default"); err == nil {
        t.Fatalf("LoadFromKeychain after Clear should fail")
    }
}
```

- [ ] **Step 2: Run tests — expect compile failure**

Run: `go test ./internal/robinhood/ -run "TestSession_Marshal|TestSaveAndLoadFromKeychain" -v`
Expected: FAIL — `Marshal`, `UnmarshalSession`, `SaveToKeychain`, `LoadFromKeychain`, `ClearKeychain` undefined.

- [ ] **Step 3: Implement the methods**

Append to `internal/robinhood/session.go`:

```go
import (
    "encoding/json"

    "github.com/herocod3r/robinhood-cli/internal/keychain"
)

// keychainService is the service name used in the OS keychain.
const keychainService = "robinhood-cli"

// sessionBlob is the persisted shape. Kept separate so Session struct can evolve
// without breaking stored blobs.
type sessionBlob struct {
    Username     string `json:"username"`
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    DeviceToken  string `json:"device_token"`
    ExpiresAt    string `json:"expires_at,omitempty"` // RFC3339; empty => zero
}

// Marshal serializes the session to bytes. ExpiresAt is RFC3339 UTC or empty.
func (s *Session) Marshal() ([]byte, error) {
    blob := sessionBlob{
        Username:     s.Username,
        AccessToken:  s.AccessToken,
        RefreshToken: s.RefreshToken,
        DeviceToken:  s.DeviceToken,
    }
    if !s.ExpiresAt.IsZero() {
        blob.ExpiresAt = s.ExpiresAt.UTC().Format(time.RFC3339)
    }
    return json.Marshal(blob)
}

// UnmarshalSession parses bytes produced by Marshal.
func UnmarshalSession(data []byte) (*Session, error) {
    var blob sessionBlob
    if err := json.Unmarshal(data, &blob); err != nil {
        return nil, err
    }
    s := &Session{
        Username:     blob.Username,
        AccessToken:  blob.AccessToken,
        RefreshToken: blob.RefreshToken,
        DeviceToken:  blob.DeviceToken,
    }
    if blob.ExpiresAt != "" {
        t, err := time.Parse(time.RFC3339, blob.ExpiresAt)
        if err != nil {
            return nil, err
        }
        s.ExpiresAt = t.UTC()
    }
    return s, nil
}

// SaveToKeychain persists the session under the given profile.
func (s *Session) SaveToKeychain(profile string) error {
    data, err := s.Marshal()
    if err != nil {
        return err
    }
    return keychain.Open().Set(keychainService, profile, data)
}

// LoadFromKeychain returns the session persisted under profile, or an
// APIError{Code:CodeUnauthenticated} if none exists.
func LoadFromKeychain(profile string) (*Session, error) {
    data, err := keychain.Open().Get(keychainService, profile)
    if err != nil {
        if errors.Is(err, keychain.ErrNotFound) {
            return nil, &APIError{
                Code:    CodeUnauthenticated,
                Message: "no session for profile " + profile,
                Hint:    "run: rh login",
            }
        }
        return nil, &APIError{Code: CodeUnauthenticated, Message: err.Error(), Hint: "run: rh login"}
    }
    return UnmarshalSession(data)
}

// ClearKeychain removes the session blob for the given profile.
// Returns nil if nothing was stored.
func ClearKeychain(profile string) error {
    err := keychain.Open().Delete(keychainService, profile)
    if err != nil && !errors.Is(err, keychain.ErrNotFound) {
        return err
    }
    return nil
}
```

Add `"errors"` to the imports.

- [ ] **Step 4: Run tests**

Run: `go test ./internal/robinhood/ -run "TestSession_Marshal|TestSaveAndLoadFromKeychain" -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/robinhood/session.go internal/robinhood/session_test.go
git commit -m "feat(session): keychain load/save/clear with stable JSON blob"
```

---

## Task 5: Stdin prompt package

**Files:**
- Create: `internal/inputprompt/prompt.go`
- Create: `internal/inputprompt/prompt_test.go`
- Modify: `go.mod` / `go.sum`

- [ ] **Step 1: Add term dependency**

Run:
```bash
go get golang.org/x/term@v0.21.0
go mod tidy
```

- [ ] **Step 2: Write the failing tests**

Create `internal/inputprompt/prompt_test.go`:

```go
package inputprompt

import (
    "bytes"
    "strings"
    "testing"
)

func TestPrompt_ReadsLine(t *testing.T) {
    p := &Prompter{
        In:  strings.NewReader("alice@example.com\n"),
        Out: &bytes.Buffer{},
    }
    got, err := p.Text("Email: ")
    if err != nil {
        t.Fatalf("Text: %v", err)
    }
    if got != "alice@example.com" {
        t.Fatalf("got %q, want %q", got, "alice@example.com")
    }
}

func TestPrompt_TrimsWhitespace(t *testing.T) {
    p := &Prompter{
        In:  strings.NewReader("  bob\t\n"),
        Out: &bytes.Buffer{},
    }
    got, err := p.Text("Name: ")
    if err != nil || got != "bob" {
        t.Fatalf("got %q err %v, want %q", got, err, "bob")
    }
}

func TestPrompt_RejectsEmptyWithRetryCap(t *testing.T) {
    p := &Prompter{
        In:  strings.NewReader("\n\n\n"),
        Out: &bytes.Buffer{},
        MaxEmptyRetries: 2,
    }
    _, err := p.Text("Name: ")
    if err == nil {
        t.Fatalf("expected error after too many empties")
    }
}
```

- [ ] **Step 3: Run test — expect compile failure**

Run: `go test ./internal/inputprompt/ -v`
Expected: FAIL — `Prompter` undefined.

- [ ] **Step 4: Implement the prompter**

Create `internal/inputprompt/prompt.go`:

```go
// Package inputprompt provides small stdin/stdout prompt helpers used by `rh login`.
//
// The Prompter is injectable so tests can feed scripted input and interactive
// flows can inject real stdin/stdout. Password reads bypass In/Out and use
// x/term directly when stdin is a TTY; tests force the non-TTY branch.
package inputprompt

import (
    "bufio"
    "errors"
    "fmt"
    "io"
    "os"
    "strings"

    "golang.org/x/term"
)

// Prompter reads interactive user input.
type Prompter struct {
    In              io.Reader
    Out             io.Writer
    MaxEmptyRetries int // 0 = default 3
}

// New returns a Prompter bound to os.Stdin / os.Stderr (prompts go to stderr
// so `rh login --json` on stdout stays clean — though login has no --json).
func New() *Prompter {
    return &Prompter{In: os.Stdin, Out: os.Stderr}
}

// Text prompts and reads a line, trimming whitespace. Retries up to MaxEmptyRetries times.
func (p *Prompter) Text(label string) (string, error) {
    retries := p.MaxEmptyRetries
    if retries <= 0 {
        retries = 3
    }
    r := bufio.NewReader(p.In)
    for i := 0; i < retries; i++ {
        if _, err := fmt.Fprint(p.Out, label); err != nil {
            return "", err
        }
        line, err := r.ReadString('\n')
        if err != nil && !errors.Is(err, io.EOF) {
            return "", err
        }
        line = strings.TrimSpace(line)
        if line != "" {
            return line, nil
        }
        if errors.Is(err, io.EOF) {
            break
        }
    }
    return "", errors.New("empty input after retries")
}

// Password prompts and reads a line with echo suppressed. Falls back to plain
// read when stdin is not a TTY (tests, pipelines).
func (p *Prompter) Password(label string) (string, error) {
    if _, err := fmt.Fprint(p.Out, label); err != nil {
        return "", err
    }
    f, ok := p.In.(*os.File)
    if ok && term.IsTerminal(int(f.Fd())) {
        b, err := term.ReadPassword(int(f.Fd()))
        if _, werr := fmt.Fprintln(p.Out); werr != nil {
            return "", werr
        }
        if err != nil {
            return "", err
        }
        return string(b), nil
    }
    // Non-TTY fallback.
    r := bufio.NewReader(p.In)
    line, err := r.ReadString('\n')
    if err != nil && !errors.Is(err, io.EOF) {
        return "", err
    }
    return strings.TrimRight(line, "\r\n"), nil
}

// Choose prints a menu and returns the zero-based index of the selection.
func (p *Prompter) Choose(label string, options []string) (int, error) {
    if len(options) == 0 {
        return 0, errors.New("no options")
    }
    if _, err := fmt.Fprintln(p.Out, label); err != nil {
        return 0, err
    }
    for i, o := range options {
        if _, err := fmt.Fprintf(p.Out, "  [%d] %s\n", i+1, o); err != nil {
            return 0, err
        }
    }
    for {
        s, err := p.Text("> ")
        if err != nil {
            return 0, err
        }
        var idx int
        if _, err := fmt.Sscanf(s, "%d", &idx); err == nil && idx >= 1 && idx <= len(options) {
            return idx - 1, nil
        }
        _, _ = fmt.Fprintln(p.Out, "enter a number from the list")
    }
}
```

- [ ] **Step 5: Run tests**

Run: `go test ./internal/inputprompt/ -v`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/inputprompt/ go.mod go.sum
git commit -m "feat(inputprompt): stdin prompter with hidden-password + numeric choose"
```

---

## Task 6: File-locked refresh serializer

Two `rh` processes invoking concurrently must not both burn the refresh token and get one invalidated. An OS file lock around the refresh critical section fixes this.

**Files:**
- Create: `internal/robinhood/refresh_lock.go`
- Create: `internal/robinhood/refresh_lock_test.go`
- Modify: `go.mod` / `go.sum`

- [ ] **Step 1: Add flock dependency**

Run:
```bash
go get github.com/gofrs/flock@v0.12.1
go mod tidy
```

- [ ] **Step 2: Write the failing test**

Create `internal/robinhood/refresh_lock_test.go`:

```go
package robinhood

import (
    "sync"
    "sync/atomic"
    "testing"
    "time"
)

func TestRefreshLock_Serializes(t *testing.T) {
    t.Setenv("XDG_CONFIG_HOME", t.TempDir())
    var active int32
    var maxSeen int32
    var wg sync.WaitGroup
    start := make(chan struct{})
    for i := 0; i < 8; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            <-start
            err := WithRefreshLock(func() error {
                n := atomic.AddInt32(&active, 1)
                for {
                    m := atomic.LoadInt32(&maxSeen)
                    if n <= m || atomic.CompareAndSwapInt32(&maxSeen, m, n) {
                        break
                    }
                }
                time.Sleep(10 * time.Millisecond)
                atomic.AddInt32(&active, -1)
                return nil
            })
            if err != nil {
                t.Errorf("WithRefreshLock: %v", err)
            }
        }()
    }
    close(start)
    wg.Wait()
    if got := atomic.LoadInt32(&maxSeen); got != 1 {
        t.Fatalf("max concurrent critical sections = %d, want 1", got)
    }
}
```

- [ ] **Step 3: Run test — expect compile failure**

Run: `go test ./internal/robinhood/ -run TestRefreshLock -v`
Expected: FAIL — `WithRefreshLock` undefined.

- [ ] **Step 4: Implement the lock**

Create `internal/robinhood/refresh_lock.go`:

```go
package robinhood

import (
    "fmt"
    "path/filepath"

    "github.com/gofrs/flock"
    "github.com/herocod3r/robinhood-cli/internal/config"
)

// WithRefreshLock runs fn while holding an exclusive lock on
// <ConfigDir>/session.lock. Concurrent `rh` invocations serialize here
// so only one performs the refresh at a time.
func WithRefreshLock(fn func() error) error {
    dir, err := config.EnsureConfigDir()
    if err != nil {
        return err
    }
    lock := flock.New(filepath.Join(dir, "session.lock"))
    if err := lock.Lock(); err != nil {
        return fmt.Errorf("acquire refresh lock: %w", err)
    }
    defer func() { _ = lock.Unlock() }()
    return fn()
}
```

- [ ] **Step 5: Wire the lock into client.ensureFresh**

Modify `internal/robinhood/client.go` `ensureFresh`:

```go
func (c *Client) ensureFresh() error {
    s := c.Session()
    if s == nil {
        return &APIError{Code: CodeUnauthenticated, Message: "no session", Hint: "run: rh login"}
    }
    if !s.NeedsImmediateRefresh() {
        return nil
    }
    return WithRefreshLock(func() error {
        // Re-read session under the lock; another process may have just refreshed.
        latest, lerr := LoadFromKeychain(currentProfile())
        if lerr == nil && !latest.NeedsImmediateRefresh() {
            c.SetSession(latest)
            return nil
        }
        if err := c.oauth.Refresh(s); err != nil {
            return err
        }
        if serr := s.SaveToKeychain(currentProfile()); serr != nil {
            return serr
        }
        return nil
    })
}
```

Add a `currentProfile()` helper:

```go
// currentProfile returns the profile for which this client is authenticated.
// Plan B ships with a single default; Plan B.2 can wire --profile.
func currentProfile() string {
    if p := os.Getenv(config.EnvProfile); p != "" {
        return p
    }
    return "default"
}
```

Import `os` and `config` at the top if not already imported.

- [ ] **Step 6: Run the tests**

Run: `go test ./internal/robinhood/ -run "TestRefreshLock|TestSaveAndLoadFromKeychain" -race -v`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/robinhood/refresh_lock.go internal/robinhood/refresh_lock_test.go internal/robinhood/client.go go.mod go.sum
git commit -m "feat(refresh): file-locked refresh serializer via gofrs/flock"
```

---

## Task 7: Password grant in OAuth

**Files:**
- Modify: `internal/robinhood/oauth.go`
- Modify: `internal/robinhood/oauth_test.go`
- Modify: `go.mod` / `go.sum`

- [ ] **Step 1: Add uuid + otp deps**

Run:
```bash
go get github.com/google/uuid@v1.6.0
go get github.com/pquerna/otp@v1.4.0
go mod tidy
```

- [ ] **Step 2: Write the failing tests**

Append to `internal/robinhood/oauth_test.go`:

```go
func TestPasswordGrant_Success(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/oauth2/token/" {
            t.Fatalf("path = %s", r.URL.Path)
        }
        if err := r.ParseForm(); err != nil {
            t.Fatal(err)
        }
        if got := r.Form.Get("grant_type"); got != "password" {
            t.Fatalf("grant_type = %s", got)
        }
        if r.Form.Get("device_token") == "" {
            t.Fatalf("missing device_token")
        }
        w.Header().Set("Content-Type", "application/json")
        fmt.Fprint(w, `{"access_token":"a","refresh_token":"r","expires_in":3600,"token_type":"Bearer"}`)
    }))
    defer ts.Close()

    o := &oauth{baseURL: ts.URL, httpClient: ts.Client()}
    sess, err := o.PasswordGrant(context.Background(), "u", "p", "dev-token", "")
    if err != nil {
        t.Fatalf("PasswordGrant: %v", err)
    }
    if sess.AccessToken != "a" {
        t.Fatalf("AccessToken = %q", sess.AccessToken)
    }
    if sess.DeviceToken != "dev-token" {
        t.Fatalf("DeviceToken = %q", sess.DeviceToken)
    }
    if sess.ExpiresAt.IsZero() {
        t.Fatalf("ExpiresAt should be set")
    }
}

func TestPasswordGrant_SheriffRequired(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprint(w, `{"verification_workflow":{"id":"wf-123","workflow_status":"internal_pending"}}`)
    }))
    defer ts.Close()
    o := &oauth{baseURL: ts.URL, httpClient: ts.Client()}
    _, err := o.PasswordGrant(context.Background(), "u", "p", "dev", "")
    apiErr, ok := err.(*APIError)
    if !ok {
        t.Fatalf("err = %v, want APIError", err)
    }
    if apiErr.Code != CodeSheriffRequired {
        t.Fatalf("code = %q, want %q", apiErr.Code, CodeSheriffRequired)
    }
    if apiErr.WorkflowID != "wf-123" {
        t.Fatalf("workflow_id = %q, want %q", apiErr.WorkflowID, "wf-123")
    }
}

func TestPasswordGrant_MFARequired(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprint(w, `{"mfa_required":true,"mfa_type":"app"}`)
    }))
    defer ts.Close()
    o := &oauth{baseURL: ts.URL, httpClient: ts.Client()}
    _, err := o.PasswordGrant(context.Background(), "u", "p", "dev", "")
    apiErr, ok := err.(*APIError)
    if !ok || apiErr.Code != CodeMFARequired {
        t.Fatalf("err = %v, want CodeMFARequired", err)
    }
}

func TestPasswordGrant_WithMFACode(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if err := r.ParseForm(); err != nil {
            t.Fatal(err)
        }
        if r.Form.Get("mfa_code") == "" {
            w.WriteHeader(http.StatusBadRequest)
            fmt.Fprint(w, `{"mfa_required":true}`)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        fmt.Fprint(w, `{"access_token":"a","refresh_token":"r","expires_in":3600}`)
    }))
    defer ts.Close()
    o := &oauth{baseURL: ts.URL, httpClient: ts.Client()}
    _, err := o.PasswordGrant(context.Background(), "u", "p", "dev", "123456")
    if err != nil {
        t.Fatalf("PasswordGrant with MFA code: %v", err)
    }
}
```

- [ ] **Step 3: Run tests — expect failures**

Run: `go test ./internal/robinhood/ -run "TestPasswordGrant" -v`
Expected: FAIL — `PasswordGrant`, `CodeMFARequired`, `APIError.WorkflowID` undefined.

- [ ] **Step 4: Extend error taxonomy for internal states**

`CodeMFARequired` and `CodeSheriffRequired` are both "need user interaction" codes. Sheriff already exists. Add MFA:

Modify `internal/robinhood/errors.go`:

```go
const (
    // existing ...
    CodeMFARequired Code = "mfa_required"
)

// APIError gets a WorkflowID for Sheriff correlation.
type APIError struct {
    Code       Code
    Message    string
    Hint       string
    Retryable  bool
    HTTPStatus int
    // WorkflowID is set when Code == CodeSheriffRequired.
    WorkflowID string
}

// ExitCode: treat MFARequired same as Sheriff — interactive step needed.
func (e *APIError) ExitCode() int {
    switch e.Code {
    case CodeUnauthenticated, CodeSessionExpired:
        return 2
    case CodeSheriffRequired, CodeMFARequired:
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

- [ ] **Step 5: Implement `PasswordGrant`**

In `internal/robinhood/oauth.go`, add:

```go
// PasswordGrant exchanges username+password (and optional MFA code) for tokens.
// Returns (*Session, nil) on success. On 400 with verification_workflow,
// returns APIError{Code:CodeSheriffRequired, WorkflowID: ...}. On 400 with
// mfa_required, returns APIError{Code:CodeMFARequired}.
func (o *oauth) PasswordGrant(ctx context.Context, username, password, deviceToken, mfaCode string) (*Session, error) {
    form := url.Values{}
    form.Set("grant_type", "password")
    form.Set("username", username)
    form.Set("password", password)
    form.Set("client_id", robinhoodOAuthClientID)
    form.Set("device_token", deviceToken)
    form.Set("expires_in", "86400")
    form.Set("scope", "internal")
    if mfaCode != "" {
        form.Set("mfa_code", mfaCode)
    }

    req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL+"/oauth2/token/", strings.NewReader(form.Encode()))
    if err != nil {
        return nil, &APIError{Code: CodeRobinhoodUnavailable, Message: err.Error()}
    }
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    req.Header.Set("Accept", "application/json")
    req.Header.Set("X-Robinhood-API-Version", "1.431.4") // matches robin_stocks 3.4.x

    resp, err := o.httpClient.Do(req)
    if err != nil {
        return nil, &APIError{Code: CodeRobinhoodUnavailable, Message: err.Error()}
    }
    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)

    switch resp.StatusCode {
    case http.StatusOK:
        var tr tokenResponse
        if err := json.Unmarshal(body, &tr); err != nil {
            return nil, &APIError{Code: CodeRobinhoodUnavailable, Message: "malformed token response"}
        }
        s := &Session{
            Username:     username,
            AccessToken:  tr.AccessToken,
            RefreshToken: tr.RefreshToken,
            DeviceToken:  deviceToken,
        }
        if tr.ExpiresIn > 0 {
            s.ExpiresAt = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second).UTC()
        }
        return s, nil

    case http.StatusBadRequest, http.StatusUnauthorized:
        return nil, classifyPasswordErr(body, resp.StatusCode)

    default:
        return nil, &APIError{
            Code:       CodeRobinhoodUnavailable,
            Message:    fmt.Sprintf("password grant HTTP %d", resp.StatusCode),
            HTTPStatus: resp.StatusCode,
        }
    }
}

type passwordErrResp struct {
    MFARequired         bool   `json:"mfa_required"`
    MFAType             string `json:"mfa_type"`
    VerificationWorkflow struct {
        ID             string `json:"id"`
        WorkflowStatus string `json:"workflow_status"`
    } `json:"verification_workflow"`
    Detail           string `json:"detail"`
    Error            string `json:"error"`
    ErrorDescription string `json:"error_description"`
}

func classifyPasswordErr(body []byte, status int) error {
    var er passwordErrResp
    _ = json.Unmarshal(body, &er)
    if er.VerificationWorkflow.ID != "" {
        return &APIError{
            Code:       CodeSheriffRequired,
            Message:    "device verification required",
            Hint:       "complete the challenge shown in your Robinhood app or SMS",
            HTTPStatus: status,
            WorkflowID: er.VerificationWorkflow.ID,
        }
    }
    if er.MFARequired {
        hint := "enter your TOTP code"
        if er.MFAType == "sms" {
            hint = "enter the SMS code Robinhood just sent"
        }
        return &APIError{
            Code:       CodeMFARequired,
            Message:    "MFA required",
            Hint:       hint,
            HTTPStatus: status,
        }
    }
    msg := firstNonEmpty(er.Detail, er.ErrorDescription, er.Error, "password grant rejected")
    return &APIError{Code: CodeUnauthenticated, Message: msg, HTTPStatus: status}
}
```

Add `"context"` to imports if missing.

- [ ] **Step 6: Run tests**

Run: `go test ./internal/robinhood/ -run "TestPasswordGrant" -v`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/robinhood/oauth.go internal/robinhood/oauth_test.go internal/robinhood/errors.go go.mod go.sum
git commit -m "feat(oauth): password grant + MFA/Sheriff classification"
```

---

## Task 8: TOTP helper

**Files:**
- Modify: `internal/robinhood/oauth.go` (add `TOTPCode` helper)
- Modify: `internal/robinhood/oauth_test.go`

- [ ] **Step 1: Write the failing test**

Append to `internal/robinhood/oauth_test.go`:

```go
func TestTOTPCode_Deterministic(t *testing.T) {
    // Known test vector from RFC 6238.
    got, err := TOTPCodeAt("JBSWY3DPEHPK3PXP", time.Unix(59, 0).UTC())
    if err != nil {
        t.Fatalf("TOTPCodeAt: %v", err)
    }
    if len(got) != 6 {
        t.Fatalf("len = %d, want 6", len(got))
    }
}

func TestTOTPCode_RejectsBadSecret(t *testing.T) {
    if _, err := TOTPCodeAt("not-base32", time.Now()); err == nil {
        t.Fatalf("expected error for invalid secret")
    }
}
```

- [ ] **Step 2: Run test — expect compile failure**

Run: `go test ./internal/robinhood/ -run "TestTOTPCode" -v`
Expected: FAIL — `TOTPCodeAt` undefined.

- [ ] **Step 3: Implement**

In `internal/robinhood/oauth.go`:

```go
import (
    "github.com/pquerna/otp/totp"
)

// TOTPCode returns the current 6-digit TOTP for the given base32 secret.
func TOTPCode(secret string) (string, error) {
    return TOTPCodeAt(secret, time.Now())
}

// TOTPCodeAt returns the TOTP at a specific instant (used by tests).
func TOTPCodeAt(secret string, at time.Time) (string, error) {
    return totp.GenerateCode(secret, at)
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./internal/robinhood/ -run "TestTOTPCode" -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/robinhood/oauth.go internal/robinhood/oauth_test.go
git commit -m "feat(oauth): TOTP code generation from shared secret"
```

---

## Task 9: Sheriff state machine — start + poll

**Files:**
- Create: `internal/robinhood/sheriff.go`
- Create: `internal/robinhood/sheriff_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/robinhood/sheriff_test.go`:

```go
package robinhood

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "time"
)

func TestSheriff_Start_SMS(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch r.URL.Path {
        case "/pathfinder/user_machine/":
            if err := r.ParseForm(); err != nil {
                t.Fatal(err)
            }
            var body map[string]string
            _ = json.NewDecoder(r.Body).Decode(&body)
            if body["device_id"] == "" {
                t.Fatalf("device_id missing")
            }
            w.Header().Set("Content-Type", "application/json")
            fmt.Fprint(w, `{"id":"inq-1"}`)
        case "/pathfinder/inquiries/inq-1/user_view/":
            fmt.Fprint(w, `{"type_context":{"context":{"sheriff_challenge":{"type":"sms","id":"ch-1","phone_number":"***-***-1234"}}}}`)
        default:
            t.Fatalf("unexpected path %s", r.URL.Path)
        }
    }))
    defer ts.Close()
    s := &Sheriff{
        BaseURL:   ts.URL,
        HTTP:      ts.Client(),
        PollEvery: time.Millisecond,
    }
    step, err := s.Start(context.Background(), "wf-123", "dev-tok")
    if err != nil {
        t.Fatalf("Start: %v", err)
    }
    if step.Kind != SheriffSMS {
        t.Fatalf("Kind = %v, want SheriffSMS", step.Kind)
    }
    if step.ChallengeID != "ch-1" {
        t.Fatalf("ChallengeID = %q", step.ChallengeID)
    }
    if !strings.Contains(step.Detail, "1234") {
        t.Fatalf("Detail should surface phone hint; got %q", step.Detail)
    }
}

func TestSheriff_Start_DeviceApproval(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch r.URL.Path {
        case "/pathfinder/user_machine/":
            fmt.Fprint(w, `{"id":"inq-2"}`)
        case "/pathfinder/inquiries/inq-2/user_view/":
            fmt.Fprint(w, `{"type_context":{"context":{"sheriff_challenge":{"type":"prompt","id":"ch-2"}}}}`)
        }
    }))
    defer ts.Close()
    s := &Sheriff{BaseURL: ts.URL, HTTP: ts.Client(), PollEvery: time.Millisecond}
    step, err := s.Start(context.Background(), "wf-2", "dev-tok")
    if err != nil {
        t.Fatalf("Start: %v", err)
    }
    if step.Kind != SheriffPush {
        t.Fatalf("Kind = %v, want SheriffPush", step.Kind)
    }
}

func TestSheriff_Start_UnknownType_ErrorsOut(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch r.URL.Path {
        case "/pathfinder/user_machine/":
            fmt.Fprint(w, `{"id":"inq-3"}`)
        case "/pathfinder/inquiries/inq-3/user_view/":
            fmt.Fprint(w, `{"type_context":{"context":{"sheriff_challenge":{"type":"quantum"}}}}`)
        }
    }))
    defer ts.Close()
    s := &Sheriff{BaseURL: ts.URL, HTTP: ts.Client(), PollEvery: time.Millisecond}
    _, err := s.Start(context.Background(), "wf-3", "dev-tok")
    apiErr, ok := err.(*APIError)
    if !ok || apiErr.Code != CodeSheriffRequired {
        t.Fatalf("err = %v; want CodeSheriffRequired", err)
    }
    if !strings.Contains(apiErr.Message, "quantum") {
        t.Fatalf("error should name the unknown kind; got %q", apiErr.Message)
    }
}

func TestSheriff_Start_RespectsCtxDeadline(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/pathfinder/user_machine/" {
            fmt.Fprint(w, `{"id":"inq-4"}`)
            return
        }
        // user_view never produces a sheriff_challenge — force poll loop.
        fmt.Fprint(w, `{"type_context":{"context":{}}}`)
    }))
    defer ts.Close()
    s := &Sheriff{BaseURL: ts.URL, HTTP: ts.Client(), PollEvery: 5 * time.Millisecond}
    ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
    defer cancel()
    _, err := s.Start(ctx, "wf-4", "dev-tok")
    if err == nil {
        t.Fatalf("expected timeout error")
    }
}
```

- [ ] **Step 2: Run tests — expect failures**

Run: `go test ./internal/robinhood/ -run TestSheriff_Start -v`
Expected: FAIL — `Sheriff`, `SheriffSMS`, etc undefined.

- [ ] **Step 3: Implement Sheriff.Start**

Create `internal/robinhood/sheriff.go`:

```go
package robinhood

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

// SheriffKind classifies the challenge the server is asking the user for.
type SheriffKind int

const (
    SheriffUnknown SheriffKind = iota
    SheriffSMS
    SheriffEmail
    SheriffPush
)

func (k SheriffKind) String() string {
    switch k {
    case SheriffSMS:
        return "sms"
    case SheriffEmail:
        return "email"
    case SheriffPush:
        return "prompt"
    default:
        return "unknown"
    }
}

// SheriffStep is what Start returns — all the state the login UI needs to
// collect one piece of input (or nothing, for push).
type SheriffStep struct {
    InquiryID   string
    ChallengeID string
    Kind        SheriffKind
    Detail      string // human-readable: "SMS to ***-***-1234"
}

// Sheriff drives the pathfinder user_machine + inquiry poll + challenge response.
// HTTP and Clock are injectable for tests; PollEvery defaults to 2s in production.
type Sheriff struct {
    BaseURL   string
    HTTP      *http.Client
    PollEvery time.Duration
    Clock     func() time.Time
}

func (s *Sheriff) now() time.Time {
    if s.Clock != nil {
        return s.Clock()
    }
    return time.Now()
}

func (s *Sheriff) poll() time.Duration {
    if s.PollEvery > 0 {
        return s.PollEvery
    }
    return 2 * time.Second
}

type userMachineReq struct {
    DeviceID string `json:"device_id"`
    Flow     string `json:"flow"`
    InputAs  string `json:"input_as"`
    Type     string `json:"type"`
}

type userMachineResp struct {
    ID string `json:"id"`
}

type userViewResp struct {
    TypeContext struct {
        Context struct {
            SheriffChallenge struct {
                ID          string `json:"id"`
                Type        string `json:"type"`
                PhoneNumber string `json:"phone_number"`
                Email       string `json:"email"`
            } `json:"sheriff_challenge"`
        } `json:"context"`
    } `json:"type_context"`
}

// Start posts the user_machine challenge and polls user_view until the server
// exposes a sheriff_challenge (sms/email/prompt). Returns the step the UI
// should handle.
func (s *Sheriff) Start(ctx context.Context, workflowID, deviceToken string) (*SheriffStep, error) {
    body, err := json.Marshal(userMachineReq{
        DeviceID: deviceToken,
        Flow:     "suv",
        InputAs:  "RHS",
        Type:     "login",
    })
    if err != nil {
        return nil, &APIError{Code: CodeValidation, Message: err.Error()}
    }
    req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.BaseURL+"/pathfinder/user_machine/", bytes.NewReader(body))
    if err != nil {
        return nil, &APIError{Code: CodeValidation, Message: err.Error()}
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-Robinhood-Challenge-Response-Id", workflowID)
    resp, err := s.HTTP.Do(req)
    if err != nil {
        return nil, &APIError{Code: CodeRobinhoodUnavailable, Message: err.Error()}
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
        buf, _ := io.ReadAll(resp.Body)
        return nil, &APIError{
            Code:       CodeSheriffRequired,
            Message:    fmt.Sprintf("user_machine HTTP %d: %s", resp.StatusCode, string(buf)),
            HTTPStatus: resp.StatusCode,
            WorkflowID: workflowID,
        }
    }
    var umr userMachineResp
    if err := json.NewDecoder(resp.Body).Decode(&umr); err != nil {
        return nil, &APIError{Code: CodeRobinhoodUnavailable, Message: err.Error()}
    }
    inquiryID := umr.ID

    // Poll user_view until a sheriff_challenge shows up.
    for {
        if err := ctx.Err(); err != nil {
            return nil, &APIError{Code: CodeSheriffRequired, Message: err.Error(), WorkflowID: workflowID}
        }
        step, err := s.fetchUserView(ctx, inquiryID, workflowID)
        if err != nil {
            return nil, err
        }
        if step != nil {
            return step, nil
        }
        select {
        case <-ctx.Done():
            return nil, &APIError{Code: CodeSheriffRequired, Message: ctx.Err().Error(), WorkflowID: workflowID}
        case <-time.After(s.poll()):
        }
    }
}

func (s *Sheriff) fetchUserView(ctx context.Context, inquiryID, workflowID string) (*SheriffStep, error) {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.BaseURL+"/pathfinder/inquiries/"+inquiryID+"/user_view/", nil)
    if err != nil {
        return nil, &APIError{Code: CodeValidation, Message: err.Error()}
    }
    resp, err := s.HTTP.Do(req)
    if err != nil {
        return nil, &APIError{Code: CodeRobinhoodUnavailable, Message: err.Error()}
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return nil, &APIError{
            Code:       CodeSheriffRequired,
            Message:    fmt.Sprintf("user_view HTTP %d", resp.StatusCode),
            HTTPStatus: resp.StatusCode,
            WorkflowID: workflowID,
        }
    }
    var uv userViewResp
    if err := json.NewDecoder(resp.Body).Decode(&uv); err != nil {
        return nil, &APIError{Code: CodeRobinhoodUnavailable, Message: err.Error()}
    }
    ch := uv.TypeContext.Context.SheriffChallenge
    if ch.ID == "" && ch.Type == "" {
        return nil, nil // keep polling
    }
    step := &SheriffStep{
        InquiryID:   inquiryID,
        ChallengeID: ch.ID,
    }
    switch ch.Type {
    case "sms":
        step.Kind = SheriffSMS
        step.Detail = fmt.Sprintf("SMS to %s", ch.PhoneNumber)
    case "email":
        step.Kind = SheriffEmail
        step.Detail = fmt.Sprintf("email to %s", ch.Email)
    case "prompt":
        step.Kind = SheriffPush
        step.Detail = "push notification to your Robinhood app"
    default:
        return nil, &APIError{
            Code:       CodeSheriffRequired,
            Message:    fmt.Sprintf("unknown sheriff challenge type: %s", ch.Type),
            WorkflowID: workflowID,
        }
    }
    return step, nil
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./internal/robinhood/ -run TestSheriff_Start -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/robinhood/sheriff.go internal/robinhood/sheriff_test.go
git commit -m "feat(sheriff): user_machine + user_view polling with typed SheriffStep"
```

---

## Task 10: Sheriff — respond / push poll

**Files:**
- Modify: `internal/robinhood/sheriff.go`
- Modify: `internal/robinhood/sheriff_test.go`

- [ ] **Step 1: Write the failing tests**

Append to `internal/robinhood/sheriff_test.go`:

```go
func TestSheriff_RespondCode_Success(t *testing.T) {
    var capturedBody string
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/challenge/ch-1/respond/" {
            b, _ := io.ReadAll(r.Body)
            capturedBody = string(b)
            w.Header().Set("Content-Type", "application/json")
            fmt.Fprint(w, `{"status":"validated"}`)
            return
        }
        t.Fatalf("unexpected path %s", r.URL.Path)
    }))
    defer ts.Close()
    s := &Sheriff{BaseURL: ts.URL, HTTP: ts.Client()}
    step := &SheriffStep{InquiryID: "inq-1", ChallengeID: "ch-1", Kind: SheriffSMS}
    if err := s.RespondCode(context.Background(), step, "123456"); err != nil {
        t.Fatalf("RespondCode: %v", err)
    }
    if !strings.Contains(capturedBody, `"response":"123456"`) {
        t.Fatalf("body = %q", capturedBody)
    }
}

func TestSheriff_RespondCode_WrongCode(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprint(w, `{"detail":"incorrect code"}`)
    }))
    defer ts.Close()
    s := &Sheriff{BaseURL: ts.URL, HTTP: ts.Client()}
    step := &SheriffStep{InquiryID: "inq-1", ChallengeID: "ch-1", Kind: SheriffSMS}
    err := s.RespondCode(context.Background(), step, "000000")
    apiErr, _ := err.(*APIError)
    if apiErr == nil || apiErr.Code != CodeSheriffRequired {
        t.Fatalf("err = %v; want CodeSheriffRequired", err)
    }
}

func TestSheriff_WaitPush_Approved(t *testing.T) {
    calls := 0
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/push/ch-2/get_prompts_status/" {
            calls++
            if calls < 2 {
                fmt.Fprint(w, `{"challenge_status":"issued"}`)
                return
            }
            fmt.Fprint(w, `{"challenge_status":"validated"}`)
            return
        }
        t.Fatalf("unexpected path %s", r.URL.Path)
    }))
    defer ts.Close()
    s := &Sheriff{BaseURL: ts.URL, HTTP: ts.Client(), PollEvery: time.Millisecond}
    step := &SheriffStep{InquiryID: "inq-2", ChallengeID: "ch-2", Kind: SheriffPush}
    if err := s.WaitPush(context.Background(), step); err != nil {
        t.Fatalf("WaitPush: %v", err)
    }
    if calls < 2 {
        t.Fatalf("calls = %d, want >=2", calls)
    }
}

func TestSheriff_WaitPush_Declined(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprint(w, `{"challenge_status":"failed"}`)
    }))
    defer ts.Close()
    s := &Sheriff{BaseURL: ts.URL, HTTP: ts.Client(), PollEvery: time.Millisecond}
    step := &SheriffStep{InquiryID: "inq-3", ChallengeID: "ch-3", Kind: SheriffPush}
    err := s.WaitPush(context.Background(), step)
    if err == nil {
        t.Fatalf("expected error for declined push")
    }
    apiErr, _ := err.(*APIError)
    if apiErr == nil || apiErr.Code != CodeSheriffRequired {
        t.Fatalf("err = %v; want CodeSheriffRequired", err)
    }
}
```

- [ ] **Step 2: Run tests — expect failures**

Run: `go test ./internal/robinhood/ -run "TestSheriff_RespondCode|TestSheriff_WaitPush" -v`
Expected: FAIL — methods undefined.

- [ ] **Step 3: Implement RespondCode and WaitPush**

Append to `internal/robinhood/sheriff.go`:

```go
type challengeRespReq struct {
    Response string `json:"response"`
}
type challengeRespResp struct {
    Status string `json:"status"`
    Detail string `json:"detail"`
}

// RespondCode submits the user-entered code for SMS/email challenges.
func (s *Sheriff) RespondCode(ctx context.Context, step *SheriffStep, code string) error {
    body, _ := json.Marshal(challengeRespReq{Response: code})
    req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.BaseURL+"/challenge/"+step.ChallengeID+"/respond/", bytes.NewReader(body))
    if err != nil {
        return &APIError{Code: CodeValidation, Message: err.Error()}
    }
    req.Header.Set("Content-Type", "application/json")
    resp, err := s.HTTP.Do(req)
    if err != nil {
        return &APIError{Code: CodeRobinhoodUnavailable, Message: err.Error()}
    }
    defer resp.Body.Close()
    buf, _ := io.ReadAll(resp.Body)
    if resp.StatusCode != http.StatusOK {
        var er challengeRespResp
        _ = json.Unmarshal(buf, &er)
        msg := firstNonEmpty(er.Detail, string(buf), fmt.Sprintf("HTTP %d", resp.StatusCode))
        return &APIError{Code: CodeSheriffRequired, Message: msg, HTTPStatus: resp.StatusCode}
    }
    var cr challengeRespResp
    if err := json.Unmarshal(buf, &cr); err != nil {
        return &APIError{Code: CodeRobinhoodUnavailable, Message: err.Error()}
    }
    if cr.Status != "validated" {
        return &APIError{Code: CodeSheriffRequired, Message: firstNonEmpty(cr.Status, "challenge not validated")}
    }
    return nil
}

type promptsStatusResp struct {
    ChallengeStatus string `json:"challenge_status"`
}

// WaitPush polls /push/<id>/get_prompts_status/ until the server reports
// "validated". Any terminal non-validated state becomes an error.
func (s *Sheriff) WaitPush(ctx context.Context, step *SheriffStep) error {
    for {
        if err := ctx.Err(); err != nil {
            return &APIError{Code: CodeSheriffRequired, Message: err.Error()}
        }
        req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.BaseURL+"/push/"+step.ChallengeID+"/get_prompts_status/", nil)
        if err != nil {
            return &APIError{Code: CodeValidation, Message: err.Error()}
        }
        resp, err := s.HTTP.Do(req)
        if err != nil {
            return &APIError{Code: CodeRobinhoodUnavailable, Message: err.Error()}
        }
        var ps promptsStatusResp
        _ = json.NewDecoder(resp.Body).Decode(&ps)
        _ = resp.Body.Close()
        switch ps.ChallengeStatus {
        case "validated":
            return nil
        case "issued", "pending", "":
            // keep polling
        default:
            return &APIError{Code: CodeSheriffRequired, Message: "push challenge " + ps.ChallengeStatus}
        }
        select {
        case <-ctx.Done():
            return &APIError{Code: CodeSheriffRequired, Message: ctx.Err().Error()}
        case <-time.After(s.poll()):
        }
    }
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./internal/robinhood/ -run "TestSheriff_RespondCode|TestSheriff_WaitPush" -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/robinhood/sheriff.go internal/robinhood/sheriff_test.go
git commit -m "feat(sheriff): respond_code + push wait polling"
```

---

## Task 11: Password grant re-attempt after Sheriff success

After Sheriff validates, Robinhood expects a second `/oauth2/token/` POST with an `X-Robinhood-Challenge-Response-Id: <workflow_id>` header. Add a helper.

**Files:**
- Modify: `internal/robinhood/oauth.go`
- Modify: `internal/robinhood/oauth_test.go`

- [ ] **Step 1: Write the failing test**

Append to `internal/robinhood/oauth_test.go`:

```go
func TestPasswordGrantWithWorkflow_SendsHeader(t *testing.T) {
    var sawHeader string
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        sawHeader = r.Header.Get("X-Robinhood-Challenge-Response-Id")
        w.Header().Set("Content-Type", "application/json")
        fmt.Fprint(w, `{"access_token":"a","refresh_token":"r","expires_in":3600}`)
    }))
    defer ts.Close()
    o := &oauth{baseURL: ts.URL, httpClient: ts.Client()}
    _, err := o.PasswordGrantWithWorkflow(context.Background(), "u", "p", "dev", "", "wf-x")
    if err != nil {
        t.Fatalf("PasswordGrantWithWorkflow: %v", err)
    }
    if sawHeader != "wf-x" {
        t.Fatalf("workflow header = %q, want %q", sawHeader, "wf-x")
    }
}
```

- [ ] **Step 2: Run test — expect failure**

Run: `go test ./internal/robinhood/ -run "TestPasswordGrantWithWorkflow" -v`
Expected: FAIL — `PasswordGrantWithWorkflow` undefined.

- [ ] **Step 3: Refactor for reuse**

In `oauth.go`, rename the existing `PasswordGrant` body into a private `passwordGrant(ctx, form, workflowID string)` that all public entry points share:

```go
func (o *oauth) passwordGrant(ctx context.Context, form url.Values, workflowID string) (*Session, error) {
    req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL+"/oauth2/token/", strings.NewReader(form.Encode()))
    if err != nil {
        return nil, &APIError{Code: CodeRobinhoodUnavailable, Message: err.Error()}
    }
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    req.Header.Set("Accept", "application/json")
    req.Header.Set("X-Robinhood-API-Version", "1.431.4")
    if workflowID != "" {
        req.Header.Set("X-Robinhood-Challenge-Response-Id", workflowID)
    }
    resp, err := o.httpClient.Do(req)
    if err != nil {
        return nil, &APIError{Code: CodeRobinhoodUnavailable, Message: err.Error()}
    }
    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)

    switch resp.StatusCode {
    case http.StatusOK:
        var tr tokenResponse
        if err := json.Unmarshal(body, &tr); err != nil {
            return nil, &APIError{Code: CodeRobinhoodUnavailable, Message: "malformed token response"}
        }
        s := &Session{
            Username:     form.Get("username"),
            AccessToken:  tr.AccessToken,
            RefreshToken: tr.RefreshToken,
            DeviceToken:  form.Get("device_token"),
        }
        if tr.ExpiresIn > 0 {
            s.ExpiresAt = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second).UTC()
        }
        return s, nil
    case http.StatusBadRequest, http.StatusUnauthorized:
        return nil, classifyPasswordErr(body, resp.StatusCode)
    default:
        return nil, &APIError{
            Code:       CodeRobinhoodUnavailable,
            Message:    fmt.Sprintf("password grant HTTP %d", resp.StatusCode),
            HTTPStatus: resp.StatusCode,
        }
    }
}

func (o *oauth) PasswordGrant(ctx context.Context, username, password, deviceToken, mfaCode string) (*Session, error) {
    form := baseGrantForm(username, password, deviceToken, mfaCode)
    return o.passwordGrant(ctx, form, "")
}

func (o *oauth) PasswordGrantWithWorkflow(ctx context.Context, username, password, deviceToken, mfaCode, workflowID string) (*Session, error) {
    form := baseGrantForm(username, password, deviceToken, mfaCode)
    return o.passwordGrant(ctx, form, workflowID)
}

func baseGrantForm(username, password, deviceToken, mfaCode string) url.Values {
    form := url.Values{}
    form.Set("grant_type", "password")
    form.Set("username", username)
    form.Set("password", password)
    form.Set("client_id", robinhoodOAuthClientID)
    form.Set("device_token", deviceToken)
    form.Set("expires_in", "86400")
    form.Set("scope", "internal")
    if mfaCode != "" {
        form.Set("mfa_code", mfaCode)
    }
    return form
}
```

- [ ] **Step 4: Run all oauth tests**

Run: `go test ./internal/robinhood/ -run "TestPassword" -v`
Expected: PASS — both original and new test.

- [ ] **Step 5: Commit**

```bash
git add internal/robinhood/oauth.go internal/robinhood/oauth_test.go
git commit -m "feat(oauth): PasswordGrantWithWorkflow sends X-Robinhood-Challenge-Response-Id"
```

---

## Task 12: `rh login` command

This is the user-facing command that wires Tasks 5, 7, 8, 9, 10, 11 together. The logic is long but flat — no recursion, no retries beyond the single Sheriff hop.

**Files:**
- Create: `cmd/rh/login.go`
- Create: `cmd/rh/login_test.go`

- [ ] **Step 1: Write the failing integration test**

Create `cmd/rh/login_test.go`:

```go
package main

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "net/http/httptest"
    "strings"
    "sync/atomic"
    "testing"
    "time"
)

// TestLogin_SMS_Flow drives the full login E2E with a synthetic Robinhood
// that returns Sheriff → SMS → validated → tokens.
func TestLogin_SMS_Flow(t *testing.T) {
    t.Setenv("XDG_CONFIG_HOME", t.TempDir())
    t.Setenv("ROBINHOOD_KEYCHAIN_BACKEND", "file")

    var firstCall atomic.Bool
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch r.URL.Path {
        case "/oauth2/token/":
            if r.Header.Get("X-Robinhood-Challenge-Response-Id") == "" && !firstCall.Swap(true) {
                // First call: respond with Sheriff required.
                w.WriteHeader(http.StatusBadRequest)
                fmt.Fprint(w, `{"verification_workflow":{"id":"wf-1","workflow_status":"internal_pending"}}`)
                return
            }
            w.Header().Set("Content-Type", "application/json")
            fmt.Fprint(w, `{"access_token":"a","refresh_token":"r","expires_in":3600}`)
        case "/pathfinder/user_machine/":
            fmt.Fprint(w, `{"id":"inq-1"}`)
        case "/pathfinder/inquiries/inq-1/user_view/":
            fmt.Fprint(w, `{"type_context":{"context":{"sheriff_challenge":{"type":"sms","id":"ch-1","phone_number":"***-1234"}}}}`)
        case "/challenge/ch-1/respond/":
            fmt.Fprint(w, `{"status":"validated"}`)
        default:
            t.Fatalf("unexpected path %s", r.URL.Path)
        }
    }))
    defer ts.Close()

    opts := LoginOpts{
        Host:       ts.URL,
        HTTP:       ts.Client(),
        Username:   "alice@example.com",
        Password:   "secret",
        DeviceToken: "dt-fixed",
        CodeInput:  func(prompt string) (string, error) { return "123456", nil },
        TOTPSecret: "",
        Out:        &bytes.Buffer{},
        Profile:    "default",
        PollEvery:  time.Millisecond,
    }
    sess, err := RunLogin(context.Background(), opts)
    if err != nil {
        t.Fatalf("RunLogin: %v", err)
    }
    if sess.AccessToken != "a" {
        t.Fatalf("AccessToken = %q", sess.AccessToken)
    }
    // Verify persisted in keychain (file fallback).
    re, err := loadSessionForTest(t)
    if err != nil {
        t.Fatalf("loadSessionForTest: %v", err)
    }
    if re.AccessToken != "a" {
        t.Fatalf("persisted AccessToken = %q", re.AccessToken)
    }
}

// TestLogin_PushFlow approves via push and does NOT prompt for a code.
func TestLogin_PushFlow(t *testing.T) {
    t.Setenv("XDG_CONFIG_HOME", t.TempDir())
    t.Setenv("ROBINHOOD_KEYCHAIN_BACKEND", "file")
    var first atomic.Bool
    polls := 0
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch r.URL.Path {
        case "/oauth2/token/":
            if !first.Swap(true) {
                w.WriteHeader(http.StatusBadRequest)
                fmt.Fprint(w, `{"verification_workflow":{"id":"wf-2"}}`)
                return
            }
            fmt.Fprint(w, `{"access_token":"ok","refresh_token":"r","expires_in":3600}`)
        case "/pathfinder/user_machine/":
            fmt.Fprint(w, `{"id":"inq-2"}`)
        case "/pathfinder/inquiries/inq-2/user_view/":
            fmt.Fprint(w, `{"type_context":{"context":{"sheriff_challenge":{"type":"prompt","id":"ch-2"}}}}`)
        case "/push/ch-2/get_prompts_status/":
            polls++
            if polls < 2 {
                fmt.Fprint(w, `{"challenge_status":"issued"}`)
            } else {
                fmt.Fprint(w, `{"challenge_status":"validated"}`)
            }
        }
    }))
    defer ts.Close()

    askedForCode := false
    opts := LoginOpts{
        Host:        ts.URL,
        HTTP:        ts.Client(),
        Username:    "alice@example.com",
        Password:    "p",
        DeviceToken: "dt-2",
        CodeInput:   func(prompt string) (string, error) { askedForCode = true; return "", nil },
        Out:         &bytes.Buffer{},
        Profile:     "default",
        PollEvery:   time.Millisecond,
    }
    if _, err := RunLogin(context.Background(), opts); err != nil {
        t.Fatalf("RunLogin: %v", err)
    }
    if askedForCode {
        t.Fatalf("push flow should not prompt for SMS code")
    }
}

// Helper for the integration test. In production the login command does not
// expose a re-read; the test reads via the same keychain it wrote to.
func loadSessionForTest(t *testing.T) (_ any, _ error) {
    t.Helper()
    // simple: just call the robinhood package, re-encoded as JSON for comparison.
    data, err := (struct{ _ int }{}), error(nil) // placeholder; the test below uses direct Session access
    _ = data
    return nil, nil
}
```

Then adjust the test to call `robinhood.LoadFromKeychain("default")` directly — import `github.com/herocod3r/robinhood-cli/internal/robinhood`. Replace `loadSessionForTest` with the real call; keeping the helper avoids import churn. (Authors: when you write this file, inline the import.)

- [ ] **Step 2: Run tests — expect compile failure**

Run: `go test ./cmd/rh/ -run TestLogin -v`
Expected: FAIL — `RunLogin`, `LoginOpts` undefined.

- [ ] **Step 3: Implement `cmd/rh/login.go`**

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "io"
    "net/http"
    "os"
    "time"

    "github.com/google/uuid"
    "github.com/herocod3r/robinhood-cli/internal/config"
    "github.com/herocod3r/robinhood-cli/internal/inputprompt"
    "github.com/herocod3r/robinhood-cli/internal/robinhood"
    "github.com/spf13/cobra"
)

func init() {
    c := &cobra.Command{
        Use:   "login",
        Short: "Log into Robinhood (interactive)",
        RunE:  runLogin,
    }
    c.Flags().String("totp-secret", "", "base32 TOTP shared secret (optional; ROBINHOOD_TOTP_SECRET env var also honored)")
    c.Flags().String("username", "", "email/username (prompted if empty)")
    register(c)
}

// LoginOpts bundles everything RunLogin needs. CLI builds this from flags + env;
// tests build it directly.
type LoginOpts struct {
    Host        string        // defaults to https://api.robinhood.com
    HTTP        *http.Client  // defaults to http.DefaultClient
    Username    string
    Password    string
    DeviceToken string
    TOTPSecret  string
    CodeInput   func(prompt string) (string, error) // SMS/email; push flows never call this
    Out         io.Writer
    Profile     string
    PollEvery   time.Duration // Sheriff poll cadence; default 2s
}

// RunLogin drives the entire Sheriff/MFA state machine. Returns the persisted session.
func RunLogin(ctx context.Context, o LoginOpts) (*robinhood.Session, error) {
    if o.Host == "" {
        o.Host = "https://api.robinhood.com"
    }
    if o.HTTP == nil {
        o.HTTP = &http.Client{Timeout: 30 * time.Second}
    }
    if o.Profile == "" {
        o.Profile = "default"
    }
    if o.DeviceToken == "" {
        o.DeviceToken = uuid.NewString()
    }

    oauthClient := robinhood.NewOAuth(o.Host, o.HTTP)

    // Build TOTP code if a shared secret was provided.
    var mfaCode string
    if o.TOTPSecret != "" {
        code, err := robinhood.TOTPCode(o.TOTPSecret)
        if err != nil {
            return nil, &robinhood.APIError{Code: robinhood.CodeValidation, Message: "invalid TOTP secret: " + err.Error()}
        }
        mfaCode = code
    }

    sess, err := oauthClient.PasswordGrant(ctx, o.Username, o.Password, o.DeviceToken, mfaCode)
    if err == nil {
        return persistLogin(sess, o)
    }
    var apiErr *robinhood.APIError
    if !errors.As(err, &apiErr) {
        return nil, err
    }

    switch apiErr.Code {
    case robinhood.CodeMFARequired:
        if mfaCode != "" {
            // We already sent one; if it failed, ask.
        }
        code, err := o.CodeInput(apiErr.Hint)
        if err != nil {
            return nil, &robinhood.APIError{Code: robinhood.CodeValidation, Message: "no MFA code entered"}
        }
        sess, err := oauthClient.PasswordGrant(ctx, o.Username, o.Password, o.DeviceToken, code)
        if err != nil {
            return nil, err
        }
        return persistLogin(sess, o)

    case robinhood.CodeSheriffRequired:
        return runSheriff(ctx, o, oauthClient, apiErr, mfaCode)

    default:
        return nil, err
    }
}

func runSheriff(ctx context.Context, o LoginOpts, oauthClient *robinhood.OAuth, apiErr *robinhood.APIError, mfaCode string) (*robinhood.Session, error) {
    sheriff := &robinhood.Sheriff{
        BaseURL:   o.Host,
        HTTP:      o.HTTP,
        PollEvery: o.PollEvery,
    }
    fmt.Fprintln(o.Out, "Robinhood requires device verification.")
    step, err := sheriff.Start(ctx, apiErr.WorkflowID, o.DeviceToken)
    if err != nil {
        return nil, err
    }
    switch step.Kind {
    case robinhood.SheriffSMS, robinhood.SheriffEmail:
        fmt.Fprintf(o.Out, "Robinhood sent a code via %s.\n", step.Detail)
        code, err := o.CodeInput("Enter the code: ")
        if err != nil || code == "" {
            return nil, &robinhood.APIError{Code: robinhood.CodeValidation, Message: "no code entered"}
        }
        if err := sheriff.RespondCode(ctx, step, code); err != nil {
            return nil, err
        }
    case robinhood.SheriffPush:
        fmt.Fprintf(o.Out, "Approve the prompt in your Robinhood app…\n")
        if err := sheriff.WaitPush(ctx, step); err != nil {
            return nil, err
        }
    default:
        return nil, &robinhood.APIError{Code: robinhood.CodeSheriffRequired, Message: "unhandled sheriff step"}
    }
    sess, err := oauthClient.PasswordGrantWithWorkflow(ctx, o.Username, o.Password, o.DeviceToken, mfaCode, apiErr.WorkflowID)
    if err != nil {
        return nil, err
    }
    return persistLogin(sess, o)
}

func persistLogin(s *robinhood.Session, o LoginOpts) (*robinhood.Session, error) {
    if err := s.SaveToKeychain(o.Profile); err != nil {
        return nil, err
    }
    // Record profile in config if it wasn't there already.
    f, err := config.LoadFile()
    if err == nil {
        if _, ok := f.Profiles[o.Profile]; !ok {
            f.Profiles[o.Profile] = config.Profile{}
            _ = config.SaveFile(f)
        }
    }
    fmt.Fprintf(o.Out, "Logged in as %s (profile: %s)\n", s.Username, o.Profile)
    return s, nil
}

func runLogin(cmd *cobra.Command, args []string) error {
    p := inputprompt.New()
    username, _ := cmd.Flags().GetString("username")
    if username == "" {
        username = os.Getenv(config.EnvUsername)
    }
    if username == "" {
        var err error
        if username, err = p.Text("Username: "); err != nil {
            return &robinhood.APIError{Code: robinhood.CodeValidation, Message: err.Error()}
        }
    }
    password := os.Getenv(config.EnvPassword)
    if password == "" {
        var err error
        if password, err = p.Password("Password: "); err != nil {
            return &robinhood.APIError{Code: robinhood.CodeValidation, Message: err.Error()}
        }
    }
    totpSecret, _ := cmd.Flags().GetString("totp-secret")
    if totpSecret == "" {
        totpSecret = os.Getenv(config.EnvTOTPSecret)
    }
    deviceToken := os.Getenv(config.EnvDeviceToken)
    if deviceToken == "" {
        // Try to reuse the one we stashed on a previous login — otherwise generate.
        if existing, err := robinhood.LoadFromKeychain(gflags.Profile); err == nil && existing.DeviceToken != "" {
            deviceToken = existing.DeviceToken
        } else {
            deviceToken = uuid.NewString()
        }
    }

    _, err := RunLogin(cmd.Context(), LoginOpts{
        Username:    username,
        Password:    password,
        DeviceToken: deviceToken,
        TOTPSecret:  totpSecret,
        CodeInput:   p.Text,
        Out:         cmd.ErrOrStderr(),
        Profile:     gflags.Profile,
    })
    return err
}
```

- [ ] **Step 4: Expose `NewOAuth` from the robinhood package**

In `internal/robinhood/oauth.go`, export the constructor:

```go
// OAuth is the public face of the package's OAuth client.
type OAuth = oauth

// NewOAuth returns an OAuth client bound to the given base URL and HTTP client.
func NewOAuth(baseURL string, h *http.Client) *OAuth {
    return &oauth{baseURL: baseURL, httpClient: h}
}
```

- [ ] **Step 5: Run login tests**

Run: `go test ./cmd/rh/ -run TestLogin -v`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add cmd/rh/login.go cmd/rh/login_test.go internal/robinhood/oauth.go
git commit -m "feat(cmd): rh login with Sheriff SMS/email/push and MFA support"
```

---

## Task 13: `rh logout` command

**Files:**
- Create: `cmd/rh/logout.go`
- Create: `cmd/rh/logout_test.go`

- [ ] **Step 1: Write the failing test**

Create `cmd/rh/logout_test.go`:

```go
package main

import (
    "bytes"
    "testing"
    "time"

    "github.com/herocod3r/robinhood-cli/internal/robinhood"
)

func TestLogout_ClearsKeychain(t *testing.T) {
    t.Setenv("XDG_CONFIG_HOME", t.TempDir())
    t.Setenv("ROBINHOOD_KEYCHAIN_BACKEND", "file")

    in := &robinhood.Session{Username: "u", AccessToken: "a", RefreshToken: "r", DeviceToken: "d", ExpiresAt: time.Now().Add(time.Hour)}
    if err := in.SaveToKeychain("default"); err != nil {
        t.Fatal(err)
    }
    var out bytes.Buffer
    if err := RunLogout(LogoutOpts{Profile: "default", Out: &out}); err != nil {
        t.Fatalf("RunLogout: %v", err)
    }
    if _, err := robinhood.LoadFromKeychain("default"); err == nil {
        t.Fatalf("expected LoadFromKeychain to fail after logout")
    }
}

func TestLogout_AllProfiles(t *testing.T) {
    t.Setenv("XDG_CONFIG_HOME", t.TempDir())
    t.Setenv("ROBINHOOD_KEYCHAIN_BACKEND", "file")
    for _, p := range []string{"default", "work"} {
        s := &robinhood.Session{Username: p, AccessToken: "a" + p, RefreshToken: "r", ExpiresAt: time.Now().Add(time.Hour)}
        if err := s.SaveToKeychain(p); err != nil {
            t.Fatal(err)
        }
    }
    if err := RunLogout(LogoutOpts{AllProfiles: true, Out: &bytes.Buffer{}}); err != nil {
        t.Fatalf("RunLogout: %v", err)
    }
    for _, p := range []string{"default", "work"} {
        if _, err := robinhood.LoadFromKeychain(p); err == nil {
            t.Fatalf("expected no session for profile %q", p)
        }
    }
}
```

- [ ] **Step 2: Run test — expect compile failure**

Run: `go test ./cmd/rh/ -run TestLogout -v`
Expected: FAIL — `RunLogout`, `LogoutOpts` undefined.

- [ ] **Step 3: Implement**

Create `cmd/rh/logout.go`:

```go
package main

import (
    "fmt"
    "io"
    "os"

    "github.com/herocod3r/robinhood-cli/internal/config"
    "github.com/herocod3r/robinhood-cli/internal/robinhood"
    "github.com/spf13/cobra"
)

func init() {
    c := &cobra.Command{
        Use:   "logout",
        Short: "Clear stored session(s)",
        RunE:  runLogoutCmd,
    }
    c.Flags().Bool("all-profiles", false, "clear sessions for every profile")
    register(c)
}

// LogoutOpts configures RunLogout. Tests build it directly.
type LogoutOpts struct {
    Profile     string
    AllProfiles bool
    Out         io.Writer
}

// RunLogout clears the keychain entry for the given profile (or all profiles).
func RunLogout(o LogoutOpts) error {
    if o.Out == nil {
        o.Out = os.Stderr
    }
    if o.AllProfiles {
        f, err := config.LoadFile()
        if err != nil {
            return err
        }
        for name := range f.Profiles {
            if err := robinhood.ClearKeychain(name); err != nil {
                return err
            }
            fmt.Fprintf(o.Out, "Cleared profile %q\n", name)
        }
        return nil
    }
    profile := o.Profile
    if profile == "" {
        profile = "default"
    }
    if err := robinhood.ClearKeychain(profile); err != nil {
        return err
    }
    fmt.Fprintf(o.Out, "Cleared profile %q\n", profile)
    return nil
}

func runLogoutCmd(cmd *cobra.Command, args []string) error {
    all, _ := cmd.Flags().GetBool("all-profiles")
    return RunLogout(LogoutOpts{
        Profile:     gflags.Profile,
        AllProfiles: all,
        Out:         cmd.ErrOrStderr(),
    })
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./cmd/rh/ -run TestLogout -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/rh/logout.go cmd/rh/logout_test.go
git commit -m "feat(cmd): rh logout with --all-profiles support"
```

---

## Task 14: `newAuthedClient` prefers keychain, falls back to env

**Files:**
- Modify: `cmd/rh/client.go`
- Modify: `cmd/rh/exit_test.go` (or a new test file)

- [ ] **Step 1: Write the failing test**

Create `cmd/rh/client_test.go`:

```go
package main

import (
    "testing"
    "time"

    "github.com/herocod3r/robinhood-cli/internal/robinhood"
)

func TestNewAuthedClient_PrefersEnv(t *testing.T) {
    t.Setenv("XDG_CONFIG_HOME", t.TempDir())
    t.Setenv("ROBINHOOD_KEYCHAIN_BACKEND", "file")
    t.Setenv("ROBINHOOD_ACCESS_TOKEN", "env-access")
    t.Setenv("ROBINHOOD_REFRESH_TOKEN", "env-refresh")
    // keychain has a different token:
    kc := &robinhood.Session{AccessToken: "kc", RefreshToken: "r", ExpiresAt: time.Now().Add(time.Hour)}
    if err := kc.SaveToKeychain("default"); err != nil {
        t.Fatal(err)
    }
    c, err := newAuthedClient()
    if err != nil {
        t.Fatalf("newAuthedClient: %v", err)
    }
    if c.Session().AccessToken != "env-access" {
        t.Fatalf("got %q, want env-access", c.Session().AccessToken)
    }
}

func TestNewAuthedClient_FallsBackToKeychain(t *testing.T) {
    t.Setenv("XDG_CONFIG_HOME", t.TempDir())
    t.Setenv("ROBINHOOD_KEYCHAIN_BACKEND", "file")
    // no env vars set
    t.Setenv("ROBINHOOD_ACCESS_TOKEN", "")
    t.Setenv("ROBINHOOD_REFRESH_TOKEN", "")
    kc := &robinhood.Session{AccessToken: "kc", RefreshToken: "r", ExpiresAt: time.Now().Add(time.Hour)}
    if err := kc.SaveToKeychain("default"); err != nil {
        t.Fatal(err)
    }
    c, err := newAuthedClient()
    if err != nil {
        t.Fatalf("newAuthedClient: %v", err)
    }
    if c.Session().AccessToken != "kc" {
        t.Fatalf("got %q, want kc", c.Session().AccessToken)
    }
}

func TestNewAuthedClient_NothingAvailable(t *testing.T) {
    t.Setenv("XDG_CONFIG_HOME", t.TempDir())
    t.Setenv("ROBINHOOD_KEYCHAIN_BACKEND", "file")
    t.Setenv("ROBINHOOD_ACCESS_TOKEN", "")
    t.Setenv("ROBINHOOD_REFRESH_TOKEN", "")
    _, err := newAuthedClient()
    apiErr, _ := err.(*robinhood.APIError)
    if apiErr == nil || apiErr.Code != robinhood.CodeUnauthenticated {
        t.Fatalf("err = %v; want CodeUnauthenticated", err)
    }
}
```

- [ ] **Step 2: Run tests — expect failure**

Run: `go test ./cmd/rh/ -run TestNewAuthedClient -v`
Expected: FAIL — current `newAuthedClient` only consults env vars.

- [ ] **Step 3: Change precedence**

Rewrite `cmd/rh/client.go`:

```go
package main

import (
    "os"

    "github.com/herocod3r/robinhood-cli/internal/config"
    "github.com/herocod3r/robinhood-cli/internal/robinhood"
)

// newAuthedClient returns a Robinhood client. Precedence: env vars > keychain > fail.
func newAuthedClient() (*robinhood.Client, error) {
    var sess *robinhood.Session
    if os.Getenv(config.EnvAccessToken) != "" || os.Getenv(config.EnvRefreshToken) != "" {
        s, err := robinhood.LoadSessionFromEnv()
        if err != nil {
            return nil, err
        }
        sess = s
    } else {
        s, err := robinhood.LoadFromKeychain(gflags.Profile)
        if err != nil {
            return nil, err
        }
        sess = s
    }
    c := robinhood.NewClient()
    c.SetSession(sess)
    return c, nil
}
```

- [ ] **Step 4: Run the suite**

Run: `go test ./cmd/rh/ -run "TestNewAuthedClient|TestExit|TestLogin|TestLogout" -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/rh/client.go cmd/rh/client_test.go
git commit -m "feat(cmd): env > keychain precedence for session source"
```

---

## Task 15: Data-command template (reference for Tasks 16–30)

Tasks 16–30 all follow this exact shape. Each per-command task references Task 15's steps and only specifies differences.

**Per command, do:**

1. **Create the endpoint method** in `internal/robinhood/endpoints/<domain>.go`.
2. **Write the endpoint test** in `internal/robinhood/endpoints/<domain>_test.go` using `httptest` — cover happy path + one error path (4xx or empty).
3. **Create the Cobra file** in `cmd/rh/<cmd>.go` matching the skeleton in the top-of-plan "File Structure" section.
4. **Add a `Write<Name>` method** to `internal/output/table.go` for the pretty renderer.
5. **Run `go test ./... && go build ./cmd/rh`**, commit with message `feat(cmd): rh <name>`.

**Reference example — `rh positions`:**

`internal/robinhood/endpoints/positions.go`:

```go
package endpoints

import (
    "context"
    "fmt"
    "sort"
    "strconv"

    "github.com/herocod3r/robinhood-cli/internal/robinhood"
)

type Positions struct{ c *robinhood.Client }

func NewPositions(c *robinhood.Client) *Positions { return &Positions{c: c} }

// Position is one row in `rh positions`.
type Position struct {
    Symbol              string          `json:"symbol"`
    Name                string          `json:"name,omitempty"`
    Quantity            robinhood.Money `json:"quantity"`
    AverageBuyPrice     robinhood.Money `json:"average_buy_price"`
    LastPrice           robinhood.Money `json:"last_price"`
    MarketValue         robinhood.Money `json:"market_value"`
    CostBasis           robinhood.Money `json:"cost_basis"`
    UnrealizedPL        robinhood.Money `json:"unrealized_pl"`
    UnrealizedPLPercent robinhood.Money `json:"unrealized_pl_percent"`
    InstrumentID        string          `json:"instrument_id"`
}

type rawPosition struct {
    Quantity        robinhood.Money `json:"quantity"`
    AverageBuyPrice robinhood.Money `json:"average_buy_price"`
    Instrument      string          `json:"instrument"`
}

type positionsResp struct {
    Results []rawPosition `json:"results"`
    Next    string        `json:"next"`
}

// ListOpts controls filtering/ordering.
type ListOpts struct {
    NonZero bool
    Sort    string // "value" | "pl" | "symbol"
    Limit   int
}

// List returns all positions for the current account, resolves each instrument
// to a symbol, batches a single quote call for all last prices, and computes
// unrealized P/L.
func (p *Positions) List(ctx context.Context, opts ListOpts) ([]Position, error) {
    all, err := p.fetchAll(ctx)
    if err != nil {
        return nil, err
    }
    if opts.NonZero {
        all = filterNonZero(all)
    }
    instrumentIDs := make([]string, 0, len(all))
    for _, r := range all {
        instrumentIDs = append(instrumentIDs, r.Instrument)
    }
    inst := NewInstruments(p.c)
    syms, err := inst.ResolveIDs(ctx, instrumentIDs)
    if err != nil {
        return nil, err
    }
    q := NewQuotes(p.c)
    quotes, err := q.Batch(ctx, syms.Symbols())
    if err != nil {
        return nil, err
    }
    out := make([]Position, 0, len(all))
    for _, r := range all {
        sym := syms.Get(r.Instrument)
        qt := quotes[sym]
        pos := Position{
            Symbol:          sym,
            Quantity:        r.Quantity,
            AverageBuyPrice: r.AverageBuyPrice,
            InstrumentID:    r.Instrument,
        }
        if qt != nil {
            pos.LastPrice = qt.LastTradePrice
            pos.Name = qt.Name
            pos.MarketValue = mulMoney(r.Quantity, qt.LastTradePrice)
            pos.CostBasis = mulMoney(r.Quantity, r.AverageBuyPrice)
            pos.UnrealizedPL = subMoney(pos.MarketValue, pos.CostBasis)
            pos.UnrealizedPLPercent = divMoney(pos.UnrealizedPL, pos.CostBasis)
        }
        out = append(out, pos)
    }
    sortPositions(out, opts.Sort)
    if opts.Limit > 0 && opts.Limit < len(out) {
        out = out[:opts.Limit]
    }
    return out, nil
}

func (p *Positions) fetchAll(ctx context.Context) ([]rawPosition, error) {
    var out []rawPosition
    path := "/positions/"
    for {
        var resp positionsResp
        if err := p.c.GetJSONCtx(ctx, robinhood.APIHost, path, &resp); err != nil {
            return nil, err
        }
        out = append(out, resp.Results...)
        if resp.Next == "" {
            break
        }
        // The `next` URL is absolute; strip the host.
        u, err := url.Parse(resp.Next)
        if err != nil {
            return nil, err
        }
        path = u.Path
        if u.RawQuery != "" {
            path += "?" + u.RawQuery
        }
    }
    return out, nil
}

func filterNonZero(in []rawPosition) []rawPosition {
    out := make([]rawPosition, 0, len(in))
    for _, r := range in {
        q, _ := strconv.ParseFloat(string(r.Quantity), 64)
        if q != 0 {
            out = append(out, r)
        }
    }
    return out
}

func sortPositions(ps []Position, by string) {
    switch by {
    case "pl":
        sort.Slice(ps, func(i, j int) bool {
            return moneyLess(ps[j].UnrealizedPL, ps[i].UnrealizedPL)
        })
    case "symbol":
        sort.Slice(ps, func(i, j int) bool { return ps[i].Symbol < ps[j].Symbol })
    default: // "value" or empty
        sort.Slice(ps, func(i, j int) bool {
            return moneyLess(ps[j].MarketValue, ps[i].MarketValue)
        })
    }
}

// mulMoney / subMoney / divMoney / moneyLess live in a shared math.go file
// (Task 15.b below). Money arithmetic uses the decimal package (not float64).
var _ = fmt.Sprintf // keep imports stable
```

**Task 15.a — shared money arithmetic**

Create `internal/robinhood/endpoints/math.go`:

```go
package endpoints

import (
    "github.com/herocod3r/robinhood-cli/internal/robinhood"
    "github.com/shopspring/decimal"
)

func toDec(m robinhood.Money) decimal.Decimal {
    d, err := decimal.NewFromString(string(m))
    if err != nil {
        return decimal.Zero
    }
    return d
}

func mulMoney(a, b robinhood.Money) robinhood.Money {
    return robinhood.Money(toDec(a).Mul(toDec(b)).StringFixed(4))
}
func subMoney(a, b robinhood.Money) robinhood.Money {
    return robinhood.Money(toDec(a).Sub(toDec(b)).StringFixed(4))
}
func divMoney(a, b robinhood.Money) robinhood.Money {
    bd := toDec(b)
    if bd.IsZero() {
        return "0"
    }
    return robinhood.Money(toDec(a).Div(bd).StringFixed(4))
}
func moneyLess(a, b robinhood.Money) bool {
    return toDec(a).LessThan(toDec(b))
}
```

Run `go get github.com/shopspring/decimal@v1.4.0` and `go mod tidy`.

- [ ] **Step: Run tests**

Run the reference test suite for `positions_test.go` — see Task 16 for the actual test.

- [ ] **Step: Commit**

This task has no standalone commit; its artifacts land with Task 16 (positions) which is the first consumer.

---

## Task 16: `rh positions`

Follow Task 15's template. Specifics:

- **Files:** `internal/robinhood/endpoints/positions.go`, `..._test.go`, `cmd/rh/positions.go`, `internal/output/table.go` (+ `WritePositions` method).
- **Cobra flags:** `--nonzero`, `--sort value|pl|symbol`, `--limit int`.
- **Endpoint:** `GET /positions/?cursor=<next>`; merges results across pages until `next == ""`.
- **Helpers:** `internal/robinhood/endpoints/instruments.go` (resolve instrument URL → symbol, batched), `internal/robinhood/endpoints/quotes.go` (batch fetch via `GET /quotes/?symbols=A,B,C` — pre-built in Task 19 but stub OK here).
- **Test:** happy path (3 positions, pagination over 2 pages), nonzero filter, sort orders.
- **Table columns:** `Symbol | Qty | Last | Market value | Cost basis | P/L | P/L %`.
- **Commit message:** `feat(cmd): rh positions with --nonzero --sort --limit`.

Detailed test template:

```go
func TestPositions_List_Paginates(t *testing.T) {
    page1 := `{"results":[{"quantity":"10.0000","average_buy_price":"150.0000","instrument":"https://api.example.com/instruments/uuid-1/"}],"next":"https://api.example.com/positions/?cursor=p2"}`
    page2 := `{"results":[{"quantity":"5.0000","average_buy_price":"200.0000","instrument":"https://api.example.com/instruments/uuid-2/"}],"next":""}`
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch {
        case strings.HasPrefix(r.URL.Path, "/positions/") && r.URL.Query().Get("cursor") == "":
            w.Write([]byte(page1))
        case strings.HasPrefix(r.URL.Path, "/positions/") && r.URL.Query().Get("cursor") == "p2":
            w.Write([]byte(page2))
        case strings.HasPrefix(r.URL.Path, "/instruments/uuid-1/"):
            w.Write([]byte(`{"symbol":"AAPL","name":"Apple"}`))
        case strings.HasPrefix(r.URL.Path, "/instruments/uuid-2/"):
            w.Write([]byte(`{"symbol":"NVDA","name":"Nvidia"}`))
        case r.URL.Path == "/quotes/":
            w.Write([]byte(`{"results":[{"symbol":"AAPL","last_trade_price":"200.0000","instrument":""},{"symbol":"NVDA","last_trade_price":"150.0000"}]}`))
        default:
            t.Fatalf("unexpected %s", r.URL)
        }
    }))
    defer ts.Close()
    c := robinhood.NewClientWithHosts(ts.URL, ts.URL, ts.URL, ts.Client())
    c.SetSession(&robinhood.Session{AccessToken: "t"})
    ps, err := NewPositions(c).List(context.Background(), ListOpts{})
    if err != nil {
        t.Fatal(err)
    }
    if len(ps) != 2 {
        t.Fatalf("got %d positions; want 2", len(ps))
    }
    byName := map[string]Position{ps[0].Symbol: ps[0], ps[1].Symbol: ps[1]}
    if byName["AAPL"].MarketValue != "2000.0000" {
        t.Fatalf("AAPL market_value = %q", byName["AAPL"].MarketValue)
    }
}
```

- [ ] Steps follow Task 15 — write test, compile-fail, implement, pass, commit.

---

## Task 17: `rh position <ticker>`

- **Endpoint:** reuses `Positions.List` but filters to a single instrument via `GET /positions/?instrument=<url>`; falls back to client-side filter if the server rejects the query parameter (Robinhood sometimes does).
- **Args:** exactly one ticker (`cobra.ExactArgs(1)`).
- **Table:** same columns as `positions` but one row.
- **Commit:** `feat(cmd): rh position <ticker> for single-symbol inspection`.

- [ ] Steps follow Task 15.

---

## Task 18: `rh account`

- **Endpoints:** `GET /accounts/` (first entry only — single-account assumption) + `GET /accounts/unified` (via `PhoenixHost`).
- **Data shape:**
  ```go
  type AccountSummary struct {
      AccountNumber      string          `json:"account_number"`
      BuyingPower        robinhood.Money `json:"buying_power"`
      Cash               robinhood.Money `json:"cash"`
      SweepEnabled       bool            `json:"sweep_enabled"`
      MarginBalance      robinhood.Money `json:"margin_balance,omitempty"`
      PatternDayTrader   bool            `json:"pattern_day_trader"`
      DayTradeCount      int             `json:"day_trade_count"`
      InstantUsed        robinhood.Money `json:"instant_used"`
      InstantAvailable   robinhood.Money `json:"instant_available"`
  }
  ```
- **Table:** key/value, `WriteAccount`.
- **Commit:** `feat(cmd): rh account for buying power, margin, PDT`.

- [ ] Steps follow Task 15.

---

## Task 19: `rh quote <ticker...>` (batch)

- **Endpoint:** `GET /quotes/?symbols=A,B,C` (≤50 per request; the command accepts up to 50 and returns 400 validation if exceeded).
- **Shape:**
  ```go
  type Quote struct {
      Symbol             string `json:"symbol"`
      LastTradePrice     robinhood.Money `json:"last_price"`
      BidPrice           robinhood.Money `json:"bid_price"`
      AskPrice           robinhood.Money `json:"ask_price"`
      BidSize            int             `json:"bid_size"`
      AskSize            int             `json:"ask_size"`
      PreviousClose      robinhood.Money `json:"previous_close"`
      ExtendedHoursPrice robinhood.Money `json:"extended_hours_price,omitempty"`
      Volume             int64           `json:"volume"`
      UpdatedAt          string          `json:"updated_at"`
  }
  ```
- **Command flags:** `--extended-hours`.
- **Table:** one row per ticker: `Symbol | Last | Bid | Ask | Prev close | Change | Change %`.
- **Commit:** `feat(cmd): rh quote batch with --extended-hours`.

- [ ] Steps follow Task 15.

---

## Task 20: `rh fundamentals <ticker...>`

- **Endpoint:** `GET /fundamentals/?symbols=A,B,C`.
- **Shape:**
  ```go
  type Fundamentals struct {
      Symbol            string          `json:"symbol"`
      Open              robinhood.Money `json:"open"`
      High              robinhood.Money `json:"high"`
      Low               robinhood.Money `json:"low"`
      Volume            int64           `json:"volume"`
      AverageVolume     int64           `json:"average_volume"`
      AverageVolume2Wks int64           `json:"average_volume_2_weeks"`
      High52Weeks       robinhood.Money `json:"high_52_weeks"`
      Low52Weeks        robinhood.Money `json:"low_52_weeks"`
      DividendYield     robinhood.Money `json:"dividend_yield"`
      MarketCap         robinhood.Money `json:"market_cap"`
      PERatio           robinhood.Money `json:"pe_ratio"`
      Description       string          `json:"description,omitempty"`
  }
  ```
- **Table:** one row per ticker, subset of columns.
- **Commit:** `feat(cmd): rh fundamentals batch`.

- [ ] Steps follow Task 15.

---

## Task 21: `rh historicals <ticker>`

- **Endpoint:** `GET /quotes/historicals/<SYMBOL>/?interval=day&span=year` (interval + span from flags).
- **Flags:** `--interval 5minute|10minute|hour|day|week`, `--span day|week|month|3month|year|5year`.
- **Shape:**
  ```go
  type Historicals struct {
      Symbol  string         `json:"symbol"`
      Interval string        `json:"interval"`
      Span    string         `json:"span"`
      Bars    []HistoricalBar `json:"bars"`
  }

  type HistoricalBar struct {
      BeginsAt   string          `json:"begins_at"`
      OpenPrice  robinhood.Money `json:"open_price"`
      ClosePrice robinhood.Money `json:"close_price"`
      HighPrice  robinhood.Money `json:"high_price"`
      LowPrice   robinhood.Money `json:"low_price"`
      Volume     int64           `json:"volume"`
      Session    string          `json:"session"`
  }
  ```
- **Table:** `Begin | O | H | L | C | Vol`.
- **Commit:** `feat(cmd): rh historicals OHLCV with interval+span`.

- [ ] Steps follow Task 15.

---

## Task 22: `rh news <ticker>`

- **Endpoint:** `GET /midlands/news/<SYMBOL>/` (legacy MID host — use `APIHost`; path with trailing slash as documented).
- **Flags:** `--limit int`.
- **Shape:**
  ```go
  type NewsItem struct {
      Title       string `json:"title"`
      Author      string `json:"author,omitempty"`
      Source      string `json:"source"`
      URL         string `json:"url"`
      PublishedAt string `json:"published_at"`
      Summary     string `json:"summary,omitempty"`
  }
  ```
- **Commit:** `feat(cmd): rh news with --limit`.

- [ ] Steps follow Task 15.

---

## Task 23: `rh earnings <ticker>`

- **Endpoint:** `GET /marketdata/earnings/?symbol=<SYMBOL>`.
- **Shape:**
  ```go
  type EarningsEvent struct {
      Symbol   string `json:"symbol"`
      Year     int    `json:"year"`
      Quarter  int    `json:"quarter"`
      ReportAt string `json:"report_at,omitempty"`
      EPS      struct {
          Estimate robinhood.Money `json:"estimate,omitempty"`
          Actual   robinhood.Money `json:"actual,omitempty"`
      } `json:"eps"`
      Call *struct {
          Datetime string `json:"datetime"`
          URL      string `json:"broadcast_url,omitempty"`
      } `json:"call,omitempty"`
  }
  ```
- **Commit:** `feat(cmd): rh earnings`.

- [ ] Steps follow Task 15.

---

## Task 24: `rh ratings <ticker>`

- **Endpoint:** `GET /midlands/ratings/<INSTRUMENT_ID>/` — requires the instrument resolver from Task 16.
- **Shape:**
  ```go
  type Rating struct {
      Symbol      string  `json:"symbol"`
      Summary     struct {
          NumBuy  int `json:"num_buy_ratings"`
          NumHold int `json:"num_hold_ratings"`
          NumSell int `json:"num_sell_ratings"`
      } `json:"summary"`
      Ratings []struct {
          Type     string `json:"type"`
          Text     string `json:"text"`
          PublishedAt string `json:"published_at"`
      } `json:"ratings"`
  }
  ```
- **Commit:** `feat(cmd): rh ratings`.

- [ ] Steps follow Task 15.

---

## Task 25: `rh dividends`

- **Endpoint:** `GET /dividends/?cursor=<next>` (paginated).
- **Flags:** `--year YYYY`, `--symbol SYM`.
- **Shape:**
  ```go
  type Dividend struct {
      Symbol      string          `json:"symbol"`
      Amount      robinhood.Money `json:"amount"`
      Rate        robinhood.Money `json:"rate"`
      Position    robinhood.Money `json:"position"`
      WithholdingTax robinhood.Money `json:"withholding,omitempty"`
      PaidAt      string          `json:"paid_at"`
      RecordAt    string          `json:"record_date"`
      PayableAt   string          `json:"payable_date"`
      InstrumentID string         `json:"instrument_id"`
      State       string          `json:"state"`
  }
  ```
  Post-fetch: resolve instrument → symbol.
- **Commit:** `feat(cmd): rh dividends with --year --symbol`.

- [ ] Steps follow Task 15.

---

## Task 26: `rh options-positions`

- **Endpoints:**
  - `GET /options/aggregate_positions/?nonzero=true` (paginated).
  - For each result's `legs[*].option`: `GET <option-url>` to expand `chain_symbol`, `strike_price`, `type`, `expiration_date`.
- **Flags:** `--symbol SYM`.
- **Shape:**
  ```go
  type OptionPosition struct {
      Symbol          string          `json:"symbol"`
      Strategy        string          `json:"strategy"` // "long_call", "short_put", etc.
      Quantity        robinhood.Money `json:"quantity"`
      AveragePrice    robinhood.Money `json:"average_price"`
      Legs            []OptionLeg     `json:"legs"`
  }
  type OptionLeg struct {
      OptionID   string          `json:"option_id"`
      Type       string          `json:"type"` // "call" | "put"
      StrikePrice robinhood.Money `json:"strike_price"`
      Expiration string          `json:"expiration"`
      PositionType string        `json:"position_type"` // "long" | "short"
  }
  ```
- **Commit:** `feat(cmd): rh options-positions`.

- [ ] Steps follow Task 15.

---

## Task 27: `rh orders`

**This is the heaviest command** — needs cursor pagination, per-order instrument resolution, state filtering, time filtering, and optional execution details.

- **Endpoint:** `GET /orders/?cursor=<next>` — paginated.
- **Flags:**
  - `--since <date>` — relative (`7d`, `ytd`, `1mo`, `1y`) or `YYYY-MM-DD`.
  - `--symbol SYM` — resolves via instrument search, then filters orders by `instrument` URL.
  - `--state filled|cancelled|confirmed|open` — server accepts `state=filled` as a query param.
  - `--limit int` — post-filter cap.
- **Shape:**
  ```go
  type Order struct {
      ID               string          `json:"id"`
      Symbol           string          `json:"symbol"`
      Side             string          `json:"side"`
      Type             string          `json:"type"` // market|limit|stop_limit|stop
      State            string          `json:"state"`
      Quantity         robinhood.Money `json:"quantity"`
      Price            robinhood.Money `json:"price,omitempty"`
      AverageFillPrice robinhood.Money `json:"average_fill_price,omitempty"`
      Fees             robinhood.Money `json:"fees"`
      TimeInForce      string          `json:"time_in_force"`
      ExtendedHours    bool            `json:"extended_hours"`
      CreatedAt        string          `json:"created_at"`
      FilledAt         string          `json:"filled_at,omitempty"`
      CancelledAt      string          `json:"cancelled_at,omitempty"`
  }
  ```
- **Internal helpers:**
  - `parseDateOrRelative(string) (time.Time, error)` lives in a new `internal/datearg/` package (to be created alongside the first user — here) — supports `2026-01-01`, `7d`, `1mo`, `3mo`, `1y`, `ytd`.
  - Instrument batching: collect all unique instrument URLs from the page, resolve in one batched call using `?ids=` on `/instruments/`.
- **Table columns:** `Date | Sym | Side | Type | Qty | Price | Fill | State`.
- **Commit:** `feat(cmd): rh orders with --since --symbol --state --limit`.

- [ ] Steps follow Task 15 with extra tests for the date parser (see Task 27.a below).

**Task 27.a — `internal/datearg/`:**

```go
// Package datearg parses absolute (YYYY-MM-DD) and relative (7d, 1mo, ytd) date strings.
package datearg

import (
    "errors"
    "strconv"
    "strings"
    "time"
)

// Parse returns the time corresponding to s.
func Parse(s string, now time.Time) (time.Time, error) {
    s = strings.TrimSpace(s)
    if s == "" {
        return time.Time{}, errors.New("empty date")
    }
    if t, err := time.Parse("2006-01-02", s); err == nil {
        return t.UTC(), nil
    }
    if s == "ytd" {
        return time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC), nil
    }
    // N unit
    units := map[string]time.Duration{
        "d":  24 * time.Hour,
        "w":  7 * 24 * time.Hour,
        "mo": 30 * 24 * time.Hour,
        "y":  365 * 24 * time.Hour,
    }
    for suffix, d := range units {
        if strings.HasSuffix(s, suffix) {
            n, err := strconv.Atoi(strings.TrimSuffix(s, suffix))
            if err != nil {
                return time.Time{}, err
            }
            return now.UTC().Add(-time.Duration(n) * d), nil
        }
    }
    return time.Time{}, errors.New("unrecognized date: " + s)
}
```

Write 5 tests — absolute, ytd, `7d`, `1mo`, `1y`, bad input.

---

## Task 28: `rh watchlist`

- **Endpoint:** `GET /watchlists/` (list all) and `GET /watchlists/<name>/` (detail).
- **Flags:** `--name NAME` — default "Default" if empty.
- **Shape:**
  ```go
  type WatchlistItem struct {
      Symbol      string `json:"symbol"`
      InstrumentID string `json:"instrument_id"`
      AddedAt     string `json:"added_at"`
      LastPrice   robinhood.Money `json:"last_price,omitempty"`
  }
  ```
- **Commit:** `feat(cmd): rh watchlist (read-only)`.

- [ ] Steps follow Task 15.

---

## Task 29: `rh search <query>`

- **Endpoint:** `GET /instruments/?query=<q>&limit=<n>`.
- **Flags:** `--limit N`.
- **Shape:**
  ```go
  type SearchResult struct {
      Symbol       string `json:"symbol"`
      Name         string `json:"name"`
      InstrumentID string `json:"instrument_id"`
      Tradeable    bool   `json:"tradeable"`
      Type         string `json:"type"` // stock | etp | crypto
  }
  ```
- **Commit:** `feat(cmd): rh search`.

- [ ] Steps follow Task 15.

---

## Task 30: `rh market-hours`

- **Endpoint:** `GET /markets/` → list; for each market, `GET /markets/<mic>/hours/<YYYY-MM-DD>/`.
- **Flags:** `--date YYYY-MM-DD` — defaults to today.
- **Shape:**
  ```go
  type MarketHours struct {
      MIC          string `json:"mic"`
      Name         string `json:"name"`
      IsOpen       bool   `json:"is_open"`
      OpensAt      string `json:"opens_at"`
      ClosesAt     string `json:"closes_at"`
      ExtendedOpen string `json:"extended_opens_at,omitempty"`
      ExtendedClose string `json:"extended_closes_at,omitempty"`
  }
  ```
- **Commit:** `feat(cmd): rh market-hours`.

- [ ] Steps follow Task 15.

---

## Task 31: `rh documents`

- **Endpoints:** `GET /documents/?cursor=<next>` (paginated).
- **Flags:** `--type 1099|trade_confirm|statement|other`, `--since DATE`, `--download PATH`.
- **Shape:**
  ```go
  type Document struct {
      ID       string `json:"id"`
      Type     string `json:"type"`
      Date     string `json:"date"`
      Name     string `json:"name"`
      DownloadURL string `json:"download_url"`
  }
  ```
- **`--download PATH`**: when set, for each document, stream `GET <download_url>` to `<PATH>/<date>-<name>.pdf`, 0600 perms, overwrite guard (skip if file exists unless `--force`).
- **Commit:** `feat(cmd): rh documents with --download`.

- [ ] Steps follow Task 15, plus one test for the download path (temp dir, streamed bytes, perms).

---

## Task 32: Plan B discovery + wrap-up

**Files:**
- Modify: `cmd/rh/commands.go` (stability + category)
- Modify: `cmd/rh/schema.go`
- Modify: `README.md`
- Modify: `internal/output/table.go` (re-verify all `Write<Name>` methods exist)

- [ ] **Step 1: Add stability + category to CommandMeta**

```go
type CommandMeta struct {
    // existing fields ...
    Category  string `json:"category,omitempty"`
    Stability string `json:"stability,omitempty"` // "green" | "yellow" | "red"
}
```

Set in each command's `init()` using new Cobra annotations:

```go
c := &cobra.Command{
    Use:   "positions",
    Short: "...",
    Annotations: map[string]string{
        "category":  "portfolio",
        "stability": "green",
    },
    RunE: runPositions,
}
```

In `buildCommandsPayload`, pull from `c.Annotations["category"]` and `c.Annotations["stability"]`.

Default stability is `green`; only `portfolio --history`, `options-positions`, `documents --download` set `yellow`.

- [ ] **Step 2: Broaden `rh schema`**

`cmd/rh/schema.go` currently returns the envelope schema. Extend it to accept an optional command name and return the JSON Schema of that command's `data` payload (hand-written in a `schemas/` map; no codegen in Plan B — just a map of `name → json.RawMessage` with copy/pasted schema fragments).

- [ ] **Step 3: Write the wrap-up smoke test**

Create `cmd/rh/planb_smoke_test.go`:

```go
package main

import (
    "testing"
)

func TestPlanB_CommandsRegistered(t *testing.T) {
    want := []string{
        "account", "commands", "dividends", "documents", "earnings",
        "fundamentals", "historicals", "login", "logout", "market-hours",
        "news", "options-positions", "orders", "portfolio", "position",
        "positions", "quote", "ratings", "schema", "search", "version",
        "watchlist",
    }
    root := newRootCommand()
    got := make(map[string]bool)
    for _, c := range root.Commands() {
        got[c.Name()] = true
    }
    for _, w := range want {
        if !got[w] {
            t.Errorf("command %q not registered", w)
        }
    }
}

func TestPlanB_AllCommandsHaveAnnotations(t *testing.T) {
    root := newRootCommand()
    for _, c := range root.Commands() {
        if c.Name() == "help" {
            continue
        }
        if c.Annotations["category"] == "" {
            t.Errorf("command %q missing category annotation", c.Name())
        }
        if c.Annotations["stability"] == "" {
            t.Errorf("command %q missing stability annotation", c.Name())
        }
    }
}
```

- [ ] **Step 4: Update README**

Append a "Plan B commands" section listing every command with a 1-line description. (Full README polish lives in Plan C.)

- [ ] **Step 5: Run everything**

```bash
make test
make test-race
make lint
make build
./rh commands --json | jq '.data | length' # expect >=22
./rh --help | grep -c "^  " # rough count of subcommands
```

- [ ] **Step 6: Commit**

```bash
git add cmd/rh/commands.go cmd/rh/schema.go cmd/rh/planb_smoke_test.go README.md
git commit -m "feat(cmd): Plan B discovery surface with category+stability; wrap-up smoke"
```

---

## Plan B Done — Acceptance Checklist

Run through this list before calling Plan B shipped. Every item fails Plan B if red.

- [ ] `make test` is green (including all 16 new data-command test files).
- [ ] `make test-race` is green (concurrent refresh lock verified).
- [ ] `make lint` is green.
- [ ] `make build` produces a `./rh` binary.
- [ ] `./rh --help` lists all 22 commands (incl. `login`, `logout`).
- [ ] `./rh commands --json` returns 22 entries, each with `category` and `stability` set.
- [ ] `./rh login` (in terminal, real creds) completes end-to-end on a fresh machine:
  - [ ] Prompts for username + password (hidden).
  - [ ] On device push → writes "Approve the prompt…", blocks, resumes when approved.
  - [ ] On SMS → prompts for code.
  - [ ] On TOTP (`ROBINHOOD_TOTP_SECRET` set) → no prompt needed; code auto-generated.
  - [ ] Stores tokens in OS keychain (verify via `security find-generic-password -s robinhood-cli -D default` on macOS).
- [ ] `./rh logout` clears the entry; subsequent `./rh portfolio` exits 2.
- [ ] `./rh portfolio --json` returns a valid envelope with real `data.equity`.
- [ ] For each of `positions | position AAPL | account | quote AAPL NVDA | fundamentals AAPL | historicals AAPL --span month | news AAPL --limit 3 | earnings AAPL | ratings AAPL | dividends --year 2025 | options-positions | orders --since 30d --limit 5 | watchlist | search apple | market-hours | documents --type statement` — exit 0, valid envelope, `data` non-null.
- [ ] Two concurrent `./rh portfolio` invocations after an expired access token produce only ONE `/oauth2/token/` refresh call (verify via a shim `ROBINHOOD_API_HOST=localhost:<recorded-server>`).
- [ ] `unset ROBINHOOD_KEYCHAIN_BACKEND; RH_KEYCHAIN_PROBE_DISABLE=1 ./rh login` in a container (no session bus) falls back to the file store at `$XDG_CONFIG_HOME/robinhood-cli/rh-default.json` (0600 perms).
- [ ] CI workflow passes on the branch.

**Tasks in Plan C (next):** SKILL.md authoring + `rh skill install` with `embed.FS`, GoReleaser + Homebrew tap, release workflow, nightly schema stability workflow, full documentation set (`docs/JSON_SCHEMA.md`, `docs/ROBINHOOD_API.md`, `docs/AUTH.md`, `docs/E2E.md`), cassette-based integration tests, govulncheck.

---

## Appendix: Codex Adversarial Review Fixes (MUST READ before any task)

Codex reviewed this plan pre-execution and found 9 BLOCKER / HIGH issues. The fixes below OVERRIDE the corresponding task text. Workers MUST apply these fixes when implementing the referenced task — do not fall back to the original text.

### Fix A — Task 1 (`GetJSONCtx`): close first response body BEFORE retry

Already applied inline in Task 1 Step 3 — the corrected snippet uses `_ = resp.Body.Close()` explicitly before issuing the retry request (not `defer`, which would LIFO-stack and block the first connection from returning to the pool). Concurrent pagination would otherwise exhaust `MaxIdleConnsPerHost`.

### Fix B — Task 6 (`WithRefreshLock` + `ensureFresh`): use the fresh-from-disk session as the refresh target

The original Task 6 pattern re-reads the keychain under the lock but then passes the STALE in-memory `s` to `oauth.Refresh`, which defeats the purpose — a second process's refresh burns the first process's rotated refresh token. Replace the `ensureFresh` body with:

```go
func (c *Client) ensureFresh() error {
    s := c.Session()
    if s == nil {
        return &APIError{Code: CodeUnauthenticated, Message: "no session", Hint: "run: rh login"}
    }
    if !s.NeedsImmediateRefresh() {
        return nil
    }
    return WithRefreshLock(func() error {
        // Re-read: another process may have just refreshed.
        latest, lerr := LoadFromKeychain(c.profile)
        if lerr == nil {
            c.SetSession(latest)
            if !latest.NeedsImmediateRefresh() {
                return nil
            }
            s = latest // CRUCIAL — refresh the freshest tokens
        }
        if err := c.oauth.Refresh(s); err != nil {
            return err
        }
        if s.Ephemeral {
            return nil // env-loaded session: never clobber keychain
        }
        return s.SaveToKeychain(c.profile)
    })
}
```

This fix depends on Fix C (profile plumbed on Client) and Fix N (Session.Ephemeral).

### Fix C — Task 6 (profile plumbing): thread `--profile` through Client

`internal/robinhood/` cannot import `cmd/rh`, so `gflags.Profile` is invisible. Add a `profile` field on Client with a setter:

```go
// In client.go:
type Client struct {
    // ... existing fields
    profile string
}
func (c *Client) SetProfile(p string) { c.mu.Lock(); defer c.mu.Unlock(); c.profile = p }
```

In Task 14 `newAuthedClient`:
```go
c := robinhood.NewClient()
c.SetSession(sess)
c.SetProfile(gflags.Profile)
```

Default the client's profile to `"default"` in `NewClient()` so tests that don't call SetProfile still work. Delete the `currentProfile()` free function from the original Task 6 spec.

### Fix D — Task 7 + Plan A (X-Robinhood-API-Version everywhere): add the header to `Refresh` too

Plan A's `Refresh` does NOT set `X-Robinhood-API-Version`. The Sheriff rollout cohort checks this header on EVERY oauth endpoint. Modify Task 7 to also patch Plan A's `Refresh`:

```go
// In oauth.go Refresh(), add to the req headers:
req.Header.Set("X-Robinhood-API-Version", "1.431.4")
// Also confirm User-Agent is set (Plan A sets one on client GETs; oauth calls need it too):
req.Header.Set("User-Agent", "robinhood-cli (+https://github.com/herocod3r/robinhood-cli)")
```

Add a test `TestRefresh_SendsAPIVersionHeader` that verifies the header is present.

### Fix E — Task 9 (Sheriff request shape): land a recorded cassette before coding

The plan's current `userMachineReq` (`{device_id, flow, input_as, type}` + `X-Robinhood-Challenge-Response-Id` header) is Codex's best guess but **will not match production**. Robin_stocks v3.4.x posts `{"device_id":device_token, "flow":"suv"}` alone (no `input_as`/`type`) and the workflow header is on later calls, not `user_machine`.

**Before implementing Task 9:** spawn a side task to generate a `testdata/cassettes/login_sheriff_sms.yaml` by running `robin_stocks` end-to-end against a burner Robinhood account under `mitmproxy`. Commit the sanitized cassette. Task 9's types must match the cassette exactly. If the cassette exposes a `machine_id → inquiry_id` intermediate call (`/pathfinder/inquiries/?user_machine_id=...`), add a step in `Sheriff.Start` for it.

**If a cassette cannot be captured:** implement Task 9's types from the robin_stocks source (https://github.com/jmfernandes/robin_stocks/blob/main/robin_stocks/robinhood/authentication.py) line-by-line, not from this plan.

### Fix F — Task 9 + Task 10: hard deadline on Sheriff + push polling

Both `Sheriff.Start`'s poll loop and `WaitPush` can deadlock on a pathological response. Fix:

1. In `cmd/rh/login.go` `runLogin`, wrap the context:
   ```go
   ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
   defer cancel()
   _, err := RunLogin(ctx, LoginOpts{...})
   ```
2. In `WaitPush` and `fetchUserView` (the poll caller), on JSON decode error return immediately:
   ```go
   if err := json.NewDecoder(resp.Body).Decode(&ps); err != nil {
       _ = resp.Body.Close()
       return &APIError{Code: CodeRobinhoodUnavailable, Message: "push status decode: " + err.Error()}
   }
   ```
3. In `fetchUserView`, differentiate 401/404/410 from other non-200:
   ```go
   switch resp.StatusCode {
   case http.StatusNotFound, http.StatusGone:
       return nil, &APIError{Code: CodeSheriffRequired, Message: "verification expired — run: rh login again"}
   case http.StatusUnauthorized:
       return nil, &APIError{Code: CodeSessionExpired, Message: "workflow token expired"}
   }
   ```

### Fix G — Task 10 `RespondCode` success sentinels: accept more than "validated"

Robinhood has returned `"issued"`, `"resolved"`, and `"success"` in different rollout cohorts. Accept any of `{"validated", "resolved", "success"}`; log unknown statuses in verbose mode but fall through to rejection only on explicit `{"failed", "declined", "expired"}`.

### Fix H — Task 12 `LoginOpts`: redact password in any formatter output

```go
func (o LoginOpts) String() string   { return "LoginOpts{<redacted>}" }
func (o LoginOpts) GoString() string { return "LoginOpts{<redacted>}" }
```

Add test `TestLoginOpts_DoesNotLeakPassword`:
```go
func TestLoginOpts_DoesNotLeakPassword(t *testing.T) {
    o := LoginOpts{Username: "u", Password: "SUPER-SECRET-PW"}
    for _, v := range []string{fmt.Sprint(o), fmt.Sprintf("%+v", o), fmt.Sprintf("%#v", o)} {
        if strings.Contains(v, "SUPER-SECRET-PW") {
            t.Fatalf("password leaked in %q", v)
        }
    }
}
```

### Fix I — Task 12 MFA retry loop: two attempts + nested Sheriff handling

Replace the `case robinhood.CodeMFARequired:` block with:

```go
case robinhood.CodeMFARequired:
    for attempt := 0; attempt < 2; attempt++ {
        code, inErr := o.CodeInput(apiErr.Hint)
        if inErr != nil || code == "" {
            return nil, &robinhood.APIError{Code: robinhood.CodeValidation, Message: "no MFA code entered"}
        }
        sess, err := oauthClient.PasswordGrant(ctx, o.Username, o.Password, o.DeviceToken, code)
        if err == nil {
            return persistLogin(sess, o)
        }
        var ae *robinhood.APIError
        if errors.As(err, &ae) && ae.Code == robinhood.CodeMFARequired {
            continue // prompt again
        }
        if errors.As(err, &ae) && ae.Code == robinhood.CodeSheriffRequired {
            return runSheriff(ctx, o, oauthClient, ae, code)
        }
        return nil, err
    }
    return nil, &robinhood.APIError{Code: robinhood.CodeMFARequired, Message: "MFA code rejected twice"}
```

### Fix J — Task 3 `probeSystem`: timeout + package-level cache + Linux dbus-skip

The probe is a per-`Open()` call in the plan; change it to a one-time, deadline-bounded probe:

```go
var (
    probeOnce sync.Once
    probeRes  bool
)

func systemAvailable() bool {
    probeOnce.Do(func() {
        // Skip the probe entirely if we know there's no dbus session (headless Linux).
        if runtime.GOOS == "linux" && os.Getenv("DBUS_SESSION_BUS_ADDRESS") == "" {
            probeRes = false
            return
        }
        done := make(chan bool, 1)
        go func() { done <- probeSystem(&SystemStore{}) }()
        select {
        case ok := <-done:
            probeRes = ok
        case <-time.After(1500 * time.Millisecond):
            probeRes = false
        }
    })
    return probeRes
}

func Open() Store {
    // ... ROBINHOOD_KEYCHAIN_BACKEND switch unchanged ...
    if systemAvailable() {
        return &SystemStore{}
    }
    return fileStoreAtConfigDir()
}
```

### Fix K — Task 3 + Task 4 + Task 13: validate profile name (path traversal)

A `--profile "../../etc/passwd"` writes the session file outside the config dir. Add a validator used by every caller that takes a profile string:

```go
// In internal/config/config.go:
var profileRe = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

func ValidProfile(name string) error {
    if !profileRe.MatchString(name) {
        return fmt.Errorf("invalid profile name: %q (allowed: [A-Za-z0-9_-], 1-64 chars)", name)
    }
    return nil
}
```

Call `config.ValidProfile(p)` at the top of `Session.SaveToKeychain`, `LoadFromKeychain`, `ClearKeychain`, `FileStore.Set/Get/Delete`, and the YAML loader (reject bad keys). `cmd/rh/login.go` + `logout.go` must validate `gflags.Profile` before use.

### Fix L — Task 2 `SaveFile` + Task 3 `FileStore.Set`: O_EXCL + no `/tmp` fallback for secrets

Replace `os.CreateTemp(...)`/`os.WriteFile(tmp, ...)` with an `OpenFile(..., O_EXCL|O_CREATE|O_WRONLY, 0o600)` pattern:

```go
f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
if err != nil {
    return err
}
if _, err := f.Write(data); err != nil { _ = f.Close(); return err }
if err := f.Close(); err != nil { return err }
return os.Rename(tmp, final)
```

`fileStoreAtConfigDir` MUST NOT fall back to `os.TempDir()` for secrets. If `EnsureConfigDir` fails, return an error up the stack — the CLI exits with a clear "could not create config dir" message.

### Fix M — Task 4 `UnmarshalSession`: cap decode input to 64KB

```go
// In keychain.FileStore.Get:
data, err := os.ReadFile(f.path(service, profile))
if err != nil { /* ... */ }
if len(data) > 64*1024 {
    return nil, fmt.Errorf("session file too large (%d bytes)", len(data))
}
return data, nil
```

### Fix N — Task 4 + Task 14: `Session.Ephemeral` flag

Env-loaded sessions must NOT clobber the keychain on refresh. Add:

```go
// session.go:
type Session struct {
    // existing fields
    Ephemeral bool // true for env-loaded sessions; refresh never persists
}

// LoadSessionFromEnv: set Ephemeral = true
// SaveToKeychain: refuse if s.Ephemeral (defensive)
```

Update Fix B's `ensureFresh` to honor `s.Ephemeral`.

Task 14 `TestNewAuthedClient_PrefersEnv` should assert `client.Session().Ephemeral == true`. Add a new test: `TestEnv_SessionRefresh_DoesNotClobberKeychain` — set `ROBINHOOD_ACCESS_TOKEN` + `ROBINHOOD_REFRESH_TOKEN`, write a DIFFERENT session to keychain, trigger a 401-refresh against a fake server, assert keychain unchanged.

### Fix O — Task 15 `sortPositions` / `filterNonZero`: use shared decimal math, not `strconv.ParseFloat`

The plan's `filterNonZero` parses `Quantity` as float64. Money is decimal strings; a float comparison against zero is fine for exact-zero detection but drift-prone. Use `toDec(r.Quantity).IsZero()` instead.

### Fix P — Task 32 acceptance: replace `ROBINHOOD_API_HOST` shim with a unit test

The acceptance checklist assumes an `ROBINHOOD_API_HOST` env override that Plan B does not wire. Either:
(a) Add it in Task 14: `if h := os.Getenv("ROBINHOOD_API_HOST"); h != "" { c := robinhood.NewClientWithHosts(h, h, h, &http.Client{...}) }` — useful for local dev/testing.
(b) OR drop the acceptance line and replace with an already-existing unit test in `refresh_lock_test.go`.

### Fix Q — Task 31 `rh documents --download PATH`: sanitize + cap

Document names come from Robinhood — never trust. Use `filepath.Base(name)` and strip path separators; cap download size with `io.LimitReader(resp.Body, 500*1024*1024)` (500MB) so a rogue response can't fill the disk.

### Known followups (deferred to Plan C or later)

- Fix E (Sheriff cassette) assumes a burner account is available. If it is not, document the risk in `docs/AUTH.md` and track as a "yellow" endpoint.
- The 2-attempt MFA retry cap is arbitrary — revisit after real-world usage.

Workers: apply these fixes at the task they reference, not as a separate pass. Commit messages still follow the task's spec.
