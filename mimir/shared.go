package mimir

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/promql/parser"
)

var (
	groupRuleNameRegexp = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-_.]*$`)
	labelNameRegexp     = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	metricNameRegexp    = regexp.MustCompile(`^[a-zA-Z_:][a-zA-Z0-9_:]*$`)
)

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
	return value.String()
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

func formatPromQLExpr(v interface{}) string {
	value, _ := parser.ParseExpr(v.(string))
	// remove spaces causing decoding issues with multiline yaml marshal/unmarshall
	return strings.TrimLeft(parser.Prettify(value), " ")
}
