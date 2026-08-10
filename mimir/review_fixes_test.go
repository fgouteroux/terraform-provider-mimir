package mimir

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TestOnlyGroupsRejectsUnknownGroup pins the guard on the destructive path:
// only_groups naming a group absent from the configuration used to shrink the
// managed set silently, which made Update compute every managed group as
// "to delete", wipe them from Mimir and report success.
func TestOnlyGroupsRejectsUnknownGroup(t *testing.T) {
	res := resourceMimirRules()
	content := "groups:\n  - name: real_group\n    rules:\n      - record: up:sum\n        expr: sum(up)\n"

	cfg := terraform.NewResourceConfigRaw(map[string]interface{}{
		"namespace":   "ns",
		contentKey:    content,
		"only_groups": []interface{}{"typo_group"},
	})

	_, err := res.Diff(context.Background(), &terraform.InstanceState{}, cfg, nil)
	if err == nil {
		t.Fatal("expected an error for an only_groups entry that matches no rule group")
	}
	if !strings.Contains(err.Error(), "typo_group") {
		t.Errorf("error should name the offending group, got: %v", err)
	}
}

// TestOnlyGroupsAcceptsKnownGroup guards against the fix over-rejecting.
func TestOnlyGroupsAcceptsKnownGroup(t *testing.T) {
	res := resourceMimirRules()
	content := "groups:\n  - name: real_group\n    rules:\n      - record: up:sum\n        expr: sum(up)\n"

	cfg := terraform.NewResourceConfigRaw(map[string]interface{}{
		"namespace":   "ns",
		contentKey:    content,
		"only_groups": []interface{}{"real_group"},
	})

	if _, err := res.Diff(context.Background(), &terraform.InstanceState{}, cfg, nil); err != nil {
		t.Fatalf("unexpected error for a valid only_groups entry: %v", err)
	}
}
