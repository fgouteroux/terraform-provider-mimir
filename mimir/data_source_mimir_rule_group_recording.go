package mimir

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gopkg.in/yaml.v3"
)

func dataSourcemimirRuleGroupRecording() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcemimirRuleGroupRecordingRead,

		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Description: "The organization id to operate on within mimir.",
			},
			"namespace": {
				Type:        schema.TypeString,
				Description: "Recording Rule group namespace",
				ForceNew:    true,
				Optional:    true,
				Default:     "default",
			},
			"name": {
				Type:         schema.TypeString,
				Description:  "Recording Rule group name",
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateGroupRuleName,
			},
			"interval": {
				Type:        schema.TypeString,
				Description: "Recording Rule group interval",
				Computed:    true,
			},
			"query_offset": {
				Type:        schema.TypeString,
				Description: "The duration by which to delay the execution of the recording rule.",
				Computed:    true,
			},
			"evaluation_delay": {
				Type:        schema.TypeString,
				Description: "**Deprecated** The duration by which to delay the execution of the recording rule.",
				Deprecated:  "With Mimir >= 2.13, replaced by query_offset. This attribute will be removed when it is no longer supported in Mimir.",
				Computed:    true,
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
						"record": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"expr": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"labels": {
							Type:        schema.TypeMap,
							Description: "Recording Rule labels",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Computed:    true,
						},
					},
				},
			},
		}, /* End schema */

	}
}

func dataSourcemimirRuleGroupRecordingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	baseMsg := fmt.Sprintf("Cannot read recording rule group '%s' -", name)
	err = handleHTTPError(err, baseMsg)
	if err != nil {
		if strings.Contains(err.Error(), "response code '404'") {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(id)

	var data recordingRuleGroup
	err = yaml.Unmarshal([]byte(jobraw), &data)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to decode recording rule group '%s' data: %v", name, err))
	}
	if err := d.Set("rule", flattenRecordingRules(data.Rules)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("interval", data.Interval); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("query_offset", data.QueryOffset); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("evaluation_delay", data.EvaluationDelay); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("source_tenants", data.SourceTenants); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
