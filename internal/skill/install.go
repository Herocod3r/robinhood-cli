package skill

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// InstallOpts controls Install behavior.
type InstallOpts struct {
	// Force causes an existing destination `robinhood-cli/` directory to be
	// removed and rewritten. Without Force, a non-empty destination returns
	// a "already exists" error.
	Force bool
}

// knownTargets maps the --target flag values to the embedded subdirectory
// inside skillsFS. Keep the values aligned with scripts/sync-skill.sh.
var knownTargets = map[string]string{
	"claude-code": "assets/claude-code/robinhood-cli",
	"codex":       "assets/codex/robinhood-cli",
}

// Install copies the skill for the given target into destRoot/robinhood-cli/.
// Returns the number of files copied.
//
// `target` must be one of the keys of knownTargets ("claude-code" or
// "codex"); unknown values return an error without touching the filesystem.
//
// `destRoot` is the parent directory (e.g. "~/.claude/skills/"); the skill
// is installed as a subdirectory named "robinhood-cli" under it.
func Install(target, destRoot string, opts InstallOpts) (int, error) {
	src, ok := knownTargets[target]
	if !ok {
		return 0, fmt.Errorf("unknown target %q (supported: claude-code, codex)", target)
	}
	finalDir := filepath.Join(destRoot, "robinhood-cli")
	exists := dirNonEmpty(finalDir)
	if exists && !opts.Force {
		return 0, fmt.Errorf("%s already exists (use --force to overwrite)", finalDir)
	}
	if exists && opts.Force {
		if err := os.RemoveAll(finalDir); err != nil {
			return 0, err
		}
	}
	if err := os.MkdirAll(finalDir, 0o755); err != nil {
		return 0, err
	}
	count := 0
	err := fs.WalkDir(skillsFS, src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(path, src)
		rel = strings.TrimPrefix(rel, "/")
		if rel == "" {
			return nil
		}
		dest := filepath.Join(finalDir, rel)
		if d.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}
		in, err := skillsFS.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, in); err != nil {
			_ = out.Close()
			return err
		}
		if err := out.Close(); err != nil {
			return err
		}
		count++
		return nil
	})
	if err != nil {
		return 0, err
	}
	if count == 0 {
		return 0, errors.New("no files copied; embedded assets likely missing")
	}
	return count, nil
}

// dirNonEmpty reports whether dir exists and contains at least one entry.
func dirNonEmpty(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	return len(entries) > 0
}
