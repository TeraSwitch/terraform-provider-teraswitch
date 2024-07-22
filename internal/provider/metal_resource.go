package provider

import (
	"bytes"
	"context"
	"fmt"
	"io"
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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &MetalResource{}
var _ resource.ResourceWithImportState = &MetalResource{}

func NewMetalResource() resource.Resource {
	return &MetalResource{}
}

// MetalResource defines the resource implementation.
type MetalResource struct {
	providerData *ProviderData
}

// MetalResourceModel describes the resource data model.
type MetalResourceModel struct {
	ID                types.Int64           `tfsdk:"id"`
	ProjectID         types.Int64           `tfsdk:"project_id"`
	RegionID          types.String          `tfsdk:"region_id"`
	DisplayName       types.String          `tfsdk:"display_name"`
	TierID            types.String          `tfsdk:"tier_id"`
	ImageID           types.String          `tfsdk:"image_id"`
	SSHKeyIDs         types.List            `tfsdk:"ssh_key_ids"`
	Password          types.String          `tfsdk:"password"`
	UserData          types.String          `tfsdk:"user_data"`
	Tags              types.List            `tfsdk:"tags"`
	MemoryGB          types.Int64           `tfsdk:"memory_gb"`
	Disks             types.Map             `tfsdk:"disks"`
	Partitions        []MetalPartitionModel `tfsdk:"partitions"`
	RaidArrays        []MetalRaidArrayModel `tfsdk:"raid_arrays"`
	IPXEURL           types.String          `tfsdk:"ipxe_url"`
	TemplateID        types.Int64           `tfsdk:"template_id"`
	ReservePricing    types.Bool            `tfsdk:"reserve_pricing"`
	IPAddresses       types.List            `tfsdk:"ip_addresses"`
	DesiredPowerState types.String          `tfsdk:"desired_power_state"`
	WaitForReady      types.Bool            `tfsdk:"wait_for_ready"`
}

type MetalRaidArrayModel struct {
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	Members    types.List   `tfsdk:"members"`
	SizeBytes  types.Int64  `tfsdk:"size_bytes"`
	FileSystem types.String `tfsdk:"file_system"`
	MountPoint types.String `tfsdk:"mount_point"`
}

type MetalPartitionModel struct {
	Name       types.String `tfsdk:"name"`
	Device     types.String `tfsdk:"device"`
	SizeBytes  types.Int64  `tfsdk:"size_bytes"`
	FileSystem types.String `tfsdk:"file_system"`
	MountPoint types.String `tfsdk:"mount_point"`
}

func i64PtrToi32Ptr(ptr *int64) *int32 {
	var i32 *int32
	if ptr != nil {
		tmp := int32(*ptr)
		i32 = &tmp
	}
	return i32
}

func (r *MetalResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_metal"
}

var fileSystemValidator validator.String = stringvalidator.OneOf(
	string(client.FileSystemUnknown),
	string(client.FileSystemUnformatted),
	string(client.FileSystemExt2),
	string(client.FileSystemExt4),
	string(client.FileSystemXfs),
	string(client.FileSystemFat32),
	string(client.FileSystemVfat),
	string(client.FileSystemSwap),
	string(client.FileSystemRamfs),
	string(client.FileSystemTmpfs),
	string(client.FileSystemBtrfs),
	string(client.FileSystemZfsroot),
)

func (r *MetalResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Metal",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Id of the network",
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
			"display_name": schema.StringAttribute{
				MarkdownDescription: "The display name of the network. This is optional.",
				Optional:            true,
			},
			"tier_id": schema.StringAttribute{
				MarkdownDescription: "The service tier to be created. For metal, this is typically the server config. For example: 7302p-64g would create a Epyc 7302P system with 64G of ram. Tier availability can be retrieved using the regions endpoints.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"image_id": schema.StringAttribute{
				MarkdownDescription: "The image to use when creating this service. Available images can be retrieved via the images endpoint.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					// TODO: allow updating
					stringplanmodifier.RequiresReplace(),
				},
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
			"user_data": schema.StringAttribute{
				MarkdownDescription: "Additional user data.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "Tags to be added to the metal service.",
				Optional:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"memory_gb": schema.Int64Attribute{
				MarkdownDescription: "The amount of memory in GB to be allocated to the metal service.",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"disks": schema.MapAttribute{
				MarkdownDescription: "Dictionary of disk names and sizes in GB. If not specified, the default configuration for the metal tier will be used. The key is the disk name and the value is the size in GB.",
				Optional:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"partitions": schema.ListNestedAttribute{
				MarkdownDescription: "Partitions to be created on the metal service. Not specifying this will result in a single root partition being created.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the partition.",
							Required:            true,
						},
						"device": schema.StringAttribute{
							MarkdownDescription: "The name of the storage device to create a partition on. It can the name of a RAID array or a physical device.",
							Required:            true,
						},
						"size_bytes": schema.Int64Attribute{
							MarkdownDescription: "The size of the partition in bytes. If not specified, the remainder of the space will be used.",
							Optional:            true,
						},
						"file_system": schema.StringAttribute{
							MarkdownDescription: "The type of filesystem for the partition to be initialized with.",
							Required:            true,
							Validators:          []validator.String{fileSystemValidator},
						},
						"mount_point": schema.StringAttribute{
							MarkdownDescription: "The mount point of the partition.",
							Required:            true,
						},
					},
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"raid_arrays": schema.ListNestedAttribute{
				MarkdownDescription: "Raid arrays to be created on the metal service. Can reference physical device names or partitions from mediums of the same class.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the raid array. For example: \"md0\"",
							Required:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "The type of the raid array.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf(
									string(client.RaidTypeNone),
									string(client.RaidTypeRaid0),
									string(client.RaidTypeRaid1),
									string(client.RaidTypeUnknown),
								),
							},
						},
						"members": schema.ListAttribute{
							MarkdownDescription: "The members of the RAID array. These can be device or partition names.",
							Required:            true,
							ElementType:         types.StringType,
						},
						"file_system": schema.StringAttribute{
							MarkdownDescription: "The type of filesystem for the RAID array to be initialized with.",
							Required:            true,
							Validators:          []validator.String{fileSystemValidator},
						},
						"mount_point": schema.StringAttribute{
							MarkdownDescription: "The mount point of the array.",
							Required:            true,
						},
						"size_bytes": schema.Int64Attribute{
							MarkdownDescription: "The size of the RAID array in bytes.",
							Optional:            true,
						},
					},
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"ipxe_url": schema.StringAttribute{
				MarkdownDescription: "The URL to the script to use when enabling iPXE boot.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"template_id": schema.Int64Attribute{
				MarkdownDescription: "Template can be specified instead of image, partitions, sshKeyId, and userData.",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"reserve_pricing": schema.BoolAttribute{
				MarkdownDescription: "Denotes if the metal service is being reserved for a whole year. If so, it gets the discounted rate",
				Optional:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},

			"ip_addresses": schema.ListAttribute{
				MarkdownDescription: "IP addresses of the metal instance.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"desired_power_state": schema.StringAttribute{
				MarkdownDescription: "The desired power state for the metal instance.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("On"),
				Validators: []validator.String{
					stringvalidator.OneOf("On", "Off"),
				},
			},
			"wait_for_ready": schema.BoolAttribute{
				MarkdownDescription: "Waits for the instance to become ready on create.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func (r *MetalResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MetalResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MetalResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projID := data.ProjectID.ValueInt64Pointer()
	if projID == nil {
		data.ProjectID = types.Int64Value(r.providerData.projectID)
	}

	body := client.CreateMetalRequest{
		ProjectId:      data.ProjectID.ValueInt64Pointer(),
		RegionId:       data.RegionID.ValueString(),
		DisplayName:    data.DisplayName.ValueStringPointer(),
		TierId:         data.TierID.ValueString(),
		ImageId:        data.ImageID.ValueStringPointer(),
		Password:       data.Password.ValueStringPointer(),
		UserData:       data.UserData.ValueStringPointer(),
		MemoryGb:       i64PtrToi32Ptr(data.MemoryGB.ValueInt64Pointer()),
		IpxeUrl:        data.IPXEURL.ValueStringPointer(),
		TemplateId:     data.TemplateID.ValueInt64Pointer(),
		Quantity:       PtrTo(int32(1)),
		ReservePricing: data.ReservePricing.ValueBoolPointer(),
	}

	resp.Diagnostics.Append(
		data.SSHKeyIDs.ElementsAs(ctx, &body.SshKeyIds, false)...,
	)
	resp.Diagnostics.Append(
		data.Tags.ElementsAs(ctx, &body.Tags, false)...,
	)
	resp.Diagnostics.Append(
		data.Disks.ElementsAs(ctx, &body.Disks, false)...,
	)

	if len(data.Partitions) > 0 {
		var parts []client.Partition
		for _, dPart := range data.Partitions {
			part := client.Partition{
				Name:       dPart.Name.ValueStringPointer(),
				Device:     dPart.Device.ValueStringPointer(),
				SizeBytes:  dPart.SizeBytes.ValueInt64Pointer(),
				FileSystem: PtrTo(client.FileSystem(dPart.FileSystem.ValueString())),
				MountPoint: dPart.MountPoint.ValueStringPointer(),
			}
			parts = append(parts, part)
		}
		body.Partitions = &parts
	}

	if len(data.RaidArrays) > 0 {
		var arrays []client.RaidArray
		for _, dRaid := range data.RaidArrays {
			arr := client.RaidArray{
				FileSystem: PtrTo(client.FileSystem(dRaid.FileSystem.ValueString())),
				MountPoint: dRaid.MountPoint.ValueStringPointer(),
				Name:       dRaid.Name.ValueStringPointer(),
				SizeBytes:  dRaid.SizeBytes.ValueInt64Pointer(),
				Type:       PtrTo(client.RaidType(dRaid.Type.ValueString())),
			}
			resp.Diagnostics.Append(
				dRaid.Members.ElementsAs(ctx, &arr.Members, false)...,
			)
			arrays = append(arrays, arr)
		}
		body.RaidArrays = &arrays
	}

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.providerData.client.PostV2MetalWithApplicationWildcardPlusJSONBodyWithResponse(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create v2 metal, got error: %s", err))
		return
	}

	if res.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to create get v2 metal, got error: %s", string(res.Body)),
		)
		return
	}

	resBody := res.JSON200.Result
	if data.WaitForReady.ValueBool() {
		final, err := r.waitInstanceReady(ctx, *resBody.Id)
		if err != nil {
			resp.Diagnostics.AddError("Client Error",
				fmt.Sprintf("Unable to wait v2 metal instance ready, got error: %s", err),
			)
			return
		}

		var diags diag.Diagnostics
		data.IPAddresses, diags = types.ListValueFrom(ctx, types.StringType, *final.IpAddresses)
		resp.Diagnostics.Append(diags...)
	}

	data.ID = types.Int64Value(*resBody.Id)

	tflog.Trace(ctx, "created v2 metal")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MetalResource) waitInstanceReady(ctx context.Context, id int64) (*client.MetalService, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(3 * time.Second):
		}

		res, err := r.providerData.client.GetV2MetalIdWithResponse(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("send get metal v2 request: %w", err)
		}

		if res.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("v2 metal returned an error: %s", string(res.Body))
		}

		status := res.JSON200.Result.Status
		if status == nil {
			continue
		}

		if *status != "Active" {
			tflog.Debug(ctx, "waiting for instance status %q, current status %q\n", map[string]interface{}{
				"want_status":    "Active",
				"current_status": *status,
			})
			continue
		}

		return res.JSON200.Result, nil
	}
}

func (r *MetalResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MetalResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.providerData.client.GetV2MetalIdWithResponse(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get v2 metal, got error: %s", err))
		return
	}

	if res.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to get v2 metal, got error: %s", string(res.Body)),
		)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MetalResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MetalResourceModel
	var state MetalResourceModel

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

	if !plan.DisplayName.Equal(state.DisplayName) && !plan.DisplayName.IsNull() {
		tflog.Debug(ctx, "display name changed, updating...", map[string]interface{}{
			"old_display_name": state.DisplayName.String(),
			"new_display_name": plan.DisplayName.String(),
		})
		res, err := r.providerData.client.PostV2MetalIdRenameWithResponse(ctx, state.ID.ValueInt64(), client.PostV2MetalIdRenameJSONRequestBody{
			Name: plan.DisplayName.ValueStringPointer(),
		})
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update v2 metal name, got error: %s", err))
			return
		}

		if res.StatusCode() != http.StatusOK {
			resp.Diagnostics.AddError("Client Error",
				fmt.Sprintf("Unable to update v2 metal name, got error: %s", string(res.Body)),
			)
			return
		}
		tflog.Trace(ctx, "display name updated")
	}

	if !plan.DesiredPowerState.Equal(state.DesiredPowerState) && !plan.DesiredPowerState.IsNull() {
		var cmd client.PowerCommand
		switch plan.DesiredPowerState.ValueString() {
		case "On":
			cmd = client.PowerOn
		case "Off":
			cmd = client.PowerOff
		default:
			resp.Diagnostics.AddError("Provider Error",
				fmt.Sprintf("unknown desired_power_state: %q", plan.DesiredPowerState.ValueString()),
			)

			goto end
		}

		tflog.Debug(ctx, "updating power state", map[string]interface{}{
			"old_power_state": state.DesiredPowerState.String(),
			"new_power_state": string(cmd),
		})

		res, err := r.providerData.client.PostV2MetalIdPowerCommandWithResponse(ctx, state.ID.ValueInt64(),
			&client.PostV2MetalIdPowerCommandParams{
				Command: PtrTo(cmd),
			},
		)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update v2 metal power status, got error: %s", err))
			return
		}

		if res.StatusCode() != http.StatusOK {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update v2 metal power status, got error: %s", string(res.Body)))
			return
		}

		tflog.Debug(ctx, "power state updated")
	}

end:
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *MetalResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MetalResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// /v2/Metal doesn't support deletion currently
	delReq, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("https://api.tsw.io/v1/Metal/%d?projectId=480", data.ID.ValueInt64()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("failed to create delete v2 metal request: %s", err),
		)
		return
	}
	delReq.Header.Add("Authorization", "Bearer "+r.providerData.apiKey)

	res, err := r.providerData.httpClient.Do(delReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("failed to send delete v2 metal request: %s", err),
		)
		return
	}
	defer res.Body.Close()

	body := bytes.NewBuffer(nil)
	_, _ = io.Copy(body, res.Body)

	if res.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("delete v2 metal request errored: %s", body.String()),
		)
		return
	}
}

func (r *MetalResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.AddError("Not supported",
		"Metal resources don't currently support importing.",
	)
	// resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
