package mimir

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/promql/parser"
	"regexp"
)

var (
	groupRuleNameRegexp = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-_.]*$`)
	labelNameRegexp     = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	metricNameRegexp    = regexp.MustCompile(`^[a-zA-Z_:][a-zA-Z0-9_:]*$`)
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

func validatePromQLExpr(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if _, err := parser.ParseExpr(value); err != nil {
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

		if !metricNameRegexp.MatchString(lvalue.(string)) {
			errors = append(errors, fmt.Errorf(
				"\"%s\": Invalid Label Value %q. Must match the regex %s", k, lvalue, metricNameRegexp))
		}
	}
	return
}

func validateAnnotations(v interface{}, k string) (ws []string, errors []error) {
	m := v.(map[string]interface{})
	for aname, _ := range m {

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
