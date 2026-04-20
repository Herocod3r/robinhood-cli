package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

// TestSkillInstall_Both exercises the full Cobra path: --target both should
// write into both the Claude Code and Codex default directories when HOME is
// redirected to a temp dir. Asserts SKILL.md lands at both expected paths.
func TestSkillInstall_Both(t *testing.T) {
	tmp := t.TempDir()
	// Redirect home so default dirs land in tmp.
	t.Setenv("HOME", tmp)

	root := newRootCommand()
	buf := &bytes.Buffer{}
	root.SetErr(buf)
	root.SetOut(buf)
	root.SetArgs([]string{"skill", "install", "--target", "both"})
	if err := root.Execute(); err != nil {
		t.Fatalf("skill install --target both: %v", err)
	}
	for _, rel := range []string{
		".claude/skills/robinhood-cli/SKILL.md",
		".codex/skills/robinhood-cli/SKILL.md",
	} {
		if _, err := os.Stat(filepath.Join(tmp, rel)); err != nil {
			t.Errorf("%s: %v", rel, err)
		}
	}
}

// TestSkillInstall_BothPlusPath_Rejects ensures combining --target both with
// --path is rejected as a validation error without touching the filesystem.
func TestSkillInstall_BothPlusPath_Rejects(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	override := t.TempDir()

	root := newRootCommand()
	buf := &bytes.Buffer{}
	root.SetErr(buf)
	root.SetOut(buf)
	root.SetArgs([]string{"skill", "install", "--target", "both", "--path", override})
	err := root.Execute()
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}
	var apiErr *robinhood.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *robinhood.APIError, got %T: %v", err, err)
	}
	if apiErr.Code != robinhood.CodeValidation {
		t.Errorf("apiErr.Code = %q, want %q", apiErr.Code, robinhood.CodeValidation)
	}
	// Verify nothing was written into the override dir.
	entries, _ := os.ReadDir(override)
	if len(entries) > 0 {
		t.Errorf("override dir should be empty, got %d entries", len(entries))
	}
	// And nothing in the default dirs either.
	for _, rel := range []string{
		".claude/skills/robinhood-cli",
		".codex/skills/robinhood-cli",
	} {
		if _, err := os.Stat(filepath.Join(tmpHome, rel)); err == nil {
			t.Errorf("%s was unexpectedly created", rel)
		}
	}
}
