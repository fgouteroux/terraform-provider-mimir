resource "mimir_alertmanager_config" "mytenant" {
  route {
    group_by = ["..."]
    group_wait = "30s"
    group_interval = "5m"
    repeat_interval = "1h"
    receiver = "pagerduty_dev"
  }
  child_route {
    matchers = ["severity=\"critical\""]

    child_route {
      receiver = "pagerduty_prod"
      matchers = ["environment=\"prod\""]
    }

    child_route {
      receiver = "pagerduty_dev"
      matchers = ["environment=\"dev\""]
    }
  }
  receiver {
    name = "pagerduty_dev"
    pagerduty_configs {
      routing_key = "secret"
      details = {
        environment = "dev"
      }
    }
  }
  receiver {
    name = "pagerduty_prod"
    pagerduty_configs {
      routing_key = "secret"
      details = {
        environment = "prod"
      }
    }
  }
}