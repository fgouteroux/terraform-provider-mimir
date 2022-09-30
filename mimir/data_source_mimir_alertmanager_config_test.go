package mimir

import (
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

var testAccDataSourceAlertmanagerConfig_basic = `
    resource "mimir_alertmanager_config" "mytenant" {
      route {
        group_by = ["..."]
        group_wait = "30s"
        group_interval = "5m"
        repeat_interval = "1h"
        receiver = "pagerduty"
      }
      receiver {
        name = "pagerduty"
        pagerduty_configs {
          routing_key = "secret"
          details = {
            environment = "dev"
          }
        }
      }
    }

	data "mimir_alertmanager_config" "mytenant" {
		depends_on = [mimir_alertmanager_config.mytenant]
	}
`
