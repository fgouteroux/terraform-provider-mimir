package mimir

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
)

// TestRedactCredentialsHidesAuthorization covers the debug dumps, which used to
// carry the bearer token and basic-auth pair straight into the Terraform log.
func TestRedactCredentialsHidesAuthorization(t *testing.T) {
	dump := "GET /api/v1/alerts HTTP/1.1\r\nHost: mimir\r\nAuthorization: Bearer super-secret-token\r\nX-Scope-OrgID: tenant\r\n\r\n"

	got := redactCredentials(dump)

	if strings.Contains(got, "super-secret-token") {
		t.Errorf("credential survived redaction:\n%s", got)
	}
	if !strings.Contains(got, "Authorization: <redacted>") {
		t.Errorf("Authorization header not redacted:\n%s", got)
	}
	if !strings.Contains(got, "X-Scope-OrgID: tenant") {
		t.Errorf("redaction ate a non-credential header:\n%s", got)
	}
}

// TestProxyLogGatedOnDebug pins that configuring a proxy stays silent unless
// debug is on. The line used to be unconditional: it wrote to stderr on every
// provider init, which GitHub Actions annotates as an error, and the proxy URL
// can embed credentials.
func TestProxyLogGatedOnDebug(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	//nolint:gosec // placeholder credentials in a proxy URL, this test never dials it
	if _, err := NewAPIClient(&apiClientOpt{
		uri: "https://mimir.example.com", proxyURL: "http://user:pass@localhost:3128",
		headers: map[string]string{}, timeout: 2, debug: false,
	}); err != nil {
		t.Fatalf("NewAPIClient: %v", err)
	}

	if strings.Contains(buf.String(), "Using proxy") {
		t.Errorf("proxy logged with debug off:\n%s", buf.String())
	}

	buf.Reset()
	if _, err := NewAPIClient(&apiClientOpt{
		uri: "https://mimir.example.com", proxyURL: "http://localhost:3128",
		headers: map[string]string{}, timeout: 2, debug: true,
	}); err != nil {
		t.Fatalf("NewAPIClient: %v", err)
	}
	if !strings.Contains(buf.String(), "Using proxy") {
		t.Error("proxy not logged with debug on")
	}
}
