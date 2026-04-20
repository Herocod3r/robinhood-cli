package main

import (
	"os"

	"github.com/herocod3r/robinhood-cli/internal/output"
	"github.com/herocod3r/robinhood-cli/internal/robinhood/endpoints"
	"github.com/spf13/cobra"
)

func init() {
	c := &cobra.Command{
		Use:   "account",
		Short: "Account summary: buying power, cash, margin, PDT status",
		Annotations: map[string]string{
			"category":  "account",
			"stability": "green",
		},
		RunE: runAccount,
	}
	register(c)
}

func runAccount(cmd *cobra.Command, args []string) error {
	client, err := newAuthedClient()
	if err != nil {
		return err
	}
	a, err := endpoints.NewAccount(client).Get(cmd.Context())
	if err != nil {
		return err
	}
	switch resolveOutputMode() {
	case output.ModeJSON:
		w := &output.JSONWriter{Out: os.Stdout}
		return w.WriteSuccess("account", a, nil)
	default:
		w := &output.TableWriter{Out: os.Stdout}
		return w.WriteAccount(a)
	}
}
