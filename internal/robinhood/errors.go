// Package robinhood is the internal Robinhood HTTP client and typed endpoints.
package robinhood

import "fmt"

// Code is the stable error taxonomy exposed in the JSON envelope and via exit codes.
type Code string

const (
	CodeUnauthenticated      Code = "unauthenticated"
	CodeSessionExpired       Code = "session_expired"
	CodeSheriffRequired      Code = "sheriff_required"
	CodeMFARequired          Code = "mfa_required"
	CodeRateLimited          Code = "rate_limited"
	CodeRobinhoodUnavailable Code = "robinhood_unavailable"
	CodeNotFound             Code = "not_found"
	CodeValidation           Code = "validation"
)

// APIError is the structured error returned from all Robinhood client operations.
type APIError struct {
	Code      Code
	Message   string
	Hint      string
	Retryable bool
	// HTTPStatus is the underlying HTTP status code, 0 if not applicable.
	HTTPStatus int
	// WorkflowID is set when Code == CodeSheriffRequired. Used by the login
	// flow to correlate Sheriff start/respond with the triggering password
	// grant attempt.
	WorkflowID string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Message == "" {
		return string(e.Code)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Is supports errors.Is matching on Code only.
func (e *APIError) Is(target error) bool {
	t, ok := target.(*APIError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// ExitCode maps the error code to the CLI exit code per spec section 6.5.
// MFARequired is grouped with Sheriff — both signal "needs interactive step".
func (e *APIError) ExitCode() int {
	switch e.Code {
	case CodeUnauthenticated, CodeSessionExpired:
		return 2
	case CodeSheriffRequired, CodeMFARequired:
		return 3
	case CodeRateLimited:
		return 4
	case CodeValidation:
		return 5
	default:
		return 1
	}
}
