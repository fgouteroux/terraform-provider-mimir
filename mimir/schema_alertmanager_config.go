package mimir

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func tlsConfigFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"server_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "ServerName extension to indicate the name of the server.",
		},
		"insecure_skip_verify": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Disable validation of the server certificate",
		},
		"min_version": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Minimum acceptable TLS version",
		},
		"max_version": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Maximum acceptable TLS version.",
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
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Configure whether HTTP requests follow HTTP 3xx redirects.",
		},
		"enable_http2": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Whether to enable HTTP2.",
		},
		"tls_config": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Configures the TLS settings.",
			Elem: &schema.Resource{
				Schema: tlsConfigFields(),
			},
		},
		"authorization": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Set the `Authorization` header configuration.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Sets the authentication type.",
					},
					"credentials": {
						Type:        schema.TypeString,
						Optional:    true,
						Sensitive:   true,
						Description: "Sets the credentials.",
					},
				},
			},
		},
		"basic_auth": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Sets the `Authorization` header with the configured username and password.",
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
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Set the OAuth 2.0 configuration.",
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
						Type:        schema.TypeList,
						Optional:    true,
						MaxItems:    1,
						Description: "Configures the TLS settings.",
						Elem: &schema.Resource{
							Schema: tlsConfigFields(),
						},
					},
					"token_url": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The URL to fetch the token from.",
					},
					"endpoint_params": {
						Type:        schema.TypeMap,
						Optional:    true,
						Description: "Parameters to append to the token URL.",
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"scopes": {
						Type:        schema.TypeList,
						Optional:    true,
						Description: "Scopes for the token request.",
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
				},
			},
		},
	}
}

func emailConfigFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"send_resolved": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Whether to notify about resolved alerts.",
		},
		"to": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The email address to send notifications to.",
		},
		"from": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The sender's address.",
		},
		"hello": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The hostname to identify to the SMTP server.",
		},
		"smarthost": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The SMTP host through which emails are sent.",
		},
		"auth_username": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "SMTP authentication username.",
		},
		"auth_password": {
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   true,
			Description: "SMTP authentication password.",
		},
		"auth_secret": {
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   true,
			Description: "SMTP authentication secret.",
		},
		"auth_identity": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "SMTP authentication identity.",
		},
		"headers": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "Further headers email header key/value pairs. Overrides any headers previously set by the notification implementation.",
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"html": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The HTML body of the email notification.",
		},
		"text": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The text body of the email notification.",
		},
		"require_tls": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "The SMTP TLS requirement.",
		},
		"tls_config": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The SMTP TLS configuration.",
			Elem: &schema.Resource{
				Schema: tlsConfigFields(),
			},
		},
	}
}

func pagerdutyConfigFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"send_resolved": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Whether to notify about resolved alerts.",
		},
		"service_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   true,
			Description: "The PagerDuty integration key (when using PagerDuty integration type `Prometheus`).",
		},
		"routing_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   true,
			Description: "The PagerDuty integration key (when using PagerDuty integration type `Events API v2`).",
		},
		"url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The URL to send API requests to",
		},
		"client": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The client identification of the Alertmanager.",
		},
		"client_url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "A backlink to the sender of the notification.",
		},
		"description": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "A description of the incident.",
		},
		"details": {
			Type:        schema.TypeMap,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "A set of arbitrary key/value pairs that provide further detail about the incident.",
		},
		"severity": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Severity of the incident.",
		},
		"class": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The class/type of the event.",
		},
		"component": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The part or component of the affected system that is broken.",
		},
		"group": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "A cluster or grouping of sources.",
		},
		"http_config": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The HTTP client's configuration.",
			Elem: &schema.Resource{
				Schema: httpConfigFields(),
			},
		},
		"images": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Images to attach to the incident.",
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
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Links to attach to the incident.",
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
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Whether to notify about resolved alerts.",
		},
		"api_secret": {
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   true,
			Description: "The API key to use when talking to the WeChat API.",
		},
		"api_url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The WeChat API URL.",
		},
		"corp_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The corp id for authentication.",
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
			Type:        schema.TypeString,
			Optional:    true,
			Description: "API request data as defined by the WeChat API.",
		},
		"message_type": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Type of the message type, supported values are `text` and `markdown`.",
		},
		"http_config": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: httpConfigFields(),
			},
			Description: "The HTTP client's configuration.",
		},
	}
}

func webhookConfigFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"send_resolved": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Whether to notify about resolved alerts.",
		},
		"url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The endpoint to send HTTP POST requests to.",
		},
		"max_alerts": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "The maximum number of alerts to include in a single webhook message. Alerts above this threshold are truncated. When leaving this at its default value of 0, all alerts are included.",
		},
		"http_config": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: httpConfigFields(),
			},
			Description: "The HTTP client's configuration.",
		},
	}
}

func webexConfigFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"send_resolved": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Whether to notify about resolved alerts.",
		},
		"api_url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The Webex Teams API URL.",
		},
		"room_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "ID of the Webex Teams room where to send the messages.",
		},
		"message": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Message template.",
		},
		"http_config": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: httpConfigFields(),
			},
			Description: "The HTTP client's configuration.",
		},
	}
}

func discordConfigFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"send_resolved": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Whether to notify about resolved alerts.",
		},
		"webhook_url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The webhook URL.",
		},
		"title": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Notification title.",
		},
		"message": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Notification message.",
		},
		"http_config": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: httpConfigFields(),
			},
			Description: "The HTTP client's configuration.",
		},
	}
}

func pushoverConfigFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"send_resolved": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Whether to notify about resolved alerts.",
		},
		"http_config": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: httpConfigFields(),
			},
			Description: "The HTTP client's configuration.",
		},
		"user_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The recipient user's user key.",
		},
		"token": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The registered application's API token.",
		},
		"title": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Notification title.",
		},
		"message": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Notification message.",
		},
		"url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: " A supplementary URL shown alongside the message.",
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
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "How often the Pushover servers will send the same notification to the user.",
			ValidateFunc: validateDuration,
			StateFunc:    formatDuration,
		},
		"expire": {
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "How long your notification will continue to be retried for, unless the user acknowledges the notification.",
			ValidateFunc: validateDuration,
			StateFunc:    formatDuration,
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
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Whether to notify about resolved alerts.",
		},
		"http_config": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: httpConfigFields(),
			},
			Description: "The HTTP client's configuration.",
		},
		"details": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "A set of arbitrary key/value pairs that provide further detail about the alert. All common labels are included as details by default.",
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"api_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   true,
			Description: "The API key to use when talking to the OpsGenie API.",
		},
		"api_url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The host to send OpsGenie API requests to.",
		},
		"message": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Alert text limited to 130 characters.",
		},
		"description": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "A description of the alert.",
		},
		"source": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "A backlink to the sender of the notification.",
		},
		"responders": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "List of responders responsible for notifications.",
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
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Comma separated list of tags attached to the notifications.",
		},
		"note": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Additional alert note.",
		},
		"priority": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Priority level of alert. Possible values are P1, P2, P3, P4, and P5.",
		},
		"update_alerts": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Whether to update message and description of the alert in OpsGenie if it already exists. By default, the alert is never updated in OpsGenie, the new message only appears in activity log.",
		},
		"entity": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Optional field that can be used to specify which domain alert is related to.",
		},
		"actions": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Comma separated list of actions that will be available for the alert.",
		},
	}
}

func slackConfigFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"send_resolved": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Whether to notify about resolved alerts.",
		},
		"http_config": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: httpConfigFields(),
			},
			Description: "The HTTP client's configuration.",
		},
		"api_url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The Slack webhook URL. Defaults to global settings if none are set here.",
		},
		"channel": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The channel or user to send notifications to.",
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
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Whether to notify about resolved alerts.",
		},
		"api_url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The Telegram API URL. If not specified, default API URL will be used.",
		},
		"bot_token": {
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   true,
			Description: "Telegram bot token",
		},
		"chat_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "ID of the chat where to send the messages.",
		},
		"message": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Message template",
		},
		"disable_notifications": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Disable telegram notifications",
		},
		"parse_mode": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Parse mode for telegram message, supported values are MarkdownV2, Markdown, HTML and empty string for plain text.",
		},
		"http_config": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: httpConfigFields(),
			},
			Description: "The HTTP client's configuration.",
		},
	}
}

func victorOpsConfigFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"send_resolved": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Whether to notify about resolved alerts.",
		},
		"http_config": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: httpConfigFields(),
			},
			Description: "The HTTP client's configuration.",
		},
		"custom_fields": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"api_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Sensitive:   true,
			Description: "The API key to use when talking to the VictorOps API.",
		},
		"api_url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The VictorOps API URL.",
		},
		"routing_key": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "A key used to map the alert to a team.",
		},
		"message_type": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Describes the behavior of the alert (CRITICAL, WARNING, INFO).",
		},
		"state_message": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Contains long explanation of the alerted problem.",
		},
		"entity_display_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Contains summary of the alerted problem.",
		},
		"monitoring_tool": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The monitoring tool the state message is from.",
		},
	}
}

func snsConfigFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"send_resolved": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Whether to notify about resolved alerts.",
		},
		"http_config": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: httpConfigFields(),
			},
			Description: "The HTTP client's configuration.",
		},
		"attributes": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "SNS message attributes.",
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"api_url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The SNS API URL. If not specified, the SNS API URL from the SNS SDK will be used.",
		},
		"topic_arn": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "SNS topic ARN. If not set, a value for the phone_number or target_arn should be set.",
		},
		"phone_number": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Phone number if message is delivered via SMS in E.164 format.",
		},
		"target_arn": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The mobile platform endpoint ARN if message is delivered via mobile notifications.",
		},
		"subject": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Subject line when the message is delivered to email endpoints.",
		},
		"message": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The message content of the SNS notification.",
		},
		"sigv4": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Configures AWS's Signature Verification 4 signing process to sign requests.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"region": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The AWS region. If blank, the region from the default credentials chain is used.",
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
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Named AWS profile used to authenticate.",
					},
					"role_arn": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "AWS Role ARN, an alternative to using AWS API keys.",
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
						Type:         schema.TypeString,
						Description:  "The time after which an alert is declared resolved if it has not been updated.",
						Optional:     true,
						Default:      "5m",
						ValidateFunc: validateDuration,
						StateFunc:    formatDuration,
					},
					"http_config": {
						Type:        schema.TypeList,
						Optional:    true,
						MaxItems:    1,
						Description: "The default HTTP client configuration",
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
					"webex_api_url": {
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
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The default SMTP From header field.",
					},
					"smtp_hello": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The default hostname to identify to the SMTP server.",
					},
					"smtp_smarthost": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The default SMTP smarthost used for sending emails, including port number.",
					},
					"smtp_auth_username": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "SMTP Auth using CRAM-MD5, LOGIN and PLAIN. If empty, Alertmanager doesn't authenticate to the SMTP server.",
					},
					"smtp_auth_password": {
						Type:        schema.TypeString,
						Optional:    true,
						Sensitive:   true,
						Description: "SMTP Auth using LOGIN and PLAIN.",
					},
					"smtp_auth_secret": {
						Type:        schema.TypeString,
						Optional:    true,
						Sensitive:   true,
						Description: "SMTP Auth using CRAM-MD5.",
					},
					"smtp_auth_identity": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "SMTP Auth using PLAIN.",
					},
					"smtp_require_tls": {
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
						Description: "The default SMTP TLS requirement.",
					},
				},
			},
		},
		"inhibit_rule": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Mutes an alert (target) matching a set of matchers when an alert (source) exists that matches another set of matchers.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"source_matchers": {
						Type:        schema.TypeList,
						Optional:    true,
						Description: "A list of matchers for which one or more alerts have to exist for the inhibition to take effect.",
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"target_matchers": {
						Type:        schema.TypeList,
						Optional:    true,
						Description: "A list of matchers that have to be fulfilled by the target alerts to be muted.",
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"equal": {
						Type:        schema.TypeList,
						Optional:    true,
						Description: "Labels that must have an equal value in the source and target alert for the inhibition to take effect.",
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
				},
			},
		},
		"time_interval": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "A list of time intervals for muting/activating routes.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Name interval of time that may be referenced in the routing tree to mute/activate particular routes for particular times of the day.",
					},
					"time_intervals": {
						Type:        schema.TypeList,
						Optional:    true,
						MaxItems:    1,
						Description: "The actual definition for an interval of time.",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"times": {
									Type:        schema.TypeList,
									Optional:    true,
									Description: "Ranges inclusive of the starting time and exclusive of the end time to make it easy to represent times that start/end on hour boundaries.",
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
									Type:        schema.TypeList,
									Optional:    true,
									Description: "A list of numerical days of the week, where the week begins on Sunday (0) and ends on Saturday (6).",
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
									Type:        schema.TypeList,
									Optional:    true,
									Description: "A list of numerical days in the month. Days begin at 1. Negative values are also accepted which begin at the end of the month.",
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
									Type:        schema.TypeList,
									Optional:    true,
									Description: "A list of calendar months identified by number, where January = 1.",
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
									Type:        schema.TypeList,
									Optional:    true,
									Description: "A numerical list of years.",
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
								"location": {
									Type:        schema.TypeString,
									Optional:    true,
									Default:     "UTC",
									Description: "A string that matches a location in the IANA time zone database.",
								},
							},
						},
					},
				},
			},
		},
		"receiver": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "A list of notification receivers.",
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
					"webex_configs": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: webexConfigFields(),
						},
					},
					"discord_configs": {
						Type:     schema.TypeList,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: discordConfigFields(),
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
			Type:        schema.TypeList,
			Required:    true,
			MaxItems:    1,
			Description: "The root node of the routing tree.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"group_by": {
						Type:        schema.TypeList,
						Optional:    true,
						Description: "The labels by which incoming alerts are grouped together.",
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"group_wait": {
						Type:         schema.TypeString,
						Required:     true,
						Description:  "How long to initially wait to send a notification for a group of alerts. Allows to wait for an inhibiting alert to arrive or collect more initial alerts for the same group.",
						ValidateFunc: validateDuration,
						StateFunc:    formatDuration,
					},
					"group_interval": {
						Type:         schema.TypeString,
						Required:     true,
						Description:  "How long to wait before sending a notification about new alerts that are added to a group of alerts for which an initial notification has already been sent.",
						ValidateFunc: validateDuration,
						StateFunc:    formatDuration,
					},
					"repeat_interval": {
						Type:         schema.TypeString,
						Required:     true,
						Description:  "How long to wait before sending a notification again if it has already been sent successfully for an alert.",
						ValidateFunc: validateDuration,
						StateFunc:    formatDuration,
					},
					"receiver": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Name of the receiver to send the notification.",
					},
					"child_route": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"group_by": {
									Type:        schema.TypeList,
									Optional:    true,
									Description: "The labels by which incoming alerts are grouped together.",
									Elem:        &schema.Schema{Type: schema.TypeString},
								},
								"matchers": {
									Type:        schema.TypeList,
									Optional:    true,
									Description: "A list of matchers that an alert has to fulfill to match the node.",
									Elem:        &schema.Schema{Type: schema.TypeString},
								},
								"group_wait": {
									Type:         schema.TypeString,
									Required:     true,
									Description:  "How long to initially wait to send a notification for a group of alerts. Allows to wait for an inhibiting alert to arrive or collect more initial alerts for the same group.",
									ValidateFunc: validateDuration,
									StateFunc:    formatDuration,
								},
								"group_interval": {
									Type:         schema.TypeString,
									Required:     true,
									Description:  "How long to wait before sending a notification about new alerts that are added to a group of alerts for which an initial notification has already been sent.",
									ValidateFunc: validateDuration,
									StateFunc:    formatDuration,
								},
								"repeat_interval": {
									Type:         schema.TypeString,
									Required:     true,
									Description:  "How long to wait before sending a notification again if it has already been sent successfully for an alert.",
									ValidateFunc: validateDuration,
									StateFunc:    formatDuration,
								},
								"receiver": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "Name of the receiver to send the notification.",
								},
								"continue": {
									Type:        schema.TypeBool,
									Optional:    true,
									Default:     false,
									Description: "Whether an alert should continue matching subsequent sibling nodes.",
								},
								"mute_time_intervals": {
									Type:        schema.TypeList,
									Optional:    true,
									Description: "Times when the route should be muted. These must match the name of a mute time interval defined in the time_interval block.",
									Elem:        &schema.Schema{Type: schema.TypeString},
								},
								"active_time_intervals": {
									Type:        schema.TypeList,
									Optional:    true,
									Description: "Times when the route should be active. These must match the name of a mute time interval defined in the time_interval block.",
									Elem:        &schema.Schema{Type: schema.TypeString},
								},
							},
						},
					},
				},
			},
		},
		"templates": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "A list of template names to use.",
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"templates_files": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "A map of key values string, where the key is the template name and the value the content of the template.",
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
	}
}

func dataSourceMimirAlertmanagerConfigSchemaV1() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Name of the alertmanager configuration. Only used for resource dependency.",
		},
		"global": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"resolve_timeout": {
						Type:        schema.TypeString,
						Description: "The time after which an alert is declared resolved if it has not been updated.",
						Optional:    true,
						Default:     "5m",
					},
					"http_config": {
						Type:        schema.TypeList,
						Computed:    true,
						Description: "The default HTTP client configuration",
						Elem: &schema.Resource{
							Schema: httpConfigFields(),
						},
					},
					"pagerduty_url": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"opsgenie_api_url": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"webex_api_url": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"wechat_api_url": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"victorops_api_url": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"telegram_api_url": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"slack_api_url": {
						Type:      schema.TypeString,
						Computed:  true,
						Sensitive: true,
					},
					"opsgenie_api_key": {
						Type:      schema.TypeString,
						Computed:  true,
						Sensitive: true,
					},
					"wechat_api_secret": {
						Type:      schema.TypeString,
						Computed:  true,
						Sensitive: true,
					},
					"wechat_api_corp_id": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"victorops_api_key": {
						Type:      schema.TypeString,
						Computed:  true,
						Sensitive: true,
					},
					"smtp_from": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The default SMTP From header field.",
					},
					"smtp_hello": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The default hostname to identify to the SMTP server.",
					},
					"smtp_smarthost": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The default SMTP smarthost used for sending emails, including port number.",
					},
					"smtp_auth_username": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "SMTP Auth using CRAM-MD5, LOGIN and PLAIN. If empty, Alertmanager doesn't authenticate to the SMTP server.",
					},
					"smtp_auth_password": {
						Type:        schema.TypeString,
						Computed:    true,
						Sensitive:   true,
						Description: "SMTP Auth using LOGIN and PLAIN.",
					},
					"smtp_auth_secret": {
						Type:        schema.TypeString,
						Computed:    true,
						Sensitive:   true,
						Description: "SMTP Auth using CRAM-MD5.",
					},
					"smtp_auth_identity": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "SMTP Auth using PLAIN.",
					},
					"smtp_require_tls": {
						Type:        schema.TypeBool,
						Computed:    true,
						Default:     nil,
						Description: "The default SMTP TLS requirement.",
					},
				},
			},
		},
		"inhibit_rule": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "Mutes an alert (target) matching a set of matchers when an alert (source) exists that matches another set of matchers.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"source_matchers": {
						Type:        schema.TypeList,
						Computed:    true,
						Description: "A list of matchers for which one or more alerts have to exist for the inhibition to take effect.",
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"target_matchers": {
						Type:        schema.TypeList,
						Computed:    true,
						Description: "A list of matchers that have to be fulfilled by the target alerts to be muted.",
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"equal": {
						Type:        schema.TypeList,
						Computed:    true,
						Description: "Labels that must have an equal value in the source and target alert for the inhibition to take effect.",
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
				},
			},
		},
		"time_interval": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "A list of time intervals for muting/activating routes.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Name interval of time that may be referenced in the routing tree to mute/activate particular routes for particular times of the day.",
					},
					"time_intervals": {
						Type:        schema.TypeList,
						Computed:    true,
						Description: "The actual definition for an interval of time.",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"times": {
									Type:        schema.TypeList,
									Computed:    true,
									Description: "Ranges inclusive of the starting time and exclusive of the end time to make it easy to represent times that start/end on hour boundaries.",
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"start_minute": {
												Type:     schema.TypeInt,
												Computed: true,
											},
											"end_minute": {
												Type:     schema.TypeInt,
												Computed: true,
											},
										},
									},
								},
								"weekdays": {
									Type:        schema.TypeList,
									Computed:    true,
									Description: "A list of numerical days of the week, where the week begins on Sunday (0) and ends on Saturday (6).",
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"begin": {
												Type:     schema.TypeInt,
												Computed: true,
											},
											"end": {
												Type:     schema.TypeInt,
												Computed: true,
											},
										},
									},
								},
								"days_of_month": {
									Type:        schema.TypeList,
									Computed:    true,
									Description: "A list of numerical days in the month. Days begin at 1. Negative values are also accepted which begin at the end of the month.",
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"begin": {
												Type:     schema.TypeInt,
												Computed: true,
											},
											"end": {
												Type:     schema.TypeInt,
												Computed: true,
											},
										},
									},
								},
								"months": {
									Type:        schema.TypeList,
									Computed:    true,
									Description: "A list of calendar months identified by number, where January = 1.",
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"begin": {
												Type:     schema.TypeInt,
												Computed: true,
											},
											"end": {
												Type:     schema.TypeInt,
												Computed: true,
											},
										},
									},
								},
								"years": {
									Type:        schema.TypeList,
									Computed:    true,
									Description: "A numerical list of years.",
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"begin": {
												Type:     schema.TypeInt,
												Computed: true,
											},
											"end": {
												Type:     schema.TypeInt,
												Computed: true,
											},
										},
									},
								},
								"location": {
									Type:        schema.TypeString,
									Optional:    true,
									Default:     "UTC",
									Description: "A string that matches a location in the IANA time zone database.",
								},
							},
						},
					},
				},
			},
		},
		"receiver": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "A list of notification receivers.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:        schema.TypeString,
						Description: "The time after which an alert is declared resolved if it has not been updated.",
						Computed:    true,
					},
					"pagerduty_configs": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Resource{
							Schema: pagerdutyConfigFields(),
						},
					},
					"email_configs": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Resource{
							Schema: emailConfigFields(),
						},
					},
					"wechat_configs": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Resource{
							Schema: weChatConfigFields(),
						},
					},
					"webhook_configs": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Resource{
							Schema: webhookConfigFields(),
						},
					},
					"webex_configs": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Resource{
							Schema: webexConfigFields(),
						},
					},
					"discord_configs": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Resource{
							Schema: discordConfigFields(),
						},
					},
					"pushover_configs": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Resource{
							Schema: pushoverConfigFields(),
						},
					},
					"opsgenie_configs": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Resource{
							Schema: opsgenieConfigFields(),
						},
					},
					"slack_configs": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Resource{
							Schema: slackConfigFields(),
						},
					},
					"telegram_configs": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Resource{
							Schema: telegramConfigFields(),
						},
					},
					"victorops_configs": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Resource{
							Schema: victorOpsConfigFields(),
						},
					},
					"sns_configs": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Resource{
							Schema: snsConfigFields(),
						},
					},
				},
			},
		},
		"route": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "The root node of the routing tree.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"group_by": {
						Type:        schema.TypeList,
						Computed:    true,
						Description: "The labels by which incoming alerts are grouped together.",
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"group_wait": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "How long to initially wait to send a notification for a group of alerts. Allows to wait for an inhibiting alert to arrive or collect more initial alerts for the same group.",
					},
					"group_interval": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "How long to wait before sending a notification about new alerts that are added to a group of alerts for which an initial notification has already been sent.",
					},
					"repeat_interval": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "How long to wait before sending a notification again if it has already been sent successfully for an alert.",
					},
					"receiver": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Name of the receiver to send the notification.",
					},
					"child_route": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"group_by": {
									Type:        schema.TypeList,
									Computed:    true,
									Description: "The labels by which incoming alerts are grouped together.",
									Elem:        &schema.Schema{Type: schema.TypeString},
								},
								"matchers": {
									Type:        schema.TypeList,
									Computed:    true,
									Description: "A list of matchers that an alert has to fulfill to match the node.",
									Elem:        &schema.Schema{Type: schema.TypeString},
								},
								"group_wait": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "How long to initially wait to send a notification for a group of alerts. Allows to wait for an inhibiting alert to arrive or collect more initial alerts for the same group.",
								},
								"group_interval": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "How long to wait before sending a notification about new alerts that are added to a group of alerts for which an initial notification has already been sent.",
								},
								"repeat_interval": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "How long to wait before sending a notification again if it has already been sent successfully for an alert.",
								},
								"receiver": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "Name of the receiver to send the notification.",
								},
								"continue": {
									Type:        schema.TypeBool,
									Computed:    true,
									Default:     nil,
									Description: "Whether an alert should continue matching subsequent sibling nodes.",
								},
								"mute_time_intervals": {
									Type:        schema.TypeList,
									Computed:    true,
									Description: "Times when the route should be muted. These must match the name of a mute time interval defined in the time_interval block.",
									Elem:        &schema.Schema{Type: schema.TypeString},
								},
								"active_time_intervals": {
									Type:        schema.TypeList,
									Computed:    true,
									Description: "Times when the route should be active. These must match the name of a mute time interval defined in the time_interval block.",
									Elem:        &schema.Schema{Type: schema.TypeString},
								},
							},
						},
					},
				},
			},
		},
		"templates": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "A list of template names to use.",
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"templates_files": {
			Type:        schema.TypeMap,
			Computed:    true,
			Description: "A map of key values string, where the key is the template name and the value the content of the template.",
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
	}
}
