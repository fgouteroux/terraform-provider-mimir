package mimir

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/promql/parser"
)

var (
	groupRuleNameRegexp = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-_.]*$`)
	labelNameRegexp     = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	metricNameRegexp    = regexp.MustCompile(`^[a-zA-Z_:][a-zA-Z0-9_:]*$`)
	validTime           = "^((([01][0-9])|(2[0-3])):[0-5][0-9])$|(^24:00$)"
	validTimeRE         = regexp.MustCompile(validTime)
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

func formatPromQLExpr(v interface{}) string {
	if enablePromQLExprFormat {
		value, _ := parser.ParseExpr(v.(string))
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
