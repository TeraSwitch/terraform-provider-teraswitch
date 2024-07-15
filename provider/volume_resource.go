package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/TeraSwitch/terraform-provider/client"
	"github.com/davecgh/go-spew/spew"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
// var _ datasource.DataSource = &VolumeResource{}
var _ resource.Resource = &VolumeResource{}
var _ resource.ResourceWithImportState = &VolumeResource{}

func NewVolumeResource() resource.Resource {
	return &VolumeResource{}
}

// VolumeResource defines the resource implementation.
type VolumeResource struct {
	providerData *ProviderData
}

// VolumeResourceModel describes the resource data model.
type VolumeResourceModel struct {
	Id          types.String `tfsdk:"id"`
	RegionID    types.String `tfsdk:"region_id"`
	DisplayName types.String `tfsdk:"display_name"`
	VolumeType  types.String `tfsdk:"volume_type"`
	Size        types.Int64  `tfsdk:"size"`
	Description types.String `tfsdk:"description"`
	ImageName   types.String `tfsdk:"image_name"`
	Status      types.String `tfsdk:"status"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

func (r *VolumeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume"
}

func (r *VolumeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Volume",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Id of the volume",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"region_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the region that the volume will be created in.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: "The display name of the volume. This is optional.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"volume_type": schema.StringAttribute{
				MarkdownDescription: "The underlying storage type of the volume, HDD or NVME.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"size": schema.Int64Attribute{
				MarkdownDescription: "The size of the volume in gibibytes (GiB).",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the volume.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"image_name": schema.StringAttribute{
				MarkdownDescription: "The name of the image to create the volume from.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the volume.",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "The time that the resource was last updated.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "The time that the resource was created.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *VolumeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

type CreateVolumeResponseApiResponse struct {
	// Message Provides additional detail about the response if one is required
	Message *string `json:"message"`

	// Result Response to the request to create a storage volume
	Result *CreateVolumeResponse `json:"result,omitempty"`

	// Success True if the request succeeded, false otherwise
	Success *bool `json:"success,omitempty"`
}

type CreateVolumeResponse struct {
	// CreatedAt The time the volume was created
	CreatedAt *string `json:"createdAt,omitempty"`

	// Description The description of the volume
	Description *string `json:"description"`

	// DisplayName The name of the volume
	DisplayName *string `json:"displayName"`

	// Region The region in which the volume was created
	Region *string `json:"region"`

	// Size The size of the volume
	Size *string `json:"size"`

	// Status The current status of the volume
	Status *string `json:"status"`

	// UpdatedAt The time the volume was last updated
	UpdatedAt *string `json:"updatedAt,omitempty"`

	// VolumeId The id of the storage volume
	VolumeId *uuid.UUID `json:"volumeId"`

	// VolumeType The type of the volume. This should be "SSD" or "HDD"
	VolumeType *string `json:"volumeType"`
}

func (r *VolumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VolumeResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.providerData.client.PostV2Volume(ctx, &client.PostV2VolumeParams{
		ProjectId: &r.providerData.projectID,
	}, client.CreateVolumeRequest{
		Description: data.Description.ValueStringPointer(),
		DisplayName: data.DisplayName.ValueStringPointer(),
		ImageName:   data.ImageName.ValueStringPointer(),
		RegionId:    data.RegionID.ValueString(),
		Size:        int32(data.Size.ValueInt64()),
		VolumeType:  data.VolumeType.ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create v2 volume, got error: %s", err))
		return
	}
	defer res.Body.Close()

	apiRes := CreateVolumeResponseApiResponse{}
	err = json.NewDecoder(res.Body).Decode(&apiRes)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create v2 volume, got error: %s", err))
		return
	}

	// fmt.Println("status code", res.StatusCode())
	// fmt.Println(string(res.Body))
	if res.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to create get v2 volume, got error: %s", *apiRes.Message),
		)
		return
	}

	// For the purposes of this example code, hardcoding a response value to
	// save into the Terraform state.
	data.Id = types.StringValue(apiRes.Result.VolumeId.String())
	data.UpdatedAt = types.StringValue(*apiRes.Result.UpdatedAt)
	data.CreatedAt = types.StringValue(*apiRes.Result.CreatedAt)
	data.Status = types.StringValue(*apiRes.Result.Status)

	tflog.Trace(ctx, "created v2 network")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VolumeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VolumeResourceModel

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

func (r *VolumeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VolumeResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type ApiResponse struct {
	// Message Provides additional detail about the response if one is required
	Message string `json:"message"`

	// Success True if the request succeeded, false otherwise
	Success bool `json:"success,omitempty"`
}

func (r *VolumeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VolumeResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.providerData.client.DeleteV2Volume(ctx, &client.DeleteV2VolumeParams{
		ProjectId: &r.providerData.projectID,
	}, client.DeleteVolumeRequest{
		RegionId: data.RegionID.ValueString(),
		VolumeId: data.Id.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to delete volume, got error: %s", err),
		)
		return
	}
	defer res.Body.Close()

	apiRes := ApiResponse{}
	err = json.NewDecoder(res.Body).Decode(&apiRes)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete v2 volume, got error: %s", err))
		return
	}

	if res.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to create get v2 volume, got error: %s", apiRes.Message),
		)
		return
	}

	tflog.Trace(ctx, "deleted v2 volume")
}

type VolumeResponseApiResponse struct {
	// Message Provides additional detail about the response if one is required
	Message *string `json:"message"`

	// Metadata For paginated responses, this object contains metadata about the list of items.
	Metadata *client.ListMetadata `json:"metadata,omitempty"`

	// Result Response to the request to create a storage volume
	Result []VolumeResponse `json:"result,omitempty"`

	// Success True if the request succeeded, false otherwise
	Success *bool `json:"success,omitempty"`
}

type VolumeResponse struct {
	// CreatedAt The time the volume was created
	CreatedAt *string `json:"createdAt,omitempty"`

	// Description The description of the volume
	Description *string `json:"description"`

	// DisplayName The name of the volume
	DisplayName *string `json:"displayName"`

	// Region The region in which the volume was created
	Region *string `json:"region"`

	// Size The size of the volume
	Size *string `json:"size"`

	// Status The current status of the volume
	Status *string `json:"status"`

	// UpdatedAt The time the volume was last updated
	UpdatedAt *string `json:"updatedAt,omitempty"`

	// VolumeId The id of the storage volume
	VolumeId uuid.UUID `json:"volumeId"`

	// VolumeType The type of the volume. This should be "SSD" or "HDD"
	VolumeType *string `json:"volumeType"`
}

func (r *VolumeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	time.Sleep(5 * time.Second)
	importID := req.ID
	fmt.Println("import id:", importID)
	// resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	ids := strings.Split(req.ID, "/")
	if len(ids) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Expected import ID to be in format 'region_id/volume_id'.",
		)
		return
	}

	regionID := ids[0]
	volumeID := ids[1]

	res, err := r.providerData.client.GetV2Volume(ctx, &client.GetV2VolumeParams{
		ProjectId: &r.providerData.projectID,
		// RegionId:  &regionID,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"HTTP error",
			err.Error(),
		)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		resp.Diagnostics.AddError(
			"HTTP error",
			err.Error(),
		)
		return
	}

	fmt.Println(string(body))
	apiRes := VolumeResponseApiResponse{}
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&apiRes)
	if err != nil {
		resp.Diagnostics.AddError("HTTP Error", fmt.Sprintf("Unable to get v2 volume, got error: %s", err))
		return
	}

	// fmt.Println("status code", res.StatusCode())
	// fmt.Println(string(res.Body))
	if res.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("HTTP Error",
			fmt.Sprintf("Unable to create get v2 volume, got error: %s", *apiRes.Message),
		)
		return
	}

	spew.Dump(apiRes)
	var vol *VolumeResponse
	for _, _vol := range apiRes.Result {
		if _vol.VolumeId.String() != volumeID {
			continue
		}

		vol = &_vol
	}

	if vol == nil {
		resp.Diagnostics.AddError("Not found",
			fmt.Sprintf("Unable to find volume with region %s and id %s", regionID, volumeID),
		)
		return
	}

	if true {
		return
	}

	volSize, err := strconv.Atoi(*vol.Size)
	if err != nil {
		resp.Diagnostics.AddError("API error",
			fmt.Sprintf("unable to convert volume size to int: %s", err),
		)
		return
	}

	attrs := map[string]attr.Value{
		"region_id":    types.StringPointerValue(vol.Region),
		"id":           types.StringValue(vol.VolumeId.String()),
		"display_name": types.StringPointerValue(vol.DisplayName),
		"size":         types.Int64Value(int64(volSize)),
		"volume_type":  types.StringPointerValue(vol.VolumeType),
		"description":  types.StringPointerValue(vol.Description),
		"status":       types.StringPointerValue(vol.Status),
		"updated_at":   types.StringPointerValue(vol.UpdatedAt),
		"created_at":   types.StringPointerValue(vol.CreatedAt),
	}

	for k, v := range attrs {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(k), v)...)
	}
}
