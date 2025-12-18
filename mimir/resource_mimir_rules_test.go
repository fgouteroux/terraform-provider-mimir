package mimir

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestValidateRuleGroupsContent_AllowsPrometheusDurations(t *testing.T) {
	ruleGroups := RuleGroups{
		Groups: []RuleGroup{
			{
				Name:     "alert_group",
				Interval: "1m",
				Rules: []Rule{
					{
						Alert: "HighErrorRate",
						Expr:  `rate(http_requests_total{status=~"5.."}[5m]) > 0`,
						For:   "1d",
					},
				},
			},
		},
	}

	if err := validateRuleGroupsContent(ruleGroups); err != nil {
		t.Fatalf("expected Prometheus-style durations to be valid, got: %v", err)
	}
}

func TestAccResourceMimirRules_Basic(t *testing.T) {
	// Init client
	client, err := NewAPIClient(setupClient())
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMimirRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceMimirRules_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirNamespaceExists("mimir_rules.rules_1", "alert_group_1", client),
					testAccCheckMimirNamespaceExists("mimir_rules.rules_1", "record_group_1", client),
					resource.TestCheckResourceAttr("mimir_rules.rules_1", "namespace", "namespace_1"),
					resource.TestCheckResourceAttr("mimir_rules.rules_1", "groups_count", "2"),
					resource.TestCheckResourceAttr("mimir_rules.rules_1", "total_rules", "3"),
					resource.TestCheckResourceAttr("mimir_rules.rules_1", "managed_groups.0", "alert_group_1"),
					resource.TestCheckResourceAttr("mimir_rules.rules_1", "managed_groups.1", "record_group_1"),
					resource.TestCheckResourceAttr("mimir_rules.rules_1", "rule_names.0", "HighCPUUsage"),
					resource.TestCheckResourceAttr("mimir_rules.rules_1", "rule_names.1", "HighMemoryUsage"),
					resource.TestCheckResourceAttr("mimir_rules.rules_1", "rule_names.2", "instance:cpu:rate5m"),
				),
			},
			{
				Config: testAccResourceMimirRules_basic_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirNamespaceExists("mimir_rules.rules_1", "alert_group_1", client),
					testAccCheckMimirNamespaceExists("mimir_rules.rules_1", "record_group_1", client),
					resource.TestCheckResourceAttr("mimir_rules.rules_1", "namespace", "namespace_1"),
					resource.TestCheckResourceAttr("mimir_rules.rules_1", "groups_count", "2"),
					resource.TestCheckResourceAttr("mimir_rules.rules_1", "total_rules", "4"),
					resource.TestCheckResourceAttr("mimir_rules.rules_1", "rule_names.2", "LowDiskSpace"),
				),
			},
		},
	})
}

func TestAccResourceMimirRules_ContentFile(t *testing.T) {
	// Init client
	client, err := NewAPIClient(setupClient())
	if err != nil {
		t.Fatal(err)
	}

	// Create temporary YAML file
	tmpfile, err := os.CreateTemp("", "rules-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	content := `groups:
  - name: file_based_alerts
    interval: 1m
    rules:
      - alert: TestAlert
        expr: up == 0
        for: 5m
        labels:
          severity: warning
`
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMimirRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceMimirRules_contentFile(tmpfile.Name()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirNamespaceExists("mimir_rules.rules_file", "file_based_alerts", client),
					resource.TestCheckResourceAttr("mimir_rules.rules_file", "namespace", "namespace_1"),
					resource.TestCheckResourceAttr("mimir_rules.rules_file", "groups_count", "1"),
					resource.TestCheckResourceAttr("mimir_rules.rules_file", "total_rules", "1"),
					resource.TestCheckResourceAttr("mimir_rules.rules_file", "rule_names.0", "TestAlert"),
				),
			},
		},
	})
}

func TestAccResourceMimirRules_OnlyGroups(t *testing.T) {
	// Init client
	client, err := NewAPIClient(setupClient())
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMimirRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceMimirRules_onlyGroups,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirNamespaceExists("mimir_rules.rules_filtered", "alert_group_1", client),
					resource.TestCheckResourceAttr("mimir_rules.rules_filtered", "namespace", "namespace_1"),
					resource.TestCheckResourceAttr("mimir_rules.rules_filtered", "groups_count", "1"),
					resource.TestCheckResourceAttr("mimir_rules.rules_filtered", "total_rules", "2"),
					resource.TestCheckResourceAttr("mimir_rules.rules_filtered", "managed_groups.0", "alert_group_1"),
				),
			},
		},
	})
}

func TestAccResourceMimirRules_IgnoreGroups(t *testing.T) {
	// Init client
	client, err := NewAPIClient(setupClient())
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMimirRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceMimirRules_ignoreGroups,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirNamespaceExists("mimir_rules.rules_ignored", "alert_group_1", client),
					resource.TestCheckResourceAttr("mimir_rules.rules_ignored", "namespace", "namespace_1"),
					resource.TestCheckResourceAttr("mimir_rules.rules_ignored", "groups_count", "1"),
					resource.TestCheckResourceAttr("mimir_rules.rules_ignored", "managed_groups.0", "alert_group_1"),
				),
			},
		},
	})
}

func TestAccResourceMimirRules_WithOrgID(t *testing.T) {
	// Init client
	client, err := NewAPIClient(setupClient())
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMimirRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceMimirRules_withOrgID,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirNamespaceExists("mimir_rules.rules_with_org", "alert_group_1", client),
					resource.TestCheckResourceAttr("mimir_rules.rules_with_org", "org_id", "another_tenant"),
					resource.TestCheckResourceAttr("mimir_rules.rules_with_org", "namespace", "namespace_1"),
					resource.TestCheckResourceAttr("mimir_rules.rules_with_org", "groups_count", "1"),
				),
			},
		},
	})
}

func TestAccResourceMimirRules_Federated(t *testing.T) {
	// Init client
	client, err := NewAPIClient(setupClient())
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMimirRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceMimirRules_federated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirNamespaceExists("mimir_rules.rules_federated", "federated_group", client),
					resource.TestCheckResourceAttr("mimir_rules.rules_federated", "namespace", "namespace_1"),
					resource.TestCheckResourceAttr("mimir_rules.rules_federated", "groups_count", "1"),
				),
			},
		},
	})
}

func TestAccResourceMimirRules_Update_AddGroup(t *testing.T) {
	// Init client
	client, err := NewAPIClient(setupClient())
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMimirRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceMimirRules_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mimir_rules.rules_1", "groups_count", "2"),
				),
			},
			{
				Config: testAccResourceMimirRules_addGroup,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirNamespaceExists("mimir_rules.rules_1", "alert_group_1", client),
					testAccCheckMimirNamespaceExists("mimir_rules.rules_1", "record_group_1", client),
					testAccCheckMimirNamespaceExists("mimir_rules.rules_1", "new_group", client),
					resource.TestCheckResourceAttr("mimir_rules.rules_1", "groups_count", "3"),
					resource.TestCheckResourceAttr("mimir_rules.rules_1", "managed_groups.2", "new_group"),
				),
			},
		},
	})
}

func TestAccResourceMimirRules_Update_RemoveGroup(t *testing.T) {
	// Init client
	client, err := NewAPIClient(setupClient())
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMimirRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceMimirRules_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mimir_rules.rules_1", "groups_count", "2"),
				),
			},
			{
				Config: testAccResourceMimirRules_removeGroup,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirNamespaceExists("mimir_rules.rules_1", "alert_group_1", client),
					resource.TestCheckResourceAttr("mimir_rules.rules_1", "groups_count", "1"),
					resource.TestCheckResourceAttr("mimir_rules.rules_1", "managed_groups.0", "alert_group_1"),
				),
			},
		},
	})
}

// Test configurations

const testAccResourceMimirRules_basic = `
resource "mimir_rules" "rules_1" {
  namespace = "namespace_1"
  
  content = <<-EOT
groups:
  - name: alert_group_1
    interval: 30s
    rules:
      - alert: HighCPUUsage
        expr: cpu_usage > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High CPU usage detected"
      - alert: HighMemoryUsage
        expr: memory_usage > 90
        for: 3m
        labels:
          severity: critical
        annotations:
          summary: "High memory usage detected"
  
  - name: record_group_1
    interval: 1m
    rules:
      - record: instance:cpu:rate5m
        expr: rate(cpu_total[5m])
        labels:
          job: monitoring
EOT
}
`

const testAccResourceMimirRules_basic_update = `
resource "mimir_rules" "rules_1" {
  namespace = "namespace_1"
  
  content = <<-EOT
groups:
  - name: alert_group_1
    interval: 30s
    rules:
      - alert: HighCPUUsage
        expr: cpu_usage > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High CPU usage detected"
      - alert: HighMemoryUsage
        expr: memory_usage > 90
        for: 3m
        labels:
          severity: critical
        annotations:
          summary: "High memory usage detected"
      - alert: LowDiskSpace
        expr: disk_free < 10
        for: 10m
        labels:
          severity: critical
        annotations:
          summary: "Low disk space"
  
  - name: record_group_1
    interval: 1m
    rules:
      - record: instance:cpu:rate5m
        expr: rate(cpu_total[5m])
        labels:
          job: monitoring
EOT
}
`

func testAccResourceMimirRules_contentFile(filepath string) string {
	return fmt.Sprintf(`
resource "mimir_rules" "rules_file" {
  namespace    = "namespace_1"
  content_file = "%s"
}
`, filepath)
}

const testAccResourceMimirRules_onlyGroups = `
resource "mimir_rules" "rules_filtered" {
  namespace = "namespace_1"
  
  only_groups = ["alert_group_1"]
  
  content = <<-EOT
groups:
  - name: alert_group_1
    rules:
      - alert: HighCPUUsage
        expr: cpu_usage > 80
      - alert: HighMemoryUsage
        expr: memory_usage > 90
  
  - name: record_group_1
    rules:
      - record: instance:cpu:rate5m
        expr: rate(cpu_total[5m])
EOT
}
`

const testAccResourceMimirRules_ignoreGroups = `
resource "mimir_rules" "rules_ignored" {
  namespace = "namespace_1"
  
  ignore_groups = ["record_group_1"]
  
  content = <<-EOT
groups:
  - name: alert_group_1
    rules:
      - alert: HighCPUUsage
        expr: cpu_usage > 80
  
  - name: record_group_1
    rules:
      - record: instance:cpu:rate5m
        expr: rate(cpu_total[5m])
EOT
}
`

const testAccResourceMimirRules_withOrgID = `
resource "mimir_rules" "rules_with_org" {
  org_id    = "another_tenant"
  namespace = "namespace_1"
  
  content = <<-EOT
groups:
  - name: alert_group_1
    rules:
      - alert: TestAlert
        expr: up == 0
EOT
}
`

const testAccResourceMimirRules_federated = `
resource "mimir_rules" "rules_federated" {
  namespace = "namespace_1"
  
  content = <<-EOT
groups:
  - name: federated_group
    source_tenants: ["tenant-a", "tenant-b"]
    rules:
      - alert: CrossTenantAlert
        expr: sum(metric) > 100
EOT
}
`

const testAccResourceMimirRules_addGroup = `
resource "mimir_rules" "rules_1" {
  namespace = "namespace_1"
  
  content = <<-EOT
groups:
  - name: alert_group_1
    interval: 30s
    rules:
      - alert: HighCPUUsage
        expr: cpu_usage > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High CPU usage detected"
      - alert: HighMemoryUsage
        expr: memory_usage > 90
        for: 3m
        labels:
          severity: critical
        annotations:
          summary: "High memory usage detected"
  
  - name: record_group_1
    interval: 1m
    rules:
      - record: instance:cpu:rate5m
        expr: rate(cpu_total[5m])
        labels:
          job: monitoring
  
  - name: new_group
    rules:
      - alert: NewAlert
        expr: new_metric > 50
EOT
}
`

const testAccResourceMimirRules_removeGroup = `
resource "mimir_rules" "rules_1" {
  namespace = "namespace_1"
  
  content = <<-EOT
groups:
  - name: alert_group_1
    interval: 30s
    rules:
      - alert: HighCPUUsage
        expr: cpu_usage > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High CPU usage detected"
      - alert: HighMemoryUsage
        expr: memory_usage > 90
        for: 3m
        labels:
          severity: critical
        annotations:
          summary: "High memory usage detected"
EOT
}
`
