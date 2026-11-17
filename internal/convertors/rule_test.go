package conv

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/fuseone-io/terraform-provider-fusescope/internal/apitypes"
	"github.com/fuseone-io/terraform-provider-fusescope/internal/schemas"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const anomalyRuleV1 = `{
  "id": "0196",
  "state": "inactive",
  "alert": "checkout requests anomaly",
  "alertType": "METRIC_BASED_ALERT",
  "ruleType": "anomaly_rule",
  "evalWindow": "10m",
  "frequency": "1m",
  "preferredChannels": ["ops"],
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
    "algorithm": "standard",
    "seasonality": "daily",
    "selectedQueryName": "A"
  }
}`

const thresholdRuleV2Alpha1 = `{
  "id": "0197",
  "state": "inactive",
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

func flattenRuleJSON(t *testing.T, body string) *schemas.RuleDataSourceModel {
	t.Helper()

	var rule apitypes.RuletypesRule
	if err := json.Unmarshal([]byte(body), &rule); err != nil {
		t.Fatalf("unmarshal rule: %v", err)
	}

	model, diags := FlattenRuletypesRule(context.Background(), &rule)
	if diags.HasError() {
		t.Fatalf("flatten: %v", diags)
	}

	return model
}

func expandRuleJSON(t *testing.T, ds *schemas.RuleDataSourceModel) map[string]any {
	t.Helper()

	model := schemas.RuleModel{
		Alert:                ds.Alert,
		AlertType:            ds.AlertType,
		Annotations:          ds.Annotations,
		Condition:            ds.Condition,
		Description:          ds.Description,
		Disabled:             ds.Disabled,
		EvalWindow:           ds.EvalWindow,
		Evaluation:           ds.Evaluation,
		Frequency:            ds.Frequency,
		Labels:               ds.Labels,
		NotificationSettings: ds.NotificationSettings,
		PreferredChannels:    ds.PreferredChannels,
		RuleType:             ds.RuleType,
		SchemaVersion:        ds.SchemaVersion,
	}

	in, diags := ExpandRuletypesPostableRule(context.Background(), model)
	if diags.HasError() {
		t.Fatalf("expand: %v", diags)
	}

	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	return out
}

func TestFlattenRuleV1(t *testing.T) {
	m := flattenRuleJSON(t, anomalyRuleV1)

	tests := []struct {
		name string
		got  any
		want any
	}{
		{name: "schema version defaults to v1 when omitted", got: m.SchemaVersion, want: types.StringValue("v1")},
		{name: "eval window", got: m.EvalWindow, want: types.StringValue("10m")},
		{name: "frequency", got: m.Frequency, want: types.StringValue("1m")},
		{name: "preferred channels", got: m.PreferredChannels.String(), want: `["ops"]`},
		{name: "target", got: m.Condition.Target, want: types.Float64Value(3)},
		{name: "op", got: m.Condition.Op, want: types.StringValue("above")},
		{name: "match type", got: m.Condition.MatchType, want: types.StringValue("at_least_once")},
		{name: "target unit", got: m.Condition.TargetUnit, want: types.StringNull()},
		{name: "seasonality", got: m.Condition.Seasonality, want: types.StringValue("daily")},
		{name: "thresholds are null", got: m.Condition.Thresholds.IsNull(), want: true},
		{name: "evaluation is null", got: m.Evaluation.IsNull(), want: true},
		{name: "notification settings are null", got: m.NotificationSettings.IsNull(), want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			equal := tt.got == tt.want
			if got, ok := tt.got.(attr.Value); ok {
				want, _ := tt.want.(attr.Value)
				equal = got.Equal(want)
			}

			if !equal {
				t.Fatalf("got %v, want %v", tt.got, tt.want)
			}
		})
	}
}

func TestExpandRuleSchemaVersions(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		want    map[string]any
		absent  []string
		condAbs []string
	}{
		{
			name: "v1 anomaly rule sends the flat condition",
			body: anomalyRuleV1,
			want: map[string]any{
				"schemaVersion":     "v1",
				"evalWindow":        "10m",
				"frequency":         "1m",
				"target":            float64(3),
				"op":                "above",
				"matchType":         "at_least_once",
				"seasonality":       "daily",
				"algorithm":         "standard",
				"selectedQueryName": "A",
			},
			absent:  []string{"evaluation", "notificationSettings"},
			condAbs: []string{"thresholds", "targetUnit"},
		},
		{
			name: "v2alpha1 rule keeps thresholds and evaluation",
			body: thresholdRuleV2Alpha1,
			want: map[string]any{
				"schemaVersion":     "v2alpha1",
				"selectedQueryName": "A",
			},
			absent:  []string{"evalWindow", "frequency", "preferredChannels"},
			condAbs: []string{"target", "op", "matchType", "targetUnit"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := expandRuleJSON(t, flattenRuleJSON(t, tt.body))
			cond, _ := payload["condition"].(map[string]any)

			for k, want := range tt.want {
				got, ok := payload[k]
				if !ok {
					got = cond[k]
				}

				if got != want {
					t.Errorf("%s = %v, want %v", k, got, want)
				}
			}

			for _, k := range tt.absent {
				if v, ok := payload[k]; ok {
					t.Errorf("%s sent as %v, want absent", k, v)
				}
			}

			for _, k := range tt.condAbs {
				if v, ok := cond[k]; ok {
					t.Errorf("condition.%s sent as %v, want absent", k, v)
				}
			}
		})
	}

	payload := expandRuleJSON(t, flattenRuleJSON(t, thresholdRuleV2Alpha1))
	for _, k := range []string{"evaluation", "notificationSettings"} {
		if _, ok := payload[k]; !ok {
			t.Errorf("v2alpha1 payload lacks %s", k)
		}
	}

	if cond, _ := payload["condition"].(map[string]any); cond["thresholds"] == nil {
		t.Errorf("v2alpha1 payload lacks condition.thresholds")
	}
}
