package mimir

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestValidRuleName(t *testing.T) {
	accept := []string{"LUN Destroyed", "Harvest Rules", `alert{env="prod"}`, "50% full", "café", "..dotstart", "a.b", "test1_alert"}
	reject := []string{"", "   ", "a\x00b", "a\nb", "a/b", ".", ".."}

	for _, s := range accept {
		if !validRuleName(s) {
			t.Errorf("validRuleName(%q) = false, want true", s)
		}
		if _, errs := validateGroupRuleName(s, "name"); len(errs) != 0 {
			t.Errorf("validateGroupRuleName(%q) rejected an accept-case: %v", s, errs)
		}
		if _, errs := validateAlertingRuleName(s, "alert"); len(errs) != 0 {
			t.Errorf("validateAlertingRuleName(%q) rejected an accept-case: %v", s, errs)
		}
	}
	for _, s := range reject {
		if validRuleName(s) {
			t.Errorf("validRuleName(%q) = true, want false", s)
		}
		if _, errs := validateGroupRuleName(s, "name"); len(errs) == 0 {
			t.Errorf("validateGroupRuleName(%q) accepted a reject-case", s)
		}
		if _, errs := validateAlertingRuleName(s, "alert"); len(errs) == 0 {
			t.Errorf("validateAlertingRuleName(%q) accepted a reject-case", s)
		}
	}
}

func TestValidateNamespace(t *testing.T) {
	accept := []string{defaultNamespace, "team a", "ns.1", "", "my-namespace"}
	reject := []string{"a/b", "a\x00b", ".", ".."}

	for _, s := range accept {
		ws, errs := validateNamespace(s, "namespace")
		if len(errs) != 0 {
			t.Errorf("validateNamespace(%q) rejected an accept-case: %v", s, errs)
		}
		if len(ws) != 0 {
			t.Errorf("validateNamespace(%q) warned on a plain value: %v", s, ws)
		}
	}
	for _, s := range reject {
		if _, errs := validateNamespace(s, "namespace"); len(errs) == 0 {
			t.Errorf("validateNamespace(%q) accepted a reject-case", s)
		}
	}

	// A valid percent-encoding sequence is ACCEPTED but warned about (older provider
	// versions let the server decode it once; it is now sent literally).
	ws, errs := validateNamespace("team%20a", "namespace")
	if len(errs) != 0 {
		t.Errorf("validateNamespace(team%%20a) rejected: %v", errs)
	}
	if len(ws) != 1 {
		t.Errorf("validateNamespace(team%%20a) warnings = %v, want exactly 1", ws)
	}
	// An invalid/bare %% is not an encoding sequence: accepted, no warning.
	ws, errs = validateNamespace("50% full", "namespace")
	if len(errs) != 0 || len(ws) != 0 {
		t.Errorf("validateNamespace(50%% full) = (%v, %v), want accepted with no warning", ws, errs)
	}
}

func TestValidateOrgID(t *testing.T) {
	accept := []string{"", "myorg", "tenant-1", "org.1"}
	reject := []string{"a\nb", "a\rb", "a\x00b", "a/b", ".", "..", "../other-tenant"}

	for _, s := range accept {
		if _, errs := validateOrgID(s, "org_id"); len(errs) != 0 {
			t.Errorf("validateOrgID(%q) rejected an accept-case: %v", s, errs)
		}
	}
	for _, s := range reject {
		if _, errs := validateOrgID(s, "org_id"); len(errs) == 0 {
			t.Errorf("validateOrgID(%q) accepted a reject-case", s)
		}
	}
}

func groupWith(name, alert string) RuleGroups {
	return RuleGroups{Groups: []RuleGroup{{Name: name, Rules: []Rule{{Alert: alert, Expr: "up"}}}}}
}

func TestValidateRuleGroupsContent(t *testing.T) {
	if err := validateRuleGroupsContent(groupWith("Harvest Rules", "LUN Destroyed")); err != nil {
		t.Errorf("expected accept for relaxed group + alert names, got: %v", err)
	}
	if err := validateRuleGroupsContent(groupWith("", "A")); err == nil {
		t.Error("expected reject for empty group name")
	}
	if err := validateRuleGroupsContent(groupWith("a/b", "A")); err == nil {
		t.Error("expected reject for '/' in group name")
	}
	if err := validateRuleGroupsContent(groupWith("g", "a/b")); err == nil {
		t.Error("expected reject for '/' in alert name")
	}
}

func TestValidateRule(t *testing.T) {
	if err := validateRule(Rule{Alert: "LUN Destroyed", Expr: "up"}, 0, 0, "g"); err != nil {
		t.Errorf("expected accept for relaxed alert name, got: %v", err)
	}
	if err := validateRule(Rule{Alert: "a/b", Expr: "up"}, 0, 0, "g"); err == nil {
		t.Error("expected reject for '/' in alert name")
	}
	// record name stays strict (metricNameRegexp): a space is invalid
	if err := validateRule(Rule{Record: "bad record", Expr: "up"}, 0, 0, "g"); err == nil {
		t.Error("expected record name to stay strict (reject space)")
	}
	if err := validateRule(Rule{Record: "valid_record", Expr: "up"}, 0, 0, "g"); err != nil {
		t.Errorf("expected valid record name accepted, got: %v", err)
	}
}

func TestParseRuleGroupID(t *testing.T) {
	// round-trip an ordinary 2-segment and 3-segment id
	cases := []struct{ org, ns, name string }{
		{"", defaultNamespace, "my group"},
		{"tenant-1", "ns a", "grp%x"},
	}
	for _, c := range cases {
		id := buildRuleGroupID(c.org, c.ns, c.name)
		org, ns, name, err := parseRuleGroupID(id)
		if err != nil {
			t.Errorf("parseRuleGroupID(%q) unexpected error: %v", id, err)
			continue
		}
		if org != c.org || ns != c.ns || name != c.name {
			t.Errorf("round-trip mismatch for %q: got (%q,%q,%q) want (%q,%q,%q)", id, org, ns, name, c.org, c.ns, c.name)
		}
	}
	// a raw '/'-in-namespace can only be expressed escaped, which decodes and is rejected
	if _, _, _, err := parseRuleGroupID("a%2Fb/c"); err == nil {
		t.Error("expected reject for a namespace containing '/' (escaped) in the id")
	}
	// a bad percent-encoding (legacy unescaped ID) falls back to the raw segment, not a hard fail
	if _, _, name, err := parseRuleGroupID("ns/50%zz"); err != nil || name != "50%zz" {
		t.Errorf("expected legacy raw fallback for bad encoding: name=%q err=%v", name, err)
	}
	// dot-segment name rejected on the parse path
	if _, _, _, err := parseRuleGroupID("ns/.."); err == nil {
		t.Error("expected reject for a '..' name in the id")
	}
	// control char in the org_id segment rejected (covers the validOrgID call site inside parse)
	if _, _, _, err := parseRuleGroupID("bad\rorg/ns/name"); err == nil {
		t.Error("expected reject for a control char in the org_id segment")
	}
	// a '/' smuggled into the org_id segment as %2F is rejected (cross-tenant traversal vector)
	if _, _, _, err := parseRuleGroupID("tenantA%2Fevil/ns/rule"); err == nil {
		t.Error("expected reject for a '/' (as %2F) in the org_id segment")
	}
	// a ".." org_id segment is rejected
	if _, _, _, err := parseRuleGroupID("../ns/rule"); err == nil {
		t.Error("expected reject for a '..' org_id segment")
	}
	// invalid segment counts are rejected (pins the default branch against a take-first-3 refactor)
	if _, _, _, err := parseRuleGroupID("noslash"); err == nil {
		t.Error("expected reject for a 1-segment id")
	}
	if _, _, _, err := parseRuleGroupID("a/b/c/d"); err == nil {
		t.Error("expected reject for a 4-segment id")
	}
	// a Unicode C1 control (NEL) in a name is rejected (covers the unicode.IsControl path)
	if validRuleName("a\u0085b") {
		t.Error("expected reject for a NEL (U+0085) control char in a name")
	}
}

// TestResourceMimirRulesImport pins the importer's re-validation guards directly (no server
// needed: the importer never touches its meta parameter). Without this, deleting the
// validOrgID/validNamespace blocks passes the entire suite (verified by mutation).
func TestResourceMimirRulesImport(t *testing.T) {
	res := resourceMimirRules()
	cases := []struct {
		id      string
		wantErr bool
		org, ns string
	}{
		{"tenant/ns", false, "tenant", "ns"},
		{"org/team a", false, "org", "team a"}, // spaces in a namespace are accepted
		{"org/..", true, "", ""},               // dot-segment namespace
		{"bad\rorg/ns", true, "", ""},          // control char in org_id
		{"a/b/c", true, "", ""},                // wrong segment count
		{"noslash", true, "", ""},              // wrong segment count
	}
	for _, c := range cases {
		d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{})
		d.SetId(c.id)
		got, err := resourceMimirRulesImport(context.Background(), d, nil)
		if (err != nil) != c.wantErr {
			t.Errorf("resourceMimirRulesImport(%q) err=%v, wantErr=%v", c.id, err, c.wantErr)
			continue
		}
		if err == nil {
			if len(got) != 1 || got[0].Get(orgIDKey).(string) != c.org || got[0].Get(namespaceKey).(string) != c.ns {
				t.Errorf("resourceMimirRulesImport(%q) set org_id/namespace = %q/%q, want %q/%q",
					c.id, got[0].Get(orgIDKey), got[0].Get(namespaceKey), c.org, c.ns)
			}
		}
	}
}

// TestGuardValidateFuncWiring pins the ATTACHMENT of the plan-time guards to every schema
// surface, not just the validator functions: a refactor dropping a `ValidateFunc:` line
// otherwise keeps the whole suite green (verified by mutation).
func TestGuardValidateFuncWiring(t *testing.T) {
	p := Provider("dev")()
	surfaces := []struct {
		kind, name string
		s          *schema.Resource
	}{
		{"resource", "mimir_rule_group_alerting", p.ResourcesMap["mimir_rule_group_alerting"]},
		{"resource", "mimir_rule_group_recording", p.ResourcesMap["mimir_rule_group_recording"]},
		{"resource", "mimir_rules", p.ResourcesMap["mimir_rules"]},
		{"data", "mimir_rule_group_alerting", p.DataSourcesMap["mimir_rule_group_alerting"]},
		{"data", "mimir_rule_group_recording", p.DataSourcesMap["mimir_rule_group_recording"]},
	}
	for _, sf := range surfaces {
		if sf.s == nil {
			t.Fatalf("%s %q not found in provider map", sf.kind, sf.name)
		}
		for _, field := range []string{namespaceKey, orgIDKey} {
			sch, ok := sf.s.Schema[field]
			if !ok {
				t.Errorf("%s %q: field %q missing from schema", sf.kind, sf.name, field)
				continue
			}
			if sch.ValidateFunc == nil {
				t.Errorf("%s %q: field %q has no ValidateFunc (plan-time guard dropped)", sf.kind, sf.name, field)
				continue
			}
			if _, errs := sch.ValidateFunc("a/b", field); len(errs) == 0 {
				t.Errorf("%s %q: field %q ValidateFunc accepted \"a/b\"", sf.kind, sf.name, field)
			}
			if _, errs := sch.ValidateFunc("valid-value", field); len(errs) != 0 {
				t.Errorf("%s %q: field %q ValidateFunc rejected a valid value: %v", sf.kind, sf.name, field, errs)
			}
		}
	}
}
