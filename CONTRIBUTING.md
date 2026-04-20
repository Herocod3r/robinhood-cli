# Contributing

Thanks for helping improve `rh`. This repo follows a short list of
conventions so contributions are fast to land.

## Ground rules

1. Fork, branch, PR to `main`.
2. `make test && make lint` must pass. For changes that touch an
   endpoint, also run `make test-fixtures` and `make test-contract`.
3. Follow the design spec at
   `docs/superpowers/specs/2026-04-18-robinhood-cli-design.md`.
4. Use conventional-commits style messages (`feat:`, `fix:`, `chore:`,
   `docs:`, `test:`, `refactor:`). Release notes are grouped by prefix.

## Development setup

### Prerequisites

- Go 1.22 or newer.
- `make`.
- `golangci-lint` (the `make lint` target installs it locally if
  missing).
- `goreleaser` (only required for local release dry-runs).
- On Linux, `libsecret-1-dev` + a running `dbus-daemon` if you want the
  OS keychain path to be exercised in local tests. Otherwise set
  `ROBINHOOD_KEYCHAIN_BACKEND=file`.

### First-time setup

```bash
git clone https://github.com/herocod3r/robinhood-cli.git
cd robinhood-cli
make test         # confirm clean baseline
make build        # produces ./rh
./rh version --json
```

## Test tiers

| Tier | Target | What it exercises |
|---|---|---|
| 1 — Unit | `make test` | In-process Go tests; hermetic; runs on every PR |
| 1 — Race | `make test-race` | Same tests with the race detector |
| 2 — Fixtures | `make test-fixtures` | Cassette-based integration (`//go:build fixtures`); replays recorded HTTP |
| 3 — Contract | `make test-contract` | JSON-schema validation of golden outputs against `docs/JSON_SCHEMA.md` |
| 4 — E2E | `docs/E2E.md` | Manual live-account checklist before each release |

Tier 1 runs on every push. Tier 2/3 run on every PR via `ci.yaml`.
Tier 4 is maintainer-only.

## Coding conventions

- **Money is a string.** `robinhood.Money` is a string that looks like
  a decimal number. Never convert to `float64`. When you need math,
  use the helpers in `internal/robinhood/endpoints/math.go`
  (`addMoney`, `subMoney`, `mulMoney`, `divMoney`, `toDec`). Decimal
  math uses `github.com/shopspring/decimal`.
- **Every HTTP call takes a `context.Context`.** Use
  `client.GetJSONCtx(ctx, …)`; don't introduce uncontextual helpers.
- **All errors surfaced to the user are `*robinhood.APIError`.**
  Cobra `RunE` functions return `error`; `main.go` maps the APIError
  to the documented exit codes. Never return `errors.New`/`fmt.Errorf`
  from a command unless it is immediately wrapped.
- **No `os.Exit` in libraries.** Only `cmd/rh/main.go` calls
  `os.Exit`, once, after rendering the error envelope.
- **0600 for anything that holds a token.** 0700 for directories.
  Use `tmp + rename` for writes. Never fall back to `os.TempDir()`
  for secrets.
- **Stable JSON.** Any new field on an existing command must be
  **additive** — existing consumers must keep working. Any rename is
  a breaking change and requires a new `schema` version
  (`robinhood-cli/vN`).

## How to add a new command

1. Write the endpoint wrapper under
   `internal/robinhood/endpoints/<name>.go`. Put the typed response
   struct there; the JSON tags in the struct are the stable output
   contract.
2. Add a unit test `<name>_test.go` using `httptest.Server` (see
   `portfolio_test.go` for the canonical shape).
3. Add the Cobra command under `cmd/rh/<name>.go`. Follow the template:

   ```go
   func init() {
       c := &cobra.Command{
           Use:   "foo",
           Short: "One-line summary",
           Annotations: map[string]string{
               "category":  "data",
               "stability": "green",
           },
           RunE: runFoo,
       }
       c.Flags().String("some-flag", "", "what it does")
       register(c)
   }

   func runFoo(cmd *cobra.Command, args []string) error {
       ctx := cmd.Context()
       client, err := newAuthedClient()
       if err != nil { return err }
       result, err := endpoints.NewFoo(client).Get(ctx, …)
       if err != nil { return err }
       return emit(cmd, "foo", result, nil)
   }
   ```

4. Add a schema block in `docs/JSON_SCHEMA.md` with heading
   `` ### `rh foo` `` — the first fenced `json` block must be the
   JSON Schema; examples go in `jsonc` fences.
5. Add a golden file `testdata/contract/foo.json` so the contract
   test has something to validate.
6. Add a row to `docs/ROBINHOOD_API.md`.
7. Update `docs/E2E.md` with a smoke line for the new command.
8. Update `CHANGELOG.md` under `[Unreleased]`.
9. Run `./scripts/sync-skill.sh` if the skill needs updating.

## How to record cassettes

Cassettes live in `testdata/cassettes/` and are consumed by the
fixtures tier. Record a new one locally:

1. Flip the recorder mode to `recorder.ModeRecordOnly` in your test
   file (temporary — do not commit).
2. Run the test against a live session.
3. Run the sanitizer — the recorder's `BeforeSaveHook` redacts
   `Authorization`, `access_token`, `refresh_token`, and
   `device_token`. Double-check the YAML before committing.
4. Switch the recorder mode back to `ModeReplayOnly` and verify the
   test passes in replay mode.
5. `git add testdata/cassettes/<name>.yaml` and open a PR — the
   gitleaks workflow will re-scan the cassette.

## Known vulnerabilities

`govulncheck ./...` runs on every PR (`vuln` job in `ci.yaml`). CI is
pinned to the Go toolchain resolved from `go.mod` (currently `go 1.25.x`
via `actions/setup-go@v5 go-version: "1.22"` which bumps forward to
satisfy `go.mod`), where no exploitable stdlib findings apply.

Local contributors running newer Go toolchains (e.g. `go1.26.0`) may see
findings from `govulncheck` that are **not in the CI matrix**. Current
known-noisy stdlib findings when building on `go1.26.0` locally:

| ID            | Package     | Fixed in  | Notes |
|---------------|-------------|-----------|-------|
| GO-2026-4947  | crypto/x509 | go1.26.2  | chain building; called via TLS |
| GO-2026-4946  | crypto/x509 | go1.26.2  | policy validation; called via TLS |
| GO-2026-4870  | crypto/tls  | go1.26.2  | TLS 1.3 KeyUpdate DoS |
| GO-2026-4866  | crypto/x509 | go1.26.2  | name-constraint auth bypass |
| GO-2026-4602  | os          | go1.26.1  | FileInfo Root escape |
| GO-2026-4601  | net/url     | go1.26.1  | IPv6 host literal parsing |
| GO-2026-4600  | crypto/x509 | go1.26.1  | malformed cert panic |
| GO-2026-4599  | crypto/x509 | go1.26.1  | email-constraint enforcement |

If you are on `go1.26.x`, update to `go1.26.2` or newer to clear them.
The CI job is authoritative — if it passes, the release toolchain is
clean. Do not silence these findings in `.golangci.yml` or skip the
job; they indicate the contributor's local stdlib is out of date.

## Release cadence

- **Patch** releases go out as needed when a bug blocks users.
- **Minor** releases go out roughly every 4-6 weeks with accumulated
  features.
- **Major** releases require an RFC in `docs/superpowers/specs/`.

See [`docs/release-process.md`](docs/release-process.md) for the full
runbook.
