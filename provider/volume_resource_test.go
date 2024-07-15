package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVolumeResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccVolumeResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teraswitch_volume.test", "region_id", "PIT1"),
					resource.TestCheckResourceAttr("teraswitch_volume.test", "display_name", "yeehaw"),
					resource.TestCheckResourceAttr("teraswitch_volume.test", "size", "20"),
					resource.TestCheckResourceAttr("teraswitch_volume.test", "volume_type", "nvme"),
					resource.TestCheckResourceAttr("teraswitch_volume.test", "description", "test 111"),
					resource.TestCheckResourceAttrSet("teraswitch_volume.test", "status"),
					resource.TestCheckResourceAttrSet("teraswitch_volume.test", "updated_at"),
					resource.TestCheckResourceAttrSet("teraswitch_volume.test", "created_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:        "teraswitch_volume.test",
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: "45/",
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
				ImportStateVerifyIgnore: []string{"configurable_attribute", "defaulted"},
			},
			// Update and Read testing
			// {
			// 	Config: testAccVolumeResourceConfig(),
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		resource.TestCheckResourceAttr("scaffolding_example.test", "configurable_attribute", "two"),
			// 	),
			// },
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccVolumeResourceConfig() string {
	return fmt.Sprintf(`
provider "teraswitch" {}

resource "teraswitch_volume" "test" {
	region_id      = "PIT1"
	display_name   = "yeehaw"
	size           = 20
	volume_type    = "nvme"
	description    = "test 111"
}
`)
}