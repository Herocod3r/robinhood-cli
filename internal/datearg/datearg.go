// Package datearg parses absolute (YYYY-MM-DD) and relative
// (7d, 2w, 1mo, 3mo, 1y, ytd) date strings used by CLI --since flags.
package datearg

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// Parse returns the UTC time corresponding to s, anchored at `now`.
//
// Supported forms:
//   - "YYYY-MM-DD"       absolute UTC calendar date
//   - "ytd"              Jan 1 of the current year
//   - "Nd" / "Nw"        N days / N weeks ago
//   - "Nmo"              N × 30 days ago (approximate months)
//   - "Ny"               N × 365 days ago (approximate years)
//
// "1mo" MUST be parsed before "1m" would be, so the month suffix is
// tested before the single-letter units.
func Parse(s string, now time.Time) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, errors.New("empty date")
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t.UTC(), nil
	}
	if s == "ytd" {
		return time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC), nil
	}
	// Order matters: "mo" before "d"/"w"/"y" so that e.g. "3mo" is not
	// misread; specifically "mo" ends in "o" so it is unambiguous versus
	// the one-letter suffixes, but map iteration order is random — iterate
	// a sorted slice instead.
	type unit struct {
		suffix string
		d      time.Duration
	}
	units := []unit{
		{"mo", 30 * 24 * time.Hour},
		{"d", 24 * time.Hour},
		{"w", 7 * 24 * time.Hour},
		{"y", 365 * 24 * time.Hour},
	}
	for _, u := range units {
		if strings.HasSuffix(s, u.suffix) {
			n, err := strconv.Atoi(strings.TrimSuffix(s, u.suffix))
			if err != nil {
				return time.Time{}, err
			}
			if n < 0 {
				return time.Time{}, errors.New("negative duration")
			}
			return now.UTC().Add(-time.Duration(n) * u.d), nil
		}
	}
	return time.Time{}, errors.New("unrecognized date: " + s)
}
