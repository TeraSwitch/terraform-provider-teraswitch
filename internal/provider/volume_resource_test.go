package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVolumeResource(t *testing.T) {
	if os.Getenv("TERASWITCH_API_KEY") == "" {
		t.Skip("Skipping, api key not provided")
		return
	}

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
				),
			},
			// ImportState testing
			{
				ResourceName:      "teraswitch_volume.test",
				ImportState:       true,
				ImportStateVerify: true,
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
				ImportStateVerifyIgnore: []string{"status"},
			},
			// Update and Read testing
			{
				Config: testAccVolumeResourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teraswitch_volume.test", "region_id", "PIT1"),
					resource.TestCheckResourceAttr("teraswitch_volume.test", "display_name", "yeehaw"),
					resource.TestCheckResourceAttr("teraswitch_volume.test", "size", "20"),
					resource.TestCheckResourceAttr("teraswitch_volume.test", "volume_type", "nvme"),
					resource.TestCheckResourceAttr("teraswitch_volume.test", "description", "test 111"),
					resource.TestCheckResourceAttrSet("teraswitch_volume.test", "status"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccVolumeResourceConfig() string {
	return `
provider "teraswitch" {}

resource "teraswitch_volume" "test" {
	region_id      = "PIT1"
	display_name   = "yeehaw"
	size           = 20
	volume_type    = "nvme"
	description    = "test 111"
}
`
}
