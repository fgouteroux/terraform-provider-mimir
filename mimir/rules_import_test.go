package mimir

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TestRulesImportRebuildsContent covers an import that could never complete:
// the importer set only org_id and namespace, then Terraform called Read, which
// needs a rule configuration and failed with "no rule configuration provided".
func TestRulesImportRebuildsContent(t *testing.T) {
	const listing = `namespace1:
  - name: group_one
    interval: 1m
    rules:
      - alert: HighErrorRate
        expr: rate(errors[5m]) > 1
        keep_firing_for: 10m
`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, rulesNamespacePath("namespace1"); got != want {
			t.Errorf("importer requested %q, want %q", got, want)
		}
		w.Header().Set("Content-Type", "application/yaml")
		_, _ = w.Write([]byte(listing))
	}))
	defer srv.Close()

	client, err := NewAPIClient(&apiClientOpt{uri: srv.URL, headers: map[string]string{}, timeout: 5})
	if err != nil {
		t.Fatalf("NewAPIClient: %v", err)
	}

	res := resourceMimirRules()
	d := res.Data(&terraform.InstanceState{ID: "tenant1/namespace1"})
	d.SetId("tenant1/namespace1")

	imported, err := resourceMimirRulesImport(context.Background(), d, client)
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}
	if len(imported) != 1 {
		t.Fatalf("got %d resources, want 1", len(imported))
	}

	got := imported[0].Get(contentKey).(string)
	if got == "" {
		t.Fatal("importer left content empty; Read would fail with \"no rule configuration provided\"")
	}
	for _, want := range []string{"name: group_one", "alert: HighErrorRate", "keep_firing_for: 10m"} {
		if !strings.Contains(got, want) {
			t.Errorf("rebuilt content is missing %q:\n%s", want, got)
		}
	}
	if imported[0].Get(namespaceKey).(string) != "namespace1" {
		t.Errorf("namespace = %q, want namespace1", imported[0].Get(namespaceKey))
	}
}

// TestDecodeNamespaceRuleGroupsAcceptsBothShapes pins the two response shapes
// the ruler has been observed returning.
func TestDecodeNamespaceRuleGroupsAcceptsBothShapes(t *testing.T) {
	keyed := "ns:\n  - name: a\n    rules:\n      - record: r\n        expr: up\n"
	flat := "- name: a\n  rules:\n    - record: r\n      expr: up\n"

	for name, body := range map[string]string{"namespace-keyed": keyed, "flat list": flat} {
		groups, err := decodeNamespaceRuleGroups(body)
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		if len(groups.Groups) != 1 || groups.Groups[0].Name != "a" {
			t.Errorf("%s: got %+v, want one group named a", name, groups.Groups)
		}
	}
}

// TestOutOfBandDeletionIsPlanned covers drift repair for inline content: the
// hash comparison used to be skipped unless content_file was set, so a group
// deleted directly in Mimir produced an empty plan and was never recreated.
func TestOutOfBandDeletionIsPlanned(t *testing.T) {
	res := resourceMimirRules()
	content := "groups:\n  - name: g1\n    rules:\n      - record: r\n        expr: up\n"

	cfg := terraform.NewResourceConfigRaw(map[string]interface{}{
		namespaceKey: "ns",
		contentKey:   content,
	})

	// State as Read leaves it once g1 has been deleted out of band: the hash is
	// computed over the surviving (empty) set, so it diverges from the config.
	state := &terraform.InstanceState{
		ID: "ns",
		Attributes: map[string]string{
			"namespace":    "ns",
			contentKey:     content,
			"content_hash": "stale-hash-from-a-namespace-missing-g1",
		},
	}

	diff, err := res.Diff(context.Background(), state, cfg, nil)
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	if diff == nil || diff.Empty() {
		t.Fatal("a rule group deleted out of band produced an empty plan; the drift is never repaired")
	}
	if _, planned := diff.Attributes["content_hash"]; !planned {
		t.Errorf("expected content_hash in the plan, got %v", diff.Attributes)
	}
}
