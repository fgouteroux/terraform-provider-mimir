provider "mimir" {
  ruler_uri = "http://127.0.0.1:8080/prometheus"
  alertmanager_uri = "http://127.0.0.1:8080"
  distributor_uri = "http://127.0.0.1:8080/distributor"
  org_id = "mytenant"
  token = "supersecrettoken"
}
