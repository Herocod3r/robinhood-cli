package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/herocod3r/robinhood-cli/internal/output"
	"github.com/herocod3r/robinhood-cli/internal/robinhood"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() {
	c := &cobra.Command{
		Use:   "commands [name]",
		Short: "List commands (for scripts/skills); pass a name for one command's detail",
		Args:  cobra.MaximumNArgs(1),
		Annotations: map[string]string{
			"category":  "meta",
			"stability": "green",
		},
		RunE: runCommands,
	}
	register(c)
}

// CommandMeta is the stable discovery payload.
type CommandMeta struct {
	Name      string     `json:"name"`
	Aliases   []string   `json:"aliases,omitempty"`
	Summary   string     `json:"summary"`
	Category  string     `json:"category,omitempty"`
	Stability string     `json:"stability,omitempty"` // "green" | "yellow" | "red"
	Args      []string   `json:"args,omitempty"`
	Flags     []FlagMeta `json:"flags,omitempty"`
}

// FlagMeta is a single flag in the discovery payload.
type FlagMeta struct {
	Name        string `json:"name"`
	Shorthand   string `json:"shorthand,omitempty"`
	Type        string `json:"type"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description,omitempty"`
}

// buildCommandsPayload walks the Cobra tree and produces the stable discovery list.
func buildCommandsPayload(root *cobra.Command) []CommandMeta {
	cmds := root.Commands()
	out := make([]CommandMeta, 0, len(cmds))
	for _, c := range cmds {
		if c.Hidden {
			continue
		}
		out = append(out, CommandMeta{
			Name:      c.Name(),
			Aliases:   c.Aliases,
			Summary:   c.Short,
			Category:  c.Annotations["category"],
			Stability: c.Annotations["stability"],
			Flags:     flagsFor(c),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// flagsFor returns the local (non-inherited) flags of a command.
func flagsFor(c *cobra.Command) []FlagMeta {
	var out []FlagMeta
	c.LocalFlags().VisitAll(func(f *pflag.Flag) {
		out = append(out, FlagMeta{
			Name:        f.Name,
			Shorthand:   f.Shorthand,
			Type:        f.Value.Type(),
			Default:     f.DefValue,
			Description: f.Usage,
		})
	})
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func runCommands(cmd *cobra.Command, args []string) error {
	root := cmd.Root()
	payload := buildCommandsPayload(root)

	if len(args) == 1 {
		name := args[0]
		for _, c := range payload {
			if c.Name == name {
				return emitSingle(c)
			}
		}
		return &robinhood.APIError{
			Code:    robinhood.CodeValidation,
			Message: fmt.Sprintf("unknown command: %s", name),
			Hint:    "run: rh commands --json",
		}
	}

	mode := resolveOutputMode()
	switch mode {
	case output.ModeJSON:
		w := &output.JSONWriter{Out: os.Stdout}
		return w.WriteSuccess("commands", payload, map[string]int{"count": len(payload)})
	default:
		return printCommandsTable(os.Stdout, payload)
	}
}

func emitSingle(c CommandMeta) error {
	w := &output.JSONWriter{Out: os.Stdout}
	return w.WriteSuccess("commands", c, nil)
}

func printCommandsTable(out io.Writer, payload []CommandMeta) error {
	for _, c := range payload {
		aliases := ""
		if len(c.Aliases) > 0 {
			aliases = " (" + strings.Join(c.Aliases, ", ") + ")"
		}
		cat := c.Category
		if cat == "" {
			cat = "-"
		}
		stab := c.Stability
		if stab == "" {
			stab = "-"
		}
		if _, err := fmt.Fprintf(out, "%-20s %-10s %-8s %s\n", c.Name+aliases, cat, stab, c.Summary); err != nil {
			return err
		}
	}
	return nil
}
