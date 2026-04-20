package endpoints

import (
	"context"
	"net/url"
	"strings"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

// EarningsEP wraps /marketdata/earnings/.
type EarningsEP struct{ c *robinhood.Client }

// NewEarnings constructs an EarningsEP.
func NewEarnings(c *robinhood.Client) *EarningsEP { return &EarningsEP{c: c} }

// EarningsEvent is one earnings row.
type EarningsEvent struct {
	Symbol   string `json:"symbol"`
	Year     int    `json:"year"`
	Quarter  int    `json:"quarter"`
	ReportAt string `json:"report_at,omitempty"`
	EPS      struct {
		Estimate robinhood.Money `json:"estimate,omitempty"`
		Actual   robinhood.Money `json:"actual,omitempty"`
	} `json:"eps"`
	Call *struct {
		Datetime string `json:"datetime"`
		URL      string `json:"broadcast_url,omitempty"`
	} `json:"call,omitempty"`
}

type earningsResp struct {
	Results []EarningsEvent `json:"results"`
}

// Get returns earnings history for a symbol.
func (e *EarningsEP) Get(ctx context.Context, symbol string) ([]EarningsEvent, error) {
	sym := strings.ToUpper(strings.TrimSpace(symbol))
	if sym == "" {
		return nil, &robinhood.APIError{Code: robinhood.CodeValidation, Message: "symbol required"}
	}
	params := url.Values{}
	params.Set("symbol", sym)
	var resp earningsResp
	if err := e.c.GetJSONCtx(ctx, robinhood.APIHost, "/marketdata/earnings/?"+params.Encode(), &resp); err != nil {
		return nil, err
	}
	for i := range resp.Results {
		if resp.Results[i].Symbol == "" {
			resp.Results[i].Symbol = sym
		}
	}
	return resp.Results, nil
}
