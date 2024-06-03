package provider

import (
	"fmt"
	"strings"

	dtrack "github.com/DependencyTrack/client-go"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// configPropertiesDataSource is the datasource implementation.
type configPropertiesDataSource struct {
	client *dtrack.Client
}

// configPropertyResource is the oidc group resource implementation.
type configPropertyResource struct {
	client *dtrack.Client
}

// configPropertiesDataSourceModel maps the data source schema data.
type configPropertiesDataSourceModel struct {
	ConfigProperties []configPropertyModel `tfsdk:"config_properties"`
}

// configPropertyModel maps configuration property schema data.
type configPropertyModel struct {
	ID    types.String `tfsdk:"id"`
	Group types.String `tfsdk:"group"`
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
	Type  types.String `tfsdk:"type"`
}

func configPropertyID(cp dtrack.ConfigProperty) string {
	return fmt.Sprintf("%s_%s", cp.GroupName, strings.ReplaceAll(cp.PropertyName, ".", "-"))
}
