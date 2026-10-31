package services

import (
	"fmt"

	"github.com/fuseone-io/terraform-provider-fusescope/internal/apiclients"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// clientFromProviderData returns nil without diagnostics when providerData is
// nil, which happens before the provider is configured.
func clientFromProviderData(providerData any) (*apiclients.WrappedClient, diag.Diagnostics) {
	var diags diag.Diagnostics
	if providerData == nil {
		return nil, diags
	}

	c, ok := providerData.(*apiclients.WrappedClient)
	if !ok {
		diags.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *apiclients.WrappedClient, got %T. This is a provider bug.", providerData),
		)
	}
	return c, diags
}
