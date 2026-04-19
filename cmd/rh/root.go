package main

import (
	"github.com/herocod3r/robinhood-cli/internal/output"
	"github.com/spf13/cobra"
)

// globalFlags holds flags applied to the root command and inherited by children.
type globalFlags struct {
	JSON    bool
	Pretty  bool
	Profile string
	NoColor bool
	Verbose bool
}

var gflags globalFlags

// newRootCommand builds the root Cobra tree. Subcommands register themselves
// by calling rootCmd.AddCommand in their file's init().
func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "rh",
		Short:        "Read-only Robinhood CLI for portfolio research",
		SilenceUsage: true,
		// SilenceErrors: main() emits exactly one envelope itself and chooses exit code.
		SilenceErrors: true,
	}
	cmd.PersistentFlags().BoolVar(&gflags.JSON, "json", false, "force JSON output")
	cmd.PersistentFlags().BoolVar(&gflags.Pretty, "pretty", false, "force pretty/table output")
	cmd.PersistentFlags().StringVar(&gflags.Profile, "profile", "default", "account profile")
	cmd.PersistentFlags().BoolVar(&gflags.NoColor, "no-color", false, "disable ANSI color")
	cmd.PersistentFlags().BoolVarP(&gflags.Verbose, "verbose", "v", false, "debug logging to stderr")

	for _, sub := range registeredCommands {
		cmd.AddCommand(sub)
	}
	return cmd
}

// registeredCommands is populated by subcommand files via init().
var registeredCommands []*cobra.Command

// register appends a subcommand to be wired at root construction.
func register(c *cobra.Command) {
	registeredCommands = append(registeredCommands, c)
}

// resolveOutputMode picks the output mode based on global flags and TTY state.
func resolveOutputMode() output.Mode {
	switch {
	case gflags.JSON:
		return output.ModeJSON
	case gflags.Pretty:
		return output.ModePretty
	default:
		return output.ChooseMode(output.ModeAuto, output.IsStdoutTTY())
	}
}
