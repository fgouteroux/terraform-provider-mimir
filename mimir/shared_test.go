package mimir

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

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

		/* Make a throw-away API object to read from the API */
		var headers map[string]string
		path := fmt.Sprintf("/config/v1/rules/%s", rs.Primary.ID)
		_, err := client.sendRequest("ruler", "GET", path, "", headers)
		if err != nil {
			return err
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
		if rs.Type != "mimir_rule_group_recording" {
			continue
		}

		var headers map[string]string
		path := fmt.Sprintf("/config/v1/rules/%s", rs.Primary.ID)
		_, err := client.sendRequest("ruler", "GET", path, "", headers)

		// If the error is equivalent to 404 not found, the widget is destroyed.
		// Otherwise return the error
		if !strings.Contains(err.Error(), "group does not exist") {
			return err
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
