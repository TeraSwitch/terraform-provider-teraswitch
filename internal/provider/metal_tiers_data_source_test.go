package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMetalTiersDataSource(t *testing.T) {
	if os.Getenv("TERASWITCH_API_KEY") == "" {
		t.Skip("Skipping, api key not provided")
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMetalTiersDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.teraswitch_metal_tiers.test", "tiers.#"),
				),
			},
		},
	})
}

func testAccMetalTiersDataSourceConfig() string {
	return `
provider "teraswitch" {}

data "teraswitch_metal_tiers" "test" {}
`
}
