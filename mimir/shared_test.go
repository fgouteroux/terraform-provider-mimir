package mimir

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"gopkg.in/yaml.v3"
)

// mimirVersion returns the Mimir version under test, skipping when MIMIR_VERSION
// is unset or unparseable. The older inline pattern
// (version.NewVersion(os.Getenv(...)) followed by .LessThan) nil-panics on an
// empty MIMIR_VERSION, which takes the whole package down under a plain
// `go test`, and its bare `return` made a skipped gate report as PASS.
func mimirVersion(t *testing.T, requirement string) *version.Version {
	t.Helper()
	v := os.Getenv("MIMIR_VERSION")
	if v == "" {
		t.Skipf("MIMIR_VERSION not set; skipping test requiring Mimir %s", requirement)
	}
	cur, err := version.NewVersion(v)
	if err != nil {
		t.Skipf("MIMIR_VERSION %q unparseable; skipping test requiring Mimir %s", v, requirement)
	}
	return cur
}

// skipBelowMimirVersion skips a test unless the Mimir under test is at least min.
func skipBelowMimirVersion(t *testing.T, min string) {
	t.Helper()
	cur := mimirVersion(t, ">= "+min)
	minV, _ := version.NewVersion(min)
	if cur.LessThan(minV) {
		t.Skipf("Mimir %s < %s; skipping test", cur, minV)
	}
}

// skipUnlessMimirVersion skips a test unless the Mimir under test is exactly
// want, for behaviour that only ever existed in a single release.
func skipUnlessMimirVersion(t *testing.T, want string) {
	t.Helper()
	cur := mimirVersion(t, "== "+want)
	wantV, _ := version.NewVersion(want)
	if !cur.Equal(wantV) {
		t.Skipf("Mimir %s is not %s; skipping test", cur, wantV)
	}
}

// testAccCheckMimirRuleGroupHasLabel GETs a rule group from the ruler API and
// asserts it carries a group-level label key=val. Used for mimir_rules, whose
// group labels live inside the YAML content and are not Terraform state
// attributes (so resource.TestCheckResourceAttr cannot see them).
func testAccCheckMimirRuleGroupHasLabel(n, group, key, val string, client *apiClient) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("mimir object not found in terraform state: %s", n)
		}
		orgID := rs.Primary.Attributes["org_id"]
		namespace := rs.Primary.Attributes["namespace"]
		headers := make(map[string]string)
		if orgID != "" {
			headers["X-Scope-OrgID"] = orgID
		}
		path := rulesGroupPath(namespace, group)
		body, err := client.sendRequest("ruler", "GET", path, "", headers)
		if err != nil {
			return err
		}
		var grp struct {
			Labels map[string]string `yaml:"labels"`
		}
		if err := yaml.Unmarshal([]byte(body), &grp); err != nil {
			return fmt.Errorf("parsing rule group %q YAML: %w", group, err)
		}
		if got := grp.Labels[key]; got != val {
			return fmt.Errorf("rule group %q label %q = %q, want %q (body: %s)", group, key, got, val, body)
		}
		return nil
	}
}

func getSetEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
		os.Setenv(key, fallback)
	}
	return value
}

func testAccCheckMimirRuleGroupExists(n string, name string, client *apiClient) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			keys := make([]string, 0, len(s.RootModule().Resources))
			for k := range s.RootModule().Resources {
				keys = append(keys, k)
			}
			return fmt.Errorf("mimir object not found in terraform state: %s. Found: %s", n, strings.Join(keys, ", "))
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("mimir object name %s not set in terraform", name)
		}

		orgID := rs.Primary.Attributes["org_id"]
		name := rs.Primary.Attributes["name"]
		namespace := rs.Primary.Attributes["namespace"]

		/* Make a throw-away API object to read from the API */
		headers := make(map[string]string)
		if orgID != "" {
			headers["X-Scope-OrgID"] = orgID
		}
		path := rulesGroupPath(namespace, name)
		_, err := client.sendRequest("ruler", "GET", path, "", headers)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckMimirNamespaceExists(n string, name string, client *apiClient) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			keys := make([]string, 0, len(s.RootModule().Resources))
			for k := range s.RootModule().Resources {
				keys = append(keys, k)
			}
			return fmt.Errorf("mimir object not found in terraform state: %s. Found: %s", n, strings.Join(keys, ", "))
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("mimir object name %s not set in terraform", name)
		}

		orgID := rs.Primary.Attributes["org_id"]
		namespace := rs.Primary.Attributes["namespace"]

		/* Make a throw-away API object to read from the API */
		headers := make(map[string]string)
		if orgID != "" {
			headers["X-Scope-OrgID"] = orgID
		}
		path := rulesNamespacePath(namespace)
		_, err := client.sendRequest("ruler", "GET", path, "", headers)
		if err != nil {
			return err
		}

		return nil
	}
}

// testAccCheckMimirRuleGroupGone verifies that a specific rule group within a
// namespace has actually been removed from Mimir (returns 404). Use this in
// Update tests that remove a group from the YAML to guard against the
// regression where Update silently keeps orphaned groups alive in Mimir while
// dropping them from Terraform state.
func testAccCheckMimirRuleGroupGone(n string, groupName string, client *apiClient) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("mimir object not found in terraform state: %s", n)
		}

		orgID := rs.Primary.Attributes["org_id"]
		namespace := rs.Primary.Attributes["namespace"]

		headers := make(map[string]string)
		if orgID != "" {
			headers["X-Scope-OrgID"] = orgID
		}
		path := rulesGroupPath(namespace, groupName)
		_, err := client.sendRequest("ruler", "GET", path, "", headers)
		if err == nil {
			return fmt.Errorf("rule group %q in namespace %q still exists in Mimir; expected 404", groupName, namespace)
		}
		if !strings.Contains(err.Error(), "response code '404'") {
			return fmt.Errorf("unexpected error checking rule group %q is gone: %w", groupName, err)
		}
		return nil
	}
}

func testAccCheckMimirRuleGroupDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*apiClient)

	// loop through the resources in state, verifying each widget
	// is destroyed
	for _, rs := range s.RootModule().Resources {
		if !strings.HasPrefix(rs.Type, "mimir_rule_group") {
			continue
		}

		orgID := rs.Primary.Attributes["org_id"]
		name := rs.Primary.Attributes["name"]
		namespace := rs.Primary.Attributes["namespace"]

		headers := make(map[string]string)
		if orgID != "" {
			headers["X-Scope-OrgID"] = orgID
		}
		path := rulesGroupPath(namespace, name)
		_, err := client.sendRequest("ruler", "GET", path, "", headers)

		// If the error is equivalent to 404 not found, the widget is destroyed.
		// Otherwise return the error
		if !strings.Contains(err.Error(), "response code '404'") {
			return err
		}
	}

	return nil
}

func testAccCheckMimirRuleDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	client := testAccProvider.Meta().(*apiClient)

	// loop through the resources in state, verifying each is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "mimir_rules" {
			continue
		}

		orgID := rs.Primary.Attributes["org_id"]
		namespace := rs.Primary.Attributes["namespace"]

		headers := make(map[string]string)
		if orgID != "" {
			headers["X-Scope-OrgID"] = orgID
		}

		// Parse managed_groups from state attributes
		// Terraform stores list items as: managed_groups.0, managed_groups.1, etc.
		managedGroupsCount, _ := strconv.Atoi(rs.Primary.Attributes["managed_groups.#"])

		for i := 0; i < managedGroupsCount; i++ {
			groupName := rs.Primary.Attributes[fmt.Sprintf("managed_groups.%d", i)]

			path := rulesGroupPath(namespace, groupName)
			_, err := client.sendRequest("ruler", "GET", path, "", headers)

			// If the error is equivalent to 404 not found, the group is destroyed.
			// Otherwise return the error
			if err != nil && !strings.Contains(err.Error(), "response code '404'") {
				return err
			}
		}
	}

	return nil
}
func setupClient() *apiClientOpt {
	headers := make(map[string]string)
	headers["X-Scope-OrgID"] = mimirOrgID

	opt := &apiClientOpt{
		uri:             mimirURI,
		rulerURI:        mimirRulerURI,
		alertmanagerURI: mimirAlertmanagerURI,
		insecure:        false,
		username:        "",
		password:        "",
		proxyURL:        "",
		token:           "",
		cert:            "",
		key:             "",
		ca:              "",
		headers:         headers,
		timeout:         2,
		debug:           true,
	}
	return opt
}
