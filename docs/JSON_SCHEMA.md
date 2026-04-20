# JSON Schema

Every `rh` command that produces data emits a stable JSON envelope. This
document defines a [JSON Schema (Draft 2020-12)](https://json-schema.org)
for each command's `data` payload, plus one example.

## Envelope

Every command emits the envelope shape defined in
`internal/output/envelope.go`:

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "envelope",
  "type": "object",
  "required": ["schema", "command", "generated_at", "data", "error"],
  "properties": {
    "schema": {"const": "robinhood-cli/v1"},
    "command": {"type": "string"},
    "generated_at": {"type": "string", "format": "date-time"},
    "data": {},
    "meta": {"type": "object"},
    "error": {
      "type": ["object", "null"],
      "properties": {
        "code": {"type": "string"},
        "message": {"type": "string"},
        "hint": {"type": "string"},
        "retryable": {"type": "boolean"}
      }
    }
  }
}
```

`error` is `null` for successful commands. `meta` is an optional
object used for counts / profile metadata.

The schemas below describe only the **`data`** payload for each command.
Combine with the envelope schema above when validating whole command
output.

Common primitives used throughout:

- Money: `{"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"}`
- Date-time: `{"type": "string", "format": "date-time"}`
- Date (YYYY-MM-DD): `{"type": "string", "format": "date"}`

---

## Meta commands

The following commands do not produce a typed `data` schema:

- `rh login` — writes progress to stderr only. No JSON `data`. Exit codes
  carry the result.
- `rh logout` — no JSON `data`.
- `rh skill install` — prints progress lines to stderr; no JSON `data`.
- `rh version` — `data` is a `{version, commit, schema}` object covered
  by the envelope's `schema` constant; not further constrained here.
- `rh commands` — `data` is the discovery payload (`CommandMeta[]`)
  defined by `cmd/rh/commands.go`; self-describing.
- `rh schema` — `data` is a copy of the envelope schema itself.

---

## Data commands

### `rh portfolio`

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "portfolio",
  "type": "object",
  "required": ["equity", "market_value", "cash", "buying_power"],
  "properties": {
    "equity": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
    "extended_hours_equity": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
    "market_value": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
    "cash": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
    "buying_power": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
    "last_core_equity": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"}
  }
}
```

Example:

```jsonc
{
  "equity": "98234.50",
  "extended_hours_equity": "98250.10",
  "market_value": "72450.10",
  "cash": "25784.40",
  "buying_power": "25784.40",
  "last_core_equity": "97500.00"
}
```

---

### `rh positions`

`data` is an array. Fields derived from `/positions/` via
`internal/robinhood/endpoints/positions.go:Position`. `last_price`,
`market_value`, `unrealized_pl`, `unrealized_pl_percent` are
back-filled from a batched quote fetch; they may be absent if the
quote lookup fails.

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "positions",
  "type": "array",
  "items": {
    "type": "object",
    "required": ["symbol", "quantity", "average_buy_price", "instrument_id"],
    "properties": {
      "symbol": {"type": "string"},
      "name": {"type": "string"},
      "quantity": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "average_buy_price": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "last_price": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "market_value": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "cost_basis": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "unrealized_pl": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "unrealized_pl_percent": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "instrument_id": {"type": "string"},
      "instrument_url": {"type": "string", "format": "uri"}
    }
  }
}
```

Example:

```jsonc
[
  {
    "symbol": "AAPL",
    "name": "Apple Inc.",
    "quantity": "50.0000",
    "average_buy_price": "150.1234",
    "last_price": "182.50",
    "market_value": "9125.00",
    "cost_basis": "7506.17",
    "unrealized_pl": "1618.83",
    "unrealized_pl_percent": "0.2156",
    "instrument_id": "450dfc6d-5510-4d40-abfb-f633b7d9be3e",
    "instrument_url": "https://api.robinhood.com/instruments/450dfc6d-5510-4d40-abfb-f633b7d9be3e/"
  }
]
```

---

### `rh position`

`rh position <symbol>` returns the same row shape as `rh positions`,
but filtered to a single symbol. `data` is the object (not an array).

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "position",
  "type": "object",
  "required": ["symbol", "quantity", "average_buy_price", "instrument_id"],
  "properties": {
    "symbol": {"type": "string"},
    "name": {"type": "string"},
    "quantity": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
    "average_buy_price": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
    "last_price": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
    "market_value": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
    "cost_basis": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
    "unrealized_pl": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
    "unrealized_pl_percent": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
    "instrument_id": {"type": "string"},
    "instrument_url": {"type": "string", "format": "uri"}
  }
}
```

Example:

```jsonc
{
  "symbol": "AAPL",
  "name": "Apple Inc.",
  "quantity": "50.0000",
  "average_buy_price": "150.1234",
  "last_price": "182.50",
  "market_value": "9125.00",
  "cost_basis": "7506.17",
  "unrealized_pl": "1618.83",
  "unrealized_pl_percent": "0.2156",
  "instrument_id": "450dfc6d-5510-4d40-abfb-f633b7d9be3e"
}
```

---

### `rh account`

Merged view from `/accounts/` and `/accounts/unified`. Definition:
`internal/robinhood/endpoints/account.go:AccountSummary`.

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "account",
  "type": "object",
  "required": ["account_number", "buying_power", "cash", "sweep_enabled", "pattern_day_trader", "day_trade_count", "instant_used", "instant_available"],
  "properties": {
    "account_number": {"type": "string"},
    "buying_power": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
    "cash": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
    "sweep_enabled": {"type": "boolean"},
    "margin_balance": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
    "pattern_day_trader": {"type": "boolean"},
    "day_trade_count": {"type": "integer", "minimum": 0},
    "instant_used": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
    "instant_available": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"}
  }
}
```

Example:

```jsonc
{
  "account_number": "5RY12345",
  "buying_power": "25784.40",
  "cash": "25784.40",
  "sweep_enabled": true,
  "margin_balance": "0.00",
  "pattern_day_trader": false,
  "day_trade_count": 0,
  "instant_used": "0.00",
  "instant_available": "1000.00"
}
```

---

### `rh quote`

`data` is an array of quote rows (even for a single symbol). Fields from
`internal/robinhood/endpoints/quotes.go:Quote`.

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "quote",
  "type": "array",
  "items": {
    "type": "object",
    "required": ["symbol", "last_price", "bid_price", "ask_price", "previous_close", "updated_at"],
    "properties": {
      "symbol": {"type": "string"},
      "name": {"type": "string"},
      "last_price": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "bid_price": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "ask_price": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "bid_size": {"type": "integer", "minimum": 0},
      "ask_size": {"type": "integer", "minimum": 0},
      "previous_close": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "extended_hours_price": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "volume": {"type": "integer", "minimum": 0},
      "updated_at": {"type": "string", "format": "date-time"}
    }
  }
}
```

Example:

```jsonc
[
  {
    "symbol": "AAPL",
    "last_price": "182.50",
    "bid_price": "182.49",
    "ask_price": "182.51",
    "bid_size": 300,
    "ask_size": 200,
    "previous_close": "180.00",
    "extended_hours_price": "182.80",
    "volume": 52300000,
    "updated_at": "2026-04-20T20:00:00Z"
  }
]
```

Up to 50 symbols per invocation (`/quotes/` limit); validated in
`Quotes.Batch`.

---

### `rh fundamentals`

`data` is an array. Fields: `internal/robinhood/endpoints/fundamentals.go:Fundamentals`.

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "fundamentals",
  "type": "array",
  "items": {
    "type": "object",
    "required": ["symbol", "open", "high", "low", "volume"],
    "properties": {
      "symbol": {"type": "string"},
      "open": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "high": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "low": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "volume": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "average_volume": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "average_volume_2_weeks": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "high_52_weeks": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "low_52_weeks": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "dividend_yield": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "market_cap": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "pe_ratio": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "description": {"type": "string"},
      "instrument": {"type": "string", "format": "uri"}
    }
  }
}
```

Example:

```jsonc
[
  {
    "symbol": "AAPL",
    "open": "180.12",
    "high": "183.50",
    "low": "179.88",
    "volume": "52300000",
    "average_volume": "58000000",
    "average_volume_2_weeks": "55000000",
    "high_52_weeks": "199.62",
    "low_52_weeks": "164.08",
    "dividend_yield": "0.0053",
    "market_cap": "2812000000000",
    "pe_ratio": "29.5",
    "description": "Apple Inc. designs, manufactures, and markets smartphones..."
  }
]
```

---

### `rh historicals`

`data` is one `Historicals` object for a single symbol. Fields:
`internal/robinhood/endpoints/historicals.go`.

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "historicals",
  "type": "object",
  "required": ["symbol", "interval", "span", "bars"],
  "properties": {
    "symbol": {"type": "string"},
    "interval": {"type": "string", "enum": ["5minute", "10minute", "hour", "day", "week"]},
    "span": {"type": "string", "enum": ["day", "week", "month", "3month", "year", "5year"]},
    "bars": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["begins_at", "open_price", "close_price", "high_price", "low_price", "volume", "session"],
        "properties": {
          "begins_at": {"type": "string", "format": "date-time"},
          "open_price": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
          "close_price": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
          "high_price": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
          "low_price": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
          "volume": {"type": "integer", "minimum": 0},
          "session": {"type": "string"}
        }
      }
    }
  }
}
```

Example:

```jsonc
{
  "symbol": "AAPL",
  "interval": "day",
  "span": "month",
  "bars": [
    {
      "begins_at": "2026-04-01T13:30:00Z",
      "open_price": "180.10",
      "close_price": "181.44",
      "high_price": "182.00",
      "low_price": "179.90",
      "volume": 52300000,
      "session": "reg"
    }
  ]
}
```

---

### `rh news`

`data` is an array of news items. Fields:
`internal/robinhood/endpoints/news.go:NewsItem`.

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "news",
  "type": "array",
  "items": {
    "type": "object",
    "required": ["title", "source", "url", "published_at"],
    "properties": {
      "title": {"type": "string"},
      "author": {"type": "string"},
      "source": {"type": "string"},
      "url": {"type": "string", "format": "uri"},
      "published_at": {"type": "string", "format": "date-time"},
      "summary": {"type": "string"},
      "preview_text": {"type": "string"}
    }
  }
}
```

Example:

```jsonc
[
  {
    "title": "Apple unveils new M5 chip",
    "author": "Jane Reporter",
    "source": "Reuters",
    "url": "https://reuters.example/article/aapl-m5",
    "published_at": "2026-04-19T14:00:00Z",
    "summary": "Apple today announced...",
    "preview_text": "Apple today announced..."
  }
]
```

---

### `rh earnings`

`data` is an array of earnings events. Fields:
`internal/robinhood/endpoints/earnings.go:EarningsEvent`.

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "earnings",
  "type": "array",
  "items": {
    "type": "object",
    "required": ["symbol", "year", "quarter", "eps"],
    "properties": {
      "symbol": {"type": "string"},
      "year": {"type": "integer"},
      "quarter": {"type": "integer", "minimum": 1, "maximum": 4},
      "report_at": {"type": "string"},
      "eps": {
        "type": "object",
        "properties": {
          "estimate": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
          "actual": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"}
        }
      },
      "call": {
        "type": ["object", "null"],
        "properties": {
          "datetime": {"type": "string", "format": "date-time"},
          "broadcast_url": {"type": "string", "format": "uri"}
        }
      }
    }
  }
}
```

Example:

```jsonc
[
  {
    "symbol": "AAPL",
    "year": 2026,
    "quarter": 1,
    "report_at": "2026-01-25",
    "eps": {"estimate": "2.10", "actual": "2.18"},
    "call": {"datetime": "2026-01-25T21:30:00Z", "broadcast_url": "https://…"}
  }
]
```

---

### `rh ratings`

`data` is a single `Rating` object. Fields:
`internal/robinhood/endpoints/ratings.go:Rating`.

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "ratings",
  "type": "object",
  "required": ["symbol", "summary", "ratings"],
  "properties": {
    "symbol": {"type": "string"},
    "summary": {
      "type": "object",
      "required": ["num_buy_ratings", "num_hold_ratings", "num_sell_ratings"],
      "properties": {
        "num_buy_ratings": {"type": "integer", "minimum": 0},
        "num_hold_ratings": {"type": "integer", "minimum": 0},
        "num_sell_ratings": {"type": "integer", "minimum": 0}
      }
    },
    "ratings": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["type", "text", "published_at"],
        "properties": {
          "type": {"type": "string"},
          "text": {"type": "string"},
          "published_at": {"type": "string", "format": "date-time"}
        }
      }
    }
  }
}
```

Example:

```jsonc
{
  "symbol": "AAPL",
  "summary": {"num_buy_ratings": 28, "num_hold_ratings": 5, "num_sell_ratings": 1},
  "ratings": [
    {"type": "buy", "text": "Upgraded on strong iPhone demand.", "published_at": "2026-04-15T12:00:00Z"}
  ]
}
```

---

### `rh dividends`

`data` is an array. Fields:
`internal/robinhood/endpoints/dividends.go:Dividend`.

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "dividends",
  "type": "array",
  "items": {
    "type": "object",
    "required": ["amount", "rate", "position", "paid_at", "record_date", "payable_date", "state"],
    "properties": {
      "symbol": {"type": "string"},
      "amount": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "rate": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "position": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "withholding": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "paid_at": {"type": "string"},
      "record_date": {"type": "string"},
      "payable_date": {"type": "string"},
      "instrument_url": {"type": "string", "format": "uri"},
      "instrument_id": {"type": "string"},
      "state": {"type": "string"}
    }
  }
}
```

Example:

```jsonc
[
  {
    "symbol": "AAPL",
    "amount": "12.00",
    "rate": "0.24",
    "position": "50.0000",
    "withholding": "0.00",
    "paid_at": "2026-02-14T00:00:00Z",
    "record_date": "2026-02-08",
    "payable_date": "2026-02-14",
    "instrument_id": "450dfc6d-5510-4d40-abfb-f633b7d9be3e",
    "state": "paid"
  }
]
```

---

### `rh options-positions`

`data` is an array of aggregated options positions. Fields:
`internal/robinhood/endpoints/options_positions.go`. Leg fields
`type`, `strike_price`, and `expiration` are fetched via per-leg
instrument lookups and may be empty strings if the lookup fails; the
outer leg entry always has `option_id` and `position_type`.

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "options-positions",
  "type": "array",
  "items": {
    "type": "object",
    "required": ["symbol", "strategy", "quantity", "average_price", "legs"],
    "properties": {
      "symbol": {"type": "string"},
      "strategy": {"type": "string"},
      "quantity": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "average_price": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "legs": {
        "type": "array",
        "items": {
          "type": "object",
          "required": ["option_id", "position_type"],
          "properties": {
            "option_id": {"type": "string"},
            "type": {"type": "string", "enum": ["call", "put", ""]},
            "strike_price": {"type": "string"},
            "expiration": {"type": "string"},
            "position_type": {"type": "string", "enum": ["long", "short"]}
          }
        }
      }
    }
  }
}
```

Example:

```jsonc
[
  {
    "symbol": "AAPL",
    "strategy": "long_call",
    "quantity": "1.0000",
    "average_price": "3.25",
    "legs": [
      {
        "option_id": "a1b2c3d4-0000-4000-8000-000000000001",
        "type": "call",
        "strike_price": "190.00",
        "expiration": "2026-06-19",
        "position_type": "long"
      }
    ]
  }
]
```

---

### `rh orders`

`data` is an array of orders. Fields:
`internal/robinhood/endpoints/orders.go:Order`.

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "orders",
  "type": "array",
  "items": {
    "type": "object",
    "required": ["id", "symbol", "side", "type", "state", "quantity", "fees", "time_in_force", "extended_hours", "created_at"],
    "properties": {
      "id": {"type": "string"},
      "symbol": {"type": "string"},
      "side": {"type": "string", "enum": ["buy", "sell"]},
      "type": {"type": "string"},
      "state": {"type": "string"},
      "quantity": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "price": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "average_fill_price": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "fees": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"},
      "time_in_force": {"type": "string"},
      "extended_hours": {"type": "boolean"},
      "created_at": {"type": "string", "format": "date-time"},
      "filled_at": {"type": "string"},
      "cancelled_at": {"type": "string"},
      "instrument": {"type": "string", "format": "uri"}
    }
  }
}
```

Example:

```jsonc
[
  {
    "id": "f0e9d8c7-0000-4000-8000-000000000001",
    "symbol": "AAPL",
    "side": "buy",
    "type": "market",
    "state": "filled",
    "quantity": "10.0000",
    "price": "180.00",
    "average_fill_price": "180.12",
    "fees": "0.00",
    "time_in_force": "gfd",
    "extended_hours": false,
    "created_at": "2026-04-01T13:30:00Z",
    "filled_at": "2026-04-01T13:30:02Z"
  }
]
```

---

### `rh watchlist`

`data` is an array. Fields:
`internal/robinhood/endpoints/watchlist.go:WatchlistItem`. `last_price`
may be omitted if the quote fetch fails.

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "watchlist",
  "type": "array",
  "items": {
    "type": "object",
    "required": ["symbol", "instrument_id", "added_at"],
    "properties": {
      "symbol": {"type": "string"},
      "instrument_id": {"type": "string"},
      "added_at": {"type": "string", "format": "date-time"},
      "last_price": {"type": "string", "pattern": "^-?\\d+(\\.\\d+)?$"}
    }
  }
}
```

Example:

```jsonc
[
  {
    "symbol": "NVDA",
    "instrument_id": "11111111-2222-3333-4444-555555555555",
    "added_at": "2025-09-01T14:22:00Z",
    "last_price": "875.00"
  }
]
```

---

### `rh search`

`data` is an array of search hits. Fields:
`internal/robinhood/endpoints/search.go:SearchResult`.

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "search",
  "type": "array",
  "items": {
    "type": "object",
    "required": ["symbol", "name", "instrument_id", "tradeable", "type"],
    "properties": {
      "symbol": {"type": "string"},
      "name": {"type": "string"},
      "instrument_id": {"type": "string"},
      "tradeable": {"type": "boolean"},
      "type": {"type": "string"}
    }
  }
}
```

Example:

```jsonc
[
  {
    "symbol": "AAPL",
    "name": "Apple Inc.",
    "instrument_id": "450dfc6d-5510-4d40-abfb-f633b7d9be3e",
    "tradeable": true,
    "type": "stock"
  }
]
```

---

### `rh market-hours`

`data` is an array of per-market hours for the requested date. Fields:
`internal/robinhood/endpoints/market_hours.go:MarketHours`.

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "market-hours",
  "type": "array",
  "items": {
    "type": "object",
    "required": ["mic", "name", "date", "is_open"],
    "properties": {
      "mic": {"type": "string"},
      "name": {"type": "string"},
      "date": {"type": "string", "format": "date"},
      "is_open": {"type": "boolean"},
      "opens_at": {"type": "string"},
      "closes_at": {"type": "string"},
      "extended_opens_at": {"type": "string"},
      "extended_closes_at": {"type": "string"}
    }
  }
}
```

Example:

```jsonc
[
  {
    "mic": "XNAS",
    "name": "NASDAQ",
    "date": "2026-04-20",
    "is_open": true,
    "opens_at": "2026-04-20T13:30:00Z",
    "closes_at": "2026-04-20T20:00:00Z",
    "extended_opens_at": "2026-04-20T08:00:00Z",
    "extended_closes_at": "2026-04-21T00:00:00Z"
  }
]
```

---

### `rh documents`

`data` is an array. Fields:
`internal/robinhood/endpoints/documents.go:Document`. Only listing is
covered by this schema — `rh documents --download` returns a
`DownloadResult` payload which is maintainer-internal.

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "documents",
  "type": "array",
  "items": {
    "type": "object",
    "required": ["id", "type", "date", "name", "download_url"],
    "properties": {
      "id": {"type": "string"},
      "type": {"type": "string"},
      "date": {"type": "string", "format": "date"},
      "name": {"type": "string"},
      "download_url": {"type": "string", "format": "uri"}
    }
  }
}
```

Example:

```jsonc
[
  {
    "id": "d0cd0cd0-1111-4000-8000-000000000042",
    "type": "statement",
    "date": "2026-03-31",
    "name": "account-statement-march.pdf",
    "download_url": "https://rh-documents.example/ticket/abc"
  }
]
```
