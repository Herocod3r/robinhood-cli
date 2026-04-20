# Robinhood API — Inventory

This is the unofficial Robinhood API surface `rh` depends on. **All endpoints
are read-only in v1** — `rh` never issues POST/PUT/DELETE calls that mutate
account state (the Sheriff `/challenge/<id>/respond/` and OAuth token POSTs are
identity operations, not trading operations).

Stability ratings are the maintainer's assessment as of the v1.0.0 release.
See [`docs/AUTH.md`](./AUTH.md) for auth details and
[`docs/JSON_SCHEMA.md`](./JSON_SCHEMA.md) for the per-command `data` shape.

## Hosts

| Alias | Base URL | Used for |
|---|---|---|
| `api` | `https://api.robinhood.com` | Most endpoints |
| `nummus` | `https://nummus.robinhood.com` | Crypto (not used in v1) |
| `phoenix` | `https://phoenix.robinhood.com` | Unified account / portfolio enrichment |

Override: set `ROBINHOOD_API_HOST=https://localhost:8443` (or similar) and
`rh` will point all three aliases at that URL — useful for mitmproxy
captures and local development.

## Endpoints

| Endpoint | Method | Host | Used by | Stability | Notes |
|---|---|---|---|---|---|
| `/oauth2/token/` (grant_type=password) | POST | api | `rh login` | yellow | Sheriff nearly always required for new devices |
| `/oauth2/token/` (grant_type=refresh_token) | POST | api | all data commands | green | Stable since 2022 |
| `/pathfinder/user_machine/` | POST | api | `rh login` | yellow | Sheriff state machine; shape changed mid-2025 — tracked against robin_stocks master |
| `/pathfinder/inquiries/<id>/user_view/` | GET | api | `rh login` | yellow | Challenge view payload |
| `/challenge/<id>/respond/` | POST | api | `rh login` | yellow | SMS/email code submission |
| `/push/<id>/get_prompts_status/` | GET | api | `rh login` | yellow | Push prompt polling; 5-min deadline on our side |
| `/portfolios/` | GET | api | `rh portfolio` | green | Returns `results[0]` (single account) |
| `/accounts/unified` | GET | phoenix | `rh portfolio`, `rh account` | green | Best-effort enrichment for buying power / cash |
| `/accounts/` | GET | api | `rh account` | green | |
| `/positions/` | GET | api | `rh positions`, `rh position` | green | Cursor-paginated via `next` |
| `/instruments/` | GET | api | resolvers (`rh positions`, `rh orders`, `rh dividends`, `rh ratings`, `rh search`, `rh watchlist`) | green | Batch lookup by URL or UUID |
| `/quotes/` | GET | api | `rh quote`, position/watchlist enrichment | green | Max 50 symbols per call (enforced client-side) |
| `/fundamentals/` | GET | api | `rh fundamentals` | green | Max 50 symbols per call |
| `/quotes/historicals/<sym>/` | GET | api | `rh historicals` | green | `interval` + `span` enum-validated client-side |
| `/midlands/news/<sym>/` | GET | api | `rh news` | green | `next` present; CLI only fetches the first page |
| `/marketdata/earnings/` | GET | api | `rh earnings` | green | Filter by `symbol` query param |
| `/midlands/ratings/<iid>/` | GET | api | `rh ratings` | green | Keyed by instrument UUID — CLI resolves the symbol first |
| `/dividends/` | GET | api | `rh dividends` | green | Cursor-paginated |
| `/options/aggregate_positions/` | GET | api | `rh options-positions` | yellow | Per-leg option instruments fetched for strike/expiration; leg shape varies |
| `/options/instruments/<id>/` | GET | api | `rh options-positions` (leg expansion) | yellow | Called once per unique leg |
| `/orders/` | GET | api | `rh orders` | green | Cursor-paginated; newest-first |
| `/watchlists/` | GET | api | `rh watchlist` | green | |
| `/watchlists/<name>/` | GET | api | `rh watchlist` | green | Membership; cursor-paginated |
| `/markets/` | GET | api | `rh market-hours` | green | Index listing; per-market `/markets/<mic>/hours/<date>/` is a best-effort follow-up |
| `/markets/<mic>/hours/<date>/` | GET | api | `rh market-hours` | green | |
| `/documents/` | GET | api | `rh documents` | green | Cursor-paginated |
| (document download URL) | GET | varies | `rh documents --download` | yellow | Signed URL the server returns; host is not one of the three aliases. 500 MB cap enforced client-side |

## Stability definitions

- **green** — stable across releases in our tracking window (≥6 months);
  safe to depend on. Any silent shape change here would be a regression we
  expect our contract tests to catch.
- **yellow** — response shape or behavior changed within the tracking window,
  or is undocumented and subject to change without notice. Our tests cover
  the observed shape; the CLI tolerates additional unknown fields.
- **red** — not used by `rh`. Listed only if referenced by another doc.

## Versioning headers

All requests send:

- `User-Agent: robinhood-cli (+https://github.com/herocod3r/robinhood-cli)`
- `Accept: application/json`
- `Authorization: Bearer <access_token>` on authenticated calls.

OAuth requests additionally send `X-Robinhood-API-Version: 1.431.4`
(required by some Sheriff rollout cohorts — see
`internal/robinhood/oauth.go`).

## Rate limits

Robinhood does not publish rate limit numbers. Observed behaviour:

- 429 with `Retry-After` header — mapped to `CodeRateLimited` (exit 4).
- The `/quotes/` and `/fundamentals/` endpoints accept up to 50 symbols
  per call. `rh` batches accordingly and surfaces `CodeValidation`
  (exit 5) if a caller requests more.

## Pagination

Every paginated endpoint returns `{ results: [...], next: "<url>" }`
with `next` empty on the last page. `rh` follows `next` in-process and
accumulates all pages before returning. See `pathFromNext` in
`internal/robinhood/endpoints/` for the URL-stripping helper.

## When a yellow endpoint breaks

1. Compare the failing response against the matching struct in
   `internal/robinhood/endpoints/<endpoint>.go`.
2. Cross-reference against robin_stocks' master branch
   (`robin_stocks/robinhood/`) — they usually catch shape changes within
   a day.
3. If the breaking change is in the Sheriff / pathfinder stack,
   `rh login` will fail with `CodeSheriffRequired` and a useful hint.
   For other endpoints you'll see either a decode error
   (`malformed … response`) or an unexpected 4xx.
4. File an issue with the response body (redact tokens and account
   numbers) and the `rh --verbose` output.
