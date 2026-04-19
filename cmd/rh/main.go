package main

import (
	"os"

	"github.com/herocod3r/robinhood-cli/internal/output"
	"github.com/herocod3r/robinhood-cli/internal/robinhood"
	"github.com/spf13/cobra"
)

func main() {
	root := newRootCommand()
	if err := root.Execute(); err != nil {
		err = normalizeErr(err)
		renderTopLevelError(root, err)
		os.Exit(exitCodeFor(err))
	}
}

// normalizeErr wraps non-APIError failures (cobra/pflag built-ins like
// unknown flag or unknown subcommand) into a *robinhood.APIError with
// CodeValidation so the JSON envelope stays within the 7-code taxonomy
// defined in spec section 6.5.
func normalizeErr(err error) error {
	if _, ok := err.(*robinhood.APIError); ok {
		return err
	}
	return &robinhood.APIError{
		Code:    robinhood.CodeValidation,
		Message: err.Error(),
		Hint:    "run: rh --help",
	}
}

// renderTopLevelError emits exactly one record describing the failure.
// In JSON mode: one envelope on stdout, nothing on stderr.
// In pretty mode: one line on stderr, nothing on stdout.
func renderTopLevelError(root *cobra.Command, err error) {
	cmdName := commandNameFor(root)
	switch resolveOutputMode() {
	case output.ModeJSON:
		w := &output.JSONWriter{Out: os.Stdout}
		_ = w.WriteError(cmdName, err)
	default:
		w := &output.TableWriter{Out: os.Stderr}
		_ = w.WriteError(cmdName, err)
	}
}

// commandNameFor returns the name of the leaf command Cobra tried to run so
// error envelopes say e.g. "command": "portfolio" rather than "rh".
func commandNameFor(root *cobra.Command) string {
	c, _, _ := root.Find(os.Args[1:])
	if c != nil && c != root {
		return c.Name()
	}
	return root.Name()
}

// exitCodeFor maps an error to the CLI exit code. See spec section 6.5.
func exitCodeFor(err error) int {
	if apiErr, ok := err.(*robinhood.APIError); ok {
		return apiErr.ExitCode()
	}
	return 1
}
