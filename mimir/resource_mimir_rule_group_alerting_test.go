package mimir

import (
	"regexp"
	"testing"

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
				ExpectError: regexp.MustCompile("not a valid duration string"),
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
}

const testAccResourceRuleGroupAlerting_basic = `
	resource "mimir_rule_group_alerting" "alert_1" {
		name = "alert_1"
		namespace = "namespace_1"
		rule {
			alert = "test1"
			expr  = "test1_metric"
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
