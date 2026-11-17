package services

import (
	"context"
	"fmt"
	"slices"

	"github.com/fuseone-io/terraform-provider-fusescope/internal/schemas"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

const (
	ruleSchemaV1       = "v1"
	ruleSchemaV2Alpha1 = "v2alpha1"
	ruleTypeAnomaly    = "anomaly_rule"

	anomalyMinTarget = 1
	anomalyMaxTarget = 7
	// anomalyBaselineAnnotation is set by the backend on alerts scored against a flat baseline.
	anomalyBaselineAnnotation = "anomaly_baseline"
	anomalyAlgorithmStandard  = "standard"
)

var (
	_ resource.ResourceWithValidateConfig = (*ruleResource)(nil)

	anomalyOps = []string{"above", "below", "outside_bounds"}
)

// ValidateConfig reports at plan time what the API refuses for each schema version and for
// anomaly rules.
func (r *ruleResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var m schemas.RuleModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &m)...)
	if resp.Diagnostics.HasError() {
		return
	}

	switch m.SchemaVersion.ValueString() {
	case ruleSchemaV1:
		validateRuleV1(m, &resp.Diagnostics)
	case ruleSchemaV2Alpha1:
		validateRuleV2Alpha1(m, &resp.Diagnostics)
	}

	if m.RuleType.ValueString() == ruleTypeAnomaly {
		validateAnomalyRule(m, &resp.Diagnostics)
	}
}

func validateRuleV1(m schemas.RuleModel, diags *diag.Diagnostics) {
	var required []ruleAttr
	v2Only := []ruleAttr{
		{path.Root("evaluation"), m.Evaluation},
		{path.Root("notification_settings"), m.NotificationSettings},
	}

	if c := m.Condition; !c.IsNull() && !c.IsUnknown() {
		cond := path.Root("condition")
		required = []ruleAttr{
			{cond.AtName("target"), c.Target},
			{cond.AtName("op"), c.Op},
			{cond.AtName("match_type"), c.MatchType},
		}
		v2Only = append(v2Only, ruleAttr{cond.AtName("thresholds"), c.Thresholds})
	}

	for _, a := range required {
		if a.value.IsNull() {
			diags.AddAttributeError(a.path, "Missing attribute",
				fmt.Sprintf("%s is required with schema_version %q.", a.path, ruleSchemaV1))
		}
	}

	for _, a := range v2Only {
		if !a.value.IsNull() {
			diags.AddAttributeError(a.path, "Attribute not used",
				fmt.Sprintf("%s is not used with schema_version %q.", a.path, ruleSchemaV1))
		}
	}
}

type ruleAttr struct {
	path  path.Path
	value attr.Value
}

func validateRuleV2Alpha1(m schemas.RuleModel, diags *diag.Diagnostics) {
	required := []ruleAttr{
		{path.Root("evaluation"), m.Evaluation},
		{path.Root("notification_settings"), m.NotificationSettings},
	}
	v1Only := []ruleAttr{
		{path.Root("eval_window"), m.EvalWindow},
		{path.Root("frequency"), m.Frequency},
		{path.Root("preferred_channels"), m.PreferredChannels},
	}

	if c := m.Condition; !c.IsNull() && !c.IsUnknown() {
		cond := path.Root("condition")
		required = append(required, ruleAttr{cond.AtName("thresholds"), c.Thresholds})
		v1Only = append(v1Only,
			ruleAttr{cond.AtName("target"), c.Target},
			ruleAttr{cond.AtName("op"), c.Op},
			ruleAttr{cond.AtName("match_type"), c.MatchType},
			ruleAttr{cond.AtName("target_unit"), c.TargetUnit},
		)
	}

	for _, a := range required {
		if a.value.IsNull() {
			diags.AddAttributeError(a.path, "Missing attribute",
				fmt.Sprintf("%s is required with schema_version %q.", a.path, ruleSchemaV2Alpha1))
		}
	}

	for _, a := range v1Only {
		if !a.value.IsNull() {
			diags.AddAttributeError(a.path, "Attribute not used",
				fmt.Sprintf("%s is only used with schema_version %q.", a.path, ruleSchemaV1))
		}
	}
}

func validateAnomalyRule(m schemas.RuleModel, diags *diag.Diagnostics) {
	if v := m.SchemaVersion; !v.IsUnknown() && v.ValueString() != ruleSchemaV1 {
		diags.AddAttributeError(path.Root("schema_version"), "Unsupported schema version",
			fmt.Sprintf("rule_type %q supports schema_version %q, not %q.", ruleTypeAnomaly, ruleSchemaV1, v.ValueString()))
	}

	if !m.Annotations.IsNull() && !m.Annotations.IsUnknown() {
		if _, ok := m.Annotations.Elements()[anomalyBaselineAnnotation]; ok {
			diags.AddAttributeError(path.Root("annotations").AtMapKey(anomalyBaselineAnnotation), "Reserved annotation",
				fmt.Sprintf("The %s annotation is set by anomaly rules.", anomalyBaselineAnnotation))
		}
	}

	c := m.Condition
	if c.IsNull() || c.IsUnknown() {
		return
	}

	cond := path.Root("condition")
	if t := c.Target; !t.IsNull() && !t.IsUnknown() && (t.ValueFloat64() < anomalyMinTarget || t.ValueFloat64() > anomalyMaxTarget) {
		diags.AddAttributeError(cond.AtName("target"), "Invalid anomaly target",
			fmt.Sprintf("An anomaly rule's target is a z-score between %d and %d, got %v.", anomalyMinTarget, anomalyMaxTarget, t.ValueFloat64()))
	}

	if a := c.Algorithm; !a.IsNull() && !a.IsUnknown() && a.ValueString() != anomalyAlgorithmStandard {
		diags.AddAttributeError(cond.AtName("algorithm"), "Unsupported anomaly algorithm",
			fmt.Sprintf("condition.algorithm must be %q, got %q.", anomalyAlgorithmStandard, a.ValueString()))
	}

	if op := c.Op; !op.IsNull() && !op.IsUnknown() && !slices.Contains(anomalyOps, op.ValueString()) {
		diags.AddAttributeError(cond.AtName("op"), "Unsupported anomaly operator",
			fmt.Sprintf("An anomaly rule's condition.op is one of %q, got %q.", anomalyOps, op.ValueString()))
	}
}
