package provider

import (
	"fmt"
	"strings"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

var (
	cloudCfg_1c1g = testAccCloudInstanceResourceConfig{
		RegionID:         PtrTo("PIT1"),
		DisplayName:      PtrTo("yeehaw"),
		TierID:           PtrTo("s1.1c1g"),
		ProjectID:        PtrTo(480),
		SSHKeyIDs:        PtrTo([]int{588}),
		Tags:             PtrTo([]string{"tag1", "tag2"}),
		BootSize:         PtrTo(64),
		ImageID:          PtrTo("ubuntu-noble"),
		SkipWaitForReady: PtrTo(false),
	}
)

func TestAccCloudComputeResource(t *testing.T) {
	cfg1 := cloudCfg_1c1g

	cfg2 := cfg1
	cfg2.DesiredPowerState = PtrTo("Off")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: cloudCfg_1c1g.String(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("teraswitch_cloud_compute.test", "id"),
					resource.TestCheckResourceAttr("teraswitch_cloud_compute.test", "project_id", fmt.Sprintf("%d", *cfg1.ProjectID)),
					resource.TestCheckResourceAttr("teraswitch_cloud_compute.test", "region_id", *cfg1.RegionID),
					resource.TestCheckResourceAttr("teraswitch_cloud_compute.test", "tier_id", *cfg1.TierID),
					resource.TestCheckResourceAttr("teraswitch_cloud_compute.test", "image_id", *cfg1.ImageID),
					resource.TestCheckResourceAttr("teraswitch_cloud_compute.test", "display_name", *cfg1.DisplayName),
					resource.TestCheckResourceAttr("teraswitch_cloud_compute.test", "ssh_key_ids.#", fmt.Sprintf("%d", len(*cfg1.SSHKeyIDs))),
					resource.TestCheckResourceAttr("teraswitch_cloud_compute.test", "ssh_key_ids.0", fmt.Sprintf("%d", (*cfg1.SSHKeyIDs)[0])),
					resource.TestCheckResourceAttr("teraswitch_cloud_compute.test", "boot_size", fmt.Sprintf("%d", *cfg1.BootSize)),
					resource.TestCheckResourceAttr("teraswitch_cloud_compute.test", "ip_addresses.#", "2"),
				),
			},
			// ImportState testing
			// {
			// 	ResourceName:      "scaffolding_example.test",
			// 	ImportState:       true,
			// 	ImportStateVerify: true,
			// 	// This is not normally necessary, but is here because this
			// 	// example code does not have an actual upstream service.
			// 	// Once the Read method is able to refresh information from
			// 	// the upstream service, this can be removed.
			// 	ImportStateVerifyIgnore: []string{"configurable_attribute", "defaulted"},
			// },
			// Update and Read testing
			{
				Config: cfg2.String(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("teraswitch_cloud_compute.test", "id"),
					resource.TestCheckResourceAttr("teraswitch_cloud_compute.test", "desired_power_state", "Off"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

type testAccCloudInstanceResourceConfig struct {
	ProjectID         *int
	RegionID          *string
	TierID            *string
	ImageID           *string
	DisplayName       *string
	SSHKeyIDs         *[]int
	Password          *string
	BootSize          *int
	UserData          *string
	Tags              *[]string
	DesiredPowerState *string
	SkipWaitForReady  *bool
}

func (c testAccCloudInstanceResourceConfig) String(t *testing.T) string {
	tpl := `
provider "teraswitch" {}

resource "teraswitch_cloud_compute" "test" {
	project_id          = {{orNull .ProjectID}}
	region_id           = {{orNull .RegionID}}
	tier_id             = {{orNull .TierID}}
	image_id            = {{orNull .ImageID}}
	display_name        = {{orNull .DisplayName}}
	ssh_key_ids         = {{orNull .SSHKeyIDs}}
	password            = {{orNull .Password}}
	boot_size           = {{orNull .BootSize}}
	user_data           = {{orNull .UserData}}
	tags                = {{orNull .Tags}}
	desired_power_state = {{orNull .DesiredPowerState}}
	skip_wait_for_ready = {{orNull .SkipWaitForReady}}
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
				return fmt.Sprintf("%q", *value)
			case *int:
				if value == nil {
					return "null"
				}
				return fmt.Sprintf("%d", *value)
			case *bool:
				if value == nil {
					return "null"
				}
				return fmt.Sprintf("%t", *value)
			case *[]string:
				if value == nil {
					return "null"
				}
				if len(*value) == 0 {
					return "[]"
				}
				var result string
				for i, item := range *value {
					if i > 0 {
						result += ", "
					}
					result += fmt.Sprintf("%q", item)
				}
				return fmt.Sprintf("[%s]", result)
			case *[]int:
				if value == nil {
					return "null"
				}
				if len(*value) == 0 {
					return "[]"
				}
				var result string
				for i, item := range *value {
					if i > 0 {
						result += ", "
					}
					result += fmt.Sprintf("%d", item)
				}
				return fmt.Sprintf("[%s]", result)
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
