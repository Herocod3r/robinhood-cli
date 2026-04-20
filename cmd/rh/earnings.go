package main

import (
	"os"

	"github.com/herocod3r/robinhood-cli/internal/output"
	"github.com/herocod3r/robinhood-cli/internal/robinhood/endpoints"
	"github.com/spf13/cobra"
)

func init() {
	c := &cobra.Command{
		Use:   "earnings <ticker>",
		Short: "Historical earnings: estimate vs actual, call info",
		Args:  cobra.ExactArgs(1),
		Annotations: map[string]string{
			"category":  "research",
			"stability": "green",
		},
		RunE: runEarnings,
	}
	register(c)
}

func runEarnings(cmd *cobra.Command, args []string) error {
	client, err := newAuthedClient()
	if err != nil {
		return err
	}
	items, err := endpoints.NewEarnings(client).Get(cmd.Context(), args[0])
	if err != nil {
		return err
	}
	switch resolveOutputMode() {
	case output.ModeJSON:
		w := &output.JSONWriter{Out: os.Stdout}
		return w.WriteSuccess("earnings", items, map[string]int{"count": len(items)})
	default:
		w := &output.TableWriter{Out: os.Stdout}
		return w.WriteEarnings(items)
	}
}
