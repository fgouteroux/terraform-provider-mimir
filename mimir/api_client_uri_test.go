package mimir

import (
	"strings"
	"testing"
)

// TestNewAPIClientTrimsEveryURI covers the per-component overrides. Only `uri`
// used to have its trailing slash removed, so a ruler/alertmanager/distributor
// URL configured with one produced a request path with a double slash.
func TestNewAPIClientTrimsEveryURI(t *testing.T) {
	client, err := NewAPIClient(&apiClientOpt{
		uri:             "https://mimir.example.com/",
		rulerURI:        "https://ruler.example.com/",
		alertmanagerURI: "https://am.example.com/",
		distributorURI:  "https://distributor.example.com/",
		headers:         map[string]string{},
		timeout:         2,
	})
	if err != nil {
		t.Fatalf("NewAPIClient: %v", err)
	}

	for name, got := range map[string]string{
		"uri":             client.uri,
		"rulerURI":        client.rulerURI,
		"alertmanagerURI": client.alertmanagerURI,
		"distributorURI":  client.distributorURI,
	} {
		if strings.HasSuffix(got, "/") {
			t.Errorf("%s = %q, trailing slash not trimmed; request paths would contain a double slash", name, got)
		}
	}
}

// TestNewAPIClientRejectsUnparseableCA pins that an unusable ca bundle fails
// loudly. AppendCertsFromPEM's return value used to be ignored, leaving an empty
// root pool and surfacing later as an unrelated TLS handshake error.
func TestNewAPIClientRejectsUnparseableCA(t *testing.T) {
	_, err := NewAPIClient(&apiClientOpt{
		uri:     "https://mimir.example.com",
		ca:      "-----BEGIN CERTIFICATE-----\nnot base64 at all\n-----END CERTIFICATE-----",
		headers: map[string]string{},
		timeout: 2,
	})
	if err == nil {
		t.Fatal("expected an error for a ca bundle containing no usable certificate")
	}
	if !strings.Contains(err.Error(), "PEM") {
		t.Errorf("error should mention the PEM bundle, got: %v", err)
	}
}
