package conv

import "github.com/fuseone-io/terraform-provider-fusescope/internal/apitypes"

// ruleSchemaVersion fills in v1, which the API omits from v1 rules.
func ruleSchemaVersion(v apitypes.RuletypesRuleSchemaVersion) string {
	if v == "" {
		return string(apitypes.RuletypesRuleSchemaVersionV1)
	}

	return string(v)
}
