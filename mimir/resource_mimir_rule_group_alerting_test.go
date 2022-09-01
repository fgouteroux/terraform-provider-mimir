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
