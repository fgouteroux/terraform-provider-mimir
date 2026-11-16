package mimir

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// The single-group resources rebuild the whole group from state on every update,
// so any group option they do not model is stripped from the group in Mimir.
// These tests set, change and then drop each option to pin the round trip.
func TestAccResourceRuleGroupAlerting_GroupOptions(t *testing.T) {
	// limit is only persisted by Mimir >= 3.0.0.
	skipBelowMimirVersion(t, "3.0.0")

	client, err := NewAPIClient(setupClient())
	if err != nil {
		t.Fatal(err)
	}

	const res = "mimir_rule_group_alerting.alert_group_options"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMimirRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupAlertingOptions(`
            query_offset = "5m"
            align_evaluation_time_on_interval = true
            limit = 10`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirRuleGroupExists(res, "alert_group_options", client),
					resource.TestCheckResourceAttr(res, queryOffsetKey, "5m"),
					resource.TestCheckResourceAttr(res, alignEvaluationTimeOnIntervalKey, "true"),
					resource.TestCheckResourceAttr(res, ruleGroupLimitKey, "10"),
				),
			},
			{
				Config: testAccRuleGroupAlertingOptions(`
            query_offset = "1m"
            limit = 20`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(res, queryOffsetKey, "1m"),
					resource.TestCheckResourceAttr(res, alignEvaluationTimeOnIntervalKey, "false"),
					resource.TestCheckResourceAttr(res, ruleGroupLimitKey, "20"),
				),
			},
			{
				Config: testAccRuleGroupAlertingOptions(""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(res, queryOffsetKey, ""),
					resource.TestCheckResourceAttr(res, alignEvaluationTimeOnIntervalKey, "false"),
					resource.TestCheckResourceAttr(res, ruleGroupLimitKey, "0"),
				),
			},
		},
	})
}

func TestAccResourceRuleGroupRecording_GroupOptions(t *testing.T) {
	// limit is only persisted by Mimir >= 3.0.0.
	skipBelowMimirVersion(t, "3.0.0")

	client, err := NewAPIClient(setupClient())
	if err != nil {
		t.Fatal(err)
	}

	const res = "mimir_rule_group_recording.record_group_options"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMimirRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupRecordingOptions(`
            align_evaluation_time_on_interval = true
            limit = 10`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirRuleGroupExists(res, "record_group_options", client),
					resource.TestCheckResourceAttr(res, alignEvaluationTimeOnIntervalKey, "true"),
					resource.TestCheckResourceAttr(res, ruleGroupLimitKey, "10"),
				),
			},
			{
				Config: testAccRuleGroupRecordingOptions(""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(res, alignEvaluationTimeOnIntervalKey, "false"),
					resource.TestCheckResourceAttr(res, ruleGroupLimitKey, "0"),
				),
			},
		},
	})
}

func testAccRuleGroupAlertingOptions(options string) string {
	return `
    resource "mimir_rule_group_alerting" "alert_group_options" {
            name = "alert_group_options"
            namespace = "namespace_1"
            interval = "1m"` + options + `
            rule {
                    alert = "test1_alert"
                    expr  = "test1_metric"
            }
    }
`
}

func testAccRuleGroupRecordingOptions(options string) string {
	return `
    resource "mimir_rule_group_recording" "record_group_options" {
            name = "record_group_options"
            namespace = "namespace_1"
            interval = "1m"` + options + `
            rule {
                    record = "test1_info"
                    expr   = "test1_metric"
            }
    }
`
}
