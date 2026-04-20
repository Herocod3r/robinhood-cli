package robinhood

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestSheriff_Start_SMS asserts the request body matches robin_stocks
// authentication.py master: {device_id, flow:"suv", input:{workflow_id}}
// and that user_view's top-level context.sheriff_challenge is parsed.
func TestSheriff_Start_SMS(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/pathfinder/user_machine/":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode user_machine body: %v", err)
			}
			if body["device_id"] == "" || body["device_id"] == nil {
				t.Fatalf("device_id missing: %v", body)
			}
			if body["flow"] != "suv" {
				t.Fatalf("flow = %v, want suv", body["flow"])
			}
			input, _ := body["input"].(map[string]any)
			if input == nil || input["workflow_id"] != "wf-123" {
				t.Fatalf("input.workflow_id = %v, want wf-123", input)
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"id":"inq-1"}`)
		case "/pathfinder/inquiries/inq-1/user_view/":
			fmt.Fprint(w, `{"context":{"sheriff_challenge":{"type":"sms","id":"ch-1","status":"issued","phone_number":"***-***-1234"}}}`)
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer ts.Close()
	s := &Sheriff{
		BaseURL:   ts.URL,
		HTTP:      ts.Client(),
		PollEvery: time.Millisecond,
	}
	step, err := s.Start(context.Background(), "wf-123", "dev-tok")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if step.Kind != SheriffSMS {
		t.Fatalf("Kind = %v, want SheriffSMS", step.Kind)
	}
	if step.ChallengeID != "ch-1" {
		t.Fatalf("ChallengeID = %q", step.ChallengeID)
	}
	if !strings.Contains(step.Detail, "1234") {
		t.Fatalf("Detail should surface phone hint; got %q", step.Detail)
	}
}

func TestSheriff_Start_DeviceApproval(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/pathfinder/user_machine/":
			fmt.Fprint(w, `{"id":"inq-2"}`)
		case "/pathfinder/inquiries/inq-2/user_view/":
			fmt.Fprint(w, `{"context":{"sheriff_challenge":{"type":"prompt","id":"ch-2","status":"issued"}}}`)
		}
	}))
	defer ts.Close()
	s := &Sheriff{BaseURL: ts.URL, HTTP: ts.Client(), PollEvery: time.Millisecond}
	step, err := s.Start(context.Background(), "wf-2", "dev-tok")
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	if step.Kind != SheriffPush {
		t.Fatalf("Kind = %v, want SheriffPush", step.Kind)
	}
}

func TestSheriff_Start_UnknownType_ErrorsOut(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/pathfinder/user_machine/":
			fmt.Fprint(w, `{"id":"inq-3"}`)
		case "/pathfinder/inquiries/inq-3/user_view/":
			fmt.Fprint(w, `{"context":{"sheriff_challenge":{"type":"quantum","id":"ch-3","status":"issued"}}}`)
		}
	}))
	defer ts.Close()
	s := &Sheriff{BaseURL: ts.URL, HTTP: ts.Client(), PollEvery: time.Millisecond}
	_, err := s.Start(context.Background(), "wf-3", "dev-tok")
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.Code != CodeSheriffRequired {
		t.Fatalf("err = %v; want CodeSheriffRequired", err)
	}
	if !strings.Contains(apiErr.Message, "quantum") {
		t.Fatalf("error should name the unknown kind; got %q", apiErr.Message)
	}
}

func TestSheriff_Start_RespectsCtxDeadline(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/pathfinder/user_machine/" {
			fmt.Fprint(w, `{"id":"inq-4"}`)
			return
		}
		// user_view never produces a sheriff_challenge — force poll loop.
		fmt.Fprint(w, `{"context":{}}`)
	}))
	defer ts.Close()
	s := &Sheriff{BaseURL: ts.URL, HTTP: ts.Client(), PollEvery: 5 * time.Millisecond}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := s.Start(ctx, "wf-4", "dev-tok")
	if err == nil {
		t.Fatalf("expected timeout error")
	}
}

// Fix F: fetchUserView must classify 404/410 as "verification expired" and
// 401 as "workflow token expired" — both as specific CodeSheriffRequired /
// CodeSessionExpired so the login UI can print a clear hint.
func TestSheriff_Start_UserView404_IsExpired(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/pathfinder/user_machine/" {
			fmt.Fprint(w, `{"id":"inq-expired"}`)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()
	s := &Sheriff{BaseURL: ts.URL, HTTP: ts.Client(), PollEvery: time.Millisecond}
	_, err := s.Start(context.Background(), "wf-x", "dev-tok")
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("err = %v", err)
	}
	if apiErr.Code != CodeSheriffRequired {
		t.Fatalf("code = %q, want CodeSheriffRequired", apiErr.Code)
	}
	if !strings.Contains(apiErr.Message, "expired") {
		t.Fatalf("message should mention expired; got %q", apiErr.Message)
	}
}

func TestSheriff_Start_UserView401_IsSessionExpired(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/pathfinder/user_machine/" {
			fmt.Fprint(w, `{"id":"inq-401"}`)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()
	s := &Sheriff{BaseURL: ts.URL, HTTP: ts.Client(), PollEvery: time.Millisecond}
	_, err := s.Start(context.Background(), "wf-x", "dev-tok")
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.Code != CodeSessionExpired {
		t.Fatalf("err = %v, want CodeSessionExpired", err)
	}
}

// Fix F: decode errors must return immediately rather than spin-poll.
func TestSheriff_Start_UserViewDecodeError_FailsFast(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/pathfinder/user_machine/" {
			fmt.Fprint(w, `{"id":"inq-bad"}`)
			return
		}
		// Write incomplete body that is not valid JSON after a few bytes.
		_, _ = io.WriteString(w, `{"context":{"sher`)
	}))
	defer ts.Close()
	s := &Sheriff{BaseURL: ts.URL, HTTP: ts.Client(), PollEvery: time.Millisecond}
	_, err := s.Start(context.Background(), "wf-x", "dev-tok")
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("err = %v", err)
	}
	if apiErr.Code != CodeRobinhoodUnavailable {
		t.Fatalf("code = %q, want CodeRobinhoodUnavailable", apiErr.Code)
	}
}
