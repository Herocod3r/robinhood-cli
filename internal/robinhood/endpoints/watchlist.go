package endpoints

import (
	"context"
	"net/url"
	"strings"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

// WatchlistEP wraps /watchlists/.
type WatchlistEP struct{ c *robinhood.Client }

// NewWatchlist constructs a WatchlistEP.
func NewWatchlist(c *robinhood.Client) *WatchlistEP { return &WatchlistEP{c: c} }

// WatchlistItem is one entry in a watchlist, optionally enriched with a quote.
type WatchlistItem struct {
	Symbol       string          `json:"symbol"`
	InstrumentID string          `json:"instrument_id"`
	AddedAt      string          `json:"added_at"`
	LastPrice    robinhood.Money `json:"last_price,omitempty"`
}

type watchlistsIndex struct {
	Results []struct {
		Name        string `json:"name"`
		URL         string `json:"url"`
		DisplayName string `json:"display_name"`
	} `json:"results"`
}

type watchlistDetail struct {
	Results []struct {
		Instrument string `json:"instrument"`
		Watchlist  string `json:"watchlist"`
		CreatedAt  string `json:"created_at"`
	} `json:"results"`
	Next string `json:"next"`
}

// Get returns the items on one watchlist (default "Default").
func (wl *WatchlistEP) Get(ctx context.Context, name string) ([]WatchlistItem, error) {
	if name == "" {
		name = "Default"
	}
	// Fetch watchlist membership.
	path := "/watchlists/" + url.PathEscape(name) + "/"
	var detail watchlistDetail
	var rows []struct {
		Instrument string
		CreatedAt  string
	}
	for {
		if err := wl.c.GetJSONCtx(ctx, robinhood.APIHost, path, &detail); err != nil {
			return nil, err
		}
		for _, r := range detail.Results {
			rows = append(rows, struct {
				Instrument string
				CreatedAt  string
			}{r.Instrument, r.CreatedAt})
		}
		if detail.Next == "" {
			break
		}
		p, err := pathFromNext(detail.Next)
		if err != nil {
			return nil, err
		}
		path = p
		detail = watchlistDetail{}
	}
	urls := make([]string, 0, len(rows))
	for _, r := range rows {
		urls = append(urls, r.Instrument)
	}
	m, err := NewInstruments(wl.c).ResolveByURL(ctx, urls)
	if err != nil {
		return nil, err
	}
	syms := m.Symbols()
	quotes := map[string]*Quote{}
	if len(syms) > 0 {
		q := NewQuotes(wl.c)
		for start := 0; start < len(syms); start += MaxBatch {
			end := start + MaxBatch
			if end > len(syms) {
				end = len(syms)
			}
			chunk, qerr := q.Batch(ctx, syms[start:end])
			if qerr != nil {
				return nil, qerr
			}
			for k, v := range chunk {
				quotes[k] = v
			}
		}
	}
	out := make([]WatchlistItem, 0, len(rows))
	for _, r := range rows {
		sym := m.Get(r.Instrument)
		item := WatchlistItem{
			Symbol:       sym,
			InstrumentID: instrumentIDFromURL(r.Instrument),
			AddedAt:      r.CreatedAt,
		}
		if q, ok := quotes[sym]; ok && q != nil {
			item.LastPrice = q.LastTradePrice
		}
		out = append(out, item)
	}
	return out, nil
}

// List returns all watchlist names the user has.
func (wl *WatchlistEP) List(ctx context.Context) ([]string, error) {
	var idx watchlistsIndex
	if err := wl.c.GetJSONCtx(ctx, robinhood.APIHost, "/watchlists/", &idx); err != nil {
		return nil, err
	}
	out := make([]string, 0, len(idx.Results))
	for _, r := range idx.Results {
		n := r.Name
		if r.DisplayName != "" {
			n = r.DisplayName
		}
		n = strings.TrimSpace(n)
		if n != "" {
			out = append(out, n)
		}
	}
	return out, nil
}
