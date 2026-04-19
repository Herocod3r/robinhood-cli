package robinhood

import (
	"errors"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	e := &APIError{Code: CodeNotFound, Message: "no such symbol"}
	if e.Error() != "not_found: no such symbol" {
		t.Errorf("got %q", e.Error())
	}
}

func TestAPIError_ExitCode(t *testing.T) {
	cases := []struct {
		code Code
		want int
	}{
		{CodeUnauthenticated, 2},
		{CodeSessionExpired, 2},
		{CodeSheriffRequired, 3},
		{CodeRateLimited, 4},
		{CodeValidation, 5},
		{CodeRobinhoodUnavailable, 1},
		{CodeNotFound, 1},
	}
	for _, c := range cases {
		e := &APIError{Code: c.code}
		if e.ExitCode() != c.want {
			t.Errorf("%s: ExitCode() = %d, want %d", c.code, e.ExitCode(), c.want)
		}
	}
}

func TestAPIError_Is(t *testing.T) {
	e := &APIError{Code: CodeNotFound}
	if !errors.Is(e, &APIError{Code: CodeNotFound}) {
		t.Error("errors.Is should match on Code")
	}
	if errors.Is(e, &APIError{Code: CodeRateLimited}) {
		t.Error("errors.Is should not match different Code")
	}
}
