package provider

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

var (
	metalDataSourceCfg = testAccMetalDataSourceConfig{
		MetalID: PtrTo("12345"),
	}
)

func TestAccMetalDataSource(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping, metal tests are not run in CI")
		return
	}

	if envMetalID := os.Getenv("TERASWITCH_TEST_METAL_ID"); envMetalID != "" {
		metalDataSourceCfg.MetalID = PtrTo(envMetalID)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: metalDataSourceCfg.String(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.teraswitch_metal.test", "id"),
					resource.TestCheckResourceAttrSet("data.teraswitch_metal.test", "project_id"),
					resource.TestCheckResourceAttrSet("data.teraswitch_metal.test", "region_id"),
					resource.TestCheckResourceAttrSet("data.teraswitch_metal.test", "tier_id"),
					resource.TestCheckResourceAttrSet("data.teraswitch_metal.test", "display_name"),
					resource.TestCheckResourceAttrSet("data.teraswitch_metal.test", "status"),
					resource.TestCheckResourceAttrSet("data.teraswitch_metal.test", "ip_addresses"),
				),
			},
			{
				Config: metalDataSourceCfg.String(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.teraswitch_metal.test", "id", *metalDataSourceCfg.MetalID),
					resource.TestCheckResourceAttrSet("data.teraswitch_metal.test", "created"),
					resource.TestCheckResourceAttrSet("data.teraswitch_metal.test", "memory_gb"),
				),
			},
		},
	})
}

type testAccMetalDataSourceConfig struct {
	MetalID *string
}

func (c testAccMetalDataSourceConfig) String(t *testing.T) string {
	tpl := `
provider "teraswitch" {}

data "teraswitch_metal" "test" {
  id = {{orNull .MetalID}}
}
`

	funcMap := template.FuncMap{
		"orNull": func(v interface{}) string {
			if v == nil {
				return "null"
			}
			switch value := v.(type) {
			case *string:
				if value == nil {
					return "null"
				}
				return *value
			default:
				require.NoError(t, fmt.Errorf("unknown type in template: %T", value))
				return ""
			}
		},
	}

	buf := strings.Builder{}
	tmpl, err := template.New("test").Funcs(funcMap).Parse(tpl)
	require.NoError(t, err)

	err = tmpl.Execute(&buf, c)
	require.NoError(t, err)

	return buf.String()
}
