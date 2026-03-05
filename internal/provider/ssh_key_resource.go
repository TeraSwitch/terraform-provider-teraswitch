package provider

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/TeraSwitch/terraform-provider/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &SshKeyResource{}
var _ resource.ResourceWithImportState = &SshKeyResource{}

func NewSshKeyResource() resource.Resource {
	return &SshKeyResource{}
}

// SshKeyResource defines the resource implementation.
type SshKeyResource struct {
	providerData *ProviderData
}

// SshKeyResourceModel describes the resource data model.
type SshKeyResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	DisplayName types.String `tfsdk:"display_name"`
	Key         types.String `tfsdk:"key"`
	ProjectID   types.Int64  `tfsdk:"project_id"`
	Created     types.String `tfsdk:"created"`
}

func (r *SshKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_key"
}

func (r *SshKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "SSH Key resource allows you to create and manage SSH keys in your project.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The ID of the SSH key.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: "The display name of the SSH key.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "The public SSH key (e.g., ssh-rsa AAAA... or ssh-ed25519 AAAA...).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the project that the SSH key belongs to.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"created": schema.StringAttribute{
				MarkdownDescription: "The date when the SSH key was created.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *SshKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.providerData = providerData
}

func (r *SshKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SshKeyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	displayName := data.DisplayName.ValueString()
	key := data.Key.ValueString()

	createReq := client.SshKey{
		DisplayName: &displayName,
		Key:         &key,
	}

	res, err := r.providerData.client.PostV2SshKeyWithResponse(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create SSH key, got error: %s", err))
		return
	}

	if res.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to create SSH key, got status %d: %s", res.StatusCode(), string(res.Body)),
		)
		return
	}

	if res.JSON200 == nil || res.JSON200.Result == nil {
		resp.Diagnostics.AddError("Client Error", "SSH key creation returned no result")
		return
	}

	sshKey := res.JSON200.Result

	if sshKey.Id != nil {
		data.ID = types.Int64Value(*sshKey.Id)
	}
	if sshKey.DisplayName != nil {
		data.DisplayName = types.StringValue(*sshKey.DisplayName)
	}
	if sshKey.Key != nil {
		data.Key = types.StringValue(*sshKey.Key)
	}
	if sshKey.ProjectId != nil {
		data.ProjectID = types.Int64Value(*sshKey.ProjectId)
	}
	if sshKey.Created != nil {
		data.Created = types.StringValue(*sshKey.Created)
	}

	tflog.Trace(ctx, "created SSH key resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SshKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SshKeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.providerData.client.GetV2SshKeyIdWithResponse(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read SSH key, got error: %s", err))
		return
	}

	if res.StatusCode() == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if res.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read SSH key, got status %d: %s", res.StatusCode(), string(res.Body)),
		)
		return
	}

	if res.JSON200 == nil || res.JSON200.Result == nil {
		resp.Diagnostics.AddError("Client Error", "SSH key not found")
		return
	}

	sshKey := res.JSON200.Result

	if sshKey.Id != nil {
		data.ID = types.Int64Value(*sshKey.Id)
	}
	if sshKey.DisplayName != nil {
		data.DisplayName = types.StringValue(*sshKey.DisplayName)
	}
	if sshKey.Key != nil {
		data.Key = types.StringValue(*sshKey.Key)
	}
	if sshKey.ProjectId != nil {
		data.ProjectID = types.Int64Value(*sshKey.ProjectId)
	}
	if sshKey.Created != nil {
		data.Created = types.StringValue(*sshKey.Created)
	}

	tflog.Trace(ctx, "read SSH key resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SshKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// SSH keys are immutable - any changes require replacement
	// This is enforced by RequiresReplace plan modifiers on all user-settable attributes
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"SSH keys cannot be updated. Any changes require creating a new SSH key.",
	)
}

func (r *SshKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SshKeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.providerData.client.DeleteV2SshKeyIdWithResponse(ctx, data.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete SSH key, got error: %s", err))
		return
	}

	if res.StatusCode() == http.StatusNotFound {
		// SSH key already deleted, nothing to do
		return
	}

	if res.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to delete SSH key, got status %d: %s", res.StatusCode(), string(res.Body)),
		)
		return
	}

	tflog.Trace(ctx, "deleted SSH key resource")
}

func (r *SshKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid ID",
			fmt.Sprintf("Unable to parse ID %q as integer: %s", req.ID, err),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
