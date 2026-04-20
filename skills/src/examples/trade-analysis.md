# Example: trade analysis (realized P&L, FIFO)

**User:** Analyze my trades this year.

**Claude (internal steps):**

```bash
rh orders --since 2026-01-01 --state filled --json
```

Sample (sanitized, truncated) response:
```json
{ "schema": "robinhood-cli/v1", "command": "orders",
  "generated_at": "2026-04-20T15:05:00Z",
  "data": [
    {"symbol":"AAPL","side":"buy", "quantity":"100","average_fill_price":"170.00","fees":"0.00","filled_at":"2026-01-15T14:30:00Z"},
    {"symbol":"AAPL","side":"buy", "quantity":"50", "average_fill_price":"180.00","fees":"0.00","filled_at":"2026-02-10T14:30:00Z"},
    {"symbol":"AAPL","side":"sell","quantity":"75", "average_fill_price":"195.00","fees":"0.00","filled_at":"2026-03-20T14:30:00Z"},
    {"symbol":"NVDA","side":"buy", "quantity":"20", "average_fill_price":"820.00","fees":"0.00","filled_at":"2026-02-01T14:30:00Z"},
    {"symbol":"NVDA","side":"sell","quantity":"10", "average_fill_price":"900.00","fees":"0.00","filled_at":"2026-04-01T14:30:00Z"}
  ],
  "meta":{"count":5,"profile":"default"}, "error": null }
```

**FIFO pairing (worked example):**

AAPL:
1. Buy 100 @ $170 — lot A (100 remaining, cost $170).
2. Buy 50 @ $180 — lot B (50 remaining, cost $180).
3. Sell 75 @ $195 — consume lot A entirely? No, 75 < 100: take 75 from lot A.
   - Realized = (195 - 170) * 75 = **$1,875**
   - Lot A now has 25 remaining.

NVDA:
1. Buy 20 @ $820 — lot A (20 remaining, cost $820).
2. Sell 10 @ $900 — take 10 from lot A.
   - Realized = (900 - 820) * 10 = **$800**

**Claude's response to user:**

> Year-to-date realized P&L (through 2026-04-20):
>
> | Symbol | Realized P&L |
> |---|---|
> | AAPL | +$1,875 |
> | NVDA | +$800   |
> | **Total** | **+$2,675** |
>
> Caveats:
> - Wash sales are not computed — consult your 1099-B.
> - This excludes options and crypto (separate endpoints).
> - Open positions not sold this year aren't counted (unrealized).
