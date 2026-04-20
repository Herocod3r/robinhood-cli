package endpoints

import (
	"context"
	"strings"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

// OptionsPositionsEP wraps /options/aggregate_positions/.
type OptionsPositionsEP struct{ c *robinhood.Client }

// NewOptionsPositions constructs an OptionsPositionsEP.
func NewOptionsPositions(c *robinhood.Client) *OptionsPositionsEP {
	return &OptionsPositionsEP{c: c}
}

// OptionPosition is one aggregated options position (e.g. one spread or single leg).
type OptionPosition struct {
	Symbol       string          `json:"symbol"`
	Strategy     string          `json:"strategy"`
	Quantity     robinhood.Money `json:"quantity"`
	AveragePrice robinhood.Money `json:"average_price"`
	Legs         []OptionLeg     `json:"legs"`
}

// OptionLeg is one leg inside an aggregate position.
type OptionLeg struct {
	OptionID     string          `json:"option_id"`
	Type         string          `json:"type"` // "call" | "put"
	StrikePrice  robinhood.Money `json:"strike_price"`
	Expiration   string          `json:"expiration"`
	PositionType string          `json:"position_type"` // "long" | "short"
}

type rawOptionLeg struct {
	Option       string `json:"option"`
	PositionType string `json:"position_type"`
}

type rawOptionAgg struct {
	Symbol       string          `json:"symbol"`
	ChainSymbol  string          `json:"chain_symbol"`
	Strategy     string          `json:"strategy"`
	Quantity     robinhood.Money `json:"quantity"`
	AveragePrice robinhood.Money `json:"average_open_price"`
	Legs         []rawOptionLeg  `json:"legs"`
}

type optAggResp struct {
	Results []rawOptionAgg `json:"results"`
	Next    string         `json:"next"`
}

type optInstrument struct {
	ID             string          `json:"id"`
	URL            string          `json:"url"`
	ChainSymbol    string          `json:"chain_symbol"`
	StrikePrice    robinhood.Money `json:"strike_price"`
	Type           string          `json:"type"`
	ExpirationDate string          `json:"expiration_date"`
}

// List returns all non-zero aggregated options positions; symbol "" returns all.
func (o *OptionsPositionsEP) List(ctx context.Context, symbol string) ([]OptionPosition, error) {
	wanted := strings.ToUpper(strings.TrimSpace(symbol))
	path := "/options/aggregate_positions/?nonzero=true"
	var aggs []rawOptionAgg
	for {
		var resp optAggResp
		if err := o.c.GetJSONCtx(ctx, robinhood.APIHost, path, &resp); err != nil {
			return nil, err
		}
		aggs = append(aggs, resp.Results...)
		if resp.Next == "" {
			break
		}
		p, err := pathFromNext(resp.Next)
		if err != nil {
			return nil, err
		}
		path = p
	}
	// Expand each leg's option URL.
	optCache := map[string]*optInstrument{}
	out := make([]OptionPosition, 0, len(aggs))
	for _, a := range aggs {
		sym := a.ChainSymbol
		if sym == "" {
			sym = a.Symbol
		}
		if wanted != "" && strings.ToUpper(sym) != wanted {
			continue
		}
		legs := make([]OptionLeg, 0, len(a.Legs))
		for _, raw := range a.Legs {
			inst, err := o.fetchOption(ctx, raw.Option, optCache)
			if err != nil {
				return nil, err
			}
			leg := OptionLeg{
				OptionID:     instrumentIDFromURL(raw.Option),
				PositionType: raw.PositionType,
			}
			if inst != nil {
				leg.Type = inst.Type
				leg.StrikePrice = inst.StrikePrice
				leg.Expiration = inst.ExpirationDate
			}
			legs = append(legs, leg)
		}
		out = append(out, OptionPosition{
			Symbol:       sym,
			Strategy:     a.Strategy,
			Quantity:     a.Quantity,
			AveragePrice: a.AveragePrice,
			Legs:         legs,
		})
	}
	return out, nil
}

// fetchOption resolves a leg URL to its option metadata, caching per-call.
func (o *OptionsPositionsEP) fetchOption(ctx context.Context, optionURL string, cache map[string]*optInstrument) (*optInstrument, error) {
	if optionURL == "" {
		return nil, nil
	}
	if cached, ok := cache[optionURL]; ok {
		return cached, nil
	}
	p, err := pathFromNext(optionURL)
	if err != nil {
		return nil, err
	}
	var inst optInstrument
	if err := o.c.GetJSONCtx(ctx, robinhood.APIHost, p, &inst); err != nil {
		return nil, err
	}
	cache[optionURL] = &inst
	return &inst, nil
}
