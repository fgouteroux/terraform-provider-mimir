resource "mimir_rule_group_recording" "test" {
  name         = "test1"
  namespace    = "namespace1"
  interval     = "6h"
  query_offset = "5m"
  # Group-level labels are added to every rule in the group (requires Mimir >= 3.0.0).
  labels = {
    team = "observability"
  }
  rule {
    expr   = "sum by (job) (http_inprogress_requests)"
    record = "job:http_inprogress_requests:sum"
  }
}