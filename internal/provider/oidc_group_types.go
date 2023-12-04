package provider

import (
	dtrack "github.com/DependencyTrack/client-go"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// oidcGroupDataSource is the datasource implementation.
type oidcGroupDataSource struct {
	client *dtrack.Client
}

// oidcGroupResource is the oidc group implementation.
type oidcGroupResource struct {
	client *dtrack.Client
}

// oidcGroupDataSourceModel maps the data source schema data.
type oidcGroupDataSourceModel struct {
	OidcGroups []oidcGroupModel `tfsdk:"oidc_groups"`
}

// oidcGroupModel maps oidc group schema data.
type oidcGroupModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	//LastUpdated types.String `tfsdk:"last_updated"`
}
