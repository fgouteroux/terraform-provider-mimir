package mimir

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gopkg.in/yaml.v3"
)

func dataSourcemimirRuleGroupAlerting() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcemimirRuleGroupAlertingRead,

		Schema: map[string]*schema.Schema{
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

func dataSourcemimirRuleGroupAlertingRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api_client)
	name := d.Get("name").(string)
	namespace := d.Get("namespace").(string)

	var headers map[string]string
	path := fmt.Sprintf("/config/v1/rules/%s/%s", namespace, name)
	jobraw, err := client.send_request("ruler", "GET", path, "", headers)

	baseMsg := fmt.Sprintf("Cannot read alerting rule group '%s' -", name)
	fullurl := fmt.Sprintf("%s%s", client.uri, path)
	err = handleHTTPError(err, jobraw, fullurl, baseMsg)
	if err != nil {
		if strings.Contains(err.Error(), "response code '404'") {
			d.SetId("")
			return nil
		}
		return err
	}

	d.SetId(fmt.Sprintf("%s/%s", namespace, name))

	var data alertingRuleGroup
	err = yaml.Unmarshal([]byte(jobraw), &data)
	if err != nil {
		return fmt.Errorf("Unable to decode alerting rule group '%s' data: %v", name, err)
	}
	if err := d.Set("rule", flattenAlertingRules(data.Rules)); err != nil {
		return err
	}

	return nil
}
