package mimir

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

// TestExpandWeChatConfigKeepsAPIURL pins the schema key the expand function
// reads. It used to read "api_url_url", which no schema declares, so a
// configured api_url never reached Mimir and every plan re-added it.
func TestExpandWeChatConfigKeepsAPIURL(t *testing.T) {
	data := zeroValues(weChatConfigFields())
	data[apiURLKey] = "https://qyapi.weixin.qq.com/cgi-bin/"

	got := expandWeChatConfig([]interface{}{data})
	if len(got) != 1 || got[0].APIURL != "https://qyapi.weixin.qq.com/cgi-bin/" {
		t.Fatalf("api_url was dropped by expandWeChatConfig: %+v", got[0])
	}
}

// TestExpandOAuth2KeepsTLSConfig covers the nested oauth2 tls_config block,
// which was declared in the schema but never read by expand nor written by
// flatten, so it was silently dropped and re-planned forever.
func TestExpandOAuth2KeepsTLSConfig(t *testing.T) {
	oauth := zeroValues(httpConfigFields()["oauth2"].Elem.(*schema.Resource).Schema)
	oauth["client_id"] = "id"
	oauth["token_url"] = "https://idp.example.com/token"
	oauth["tls_config"] = []interface{}{map[string]interface{}{
		"server_name": "idp.example.com", "insecure_skip_verify": false,
		"min_version": "", "max_version": "",
	}}

	expanded := expandHTTPConfigOAuth2([]interface{}{oauth})
	if expanded == nil || expanded.TLSConfig == nil {
		t.Fatal("oauth2 tls_config was dropped by expandHTTPConfigOAuth2")
	}
	if expanded.TLSConfig.ServerName != "idp.example.com" {
		t.Errorf("server_name = %q, want idp.example.com", expanded.TLSConfig.ServerName)
	}

	back := flattenHTTPConfigOAuth2(expanded)
	if _, ok := back[0].(map[string]interface{})["tls_config"]; !ok {
		t.Error("oauth2 tls_config was dropped by flattenHTTPConfigOAuth2")
	}
}
