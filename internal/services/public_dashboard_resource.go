package services

import (
	"context"
	"fmt"

	"github.com/fuseone-io/terraform-provider-fusescope/internal/apiclients"
	conv "github.com/fuseone-io/terraform-provider-fusescope/internal/convertors"
	"github.com/fuseone-io/terraform-provider-fusescope/internal/schemas"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

var (
	_ resource.Resource                = (*publicDashboardResource)(nil)
	_ resource.ResourceWithConfigure   = (*publicDashboardResource)(nil)
	_ resource.ResourceWithImportState = (*publicDashboardResource)(nil)
)

type publicDashboardResource struct {
	api *apiclients.WrappedClient
}

// NewPublicDashboardResource returns the framework resource constructor.
func NewPublicDashboardResource() resource.Resource {
	return &publicDashboardResource{}
}

func (r *publicDashboardResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_public_dashboard"
}

func (r *publicDashboardResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schemas.PublicDashboardResourceSchema(ctx)
}

func (r *publicDashboardResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	c, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if c != nil {
		r.api = c
	}
}

func (r *publicDashboardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan schemas.PublicDashboardModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dashboardID := plan.DashboardId.ValueString()
	if err := r.api.CreatePublicDashboard(ctx, dashboardID, conv.ExpandPublicDashboard(plan)); err != nil {
		resp.Diagnostics.AddError("Create public_dashboard", err.Error())
		return
	}

	r.refresh(ctx, dashboardID, &resp.State, &resp.Diagnostics, "Refresh public_dashboard after create")
}

func (r *publicDashboardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state schemas.PublicDashboardModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dashboardID := state.DashboardId.ValueString()
	pd, err := r.api.GetPublicDashboard(ctx, dashboardID)
	if apiclients.IsNotFound(err) {
		resp.Diagnostics.AddWarning("Public dashboard not found", fmt.Sprintf("Removing public_dashboard %q from state.", dashboardID))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read public_dashboard", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, conv.FlattenPublicDashboard(dashboardID, pd))...)
}

func (r *publicDashboardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan schemas.PublicDashboardModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dashboardID := plan.DashboardId.ValueString()
	if err := r.api.UpdatePublicDashboard(ctx, dashboardID, conv.ExpandPublicDashboard(plan)); err != nil {
		resp.Diagnostics.AddError("Update public_dashboard", err.Error())
		return
	}

	r.refresh(ctx, dashboardID, &resp.State, &resp.Diagnostics, "Refresh public_dashboard after update")
}

func (r *publicDashboardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state schemas.PublicDashboardModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.api.DeletePublicDashboard(ctx, state.DashboardId.ValueString())
	if err != nil && !apiclients.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete public_dashboard", err.Error())
	}
}

// ImportState takes the dashboard ID.
func (r *publicDashboardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resource.ImportStatePassthroughID(ctx, path.Root("dashboard_id"), req, resp)
}

func (r *publicDashboardResource) refresh(ctx context.Context, dashboardID string, state *tfsdk.State, diags *diag.Diagnostics, summary string) {
	pd, err := r.api.GetPublicDashboard(ctx, dashboardID)
	if err != nil {
		diags.AddError(summary, err.Error())
		return
	}

	diags.Append(state.Set(ctx, conv.FlattenPublicDashboard(dashboardID, pd))...)
}
