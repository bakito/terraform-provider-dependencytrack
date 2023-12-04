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
	_ resource.Resource                = &oidcGroupResource{}
	_ resource.ResourceWithConfigure   = &oidcGroupResource{}
	_ resource.ResourceWithImportState = &oidcGroupResource{}
)

// NewOidcGroupResource is a helper function to simplify the provider implementation.
func NewOidcGroupResource() resource.Resource {
	return &oidcGroupResource{}
}

// Configure adds the provider configured client to the resource.
func (r *oidcGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Metadata returns the oidcGroup type name.
func (r *oidcGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oidc_group"
}

// Schema defines the schema for the resource.
func (r *oidcGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			//"last_updated": schema.StringAttribute{
			//	Computed: true,
			//},
			"name": schema.StringAttribute{
				Required: true,
			},
			"teams": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// Create creates the oidcGroup and sets the initial Terraform state.
func (r *oidcGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan oidcGroupModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	oidcGroup := dtrack.OIDCGroup{
		Name: plan.Name.ValueString(),
	}

	// Create new oidcGroup
	result, err := r.client.OIDC.CreateGroup(ctx, oidcGroup.Name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating oidcGroup",
			"Could not create oidcGroup, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(result.UUID.String())
	//plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Get refreshed order value from DependencyTrack
	teams, err := dtrack.FetchAll(func(po dtrack.PageOptions) (dtrack.Page[dtrack.Team], error) {
		return r.client.Team.GetAll(ctx, po)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting Teams",
			"Could not get Teams, unexpected error: "+err.Error(),
		)
		return
	}

	teamMap := make(map[string]dtrack.Team)
	for i := range teams {
		t := teams[i]
		teamMap[t.Name] = t
	}

	for _, t := range plan.Teams.Elements() {
		name := valueString(t)
		team, ok := teamMap[name]
		if !ok {
			resp.Diagnostics.AddError(
				"Error mapping Team",
				fmt.Sprintf("Could not create OIDC Group - Team Mapping, team %q not found", name),
			)
			return
		}

		_, err = r.client.OIDC.AddTeamMapping(ctx, dtrack.OIDCMappingRequest{
			Group: result.UUID,
			Team:  team.UUID,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating OIDC Group - Team Mapping",
				"Could not create OIDC Group - Team Mapping, unexpected error: "+err.Error(),
			)
			return
		}
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func valueString(t attr.Value) string {
	if value, ok := t.(types.String); ok {
		return value.ValueString()
	}
	return ""
}

// Read refreshes the Terraform state with the latest data.
func (r *oidcGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state oidcGroupModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed order value from DependencyTrack
	groups, err := dtrack.FetchAll(func(po dtrack.PageOptions) (dtrack.Page[dtrack.OIDCGroup], error) {
		return r.client.OIDC.GetAllGroups(ctx, po)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading DependencyTrack OIDC Groups",
			"Could not read DependencyTrack OIDC Groups: "+err.Error(),
		)
		return
	}

	var group *dtrack.OIDCGroup
	for i := range groups {
		r := groups[i]
		if state.ID.ValueString() == r.UUID.String() {
			group = &r
		}
	}
	if group == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Overwrite items with refreshed state
	state.ID = types.StringValue(group.UUID.String())
	state.Name = types.StringValue(group.Name)

	teams, err := dtrack.FetchAll(func(po dtrack.PageOptions) (dtrack.Page[dtrack.Team], error) {
		return r.client.OIDC.GetAllTeamsOf(ctx, *group, po)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read DependencyTrack Teams",
			err.Error(),
		)
		return
	}

	var teamNames []attr.Value
	for _, team := range teams {
		teamNames = append(teamNames, types.StringValue(team.Name))
	}
	state.Teams = types.SetValueMust(types.StringType, teamNames)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the oidcGroup and sets the updated Terraform state on success.
func (r *oidcGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	// Retrieve values from plan
	var plan oidcGroupModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	oidcGroup := dtrack.OIDCGroup{
		UUID: uuid.MustParse(plan.ID.ValueString()),
		Name: plan.Name.ValueString(),
	}

	// Update existing oidcGroup
	_, err := r.client.OIDC.UpdateGroup(ctx, oidcGroup)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating oidcGroup",
			"Could not update oidcGroup, unexpected error: "+err.Error(),
		)
		return
	}

	// Get refreshed order value from DependencyTrack
	groupTeams, err := fetchAllMappedByUI(
		func(po dtrack.PageOptions) (dtrack.Page[dtrack.Team], error) {
			return r.client.OIDC.GetAllTeamsOf(ctx, oidcGroup, po)
		},
		func(it dtrack.Team) string {
			return it.Name
		},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting Teams of Group",
			fmt.Sprintf("Could not get Teams of Group %v, unexpected error: %v", oidcGroup.UUID, err),
		)
		return
	}

	allTeams, err := fetchAllMappedByUI(
		func(po dtrack.PageOptions) (dtrack.Page[dtrack.Team], error) {
			return r.client.Team.GetAll(ctx, po)
		},
		func(it dtrack.Team) string {
			return it.Name

		},
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting Teams ",
			fmt.Sprintf("Could not get Teams, unexpected error: %v", err),
		)
		return
	}

	for _, t := range plan.Teams.Elements() {

		name := valueString(t)
		_, ok := groupTeams[name]
		if !ok {
			team, ok := allTeams[name]
			if !ok {
				resp.Diagnostics.AddError(
					"Error finding Team",
					fmt.Sprintf("Could not find Team %q", name),
				)
				return
			}

			_, err = r.client.OIDC.AddTeamMapping(ctx, dtrack.OIDCMappingRequest{
				Group: oidcGroup.UUID,
				Team:  team.UUID,
			})
			if err != nil {
				resp.Diagnostics.AddError(
					"Error creating OIDC Group - Team Mapping",
					fmt.Sprintf("Could not create OIDC Group (%s) - Team Mapping (%s/%s), unexpected error: %v", oidcGroup.UUID, name, team.UUID, err),
				)
				return
			}
		} else {
			delete(groupTeams, name)
		}
	}

	for teamName := range groupTeams {
		team, ok := allTeams[teamName]
		if !ok {
			resp.Diagnostics.AddError(
				"Error finding Team",
				fmt.Sprintf("Could not find Team %q", teamName),
			)
			return
		}
		for _, m := range team.MappedOIDCGroups {
			if m.Group.UUID.String() == plan.ID.ValueString() {
				err = r.client.OIDC.RemoveTeamMapping(ctx, m.UUID)
				if err != nil {
					resp.Diagnostics.AddError(
						"Error creating OIDC Group - Team Mapping",
						"Could not create OIDC Group - Team Mapping, unexpected error: "+err.Error(),
					)
					return
				}
			}
		}
	}

	//plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the oidcGroup and removes the Terraform state on success.
func (r *oidcGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state oidcGroupModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing order
	err := r.client.OIDC.DeleteGroup(ctx, uuid.MustParse(state.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting DependencyTrack Repository",
			"Could not delete oidcGroup, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *oidcGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
