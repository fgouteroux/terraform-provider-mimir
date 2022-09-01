package mimir

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
)

var (
	groupRuleNameRegexp     = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-_.]*$`)
	alertingRuleNameRegexp  = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	recordingRuleNameRegexp = regexp.MustCompile(`^[a-zA-Z_:][a-zA-Z0-9_:]*$`)
)

func jsonPrettyPrint(input []byte) string {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(input), "", "  ")
	if err != nil {
		return string(input)
	}
	return out.String()
}

func handleHTTPError(err error, body string, url, baseMsg string) error {
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

	if !groupRuleNameRegexp.MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"\"%s\": Invalid Group Rule Name %q. Must match the regex %s", k, value, groupRuleNameRegexp))
	}

	return
}

// SliceFind takes a slice and looks for an element in it. If found it will
// return true otherwise false.
func SliceFind(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
