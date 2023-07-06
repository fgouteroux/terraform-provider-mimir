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