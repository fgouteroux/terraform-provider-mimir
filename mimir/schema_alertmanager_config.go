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

func emailConfigFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
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
	}
}

func pagerdutyConfigFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
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
		"images": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"src": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"alt": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"href": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
		"links": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"text": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"href": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
	}
}

func weChatConfigFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"send_resolved": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"api_secret": {
			Type:      schema.TypeString,
			Optional:  true,
			Sensitive: true,
		},
		"api_url": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"corp_id": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"agent_id": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"to_user": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"to_party": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"to_tag": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"message": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"message_type": {
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
	}
}

func webhookConfigFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"send_resolved": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"url": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"max_alerts": {
			Type:     schema.TypeInt,
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
	}
}

func pushoverConfigFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"send_resolved": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"http_config": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: httpConfigFields(),
			},
		},
		"user_key": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"token": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"message": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"url": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"url_title": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"sound": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"priority": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"retry": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"expire": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"html": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
	}
}

func opsgenieConfigFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"send_resolved": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
		"http_config": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: httpConfigFields(),
			},
		},
		"details": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"api_key": {
			Type:      schema.TypeString,
			Optional:  true,
			Sensitive: true,
		},
		"api_url": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"message": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"description": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"source": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"responders": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"name": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"username": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"type": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
		"tags": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"note": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"priority": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"update_alerts": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"entity": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"actions": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
}

func slackConfigFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"send_resolved": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
		"http_config": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: httpConfigFields(),
			},
		},
		"api_url": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"channel": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"username": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"color": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_link": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"pretext": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"text": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"short_fields": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"footer": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"fallback": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"callback_id": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"icon_emoji": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"icon_url": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"image_url": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"thumb_url": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"link_names": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"mrkdwn_in": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"fields": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"title": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"value": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"short": {
						Type:     schema.TypeBool,
						Optional: true,
						Default:  false,
					},
				},
			},
		},
		"actions": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"text": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"url": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"style": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"name": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"value": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"confirm": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"text": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"title": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"ok_text": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"dismiss_text": {
									Type:     schema.TypeString,
									Optional: true,
								},
							},
						},
					},
				},
			},
		},
	}
}

func telegramConfigFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"send_resolved": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"api_url": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"bot_token": {
			Type:      schema.TypeString,
			Optional:  true,
			Sensitive: true,
		},
		"chat_id": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"message": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"disable_notifications": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		"parse_mode": {
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
	}
}

func victorOpsConfigFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"send_resolved": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
		"http_config": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: httpConfigFields(),
			},
		},
		"custom_fields": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"api_key": {
			Type:      schema.TypeString,
			Optional:  true,
			Sensitive: true,
		},
		"api_url": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"routing_key": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"message_type": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"state_message": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"entity_display_name": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"monitoring_tool": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
}

func snsConfigFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"send_resolved": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
		"http_config": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: httpConfigFields(),
			},
		},
		"attributes": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"api_url": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"topic_arn": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"phone_number": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"target_arn": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"subject": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"message": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"sigv4": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"region": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"access_key": {
						Type:      schema.TypeString,
						Optional:  true,
						Sensitive: true,
					},
					"secret_key": {
						Type:      schema.TypeString,
						Optional:  true,
						Sensitive: true,
					},
					"profile": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"role_arn": {
						Type:     schema.TypeString,
						Optional: true,
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
					"opsgenie_api_url": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"wechat_api_url": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"victorops_api_url": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"telegram_api_url": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"slack_api_url": {
						Type:      schema.TypeString,
						Optional:  true,
						Sensitive: true,
					},
					"opsgenie_api_key": {
						Type:      schema.TypeString,
						Optional:  true,
						Sensitive: true,
					},
					"wechat_api_secret": {
						Type:      schema.TypeString,
						Optional:  true,
						Sensitive: true,
					},
					"wechat_api_corp_id": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"victorops_api_key": {
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
							Schema: pagerdutyConfigFields(),
						},
					},
					"email_configs": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: emailConfigFields(),
						},
					},
					"wechat_configs": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: weChatConfigFields(),
						},
					},
					"webhook_configs": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: webhookConfigFields(),
						},
					},
					"pushover_configs": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: pushoverConfigFields(),
						},
					},
					"opsgenie_configs": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: opsgenieConfigFields(),
						},
					},
					"slack_configs": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: slackConfigFields(),
						},
					},
					"telegram_configs": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: telegramConfigFields(),
						},
					},
					"victorops_configs": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: victorOpsConfigFields(),
						},
					},
					"sns_configs": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: snsConfigFields(),
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
