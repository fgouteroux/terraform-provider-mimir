package mimir

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestValidateHTTPHeaderName(t *testing.T) {
	for name, wantErr := range map[string]bool{
		"X-Custom":      false,
		"x-team":        false,
		"Authorization": true,
		"content-type":  true,
		"User-Agent":    true,
		"":              true,
	} {
		if _, errs := validateHTTPHeaderName(name, "name"); (len(errs) > 0) != wantErr {
			t.Errorf("validateHTTPHeaderName(%q) errors = %v, wantErr %v", name, errs, wantErr)
		}
	}
}

// The new http_config and oauth2 fields must survive expand then flatten, and
// stay out of the configuration when unset.
func TestHTTPConfigFieldsRoundTrip(t *testing.T) {
	httpSchema := httpConfigFields()
	oauth2Schema := httpSchema["oauth2"].Elem.(*schema.Resource).Schema

	oauth2Cfg := zeroValues(oauth2Schema)
	oauth2Cfg["client_id"] = "id"
	oauth2Cfg["token_url"] = "https://idp.example.com/token"
	oauth2Cfg["grant_type"] = oauth2GrantTypeJWTBearer
	oauth2Cfg["client_certificate_key_id"] = "kid"
	oauth2Cfg["client_certificate_key"] = "key"
	oauth2Cfg["signature_algorithm"] = "RS384"
	oauth2Cfg["iss"] = "issuer"
	oauth2Cfg["audience"] = "aud"
	oauth2Cfg["claims"] = map[string]interface{}{"tenant": "a"}

	cfg := zeroValues(httpSchema)
	cfg["proxy_url"] = "http://proxy:3128"
	cfg["no_proxy"] = "localhost"
	cfg["proxy_connect_header"] = schema.NewSet(schema.HashResource(httpSchema["proxy_connect_header"].Elem.(*schema.Resource)), []interface{}{
		map[string]interface{}{"name": "X-Proxy", "values": []interface{}{"p"}},
	})
	cfg["http_headers"] = schema.NewSet(schema.HashResource(httpSchema["http_headers"].Elem.(*schema.Resource)), []interface{}{
		map[string]interface{}{"name": "X-Team", "values": []interface{}{"infra"}, "secrets": []interface{}{"s3cr3t"}},
	})
	cfg["oauth2"] = []interface{}{oauth2Cfg}

	conf := expandHTTPConfig([]interface{}{cfg})
	yamlOut := marshalYAML(t, conf)
	for _, want := range []string{
		"no_proxy: localhost",
		"X-Proxy:",
		"http_headers:",
		"X-Team:",
		"secrets:",
		"grant_type: " + oauth2GrantTypeJWTBearer,
		"client_certificate_key_id: kid",
		"signature_algorithm: RS384",
		"tenant: a",
	} {
		if !regexp.MustCompile(regexp.QuoteMeta(want)).MatchString(yamlOut) {
			t.Errorf("expanded config misses %q:\n%s", want, yamlOut)
		}
	}

	out := flattenHTTPConfig(conf)[0].(map[string]interface{})
	if out["no_proxy"] != "localhost" {
		t.Errorf("no_proxy = %v", out["no_proxy"])
	}
	headers := out["http_headers"].([]interface{})
	if len(headers) != 1 || headers[0].(map[string]interface{})["name"] != "X-Team" {
		t.Errorf("http_headers = %v", headers)
	}
	o := out["oauth2"].([]interface{})[0].(map[string]interface{})
	if o["grant_type"] != oauth2GrantTypeJWTBearer || o["claims"].(map[string]interface{})["tenant"] != "a" {
		t.Errorf("oauth2 = %v", o)
	}

	plain := expandHTTPConfig([]interface{}{zeroValues(httpSchema)})
	for _, unwanted := range []string{"no_proxy", "proxy_from_environment", "proxy_connect_header", "http_headers"} {
		if regexp.MustCompile(unwanted).MatchString(marshalYAML(t, plain)) {
			t.Errorf("unset %s must stay out of the configuration", unwanted)
		}
	}
}

func TestAccResourceAlertmanagerConfig_HTTPConfigFields(t *testing.T) {
	client, err := NewAPIClient(setupClient())
	if err != nil {
		t.Fatal(err)
	}

	const res = "mimir_alertmanager_config.mytenant"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMimirAlertmanagerConfigDestroy,
		Steps: []resource.TestStep{
			{
				// Rejected at plan time: runs first so the last step leaves a
				// valid configuration for the post-test destroy.
				Config: `
    resource "mimir_alertmanager_config" "mytenant" {
      route {
        group_wait      = "30s"
        group_interval  = "5m"
        repeat_interval = "1h"
        receiver        = "webhook"
      }
      receiver {
        name = "webhook"
        webhook_configs {
          url = "http://example.com/hook"
          http_config {
            http_headers {
              name   = "Authorization"
              values = ["x"]
            }
          }
        }
      }
    }`,
				ExpectError: regexp.MustCompile(`setting header "Authorization" is not allowed`),
			},
			{
				Config: testAccResourceAlertmanagerConfigHTTPConfigFields,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMimirAlertmanagerConfigExists(res, client),
					resource.TestCheckResourceAttr(res, "receiver.0.webhook_configs.0.http_config.0.no_proxy", "localhost"),
					resource.TestCheckResourceAttr(res, "receiver.0.webhook_configs.0.http_config.0.http_headers.#", "2"),
					resource.TestCheckResourceAttr(res, "receiver.0.webhook_configs.0.http_config.0.proxy_connect_header.#", "1"),
					resource.TestCheckResourceAttr(res, "receiver.1.webhook_configs.0.http_config.0.oauth2.0.grant_type", oauth2GrantTypeJWTBearer),
					resource.TestCheckResourceAttr(res, "receiver.1.webhook_configs.0.http_config.0.oauth2.0.claims.tenant", "mytenant"),
				),
			},
			{
				// Headers come back from Mimir as a map: the set must not diff.
				Config:   testAccResourceAlertmanagerConfigHTTPConfigFields,
				PlanOnly: true,
			},
		},
	})
}

const testAccResourceAlertmanagerConfigHTTPConfigFields = `
    resource "mimir_alertmanager_config" "mytenant" {
      route {
        group_by        = ["..."]
        group_wait      = "30s"
        group_interval  = "5m"
        repeat_interval = "1h"
        receiver        = "webhook"
      }
      receiver {
        name = "webhook"
        webhook_configs {
          url = "http://example.com/hook"
          http_config {
            proxy_url = "http://proxy.example.com:3128"
            no_proxy  = "localhost"
            proxy_connect_header {
              name   = "X-Proxy-Token"
              values = ["proxy"]
            }
            http_headers {
              name   = "X-Team"
              values = ["infra"]
            }
            http_headers {
              name    = "X-Api-Key"
              secrets = ["s3cr3t"]
            }
          }
        }
      }
      receiver {
        name = "oauth"
        webhook_configs {
          url = "http://example.com/oauth"
          http_config {
            oauth2 {
              client_id           = "client"
              token_url           = "https://idp.example.com/token"
              grant_type          = "urn:ietf:params:oauth:grant-type:jwt-bearer"
              signature_algorithm = "RS256"
              audience            = "https://idp.example.com"
              claims = {
                tenant = "mytenant"
              }
            }
          }
        }
      }
    }
`
