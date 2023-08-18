package provider

import (
	"context"
	"os"

	"github.com/capybaradevcloud/terraform-provider-cdcovhns/internal/api"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &CDCOvhNSProvider{}

type CDCOvhNSProvider struct {
	version string
}

type CDCOvhNSProviderModel struct {
	Endpoint          types.String `tfsdk:"endpoint"`
	ApplicationKey    types.String `tfsdk:"application_key"`
	ApplicationSecret types.String `tfsdk:"application_secret"`
	ConsumerKey       types.String `tfsdk:"consumer_key"`
}

func (p *CDCOvhNSProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "cdcovhns"
	resp.Version = p.version
}

func (p *CDCOvhNSProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "The OVH API endpoint to target (eg: \"ovh-eu\"). " +
					"Can also be configured using the `OVH_ENDPOINT` environment variable.",
				Optional: true,
			},
			"application_key": schema.StringAttribute{
				MarkdownDescription: "The OVH API Application Key. " +
					"Can also be configured using the `OVH_APPLICATION_KEY` environment variable.",
				Sensitive: true,
				Optional:  true,
			},
			"application_secret": schema.StringAttribute{
				MarkdownDescription: "The OVH API Application Secret. " +
					"Can also be configured using the `OVH_SECRET_KEY` environment variable.",
				Sensitive: true,
				Optional:  true,
			},
			"consumer_key": schema.StringAttribute{
				MarkdownDescription: "The OVH API Consumer key. " +
					"Can also be configured using the `OVH_CONSUMER_KEY` environment variable.",
				Sensitive: true,
				Optional:  true,
			},
		},
	}
}

func (p *CDCOvhNSProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data CDCOvhNSProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	endpoint := os.Getenv("OVH_ENDPOINT")
	if data.Endpoint.ValueString() != "" {
		endpoint = data.Endpoint.ValueString()
	}

	applicationKey := os.Getenv("OVH_APPLICATION_KEY")
	if data.ApplicationKey.ValueString() != "" {
		applicationKey = data.ApplicationKey.ValueString()
	}

	applicationSecret := os.Getenv("OVH_APPLICATION_SECRET")
	if data.ApplicationSecret.ValueString() != "" {
		applicationSecret = data.ApplicationSecret.ValueString()
	}

	consumerKey := os.Getenv("OVH_CONSUMER_KEY")
	if data.ConsumerKey.ValueString() != "" {
		consumerKey = data.ConsumerKey.ValueString()
	}

	ovhData := api.OVHCredentials{
		Endpoint:          endpoint,
		ApplicationKey:    applicationKey,
		ApplicationSecret: applicationSecret,
		ConsumerKey:       consumerKey,
	}

	client, err := api.GetClient(ovhData, ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create OVH API Client",
			"OVH client seems to be misconfigured: "+err.Error(),
		)
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *CDCOvhNSProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCDCOvhNSResource,
	}
}

func (p *CDCOvhNSProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &CDCOvhNSProvider{
			version: version,
		}
	}
}
