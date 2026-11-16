data "mimir_tenant_limits" "mytenant" {
  org_id = "mytenant"
}

output "max_series" {
  value = data.mimir_tenant_limits.mytenant.max_global_series_per_user
}
