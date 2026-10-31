package apiclients

import (
	"context"
	"net/http"
	"net/url"
)

const rumApplicationsPath = "/api/v1/rum/applications"

type RUMPIIPolicy struct {
	DropQuery      bool     `json:"dropQuery"`
	QueryAllowlist []string `json:"queryAllowlist"`
	URLAllowlist   []string `json:"urlAllowlist"`
	GeoLookup      bool     `json:"geoLookup"`
	RequireConsent bool     `json:"requireConsent"`
	HashUserID     bool     `json:"hashUserId"`
}

type RUMSourceMapSettings struct {
	Retention          string `json:"retention"`
	KeepNewestReleases bool   `json:"keepNewestReleases"`
}

type RUMApplication struct {
	ID                   string               `json:"id"`
	OrgID                string               `json:"orgId"`
	Name                 string               `json:"name"`
	Platform             string               `json:"platform"`
	Environment          string               `json:"environment"`
	AllowedOrigins       []string             `json:"allowedOrigins"`
	AllowedPackageNames  []string             `json:"allowedPackageNames"`
	TracePropagationURLs []string             `json:"tracePropagationUrls"`
	SampleRate           float64              `json:"sampleRate"`
	PIIPolicy            RUMPIIPolicy         `json:"piiPolicy"`
	SourceMaps           RUMSourceMapSettings `json:"sourceMaps"`
}

type RUMApplicationKey struct {
	ID            string `json:"id"`
	ApplicationID string `json:"applicationId"`
	Last4         string `json:"last4"`
	Key           string `json:"key"`
}

type RUMCreatedApplication struct {
	RUMApplication
	Key *RUMApplicationKey `json:"key"`
}

// RUMApplicationInput is the create and update body. Platform is set on create
// only: the backend does not change an application's platform.
type RUMApplicationInput struct {
	Name                 string                `json:"name"`
	Platform             string                `json:"platform,omitempty"`
	Environment          string                `json:"environment"`
	AllowedOrigins       []string              `json:"allowedOrigins"`
	AllowedPackageNames  []string              `json:"allowedPackageNames"`
	TracePropagationURLs []string              `json:"tracePropagationUrls"`
	SampleRate           float64               `json:"sampleRate"`
	PIIPolicy            *RUMPIIPolicy         `json:"piiPolicy,omitempty"`
	SourceMaps           *RUMSourceMapSettings `json:"sourceMaps,omitempty"`
}

func rumApplicationPath(id string) string {
	return rumApplicationsPath + "/" + url.PathEscape(id)
}

func (c *WrappedClient) CreateRUMApplication(ctx context.Context, in RUMApplicationInput) (*RUMCreatedApplication, error) {
	out := &RUMCreatedApplication{}
	if err := c.doJSON(ctx, http.MethodPost, rumApplicationsPath, in, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *WrappedClient) GetRUMApplication(ctx context.Context, id string) (*RUMApplication, error) {
	out := &RUMApplication{}
	if err := c.doJSON(ctx, http.MethodGet, rumApplicationPath(id), nil, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *WrappedClient) UpdateRUMApplication(ctx context.Context, id string, in RUMApplicationInput) (*RUMApplication, error) {
	out := &RUMApplication{}
	if err := c.doJSON(ctx, http.MethodPut, rumApplicationPath(id), in, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *WrappedClient) DeleteRUMApplication(ctx context.Context, id string) error {
	return c.doJSON(ctx, http.MethodDelete, rumApplicationPath(id), nil, nil)
}
