package provider

import (
	dtrack "github.com/DependencyTrack/client-go"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// repositoryDataSource is the datasource implementation.
type repositoryDataSource struct {
	client *dtrack.Client
}

// repositoryResource is the resource implementation.
type repositoryResource struct {
	client *dtrack.Client
}

// repositoryDataSourceModel maps the data source schema data.
type repositoryDataSourceModel struct {
	Repositories []repositoryModel `tfsdk:"repositories"`
}

// repositoryModel maps repository schema data.
type repositoryModel struct {
	ID                     types.String `tfsdk:"id"`
	Type                   types.String `tfsdk:"type"`
	Identifier             types.String `tfsdk:"identifier"`
	Url                    types.String `tfsdk:"url"`
	ResolutionOrder        types.Int64  `tfsdk:"resolution_order"`
	Enabled                types.Bool   `tfsdk:"enabled"`
	Internal               types.Bool   `tfsdk:"internal"`
	AuthenticationRequired types.Bool   `tfsdk:"authentication_required"`
	Username               types.String `tfsdk:"username"`
	Password               types.String `tfsdk:"password"`
	LastUpdated            types.String `tfsdk:"last_updated"`
}
