package main

import (
	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

// newAuthedClient returns a Robinhood client with a Session loaded from env vars.
// Returns the raw *robinhood.APIError on failure so main() can map to exit codes.
func newAuthedClient() (*robinhood.Client, error) {
	sess, err := robinhood.LoadSessionFromEnv()
	if err != nil {
		return nil, err
	}
	c := robinhood.NewClient()
	c.SetSession(sess)
	return c, nil
}
