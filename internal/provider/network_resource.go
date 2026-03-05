package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/TeraSwitch/terraform-provider/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &NetworkResource{}
var _ resource.ResourceWithImportState = &NetworkResource{}

func NewNetworkResource() resource.Resource {
	return &NetworkResource{}
}

// NetworkResource defines the resource implementation.
type NetworkResource struct {
	providerData *ProviderData
}

// NetworkResourceModel describes the resource data model.
type NetworkResourceModel struct {
	ID           types.String `tfsdk:"id"`
	RegionID     types.String `tfsdk:"region_id"`
	DisplayName  types.String `tfsdk:"display_name"`
	V4Subnet     types.String `tfsdk:"v4_subnet"`
	V4SubnetMask types.String `tfsdk:"v4_subnet_mask"`
}

func (r *NetworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

func (r *NetworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Network (DO NOT USE)",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ID of the network",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"region_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the region that the network will be created in",
				Required:            true,
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: "The display name of the network. This is optional",
				Optional:            true,
			},
			"v4_subnet": schema.StringAttribute{
				MarkdownDescription: "The IPv4 subnet that the network will use",
				Required:            true,
			},
			"v4_subnet_mask": schema.StringAttribute{
				MarkdownDescription: "The IPv4 subnet mask that the network will use",
				Required:            true,
			},
		},
	}
}

func (r *NetworkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.providerData = client
}

func (r *NetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NetworkResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.providerData.client.PostV2NetworkWithResponse(ctx, &client.PostV2NetworkParams{
		ProjectId: &r.providerData.projectID,
	}, client.PostV2NetworkJSONRequestBody{
		DisplayName:  data.DisplayName.ValueStringPointer(),
		RegionId:     data.RegionID.ValueStringPointer(),
		V4Subnet:     data.V4Subnet.ValueStringPointer(),
		V4SubnetMask: data.V4SubnetMask.ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create network, got error: %s", err))
		return
	}

	// Accept both 200 OK and 201 Created as success
	if res.StatusCode() != http.StatusOK && res.StatusCode() != http.StatusCreated {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to create network, got status %d: %s", res.StatusCode(), string(res.Body)),
		)
		return
	}

	// Handle response - JSON200 is only populated for status 200, manually decode for 201
	var networkID *string
	if res.JSON200 != nil && res.JSON200.Result != nil {
		networkID = res.JSON200.Result.Id
	} else if res.StatusCode() == http.StatusCreated && len(res.Body) > 0 {
		// Manually decode for 201 Created
		var createResp client.CreateNetworkResponse
		if err := json.Unmarshal(res.Body, &createResp); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to decode network response: %s", err))
			return
		}
		if createResp.Result != nil {
			networkID = createResp.Result.Id
		}
	}

	if networkID == nil {
		resp.Diagnostics.AddError("Client Error", "Network creation returned no ID")
		return
	}

	data.ID = types.StringPointerValue(networkID)

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NetworkResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NetworkResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data NetworkResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *NetworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
