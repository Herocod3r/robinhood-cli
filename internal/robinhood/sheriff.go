package robinhood

// Sheriff drives Robinhood's "user_machine" device-verification workflow
// (SMS / email codes + app-prompt approvals). Request and response shapes
// mirror robin_stocks authentication.py as of 2026-04-20 (master branch,
// https://github.com/jmfernandes/robin_stocks/blob/master/robin_stocks/
// robinhood/authentication.py) — see _validate_sherrif_id. Key points:
//
//   - user_machine POST body is {device_id, flow:"suv", input:{workflow_id}}.
//     No X-Robinhood-Challenge-Response-Id header here; that header belongs
//     on the final /oauth2/token/ re-attempt (Task 11).
//   - user_view returns context.sheriff_challenge at the TOP LEVEL (not
//     under type_context). robin_stocks reads data["context"]["sheriff_
//     challenge"] directly.
//   - sheriff_challenge has: {id, type, status, phone_number?, email?}.
//     type ∈ {"sms", "email", "prompt"}; we accept robin_stocks's variants.
//
// This file implements Sheriff.Start (user_machine + user_view poll). The
// respond/push code path is in Sheriff.RespondCode and Sheriff.WaitPush
// (Task 10).
//
// Fix F: the poll loop respects ctx.Done() and fails fast on JSON decode
// errors or stale 404/410/401 responses, so the login CLI can surface a
// clear hint rather than hang for the 5-minute outer deadline.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// SheriffKind classifies the challenge the server is asking the user for.
type SheriffKind int

const (
	SheriffUnknown SheriffKind = iota
	SheriffSMS
	SheriffEmail
	SheriffPush
)

func (k SheriffKind) String() string {
	switch k {
	case SheriffSMS:
		return "sms"
	case SheriffEmail:
		return "email"
	case SheriffPush:
		return "prompt"
	default:
		return "unknown"
	}
}

// SheriffStep is what Start returns — the state the login UI needs to
// collect one piece of input from the user (or wait for push).
type SheriffStep struct {
	InquiryID   string
	ChallengeID string
	Kind        SheriffKind
	// Detail is a human-readable hint: "SMS to ***-***-1234".
	Detail string
	// Status mirrors the sheriff_challenge.status field (e.g. "issued",
	// "validated"); used by Task 10 callers to short-circuit when the
	// challenge is already validated.
	Status string
}

// Sheriff drives the user_machine + inquiry poll + challenge response flow.
// HTTP and Clock are injectable for tests; PollEvery defaults to 2s.
type Sheriff struct {
	BaseURL   string
	HTTP      *http.Client
	PollEvery time.Duration
	Clock     func() time.Time
}

func (s *Sheriff) poll() time.Duration {
	if s.PollEvery > 0 {
		return s.PollEvery
	}
	return 2 * time.Second
}

// userMachineReq matches robin_stocks authentication.py master:
//
//	machine_payload = {
//	    'device_id': device_token,
//	    'flow': 'suv',
//	    'input': {'workflow_id': workflow_id},
//	}
type userMachineReq struct {
	DeviceID string             `json:"device_id"`
	Flow     string             `json:"flow"`
	Input    userMachineReqBody `json:"input"`
}

type userMachineReqBody struct {
	WorkflowID string `json:"workflow_id"`
}

type userMachineResp struct {
	ID string `json:"id"`
}

// userViewResp matches robin_stocks: inquiries_response["context"]
// ["sheriff_challenge"]. Note there is NO type_context wrapper here.
type userViewResp struct {
	Context struct {
		SheriffChallenge struct {
			ID          string `json:"id"`
			Type        string `json:"type"`
			Status      string `json:"status"`
			PhoneNumber string `json:"phone_number"`
			Email       string `json:"email"`
		} `json:"sheriff_challenge"`
	} `json:"context"`
}

// Start posts /pathfinder/user_machine/ and polls /pathfinder/inquiries/
// {id}/user_view/ until the server exposes a sheriff_challenge. Returns
// the step the caller should handle (SMS/email code entry, or push wait).
//
// Context cancellation aborts the poll immediately. See Fix F.
func (s *Sheriff) Start(ctx context.Context, workflowID, deviceToken string) (*SheriffStep, error) {
	body, err := json.Marshal(userMachineReq{
		DeviceID: deviceToken,
		Flow:     "suv",
		Input:    userMachineReqBody{WorkflowID: workflowID},
	})
	if err != nil {
		return nil, &APIError{Code: CodeValidation, Message: err.Error()}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.BaseURL+"/pathfinder/user_machine/", bytes.NewReader(body))
	if err != nil {
		return nil, &APIError{Code: CodeValidation, Message: err.Error()}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "robinhood-cli (+https://github.com/herocod3r/robinhood-cli)")
	// Note: robin_stocks authentication.py master does NOT set
	// X-Robinhood-Challenge-Response-Id here. That header rides on the
	// final /oauth2/token/ re-attempt (Task 11) instead.

	resp, err := s.HTTP.Do(req)
	if err != nil {
		return nil, &APIError{Code: CodeRobinhoodUnavailable, Message: err.Error()}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		buf, _ := io.ReadAll(resp.Body)
		return nil, &APIError{
			Code:       CodeSheriffRequired,
			Message:    fmt.Sprintf("user_machine HTTP %d: %s", resp.StatusCode, string(buf)),
			HTTPStatus: resp.StatusCode,
			WorkflowID: workflowID,
		}
	}
	var umr userMachineResp
	if err := json.NewDecoder(resp.Body).Decode(&umr); err != nil {
		return nil, &APIError{Code: CodeRobinhoodUnavailable, Message: "user_machine decode: " + err.Error()}
	}
	if umr.ID == "" {
		return nil, &APIError{Code: CodeSheriffRequired, Message: "user_machine response missing id", WorkflowID: workflowID}
	}
	inquiryID := umr.ID

	// Poll user_view until sheriff_challenge is populated or ctx fires.
	for {
		if err := ctx.Err(); err != nil {
			return nil, &APIError{Code: CodeSheriffRequired, Message: err.Error(), WorkflowID: workflowID}
		}
		step, err := s.fetchUserView(ctx, inquiryID, workflowID)
		if err != nil {
			return nil, err
		}
		if step != nil {
			return step, nil
		}
		select {
		case <-ctx.Done():
			return nil, &APIError{Code: CodeSheriffRequired, Message: ctx.Err().Error(), WorkflowID: workflowID}
		case <-time.After(s.poll()):
		}
	}
}

// fetchUserView makes a single GET to the inquiries user_view endpoint.
// Returns (nil, nil) when the sheriff_challenge isn't ready yet — caller
// should poll again. Returns a terminal APIError on decode failure or when
// the server signals the workflow is gone (401/404/410) per Fix F.
func (s *Sheriff) fetchUserView(ctx context.Context, inquiryID, workflowID string) (*SheriffStep, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.BaseURL+"/pathfinder/inquiries/"+inquiryID+"/user_view/", nil)
	if err != nil {
		return nil, &APIError{Code: CodeValidation, Message: err.Error()}
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "robinhood-cli (+https://github.com/herocod3r/robinhood-cli)")
	resp, err := s.HTTP.Do(req)
	if err != nil {
		return nil, &APIError{Code: CodeRobinhoodUnavailable, Message: err.Error()}
	}
	defer resp.Body.Close()
	// Fix F: specific handling for stale workflow states so the CLI can
	// print a clear hint instead of spinning.
	switch resp.StatusCode {
	case http.StatusOK:
		// fall through
	case http.StatusNotFound, http.StatusGone:
		return nil, &APIError{
			Code:       CodeSheriffRequired,
			Message:    "verification expired — run: rh login again",
			HTTPStatus: resp.StatusCode,
			WorkflowID: workflowID,
		}
	case http.StatusUnauthorized:
		return nil, &APIError{
			Code:       CodeSessionExpired,
			Message:    "workflow token expired",
			HTTPStatus: resp.StatusCode,
			WorkflowID: workflowID,
		}
	default:
		return nil, &APIError{
			Code:       CodeSheriffRequired,
			Message:    fmt.Sprintf("user_view HTTP %d", resp.StatusCode),
			HTTPStatus: resp.StatusCode,
			WorkflowID: workflowID,
		}
	}
	var uv userViewResp
	if err := json.NewDecoder(resp.Body).Decode(&uv); err != nil {
		// Fix F: decode errors are terminal. Don't spin on malformed JSON.
		return nil, &APIError{Code: CodeRobinhoodUnavailable, Message: "user_view decode: " + err.Error()}
	}
	ch := uv.Context.SheriffChallenge
	if ch.ID == "" && ch.Type == "" {
		return nil, nil // keep polling
	}
	step := &SheriffStep{
		InquiryID:   inquiryID,
		ChallengeID: ch.ID,
		Status:      ch.Status,
	}
	switch ch.Type {
	case "sms":
		step.Kind = SheriffSMS
		step.Detail = fmt.Sprintf("SMS to %s", ch.PhoneNumber)
	case "email":
		step.Kind = SheriffEmail
		step.Detail = fmt.Sprintf("email to %s", ch.Email)
	case "prompt":
		step.Kind = SheriffPush
		step.Detail = "push notification to your Robinhood app"
	default:
		return nil, &APIError{
			Code:       CodeSheriffRequired,
			Message:    fmt.Sprintf("unknown sheriff challenge type: %s", ch.Type),
			WorkflowID: workflowID,
		}
	}
	return step, nil
}

type challengeRespReq struct {
	Response string `json:"response"`
}

type challengeRespResp struct {
	Status string `json:"status"`
	Detail string `json:"detail"`
}

// sheriffSuccessStatuses lists the status sentinels that mean "the user
// cleared the challenge". Robinhood has returned each of these in
// different rollout cohorts — see Codex Fix G.
var sheriffSuccessStatuses = map[string]bool{
	"validated": true,
	"resolved":  true,
	"success":   true,
}

// sheriffFailureStatuses lists the status sentinels that mean "the
// challenge cannot be completed" (Fix G). Unknown statuses are treated as
// "continue / inconclusive" — the caller will re-check later.
var sheriffFailureStatuses = map[string]bool{
	"failed":   true,
	"declined": true,
	"expired":  true,
}

// RespondCode submits the user-entered code for SMS/email challenges.
// Success when status ∈ {validated, resolved, success}. Rejects only
// explicit failure sentinels; unknown statuses are accepted so the next
// step of the login loop can re-verify against the workflow.
func (s *Sheriff) RespondCode(ctx context.Context, step *SheriffStep, code string) error {
	body, _ := json.Marshal(challengeRespReq{Response: code})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.BaseURL+"/challenge/"+step.ChallengeID+"/respond/", bytes.NewReader(body))
	if err != nil {
		return &APIError{Code: CodeValidation, Message: err.Error()}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "robinhood-cli (+https://github.com/herocod3r/robinhood-cli)")
	resp, err := s.HTTP.Do(req)
	if err != nil {
		return &APIError{Code: CodeRobinhoodUnavailable, Message: err.Error()}
	}
	defer resp.Body.Close()
	buf, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var er challengeRespResp
		_ = json.Unmarshal(buf, &er)
		msg := firstNonEmpty(er.Detail, strings.TrimSpace(string(buf)), fmt.Sprintf("HTTP %d", resp.StatusCode))
		return &APIError{Code: CodeSheriffRequired, Message: msg, HTTPStatus: resp.StatusCode}
	}
	var cr challengeRespResp
	if err := json.Unmarshal(buf, &cr); err != nil {
		// Fix F: decode failures are terminal, not poll-again.
		return &APIError{Code: CodeRobinhoodUnavailable, Message: "respond decode: " + err.Error()}
	}
	if sheriffSuccessStatuses[cr.Status] {
		return nil
	}
	if sheriffFailureStatuses[cr.Status] {
		msg := firstNonEmpty(cr.Detail, "challenge "+cr.Status)
		return &APIError{Code: CodeSheriffRequired, Message: msg}
	}
	// Unknown status: accept optimistically. The outer login flow will
	// re-verify against the workflow state; we don't want to reject on a
	// newly-added status string.
	return nil
}

type promptsStatusResp struct {
	ChallengeStatus string `json:"challenge_status"`
}

// WaitPush polls /push/<id>/get_prompts_status/ until the server reports
// a success sentinel. Any explicit failure sentinel becomes an error;
// unknown statuses continue polling. Context cancellation aborts the poll.
func (s *Sheriff) WaitPush(ctx context.Context, step *SheriffStep) error {
	for {
		if err := ctx.Err(); err != nil {
			return &APIError{Code: CodeSheriffRequired, Message: err.Error()}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.BaseURL+"/push/"+step.ChallengeID+"/get_prompts_status/", nil)
		if err != nil {
			return &APIError{Code: CodeValidation, Message: err.Error()}
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "robinhood-cli (+https://github.com/herocod3r/robinhood-cli)")
		resp, err := s.HTTP.Do(req)
		if err != nil {
			return &APIError{Code: CodeRobinhoodUnavailable, Message: err.Error()}
		}
		var ps promptsStatusResp
		decodeErr := json.NewDecoder(resp.Body).Decode(&ps)
		_ = resp.Body.Close()
		if decodeErr != nil {
			// Fix F: decode failures are terminal, not poll-again.
			return &APIError{Code: CodeRobinhoodUnavailable, Message: "push status decode: " + decodeErr.Error()}
		}
		if sheriffSuccessStatuses[ps.ChallengeStatus] {
			return nil
		}
		if sheriffFailureStatuses[ps.ChallengeStatus] {
			return &APIError{Code: CodeSheriffRequired, Message: "push challenge " + ps.ChallengeStatus}
		}
		// issued / pending / "" / unknown -> keep polling.
		select {
		case <-ctx.Done():
			return &APIError{Code: CodeSheriffRequired, Message: ctx.Err().Error()}
		case <-time.After(s.poll()):
		}
	}
}
