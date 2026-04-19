package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/herocod3r/robinhood-cli/internal/robinhood/endpoints"
)

func TestTable_Portfolio(t *testing.T) {
	var buf bytes.Buffer
	w := &TableWriter{Out: &buf}

	if err := w.WritePortfolio(&endpoints.PortfolioSummary{
		Equity:              "98234.50",
		ExtendedHoursEquity: "98240.12",
		MarketValue:         "72450.10",
		Cash:                "25784.40",
		BuyingPower:         "25784.40",
	}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"Equity", "Cash", "98234.50", "25784.40"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output, got:\n%s", want, out)
		}
	}
}

func TestTable_Error(t *testing.T) {
	var buf bytes.Buffer
	w := &TableWriter{Out: &buf}
	err := w.WriteError("portfolio", &testErr{msg: "boom", hint: "run: rh login"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "boom") {
		t.Errorf("expected 'boom' in error output: %s", buf.String())
	}
}

type testErr struct{ msg, hint string }

func (e *testErr) Error() string { return e.msg }
