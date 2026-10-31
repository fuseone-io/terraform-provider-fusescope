package apiclients

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type recordedRequest struct {
	method string
	path   string
	apiKey string
	body   map[string]any
}

func newTestClient(t *testing.T, status int, response string, got *recordedRequest) *WrappedClient {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.method = r.Method
		got.path = r.URL.Path
		got.apiKey = r.Header.Get(apiKeyHeader)
		raw, _ := io.ReadAll(r.Body)
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, &got.body); err != nil {
				t.Errorf("request body is not JSON: %v", err)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(response))
	}))
	t.Cleanup(srv.Close)

	c, err := New(srv.URL, "test-key", "terraform", "test", 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

const rumApplicationJSON = `{
	"id": "app-1",
	"orgId": "org-1",
	"name": "shop",
	"platform": "web",
	"environment": "production",
	"allowedOrigins": ["https://shop.example.com"],
	"allowedPackageNames": [],
	"tracePropagationUrls": [],
	"sampleRate": 0.5,
	"piiPolicy": {"dropQuery": true, "queryAllowlist": ["page"], "urlAllowlist": [], "geoLookup": false, "requireConsent": true, "hashUserId": false},
	"sourceMaps": {"retention": "90d", "keepNewestReleases": true},
	"createdBy": "admin@example.com",
	"firstSeenAt": null
}`

func TestRUMApplicationClient(t *testing.T) {
	ctx := context.Background()
	input := RUMApplicationInput{
		Name:                 "shop",
		Platform:             "web",
		Environment:          "production",
		AllowedOrigins:       []string{"https://shop.example.com"},
		AllowedPackageNames:  []string{},
		TracePropagationURLs: []string{},
		SampleRate:           0.5,
		PIIPolicy:            &RUMPIIPolicy{DropQuery: true, QueryAllowlist: []string{"page"}, URLAllowlist: []string{}, RequireConsent: true},
		SourceMaps:           &RUMSourceMapSettings{Retention: "90d", KeepNewestReleases: true},
	}

	t.Run("create returns the application and its first key", func(t *testing.T) {
		var got recordedRequest
		created := `{"status":"success","data":` + rumApplicationJSON[:len(rumApplicationJSON)-1] + `,"key":{"id":"key-1","applicationId":"app-1","last4":"wxyz","status":"active","key":"fsr_secret_wxyz","createdAt":"2026-10-31T00:00:00Z","expiresAt":null,"lastUsedAt":null}}}`
		c := newTestClient(t, http.StatusCreated, created, &got)

		app, err := c.CreateRUMApplication(ctx, input)
		if err != nil {
			t.Fatal(err)
		}

		if got.method != http.MethodPost || got.path != "/api/v1/rum/applications" {
			t.Fatalf("request = %s %s", got.method, got.path)
		}
		if got.apiKey != "test-key" {
			t.Fatalf("api key header = %q", got.apiKey)
		}
		if got.body["platform"] != "web" || got.body["sampleRate"] != 0.5 {
			t.Fatalf("body = %v", got.body)
		}
		if app.ID != "app-1" || app.PIIPolicy.QueryAllowlist[0] != "page" || !app.PIIPolicy.RequireConsent {
			t.Fatalf("application = %+v", app.RUMApplication)
		}
		if app.Key == nil || app.Key.Key != "fsr_secret_wxyz" || app.Key.ID != "key-1" {
			t.Fatalf("key = %+v", app.Key)
		}
	})

	t.Run("get decodes the envelope", func(t *testing.T) {
		var got recordedRequest
		c := newTestClient(t, http.StatusOK, `{"status":"success","data":`+rumApplicationJSON+`}`, &got)

		app, err := c.GetRUMApplication(ctx, "app-1")
		if err != nil {
			t.Fatal(err)
		}
		if got.method != http.MethodGet || got.path != "/api/v1/rum/applications/app-1" {
			t.Fatalf("request = %s %s", got.method, got.path)
		}
		if app.Name != "shop" || app.SourceMaps.Retention != "90d" || app.OrgID != "org-1" {
			t.Fatalf("application = %+v", app)
		}
	})

	t.Run("get reports not found", func(t *testing.T) {
		var got recordedRequest
		c := newTestClient(t, http.StatusNotFound, `{"status":"error","error":{"code":"rum_application_not_found","message":"not found"}}`, &got)

		_, err := c.GetRUMApplication(ctx, "gone")
		if !IsNotFound(err) {
			t.Fatalf("IsNotFound(%v) = false", err)
		}
	})

	t.Run("update omits the platform", func(t *testing.T) {
		var got recordedRequest
		c := newTestClient(t, http.StatusOK, `{"status":"success","data":`+rumApplicationJSON+`}`, &got)

		update := input
		update.Platform = ""
		if _, err := c.UpdateRUMApplication(ctx, "app-1", update); err != nil {
			t.Fatal(err)
		}
		if got.method != http.MethodPut || got.path != "/api/v1/rum/applications/app-1" {
			t.Fatalf("request = %s %s", got.method, got.path)
		}
		if _, ok := got.body["platform"]; ok {
			t.Fatalf("update body carries platform: %v", got.body)
		}
	})

	t.Run("delete accepts no content", func(t *testing.T) {
		var got recordedRequest
		c := newTestClient(t, http.StatusNoContent, ``, &got)

		if err := c.DeleteRUMApplication(ctx, "app-1"); err != nil {
			t.Fatal(err)
		}
		if got.method != http.MethodDelete || got.path != "/api/v1/rum/applications/app-1" {
			t.Fatalf("request = %s %s", got.method, got.path)
		}
	})
}
