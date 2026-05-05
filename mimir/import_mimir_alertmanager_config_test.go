package mimir

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccImportAlertmanagerConfig_Basic(t *testing.T) {
	resourceName := "mimir_alertmanager_config.mytenant"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAlertmanagerConfig_basic,
			},

			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"org_id"},
			},
		},
	})
}

func TestAccImportAlertmanagerConfig_WithOrgID(t *testing.T) {
	resourceName := "mimir_alertmanager_config.another_tenant"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAlertmanagerConfig_WithOrgID,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "another_tenant",
			},
		},
	})
}
