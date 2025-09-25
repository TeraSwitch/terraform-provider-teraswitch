package provider

import (
	"context"
	"net/http"
	"os"
	"strconv"

	"github.com/TeraSwitch/terraform-provider/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &TeraswitchProvider{}
var _ provider.ProviderWithFunctions = &TeraswitchProvider{}

// TeraswitchProvider defines the provider implementation.
type TeraswitchProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type ProviderData struct {
	httpClient *http.Client
	projectID  int64
	apiKey     string
	client     *client.ClientWithResponses
}

// TeraswitchProviderModel describes the provider data model.
type TeraswitchProviderModel struct {
	APIKey    types.String `tfsdk:"api_key"`
	ProjectID types.Int64  `tfsdk:"project_id"`
}

func (p *TeraswitchProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "teraswitch"
	resp.Version = p.version
}

func (p *TeraswitchProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "API key generated from beta.tsw.io",
				Optional:            true,
				Sensitive:           true,
			},
			"project_id": schema.Int64Attribute{
				MarkdownDescription: "Project ID from beta.tsw.io. Used as the default if a project id isn't supplied on a resource.",
				Optional:            true,
			},
		},
	}
}

func (p *TeraswitchProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data TeraswitchProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.APIKey.IsNull() {
		apiKeyEnv, ok := os.LookupEnv("TERASWITCH_API_KEY")
		if !ok {
			resp.Diagnostics.AddError(
				"api_key is required",
				"Expected api_key to be set on the provider configuration.",
			)
			return
		}
		data.APIKey = types.StringValue(apiKeyEnv)
	}

	if data.ProjectID.IsNull() {
		projectIDEnv, ok := os.LookupEnv("TERASWITCH_PROJECT_ID")
		if ok {
			projID, err := strconv.ParseInt(projectIDEnv, 10, 64)
			if err != nil {
				resp.Diagnostics.AddError(
					"project_id invalid",
					"Expected project_id to be a valid int64: "+err.Error(),
				)
				return
			}
			data.ProjectID = types.Int64Value(projID)
		}
	}

	httpClient := &http.Client{}

	reqClient, err := client.NewClientWithResponses("https://api.tsw.io",
		client.WithHTTPClient(httpClient),
		client.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Add("Authorization", "Bearer "+data.APIKey.ValueString())
			return nil
		}),
	)
	if err != nil {
		resp.Diagnostics.AddError("failed to create client", err.Error())
		return
	}

	pd := &ProviderData{
		client:     reqClient,
		httpClient: httpClient,
		projectID:  data.ProjectID.ValueInt64(),
		apiKey:     data.APIKey.ValueString(),
	}

	// Example client configuration for data sources and resources
	resp.DataSourceData = pd
	resp.ResourceData = pd
}

func (p *TeraswitchProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewNetworkResource,
		NewVolumeResource,
		NewMetalResource,
		NewCloudComputeResource,
	}
}

func (p *TeraswitchProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewMetalDataSource,
	}
}

func (p *TeraswitchProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &TeraswitchProvider{
			version: version,
		}
	}
}
