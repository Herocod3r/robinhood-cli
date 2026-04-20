package endpoints

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

// DocumentsEP wraps /documents/.
type DocumentsEP struct{ c *robinhood.Client }

// NewDocuments constructs a DocumentsEP.
func NewDocuments(c *robinhood.Client) *DocumentsEP { return &DocumentsEP{c: c} }

// Document is one document row.
type Document struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Date        string `json:"date"`
	Name        string `json:"name"`
	DownloadURL string `json:"download_url"`
}

type documentsResp struct {
	Results []Document `json:"results"`
	Next    string     `json:"next"`
}

// DocumentOpts filters the listing.
type DocumentOpts struct {
	Type  string    // "" = all types
	Since time.Time // zero = no time filter
}

// List returns documents, optionally filtered by type + since.
func (d *DocumentsEP) List(ctx context.Context, opts DocumentOpts) ([]Document, error) {
	path := "/documents/"
	if opts.Type != "" {
		params := url.Values{}
		params.Set("type", opts.Type)
		path += "?" + params.Encode()
	}
	var out []Document
	for {
		var resp documentsResp
		if err := d.c.GetJSONCtx(ctx, robinhood.APIHost, path, &resp); err != nil {
			return nil, err
		}
		for _, doc := range resp.Results {
			if !opts.Since.IsZero() {
				if t, err := time.Parse("2006-01-02", doc.Date); err == nil {
					if t.Before(opts.Since) {
						continue
					}
				}
			}
			out = append(out, doc)
		}
		if resp.Next == "" {
			break
		}
		p, err := pathFromNext(resp.Next)
		if err != nil {
			return nil, err
		}
		path = p
	}
	return out, nil
}

// MaxDownloadBytes caps any single document download per Fix Q.
const MaxDownloadBytes = int64(500 * 1024 * 1024)

// DownloadOpts controls download behaviour.
type DownloadOpts struct {
	Dir   string
	Force bool
}

// DownloadResult is the outcome of one download attempt.
type DownloadResult struct {
	Document Document `json:"document"`
	Path     string   `json:"path"`
	Bytes    int64    `json:"bytes"`
	Skipped  bool     `json:"skipped"`
	Reason   string   `json:"reason,omitempty"`
}

// Download streams each document into opts.Dir.
// Per Fix Q: name is sanitized via filepath.Base, download is capped at
// 500 MB via io.LimitReader, file is created with 0600 perms, and an
// existing file refuses to overwrite unless opts.Force is true.
//
// httpClient lets callers inject a test client. When nil, a fresh
// http.Client is used (downloads go to the URL Robinhood provides,
// not the API host).
func (d *DocumentsEP) Download(ctx context.Context, docs []Document, opts DownloadOpts, httpClient *http.Client) ([]DownloadResult, error) {
	if opts.Dir == "" {
		return nil, errors.New("download dir required")
	}
	if err := os.MkdirAll(opts.Dir, 0700); err != nil {
		return nil, err
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	results := make([]DownloadResult, 0, len(docs))
	for _, doc := range docs {
		safeName := filepath.Base(strings.ReplaceAll(doc.Name, "\\", "/"))
		// filepath.Base turns empty/dot into ".", reject that.
		if safeName == "" || safeName == "." || safeName == string(os.PathSeparator) {
			safeName = doc.ID + ".pdf"
		}
		// Strip any path separators a malicious name might retain.
		safeName = strings.ReplaceAll(safeName, string(os.PathSeparator), "_")
		datePrefix := strings.TrimSpace(doc.Date)
		var fname string
		if datePrefix != "" {
			fname = datePrefix + "-" + safeName
		} else {
			fname = safeName
		}
		if !strings.Contains(fname, ".") {
			fname += ".pdf"
		}
		target := filepath.Join(opts.Dir, fname)
		if _, err := os.Stat(target); err == nil && !opts.Force {
			results = append(results, DownloadResult{Document: doc, Path: target, Skipped: true, Reason: "exists"})
			continue
		}
		n, err := streamDocument(ctx, httpClient, doc.DownloadURL, target)
		if err != nil {
			return results, err
		}
		results = append(results, DownloadResult{Document: doc, Path: target, Bytes: n})
	}
	return results, nil
}

func streamDocument(ctx context.Context, httpClient *http.Client, url, target string) (int64, error) {
	if url == "" {
		return 0, errors.New("empty download URL")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("download: HTTP %d", resp.StatusCode)
	}
	// 0600 per Fix Q.
	f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	limited := io.LimitReader(resp.Body, MaxDownloadBytes+1)
	n, err := io.Copy(f, limited)
	if err != nil {
		return n, err
	}
	if n > MaxDownloadBytes {
		_ = os.Remove(target)
		return 0, fmt.Errorf("download exceeded cap of %d bytes", MaxDownloadBytes)
	}
	return n, nil
}
