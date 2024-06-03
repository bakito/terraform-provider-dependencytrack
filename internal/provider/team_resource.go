package provider

import (
	"context"
	"fmt"

	dtrack "github.com/DependencyTrack/client-go"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &teamResource{}
	_ resource.ResourceWithConfigure   = &teamResource{}
	_ resource.ResourceWithImportState = &teamResource{}
)

// NewTeamResource is a helper function to simplify the provider implementation.
func NewTeamResource() resource.Resource {
	return &teamResource{}
}

// Configure adds the provider configured client to the resource.
func (r *teamResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

// Metadata returns the repository type name.
func (r *teamResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

// Schema defines the schema for the resource.
func (r *teamResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"permissions": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// Create creates the repository and sets the initial Terraform state.
func (r *teamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan teamModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed order value from DependencyTrack
	allTeams, err := dtrack.FetchAll(func(po dtrack.PageOptions) (dtrack.Page[dtrack.Team], error) {
		return r.client.Team.GetAll(ctx, po)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting all Teams",
			"Could not get Teams, unexpected error: "+err.Error(),
		)
		return
	}

	for _, existing := range allTeams {
		if existing.Name == plan.Name.ValueString() {
			resp.Diagnostics.AddError(
				"Error creating team",
				fmt.Sprintf("A team with name %q exists already with UUID %q",
					plan.Name.ValueString(), existing.UUID.String()),
			)
			return
		}
	}

	team := dtrack.Team{
		Name:        plan.Name.ValueString(),
		Permissions: []dtrack.Permission{},
	}

	for _, p := range plan.Permissions.Elements() {
		team.Permissions = append(team.Permissions, dtrack.Permission{Name: valueString(p)})
	}

	// Create new team
	result, err := r.client.Team.Create(ctx, team)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating team",
			"Could not create team, unexpected error: "+err.Error(),
		)
		return
	}

	permissions, err := fetchAllMappedByUI(
		func(po dtrack.PageOptions) (dtrack.Page[dtrack.Permission], error) {
			return r.client.Permission.GetAll(ctx, po)
		},
		func(it dtrack.Permission) string {
			return it.Name
		},
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting Permission",
			fmt.Sprintf("Could not get Permissions, unexpected error: %v", err),
		)
		return
	}

	for _, perm := range team.Permissions {
		p, ok := permissions[perm.Name]
		if !ok {
			resp.Diagnostics.AddError(
				"Error add permission to team",
				fmt.Sprintf("Could not add Permission to team, permission %q not found", perm.Name),
			)
			return
		}

		_, err := r.client.Permission.AddPermissionToTeam(ctx, p, result.UUID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error adding Permission to Team",
				"Could not add Permission to Team, unexpected error: "+err.Error(),
			)
			return
		}
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(result.UUID.String())

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *teamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state teamModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed order value from DependencyTrack
	team, err := r.client.Team.Get(ctx, uuid.MustParse(state.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading DependencyTrack Repositories",
			"Could not read DependencyTrack repositories: "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	state.ID = types.StringValue(team.UUID.String())
	state.Name = types.StringValue(team.Name)

	var permissionNames []attr.Value
	for _, p := range team.Permissions {
		permissionNames = append(permissionNames, types.StringValue(p.Name))
	}
	state.Permissions = types.SetValueMust(types.StringType, permissionNames)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the repository and sets the updated Terraform state on success.
func (r *teamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	// Retrieve values from plan
	var plan teamModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	team := dtrack.Team{
		UUID:        uuid.MustParse(plan.ID.ValueString()),
		Name:        plan.Name.ValueString(),
		Permissions: []dtrack.Permission{},
	}

	for _, p := range plan.Permissions.Elements() {
		team.Permissions = append(team.Permissions, dtrack.Permission{Name: valueString(p)})
	}

	// Update existing repository
	updatedTeam, err := r.client.Team.Update(ctx, team)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating team",
			"Could not update team, unexpected error: "+err.Error(),
		)
		return
	}

	permissions, err := fetchAllMappedByUI(
		func(po dtrack.PageOptions) (dtrack.Page[dtrack.Permission], error) {
			return r.client.Permission.GetAll(ctx, po)
		},
		func(it dtrack.Permission) string {
			return it.Name
		},
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting Permission",
			fmt.Sprintf("Could not get Permissions, unexpected error: %v", err),
		)
		return
	}

	teamPermissions := mapByID(updatedTeam.Permissions, func(it dtrack.Permission) string {
		return it.Name
	})

	for _, p := range plan.Permissions.Elements() {
		name := valueString(p)
		if _, ok := teamPermissions[name]; !ok {
			perm, ok := permissions[name]
			if !ok {
				resp.Diagnostics.AddError(
					"Error add permission to team",
					fmt.Sprintf("Could not add Permission to team, permission %q not found", name),
				)
				return
			}
			_, err := r.client.Permission.AddPermissionToTeam(ctx, perm, team.UUID)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error adding Permission to Team",
					"Could not add Permission to Team, unexpected error: "+err.Error(),
				)
				return
			}
		} else {
			delete(teamPermissions, name)
		}
	}

	for _, perm := range teamPermissions {
		p, ok := permissions[perm.Name]
		if !ok {
			resp.Diagnostics.AddError(
				"Error add permission to team",
				fmt.Sprintf("Could not remove Permission from team, permission %q not found", perm.Name),
			)
			return
		}
		_, err := r.client.Permission.RemovePermissionFromTeam(ctx, p, team.UUID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error removing Permission from Team",
				"Could not remove Permission from Team, unexpected error: "+err.Error(),
			)
			return
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the repository and removes the Terraform state on success.
func (r *teamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state teamModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	team := dtrack.Team{
		UUID: uuid.MustParse(state.ID.ValueString()),
		Name: state.Name.ValueString(),
	}

	// Delete existing order
	err := r.client.Team.Delete(ctx, team)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting DependencyTrack Team",
			"Could not delete team, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *teamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
