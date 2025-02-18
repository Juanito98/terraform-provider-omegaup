// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"terraform-provider-omegaup/internal/apiclient"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure OmegaUpProvider satisfies various provider interfaces.
var _ provider.Provider = &OmegaUpProvider{}
var _ provider.ProviderWithFunctions = &OmegaUpProvider{}
var _ provider.ProviderWithEphemeralResources = &OmegaUpProvider{}

// OmegaUpProvider defines the provider implementation.
type OmegaUpProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// OmegaUpProviderModel describes the provider data model.
type OmegaUpProviderModel struct {
	BaseURL  types.String `tfsdk:"base_url"`
	Username types.String `tfsdk:"username"`
	ApiToken types.String `tfsdk:"api_token"`
}

func (p *OmegaUpProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "omegaup"
	resp.Version = p.version
}

func (p *OmegaUpProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				MarkdownDescription: "Base URL for OmegaUp API.",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "OmegaUp username",
				Optional:            true,
			},
			"api_token": schema.StringAttribute{
				MarkdownDescription: "OmegaUp API token",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *OmegaUpProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data OmegaUpProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// if data.Endpoint.IsNull() { /* ... */ }

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.
	if data.BaseURL.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("base_url"),
			"Unknown OmegaUp base url",
			"The provider cannot create the OmegaUp API client as there is an unknown configuration value for the OmegaUp base url.",
		)
	}
	if data.ApiToken.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token"),
			"Unknown OmegaUp api token",
			"The provider cannot create the OmegaUp API client as there is an unknown configuration value for the OmegaUp api token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the OMEGAUP_API_TOKEN environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.
	baseURL := ""
	apiToken := os.Getenv("OMEGAUP_API_TOKEN")

	if !data.BaseURL.IsNull() {
		baseURL = data.BaseURL.ValueString()
	}
	if !data.ApiToken.IsNull() {
		apiToken = data.ApiToken.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.
	if apiToken == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("user"),
			"Missing OmegaUp api token",
			"The provider cannot create the OmegaUp API client as there is a missing or empty value for the OmegaUp api token. "+
				"Set the host value in the configuration or use the OMEGAUP_API_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Example client configuration for data sources and resources
	client := apiclient.NewClient(apiToken, baseURL)
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *OmegaUpProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewGroupResource,
		NewGroupMemberResource,
		NewIdentityResource,
		NewIdentitiesResource,
	}
}

func (p *OmegaUpProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{}
}

func (p *OmegaUpProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *OmegaUpProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &OmegaUpProvider{
			version: version,
		}
	}
}
