# Anomaly rules use schema_version v1: the condition holds the target, op and
# match type directly. The target is a z-score from 1 to 7; target_unit is
# ignored because the score is unitless.
resource "fusescope_notification_channel" "anomalies" {
  name         = "webhook-anomalies"
  display_name = "webhook-anomalies"

  config = {
    webhook = {
      kind = "webhook"
      spec = {
        url           = "https://example.com/webhook-anomalies"
        send_resolved = true
      }
    }
  }
}

resource "fusescope_rule" "request_rate_anomaly" {
  alert          = "checkout request rate anomaly"
  alert_type     = "METRIC_BASED_ALERT"
  rule_type      = "anomaly_rule"
  schema_version = "v1"

  eval_window        = "10m"
  frequency          = "1m"
  preferred_channels = [fusescope_notification_channel.anomalies.display_name]

  condition = {
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
                    metric_name       = "http.server.request.duration"
                    time_aggregation  = "rate"
                    space_aggregation = "sum"
                  }
                ]

                # The anomaly function must be the query's last function.
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
    seasonality         = "daily"
  }

  labels = {
    severity = "warning"
  }

  annotations = {
    summary = "Request rate z-score {{$value}} is above {{$threshold}}"
  }
}
