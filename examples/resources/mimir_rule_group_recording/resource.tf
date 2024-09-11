resource "mimir_rule_group_recording" "test" {
  name         = "test1"
  namespace    = "namespace1"
  interval     = "6h"
  query_offset = "5m"
  rule {
    expr   = "sum by (job) (http_inprogress_requests)"
    record = "job:http_inprogress_requests:sum"
  }
}