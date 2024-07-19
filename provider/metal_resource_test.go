package provider

import (
	"fmt"
	"strings"
	"testing"
	"text/template"

	"github.com/TeraSwitch/terraform-provider/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

var (
	metalCfg2388g = testAccMetalResourceConfig{
		RegionID:    PtrTo("LAX1"),
		DisplayName: PtrTo("yeehaw"),
		TierID:      PtrTo("2388g"),
		ProjectID:   PtrTo(480),
		SSHKeyIDs:   PtrTo([]int{588}),
		Tags:        PtrTo([]string{"tag1", "tag2"}),
		MemoryGB:    PtrTo(64),
		Disks: PtrTo(map[string]string{
			"nvme0n1": "960g",
			"nvme1n1": "960g",
		}),
		RaidArrays: PtrTo([]RaidArray{
			{
				Name:       PtrTo("md0"),
				Type:       PtrTo("Raid1"),
				Members:    PtrTo([]string{"nvme0n1", "nvme1n1"}),
				FileSystem: PtrTo(string(client.FileSystemExt4)),
				MountPoint: PtrTo("/"),
			},
		}),
		ImageID:        PtrTo("ubuntu-noble"),
		ReservePricing: PtrTo(false),
	}
	_ = metalCfg2388g

	metalCfg7950 = testAccMetalResourceConfig{
		RegionID:    PtrTo("SLC1"),
		DisplayName: PtrTo("yeehaw-amd"),
		TierID:      PtrTo("9254"),
		ProjectID:   PtrTo(480),
		SSHKeyIDs:   PtrTo([]int{588}),
		Tags:        PtrTo([]string{"tag1", "tag2"}),
		MemoryGB:    PtrTo(384),
		Disks: PtrTo(map[string]string{
			"nvme0n1": "480g", // boss card
			"nvme1n1": "3.84t",
			"nvme2n1": "3.84t",
		}),
		RaidArrays: PtrTo([]RaidArray{
			{
				Name:       PtrTo("md0"),
				Type:       PtrTo("Raid1"),
				Members:    PtrTo([]string{"nvme1n1", "nvme2n1"}),
				FileSystem: PtrTo(string(client.FileSystemExt4)),
				MountPoint: PtrTo("/"),
			},
		}),
		ImageID:        PtrTo("ubuntu-noble"),
		ReservePricing: PtrTo(false),
		WaitForReady:   PtrTo(true),
	}
)

func TestAccMetalResource(t *testing.T) {
	cfg1 := metalCfg7950

	cfg2 := cfg1
	cfg2.DisplayName = PtrTo("yeehaw2")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: cfg1.String(t),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("teraswitch_metal.test", "region_id", *cfg1.RegionID),
					resource.TestCheckResourceAttr("teraswitch_metal.test", "display_name", *cfg1.DisplayName),
					resource.TestCheckResourceAttr("teraswitch_metal.test", "tier_id", *cfg1.TierID),
					resource.TestCheckResourceAttrSet("teraswitch_metal.test", "id"),
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
					resource.TestCheckResourceAttr("teraswitch_metal.test", "display_name", *cfg2.DisplayName),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

type testAccMetalResourceConfig struct {
	RegionID          *string
	DisplayName       *string
	TierID            *string
	ProjectID         *int
	SSHKeyIDs         *[]int
	Password          *string
	Tags              *[]string
	MemoryGB          *int
	Disks             *map[string]string
	RaidArrays        *[]RaidArray
	ImageID           *string
	ReservePricing    *bool
	DesiredPowerState *string
	WaitForReady      *bool
}

type RaidArray struct {
	Name       *string
	Type       *string
	Members    *[]string
	FileSystem *string
	MountPoint *string
	SizeBytes  *int
}

func (c testAccMetalResourceConfig) String(t *testing.T) string {
	tpl := `
provider "teraswitch" {}

resource "teraswitch_metal" "test" {
	region_id    = {{orNull .RegionID}}
	display_name = {{orNull .DisplayName}}
	tier_id      = {{orNull .TierID}}
	project_id   = {{orNull .ProjectID}}
	ssh_key_ids  = {{orNull .SSHKeyIDs}}
	password     = {{orNull .Password}}
	tags         = {{orNull .Tags}}
	memory_gb    = {{orNull .MemoryGB}}
	disks = {
		{{- range $disk, $size := .Disks }}
		"{{$disk}}": "{{$size}}",
		{{- end }}
	}
	raid_arrays = [
		{{- range .RaidArrays }}
		{
			name = {{orNull .Name}}
			type = {{orNull .Type}}
			members = {{orNull .Members }}
			file_system = {{orNull .FileSystem}}
			mount_point = {{orNull .MountPoint}}
			size_bytes = {{orNull .SizeBytes}}
		},
		{{- end }}
	]
	image_id        = {{orNull .ImageID}}
	reserve_pricing = {{orNull .ReservePricing}}
	desired_power_state = {{orNull .DesiredPowerState}}
	wait_for_ready = {{orNull .WaitForReady}}
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
