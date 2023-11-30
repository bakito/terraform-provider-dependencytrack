package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	dtrack "github.com/DependencyTrack/client-go"
	"github.com/google/uuid"
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
	_ resource.Resource              = &repositoryResource{}
	_ resource.ResourceWithConfigure = &repositoryResource{}
)

// NewRepositoryResource is a helper function to simplify the provider implementation.
func NewRepositoryResource() resource.Resource {
	return &repositoryResource{}
}

// repositoryResource is the repository implementation.
type repositoryResource struct {
	client *dtrack.Client
}

// Configure adds the provider configured client to the repository.
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

// Schema defines the schema for the repository.
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
			"username": schema.StringAttribute{
				Optional: true,
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

	repository := dtrack.Repository{
		Type:       dtrack.RepositoryType(plan.Type.ValueString()),
		Identifier: plan.Identifier.ValueString(),
		Url:        plan.Url.ValueString(),
		Enabled:    plan.Enabled.ValueBool(),
		Internal:   plan.Internal.ValueBool(),
		Username:   plan.Username.ValueString(),
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
		repo = &repos[i]
	}
	if repo == nil {
		resp.Diagnostics.AddError(
			"Error Reading DependencyTrack Repositories",
			"Could not read DependencyTrack repository ID "+state.ID.ValueString()+": "+err.Error(),
		)
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
	if repo.Username != "" {
		state.Username = types.StringValue(repo.Username)
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
		UUID:       uuid.MustParse(plan.ID.ValueString()),
		Type:       dtrack.RepositoryType(plan.Type.ValueString()),
		Identifier: plan.Identifier.ValueString(),
		Url:        plan.Url.ValueString(),
		Enabled:    plan.Enabled.ValueBool(),
		Internal:   plan.Internal.ValueBool(),
		Username:   plan.Username.ValueString(),
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
	err := r.client.Repository.Delete(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting DependnecyTrack Repository",
			"Could not delete repository, unexpected error: "+err.Error(),
		)
		return
	}
}

var repositoryTypes = []string{"CPAN", "MAVEN", "NPM", "GEM", "PYPI", "NUGET", "HEX", "COMPOSER", "CARGO", "GO_MODULES", "UNSUPPORTED"}

type repositoryTypeValidator struct {
}

func (r repositoryTypeValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Available Type Values: %s", strings.Join(repositoryTypes, ", "))
}

func (r repositoryTypeValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("# Available Type Values: %s\n\n- ", strings.Join(repositoryTypes, "\n- "))
}

func (r repositoryTypeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	tp := req.ConfigValue.ValueString()
	for _, t := range repositoryTypes {
		if t == tp {
			return
		}
	}
	resp.Diagnostics.AddError(
		fmt.Sprintf("Unknown Repository Type: %q", tp),
		r.Description(ctx),
	)
}
