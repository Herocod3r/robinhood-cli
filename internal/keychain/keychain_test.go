package keychain

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileStore_SetGetDelete(t *testing.T) {
	dir := t.TempDir()
	s := &FileStore{Dir: dir}
	if err := s.Set("rh", "default", []byte(`{"x":1}`)); err != nil {
		t.Fatalf("Set: %v", err)
	}
	// 0600 perms on the session file.
	info, err := os.Stat(filepath.Join(dir, "rh-default.json"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("perms = %o, want 0600", info.Mode().Perm())
	}
	got, err := s.Get("rh", "default")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if string(got) != `{"x":1}` {
		t.Fatalf("got %q, want %q", got, `{"x":1}`)
	}
	if err := s.Delete("rh", "default"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := s.Get("rh", "default"); err == nil {
		t.Fatalf("Get after Delete should fail")
	}
}

func TestFallback_PicksFileWhenSystemUnavailable(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("ROBINHOOD_KEYCHAIN_BACKEND", "file") // forces file backend
	s, err := Open()
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, ok := s.(*FileStore); !ok {
		t.Fatalf("expected FileStore, got %T", s)
	}
}

func TestFileStore_RejectsInvalidProfile(t *testing.T) {
	s := &FileStore{Dir: t.TempDir()}
	bad := []string{"", "../etc/passwd", "a/b", "a b"}
	for _, b := range bad {
		if err := s.Set("rh", b, []byte("x")); err == nil {
			t.Errorf("Set(%q) should have failed", b)
		}
		if _, err := s.Get("rh", b); err == nil {
			t.Errorf("Get(%q) should have failed", b)
		}
		if err := s.Delete("rh", b); err == nil {
			t.Errorf("Delete(%q) should have failed", b)
		}
	}
}

func TestFileStore_Get_RejectsOversize(t *testing.T) {
	dir := t.TempDir()
	s := &FileStore{Dir: dir}
	big := make([]byte, 64*1024+1)
	for i := range big {
		big[i] = 'x'
	}
	// Bypass Set so we land an oversize file on disk.
	if err := os.WriteFile(filepath.Join(dir, "rh-default.json"), big, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get("rh", "default"); err == nil {
		t.Fatal("Get should reject oversize file")
	}
}
