// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure DeepmergeProvider satisfies various provider interfaces.
var _ provider.Provider = &DeepmergeProvider{}
var _ provider.ProviderWithFunctions = &DeepmergeProvider{}

// DeepmergeProvider defines the provider implementation.
type DeepmergeProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// DeepmergeProviderModel describes the provider data model.
type DeepmergeProviderModel struct {
}

func (p *DeepmergeProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "deepmerge"
	resp.Version = p.version
}

func (p *DeepmergeProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema.Description = `Deepmerge functions`
}

func (p *DeepmergeProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
}

func (p *DeepmergeProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *DeepmergeProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *DeepmergeProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewMergoFunction,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &DeepmergeProvider{
			version: version,
		}
	}
}
