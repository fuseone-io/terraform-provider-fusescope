package services

import (
	"context"

	"github.com/fuseone-io/terraform-provider-fusescope/internal/apiclients"
	conv "github.com/fuseone-io/terraform-provider-fusescope/internal/convertors"
	"github.com/fuseone-io/terraform-provider-fusescope/internal/schemas"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var (
	_ datasource.DataSource              = (*rumApplicationDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*rumApplicationDataSource)(nil)
)

type rumApplicationDataSource struct {
	api *apiclients.WrappedClient
}

// NewRUMApplicationDataSource returns the framework datasource constructor.
func NewRUMApplicationDataSource() datasource.DataSource {
	return &rumApplicationDataSource{}
}

func (d *rumApplicationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rum_application"
}

func (d *rumApplicationDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schemas.RUMApplicationDataSourceSchema(ctx)
}

func (d *rumApplicationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	c, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	if c != nil {
		d.api = c
	}
}

func (d *rumApplicationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg schemas.RUMApplicationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := d.api.GetRUMApplication(ctx, cfg.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read rum_application datasource", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, conv.FlattenRUMApplication(app))...)
}
