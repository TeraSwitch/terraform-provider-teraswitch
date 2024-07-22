package provider

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/TeraSwitch/terraform-provider/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
		MarkdownDescription: "Cloud Compute Instance",

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
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.providerData = client
}

func (r *CloudComputeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CloudComputeResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	body := client.PostV2InstanceJSONRequestBody{
		ProjectId:   data.ProjectID.ValueInt64Pointer(),
		RegionId:    data.RegionID.ValueString(),
		TierId:      data.TierID.ValueString(),
		ImageId:     data.ImageID.ValueStringPointer(),
		DisplayName: data.DisplayName.ValueStringPointer(),
		Password:    data.Password.ValueStringPointer(),
		BootSize:    i64PtrToi32Ptr(data.BootSize.ValueInt64Pointer()),
		UserData:    data.UserData.ValueStringPointer(),
	}

	resp.Diagnostics.Append(
		data.SSHKeyIDs.ElementsAs(ctx, &body.SshKeyIds, false)...,
	)
	resp.Diagnostics.Append(
		data.Tags.ElementsAs(ctx, &body.Tags, false)...,
	)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.providerData.client.PostV2InstanceWithResponse(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create get v2 instance, got error: %s", err))
		return
	}

	if res.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to create get v2 instance, got error: %s", string(res.Body)),
		)
		return
	}

	resBody := res.JSON200.Result
	data.ID = types.Int64PointerValue(resBody.Id)

	if !data.SkipWaitForReady.ValueBool() {
		final, err := r.waitInstanceStatus(ctx, *resBody.Id, "Active")
		if err != nil {
			resp.Diagnostics.AddError("Client Error",
				fmt.Sprintf("Unable to wait v2 instance ready, got error: %s", err),
			)
			return
		}

		var diags diag.Diagnostics
		data.IPAddresses, diags = types.ListValueFrom(ctx, types.StringType, *final.IpAddresses)
		resp.Diagnostics.Append(diags...)
	}

	tflog.Trace(ctx, "created a v2 instance")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CloudComputeResource) waitInstanceStatus(ctx context.Context, id int64, status string) (*client.CloudService, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(3 * time.Second):
		}

		res, err := r.providerData.client.GetV2InstanceIdWithResponse(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("send get metal v2 request: %w", err)
		}

		if res.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("v2 metal returned an error: %s", string(res.Body))
		}

		gotStatus := res.JSON200.Result.Status
		if gotStatus == nil {
			continue
		}

		if *gotStatus != status {
			tflog.Debug(ctx, "waiting for instance status %q, current status %q\n", map[string]interface{}{
				"want_status":    status,
				"current_status": *gotStatus,
			})
			continue
		}

		return res.JSON200.Result, nil
	}
}

func (r *CloudComputeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CloudComputeResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
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

			goto end
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
end:
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
