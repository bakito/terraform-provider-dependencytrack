package provider

import (
	"context"
	"fmt"

	dtrack "github.com/DependencyTrack/client-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &configPropertiesDataSource{}
	_ datasource.DataSourceWithConfigure = &configPropertiesDataSource{}
)

func NewConfigPropertiesDataSource() datasource.DataSource {
	return &configPropertiesDataSource{}
}

func (d *configPropertiesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*dtrack.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *dtrack.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *configPropertiesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_properties"
}

// Schema defines the schema for the data source.
func (d *configPropertiesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"config_properties": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"group": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"type": schema.StringAttribute{
							Computed: true,
						},
						"value": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *configPropertiesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state configPropertiesDataSourceModel

	properties, err := dtrack.FetchAll(func(po dtrack.PageOptions) (dtrack.Page[dtrack.ConfigProperty], error) {
		return d.client.ConfigProperty.GetAllConfigProperties(ctx, po)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read DependencyTrack Config Properties Groups",
			err.Error(),
		)
		return
	}

	// Map response body to model
	for _, property := range properties {
		configPropertyState := configPropertyModel{
			ID:    types.StringValue(configPropertyID(property)),
			Group: types.StringValue(property.GroupName),
			Name:  types.StringValue(property.PropertyName),
			Type:  types.StringValue(property.PropertyType),
			Value: types.StringValue(property.PropertyValue),
		}

		state.ConfigProperties = append(state.ConfigProperties, configPropertyState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
