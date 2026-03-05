package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/TeraSwitch/terraform-provider/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &RegionsDataSource{}

func NewRegionsDataSource() datasource.DataSource {
	return &RegionsDataSource{}
}

// RegionsDataSource defines the data source implementation.
type RegionsDataSource struct {
	providerData *ProviderData
}

// RegionModel describes a single region.
type RegionModel struct {
	ID           types.String   `tfsdk:"id"`
	Name         types.String   `tfsdk:"name"`
	Country      types.String   `tfsdk:"country"`
	City         types.String   `tfsdk:"city"`
	Location     types.String   `tfsdk:"location"`
	ServiceTypes []types.String `tfsdk:"service_types"`
}

// RegionsDataSourceModel describes the data source data model.
type RegionsDataSourceModel struct {
	ServiceType types.String  `tfsdk:"service_type"`
	Regions     []RegionModel `tfsdk:"regions"`
}

func (d *RegionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_regions"
}

func (d *RegionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Regions data source allows you to retrieve all available regions where services can be deployed.",

		Attributes: map[string]schema.Attribute{
			"service_type": schema.StringAttribute{
				MarkdownDescription: "Filter regions by service type. Valid values are: Metal, Instance, ObjectStorage, BlockStorage.",
				Optional:            true,
			},
			"regions": schema.ListNestedAttribute{
				MarkdownDescription: "List of available regions.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The ID of the region (e.g., us-east-1).",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The display name of the region.",
							Computed:            true,
						},
						"country": schema.StringAttribute{
							MarkdownDescription: "The country where the region is located.",
							Computed:            true,
						},
						"city": schema.StringAttribute{
							MarkdownDescription: "The city where the region is located.",
							Computed:            true,
						},
						"location": schema.StringAttribute{
							MarkdownDescription: "The geographic location of the region (e.g., North America, Europe).",
							Computed:            true,
						},
						"service_types": schema.ListAttribute{
							MarkdownDescription: "List of service types available in this region.",
							Computed:            true,
							ElementType:         types.StringType,
						},
					},
				},
			},
		},
	}
}

func (d *RegionsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.providerData = providerData
}

func (d *RegionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RegionsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	params := &client.GetV2RegionParams{}
	if !data.ServiceType.IsNull() && !data.ServiceType.IsUnknown() {
		serviceType := data.ServiceType.ValueString()
		params.ServiceType = &serviceType
	}

	res, err := d.providerData.client.GetV2RegionWithResponse(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read regions, got error: %s", err))
		return
	}

	if res.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read regions, got status %d: %s", res.StatusCode(), string(res.Body)),
		)
		return
	}

	if res.JSON200 == nil || res.JSON200.Result == nil {
		data.Regions = []RegionModel{}
	} else {
		regions := make([]RegionModel, 0, len(*res.JSON200.Result))
		for _, region := range *res.JSON200.Result {
			regionModel := RegionModel{}

			if region.Id != nil {
				regionModel.ID = types.StringValue(*region.Id)
			} else {
				regionModel.ID = types.StringNull()
			}

			if region.Name != nil {
				regionModel.Name = types.StringValue(*region.Name)
			} else {
				regionModel.Name = types.StringNull()
			}

			if region.Country != nil {
				regionModel.Country = types.StringValue(*region.Country)
			} else {
				regionModel.Country = types.StringNull()
			}

			if region.City != nil {
				regionModel.City = types.StringValue(*region.City)
			} else {
				regionModel.City = types.StringNull()
			}

			if region.Location != nil {
				regionModel.Location = types.StringValue(*region.Location)
			} else {
				regionModel.Location = types.StringNull()
			}

			if region.ServiceTypes != nil {
				serviceTypes := make([]types.String, 0, len(*region.ServiceTypes))
				for _, st := range *region.ServiceTypes {
					serviceTypes = append(serviceTypes, types.StringValue(st))
				}
				regionModel.ServiceTypes = serviceTypes
			} else {
				regionModel.ServiceTypes = []types.String{}
			}

			regions = append(regions, regionModel)
		}
		data.Regions = regions
	}

	tflog.Trace(ctx, "read regions data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
