# Authentication

`rh` talks to Robinhood's unofficial mobile/web API over OAuth password grant
with optional device verification ("Sheriff") and MFA. This document describes
what the CLI actually does, where secrets live, and how to recover when
something breaks.

## Threat model

**What `rh` protects:**

- Tokens at rest are stored in the OS keychain when available (macOS Keychain,
  GNOME libsecret / KWallet, Windows Credential Manager). When the OS keychain
  is unavailable, tokens fall back to a `0600` JSON file under the application
  config directory.
- Passwords are **never** persisted. They are read into memory for the login
  POST and discarded.
- `LoginOpts` has custom `String()` / `GoString()` methods that redact every
  field, so no formatter, panic trace, or error envelope can accidentally leak
  credentials (see `cmd/rh/login.go`).
- Cassettes recorded for tests are sanitized; CI scans them with gitleaks
  using rules in `.gitleaks.toml`.

**What `rh` does NOT protect:**

- A malicious process running as your user can read your OS keychain entry
  (same threat model as any CLI that uses the system keychain).
- `rh` cannot harden Robinhood's servers. Rate limits, account locks, and
  Sheriff challenges come from Robinhood and are policy-driven.
- The unofficial API has no documented deprecation contract. A server-side
  change can break authentication at any time.

## Session lifecycle

```
+------------+   rh login    +--------+  ensureFresh   +--------+
| no session | ------------> | active | -------------> | active |
+------------+               +--------+   (refresh)    +--------+
      ^                          |  \                        |
      |                          |   \ 401 on request        |
      |                          |    \---(retry once)       |
      |                          |                           |
      | rh logout                v                           v
      +------------------- +-----------+    refresh      +--------+
                           | expired / | --------------> | active |
                           | unknown   |    rejected     +--------+
                           +-----------+        \
                                                 \--> CodeSessionExpired
                                                      (exit 2)
```

Implemented in `internal/robinhood/client.go` (`ensureFresh`) and
`internal/robinhood/oauth.go` (`Refresh`). The process-local cache of the
`Session` struct is kept in `Client` behind a mutex; multi-process coordination
is handled by a file lock (see "Concurrent refresh" below).

## Interactive `rh login`

Entry point: `cmd/rh/login.go` → `RunLogin` → `internal/robinhood.oauth.passwordGrant`.

High-level flow:

1. Read `--username` / prompt / `ROBINHOOD_USERNAME`.
2. Read `ROBINHOOD_PASSWORD` or prompt (no echo).
3. Resolve device token in this order: `ROBINHOOD_DEVICE_TOKEN` env var →
   existing keychain entry → new random 16-byte hex token generated via
   `crypto/rand`.
4. If `--totp-secret` or `ROBINHOOD_TOTP_SECRET` is set, generate the current
   TOTP code via `github.com/pquerna/otp/totp`.
5. POST `/oauth2/token/` with `grant_type=password`, `client_id` (the
   public Robinhood iOS/Android client id), `username`, `password`,
   `device_token`, `expires_in=86400`, `scope=internal`, and optional
   `mfa_code`. Headers include `X-Robinhood-API-Version: 1.431.4` (required
   by some Sheriff rollout cohorts) and `User-Agent: robinhood-cli (+…)`.

The response is classified by `classifyPasswordErr`:

- **200 OK** — tokens returned, `ExpiresAt = now + expires_in`, saved to
  keychain under the active profile. Done.
- **400/401 with `verification_workflow.id` (or top-level `id`)** —
  `CodeSheriffRequired` + `WorkflowID`. Triggers the Sheriff sub-flow.
- **400/401 with `mfa_required: true`** — `CodeMFARequired`. CLI prompts
  for the code and re-POSTs. Two attempts allowed before giving up.
- **other 4xx** — `CodeUnauthenticated` with the server's error message.

### Sheriff workflow

Handled in `internal/robinhood/sheriff.go` via `cmd/rh/login.go:runSheriff`.

1. POST `/pathfinder/user_machine/` with the workflow id → returns the
   next step URL (`SheriffSMS`, `SheriffEmail`, or `SheriffPush`).
2. GET that URL → `/pathfinder/inquiries/<id>/user_view/` to get the
   challenge payload.
3. For SMS/email: prompt the user for the code, POST `/challenge/<id>/respond/`.
4. For push: poll `/push/<id>/get_prompts_status/` every `PollEvery`
   (default 2s) with a 5-minute deadline (`Fix F`).
5. On success, re-issue the password grant with the
   `X-Robinhood-Challenge-Response-Id: <workflowID>` header. The server
   recognises the completed challenge and returns tokens.

### MFA workflow (TOTP / SMS)

If Robinhood returns `mfa_required`, the CLI prompts (or uses the
TOTP secret to compute a code) and retries with `mfa_code` set.
Up to two attempts. If the second attempt surfaces `sheriff_required`,
the Sheriff flow kicks in from that point.

## Silent refresh

Entry point: `Client.ensureFresh` in `internal/robinhood/client.go`.

- `Session.NeedsImmediateRefresh()` returns true only when no access token
  is present (but a refresh token is), or when a *known* `ExpiresAt` is
  within the 5-minute buffer. Unknown expiry ⇒ trust the access token.
- If a refresh is needed and the session is **not ephemeral**
  (`Session.Ephemeral == false`), refresh runs inside `WithRefreshLock`
  (`internal/robinhood/refresh_lock.go`) — a file-backed lock under the
  config dir that serialises concurrent `rh` processes.
  Under the lock the client re-reads the keychain first, so if another
  process just rotated tokens this process picks those up instead of
  racing the refresh.
- If the session **is ephemeral** (loaded from `ROBINHOOD_ACCESS_TOKEN` /
  `ROBINHOOD_REFRESH_TOKEN`), the refresh is serialised in-process only
  and the refreshed tokens are **never written back to the keychain** —
  env callers own their own persistence.

`GetJSONCtx` also retries once on a 401 response from a data endpoint:
drain + close the first body (to free the connection slot), call
`Refresh`, and re-issue the request. A second 401 returns
`CodeUnauthenticated` — at that point the refresh token itself is
invalid and the user must re-login.

## Storage

### OS keychain

- Service name: `robinhood-cli`
- Account: the profile name (default: `default`)
- Value: a JSON blob with `{username, access_token, refresh_token,
  device_token, expires_at}` (see `sessionBlob` in
  `internal/robinhood/session.go`).

Passwords and account numbers are never stored.

Backend is selected by `internal/keychain/keychain.go:Open`:

1. `ROBINHOOD_KEYCHAIN_BACKEND=file` → force the file fallback.
2. `ROBINHOOD_KEYCHAIN_BACKEND=system` → force the OS keychain (fails hard
   if unavailable).
3. Default: run a one-shot probe (`probeSystem`) with a 1.5 s deadline. On
   headless Linux with no `DBUS_SESSION_BUS_ADDRESS`, the probe
   short-circuits to false. If the probe passes, use the OS keychain;
   otherwise fall back to the file backend.

### File fallback

Location:

```
$XDG_CONFIG_HOME/robinhood-cli/robinhood-cli-<profile>.json
```

If `$XDG_CONFIG_HOME` is unset, `$HOME/.config/robinhood-cli/…` is used
(see `internal/config/config.go:ConfigDir`). The directory is created
at `0700`, the file at `0600`, written via a `tmp + rename` sequence so
a crash can't leave a half-written blob. Max size: 64 KiB
(`maxBlobBytes`). The config dir is **never** substituted with
`os.TempDir()` — if it can't be created, `rh` returns an error rather
than writing secrets to `/tmp`.

## Env-var override

Env vars are consumed in precedence order **env > keychain > fail**
(see `cmd/rh/client.go:newAuthedClient`):

| Var | Purpose |
|---|---|
| `ROBINHOOD_USERNAME` | Skip the username prompt |
| `ROBINHOOD_PASSWORD` | Skip the password prompt (headless login) |
| `ROBINHOOD_TOTP_SECRET` | Base32 shared secret; CLI generates the code |
| `ROBINHOOD_DEVICE_TOKEN` | Reuse a known device_token |
| `ROBINHOOD_ACCESS_TOKEN` | Bypass keychain; use this token directly |
| `ROBINHOOD_REFRESH_TOKEN` | Bypass keychain; use this refresh token |
| `ROBINHOOD_PROFILE` | Default profile (same as `--profile`) |
| `ROBINHOOD_KEYCHAIN_BACKEND` | `system` or `file` — force one backend |
| `ROBINHOOD_API_HOST` | Override the API/nummus/phoenix hosts (mitmproxy, tests) |

If **either** `ROBINHOOD_ACCESS_TOKEN` or `ROBINHOOD_REFRESH_TOKEN` is set,
the session is built from env (`LoadSessionFromEnv`), marked `Ephemeral:
true`, and keychain reads are skipped. This is the intended entrypoint
for CI, containers, and the skill install running inside an AI assistant
sandbox.

## Error codes

Taxonomy lives in `internal/robinhood/errors.go`. There are **7 codes**;
the last three rows are grouped under exit code 1.

| Code | Exit | Meaning | Typical user action |
|---|---:|---|---|
| `unauthenticated` | 2 | No session / keychain empty | `rh login` |
| `session_expired` | 2 | Refresh rejected | `rh login` (refresh token is dead) |
| `sheriff_required` | 3 | Device verification required | Finish the Robinhood app / SMS prompt |
| `mfa_required` | 3 | MFA code required | Enter TOTP or SMS code |
| `rate_limited` | 4 | Robinhood 429 | Retry after `Retry-After` seconds |
| `validation` | 5 | Bad flags/args | Fix the invocation |
| `robinhood_unavailable` | 1 | 5xx or network failure | Retry later |
| `not_found` | 1 | 404 from Robinhood | Check the ticker/ID |

Mapping is implemented by `APIError.ExitCode` — both
`sheriff_required` and `mfa_required` return 3 because both signal
"needs interactive step". The JSON envelope's `error.code` always
contains the string form.

## Multi-profile

Every command accepts `--profile <name>`; the env var `ROBINHOOD_PROFILE`
sets the default. Profile names must match `^[A-Za-z0-9_-]{1,64}$`
(`config.ValidProfile`) — this is a path-traversal guard since the profile
name becomes part of the file-fallback filename.

```bash
rh --profile work   login
rh --profile work   portfolio --json
rh --profile home   login
rh --profile home   portfolio --json
```

The CLI writes an entry to `$XDG_CONFIG_HOME/robinhood-cli/config.yaml`
under `profiles.<name>` when a new profile logs in (`cmd/rh/login.go:
persistLogin`). Use `rh logout --all-profiles` to clear every stored
session.

## Troubleshooting

| Symptom | Likely cause | Fix |
|---|---|---|
| `refresh rejected` | Refresh token revoked or expired | `rh logout && rh login` |
| `sheriff_required` but no prompt arrived | App notifications off, or the prompt went to an old device | Check the Robinhood mobile app; ensure push is enabled; retry |
| `401 after refresh` | Token possibly compromised, or clock skew | `rh logout --all-profiles` then `rh login`; fix the system clock |
| `failed to open keychain` on Linux | No dbus / libsecret | Install `libsecret` + `gnome-keyring` **or** export `ROBINHOOD_KEYCHAIN_BACKEND=file` |
| `invalid profile name` | Profile contains `/`, spaces, or is longer than 64 chars | Rename to `[A-Za-z0-9_-]{1,64}` |
| Intermittent 429 | Too many concurrent `rh` processes | The refresh lock serialises refreshes, but data calls are not throttled — add backoff in your script |
| "too many symbols" on `rh quote` | `/quotes/` is capped at 50 per call | Split the symbol list |
| `rh login` hangs on push | Waiting on the Robinhood app approval | There's a 5-minute deadline; approve in the app or Ctrl-C |
| Credentials leak into shell history | `ROBINHOOD_PASSWORD=...` was typed on the CLI | Put it in a `.env` file, or let `rh login` prompt interactively |
