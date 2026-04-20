package main

import (
	"os"
	"strings"

	"github.com/herocod3r/robinhood-cli/internal/output"
	"github.com/herocod3r/robinhood-cli/internal/robinhood/endpoints"
	"github.com/spf13/cobra"
)

var quoteExtended bool

func init() {
	c := &cobra.Command{
		Use:   "quote <ticker> [ticker...]",
		Short: "Real-time quote for up to 50 tickers",
		Args:  cobra.MinimumNArgs(1),
		Annotations: map[string]string{
			"category":  "market",
			"stability": "green",
		},
		RunE: runQuote,
	}
	c.Flags().BoolVar(&quoteExtended, "extended-hours", false, "include extended-hours price when available")
	register(c)
}

func runQuote(cmd *cobra.Command, args []string) error {
	client, err := newAuthedClient()
	if err != nil {
		return err
	}
	symbols := make([]string, 0, len(args))
	for _, a := range args {
		symbols = append(symbols, strings.ToUpper(a))
	}
	m, err := endpoints.NewQuotes(client).Batch(cmd.Context(), symbols)
	if err != nil {
		return err
	}
	// Preserve caller order.
	ordered := make([]*endpoints.Quote, 0, len(symbols))
	for _, s := range symbols {
		if q, ok := m[s]; ok {
			ordered = append(ordered, q)
		}
	}
	switch resolveOutputMode() {
	case output.ModeJSON:
		w := &output.JSONWriter{Out: os.Stdout}
		return w.WriteSuccess("quote", ordered, nil)
	default:
		w := &output.TableWriter{Out: os.Stdout}
		return w.WriteQuotes(ordered, quoteExtended)
	}
}
