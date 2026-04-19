package output

import (
	"bytes"
	"testing"
)

func TestChoose_ExplicitJSON(t *testing.T) {
	if ChooseMode(ModeJSON, false) != ModeJSON {
		t.Error("explicit JSON must stick")
	}
}

func TestChoose_ExplicitPretty(t *testing.T) {
	if ChooseMode(ModePretty, true) != ModePretty {
		t.Error("explicit Pretty must stick")
	}
}

func TestChoose_AutoTTY(t *testing.T) {
	if ChooseMode(ModeAuto, true) != ModePretty {
		t.Error("auto + tty should yield Pretty")
	}
	if ChooseMode(ModeAuto, false) != ModeJSON {
		t.Error("auto + non-tty should yield JSON")
	}
}

func TestNewWriter_JSON(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(ModeJSON, &buf)
	_, ok := w.(*JSONWriter)
	if !ok {
		t.Errorf("ModeJSON should produce *JSONWriter, got %T", w)
	}
}

func TestNewWriter_Pretty(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(ModePretty, &buf)
	_, ok := w.(*TableWriter)
	if !ok {
		t.Errorf("ModePretty should produce *TableWriter, got %T", w)
	}
}
