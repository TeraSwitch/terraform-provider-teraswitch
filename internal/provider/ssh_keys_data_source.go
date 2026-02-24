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
var _ datasource.DataSource = &SshKeysDataSource{}

func NewSshKeysDataSource() datasource.DataSource {
	return &SshKeysDataSource{}
}

// SshKeysDataSource defines the data source implementation.
type SshKeysDataSource struct {
	providerData *ProviderData
}

// SshKeyModel describes a single SSH key.
type SshKeyModel struct {
	ID          types.Int64  `tfsdk:"id"`
	DisplayName types.String `tfsdk:"display_name"`
	Key         types.String `tfsdk:"key"`
	ProjectID   types.Int64  `tfsdk:"project_id"`
	Created     types.String `tfsdk:"created"`
}

// SshKeysDataSourceModel describes the data source data model.
type SshKeysDataSourceModel struct {
	SshKeys []SshKeyModel `tfsdk:"ssh_keys"`
}

func (d *SshKeysDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_keys"
}

func (d *SshKeysDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "SSH Keys data source allows you to retrieve all SSH keys in your project.",

		Attributes: map[string]schema.Attribute{
			"ssh_keys": schema.ListNestedAttribute{
				MarkdownDescription: "List of SSH keys in the project.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "The ID of the SSH key.",
							Computed:            true,
						},
						"display_name": schema.StringAttribute{
							MarkdownDescription: "The display name of the SSH key.",
							Computed:            true,
						},
						"key": schema.StringAttribute{
							MarkdownDescription: "The public SSH key.",
							Computed:            true,
						},
						"project_id": schema.Int64Attribute{
							MarkdownDescription: "The ID of the project that the SSH key belongs to.",
							Computed:            true,
						},
						"created": schema.StringAttribute{
							MarkdownDescription: "The date when the SSH key was created.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *SshKeysDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SshKeysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SshKeysDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := d.providerData.client.GetV2SshKeyWithResponse(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read SSH keys, got error: %s", err))
		return
	}

	if res.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read SSH keys, got error: %s", string(res.Body)),
		)
		return
	}

	if res.JSON200 == nil || res.JSON200.Result == nil {
		data.SshKeys = []SshKeyModel{}
	} else {
		sshKeys := make([]SshKeyModel, 0, len(*res.JSON200.Result))
		for _, key := range *res.JSON200.Result {
			sshKey := SshKeyModel{}

			if key.Id != nil {
				sshKey.ID = types.Int64Value(*key.Id)
			} else {
				sshKey.ID = types.Int64Null()
			}

			if key.DisplayName != nil {
				sshKey.DisplayName = types.StringValue(*key.DisplayName)
			} else {
				sshKey.DisplayName = types.StringNull()
			}

			if key.Key != nil {
				sshKey.Key = types.StringValue(*key.Key)
			} else {
				sshKey.Key = types.StringNull()
			}

			if key.ProjectId != nil {
				sshKey.ProjectID = types.Int64Value(*key.ProjectId)
			} else {
				sshKey.ProjectID = types.Int64Null()
			}

			if key.Created != nil {
				sshKey.Created = types.StringValue(*key.Created)
			} else {
				sshKey.Created = types.StringNull()
			}

			sshKeys = append(sshKeys, sshKey)
		}
		data.SshKeys = sshKeys
	}

	tflog.Trace(ctx, "read SSH keys data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
