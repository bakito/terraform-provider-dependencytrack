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
	_ datasource.DataSource              = &repositoryDataSource{}
	_ datasource.DataSourceWithConfigure = &repositoryDataSource{}
)

func NewRepositoryDataSource() datasource.DataSource {
	return &repositoryDataSource{}
}

func (d *repositoryDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *repositoryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repositories"
}

// Schema defines the schema for the data source.
func (d *repositoryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"repositories": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"identifier": schema.StringAttribute{
							Computed: true,
						},
						"type": schema.StringAttribute{
							Computed: true,
						},
						"url": schema.StringAttribute{
							Computed: true,
						},
						"resolution_order": schema.Int64Attribute{
							Computed: true,
						},
						"enabled": schema.BoolAttribute{
							Computed: true,
						},
						"internal": schema.BoolAttribute{
							Computed: true,
						},
						"username": schema.StringAttribute{
							Computed: true,
						},
						"password": schema.StringAttribute{
							Computed:  true,
							Sensitive: true,
						},
						"last_updated": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *repositoryDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state repositoryDataSourceModel

	repos, err := dtrack.FetchAll(func(po dtrack.PageOptions) (dtrack.Page[dtrack.Repository], error) {
		return d.client.Repository.GetAll(ctx, po)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read DependencyTrack Repositories",
			err.Error(),
		)
		return
	}

	// Map response body to model
	for _, repo := range repos {
		repositoryState := repositoryModel{
			ID:              types.StringValue(repo.UUID.String()),
			Type:            types.StringValue(string(repo.Type)),
			Identifier:      types.StringValue(repo.Identifier),
			Url:             types.StringValue(repo.Url),
			ResolutionOrder: types.Int64Value(int64(repo.ResolutionOrder)),
			Enabled:         types.BoolValue(repo.Enabled),
			Internal:        types.BoolValue(repo.Internal),
			Username:        types.StringValue(repo.Username),
			Password:        types.StringValue(repo.Password),
		}

		state.Repositories = append(state.Repositories, repositoryState)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
