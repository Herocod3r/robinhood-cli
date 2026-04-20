package main

import (
	"os"
	"time"

	"github.com/herocod3r/robinhood-cli/internal/datearg"
	"github.com/herocod3r/robinhood-cli/internal/output"
	"github.com/herocod3r/robinhood-cli/internal/robinhood"
	"github.com/herocod3r/robinhood-cli/internal/robinhood/endpoints"
	"github.com/spf13/cobra"
)

var (
	ordersSince  string
	ordersSymbol string
	ordersState  string
	ordersLimit  int
)

func init() {
	c := &cobra.Command{
		Use:   "orders",
		Short: "List order history (filters: --since --symbol --state --limit)",
		Annotations: map[string]string{
			"category":  "portfolio",
			"stability": "green",
		},
		RunE: runOrders,
	}
	c.Flags().StringVar(&ordersSince, "since", "", "YYYY-MM-DD or relative (7d|2w|1mo|1y|ytd)")
	c.Flags().StringVar(&ordersSymbol, "symbol", "", "filter to one ticker")
	c.Flags().StringVar(&ordersState, "state", "", "filled|cancelled|confirmed|open")
	c.Flags().IntVar(&ordersLimit, "limit", 0, "cap output (0=all)")
	register(c)
}

func runOrders(cmd *cobra.Command, args []string) error {
	client, err := newAuthedClient()
	if err != nil {
		return err
	}
	opts := endpoints.OrderOpts{
		Symbol: ordersSymbol,
		State:  ordersState,
		Limit:  ordersLimit,
	}
	if ordersSince != "" {
		t, perr := datearg.Parse(ordersSince, time.Now())
		if perr != nil {
			return &robinhood.APIError{Code: robinhood.CodeValidation, Message: "invalid --since: " + perr.Error()}
		}
		opts.Since = t
	}
	rows, err := endpoints.NewOrders(client).List(cmd.Context(), opts)
	if err != nil {
		return err
	}
	switch resolveOutputMode() {
	case output.ModeJSON:
		w := &output.JSONWriter{Out: os.Stdout}
		return w.WriteSuccess("orders", rows, map[string]int{"count": len(rows)})
	default:
		w := &output.TableWriter{Out: os.Stdout}
		return w.WriteOrders(rows)
	}
}
