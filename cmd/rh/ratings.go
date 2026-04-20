package main

import (
	"os"

	"github.com/herocod3r/robinhood-cli/internal/output"
	"github.com/herocod3r/robinhood-cli/internal/robinhood/endpoints"
	"github.com/spf13/cobra"
)

func init() {
	c := &cobra.Command{
		Use:   "ratings <ticker>",
		Short: "Analyst buy/hold/sell summary plus recent comments",
		Args:  cobra.ExactArgs(1),
		Annotations: map[string]string{
			"category":  "research",
			"stability": "green",
		},
		RunE: runRatings,
	}
	register(c)
}

func runRatings(cmd *cobra.Command, args []string) error {
	client, err := newAuthedClient()
	if err != nil {
		return err
	}
	rating, err := endpoints.NewRatings(client).Get(cmd.Context(), args[0])
	if err != nil {
		return err
	}
	switch resolveOutputMode() {
	case output.ModeJSON:
		w := &output.JSONWriter{Out: os.Stdout}
		return w.WriteSuccess("ratings", rating, nil)
	default:
		w := &output.TableWriter{Out: os.Stdout}
		return w.WriteRatings(rating)
	}
}
