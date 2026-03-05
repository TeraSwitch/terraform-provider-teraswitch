package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSshKeysDataSource(t *testing.T) {
	if os.Getenv("TERASWITCH_API_KEY") == "" {
		t.Skip("Skipping, api key not provided")
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSshKeysDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.teraswitch_ssh_keys.test", "ssh_keys.#"),
				),
			},
		},
	})
}

func testAccSshKeysDataSourceConfig() string {
	return `
provider "teraswitch" {}

data "teraswitch_ssh_keys" "test" {}
`
}
