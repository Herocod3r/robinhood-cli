package endpoints

import "context"

// SearchResult is the flattened Instruments.Search row for `rh search`.
type SearchResult struct {
	Symbol       string `json:"symbol"`
	Name         string `json:"name"`
	InstrumentID string `json:"instrument_id"`
	Tradeable    bool   `json:"tradeable"`
	Type         string `json:"type"`
}

// Search runs a free-text search and returns the typed rows.
func Search(ctx context.Context, i *Instruments, q string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := i.Search(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	out := make([]SearchResult, 0, len(rows))
	for _, r := range rows {
		n := r.Name
		if n == "" {
			n = r.SimpleName
		}
		out = append(out, SearchResult{
			Symbol:       r.Symbol,
			Name:         n,
			InstrumentID: r.ID,
			Tradeable:    r.Tradeable,
			Type:         r.InstrumentType,
		})
	}
	return out, nil
}
