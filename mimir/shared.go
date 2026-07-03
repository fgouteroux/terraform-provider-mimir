package mimir

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/promql/parser"
)

var (
	labelNameRegexp  = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	metricNameRegexp = regexp.MustCompile(`^[a-zA-Z_:][a-zA-Z0-9_:]*$`)
	validTime        = "^((([01][0-9])|(2[0-3])):[0-5][0-9])$|(^24:00$)"
	validTimeRE      = regexp.MustCompile(validTime)
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
	}

	return
}

func validatePromQLExpr(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if _, err := parser.NewParser(parser.Options{}).ParseExpr(value); err != nil {
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

func formatPromQLExpr(v interface{}) string {
	if enablePromQLExprFormat {
		value, _ := parser.NewParser(parser.Options{}).ParseExpr(v.(string))
		// remove spaces causing decoding issues with multiline yaml marshal/unmarshall
		return strings.TrimLeft(parser.Prettify(value), " ")
	}
	return v.(string)
}

// Converts a string of the form "HH:MM" into the number of minutes elapsed in the day.
func parseTime(in string) (mins int) {
	timestampComponents := strings.Split(in, ":")
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
