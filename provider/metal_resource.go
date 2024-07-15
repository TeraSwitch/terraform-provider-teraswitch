package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/TeraSwitch/terraform-provider/client"
	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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
	Id             types.Int64           `tfsdk:"id"`
	ProjectID      types.Int64           `tfsdk:"project_id"`
	RegionID       types.String          `tfsdk:"region_id"`
	DisplayName    types.String          `tfsdk:"display_name"`
	TierID         types.String          `tfsdk:"tier_id"`
	ImageID        types.String          `tfsdk:"image_id"`
	SSHKeyIDs      types.List            `tfsdk:"ssh_key_ids"`
	Password       types.String          `tfsdk:"password"`
	UserData       types.String          `tfsdk:"user_data"`
	Tags           types.List            `tfsdk:"tags"`
	MemoryGB       types.Int64           `tfsdk:"memory_gb"`
	Disks          types.Map             `tfsdk:"disks"`
	Partitions     []MetalPartitionModel `tfsdk:"partitions"`
	RaidArrays     []MetalRaidArrayModel `tfsdk:"raid_arrays"`
	IPXEURL        types.String          `tfsdk:"ipxe_url"`
	TemplateID     types.Int64           `tfsdk:"template_id"`
	Quantity       types.Int64           `tfsdk:"quantity"`
	ReservePricing types.Bool            `tfsdk:"reserve_pricing"`
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
			},
			"region_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the region that the metal will be created in.",
				Required:            true,
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: "The display name of the network. This is optional.",
				Optional:            true,
			},
			"tier_id": schema.StringAttribute{
				MarkdownDescription: "The service tier to be created. For metal, this is typically the server config. For example: 7302p-64g would create a Epyc 7302P system with 64G of ram. Tier availability can be retrieved using the regions endpoints.",
				Required:            true,
			},
			"image_id": schema.StringAttribute{
				MarkdownDescription: "The image to use when creating this service. Available images can be retrieved via the images endpoint.",
				Optional:            true,
			},
			"ssh_key_ids": schema.ListAttribute{
				MarkdownDescription: "The SSH key ids to be added to the service. These keys will be added to the authorized_keys file for the root user.",
				Optional:            true,
				ElementType:         types.Int64Type,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The password to be set for the root user. If not provided, a random password will be generated.",
				Optional:            true,
			},
			"user_data": schema.StringAttribute{
				MarkdownDescription: "Additional user data.",
				Optional:            true,
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "Tags to be added to the metal service.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"memory_gb": schema.Int64Attribute{
				MarkdownDescription: "The amount of memory in GB to be allocated to the metal service.",
				Optional:            true,
			},
			"disks": schema.MapAttribute{
				MarkdownDescription: "Dictionary of disk names and sizes in GB. If not specified, the default configuration for the metal tier will be used. The key is the disk name and the value is the size in GB.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"partitions": schema.ListNestedAttribute{
				MarkdownDescription: "Partitions to be created on the metal service. Not specifying this will result in a single root partition being created.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "",
						},
						"device": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "",
						},
						"size_bytes": schema.Int64Attribute{
							Optional: true,
						},
						"file_system": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "",
						},
						"mount_point": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "",
						},
					},
				},
			},
			"raid_arrays": schema.ListNestedAttribute{
				MarkdownDescription: "Raid arrays to be created on the metal service. Can reference physical device names or partitions from mediums of the same class.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "",
						},
						"type": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "",
						},
						"members": schema.ListAttribute{
							Required:            true,
							MarkdownDescription: "",
							ElementType:         types.StringType,
						},
						"file_system": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "",
						},
						"mount_point": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "",
						},
						"size_bytes": schema.Int64Attribute{
							Optional: true,
						},
					},
				},
			},
			"ipxe_url": schema.StringAttribute{
				MarkdownDescription: "The URL to the script to use when enabling iPXE boot.",
				Optional:            true,
			},
			"template_id": schema.Int64Attribute{
				MarkdownDescription: "Template can be specified instead of image, partitions, sshKeyId, and userData.",
				Optional:            true,
			},
			"quantity": schema.Int64Attribute{
				MarkdownDescription: "The number of services to be created. By default, one will be created.",
				Optional:            true,
			},
			"reserve_pricing": schema.BoolAttribute{
				MarkdownDescription: "Denotes if the metal service is being reserved for a whole year. If so, it gets the discounted rate",
				Optional:            true,
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
		projID = &r.providerData.projectID
	}

	body := client.CreateMetalRequest{
		ProjectId:      projID,
		RegionId:       data.RegionID.ValueString(),
		DisplayName:    data.DisplayName.ValueStringPointer(),
		TierId:         data.TierID.ValueString(),
		ImageId:        data.ImageID.ValueStringPointer(),
		Password:       data.Password.ValueStringPointer(),
		UserData:       data.UserData.ValueStringPointer(),
		MemoryGb:       i64PtrToi32Ptr(data.MemoryGB.ValueInt64Pointer()),
		IpxeUrl:        data.IPXEURL.ValueStringPointer(),
		TemplateId:     data.TemplateID.ValueInt64Pointer(),
		Quantity:       i64PtrToi32Ptr(data.Quantity.ValueInt64Pointer()),
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
				FileSystem: PtrTo(client.FileSystem(3)), // TODO: fix
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
				FileSystem: PtrTo(client.FileSystem(3)), // TODO: fix
				MountPoint: dRaid.MountPoint.ValueStringPointer(),
				Name:       dRaid.Name.ValueStringPointer(),
				SizeBytes:  dRaid.SizeBytes.ValueInt64Pointer(),
				Type:       PtrTo(client.RaidType(3)),
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

	spew.Dump(res.JSON200)

	fmt.Println(string(res.Body))
	if res.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to create get v2 network, got error: %s", string(res.Body)),
		)
		return
	}

	resBody := res.JSON200.Result

	// For the purposes of this example code, hardcoding a response value to
	// save into the Terraform state.
	data.Id = types.Int64Value(*resBody.Id)

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MetalResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MetalResourceModel

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

func (r *MetalResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MetalResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MetalResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MetalResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete example, got error: %s", err))
	//     return
	// }
}

func (r *MetalResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}