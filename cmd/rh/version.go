package main

import (
	"fmt"
	"os"

	"github.com/herocod3r/robinhood-cli/internal/buildinfo"
	"github.com/herocod3r/robinhood-cli/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	c := &cobra.Command{
		Use:   "version",
		Short: "Print CLI version and schema version",
		RunE:  runVersion,
	}
	register(c)
}

type versionPayload struct {
	Version       string `json:"version"`
	Commit        string `json:"commit"`
	SchemaVersion string `json:"schema_version"`
}

func runVersion(cmd *cobra.Command, args []string) error {
	payload := versionPayload{
		Version:       buildinfo.Version,
		Commit:        buildinfo.Commit,
		SchemaVersion: buildinfo.SchemaVersion,
	}

	mode := resolveOutputMode()
	switch mode {
	case output.ModeJSON:
		w := &output.JSONWriter{Out: os.Stdout}
		return w.WriteSuccess("version", payload, nil)
	default:
		_, err := fmt.Fprintf(os.Stdout, "rh %s (%s) — schema %s\n",
			payload.Version, payload.Commit, payload.SchemaVersion)
		return err
	}
}
