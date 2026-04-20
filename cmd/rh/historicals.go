package main

import (
	"os"

	"github.com/herocod3r/robinhood-cli/internal/output"
	"github.com/herocod3r/robinhood-cli/internal/robinhood/endpoints"
	"github.com/spf13/cobra"
)

var (
	historicalsInterval string
	historicalsSpan     string
)

func init() {
	c := &cobra.Command{
		Use:   "historicals <ticker>",
		Short: "OHLCV bars (interval × span)",
		Args:  cobra.ExactArgs(1),
		Annotations: map[string]string{
			"category":  "market",
			"stability": "green",
		},
		RunE: runHistoricals,
	}
	c.Flags().StringVar(&historicalsInterval, "interval", "day", "5minute|10minute|hour|day|week")
	c.Flags().StringVar(&historicalsSpan, "span", "year", "day|week|month|3month|year|5year")
	register(c)
}

func runHistoricals(cmd *cobra.Command, args []string) error {
	client, err := newAuthedClient()
	if err != nil {
		return err
	}
	h, err := endpoints.NewHistoricals(client).Get(cmd.Context(), args[0], historicalsInterval, historicalsSpan)
	if err != nil {
		return err
	}
	switch resolveOutputMode() {
	case output.ModeJSON:
		w := &output.JSONWriter{Out: os.Stdout}
		return w.WriteSuccess("historicals", h, map[string]int{"count": len(h.Bars)})
	default:
		w := &output.TableWriter{Out: os.Stdout}
		return w.WriteHistoricals(h)
	}
}
