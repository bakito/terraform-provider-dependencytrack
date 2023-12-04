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
	_ datasource.DataSource              = &oidcGroupDataSource{}
	_ datasource.DataSourceWithConfigure = &oidcGroupDataSource{}
)

func NewOidcGroupataSource() datasource.DataSource {
	return &oidcGroupDataSource{}
}

func (d *oidcGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *oidcGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oidc_groups"
}

// Schema defines the schema for the data source.
func (d *oidcGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"repositories": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *oidcGroupDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state oidcGroupDataSourceModel

	groups, err := dtrack.FetchAll(func(po dtrack.PageOptions) (dtrack.Page[dtrack.OIDCGroup], error) {
		return d.client.OIDC.GetAllGroups(ctx, po)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read DependencyTrack Repositories",
			err.Error(),
		)
		return
	}

	// Map response body to model
	for _, group := range groups {
		oidcgroupState := oidcGroupModel{
			ID:   types.StringValue(group.UUID.String()),
			Name: types.StringValue(group.Name),
		}

		state.Repositories = append(state.Repositories, oidcgroupState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
