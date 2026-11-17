# Anomaly rule alerting on deviation from a learned seasonal baseline.
resource "fusescope_notification_channel" "slack" {
  name         = "slack-scenario-02"
  display_name = "slack-scenario-02"

  config = {
    webhook = {
      kind = "webhook"
      spec = {
        url           = "https://example.com/webhook"
        send_resolved = true
      }
    }
  }
}

resource "fusescope_rule" "scenario_02" {
  alert          = "testdata-anomaly-seasonality"
  alert_type     = "METRIC_BASED_ALERT"
  rule_type      = "anomaly_rule"
  schema_version = "v1"

  eval_window        = "5m"
  frequency          = "1m"
  preferred_channels = [fusescope_notification_channel.slack.display_name]

  condition = {
    algorithm           = "standard"
    seasonality         = "daily"
    require_min_points  = true
    required_num_points = 4

    composite_query = {
      panel_type = "graph"
      query_type = "builder"

      queries = [
        {
          builder_query = {
            type = "builder_query"
            spec = {
              metrics = {
                name   = "A"
                signal = "metrics"

                aggregations = [
                  {
                    metric_name       = "http.server.request.count"
                    space_aggregation = "sum"
                    time_aggregation  = "rate"
                    temporality       = "delta"
                  }
                ]

                functions = [
                  {
                    name = "anomaly"
                    args = [
                      {
                        name  = "z_score_threshold"
                        value = jsonencode(3)
                      }
                    ]
                  }
                ]
              }
            }
          }
        }
      ]
    }

    selected_query_name = "A"
    target              = 3
    op                  = "above"
    match_type          = "at_least_once"
  }
}
