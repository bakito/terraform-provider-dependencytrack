package provider

import (
	"context"
	"fmt"
	"time"

	dtrack "github.com/DependencyTrack/client-go"
	"github.com/google/uuid"
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
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
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
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
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
			"Error Reading DependencyTrack Repositories",
			"Could not read DependencyTrack repositories: "+err.Error(),
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

	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

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
	err := r.client.Repository.Delete(ctx, state.ID.ValueString())
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
