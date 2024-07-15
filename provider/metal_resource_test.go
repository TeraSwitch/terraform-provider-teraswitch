package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMetalResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccMetalResourceConfigReal(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teraswitch_metal.test", "region_id", "LAX1"),
					resource.TestCheckResourceAttr("teraswitch_metal.test", "display_name", "yeehaw"),
					resource.TestCheckResourceAttr("teraswitch_metal.test", "tier_id", "2388g"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "scaffolding_example.test",
				ImportState:       true,
				ImportStateVerify: true,
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
				ImportStateVerifyIgnore: []string{"configurable_attribute", "defaulted"},
			},
			// Update and Read testing
			{
				Config: testAccExampleResourceConfig("two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("scaffolding_example.test", "configurable_attribute", "two"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccMetalResourceConfigReal() string {
	return fmt.Sprintf(`
provider "teraswitch" {}

resource "teraswitch_metal" "test" {
	region_id    = "LAX1"
	display_name = "yeehaw"
	tier_id      = "2388g"
	project_id   = 480
	ssh_key_ids  = [588]
	# password     = "wnp3xev!vaq7duh-PUY"
	tags         = ["tag1", "tag2"]
	memory_gb    = 64
	disks = {
		"nvme0n1": "960g",
		"nvme1n1": "960g",
	}
	raid_arrays = [{
        name = "md0"
        type = "Raid1"
        members = [
            "nvme0n1",
            "nvme1n1",
        ]
        file_system = "Ext4"
        mount_point = "/"
		# size_bytes = null
	}]
	image_id = "ubuntu-noble"
	quantity        = 1
	reserve_pricing = false
}
`)
}
