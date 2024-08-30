package mimir

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

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
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: resourceMimirAlertmanagerConfigSchemaV1(),
	}
}

func resourcemimirAlertmanagerConfigCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	if !overwriteAlertmanagerConfig {
		alertmanagerConfigExists := true
		resp, err := client.sendRequest("alertmanager", "GET", apiAlertsPath, "", make(map[string]string))
		baseMsg := "Cannot read alertmanager config"
		err = handleHTTPError(err, baseMsg)
		if err != nil {
			if strings.Contains(err.Error(), "response code '404'") {
				alertmanagerConfigExists = false
			} else {
				return diag.FromErr(err)
			}
		}

		// Check if an empty config has been set
		if _, isEmpty := alertmanagerEmptyConfigCheck(d, resp); isEmpty {
			alertmanagerConfigExists = false
		}

		if alertmanagerConfigExists {
			return diag.Errorf("alertmanager config already exists")
		}
	}

	_, err := alertmanagerConfigCreateUpdate(client, d, apiAlertsPath)
	baseMsg := "Cannot create alertmanager config"
	err = handleHTTPError(err, baseMsg)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(client.headers["X-Scope-OrgID"])

	// Retry read as mimir api could return a 404 status code caused by the event change notification propagation.
	// Add delay of <alertmanagerReadDelayAfterChange> * time.Second) between each retry with a <alertmanagerReadRetryAfterChange> max retries.
	for i := 1; i <= alertmanagerReadRetryAfterChange; i++ {
		result := resourcemimirAlertmanagerConfigRead(ctx, d, meta)
		if len(result) > 0 && !result.HasError() {
			log.Printf("[WARN] Alertmanager config previously created not found (%d/3)", i)
			time.Sleep(alertmanagerReadDelayAfterChangeDuration)
			continue
		}
		return result
	}
	return resourcemimirAlertmanagerConfigRead(ctx, d, meta)
}

func resourcemimirAlertmanagerConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	resp, err := alertmanagerConfigRead(meta)
	if err != nil {
		if d.IsNewResource() && strings.Contains(err.Error(), "response code '404'") {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("Alertmanager config not found. You should increase the provider parameter 'alertmanager_read_delay_after_change' (current: %s)", alertmanagerReadDelayAfterChange),
			})
			return diags
		} else if !d.IsNewResource() && strings.Contains(err.Error(), "response code '404'") {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("Alertmanager config (id: %s) not found, removing from state", d.Id()),
			})
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	// Check if an empty config has been set
	if diag, isEmpty := alertmanagerEmptyConfigCheck(d, resp); isEmpty {
		return diag
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

	return diags
}

func resourcemimirAlertmanagerConfigUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)
	_, err := alertmanagerConfigCreateUpdate(client, d, apiAlertsPath)
	baseMsg := "Cannot update alertmanager config"
	err = handleHTTPError(err, baseMsg)
	if err != nil {
		return diag.FromErr(err)
	}
	// Add time delay before read to wait the event change notification propagation to finish
	time.Sleep(alertmanagerReadDelayAfterChangeDuration)
	return resourcemimirAlertmanagerConfigRead(ctx, d, meta)
}

func resourcemimirAlertmanagerConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)
	_, err := client.sendRequest("alertmanager", "DELETE", apiAlertsPath, "", make(map[string]string))
	if err != nil {
		return diag.FromErr(fmt.Errorf(
			"cannot delete alertmanager config from %s: %v",
			fmt.Sprintf("%s%s", client.uri, apiAlertsPath),
			err))
	}

	// Retry read as mimir api could return a 200 status code but the alertmanager config still exist because of the event change notification propagation latency.
	// Add delay of <alertmanagerReadDelayAfterChange> * time.Second) between each retry with a <alertmanagerReadRetryAfterChange> max retries.
	for i := 1; i <= alertmanagerReadRetryAfterChange; i++ {
		_, err := alertmanagerConfigRead(meta)
		if err == nil {
			log.Printf("[WARN] Alertmanager config previously deleted still exist (%d/3)", i)
			time.Sleep(alertmanagerReadDelayAfterChangeDuration)
			continue
		} else if strings.Contains(err.Error(), "response code '404'") {
			break
		}
		return diag.FromErr(err)
	}
	d.SetId("")
	return diag.Diagnostics{}
}

func alertmanagerConfigRead(meta interface{}) (string, error) {
	client := meta.(*apiClient)
	resp, err := client.sendRequest("alertmanager", "GET", apiAlertsPath, "", make(map[string]string))
	baseMsg := "Cannot read alertmanager config"
	return resp, handleHTTPError(err, baseMsg)
}

func alertmanagerConfigCreateUpdate(client *apiClient, d *schema.ResourceData, path string) (string, error) {
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

	resp, err := client.sendRequest("alertmanager", "POST", path, string(dataBytes), headers)

	return resp, err
}

func alertmanagerEmptyConfigCheck(d *schema.ResourceData, data string) (diag.Diagnostics, bool) {
	var isEmpty bool
	var diags diag.Diagnostics
	var alertmanagerUserConf alertmanagerUserConfig
	err := yaml.Unmarshal([]byte(data), &alertmanagerUserConf)
	if err != nil {
		return diag.FromErr(err), isEmpty
	}

	if alertmanagerUserConf.AlertmanagerConfig == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  fmt.Sprintf("Alertmanager config (id: %s) is empty, removing from state", d.Id()),
		})
		d.SetId("")
		isEmpty = true
	}
	return diags, isEmpty
}
