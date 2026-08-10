package mimir

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gopkg.in/yaml.v3"
)

// TestRuleGroupPreservesRulefmtFields guards the fields of Mimir's vendored
// rulefmt.RuleGroup / rulefmt.Rule that this resource must round-trip. It parses
// the user's YAML into its own structs and marshals them back out before the
// POST, so any field missing from the Go model is silently stripped from the
// rule group that reaches Mimir.
func TestRuleGroupPreservesRulefmtFields(t *testing.T) {
	const content = `groups:
  - name: example
    interval: 1m
    query_offset: 2m
    evaluation_delay: 3m
    align_evaluation_time_on_interval: true
    limit: 10
    source_tenants: [tenant-a]
    labels:
      team: core
    rules:
      - alert: HighErrorRate
        expr: rate(errors[5m]) > 1
        for: 5m
        keep_firing_for: 10m
        labels:
          severity: critical
        annotations:
          summary: too many errors
      - record: job:errors:rate5m
        expr: rate(errors[5m])
`

	var parsed RuleGroups
	if err := yaml.Unmarshal([]byte(content), &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(parsed.Groups) != 1 {
		t.Fatalf("got %d groups, want 1", len(parsed.Groups))
	}

	out, err := yaml.Marshal(parsed.Groups[0])
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(out)

	for _, want := range []string{
		"query_offset: 2m",
		"evaluation_delay: 3m",
		"align_evaluation_time_on_interval: true",
		"keep_firing_for: 10m",
		"limit: 10",
		"source_tenants:",
		"team: core",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("field %q was dropped when re-marshalling the rule group:\n%s", want, got)
		}
	}
}

// TestAlertingDataSourceExposesKeepFiringFor pins that the alerting data source
// declares every attribute the shared flattenAlertingRules writes. A missing
// attribute makes d.Set fail, so reading any group whose rules use it errors out
// instead of returning the group.
func TestAlertingDataSourceExposesKeepFiringFor(t *testing.T) {
	ds := dataSourcemimirRuleGroupAlerting()
	ruleSchema := ds.Schema["rule"].Elem.(*schema.Resource).Schema

	flattened := flattenAlertingRules([]alertingRule{{
		Alert:         "HighErrorRate",
		Expr:          "rate(errors[5m]) > 1",
		KeepFiringFor: "10m",
	}})

	for key := range flattened[0] {
		if _, declared := ruleSchema[key]; !declared {
			t.Errorf("flattenAlertingRules writes %q but the alerting data source does not declare it; d.Set would fail", key)
		}
	}
}
