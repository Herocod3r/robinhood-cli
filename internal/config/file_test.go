package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFile_Missing_ReturnsDefaults(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	f, err := LoadFile()
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	if f.DefaultProfile != "default" {
		t.Fatalf("DefaultProfile = %q, want %q", f.DefaultProfile, "default")
	}
	if _, ok := f.Profiles["default"]; !ok {
		t.Fatalf("Profiles should contain 'default'")
	}
}

func TestLoadFile_Existing_ReturnsParsed(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	appDir := filepath.Join(dir, AppName)
	if err := os.MkdirAll(appDir, 0o700); err != nil {
		t.Fatal(err)
	}
	cfg := []byte("default_profile: work\nprofiles:\n  work:\n    account_number: \"9876\"\n  home:\n    account_number: \"1234\"\n")
	if err := os.WriteFile(filepath.Join(appDir, "config.yaml"), cfg, 0o600); err != nil {
		t.Fatal(err)
	}
	f, err := LoadFile()
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	if f.DefaultProfile != "work" {
		t.Fatalf("DefaultProfile = %q, want %q", f.DefaultProfile, "work")
	}
	if f.Profiles["work"].AccountNumber != "9876" {
		t.Fatalf("work.AccountNumber = %q, want 9876", f.Profiles["work"].AccountNumber)
	}
}

func TestSaveFile_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	in := &File{
		DefaultProfile: "home",
		Profiles:       map[string]Profile{"home": {AccountNumber: "1234"}},
	}
	if err := SaveFile(in); err != nil {
		t.Fatalf("SaveFile: %v", err)
	}
	out, err := LoadFile()
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}
	if out.DefaultProfile != "home" || out.Profiles["home"].AccountNumber != "1234" {
		t.Fatalf("round-trip lost data: %+v", out)
	}
	// Verify 0600 perms on the written file.
	info, err := os.Stat(filepath.Join(dir, AppName, "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("config.yaml perms = %o, want 0600", info.Mode().Perm())
	}
}

func TestLoadFile_RejectsInvalidProfileKeys(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	appDir := filepath.Join(dir, AppName)
	if err := os.MkdirAll(appDir, 0o700); err != nil {
		t.Fatal(err)
	}
	cfg := []byte("default_profile: default\nprofiles:\n  \"../../evil\":\n    account_number: \"1\"\n")
	if err := os.WriteFile(filepath.Join(appDir, "config.yaml"), cfg, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadFile(); err == nil {
		t.Fatal("expected error for invalid profile key")
	}
}

func TestValidProfile(t *testing.T) {
	good := []string{"default", "work", "home-1", "a_b", "A1", "x"}
	for _, g := range good {
		if err := ValidProfile(g); err != nil {
			t.Errorf("ValidProfile(%q) = %v, want nil", g, err)
		}
	}
	bad := []string{"", "../etc/passwd", "a/b", "a b", "a.b", "a:b", string(make([]byte, 65))}
	for _, b := range bad {
		if err := ValidProfile(b); err == nil {
			t.Errorf("ValidProfile(%q) = nil, want error", b)
		}
	}
}
