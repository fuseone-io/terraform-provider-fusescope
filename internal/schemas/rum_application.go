package schemas

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	rumDefaultSourceMapRetention = "90d"
	rumDefaultSampleRate         = 1.0
)

var (
	rumPlatforms           = []string{"web", "android"}
	rumSourceMapRetentions = []string{"30d", "90d", "1y", "never"}

	rumPIIPolicyAttrTypes = map[string]attr.Type{
		"drop_query":      types.BoolType,
		"query_allowlist": types.ListType{ElemType: types.StringType},
		"url_allowlist":   types.ListType{ElemType: types.StringType},
		"geo_lookup":      types.BoolType,
		"require_consent": types.BoolType,
		"hash_user_id":    types.BoolType,
	}

	rumSourceMapsAttrTypes = map[string]attr.Type{
		"retention":            types.StringType,
		"keep_newest_releases": types.BoolType,
	}
)

type RUMPIIPolicyModel struct {
	DropQuery      types.Bool     `tfsdk:"drop_query"`
	QueryAllowlist []types.String `tfsdk:"query_allowlist"`
	UrlAllowlist   []types.String `tfsdk:"url_allowlist"`
	GeoLookup      types.Bool     `tfsdk:"geo_lookup"`
	RequireConsent types.Bool     `tfsdk:"require_consent"`
	HashUserId     types.Bool     `tfsdk:"hash_user_id"`
}

type RUMSourceMapsModel struct {
	Retention          types.String `tfsdk:"retention"`
	KeepNewestReleases types.Bool   `tfsdk:"keep_newest_releases"`
}

type RUMApplicationModel struct {
	Id                   types.String        `tfsdk:"id"`
	OrgId                types.String        `tfsdk:"org_id"`
	Name                 types.String        `tfsdk:"name"`
	Platform             types.String        `tfsdk:"platform"`
	Environment          types.String        `tfsdk:"environment"`
	AllowedOrigins       []types.String      `tfsdk:"allowed_origins"`
	AllowedPackageNames  []types.String      `tfsdk:"allowed_package_names"`
	TracePropagationUrls []types.String      `tfsdk:"trace_propagation_urls"`
	SampleRate           types.Float64       `tfsdk:"sample_rate"`
	PiiPolicy            *RUMPIIPolicyModel  `tfsdk:"pii_policy"`
	SourceMaps           *RUMSourceMapsModel `tfsdk:"source_maps"`
	Key                  types.String        `tfsdk:"key"`
	KeyId                types.String        `tfsdk:"key_id"`
}

type RUMApplicationDataSourceModel struct {
	Id                   types.String        `tfsdk:"id"`
	OrgId                types.String        `tfsdk:"org_id"`
	Name                 types.String        `tfsdk:"name"`
	Platform             types.String        `tfsdk:"platform"`
	Environment          types.String        `tfsdk:"environment"`
	AllowedOrigins       []types.String      `tfsdk:"allowed_origins"`
	AllowedPackageNames  []types.String      `tfsdk:"allowed_package_names"`
	TracePropagationUrls []types.String      `tfsdk:"trace_propagation_urls"`
	SampleRate           types.Float64       `tfsdk:"sample_rate"`
	PiiPolicy            *RUMPIIPolicyModel  `tfsdk:"pii_policy"`
	SourceMaps           *RUMSourceMapsModel `tfsdk:"source_maps"`
}

func emptyStringList() types.List {
	return types.ListValueMust(types.StringType, []attr.Value{})
}

func optionalStringList(description string) schema.ListAttribute {
	return schema.ListAttribute{
		ElementType: types.StringType,
		Optional:    true,
		Computed:    true,
		Default:     listdefault.StaticValue(emptyStringList()),
		Description: description,
	}
}

func RUMApplicationResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description: "A Real User Monitoring application. Creating one also creates its first public key, exposed as `key`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"org_id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"platform": schema.StringAttribute{
				Required:      true,
				Description:   "`web` or `android`. Changing it replaces the application.",
				Validators:    []validator.String{stringvalidator.OneOf(rumPlatforms...)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"environment": schema.StringAttribute{
				Required: true,
			},
			"allowed_origins":        optionalStringList("Browser origins allowed to ingest, e.g. `https://shop.example.com` or `https://*.example.com`; `*` allows any. Web only."),
			"allowed_package_names":  optionalStringList("Android package names allowed to ingest, e.g. `com.example.shop` or `com.example.*`; `*` allows any. Android only."),
			"trace_propagation_urls": optionalStringList("URL prefixes of APIs the SDK sends `traceparent` to."),
			"sample_rate": schema.Float64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     float64default.StaticFloat64(rumDefaultSampleRate),
				Description: "Fraction of sessions the SDK records, from 0 to 1. Defaults to 1.",
				Validators:  []validator.Float64{float64validator.Between(0, 1)},
			},
			"pii_policy": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "What URLs keep before storage.",
				Default: objectdefault.StaticValue(types.ObjectValueMust(rumPIIPolicyAttrTypes, map[string]attr.Value{
					"drop_query":      types.BoolValue(true),
					"query_allowlist": emptyStringList(),
					"url_allowlist":   emptyStringList(),
					"geo_lookup":      types.BoolValue(false),
					"require_consent": types.BoolValue(false),
					"hash_user_id":    types.BoolValue(false),
				})),
				Attributes: map[string]schema.Attribute{
					"drop_query": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(true),
						Description: "Drop query parameters that are not in `query_allowlist`. Defaults to true.",
					},
					"query_allowlist": optionalStringList("Query parameter names kept when `drop_query` is true."),
					"url_allowlist":   optionalStringList("URL prefixes kept verbatim, query and fragment included."),
					"geo_lookup": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Derive the country from the client IP. The IP itself is never stored.",
					},
					"require_consent": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "The SDK starts only after the consent hook allows it.",
					},
					"hash_user_id": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "The SDK hashes the id passed to `setUser()`.",
					},
				},
			},
			"source_maps": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Source map and R8 mapping retention.",
				Default: objectdefault.StaticValue(types.ObjectValueMust(rumSourceMapsAttrTypes, map[string]attr.Value{
					"retention":            types.StringValue(rumDefaultSourceMapRetention),
					"keep_newest_releases": types.BoolValue(true),
				})),
				Attributes: map[string]schema.Attribute{
					"retention": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString(rumDefaultSourceMapRetention),
						Description: "Delete a release's maps when unused for this long: `30d`, `90d`, `1y` or `never`. Defaults to `90d`.",
						Validators:  []validator.String{stringvalidator.OneOf(rumSourceMapRetentions...)},
					},
					"keep_newest_releases": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(true),
						Description: "Never delete the 3 most recently uploaded releases. Defaults to true.",
					},
				},
			},
			"key": schema.StringAttribute{
				Computed:      true,
				Sensitive:     true,
				Description:   "The application's first public key. The backend returns it only on create, so it is null after an import.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"key_id": schema.StringAttribute{
				Computed:      true,
				Description:   "ID of `key`.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func RUMApplicationDataSourceSchema(_ context.Context) dsschema.Schema {
	list := func(description string) dsschema.ListAttribute {
		return dsschema.ListAttribute{ElementType: types.StringType, Computed: true, Description: description}
	}

	return dsschema.Schema{
		Description: "Reads a Real User Monitoring application. Its keys are not readable.",
		Attributes: map[string]dsschema.Attribute{
			"id":                     dsschema.StringAttribute{Required: true},
			"org_id":                 dsschema.StringAttribute{Computed: true},
			"name":                   dsschema.StringAttribute{Computed: true},
			"platform":               dsschema.StringAttribute{Computed: true},
			"environment":            dsschema.StringAttribute{Computed: true},
			"allowed_origins":        list("Browser origins allowed to ingest."),
			"allowed_package_names":  list("Android package names allowed to ingest."),
			"trace_propagation_urls": list("URL prefixes of APIs the SDK sends `traceparent` to."),
			"sample_rate":            dsschema.Float64Attribute{Computed: true},
			"pii_policy": dsschema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]dsschema.Attribute{
					"drop_query":      dsschema.BoolAttribute{Computed: true},
					"query_allowlist": list(""),
					"url_allowlist":   list(""),
					"geo_lookup":      dsschema.BoolAttribute{Computed: true},
					"require_consent": dsschema.BoolAttribute{Computed: true},
					"hash_user_id":    dsschema.BoolAttribute{Computed: true},
				},
			},
			"source_maps": dsschema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]dsschema.Attribute{
					"retention":            dsschema.StringAttribute{Computed: true},
					"keep_newest_releases": dsschema.BoolAttribute{Computed: true},
				},
			},
		},
	}
}
