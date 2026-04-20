package endpoints

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

// Quotes wraps /quotes/ — the batched real-time market data endpoint.
type Quotes struct{ c *robinhood.Client }

// NewQuotes constructs a Quotes endpoint.
func NewQuotes(c *robinhood.Client) *Quotes { return &Quotes{c: c} }

// Quote is the subset of /quotes/ fields exposed by `rh quote`.
type Quote struct {
	Symbol             string          `json:"symbol"`
	Name               string          `json:"name,omitempty"`
	LastTradePrice     robinhood.Money `json:"last_price"`
	BidPrice           robinhood.Money `json:"bid_price"`
	AskPrice           robinhood.Money `json:"ask_price"`
	BidSize            int             `json:"bid_size"`
	AskSize            int             `json:"ask_size"`
	PreviousClose      robinhood.Money `json:"previous_close"`
	ExtendedHoursPrice robinhood.Money `json:"extended_hours_price,omitempty"`
	Volume             int64           `json:"volume"`
	UpdatedAt          string          `json:"updated_at"`
}

// rawQuote mirrors the API's field names so a single unmarshal works against
// Robinhood and our typed consumer shape stays stable.
type rawQuote struct {
	Symbol                      string          `json:"symbol"`
	InstrumentID                string          `json:"instrument_id"`
	LastTradePrice              robinhood.Money `json:"last_trade_price"`
	LastExtendedHoursTradePrice robinhood.Money `json:"last_extended_hours_trade_price"`
	BidPrice                    robinhood.Money `json:"bid_price"`
	AskPrice                    robinhood.Money `json:"ask_price"`
	BidSize                     int             `json:"bid_size"`
	AskSize                     int             `json:"ask_size"`
	PreviousClose               robinhood.Money `json:"previous_close"`
	Volume                      robinhood.Money `json:"volume"`
	UpdatedAt                   string          `json:"updated_at"`
}

type quotesResp struct {
	Results []rawQuote `json:"results"`
}

// MaxBatch is the documented ceiling for /quotes/ symbols per request.
const MaxBatch = 50

// Batch fetches quotes for up to MaxBatch symbols; returns a symbol→Quote map.
// Unknown symbols are omitted.
func (q *Quotes) Batch(ctx context.Context, symbols []string) (map[string]*Quote, error) {
	out := make(map[string]*Quote, len(symbols))
	if len(symbols) == 0 {
		return out, nil
	}
	if len(symbols) > MaxBatch {
		return nil, &robinhood.APIError{
			Code:    robinhood.CodeValidation,
			Message: fmt.Sprintf("too many symbols (%d); max %d per request", len(symbols), MaxBatch),
		}
	}
	params := url.Values{}
	params.Set("symbols", strings.Join(symbols, ","))
	var resp quotesResp
	if err := q.c.GetJSONCtx(ctx, robinhood.APIHost, "/quotes/?"+params.Encode(), &resp); err != nil {
		return nil, err
	}
	for _, r := range resp.Results {
		out[r.Symbol] = &Quote{
			Symbol:             r.Symbol,
			LastTradePrice:     r.LastTradePrice,
			BidPrice:           r.BidPrice,
			AskPrice:           r.AskPrice,
			BidSize:            r.BidSize,
			AskSize:            r.AskSize,
			PreviousClose:      r.PreviousClose,
			ExtendedHoursPrice: r.LastExtendedHoursTradePrice,
			UpdatedAt:          r.UpdatedAt,
		}
	}
	return out, nil
}

// Single fetches one symbol's quote.
func (q *Quotes) Single(ctx context.Context, symbol string) (*Quote, error) {
	m, err := q.Batch(ctx, []string{symbol})
	if err != nil {
		return nil, err
	}
	if quote, ok := m[symbol]; ok {
		return quote, nil
	}
	return nil, &robinhood.APIError{Code: robinhood.CodeNotFound, Message: "no quote for " + symbol}
}
