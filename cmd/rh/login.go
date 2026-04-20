package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/herocod3r/robinhood-cli/internal/config"
	"github.com/herocod3r/robinhood-cli/internal/inputprompt"
	"github.com/herocod3r/robinhood-cli/internal/robinhood"
	"github.com/spf13/cobra"
)

func init() {
	c := &cobra.Command{
		Use:   "login",
		Short: "Log into Robinhood (interactive)",
		RunE:  runLogin,
	}
	c.Flags().String("totp-secret", "", "base32 TOTP shared secret (optional; ROBINHOOD_TOTP_SECRET env var also honored)")
	c.Flags().String("username", "", "email/username (prompted if empty)")
	register(c)
}

// LoginOpts bundles everything RunLogin needs. CLI builds this from flags + env;
// tests build it directly.
//
// The Password field is intentionally NOT printed by any formatter — see the
// String and GoString methods below (Fix H). Tests guard this with
// TestLoginOpts_DoesNotLeakPassword.
type LoginOpts struct {
	Host        string       // defaults to https://api.robinhood.com
	HTTP        *http.Client // defaults to http.DefaultClient
	Username    string
	Password    string
	DeviceToken string
	TOTPSecret  string
	CodeInput   func(prompt string) (string, error) // SMS/email; push flows never call this
	Out         io.Writer
	Profile     string
	PollEvery   time.Duration // Sheriff poll cadence; default 2s
}

// String and GoString redact every field (not just Password) so no formatter
// accidentally leaks credentials in logs, panic traces, or error envelopes.
// See Fix H.
func (o LoginOpts) String() string   { return "LoginOpts{<redacted>}" }
func (o LoginOpts) GoString() string { return "LoginOpts{<redacted>}" }

// RunLogin drives the entire Sheriff/MFA state machine. Returns the persisted session.
func RunLogin(ctx context.Context, o LoginOpts) (*robinhood.Session, error) {
	if o.Host == "" {
		o.Host = "https://api.robinhood.com"
	}
	if o.HTTP == nil {
		o.HTTP = &http.Client{Timeout: 30 * time.Second}
	}
	if o.Profile == "" {
		o.Profile = "default"
	}
	if err := config.ValidProfile(o.Profile); err != nil {
		return nil, &robinhood.APIError{Code: robinhood.CodeValidation, Message: err.Error()}
	}
	if o.DeviceToken == "" {
		o.DeviceToken = newDeviceToken()
	}

	oauthClient := robinhood.NewOAuth(o.Host, o.HTTP)

	// Build TOTP code if a shared secret was provided.
	var mfaCode string
	if o.TOTPSecret != "" {
		code, err := robinhood.TOTPCode(o.TOTPSecret)
		if err != nil {
			return nil, &robinhood.APIError{Code: robinhood.CodeValidation, Message: "invalid TOTP secret: " + err.Error()}
		}
		mfaCode = code
	}

	sess, err := oauthClient.PasswordGrant(ctx, o.Username, o.Password, o.DeviceToken, mfaCode)
	if err == nil {
		return persistLogin(sess, o)
	}
	var apiErr *robinhood.APIError
	if !errors.As(err, &apiErr) {
		return nil, err
	}

	switch apiErr.Code {
	case robinhood.CodeMFARequired:
		// Fix I: two-attempt loop with nested Sheriff handling on retry.
		for attempt := 0; attempt < 2; attempt++ {
			code, inErr := o.CodeInput(apiErr.Hint)
			if inErr != nil || code == "" {
				return nil, &robinhood.APIError{Code: robinhood.CodeValidation, Message: "no MFA code entered"}
			}
			sess, pgErr := oauthClient.PasswordGrant(ctx, o.Username, o.Password, o.DeviceToken, code)
			if pgErr == nil {
				return persistLogin(sess, o)
			}
			var ae *robinhood.APIError
			if errors.As(pgErr, &ae) && ae.Code == robinhood.CodeMFARequired {
				apiErr = ae
				continue // prompt again
			}
			if errors.As(pgErr, &ae) && ae.Code == robinhood.CodeSheriffRequired {
				return runSheriff(ctx, o, oauthClient, ae, code)
			}
			return nil, pgErr
		}
		return nil, &robinhood.APIError{Code: robinhood.CodeMFARequired, Message: "MFA code rejected twice"}

	case robinhood.CodeSheriffRequired:
		return runSheriff(ctx, o, oauthClient, apiErr, mfaCode)

	default:
		return nil, err
	}
}

// runSheriff handles a single Sheriff hop: device verification via SMS,
// email, or push prompt, followed by the post-challenge token re-exchange.
func runSheriff(ctx context.Context, o LoginOpts, oauthClient *robinhood.OAuth, apiErr *robinhood.APIError, mfaCode string) (*robinhood.Session, error) {
	sheriff := &robinhood.Sheriff{
		BaseURL:   o.Host,
		HTTP:      o.HTTP,
		PollEvery: o.PollEvery,
	}
	fmt.Fprintln(o.Out, "Robinhood requires device verification.")
	step, err := sheriff.Start(ctx, apiErr.WorkflowID, o.DeviceToken)
	if err != nil {
		return nil, err
	}
	switch step.Kind {
	case robinhood.SheriffSMS, robinhood.SheriffEmail:
		fmt.Fprintf(o.Out, "Robinhood sent a code via %s.\n", step.Detail)
		code, err := o.CodeInput("Enter the code: ")
		if err != nil || code == "" {
			return nil, &robinhood.APIError{Code: robinhood.CodeValidation, Message: "no code entered"}
		}
		if err := sheriff.RespondCode(ctx, step, code); err != nil {
			return nil, err
		}
	case robinhood.SheriffPush:
		fmt.Fprintf(o.Out, "Approve the prompt in your Robinhood app...\n")
		if err := sheriff.WaitPush(ctx, step); err != nil {
			return nil, err
		}
	default:
		return nil, &robinhood.APIError{Code: robinhood.CodeSheriffRequired, Message: "unhandled sheriff step"}
	}
	sess, err := oauthClient.PasswordGrantWithWorkflow(ctx, o.Username, o.Password, o.DeviceToken, mfaCode, apiErr.WorkflowID)
	if err != nil {
		return nil, err
	}
	return persistLogin(sess, o)
}

// persistLogin writes the session to the keychain and records the profile in
// config.yaml if it wasn't already present.
func persistLogin(s *robinhood.Session, o LoginOpts) (*robinhood.Session, error) {
	if err := s.SaveToKeychain(o.Profile); err != nil {
		return nil, err
	}
	// Record profile in config if it wasn't there already.
	f, err := config.LoadFile()
	if err == nil {
		if _, ok := f.Profiles[o.Profile]; !ok {
			f.Profiles[o.Profile] = config.Profile{}
			_ = config.SaveFile(f)
		}
	}
	fmt.Fprintf(o.Out, "Logged in as %s (profile: %s)\n", s.Username, o.Profile)
	return s, nil
}

// newDeviceToken returns a random 32-hex-char device identifier. Robinhood
// accepts any stable string here; we generate our own rather than pulling in
// a UUID dep.
func newDeviceToken() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// Fall back to a timestamp-based token if the kernel RNG is unavailable.
		return fmt.Sprintf("rh-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

func runLogin(cmd *cobra.Command, args []string) error {
	// Fix K: validate profile before any use.
	if err := config.ValidProfile(gflags.Profile); err != nil {
		return &robinhood.APIError{Code: robinhood.CodeValidation, Message: err.Error()}
	}
	// Fix F: hard 5-minute deadline so Sheriff / push polling can't hang forever.
	ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
	defer cancel()

	p := inputprompt.New()
	username, _ := cmd.Flags().GetString("username")
	if username == "" {
		username = os.Getenv(config.EnvUsername)
	}
	if username == "" {
		var err error
		if username, err = p.Text("Username: "); err != nil {
			return &robinhood.APIError{Code: robinhood.CodeValidation, Message: err.Error()}
		}
	}
	password := os.Getenv(config.EnvPassword)
	if password == "" {
		var err error
		if password, err = p.Password("Password: "); err != nil {
			return &robinhood.APIError{Code: robinhood.CodeValidation, Message: err.Error()}
		}
	}
	totpSecret, _ := cmd.Flags().GetString("totp-secret")
	if totpSecret == "" {
		totpSecret = os.Getenv(config.EnvTOTPSecret)
	}
	deviceToken := os.Getenv(config.EnvDeviceToken)
	if deviceToken == "" {
		// Try to reuse the one we stashed on a previous login — otherwise generate.
		if existing, err := robinhood.LoadFromKeychain(gflags.Profile); err == nil && existing.DeviceToken != "" {
			deviceToken = existing.DeviceToken
		} else {
			deviceToken = newDeviceToken()
		}
	}

	// Fix P: honor ROBINHOOD_API_HOST for local dev / mitmproxy captures.
	host := os.Getenv("ROBINHOOD_API_HOST")

	_, err := RunLogin(ctx, LoginOpts{
		Host:        host,
		Username:    username,
		Password:    password,
		DeviceToken: deviceToken,
		TOTPSecret:  totpSecret,
		CodeInput:   p.Text,
		Out:         cmd.ErrOrStderr(),
		Profile:     gflags.Profile,
	})
	return err
}
