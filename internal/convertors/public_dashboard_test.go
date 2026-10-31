package conv

import (
	"reflect"
	"testing"

	"github.com/fuseone-io/terraform-provider-fusescope/internal/apiclients"
	"github.com/fuseone-io/terraform-provider-fusescope/internal/schemas"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPublicDashboardConversion(t *testing.T) {
	tests := []struct {
		name  string
		model schemas.PublicDashboardModel
		want  apiclients.PublicDashboardInput
	}{
		{
			name: "custom time range",
			model: schemas.PublicDashboardModel{
				DashboardId:      types.StringValue("dash-1"),
				TimeRangeEnabled: types.BoolValue(true),
				DefaultTimeRange: types.StringValue("6h"),
			},
			want: apiclients.PublicDashboardInput{TimeRangeEnabled: true, DefaultTimeRange: "6h"},
		},
		{
			name: "fixed time range",
			model: schemas.PublicDashboardModel{
				DashboardId:      types.StringValue("dash-1"),
				TimeRangeEnabled: types.BoolValue(false),
				DefaultTimeRange: types.StringValue("30m"),
			},
			want: apiclients.PublicDashboardInput{DefaultTimeRange: "30m"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExpandPublicDashboard(tt.model); got != tt.want {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}

	t.Run("flatten uses the dashboard ID as the resource ID", func(t *testing.T) {
		got := FlattenPublicDashboard("dash-1", &apiclients.PublicDashboard{TimeRangeEnabled: true, DefaultTimeRange: "1h", PublicPath: "/public/dashboard/pub-1"})
		want := &schemas.PublicDashboardModel{
			Id:               types.StringValue("dash-1"),
			DashboardId:      types.StringValue("dash-1"),
			TimeRangeEnabled: types.BoolValue(true),
			DefaultTimeRange: types.StringValue("1h"),
			PublicPath:       types.StringValue("/public/dashboard/pub-1"),
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %+v, want %+v", got, want)
		}
	})
}
