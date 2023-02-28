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
					},
				},
			},
		}, /* End schema */

	}
}

func dataSourcemimirRuleGroupRecordingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api_client)
	name := d.Get("name").(string)
	namespace := d.Get("namespace").(string)

	var headers map[string]string
	path := fmt.Sprintf("/config/v1/rules/%s/%s", namespace, name)
	jobraw, err := client.send_request("ruler", "GET", path, "", headers)

	baseMsg := fmt.Sprintf("Cannot read recording rule group '%s' -", name)
	fullurl := fmt.Sprintf("%s%s", client.uri, path)
	err = handleHTTPError(err, jobraw, fullurl, baseMsg)
	if err != nil {
		if strings.Contains(err.Error(), "response code '404'") {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s/%s", namespace, name))

	var data recordingRuleGroup
	err = yaml.Unmarshal([]byte(jobraw), &data)
	if err != nil {
		return diag.FromErr(fmt.Errorf("Unable to decode recording rule group '%s' data: %v", name, err))
	}
	if err := d.Set("rule", flattenRecordingRules(data.Rules)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
