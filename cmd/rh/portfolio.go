package main

import (
	"os"

	"github.com/herocod3r/robinhood-cli/internal/output"
	"github.com/herocod3r/robinhood-cli/internal/robinhood/endpoints"
	"github.com/spf13/cobra"
)

func init() {
	c := &cobra.Command{
		Use:     "portfolio",
		Aliases: []string{"pf"},
		Short:   "Show account portfolio value, equity, cash, buying power",
		RunE:    runPortfolio,
	}
	register(c)
}

func runPortfolio(cmd *cobra.Command, args []string) error {
	client, err := newAuthedClient()
	if err != nil {
		return err // main() writes the envelope
	}

	summary, err := endpoints.NewPortfolio(client).Get()
	if err != nil {
		return err // main() writes the envelope
	}

	switch resolveOutputMode() {
	case output.ModeJSON:
		w := &output.JSONWriter{Out: os.Stdout}
		return w.WriteSuccess("portfolio", summary, nil)
	default:
		w := &output.TableWriter{Out: os.Stdout}
		return w.WritePortfolio(summary)
	}
}
