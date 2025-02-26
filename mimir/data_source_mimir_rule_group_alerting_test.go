package mimir

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceRuleGroupAlerting_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceRuleGroupAlerting_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.mimir_rule_group_alerting.alert_1", "name", "alert_1"),
					resource.TestCheckResourceAttr("data.mimir_rule_group_alerting.alert_1", "namespace", "namespace_1"),
				),
			},
		},
	})
}

var testAccDataSourceRuleGroupAlerting_basic = fmt.Sprintf(`
	%s

	data "mimir_rule_group_alerting" "alert_1" {
		name = "${mimir_rule_group_alerting.alert_1.name}"
		namespace = "${mimir_rule_group_alerting.alert_1.namespace}"
	}
`, testAccResourceRuleGroupAlerting_basic)

func TestAccDataSourceRuleGroupAlerting_withOrgID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceRuleGroupAlerting_withOrgID,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.mimir_rule_group_alerting.alert_1_withOrgID", "org_id", "another_tenant"),
					resource.TestCheckResourceAttr("data.mimir_rule_group_alerting.alert_1_withOrgID", "name", "alert_1_withOrgID"),
					resource.TestCheckResourceAttr("data.mimir_rule_group_alerting.alert_1_withOrgID", "namespace", "namespace_1"),
				),
			},
		},
	})
}

var testAccDataSourceRuleGroupAlerting_withOrgID = fmt.Sprintf(`
	%s

	data "mimir_rule_group_alerting" "alert_1_withOrgID" {
		org_id = "${mimir_rule_group_alerting.alert_1_withOrgID.org_id}"
		name = "${mimir_rule_group_alerting.alert_1_withOrgID.name}"
		namespace = "${mimir_rule_group_alerting.alert_1_withOrgID.namespace}"
	}
`, testAccResourceRuleGroupAlerting_withOrgID)
