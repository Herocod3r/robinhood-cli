package main

import (
	"os"

	"github.com/herocod3r/robinhood-cli/internal/output"
	"github.com/herocod3r/robinhood-cli/internal/robinhood/endpoints"
	"github.com/spf13/cobra"
)

var searchLimit int

func init() {
	c := &cobra.Command{
		Use:   "search <query>",
		Short: "Search the Robinhood instrument catalog by name or symbol",
		Args:  cobra.MinimumNArgs(1),
		Annotations: map[string]string{
			"category":  "research",
			"stability": "green",
		},
		RunE: runSearch,
	}
	c.Flags().IntVar(&searchLimit, "limit", 10, "cap results")
	register(c)
}

func runSearch(cmd *cobra.Command, args []string) error {
	client, err := newAuthedClient()
	if err != nil {
		return err
	}
	q := args[0]
	for _, a := range args[1:] {
		q += " " + a
	}
	rows, err := endpoints.Search(cmd.Context(), endpoints.NewInstruments(client), q, searchLimit)
	if err != nil {
		return err
	}
	switch resolveOutputMode() {
	case output.ModeJSON:
		w := &output.JSONWriter{Out: os.Stdout}
		return w.WriteSuccess("search", rows, map[string]int{"count": len(rows)})
	default:
		w := &output.TableWriter{Out: os.Stdout}
		return w.WriteSearch(rows)
	}
}
