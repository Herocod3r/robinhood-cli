package output

import (
	"fmt"
	"io"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
	"github.com/herocod3r/robinhood-cli/internal/robinhood/endpoints"
	"github.com/olekukonko/tablewriter"
)

// TableWriter renders structured data as ASCII tables suitable for a terminal.
type TableWriter struct {
	Out io.Writer
}

// WritePortfolio renders a portfolio summary as a key/value table.
func (w *TableWriter) WritePortfolio(p *endpoints.PortfolioSummary) error {
	t := tablewriter.NewWriter(w.Out)
	t.SetHeader([]string{"Field", "Value"})
	t.SetBorder(false)
	t.Append([]string{"Equity", string(p.Equity)})
	t.Append([]string{"Extended hours equity", string(p.ExtendedHoursEquity)})
	t.Append([]string{"Market value", string(p.MarketValue)})
	t.Append([]string{"Cash", string(p.Cash)})
	t.Append([]string{"Buying power", string(p.BuyingPower)})
	t.Render()
	return nil
}

// WritePositions renders the positions table with live quotes + P/L columns.
func (w *TableWriter) WritePositions(ps []endpoints.Position) error {
	t := tablewriter.NewWriter(w.Out)
	t.SetHeader([]string{"Symbol", "Qty", "Last", "Market Value", "Cost Basis", "P/L", "P/L %"})
	t.SetBorder(false)
	for _, p := range ps {
		t.Append([]string{
			p.Symbol,
			string(p.Quantity),
			string(p.LastPrice),
			string(p.MarketValue),
			string(p.CostBasis),
			string(p.UnrealizedPL),
			string(p.UnrealizedPLPercent),
		})
	}
	t.Render()
	return nil
}

// WriteQuotes renders a slice of quotes as a compact one-row-per-symbol table.
func (w *TableWriter) WriteQuotes(qs []*endpoints.Quote, extended bool) error {
	t := tablewriter.NewWriter(w.Out)
	headers := []string{"Symbol", "Last", "Bid", "Ask", "Prev Close", "Change", "Change %"}
	if extended {
		headers = append(headers, "Ext Hours")
	}
	t.SetHeader(headers)
	t.SetBorder(false)
	for _, q := range qs {
		change := subDecimal(q.LastTradePrice, q.PreviousClose)
		changePct := divPercent(q.LastTradePrice, q.PreviousClose)
		row := []string{
			q.Symbol,
			string(q.LastTradePrice),
			string(q.BidPrice),
			string(q.AskPrice),
			string(q.PreviousClose),
			change,
			changePct,
		}
		if extended {
			row = append(row, string(q.ExtendedHoursPrice))
		}
		t.Append(row)
	}
	t.Render()
	return nil
}

// WriteAccount renders AccountSummary as a key/value table.
func (w *TableWriter) WriteAccount(a *endpoints.AccountSummary) error {
	t := tablewriter.NewWriter(w.Out)
	t.SetHeader([]string{"Field", "Value"})
	t.SetBorder(false)
	t.Append([]string{"Account", a.AccountNumber})
	t.Append([]string{"Buying power", string(a.BuyingPower)})
	t.Append([]string{"Cash", string(a.Cash)})
	t.Append([]string{"Margin balance", string(a.MarginBalance)})
	t.Append([]string{"Sweep enabled", fmt.Sprintf("%t", a.SweepEnabled)})
	t.Append([]string{"Pattern day trader", fmt.Sprintf("%t", a.PatternDayTrader)})
	t.Append([]string{"Day trade count", fmt.Sprintf("%d", a.DayTradeCount)})
	t.Append([]string{"Instant used", string(a.InstantUsed)})
	t.Append([]string{"Instant available", string(a.InstantAvailable)})
	t.Render()
	return nil
}

// WriteFundamentals renders a subset of fundamentals columns.
func (w *TableWriter) WriteFundamentals(rows []endpoints.Fundamentals) error {
	t := tablewriter.NewWriter(w.Out)
	t.SetHeader([]string{"Symbol", "Open", "High", "Low", "52wk Hi", "52wk Lo", "Mkt Cap", "P/E", "Div Yield"})
	t.SetBorder(false)
	for _, r := range rows {
		t.Append([]string{
			r.Symbol,
			string(r.Open),
			string(r.High),
			string(r.Low),
			string(r.High52Weeks),
			string(r.Low52Weeks),
			string(r.MarketCap),
			string(r.PERatio),
			string(r.DividendYield),
		})
	}
	t.Render()
	return nil
}

// WriteHistoricals renders OHLCV bars; most recent last.
func (w *TableWriter) WriteHistoricals(h *endpoints.Historicals) error {
	t := tablewriter.NewWriter(w.Out)
	t.SetHeader([]string{"Begin", "O", "H", "L", "C", "Vol"})
	t.SetBorder(false)
	for _, b := range h.Bars {
		t.Append([]string{
			b.BeginsAt,
			string(b.OpenPrice),
			string(b.HighPrice),
			string(b.LowPrice),
			string(b.ClosePrice),
			fmt.Sprintf("%d", b.Volume),
		})
	}
	t.Render()
	return nil
}

// WriteNews renders news items as a short table (title truncated).
func (w *TableWriter) WriteNews(items []endpoints.NewsItem) error {
	t := tablewriter.NewWriter(w.Out)
	t.SetHeader([]string{"Published", "Source", "Title"})
	t.SetBorder(false)
	for _, it := range items {
		t.Append([]string{
			truncate(it.PublishedAt, 19),
			truncate(it.Source, 16),
			truncate(it.Title, 80),
		})
	}
	t.Render()
	return nil
}

// WriteError renders a CLI-friendly error line. Always writes to Out even if that
// is stderr — caller decides.
func (w *TableWriter) WriteError(command string, err error) error {
	if apiErr, ok := err.(*robinhood.APIError); ok {
		_, perr := fmt.Fprintf(w.Out, "error (%s): %s\n", apiErr.Code, apiErr.Message)
		if perr != nil {
			return perr
		}
		if apiErr.Hint != "" {
			_, _ = fmt.Fprintf(w.Out, "hint: %s\n", apiErr.Hint)
		}
		return nil
	}
	_, perr := fmt.Fprintf(w.Out, "error (%s): %v\n", command, err)
	return perr
}
