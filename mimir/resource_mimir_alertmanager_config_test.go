package mimir

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"strings"
	"testing"
)

func testAccCheckMimirAlertmanagerConfigExists(n string, client *api_client) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			keys := make([]string, 0, len(s.RootModule().Resources))
			for k := range s.RootModule().Resources {
				keys = append(keys, k)
			}
			return fmt.Errorf("mimir object not found in terraform state: %s. Found: %s", n, strings.Join(keys, ", "))
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("mimir object not set in terraform")
		}

		/* Make a throw-away API object to read from the API */
		_, err := client.send_request("alertmanager", "GET", "/api/v1/alerts", "", make(map[string]string))
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckMimirAlertmanagerConfigDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*api_client)

	// loop through the resources in state, verifying each widget
	// is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "mimir_alertmanager_config" {
			continue
		}
		_, err := client.send_request("alertmanager", "GET", "/api/v1/alerts", "", make(map[string]string))
		// If the error is equivalent to 404 not found, the widget is destroyed.
		// Otherwise return the error
		if !strings.Contains(err.Error(), "not found") {
			return err
		}
	}

	return nil
}

func TestAccResourceAlertmanagerConfig_Basic(t *testing.T) {
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
				Config: testAccResourceAlertmanagerConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirAlertmanagerConfigExists("mimir_alertmanager_config.mytenant", client),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_by.0", "..."),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_wait", "30s"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_interval", "5m"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.repeat_interval", "1h"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.receiver", "pagerduty"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.name", "pagerduty"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.pagerduty_configs.0.routing_key", "secret"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.pagerduty_configs.0.details.environment", "dev"),
				),
			},
			{
				Config: testAccResourceAlertmanagerConfig_basic_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirAlertmanagerConfigExists("mimir_alertmanager_config.mytenant", client),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_by.0", "..."),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_wait", "1m"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_interval", "5m"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.repeat_interval", "1h"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.receiver", "pagerduty"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.child_route.0.group_by.0", "..."),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.child_route.0.group_wait", "30s"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.child_route.0.group_interval", "15m"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.child_route.0.repeat_interval", "1h"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.child_route.0.receiver", "pagerduty"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.name", "pagerduty"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.pagerduty_configs.0.routing_key", "secret"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.pagerduty_configs.0.details.environment", "dev"),
				),
			},
			{
				Config: testAccResourceAlertmanagerConfig_global,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirAlertmanagerConfigExists("mimir_alertmanager_config.mytenant", client),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "global.0.resolve_timeout", "5m"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "global.0.http_config.0.follow_redirects", "true"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_by.0", "..."),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_wait", "30s"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_interval", "5m"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.repeat_interval", "1h"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.receiver", "pagerduty"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.name", "pagerduty"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.pagerduty_configs.0.routing_key", "secret"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.pagerduty_configs.0.details.environment", "dev"),
				),
			},
			{
				Config: testAccResourceAlertmanagerConfig_global_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirAlertmanagerConfigExists("mimir_alertmanager_config.mytenant", client),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "global.0.resolve_timeout", "15m"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "global.0.http_config.0.follow_redirects", "true"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_by.0", "..."),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_wait", "30s"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_interval", "5m"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.repeat_interval", "1h"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.receiver", "pagerduty"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.name", "pagerduty"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.pagerduty_configs.0.routing_key", "secret"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.pagerduty_configs.0.details.environment", "dev"),
				),
			},
			{
				Config: testAccResourceAlertmanagerConfig_inhibit_rule,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirAlertmanagerConfigExists("mimir_alertmanager_config.mytenant", client),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "inhibit_rule.0.source_matchers.#", "1"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "inhibit_rule.0.source_matchers.0", "severity=\"critical\""),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "inhibit_rule.0.target_matchers.#", "2"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "inhibit_rule.0.target_matchers.0", "ignore_inhibit!=\"true\""),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "inhibit_rule.0.target_matchers.1", "severity=\"warning\""),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "inhibit_rule.0.equal.#", "1"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "inhibit_rule.0.equal.0", "alertname"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_by.0", "..."),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_wait", "30s"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_interval", "5m"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.repeat_interval", "1h"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.receiver", "pagerduty"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.name", "pagerduty"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.pagerduty_configs.0.routing_key", "secret"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.pagerduty_configs.0.details.environment", "dev"),
				),
			},
			{
				Config: testAccResourceAlertmanagerConfig_inhibit_rule_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirAlertmanagerConfigExists("mimir_alertmanager_config.mytenant", client),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "inhibit_rule.0.source_matchers.#", "1"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "inhibit_rule.0.source_matchers.0", "severity=\"warning\""),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "inhibit_rule.0.target_matchers.#", "1"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "inhibit_rule.0.target_matchers.0", "ignore_inhibit!=\"true\""),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "inhibit_rule.0.equal.#", "1"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "inhibit_rule.0.equal.0", "alertname"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_by.0", "..."),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_wait", "30s"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_interval", "5m"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.repeat_interval", "1h"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.receiver", "pagerduty"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.name", "pagerduty"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.pagerduty_configs.0.routing_key", "secret"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.pagerduty_configs.0.details.environment", "dev"),
				),
			},
			{

				Config: testAccResourceAlertmanagerConfig_time_interval,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirAlertmanagerConfigExists("mimir_alertmanager_config.mytenant", client),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "time_interval.0.name", "offhours"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "time_interval.0.time_intervals.0.weekdays.0.begin", "0"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "time_interval.0.time_intervals.0.weekdays.0.end", "6"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_by.0", "..."),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_wait", "30s"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_interval", "5m"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.repeat_interval", "1h"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.receiver", "pagerduty"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.name", "pagerduty"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.pagerduty_configs.0.routing_key", "secret"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.pagerduty_configs.0.details.environment", "dev"),
				),
			},
			{
				Config: testAccResourceAlertmanagerConfig_time_interval_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirAlertmanagerConfigExists("mimir_alertmanager_config.mytenant", client),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "time_interval.0.name", "offhours"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "time_interval.0.time_intervals.0.weekdays.0.begin", "5"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "time_interval.0.time_intervals.0.weekdays.0.end", "6"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_by.0", "..."),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_wait", "30s"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_interval", "5m"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.repeat_interval", "1h"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.receiver", "pagerduty"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.name", "pagerduty"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.pagerduty_configs.0.routing_key", "secret"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.pagerduty_configs.0.details.environment", "dev"),
				),
			},
			{

				Config: testAccResourceAlertmanagerConfig_templates_files,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirAlertmanagerConfigExists("mimir_alertmanager_config.mytenant", client),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "templates.#", "1"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "templates.0", "default_template"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "templates_files.default_template", "default template text file"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_by.0", "..."),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_wait", "30s"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_interval", "5m"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.repeat_interval", "1h"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.receiver", "pagerduty"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.name", "pagerduty"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.pagerduty_configs.0.routing_key", "secret"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.pagerduty_configs.0.details.environment", "dev"),
				),
			},
			{
				Config: testAccResourceAlertmanagerConfig_templates_files_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirAlertmanagerConfigExists("mimir_alertmanager_config.mytenant", client),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "templates.#", "1"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "templates.0", "default_template"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "templates_files.default_template", "updated template text file"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_by.0", "..."),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_wait", "30s"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_interval", "5m"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.repeat_interval", "1h"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.receiver", "pagerduty"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.name", "pagerduty"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.pagerduty_configs.0.routing_key", "secret"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.pagerduty_configs.0.details.environment", "dev"),
				),
			},
		},
	})
}

const testAccResourceAlertmanagerConfig_basic = `
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
`

const testAccResourceAlertmanagerConfig_basic_update = `
    resource "mimir_alertmanager_config" "mytenant" {
      route {
        group_by = ["..."]
        group_wait = "1m"
        group_interval = "5m"
        repeat_interval = "1h"
        receiver = "pagerduty"
        child_route {
          group_by = ["..."]
          group_wait = "30s"
          group_interval = "15m"
          repeat_interval = "1h"
          receiver = "pagerduty"
        }
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
`

const testAccResourceAlertmanagerConfig_global = `
    resource "mimir_alertmanager_config" "mytenant" {
      global {
        resolve_timeout = "5m"
        http_config {
          follow_redirects = true
        }
      }
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
`

const testAccResourceAlertmanagerConfig_global_update = `
    resource "mimir_alertmanager_config" "mytenant" {
      global {
        resolve_timeout = "15m"
        http_config {
          follow_redirects = true
        }
      }
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
`

const testAccResourceAlertmanagerConfig_inhibit_rule = `
    resource "mimir_alertmanager_config" "mytenant" {
      inhibit_rule {
        source_matchers = ["severity=\"critical\""]
        target_matchers = ["ignore_inhibit!=\"true\"", "severity=\"warning\""]
        equal = ["alertname"]
      }
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
`

const testAccResourceAlertmanagerConfig_inhibit_rule_update = `
    resource "mimir_alertmanager_config" "mytenant" {
      inhibit_rule {
        source_matchers = ["severity=\"warning\""]
        target_matchers = ["ignore_inhibit!=\"true\""]
        equal = ["alertname"]
      }
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
`

const testAccResourceAlertmanagerConfig_time_interval = `
    resource "mimir_alertmanager_config" "mytenant" {
      time_interval {
        name = "offhours"
        time_intervals {
          weekdays {
            begin = 0
            end = 6
          }
        }
      }
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
`

const testAccResourceAlertmanagerConfig_time_interval_update = `
    resource "mimir_alertmanager_config" "mytenant" {
      time_interval {
        name = "offhours"
        time_intervals {
          weekdays {
            begin = 5
            end = 6
          }
        }
      }
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
`

const testAccResourceAlertmanagerConfig_templates_files = `
    resource "mimir_alertmanager_config" "mytenant" {
      templates = ["default_template"]
      templates_files = {
        default_template = "default template text file"
      }
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
`

const testAccResourceAlertmanagerConfig_templates_files_update = `
    resource "mimir_alertmanager_config" "mytenant" {
      templates = ["default_template"]
      templates_files = {
        default_template = "updated template text file"
      }
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
`

func TestAccResourceAlertmanagerConfig_PagerdutyReceiver(t *testing.T) {
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
				Config: testAccResourceAlertmanagerConfig_PagerdutyReceiver,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirAlertmanagerConfigExists("mimir_alertmanager_config.mytenant", client),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_by.0", "..."),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_wait", "30s"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_interval", "5m"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.repeat_interval", "1h"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.receiver", "pagerduty"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.name", "pagerduty"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.pagerduty_configs.0.routing_key", "secret"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.pagerduty_configs.0.details.environment", "dev"),
				),
			},
			{
				Config: testAccResourceAlertmanagerConfig_PagerdutyReceiver_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirAlertmanagerConfigExists("mimir_alertmanager_config.mytenant", client),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_by.0", "..."),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_wait", "30s"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_interval", "5m"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.repeat_interval", "1h"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.receiver", "pagerduty"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.name", "pagerduty"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.pagerduty_configs.0.routing_key", "secret"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.pagerduty_configs.0.details.environment", "dev2"),
				),
			},
		},
	})
}

const testAccResourceAlertmanagerConfig_PagerdutyReceiver = `
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
`

const testAccResourceAlertmanagerConfig_PagerdutyReceiver_update = `
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
            environment = "dev2"
          }
        }
      }
    }
`

func TestAccResourceAlertmanagerConfig_EmailReceiver(t *testing.T) {
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
				Config: testAccResourceAlertmanagerConfig_EmailReceiver,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirAlertmanagerConfigExists("mimir_alertmanager_config.mytenant", client),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_by.0", "..."),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_wait", "30s"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_interval", "5m"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.repeat_interval", "1h"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.receiver", "email"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.name", "email"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.email_configs.0.to", "user@example.com"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.email_configs.0.from", "no-reply@example.com"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.email_configs.0.smarthost", "smtp.example.com:25"),
				),
			},
			{
				Config: testAccResourceAlertmanagerConfig_EmailReceiver_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirAlertmanagerConfigExists("mimir_alertmanager_config.mytenant", client),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_by.0", "..."),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_wait", "30s"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.group_interval", "5m"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.repeat_interval", "1h"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "route.0.receiver", "email"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.name", "email"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.email_configs.0.to", "user2@example.com"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.email_configs.0.from", "no-reply@example.com"),
					resource.TestCheckResourceAttr("mimir_alertmanager_config.mytenant", "receiver.0.email_configs.0.smarthost", "smtp.example.com:25"),
				),
			},
		},
	})
}

const testAccResourceAlertmanagerConfig_EmailReceiver = `
    resource "mimir_alertmanager_config" "mytenant" {
      route {
        group_by = ["..."]
        group_wait = "30s"
        group_interval = "5m"
        repeat_interval = "1h"
        receiver = "email"
      }
      receiver {
        name = "email"
        email_configs {
          to = "user@example.com"
          from = "no-reply@example.com"
          smarthost = "smtp.example.com:25"
        }
      }
    }
`

const testAccResourceAlertmanagerConfig_EmailReceiver_update = `
    resource "mimir_alertmanager_config" "mytenant" {
      route {
        group_by = ["..."]
        group_wait = "30s"
        group_interval = "5m"
        repeat_interval = "1h"
        receiver = "email"
      }
      receiver {
        name = "email"
        email_configs {
          to = "user2@example.com"
          from = "no-reply@example.com"
          smarthost = "smtp.example.com:25"
        }
      }
    }
`
