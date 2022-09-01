package mimir

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func tlsConfigFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"server_name": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"insecure_skip_verify": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
	}
}

func httpConfigFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"bearer_token": {
			Type:      schema.TypeString,
			Optional:  true,
			Sensitive: true,
		},
		"proxy_url": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"follow_redirects": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
		"tls_config": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: tlsConfigFields(),
			},
		},
		"authorization": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"credentials": {
						Type:      schema.TypeString,
						Optional:  true,
						Sensitive: true,
					},
				},
			},
		},
		"basic_auth": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"username": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"password": {
						Type:      schema.TypeString,
						Optional:  true,
						Sensitive: true,
					},
				},
			},
		},
		"oauth2": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"client_id": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"client_secret": {
						Type:      schema.TypeString,
						Optional:  true,
						Sensitive: true,
					},
					"tls_config": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: tlsConfigFields(),
						},
					},
					"token_url": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"endpoint_params": {
						Type:     schema.TypeMap,
						Optional: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
					"scopes": {
						Type:     schema.TypeList,
						Optional: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
				},
			},
		},
	}
}

func resourceMimirAlertmanagerConfigSchemaV1() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"global": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"resolve_timeout": {
						Type:        schema.TypeString,
						Description: "The time after which an alert is declared resolved if it has not been updated.",
						Optional:    true,
						Default:     "5m",
					},
					"http_config": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: httpConfigFields(),
						},
					},
					"pagerduty_url": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"slack_api_url": {
						Type:      schema.TypeString,
						Optional:  true,
						Sensitive: true,
					},
					"smtp_from": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"smtp_hello": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"smtp_smarthost": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"smtp_auth_username": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"smtp_auth_password": {
						Type:      schema.TypeString,
						Optional:  true,
						Sensitive: true,
					},
					"smtp_auth_secret": {
						Type:      schema.TypeString,
						Optional:  true,
						Sensitive: true,
					},
					"smtp_auth_identity": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"smtp_require_tls": {
						Type:     schema.TypeBool,
						Optional: true,
						Default:  false,
					},
				},
			},
		},
		"inhibit_rule": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"source_matchers": {
						Type:     schema.TypeList,
						Optional: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
					"target_matchers": {
						Type:     schema.TypeList,
						Optional: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
					"equal": {
						Type:     schema.TypeList,
						Optional: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
				},
			},
		},
		"time_interval": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"time_intervals": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"times": {
									Type:     schema.TypeList,
									Optional: true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"start_minute": {
												Type:     schema.TypeInt,
												Optional: true,
											},
											"end_minute": {
												Type:     schema.TypeInt,
												Optional: true,
											},
										},
									},
								},
								"weekdays": {
									Type:     schema.TypeList,
									Optional: true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"begin": {
												Type:     schema.TypeInt,
												Optional: true,
											},
											"end": {
												Type:     schema.TypeInt,
												Optional: true,
											},
										},
									},
								},
								"days_of_month": {
									Type:     schema.TypeList,
									Optional: true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"begin": {
												Type:     schema.TypeInt,
												Optional: true,
											},
											"end": {
												Type:     schema.TypeInt,
												Optional: true,
											},
										},
									},
								},
								"months": {
									Type:     schema.TypeList,
									Optional: true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"begin": {
												Type:     schema.TypeInt,
												Optional: true,
											},
											"end": {
												Type:     schema.TypeInt,
												Optional: true,
											},
										},
									},
								},
								"years": {
									Type:     schema.TypeList,
									Optional: true,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"begin": {
												Type:     schema.TypeInt,
												Optional: true,
											},
											"end": {
												Type:     schema.TypeInt,
												Optional: true,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		"receiver": {
			Type:     schema.TypeList,
			Required: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:        schema.TypeString,
						Description: "The time after which an alert is declared resolved if it has not been updated.",
						Required:    true,
					},
					"pagerduty_configs": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"send_resolved": {
									Type:     schema.TypeBool,
									Optional: true,
									Default:  true,
								},
								"service_key": {
									Type:      schema.TypeString,
									Optional:  true,
									Sensitive: true,
								},
								"routing_key": {
									Type:      schema.TypeString,
									Optional:  true,
									Sensitive: true,
								},
								"url": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"client": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"client_url": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"description": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"details": {
									Type:     schema.TypeMap,
									Optional: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
								"severity": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"class": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"component": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"group": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"http_config": {
									Type:     schema.TypeList,
									Optional: true,
									MaxItems: 1,
									Elem: &schema.Resource{
										Schema: httpConfigFields(),
									},
								},
							},
						},
					},
					"email_configs": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"send_resolved": {
									Type:     schema.TypeBool,
									Optional: true,
									Default:  true,
								},
								"to": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"from": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"hello": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"smarthost": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"auth_username": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"auth_password": {
									Type:      schema.TypeString,
									Optional:  true,
									Sensitive: true,
								},
								"auth_secret": {
									Type:      schema.TypeString,
									Optional:  true,
									Sensitive: true,
								},
								"auth_identity": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"headers": {
									Type:     schema.TypeMap,
									Optional: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
								"html": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"text": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"require_tls": {
									Type:     schema.TypeBool,
									Optional: true,
									Default:  false,
								},
								"tls_config": {
									Type:     schema.TypeList,
									Optional: true,
									MaxItems: 1,
									Elem: &schema.Resource{
										Schema: tlsConfigFields(),
									},
								},
							},
						},
					},
				},
			},
		},
		"route": {
			Type:     schema.TypeList,
			Required: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"group_by": {
						Type:     schema.TypeList,
						Optional: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
					},
					"group_wait": {
						Type:     schema.TypeString,
						Required: true,
					},
					"group_interval": {
						Type:     schema.TypeString,
						Required: true,
					},
					"repeat_interval": {
						Type:     schema.TypeString,
						Required: true,
					},
					"receiver": {
						Type:     schema.TypeString,
						Required: true,
					},
					"continue": {
						Type:     schema.TypeBool,
						Optional: true,
						Default:  false,
					},
					"child_route": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"group_by": {
									Type:     schema.TypeList,
									Optional: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
								"matchers": {
									Type:     schema.TypeList,
									Optional: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
								"group_wait": {
									Type:     schema.TypeString,
									Required: true,
								},
								"group_interval": {
									Type:     schema.TypeString,
									Required: true,
								},
								"repeat_interval": {
									Type:     schema.TypeString,
									Required: true,
								},
								"receiver": {
									Type:     schema.TypeString,
									Required: true,
								},
								"continue": {
									Type:     schema.TypeBool,
									Optional: true,
									Default:  false,
								},
								"mute_time_intervals": {
									Type:     schema.TypeList,
									Optional: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
								"active_time_intervals": {
									Type:     schema.TypeList,
									Optional: true,
									Elem:     &schema.Schema{Type: schema.TypeString},
								},
							},
						},
					},
				},
			},
		},
		"templates": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"templates_files": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
	}
}
