package conv

import (
	"reflect"
	"testing"

	"github.com/fuseone-io/terraform-provider-fusescope/internal/apiclients"
	"github.com/fuseone-io/terraform-provider-fusescope/internal/schemas"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func stringValues(v ...string) []types.String {
	out := make([]types.String, 0, len(v))
	for _, s := range v {
		out = append(out, types.StringValue(s))
	}
	return out
}

func TestExpandRUMApplication(t *testing.T) {
	full := schemas.RUMApplicationModel{
		Name:                 types.StringValue("shop"),
		Platform:             types.StringValue("web"),
		Environment:          types.StringValue("production"),
		AllowedOrigins:       stringValues("https://shop.example.com"),
		AllowedPackageNames:  stringValues(),
		TracePropagationUrls: stringValues("https://api.example.com"),
		SampleRate:           types.Float64Value(0.25),
		PiiPolicy: &schemas.RUMPIIPolicyModel{
			DropQuery:      types.BoolValue(true),
			QueryAllowlist: stringValues("page"),
			UrlAllowlist:   stringValues(),
			GeoLookup:      types.BoolValue(true),
			RequireConsent: types.BoolValue(false),
			HashUserId:     types.BoolValue(true),
		},
		SourceMaps: &schemas.RUMSourceMapsModel{
			Retention:          types.StringValue("1y"),
			KeepNewestReleases: types.BoolValue(false),
		},
	}

	tests := []struct {
		name         string
		model        schemas.RUMApplicationModel
		withPlatform bool
		want         apiclients.RUMApplicationInput
	}{
		{
			name:         "create sends every field",
			model:        full,
			withPlatform: true,
			want: apiclients.RUMApplicationInput{
				Name:                 "shop",
				Platform:             "web",
				Environment:          "production",
				AllowedOrigins:       []string{"https://shop.example.com"},
				AllowedPackageNames:  []string{},
				TracePropagationURLs: []string{"https://api.example.com"},
				SampleRate:           0.25,
				PIIPolicy:            &apiclients.RUMPIIPolicy{DropQuery: true, QueryAllowlist: []string{"page"}, URLAllowlist: []string{}, GeoLookup: true, HashUserID: true},
				SourceMaps:           &apiclients.RUMSourceMapSettings{Retention: "1y"},
			},
		},
		{
			name:  "update leaves the platform out",
			model: full,
			want: apiclients.RUMApplicationInput{
				Name:                 "shop",
				Environment:          "production",
				AllowedOrigins:       []string{"https://shop.example.com"},
				AllowedPackageNames:  []string{},
				TracePropagationURLs: []string{"https://api.example.com"},
				SampleRate:           0.25,
				PIIPolicy:            &apiclients.RUMPIIPolicy{DropQuery: true, QueryAllowlist: []string{"page"}, URLAllowlist: []string{}, GeoLookup: true, HashUserID: true},
				SourceMaps:           &apiclients.RUMSourceMapSettings{Retention: "1y"},
			},
		},
		{
			name: "unset lists and blocks are sent empty or left to server defaults",
			model: schemas.RUMApplicationModel{
				Name:        types.StringValue("android-shop"),
				Platform:    types.StringValue("android"),
				Environment: types.StringValue("staging"),
				SampleRate:  types.Float64Null(),
			},
			withPlatform: true,
			want: apiclients.RUMApplicationInput{
				Name:                 "android-shop",
				Platform:             "android",
				Environment:          "staging",
				AllowedOrigins:       []string{},
				AllowedPackageNames:  []string{},
				TracePropagationURLs: []string{},
				SampleRate:           1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpandRUMApplication(tt.model, tt.withPlatform)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got  %+v\nwant %+v", got, tt.want)
			}
		})
	}
}

func TestFlattenRUMApplication(t *testing.T) {
	app := &apiclients.RUMApplication{
		ID:                   "app-1",
		OrgID:                "org-1",
		Name:                 "shop",
		Platform:             "web",
		Environment:          "production",
		AllowedOrigins:       []string{"https://shop.example.com"},
		AllowedPackageNames:  nil,
		TracePropagationURLs: []string{},
		SampleRate:           0.5,
		PIIPolicy:            apiclients.RUMPIIPolicy{DropQuery: true, QueryAllowlist: nil, URLAllowlist: []string{"https://shop.example.com/search"}, RequireConsent: true},
		SourceMaps:           apiclients.RUMSourceMapSettings{Retention: "90d", KeepNewestReleases: true},
	}

	want := &schemas.RUMApplicationDataSourceModel{
		Id:                   types.StringValue("app-1"),
		OrgId:                types.StringValue("org-1"),
		Name:                 types.StringValue("shop"),
		Platform:             types.StringValue("web"),
		Environment:          types.StringValue("production"),
		AllowedOrigins:       stringValues("https://shop.example.com"),
		AllowedPackageNames:  stringValues(),
		TracePropagationUrls: stringValues(),
		SampleRate:           types.Float64Value(0.5),
		PiiPolicy: &schemas.RUMPIIPolicyModel{
			DropQuery:      types.BoolValue(true),
			QueryAllowlist: stringValues(),
			UrlAllowlist:   stringValues("https://shop.example.com/search"),
			GeoLookup:      types.BoolValue(false),
			RequireConsent: types.BoolValue(true),
			HashUserId:     types.BoolValue(false),
		},
		SourceMaps: &schemas.RUMSourceMapsModel{
			Retention:          types.StringValue("90d"),
			KeepNewestReleases: types.BoolValue(true),
		},
	}

	got := FlattenRUMApplication(app)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got  %+v\nwant %+v", got, want)
	}
}
