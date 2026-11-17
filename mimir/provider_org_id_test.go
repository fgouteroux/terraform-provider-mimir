package mimir

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// The provider's org_id becomes the X-Scope-OrgID of every request that has no
// resource-level org_id, so it follows the same tenant ID rules.
func TestProviderOrgIDValidation(t *testing.T) {
	for _, tc := range []struct {
		orgID   string
		wantErr bool
	}{
		{"tenant1", false},
		{"tenant-a|tenant_b", false},
		{"my tenant", true},
		{"..", true},
		{"tenant/a", true},
		{strings.Repeat("a", maxTenantIDLength+1), true},
	} {
		t.Run(tc.orgID, func(t *testing.T) {
			p := Provider("test")()
			cfg := terraform.NewResourceConfigRaw(map[string]interface{}{
				"uri":    "http://localhost:8080",
				"org_id": tc.orgID,
			})

			if diags := p.Validate(cfg); diags.HasError() != tc.wantErr {
				t.Errorf("Validate(%q) = %v, wantErr %v", tc.orgID, diags, tc.wantErr)
			}
		})
	}
}

// A value set through MIMIR_ORG_ID is checked as well.
func TestProviderOrgIDFromEnvValidation(t *testing.T) {
	t.Setenv("MIMIR_ORG_ID", "my tenant")
	p := Provider("test")()
	cfg := terraform.NewResourceConfigRaw(map[string]interface{}{"uri": "http://localhost:8080"})

	diags := p.Validate(cfg)
	if !diags.HasError() {
		t.Fatal("Validate accepted an invalid MIMIR_ORG_ID")
	}
	if got := diags[0].Summary; !strings.Contains(got, "my tenant") {
		t.Errorf("error does not name the value: %q", got)
	}
}
