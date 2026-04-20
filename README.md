# rh — Robinhood CLI

Open-source, read-only Robinhood CLI for portfolio research and trade
analysis. Ships with a Claude Code / Codex skill so your AI assistant can
answer questions about your portfolio.

> **⚠️ Research tool only.** `rh` is read-only — it never trades, moves
> cash, or mutates watchlists. It uses Robinhood's unofficial API and may
> break without notice.

## Status

**v1.0.0** — all 17 data commands, interactive login, skill install,
multi-profile, GoReleaser pipeline, nightly schema drift detection.
See [`CHANGELOG.md`](CHANGELOG.md).

## Install

### Homebrew (macOS, Linux)

```bash
brew install herocod3r/tap/rh
```

### `go install`

```bash
go install github.com/herocod3r/robinhood-cli/cmd/rh@latest
```

### GitHub releases

Download the archive for your platform from the
[Releases page](https://github.com/herocod3r/robinhood-cli/releases).
Each archive contains `rh`, `LICENSE`, `README.md`, `CHANGELOG.md`,
and the `skills/` tree.

## Quick start

```bash
rh login                     # interactive Sheriff / MFA flow
rh portfolio --json
rh positions --json
rh quote AAPL NVDA --json
```

All data commands accept `--json` (stable shape — see
[`docs/JSON_SCHEMA.md`](docs/JSON_SCHEMA.md)) and a default human table.

## Skills

`rh` ships a [Claude Code](https://docs.claude.com/claude-code) / Codex
skill bundle so your AI assistant can drive the CLI with tested
workflows. Install with one command:

```bash
rh skill install --target claude-code     # ~/.claude/skills/robinhood-cli/
rh skill install --target codex           # ~/.codex/skills/robinhood-cli/
rh skill install --target both            # both at once
```

Start a fresh assistant session after install. Ask "how's my portfolio?"
and the skill will run the right commands and narrate the output.

## Shell completions

Cobra ships completions out of the box:

```bash
# bash
rh completion bash > /etc/bash_completion.d/rh

# zsh
rh completion zsh > "${fpath[1]}/_rh"

# fish
rh completion fish > ~/.config/fish/completions/rh.fish
```

## Environment variables

| Var | Purpose |
|---|---|
| `ROBINHOOD_USERNAME` | Skip the username prompt on `rh login` |
| `ROBINHOOD_PASSWORD` | Skip the password prompt (headless) |
| `ROBINHOOD_TOTP_SECRET` | Base32 shared secret; `rh` generates the code |
| `ROBINHOOD_DEVICE_TOKEN` | Reuse a known device token across machines |
| `ROBINHOOD_ACCESS_TOKEN` | Bypass keychain; use this token directly |
| `ROBINHOOD_REFRESH_TOKEN` | Bypass keychain; use this refresh token |
| `ROBINHOOD_PROFILE` | Default profile (same as `--profile`) |
| `ROBINHOOD_KEYCHAIN_BACKEND` | `system` (force OS keychain) or `file` (force file fallback) |
| `ROBINHOOD_API_HOST` | Override all three API hosts — handy for mitmproxy / local dev |

See [`docs/AUTH.md`](docs/AUTH.md) for storage details, precedence
(`env > keychain > fail`), and the full auth state machine.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | OK |
| `1` | Runtime or network error (`robinhood_unavailable`, `not_found`, other) |
| `2` | Not authenticated / session expired (`unauthenticated`, `session_expired`) |
| `3` | Interactive step required (`sheriff_required`, `mfa_required`) |
| `4` | Rate limited (`rate_limited`) |
| `5` | Bad flags / args (`validation`) |

Mapping lives in `internal/robinhood/errors.go:APIError.ExitCode` — the
JSON envelope's `error.code` always carries the string form.

## Documentation

- [`docs/AUTH.md`](docs/AUTH.md) — auth state machine, storage, env vars, troubleshooting.
- [`docs/JSON_SCHEMA.md`](docs/JSON_SCHEMA.md) — JSON Schema for every command's `data` payload.
- [`docs/ROBINHOOD_API.md`](docs/ROBINHOOD_API.md) — endpoint inventory + stability ratings.
- [`docs/E2E.md`](docs/E2E.md) — pre-release QA checklist.
- [`docs/release-process.md`](docs/release-process.md) — maintainer release runbook.
- [`SECURITY.md`](SECURITY.md) — reporting policy, scope, secure defaults.
- Design specs + plans live under `docs/superpowers/`.

## Development

```bash
make test            # unit tests
make test-race       # with race detector
make test-fixtures   # cassette-based integration (Tier 2)
make test-contract   # JSON-schema contract tests (Tier 3)
make lint            # golangci-lint (errcheck, govet, staticcheck, gosec, revive, ...)
make build           # local binary
```

See [`CONTRIBUTING.md`](CONTRIBUTING.md) for the full dev setup, coding
conventions, and how to add a new command.

## License

[MIT](LICENSE)
