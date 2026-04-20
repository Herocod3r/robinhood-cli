package main

import (
	"os"
	"strings"

	"github.com/herocod3r/robinhood-cli/internal/output"
	"github.com/herocod3r/robinhood-cli/internal/robinhood"
	"github.com/herocod3r/robinhood-cli/internal/robinhood/endpoints"
	"github.com/spf13/cobra"
)

func init() {
	c := &cobra.Command{
		Use:   "position <ticker>",
		Short: "Show a single position (qty, last price, cost basis, P/L)",
		Args:  cobra.ExactArgs(1),
		Annotations: map[string]string{
			"category":  "portfolio",
			"stability": "green",
		},
		RunE: runPosition,
	}
	register(c)
}

func runPosition(cmd *cobra.Command, args []string) error {
	client, err := newAuthedClient()
	if err != nil {
		return err
	}
	sym := strings.ToUpper(args[0])
	ps, err := endpoints.NewPositions(client).List(cmd.Context(), endpoints.ListOpts{Symbol: sym})
	if err != nil {
		return err
	}
	if len(ps) == 0 {
		return &robinhood.APIError{
			Code:    robinhood.CodeNotFound,
			Message: "no position for " + sym,
			Hint:    "run: rh positions",
		}
	}
	switch resolveOutputMode() {
	case output.ModeJSON:
		w := &output.JSONWriter{Out: os.Stdout}
		return w.WriteSuccess("position", ps[0], nil)
	default:
		w := &output.TableWriter{Out: os.Stdout}
		return w.WritePositions(ps)
	}
}
