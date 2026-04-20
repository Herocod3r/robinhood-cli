package endpoints

import (
	"context"
	"strings"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

// RatingsEP wraps /midlands/ratings/.
type RatingsEP struct{ c *robinhood.Client }

// NewRatings constructs a RatingsEP.
func NewRatings(c *robinhood.Client) *RatingsEP { return &RatingsEP{c: c} }

// Rating is the merged view for one symbol's ratings.
type Rating struct {
	Symbol  string `json:"symbol"`
	Summary struct {
		NumBuy  int `json:"num_buy_ratings"`
		NumHold int `json:"num_hold_ratings"`
		NumSell int `json:"num_sell_ratings"`
	} `json:"summary"`
	Ratings []RatingComment `json:"ratings"`
}

// RatingComment is one analyst comment.
type RatingComment struct {
	Type        string `json:"type"`
	Text        string `json:"text"`
	PublishedAt string `json:"published_at"`
}

type rawRatings struct {
	Summary struct {
		NumBuyRatings  int `json:"num_buy_ratings"`
		NumHoldRatings int `json:"num_hold_ratings"`
		NumSellRatings int `json:"num_sell_ratings"`
	} `json:"summary"`
	Ratings []RatingComment `json:"ratings"`
}

// Get resolves the symbol to an instrument UUID, then fetches its rating.
func (r *RatingsEP) Get(ctx context.Context, symbol string) (*Rating, error) {
	sym := strings.ToUpper(strings.TrimSpace(symbol))
	if sym == "" {
		return nil, &robinhood.APIError{Code: robinhood.CodeValidation, Message: "symbol required"}
	}
	// Robinhood's /midlands/ratings/ is keyed by instrument UUID; resolve first.
	results, err := NewInstruments(r.c).Search(ctx, sym, 10)
	if err != nil {
		return nil, err
	}
	var id string
	for _, res := range results {
		if strings.EqualFold(res.Symbol, sym) {
			id = res.ID
			break
		}
	}
	if id == "" {
		return nil, &robinhood.APIError{Code: robinhood.CodeNotFound, Message: "no instrument for " + sym}
	}
	var raw rawRatings
	if err := r.c.GetJSONCtx(ctx, robinhood.APIHost, "/midlands/ratings/"+id+"/", &raw); err != nil {
		return nil, err
	}
	out := &Rating{Symbol: sym, Ratings: raw.Ratings}
	out.Summary.NumBuy = raw.Summary.NumBuyRatings
	out.Summary.NumHold = raw.Summary.NumHoldRatings
	out.Summary.NumSell = raw.Summary.NumSellRatings
	return out, nil
}
