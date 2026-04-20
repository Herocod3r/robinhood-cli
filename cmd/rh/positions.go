package main

import (
	"os"

	"github.com/herocod3r/robinhood-cli/internal/output"
	"github.com/herocod3r/robinhood-cli/internal/robinhood/endpoints"
	"github.com/spf13/cobra"
)

var (
	positionsNonZero bool
	positionsSort    string
	positionsLimit   int
)

func init() {
	c := &cobra.Command{
		Use:   "positions",
		Short: "List all equity positions with live quotes and unrealized P/L",
		Annotations: map[string]string{
			"category":  "portfolio",
			"stability": "green",
		},
		RunE: runPositions,
	}
	c.Flags().BoolVar(&positionsNonZero, "nonzero", false, "hide positions with zero quantity")
	c.Flags().StringVar(&positionsSort, "sort", "value", "sort by: value|pl|symbol")
	c.Flags().IntVar(&positionsLimit, "limit", 0, "cap output (0=all)")
	register(c)
}

func runPositions(cmd *cobra.Command, args []string) error {
	client, err := newAuthedClient()
	if err != nil {
		return err
	}
	ps, err := endpoints.NewPositions(client).List(cmd.Context(), endpoints.ListOpts{
		NonZero: positionsNonZero,
		Sort:    positionsSort,
		Limit:   positionsLimit,
	})
	if err != nil {
		return err
	}
	switch resolveOutputMode() {
	case output.ModeJSON:
		w := &output.JSONWriter{Out: os.Stdout}
		return w.WriteSuccess("positions", ps, map[string]int{"count": len(ps)})
	default:
		w := &output.TableWriter{Out: os.Stdout}
		return w.WritePositions(ps)
	}
}
