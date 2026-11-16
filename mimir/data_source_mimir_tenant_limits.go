package mimir

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const userLimitsPath = "/api/v1/user_limits"

// tenantLimitAttributes lists the fields of Mimir's /api/v1/user_limits
// response (pkg/util/validation/user_limits_handler.go). Each attribute keeps
// the API's JSON name.
var tenantLimitAttributes = []struct {
	name        string
	kind        schema.ValueType
	description string
}{
	{"compactor_blocks_retention_period_seconds", schema.TypeInt, "Retention of the tenant's blocks, in seconds. 0 means no retention limit."},
	{"ingestion_rate", schema.TypeFloat, "Per-tenant ingestion rate limit, in samples per second."},
	{"ingestion_burst_size", schema.TypeInt, "Per-tenant allowed ingestion burst size, in samples."},
	{"ingestion_burst_factor", schema.TypeFloat, "Per-tenant burst factor, multiplied by the ingestion rate to give the burst size."},
	{"max_global_series_per_user", schema.TypeInt, "Maximum number of in-memory series per tenant, across the cluster."},
	{"max_global_series_per_metric", schema.TypeInt, "Maximum number of in-memory series per metric name, across the cluster."},
	{"max_global_exemplars_per_user", schema.TypeInt, "Maximum number of in-memory exemplars per tenant, across the cluster."},
	{"max_fetched_chunks_per_query", schema.TypeInt, "Maximum number of chunks a single query may fetch."},
	{"max_fetched_series_per_query", schema.TypeInt, "Maximum number of series a single query may fetch."},
	{"max_fetched_chunk_bytes_per_query", schema.TypeInt, "Maximum size of the chunks a single query may fetch, in bytes."},
	{"ruler_max_rules_per_rule_group", schema.TypeInt, "Maximum number of rules per rule group."},
	{"ruler_max_rule_groups_per_tenant", schema.TypeInt, "Maximum number of rule groups per tenant."},
	{"alertmanager_notification_rate_limit", schema.TypeFloat, "Per-tenant rate limit for sending notifications, in notifications per second." + mimirMin32Descr},
	{"alertmanager_max_dispatcher_aggregation_groups", schema.TypeInt, "Maximum number of alert aggregation groups in the tenant's Alertmanager dispatcher." + mimirMin32Descr},
	{"alertmanager_max_templates_count", schema.TypeInt, "Maximum number of templates in the tenant's Alertmanager configuration." + mimirMin32Descr},
	{"alertmanager_max_alerts_count", schema.TypeInt, "Maximum number of alerts the tenant's Alertmanager holds." + mimirMin32Descr},
}

func dataSourcemimirTenantLimits() *schema.Resource {
	s := map[string]*schema.Schema{
		orgIDKey: {
			Type:         schema.TypeString,
			Optional:     true,
			Description:  orgIDDescription,
			ValidateFunc: validateOrgID,
		},
	}
	for _, attr := range tenantLimitAttributes {
		s[attr.name] = &schema.Schema{
			Type:        attr.kind,
			Computed:    true,
			Description: attr.description,
		}
	}

	return &schema.Resource{
		Description: "The effective limits of a tenant, as returned by Mimir's /api/v1/user_limits endpoint: the tenant's runtime overrides merged over the defaults. Mimir only serves this endpoint when it runs with -runtime-config.file. The request goes to alertmanager_uri, or uri when it is unset, since the endpoint is served at the root of every Mimir component.",
		ReadContext: dataSourcemimirTenantLimitsRead,
		Schema:      s,
	}
}

func dataSourcemimirTenantLimitsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)
	orgID := d.Get(orgIDKey).(string)

	headers := map[string]string{"Accept": "application/json"}
	if orgID != "" {
		headers["X-Scope-OrgID"] = orgID
	} else {
		orgID = client.headers["X-Scope-OrgID"]
	}

	body, err := client.sendRequest("alertmanager", "GET", userLimitsPath, "", headers)
	if errors.Is(err, errRouteNotFound) {
		return diag.FromErr(fmt.Errorf("cannot read tenant limits: %w. Mimir only serves %s when it runs with -runtime-config.file", err, userLimitsPath))
	}
	if err = handleHTTPError(err, "Cannot read tenant limits"); err != nil {
		return diag.FromErr(err)
	}

	var limits map[string]json.Number
	decoder := json.NewDecoder(strings.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&limits); err != nil {
		return diag.FromErr(fmt.Errorf("unable to decode tenant limits: %v", err))
	}

	for _, attr := range tenantLimitAttributes {
		value, present := limits[attr.name]
		if !present {
			// Older Mimir versions do not return every field.
			value = "0"
		}
		var setErr error
		switch attr.kind {
		case schema.TypeFloat:
			f, err := value.Float64()
			if err != nil {
				return diag.FromErr(fmt.Errorf("tenant limit %s: %v", attr.name, err))
			}
			setErr = d.Set(attr.name, f)
		default:
			i, err := value.Int64()
			if err != nil {
				return diag.FromErr(fmt.Errorf("tenant limit %s: %v", attr.name, err))
			}
			setErr = d.Set(attr.name, int(i))
		}
		if setErr != nil {
			return diag.FromErr(setErr)
		}
	}

	d.SetId(orgID)
	return nil
}
