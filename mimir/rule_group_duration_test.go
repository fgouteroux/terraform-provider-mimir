package mimir

import "testing"

// TestRuleGroupDurationsNormalizeState pins StateFunc: formatDuration on every
// rule-group duration attribute. Without it a non-canonical value such as "60s"
// stays verbatim in state while the normalized form differs, producing a
// permanent diff — the defect that broke CI for the pushover ttl attribute.
func TestRuleGroupDurationsNormalizeState(t *testing.T) {
	alerting := resourcemimirRuleGroupAlerting().Schema
	if alerting[intervalKey].StateFunc == nil {
		t.Error("alerting rule group interval has no StateFunc: formatDuration")
	}

	recording := resourcemimirRuleGroupRecording().Schema
	for _, attr := range []string{intervalKey, queryOffsetKey, "evaluation_delay"} {
		entry, declared := recording[attr]
		if !declared {
			t.Errorf("recording rule group has no %q attribute", attr)
			continue
		}
		if entry.StateFunc == nil {
			t.Errorf("recording rule group %q has no StateFunc: formatDuration", attr)
		}
	}
}
