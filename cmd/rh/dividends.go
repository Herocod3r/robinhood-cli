package main

import (
	"os"

	"github.com/herocod3r/robinhood-cli/internal/output"
	"github.com/herocod3r/robinhood-cli/internal/robinhood/endpoints"
	"github.com/spf13/cobra"
)

var (
	dividendsYear   int
	dividendsSymbol string
)

func init() {
	c := &cobra.Command{
		Use:   "dividends",
		Short: "List dividend events (paid/pending) with optional year + symbol filters",
		Annotations: map[string]string{
			"category":  "portfolio",
			"stability": "green",
		},
		RunE: runDividends,
	}
	c.Flags().IntVar(&dividendsYear, "year", 0, "filter to paid-at year (e.g. 2025)")
	c.Flags().StringVar(&dividendsSymbol, "symbol", "", "filter to a single ticker")
	register(c)
}

func runDividends(cmd *cobra.Command, args []string) error {
	client, err := newAuthedClient()
	if err != nil {
		return err
	}
	rows, err := endpoints.NewDividends(client).List(cmd.Context(), endpoints.DividendOpts{
		Year:   dividendsYear,
		Symbol: dividendsSymbol,
	})
	if err != nil {
		return err
	}
	switch resolveOutputMode() {
	case output.ModeJSON:
		w := &output.JSONWriter{Out: os.Stdout}
		return w.WriteSuccess("dividends", rows, map[string]int{"count": len(rows)})
	default:
		w := &output.TableWriter{Out: os.Stdout}
		return w.WriteDividends(rows)
	}
}
