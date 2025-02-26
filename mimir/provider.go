package mimir

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	apiAlertsPath                            = "/api/v1/alerts"
	enablePromQLExprFormat                   bool
	overwriteAlertmanagerConfig              bool
	overwriteRuleGroupConfig                 bool
	ruleGroupReadDelayAfterChange            string
	alertmanagerReadDelayAfterChange         string
	ruleGroupReadDelayAfterChangeDuration    time.Duration
	alertmanagerReadDelayAfterChangeDuration time.Duration
	ruleGroupReadRetryAfterChange            int
	alertmanagerReadRetryAfterChange         int
)

func Provider(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"uri": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("MIMIR_URI", nil),
					Description: "mimir base url",
				},
				"ruler_uri": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("MIMIR_RULER_URI", nil),
					Description: "mimir ruler base url",
				},
				"alertmanager_uri": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("MIMIR_ALERTMANAGER_URI", nil),
					Description: "mimir alertmanager base url",
				},
				"distributor_uri": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("MIMIR_DISTRIBUTOR_URI", nil),
					Description: "mimir distributor base url",
				},
				"org_id": {
					Type:        schema.TypeString,
					Required:    true,
					DefaultFunc: schema.EnvDefaultFunc("MIMIR_ORG_ID", nil),
					Description: "The default organization id to operate on within mimir. For resources that have an org_id attribute, the resource-level attribute has priority. May alternatively be set via the MIMIR_ORG_ID environment variable.",
				},
				"token": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("MIMIR_TOKEN", nil),
					Description: "When set, will use this token for Bearer auth to the API.",
				},
				"username": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("MIMIR_USERNAME", nil),
					Description: "When set, will use this username for BASIC auth to the API.",
				},
				"password": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("MIMIR_PASSWORD", nil),
					Description: "When set, will use this password for BASIC auth to the API.",
				},
				"proxy_url": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "URL to the proxy to be used for all API requests",
					DefaultFunc: schema.EnvDefaultFunc("MIMIR_PROXY_URL", nil),
				},
				"insecure": {
					Type:        schema.TypeBool,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("MIMIR_INSECURE", nil),
					Description: "When using https, this disables TLS verification of the host.",
				},
				"cert": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("MIMIR_TLS_CERT", nil),
					Description: "Client cert (filepath or inline) for TLS client authentication.",
				},
				"key": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("MIMIR_TLS_KEY", nil),
					Description: "Client key (filepath or inline) for TLS client authentication.",
				},
				"ca": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("MIMIR_TLS_CA", nil),
					Description: "Client ca (filepath or inline) for TLS client authentication.",
				},
				"headers": {
					Type:        schema.TypeMap,
					Elem:        schema.TypeString,
					Optional:    true,
					Description: "A map of header names and values to set on all outbound requests.",
				},
				"timeout": {
					Type:        schema.TypeInt,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("MIMIR_TIMEOUT", 60),
					Description: "When set, will cause requests taking longer than this time (in seconds) to be aborted.",
				},
				"debug": {
					Type:        schema.TypeBool,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("MIMIR_DEBUG", false),
					Description: "Enable debug mode to trace requests executed.",
				},
				"format_promql_expr": {
					Type:        schema.TypeBool,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("MIMIR_FORMAT_PROMQL_EXPR", false),
					Description: "Enable the formatting of PromQL expression.",
				},
				"overwrite_alertmanager_config": {
					Type:        schema.TypeBool,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("MIMIR_OVERWRITE_ALERTMANAGER_CONFIG", false),
					Description: "Overwrite the current alertmanager config on create.",
				},
				"overwrite_rule_group_config": {
					Type:        schema.TypeBool,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("MIMIR_OVERWRITE_RULE_GROUP_CONFIG", false),
					Description: "Overwrite the current rule group (alerting/recording) config on create.",
				},
				"rule_group_read_delay_after_change": {
					Type:         schema.TypeString,
					Optional:     true,
					DefaultFunc:  schema.EnvDefaultFunc("MIMIR_RULE_GROUP_READ_DELAY_AFTER_CHANGE", "1s"),
					Description:  "When set, add a delay (time duration) to read the rule group after a change.",
					ValidateFunc: validateDuration,
				},
				"rule_group_read_retry_after_change": {
					Type:        schema.TypeInt,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("MIMIR_RULE_GROUP_READ_RETRY_AFTER_CHANGE", 3),
					Description: "Max retries to read the rule group after a change.",
				},
				"alertmanager_read_delay_after_change": {
					Type:         schema.TypeString,
					Optional:     true,
					DefaultFunc:  schema.EnvDefaultFunc("MIMIR_ALERTMANAGER_READ_DELAY_AFTER_CHANGE", "1s"),
					Description:  "When set, add a delay (time duration) to read the alertmanager config after a change.",
					ValidateFunc: validateDuration,
				},
				"alertmanager_read_retry_after_change": {
					Type:        schema.TypeInt,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("MIMIR_ALERTMANAGER_READ_RETRY_AFTER_CHANGE", 3),
					Description: "Max retries to read the alertmanager config after a change.",
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"mimir_alertmanager_config":      dataSourcemimirAlertmanagerConfig(),
				"mimir_rule_group_alerting":      dataSourcemimirRuleGroupAlerting(),
				"mimir_rule_group_recording":     dataSourcemimirRuleGroupRecording(),
				"mimir_distributor_tenant_stats": dataSourcemimirDistributorTenantStats(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"mimir_alertmanager_config":  resourcemimirAlertmanagerConfig(),
				"mimir_rule_group_alerting":  resourcemimirRuleGroupAlerting(),
				"mimir_rule_group_recording": resourcemimirRuleGroupRecording(),
			},
		}
		p.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
			p.UserAgent("terraform-provider-mimir", version)
			return providerConfigure(version, p, d)
		}
		return p
	}
}

func providerConfigure(version string, p *schema.Provider, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	headers := make(map[string]string)
	if initHeaders := d.Get("headers"); initHeaders != nil {
		for k, v := range initHeaders.(map[string]interface{}) {
			headers[k] = v.(string)
		}
	}
	orgID := d.Get("org_id").(string)
	if orgID != "" {
		headers["X-Scope-OrgID"] = orgID
	}
	headers["User-Agent"] = p.UserAgent("terraform-provider-mimir", version)

	opt := &apiClientOpt{
		token:           d.Get("token").(string),
		username:        d.Get("username").(string),
		password:        d.Get("password").(string),
		proxyURL:        d.Get("proxy_url").(string),
		cert:            d.Get("cert").(string),
		key:             d.Get("key").(string),
		ca:              d.Get("ca").(string),
		insecure:        d.Get("insecure").(bool),
		uri:             d.Get("uri").(string),
		rulerURI:        d.Get("ruler_uri").(string),
		alertmanagerURI: d.Get("alertmanager_uri").(string),
		distributorURI:  d.Get("distributor_uri").(string),
		headers:         headers,
		timeout:         d.Get("timeout").(int),
		debug:           d.Get("debug").(bool),
	}

	enablePromQLExprFormat = d.Get("format_promql_expr").(bool)
	overwriteAlertmanagerConfig = d.Get("overwrite_alertmanager_config").(bool)
	overwriteRuleGroupConfig = d.Get("overwrite_rule_group_config").(bool)
	ruleGroupReadDelayAfterChange = d.Get("rule_group_read_delay_after_change").(string)
	alertmanagerReadDelayAfterChange = d.Get("alertmanager_read_delay_after_change").(string)
	ruleGroupReadDelayAfterChangeDuration, _ = time.ParseDuration(d.Get("rule_group_read_delay_after_change").(string))
	alertmanagerReadDelayAfterChangeDuration, _ = time.ParseDuration(d.Get("alertmanager_read_delay_after_change").(string))
	ruleGroupReadRetryAfterChange = d.Get("rule_group_read_retry_after_change").(int)
	alertmanagerReadRetryAfterChange = d.Get("alertmanager_read_retry_after_change").(int)

	client, err := NewAPIClient(opt)
	return client, diag.FromErr(err)
}
