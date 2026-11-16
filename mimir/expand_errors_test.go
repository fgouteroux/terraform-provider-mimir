package mimir

import (
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// withValues returns a one-element block of zero values overridden by values.
func withValues(fields map[string]*schema.Schema, values map[string]interface{}) []interface{} {
	data := zeroValues(fields)
	for key, value := range values {
		data[key] = value
	}
	return []interface{}{data}
}

func TestValidateAlertmanagerURL(t *testing.T) {
	for value, wantErr := range map[string]bool{
		"":                             false,
		"https://events.pagerduty.com": false,
		"http://localhost:8080/api":    false,
		"ftp://example.com":            true,
		"example.com":                  true,
		"https://":                     true,
		"http://[::1":                  true,
	} {
		if _, errs := validateAlertmanagerURL(value, "url"); (len(errs) > 0) != wantErr {
			t.Errorf("validateAlertmanagerURL(%q) errors = %v, wantErr %v", value, errs, wantErr)
		}
	}
}

func TestValidateHostPort(t *testing.T) {
	for value, wantErr := range map[string]bool{
		"":                     false,
		"smtp.example.com:587": false,
		"[::1]:25":             false,
		"smtp.example.com":     true,
		"smtp.example.com:":    true,
	} {
		if _, errs := validateHostPort(value, "smarthost"); (len(errs) > 0) != wantErr {
			t.Errorf("validateHostPort(%q) errors = %v, wantErr %v", value, errs, wantErr)
		}
	}
}

func TestValidateLocation(t *testing.T) {
	for value, wantErr := range map[string]bool{
		"UTC":           false,
		"Europe/Paris":  false,
		"Mars/Olympus":  true,
		"not a zone !!": true,
	} {
		if _, errs := validateLocation(value, "location"); (len(errs) > 0) != wantErr {
			t.Errorf("validateLocation(%q) errors = %v, wantErr %v", value, errs, wantErr)
		}
	}
}

// Expand used to drop values it could not parse, sending a configuration
// without them. It now reports them.
func TestExpandReportsInvalidValues(t *testing.T) {
	_, err := expandGlobalConfig(withValues(globalConfigFieldsForTest(), map[string]interface{}{
		"smtp_smarthost": "smtp.example.com",
	}))
	if err == nil || !strings.Contains(err.Error(), "smtp_smarthost") {
		t.Errorf("expandGlobalConfig: want a smtp_smarthost error, got %v", err)
	}

	_, err = expandReceiverConfig([]interface{}{map[string]interface{}{
		"name": "hook",
		"webhook_configs": withValues(webhookConfigFields(), map[string]interface{}{
			"url":      "http://example.com",
			timeoutKey: "soon",
		}),
	}})
	if err == nil || !strings.Contains(err.Error(), `receiver "hook"`) {
		t.Errorf("expandReceiverConfig: want an error naming the receiver, got %v", err)
	}

	_, err = expandMuteTimeIntervalConfig([]interface{}{map[string]interface{}{
		"name": "weekend",
		"time_intervals": []interface{}{map[string]interface{}{
			"location": "Mars/Olympus",
		}},
	}})
	if err == nil || !strings.Contains(err.Error(), `time_interval "weekend"`) {
		t.Errorf("expandMuteTimeIntervalConfig: want an error naming the interval, got %v", err)
	}
}

func TestAccResourceAlertmanagerConfig_InvalidSmarthost(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
    resource "mimir_alertmanager_config" "mytenant" {
      global {
        smtp_smarthost = "smtp.example.com"
      }
      route {
        group_wait      = "30s"
        group_interval  = "5m"
        repeat_interval = "1h"
        receiver        = "email"
      }
      receiver {
        name = "email"
      }
    }`,
				ExpectError: regexp.MustCompile(`invalid address "smtp.example.com", expected host:port`),
			},
		},
	})
}
