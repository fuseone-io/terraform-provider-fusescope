resource "fusescope_notification_channel" "logs" {
  name         = "webhook-logs"
  display_name = "webhook-logs"

  config = {
    webhook = {
      kind = "webhook"
      spec = {
        url           = "https://example.com/webhook-logs"
        send_resolved = true
      }
    }
  }
}

resource "fusescope_rule" "error_logs" {
  alert      = "Error logs per service"
  alert_type = "LOGS_BASED_ALERT"
  rule_type  = "threshold_rule"

  annotations = {
    summary     = "{{$service.name}} logged {{$value}} errors in 5m"
    description = "{{$service.name}} in {{$deployment.environment}} logged {{$value}} ERROR/FATAL lines in the last 5 minutes (threshold {{$threshold}})."
  }

  condition = {
    composite_query = {
      panel_type = "graph"
      query_type = "builder"

      queries = [
        {
          builder_query = {
            type = "builder_query"
            spec = {
              logs = {
                name   = "A"
                signal = "logs"

                aggregations = [
                  { expression = "count()" }
                ]

                filter = {
                  expression = "severity_text IN ('ERROR', 'FATAL')"
                }

                group_by = [
                  {
                    name          = "service.name"
                    field_context = "resource"
                  },
                  {
                    name          = "deployment.environment"
                    field_context = "resource"
                  }
                ]
              }
            }
          }
        }
      ]
    }

    selected_query_name = "A"

    thresholds = {
      basic = {
        kind = "basic"
        spec = [
          {
            name       = "warning"
            op         = "above"
            match_type = "in_total"
            target     = 50
            channels   = [fusescope_notification_channel.logs.display_name]
          }
        ]
      }
    }
  }

  evaluation = {
    rolling = {
      kind = "rolling"
      spec = {
        eval_window = "5m"
        frequency   = "1m"
      }
    }
  }

  notification_settings = {
    group_by = ["service.name", "deployment.environment"]
  }

  schema_version = "v2alpha1"
}
