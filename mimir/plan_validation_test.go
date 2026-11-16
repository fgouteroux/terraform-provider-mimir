package mimir

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestValidateRuleGroupsContentMimirChecks(t *testing.T) {
	cases := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name:    "query_offset and evaluation_delay differ",
			content: "groups:\n- name: g\n  query_offset: 1m\n  evaluation_delay: 2m\n  rules:\n  - record: r\n    expr: up\n",
			wantErr: "evaluation_delay are set to different values",
		},
		{
			name:    "invalid query_offset",
			content: "groups:\n- name: g\n  query_offset: 1x\n  rules:\n  - record: r\n    expr: up\n",
			wantErr: "invalid query_offset",
		},
		{
			name:    "negative limit",
			content: "groups:\n- name: g\n  limit: -1\n  rules:\n  - record: r\n    expr: up\n",
			wantErr: "limit must not be negative",
		},
		{
			name:    "invalid group label name",
			content: "groups:\n- name: g\n  labels:\n    bad-name: x\n  rules:\n  - record: r\n    expr: up\n",
			wantErr: "invalid group label name",
		},
		{
			name:    "keep_firing_for on a recording rule",
			content: "groups:\n- name: g\n  rules:\n  - record: r\n    expr: up\n    keep_firing_for: 5m\n",
			wantErr: "cannot have 'keep_firing_for'",
		},
		{
			name:    "invalid PromQL",
			content: "groups:\n- name: g\n  rules:\n  - alert: A\n    expr: rate(up[5m]\n",
			wantErr: "could not parse expression",
		},
		{
			name:    "reserved __name__ label",
			content: "groups:\n- name: g\n  rules:\n  - alert: A\n    expr: up\n    labels:\n      __name__: x\n",
			wantErr: "invalid label name: __name__",
		},
		{
			name:    "broken annotation template",
			content: "groups:\n- name: g\n  rules:\n  - alert: A\n    expr: up\n    annotations:\n      summary: '{{ $labels.instance '\n",
			wantErr: "template",
		},
		{
			name:    "valid group",
			content: "groups:\n- name: g\n  query_offset: 1m\n  evaluation_delay: 1m\n  limit: 5\n  rules:\n  - alert: A\n    expr: up == 0\n    for: 5m\n    keep_firing_for: 1m\n    annotations:\n      summary: '{{ $labels.instance }} down'\n",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, errs := validateYAMLContent(c.content, "content")
			if c.wantErr == "" {
				if len(errs) > 0 {
					t.Fatalf("unexpected errors: %v", errs)
				}
				return
			}
			if len(errs) == 0 || !strings.Contains(errs[0].Error(), c.wantErr) {
				t.Fatalf("want an error containing %q, got %v", c.wantErr, errs)
			}
		})
	}
}

func TestValidateYAMLContentWarnsOnPartialResponse(t *testing.T) {
	warns, errs := validateYAMLContent("groups:\n- name: g\n  partial_response_strategy: true\n  rules:\n  - record: r\n    expr: up\n", "content")
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(warns) != 1 || !strings.Contains(warns[0], "partial_response_strategy") {
		t.Fatalf("want one partial_response_strategy warning, got %v", warns)
	}
}

func TestValidateAlertmanagerMatchers(t *testing.T) {
	for value, wantErr := range map[string]bool{
		`severity="critical"`:        false,
		`severity=~"crit.*"`:         false,
		`{team="infra", env!="dev"}`: false,
		`service="fóo"`:              false,
		`severity=`:                  false,
		`severity`:                   true,
		`severity=~"("`:              true,
	} {
		if _, errs := validateAlertmanagerMatchers(value, "matchers"); (len(errs) > 0) != wantErr {
			t.Errorf("validateAlertmanagerMatchers(%q) errors = %v, wantErr %v", value, errs, wantErr)
		}
	}
}

func TestValidateTemplateNames(t *testing.T) {
	for name, wantErr := range map[string]bool{
		"default.tmpl":   false,
		"*.tmpl":         false,
		"a/b.tmpl":       true,
		"/etc/am/x.tmpl": true,
		"../escape.tmpl": true,
		"./default.tmpl": true,
	} {
		if _, errs := validateTemplateName(name, "templates"); (len(errs) > 0) != wantErr {
			t.Errorf("validateTemplateName(%q) errors = %v, wantErr %v", name, errs, wantErr)
		}
	}

	if _, errs := validateTemplateFiles(map[string]interface{}{"ok.tmpl": "x", "dir/bad.tmpl": "y"}, "templates_files"); len(errs) != 1 {
		t.Errorf("validateTemplateFiles: want exactly one error, got %v", errs)
	}
}

func TestValidateOrgIDMimirTenantRules(t *testing.T) {
	accept := []string{"tenant-1", "org_(1)", "a.b", "team-a|team-b", "Tenant!*'"}
	reject := []string{"my org", "tenant@corp", "café", "a||b", "a|", strings.Repeat("a", maxTenantIDLength+1)}
	for _, s := range accept {
		if _, errs := validateOrgID(s, "org_id"); len(errs) != 0 {
			t.Errorf("validateOrgID(%q) rejected a valid tenant: %v", s, errs)
		}
	}
	for _, s := range reject {
		if _, errs := validateOrgID(s, "org_id"); len(errs) == 0 {
			t.Errorf("validateOrgID(%q) accepted an invalid tenant", s)
		}
	}
}

func TestDataSourceTenantLimitsRead(t *testing.T) {
	var gotOrgID string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != userLimitsPath {
			http.NotFound(w, r)
			return
		}
		gotOrgID = r.Header.Get("X-Scope-OrgID")
		// A 3.1-shaped response: no alertmanager_* fields.
		_, _ = w.Write([]byte(`{"compactor_blocks_retention_period_seconds":0,"ingestion_rate":10000,"ingestion_burst_size":200000,"ingestion_burst_factor":0,"max_global_series_per_user":150000,"max_global_series_per_metric":0,"max_global_exemplars_per_user":0,"max_fetched_chunks_per_query":2000000,"max_fetched_series_per_query":0,"max_fetched_chunk_bytes_per_query":0,"ruler_max_rules_per_rule_group":20,"ruler_max_rule_groups_per_tenant":70}`))
	}))
	defer srv.Close()

	client, err := NewAPIClient(&apiClientOpt{uri: srv.URL, headers: map[string]string{"X-Scope-OrgID": "default"}, timeout: 2})
	if err != nil {
		t.Fatalf("NewAPIClient: %v", err)
	}

	res := dataSourcemimirTenantLimits()
	d := schema.TestResourceDataRaw(t, res.Schema, map[string]interface{}{orgIDKey: "mytenant"})
	if diags := res.ReadContext(t.Context(), d, client); diags.HasError() {
		t.Fatalf("read: %v", diags)
	}

	if gotOrgID != "mytenant" {
		t.Errorf("X-Scope-OrgID = %q, want mytenant", gotOrgID)
	}
	if d.Id() != "mytenant" {
		t.Errorf("id = %q, want mytenant", d.Id())
	}
	for attr, want := range map[string]interface{}{
		"ingestion_rate":                   10000.0,
		"max_global_series_per_user":       150000,
		"ruler_max_rule_groups_per_tenant": 70,
		"alertmanager_max_templates_count": 0,
	} {
		if got := d.Get(attr); got != want {
			t.Errorf("%s = %v, want %v", attr, got, want)
		}
	}
}

func TestAccDataSourceTenantLimits(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "mimir_tenant_limits" "this" {}`,
				Check: resource.ComposeTestCheckFunc(
					// Mimir's defaults: -distributor.ingestion-rate-limit and
					// -ingester.max-global-series-per-user.
					resource.TestCheckResourceAttr("data.mimir_tenant_limits.this", "ingestion_rate", "10000"),
					resource.TestCheckResourceAttr("data.mimir_tenant_limits.this", "max_global_series_per_user", "150000"),
				),
			},
		},
	})
}

func TestAccResourceRuleGroupAlerting_MimirRuleChecks(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
    resource "mimir_rule_group_alerting" "template_error" {
        name      = "template_error"
        namespace = "namespace_1"
        rule {
            alert = "A"
            expr  = "up == 0"
            annotations = {
                summary = "{{ $labels.instance "
            }
        }
    }`,
				ExpectError: regexp.MustCompile(`template`),
			},
			{
				Config: `
    resource "mimir_rule_group_recording" "name_label" {
        name      = "name_label"
        namespace = "namespace_1"
        rule {
            record = "r"
            expr   = "up"
            labels = {
                __name__ = "x"
            }
        }
    }`,
				ExpectError: regexp.MustCompile(`invalid label name: __name__`),
			},
		},
	})
}
