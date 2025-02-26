package mimir

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gopkg.in/yaml.v3"
)

func dataSourcemimirRuleGroupAlerting() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcemimirRuleGroupAlertingRead,

		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Description: "The Organization ID. If not set, the Org ID defined in the provider block will be used.",
			},
			"namespace": {
				Type:        schema.TypeString,
				Description: "Alerting Rule group namespace",
				ForceNew:    true,
				Optional:    true,
				Default:     "default",
			},
			"name": {
				Type:         schema.TypeString,
				Description:  "Alerting Rule group name",
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateGroupRuleName,
			},
			"source_tenants": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Allows aggregating data from multiple tenants while evaluating a rule group.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"rule": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"alert": {
							Type:        schema.TypeString,
							Description: "Alerting Rule name",
							Computed:    true,
						},
						"expr": {
							Type:        schema.TypeString,
							Description: "Alerting Rule query",
							Computed:    true,
						},
						"for": {
							Type:        schema.TypeString,
							Description: "Alerting Rule duration",
							Computed:    true,
						},
						"annotations": {
							Type:        schema.TypeMap,
							Description: "Alerting Rule annotations",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
						},
						"labels": {
							Type:        schema.TypeMap,
							Description: "Alerting Rule labels",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
						},
					},
				},
			},
		}, /* End schema */

	}
}

func dataSourcemimirRuleGroupAlertingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)
	name := d.Get("name").(string)
	namespace := d.Get("namespace").(string)
	orgID := d.Get("org_id").(string)

	id := fmt.Sprintf("%s/%s", namespace, name)

	headers := make(map[string]string)
	if orgID != "" {
		headers["X-Scope-OrgID"] = orgID
		id = fmt.Sprintf("%s/%s/%s", orgID, namespace, name)
	}
	path := fmt.Sprintf("/config/v1/rules/%s/%s", namespace, name)
	jobraw, err := client.sendRequest("ruler", "GET", path, "", headers)

	baseMsg := fmt.Sprintf("Cannot read alerting rule group '%s' -", name)
	err = handleHTTPError(err, baseMsg)
	if err != nil {
		if strings.Contains(err.Error(), "response code '404'") {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(id)

	var data alertingRuleGroup
	err = yaml.Unmarshal([]byte(jobraw), &data)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to decode alerting rule group '%s' data: %v", name, err))
	}
	if err := d.Set("rule", flattenAlertingRules(data.Rules)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("source_tenants", data.SourceTenants); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
