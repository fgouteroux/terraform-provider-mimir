package mimir

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/prometheus/alertmanager/matcher/compat"
	"github.com/prometheus/common/model"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/prometheus/model/rulefmt"
	"github.com/prometheus/prometheus/promql/parser"
	"gopkg.in/yaml.v3"
)

var (
	labelNameRegexp  = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	metricNameRegexp = regexp.MustCompile(`^[a-zA-Z_:][a-zA-Z0-9_:]*$`)
	validTime        = "^((([01][0-9])|(2[0-3])):[0-5][0-9])$|(^24:00$)"
	validTimeRE      = regexp.MustCompile(validTime)

	promQLParserOptions = parser.Options{
		EnableExperimentalFunctions: true,
		EnableBinopFillModifiers:    true,
	}
)

// hasControlChar reports whether s contains any Unicode control character (category Cc:
// the ASCII C0 controls, DEL, and C1 controls such as NEL U+0085). Keeps control bytes out
// of names, namespaces, and the X-Scope-OrgID header. unicode.IsControl is used because the
// ASCII regex class [:cntrl:] would miss the C1 range.
func hasControlChar(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}

// validRuleName is the permissive validity guard for alert and rule-group names. It accepts
// a name that is non-empty, not whitespace-only, not the "."/".." dot-segment, and free of
// (Unicode) control characters and '/'. Dot-segments are rejected because routers and proxies
// normalize them as path traversal, breaking the read/update/delete round-trip.
func validRuleName(s string) bool {
	if s == "." || s == ".." {
		return false
	}
	return strings.TrimSpace(s) != "" && !strings.ContainsRune(s, '/') && !hasControlChar(s)
}

// validNamespace is the narrower namespace guard: reject '/', control characters, and the
// "."/".." dot-segments; empty is allowed (the field defaults to "default"). Rejecting '/'
// is a security requirement: a '/' in a namespace collides with the "/"-joined Terraform
// resource ID and would fabricate an X-Scope-OrgID tenant, causing cross-tenant read/delete.
func validNamespace(s string) bool {
	if s == "." || s == ".." {
		return false
	}
	return !strings.ContainsRune(s, '/') && !hasControlChar(s)
}

// validOrgID applies the same rule as validNamespace: reject '/', control characters, and
// the "."/".." dot-segments (empty is allowed). org_id becomes the X-Scope-OrgID tenant and
// is used by Mimir as a storage-path component, so a '/' or "."/".." is a cross-tenant
// path-traversal/collision primitive — and a '/' would also break the "orgID/namespace/name"
// Terraform ID. (Control chars are additionally rejected to prevent header injection.)
func validOrgID(s string) bool {
	if s == "." || s == ".." {
		return false
	}
	return !strings.ContainsRune(s, '/') && !hasControlChar(s)
}

// rulesGroupPath builds a ruler-API path that addresses a rule group by name, with each
// segment URL-path-escaped so a relaxed name (spaces, %, #, ?) survives read/delete instead
// of mis-decoding. Escaping happens here, never inside sendRequest.
func rulesGroupPath(namespace, name string) string {
	return "/config/v1/rules/" + url.PathEscape(namespace) + "/" + url.PathEscape(name)
}

// rulesNamespacePath builds the namespace-only ruler-API path used by create/update (the
// group name travels in the request body there), with the namespace segment escaped.
func rulesNamespacePath(namespace string) string {
	return "/config/v1/rules/" + url.PathEscape(namespace)
}

// buildRuleGroupID encodes the typed-resource Terraform ID with each segment URL-escaped,
// so a delimiter-colliding character in a segment cannot shift the parse. An empty orgID
// yields the 2-segment form.
func buildRuleGroupID(orgID, namespace, name string) string {
	if orgID != "" {
		return url.PathEscape(orgID) + "/" + url.PathEscape(namespace) + "/" + url.PathEscape(name)
	}
	return url.PathEscape(namespace) + "/" + url.PathEscape(name)
}

// parseRuleGroupID splits and unescapes a typed-resource Terraform ID and re-validates every
// segment. A raw '/' is always the delimiter, so a literal '/' inside a value can only appear
// as %2F, which decodes and is then rejected by validNamespace — closing the terraform-import
// and legacy-state cross-tenant vectors that schema ValidateFuncs never see. It rejects an ID
// that PathUnescape cannot decode, or whose orgID/namespace/name fail their guards.
func parseRuleGroupID(id string) (orgID, namespace, name string, err error) {
	idArr := strings.Split(id, "/")
	var raw [3]string // orgID, namespace, name
	switch len(idArr) {
	case 2:
		raw = [3]string{"", idArr[0], idArr[1]}
	case 3:
		raw = [3]string{idArr[0], idArr[1], idArr[2]}
	default:
		return "", "", "", fmt.Errorf("invalid id format: expected 'namespace/name' or 'org_id/namespace/name', got %q", id)
	}
	var dec [3]string
	for i, seg := range raw {
		d, uerr := url.PathUnescape(seg)
		if uerr != nil {
			// Legacy IDs were stored unescaped, so a bare '%' (e.g. namespace "50% teams")
			// is not valid percent-encoding. Fall back to the raw segment rather than
			// hard-failing Read; the guards below still reject '/'/control/dot-segments,
			// and rulesGroupPath re-escapes the value on the wire.
			d = seg
		}
		dec[i] = d
	}
	orgID, namespace, name = dec[0], dec[1], dec[2]
	if !validOrgID(orgID) {
		return "", "", "", fmt.Errorf("invalid id %q: org_id must contain no control characters", id)
	}
	if !validNamespace(namespace) {
		return "", "", "", fmt.Errorf("invalid id %q: namespace %q must not be \".\"/\"..\" and must contain no control characters or '/'", id, namespace)
	}
	if !validRuleName(name) {
		return "", "", "", fmt.Errorf("invalid id %q: name %q must be non-empty, not whitespace-only, not \".\"/\"..\", and contain no control characters or '/'", id, name)
	}
	return orgID, namespace, name, nil
}

func handleHTTPError(err error, baseMsg string) error {
	if err != nil {
		return fmt.Errorf("%s %v", baseMsg, err)
	}

	return nil
}

// Array to String Array
func expandStringArray(v []interface{}) []string {
	var m []string
	for _, val := range v {
		m = append(m, val.(string))
	}

	return m
}

// Map to String Map
func expandStringMap(v map[string]interface{}) map[string]string {
	m := make(map[string]string)
	for key, val := range v {
		m[key] = val.(string)
	}

	return m
}

func validateGroupRuleName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if !validRuleName(value) {
		errors = append(errors, fmt.Errorf(
			"\"%s\": Invalid Group Rule Name %q: must be non-empty, not whitespace-only, not \".\" or \"..\", and contain no control characters or '/'", k, value))
	}

	return
}

// pctEncodedSeqRegexp matches a valid percent-encoding sequence (e.g. "%20"). Used only to
// WARN: older provider versions interpolated the namespace raw into the ruler URL, so the
// server percent-decoded it once (HCL "team%20a" silently addressed namespace "team a").
// Values are now sent literally, matching mimirtool's behavior.
var pctEncodedSeqRegexp = regexp.MustCompile(`%[0-9a-fA-F]{2}`)

func validateNamespace(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if !validNamespace(value) {
		errors = append(errors, fmt.Errorf(
			"\"%s\": Invalid Namespace %q: must not be \".\" or \"..\" and must contain no control characters or '/'", k, value))
		return
	}

	if pctEncodedSeqRegexp.MatchString(value) {
		if decoded, err := url.PathUnescape(value); err == nil && decoded != value {
			ws = append(ws, fmt.Sprintf(
				"%q: namespace %q contains a percent-encoding sequence and is now sent literally (matching mimirtool); older provider versions let the server decode it once. If you meant %q, use that value instead — it addresses the same server-side namespace as before.",
				k, value, decoded))
		}
	}

	return
}

func validateOrgID(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if !validOrgID(value) {
		errors = append(errors, fmt.Errorf(
			"\"%s\": Invalid Org ID %q: must not be \".\"/\"..\" and must contain no control characters or '/'", k, value))
		return
	}
	if value == "" {
		return
	}
	for _, tenantID := range strings.Split(value, "|") {
		if err := mimirTenantIDError(tenantID); err != nil {
			errors = append(errors, fmt.Errorf("\"%s\": Invalid Org ID %q: %v", k, value, err))
		}
	}

	return
}

// maxTenantIDLength and tenantIDSpecialChars mirror dskit's tenant.ValidTenantID,
// which Mimir applies to every X-Scope-OrgID (each "|"-separated tenant of a
// federated request).
const (
	maxTenantIDLength    = 150
	tenantIDSpecialChars = "!-_.*'()"
)

func mimirTenantIDError(tenantID string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant ID must not be empty")
	}
	for i := 0; i < len(tenantID); i++ {
		c := tenantID[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || strings.IndexByte(tenantIDSpecialChars, c) >= 0 {
			continue
		}
		return fmt.Errorf("tenant ID %q contains unsupported character %q at position %d: only letters, digits and %s are allowed", tenantID, c, i, tenantIDSpecialChars)
	}
	if len(tenantID) > maxTenantIDLength {
		return fmt.Errorf("tenant ID %q is too long: max %d characters", tenantID, maxTenantIDLength)
	}
	if tenantID == "." || tenantID == ".." {
		return fmt.Errorf("tenant ID must not be '.' or '..'")
	}
	return nil
}

func validatePromQLExpr(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if _, err := parser.NewParser(promQLParserOptions).ParseExpr(value); err != nil {
		errors = append(errors, fmt.Errorf(
			"\"%s\": Invalid PromQL expression %q: %v", k, value, err))
	}

	return
}

func validateLabels(v interface{}, k string) (ws []string, errors []error) {
	m := v.(map[string]interface{})
	for lname, lvalue := range m {
		if !labelNameRegexp.MatchString(lname) {
			errors = append(errors, fmt.Errorf(
				"\"%s\": Invalid Label Name %q. Must match the regex %s", k, lname, labelNameRegexp))
		}

		if !utf8.ValidString(lvalue.(string)) {
			errors = append(errors, fmt.Errorf(
				"\"%s\": Invalid Label Value %q: not a valid UTF8 string", k, lvalue))
		}
	}
	return
}

func validateAnnotations(v interface{}, k string) (ws []string, errors []error) {
	m := v.(map[string]interface{})
	for aname := range m {
		if !labelNameRegexp.MatchString(aname) {
			errors = append(errors, fmt.Errorf(
				"\"%s\": Invalid Annotation Name %q. Must match the regex %s", k, aname, labelNameRegexp))
		}
	}
	return
}

func validateDuration(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if value == "" {
		return
	}

	if _, err := model.ParseDuration(value); err != nil {
		errors = append(errors, fmt.Errorf("\"%s\": %v", k, err))
	}

	return
}

func formatDuration(v interface{}) string {
	value, _ := model.ParseDuration(v.(string))
	str := value.String()
	if str == "0s" {
		return ""
	}
	return str
}

// flattenDuration renders a duration held as a time.Duration back into state.
//
// Configured values reach state through formatDuration, which normalizes with
// model.Duration. Rendering with time.Duration.String() instead disagrees
// whenever the value carries trailing zero sub-units or exceeds a day ("1h" vs
// "1h0m0s", "1d" vs "24h0m0s"), which shows up as a permanent diff. Zero maps to
// the empty string for the same reason formatDuration does it: an unset optional
// attribute must not read back as "0s".
func flattenDuration(d time.Duration) string {
	if d == 0 {
		return ""
	}
	return model.Duration(d).String()
}

func formatPromQLExpr(v interface{}) string {
	if enablePromQLExprFormat {
		value, _ := parser.NewParser(promQLParserOptions).ParseExpr(v.(string))
		// remove spaces causing decoding issues with multiline yaml marshal/unmarshall
		return strings.TrimLeft(parser.Prettify(value), " ")
	}
	return v.(string)
}

// Converts a string of the form "HH:MM" into the number of minutes elapsed in the day.
func parseTime(in string) (mins int) {
	// start_time and end_time are both Optional and validateTime returns early
	// for "", so an unset half of a times block reaches this function as an
	// empty string. Indexing [1] unconditionally then panics the provider.
	timestampComponents := strings.Split(in, ":")
	if len(timestampComponents) != 2 {
		return 0
	}
	timeStampHours, _ := strconv.Atoi(timestampComponents[0])
	timeStampMinutes, _ := strconv.Atoi(timestampComponents[1])

	// Timestamps are stored as minutes elapsed in the day, so multiply hours by 60.
	mins = timeStampHours*60 + timeStampMinutes
	return mins
}

func validateTime(v interface{}, k string) (ws []string, errors []error) {
	in := v.(string)

	if in == "" {
		return
	}

	if !validTimeRE.MatchString(in) {
		errors = append(errors, fmt.Errorf("\"%s\": couldn't parse timestamp %s, invalid format", k, in))
		return
	}
	timestampComponents := strings.Split(in, ":")
	if len(timestampComponents) != 2 {
		errors = append(errors, fmt.Errorf("\"%s\": invalid timestamp format: %s", k, in))
		return
	}
	timeStampHours, err := strconv.Atoi(timestampComponents[0])
	if err != nil {
		errors = append(errors, fmt.Errorf("\"%s\": %v", k, err))
	}
	timeStampMinutes, err := strconv.Atoi(timestampComponents[1])
	if err != nil {
		errors = append(errors, fmt.Errorf("\"%s\": %v", k, err))
	}
	if timeStampHours < 0 || timeStampHours > 24 || timeStampMinutes < 0 || timeStampMinutes > 60 {
		errors = append(errors, fmt.Errorf("\"%s\": timestamp %s out of range", k, in))
	}
	return
}

// validateYAMLDocument checks that a string attribute holds a YAML (or JSON)
// document, so a malformed webhook payload fails at plan time.
func validateYAMLDocument(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if value == "" {
		return
	}
	var doc interface{}
	if err := yaml.Unmarshal([]byte(value), &doc); err != nil {
		errors = append(errors, fmt.Errorf("%q: invalid YAML or JSON document: %v", k, err))
	}
	return
}

// expandYAMLDocument decodes a YAML/JSON attribute into the structure the
// Alertmanager configuration embeds. validateYAMLDocument has already rejected
// malformed input at plan time, so a decode error here is only logged.
func expandYAMLDocument(value string) interface{} {
	if value == "" {
		return nil
	}
	var doc interface{}
	if err := yaml.Unmarshal([]byte(value), &doc); err != nil {
		log.Printf("[WARN] cannot decode YAML document, sending it unset: %v", err)
		return nil
	}
	return doc
}

// flattenYAMLDocument renders an embedded structure back to the YAML string
// kept in state. suppressEquivalentYAML hides the formatting difference with
// the configured value (JSON, key order, indentation).
func flattenYAMLDocument(doc interface{}) string {
	if doc == nil {
		return ""
	}
	out, err := yaml.Marshal(doc)
	if err != nil {
		log.Printf("[WARN] cannot encode YAML document: %v", err)
		return ""
	}
	return string(out)
}

// suppressEquivalentYAML reports whether two YAML/JSON documents decode to the
// same value, so reformatting alone never shows as a diff.
func suppressEquivalentYAML(_, oldValue, newValue string, _ *schema.ResourceData) bool {
	if oldValue == newValue {
		return true
	}
	var oldDoc, newDoc interface{}
	if err := yaml.Unmarshal([]byte(oldValue), &oldDoc); err != nil {
		return false
	}
	if err := yaml.Unmarshal([]byte(newValue), &newDoc); err != nil {
		return false
	}
	return reflect.DeepEqual(oldDoc, newDoc)
}

// mimirRuleErrors runs the rule checks Mimir's ruler applies on upload
// (rulefmt.Rule.Validate with Mimir's parser options and the default legacy
// name validation scheme): PromQL parsing, record/alert exclusivity, fields a
// recording rule may not carry, label and annotation names, and the parsing of
// alert label and annotation templates. Durations are validated by the caller.
func mimirRuleErrors(rule rulefmt.Rule) []error {
	var errs []error
	for _, we := range rule.Validate(rulefmt.RuleNode{}, model.LegacyValidation, parser.NewParser(promQLParserOptions)) {
		errs = append(errs, we.Unwrap())
	}
	return errs
}

// Rule block attributes of each single-group resource, as read by
// customizeDiffMimirRules. Reading an attribute a schema does not declare
// panics, so each resource passes its own list.
var (
	alertingRuleAttributes  = []string{"alert", "expr", "for", "keep_firing_for", labelsKey, "annotations"}
	recordingRuleAttributes = []string{"record", "expr", labelsKey}
)

// customizeDiffMimirRules runs mimirRuleErrors on every rule block of a
// single-group resource, so what the per-attribute validators cannot see (alert
// templates in labels and annotations, the reserved __name__ label) fails at
// plan instead of at apply. A rule holding a value not known until apply is
// skipped; its per-attribute validators still run.
func customizeDiffMimirRules(attributes []string) schema.CustomizeDiffFunc {
	declared := map[string]bool{}
	for _, attr := range attributes {
		declared[attr] = true
	}
	return func(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
		var errs []error
		for i := 0; i < diff.Get("rule.#").(int); i++ {
			if err := validateRuleBlock(diff, i, attributes, declared); err != nil {
				errs = append(errs, err)
			}
		}
		return errors.Join(errs...)
	}
}

func validateRuleBlock(diff *schema.ResourceDiff, i int, attributes []string, declared map[string]bool) error {
	prefix := fmt.Sprintf("rule.%d.", i)
	for _, attr := range attributes {
		if !diff.NewValueKnown(prefix + attr) {
			return nil
		}
	}

	get := func(attr string) (interface{}, bool) {
		if !declared[attr] {
			return nil, false
		}
		return diff.GetOk(prefix + attr)
	}

	rule := rulefmt.Rule{Expr: diff.Get(prefix + "expr").(string)}
	name := ""
	if v, ok := get("alert"); ok {
		rule.Alert = v.(string)
		name = rule.Alert
	}
	if v, ok := get("record"); ok {
		rule.Record = v.(string)
		name = rule.Record
	}
	for attr, target := range map[string]*model.Duration{"for": &rule.For, "keep_firing_for": &rule.KeepFiringFor} {
		v, ok := get(attr)
		if !ok {
			continue
		}
		d, err := model.ParseDuration(v.(string))
		if err != nil {
			// Already reported by the attribute's own ValidateFunc.
			continue
		}
		*target = d
	}
	if v, ok := get(labelsKey); ok {
		rule.Labels = expandStringMap(v.(map[string]interface{}))
	}
	if v, ok := get("annotations"); ok {
		rule.Annotations = expandStringMap(v.(map[string]interface{}))
	}

	var errs []error
	for _, err := range mimirRuleErrors(rule) {
		errs = append(errs, fmt.Errorf("rule %d (%s): %w", i, name, err))
	}
	return errors.Join(errs...)
}

// alertmanagerMatchersParser parses matchers the way Mimir does by default:
// UTF-8 strict mode is off, so Alertmanager runs in fallback mode and accepts
// a matcher that either the UTF-8 or the classic parser accepts.
var alertmanagerMatchersParser = compat.FallbackMatchersParser(promslog.NewNopLogger())

// validateAlertmanagerMatchers rejects a matchers entry Mimir would refuse.
func validateAlertmanagerMatchers(v interface{}, k string) (ws []string, errs []error) {
	value := v.(string)
	if _, err := alertmanagerMatchersParser(value, "config"); err != nil {
		errs = append(errs, fmt.Errorf("%q: invalid matcher %q: %v", k, value, err))
	}
	return
}

// validateTemplateName mirrors Mimir's validateTemplateFilename: a template
// name must be a bare file name, since Mimir writes templates into a per-tenant
// directory and refuses anything that could point outside it.
func validateTemplateName(v interface{}, k string) (ws []string, errs []error) {
	name := v.(string)
	if filepath.Base(name) != name || filepath.Dir(filepath.Clean(name)) != "." {
		errs = append(errs, fmt.Errorf("%q: invalid template name %q: the template name cannot contain any path", k, name))
	}
	return
}

// validateTemplateFiles applies validateTemplateName to every templates_files key.
func validateTemplateFiles(v interface{}, k string) (ws []string, errs []error) {
	for name := range v.(map[string]interface{}) {
		_, nameErrs := validateTemplateName(name, k)
		errs = append(errs, nameErrs...)
	}
	return
}
