package conv

import (
	"github.com/fuseone-io/terraform-provider-fusescope/internal/apiclients"
	"github.com/fuseone-io/terraform-provider-fusescope/internal/schemas"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ExpandRUMApplication builds the create body, or the update body when
// withPlatform is false. Lists are always sent so that removing every entry
// clears them; an unset pii_policy or source_maps is left to server defaults.
func ExpandRUMApplication(m schemas.RUMApplicationModel, withPlatform bool) apiclients.RUMApplicationInput {
	in := apiclients.RUMApplicationInput{
		Name:                 m.Name.ValueString(),
		Environment:          m.Environment.ValueString(),
		AllowedOrigins:       expandStrings(m.AllowedOrigins),
		AllowedPackageNames:  expandStrings(m.AllowedPackageNames),
		TracePropagationURLs: expandStrings(m.TracePropagationUrls),
		SampleRate:           1,
	}

	if withPlatform {
		in.Platform = m.Platform.ValueString()
	}

	if !m.SampleRate.IsNull() && !m.SampleRate.IsUnknown() {
		in.SampleRate = m.SampleRate.ValueFloat64()
	}

	if p := m.PiiPolicy; p != nil {
		in.PIIPolicy = &apiclients.RUMPIIPolicy{
			DropQuery:      p.DropQuery.ValueBool(),
			QueryAllowlist: expandStrings(p.QueryAllowlist),
			URLAllowlist:   expandStrings(p.UrlAllowlist),
			GeoLookup:      p.GeoLookup.ValueBool(),
			RequireConsent: p.RequireConsent.ValueBool(),
			HashUserID:     p.HashUserId.ValueBool(),
		}
	}

	if s := m.SourceMaps; s != nil {
		in.SourceMaps = &apiclients.RUMSourceMapSettings{
			Retention:          s.Retention.ValueString(),
			KeepNewestReleases: s.KeepNewestReleases.ValueBool(),
		}
	}

	return in
}

func FlattenRUMApplication(app *apiclients.RUMApplication) *schemas.RUMApplicationDataSourceModel {
	return &schemas.RUMApplicationDataSourceModel{
		Id:                   types.StringValue(app.ID),
		OrgId:                types.StringValue(app.OrgID),
		Name:                 types.StringValue(app.Name),
		Platform:             types.StringValue(app.Platform),
		Environment:          types.StringValue(app.Environment),
		AllowedOrigins:       flattenStrings(app.AllowedOrigins),
		AllowedPackageNames:  flattenStrings(app.AllowedPackageNames),
		TracePropagationUrls: flattenStrings(app.TracePropagationURLs),
		SampleRate:           types.Float64Value(app.SampleRate),
		PiiPolicy: &schemas.RUMPIIPolicyModel{
			DropQuery:      types.BoolValue(app.PIIPolicy.DropQuery),
			QueryAllowlist: flattenStrings(app.PIIPolicy.QueryAllowlist),
			UrlAllowlist:   flattenStrings(app.PIIPolicy.URLAllowlist),
			GeoLookup:      types.BoolValue(app.PIIPolicy.GeoLookup),
			RequireConsent: types.BoolValue(app.PIIPolicy.RequireConsent),
			HashUserId:     types.BoolValue(app.PIIPolicy.HashUserID),
		},
		SourceMaps: &schemas.RUMSourceMapsModel{
			Retention:          types.StringValue(app.SourceMaps.Retention),
			KeepNewestReleases: types.BoolValue(app.SourceMaps.KeepNewestReleases),
		},
	}
}

// RUMApplicationResourceFromDS keeps the key fields from prior, since the
// backend never returns the key after create.
func RUMApplicationResourceFromDS(ds *schemas.RUMApplicationDataSourceModel, key, keyID types.String) *schemas.RUMApplicationModel {
	return &schemas.RUMApplicationModel{
		Id:                   ds.Id,
		OrgId:                ds.OrgId,
		Name:                 ds.Name,
		Platform:             ds.Platform,
		Environment:          ds.Environment,
		AllowedOrigins:       ds.AllowedOrigins,
		AllowedPackageNames:  ds.AllowedPackageNames,
		TracePropagationUrls: ds.TracePropagationUrls,
		SampleRate:           ds.SampleRate,
		PiiPolicy:            ds.PiiPolicy,
		SourceMaps:           ds.SourceMaps,
		Key:                  key,
		KeyId:                keyID,
	}
}

func expandStrings(in []types.String) []string {
	out := make([]string, 0, len(in))
	for _, v := range in {
		out = append(out, v.ValueString())
	}
	return out
}

func flattenStrings(in []string) []types.String {
	out := make([]types.String, 0, len(in))
	for _, v := range in {
		out = append(out, types.StringValue(v))
	}
	return out
}
