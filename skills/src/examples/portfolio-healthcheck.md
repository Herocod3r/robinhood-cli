# Example: portfolio health check

**User:** How's my portfolio looking today?

**Claude (internal steps):**

```bash
rh portfolio --json
rh positions --nonzero --sort value --json
```

`rh portfolio --json` returns:
```json
{ "schema": "robinhood-cli/v1", "command": "portfolio",
  "generated_at": "2026-04-20T15:00:00Z",
  "data": {"equity": "98234.50", "day_change": "342.15", "day_change_percent": "0.0035",
           "cash": "1784.40", "buying_power": "3568.80"}, "error": null }
```

`rh positions --nonzero --sort value --json` returns the top positions:
```json
{ "schema": "robinhood-cli/v1", "command": "positions",
  "generated_at": "2026-04-20T15:00:00Z",
  "data": [
    {"symbol":"AAPL","quantity":"200","market_value":"32400.00","day_change_percent":"0.0080"},
    {"symbol":"NVDA","quantity":"25", "market_value":"18700.00","day_change_percent":"-0.0120"},
    {"symbol":"SPY", "quantity":"30", "market_value":"14900.00","day_change_percent":"0.0040"}
  ],
  "meta": {"count":3,"profile":"default"}, "error": null }
```

**Claude's response to user:**

> Your portfolio is $98,234.50 as of 15:00 UTC — up $342.15 (+0.35%) today.
>
> Top positions by market value:
> - AAPL: $32,400 (33% of portfolio) — up 0.8% today
> - NVDA: $18,700 (19%) — down 1.2%
> - SPY: $14,900 (15%) — up 0.4%
>
> Concentration flag: AAPL is 33% of your portfolio. Single-name risk is high.
>
> You have $1,784.40 in cash and $3,568.80 of buying power.

## Why this shape works

- Answer first (dollar figure + freshness), details second.
- Percentages are computed locally from the same JSON, not from a third call.
- Concentration alerts use the `references/sector-concentration.md`
  thresholds (>20% single-name, >40% single-sector).
- No buy/sell recommendation is offered. Facts only.
