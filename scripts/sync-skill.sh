#!/usr/bin/env bash
# sync-skill.sh: regenerates the embedded skill tree at
#   internal/skill/assets/{claude-code,codex}/robinhood-cli/
# from skills/src/. The only per-target substitution is the
# {{INVOCATION_PREFIX}} placeholder inside SKILL.md.tmpl (Claude Code: "/",
# Codex: "@"). References and examples are copied verbatim.
#
# The generated tree is committed to the repo for deterministic releases
# (the rh binary embeds it via //go:embed all:assets). CI verifies the
# script is a no-op against the committed tree.
#
# Portable across bash 3.2 (macOS default) and bash 4+ (Linux/CI) — avoids
# associative arrays so it runs on `/usr/bin/env bash` everywhere.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SRC="${REPO_ROOT}/skills/src"

# Parallel arrays: TARGETS[i] + PREFIXES[i].
TARGETS=("claude-code" "codex")
PREFIXES=("/" "@")

for i in "${!TARGETS[@]}"; do
    target="${TARGETS[$i]}"
    prefix="${PREFIXES[$i]}"
    dest="${REPO_ROOT}/internal/skill/assets/${target}/robinhood-cli"

    rm -rf "${dest}"
    mkdir -p "${dest}"

    # Copy references and examples verbatim.
    cp -R "${SRC}/references" "${dest}/references"
    cp -R "${SRC}/examples"   "${dest}/examples"

    # SKILL.md — substitute the prefix.
    # Use | as delimiter since / is a common value.
    sed "s|{{INVOCATION_PREFIX}}|${prefix}|g" "${SRC}/SKILL.md.tmpl" > "${dest}/SKILL.md"
done

echo "synced internal/skill/assets/{claude-code,codex}/robinhood-cli"
