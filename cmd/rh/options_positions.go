package main

import (
	"os"

	"github.com/herocod3r/robinhood-cli/internal/output"
	"github.com/herocod3r/robinhood-cli/internal/robinhood/endpoints"
	"github.com/spf13/cobra"
)

var optionsPositionsSymbol string

func init() {
	c := &cobra.Command{
		Use:   "options-positions",
		Short: "List options aggregate positions with per-leg strike/type/expiry",
		Annotations: map[string]string{
			"category":  "portfolio",
			"stability": "yellow",
		},
		RunE: runOptionsPositions,
	}
	c.Flags().StringVar(&optionsPositionsSymbol, "symbol", "", "filter to a single ticker")
	register(c)
}

func runOptionsPositions(cmd *cobra.Command, args []string) error {
	client, err := newAuthedClient()
	if err != nil {
		return err
	}
	rows, err := endpoints.NewOptionsPositions(client).List(cmd.Context(), optionsPositionsSymbol)
	if err != nil {
		return err
	}
	switch resolveOutputMode() {
	case output.ModeJSON:
		w := &output.JSONWriter{Out: os.Stdout}
		return w.WriteSuccess("options-positions", rows, map[string]int{"count": len(rows)})
	default:
		w := &output.TableWriter{Out: os.Stdout}
		return w.WriteOptionsPositions(rows)
	}
}
