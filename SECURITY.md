# Security policy

## Supported versions

We support the latest minor release of `rh`. Older versions receive no
backports or security patches.

## Reporting a vulnerability

Open a [private security advisory on GitHub](https://github.com/herocod3r/robinhood-cli/security/advisories/new)
or email the maintainer listed in `CODEOWNERS`.

Please include:

- Steps to reproduce.
- Affected `rh` version (`rh version --json | jq .data.version`).
- Impact assessment — what an attacker can do with the flaw.
- Suggested mitigation, if you have one.

**Target response:** we aim to acknowledge within 72 hours. Fix timeline
depends on severity — critical issues get immediate patch releases,
lower-severity issues are bundled into the next minor.

## Scope

**In scope:**

- The `rh` CLI binary and its Go source tree.
- The skill bundles under `skills/src/` and the embedded assets under
  `internal/skill/assets/`.
- GitHub release artifacts (archives, checksums, Homebrew formula).
- CI workflows under `.github/workflows/`.

**Out of scope:**

- Robinhood's servers and mobile apps. Report those to Robinhood directly.
- The Go standard library and third-party modules — file those upstream,
  but do let us know so we can bump the pin.
- Issues that require user-level code execution on the victim's machine
  (we consider that already-pwned).

There is **no bug bounty** for v1. We appreciate responsible disclosure
and will credit reporters in release notes.

## Secure defaults

- **Tokens in the OS keychain.** When the OS keychain is available,
  tokens live in macOS Keychain / GNOME libsecret / Windows Credential
  Manager. When it is not, tokens fall back to a file at
  `$XDG_CONFIG_HOME/robinhood-cli/robinhood-cli-<profile>.json` with
  `0600` permissions; the directory is `0700`. Writes use a
  `tmp + rename` sequence so a crash cannot leave a half-written blob.
  See [`docs/AUTH.md`](docs/AUTH.md).
- **Passwords never persisted.** `rh login` reads the password into
  memory, uses it for the token POST, and discards it. `LoginOpts` has
  custom `String()` / `GoString()` that redact every field so no
  formatter can leak it.
- **No password in logs.** `rh --verbose` logs HTTP URLs and status
  codes but never request bodies for OAuth endpoints and never
  `Authorization` header values.
- **Path-traversal guard on profiles.** Profile names are constrained
  to `[A-Za-z0-9_-]{1,64}` before they touch the filesystem.
- **Download cap.** `rh documents --download` streams through an
  `io.LimitReader` capped at 500 MB (Fix Q in the Plan B design).
- **Cassettes are sanitized.** Recorded HTTP cassettes in
  `testdata/cassettes/` are scrubbed of `access_token`,
  `refresh_token`, `device_token`, and Authorization headers. CI runs
  gitleaks (`.github/workflows/cassettes.yaml`) with custom rules in
  `.gitleaks.toml` to catch anything that slips through.
- **Dependency scanning.** `govulncheck` runs on every PR.

## What to do if you think a token leaked

1. `rh logout --all-profiles` — clears every local session.
2. Change your Robinhood password in the app; that invalidates all
   outstanding refresh tokens on the server side.
3. Re-login from the devices you trust.
4. If you think a cassette in `testdata/` contains live credentials,
   open a security advisory immediately so we can force-push the scrub.
