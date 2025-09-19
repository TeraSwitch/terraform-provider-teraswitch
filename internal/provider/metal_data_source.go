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
var _ datasource.DataSource = &MetalDataSource{}

func NewMetalDataSource() datasource.DataSource {
	return &MetalDataSource{}
}

// MetalDataSource defines the data source implementation.
type MetalDataSource struct {
	providerData *ProviderData
}

// MetalDataSourceModel describes the data source data model.
type MetalDataSourceModel struct {
	ID                 types.Int64   `tfsdk:"id"`
	ProjectID          types.Int64   `tfsdk:"project_id"`
	RegionID           types.String  `tfsdk:"region_id"`
	DisplayName        types.String  `tfsdk:"display_name"`
	TierID             types.String  `tfsdk:"tier_id"`
	ImageID            types.String  `tfsdk:"image_id"`
	Status             types.String  `tfsdk:"status"`
	PowerState         types.String  `tfsdk:"power_state"`
	CurrentTask        types.String  `tfsdk:"current_task"`
	IPAddresses        types.List    `tfsdk:"ip_addresses"`
	IPv4DefaultGateway types.String  `tfsdk:"ipv4_default_gateway"`
	IPv6DefaultGateway types.String  `tfsdk:"ipv6_default_gateway"`
	MemoryGB           types.Int64   `tfsdk:"memory_gb"`
	Tags               types.List    `tfsdk:"tags"`
	ReservePricing     types.Bool    `tfsdk:"reserve_pricing"`
	ActiveDate         types.String  `tfsdk:"active_date"`
	TerminationDate    types.String  `tfsdk:"termination_date"`
	MonthlyPrice       types.Float64 `tfsdk:"monthly_price"`
	HourlyPrice        types.Float64 `tfsdk:"hourly_price"`
	Created            types.String  `tfsdk:"created"`
}

func (d *MetalDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_metal"
}

func (d *MetalDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Metal data source allows you to retrieve information about a specific metal service.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the metal service to retrieve.",
				Required:            true,
			},
			"project_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the project that the metal service belongs to.",
				Computed:            true,
			},
			"region_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the region where the metal service is located.",
				Computed:            true,
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: "The display name of the metal service.",
				Computed:            true,
			},
			"tier_id": schema.StringAttribute{
				MarkdownDescription: "The service tier of the metal service.",
				Computed:            true,
			},
			"image_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the OS image applied to the metal service.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The current status of the metal service.",
				Computed:            true,
			},
			"power_state": schema.StringAttribute{
				MarkdownDescription: "The power state of the metal service.",
				Computed:            true,
			},
			"current_task": schema.StringAttribute{
				MarkdownDescription: "The current task being performed on the metal service.",
				Computed:            true,
			},
			"ip_addresses": schema.ListAttribute{
				MarkdownDescription: "The IP addresses associated with the metal service.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"ipv4_default_gateway": schema.StringAttribute{
				MarkdownDescription: "The IPv4 default gateway for the metal service.",
				Computed:            true,
			},
			"ipv6_default_gateway": schema.StringAttribute{
				MarkdownDescription: "The IPv6 default gateway for the metal service.",
				Computed:            true,
			},
			"memory_gb": schema.Int64Attribute{
				MarkdownDescription: "The amount of memory in GB allocated to the metal service.",
				Computed:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "Tags associated with the metal service.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"reserve_pricing": schema.BoolAttribute{
				MarkdownDescription: "Whether the metal service is using reserve pricing.",
				Computed:            true,
			},
			"active_date": schema.StringAttribute{
				MarkdownDescription: "The date when the metal service became active.",
				Computed:            true,
			},
			"termination_date": schema.StringAttribute{
				MarkdownDescription: "The date when the metal service was terminated.",
				Computed:            true,
			},
			"monthly_price": schema.Float64Attribute{
				MarkdownDescription: "The current monthly price for the metal service.",
				Computed:            true,
			},
			"hourly_price": schema.Float64Attribute{
				MarkdownDescription: "The current hourly price for the metal service.",
				Computed:            true,
			},
			"created": schema.StringAttribute{
				MarkdownDescription: "The date when the metal service was created.",
				Computed:            true,
			},
		},
	}
}

func (d *MetalDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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

func (d *MetalDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MetalDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get the metal service by ID
	res, err := d.providerData.client.GetV2MetalIdWithResponse(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read metal service, got error: %s", err))
		return
	}

	if res.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read metal service, got error: %s", string(res.Body)),
		)
		return
	}

	metalService := res.JSON200.Result
	if metalService == nil {
		resp.Diagnostics.AddError("Client Error", "Metal service not found")
		return
	}

	// Map the API response to the data source model
	data.ID = types.Int64Value(*metalService.Id)
	if metalService.ProjectId != nil {
		data.ProjectID = types.Int64Value(*metalService.ProjectId)
	} else {
		data.ProjectID = types.Int64Null()
	}

	if metalService.RegionId != nil {
		data.RegionID = types.StringValue(*metalService.RegionId)
	} else {
		data.RegionID = types.StringNull()
	}

	if metalService.DisplayName != nil {
		data.DisplayName = types.StringValue(*metalService.DisplayName)
	} else {
		data.DisplayName = types.StringNull()
	}

	if metalService.TierId != nil {
		data.TierID = types.StringValue(*metalService.TierId)
	} else {
		data.TierID = types.StringNull()
	}

	if metalService.ImageId != nil {
		data.ImageID = types.StringValue(*metalService.ImageId)
	} else {
		data.ImageID = types.StringNull()
	}

	if metalService.Status != nil {
		data.Status = types.StringValue(*metalService.Status)
	} else {
		data.Status = types.StringNull()
	}

	if metalService.PowerState != nil {
		data.PowerState = types.StringValue(*metalService.PowerState)
	} else {
		data.PowerState = types.StringNull()
	}

	if metalService.CurrentTask != nil {
		data.CurrentTask = types.StringValue(*metalService.CurrentTask)
	} else {
		data.CurrentTask = types.StringNull()
	}

	if metalService.IpAddresses != nil {
		ipList, diags := types.ListValueFrom(ctx, types.StringType, *metalService.IpAddresses)
		resp.Diagnostics.Append(diags...)
		data.IPAddresses = ipList
	} else {
		data.IPAddresses = types.ListNull(types.StringType)
	}

	if metalService.Ipv4DefaultGateway != nil {
		data.IPv4DefaultGateway = types.StringValue(*metalService.Ipv4DefaultGateway)
	} else {
		data.IPv4DefaultGateway = types.StringNull()
	}

	if metalService.Ipv6DefaultGateway != nil {
		data.IPv6DefaultGateway = types.StringValue(*metalService.Ipv6DefaultGateway)
	} else {
		data.IPv6DefaultGateway = types.StringNull()
	}

	if metalService.MemoryGb != nil {
		data.MemoryGB = types.Int64Value(int64(*metalService.MemoryGb))
	} else {
		data.MemoryGB = types.Int64Null()
	}

	if metalService.Tags != nil {
		tagsList, diags := types.ListValueFrom(ctx, types.StringType, *metalService.Tags)
		resp.Diagnostics.Append(diags...)
		data.Tags = tagsList
	} else {
		data.Tags = types.ListNull(types.StringType)
	}

	if metalService.ReservePricing != nil {
		data.ReservePricing = types.BoolValue(*metalService.ReservePricing)
	} else {
		data.ReservePricing = types.BoolNull()
	}

	if metalService.ActiveDate != nil {
		data.ActiveDate = types.StringValue(*metalService.ActiveDate)
	} else {
		data.ActiveDate = types.StringNull()
	}

	if metalService.TerminationDate != nil {
		data.TerminationDate = types.StringValue(*metalService.TerminationDate)
	} else {
		data.TerminationDate = types.StringNull()
	}

	if metalService.MonthlyPrice != nil {
		data.MonthlyPrice = types.Float64Value(*metalService.MonthlyPrice)
	} else {
		data.MonthlyPrice = types.Float64Null()
	}

	if metalService.HourlyPrice != nil {
		data.HourlyPrice = types.Float64Value(*metalService.HourlyPrice)
	} else {
		data.HourlyPrice = types.Float64Null()
	}

	if metalService.Created != nil {
		data.Created = types.StringValue(*metalService.Created)
	} else {
		data.Created = types.StringNull()
	}

	tflog.Trace(ctx, "read metal data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
