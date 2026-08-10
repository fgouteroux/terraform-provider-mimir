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

	// mimirMin31Descr flags attributes backed by an Alertmanager field that Mimir
	// only vendors from 3.1.0 onwards. Mimir rejects a whole configuration when it
	// carries an unknown key, so setting one of these against an older Mimir fails.
	mimirMin31Descr = " Requires Mimir >= 3.1.0."
)
