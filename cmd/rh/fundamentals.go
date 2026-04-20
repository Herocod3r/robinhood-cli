package main

import (
	"os"
	"strings"

	"github.com/herocod3r/robinhood-cli/internal/output"
	"github.com/herocod3r/robinhood-cli/internal/robinhood/endpoints"
	"github.com/spf13/cobra"
)

func init() {
	c := &cobra.Command{
		Use:   "fundamentals <ticker> [ticker...]",
		Short: "Company fundamentals (market cap, P/E, 52-wk hi/lo, etc.)",
		Args:  cobra.MinimumNArgs(1),
		Annotations: map[string]string{
			"category":  "research",
			"stability": "green",
		},
		RunE: runFundamentals,
	}
	register(c)
}

func runFundamentals(cmd *cobra.Command, args []string) error {
	client, err := newAuthedClient()
	if err != nil {
		return err
	}
	upper := make([]string, len(args))
	for i, a := range args {
		upper[i] = strings.ToUpper(a)
	}
	rows, err := endpoints.NewFundamentals(client).Batch(cmd.Context(), upper)
	if err != nil {
		return err
	}
	switch resolveOutputMode() {
	case output.ModeJSON:
		w := &output.JSONWriter{Out: os.Stdout}
		return w.WriteSuccess("fundamentals", rows, nil)
	default:
		w := &output.TableWriter{Out: os.Stdout}
		return w.WriteFundamentals(rows)
	}
}
