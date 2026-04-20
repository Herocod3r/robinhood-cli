package skill

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstall_WritesAllFiles(t *testing.T) {
	dest := t.TempDir()
	count, err := Install("claude-code", dest, InstallOpts{})
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if count == 0 {
		t.Fatalf("count = 0, expected >0 files copied")
	}
	// SKILL.md exists and contains the front-matter name key.
	data, err := os.ReadFile(filepath.Join(dest, "robinhood-cli", "SKILL.md"))
	if err != nil {
		t.Fatalf("read SKILL.md: %v", err)
	}
	if !strings.Contains(string(data), "name: robinhood-cli") {
		t.Fatalf("SKILL.md missing frontmatter")
	}
	// references/ sub-dir exists.
	if _, err := os.Stat(filepath.Join(dest, "robinhood-cli", "references")); err != nil {
		t.Fatalf("references/ missing: %v", err)
	}
}

func TestInstall_UnknownTarget(t *testing.T) {
	if _, err := Install("ghcp-copilot", t.TempDir(), InstallOpts{}); err == nil {
		t.Fatalf("expected error for unknown target")
	}
}

func TestInstall_ConflictWithoutForceErrors(t *testing.T) {
	dest := t.TempDir()
	if _, err := Install("claude-code", dest, InstallOpts{}); err != nil {
		t.Fatal(err)
	}
	// install again without force -> error.
	if _, err := Install("claude-code", dest, InstallOpts{}); err == nil {
		t.Fatalf("expected conflict error on second install without force")
	}
}

func TestInstall_ForceOverwrites(t *testing.T) {
	dest := t.TempDir()
	if _, err := Install("claude-code", dest, InstallOpts{}); err != nil {
		t.Fatal(err)
	}
	// manually corrupt one file.
	p := filepath.Join(dest, "robinhood-cli", "SKILL.md")
	if err := os.WriteFile(p, []byte("corrupted"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Install("claude-code", dest, InstallOpts{Force: true}); err != nil {
		t.Fatalf("Install force: %v", err)
	}
	data, _ := os.ReadFile(p)
	if strings.Contains(string(data), "corrupted") {
		t.Fatalf("force did not overwrite")
	}
}

func TestInstall_WrittenFilesAre0644(t *testing.T) {
	dest := t.TempDir()
	if _, err := Install("claude-code", dest, InstallOpts{}); err != nil {
		t.Fatal(err)
	}
	_ = filepath.WalkDir(filepath.Join(dest, "robinhood-cli"), func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if info.Mode().Perm() != 0o644 {
			t.Errorf("%s perms = %o, want 0644", path, info.Mode().Perm())
		}
		return nil
	})
}
