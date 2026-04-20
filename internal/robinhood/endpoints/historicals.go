package endpoints

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

// HistoricalsEP wraps /quotes/historicals/.
type HistoricalsEP struct{ c *robinhood.Client }

// NewHistoricals constructs a HistoricalsEP.
func NewHistoricals(c *robinhood.Client) *HistoricalsEP { return &HistoricalsEP{c: c} }

// Historicals is the typed OHLCV response.
type Historicals struct {
	Symbol   string          `json:"symbol"`
	Interval string          `json:"interval"`
	Span     string          `json:"span"`
	Bars     []HistoricalBar `json:"bars"`
}

// HistoricalBar is one OHLCV candle.
type HistoricalBar struct {
	BeginsAt   string          `json:"begins_at"`
	OpenPrice  robinhood.Money `json:"open_price"`
	ClosePrice robinhood.Money `json:"close_price"`
	HighPrice  robinhood.Money `json:"high_price"`
	LowPrice   robinhood.Money `json:"low_price"`
	Volume     int64           `json:"volume"`
	Session    string          `json:"session"`
}

type historicalsResp struct {
	Symbol     string          `json:"symbol"`
	Interval   string          `json:"interval"`
	Span       string          `json:"span"`
	Historicals []HistoricalBar `json:"historicals"`
}

// ValidIntervals enumerates server-accepted intervals.
var ValidIntervals = map[string]bool{
	"5minute": true, "10minute": true, "hour": true, "day": true, "week": true,
}

// ValidSpans enumerates server-accepted spans.
var ValidSpans = map[string]bool{
	"day": true, "week": true, "month": true, "3month": true, "year": true, "5year": true,
}

// Get fetches OHLCV bars for one symbol.
func (h *HistoricalsEP) Get(ctx context.Context, symbol, interval, span string) (*Historicals, error) {
	sym := strings.ToUpper(symbol)
	if sym == "" {
		return nil, &robinhood.APIError{Code: robinhood.CodeValidation, Message: "symbol required"}
	}
	if interval == "" {
		interval = "day"
	}
	if span == "" {
		span = "year"
	}
	if !ValidIntervals[interval] {
		return nil, &robinhood.APIError{Code: robinhood.CodeValidation, Message: fmt.Sprintf("invalid interval %q", interval)}
	}
	if !ValidSpans[span] {
		return nil, &robinhood.APIError{Code: robinhood.CodeValidation, Message: fmt.Sprintf("invalid span %q", span)}
	}
	params := url.Values{}
	params.Set("interval", interval)
	params.Set("span", span)
	path := "/quotes/historicals/" + sym + "/?" + params.Encode()
	var resp historicalsResp
	if err := h.c.GetJSONCtx(ctx, robinhood.APIHost, path, &resp); err != nil {
		return nil, err
	}
	return &Historicals{
		Symbol:   sym,
		Interval: interval,
		Span:     span,
		Bars:     resp.Historicals,
	}, nil
}
