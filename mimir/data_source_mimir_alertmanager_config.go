package mimir

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gopkg.in/yaml.v3"
)

func dataSourcemimirAlertmanagerConfig() *schema.Resource {
	return &schema.Resource{
		Read:   dataSourcemimirAlertmanagerConfigRead,
		Schema: dataSourceMimirAlertmanagerConfigSchemaV1(),
	}
}

func dataSourcemimirAlertmanagerConfigRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api_client)
	name := d.Get("name").(string)
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
		return err
	}

	d.SetId(client.headers["X-Scope-OrgID"])

	if name == "" {
		name = client.headers["X-Scope-OrgID"]
	}

	d.Set("name", name)

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

	return nil
}
