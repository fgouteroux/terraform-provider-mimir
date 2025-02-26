package mimir

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAlertmanagerConfig_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAlertmanagerConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.mimir_alertmanager_config.mytenant", "route.0.group_by.0", "..."),
					resource.TestCheckResourceAttr("data.mimir_alertmanager_config.mytenant", "route.0.group_wait", "30s"),
					resource.TestCheckResourceAttr("data.mimir_alertmanager_config.mytenant", "route.0.group_interval", "5m"),
					resource.TestCheckResourceAttr("data.mimir_alertmanager_config.mytenant", "route.0.repeat_interval", "1h"),
					resource.TestCheckResourceAttr("data.mimir_alertmanager_config.mytenant", "route.0.receiver", "pagerduty"),
					resource.TestCheckResourceAttr("data.mimir_alertmanager_config.mytenant", "receiver.0.name", "pagerduty"),
					resource.TestCheckResourceAttr("data.mimir_alertmanager_config.mytenant", "receiver.0.pagerduty_configs.0.routing_key", "secret"),
					resource.TestCheckResourceAttr("data.mimir_alertmanager_config.mytenant", "receiver.0.pagerduty_configs.0.details.environment", "dev"),
				),
			},
		},
	})
}

var testAccDataSourceAlertmanagerConfig_basic = fmt.Sprintf(`
	%s

	data "mimir_alertmanager_config" "mytenant" {
		name = "${mimir_alertmanager_config.mytenant.id}"
	}
`, testAccResourceAlertmanagerConfig_basic)

func TestAccDataSourceAlertmanagerConfig_WithOrgID(t *testing.T) {
	// Init client
	client, err := NewAPIClient(setupClient())
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMimirAlertmanagerConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAlertmanagerConfig_WithOrgID,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirAlertmanagerConfigExists("mimir_alertmanager_config.another_tenant", client),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.another_tenant", "org_id", "another_tenant"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.another_tenant", "route.0.group_by.0", "..."),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.another_tenant", "route.0.group_wait", "30s"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.another_tenant", "route.0.group_interval", "5m"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.another_tenant", "route.0.repeat_interval", "1h"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.another_tenant", "route.0.receiver", "test"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.another_tenant", "receiver.0.name", "test"),
				),
			},
		},
	})
}

var testAccDataSourceAlertmanagerConfig_WithOrgID = fmt.Sprintf(`
	%s

	data "mimir_alertmanager_config" "another_tenant" {
		org_id = "another_tenant"
		name = "${mimir_alertmanager_config.another_tenant.id}"
	}
`, testAccResourceAlertmanagerConfig_WithOrgID)
