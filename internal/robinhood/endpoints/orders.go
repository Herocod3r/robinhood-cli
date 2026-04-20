package endpoints

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

// OrdersEP wraps /orders/.
type OrdersEP struct{ c *robinhood.Client }

// NewOrders constructs an OrdersEP.
func NewOrders(c *robinhood.Client) *OrdersEP { return &OrdersEP{c: c} }

// Order is one equity order row.
type Order struct {
	ID               string          `json:"id"`
	Symbol           string          `json:"symbol"`
	Side             string          `json:"side"`
	Type             string          `json:"type"`
	State            string          `json:"state"`
	Quantity         robinhood.Money `json:"quantity"`
	Price            robinhood.Money `json:"price,omitempty"`
	AverageFillPrice robinhood.Money `json:"average_fill_price,omitempty"`
	Fees             robinhood.Money `json:"fees"`
	TimeInForce      string          `json:"time_in_force"`
	ExtendedHours    bool            `json:"extended_hours"`
	CreatedAt        string          `json:"created_at"`
	FilledAt         string          `json:"filled_at,omitempty"`
	CancelledAt      string          `json:"cancelled_at,omitempty"`
	InstrumentURL    string          `json:"instrument,omitempty"`
}

type rawOrder struct {
	ID               string          `json:"id"`
	Side             string          `json:"side"`
	Type             string          `json:"type"`
	State            string          `json:"state"`
	Quantity         robinhood.Money `json:"quantity"`
	Price            robinhood.Money `json:"price"`
	AverageFillPrice robinhood.Money `json:"average_price"`
	Fees             robinhood.Money `json:"fees"`
	TimeInForce      string          `json:"time_in_force"`
	ExtendedHours    bool            `json:"extended_hours"`
	CreatedAt        string          `json:"created_at"`
	UpdatedAt        string          `json:"updated_at"`
	Instrument       string          `json:"instrument"`
	Executions       []struct {
		SettlementDate string `json:"settlement_date"`
		Timestamp      string `json:"timestamp"`
	} `json:"executions,omitempty"`
}

type ordersResp struct {
	Results []rawOrder `json:"results"`
	Next    string     `json:"next"`
}

// OrderOpts filters the results.
type OrderOpts struct {
	Since  time.Time
	Symbol string
	State  string
	Limit  int
}

// List fetches orders paginated + filtered.
func (o *OrdersEP) List(ctx context.Context, opts OrderOpts) ([]Order, error) {
	params := url.Values{}
	if opts.State != "" {
		params.Set("state", strings.ToLower(opts.State))
	}
	path := "/orders/"
	if q := params.Encode(); q != "" {
		path += "?" + q
	}
	var raws []rawOrder
	for {
		var resp ordersResp
		if err := o.c.GetJSONCtx(ctx, robinhood.APIHost, path, &resp); err != nil {
			return nil, err
		}
		raws = append(raws, resp.Results...)
		if resp.Next == "" {
			break
		}
		p, err := pathFromNext(resp.Next)
		if err != nil {
			return nil, err
		}
		path = p
		// Enforce since as an early cutoff: if the last order is older than
		// since AND orders come back newest-first, we can stop.
		if !opts.Since.IsZero() && len(raws) > 0 {
			if t, ok := parseOrderTime(raws[len(raws)-1].CreatedAt); ok && t.Before(opts.Since) {
				break
			}
		}
	}
	urls := make([]string, 0, len(raws))
	for _, r := range raws {
		urls = append(urls, r.Instrument)
	}
	sym, err := NewInstruments(o.c).ResolveByURL(ctx, urls)
	if err != nil {
		return nil, err
	}
	wantedSym := strings.ToUpper(strings.TrimSpace(opts.Symbol))
	out := make([]Order, 0, len(raws))
	for _, r := range raws {
		if !opts.Since.IsZero() {
			if t, ok := parseOrderTime(r.CreatedAt); ok && t.Before(opts.Since) {
				continue
			}
		}
		s := sym.Get(r.Instrument)
		if wantedSym != "" && s != wantedSym {
			continue
		}
		filled := ""
		if r.State == "filled" {
			filled = r.UpdatedAt
		}
		cancelled := ""
		if r.State == "cancelled" || r.State == "canceled" {
			cancelled = r.UpdatedAt
		}
		out = append(out, Order{
			ID:               r.ID,
			Symbol:           s,
			Side:             r.Side,
			Type:             r.Type,
			State:            r.State,
			Quantity:         r.Quantity,
			Price:            r.Price,
			AverageFillPrice: r.AverageFillPrice,
			Fees:             r.Fees,
			TimeInForce:      r.TimeInForce,
			ExtendedHours:    r.ExtendedHours,
			CreatedAt:        r.CreatedAt,
			FilledAt:         filled,
			CancelledAt:      cancelled,
			InstrumentURL:    r.Instrument,
		})
	}
	if opts.Limit > 0 && opts.Limit < len(out) {
		out = out[:opts.Limit]
	}
	return out, nil
}

// parseOrderTime accepts both RFC3339 with fractional seconds and plain date.
func parseOrderTime(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
