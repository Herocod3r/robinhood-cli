# Sector concentration

## Input

`rh positions --nonzero --json`

`rh` does not return sector on positions in v1. Two options:

1. Ask the user to provide a sector map (small portfolios).
2. Enrich via `rh fundamentals <ticker>...` — `sector` and `industry` fields
   on fundamentals responses. Batch into groups of 50 symbols.

## Algorithm

```
sector_value = 0
for position in positions:
    sector = fundamentals_by_symbol[position.symbol].sector
    sector_value[sector] += decimal(position.market_value)
total = sum(decimal(p.market_value) for p in positions)
for sector, value in sector_value.items():
    percent = value / total
```

## What to surface

- Top 3 sectors with percentage.
- Any sector >40% — "heavy concentration in <sector>".
- Any single position >20% — "high single-name risk in <sym>".

## Caveats

- `fundamentals.sector` is a free-text field; occasionally missing. When
  missing, bucket under "unclassified" and note it.
- Sector is as of the last fundamentals update from Robinhood; typically
  refreshed daily.
