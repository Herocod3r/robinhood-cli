package main

import (
	"encoding/json"
	"testing"
)

func TestCommandsOutput_HasPortfolio(t *testing.T) {
	root := newRootCommand()
	data := buildCommandsPayload(root)

	var found bool
	for _, c := range data {
		if c.Name == "portfolio" {
			found = true
			if len(c.Aliases) == 0 || c.Aliases[0] != "pf" {
				t.Errorf("portfolio aliases = %v", c.Aliases)
			}
		}
	}
	if !found {
		t.Error("portfolio should appear in commands payload")
	}
}

func TestCommandsOutput_JSONRoundTrip(t *testing.T) {
	root := newRootCommand()
	data := buildCommandsPayload(root)
	b, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	var out []CommandMeta
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if len(out) != len(data) {
		t.Errorf("round-trip len mismatch")
	}
}
