package robinhood

// Money is a decimal-string representation of monetary values.
// We preserve Robinhood's string form; conversion to a numeric type is a
// caller concern (and should use a decimal library, not float64).
type Money string
