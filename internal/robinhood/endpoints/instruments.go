package endpoints

import (
	"context"
	"net/url"
	"strings"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

// Instruments resolves Robinhood instrument IDs / URLs to symbols and names.
// It batches requests via `GET /instruments/?ids=<csv>` and caches resolutions
// for the lifetime of the struct.
type Instruments struct {
	c *robinhood.Client
}

// NewInstruments constructs an Instruments resolver.
func NewInstruments(c *robinhood.Client) *Instruments { return &Instruments{c: c} }

// Instrument is the subset of the Robinhood /instruments/ payload we use.
type Instrument struct {
	ID             string `json:"id"`
	URL            string `json:"url"`
	Symbol         string `json:"symbol"`
	Name           string `json:"name,omitempty"`
	SimpleName     string `json:"simple_name,omitempty"`
	Tradeable      bool   `json:"tradeable"`
	InstrumentType string `json:"type,omitempty"`
}

type instrumentsResp struct {
	Results []Instrument `json:"results"`
	Next    string       `json:"next"`
}

// InstrumentMap holds a batch of resolved instruments and indexes them by
// both their UUID and their canonical URL.
type InstrumentMap struct {
	byID  map[string]Instrument
	byURL map[string]Instrument
}

// Get returns the symbol for an instrument URL or UUID; empty string on miss.
func (m *InstrumentMap) Get(idOrURL string) string {
	if m == nil {
		return ""
	}
	if inst, ok := m.byURL[idOrURL]; ok {
		return inst.Symbol
	}
	if inst, ok := m.byID[idOrURL]; ok {
		return inst.Symbol
	}
	// Allow callers to pass a URL whose path ends with /<uuid>/.
	if uid := instrumentIDFromURL(idOrURL); uid != "" {
		if inst, ok := m.byID[uid]; ok {
			return inst.Symbol
		}
	}
	return ""
}

// Lookup returns the full instrument record (zero value + false on miss).
func (m *InstrumentMap) Lookup(idOrURL string) (Instrument, bool) {
	if m == nil {
		return Instrument{}, false
	}
	if inst, ok := m.byURL[idOrURL]; ok {
		return inst, true
	}
	if inst, ok := m.byID[idOrURL]; ok {
		return inst, true
	}
	if uid := instrumentIDFromURL(idOrURL); uid != "" {
		if inst, ok := m.byID[uid]; ok {
			return inst, true
		}
	}
	return Instrument{}, false
}

// Symbols returns the deduped list of symbols in the map.
func (m *InstrumentMap) Symbols() []string {
	if m == nil {
		return nil
	}
	seen := make(map[string]bool, len(m.byID))
	out := make([]string, 0, len(m.byID))
	for _, inst := range m.byID {
		if inst.Symbol == "" || seen[inst.Symbol] {
			continue
		}
		seen[inst.Symbol] = true
		out = append(out, inst.Symbol)
	}
	return out
}

// ResolveIDs fetches instrument metadata for every unique UUID in ids.
// It batches using /instruments/?ids=<csv>.
func (i *Instruments) ResolveIDs(ctx context.Context, ids []string) (*InstrumentMap, error) {
	unique := dedup(ids)
	m := &InstrumentMap{byID: map[string]Instrument{}, byURL: map[string]Instrument{}}
	if len(unique) == 0 {
		return m, nil
	}
	// Keep URLs below typical 8K limit.
	const batchSize = 50
	for start := 0; start < len(unique); start += batchSize {
		end := start + batchSize
		if end > len(unique) {
			end = len(unique)
		}
		path := "/instruments/?ids=" + url.QueryEscape(strings.Join(unique[start:end], ","))
		if err := i.fetchPage(ctx, path, m); err != nil {
			return nil, err
		}
	}
	return m, nil
}

// ResolveByURL fetches instrument metadata given a list of fully-qualified
// instrument URLs (e.g. returned by /positions/). It extracts the UUID from
// each URL and delegates to ResolveIDs.
func (i *Instruments) ResolveByURL(ctx context.Context, urls []string) (*InstrumentMap, error) {
	ids := make([]string, 0, len(urls))
	for _, u := range urls {
		if id := instrumentIDFromURL(u); id != "" {
			ids = append(ids, id)
		}
	}
	return i.ResolveIDs(ctx, ids)
}

// Search runs a free-text query against /instruments/?query=.
func (i *Instruments) Search(ctx context.Context, q string, limit int) ([]Instrument, error) {
	params := url.Values{}
	params.Set("query", q)
	var resp instrumentsResp
	if err := i.c.GetJSONCtx(ctx, robinhood.APIHost, "/instruments/?"+params.Encode(), &resp); err != nil {
		return nil, err
	}
	if limit > 0 && limit < len(resp.Results) {
		resp.Results = resp.Results[:limit]
	}
	return resp.Results, nil
}

// fetchPage follows resp.Next until empty, storing all rows in m.
func (i *Instruments) fetchPage(ctx context.Context, path string, m *InstrumentMap) error {
	for {
		var resp instrumentsResp
		if err := i.c.GetJSONCtx(ctx, robinhood.APIHost, path, &resp); err != nil {
			return err
		}
		for _, inst := range resp.Results {
			if inst.ID != "" {
				m.byID[inst.ID] = inst
			}
			if inst.URL != "" {
				m.byURL[inst.URL] = inst
			}
		}
		if resp.Next == "" {
			return nil
		}
		np, err := pathFromNext(resp.Next)
		if err != nil {
			return err
		}
		path = np
	}
}

// instrumentIDFromURL strips the UUID from a URL whose path ends /instruments/<uuid>/.
// It returns "" when the path doesn't match.
func instrumentIDFromURL(u string) string {
	if u == "" {
		return ""
	}
	parsed, err := url.Parse(u)
	if err != nil {
		return ""
	}
	p := strings.TrimSuffix(parsed.Path, "/")
	slash := strings.LastIndex(p, "/")
	if slash < 0 {
		return ""
	}
	return p[slash+1:]
}

// pathFromNext returns the path+query part of an absolute Robinhood URL.
// The rh Client prepends the base host; using the relative portion ensures
// pagination hits the host the test/prod client is configured for.
func pathFromNext(next string) (string, error) {
	u, err := url.Parse(next)
	if err != nil {
		return "", err
	}
	p := u.Path
	if u.RawQuery != "" {
		p += "?" + u.RawQuery
	}
	return p, nil
}

// dedup removes empty strings and duplicates while preserving order.
func dedup(in []string) []string {
	seen := make(map[string]bool, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}
