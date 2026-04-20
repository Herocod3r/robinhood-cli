# robinhood-cli — Design Spec

**Date:** 2026-04-18
**Status:** Draft, pre-implementation
**Authors:** herocod3r (user), brainstormed with Claude Code
**Repo:** `herocod3r/robinhood-cli`
**License:** MIT

## 1. Purpose and scope

Build an open-source Go CLI (`rh`) for read-only Robinhood portfolio research, plus a Claude Code + Codex skill that drives it. The CLI is human-usable and skill-usable; the skill is a thin prompt layer that shells out to `rh --json`.

**v1 is read-only.** Trading (v2) is out of scope but design choices preserve the path:
- The client layer stays pure Go with no hard-coded read assumption
- Command tree is flat and extensible
- Auth state machine already handles the surface writes would require
- JSON schema versioning allows additive evolution

**Explicitly not v1:**
- Writes of any kind (orders, watchlist add/remove, ACH)
- MCP server (the reference [verygoodplugins/robinhood-mcp](https://github.com/verygoodplugins/robinhood-mcp) already fills that niche)
- TUI / interactive dashboard
- Published Go library (`internal/` only until API stabilizes)
- Web UI, daemon, background polling

## 2. Reference and research

- Reference MCP: [verygoodplugins/robinhood-mcp](https://github.com/verygoodplugins/robinhood-mcp) — Python, wraps `robin_stocks`, 13 read-only tools
- Python reference client: [jmfernandes/robin_stocks](https://github.com/jmfernandes/robin_stocks) — source of truth for current Robinhood API behavior
- Private API docs (stale but useful): [sanko/Robinhood](https://github.com/sanko/Robinhood)
- Most complete existing Go client: [andrewstuart/go-robinhood](https://github.com/andrewstuart/go-robinhood) — **does not implement the mandatory Sheriff workflow introduced Feb 2025**. We will not depend on it.

Full endpoint inventory in `docs/ROBINHOOD_API.md` (to be written alongside implementation).

## 3. Key constraints derived from research

- **Sheriff/SUV challenge workflow is mandatory** for new-device logins since Feb 2025 (`robin_stocks` 3.4.0, commit `dfcf055`)
- **Stable `device_token`** per device is required; rotating tokens triggers 400s
- **Session persistence effectively required** — `store_session=False` path is broken as of Jan 2026 (robin_stocks #1643)
- **Refresh-token-first auth**: fresh password grant is increasingly unreliable (robin_stocks #1640, Dec 2025)
- **No real-time streaming, no L2 book, no tick data, no tax lots** — these are not available in the unofficial API regardless of implementation
- **Some endpoints are yellow-stable** (portfolios/historicals, options-order history). We document stability per-command

## 4. Architecture

Three layers + one sidecar (the skill):

```
┌─────────────────────────────────────────────────────────┐
│  skills/  (Claude Code skill + Codex skill)             │
│  shells out to `rh`, reads JSON, summarizes for user    │
└─────────────────────────────────────────────────────────┘
                          │  exec
                          ▼
┌──────────────────┬──────────────────┬──────────────────┐
│  cmd/rh          │  internal/cli    │  internal/output │
│  cobra commands  │  flag parsing    │  pretty / JSON   │
│  one file/cmd    │  shared ctx      │  isatty switch   │
└──────────────────┴──────────────────┴──────────────────┘
                          │
                          ▼
┌──────────────────────────────────────────────────────────┐
│  internal/robinhood  — our own HTTP client               │
│  • oauth.go    (login, Sheriff, refresh, device token)   │
│  • session.go  (keychain store + file-lock refresh)      │
│  • client.go   (typed calls, 401→refresh→retry)          │
│  • types.go    (response structs)                        │
│  • endpoints/  (one file per domain)                     │
└──────────────────────────────────────────────────────────┘
                          │
                          ▼
    api.robinhood.com, nummus.robinhood.com,
    phoenix.robinhood.com
```

**Principles:**
- `internal/robinhood` is usable as a Go package (within the module) — `cmd/rh` is a thin caller. This preserves the v2 path (trading CLI, MCP server, etc.).
- Each command = one Cobra file + one client call + one output formatter. No business logic in `cmd/rh`.
- Output layer is the only code aware of terminals. The client returns structs; `internal/output` renders.
- The skill is pure Markdown — no code, no MCP server. It tells the model when and how to shell out.

## 5. Project layout

```
robinhood-cli/
├── cmd/rh/
│   ├── main.go              # wires Cobra root + subcommands
│   ├── login.go             # rh login (interactive Sheriff)
│   ├── logout.go
│   ├── version.go           # rh version --json (compat check)
│   ├── commands.go          # rh commands --json (discovery)
│   ├── schema.go            # rh schema [cmd] --json
│   ├── skill.go             # rh skill install --target
│   ├── portfolio.go
│   ├── positions.go
│   ├── position.go
│   ├── account.go
│   ├── quote.go
│   ├── fundamentals.go
│   ├── historicals.go
│   ├── news.go
│   ├── earnings.go
│   ├── ratings.go
│   ├── dividends.go
│   ├── options_positions.go
│   ├── orders.go
│   ├── watchlist.go
│   ├── search.go
│   ├── market_hours.go
│   └── documents.go
│
├── internal/
│   ├── robinhood/
│   │   ├── client.go
│   │   ├── oauth.go
│   │   ├── session.go
│   │   ├── errors.go
│   │   ├── types.go
│   │   └── endpoints/
│   │       ├── portfolio.go
│   │       ├── positions.go
│   │       ├── quotes.go
│   │       ├── fundamentals.go
│   │       ├── historicals.go
│   │       ├── news.go
│   │       ├── earnings.go
│   │       ├── ratings.go
│   │       ├── dividends.go
│   │       ├── options.go
│   │       ├── orders.go
│   │       ├── watchlist.go
│   │       ├── instruments.go
│   │       ├── account.go
│   │       └── documents.go
│   ├── config/
│   ├── output/
│   │   ├── output.go        # Writer interface; TTY detection
│   │   ├── json.go          # stable JSON envelope
│   │   └── table.go         # tables (lipgloss or tablewriter)
│   └── keychain/            # zalando/go-keyring wrapper + file fallback
│
├── skills/
│   ├── claude-code/robinhood-cli/
│   │   ├── SKILL.md
│   │   ├── references/
│   │   └── examples/
│   └── codex/robinhood-cli/ # synced copy
│
├── docs/
│   ├── AUTH.md
│   ├── JSON_SCHEMA.md
│   ├── ROBINHOOD_API.md
│   ├── E2E.md
│   └── superpowers/specs/   # this doc lives here
│
├── testdata/cassettes/      # sanitized go-vcr fixtures
│
├── .goreleaser.yaml
├── .github/workflows/
│   ├── ci.yaml
│   ├── release.yaml
│   └── schema.yaml          # nightly schema stability
├── Makefile
├── go.mod
├── LICENSE                  # MIT
├── README.md
├── CHANGELOG.md
├── CONTRIBUTING.md
└── SECURITY.md
```

**Key choices:**
- Single repo, CLI + skills together. One release cycle, skill versions pin to CLI versions.
- `internal/` (not `pkg/`) until API stabilizes.
- One file per command — grep-friendly, contributor-friendly.
- Endpoints grouped by API domain, not per-command — keeps struct definitions near their producers.

## 6. Authentication

Two flows: interactive login (rare) and silent refresh (every invocation).

### 6.1 Interactive — `rh login`

```
1. Prompt username + password (stdin, password hidden)
2. Generate device_token (UUIDv4) if none in keychain
3. POST /oauth2/token/ {grant_type: password, device_token, …}

Branch on response:
  (a) 200 + tokens
      → store in keychain; print "Logged in as <email>"
  (b) 400 + {verification_workflow: {id, workflow_status}}
      → Sheriff flow (below)
  (c) 400 + {mfa_required: true}
      → prompt TOTP (or read $ROBINHOOD_TOTP_SECRET; generate);
        retry with mfa_code
  (d) other 4xx
      → parsed error + docs link; exit 1

Sheriff flow (branch b):
  1. POST /pathfinder/user_machine/ {device_id: device_token, …}
     → returns inquiry_id
  2. Poll GET /pathfinder/inquiries/{id}/user_view/ every 2s
     → exposes one of: sms, email, device_approval (push)
  3. Show what's being requested:
     "Robinhood sent a push to your device. Approve to continue." [spinner]
     or "Robinhood sent a code to ***-***-1234. Enter it: "
  4. For SMS/email: POST /challenge/{id}/respond/ {response: code}
     For push: continue polling /get_prompts_status/
  5. On challenge success: re-POST /oauth2/token/ with workflow header
     → receive tokens → store
```

### 6.2 Silent — every command

```
1. Load session from keychain (or env override)
   → missing → exit 2 "run: rh login"
2. Check access_token expiry
   ├ >5min remaining: use as-is
   └ <5min: → Refresh
3. Make API call
   ├ 200: render, exit 0
   ├ 401: Refresh, retry ONCE, if 401 again → exit 2
   └ other: typed error → render → exit 1/4/5

Refresh (file-locked to prevent races):
  flock(session.lock)
  POST /oauth2/token/ {grant_type: refresh_token, …}
  ├ 200: update keychain; return
  └ 4xx: exit 2 "session expired — run: rh login"
  funlock()
```

### 6.3 Storage

- **Keychain** (one item per profile; default profile `"default"`):
  `access_token`, `refresh_token`, `expires_at`, `device_token`, `username`
- **Config file** `~/.config/robinhood-cli/config.yaml`:
  `default_profile`, `profiles: {default: {account_number: …}}`
- **Lock file** `~/.config/robinhood-cli/session.lock` (flock only; no data)

### 6.4 Env-var override (skill, CI, headless)

- `ROBINHOOD_USERNAME`, `ROBINHOOD_PASSWORD`, `ROBINHOOD_TOTP_SECRET`, `ROBINHOOD_DEVICE_TOKEN`
- `ROBINHOOD_ACCESS_TOKEN` + `ROBINHOOD_REFRESH_TOKEN` (bypass keychain; container use)
- Precedence: env vars > keychain > fail

### 6.5 Error taxonomy (JSON `error.code`)

| Code | Exit | Meaning |
|---|---|---|
| `unauthenticated` | 2 | No session |
| `session_expired` | 2 | Refresh failed |
| `sheriff_required` | 3 | Interactive step needed |
| `rate_limited` | 4 | 429 from Robinhood |
| `robinhood_unavailable` | 1 | 5xx |
| `not_found` | 1 | Missing symbol/order/etc |
| `validation` | 5 | Bad input |

## 7. Command tree

Flat tree. Cobra aliases for short forms. Global flags: `--json`, `--pretty`, `--profile`, `--no-color`, `-v/--verbose`, `--help`.

| Command | Alias | Args | Key flags | Notes |
|---|---|---|---|---|
| `rh login` | — | — | `--totp-secret`, `--username` | Interactive by default |
| `rh logout` | — | — | `--all-profiles` | Clears keychain |
| `rh version` | — | — | — | Emits schema version |
| `rh commands` | — | `[name]` | — | Discovery (skill uses this) |
| `rh schema` | — | `[cmd]` | — | Emits JSON schema |
| `rh skill install` | — | — | `--target claude-code\|codex\|both` | |
| `rh portfolio` | `pf` | — | `--history <span>` | Yellow on `--history` (portfolios/historicals) |
| `rh positions` | `ps` | — | `--nonzero`, `--sort value\|pl\|symbol`, `--limit N` | |
| `rh position` | `p` | `<ticker>` | — | Faster than `positions\|grep` |
| `rh account` | `acct` | — | — | Buying power, margin, PDT, sweep |
| `rh quote` | `q` | `<ticker...>` | `--extended-hours` | Batch up to 50 symbols |
| `rh fundamentals` | `f` | `<ticker...>` | — | Batch |
| `rh historicals` | `h` | `<ticker>` | `--interval`, `--span` | OHLCV |
| `rh news` | `n` | `<ticker>` | `--limit N` | |
| `rh earnings` | `e` | `<ticker>` | — | |
| `rh ratings` | `r` | `<ticker>` | — | |
| `rh dividends` | `div` | — | `--year YYYY`, `--symbol SYM` | |
| `rh options-positions` | `opts` | — | `--symbol SYM` | |
| `rh orders` | `o` | — | `--since`, `--symbol`, `--state`, `--limit` | **Core of "analyze my trades"** |
| `rh watchlist` | `wl` | — | `--name NAME` | Read-only in v1 |
| `rh search` | `s` | `<query>` | `--limit N` | |
| `rh market-hours` | `mh` | — | `--date YYYY-MM-DD` | |
| `rh documents` | `docs` | — | `--type`, `--since`, `--download PATH` | |

**Conventions:**
- Tickers case-insensitive on input, uppercase in output
- Dates accept `YYYY-MM-DD` or relative (`7d`, `1mo`, `ytd`, `1y`)
- Batch commands accept ≤50 symbols (Robinhood limit)
- Yellow-stable endpoints print a stderr warning in `--verbose` mode

**Not in v1:** `rh config`, `rh profile`, `rh export`. Hand-edit YAML, use `--profile`, pipe JSON through `jq`.

## 8. Output format

TTY-aware auto-switch (`gh`/`kubectl` model):

- `rh positions` (TTY) → pretty table
- `rh positions | jq …` (pipe) → JSON auto
- `--json` / `--pretty` override

### 8.1 JSON envelope

```json
{
  "schema": "robinhood-cli/v1",
  "command": "positions",
  "generated_at": "2026-04-18T18:42:03Z",
  "data": <command-specific>,
  "meta": { "count": 23, "profile": "default" },
  "error": null
}
```

On error:

```json
{
  "schema": "robinhood-cli/v1",
  "command": "positions",
  "generated_at": "2026-04-18T18:42:03Z",
  "data": null,
  "meta": null,
  "error": {
    "code": "session_expired",
    "message": "refresh token rejected",
    "hint": "run: rh login",
    "retryable": false
  }
}
```

### 8.2 Type conventions

- **Money**: decimal strings (`"123.4500"`), never floats; preserves Robinhood's precision
- **Percents**: decimal strings (`"-0.0234"` = -2.34%); never percent-string mix
- **Quantities**: decimal strings (fractional shares)
- **Timestamps**: ISO-8601 UTC with `Z` (no zone offsets)

### 8.3 Representative schemas (abbreviated)

```jsonc
// rh portfolio
{
  "equity": "98234.50",
  "extended_hours_equity": "98240.12",
  "market_value": "72450.10",
  "cash": "25784.40",
  "buying_power": "25784.40",
  "day_change": "342.15",
  "day_change_percent": "0.0035",
  "total_return": "12450.00",
  "total_return_percent": "0.1449"
}

// rh positions — data is array
[
  {
    "symbol": "AAPL",
    "name": "Apple Inc.",
    "quantity": "10.0000",
    "average_buy_price": "152.3400",
    "last_price": "195.2000",
    "market_value": "1952.00",
    "cost_basis": "1523.40",
    "unrealized_pl": "428.60",
    "unrealized_pl_percent": "0.2814",
    "day_change": "12.40",
    "day_change_percent": "0.0064",
    "instrument_id": "450dfc6d-…",
    "account": "default"
  }
]

// rh orders
[
  {
    "id": "8f2d…",
    "symbol": "NVDA",
    "side": "buy",
    "type": "limit",
    "state": "filled",
    "quantity": "5.0000",
    "price": "110.2500",
    "average_fill_price": "109.8700",
    "fees": "0.00",
    "time_in_force": "gfd",
    "extended_hours": false,
    "created_at": "2026-03-14T14:22:08Z",
    "filled_at": "2026-03-14T14:22:11Z"
  }
]

// rh quote
{
  "symbol": "AAPL",
  "last_price": "195.20",
  "bid_price": "195.18",
  "ask_price": "195.22",
  "bid_size": 100,
  "ask_size": 200,
  "previous_close": "192.80",
  "day_change": "2.40",
  "day_change_percent": "0.0124",
  "extended_hours_price": "195.60",
  "volume": 48235100,
  "updated_at": "2026-04-18T18:42:01Z"
}
```

Full schemas: `docs/JSON_SCHEMA.md`.

### 8.4 Versioning

- `schema: "robinhood-cli/v1"` stays across all v1.x.y CLI releases
- Additive changes (new fields) allowed in minor releases
- Renames/removals/type changes require `v2` bump and deprecation period
- Per-field deprecation: ship replacement alongside old field for one minor cycle, remove on next major

### 8.5 What we don't do

- No pagination envelope — we unfold internally and return full arrays (flag `meta.truncated: true` only if >10k items)
- No HATEOAS links
- No embedded related objects — `positions` does not include full quote data; caller can batch `rh quote` if needed

## 9. Discovery surface

The skill uses CLI-side discovery so SKILL.md stays small and doesn't rot.

| Command | Purpose |
|---|---|
| `rh version --json` | Compat check: `{"version":"…","schema":"robinhood-cli/v1"}` |
| `rh commands --json` | All commands, summaries, flags, args, stability, category |
| `rh commands <name> --json` | One command, full flag spec |
| `rh schema [cmd] --json` | JSON schema for output data |
| `rh <cmd> --help` / `rh help <cmd>` | Plain-text help (Cobra default) |

**`rh commands --json` shape:**

```json
{
  "schema": "robinhood-cli/v1",
  "command": "commands",
  "data": [
    {
      "name": "orders",
      "aliases": ["o"],
      "summary": "List order history",
      "category": "trades",
      "args": [],
      "flags": [
        {"name": "since", "type": "date", "default": "", "description": "YYYY-MM-DD or relative"},
        {"name": "symbol", "type": "string", "default": ""},
        {"name": "state", "type": "enum", "values": ["filled","cancelled","open"], "default": ""},
        {"name": "limit", "type": "int", "default": 0}
      ],
      "schema_ref": "schemas/orders",
      "stability": "green"
    }
  ]
}
```

This surface is generated from Cobra's command tree via reflection — no hand-maintained table.

## 10. Skill design

**One skill, two install targets.** Same `SKILL.md` ships to `skills/claude-code/robinhood-cli/` and `skills/codex/robinhood-cli/`. A build script copies from a single source of truth at release time.

### 10.1 `SKILL.md` structure

```markdown
---
name: robinhood-cli
description: Use when the user asks about their Robinhood portfolio, positions,
  trades, dividends, or investment performance. Shells out to the `rh` CLI
  (read-only) and analyzes JSON output. Do NOT activate for general stock-market
  questions unless the user explicitly asks about their holdings.
---

# robinhood-cli

You have access to the `rh` CLI. Read-only — you cannot place trades.
Always invoke with `--json` and parse stdout. Check exit code first.

## On first use per session

1. `rh version --json` — confirm `schema: "robinhood-cli/v1"`
2. `rh commands --json` — discover available commands, flags, stability

## Invariants

1. Never run `rh login` yourself. On exit code 2 or 3, tell the user to run
   it in their terminal and stop.
2. Batch quotes — `rh quote A B C` is one call
3. Prefer `rh position <sym>` over `rh positions | filter`
4. Do not cache between turns; re-invoke
5. Money is strings; compare as decimals, not floats

## Exit codes

| Code | Meaning | Action |
|---|---|---|
| 0 | OK | Parse `.data`, answer |
| 2 | Not logged in / expired | Tell user: `run: rh login`. Stop. |
| 3 | Sheriff required | Same as 2. Stop. |
| 4 | Rate limited | Wait `.error.hint` interval, retry once |
| 5 | Bad args | Re-read `rh <cmd> --help` |

## Workflows (guidance — CLI is authoritative via `rh commands`)

### "How's my portfolio?"
    rh portfolio --json
    rh positions --nonzero --sort value --json
Compute sector concentration; top gainers/losers; flag >20% single-position.

### "Analyze my trades this year"
    rh orders --since 2026-01-01 --state filled --json
Group by symbol; compute realized P&L (see references/trade-analysis.md).
Flag wash sales as a caveat ("not computed").

### "Should I add more <TICKER>?"
    rh position TICKER --json
    rh fundamentals TICKER --json
    rh quote TICKER --json
    rh news TICKER --limit 5 --json
Summarize. Do not give buy/sell advice.

### "Compare <A> <B> <C>"
    rh fundamentals A B C --json
    rh quote A B C --json
Side-by-side table.

### "What dividends have I received?"
    rh dividends --year 2026 --json
Group, sum, top-5 contributors.

## When NOT to shell out

- General market questions → general knowledge
- Stocks user doesn't hold, general research → only if user asks for data
- If user hasn't said this is their account → ask first

## Presenting

- Answer first, commands second
- Brief breakdown tables; reserve long JSON for user request
- Cite `generated_at` timestamp when price-sensitive
- Note material divergence between `extended_hours_price` and `last_price`

## References

- `references/trade-analysis.md` — buy/sell lot pairing, realized P&L
- `references/sector-concentration.md`
- Full output schema: emit `rh schema --json`; authoritative doc in the CLI repo at `docs/JSON_SCHEMA.md`
```

### 10.2 Skill directory

```
skills/claude-code/robinhood-cli/
  SKILL.md
  references/
    trade-analysis.md
    sector-concentration.md
    output-schema.md
  examples/
    portfolio-healthcheck.md
    trade-analysis.md
```

Note: `rh skill` is a Cobra group with `install` as a subcommand; all other commands stay flat. Codex skill copy is identical in content (generated from the same source by `scripts/sync-skill.sh`).

### 10.3 Installation

Users don't need to clone the repo. `rh skill install --target claude-code|codex|both` writes the skill from the binary's embedded `skills/` (via `embed.FS`) to the right platform path.

## 11. Distribution and release

### 11.1 GoReleaser

- Cross-compile darwin/linux/windows × amd64/arm64
- tar.gz (zip for Windows) archives include `LICENSE`, `README.md`, `skills/**`
- Homebrew tap at `herocod3r/homebrew-tap` (PR automation via GoReleaser)
- `ldflags` inject version + commit for `rh version`

### 11.2 Install paths, by friction

1. `brew install herocod3r/tap/rh`
2. `go install github.com/herocod3r/robinhood-cli/cmd/rh@latest`
3. GitHub release archives (Windows and Brew-less Linux)
4. `curl | sh` installer — defer until v1.1 justifies scripting effort

### 11.3 GitHub Actions

- **`ci.yaml`** (PR/push): `go vet`, `golangci-lint`, `go test ./...`, `govulncheck`, `goreleaser check`
- **`release.yaml`** (`v*.*.*` tag): GoReleaser → GitHub release → Homebrew tap PR
- **`schema.yaml`** (nightly): runs `rh commands --json` on HEAD vs latest release; fails on field renames/removals

### 11.4 Versioning

- SemVer for the CLI
- JSON schema version decoupled (`robinhood-cli/v1` stays across all v1.x.y CLI releases)
- Only breaking schema changes bump the schema namespace to `v2`

### 11.5 Release cadence

Ship when ready. No calendar pressure.

### 11.6 Security

- `SECURITY.md` with disclosure email; 72h response target
- No bug bounty in v1

## 12. Testing strategy

Four tiers with clear ownership.

### 12.1 Tier 1 — unit (fast, hermetic)

- `httptest.Server` per test
- Every endpoint: happy, 4xx, 5xx, malformed, missing fields
- Auth state machine: each Sheriff branch (sms/email/push) and MFA/TOTP
- Token refresh: expired access, valid refresh, invalid refresh, concurrent race
- Output formatters: snapshot-compare on JSON and table rendering

Target: every file in `internal/robinhood/endpoints/` has ≥1 test; every branch in `oauth.go` has ≥1 test.

### 12.2 Tier 2 — recorded fixtures (drift detection)

- `dnaeon/go-vcr` cassettes in `testdata/cassettes/`
- One cassette per endpoint; refreshed quarterly by a maintainer
- Sanitizer strips tokens, account numbers, SSN-like patterns, names, emails
- CI runs a secrets scanner on cassettes (e.g., `gitleaks`)

### 12.3 Tier 3 — JSON schema contract

- Generated outputs diffed against `docs/JSON_SCHEMA.md`
- Additions pass; renames/removals fail unless schema version bumps
- Runs nightly (see `schema.yaml`)

### 12.4 Tier 4 — E2E with real auth (manual, gated)

- `make e2e` or `workflow_dispatch` action with encrypted test account
- Checklist in `docs/E2E.md`: fresh login → Sheriff → TOTP → first command → logout → refresh-path login → 5 commands → concurrent invocations → expired-token recovery
- Maintainer runs before every release; results in release PR

### 12.5 Tooling

- `make test` — Tier 1
- `make test-fixtures` — Tier 1 + 2
- `make test-contract` — Tier 3
- `make e2e` — Tier 4
- `golangci-lint`: `errcheck`, `govet`, `staticcheck`, `gosec`, `revive`
- Pre-commit: `go vet`, `gofmt`, fast tests

### 12.6 Not in v1

- Fuzzing, mutation testing, property-based tests, load tests

## 13. Open questions (track and resolve during implementation)

1. **`rh portfolio --history` default span** — `day` or `week`? Pick during implementation based on endpoint behavior.
2. **Windows keychain parity** — `go-keyring` uses `wincred`; verify it handles >2KB tokens (there's a known-ish limit).
3. **Cobra vs `urfave/cli`** — spec assumes Cobra (mature, common). Revisit if Cobra pain emerges.
4. **`rh skill install` over-write behavior** — prompt? Force flag? Defer to plan.
5. **Cassette sanitation policy** — exact regex list needs review before first cassette commit.

## 14. Out of scope for v1 (explicit)

- Any write operation (orders, cancels, watchlist mutation, ACH)
- Real-time streaming (not available in Robinhood's unofficial API)
- Tax-lot accounting (not available via API)
- MCP server mode
- GUI/TUI
- Non-English locales
- Options chains explorer (we expose options-positions; full chain exploration is v2)

## 15. Success criteria for v1

1. `rh login` works on a fresh device without prior setup (full Sheriff flow)
2. All 17 read commands return structurally-correct JSON matching `docs/JSON_SCHEMA.md`
3. `rh` survives a week of daily use without re-login (refresh path works)
4. Skill installed on a fresh Claude Code session answers the five workflows in Section 10 without hand-holding
5. GoReleaser tags a release and Homebrew install succeeds on macOS and Linux
6. Tier 1 tests pass on CI for every PR
