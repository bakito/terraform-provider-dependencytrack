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
	_ datasource.DataSource              = &configPropertyDataSource{}
	_ datasource.DataSourceWithConfigure = &configPropertyDataSource{}
)

func NewConfigPropertyDataSource() datasource.DataSource {
	return &configPropertyDataSource{}
}

func (d *configPropertyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *configPropertyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_property"
}

// Schema defines the schema for the data source.
func (d *configPropertyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"config_property": schema.ListNestedAttribute{
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

func (d *configPropertyDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state configPropertyDataSourceModel

	properties, err := dtrack.FetchAll(func(po dtrack.PageOptions) (dtrack.Page[dtrack.ConfigProperty], error) {
		return d.client.ConfigProperty.GetAllConfigProperties(ctx, po)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read DependencyTrack OIDC Groups",
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

		state.configProperty = append(state.configProperty, configPropertyState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
