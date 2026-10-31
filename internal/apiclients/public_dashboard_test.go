package apiclients

import (
	"context"
	"net/http"
	"testing"
)

func TestPublicDashboardClient(t *testing.T) {
	ctx := context.Background()
	input := PublicDashboardInput{TimeRangeEnabled: true, DefaultTimeRange: "1h"}

	tests := []struct {
		name       string
		status     int
		response   string
		call       func(c *WrappedClient) error
		wantMethod string
		wantBody   bool
	}{
		{
			name:     "create",
			status:   http.StatusCreated,
			response: `{"status":"success","data":{"id":"pub-1"}}`,
			call: func(c *WrappedClient) error {
				return c.CreatePublicDashboard(ctx, "dash-1", input)
			},
			wantMethod: http.MethodPost,
			wantBody:   true,
		},
		{
			name:   "update",
			status: http.StatusNoContent,
			call: func(c *WrappedClient) error {
				return c.UpdatePublicDashboard(ctx, "dash-1", input)
			},
			wantMethod: http.MethodPut,
			wantBody:   true,
		},
		{
			name:   "delete",
			status: http.StatusNoContent,
			call: func(c *WrappedClient) error {
				return c.DeletePublicDashboard(ctx, "dash-1")
			},
			wantMethod: http.MethodDelete,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got recordedRequest
			c := newTestClient(t, tt.status, tt.response, &got)

			if err := tt.call(c); err != nil {
				t.Fatal(err)
			}
			if got.method != tt.wantMethod || got.path != "/api/v1/dashboards/dash-1/public" {
				t.Fatalf("request = %s %s", got.method, got.path)
			}
			if tt.wantBody && (got.body["timeRangeEnabled"] != true || got.body["defaultTimeRange"] != "1h") {
				t.Fatalf("body = %v", got.body)
			}
		})
	}

	t.Run("get", func(t *testing.T) {
		var got recordedRequest
		c := newTestClient(t, http.StatusOK, `{"status":"success","data":{"timeRangeEnabled":true,"defaultTimeRange":"1h","publicPath":"/public/dashboard/pub-1"}}`, &got)

		pd, err := c.GetPublicDashboard(ctx, "dash-1")
		if err != nil {
			t.Fatal(err)
		}
		if got.method != http.MethodGet || got.path != "/api/v1/dashboards/dash-1/public" {
			t.Fatalf("request = %s %s", got.method, got.path)
		}
		if !pd.TimeRangeEnabled || pd.DefaultTimeRange != "1h" || pd.PublicPath != "/public/dashboard/pub-1" {
			t.Fatalf("public dashboard = %+v", pd)
		}
	})

	t.Run("get reports not found", func(t *testing.T) {
		var got recordedRequest
		c := newTestClient(t, http.StatusNotFound, `{"status":"error","error":{"code":"public_dashboard_not_found","message":"public dashboard not found"}}`, &got)

		if _, err := c.GetPublicDashboard(ctx, "dash-1"); !IsNotFound(err) {
			t.Fatalf("IsNotFound(%v) = false", err)
		}
	})
}
