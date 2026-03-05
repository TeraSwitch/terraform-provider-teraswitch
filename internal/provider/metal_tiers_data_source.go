package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &MetalTiersDataSource{}

func NewMetalTiersDataSource() datasource.DataSource {
	return &MetalTiersDataSource{}
}

// MetalTiersDataSource defines the data source implementation.
type MetalTiersDataSource struct {
	providerData *ProviderData
}

// MetalTierModel describes a single metal tier.
type MetalTierModel struct {
	ID             types.String  `tfsdk:"id"`
	CPU            types.String  `tfsdk:"cpu"`
	CPUDescription types.String  `tfsdk:"cpu_description"`
	HourlyPrice    types.Float64 `tfsdk:"hourly_price"`
	MonthlyPrice   types.Float64 `tfsdk:"monthly_price"`
}

// MetalTiersDataSourceModel describes the data source data model.
type MetalTiersDataSourceModel struct {
	Tiers []MetalTierModel `tfsdk:"tiers"`
}

func (d *MetalTiersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_metal_tiers"
}

func (d *MetalTiersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Metal Tiers data source allows you to retrieve all available metal server tiers with pricing.",

		Attributes: map[string]schema.Attribute{
			"tiers": schema.ListNestedAttribute{
				MarkdownDescription: "List of available metal tiers.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The ID of the metal tier (e.g., c3.small.x86).",
							Computed:            true,
						},
						"cpu": schema.StringAttribute{
							MarkdownDescription: "The CPU model for the tier.",
							Computed:            true,
						},
						"cpu_description": schema.StringAttribute{
							MarkdownDescription: "Description of the CPU in terms of cores and threads (e.g., 4c / 8t).",
							Computed:            true,
						},
						"hourly_price": schema.Float64Attribute{
							MarkdownDescription: "The hourly price for the tier.",
							Computed:            true,
						},
						"monthly_price": schema.Float64Attribute{
							MarkdownDescription: "The monthly price for the tier.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *MetalTiersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.providerData = client
}

func (d *MetalTiersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MetalTiersDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := d.providerData.client.GetV2MetalTiersWithResponse(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read metal tiers, got error: %s", err))
		return
	}

	if res.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read metal tiers, got error: %s", string(res.Body)),
		)
		return
	}

	if res.JSON200 == nil || res.JSON200.Result == nil {
		data.Tiers = []MetalTierModel{}
	} else {
		tiers := make([]MetalTierModel, 0, len(*res.JSON200.Result))
		for _, tier := range *res.JSON200.Result {
			tierModel := MetalTierModel{}

			if tier.Id != nil {
				tierModel.ID = types.StringValue(*tier.Id)
			} else {
				tierModel.ID = types.StringNull()
			}

			if tier.Cpu != nil {
				tierModel.CPU = types.StringValue(*tier.Cpu)
			} else {
				tierModel.CPU = types.StringNull()
			}

			if tier.CpuDescription != nil {
				tierModel.CPUDescription = types.StringValue(*tier.CpuDescription)
			} else {
				tierModel.CPUDescription = types.StringNull()
			}

			if tier.HourlyPrice != nil {
				tierModel.HourlyPrice = types.Float64Value(*tier.HourlyPrice)
			} else {
				tierModel.HourlyPrice = types.Float64Null()
			}

			if tier.MonthlyPrice != nil {
				tierModel.MonthlyPrice = types.Float64Value(*tier.MonthlyPrice)
			} else {
				tierModel.MonthlyPrice = types.Float64Null()
			}

			tiers = append(tiers, tierModel)
		}
		data.Tiers = tiers
	}

	tflog.Trace(ctx, "read metal tiers data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
