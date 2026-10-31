package apiclients

import (
	"context"
	"net/http"
	"net/url"
)

type PublicDashboardInput struct {
	TimeRangeEnabled bool   `json:"timeRangeEnabled"`
	DefaultTimeRange string `json:"defaultTimeRange"`
}

type PublicDashboard struct {
	TimeRangeEnabled bool   `json:"timeRangeEnabled"`
	DefaultTimeRange string `json:"defaultTimeRange"`
	PublicPath       string `json:"publicPath"`
}

func publicDashboardPath(dashboardID string) string {
	return "/api/v1/dashboards/" + url.PathEscape(dashboardID) + "/public"
}

func (c *WrappedClient) CreatePublicDashboard(ctx context.Context, dashboardID string, in PublicDashboardInput) error {
	return c.doJSON(ctx, http.MethodPost, publicDashboardPath(dashboardID), in, nil)
}

func (c *WrappedClient) GetPublicDashboard(ctx context.Context, dashboardID string) (*PublicDashboard, error) {
	out := &PublicDashboard{}
	if err := c.doJSON(ctx, http.MethodGet, publicDashboardPath(dashboardID), nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *WrappedClient) UpdatePublicDashboard(ctx context.Context, dashboardID string, in PublicDashboardInput) error {
	return c.doJSON(ctx, http.MethodPut, publicDashboardPath(dashboardID), in, nil)
}

func (c *WrappedClient) DeletePublicDashboard(ctx context.Context, dashboardID string) error {
	return c.doJSON(ctx, http.MethodDelete, publicDashboardPath(dashboardID), nil, nil)
}
