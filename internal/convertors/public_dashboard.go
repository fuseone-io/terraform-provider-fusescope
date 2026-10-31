package conv

import (
	"github.com/fuseone-io/terraform-provider-fusescope/internal/apiclients"
	"github.com/fuseone-io/terraform-provider-fusescope/internal/schemas"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ExpandPublicDashboard(m schemas.PublicDashboardModel) apiclients.PublicDashboardInput {
	return apiclients.PublicDashboardInput{
		TimeRangeEnabled: m.TimeRangeEnabled.ValueBool(),
		DefaultTimeRange: m.DefaultTimeRange.ValueString(),
	}
}

func FlattenPublicDashboard(dashboardID string, pd *apiclients.PublicDashboard) *schemas.PublicDashboardModel {
	return &schemas.PublicDashboardModel{
		Id:               types.StringValue(dashboardID),
		DashboardId:      types.StringValue(dashboardID),
		TimeRangeEnabled: types.BoolValue(pd.TimeRangeEnabled),
		DefaultTimeRange: types.StringValue(pd.DefaultTimeRange),
		PublicPath:       types.StringValue(pd.PublicPath),
	}
}
