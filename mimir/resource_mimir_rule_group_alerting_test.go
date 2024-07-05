package mimir

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceRuleGroupAlerting_expectValidationError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceRuleGroupAlerting_expectNameValidationError,
				ExpectError: regexp.MustCompile("Invalid Group Rule Name"),
			},
			{
				Config:      testAccResourceRuleGroupAlerting_expectRuleNameValidationError,
				ExpectError: regexp.MustCompile("Invalid Alerting Rule Name"),
			},
			{
				Config:      testAccResourceRuleGroupAlerting_expectPromQLValidationError,
				ExpectError: regexp.MustCompile("Invalid PromQL expression"),
			},
			{
				Config:      testAccResourceRuleGroupAlerting_expectDurationValidationError,
				ExpectError: regexp.MustCompile("unknown unit"),
			},
			{
				Config:      testAccResourceRuleGroupAlerting_expectLabelNameValidationError,
				ExpectError: regexp.MustCompile("Invalid Label Name"),
			},
			{
				Config:      testAccResourceRuleGroupAlerting_expectAnnotationNameValidationError,
				ExpectError: regexp.MustCompile("Invalid Annotation Name"),
			},
		},
	})
}

const testAccResourceRuleGroupAlerting_expectNameValidationError = `
	resource "mimir_rule_group_alerting" "alert_1" {
		name = "alert-@error" 
		namespace = "namespace_1"
		rule {
			alert = "test1_alert"
			expr   = "test1_metric"
		}
	}
`

const testAccResourceRuleGroupAlerting_expectRuleNameValidationError = `
	resource "mimir_rule_group_alerting" "alert_1" {
		name = "alert_1"
		namespace = "namespace_1"
		rule {
			alert = "test1 alert"
			expr   = "test1_metric"
		}
	}
`

const testAccResourceRuleGroupAlerting_expectPromQLValidationError = `
	resource "mimir_rule_group_alerting" "alert_1" {
		name = "alert_1"
		namespace = "namespace_1"
		rule {
			alert = "test1_alert"
			expr   = "rate(hi)"
		}
	}
`

const testAccResourceRuleGroupAlerting_expectDurationValidationError = `
	resource "mimir_rule_group_alerting" "alert_1" {
		name = "alert_1"
		namespace = "namespace_1"
		rule {
			alert = "test1_alert"
			expr  = "test1_metric"
			for   = "3months"
		}
	}
`

const testAccResourceRuleGroupAlerting_expectLabelNameValidationError = `
	resource "mimir_rule_group_alerting" "alert_1" {
		name = "alert_1"
		namespace = "namespace_1"
		rule {
			alert = "test1_alert"
			expr   = "test1_metric"
			labels = {
				 ins-tance = "localhost"
			}
		}
	}
`

const testAccResourceRuleGroupAlerting_expectAnnotationNameValidationError = `
	resource "mimir_rule_group_alerting" "alert_1" {
		name = "alert_1"
		namespace = "namespace_1"
		rule {
			alert = "test1_alert"
			expr   = "test1_metric"
			annotations = {
				 ins-tance = "localhost"
			}
		}
	}
`

func TestAccResourceRuleGroupAlerting_Basic(t *testing.T) {
	// Init client
	client, err := NewAPIClient(setupClient())
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMimirRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRuleGroupAlerting_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirRuleGroupExists("mimir_rule_group_alerting.alert_1", "alert_1", client),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1", "name", "alert_1"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1", "namespace", "namespace_1"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1", "rule.0.alert", "test1"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1", "rule.0.expr", "test1_metric"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1", "rule.1.alert", "test2"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1", "rule.1.expr", "test2_metric"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1", "rule.1.for", ""),
				),
			},
			{
				Config: testAccResourceRuleGroupAlerting_basic_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirRuleGroupExists("mimir_rule_group_alerting.alert_1", "alert_1", client),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1", "name", "alert_1"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1", "namespace", "namespace_1"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1", "rule.0.alert", "test1"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1", "rule.0.expr", "test1_metric"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1", "rule.1.alert", "test2"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1", "rule.1.expr", "test2_metric"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1", "rule.1.for", "1m"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1", "rule.1.labels.severity", "critical"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1", "rule.1.annotations.summary", "test 2 alert summary"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1", "rule.1.annotations.description", "test 2 alert description"),
				),
			},
			{
				Config: testAccResourceRuleGroupAlerting_interval,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirRuleGroupExists("mimir_rule_group_alerting.alert_1_interval", "alert_1_interval", client),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_interval", "name", "alert_1_interval"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_interval", "namespace", "namespace_1"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_interval", "rule.0.alert", "test1_info"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_interval", "rule.0.expr", "test1_metric"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_interval", "interval", "6h"),
				),
			},
			{
				Config: testAccResourceRuleGroupAlerting_interval_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirRuleGroupExists("mimir_rule_group_alerting.alert_1_interval", "alert_1_interval", client),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_interval", "name", "alert_1_interval"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_interval", "namespace", "namespace_1"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_interval", "rule.0.alert", "test1_info"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_interval", "rule.0.expr", "test1_metric"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_interval", "interval", "10m"),
				),
			},
		},
	})
}

func TestAccResourceRuleGroupAlerting_Federated(t *testing.T) {
	// Init client
	client, err := NewAPIClient(setupClient())
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMimirRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRuleGroupAlerting_federated_rule_group,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirRuleGroupExists("mimir_rule_group_alerting.alert_1_federated_rule_group", "alert_1_federated_rule_group", client),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_federated_rule_group", "name", "alert_1_federated_rule_group"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_federated_rule_group", "namespace", "namespace_1"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_federated_rule_group", "source_tenants.0", "tenant-a"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_federated_rule_group", "source_tenants.1", "tenant-b"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_federated_rule_group", "rule.0.alert", "test1"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_federated_rule_group", "rule.0.expr", "test1_metric"),
				),
			},
			{
				Config: testAccResourceRuleGroupAlerting_federated_rule_group_tenant_change,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirRuleGroupExists("mimir_rule_group_alerting.alert_1_federated_rule_group", "alert_1_federated_rule_group", client),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_federated_rule_group", "name", "alert_1_federated_rule_group"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_federated_rule_group", "namespace", "namespace_1"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_federated_rule_group", "source_tenants.0", "tenant-a"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_federated_rule_group", "source_tenants.1", "tenant-c"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_federated_rule_group", "source_tenants.2", "tenant-d"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_federated_rule_group", "rule.0.alert", "test1"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_federated_rule_group", "rule.0.expr", "test1_metric"),
				),
			},
			{
				Config: testAccResourceRuleGroupAlerting_federated_rule_group_rule_change,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirRuleGroupExists("mimir_rule_group_alerting.alert_1_federated_rule_group", "alert_1_federated_rule_group", client),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_federated_rule_group", "name", "alert_1_federated_rule_group"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_federated_rule_group", "namespace", "namespace_1"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_federated_rule_group", "source_tenants.0", "tenant-a"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_federated_rule_group", "source_tenants.1", "tenant-c"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_federated_rule_group", "source_tenants.2", "tenant-d"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_federated_rule_group", "rule.0.alert", "test2"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_federated_rule_group", "rule.0.expr", "test2_metric"),
				),
			},
		},
	})
}

func TestAccResourceRuleGroupAlerting_PromQLValidation_HistogramAvg(t *testing.T) {

	/* Skip this test if mimir version is older than 2.12.0

	=== RUN   TestAccResourceRuleGroupAlerting_PromQLValidation_HistogramAvg
		resource_mimir_rule_group_alerting_test.go:245: Step 1/1 error: Error running apply: exit status 1
			2024/07/05 09:20:03 [DEBUG] Using modified User-Agent: Terraform/0.12.31 HashiCorp-terraform-exec/0.18.1

			Error: Cannot create alerting rule group 'alert_1_histogram_avg_rule_group' (namespace: namespace_1) - unexpected response code '400': 4:13: group "alert_1_histogram_avg_rule_group", rule 0, "test_histogram_avg": could not parse expression: 1:1: parse error: unknown function with name "histogram_avg"


			on terraform_plugin_test.tf line 2, in resource "mimir_rule_group_alerting" "alert_1_histogram_avg_rule_group":
			2: 	resource "mimir_rule_group_alerting" "alert_1_histogram_avg_rule_group" {


	--- FAIL: TestAccResourceRuleGroupAlerting_PromQLValidation_HistogramAvg (0.28s)

	*/
	currentVersion, _ := version.NewVersion(os.Getenv("MIMIR_VERSION"))
	minVersion, _ := version.NewVersion("2.12.0")

	if currentVersion.LessThan(minVersion) {
		fmt.Printf("Skipping PromQL HistogramAvg tests (current version '%s' is less than '%s')\n", currentVersion, minVersion)
		return
	}

	// Init client
	client, err := NewAPIClient(setupClient())
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMimirRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRuleGroupAlerting_promql_validation_histogram_avg,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirRuleGroupExists("mimir_rule_group_alerting.alert_1_histogram_avg_rule_group", "alert_1_histogram_avg_rule_group", client),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_histogram_avg_rule_group", "name", "alert_1_histogram_avg_rule_group"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_histogram_avg_rule_group", "namespace", "namespace_1"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_histogram_avg_rule_group", "rule.0.alert", "test_histogram_avg"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_histogram_avg_rule_group", "rule.0.expr", "histogram_avg(rate(test_metric[5m])) > 1"),
				),
			},
		},
	})
}

func TestAccResourceRuleGroupAlerting_FormatPromQLExpr(t *testing.T) {
	// Init client
	client, err := NewAPIClient(setupClient())
	if err != nil {
		t.Fatal(err)
	}
	os.Setenv("MIMIR_FORMAT_PROMQL_EXPR", "true")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMimirRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRuleGroupAlerting_prettify_promql_expr,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirRuleGroupExists("mimir_rule_group_alerting.alert_1_prettify", "alert_1_prettify", client),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_prettify", "name", "alert_1_prettify"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_prettify", "namespace", "namespace_1"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_prettify", "rule.0.alert", "checkPrettifyPromQL"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.alert_1_prettify", "rule.0.expr", "up == 0\nunless\n  my_very_very_long_useless_metric_that_mean_nothing_but_necessary_to_check_prettify_promql > 300"),
				),
			},
		},
	})
	os.Setenv("MIMIR_FORMAT_PROMQL_EXPR", "false")
}

const testAccResourceRuleGroupAlerting_basic = `
	resource "mimir_rule_group_alerting" "alert_1" {
		name = "alert_1"
		namespace = "namespace_1"
		rule {
			alert = "test1"
			expr  = "test1_metric"
		}
		rule {
			alert = "test2"
			expr   = "test2_metric"
			for    = "0s"
		}
	}
`

const testAccResourceRuleGroupAlerting_basic_update = `
	resource "mimir_rule_group_alerting" "alert_1" {
		name = "alert_1"
		namespace = "namespace_1"
		rule {
			alert = "test1"
			expr  = "test1_metric"
		}
		rule {
			alert = "test2"
			expr   = "test2_metric"
			for = "1m"
			labels = {
				severity = "critical"
			}
			annotations = {
				summary = "test 2 alert summary"
				description = "test 2 alert description"
			}
		}
	}
`

const testAccResourceRuleGroupAlerting_interval = `
    resource "mimir_rule_group_alerting" "alert_1_interval" {
            name = "alert_1_interval"
            namespace = "namespace_1"
            interval = "6h"
            rule {
                    alert = "test1_info"
                    expr  = "test1_metric"
            }
    }
`

const testAccResourceRuleGroupAlerting_interval_update = `
    resource "mimir_rule_group_alerting" "alert_1_interval" {
            name = "alert_1_interval"
            namespace = "namespace_1"
            interval = "10m"
            rule {
                    alert = "test1_info"
                    expr  = "test1_metric"
            }
    }
`

const testAccResourceRuleGroupAlerting_prettify_promql_expr = `
	resource "mimir_rule_group_alerting" "alert_1_prettify" {
		name = "alert_1_prettify"
		namespace = "namespace_1"
		rule {
			alert = "checkPrettifyPromQL"
			expr  = "up==0 unless my_very_very_long_useless_metric_that_mean_nothing_but_necessary_to_check_prettify_promql > 300"
		}
	}
`

const testAccResourceRuleGroupAlerting_federated_rule_group = `
	resource "mimir_rule_group_alerting" "alert_1_federated_rule_group" {
		name = "alert_1_federated_rule_group"
		source_tenants = ["tenant-a", "tenant-b"]
		namespace = "namespace_1"
		rule {
			alert = "test1"
			expr  = "test1_metric"
		}
	}
`

const testAccResourceRuleGroupAlerting_federated_rule_group_tenant_change = `
	resource "mimir_rule_group_alerting" "alert_1_federated_rule_group" {
		name = "alert_1_federated_rule_group"
		source_tenants = ["tenant-a", "tenant-c", "tenant-d"]
		namespace = "namespace_1"
		rule {
			alert = "test1"
			expr  = "test1_metric"
		}
	}
`

const testAccResourceRuleGroupAlerting_federated_rule_group_rule_change = `
	resource "mimir_rule_group_alerting" "alert_1_federated_rule_group" {
		name = "alert_1_federated_rule_group"
		source_tenants = ["tenant-a", "tenant-c", "tenant-d"]
		namespace = "namespace_1"
		rule {
			alert = "test2"
			expr  = "test2_metric"
		}
	}
`

const testAccResourceRuleGroupAlerting_promql_validation_histogram_avg = `
	resource "mimir_rule_group_alerting" "alert_1_histogram_avg_rule_group" {
		name = "alert_1_histogram_avg_rule_group"
		namespace = "namespace_1"
		rule {
			alert = "test_histogram_avg"
			expr  = "histogram_avg(rate(test_metric[5m])) > 1"
		}
	}
`
