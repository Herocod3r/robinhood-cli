package main

import (
	"os"

	"github.com/herocod3r/robinhood-cli/internal/output"
	"github.com/herocod3r/robinhood-cli/internal/robinhood/endpoints"
	"github.com/spf13/cobra"
)

var marketHoursDate string

func init() {
	c := &cobra.Command{
		Use:   "market-hours",
		Short: "Show market open/close for each exchange on a given date",
		Annotations: map[string]string{
			"category":  "market",
			"stability": "green",
		},
		RunE: runMarketHours,
	}
	c.Flags().StringVar(&marketHoursDate, "date", "", "YYYY-MM-DD (default today UTC)")
	register(c)
}

func runMarketHours(cmd *cobra.Command, args []string) error {
	client, err := newAuthedClient()
	if err != nil {
		return err
	}
	rows, err := endpoints.NewMarkets(client).List(cmd.Context(), marketHoursDate)
	if err != nil {
		return err
	}
	switch resolveOutputMode() {
	case output.ModeJSON:
		w := &output.JSONWriter{Out: os.Stdout}
		return w.WriteSuccess("market-hours", rows, map[string]int{"count": len(rows)})
	default:
		w := &output.TableWriter{Out: os.Stdout}
		return w.WriteMarketHours(rows)
	}
}
