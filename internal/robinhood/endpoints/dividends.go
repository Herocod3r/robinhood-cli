package endpoints

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

// DividendsEP wraps /dividends/.
type DividendsEP struct{ c *robinhood.Client }

// NewDividends constructs a DividendsEP.
func NewDividends(c *robinhood.Client) *DividendsEP { return &DividendsEP{c: c} }

// Dividend is one dividend event (paid/pending/reinvested).
type Dividend struct {
	Symbol         string          `json:"symbol,omitempty"`
	Amount         robinhood.Money `json:"amount"`
	Rate           robinhood.Money `json:"rate"`
	Position       robinhood.Money `json:"position"`
	WithholdingTax robinhood.Money `json:"withholding,omitempty"`
	PaidAt         string          `json:"paid_at"`
	RecordAt       string          `json:"record_date"`
	PayableAt      string          `json:"payable_date"`
	InstrumentURL  string          `json:"instrument_url,omitempty"`
	InstrumentID   string          `json:"instrument_id,omitempty"`
	State          string          `json:"state"`
}

type rawDividend struct {
	Amount         robinhood.Money `json:"amount"`
	Rate           robinhood.Money `json:"rate"`
	Position       robinhood.Money `json:"position"`
	Withholding    robinhood.Money `json:"withholding"`
	PaidAt         string          `json:"paid_at"`
	RecordDate     string          `json:"record_date"`
	PayableDate    string          `json:"payable_date"`
	Instrument     string          `json:"instrument"`
	State          string          `json:"state"`
}

type dividendsResp struct {
	Results []rawDividend `json:"results"`
	Next    string        `json:"next"`
}

// ListOpts filters what List returns.
type DividendOpts struct {
	Year   int    // 0 = no year filter
	Symbol string // "" = no symbol filter
}

// List returns dividend events, optionally filtered by year + symbol.
func (d *DividendsEP) List(ctx context.Context, opts DividendOpts) ([]Dividend, error) {
	raws, err := d.fetchAll(ctx)
	if err != nil {
		return nil, err
	}
	// Resolve instrument URLs to symbols.
	urls := make([]string, 0, len(raws))
	for _, r := range raws {
		urls = append(urls, r.Instrument)
	}
	m, err := NewInstruments(d.c).ResolveByURL(ctx, urls)
	if err != nil {
		return nil, err
	}
	out := make([]Dividend, 0, len(raws))
	wantedSym := strings.ToUpper(strings.TrimSpace(opts.Symbol))
	for _, r := range raws {
		sym := m.Get(r.Instrument)
		if wantedSym != "" && sym != wantedSym {
			continue
		}
		if opts.Year != 0 {
			if yr := parseYearFromDate(r.PaidAt); yr != opts.Year {
				continue
			}
		}
		out = append(out, Dividend{
			Symbol:         sym,
			Amount:         r.Amount,
			Rate:           r.Rate,
			Position:       r.Position,
			WithholdingTax: r.Withholding,
			PaidAt:         r.PaidAt,
			RecordAt:       r.RecordDate,
			PayableAt:      r.PayableDate,
			InstrumentURL:  r.Instrument,
			InstrumentID:   instrumentIDFromURL(r.Instrument),
			State:          r.State,
		})
	}
	return out, nil
}

func (d *DividendsEP) fetchAll(ctx context.Context) ([]rawDividend, error) {
	var out []rawDividend
	path := "/dividends/"
	for {
		var resp dividendsResp
		if err := d.c.GetJSONCtx(ctx, robinhood.APIHost, path, &resp); err != nil {
			return nil, err
		}
		out = append(out, resp.Results...)
		if resp.Next == "" {
			break
		}
		p, err := pathFromNext(resp.Next)
		if err != nil {
			return nil, err
		}
		path = p
	}
	return out, nil
}

// parseYearFromDate extracts YYYY from an RFC3339 or YYYY-MM-DD string; 0 on error.
func parseYearFromDate(s string) int {
	if s == "" {
		return 0
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.Year()
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t.Year()
	}
	// Fallback: leading 4 chars.
	if len(s) >= 4 {
		n, err := strconv.Atoi(s[:4])
		if err == nil {
			return n
		}
	}
	return 0
}
