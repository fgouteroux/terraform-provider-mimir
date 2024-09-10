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
			"interval": {
				Type:         schema.TypeString,
				Description:  "Recording Rule group interval",
				Optional:     true,
				ValidateFunc: validateDuration,
			},
			"query_offset": {
				Type:          schema.TypeString,
				Description:   "The duration by which to delay the execution of the recording rule.",
				Optional:      true,
				ConflictsWith: []string{"evaluation_delay"},
				ValidateFunc:  validateDuration,
			},
			"evaluation_delay": {
				Type:          schema.TypeString,
				Description:   "**Deprecated** The duration by which to delay the execution of the recording rule.",
				Optional:      true,
				Deprecated:    "With Mimir >= 2.13, replaced by query_offset. This attribute will be removed in the next major version of this provider.",
				ConflictsWith: []string{"query_offset"},
				ValidateFunc:  validateDuration,
			},
			"source_tenants": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Allows aggregating data from multiple tenants while evaluating a rule group.",
				Elem:        &schema.Schema{Type: schema.TypeString},
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

	if !overwriteRuleGroupConfig {
		ruleGroupConfigExists := true

		path := fmt.Sprintf("/config/v1/rules/%s/%s", namespace, name)
		_, err := client.sendRequest("ruler", "GET", path, "", make(map[string]string))
		baseMsg := fmt.Sprintf("Cannot create recording rule group '%s' (namespace: %s) -", name, namespace)
		err = handleHTTPError(err, baseMsg)
		if err != nil {
			if strings.Contains(err.Error(), "response code '404'") {
				ruleGroupConfigExists = false
			} else {
				return diag.FromErr(err)
			}
		}

		if ruleGroupConfigExists {
			return diag.Errorf("recording rule group '%s' (namespace: %s) already exists", name, namespace)
		}
	}

	rules := &recordingRuleGroup{
		Name:            name,
		Interval:        d.Get("interval").(string),
		EvaluationDelay: d.Get("evaluation_delay").(string),
		QueryOffset:     d.Get("query_offset").(string),
		SourceTenants:   expandStringArray(d.Get("source_tenants").([]interface{})),
		Rules:           expandRecordingRules(d.Get("rule").([]interface{})),
	}
	// if ed, ok := d.GetOk("evaluation_delay"); ok {
	// 	rules.EvaluationDelay = ed.(string)
	// } else {
	// 	rules.QueryOffset = d.Get("query_offset").(string)
	// }
	data, _ := yaml.Marshal(rules)
	headers := map[string]string{"Content-Type": "application/yaml"}

	path := fmt.Sprintf("/config/v1/rules/%s", namespace)
	_, err := client.sendRequest("ruler", "POST", path, string(data), headers)
	baseMsg := fmt.Sprintf("Cannot create recording rule group '%s' (namespace: %s) -", name, namespace)
	err = handleHTTPError(err, baseMsg)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("%s/%s", namespace, name))

	// Retry read as mimir api could return a 404 status code caused by the event change notification propagation.
	// Add delay of <ruleGroupReadDelayAfterChange> * time.Second) between each retry with a <ruleGroupReadRetryAfterChange> max retries.
	for i := 1; i <= ruleGroupReadRetryAfterChange; i++ {
		result := resourcemimirRuleGroupRecordingRead(ctx, d, meta)
		if len(result) > 0 && !result.HasError() {
			log.Printf("[WARN] Recording rule group previously created'%s' not found (%d/3)", name, i)
			time.Sleep(ruleGroupReadDelayAfterChangeDuration)
			continue
		}
		return result
	}
	return resourcemimirRuleGroupRecordingRead(ctx, d, meta)
}

func resourcemimirRuleGroupRecordingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*apiClient)

	// use id as read is also called by import
	idArr := strings.Split(d.Id(), "/")
	namespace := idArr[0]
	name := idArr[1]

	var headers map[string]string
	path := fmt.Sprintf("/config/v1/rules/%s/%s", namespace, name)
	jobraw, err := client.sendRequest("ruler", "GET", path, "", headers)

	baseMsg := fmt.Sprintf("Cannot read recording rule group '%s' (namespace: %s) -", name, namespace)
	err = handleHTTPError(err, baseMsg)
	if err != nil {
		if d.IsNewResource() && strings.Contains(err.Error(), "response code '404'") {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("Recording rule group '%s' (namespace: %s) not found. You should increase the provider parameter 'rule_group_read_delay_after_change' (current: %s)", name, namespace, ruleGroupReadDelayAfterChange),
			})
			return diags
		} else if !d.IsNewResource() && strings.Contains(err.Error(), "response code '404'") {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("Recording rule group '%s' (id: %s) not found, removing from state", name, d.Id()),
			})
			d.SetId("")
			return diags
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
	err = d.Set("interval", data.Interval)
	if err != nil {
		return diag.FromErr(err)
	}
	if _, ok := d.GetOk("evaluation_delay"); ok {
		err = d.Set("evaluation_delay", data.EvaluationDelay)
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		err = d.Set("query_offset", data.QueryOffset)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	err = d.Set("source_tenants", data.SourceTenants)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourcemimirRuleGroupRecordingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.HasChanges("rule", "interval", "query_offset", "evaluation_delay", "source_tenants") {
		client := meta.(*apiClient)
		name := d.Get("name").(string)
		namespace := d.Get("namespace").(string)

		rules := &recordingRuleGroup{
			Name:          name,
			Interval:      d.Get("interval").(string),
			SourceTenants: expandStringArray(d.Get("source_tenants").([]interface{})),
			Rules:         expandRecordingRules(d.Get("rule").([]interface{})),
		}
		if ed, ok := d.GetOk("evaluation_delay"); ok {
			rules.EvaluationDelay = ed.(string)
		} else {
			rules.QueryOffset = d.Get("query_offset").(string)
		}
		data, _ := yaml.Marshal(rules)
		headers := map[string]string{"Content-Type": "application/yaml"}

		path := fmt.Sprintf("/config/v1/rules/%s", namespace)
		_, err := client.sendRequest("ruler", "POST", path, string(data), headers)
		baseMsg := fmt.Sprintf("Cannot update recording rule group '%s' (namespace: %s)  -", name, namespace)
		err = handleHTTPError(err, baseMsg)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	// Add time delay before read to wait the event change notification propagation to finish
	time.Sleep(ruleGroupReadDelayAfterChangeDuration)
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
	Name            string          `yaml:"name"`
	Interval        string          `yaml:"interval,omitempty"`
	QueryOffset     string          `yaml:"query_offset,omitempty"`
	EvaluationDelay string          `yaml:"evaluation_delay,omitempty"`
	Rules           []recordingRule `yaml:"rules"`
	SourceTenants   []string        `yaml:"source_tenants,omitempty"`
}
