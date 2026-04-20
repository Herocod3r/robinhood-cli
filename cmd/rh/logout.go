package main

import (
	"fmt"
	"io"
	"os"

	"github.com/herocod3r/robinhood-cli/internal/config"
	"github.com/herocod3r/robinhood-cli/internal/robinhood"
	"github.com/spf13/cobra"
)

func init() {
	c := &cobra.Command{
		Use:   "logout",
		Short: "Clear stored session(s)",
		Annotations: map[string]string{
			"category":  "auth",
			"stability": "green",
		},
		RunE: runLogoutCmd,
	}
	c.Flags().Bool("all-profiles", false, "clear sessions for every profile")
	register(c)
}

// LogoutOpts configures RunLogout. Tests build it directly.
type LogoutOpts struct {
	Profile     string
	AllProfiles bool
	Out         io.Writer
}

// RunLogout clears the keychain entry for the given profile (or all profiles).
// Invalid profile names are rejected with a CodeValidation APIError (Fix K).
func RunLogout(o LogoutOpts) error {
	if o.Out == nil {
		o.Out = os.Stderr
	}
	if o.AllProfiles {
		f, err := config.LoadFile()
		if err != nil {
			return err
		}
		for name := range f.Profiles {
			// LoadFile already validated names, but clear the keychain under the
			// same guard so injected map keys can't escape.
			if err := config.ValidProfile(name); err != nil {
				return &robinhood.APIError{Code: robinhood.CodeValidation, Message: err.Error()}
			}
			if err := robinhood.ClearKeychain(name); err != nil {
				return err
			}
			fmt.Fprintf(o.Out, "Cleared profile %q\n", name)
		}
		return nil
	}
	profile := o.Profile
	if profile == "" {
		profile = "default"
	}
	if err := config.ValidProfile(profile); err != nil {
		return &robinhood.APIError{Code: robinhood.CodeValidation, Message: err.Error()}
	}
	if err := robinhood.ClearKeychain(profile); err != nil {
		return err
	}
	fmt.Fprintf(o.Out, "Cleared profile %q\n", profile)
	return nil
}

func runLogoutCmd(cmd *cobra.Command, args []string) error {
	// Fix K: validate profile before any use.
	if err := config.ValidProfile(gflags.Profile); err != nil {
		return &robinhood.APIError{Code: robinhood.CodeValidation, Message: err.Error()}
	}
	all, _ := cmd.Flags().GetBool("all-profiles")
	return RunLogout(LogoutOpts{
		Profile:     gflags.Profile,
		AllProfiles: all,
		Out:         cmd.ErrOrStderr(),
	})
}
