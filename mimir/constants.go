package mimir

const (
	orgIDKey          = "org_id"
	orgIDDescription  = "The Organization ID. If not set, the Org ID defined in the provider block will be used."
	namespaceKey      = "namespace"
	defaultNamespace  = "default"
	intervalKey       = "interval"
	labelsKey         = "labels"
	contentTypeHeader = "Content-Type"
	contentTypeYAML   = "application/yaml"

	receiverKey       = "receiver"
	sendResolvedKey   = "send_resolved"
	sendResolvedDescr = "Whether to notify about resolved alerts."
	titleKey          = "title"
	httpConfigKey     = "http_config"
	httpConfigDescr   = "The HTTP client's configuration."
	apiURLKey         = "api_url"
	messageKey        = "message"
	beginKey          = "begin"
	timeoutKey        = "timeout"
	usernameKey       = "username"
	contentKey        = "content"
	sourceTenantsKey  = "source_tenants"
	queryOffsetKey    = "query_offset"

	alignEvaluationTimeOnIntervalKey   = "align_evaluation_time_on_interval"
	alignEvaluationTimeOnIntervalDescr = "Whether to align the group's evaluation timestamps on multiples of its interval."
	ruleGroupLimitKey                  = "limit"
	ruleGroupLimitDescr                = "Maximum number of alerts or series a rule in the group may produce; 0 means no limit. Requires Mimir >= 3.0.0 to be persisted (older Mimir accepts but drops it)."
	queryOffsetDescr                   = "The duration by which to delay the evaluation of the rules in the group."

	// mimirMin31Descr flags attributes backed by an Alertmanager field that Mimir
	// only vendors from 3.1.0 onwards. Mimir rejects a whole configuration when it
	// carries an unknown key, so setting one of these against an older Mimir fails.
	mimirMin31Descr = " Requires Mimir >= 3.1.0."
	mimirMin32Descr = " Requires Mimir >= 3.2.0."
)
