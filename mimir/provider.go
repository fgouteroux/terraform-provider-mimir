package mimir

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func Provider() *schema.Provider {
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
			"org_id": {
				Type:         schema.TypeString,
				Required:     true,
				DefaultFunc:  schema.EnvDefaultFunc("MIMIR_ORG_ID", nil),
				Description:  "The organization id to operate on within mimir.",
				ValidateFunc: validation.StringIsNotEmpty,
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
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("MIMIR_INSECURE", nil),
				Description: "When using https, this disables TLS verification of the host.",
			},
			"cert": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Client cert for client authentication",
			},
			"key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Client key for client authentication",
			},
			"ca": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Client ca for client authentication",
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
				Default:     60,
				Description: "When set, will cause requests taking longer than this time (in seconds) to be aborted.",
			},
			"debug": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("MIMIR_DEBUG", false),
				Description: "Enable debug mode to trace requests executed.",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"mimir_alertmanager_config":  dataSourcemimirAlertmanagerConfig(),
			"mimir_rule_group_alerting":  dataSourcemimirRuleGroupAlerting(),
			"mimir_rule_group_recording": dataSourcemimirRuleGroupRecording(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"mimir_alertmanager_config":  resourcemimirAlertmanagerConfig(),
			"mimir_rule_group_alerting":  resourcemimirRuleGroupAlerting(),
			"mimir_rule_group_recording": resourcemimirRuleGroupRecording(),
		},
	}
	p.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return providerConfigure(d, p.TerraformVersion)
	}
	return p
}

func providerConfigure(d *schema.ResourceData, terraformVersion string) (interface{}, diag.Diagnostics) {

	headers := make(map[string]string)
	if i_headers := d.Get("headers"); i_headers != nil {
		for k, v := range i_headers.(map[string]interface{}) {
			headers[k] = v.(string)
		}
	}
	headers["X-Scope-OrgID"] = d.Get("org_id").(string)

	opt := &apiClientOpt{
		token:            d.Get("token").(string),
		username:         d.Get("username").(string),
		password:         d.Get("password").(string),
		cert:             d.Get("cert").(string),
		key:              d.Get("key").(string),
		ca:               d.Get("ca").(string),
		insecure:         d.Get("insecure").(bool),
		uri:              d.Get("uri").(string),
		ruler_uri:        d.Get("ruler_uri").(string),
		alertmanager_uri: d.Get("alertmanager_uri").(string),
		headers:          headers,
		timeout:          d.Get("timeout").(int),
		debug:            d.Get("debug").(bool),
	}

	client, err := NewAPIClient(opt)
	return client, diag.FromErr(err)
}
