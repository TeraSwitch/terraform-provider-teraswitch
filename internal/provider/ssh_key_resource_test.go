package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSshKeyResource(t *testing.T) {
	if os.Getenv("TERASWITCH_API_KEY") == "" {
		t.Skip("Skipping, api key not provided")
		return
	}

	rName := acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSshKeyResourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teraswitch_ssh_key.test", "display_name", fmt.Sprintf("tf-acc-test-%s", rName)),
					resource.TestCheckResourceAttrSet("teraswitch_ssh_key.test", "id"),
					resource.TestCheckResourceAttrSet("teraswitch_ssh_key.test", "key"),
					resource.TestCheckResourceAttrSet("teraswitch_ssh_key.test", "project_id"),
					resource.TestCheckResourceAttrSet("teraswitch_ssh_key.test", "created"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "teraswitch_ssh_key.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"created"}, // API returns different timestamp formats
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccSshKeyResourceConfig(rName string) string {
	return fmt.Sprintf(`
provider "teraswitch" {}

resource "teraswitch_ssh_key" "test" {
	display_name = "tf-acc-test-%s"
	key          = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAccTestKeyForTerraformProviderTesting12345678 acc-test@terraform"
}
`, rName)
}
