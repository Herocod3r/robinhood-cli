package endpoints

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

// FundamentalsEP wraps /fundamentals/.
type FundamentalsEP struct{ c *robinhood.Client }

// NewFundamentals constructs a FundamentalsEP.
func NewFundamentals(c *robinhood.Client) *FundamentalsEP { return &FundamentalsEP{c: c} }

// Fundamentals is the subset of /fundamentals/ fields used by the CLI.
type Fundamentals struct {
	Symbol            string          `json:"symbol"`
	Open              robinhood.Money `json:"open"`
	High              robinhood.Money `json:"high"`
	Low               robinhood.Money `json:"low"`
	Volume            robinhood.Money `json:"volume"`
	AverageVolume     robinhood.Money `json:"average_volume"`
	AverageVolume2Wks robinhood.Money `json:"average_volume_2_weeks"`
	High52Weeks       robinhood.Money `json:"high_52_weeks"`
	Low52Weeks        robinhood.Money `json:"low_52_weeks"`
	DividendYield     robinhood.Money `json:"dividend_yield"`
	MarketCap         robinhood.Money `json:"market_cap"`
	PERatio           robinhood.Money `json:"pe_ratio"`
	Description       string          `json:"description,omitempty"`
	InstrumentURL     string          `json:"instrument,omitempty"`
}

type fundamentalsResp struct {
	Results []Fundamentals `json:"results"`
}

// Batch returns fundamentals for up to 50 tickers.
func (f *FundamentalsEP) Batch(ctx context.Context, symbols []string) ([]Fundamentals, error) {
	if len(symbols) == 0 {
		return nil, nil
	}
	if len(symbols) > MaxBatch {
		return nil, &robinhood.APIError{
			Code:    robinhood.CodeValidation,
			Message: fmt.Sprintf("too many symbols (%d); max %d per request", len(symbols), MaxBatch),
		}
	}
	upper := make([]string, len(symbols))
	for i, s := range symbols {
		upper[i] = strings.ToUpper(s)
	}
	params := url.Values{}
	params.Set("symbols", strings.Join(upper, ","))
	var resp fundamentalsResp
	if err := f.c.GetJSONCtx(ctx, robinhood.APIHost, "/fundamentals/?"+params.Encode(), &resp); err != nil {
		return nil, err
	}
	// Server may omit the symbol in some rollouts; backfill from caller order.
	for i := range resp.Results {
		if resp.Results[i].Symbol == "" && i < len(upper) {
			resp.Results[i].Symbol = upper[i]
		}
	}
	return resp.Results, nil
}
