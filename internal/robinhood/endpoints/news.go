package endpoints

import (
	"context"
	"strings"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

// NewsEP wraps the midlands news endpoint.
type NewsEP struct{ c *robinhood.Client }

// NewNews constructs a NewsEP.
func NewNews(c *robinhood.Client) *NewsEP { return &NewsEP{c: c} }

// NewsItem is one news card.
type NewsItem struct {
	Title       string `json:"title"`
	Author      string `json:"author,omitempty"`
	Source      string `json:"source"`
	URL         string `json:"url"`
	PublishedAt string `json:"published_at"`
	Summary     string `json:"summary,omitempty"`
	PreviewText string `json:"preview_text,omitempty"`
}

type newsResp struct {
	Results []NewsItem `json:"results"`
	Next    string     `json:"next"`
}

// Get returns the first page of news items for a symbol, optionally capped.
func (n *NewsEP) Get(ctx context.Context, symbol string, limit int) ([]NewsItem, error) {
	sym := strings.ToUpper(strings.TrimSpace(symbol))
	if sym == "" {
		return nil, &robinhood.APIError{Code: robinhood.CodeValidation, Message: "symbol required"}
	}
	path := "/midlands/news/" + sym + "/"
	var resp newsResp
	if err := n.c.GetJSONCtx(ctx, robinhood.APIHost, path, &resp); err != nil {
		return nil, err
	}
	if limit > 0 && limit < len(resp.Results) {
		resp.Results = resp.Results[:limit]
	}
	return resp.Results, nil
}
