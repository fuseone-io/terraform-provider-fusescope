package provider

import (
	"context"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestProviderRegistersTypes(t *testing.T) {
	ctx := context.Background()
	p := &fusescopeProvider{name: "fusescope"}

	var resources []string
	for _, newResource := range p.Resources(ctx) {
		r := newResource()

		meta := &resource.MetadataResponse{}
		r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "fusescope"}, meta)
		resources = append(resources, meta.TypeName)

		schema := &resource.SchemaResponse{}
		r.Schema(ctx, resource.SchemaRequest{}, schema)
		if schema.Diagnostics.HasError() {
			t.Fatalf("%s schema: %v", meta.TypeName, schema.Diagnostics)
		}
		if diags := schema.Schema.ValidateImplementation(ctx); diags.HasError() {
			t.Fatalf("%s schema: %v", meta.TypeName, diags)
		}
	}

	var dataSources []string
	for _, newDataSource := range p.DataSources(ctx) {
		d := newDataSource()

		meta := &datasource.MetadataResponse{}
		d.Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "fusescope"}, meta)
		dataSources = append(dataSources, meta.TypeName)

		schema := &datasource.SchemaResponse{}
		d.Schema(ctx, datasource.SchemaRequest{}, schema)
		if diags := schema.Schema.ValidateImplementation(ctx); diags.HasError() {
			t.Fatalf("%s schema: %v", meta.TypeName, diags)
		}
	}

	tests := []struct {
		name       string
		registered []string
		want       string
	}{
		{name: "rum application resource", registered: resources, want: "fusescope_rum_application"},
		{name: "rum application data source", registered: dataSources, want: "fusescope_rum_application"},
		{name: "public dashboard resource", registered: resources, want: "fusescope_public_dashboard"},
		{name: "role resource is not offered", registered: resources, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.want == "" {
				if slices.Contains(tt.registered, "fusescope_role") {
					t.Fatalf("fusescope_role resource is registered, but FuseScope cannot manage custom roles")
				}
				return
			}
			if !slices.Contains(tt.registered, tt.want) {
				t.Fatalf("%s not registered; got %v", tt.want, tt.registered)
			}
		})
	}
}
