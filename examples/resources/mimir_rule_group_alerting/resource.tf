resource "mimir_rule_group_alerting" "test" {
  name      = "test1"
  namespace = "namespace1"
  # Group-level labels are added to every rule in the group (requires Mimir >= 3.0.0).
  labels = {
    target_channel = "Storage"
  }
  rule {
    alert       = "HighRequestLatency"
    expr        = "job:request_latency_seconds:mean5m{job=\"myjob\"} > 0.5"
    for         = "10m"
    labels      = {
      severity = "warning"
    }
    annotations = {
      summary = "High request latency"
    }
  }
}