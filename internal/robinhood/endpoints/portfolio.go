// Package endpoints holds typed Robinhood REST endpoint wrappers.
package endpoints

import (
	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

// Portfolio wraps portfolio-related endpoints.
type Portfolio struct {
	c *robinhood.Client
}

// NewPortfolio constructs a Portfolio endpoint bound to the given client.
func NewPortfolio(c *robinhood.Client) *Portfolio {
	return &Portfolio{c: c}
}

// PortfolioSummary is the merged view presented by `rh portfolio`.
type PortfolioSummary struct {
	Equity              robinhood.Money `json:"equity"`
	ExtendedHoursEquity robinhood.Money `json:"extended_hours_equity"`
	MarketValue         robinhood.Money `json:"market_value"`
	Cash                robinhood.Money `json:"cash"`
	BuyingPower         robinhood.Money `json:"buying_power"`
	LastCoreEquity      robinhood.Money `json:"last_core_equity,omitempty"`
}

type portfolioResultsResp struct {
	Results []struct {
		Account             string          `json:"account"`
		Equity              robinhood.Money `json:"equity"`
		ExtendedHoursEquity robinhood.Money `json:"extended_hours_equity"`
		MarketValue         robinhood.Money `json:"market_value"`
		LastCoreEquity      robinhood.Money `json:"last_core_equity"`
	} `json:"results"`
}

type unifiedAccountResp struct {
	TotalEquity robinhood.Money `json:"total_equity"`
	Cash        robinhood.Money `json:"cash"`
	BuyingPower robinhood.Money `json:"buying_power"`
}

// Get returns a merged PortfolioSummary combining /portfolios/ and /accounts/unified.
func (p *Portfolio) Get() (*PortfolioSummary, error) {
	var pr portfolioResultsResp
	if err := p.c.GetJSON(robinhood.APIHost, "/portfolios/", &pr); err != nil {
		return nil, err
	}
	if len(pr.Results) == 0 {
		return nil, &robinhood.APIError{
			Code:    robinhood.CodeNotFound,
			Message: "no portfolios returned for this account",
		}
	}
	first := pr.Results[0]

	var uni unifiedAccountResp
	if err := p.c.GetJSON(robinhood.PhoenixHost, "/accounts/unified", &uni); err != nil {
		return nil, err
	}

	return &PortfolioSummary{
		Equity:              first.Equity,
		ExtendedHoursEquity: first.ExtendedHoursEquity,
		MarketValue:         first.MarketValue,
		LastCoreEquity:      first.LastCoreEquity,
		Cash:                uni.Cash,
		BuyingPower:         uni.BuyingPower,
	}, nil
}
