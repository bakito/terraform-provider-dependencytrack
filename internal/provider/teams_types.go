package provider

import (
	dtrack "github.com/DependencyTrack/client-go"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// teamsDataSource is the datasource implementation.
type teamsDataSource struct {
	client *dtrack.Client
}

// teamDataSource is the datasource implementation.
type teamDataSource struct {
	client *dtrack.Client
}

// teamResource is the resource implementation.
type teamResource struct {
	client *dtrack.Client
}

// teamDataSourceModel maps the data source schema data.
type teamsDataSourceModel struct {
	Teams []teamModel `tfsdk:"teams"`
}

// teamModel maps oidc group schema data.
type teamModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}
