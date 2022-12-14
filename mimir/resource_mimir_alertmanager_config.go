package mimir

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gopkg.in/yaml.v3"
)

func resourcemimirAlertmanagerConfig() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcemimirAlertmanagerConfigCreate,
		ReadContext:   resourcemimirAlertmanagerConfigRead,
		UpdateContext: resourcemimirAlertmanagerConfigUpdate,
		DeleteContext: resourcemimirAlertmanagerConfigDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: resourceMimirAlertmanagerConfigSchemaV1(),
	}
}

func resourcemimirAlertmanagerConfigCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api_client)
	path := "/api/v1/alerts"
	resp, err := alertmanagerConfigCreateUpdate(client, d, path)
	baseMsg := "Cannot create alertmanager config"
	fullurl := fmt.Sprintf("%s%s", client.uri, path)
	err = handleHTTPError(err, resp, fullurl, baseMsg)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(client.headers["X-Scope-OrgID"])
	return resourcemimirAlertmanagerConfigRead(ctx, d, meta)
}

func resourcemimirAlertmanagerConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api_client)
	path := "/api/v1/alerts"
	resp, err := client.send_request("alertmanager", "GET", path, "", make(map[string]string))
	baseMsg := "Cannot read alertmanager config"
	fullurl := fmt.Sprintf("%s%s", client.uri, path)
	err = handleHTTPError(err, resp, fullurl, baseMsg)
	if err != nil {
		if strings.Contains(err.Error(), "response code '404'") {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	var alertmanagerUserConf alertmanagerUserConfig
	yaml.Unmarshal([]byte(resp), &alertmanagerUserConf)

	var alertmanagerConf alertmanagerConfig
	yaml.Unmarshal([]byte(alertmanagerUserConf.AlertmanagerConfig), &alertmanagerConf)

	if alertmanagerConf.Global != nil {
		d.Set("global", flattenGlobalConfig(alertmanagerConf.Global))
	}
	d.Set("time_interval", flattenMuteTimeIntervalConfig(alertmanagerConf.MuteTimeIntervals))
	d.Set("inhibit_rule", flattenInhibitRuleConfig(alertmanagerConf.InhibitRules))
	d.Set("receiver", flattenReceiverConfig(alertmanagerConf.Receivers))
	d.Set("route", flattenRouteConfig(alertmanagerConf.Route))
	d.Set("templates", alertmanagerConf.Templates)
	d.Set("templates_files", alertmanagerUserConf.TemplateFiles)

	return diag.Diagnostics{}
}

func resourcemimirAlertmanagerConfigUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api_client)
	path := "/api/v1/alerts"
	resp, err := alertmanagerConfigCreateUpdate(client, d, path)
	baseMsg := "Cannot update alertmanager config"
	fullurl := fmt.Sprintf("%s%s", client.uri, path)
	err = handleHTTPError(err, resp, fullurl, baseMsg)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourcemimirAlertmanagerConfigRead(ctx, d, meta)
}

func resourcemimirAlertmanagerConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api_client)
	path := "/api/v1/alerts"
	_, err := client.send_request("alertmanager", "DELETE", path, "", make(map[string]string))
	if err != nil {
		return diag.FromErr(fmt.Errorf(
			"Cannot delete alertmanager config from %s: %v",
			fmt.Sprintf("%s%s", client.uri, path),
			err))
	}
	d.SetId("")

	return diag.Diagnostics{}
}

func alertmanagerConfigCreateUpdate(client *api_client, d *schema.ResourceData, path string) (string, error) {
	headers := map[string]string{"Content-Type": "application/yaml"}

	alertmanagerConf := &alertmanagerConfig{
		Global:            expandGlobalConfig(d.Get("global").([]interface{})),
		MuteTimeIntervals: expandMuteTimeIntervalConfig(d.Get("time_interval").([]interface{})),
		InhibitRules:      expandInhibitRuleConfig(d.Get("inhibit_rule").([]interface{})),
		Receivers:         expandReceiverConfig(d.Get("receiver").([]interface{})),
		Route:             expandRouteConfig(d.Get("route").([]interface{})),
		Templates:         expandStringArray(d.Get("templates").([]interface{})),
	}
	alertmanagerConfBytes, _ := yaml.Marshal(&alertmanagerConf)

	alertmanagerUserConf := &alertmanagerUserConfig{
		TemplateFiles:      expandStringMap(d.Get("templates_files").(map[string]interface{})),
		AlertmanagerConfig: string(alertmanagerConfBytes),
	}

	dataBytes, _ := yaml.Marshal(&alertmanagerUserConf)

	resp, err := client.send_request("alertmanager", "POST", path, string(dataBytes), headers)

	return resp, err
}
