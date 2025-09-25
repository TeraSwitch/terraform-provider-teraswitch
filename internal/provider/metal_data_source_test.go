package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMetalDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccMetalDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.teraswitch_metal.test", "id"),
					resource.TestCheckResourceAttrSet("data.teraswitch_metal.test", "project_id"),
					resource.TestCheckResourceAttrSet("data.teraswitch_metal.test", "region_id"),
					resource.TestCheckResourceAttrSet("data.teraswitch_metal.test", "tier_id"),
				),
			},
		},
	})
}

const testAccMetalDataSourceConfig = `
data "teraswitch_metal" "test" {
  id = 1234  # Replace with a valid metal service ID for testing
}
`
