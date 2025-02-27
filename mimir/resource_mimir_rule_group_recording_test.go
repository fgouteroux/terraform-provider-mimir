package mimir

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/go-version"
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
			{
				Config: testAccResourceRuleGroupRecording_interval_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirRuleGroupExists("mimir_rule_group_recording.record_1_interval", "record_1_interval", client),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_interval", "name", "record_1_interval"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_interval", "namespace", "namespace_1"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_interval", "rule.0.record", "test1_info"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_interval", "rule.0.expr", "test1_metric"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_interval", "interval", "10m"),
				),
			},
		},
	})
}

func TestAccResourceRuleGroupRecording_EvaluationDelay(t *testing.T) {
	currentVersion, _ := version.NewVersion(os.Getenv("MIMIR_VERSION"))
	validVersion, _ := version.NewVersion("2.12.0")

	if !currentVersion.Equal(validVersion) {
		fmt.Printf("Skipping ruler evaluation delay tests (current version '%s' is not '%s')\n", currentVersion, validVersion)
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
				Config: testAccResourceRuleGroupRecording_evaluation_delay,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirRuleGroupExists("mimir_rule_group_recording.record_1_evaluation_delay", "record_1_evaluation_delay", client),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_evaluation_delay", "name", "record_1_evaluation_delay"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_evaluation_delay", "namespace", "namespace_1"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_evaluation_delay", "rule.0.record", "test1_info"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_evaluation_delay", "rule.0.expr", "test1_metric"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_evaluation_delay", "evaluation_delay", "5m"),
				),
			},
			{
				Config: testAccResourceRuleGroupRecording_evaluation_delay_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirRuleGroupExists("mimir_rule_group_recording.record_1_evaluation_delay", "record_1_evaluation_delay", client),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_evaluation_delay", "name", "record_1_evaluation_delay"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_evaluation_delay", "namespace", "namespace_1"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_evaluation_delay", "rule.0.record", "test1_info"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_evaluation_delay", "rule.0.expr", "test1_metric"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_evaluation_delay", "evaluation_delay", "1m"),
				),
			},
		},
	})
}

func TestAccResourceRuleGroupRecording_QueryOffset(t *testing.T) {
	currentVersion, _ := version.NewVersion(os.Getenv("MIMIR_VERSION"))
	minVersion, _ := version.NewVersion("2.13.0")

	if currentVersion.LessThan(minVersion) {
		fmt.Printf("Skipping ruler query offset tests (current version '%s' is less than '%s')\n", currentVersion, minVersion)
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
				Config: testAccResourceRuleGroupRecording_query_offset,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirRuleGroupExists("mimir_rule_group_recording.record_1_query_offset", "record_1_query_offset", client),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_query_offset", "name", "record_1_query_offset"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_query_offset", "namespace", "namespace_1"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_query_offset", "rule.0.record", "test1_info"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_query_offset", "rule.0.expr", "test1_metric"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_query_offset", "query_offset", "5m"),
				),
			},
			{
				Config: testAccResourceRuleGroupRecording_query_offset_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirRuleGroupExists("mimir_rule_group_recording.record_1_query_offset", "record_1_query_offset", client),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_query_offset", "name", "record_1_query_offset"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_query_offset", "namespace", "namespace_1"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_query_offset", "rule.0.record", "test1_info"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_query_offset", "rule.0.expr", "test1_metric"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_query_offset", "query_offset", "1m"),
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
			{
				Config: testAccResourceRuleGroupRecording_federated_rule_group_tenant_change,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirRuleGroupExists("mimir_rule_group_recording.record_1_federated_rule_group", "record_1_federated_rule_group", client),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_federated_rule_group", "name", "record_1_federated_rule_group"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_federated_rule_group", "namespace", "namespace_1"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_federated_rule_group", "source_tenants.0", "tenant-a"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_federated_rule_group", "source_tenants.1", "tenant-c"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_federated_rule_group", "rule.0.record", "test1_info"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_federated_rule_group", "rule.0.expr", "test1_metric"),
				),
			},
			{
				Config: testAccResourceRuleGroupRecording_federated_rule_group_rule_change,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirRuleGroupExists("mimir_rule_group_recording.record_1_federated_rule_group", "record_1_federated_rule_group", client),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_federated_rule_group", "name", "record_1_federated_rule_group"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_federated_rule_group", "namespace", "namespace_1"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_federated_rule_group", "source_tenants.0", "tenant-a"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_federated_rule_group", "source_tenants.1", "tenant-c"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_federated_rule_group", "rule.0.record", "test2_info"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_federated_rule_group", "rule.0.expr", "test2_metric"),
				),
			},
		},
	})
}

func TestAccResourceRuleGroupRecording_WithOrgID(t *testing.T) {
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
				Config: testAccResourceRuleGroupRecording_withOrgID,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirRuleGroupExists("mimir_rule_group_recording.record_1_withOrgID", "record_1_withOrgID", client),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_withOrgID", "org_id", "another_tenant"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_withOrgID", "name", "record_1_withOrgID"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_withOrgID", "namespace", "namespace_1"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_withOrgID", "rule.0.record", "test1_info"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.record_1_withOrgID", "rule.0.expr", "test1_metric"),
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

const testAccResourceRuleGroupRecording_federated_rule_group_tenant_change = `
	resource "mimir_rule_group_recording" "record_1_federated_rule_group" {
		name = "record_1_federated_rule_group"
		namespace = "namespace_1"
		source_tenants = ["tenant-a", "tenant-c"]
		rule {
			record = "test1_info"
			expr   = "test1_metric"
		}
	}
`

const testAccResourceRuleGroupRecording_federated_rule_group_rule_change = `
	resource "mimir_rule_group_recording" "record_1_federated_rule_group" {
		name = "record_1_federated_rule_group"
		namespace = "namespace_1"
		source_tenants = ["tenant-a", "tenant-c"]
		rule {
			record = "test2_info"
			expr   = "test2_metric"
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

const testAccResourceRuleGroupRecording_interval_update = `
    resource "mimir_rule_group_recording" "record_1_interval" {
            name = "record_1_interval"
            namespace = "namespace_1"
            interval = "10m"
            rule {
                    record = "test1_info"
                    expr   = "test1_metric"
            }
    }
`

const testAccResourceRuleGroupRecording_evaluation_delay = `
    resource "mimir_rule_group_recording" "record_1_evaluation_delay" {
            name = "record_1_evaluation_delay"
            namespace = "namespace_1"
            evaluation_delay = "5m"
            rule {
                    record = "test1_info"
                    expr   = "test1_metric"
            }
    }
`

const testAccResourceRuleGroupRecording_evaluation_delay_update = `
    resource "mimir_rule_group_recording" "record_1_evaluation_delay" {
            name = "record_1_evaluation_delay"
            namespace = "namespace_1"
            evaluation_delay = "1m"
            rule {
                    record = "test1_info"
                    expr   = "test1_metric"
            }
    }
`

const testAccResourceRuleGroupRecording_query_offset = `
    resource "mimir_rule_group_recording" "record_1_query_offset" {
            name = "record_1_query_offset"
            namespace = "namespace_1"
            query_offset = "5m"
            rule {
                    record = "test1_info"
                    expr   = "test1_metric"
            }
    }
`

const testAccResourceRuleGroupRecording_query_offset_update = `
    resource "mimir_rule_group_recording" "record_1_query_offset" {
            name = "record_1_query_offset"
            namespace = "namespace_1"
            query_offset = "1m"
            rule {
                    record = "test1_info"
                    expr   = "test1_metric"
            }
    }
`

const testAccResourceRuleGroupRecording_withOrgID = `
    resource "mimir_rule_group_recording" "record_1_withOrgID" {
            org_id = "another_tenant"
            name = "record_1_withOrgID"
            namespace = "namespace_1"
            rule {
                record = "test1_info"
                expr  = "test1_metric"
            }
    }
`
