package vcr

import (
	"net/http"
	"strings"
	"testing"

	"gopkg.in/dnaeon/go-vcr.v3/cassette"
)

func TestSanitizeHook_RedactsAuthorizationHeader(t *testing.T) {
	i := &cassette.Interaction{
		Request: cassette.Request{
			Headers: http.Header{"Authorization": []string{"Bearer abc.def.ghi.long-secret-token-value"}},
		},
		Response: cassette.Response{
			Headers: http.Header{"Content-Type": []string{"application/json"}},
			Body:    `{"ok":true}`,
		},
	}
	if err := SanitizeHook(i); err != nil {
		t.Fatal(err)
	}
	if got := i.Request.Headers.Get("Authorization"); got != "Bearer REDACTED" {
		t.Errorf("Authorization = %q, want %q", got, "Bearer REDACTED")
	}
}

func TestSanitizeHook_RedactsJSONTokenFields(t *testing.T) {
	i := &cassette.Interaction{
		Request: cassette.Request{
			Headers: http.Header{},
			Body:    `{"grant_type":"refresh_token","refresh_token":"r-0123456789abcdef"}`,
		},
		Response: cassette.Response{
			Headers: http.Header{},
			Body:    `{"access_token":"at-0123456789","refresh_token":"rt-0123456789","device_token":"01234567-89ab-cdef-0123-456789abcdef","expires_in":3600}`,
		},
	}
	if err := SanitizeHook(i); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(i.Request.Body, "r-0123456789abcdef") {
		t.Errorf("request body still contains refresh_token value: %s", i.Request.Body)
	}
	if strings.Contains(i.Response.Body, "at-0123456789") ||
		strings.Contains(i.Response.Body, "rt-0123456789") ||
		strings.Contains(i.Response.Body, "01234567-89ab-cdef-0123-456789abcdef") {
		t.Errorf("response body still contains secrets: %s", i.Response.Body)
	}
	if !strings.Contains(i.Response.Body, `"access_token":"REDACTED"`) {
		t.Errorf("access_token not redacted: %s", i.Response.Body)
	}
}

func TestSanitizeHook_RedactsBearerInBody(t *testing.T) {
	i := &cassette.Interaction{
		Response: cassette.Response{
			Headers: http.Header{},
			Body:    `Server echoed: Bearer abcdefghijklmnop.q.r`,
		},
	}
	if err := SanitizeHook(i); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(i.Response.Body, "Bearer REDACTED") {
		t.Errorf("bearer not redacted: %s", i.Response.Body)
	}
}

func TestSanitizeHook_Idempotent(t *testing.T) {
	i := &cassette.Interaction{
		Request: cassette.Request{
			Headers: http.Header{"Authorization": []string{"Bearer REDACTED"}},
		},
		Response: cassette.Response{
			Headers: http.Header{},
			Body:    `{"access_token":"REDACTED"}`,
		},
	}
	if err := SanitizeHook(i); err != nil {
		t.Fatal(err)
	}
	if i.Request.Headers.Get("Authorization") != "Bearer REDACTED" {
		t.Errorf("not idempotent: %q", i.Request.Headers.Get("Authorization"))
	}
	if i.Response.Body != `{"access_token":"REDACTED"}` {
		t.Errorf("body mutated: %s", i.Response.Body)
	}
}
