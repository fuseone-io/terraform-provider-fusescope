package services

import (
	"context"
	"fmt"

	"github.com/fuseone-io/terraform-provider-fusescope/internal/apiclients"
	conv "github.com/fuseone-io/terraform-provider-fusescope/internal/convertors"
	"github.com/fuseone-io/terraform-provider-fusescope/internal/schemas"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = (*rumApplicationResource)(nil)
	_ resource.ResourceWithConfigure   = (*rumApplicationResource)(nil)
	_ resource.ResourceWithImportState = (*rumApplicationResource)(nil)
)

type rumApplicationResource struct {
	api *apiclients.WrappedClient
}

// NewRUMApplicationResource returns the framework resource constructor.
func NewRUMApplicationResource() resource.Resource {
	return &rumApplicationResource{}
}

func (r *rumApplicationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rum_application"
}

func (r *rumApplicationResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schemas.RUMApplicationResourceSchema(ctx)
}

func (r *rumApplicationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	c, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if c != nil {
		r.api = c
	}
}

func (r *rumApplicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan schemas.RUMApplicationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.api.CreateRUMApplication(ctx, conv.ExpandRUMApplication(plan, true))
	if err != nil {
		resp.Diagnostics.AddError("Create rum_application", err.Error())
		return
	}

	key, keyID := types.StringNull(), types.StringNull()
	if created.Key != nil {
		key, keyID = types.StringValue(created.Key.Key), types.StringValue(created.Key.ID)
	}

	state := conv.RUMApplicationResourceFromDS(conv.FlattenRUMApplication(&created.RUMApplication), key, keyID)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *rumApplicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state schemas.RUMApplicationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	app, err := r.api.GetRUMApplication(ctx, id)
	if apiclients.IsNotFound(err) {
		resp.Diagnostics.AddWarning("RUM application not found", fmt.Sprintf("Removing rum_application %q from state.", id))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read rum_application", err.Error())
		return
	}

	next := conv.RUMApplicationResourceFromDS(conv.FlattenRUMApplication(app), state.Key, state.KeyId)
	resp.Diagnostics.Append(resp.State.Set(ctx, next)...)
}

func (r *rumApplicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state schemas.RUMApplicationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := r.api.UpdateRUMApplication(ctx, state.Id.ValueString(), conv.ExpandRUMApplication(plan, false))
	if err != nil {
		resp.Diagnostics.AddError("Update rum_application", err.Error())
		return
	}

	next := conv.RUMApplicationResourceFromDS(conv.FlattenRUMApplication(app), state.Key, state.KeyId)
	resp.Diagnostics.Append(resp.State.Set(ctx, next)...)
}

func (r *rumApplicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state schemas.RUMApplicationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.api.DeleteRUMApplication(ctx, state.Id.ValueString())
	if err != nil && !apiclients.IsNotFound(err) {
		resp.Diagnostics.AddError("Delete rum_application", err.Error())
	}
}

func (r *rumApplicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
