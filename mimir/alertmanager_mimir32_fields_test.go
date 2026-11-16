package mimir

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestSuppressEquivalentYAML(t *testing.T) {
	cases := []struct {
		oldValue, newValue string
		want               bool
	}{
		{"text: hi\nstatus: '{{ .Status }}'\n", `{"status":"{{ .Status }}","text":"hi"}`, true},
		{"a: 1\n", "a: 2\n", false},
		{"a: 1\n", "", false},
		{"", "", true},
		{"a: [1\n", "a: [1\n", true},
		{"a: [1\n", "a: 1\n", false},
	}
	for _, c := range cases {
		if got := suppressEquivalentYAML("", c.oldValue, c.newValue, nil); got != c.want {
			t.Errorf("suppressEquivalentYAML(%q, %q) = %v, want %v", c.oldValue, c.newValue, got, c.want)
		}
	}
}

func TestValidateYAMLDocument(t *testing.T) {
	for value, wantErr := range map[string]bool{
		"":                         false,
		`{"text":"{{ .Status }}"}`: false,
		"text: hi\n":               false,
		"text: [hi\n":              true,
	} {
		if _, errs := validateYAMLDocument(value, "payload"); (len(errs) > 0) != wantErr {
			t.Errorf("validateYAMLDocument(%q) errors = %v, wantErr %v", value, errs, wantErr)
		}
	}
}

// The Mimir 3.2 fields must survive expand then flatten, and stay out of the
// configuration when unset so older Mimir versions do not reject it.
func TestAlertmanagerMimir32FieldsRoundTrip(t *testing.T) {
	in := []interface{}{map[string]interface{}{
		"name":    "team",
		labelsKey: map[string]interface{}{"team": "infra"},
		"webhook_configs": []interface{}{map[string]interface{}{
			"url":     "http://example.com",
			"payload": `{"text":"{{ .Status }}"}`,
		}},
		"slack_configs": []interface{}{map[string]interface{}{
			"api_url":        "https://slack.com/api/chat.postMessage",
			"update_message": true,
		}},
	}}

	receivers, err := expandReceiverConfig(in)
	if err != nil {
		t.Fatal(err)
	}
	if got := receivers[0].Labels["team"]; got != "infra" {
		t.Errorf("labels not expanded: %v", receivers[0].Labels)
	}
	if !receivers[0].SlackConfigs[0].UpdateMessage {
		t.Error("update_message not expanded")
	}

	out := flattenReceiverConfig(receivers)[0].(map[string]interface{})
	payload := out["webhook_configs"].([]interface{})[0].(map[string]interface{})["payload"].(string)
	if !suppressEquivalentYAML("", payload, `{"text":"{{ .Status }}"}`, nil) {
		t.Errorf("payload round trip changed the document: %q", payload)
	}
	if out["slack_configs"].([]interface{})[0].(map[string]interface{})["update_message"] != true {
		t.Error("update_message not flattened")
	}

	unset, err := expandReceiverConfig([]interface{}{map[string]interface{}{
		"name":            "plain",
		labelsKey:         map[string]interface{}{},
		"webhook_configs": []interface{}{map[string]interface{}{"url": "http://example.com", "payload": ""}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if unset[0].Labels != nil || unset[0].WebhookConfigs[0].Payload != nil {
		t.Errorf("unset fields must stay nil, got labels=%v payload=%v", unset[0].Labels, unset[0].WebhookConfigs[0].Payload)
	}
}

func TestAccResourceAlertmanagerConfig_Mimir32Fields(t *testing.T) {
	skipBelowMimirVersion(t, "3.2.0")

	client, err := NewAPIClient(setupClient())
	if err != nil {
		t.Fatal(err)
	}

	const res = "mimir_alertmanager_config.mytenant"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMimirAlertmanagerConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAlertmanagerConfigMimir32Fields,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirAlertmanagerConfigExists(res, client),
					resource.TestCheckResourceAttr(res, "receiver.0.labels.team", "infra"),
					resource.TestMatchResourceAttr(res, "receiver.0.webhook_configs.0.payload", regexp.MustCompile(`Status`)),
					resource.TestCheckResourceAttr(res, "receiver.1.slack_configs.0.update_message", "true"),
				),
			},
			{
				// Re-applying the same configuration must be a no-op: the JSON
				// payload is read back as YAML and has to compare equal.
				Config:   testAccResourceAlertmanagerConfigMimir32Fields,
				PlanOnly: true,
			},
		},
	})
}

const testAccResourceAlertmanagerConfigMimir32Fields = `
    resource "mimir_alertmanager_config" "mytenant" {
      route {
        group_by = ["..."]
        group_wait = "30s"
        group_interval = "5m"
        repeat_interval = "1h"
        receiver = "webhook"
      }
      receiver {
        name = "webhook"
        labels = {
          team = "infra"
        }
        webhook_configs {
          url = "http://example.com/hook"
          payload = jsonencode({
            text   = "{{ .Status }}"
            alerts = "{{ len .Alerts }}"
          })
        }
      }
      receiver {
        name = "slack"
        slack_configs {
          api_url = "https://slack.com/api/chat.postMessage"
          channel = "#alerts"
          update_message = true
        }
      }
    }
`
