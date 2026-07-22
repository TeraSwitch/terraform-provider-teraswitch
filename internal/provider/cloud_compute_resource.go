package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/TeraSwitch/terraform-provider/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &CloudComputeResource{}
var _ resource.ResourceWithImportState = &CloudComputeResource{}

func NewCloudComputeResource() resource.Resource {
	return &CloudComputeResource{}
}

// CloudComputeResource defines the resource implementation.
type CloudComputeResource struct {
	providerData *ProviderData
}

// CloudComputeResourceModel describes the resource data model.
type CloudComputeResourceModel struct {
	ID                types.Int64  `tfsdk:"id"`
	ProjectID         types.Int64  `tfsdk:"project_id"`
	RegionID          types.String `tfsdk:"region_id"`
	TierID            types.String `tfsdk:"tier_id"`
	ImageID           types.String `tfsdk:"image_id"`
	DisplayName       types.String `tfsdk:"display_name"`
	SSHKeyIDs         types.List   `tfsdk:"ssh_key_ids"`
	Password          types.String `tfsdk:"password"`
	BootSize          types.Int64  `tfsdk:"boot_size"`
	UserData          types.String `tfsdk:"user_data"`
	Tags              types.List   `tfsdk:"tags"`
	IPAddresses       types.List   `tfsdk:"ip_addresses"`
	DesiredPowerState types.String `tfsdk:"desired_power_state"`
	SkipWaitForReady  types.Bool   `tfsdk:"skip_wait_for_ready"`
}

func (r *CloudComputeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_compute"
}

func (r *CloudComputeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Cloud Compute Instance. **Deprecated:** cloud instances are no longer orderable; creating a new instance returns an error. Existing instances remain manageable (read, power, destroy).",
		DeprecationMessage:  "Cloud instances are no longer orderable. New teraswitch_cloud_compute resources cannot be created; existing instances remain manageable (read, power, destroy).",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Id of the compute instance",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the project that the metal will be created in.",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"region_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the region that the metal will be created in.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tier_id": schema.StringAttribute{
				MarkdownDescription: "The service tier to be created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"image_id": schema.StringAttribute{
				MarkdownDescription: "The image to use when creating this service. Available images can be retrieved via the images endpoint.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: "The display name of the instance.",
				Required:            true,
			},
			"ssh_key_ids": schema.ListAttribute{
				MarkdownDescription: "The SSH key ids to be added to the service. These keys will be added to the authorized_keys file for the root user.",
				Optional:            true,
				ElementType:         types.Int64Type,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.AtLeastOneOf(path.MatchRoot("password")),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The password to be set for the root user. If not provided, a random password will be generated.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.AtLeastOneOf(path.MatchRoot("ssh_key_ids")),
				},
			},
			"boot_size": schema.Int64Attribute{
				MarkdownDescription: "The size of the boot disk.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"user_data": schema.StringAttribute{
				MarkdownDescription: "Additional user data.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "Tags to be added to the instance.",
				Optional:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},

			"ip_addresses": schema.ListAttribute{
				MarkdownDescription: "IP addresses of the instance.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"desired_power_state": schema.StringAttribute{
				MarkdownDescription: "The desired power state for the compute instance.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("On"),
				Validators: []validator.String{
					stringvalidator.OneOf("On", "Off"),
				},
			},
			"skip_wait_for_ready": schema.BoolAttribute{
				MarkdownDescription: "Skips waiting for the instance to become ready on create. `ip_addresses` will be nil on initial create.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func (r *CloudComputeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CloudComputeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Cloud instances are no longer orderable (PTL-1622/PTL-1623); the create API returns
	// HTTP 410 Gone. Fail fast with an actionable message instead of a wrapped 410 body.
	resp.Diagnostics.AddError(
		"Cloud instances are no longer orderable",
		"Creating cloud compute instances has been retired and the TeraSwitch API no longer "+
			"accepts new instance orders. Existing instances remain fully manageable (read, power, "+
			"and destroy), but new teraswitch_cloud_compute resources cannot be created.",
	)
}

func (r *CloudComputeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CloudComputeResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CloudComputeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CloudComputeResourceModel
	var state CloudComputeResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the current state
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.IPAddresses = state.IPAddresses

	if !plan.DesiredPowerState.Equal(state.DesiredPowerState) {
		var cmd client.PowerCommand
		switch plan.DesiredPowerState.ValueString() {
		case "On":
			cmd = client.PowerOn
		case "Off":
			cmd = client.PowerOff
		default:
			resp.Diagnostics.AddError("Provider Error",
				fmt.Sprintf("unknown desired_power_state: %s", plan.DesiredPowerState.ValueString()),
			)

			return
		}

		tflog.Debug(ctx, "updating power state", map[string]interface{}{
			"old_power_state": state.DesiredPowerState.String(),
			"new_power_state": string(cmd),
		})

		res, err := r.providerData.client.PostV2InstanceIdPowerCommandWithResponse(ctx, state.ID.ValueInt64(),
			&client.PostV2InstanceIdPowerCommandParams{
				Command: PtrTo(cmd),
			},
		)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update v2 metal power status, got error: %s", err))
			return
		}

		if res.StatusCode() != http.StatusOK {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update v2 instance power status, got error: %s", string(res.Body)))
			return
		}

		tflog.Trace(ctx, "power state updated")
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CloudComputeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CloudComputeResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.providerData.client.DeleteV2InstanceIdWithResponse(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to delete v2 instance, got error: %s", err),
		)
		return
	}

	if res.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to delete v2 instance, got error: %s", string(res.Body)),
		)
		return
	}
}

func (r *CloudComputeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.AddError("Not supported",
		"Compute instance resources don't currently support importing.",
	)
	// resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
