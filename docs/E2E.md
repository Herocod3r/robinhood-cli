# End-to-end release checklist

The manual QA checklist a maintainer runs before every release.

Run on **macOS and Linux** (once each, ideally on a fresh host). Each step
has pass/fail; paste the checklist into the release PR body so reviewers
can see exactly which rows were exercised.

## Environment

- [ ] `go version` reports 1.22 or later.
- [ ] `rh version --json | jq .schema` is `"robinhood-cli/v1"`.
- [ ] `rh version --json | jq .data.version` matches the tag you're cutting.

## Fresh login (Sheriff / SMS)

- [ ] `rh logout --all-profiles`
- [ ] `rh login` prompts for username + password.
- [ ] Sheriff flow triggers; pick SMS on your phone and enter the code.
- [ ] (macOS only) `security find-generic-password -s robinhood-cli -a default`
      returns a non-empty result. **Do not paste the blob into the PR.**
- [ ] `rh portfolio --json | jq .error` is `null`.
- [ ] `rh portfolio --json | jq .data.equity` is a non-null numeric string.

## Push flow

- [ ] On a second machine (or after deleting the keychain entry):
      `rh login` triggers a push prompt. Approving it resumes login.

## Email verification

- [ ] Some cohorts surface email challenges instead of SMS.
      If offered, verify the code and confirm login succeeds.

## TOTP (headless login)

- [ ] Enable TOTP in your Robinhood 2FA settings and note the base32 secret.
- [ ] `ROBINHOOD_TOTP_SECRET=<secret> rh login` (with username/password via
      env or prompt) succeeds without further MFA prompts.

## Refresh path

- [ ] Wait >1 hour after the last login.
- [ ] `rh portfolio --json` still returns `error: null`.
- [ ] `rh -v portfolio` (verbose) shows the refresh call followed by the
      portfolio call.

## Read commands

Each of these must exit 0 and produce a non-empty `data`:

- [ ] `rh positions --json`
- [ ] `rh position AAPL --json`
- [ ] `rh account --json`
- [ ] `rh quote AAPL NVDA --json`
- [ ] `rh fundamentals AAPL --json`
- [ ] `rh historicals AAPL --span month --json`
- [ ] `rh news AAPL --limit 3 --json`
- [ ] `rh earnings AAPL --json`
- [ ] `rh ratings AAPL --json`
- [ ] `rh dividends --year 2025 --json`
- [ ] `rh options-positions --json`
- [ ] `rh orders --limit 5 --json`
- [ ] `rh watchlist --json`
- [ ] `rh search apple --json`
- [ ] `rh market-hours --json`
- [ ] `rh documents --type statement --json`

## Error-path smoke

- [ ] `rh position ZZZZZZ --json` exits non-zero with `error.code` set.
- [ ] `rh quote AAPL BBBB CCCC … (51 symbols)` exits 5 (validation).
- [ ] With `ROBINHOOD_ACCESS_TOKEN=bogus rh portfolio --json` exits 2 after
      a refresh failure.

## Skill install

- [ ] `rh skill install --target claude-code`
- [ ] New Claude Code session: `/robinhood-cli` skill is visible; running
      "How's my portfolio?" succeeds and returns a summary.
- [ ] `rh skill install --target codex`
- [ ] Codex session completes the same query.
- [ ] `rh skill install --target both` installs to both locations.

## Cleanup

- [ ] `rh logout` removes the keychain entry.
- [ ] `rh portfolio --json` now exits 2 with `code: unauthenticated`.
