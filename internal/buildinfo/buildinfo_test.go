package buildinfo

import "testing"

func TestDefaults(t *testing.T) {
	if Version == "" {
		t.Error("Version should have a default")
	}
	if Commit == "" {
		t.Error("Commit should have a default")
	}
	if SchemaVersion != "robinhood-cli/v1" {
		t.Errorf("SchemaVersion = %q, want %q", SchemaVersion, "robinhood-cli/v1")
	}
}
