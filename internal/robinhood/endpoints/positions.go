package endpoints

import (
	"context"
	"net/url"
	"sort"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

// Positions wraps the /positions/ endpoint.
type Positions struct{ c *robinhood.Client }

// NewPositions constructs a Positions endpoint.
func NewPositions(c *robinhood.Client) *Positions { return &Positions{c: c} }

// Position is one row in `rh positions`.
type Position struct {
	Symbol              string          `json:"symbol"`
	Name                string          `json:"name,omitempty"`
	Quantity            robinhood.Money `json:"quantity"`
	AverageBuyPrice     robinhood.Money `json:"average_buy_price"`
	LastPrice           robinhood.Money `json:"last_price"`
	MarketValue         robinhood.Money `json:"market_value"`
	CostBasis           robinhood.Money `json:"cost_basis"`
	UnrealizedPL        robinhood.Money `json:"unrealized_pl"`
	UnrealizedPLPercent robinhood.Money `json:"unrealized_pl_percent"`
	InstrumentID        string          `json:"instrument_id"`
	InstrumentURL       string          `json:"instrument_url,omitempty"`
}

type rawPosition struct {
	Quantity        robinhood.Money `json:"quantity"`
	AverageBuyPrice robinhood.Money `json:"average_buy_price"`
	Instrument      string          `json:"instrument"`
}

type positionsResp struct {
	Results []rawPosition `json:"results"`
	Next    string        `json:"next"`
}

// ListOpts controls filtering/ordering for `rh positions`.
type ListOpts struct {
	NonZero bool
	Sort    string // "value" | "pl" | "symbol"
	Limit   int
	Symbol  string // filter to a single ticker (for `rh position <ticker>`)
}

// List returns all positions for the current account, enriched with the live
// last price so unrealized P/L can be computed.
func (p *Positions) List(ctx context.Context, opts ListOpts) ([]Position, error) {
	all, err := p.fetchAll(ctx)
	if err != nil {
		return nil, err
	}
	if opts.NonZero {
		all = filterNonZero(all)
	}
	if len(all) == 0 {
		return []Position{}, nil
	}
	urls := make([]string, 0, len(all))
	for _, r := range all {
		urls = append(urls, r.Instrument)
	}
	inst := NewInstruments(p.c)
	syms, err := inst.ResolveByURL(ctx, urls)
	if err != nil {
		return nil, err
	}
	symbolList := syms.Symbols()
	// Optional symbol filter — applied after resolve, before quote.
	if opts.Symbol != "" {
		wanted := map[string]bool{opts.Symbol: true}
		filteredRaw := make([]rawPosition, 0, 1)
		for _, r := range all {
			if wanted[syms.Get(r.Instrument)] {
				filteredRaw = append(filteredRaw, r)
			}
		}
		all = filteredRaw
		symbolList = []string{opts.Symbol}
	}
	quotes := map[string]*Quote{}
	if len(symbolList) > 0 {
		q := NewQuotes(p.c)
		// Batch in MaxBatch-sized chunks.
		for start := 0; start < len(symbolList); start += MaxBatch {
			end := start + MaxBatch
			if end > len(symbolList) {
				end = len(symbolList)
			}
			chunk, qerr := q.Batch(ctx, symbolList[start:end])
			if qerr != nil {
				return nil, qerr
			}
			for k, v := range chunk {
				quotes[k] = v
			}
		}
	}
	out := make([]Position, 0, len(all))
	for _, r := range all {
		sym := syms.Get(r.Instrument)
		meta, _ := syms.Lookup(r.Instrument)
		pos := Position{
			Symbol:          sym,
			Name:            meta.Name,
			Quantity:        r.Quantity,
			AverageBuyPrice: r.AverageBuyPrice,
			InstrumentID:    instrumentIDFromURL(r.Instrument),
			InstrumentURL:   r.Instrument,
			CostBasis:       mulMoney(r.Quantity, r.AverageBuyPrice),
		}
		if qt, ok := quotes[sym]; ok && qt != nil {
			pos.LastPrice = qt.LastTradePrice
			pos.MarketValue = mulMoney(r.Quantity, qt.LastTradePrice)
			pos.UnrealizedPL = subMoney(pos.MarketValue, pos.CostBasis)
			pos.UnrealizedPLPercent = divMoney(pos.UnrealizedPL, pos.CostBasis)
		}
		out = append(out, pos)
	}
	sortPositions(out, opts.Sort)
	if opts.Limit > 0 && opts.Limit < len(out) {
		out = out[:opts.Limit]
	}
	return out, nil
}

func (p *Positions) fetchAll(ctx context.Context) ([]rawPosition, error) {
	var out []rawPosition
	path := "/positions/"
	for {
		var resp positionsResp
		if err := p.c.GetJSONCtx(ctx, robinhood.APIHost, path, &resp); err != nil {
			return nil, err
		}
		out = append(out, resp.Results...)
		if resp.Next == "" {
			break
		}
		u, err := url.Parse(resp.Next)
		if err != nil {
			return nil, err
		}
		path = u.Path
		if u.RawQuery != "" {
			path += "?" + u.RawQuery
		}
	}
	return out, nil
}

// filterNonZero drops positions with a zero quantity. Uses decimal math (Fix O).
func filterNonZero(in []rawPosition) []rawPosition {
	out := make([]rawPosition, 0, len(in))
	for _, r := range in {
		if !toDec(r.Quantity).IsZero() {
			out = append(out, r)
		}
	}
	return out
}

func sortPositions(ps []Position, by string) {
	switch by {
	case "pl":
		sort.SliceStable(ps, func(i, j int) bool {
			return moneyLess(ps[j].UnrealizedPL, ps[i].UnrealizedPL)
		})
	case "symbol":
		sort.SliceStable(ps, func(i, j int) bool { return ps[i].Symbol < ps[j].Symbol })
	default: // "value" or empty
		sort.SliceStable(ps, func(i, j int) bool {
			return moneyLess(ps[j].MarketValue, ps[i].MarketValue)
		})
	}
}
