package main

import (
	"os"

	"github.com/herocod3r/robinhood-cli/internal/output"
	"github.com/herocod3r/robinhood-cli/internal/robinhood/endpoints"
	"github.com/spf13/cobra"
)

var newsLimit int

func init() {
	c := &cobra.Command{
		Use:   "news <ticker>",
		Short: "Recent news articles for a symbol",
		Args:  cobra.ExactArgs(1),
		Annotations: map[string]string{
			"category":  "research",
			"stability": "green",
		},
		RunE: runNews,
	}
	c.Flags().IntVar(&newsLimit, "limit", 10, "max items (0=all)")
	register(c)
}

func runNews(cmd *cobra.Command, args []string) error {
	client, err := newAuthedClient()
	if err != nil {
		return err
	}
	items, err := endpoints.NewNews(client).Get(cmd.Context(), args[0], newsLimit)
	if err != nil {
		return err
	}
	switch resolveOutputMode() {
	case output.ModeJSON:
		w := &output.JSONWriter{Out: os.Stdout}
		return w.WriteSuccess("news", items, map[string]int{"count": len(items)})
	default:
		w := &output.TableWriter{Out: os.Stdout}
		return w.WriteNews(items)
	}
}
