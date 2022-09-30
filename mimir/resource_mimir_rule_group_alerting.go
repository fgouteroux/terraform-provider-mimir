package mimir

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gopkg.in/yaml.v3"
)

func resourcemimirRuleGroupAlerting() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcemimirRuleGroupAlertingCreate,
		ReadContext:   resourcemimirRuleGroupAlertingRead,
		UpdateContext: resourcemimirRuleGroupAlertingUpdate,
		DeleteContext: resourcemimirRuleGroupAlertingDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
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
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"alert": {
							Type:         schema.TypeString,
							Description:  "The name of the alert.",
							Required:     true,
							ValidateFunc: validateAlertingRuleName,
						},
						"expr": {
							Type:         schema.TypeString,
							Description:  "The PromQL expression to evaluate.",
							Required:     true,
							ValidateFunc: validatePromQLExpr,
						},
						"for": {
							Type:         schema.TypeString,
							Description:  "The duration for which the condition must be true before an alert fires.",
							Optional:     true,
							ValidateFunc: validateDuration,
						},
						"annotations": {
							Type:         schema.TypeMap,
							Description:  "Annotations to add to each alert.",
							Optional:     true,
							Elem:         &schema.Schema{Type: schema.TypeString},
							ValidateFunc: validateAnnotations,
						},
						"labels": {
							Type:         schema.TypeMap,
							Description:  "Labels to add or overwrite for each alert.",
							Optional:     true,
							Elem:         &schema.Schema{Type: schema.TypeString},
							ValidateFunc: validateLabels,
						},
					},
				},
			},
		}, /* End schema */
	}
}

func resourcemimirRuleGroupAlertingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api_client)
	name := d.Get("name").(string)
	namespace := d.Get("namespace").(string)

	rules := &alertingRuleGroup{
		Name:  name,
		Rules: expandAlertingRules(d.Get("rule").([]interface{})),
	}
	data, _ := yaml.Marshal(rules)
	headers := map[string]string{"Content-Type": "application/yaml"}

	path := fmt.Sprintf("/config/v1/rules/%s", namespace)
	jobraw, err := client.send_request("ruler", "POST", path, string(data), headers)
	baseMsg := fmt.Sprintf("Cannot create alerting rule group '%s' -", name)
	fullurl := fmt.Sprintf("%s%s", client.uri, path)
	err = handleHTTPError(err, jobraw, fullurl, baseMsg)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("%s/%s", namespace, name))
	return resourcemimirRuleGroupAlertingRead(ctx, d, meta)
}

func resourcemimirRuleGroupAlertingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	err := ruleAlertingRead(ctx, d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.Diagnostics{}
}

func resourcemimirRuleGroupAlertingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.HasChange("rule") {
		client := meta.(*api_client)
		name := d.Get("name").(string)
		namespace := d.Get("namespace").(string)

		rules := &alertingRuleGroup{
			Name:  name,
			Rules: expandAlertingRules(d.Get("rule").([]interface{})),
		}
		data, _ := yaml.Marshal(rules)
		headers := map[string]string{"Content-Type": "application/yaml"}

		path := fmt.Sprintf("/config/v1/rules/%s", namespace)
		jobraw, err := client.send_request("ruler", "POST", path, string(data), headers)
		baseMsg := fmt.Sprintf("Cannot update alerting rule group '%s' -", name)
		fullurl := fmt.Sprintf("%s%s", client.uri, path)
		err = handleHTTPError(err, jobraw, fullurl, baseMsg)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return resourcemimirRuleGroupAlertingRead(ctx, d, meta)
}

func resourcemimirRuleGroupAlertingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api_client)
	name := d.Get("name").(string)
	namespace := d.Get("namespace").(string)
	var headers map[string]string
	path := fmt.Sprintf("/config/v1/rules/%s/%s", namespace, name)
	_, err := client.send_request("ruler", "DELETE", path, "", headers)
	if err != nil {
		return diag.FromErr(fmt.Errorf(
			"Cannot delete alerting rule group '%s' from %s: %v",
			name,
			fmt.Sprintf("%s%s", client.uri, path),
			err))
	}
	d.SetId("")

	return diag.Diagnostics{}
}

func ruleAlertingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api_client)

	// use id as read is also called by import
	id_arr := strings.Split(d.Id(), "/")
	namespace := id_arr[0]
	name := id_arr[1]

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

	var data alertingRuleGroup
	err = yaml.Unmarshal([]byte(jobraw), &data)
	if err != nil {
		return fmt.Errorf("Unable to decode alerting namespace rule group '%s' data: %v", name, err)
	}

	if err := d.Set("rule", flattenAlertingRules(data.Rules)); err != nil {
		return err
	}

	d.Set("namespace", namespace)
	d.Set("name", name)

	return nil
}

func expandAlertingRules(v []interface{}) []alertingRule {
	var rules []alertingRule

	for _, v := range v {
		var rule alertingRule
		data := v.(map[string]interface{})

		if raw, ok := data["alert"]; ok {
			rule.Alert = raw.(string)
		}

		if raw, ok := data["expr"]; ok {
			rule.Expr = raw.(string)
		}

		if raw, ok := data["for"]; ok {
			if raw.(string) != "" {
				rule.For = raw.(string)
			}
		}

		if raw, ok := data["labels"]; ok {
			if len(raw.(map[string]interface{})) > 0 {
				rule.Labels = expandStringMap(raw.(map[string]interface{}))
			}
		}

		if raw, ok := data["annotations"]; ok {
			if len(raw.(map[string]interface{})) > 0 {
				rule.Annotations = expandStringMap(raw.(map[string]interface{}))
			}
		}

		rules = append(rules, rule)
	}

	return rules
}

func flattenAlertingRules(v []alertingRule) []map[string]interface{} {
	var rules []map[string]interface{}

	if v == nil {
		return rules
	}

	for _, v := range v {
		rule := make(map[string]interface{})
		rule["alert"] = v.Alert
		rule["expr"] = v.Expr

		if v.For != "" {
			rule["for"] = v.For
		}
		if v.Labels != nil {
			rule["labels"] = v.Labels
		}
		if v.Annotations != nil {
			rule["annotations"] = v.Annotations
		}

		rules = append(rules, rule)

	}

	return rules
}

func validateAlertingRuleName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if !labelNameRegexp.MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"\"%s\": Invalid Alerting Rule Name %q. Must match the regex %s", k, value, labelNameRegexp))
	}

	return
}

type alertingRule struct {
	Alert       string            `yaml:"alert"`
	Expr        string            `yaml:"expr"`
	For         string            `yaml:"for,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

type alertingRuleGroup struct {
	Name  string         `yaml:"name"`
	Rules []alertingRule `yaml:"rules"`
}
