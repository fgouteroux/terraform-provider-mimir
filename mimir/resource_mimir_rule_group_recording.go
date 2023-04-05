package mimir

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gopkg.in/yaml.v3"
)

func resourcemimirRuleGroupRecording() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcemimirRuleGroupRecordingCreate,
		ReadContext:   resourcemimirRuleGroupRecordingRead,
		UpdateContext: resourcemimirRuleGroupRecordingUpdate,
		DeleteContext: resourcemimirRuleGroupRecordingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
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
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"record": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The name of the time series to output to.",
							ValidateFunc: validateRecordingRuleName,
						},
						"expr": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "The PromQL expression to evaluate.",
							ValidateFunc: validatePromQLExpr,
							StateFunc:    formatPromQLExpr,
						},
						"labels": {
							Type:         schema.TypeMap,
							Description:  "Labels to add or overwrite before storing the result.",
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

func resourcemimirRuleGroupRecordingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)
	name := d.Get("name").(string)
	namespace := d.Get("namespace").(string)

	rules := &recordingRuleGroup{
		Name:  name,
		Rules: expandRecordingRules(d.Get("rule").([]interface{})),
	}
	data, _ := yaml.Marshal(rules)
	headers := map[string]string{"Content-Type": "application/yaml"}

	path := fmt.Sprintf("/config/v1/rules/%s", namespace)
	_, err := client.sendRequest("ruler", "POST", path, string(data), headers)
	baseMsg := fmt.Sprintf("Cannot create recording rule group '%s' -", name)
	err = handleHTTPError(err, baseMsg)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("%s/%s", namespace, name))
	return resourcemimirRuleGroupRecordingRead(ctx, d, meta)
}

func resourcemimirRuleGroupRecordingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)

	// use id as read is also called by import
	idArr := strings.Split(d.Id(), "/")
	namespace := idArr[0]
	name := idArr[1]

	var headers map[string]string
	path := fmt.Sprintf("/config/v1/rules/%s/%s", namespace, name)
	jobraw, err := client.sendRequest("ruler", "GET", path, "", headers)

	baseMsg := fmt.Sprintf("Cannot read recording rule group '%s' -", name)
	err = handleHTTPError(err, baseMsg)
	if err != nil {
		if strings.Contains(err.Error(), "response code '404'") {
			d.SetId("")
			return diag.Diagnostics{}
		}
		return diag.FromErr(err)
	}

	var data recordingRuleGroup
	err = yaml.Unmarshal([]byte(jobraw), &data)
	if err != nil {
		return diag.FromErr(fmt.Errorf("unable to decode recording namespace rule group '%s' data: %v", name, err))
	}

	if err := d.Set("rule", flattenRecordingRules(data.Rules)); err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("namespace", namespace)
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("name", name)
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

func resourcemimirRuleGroupRecordingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.HasChange("rule") {
		client := meta.(*apiClient)
		name := d.Get("name").(string)
		namespace := d.Get("namespace").(string)

		rules := &recordingRuleGroup{
			Name:  name,
			Rules: expandRecordingRules(d.Get("rule").([]interface{})),
		}
		data, _ := yaml.Marshal(rules)
		headers := map[string]string{"Content-Type": "application/yaml"}

		path := fmt.Sprintf("/config/v1/rules/%s", namespace)
		_, err := client.sendRequest("ruler", "POST", path, string(data), headers)
		baseMsg := fmt.Sprintf("Cannot update recording rule group '%s' -", name)
		err = handleHTTPError(err, baseMsg)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return resourcemimirRuleGroupRecordingRead(ctx, d, meta)
}

func resourcemimirRuleGroupRecordingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*apiClient)
	name := d.Get("name").(string)
	namespace := d.Get("namespace").(string)
	var headers map[string]string
	path := fmt.Sprintf("/config/v1/rules/%s/%s", namespace, name)
	_, err := client.sendRequest("ruler", "DELETE", path, "", headers)
	if err != nil {
		return diag.FromErr(fmt.Errorf(
			"cannot delete recording rule group '%s' from %s: %v",
			name,
			fmt.Sprintf("%s%s", client.uri, path),
			err))
	}
	d.SetId("")

	return diag.Diagnostics{}
}

func expandRecordingRules(v []interface{}) []recordingRule {
	var rules []recordingRule

	for _, v := range v {
		var rule recordingRule
		data := v.(map[string]interface{})

		if raw, ok := data["record"]; ok {
			rule.Record = raw.(string)
		}

		if raw, ok := data["expr"]; ok {
			rule.Expr = formatPromQLExpr(raw)
		}

		if raw, ok := data["labels"]; ok {
			if len(raw.(map[string]interface{})) > 0 {
				rule.Labels = expandStringMap(raw.(map[string]interface{}))
			}
		}

		rules = append(rules, rule)
	}

	return rules
}

func flattenRecordingRules(v []recordingRule) []map[string]interface{} {
	var rules []map[string]interface{}

	if v == nil {
		return rules
	}

	for _, v := range v {
		rule := make(map[string]interface{})
		rule["record"] = v.Record
		rule["expr"] = formatPromQLExpr(v.Expr)

		if v.Labels != nil {
			rule["labels"] = v.Labels
		}

		rules = append(rules, rule)
	}

	return rules
}

func validateRecordingRuleName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if !metricNameRegexp.MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"\"%s\": Invalid Recording Rule Name %q. Must match the regex %s", k, value, metricNameRegexp))
	}

	return
}

type recordingRule struct {
	Record string            `json:"record"`
	Expr   string            `json:"expr"`
	Labels map[string]string `yaml:"labels,omitempty"`
}

type recordingRuleGroup struct {
	Name  string          `yaml:"name"`
	Rules []recordingRule `yaml:"rules"`
}
