package mimir

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// routeSchemaOf returns the root route block of a resource or data source.
func routeSchemaOf(s map[string]*schema.Schema) map[string]*schema.Schema {
	return s["route"].Elem.(*schema.Resource).Schema
}

// TestRootRouteRejectsChildOnlyAttributes pins that the root route block does
// NOT declare matchers, mute_time_intervals or active_time_intervals. Alertmanager
// refuses all three on the root route ("root route must not have any matchers",
// config.go), so the server can never return them and declaring them as Optional
// would let a configuration pass plan and then fail at apply with an opaque 400.
// They belong to child_route only.
func TestRootRouteRejectsChildOnlyAttributes(t *testing.T) {
	for name, s := range map[string]map[string]*schema.Schema{
		"resource":    routeSchemaOf(resourceMimirAlertmanagerConfigSchemaV1()),
		"data source": routeSchemaOf(dataSourceMimirAlertmanagerConfigSchemaV1()),
	} {
		for _, attr := range []string{"matchers", "mute_time_intervals", "active_time_intervals"} {
			if _, declared := s[attr]; declared {
				t.Errorf("%s root route declares %q, which Alertmanager rejects on the root route", name, attr)
			}
		}
	}

	// The same attributes must stay available on child routes.
	child := childRouteSchema(1)
	for _, attr := range []string{"matchers", "mute_time_intervals", "active_time_intervals"} {
		if _, declared := child[attr]; !declared {
			t.Errorf("child route no longer declares %q", attr)
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
