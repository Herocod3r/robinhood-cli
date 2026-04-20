# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org).

## [Unreleased]

## [1.0.0] - 2026-04-XX

### Added

- `rh login` with Sheriff / SUV interactive workflow (SMS, email, push)
  and TOTP MFA support.
- `rh logout` with `--all-profiles` to clear every stored session.
- 17 read-only data commands: `portfolio`, `positions`, `position`,
  `account`, `quote`, `fundamentals`, `historicals`, `news`,
  `earnings`, `ratings`, `dividends`, `options-positions`, `orders`,
  `watchlist`, `search`, `market-hours`, `documents`.
- Discovery commands: `rh version`, `rh commands`, `rh schema`.
- `rh skill install --target claude-code | codex | both` — embeds the
  skill bundle in the binary and copies it into the target's skills
  directory.
- Stable JSON envelope (`schema: "robinhood-cli/v1"`) with a 7-code
  error taxonomy (`unauthenticated`, `session_expired`,
  `sheriff_required`, `mfa_required`, `rate_limited`,
  `robinhood_unavailable`, `not_found`, `validation`) and exit-code
  mapping (2 = unauth/expired, 3 = sheriff/mfa, 4 = rate, 5 =
  validation, 1 = everything else).
- OS keychain storage (macOS Keychain, GNOME libsecret, Windows
  Credential Manager) with a `0600` JSON fallback under
  `$XDG_CONFIG_HOME/robinhood-cli/`.
- File-locked refresh so concurrent `rh` processes never race the
  same refresh token.
- Multi-profile support via `--profile` / `ROBINHOOD_PROFILE`.
- GoReleaser pipeline producing darwin / linux / windows archives for
  `amd64` and `arm64`, a Homebrew tap formula, and checksums.
- GitHub Actions: `ci.yaml` (vet + lint + test + govulncheck),
  `schema.yaml` (nightly schema-stability diff), `release.yaml`
  (tag-driven), `cassettes.yaml` (gitleaks scan on cassette changes).
- Cassette-based integration tests (Tier 2) under `testdata/cassettes/`.
- JSON-schema contract tests (Tier 3) that validate golden outputs
  against `docs/JSON_SCHEMA.md`.

### Documentation

- `docs/AUTH.md` — authentication state machine, storage, env-vars,
  error taxonomy, troubleshooting.
- `docs/JSON_SCHEMA.md` — JSON Schema Draft 2020-12 for every command's
  `data` payload.
- `docs/ROBINHOOD_API.md` — endpoint inventory with per-endpoint
  stability ratings.
- `docs/E2E.md` — the pre-release manual QA checklist.
- `docs/release-process.md` — the maintainer release runbook.
- `SECURITY.md` — reporting, scope, secure defaults.
- Expanded `README.md` and `CONTRIBUTING.md`.

### Known limitations

- **No writes in v1.** Trading, watchlist mutation, ACH moves,
  crypto — all out of scope. A `--write`-flagged plane may ship in a
  later plan with a very different auth posture.
- **No real-time streaming.** All data is pull-based; there is no
  websocket or SSE client.
- **Wash-sale accounting is not computed.** `rh orders` returns the
  raw order history; consult your 1099 for tax-correct cost basis.
- **Unofficial API.** Endpoint stability can change without notice;
  see `docs/ROBINHOOD_API.md` for the risk assessment.
