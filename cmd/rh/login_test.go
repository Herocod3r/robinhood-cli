package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/herocod3r/robinhood-cli/internal/robinhood"
)

// TestLogin_SMS_Flow drives the full login E2E with a synthetic Robinhood
// that returns Sheriff → SMS → validated → tokens.
//
// Deviation from the plan text: the Sheriff implementation reads
// `context.sheriff_challenge` at the TOP LEVEL of the user_view payload
// (matching robin_stocks master), not the nested
// `type_context.context.sheriff_challenge` shape the plan snippet used.
// See Sheriff.fetchUserView in internal/robinhood/sheriff.go.
func TestLogin_SMS_Flow(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("ROBINHOOD_KEYCHAIN_BACKEND", "file")

	var firstCall atomic.Bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token/":
			if r.Header.Get("X-Robinhood-Challenge-Response-Id") == "" && !firstCall.Swap(true) {
				// First call: respond with Sheriff required.
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(w, `{"verification_workflow":{"id":"wf-1","workflow_status":"internal_pending"}}`)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"access_token":"a","refresh_token":"r","expires_in":3600}`)
		case "/pathfinder/user_machine/":
			fmt.Fprint(w, `{"id":"inq-1"}`)
		case "/pathfinder/inquiries/inq-1/user_view/":
			fmt.Fprint(w, `{"context":{"sheriff_challenge":{"type":"sms","id":"ch-1","phone_number":"***-1234"}}}`)
		case "/challenge/ch-1/respond/":
			fmt.Fprint(w, `{"status":"validated"}`)
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	opts := LoginOpts{
		Host:        ts.URL,
		HTTP:        ts.Client(),
		Username:    "alice@example.com",
		Password:    "secret",
		DeviceToken: "dt-fixed",
		CodeInput:   func(prompt string) (string, error) { return "123456", nil },
		TOTPSecret:  "",
		Out:         &bytes.Buffer{},
		Profile:     "default",
		PollEvery:   time.Millisecond,
	}
	sess, err := RunLogin(context.Background(), opts)
	if err != nil {
		t.Fatalf("RunLogin: %v", err)
	}
	if sess.AccessToken != "a" {
		t.Fatalf("AccessToken = %q", sess.AccessToken)
	}
	// Verify persisted in keychain (file fallback).
	re, err := robinhood.LoadFromKeychain("default")
	if err != nil {
		t.Fatalf("LoadFromKeychain: %v", err)
	}
	if re.AccessToken != "a" {
		t.Fatalf("persisted AccessToken = %q", re.AccessToken)
	}
}

// TestLogin_PushFlow approves via push and does NOT prompt for a code.
func TestLogin_PushFlow(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("ROBINHOOD_KEYCHAIN_BACKEND", "file")
	var first atomic.Bool
	polls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token/":
			if !first.Swap(true) {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(w, `{"verification_workflow":{"id":"wf-2"}}`)
				return
			}
			fmt.Fprint(w, `{"access_token":"ok","refresh_token":"r","expires_in":3600}`)
		case "/pathfinder/user_machine/":
			fmt.Fprint(w, `{"id":"inq-2"}`)
		case "/pathfinder/inquiries/inq-2/user_view/":
			fmt.Fprint(w, `{"context":{"sheriff_challenge":{"type":"prompt","id":"ch-2"}}}`)
		case "/push/ch-2/get_prompts_status/":
			polls++
			if polls < 2 {
				fmt.Fprint(w, `{"challenge_status":"issued"}`)
			} else {
				fmt.Fprint(w, `{"challenge_status":"validated"}`)
			}
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	askedForCode := false
	opts := LoginOpts{
		Host:        ts.URL,
		HTTP:        ts.Client(),
		Username:    "alice@example.com",
		Password:    "p",
		DeviceToken: "dt-2",
		CodeInput:   func(prompt string) (string, error) { askedForCode = true; return "", nil },
		Out:         &bytes.Buffer{},
		Profile:     "default",
		PollEvery:   time.Millisecond,
	}
	if _, err := RunLogin(context.Background(), opts); err != nil {
		t.Fatalf("RunLogin: %v", err)
	}
	if askedForCode {
		t.Fatalf("push flow should not prompt for SMS code")
	}
}

// TestLoginOpts_DoesNotLeakPassword — Fix H. LoginOpts must redact its
// password in String/GoString and default %+v formatters.
func TestLoginOpts_DoesNotLeakPassword(t *testing.T) {
	o := LoginOpts{Username: "u", Password: "SUPER-SECRET-PW"}
	for _, v := range []string{fmt.Sprint(o), fmt.Sprintf("%+v", o), fmt.Sprintf("%#v", o)} {
		if strings.Contains(v, "SUPER-SECRET-PW") {
			t.Fatalf("password leaked in %q", v)
		}
	}
}
