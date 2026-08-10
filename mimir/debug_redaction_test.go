package mimir

import (
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
