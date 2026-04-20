package endpoints

import (
	"context"
	"strings"
	"time"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

// MarketsEP wraps /markets/.
type MarketsEP struct{ c *robinhood.Client }

// NewMarkets constructs a MarketsEP.
func NewMarkets(c *robinhood.Client) *MarketsEP { return &MarketsEP{c: c} }

// MarketHours is the merged hours for one market on one date.
type MarketHours struct {
	MIC           string `json:"mic"`
	Name          string `json:"name"`
	Date          string `json:"date"`
	IsOpen        bool   `json:"is_open"`
	OpensAt       string `json:"opens_at"`
	ClosesAt      string `json:"closes_at"`
	ExtendedOpen  string `json:"extended_opens_at,omitempty"`
	ExtendedClose string `json:"extended_closes_at,omitempty"`
}

type marketListing struct {
	Results []struct {
		MIC       string `json:"mic"`
		Name      string `json:"name"`
		Acronym   string `json:"acronym"`
		TodaysURL string `json:"todays_hours"`
	} `json:"results"`
}

type marketHoursResp struct {
	Date          string `json:"date"`
	IsOpen        bool   `json:"is_open"`
	OpensAt       string `json:"opens_at"`
	ClosesAt      string `json:"closes_at"`
	ExtendedOpen  string `json:"extended_opens_at"`
	ExtendedClose string `json:"extended_closes_at"`
}

// List returns market hours for `date` (or today) across all markets.
func (m *MarketsEP) List(ctx context.Context, date string) ([]MarketHours, error) {
	if strings.TrimSpace(date) == "" {
		date = time.Now().UTC().Format("2006-01-02")
	}
	var idx marketListing
	if err := m.c.GetJSONCtx(ctx, robinhood.APIHost, "/markets/", &idx); err != nil {
		return nil, err
	}
	out := make([]MarketHours, 0, len(idx.Results))
	for _, mk := range idx.Results {
		path := "/markets/" + mk.MIC + "/hours/" + date + "/"
		var hr marketHoursResp
		if err := m.c.GetJSONCtx(ctx, robinhood.APIHost, path, &hr); err != nil {
			continue // best-effort: skip markets with hours unavailable
		}
		out = append(out, MarketHours{
			MIC:           mk.MIC,
			Name:          mk.Name,
			Date:          hr.Date,
			IsOpen:        hr.IsOpen,
			OpensAt:       hr.OpensAt,
			ClosesAt:      hr.ClosesAt,
			ExtendedOpen:  hr.ExtendedOpen,
			ExtendedClose: hr.ExtendedClose,
		})
	}
	return out, nil
}
