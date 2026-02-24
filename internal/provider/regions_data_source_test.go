package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRegionsDataSource(t *testing.T) {
	if os.Getenv("TERASWITCH_API_KEY") == "" {
		t.Skip("Skipping, api key not provided")
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegionsDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.teraswitch_regions.test", "regions.#"),
				),
			},
		},
	})
}

func TestAccRegionsDataSourceWithFilter(t *testing.T) {
	if os.Getenv("TERASWITCH_API_KEY") == "" {
		t.Skip("Skipping, api key not provided")
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegionsDataSourceConfigWithFilter(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.teraswitch_regions.metal", "regions.#"),
				),
			},
		},
	})
}

func testAccRegionsDataSourceConfig() string {
	return `
provider "teraswitch" {}

data "teraswitch_regions" "test" {}
`
}

func testAccRegionsDataSourceConfigWithFilter() string {
	return `
provider "teraswitch" {}

data "teraswitch_regions" "metal" {
  service_type = "Metal"
}
`
}
