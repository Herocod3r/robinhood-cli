package endpoints

import (
	"context"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

// Account wraps the accounts endpoints.
type Account struct{ c *robinhood.Client }

// NewAccount constructs an Account endpoint.
func NewAccount(c *robinhood.Client) *Account { return &Account{c: c} }

// AccountSummary is the merged view for `rh account`.
type AccountSummary struct {
	AccountNumber    string          `json:"account_number"`
	BuyingPower      robinhood.Money `json:"buying_power"`
	Cash             robinhood.Money `json:"cash"`
	SweepEnabled     bool            `json:"sweep_enabled"`
	MarginBalance    robinhood.Money `json:"margin_balance,omitempty"`
	PatternDayTrader bool            `json:"pattern_day_trader"`
	DayTradeCount    int             `json:"day_trade_count"`
	InstantUsed      robinhood.Money `json:"instant_used"`
	InstantAvailable robinhood.Money `json:"instant_available"`
}

type accountsResp struct {
	Results []struct {
		AccountNumber    string          `json:"account_number"`
		BuyingPower      robinhood.Money `json:"buying_power"`
		Cash             robinhood.Money `json:"cash"`
		CashHeldForOrders robinhood.Money `json:"cash_held_for_orders"`
		SweepEnabled     bool            `json:"sweep_enabled"`
		MarginBalances   *struct {
			DayTradeCount     int             `json:"day_trade_count"`
			MarginBalance     robinhood.Money `json:"margin_balance"`
			UnallocatedMargin robinhood.Money `json:"unallocated_margin_cash"`
		} `json:"margin_balances,omitempty"`
		InstantEligibility *struct {
			Reason string `json:"reason"`
		} `json:"instant_eligibility,omitempty"`
		PatternDayTrader bool            `json:"is_pinnacle_account"` // best-effort
		InstantUsed      robinhood.Money `json:"uncleared_deposits"`
		InstantAvailable robinhood.Money `json:"unsettled_funds"`
	} `json:"results"`
}

type unifiedAcctResp struct {
	BuyingPower    robinhood.Money `json:"buying_power"`
	Cash           robinhood.Money `json:"cash"`
	PatternDay     bool            `json:"pattern_day_trader"`
	DayTradeCount  int             `json:"day_trade_count"`
}

// Get returns a merged AccountSummary.
func (a *Account) Get(ctx context.Context) (*AccountSummary, error) {
	var ar accountsResp
	if err := a.c.GetJSONCtx(ctx, robinhood.APIHost, "/accounts/", &ar); err != nil {
		return nil, err
	}
	if len(ar.Results) == 0 {
		return nil, &robinhood.APIError{Code: robinhood.CodeNotFound, Message: "no account"}
	}
	first := ar.Results[0]

	sum := &AccountSummary{
		AccountNumber:    first.AccountNumber,
		BuyingPower:      first.BuyingPower,
		Cash:             first.Cash,
		SweepEnabled:     first.SweepEnabled,
		PatternDayTrader: first.PatternDayTrader,
		InstantUsed:      first.InstantUsed,
		InstantAvailable: first.InstantAvailable,
	}
	if first.MarginBalances != nil {
		sum.MarginBalance = first.MarginBalances.MarginBalance
		sum.DayTradeCount = first.MarginBalances.DayTradeCount
	}

	// Unified enrichment is best-effort; ignore failures.
	var uni unifiedAcctResp
	if err := a.c.GetJSONCtx(ctx, robinhood.PhoenixHost, "/accounts/unified", &uni); err == nil {
		if uni.BuyingPower != "" {
			sum.BuyingPower = uni.BuyingPower
		}
		if uni.Cash != "" {
			sum.Cash = uni.Cash
		}
		if uni.PatternDay {
			sum.PatternDayTrader = true
		}
		if uni.DayTradeCount > sum.DayTradeCount {
			sum.DayTradeCount = uni.DayTradeCount
		}
	}
	return sum, nil
}
