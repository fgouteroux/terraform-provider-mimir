package mimir

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

// TestEmailRequiresTLSByDefault pins the SMTP default. require_tls used to
// default to false and was always emitted, so the generated configuration
// disabled STARTTLS even though upstream Alertmanager defaults it to true.
func TestEmailRequiresTLSByDefault(t *testing.T) {
	if got := emailConfigFields()["require_tls"].Default; got != true {
		t.Errorf("email require_tls defaults to %v, want true", got)
	}

	global := resourceMimirAlertmanagerConfigSchemaV1()["global"].Elem.(*schema.Resource).Schema
	if got := global["smtp_require_tls"].Default; got != true {
		t.Errorf("global smtp_require_tls defaults to %v, want true", got)
	}
}

// TestReceiverSecretsAreSensitive keeps credentials out of plan output.
func TestReceiverSecretsAreSensitive(t *testing.T) {
	for name, fields := range map[string]map[string]*schema.Schema{
		"pushover":  pushoverConfigFields(),
		"slack":     slackConfigFields(),
		"discord":   discordConfigFields(),
		"msteams":   msteamsConfigFields(),
		"msteamsv2": msteamsv2ConfigFields(),
	} {
		for _, attr := range []string{"token", "user_key", apiURLKey, "webhook_url"} {
			entry, declared := fields[attr]
			if !declared {
				continue
			}
			if !entry.Sensitive {
				t.Errorf("%s %q is not marked Sensitive; the credential is printed in plan output", name, attr)
			}
		}
	}
}
