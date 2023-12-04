package provider

import (
	dtrack "github.com/DependencyTrack/client-go"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// oidcGroupMappingResource is the oidc group implementation.
type oidcGroupMappingResource struct {
	client *dtrack.Client
}

// oidcGroupMappingDataSourceModel maps the data source schema data.
type oidcGroupMappingDataSourceModel struct {
	Repositories []oidcGroupMappingModel `tfsdk:"oicd_groups"`
}

// oidcGroupMappingModel maps oidc group schema data.
type oidcGroupMappingModel struct {
	ID          types.String `tfsdk:"id"`
	LastUpdated types.String `tfsdk:"last_updated"`
}
