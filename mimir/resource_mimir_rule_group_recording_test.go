package mimir

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceRuleGroupRecording_expectValidationError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceRuleGroupRecording_expectNameValidationError,
				ExpectError: regexp.MustCompile("Invalid Group Rule Name"),
			},
			{
				Config:      testAccResourceRuleGroupRecording_expectRuleNameValidationError,
				ExpectError: regexp.MustCompile("Invalid Recording Rule Name"),
			},
			{
				Config:      testAccResourceRuleGroupRecording_expectPromQLValidationError,
				ExpectError: regexp.MustCompile("Invalid PromQL expression"),
			},
			{
				Config:      testAccResourceRuleGroupRecording_expectLabelNameValidationError,
				ExpectError: regexp.MustCompile("Invalid Label Name"),
			},
		},
	})
}

const testAccResourceRuleGroupRecording_expectNameValidationError = `
	resource "mimir_rule_group_recording" "record_1" {
		name = "record_1-@error"
		namespace = "namespace_1"
		rule {
			record = "test1_info"
			expr   = "test1_metric"
		}
	}
`
const testAccResourceRuleGroupRecording_expectRuleNameValidationError = `
	resource "mimir_rule_group_recording" "record_1" {
		name = "record_1"
		namespace = "namespace_1"
		rule {
			record = "test1_info;error"
			expr   = "test1_metric"
		}
	}
`

const testAccResourceRuleGroupRecording_expectPromQLValidationError = `
	resource "mimir_rule_group_recording" "record_1" {
		name = "record_1"
		namespace = "namespace_1"
		rule {
			record = "test1_info"
			expr   = "rate(hi)"
		}
	}
`
const testAccResourceRuleGroupRecording_expectLabelNameValidationError = `
	resource "mimir_rule_group_recording" "record_1" {
		name = "record_1"
		namespace = "namespace_1"
		rule {
			record = "test1_info"
			expr   = "test1_metric"
			labels = {
				 ins-tance = "localhost"
			}
		}
	}
`

func TestAccResourceRuleGroupRecording_Basic(t *testing.T) {
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
				Config: testAccResourceRuleGroupRecording_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirRuleGroupExists("mimir_rule_group_recording.record_1", "record_1", client),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1", "name", "record_1"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1", "namespace", "namespace_1"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1", "rule.0.record", "test1_info"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1", "rule.0.expr", "test1_metric"),
				),
			},
			{
				Config: testAccResourceRuleGroupRecording_basic_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirRuleGroupExists("mimir_rule_group_recording.record_1", "record_1", client),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1", "name", "record_1"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1", "namespace", "namespace_1"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1", "rule.0.record", "test1_info"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1", "rule.0.expr", "test1_metric"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1", "rule.1.record", "test2_info"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1", "rule.1.expr", "test2_metric"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1", "rule.1.labels.key1", "val1"),
				),
			},
			{
				Config: testAccResourceRuleGroupRecording_interval,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirRuleGroupExists("mimir_rule_group_recording.record_1_interval", "record_1_interval", client),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_interval", "name", "record_1_interval"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_interval", "namespace", "namespace_1"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_interval", "rule.0.record", "test1_info"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_interval", "rule.0.expr", "test1_metric"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_interval", "interval", "6h"),
				),
			},
		},
	})
}

func TestAccResourceRuleGroupRecording_Federated(t *testing.T) {
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
				Config: testAccResourceRuleGroupRecording_federated_rule_group,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirRuleGroupExists("mimir_rule_group_recording.record_1_federated_rule_group", "record_1_federated_rule_group", client),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_federated_rule_group", "name", "record_1_federated_rule_group"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_federated_rule_group", "namespace", "namespace_1"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_federated_rule_group", "source_tenants.0", "tenant-a"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_federated_rule_group", "source_tenants.1", "tenant-b"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_federated_rule_group", "rule.0.record", "test1_info"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_federated_rule_group", "rule.0.expr", "test1_metric"),
				),
			},
		},
	})
}

const testAccResourceRuleGroupRecording_basic = `
	resource "mimir_rule_group_recording" "record_1" {
		name = "record_1"
		namespace = "namespace_1"
		rule {
			record = "test1_info"
			expr   = "test1_metric"
		}
	}
`

const testAccResourceRuleGroupRecording_basic_update = `
	resource "mimir_rule_group_recording" "record_1" {
		name = "record_1"
		namespace = "namespace_1"
		rule {
			record = "test1_info"
			expr   = "test1_metric"
		}
		rule {
			record = "test2_info"
			expr   = "test2_metric"
			labels = {
				key1 = "val1"
			}
		}
	}
`

const testAccResourceRuleGroupRecording_federated_rule_group = `
	resource "mimir_rule_group_recording" "record_1_federated_rule_group" {
		name = "record_1_federated_rule_group"
		namespace = "namespace_1"
		source_tenants = ["tenant-a", "tenant-b"]
		rule {
			record = "test1_info"
			expr   = "test1_metric"
		}
	}
`

const testAccResourceRuleGroupRecording_interval = `
    resource "mimir_rule_group_recording" "record_1_interval" {
            name = "record_1_interval"
            namespace = "namespace_1"
            interval = "6h"
            rule {
                    record = "test1_info"
                    expr   = "test1_metric"
            }
    }
`
