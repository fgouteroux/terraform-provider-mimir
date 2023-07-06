package mimir

import (
	"net"
	"net/url"
	"time"

	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/alertmanager/timeinterval"
	"github.com/prometheus/common/model"
)

func expandHTTPConfigOAuth2(v interface{}) *oauth2 {
	var oauth2Conf *oauth2
	data := v.([]interface{})
	if len(data) != 0 && data[0] != nil {
		oauth2Conf = &oauth2{}
		cfg := data[0].(map[string]interface{})
		oauth2Conf.ClientID = cfg["client_id"].(string)
		oauth2Conf.ClientSecret = cfg["client_secret"].(string)
		oauth2Conf.TokenURL = cfg["token_url"].(string)
		oauth2Conf.Scopes = expandStringArray(cfg["scopes"].([]interface{}))
		oauth2Conf.EndpointParams = expandStringMap(cfg["endpoint_params"].(map[string]interface{}))
	}
	return oauth2Conf
}

func flattenHTTPConfigOAuth2(v *oauth2) []interface{} {
	oauth2Conf := make(map[string]interface{})
	if v != nil {
		oauth2Conf["client_id"] = v.ClientID
		oauth2Conf["client_secret"] = v.ClientSecret
		oauth2Conf["token_url"] = v.TokenURL
		oauth2Conf["scopes"] = v.Scopes
		oauth2Conf["endpoint_params"] = v.EndpointParams
	}
	return []interface{}{oauth2Conf}
}

func expandHTTPConfigBasicAuth(v interface{}) *basicAuth {
	var basicAuthConf *basicAuth
	data := v.([]interface{})
	if len(data) != 0 && data[0] != nil {
		basicAuthConf = &basicAuth{}
		cfg := data[0].(map[string]interface{})
		basicAuthConf.Username = cfg["username"].(string)
		basicAuthConf.Password = cfg["password"].(string)
	}
	return basicAuthConf
}

func flattenHTTPConfigBasicAuth(v *basicAuth) []interface{} {
	basicAuthConf := make(map[string]interface{})
	if v != nil {
		basicAuthConf["username"] = v.Username
		basicAuthConf["password"] = v.Password
	}
	return []interface{}{basicAuthConf}
}

func expandHTTPConfigAuthorization(v interface{}) *authorization {
	var authConf *authorization
	data := v.([]interface{})
	if len(data) != 0 && data[0] != nil {
		authConf = &authorization{}
		cfg := data[0].(map[string]interface{})
		authConf.Type = cfg["type"].(string)
		authConf.Credentials = cfg["credentials"].(string)
	}
	return authConf
}

func flattenHTTPConfigAuthorization(v *authorization) []interface{} {
	authConf := make(map[string]interface{})
	if v != nil {
		authConf["type"] = v.Type
		authConf["credentials"] = v.Credentials
	}
	return []interface{}{authConf}
}

func expandTLSConfig(v interface{}) *tlsConfig {
	var tlsConf *tlsConfig
	data := v.([]interface{})
	if len(data) != 0 && data[0] != nil {
		tlsConf = &tlsConfig{}
		cfg := data[0].(map[string]interface{})
		tlsConf.ServerName = cfg["server_name"].(string)
		tlsConf.InsecureSkipVerify = cfg["insecure_skip_verify"].(bool)
		tlsConf.MinVersion = cfg["min_version"].(string)
		tlsConf.MaxVersion = cfg["max_version"].(string)
	}
	return tlsConf
}

func flattenTLSConfig(v *tlsConfig) []interface{} {
	tlsConf := make(map[string]interface{})
	if v != nil {
		tlsConf["server_name"] = v.ServerName
		tlsConf["insecure_skip_verify"] = v.InsecureSkipVerify
		tlsConf["min_version"] = v.MinVersion
		tlsConf["max_version"] = v.MaxVersion
	}
	return []interface{}{tlsConf}
}

func expandHTTPConfig(v interface{}) *httpClientConfig {
	var httpConf *httpClientConfig
	data := v.([]interface{})
	if len(data) != 0 && data[0] != nil {
		httpConf = &httpClientConfig{}
		cfg := data[0].(map[string]interface{})
		httpConf.ProxyURL = cfg["proxy_url"].(string)
		httpConf.FollowRedirects = new(bool)
		*httpConf.FollowRedirects = cfg["follow_redirects"].(bool)
		httpConf.EnableHTTP2 = new(bool)
		*httpConf.EnableHTTP2 = cfg["enable_http2"].(bool)
		httpConf.BearerToken = cfg["bearer_token"].(string)

		if len(cfg["authorization"].([]interface{})) > 0 {
			httpConf.Authorization = expandHTTPConfigAuthorization(cfg["authorization"].([]interface{}))
		}

		if len(cfg["basic_auth"].([]interface{})) > 0 {
			httpConf.BasicAuth = expandHTTPConfigBasicAuth(cfg["basic_auth"].([]interface{}))
		}

		if len(cfg["oauth2"].([]interface{})) > 0 {
			httpConf.OAuth2 = expandHTTPConfigOAuth2(cfg["oauth2"].([]interface{}))
		}

		if len(cfg["tls_config"].([]interface{})) > 0 {
			httpConf.TLSConfig = expandTLSConfig(cfg["tls_config"].([]interface{}))
		}
	}

	return httpConf
}

func flattenHTTPConfig(v *httpClientConfig) []interface{} {
	httpConf := make(map[string]interface{})

	if v != nil {
		httpConf["proxy_url"] = v.ProxyURL
		httpConf["bearer_token"] = v.BearerToken

		if v.FollowRedirects != nil {
			httpConf["follow_redirects"] = v.FollowRedirects
		}

		if v.EnableHTTP2 != nil {
			httpConf["enable_http2"] = v.EnableHTTP2
		}

		if v.BasicAuth != nil {
			httpConf["basic_auth"] = flattenHTTPConfigBasicAuth(v.BasicAuth)
		}
		if v.OAuth2 != nil {
			httpConf["oauth2"] = flattenHTTPConfigOAuth2(v.OAuth2)
		}
		if v.Authorization != nil {
			httpConf["authorization"] = flattenHTTPConfigAuthorization(v.Authorization)
		}

		if v.TLSConfig != nil {
			httpConf["tls_config"] = flattenTLSConfig(v.TLSConfig)
		}
	}
	return []interface{}{httpConf}
}

func expandGlobalConfig(v interface{}) *globalConfig {
	var globalConf *globalConfig
	data := v.([]interface{})
	if len(data) != 0 && data[0] != nil {
		globalConf = &globalConfig{}
		cfg := data[0].(map[string]interface{})
		resolveTimeout, _ := model.ParseDuration(cfg["resolve_timeout"].(string))
		globalConf.ResolveTimeout = new(model.Duration)
		*globalConf.ResolveTimeout = resolveTimeout

		pagerdutyURL, _ := url.Parse(cfg["pagerduty_url"].(string))
		if pagerdutyURL.String() != "" {
			globalConf.PagerdutyURL = &config.URL{URL: pagerdutyURL}
		}

		slackAPIURL, _ := url.Parse(cfg["slack_api_url"].(string))
		if slackAPIURL.String() != "" {
			globalConf.SlackAPIURL = &config.URL{URL: slackAPIURL}
		}

		globalConf.SMTPFrom = cfg["smtp_from"].(string)
		globalConf.SMTPHello = cfg["smtp_hello"].(string)

		var hp config.HostPort
		hp.Host, hp.Port, _ = net.SplitHostPort(cfg["smtp_smarthost"].(string))
		globalConf.SMTPSmarthost = hp

		globalConf.SMTPAuthUsername = cfg["smtp_auth_username"].(string)
		globalConf.SMTPAuthPassword = cfg["smtp_auth_password"].(string)
		globalConf.SMTPAuthSecret = cfg["smtp_auth_secret"].(string)
		globalConf.SMTPAuthIdentity = cfg["smtp_auth_identity"].(string)
		globalConf.SMTPRequireTLS = new(bool)
		*globalConf.SMTPRequireTLS = cfg["smtp_require_tls"].(bool)
		globalConf.HTTPConfig = expandHTTPConfig(cfg["http_config"])

		globalConf.OpsGenieAPIKey = cfg["opsgenie_api_key"].(string)
		opsGenieAPIURL, _ := url.Parse(cfg["opsgenie_api_url"].(string))
		if opsGenieAPIURL.String() != "" {
			globalConf.OpsGenieAPIURL = &config.URL{URL: opsGenieAPIURL}
		}

		webexAPIURL, _ := url.Parse(cfg["webex_api_url"].(string))
		if webexAPIURL.String() != "" {
			globalConf.WebexAPIURL = &config.URL{URL: webexAPIURL}
		}

		globalConf.WeChatAPISecret = cfg["wechat_api_secret"].(string)
		globalConf.WeChatAPICorpID = cfg["wechat_api_corp_id"].(string)
		weChatAPIURL, _ := url.Parse(cfg["wechat_api_url"].(string))
		if weChatAPIURL.String() != "" {
			globalConf.WeChatAPIURL = &config.URL{URL: weChatAPIURL}
		}

		globalConf.VictorOpsAPIKey = cfg["victorops_api_key"].(string)
		victorOpsAPIURL, _ := url.Parse(cfg["victorops_api_url"].(string))
		if victorOpsAPIURL.String() != "" {
			globalConf.VictorOpsAPIURL = &config.URL{URL: victorOpsAPIURL}
		}

		telegramAPIURL, _ := url.Parse(cfg["telegram_api_url"].(string))
		if telegramAPIURL.String() != "" {
			globalConf.TelegramAPIURL = &config.URL{URL: telegramAPIURL}
		}
	}
	return globalConf
}

func flattenGlobalConfig(v *globalConfig) []interface{} {
	globalConf := make(map[string]interface{})

	if v != nil {
		if v.ResolveTimeout != nil {
			globalConf["resolve_timeout"] = v.ResolveTimeout.String()
		}

		if v.PagerdutyURL != nil {
			globalConf["pagerduty_url"] = v.PagerdutyURL.URL.String()
		}

		if v.SlackAPIURL != nil {
			globalConf["slack_api_url"] = v.SlackAPIURL.URL.String()
		}

		if v.OpsGenieAPIURL != nil {
			globalConf["opsgenie_api_url"] = v.OpsGenieAPIURL.URL.String()
		}

		if v.WebexAPIURL != nil {
			globalConf["webex_api_url"] = v.WebexAPIURL.URL.String()
		}

		if v.WeChatAPIURL != nil {
			globalConf["wechat_api_url"] = v.WeChatAPIURL.URL.String()
		}
		if v.VictorOpsAPIURL != nil {
			globalConf["victorops_api_url"] = v.VictorOpsAPIURL.URL.String()
		}

		if v.TelegramAPIURL != nil {
			globalConf["telegram_api_url"] = v.TelegramAPIURL.URL.String()
		}

		globalConf["opsgenie_api_key"] = v.OpsGenieAPIKey
		globalConf["wechat_api_secret"] = v.WeChatAPISecret
		globalConf["wechat_api_corp_id"] = v.WeChatAPICorpID
		globalConf["victorops_api_key"] = v.VictorOpsAPIKey

		if v.HTTPConfig != nil {
			globalConf["http_config"] = flattenHTTPConfig(v.HTTPConfig)
		}
		globalConf["smtp_from"] = v.SMTPFrom
		globalConf["smtp_hello"] = v.SMTPHello
		globalConf["smtp_smarthost"] = v.SMTPSmarthost.String()
		globalConf["smtp_auth_username"] = v.SMTPAuthUsername
		globalConf["smtp_auth_password"] = v.SMTPAuthPassword
		globalConf["smtp_auth_secret"] = v.SMTPAuthSecret
		globalConf["smtp_auth_identity"] = v.SMTPAuthIdentity

		if v.SMTPRequireTLS != nil {
			globalConf["smtp_require_tls"] = v.SMTPRequireTLS
		}
	}

	return []interface{}{globalConf}
}

func expandReceiverConfig(v []interface{}) []*receiver {
	var receiverConf []*receiver

	for _, v := range v {
		cfg := &receiver{}
		data := v.(map[string]interface{})

		if raw, ok := data["name"]; ok {
			cfg.Name = raw.(string)
		}
		if raw, ok := data["pagerduty_configs"]; ok {
			cfg.PagerdutyConfigs = expandPagerdutyConfig(raw.([]interface{}))
		}
		if raw, ok := data["email_configs"]; ok {
			cfg.EmailConfigs = expandEmailConfig(raw.([]interface{}))
		}
		if raw, ok := data["wechat_configs"]; ok {
			cfg.WeChatConfigs = expandWeChatConfig(raw.([]interface{}))
		}
		if raw, ok := data["webhook_configs"]; ok {
			cfg.WebhookConfigs = expandWebhookConfig(raw.([]interface{}))
		}
		if raw, ok := data["webex_configs"]; ok {
			cfg.WebexConfigs = expandWebexConfig(raw.([]interface{}))
		}
		if raw, ok := data["discord_configs"]; ok {
			cfg.DiscordConfigs = expandDiscordConfig(raw.([]interface{}))
		}
		if raw, ok := data["pushover_configs"]; ok {
			cfg.PushoverConfigs = expandPushoverConfig(raw.([]interface{}))
		}
		if raw, ok := data["opsgenie_configs"]; ok {
			cfg.OpsgenieConfigs = expandOpsgenieConfig(raw.([]interface{}))
		}
		if raw, ok := data["slack_configs"]; ok {
			cfg.SlackConfigs = expandSlackConfig(raw.([]interface{}))
		}
		if raw, ok := data["telegram_configs"]; ok {
			cfg.TelegramConfigs = expandTelegramConfig(raw.([]interface{}))
		}
		if raw, ok := data["victorops_configs"]; ok {
			cfg.VictorOpsConfigs = expandVictorOpsConfig(raw.([]interface{}))
		}
		if raw, ok := data["sns_configs"]; ok {
			cfg.SNSConfigs = expandSnsConfig(raw.([]interface{}))
		}
		receiverConf = append(receiverConf, cfg)
	}
	return receiverConf
}

func flattenReceiverConfig(v []*receiver) []interface{} {
	var receiverConf []interface{}

	if v == nil {
		return receiverConf
	}

	for _, v := range v {
		cfg := make(map[string]interface{})
		cfg["name"] = v.Name
		cfg["pagerduty_configs"] = flattenPagerdutyConfig(v.PagerdutyConfigs)
		cfg["email_configs"] = flattenEmailConfig(v.EmailConfigs)
		cfg["wechat_configs"] = flattenWeChatConfig(v.WeChatConfigs)
		cfg["webhook_configs"] = flattenWebhookConfig(v.WebhookConfigs)
		cfg["webex_configs"] = flattenWebexConfig(v.WebexConfigs)
		cfg["discord_configs"] = flattenDiscordConfig(v.DiscordConfigs)
		cfg["pushover_configs"] = flattenPushoverConfig(v.PushoverConfigs)
		cfg["opsgenie_configs"] = flattenOpsgenieConfig(v.OpsgenieConfigs)
		cfg["slack_configs"] = flattenSlackConfig(v.SlackConfigs)
		cfg["telegram_configs"] = flattenTelegramConfig(v.TelegramConfigs)
		cfg["victorops_configs"] = flattenVictorOpsConfig(v.VictorOpsConfigs)
		cfg["sns_configs"] = flattenSnsConfig(v.SNSConfigs)
		receiverConf = append(receiverConf, cfg)
	}
	return receiverConf
}

func expandSnsSigV4Config(v interface{}) *sigV4Config {
	sigV4ConfigConf := &sigV4Config{}
	data := v.([]interface{})
	if len(data) != 0 && data[0] != nil {
		cfg := data[0].(map[string]interface{})
		sigV4ConfigConf.Region = cfg["region"].(string)
		sigV4ConfigConf.AccessKey = cfg["access_key"].(string)
		sigV4ConfigConf.SecretKey = cfg["secret_key"].(string)
		sigV4ConfigConf.Profile = cfg["profile"].(string)
		sigV4ConfigConf.RoleARN = cfg["role_arn"].(string)
	}
	return sigV4ConfigConf
}

func flattenSnsSigV4Config(v *sigV4Config) []interface{} {
	sigV4ConfigConf := make(map[string]interface{})
	sigV4ConfigConf["region"] = v.Region
	sigV4ConfigConf["access_key"] = v.AccessKey
	sigV4ConfigConf["secret_key"] = v.SecretKey
	sigV4ConfigConf["profile"] = v.Profile
	sigV4ConfigConf["role_arn"] = v.RoleARN
	return []interface{}{sigV4ConfigConf}
}

func expandSnsConfig(v []interface{}) []*snsConfig {
	var snsConf []*snsConfig

	for _, v := range v {
		cfg := &snsConfig{}
		data := v.(map[string]interface{})

		if raw, ok := data["send_resolved"]; ok {
			cfg.VSendResolved = new(bool)
			*cfg.VSendResolved = raw.(bool)
		}
		if raw, ok := data["http_config"]; ok {
			cfg.HTTPConfig = expandHTTPConfig(raw)
		}
		if raw, ok := data["sigv4"]; ok {
			cfg.Sigv4 = expandSnsSigV4Config(raw)
		}
		if raw, ok := data["api_url"]; ok {
			cfg.APIUrl = raw.(string)
		}
		if raw, ok := data["topic_arn"]; ok {
			cfg.TopicARN = raw.(string)
		}
		if raw, ok := data["phone_number"]; ok {
			cfg.PhoneNumber = raw.(string)
		}
		if raw, ok := data["target_arn"]; ok {
			cfg.TargetARN = raw.(string)
		}
		if raw, ok := data["subject"]; ok {
			cfg.Subject = raw.(string)
		}
		if raw, ok := data["message"]; ok {
			cfg.Message = raw.(string)
		}
		if raw, ok := data["attributes"]; ok {
			cfg.Attributes = expandStringMap(raw.(map[string]interface{}))
		}
		snsConf = append(snsConf, cfg)
	}
	return snsConf
}

func flattenSnsConfig(v []*snsConfig) []interface{} {
	var snsConf []interface{}

	if v == nil {
		return snsConf
	}

	for _, v := range v {
		cfg := make(map[string]interface{})
		cfg["send_resolved"] = v.VSendResolved
		if v.HTTPConfig != nil {
			cfg["http_config"] = flattenHTTPConfig(v.HTTPConfig)
		}
		cfg["sigv4"] = flattenSnsSigV4Config(v.Sigv4)
		cfg["api_url"] = v.APIUrl
		cfg["topic_arn"] = v.TopicARN
		cfg["phone_number"] = v.PhoneNumber
		cfg["target_arn"] = v.TargetARN
		cfg["subject"] = v.Subject
		cfg["message"] = v.Message
		cfg["attributes"] = v.Attributes
		snsConf = append(snsConf, cfg)
	}
	return snsConf
}

func expandVictorOpsConfig(v []interface{}) []*victorOpsConfig {
	var victorOpsConf []*victorOpsConfig

	for _, v := range v {
		cfg := &victorOpsConfig{}
		data := v.(map[string]interface{})

		if raw, ok := data["send_resolved"]; ok {
			cfg.VSendResolved = new(bool)
			*cfg.VSendResolved = raw.(bool)
		}
		if raw, ok := data["http_config"]; ok {
			cfg.HTTPConfig = expandHTTPConfig(raw)
		}
		if raw, ok := data["api_key"]; ok {
			cfg.APIKey = raw.(string)
		}
		if raw, ok := data["api_url"]; ok {
			cfg.APIURL = raw.(string)
		}
		if raw, ok := data["routing_key"]; ok {
			cfg.RoutingKey = raw.(string)
		}
		if raw, ok := data["message_type"]; ok {
			cfg.MessageType = raw.(string)
		}
		if raw, ok := data["state_message"]; ok {
			cfg.StateMessage = raw.(string)
		}
		if raw, ok := data["entity_display_name"]; ok {
			cfg.EntityDisplayName = raw.(string)
		}
		if raw, ok := data["monitoring_tool"]; ok {
			cfg.MonitoringTool = raw.(string)
		}
		if raw, ok := data["custom_fields"]; ok {
			cfg.CustomFields = expandStringMap(raw.(map[string]interface{}))
		}
		victorOpsConf = append(victorOpsConf, cfg)
	}
	return victorOpsConf
}

func flattenVictorOpsConfig(v []*victorOpsConfig) []interface{} {
	var victorOpsConf []interface{}

	if v == nil {
		return victorOpsConf
	}

	for _, v := range v {
		cfg := make(map[string]interface{})
		cfg["send_resolved"] = v.VSendResolved
		if v.HTTPConfig != nil {
			cfg["http_config"] = flattenHTTPConfig(v.HTTPConfig)
		}
		cfg["api_key"] = v.APIKey
		cfg["api_url"] = v.APIURL
		cfg["routing_key"] = v.RoutingKey
		cfg["message_type"] = v.MessageType
		cfg["state_message"] = v.StateMessage
		cfg["entity_display_name"] = v.EntityDisplayName
		cfg["monitoring_tool"] = v.MonitoringTool
		cfg["custom_fields"] = v.CustomFields
		victorOpsConf = append(victorOpsConf, cfg)
	}
	return victorOpsConf
}

func expandTelegramConfig(v []interface{}) []*telegramConfig {
	var telegramConf []*telegramConfig

	for _, v := range v {
		cfg := &telegramConfig{}
		data := v.(map[string]interface{})

		if raw, ok := data["send_resolved"]; ok {
			cfg.VSendResolved = new(bool)
			*cfg.VSendResolved = raw.(bool)
		}
		if raw, ok := data["http_config"]; ok {
			cfg.HTTPConfig = expandHTTPConfig(raw)
		}
		if raw, ok := data["api_url"]; ok {
			cfg.APIUrl = raw.(string)
		}
		if raw, ok := data["bot_token"]; ok {
			cfg.BotToken = raw.(string)
		}
		if raw, ok := data["chat_id"]; ok {
			cfg.ChatID = int64(raw.(int))
		}
		if raw, ok := data["message"]; ok {
			cfg.Message = raw.(string)
		}
		if raw, ok := data["disable_notifications"]; ok {
			cfg.DisableNotifications = raw.(bool)
		}
		if raw, ok := data["parse_mode"]; ok {
			cfg.ParseMode = raw.(string)
		}

		telegramConf = append(telegramConf, cfg)
	}
	return telegramConf
}

func flattenTelegramConfig(v []*telegramConfig) []interface{} {
	var telegramConf []interface{}

	if v == nil {
		return telegramConf
	}

	for _, v := range v {
		cfg := make(map[string]interface{})
		cfg["send_resolved"] = v.VSendResolved
		if v.HTTPConfig != nil {
			cfg["http_config"] = flattenHTTPConfig(v.HTTPConfig)
		}
		cfg["api_url"] = v.APIUrl
		cfg["bot_token"] = v.BotToken
		cfg["chat_id"] = v.ChatID
		cfg["message"] = v.Message
		cfg["disable_notifications"] = v.DisableNotifications
		cfg["parse_mode"] = v.ParseMode
		telegramConf = append(telegramConf, cfg)
	}
	return telegramConf
}

func expandOpsgenieResponder(v []interface{}) []opsgenieResponder {
	var opsgenieResponderConf []opsgenieResponder

	for _, v := range v {
		var cfg opsgenieResponder
		data := v.(map[string]interface{})

		if raw, ok := data["username"]; ok {
			cfg.Username = raw.(string)
		}
		if raw, ok := data["name"]; ok {
			cfg.Name = raw.(string)
		}
		if raw, ok := data["type"]; ok {
			cfg.Type = raw.(string)
		}
		if raw, ok := data["id"]; ok {
			cfg.ID = raw.(string)
		}
		opsgenieResponderConf = append(opsgenieResponderConf, cfg)
	}
	return opsgenieResponderConf
}

func flattenOpsgenieResponder(v []opsgenieResponder) []interface{} {
	var opsgenieResponderConf []interface{}

	if v == nil {
		return opsgenieResponderConf
	}

	for _, v := range v {
		cfg := make(map[string]interface{})
		cfg["username"] = v.Username
		cfg["name"] = v.Name
		cfg["type"] = v.Type
		cfg["id"] = v.ID
		opsgenieResponderConf = append(opsgenieResponderConf, cfg)
	}
	return opsgenieResponderConf
}

func expandOpsgenieConfig(v []interface{}) []*opsgenieConfig {
	var opsgenieConf []*opsgenieConfig

	for _, v := range v {
		cfg := &opsgenieConfig{}
		data := v.(map[string]interface{})
		if raw, ok := data["send_resolved"]; ok {
			cfg.VSendResolved = new(bool)
			*cfg.VSendResolved = raw.(bool)
		}
		if raw, ok := data["http_config"]; ok {
			cfg.HTTPConfig = expandHTTPConfig(raw)
		}
		if raw, ok := data["api_key"]; ok {
			cfg.APIKey = raw.(string)
		}
		if raw, ok := data["api_url"]; ok {
			cfg.APIURL = raw.(string)
		}
		if raw, ok := data["message"]; ok {
			cfg.Message = raw.(string)
		}
		if raw, ok := data["description"]; ok {
			cfg.Description = raw.(string)
		}
		if raw, ok := data["source"]; ok {
			cfg.Source = raw.(string)
		}
		if raw, ok := data["details"]; ok {
			cfg.Details = expandStringMap(raw.(map[string]interface{}))
		}
		if raw, ok := data["responders"]; ok {
			cfg.Responders = expandOpsgenieResponder(raw.([]interface{}))
		}
		if raw, ok := data["tags"]; ok {
			cfg.Tags = raw.(string)
		}
		if raw, ok := data["note"]; ok {
			cfg.Note = raw.(string)
		}
		if raw, ok := data["priority"]; ok {
			cfg.Priority = raw.(string)
		}
		if raw, ok := data["update_alerts"]; ok {
			cfg.UpdateAlerts = new(bool)
			*cfg.UpdateAlerts = raw.(bool)
		}
		if raw, ok := data["entity"]; ok {
			cfg.Entity = raw.(string)
		}
		if raw, ok := data["actions"]; ok {
			cfg.Actions = raw.(string)
		}
		opsgenieConf = append(opsgenieConf, cfg)
	}
	return opsgenieConf
}

func flattenOpsgenieConfig(v []*opsgenieConfig) []interface{} {
	var opsgenieConf []interface{}

	if v == nil {
		return opsgenieConf
	}

	for _, v := range v {
		cfg := make(map[string]interface{})
		cfg["send_resolved"] = v.VSendResolved
		if v.HTTPConfig != nil {
			cfg["http_config"] = flattenHTTPConfig(v.HTTPConfig)
		}
		cfg["details"] = v.Details
		cfg["api_key"] = v.APIKey
		cfg["api_url"] = v.APIURL
		cfg["message"] = v.Message
		cfg["description"] = v.Description
		cfg["source"] = v.Source
		cfg["responders"] = flattenOpsgenieResponder(v.Responders)
		cfg["tags"] = v.Tags
		cfg["note"] = v.Note
		cfg["priority"] = v.Priority
		cfg["update_alerts"] = v.UpdateAlerts
		cfg["entity"] = v.Entity
		cfg["actions"] = v.Actions
		opsgenieConf = append(opsgenieConf, cfg)
	}
	return opsgenieConf
}

func expandWebhookConfig(v []interface{}) []*webhookConfig {
	var webhookConf []*webhookConfig

	for _, v := range v {
		cfg := &webhookConfig{}
		data := v.(map[string]interface{})
		if raw, ok := data["send_resolved"]; ok {
			cfg.VSendResolved = new(bool)
			*cfg.VSendResolved = raw.(bool)
		}
		if raw, ok := data["http_config"]; ok {
			cfg.HTTPConfig = expandHTTPConfig(raw)
		}
		if raw, ok := data["url"]; ok {
			cfg.URL = raw.(string)
		}
		if raw, ok := data["max_alerts"]; ok {
			cfg.MaxAlerts = int32(raw.(int))
		}
		webhookConf = append(webhookConf, cfg)
	}
	return webhookConf
}

func flattenWebhookConfig(v []*webhookConfig) []interface{} {
	var webhookConf []interface{}

	if v == nil {
		return webhookConf
	}

	for _, v := range v {
		cfg := make(map[string]interface{})
		cfg["send_resolved"] = v.VSendResolved
		if v.HTTPConfig != nil {
			cfg["http_config"] = flattenHTTPConfig(v.HTTPConfig)
		}
		cfg["url"] = v.URL
		cfg["max_alerts"] = v.MaxAlerts
		webhookConf = append(webhookConf, cfg)
	}
	return webhookConf
}

func expandWebexConfig(v []interface{}) []*webexConfig {
	var webexConf []*webexConfig

	for _, v := range v {
		cfg := &webexConfig{}
		data := v.(map[string]interface{})
		if raw, ok := data["send_resolved"]; ok {
			cfg.VSendResolved = new(bool)
			*cfg.VSendResolved = raw.(bool)
		}
		if raw, ok := data["http_config"]; ok {
			cfg.HTTPConfig = expandHTTPConfig(raw)
		}
		if raw, ok := data["api_url"]; ok {
			cfg.APIURL = raw.(string)
		}
		if raw, ok := data["room_id"]; ok {
			cfg.RoomID = raw.(string)
		}
		if raw, ok := data["message"]; ok {
			cfg.Message = raw.(string)
		}
		webexConf = append(webexConf, cfg)
	}
	return webexConf
}

func flattenWebexConfig(v []*webexConfig) []interface{} {
	var webexConf []interface{}

	if v == nil {
		return webexConf
	}

	for _, v := range v {
		cfg := make(map[string]interface{})
		cfg["send_resolved"] = v.VSendResolved
		if v.HTTPConfig != nil {
			cfg["http_config"] = flattenHTTPConfig(v.HTTPConfig)
		}
		cfg["api_url"] = v.APIURL
		cfg["room_id"] = v.RoomID
		cfg["message"] = v.Message
		webexConf = append(webexConf, cfg)
	}
	return webexConf
}

func expandDiscordConfig(v []interface{}) []*discordConfig {
	var discordConf []*discordConfig

	for _, v := range v {
		cfg := &discordConfig{}
		data := v.(map[string]interface{})
		if raw, ok := data["send_resolved"]; ok {
			cfg.VSendResolved = new(bool)
			*cfg.VSendResolved = raw.(bool)
		}
		if raw, ok := data["http_config"]; ok {
			cfg.HTTPConfig = expandHTTPConfig(raw)
		}
		if raw, ok := data["webhook_url"]; ok {
			cfg.WebhookURL = raw.(string)
		}
		if raw, ok := data["title"]; ok {
			cfg.Title = raw.(string)
		}
		if raw, ok := data["message"]; ok {
			cfg.Message = raw.(string)
		}
		discordConf = append(discordConf, cfg)
	}
	return discordConf
}

func flattenDiscordConfig(v []*discordConfig) []interface{} {
	var discordConf []interface{}

	if v == nil {
		return discordConf
	}

	for _, v := range v {
		cfg := make(map[string]interface{})
		cfg["send_resolved"] = v.VSendResolved
		if v.HTTPConfig != nil {
			cfg["http_config"] = flattenHTTPConfig(v.HTTPConfig)
		}
		cfg["webhook_url"] = v.WebhookURL
		cfg["title"] = v.Title
		cfg["message"] = v.Message
		discordConf = append(discordConf, cfg)
	}
	return discordConf
}

func expandWeChatConfig(v []interface{}) []*weChatConfig {
	var weChatConf []*weChatConfig

	for _, v := range v {
		cfg := &weChatConfig{}
		data := v.(map[string]interface{})
		if raw, ok := data["send_resolved"]; ok {
			cfg.VSendResolved = new(bool)
			*cfg.VSendResolved = raw.(bool)
		}
		if raw, ok := data["http_config"]; ok {
			cfg.HTTPConfig = expandHTTPConfig(raw)
		}
		if raw, ok := data["api_secret"]; ok {
			cfg.APISecret = raw.(string)
		}
		if raw, ok := data["api_url_url"]; ok {
			cfg.APIURL = raw.(string)
		}
		if raw, ok := data["corp_id"]; ok {
			cfg.CorpID = raw.(string)
		}
		if raw, ok := data["agent_id"]; ok {
			cfg.AgentID = raw.(string)
		}
		if raw, ok := data["to_user"]; ok {
			cfg.ToUser = raw.(string)
		}
		if raw, ok := data["to_party"]; ok {
			cfg.ToParty = raw.(string)
		}
		if raw, ok := data["to_tag"]; ok {
			cfg.ToTag = raw.(string)
		}
		if raw, ok := data["message"]; ok {
			cfg.Message = raw.(string)
		}
		if raw, ok := data["message_type"]; ok {
			cfg.MessageType = raw.(string)
		}
		weChatConf = append(weChatConf, cfg)
	}
	return weChatConf
}

func flattenWeChatConfig(v []*weChatConfig) []interface{} {
	var weChatConf []interface{}

	if v == nil {
		return weChatConf
	}

	for _, v := range v {
		cfg := make(map[string]interface{})
		cfg["send_resolved"] = v.VSendResolved
		if v.HTTPConfig != nil {
			cfg["http_config"] = flattenHTTPConfig(v.HTTPConfig)
		}
		cfg["api_secret"] = v.APISecret
		cfg["api_url"] = v.APIURL
		cfg["corp_id"] = v.CorpID
		cfg["agent_id"] = v.AgentID
		cfg["to_user"] = v.ToUser
		cfg["to_party"] = v.ToParty
		cfg["to_tag"] = v.ToTag
		cfg["message"] = v.Message
		cfg["message_type"] = v.MessageType
		weChatConf = append(weChatConf, cfg)
	}
	return weChatConf
}

func expandEmailConfig(v []interface{}) []*emailConfig {
	var emailConf []*emailConfig

	for _, v := range v {
		cfg := &emailConfig{}
		data := v.(map[string]interface{})

		if raw, ok := data["send_resolved"]; ok {
			cfg.VSendResolved = new(bool)
			*cfg.VSendResolved = raw.(bool)
		}
		if raw, ok := data["to"]; ok {
			cfg.To = raw.(string)
		}
		if raw, ok := data["from"]; ok {
			cfg.From = raw.(string)
		}
		if raw, ok := data["hello"]; ok {
			cfg.Hello = raw.(string)
		}
		if raw, ok := data["smarthost"]; ok {
			var hp config.HostPort
			hp.Host, hp.Port, _ = net.SplitHostPort(raw.(string))
			cfg.Smarthost = hp
		}

		if raw, ok := data["auth_username"]; ok {
			cfg.AuthUsername = raw.(string)
		}
		if raw, ok := data["auth_password"]; ok {
			cfg.AuthPassword = raw.(string)
		}
		if raw, ok := data["auth_secret"]; ok {
			cfg.AuthSecret = raw.(string)
		}
		if raw, ok := data["auth_identity"]; ok {
			cfg.AuthIdentity = raw.(string)
		}
		if raw, ok := data["headers"]; ok {
			cfg.Headers = expandStringMap(raw.(map[string]interface{}))
		}
		if raw, ok := data["html"]; ok {
			cfg.HTML = raw.(string)
		}
		if raw, ok := data["text"]; ok {
			cfg.Text = raw.(string)
		}
		if raw, ok := data["require_tls"]; ok {
			cfg.RequireTLS = new(bool)
			*cfg.RequireTLS = raw.(bool)
		}
		emailConf = append(emailConf, cfg)
	}
	return emailConf
}

func flattenEmailConfig(v []*emailConfig) []interface{} {
	var emailConf []interface{}

	if v == nil {
		return emailConf
	}

	for _, v := range v {
		cfg := make(map[string]interface{})
		cfg["send_resolved"] = v.VSendResolved
		cfg["to"] = v.To
		cfg["from"] = v.From
		cfg["hello"] = v.Hello
		cfg["smarthost"] = v.Smarthost.String()
		cfg["auth_username"] = v.AuthUsername
		cfg["auth_password"] = v.AuthPassword
		cfg["auth_secret"] = v.AuthSecret
		cfg["auth_identity"] = v.AuthIdentity
		cfg["headers"] = v.Headers
		cfg["html"] = v.HTML
		cfg["text"] = v.Text
		cfg["require_tls"] = v.RequireTLS
		emailConf = append(emailConf, cfg)
	}
	return emailConf
}

func expandSlackConfigConfirmationField(v interface{}) *slackConfirmationField {
	var slackConfirmationFieldConf *slackConfirmationField
	data := v.([]interface{})
	if len(data) != 0 && data[0] != nil {
		slackConfirmationFieldConf = &slackConfirmationField{}
		cfg := data[0].(map[string]interface{})
		slackConfirmationFieldConf.Text = cfg["text"].(string)
		slackConfirmationFieldConf.Title = cfg["title"].(string)
		slackConfirmationFieldConf.OkText = cfg["ok_text"].(string)
		slackConfirmationFieldConf.DismissText = cfg["dismiss_text"].(string)
	}
	return slackConfirmationFieldConf
}

func flattenSlackConfigConfirmationField(v *slackConfirmationField) []interface{} {
	slackConfirmationFieldConf := make(map[string]interface{})
	if v != nil {
		slackConfirmationFieldConf["text"] = v.Text
		slackConfirmationFieldConf["title"] = v.Title
		slackConfirmationFieldConf["ok_text"] = v.OkText
		slackConfirmationFieldConf["dismiss_text"] = v.DismissText
	}
	return []interface{}{slackConfirmationFieldConf}
}

func expandSlackConfigFields(v []interface{}) []slackField {
	var slackFieldConf []slackField

	for _, v := range v {
		var cfg slackField
		data := v.(map[string]interface{})

		if raw, ok := data["title"]; ok {
			cfg.Title = raw.(string)
		}
		if raw, ok := data["value"]; ok {
			cfg.Value = raw.(string)
		}
		if raw, ok := data["short"]; ok {
			cfg.Short = raw.(bool)
		}
		slackFieldConf = append(slackFieldConf, cfg)
	}
	return slackFieldConf
}

func flattenSlackConfigFields(v []slackField) []interface{} {
	var slackFieldConf []interface{}

	if v == nil {
		return slackFieldConf
	}

	for _, v := range v {
		cfg := make(map[string]interface{})
		cfg["title"] = v.Title
		cfg["value"] = v.Value
		cfg["short"] = v.Short
		slackFieldConf = append(slackFieldConf, cfg)
	}
	return slackFieldConf
}

func expandSlackConfigActions(v []interface{}) []slackAction {
	var slackActionConf []slackAction

	for _, v := range v {
		var cfg slackAction
		data := v.(map[string]interface{})

		if raw, ok := data["type"]; ok {
			cfg.Type = raw.(string)
		}
		if raw, ok := data["text"]; ok {
			cfg.Text = raw.(string)
		}
		if raw, ok := data["url"]; ok {
			cfg.URL = raw.(string)
		}
		if raw, ok := data["style"]; ok {
			cfg.Style = raw.(string)
		}
		if raw, ok := data["name"]; ok {
			cfg.Name = raw.(string)
		}
		if raw, ok := data["value"]; ok {
			cfg.Value = raw.(string)
		}
		if raw, ok := data["confirm"]; ok {
			cfg.ConfirmField = expandSlackConfigConfirmationField(raw.([]interface{}))
		}
		slackActionConf = append(slackActionConf, cfg)
	}
	return slackActionConf
}

func flattenSlackConfigActions(v []slackAction) []interface{} {
	var slackActionConf []interface{}

	if v == nil {
		return slackActionConf
	}

	for _, v := range v {
		cfg := make(map[string]interface{})
		cfg["type"] = v.Type
		cfg["text"] = v.Text
		cfg["url"] = v.URL
		cfg["style"] = v.Style
		cfg["name"] = v.Name
		cfg["value"] = v.Value
		cfg["confirm"] = flattenSlackConfigConfirmationField(v.ConfirmField)
		slackActionConf = append(slackActionConf, cfg)
	}
	return slackActionConf
}

func expandSlackConfig(v []interface{}) []*slackConfig {
	var slackConf []*slackConfig

	for _, v := range v {
		cfg := &slackConfig{}
		data := v.(map[string]interface{})

		if raw, ok := data["send_resolved"]; ok {
			cfg.VSendResolved = new(bool)
			*cfg.VSendResolved = raw.(bool)
		}
		if raw, ok := data["http_config"]; ok {
			cfg.HTTPConfig = expandHTTPConfig(raw)
		}
		if raw, ok := data["actions"]; ok {
			cfg.Actions = expandSlackConfigActions(raw.([]interface{}))
		}
		if raw, ok := data["fields"]; ok {
			cfg.Fields = expandSlackConfigFields(raw.([]interface{}))
		}
		if raw, ok := data["api_url"]; ok {
			cfg.APIURL = raw.(string)
		}
		if raw, ok := data["channel"]; ok {
			cfg.Channel = raw.(string)
		}
		if raw, ok := data["username"]; ok {
			cfg.Username = raw.(string)
		}
		if raw, ok := data["color"]; ok {
			cfg.Color = raw.(string)
		}
		if raw, ok := data["title"]; ok {
			cfg.Title = raw.(string)
		}
		if raw, ok := data["title_link"]; ok {
			cfg.TitleLink = raw.(string)
		}
		if raw, ok := data["pretext"]; ok {
			cfg.Pretext = raw.(string)
		}
		if raw, ok := data["text"]; ok {
			cfg.Text = raw.(string)
		}
		if raw, ok := data["footer"]; ok {
			cfg.Footer = raw.(string)
		}
		if raw, ok := data["fallback"]; ok {
			cfg.Fallback = raw.(string)
		}
		if raw, ok := data["callback_id"]; ok {
			cfg.CallbackID = raw.(string)
		}
		if raw, ok := data["icon_emoji"]; ok {
			cfg.IconEmoji = raw.(string)
		}
		if raw, ok := data["icon_url"]; ok {
			cfg.IconURL = raw.(string)
		}
		if raw, ok := data["image_url"]; ok {
			cfg.ImageURL = raw.(string)
		}
		if raw, ok := data["thumb_url"]; ok {
			cfg.ThumbURL = raw.(string)
		}
		if raw, ok := data["short_fields"]; ok {
			cfg.ShortFields = raw.(bool)
		}
		if raw, ok := data["link_names"]; ok {
			cfg.LinkNames = raw.(bool)
		}
		if raw, ok := data["mrkdwn_in"]; ok {
			cfg.MrkdwnIn = expandStringArray(raw.([]interface{}))
		}

		slackConf = append(slackConf, cfg)
	}
	return slackConf
}

func flattenSlackConfig(v []*slackConfig) []interface{} {
	var slackConf []interface{}

	if v == nil {
		return slackConf
	}

	for _, v := range v {
		cfg := make(map[string]interface{})
		cfg["send_resolved"] = v.VSendResolved
		cfg["fields"] = flattenSlackConfigFields(v.Fields)
		cfg["actions"] = flattenSlackConfigActions(v.Actions)
		if v.HTTPConfig != nil {
			cfg["http_config"] = flattenHTTPConfig(v.HTTPConfig)
		}
		cfg["api_url"] = v.APIURL
		cfg["channel"] = v.Channel
		cfg["username"] = v.Username
		cfg["color"] = v.Color
		cfg["title"] = v.Title
		cfg["title_link"] = v.TitleLink
		cfg["pretext"] = v.Pretext
		cfg["text"] = v.Text
		cfg["footer"] = v.Footer
		cfg["fallback"] = v.Fallback
		cfg["callback_id"] = v.CallbackID
		cfg["icon_emoji"] = v.IconEmoji
		cfg["icon_url"] = v.IconURL
		cfg["image_url"] = v.ImageURL
		cfg["thumb_url"] = v.ThumbURL
		cfg["short_fields"] = v.ShortFields
		cfg["link_names"] = v.LinkNames
		cfg["mrkdwn_in"] = v.MrkdwnIn
		slackConf = append(slackConf, cfg)
	}
	return slackConf
}

func expandPagerdutyConfigLinks(v []interface{}) []pagerdutyLink {
	var pagerdutyLinkConf []pagerdutyLink

	for _, v := range v {
		var cfg pagerdutyLink
		data := v.(map[string]interface{})

		if raw, ok := data["text"]; ok {
			cfg.Text = raw.(string)
		}
		if raw, ok := data["href"]; ok {
			cfg.Href = raw.(string)
		}
		pagerdutyLinkConf = append(pagerdutyLinkConf, cfg)
	}
	return pagerdutyLinkConf
}

func flattenPagerdutyConfigLinks(v []pagerdutyLink) []interface{} {
	var pagerdutyLinkConf []interface{}

	if v == nil {
		return pagerdutyLinkConf
	}

	for _, v := range v {
		cfg := make(map[string]interface{})
		cfg["text"] = v.Text
		cfg["href"] = v.Href
		pagerdutyLinkConf = append(pagerdutyLinkConf, cfg)
	}
	return pagerdutyLinkConf
}

func expandPagerdutyConfigImages(v []interface{}) []pagerdutyImage {
	var pagerdutyImageConf []pagerdutyImage

	for _, v := range v {
		var cfg pagerdutyImage
		data := v.(map[string]interface{})

		if raw, ok := data["src"]; ok {
			cfg.Src = raw.(string)
		}
		if raw, ok := data["alt"]; ok {
			cfg.Alt = raw.(string)
		}
		if raw, ok := data["href"]; ok {
			cfg.Href = raw.(string)
		}
		pagerdutyImageConf = append(pagerdutyImageConf, cfg)
	}
	return pagerdutyImageConf
}

func flattenPagerdutyConfigImages(v []pagerdutyImage) []interface{} {
	var pagerdutyImageConf []interface{}

	if v == nil {
		return pagerdutyImageConf
	}

	for _, v := range v {
		cfg := make(map[string]interface{})
		cfg["src"] = v.Src
		cfg["alt"] = v.Alt
		cfg["href"] = v.Href
		pagerdutyImageConf = append(pagerdutyImageConf, cfg)
	}
	return pagerdutyImageConf
}

func expandPagerdutyConfig(v []interface{}) []*pagerdutyConfig {
	var pagerdutyConf []*pagerdutyConfig

	for _, v := range v {
		cfg := &pagerdutyConfig{}
		data := v.(map[string]interface{})

		if raw, ok := data["send_resolved"]; ok {
			cfg.VSendResolved = new(bool)
			*cfg.VSendResolved = raw.(bool)
		}
		if raw, ok := data["routing_key"]; ok {
			cfg.RoutingKey = raw.(string)
		}
		if raw, ok := data["service_key"]; ok {
			cfg.ServiceKey = raw.(string)
		}
		if raw, ok := data["url"]; ok {
			cfg.URL = raw.(string)
		}
		if raw, ok := data["http_config"]; ok {
			cfg.HTTPConfig = expandHTTPConfig(raw)
		}
		if raw, ok := data["images"]; ok {
			cfg.Images = expandPagerdutyConfigImages(raw.([]interface{}))
		}
		if raw, ok := data["links"]; ok {
			cfg.Links = expandPagerdutyConfigLinks(raw.([]interface{}))
		}
		if raw, ok := data["client"]; ok {
			cfg.Client = raw.(string)
		}
		if raw, ok := data["client_url"]; ok {
			cfg.ClientURL = raw.(string)
		}
		if raw, ok := data["description"]; ok {
			cfg.Description = raw.(string)
		}
		if raw, ok := data["severity"]; ok {
			cfg.Severity = raw.(string)
		}
		if raw, ok := data["class"]; ok {
			cfg.Class = raw.(string)
		}
		if raw, ok := data["component"]; ok {
			cfg.Component = raw.(string)
		}
		if raw, ok := data["group"]; ok {
			cfg.Group = raw.(string)
		}
		if raw, ok := data["details"]; ok {
			cfg.Details = expandStringMap(raw.(map[string]interface{}))
		}
		pagerdutyConf = append(pagerdutyConf, cfg)
	}
	return pagerdutyConf
}

func flattenPagerdutyConfig(v []*pagerdutyConfig) []interface{} {
	var pagerdutyConf []interface{}

	if v == nil {
		return pagerdutyConf
	}

	for _, v := range v {
		cfg := make(map[string]interface{})
		cfg["send_resolved"] = v.VSendResolved
		cfg["service_key"] = v.ServiceKey
		cfg["routing_key"] = v.RoutingKey
		if v.HTTPConfig != nil {
			cfg["http_config"] = flattenHTTPConfig(v.HTTPConfig)
		}
		cfg["url"] = v.URL
		cfg["client"] = v.Client
		cfg["client_url"] = v.ClientURL
		cfg["description"] = v.Description
		cfg["severity"] = v.Severity
		cfg["class"] = v.Class
		cfg["component"] = v.Component
		cfg["group"] = v.Group
		cfg["details"] = v.Details
		cfg["images"] = flattenPagerdutyConfigImages(v.Images)
		cfg["links"] = flattenPagerdutyConfigLinks(v.Links)
		pagerdutyConf = append(pagerdutyConf, cfg)
	}
	return pagerdutyConf
}

func expandPushoverConfig(v []interface{}) []*pushoverConfig {
	var pushoverConf []*pushoverConfig

	for _, v := range v {
		cfg := &pushoverConfig{}
		data := v.(map[string]interface{})

		if raw, ok := data["send_resolved"]; ok {
			cfg.VSendResolved = new(bool)
			*cfg.VSendResolved = raw.(bool)
		}
		if raw, ok := data["http_config"]; ok {
			cfg.HTTPConfig = expandHTTPConfig(raw)
		}
		if raw, ok := data["user_key"]; ok {
			cfg.UserKey = raw.(string)
		}
		if raw, ok := data["token"]; ok {
			cfg.Token = raw.(string)
		}
		if raw, ok := data["title"]; ok {
			cfg.Title = raw.(string)
		}
		if raw, ok := data["message"]; ok {
			cfg.Message = raw.(string)
		}
		if raw, ok := data["url"]; ok {
			cfg.URL = raw.(string)
		}
		if raw, ok := data["url_title"]; ok {
			cfg.URLTitle = raw.(string)
		}
		if raw, ok := data["sound"]; ok {
			cfg.Sound = raw.(string)
		}
		if raw, ok := data["priority"]; ok {
			cfg.Priority = raw.(string)
		}
		if raw, ok := data["retry"]; ok {
			retry, _ := time.ParseDuration(raw.(string))
			cfg.Retry = retry
		}
		if raw, ok := data["expire"]; ok {
			expire, _ := time.ParseDuration(raw.(string))
			cfg.Expire = expire
		}
		if raw, ok := data["html"]; ok {
			cfg.HTML = raw.(bool)
		}
		pushoverConf = append(pushoverConf, cfg)
	}
	return pushoverConf
}

func flattenPushoverConfig(v []*pushoverConfig) []interface{} {
	var pushoverConf []interface{}

	if v == nil {
		return pushoverConf
	}

	for _, v := range v {
		cfg := make(map[string]interface{})
		cfg["send_resolved"] = v.VSendResolved
		if v.HTTPConfig != nil {
			cfg["http_config"] = flattenHTTPConfig(v.HTTPConfig)
		}
		cfg["user_key"] = v.UserKey
		cfg["token"] = v.Token
		cfg["title"] = v.Title
		cfg["message"] = v.Message
		cfg["url"] = v.URL
		cfg["url_title"] = v.URLTitle
		cfg["sound"] = v.Sound
		cfg["priority"] = v.Priority
		cfg["retry"] = v.Retry.String()
		cfg["expire"] = v.Expire.String()
		cfg["html"] = v.HTML

		pushoverConf = append(pushoverConf, cfg)
	}
	return pushoverConf
}

func expandRouteConfig(v interface{}) *route {
	routeConf := &route{}
	data := v.([]interface{})
	if len(data) != 0 && data[0] != nil {
		cfg := data[0].(map[string]interface{})
		if raw, ok := cfg["receiver"]; ok {
			routeConf.Receiver = raw.(string)
		}
		if raw, ok := cfg["group_by"]; ok {
			routeConf.GroupByStr = expandStringArray(raw.([]interface{}))
		}
		if raw, ok := cfg["matchers"]; ok {
			routeConf.Matchers = expandStringArray(raw.([]interface{}))
		}
		if raw, ok := cfg["child_route"]; ok {
			var routes []*route
			for _, item := range raw.([]interface{}) {
				routes = append(routes, expandChildRouteConfig([]interface{}{item.(map[string]interface{})}))
			}
			routeConf.Routes = routes
		}
		if raw, ok := cfg["group_wait"]; ok {
			routeConf.GroupWait = raw.(string)
		}
		if raw, ok := cfg["group_interval"]; ok {
			routeConf.GroupInterval = raw.(string)
		}
		if raw, ok := cfg["repeat_interval"]; ok {
			routeConf.RepeatInterval = raw.(string)
		}
		if raw, ok := cfg["mute_time_intervals"]; ok {
			routeConf.MuteTimeIntervals = expandStringArray(raw.([]interface{}))
		}
		if raw, ok := cfg["active_time_intervals"]; ok {
			routeConf.ActiveTimeIntervals = expandStringArray(raw.([]interface{}))
		}
	}
	return routeConf
}

func expandChildRouteConfig(v interface{}) *route {
	routeConf := &route{}
	data := v.([]interface{})
	if len(data) != 0 && data[0] != nil {
		cfg := data[0].(map[string]interface{})
		if raw, ok := cfg["receiver"]; ok {
			routeConf.Receiver = raw.(string)
		}
		if raw, ok := cfg["group_by"]; ok {
			routeConf.GroupByStr = expandStringArray(raw.([]interface{}))
		}
		if raw, ok := cfg["matchers"]; ok {
			routeConf.Matchers = expandStringArray(raw.([]interface{}))
		}
		if raw, ok := cfg["continue"]; ok {
			routeConf.Continue = raw.(bool)
		}
		if raw, ok := cfg["group_wait"]; ok {
			routeConf.GroupWait = raw.(string)
		}
		if raw, ok := cfg["group_interval"]; ok {
			routeConf.GroupInterval = raw.(string)
		}
		if raw, ok := cfg["repeat_interval"]; ok {
			routeConf.RepeatInterval = raw.(string)
		}
		if raw, ok := cfg["mute_time_intervals"]; ok {
			routeConf.MuteTimeIntervals = expandStringArray(raw.([]interface{}))
		}
		if raw, ok := cfg["active_time_intervals"]; ok {
			routeConf.ActiveTimeIntervals = expandStringArray(raw.([]interface{}))
		}
	}
	return routeConf
}

func flattenRouteConfig(v *route) []interface{} {
	routeConf := make(map[string]interface{})

	routeConf["receiver"] = v.Receiver

	if len(v.GroupByStr) > 0 {
		routeConf["group_by"] = v.GroupByStr
	}

	if len(v.Matchers) > 0 {
		routeConf["matchers"] = v.Matchers
	}

	if v.Routes != nil {
		var routes []interface{}
		for _, route := range v.Routes {
			routes = append(routes, flattenChildRouteConfig(route)[0])
		}
		routeConf["child_route"] = routes
	}
	routeConf["group_wait"] = v.GroupWait
	routeConf["group_interval"] = v.GroupInterval
	routeConf["repeat_interval"] = v.RepeatInterval

	if len(v.MuteTimeIntervals) > 0 {
		routeConf["mute_time_intervals"] = v.MuteTimeIntervals
	}
	if len(v.ActiveTimeIntervals) > 0 {
		routeConf["active_time_intervals"] = v.ActiveTimeIntervals
	}

	return []interface{}{routeConf}
}

func flattenChildRouteConfig(v *route) []interface{} {
	routeConf := make(map[string]interface{})

	routeConf["receiver"] = v.Receiver

	if len(v.GroupByStr) > 0 {
		routeConf["group_by"] = v.GroupByStr
	}

	if len(v.Matchers) > 0 {
		routeConf["matchers"] = v.Matchers
	}

	routeConf["continue"] = v.Continue
	routeConf["group_wait"] = v.GroupWait
	routeConf["group_interval"] = v.GroupInterval
	routeConf["repeat_interval"] = v.RepeatInterval

	if len(v.MuteTimeIntervals) > 0 {
		routeConf["mute_time_intervals"] = v.MuteTimeIntervals
	}
	if len(v.ActiveTimeIntervals) > 0 {
		routeConf["active_time_intervals"] = v.ActiveTimeIntervals
	}

	return []interface{}{routeConf}
}

func expandInhibitRuleConfig(v []interface{}) []*inhibitRule {
	var inhibitRuleConf []*inhibitRule

	for _, v := range v {
		cfg := &inhibitRule{}
		data := v.(map[string]interface{})

		if raw, ok := data["source_matchers"]; ok {
			cfg.SourceMatchers = expandStringArray(raw.([]interface{}))
		}
		if raw, ok := data["target_matchers"]; ok {
			cfg.TargetMatchers = expandStringArray(raw.([]interface{}))
		}
		if raw, ok := data["equal"]; ok {
			cfg.Equal = expandStringArray(raw.([]interface{}))
		}
		inhibitRuleConf = append(inhibitRuleConf, cfg)
	}
	return inhibitRuleConf
}

func flattenInhibitRuleConfig(v []*inhibitRule) []interface{} {
	var inhibitRuleConf []interface{}

	if v == nil {
		return inhibitRuleConf
	}

	for _, v := range v {
		cfg := make(map[string]interface{})
		cfg["source_matchers"] = v.SourceMatchers
		cfg["target_matchers"] = v.TargetMatchers
		cfg["equal"] = v.Equal
		inhibitRuleConf = append(inhibitRuleConf, cfg)
	}
	return inhibitRuleConf
}

func expandMuteTimeIntervalConfig(v []interface{}) []*muteTimeInterval {
	var muteTimeIntervalConf []*muteTimeInterval

	for _, v := range v {
		cfg := &muteTimeInterval{}
		data := v.(map[string]interface{})

		if raw, ok := data["name"]; ok {
			cfg.Name = raw.(string)
		}
		if raw, ok := data["time_intervals"]; ok {
			cfg.TimeIntervals = expandTimeIntervalConfig(raw.([]interface{}))
		}
		muteTimeIntervalConf = append(muteTimeIntervalConf, cfg)
	}
	return muteTimeIntervalConf
}

func flattenMuteTimeIntervalConfig(v []*muteTimeInterval) []interface{} {
	var muteTimeIntervalConf []interface{}

	if v == nil {
		return muteTimeIntervalConf
	}

	for _, v := range v {
		cfg := make(map[string]interface{})
		cfg["name"] = v.Name
		cfg["time_intervals"] = flattenTimeIntervalConfig(v.TimeIntervals)
		muteTimeIntervalConf = append(muteTimeIntervalConf, cfg)
	}
	return muteTimeIntervalConf
}

func expandTimeIntervalConfig(v []interface{}) []timeinterval.TimeInterval {
	var timeIntervalConf []timeinterval.TimeInterval

	for _, v := range v {
		var cfg timeinterval.TimeInterval
		data := v.(map[string]interface{})

		if raw, ok := data["times"]; ok {
			cfg.Times = expandTimeRange(raw.([]interface{}))
		}
		if raw, ok := data["weekdays"]; ok {
			cfg.Weekdays = expandWeekdayRange(raw.([]interface{}))
		}
		if raw, ok := data["days_of_month"]; ok {
			cfg.DaysOfMonth = expandDayOfMonthRange(raw.([]interface{}))
		}
		if raw, ok := data["months"]; ok {
			cfg.Months = expandMonthRange(raw.([]interface{}))
		}
		if raw, ok := data["years"]; ok {
			cfg.Years = expandYearRange(raw.([]interface{}))
		}
		if raw, ok := data["location"]; ok {
			loc, _ := time.LoadLocation(raw.(string))
			cfg.Location = new(timeinterval.Location)
			*cfg.Location = timeinterval.Location{Location: loc}
		}
		timeIntervalConf = append(timeIntervalConf, cfg)
	}
	return timeIntervalConf
}

func expandTimeRange(v []interface{}) []timeinterval.TimeRange {
	var timeRangeConf []timeinterval.TimeRange

	for _, v := range v {
		var cfg timeinterval.TimeRange
		data := v.(map[string]interface{})

		if raw, ok := data["start_minute"]; ok {
			cfg.StartMinute = raw.(int)
		}
		if raw, ok := data["end_minute"]; ok {
			cfg.EndMinute = raw.(int)
		}

		timeRangeConf = append(timeRangeConf, cfg)
	}
	return timeRangeConf
}

func flattenTimeRange(v []timeinterval.TimeRange) []interface{} {
	var timeRangeConf []interface{}

	if v == nil {
		return timeRangeConf
	}

	for _, v := range v {
		cfg := make(map[string]interface{})
		cfg["start_minute"] = v.StartMinute
		cfg["end_minute"] = v.EndMinute
		timeRangeConf = append(timeRangeConf, cfg)
	}
	return timeRangeConf
}

func expandWeekdayRange(v []interface{}) []timeinterval.WeekdayRange {
	var inclusiveRangeConf []timeinterval.WeekdayRange

	for _, v := range v {
		var cfg timeinterval.WeekdayRange
		data := v.(map[string]interface{})

		if raw, ok := data["begin"]; ok {
			cfg.Begin = raw.(int)
		}
		if raw, ok := data["end"]; ok {
			cfg.End = raw.(int)
		}

		inclusiveRangeConf = append(inclusiveRangeConf, cfg)
	}
	return inclusiveRangeConf
}

func flattenWeekdayRange(v []timeinterval.WeekdayRange) []interface{} {
	var inclusiveRangeConf []interface{}

	if v == nil {
		return inclusiveRangeConf
	}

	for _, v := range v {
		cfg := make(map[string]interface{})
		cfg["begin"] = v.Begin
		cfg["end"] = v.End
		inclusiveRangeConf = append(inclusiveRangeConf, cfg)
	}
	return inclusiveRangeConf
}

func expandDayOfMonthRange(v []interface{}) []timeinterval.DayOfMonthRange {
	var inclusiveRangeConf []timeinterval.DayOfMonthRange

	for _, v := range v {
		var cfg timeinterval.DayOfMonthRange
		data := v.(map[string]interface{})

		if raw, ok := data["begin"]; ok {
			cfg.Begin = raw.(int)
		}
		if raw, ok := data["end"]; ok {
			cfg.End = raw.(int)
		}

		inclusiveRangeConf = append(inclusiveRangeConf, cfg)
	}
	return inclusiveRangeConf
}

func flattenDayOfMonthRange(v []timeinterval.DayOfMonthRange) []interface{} {
	var inclusiveRangeConf []interface{}

	if v == nil {
		return inclusiveRangeConf
	}

	for _, v := range v {
		cfg := make(map[string]interface{})
		cfg["begin"] = v.Begin
		cfg["end"] = v.End
		inclusiveRangeConf = append(inclusiveRangeConf, cfg)
	}
	return inclusiveRangeConf
}

func expandMonthRange(v []interface{}) []timeinterval.MonthRange {
	var inclusiveRangeConf []timeinterval.MonthRange

	for _, v := range v {
		var cfg timeinterval.MonthRange
		data := v.(map[string]interface{})

		if raw, ok := data["begin"]; ok {
			cfg.Begin = raw.(int)
		}
		if raw, ok := data["end"]; ok {
			cfg.End = raw.(int)
		}

		inclusiveRangeConf = append(inclusiveRangeConf, cfg)
	}
	return inclusiveRangeConf
}

func flattenMonthRange(v []timeinterval.MonthRange) []interface{} {
	var inclusiveRangeConf []interface{}

	if v == nil {
		return inclusiveRangeConf
	}

	for _, v := range v {
		cfg := make(map[string]interface{})
		cfg["begin"] = v.Begin
		cfg["end"] = v.End
		inclusiveRangeConf = append(inclusiveRangeConf, cfg)
	}
	return inclusiveRangeConf
}

func expandYearRange(v []interface{}) []timeinterval.YearRange {
	var inclusiveRangeConf []timeinterval.YearRange

	for _, v := range v {
		var cfg timeinterval.YearRange
		data := v.(map[string]interface{})

		if raw, ok := data["begin"]; ok {
			cfg.Begin = raw.(int)
		}
		if raw, ok := data["end"]; ok {
			cfg.End = raw.(int)
		}

		inclusiveRangeConf = append(inclusiveRangeConf, cfg)
	}
	return inclusiveRangeConf
}

func flattenYearRange(v []timeinterval.YearRange) []interface{} {
	var inclusiveRangeConf []interface{}

	if v == nil {
		return inclusiveRangeConf
	}

	for _, v := range v {
		cfg := make(map[string]interface{})
		cfg["begin"] = v.Begin
		cfg["end"] = v.End
		inclusiveRangeConf = append(inclusiveRangeConf, cfg)
	}
	return inclusiveRangeConf
}

func flattenTimeIntervalConfig(v []timeinterval.TimeInterval) []interface{} {
	var timeIntervalConf []interface{}

	if v == nil {
		return timeIntervalConf
	}

	for _, v := range v {
		cfg := make(map[string]interface{})
		cfg["times"] = flattenTimeRange(v.Times)
		cfg["weekdays"] = flattenWeekdayRange(v.Weekdays)
		cfg["days_of_month"] = flattenDayOfMonthRange(v.DaysOfMonth)
		cfg["months"] = flattenMonthRange(v.Months)
		cfg["years"] = flattenYearRange(v.Years)
		cfg["location"] = v.Location.String()
		timeIntervalConf = append(timeIntervalConf, cfg)
	}
	return timeIntervalConf
}
