package main

import (
	"os"

	"github.com/herocod3r/robinhood-cli/internal/output"
	"github.com/herocod3r/robinhood-cli/internal/robinhood/endpoints"
	"github.com/spf13/cobra"
)

var watchlistName string

func init() {
	c := &cobra.Command{
		Use:   "watchlist",
		Short: "Show a watchlist's members with live quotes",
		Annotations: map[string]string{
			"category":  "research",
			"stability": "green",
		},
		RunE: runWatchlist,
	}
	c.Flags().StringVar(&watchlistName, "name", "Default", "watchlist name")
	register(c)
}

func runWatchlist(cmd *cobra.Command, args []string) error {
	client, err := newAuthedClient()
	if err != nil {
		return err
	}
	items, err := endpoints.NewWatchlist(client).Get(cmd.Context(), watchlistName)
	if err != nil {
		return err
	}
	switch resolveOutputMode() {
	case output.ModeJSON:
		w := &output.JSONWriter{Out: os.Stdout}
		return w.WriteSuccess("watchlist", items, map[string]int{"count": len(items)})
	default:
		w := &output.TableWriter{Out: os.Stdout}
		return w.WriteWatchlist(items)
	}
}
