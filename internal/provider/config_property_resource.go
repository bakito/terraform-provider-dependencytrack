package provider

import (
	"context"
	"fmt"

	dtrack "github.com/DependencyTrack/client-go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &configPropertyResource{}
	_ resource.ResourceWithConfigure   = &configPropertyResource{}
	_ resource.ResourceWithImportState = &configPropertyResource{}
)

// NewConfigPropertyResource is a helper function to simplify the provider implementation.
func NewConfigPropertyResource() resource.Resource {
	return &configPropertyResource{}
}

// Configure adds the provider configured client to the resource.
func (r *configPropertyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Metadata returns the configProperty type name.
func (r *configPropertyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_property"
}

// Schema defines the schema for the resource.
func (r *configPropertyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"group": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"type": schema.StringAttribute{
				Required: true,
			},
			"value": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

// Create creates the configProperty and sets the initial Terraform state.
func (r *configPropertyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError(
		"Creating Configuration Properties is not supported",
		"Configuration Properties can not be created",
	)
}

// Read refreshes the Terraform state with the latest data.
func (r *configPropertyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state configPropertyModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed order value from DependencyTrack
	properties, err := r.client.Config.GetAll(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading DependencyTrack Configuration Properties",
			"Could not read DependencyTrack Configuration Properties: "+err.Error(),
		)
		return
	}

	var stateProperty *dtrack.ConfigProperty
	for _, property := range properties {
		if state.ID.ValueString() == configPropertyID(property) {
			stateProperty = &property
		}
	}

	if stateProperty == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Overwrite items with refreshed state
	state.ID = types.StringValue(configPropertyID(*stateProperty))
	state.Group = types.StringValue(stateProperty.GroupName)
	state.Name = types.StringValue(stateProperty.Name)
	state.Type = types.StringValue(stateProperty.Type)
	state.Value = types.StringValue(stateProperty.Value)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the configProperty and sets the updated Terraform state on success.
func (r *configPropertyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan configPropertyModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	configProperty := dtrack.ConfigProperty{
		GroupName: plan.Group.ValueString(),
		Name:      plan.Name.ValueString(),
		Type:      plan.Type.ValueString(),
		Value:     plan.Value.ValueString(),
	}

	// Update existing configProperty
	_, err := r.client.Config.Update(ctx, configProperty)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating config property",
			"Could not update config property, unexpected error: "+err.Error(),
		)
		return
	}

	// plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the configProperty and removes the Terraform state on success.
func (r *configPropertyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddError(
		"Deleting Configuration Properties is not supported",
		"Configuration Properties can not be deleted",
	)
}

func (r *configPropertyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
