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
var _ datasource.DataSource = &TagsDataSource{}

func NewTagsDataSource() datasource.DataSource {
	return &TagsDataSource{}
}

// TagsDataSource defines the data source implementation.
type TagsDataSource struct {
	providerData *ProviderData
}

// TagsDataSourceModel describes the data source data model.
type TagsDataSourceModel struct {
	Tags []types.String `tfsdk:"tags"`
}

func (d *TagsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tags"
}

func (d *TagsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Tags data source allows you to retrieve all tags in use across your project.",

		Attributes: map[string]schema.Attribute{
			"tags": schema.ListAttribute{
				MarkdownDescription: "List of tags in use across the project.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (d *TagsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TagsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TagsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := d.providerData.client.GetV2TagsWithResponse(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read tags, got error: %s", err))
		return
	}

	if res.StatusCode() != http.StatusOK {
		resp.Diagnostics.AddError("Client Error",
			fmt.Sprintf("Unable to read tags, got error: %s", string(res.Body)),
		)
		return
	}

	if res.JSON200 == nil || res.JSON200.Result == nil {
		data.Tags = []types.String{}
	} else {
		tags := make([]types.String, 0, len(*res.JSON200.Result))
		for _, tag := range *res.JSON200.Result {
			tags = append(tags, types.StringValue(tag))
		}
		data.Tags = tags
	}

	tflog.Trace(ctx, "read tags data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
