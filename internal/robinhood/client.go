package robinhood

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Host enumerates the Robinhood API hosts we talk to.
type Host int

const (
	APIHost     Host = iota // api.robinhood.com
	NummusHost              // nummus.robinhood.com  (crypto)
	PhoenixHost             // phoenix.robinhood.com (unified account)
)

// Default production hosts.
const (
	defaultAPIBase     = "https://api.robinhood.com"
	defaultNummusBase  = "https://nummus.robinhood.com"
	defaultPhoenixBase = "https://phoenix.robinhood.com"
)

// Client is the HTTP client for Robinhood's unofficial API.
type Client struct {
	apiBase     string
	nummusBase  string
	phoenixBase string
	http        *http.Client
	oauth       *oauth

	mu      sync.Mutex
	session *Session
}

// NewClient returns a client pointed at production hosts.
func NewClient() *Client {
	return NewClientWithHosts(defaultAPIBase, defaultNummusBase, defaultPhoenixBase, &http.Client{Timeout: 30 * time.Second})
}

// NewClientWithHosts lets tests override hosts and the underlying http.Client.
func NewClientWithHosts(apiBase, nummusBase, phoenixBase string, h *http.Client) *Client {
	return &Client{
		apiBase:     apiBase,
		nummusBase:  nummusBase,
		phoenixBase: phoenixBase,
		http:        h,
		oauth:       &oauth{baseURL: apiBase, httpClient: h},
	}
}

// SetSession installs the active session.
func (c *Client) SetSession(s *Session) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.session = s
}

// Session returns the active session (may be nil).
func (c *Client) Session() *Session {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.session
}

// baseFor returns the base URL for the given host.
func (c *Client) baseFor(h Host) string {
	switch h {
	case NummusHost:
		return c.nummusBase
	case PhoenixHost:
		return c.phoenixBase
	default:
		return c.apiBase
	}
}

// ensureFresh refreshes pre-emptively if the session is expired.
func (c *Client) ensureFresh() error {
	c.mu.Lock()
	s := c.session
	c.mu.Unlock()
	if s == nil {
		return &APIError{Code: CodeUnauthenticated, Message: "no session", Hint: "run: rh login"}
	}
	// Only proactively refresh when we KNOW we need to:
	// - no access token (but refresh available), or
	// - access token present with a known-expired ExpiresAt.
	// Unknown expiry (env-loaded sessions) trusts the access token and
	// lets 401 drive refresh. Otherwise every CLI invocation would burn
	// the refresh token once before trying the live access token.
	if s.NeedsImmediateRefresh() {
		return c.oauth.Refresh(s)
	}
	return nil
}

// getJSON does a GET, auto-refreshes once on 401, and decodes into out.
// It is the private workhorse used by all endpoint helpers.
func (c *Client) getJSON(host Host, path string, out any) error {
	if err := c.ensureFresh(); err != nil {
		return err
	}
	resp, err := c.do(host, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		// One retry after refresh.
		if rerr := c.oauth.Refresh(c.Session()); rerr != nil {
			return rerr
		}
		resp2, err := c.do(host, http.MethodGet, path, nil)
		if err != nil {
			return err
		}
		defer resp2.Body.Close()
		if resp2.StatusCode == http.StatusUnauthorized {
			return &APIError{Code: CodeUnauthenticated, Message: "401 after refresh", Hint: "run: rh login"}
		}
		return decodeOrMap(resp2, out)
	}

	return decodeOrMap(resp, out)
}

// do builds and sends a single request with the Authorization header.
func (c *Client) do(host Host, method, path string, body io.Reader) (*http.Response, error) {
	s := c.Session()
	req, err := http.NewRequest(method, c.baseFor(host)+path, body)
	if err != nil {
		return nil, &APIError{Code: CodeValidation, Message: err.Error()}
	}
	if s != nil && s.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.AccessToken)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "robinhood-cli (+https://github.com/herocod3r/robinhood-cli)")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, &APIError{Code: CodeRobinhoodUnavailable, Message: err.Error()}
	}
	return resp, nil
}

// decodeOrMap converts non-2xx responses into typed APIErrors, and decodes 2xx into out.
func decodeOrMap(resp *http.Response, out any) error {
	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		if out == nil {
			return nil
		}
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return &APIError{Code: CodeRobinhoodUnavailable, Message: fmt.Sprintf("decode: %v", err)}
		}
		return nil
	case resp.StatusCode == http.StatusNotFound:
		return &APIError{Code: CodeNotFound, Message: "not found", HTTPStatus: resp.StatusCode}
	case resp.StatusCode == http.StatusTooManyRequests:
		retry := resp.Header.Get("Retry-After")
		msg := "rate limited"
		hint := "retry in 30s"
		if retry != "" {
			hint = "retry in " + retry + "s"
		}
		return &APIError{Code: CodeRateLimited, Message: msg, Hint: hint, Retryable: true, HTTPStatus: resp.StatusCode}
	case resp.StatusCode >= 500:
		return &APIError{Code: CodeRobinhoodUnavailable, Message: fmt.Sprintf("HTTP %d", resp.StatusCode), HTTPStatus: resp.StatusCode}
	default:
		buf, _ := io.ReadAll(resp.Body)
		return &APIError{Code: CodeValidation, Message: strings.TrimSpace(string(buf)), HTTPStatus: resp.StatusCode}
	}
}
