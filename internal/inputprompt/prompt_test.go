package inputprompt

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrompt_ReadsLine(t *testing.T) {
	p := &Prompter{
		In:  strings.NewReader("alice@example.com\n"),
		Out: &bytes.Buffer{},
	}
	got, err := p.Text("Email: ")
	if err != nil {
		t.Fatalf("Text: %v", err)
	}
	if got != "alice@example.com" {
		t.Fatalf("got %q, want %q", got, "alice@example.com")
	}
}

func TestPrompt_TrimsWhitespace(t *testing.T) {
	p := &Prompter{
		In:  strings.NewReader("  bob\t\n"),
		Out: &bytes.Buffer{},
	}
	got, err := p.Text("Name: ")
	if err != nil || got != "bob" {
		t.Fatalf("got %q err %v, want %q", got, err, "bob")
	}
}

func TestPrompt_RejectsEmptyWithRetryCap(t *testing.T) {
	p := &Prompter{
		In:              strings.NewReader("\n\n\n"),
		Out:             &bytes.Buffer{},
		MaxEmptyRetries: 2,
	}
	_, err := p.Text("Name: ")
	if err == nil {
		t.Fatalf("expected error after too many empties")
	}
}

func TestPrompt_PasswordNonTTY(t *testing.T) {
	p := &Prompter{
		In:  strings.NewReader("hunter2\n"),
		Out: &bytes.Buffer{},
	}
	got, err := p.Password("Password: ")
	if err != nil {
		t.Fatalf("Password: %v", err)
	}
	if got != "hunter2" {
		t.Fatalf("got %q, want %q", got, "hunter2")
	}
}

func TestPrompt_ChooseByNumber(t *testing.T) {
	p := &Prompter{
		In:  strings.NewReader("2\n"),
		Out: &bytes.Buffer{},
	}
	idx, err := p.Choose("Pick:", []string{"a", "b", "c"})
	if err != nil {
		t.Fatalf("Choose: %v", err)
	}
	if idx != 1 {
		t.Fatalf("idx = %d, want 1", idx)
	}
}
