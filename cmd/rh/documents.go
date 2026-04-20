package main

import (
	"net/http"
	"os"
	"time"

	"github.com/herocod3r/robinhood-cli/internal/datearg"
	"github.com/herocod3r/robinhood-cli/internal/output"
	"github.com/herocod3r/robinhood-cli/internal/robinhood"
	"github.com/herocod3r/robinhood-cli/internal/robinhood/endpoints"
	"github.com/spf13/cobra"
)

var (
	documentsType     string
	documentsSince    string
	documentsDownload string
	documentsForce    bool
)

func init() {
	c := &cobra.Command{
		Use:   "documents",
		Short: "List (and optionally download) account documents: 1099, statements, trade confirms",
		Annotations: map[string]string{
			"category":  "account",
			"stability": "yellow",
		},
		RunE: runDocuments,
	}
	c.Flags().StringVar(&documentsType, "type", "", "filter: 1099|trade_confirm|statement|other")
	c.Flags().StringVar(&documentsSince, "since", "", "YYYY-MM-DD or relative (ytd|7d|1y)")
	c.Flags().StringVar(&documentsDownload, "download", "", "download each document into PATH")
	c.Flags().BoolVar(&documentsForce, "force", false, "overwrite existing files when --download is set")
	register(c)
}

func runDocuments(cmd *cobra.Command, args []string) error {
	client, err := newAuthedClient()
	if err != nil {
		return err
	}
	opts := endpoints.DocumentOpts{Type: documentsType}
	if documentsSince != "" {
		t, perr := datearg.Parse(documentsSince, time.Now())
		if perr != nil {
			return &robinhood.APIError{Code: robinhood.CodeValidation, Message: "invalid --since: " + perr.Error()}
		}
		opts.Since = t
	}
	docs, err := endpoints.NewDocuments(client).List(cmd.Context(), opts)
	if err != nil {
		return err
	}
	if documentsDownload != "" {
		res, derr := endpoints.NewDocuments(client).Download(cmd.Context(), docs, endpoints.DownloadOpts{
			Dir:   documentsDownload,
			Force: documentsForce,
		}, &http.Client{Timeout: 5 * time.Minute})
		if derr != nil {
			return derr
		}
		switch resolveOutputMode() {
		case output.ModeJSON:
			w := &output.JSONWriter{Out: os.Stdout}
			return w.WriteSuccess("documents", res, map[string]int{"count": len(res)})
		default:
			w := &output.TableWriter{Out: os.Stdout}
			return w.WriteDocumentDownloads(res)
		}
	}
	switch resolveOutputMode() {
	case output.ModeJSON:
		w := &output.JSONWriter{Out: os.Stdout}
		return w.WriteSuccess("documents", docs, map[string]int{"count": len(docs)})
	default:
		w := &output.TableWriter{Out: os.Stdout}
		return w.WriteDocuments(docs)
	}
}
