package main

import (
	"testing"
)

// TestPlanB_CommandsRegistered verifies all 22 Plan B commands live on the root.
func TestPlanB_CommandsRegistered(t *testing.T) {
	want := []string{
		"account", "commands", "dividends", "documents", "earnings",
		"fundamentals", "historicals", "login", "logout", "market-hours",
		"news", "options-positions", "orders", "portfolio", "position",
		"positions", "quote", "ratings", "schema", "search", "version",
		"watchlist",
	}
	root := newRootCommand()
	got := make(map[string]bool)
	for _, c := range root.Commands() {
		got[c.Name()] = true
	}
	for _, w := range want {
		if !got[w] {
			t.Errorf("command %q not registered", w)
		}
	}
}

// TestPlanB_AllCommandsHaveAnnotations enforces the wrap-up rule: every
// non-help command must set category + stability annotations.
func TestPlanB_AllCommandsHaveAnnotations(t *testing.T) {
	root := newRootCommand()
	for _, c := range root.Commands() {
		if c.Name() == "help" {
			continue
		}
		if c.Annotations["category"] == "" {
			t.Errorf("command %q missing category annotation", c.Name())
		}
		if c.Annotations["stability"] == "" {
			t.Errorf("command %q missing stability annotation", c.Name())
		}
	}
}

// TestPlanB_PayloadSurfacesCategoryStability verifies the JSON discovery
// payload echoes both annotations.
func TestPlanB_PayloadSurfacesCategoryStability(t *testing.T) {
	root := newRootCommand()
	payload := buildCommandsPayload(root)
	for _, p := range payload {
		if p.Name == "help" {
			continue
		}
		if p.Category == "" {
			t.Errorf("payload for %q missing category", p.Name)
		}
		if p.Stability == "" {
			t.Errorf("payload for %q missing stability", p.Name)
		}
	}
}
