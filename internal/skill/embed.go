// Package skill ships the Claude Code + Codex skill bundles embedded in the
// rh binary. `rh skill install` copies the files out to the user's editor
// config path.
//
// The embed tree layout is:
//
//	internal/skill/assets/
//	  claude-code/robinhood-cli/SKILL.md + references/ + examples/
//	  codex/robinhood-cli/SKILL.md + references/ + examples/
//
// The tree is regenerated from skills/src/ via scripts/sync-skill.sh and is
// committed to the repo for deterministic releases — CI verifies the script
// is a no-op against the committed tree.
package skill

import "embed"

// skillsFS is the embedded asset tree. The //go:embed directive below uses
// the `all:` prefix so dotfiles and files starting with `_` are included too
// (currently none, but cheap insurance).

//go:embed all:assets
var skillsFS embed.FS
