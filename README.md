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
