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
