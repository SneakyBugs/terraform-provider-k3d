package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// Ensure K3dProvider satisfies various provider interfaces.
var _ provider.Provider = &K3dProvider{}
var _ provider.ProviderWithMetadata = &K3dProvider{}

// K3dProvider defines the provider implementation.
type K3dProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// K3dProviderModel describes the provider data model.
type K3dProviderModel struct{}

func (p *K3dProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "k3d"
	resp.Version = p.version
}

func (p *K3dProvider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "This provider manages development Kubernetes clusters in Docker with k3d. " +
			"Managing k3d clusters in Terraform allows you to provision development clusters " +
			"and deploy additional software (such as a database for your app) in a single action.\n" +
			"\n" +
			"The idea behind this provider is to automate everything before `tilt up` with Terraform.\n" +
			"\n" +
			"## Quick Start\n" +
			"\n" +
			"Make sure to install k3d, [see the installation guide](https://developer.hashicorp.com/terraform/tutorials/aws-get-started/install-cli#install-terraform).\n" +
			"\n" +
			"You may need to run Terraform with `sudo` because k3d uses Docker.\n" +
			"\n" +
			"The example below creates a cluster and deploys a Postgres instance on it. " +
			"It can be adapted to deploy any services your app needs for development with minimal effort.",
		Attributes: map[string]tfsdk.Attribute{},
	}, nil
}

func (p *K3dProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data K3dProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// if data.Endpoint.IsNull() { /* ... */ }

	// Example client configuration for data sources and resources
	client := http.DefaultClient
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *K3dProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewClusterResource,
	}
}

func (p *K3dProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &K3dProvider{
			version: version,
		}
	}
}
