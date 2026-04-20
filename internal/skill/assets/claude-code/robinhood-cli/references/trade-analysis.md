# Trade analysis — realized P&L (FIFO)

## Input

`rh orders --state filled --since <date> --json`

`.data[*]` fields you need: `symbol`, `side`, `quantity`, `average_fill_price`, `fees`, `filled_at`.

## Algorithm

For each symbol, group orders by symbol and sort by `filled_at` ascending.
Maintain a FIFO queue of open lots (`remaining_qty`, `cost_basis_per_share`).

For each buy: append a new lot.
For each sell:
  while `sell_qty > 0`:
    take the head lot, take `min(sell_qty, lot.remaining_qty)`.
    proceeds_per_share = average_fill_price
    realized_pl += (proceeds_per_share - lot.cost_basis_per_share) * taken_qty - fees_alloc
    reduce the lot; drop it if empty.

`fees_alloc` = fees * taken_qty / total_sell_qty (prorate across the sell's portions).

## Decimal discipline

All arithmetic must use a decimal type — Money values are strings for a reason.
In Python: `decimal.Decimal`. In JS: `decimal.js` or `bignumber.js`. Do NOT
parse as `Number`/`float64`.

## Caveats (surface to user)

- **Wash sales are not computed.** Robinhood's 1099-B is the authoritative
  record for tax purposes. Mention this.
- **Options assignments and crypto transactions are not in `/orders/`.** They
  appear in separate endpoints and are outside this analysis.
- **Splits and dividends** that affect cost basis are not adjusted here.
  Consult `rh dividends` and the user's account settings for split history.
