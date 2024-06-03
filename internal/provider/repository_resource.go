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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &repositoryResource{}
	_ resource.ResourceWithConfigure   = &repositoryResource{}
	_ resource.ResourceWithImportState = &repositoryResource{}
)

// NewRepositoryResource is a helper function to simplify the provider implementation.
func NewRepositoryResource() resource.Resource {
	return &repositoryResource{}
}

// Configure adds the provider configured client to the resource.
func (r *repositoryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *repositoryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_repository"
}

// Schema defines the schema for the resource.
func (r *repositoryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"identifier": schema.StringAttribute{
				Required: true,
			},
			"type": schema.StringAttribute{
				Required:   true,
				Validators: []validator.String{&repositoryTypeValidator{}},
			},
			"url": schema.StringAttribute{
				Required: true,
			},
			"resolution_order": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Required: true,
			},
			"internal": schema.BoolAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"authentication_required": schema.BoolAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"password": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Create creates the repository and sets the initial Terraform state.
func (r *repositoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan repositoryModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed order value from DependencyTrack
	allRepos, err := dtrack.FetchAll(func(po dtrack.PageOptions) (dtrack.Page[dtrack.Repository], error) {
		return r.client.Repository.GetAll(ctx, po)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting all Repositories",
			"Could not get Repositories, unexpected error: "+err.Error(),
		)
		return
	}

	for _, existing := range allRepos {
		if existing.Identifier == plan.Identifier.ValueString() &&
			existing.Type == dtrack.RepositoryType(plan.Type.ValueString()) {
			resp.Diagnostics.AddError(
				"Error creating team",
				fmt.Sprintf("A repository with identifier %q exists already with UUID %q",
					plan.Identifier.ValueString(), existing.UUID.String()),
			)
			return
		}
	}

	repository := dtrack.Repository{
		Type:                   dtrack.RepositoryType(plan.Type.ValueString()),
		Identifier:             plan.Identifier.ValueString(),
		Url:                    plan.Url.ValueString(),
		Enabled:                plan.Enabled.ValueBool(),
		Internal:               plan.Internal.ValueBool(),
		AuthenticationRequired: plan.AuthenticationRequired.ValueBool(),
		Username:               plan.Username.ValueString(),
		Password:               plan.Password.ValueString(),
	}

	// Create new repository
	result, err := r.client.Repository.Create(ctx, repository)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating repository",
			"Could not create repository, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(result.UUID.String())
	plan.ResolutionOrder = types.Int64Value(int64(result.ResolutionOrder))
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *repositoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state repositoryModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed order value from DependencyTrack
	repos, err := dtrack.FetchAll(func(po dtrack.PageOptions) (dtrack.Page[dtrack.Repository], error) {
		return r.client.Repository.GetByType(ctx, dtrack.RepositoryType(state.Type.ValueString()), po)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading DependencyTrack Repositories",
			"Could not read DependencyTrack repositories: "+err.Error(),
		)
		return
	}

	var repo *dtrack.Repository
	for i := range repos {
		r := repos[i]
		if state.ID.ValueString() == r.UUID.String() {
			repo = &r
		}
	}
	if repo == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Overwrite items with refreshed state
	state.ID = types.StringValue(repo.UUID.String())
	state.Type = types.StringValue(string(repo.Type))
	state.Identifier = types.StringValue(repo.Identifier)
	state.Url = types.StringValue(repo.Url)
	state.ResolutionOrder = types.Int64Value(int64(repo.ResolutionOrder))
	state.Enabled = types.BoolValue(repo.Enabled)
	if repo.Internal {
		state.Internal = types.BoolValue(repo.Internal)
	}
	state.AuthenticationRequired = types.BoolValue(repo.AuthenticationRequired)
	if repo.Username != "" {
		state.Username = types.StringValue(repo.Username)
	}
	if repo.Password != "" {
		state.Password = types.StringValue(repo.Password)
	}
	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the repository and sets the updated Terraform state on success.
func (r *repositoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan repositoryModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	repository := dtrack.Repository{
		UUID:                   uuid.MustParse(plan.ID.ValueString()),
		Type:                   dtrack.RepositoryType(plan.Type.ValueString()),
		Identifier:             plan.Identifier.ValueString(),
		Url:                    plan.Url.ValueString(),
		Enabled:                plan.Enabled.ValueBool(),
		Internal:               plan.Internal.ValueBool(),
		AuthenticationRequired: plan.AuthenticationRequired.ValueBool(),
		Username:               plan.Username.ValueString(),
		Password:               plan.Password.ValueString(),
	}

	// Update existing repository
	result, err := r.client.Repository.Update(ctx, repository)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating repository",
			"Could not update repository, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ResolutionOrder = types.Int64Value(int64(result.ResolutionOrder))
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the repository and removes the Terraform state on success.
func (r *repositoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state repositoryModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing order
	err := r.client.Repository.Delete(ctx, uuid.MustParse(state.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting DependencyTrack Repository",
			"Could not delete repository, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *repositoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
