package mimir

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// routeSchemaOf returns the root route block of a resource or data source.
func routeSchemaOf(s map[string]*schema.Schema) map[string]*schema.Schema {
	return s["route"].Elem.(*schema.Resource).Schema
}

// TestRootRouteDeclaresEveryFlattenedKey pins that the root route block declares
// every key flattenRouteConfig writes. It used to omit matchers,
// mute_time_intervals and active_time_intervals even though both expand and
// flatten handled them, so d.Set failed for any root route using one.
func TestRootRouteDeclaresEveryFlattenedKey(t *testing.T) {
	flattened := flattenRouteConfig(&route{
		Receiver:            "team",
		GroupByStr:          []string{"alertname"},
		Matchers:            []string{"severity=critical"},
		MuteTimeIntervals:   []string{"offhours"},
		ActiveTimeIntervals: []string{"workhours"},
	})
	written := flattened[0].(map[string]interface{})

	for name, s := range map[string]map[string]*schema.Schema{
		"resource":    routeSchemaOf(resourceMimirAlertmanagerConfigSchemaV1()),
		"data source": routeSchemaOf(dataSourceMimirAlertmanagerConfigSchemaV1()),
	} {
		for key := range written {
			if _, declared := s[key]; !declared {
				t.Errorf("%s root route: flattenRouteConfig writes %q but the schema does not declare it; d.Set would fail", name, key)
			}
		}
	}
}

// TestOrgIDIsValidatedEverywhere keeps org_id from reaching the X-Scope-OrgID
// header unchecked: it is a tenant identifier and a storage path component, so a
// '/' or a control character is a cross-tenant primitive.
func TestOrgIDIsValidatedEverywhere(t *testing.T) {
	for name, s := range map[string]map[string]*schema.Schema{
		"alertmanager resource":    resourceMimirAlertmanagerConfigSchemaV1(),
		"alertmanager data source": dataSourceMimirAlertmanagerConfigSchemaV1(),
		"distributor stats source": dataSourcemimirDistributorTenantStats().Schema,
		"rule group alerting":      resourcemimirRuleGroupAlerting().Schema,
		"rule group recording":     resourcemimirRuleGroupRecording().Schema,
		"rules":                    resourceMimirRules().Schema,
	} {
		entry, declared := s[orgIDKey]
		if !declared {
			continue
		}
		if entry.ValidateFunc == nil {
			t.Errorf("%s: org_id has no ValidateFunc; an unchecked tenant reaches X-Scope-OrgID", name)
		}
	}
}
