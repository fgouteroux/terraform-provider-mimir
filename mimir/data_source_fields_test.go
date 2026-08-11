package mimir

import "testing"

// TestRuleGroupDataSourcesExposeGroupFields pins the group-level attributes the
// data sources must surface. The alerting one exposed neither interval nor the
// group labels, so a group using them could not be read back in full.
func TestRuleGroupDataSourcesExposeGroupFields(t *testing.T) {
	alerting := dataSourcemimirRuleGroupAlerting().Schema
	for _, attr := range []string{intervalKey, labelsKey, sourceTenantsKey} {
		if _, declared := alerting[attr]; !declared {
			t.Errorf("alerting data source does not expose %q", attr)
		}
	}

	recording := dataSourcemimirRuleGroupRecording().Schema
	for _, attr := range []string{intervalKey, labelsKey, queryOffsetKey, sourceTenantsKey} {
		if _, declared := recording[attr]; !declared {
			t.Errorf("recording data source does not expose %q", attr)
		}
	}
}
