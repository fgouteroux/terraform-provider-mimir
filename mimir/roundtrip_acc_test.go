package mimir

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"gopkg.in/yaml.v3"
)

func TestCountRuleGroups(t *testing.T) {
	cases := []struct {
		body    string
		want    int
		wantErr bool
	}{
		{"", 0, false},
		{"null\n", 0, false},
		{"[]\n", 0, false},
		{"{}\n", 0, false},
		{"namespace_sc:\n- name: g1\n", 1, false},
		{"namespace_sc:\n- name: g1\n- name: g2\n", 2, false},
		{"- name: g1\n", 1, false}, // flat list
		{"not-a-collection", 0, true},
	}
	for _, c := range cases {
		got, err := countRuleGroups(c.body)
		if (err != nil) != c.wantErr {
			t.Errorf("countRuleGroups(%q) err=%v, wantErr=%v", c.body, err, c.wantErr)
		}
		if err == nil && got != c.want {
			t.Errorf("countRuleGroups(%q) = %d, want %d", c.body, got, c.want)
		}
	}
}

// countRuleGroups decodes a ruler namespace-listing response and returns the group count. It
// accepts both observed shapes — a namespace-keyed map ({<ns>: [{name: ...}]}) and a flat list
// ([{name: ...}]) — and returns 0 for the legitimately-empty forms ("", "null", "[]", "{}").
// It errors ONLY when the body decodes as neither (a genuine shape/parse mismatch), so a real
// empty namespace never false-fails. The exact live shape is confirmed in CI (2.17.10 + 3.0.6);
// TestCountRuleGroups pins this decode logic without a network.
func countRuleGroups(body string) (int, error) {
	type group struct {
		Name string `yaml:"name"`
	}
	var byNamespace map[string][]group
	if err := yaml.Unmarshal([]byte(body), &byNamespace); err == nil {
		n := 0
		for _, groups := range byNamespace {
			n += len(groups)
		}
		return n, nil
	}
	var flat []group
	if err := yaml.Unmarshal([]byte(body), &flat); err == nil {
		return len(flat), nil
	}
	return 0, fmt.Errorf("could not decode ruler namespace listing as map or list: %q", body)
}

// testAccCheckMimirNamespaceGroupCount lists a namespace via the ruler API and asserts the
// number of groups. Terraform state/ID stability alone cannot prove the backend has no orphan
// at a mis-escaped path, so this out-of-band count guards the escaping/update paths.
func testAccCheckMimirNamespaceGroupCount(namespace string, want int, client *apiClient) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		body, err := client.sendRequest("ruler", "GET", rulesNamespacePath(namespace), "", map[string]string{})
		if err != nil {
			return fmt.Errorf("listing namespace %q: %w", namespace, err)
		}
		got, err := countRuleGroups(body)
		if err != nil {
			return fmt.Errorf("namespace %q: %w", namespace, err)
		}
		if got != want {
			return fmt.Errorf("namespace %q group count = %d, want %d", namespace, got, want)
		}
		return nil
	}
}

// AC4 regression guard: a space-containing name is the one class that round-trips today by
// luck; prove the escaping change did not break it, through apply -> read -> destroy.
func TestAccResourceRuleGroupAlerting_SpecialNameRoundTrip(t *testing.T) {
	client, err := NewAPIClient(setupClient())
	if err != nil {
		t.Fatal(err)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMimirRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoundTrip_specialName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirRuleGroupExists("mimir_rule_group_alerting.special", "Harvest Rules", client),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.special", "name", "Harvest Rules"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.special", "rule.0.alert", "LUN Destroyed"),
				),
			},
			{
				Config:   testAccRoundTrip_specialName,
				PlanOnly: true, // re-apply is a clean no-op
			},
		},
	})
}

// AC4 new class: %/# names do not round-trip today. Full lifecycle incl an in-place update,
// with an out-of-band count proving the update targets the existing group (no orphan).
func TestAccResourceRuleGroupAlerting_SpecialCharRoundTrip(t *testing.T) {
	client, err := NewAPIClient(setupClient())
	if err != nil {
		t.Fatal(err)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMimirRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoundTrip_specialChar,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirRuleGroupExists("mimir_rule_group_alerting.sc", "a%b#c", client),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.sc", "name", "a%b#c"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.sc", "rule.0.alert", "alert %x"),
					testAccCheckMimirNamespaceGroupCount("namespace_sc", 1, client),
				),
			},
			{
				Config: testAccRoundTrip_specialCharUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirRuleGroupExists("mimir_rule_group_alerting.sc", "a%b#c", client),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.sc", "rule.0.expr", "up > 1"),
					// update must target the existing group, not orphan/duplicate it
					testAccCheckMimirNamespaceGroupCount("namespace_sc", 1, client),
				),
			},
		},
	})
}

// AC5: a namespace with a space round-trips (accepted, escaped); a namespace with '/' is
// rejected at plan time (so the cross-tenant ID-misparse cannot occur).
func TestAccResourceRuleGroupAlerting_NamespaceRoundTrip(t *testing.T) {
	client, err := NewAPIClient(setupClient())
	if err != nil {
		t.Fatal(err)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMimirRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccRoundTrip_nsSlash,
				ExpectError: regexp.MustCompile("Invalid Namespace"),
			},
			{
				Config: testAccRoundTrip_nsSpace,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirRuleGroupExists("mimir_rule_group_alerting.nss", "grp", client),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.nss", "namespace", "team space"),
				),
			},
		},
	})
}

// AC5 / ID-hardening: the escaped-ID scheme round-trips through import for ordinary names.
func TestAccResourceRuleGroupAlerting_ImportRoundTrip(t *testing.T) {
	client, err := NewAPIClient(setupClient())
	if err != nil {
		t.Fatal(err)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMimirRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoundTrip_specialName,
				Check:  testAccCheckMimirRuleGroupExists("mimir_rule_group_alerting.special", "Harvest Rules", client),
			},
			{
				ResourceName:      "mimir_rule_group_alerting.special",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// AC5 / ID-hardening at the import wiring: a crafted import ID with a '/'-in-namespace (as
// %2F) is rejected through the real Importer/Read path, not only at the parseRuleGroupID
// unit level — so a regression in the Importer wiring would be caught.
func TestAccResourceRuleGroupAlerting_ImportRejectRoundTrip(t *testing.T) {
	client, err := NewAPIClient(setupClient())
	if err != nil {
		t.Fatal(err)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMimirRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoundTrip_specialName,
				Check:  testAccCheckMimirRuleGroupExists("mimir_rule_group_alerting.special", "Harvest Rules", client),
			},
			{
				ResourceName:  "mimir_rule_group_alerting.special",
				ImportState:   true,
				ImportStateId: "team%2Fevil/grp", // '/'-in-namespace (as %2F) must be rejected
				ExpectError:   regexp.MustCompile("namespace .* must not"),
			},
		},
	})
}

// AC1/AC4 for the typed recording resource: a special group name with an ordinary (strict)
// record name.
func TestAccResourceRuleGroupRecording_SpecialNameRoundTrip(t *testing.T) {
	client, err := NewAPIClient(setupClient())
	if err != nil {
		t.Fatal(err)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMimirRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoundTrip_recordingSpecial,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirRuleGroupExists("mimir_rule_group_recording.rt", "Rec Group %x", client),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.rt", "name", "Rec Group %x"),
					resource.TestCheckResourceAttr("mimir_rule_group_recording.rt", "rule.0.record", "valid_record"),
				),
			},
		},
	})
}

const testAccRoundTrip_specialName = `
resource "mimir_rule_group_alerting" "special" {
  name      = "Harvest Rules"
  namespace = "namespace_special"
  rule {
    alert = "LUN Destroyed"
    expr  = "up"
  }
}
`

const testAccRoundTrip_specialChar = `
resource "mimir_rule_group_alerting" "sc" {
  name      = "a%b#c"
  namespace = "namespace_sc"
  rule {
    alert = "alert %x"
    expr  = "up"
  }
}
`

const testAccRoundTrip_specialCharUpdate = `
resource "mimir_rule_group_alerting" "sc" {
  name      = "a%b#c"
  namespace = "namespace_sc"
  rule {
    alert = "alert %x"
    expr  = "up > 1"
  }
}
`

const testAccRoundTrip_nsSpace = `
resource "mimir_rule_group_alerting" "nss" {
  name      = "grp"
  namespace = "team space"
  rule {
    alert = "a"
    expr  = "up"
  }
}
`

const testAccRoundTrip_nsSlash = `
resource "mimir_rule_group_alerting" "nsslash" {
  name      = "grp"
  namespace = "team/space"
  rule {
    alert = "a"
    expr  = "up"
  }
}
`

const testAccRoundTrip_recordingSpecial = `
resource "mimir_rule_group_recording" "rt" {
  name      = "Rec Group %x"
  namespace = "namespace_rec"
  rule {
    record = "valid_record"
    expr   = "up"
  }
}
`

// Pins the upstream rulefmt scope of the braces "common-mistake check" (prometheus#15851):
// it applies to RECORD names only, so alert and group names containing braces are valid and
// must round-trip. (Record names with braces are rejected by this provider's own strict
// metric-name validation, matching the server.)
func TestAccResourceRuleGroupAlerting_BracesNameRoundTrip(t *testing.T) {
	client, err := NewAPIClient(setupClient())
	if err != nil {
		t.Fatal(err)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMimirRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoundTrip_bracesName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirRuleGroupExists("mimir_rule_group_alerting.braces", "Braces {Group}", client),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.braces", "name", "Braces {Group}"),
					resource.TestCheckResourceAttr("mimir_rule_group_alerting.braces", "rule.0.alert", `alert{env="prod"}`),
				),
			},
			{
				Config:   testAccRoundTrip_bracesName,
				PlanOnly: true, // re-apply is a clean no-op
			},
		},
	})
}

const testAccRoundTrip_bracesName = `
resource "mimir_rule_group_alerting" "braces" {
  name      = "Braces {Group}"
  namespace = "namespace_braces"
  rule {
    alert = "alert{env=\"prod\"}"
    expr  = "up"
  }
}
`
