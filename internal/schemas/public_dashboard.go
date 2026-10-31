package schemas

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PublicDashboardModel struct {
	Id               types.String `tfsdk:"id"`
	DashboardId      types.String `tfsdk:"dashboard_id"`
	TimeRangeEnabled types.Bool   `tfsdk:"time_range_enabled"`
	DefaultTimeRange types.String `tfsdk:"default_time_range"`
	PublicPath       types.String `tfsdk:"public_path"`
}

func PublicDashboardResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description: "Publishes a dashboard at a public, unauthenticated URL.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Same as `dashboard_id`.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"dashboard_id": schema.StringAttribute{
				Required:      true,
				Description:   "ID of the dashboard to publish. Changing it replaces the resource.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"time_range_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Let viewers pick the time range, up to 30 days. Defaults to false.",
			},
			"default_time_range": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("30m"),
				Description: "Time range shown by default, as a Go duration up to 720h (e.g. `30m`, `6h`). Defaults to `30m`.",
			},
			"public_path": schema.StringAttribute{
				Computed:      true,
				Description:   "Path of the public page, relative to the FuseScope URL.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}
