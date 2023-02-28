package mimir

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gopkg.in/yaml.v3"
)

func dataSourcemimirAlertmanagerConfig() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcemimirAlertmanagerConfigRead,
		Schema:      dataSourceMimirAlertmanagerConfigSchemaV1(),
	}
}

func dataSourcemimirAlertmanagerConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)
	name := d.Get("name").(string)
	resp, err := client.sendRequest("alertmanager", "GET", apiAlertsPath, "", make(map[string]string))
	baseMsg := "Cannot read alertmanager config"
	err = handleHTTPError(err, baseMsg)
	if err != nil {
		if strings.Contains(err.Error(), "response code '404'") {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(client.headers["X-Scope-OrgID"])

	if name == "" {
		name = client.headers["X-Scope-OrgID"]
	}

	if err := d.Set("name", name); err != nil {
		return diag.Errorf("error setting item: %v", err)
	}

	var alertmanagerUserConf alertmanagerUserConfig
	err = yaml.Unmarshal([]byte(resp), &alertmanagerUserConf)
	if err != nil {
		return diag.FromErr(err)
	}

	var alertmanagerConf alertmanagerConfig
	err = yaml.Unmarshal([]byte(alertmanagerUserConf.AlertmanagerConfig), &alertmanagerConf)
	if err != nil {
		return diag.FromErr(err)
	}

	if alertmanagerConf.Global != nil {
		if err := d.Set("global", flattenGlobalConfig(alertmanagerConf.Global)); err != nil {
			return diag.Errorf("error setting item: %v", err)
		}
	}
	if err := d.Set("time_interval", flattenMuteTimeIntervalConfig(alertmanagerConf.MuteTimeIntervals)); err != nil {
		return diag.Errorf("error setting item: %v", err)
	}
	if err := d.Set("inhibit_rule", flattenInhibitRuleConfig(alertmanagerConf.InhibitRules)); err != nil {
		return diag.Errorf("error setting item: %v", err)
	}
	if err := d.Set("receiver", flattenReceiverConfig(alertmanagerConf.Receivers)); err != nil {
		return diag.Errorf("error setting item: %v", err)
	}
	if err := d.Set("route", flattenRouteConfig(alertmanagerConf.Route)); err != nil {
		return diag.Errorf("error setting item: %v", err)
	}
	if err := d.Set("templates", alertmanagerConf.Templates); err != nil {
		return diag.Errorf("error setting item: %v", err)
	}
	if err := d.Set("templates_files", alertmanagerUserConf.TemplateFiles); err != nil {
		return diag.Errorf("error setting item: %v", err)
	}

	return nil
}
