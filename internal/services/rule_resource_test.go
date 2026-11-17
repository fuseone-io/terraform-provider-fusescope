package services

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/fuseone-io/terraform-provider-fusescope/internal/apitypes"
	conv "github.com/fuseone-io/terraform-provider-fusescope/internal/convertors"
	"github.com/fuseone-io/terraform-provider-fusescope/internal/customtypes"
	"github.com/fuseone-io/terraform-provider-fusescope/internal/schemas"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

const anomalyRule = `{
  "id": "0196",
  "alert": "checkout requests anomaly",
  "alertType": "METRIC_BASED_ALERT",
  "ruleType": "anomaly_rule",
  "schemaVersion": "v1",
  "condition": {
    "compositeQuery": {
      "queryType": "builder",
      "panelType": "graph",
      "queries": [{
        "type": "builder_query",
        "spec": {
          "name": "A",
          "signal": "metrics",
          "aggregations": [{"metricName": "http.server.requests", "timeAggregation": "rate", "spaceAggregation": "sum"}],
          "functions": [{"name": "anomaly", "args": [{"name": "z_score_threshold", "value": 3}]}]
        }
      }]
    },
    "target": 3,
    "op": "above",
    "matchType": "at_least_once",
    "seasonality": "hourly",
    "selectedQueryName": "A"
  }
}`

const thresholdRule = `{
  "id": "0197",
  "alert": "cpu high",
  "alertType": "METRIC_BASED_ALERT",
  "ruleType": "threshold_rule",
  "schemaVersion": "v2alpha1",
  "condition": {
    "compositeQuery": {
      "queryType": "builder",
      "panelType": "graph",
      "queries": [{
        "type": "builder_query",
        "spec": {
          "name": "A",
          "signal": "metrics",
          "aggregations": [{"metricName": "system.cpu.utilization", "timeAggregation": "avg", "spaceAggregation": "avg"}]
        }
      }]
    },
    "selectedQueryName": "A",
    "thresholds": {"kind": "basic", "spec": [{"name": "critical", "target": 0.9, "op": "above", "matchType": "at_least_once"}]}
  },
  "evaluation": {"kind": "rolling", "spec": {"evalWindow": "5m", "frequency": "1m"}},
  "notificationSettings": {"usePolicy": false}
}`

func ruleModel(t *testing.T, body string) schemas.RuleModel {
	t.Helper()

	var rule apitypes.RuletypesRule
	if err := json.Unmarshal([]byte(body), &rule); err != nil {
		t.Fatalf("unmarshal rule: %v", err)
	}

	ds, diags := conv.FlattenRuletypesRule(context.Background(), &rule)
	if diags.HasError() {
		t.Fatalf("flatten: %v", diags)
	}

	return *ruleResourceFromDS(ds)
}

func validateRule(t *testing.T, m schemas.RuleModel) diag.Diagnostics {
	t.Helper()
	ctx := context.Background()

	s := schemas.RuleResourceSchema(ctx)
	state := tfsdk.State{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(ctx), nil)}
	if diags := state.Set(ctx, &m); diags.HasError() {
		t.Fatalf("build config: %v", diags)
	}

	resp := &resource.ValidateConfigResponse{}
	(&ruleResource{}).ValidateConfig(ctx, resource.ValidateConfigRequest{Config: tfsdk.Config{Schema: s, Raw: state.Raw}}, resp)

	return resp.Diagnostics
}

func TestRuleValidateConfig(t *testing.T) {
	withCondition := func(m schemas.RuleModel, edit func(*customtypes.RuletypesRuleConditionValue)) schemas.RuleModel {
		c := m.Condition
		edit(&c)
		m.Condition = c
		return m
	}
	anomaly := ruleModel(t, anomalyRule)
	threshold := ruleModel(t, thresholdRule)

	tests := []struct {
		name      string
		model     schemas.RuleModel
		wantPaths []string
	}{
		{name: "v1 anomaly rule", model: anomaly},
		{name: "v2alpha1 threshold rule", model: threshold},
		{
			name:  "v1 threshold rule may exceed the anomaly target range",
			model: withCondition(func() schemas.RuleModel { m := anomaly; m.RuleType = types.StringValue("threshold_rule"); return m }(), func(c *customtypes.RuletypesRuleConditionValue) { c.Target = types.Float64Value(50) }),
		},
		{
			name:      "anomaly target below 1",
			model:     withCondition(anomaly, func(c *customtypes.RuletypesRuleConditionValue) { c.Target = types.Float64Value(0.5) }),
			wantPaths: []string{"condition.target"},
		},
		{
			name:      "anomaly target above 7",
			model:     withCondition(anomaly, func(c *customtypes.RuletypesRuleConditionValue) { c.Target = types.Float64Value(8) }),
			wantPaths: []string{"condition.target"},
		},
		{
			name:      "anomaly algorithm other than standard",
			model:     withCondition(anomaly, func(c *customtypes.RuletypesRuleConditionValue) { c.Algorithm = types.StringValue("zscore") }),
			wantPaths: []string{"condition.algorithm"},
		},
		{
			name:      "anomaly op equal",
			model:     withCondition(anomaly, func(c *customtypes.RuletypesRuleConditionValue) { c.Op = types.StringValue("equal") }),
			wantPaths: []string{"condition.op"},
		},
		{
			name: "anomaly_baseline annotation is reserved",
			model: func() schemas.RuleModel {
				m := anomaly
				m.Annotations = types.MapValueMust(types.StringType, map[string]attr.Value{"anomaly_baseline": types.StringValue("flat")})
				return m
			}(),
			wantPaths: []string{`annotations["anomaly_baseline"]`},
		},
		{
			name:      "anomaly rule on v2alpha1",
			model:     func() schemas.RuleModel { m := threshold; m.RuleType = types.StringValue("anomaly_rule"); return m }(),
			wantPaths: []string{"schema_version"},
		},
		{
			name: "v1 needs target, op and match type",
			model: withCondition(anomaly, func(c *customtypes.RuletypesRuleConditionValue) {
				c.Target = types.Float64Null()
				c.Op = types.StringNull()
				c.MatchType = types.StringNull()
			}),
			wantPaths: []string{"condition.target", "condition.op", "condition.match_type"},
		},
		{
			name:      "v1 refuses the v2alpha1 blocks",
			model:     func() schemas.RuleModel { m := threshold; m.SchemaVersion = types.StringValue("v1"); return m }(),
			wantPaths: []string{"condition.thresholds", "evaluation", "notification_settings"},
		},
		{
			name: "v2alpha1 needs thresholds",
			model: withCondition(threshold, func(c *customtypes.RuletypesRuleConditionValue) {
				c.Thresholds = types.ObjectNull(customtypes.RuletypesRuleThresholdDataValue{}.AttributeTypes(context.Background()))
			}),
			wantPaths: []string{"condition.thresholds"},
		},
		{
			name: "v2alpha1 refuses the v1 fields",
			model: withCondition(func() schemas.RuleModel {
				m := threshold
				m.EvalWindow = types.StringValue("5m")
				return m
			}(), func(c *customtypes.RuletypesRuleConditionValue) { c.Target = types.Float64Value(1) }),
			wantPaths: []string{"eval_window", "condition.target"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diags := validateRule(t, tt.model)

			var got []string
			for _, d := range diags.Errors() {
				if p, ok := d.(diag.DiagnosticWithPath); ok {
					got = append(got, p.Path().String())
				}
			}

			if len(tt.wantPaths) == 0 && diags.HasError() {
				t.Fatalf("unexpected errors: %v", diags)
			}

			for _, want := range tt.wantPaths {
				found := false
				for _, g := range got {
					found = found || g == want
				}

				if !found {
					t.Errorf("no error at %s; errors at [%s]: %v", want, strings.Join(got, ", "), diags)
				}
			}
		})
	}
}
