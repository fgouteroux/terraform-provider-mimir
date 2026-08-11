package mimir

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/prometheus/common/model"
	"gopkg.in/yaml.v3"
)

// zeroValues mirrors what the Terraform SDK hands an expand function for a block
// whose optional attributes are all left unset: every declared key is present,
// holding the zero value of its type. Deriving it from the schema keeps this
// helper correct as attributes are added.
func zeroValues(s map[string]*schema.Schema) map[string]interface{} {
	data := make(map[string]interface{}, len(s))
	for key, attr := range s {
		switch attr.Type {
		case schema.TypeBool:
			data[key] = false
		case schema.TypeInt:
			data[key] = 0
		case schema.TypeFloat:
			data[key] = 0.0
		case schema.TypeList, schema.TypeSet:
			data[key] = []interface{}{}
		case schema.TypeMap:
			data[key] = map[string]interface{}{}
		default:
			data[key] = ""
		}
	}
	return data
}

func globalConfigFieldsForTest() map[string]*schema.Schema {
	return resourceMimirAlertmanagerConfigSchemaV1()["global"].Elem.(*schema.Resource).Schema
}

func inhibitRuleFieldsForTest() map[string]*schema.Schema {
	return resourceMimirAlertmanagerConfigSchemaV1()["inhibit_rule"].Elem.(*schema.Resource).Schema
}

func marshalYAML(t *testing.T, v interface{}) string {
	t.Helper()
	out, err := yaml.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return string(out)
}

// TestExpandAlertmanagerConfigOmitsUnsetKeys is the guard for Mimir version
// compatibility. Mimir refuses a configuration carrying a key it does not know,
// so an attribute the practitioner never set must not reach the generated YAML —
// otherwise every user on a Mimir older than the one that introduced the field
// would have their configuration rejected.
func TestExpandAlertmanagerConfigOmitsUnsetKeys(t *testing.T) {
	unset := func(fields map[string]*schema.Schema) []interface{} {
		return []interface{}{zeroValues(fields)}
	}

	cfg := alertmanagerConfig{
		Global: expandGlobalConfig(unset(globalConfigFieldsForTest())),
		Receivers: []*receiver{{
			Name:             "everything-unset",
			WebhookConfigs:   expandWebhookConfig(unset(webhookConfigFields())),
			DiscordConfigs:   expandDiscordConfig(unset(discordConfigFields())),
			EmailConfigs:     expandEmailConfig(unset(emailConfigFields())),
			PagerdutyConfigs: expandPagerdutyConfig(unset(pagerdutyConfigFields())),
			SlackConfigs:     expandSlackConfig(unset(slackConfigFields())),
			PushoverConfigs:  expandPushoverConfig(unset(pushoverConfigFields())),
			TelegramConfigs:  expandTelegramConfig(unset(telegramConfigFields())),
		}},
		InhibitRules: expandInhibitRuleConfig(unset(inhibitRuleFieldsForTest())),
	}

	got := marshalYAML(t, cfg)

	// Every key added for a Mimir version newer than the oldest one this provider
	// targets. None may appear when the corresponding attribute is unset.
	for _, key := range []string{
		"timeout:",
		"source:",
		"device:",
		"ttl:",
		"monospace:",
		"content:",
		"avatar_url:",
		"message_thread_id:",
		"app_token:",
		"app_url:",
		"message_text:",
		"force_implicit_tls:",
		"threading:",
		"smtp_tls_config:",
		"smtp_force_implicit_tls:",
		"slack_app_token:",
		"slack_app_url:",
		"telegram_bot_token:",
	} {
		// Anchor on the start of the line so that, say, "timeout:" does not match
		// inside "resolve_timeout:".
		found := regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(key)).MatchString(got)
		if found {
			t.Errorf("unset attribute produced key %q in generated configuration:\n%s", key, got)
		}
	}
}

func TestExpandAlertmanagerConfigEmitsSetKeys(t *testing.T) {
	with := func(fields map[string]*schema.Schema, overrides map[string]interface{}) []interface{} {
		data := zeroValues(fields)
		for key, value := range overrides {
			if _, declared := data[key]; !declared {
				t.Fatalf("attribute %q is not declared in the schema", key)
			}
			data[key] = value
		}
		return []interface{}{data}
	}

	cfg := alertmanagerConfig{
		//nolint:gosec // placeholder credentials, this test never reaches a network
		Global: expandGlobalConfig(with(globalConfigFieldsForTest(), map[string]interface{}{
			"smtp_force_implicit_tls": true,
			"slack_app_token":         "xapp-1-token",
			"slack_app_url":           "https://slack.example.com/api",
			"telegram_bot_token":      "bot-token",
			"smtp_tls_config":         []interface{}{map[string]interface{}{"server_name": "smtp.example.com", "insecure_skip_verify": false, "min_version": "", "max_version": ""}},
		})),
		Receivers: []*receiver{{
			Name: "everything-set",
			WebhookConfigs: expandWebhookConfig(with(webhookConfigFields(), map[string]interface{}{
				"url":     "https://example.com/hook",
				"timeout": "10s",
			})),
			DiscordConfigs: expandDiscordConfig(with(discordConfigFields(), map[string]interface{}{
				"content":    "content template",
				usernameKey:  "alertmanager",
				"avatar_url": "https://example.com/avatar.png",
			})),
			EmailConfigs: expandEmailConfig(with(emailConfigFields(), map[string]interface{}{
				"to":                 "ops@example.com",
				"force_implicit_tls": true,
				"threading":          []interface{}{map[string]interface{}{"enabled": true, "thread_by_date": "daily"}},
				"tls_config":         []interface{}{map[string]interface{}{"server_name": "mail.example.com", "insecure_skip_verify": false, "min_version": "", "max_version": ""}},
			})),
			PagerdutyConfigs: expandPagerdutyConfig(with(pagerdutyConfigFields(), map[string]interface{}{
				"source":  "mimir",
				"timeout": "15s",
			})),
			//nolint:gosec // placeholder credentials, this test never reaches a network
			SlackConfigs: expandSlackConfig(with(slackConfigFields(), map[string]interface{}{
				"app_token":    "xapp-1-token",
				"app_url":      "https://slack.example.com/api",
				"message_text": "text template",
				"timeout":      "20s",
			})),
			PushoverConfigs: expandPushoverConfig(with(pushoverConfigFields(), map[string]interface{}{
				"device":    "phone",
				"ttl":       "1h",
				"monospace": true,
			})),
			TelegramConfigs: expandTelegramConfig(with(telegramConfigFields(), map[string]interface{}{
				"message_thread_id": 42,
			})),
		}},
		InhibitRules: expandInhibitRuleConfig(with(inhibitRuleFieldsForTest(), map[string]interface{}{
			"name": "cluster-down",
		})),
	}

	got := marshalYAML(t, cfg)

	for _, want := range []string{
		"timeout: 10s",
		"timeout: 15s",
		"timeout: 20s",
		"source: mimir",
		"device: phone",
		"ttl: 1h0m0s",
		"monospace: true",
		"content: content template",
		"avatar_url: https://example.com/avatar.png",
		"message_thread_id: 42",
		"app_token: xapp-1-token",
		"app_url: https://slack.example.com/api",
		"message_text: text template",
		"force_implicit_tls: true",
		"threading:",
		"thread_by_date: daily",
		"server_name: mail.example.com",
		"smtp_tls_config:",
		"server_name: smtp.example.com",
		"smtp_force_implicit_tls: true",
		"slack_app_token: xapp-1-token",
		"slack_app_url: https://slack.example.com/api",
		"telegram_bot_token: bot-token",
		"name: cluster-down",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("set attribute did not produce %q in generated configuration:\n%s", want, got)
		}
	}
}

// TestExpandAlertmanagerConfigUsesTimeIntervals pins that generated
// configurations use the current spelling of the top-level time interval list.
func TestExpandAlertmanagerConfigUsesTimeIntervals(t *testing.T) {
	interval := []interface{}{map[string]interface{}{
		"name":           "weekends",
		"time_intervals": []interface{}{},
	}}

	got := marshalYAML(t, alertmanagerConfig{
		TimeIntervals: expandMuteTimeIntervalConfig(interval),
	})

	if !strings.Contains(got, "time_intervals:") {
		t.Errorf("expected time_intervals in generated configuration:\n%s", got)
	}
	if strings.Contains(got, "mute_time_intervals:") {
		t.Errorf("expected no mute_time_intervals in generated configuration:\n%s", got)
	}
}

// TestAllTimeIntervals covers reading a configuration this provider did not
// write: Alertmanager accepts either spelling, so both have to reach the state.
func TestAllTimeIntervals(t *testing.T) {
	deprecated := &muteTimeInterval{Name: "deprecated"}
	current := &muteTimeInterval{Name: "current"}

	cases := []struct {
		name  string
		input alertmanagerConfig
		want  []string
	}{
		{"neither", alertmanagerConfig{}, nil},
		{"deprecated only", alertmanagerConfig{MuteTimeIntervals: []*muteTimeInterval{deprecated}}, []string{"deprecated"}},
		{"current only", alertmanagerConfig{TimeIntervals: []*muteTimeInterval{current}}, []string{"current"}},
		{
			"both",
			alertmanagerConfig{
				MuteTimeIntervals: []*muteTimeInterval{deprecated},
				TimeIntervals:     []*muteTimeInterval{current},
			},
			[]string{"deprecated", "current"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := allTimeIntervals(c.input)
			if len(got) != len(c.want) {
				t.Fatalf("got %d intervals, want %d", len(got), len(c.want))
			}
			for i, name := range c.want {
				if got[i].Name != name {
					t.Errorf("interval %d = %q, want %q", i, got[i].Name, name)
				}
			}
		})
	}

	// Merging must not write through to the slice held by the configuration.
	cfg := alertmanagerConfig{
		MuteTimeIntervals: []*muteTimeInterval{deprecated},
		TimeIntervals:     []*muteTimeInterval{current},
	}
	_ = allTimeIntervals(cfg)
	if len(cfg.MuteTimeIntervals) != 1 {
		t.Errorf("allTimeIntervals mutated MuteTimeIntervals: got %d entries, want 1", len(cfg.MuteTimeIntervals))
	}
}

// TestFlattenDurationMatchesStateFunc pins the round-trip invariant behind the
// pushover ttl acceptance failure: a duration written back to state by Read must
// render exactly as formatDuration renders the configured value, or Terraform
// reports a permanent diff. Rendering a time.Duration with its own String method
// breaks this for any value with trailing zero sub-units or longer than a day.
func TestFlattenDurationMatchesStateFunc(t *testing.T) {
	for _, configured := range []string{"", "10s", "90s", "5m", "1h", "1h30m", "1d", "2w"} {
		parsed, _ := model.ParseDuration(configured)

		want := formatDuration(configured)
		got := flattenDuration(time.Duration(parsed))

		if got != want {
			t.Errorf("configured %q: state written by Read is %q, config normalizes to %q", configured, got, want)
		}
	}
}

// TestFlattenSlackActionOmitsAbsentConfirm covers the confirm sub-block, which
// was flattened unconditionally: a nil pointer still produced a one-element list
// and the plan kept trying to remove a block the user never wrote.
func TestFlattenSlackActionOmitsAbsentConfirm(t *testing.T) {
	got := flattenSlackConfigActions([]slackAction{{
		Type: "button", Text: "Runbook", URL: "https://runbook.example.com",
	}})

	action := got[0].(map[string]interface{})
	if _, present := action["confirm"]; present {
		t.Errorf("confirm block written for an action that declares none: %v", action)
	}

	withConfirm := flattenSlackConfigActions([]slackAction{{
		Type: "button", ConfirmField: &slackConfirmationField{Text: "sure?"},
	}})
	if _, present := withConfirm[0].(map[string]interface{})["confirm"]; !present {
		t.Error("confirm block dropped for an action that declares one")
	}
}
