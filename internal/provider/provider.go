package provider

import (
	"context"

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
				MarkdownDescription: "The OVH API endpoint to target (eg: \"ovh-eu\").",
				Required:            true,
			},
			"application_key": schema.StringAttribute{
				MarkdownDescription: "The OVH API Application Key.",
				Required:            true,
				Sensitive:           true,
			},
			"application_secret": schema.StringAttribute{
				MarkdownDescription: "The OVH API Application Secret.",
				Required:            true,
				Sensitive:           true,
			},
			"consumer_key": schema.StringAttribute{
				MarkdownDescription: "The OVH API Consumer key.",
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *CDCOvhNSProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data CDCOvhNSProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	ovhData := api.OVHCredentials{
		Endpoint:          data.Endpoint.ValueString(),
		ApplicationKey:    data.ApplicationKey.ValueString(),
		ApplicationSecret: data.ApplicationSecret.ValueString(),
		ConsumerKey:       data.ConsumerKey.ValueString(),
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
