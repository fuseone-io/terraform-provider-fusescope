# Each threshold notifies its own channels; the threshold name becomes the
# alert's severity label.
resource "fusescope_notification_channel" "team" {
  name         = "webhook-team"
  display_name = "webhook-team"

  config = {
    webhook = {
      kind = "webhook"
      spec = {
        url           = "https://example.com/webhook-team"
        send_resolved = true
      }
    }
  }
}

resource "fusescope_notification_channel" "oncall" {
  name         = "webhook-oncall"
  display_name = "webhook-oncall"

  config = {
    webhook = {
      kind = "webhook"
      spec = {
        url           = "https://example.com/webhook-oncall"
        send_resolved = true
      }
    }
  }
}

resource "fusescope_rule" "p99_latency" {
  alert      = "p99 latency per service"
  alert_type = "TRACES_BASED_ALERT"
  rule_type  = "threshold_rule"

  annotations = {
    summary     = "{{$service.name}} p99 latency is {{$value}}"
    description = "p99 latency of {{$service.name}} server spans crossed {{$threshold}}."
  }

  condition = {
    composite_query = {
      panel_type = "graph"
      query_type = "builder"
      unit       = "ns"

      queries = [
        {
          builder_query = {
            type = "builder_query"
            spec = {
              traces = {
                name   = "A"
                signal = "traces"

                aggregations = [
                  { expression = "p99(duration_nano)" }
                ]

                filter = {
                  expression = "kind_string = 'Server' AND deployment.environment = 'production'"
                }

                group_by = [
                  {
                    name          = "service.name"
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
            name            = "warning"
            op              = "above"
            match_type      = "on_average"
            target          = 500
            target_unit     = "ms"
            recovery_target = 400
            channels        = [fusescope_notification_channel.team.display_name]
          },
          {
            name        = "critical"
            op          = "above"
            match_type  = "on_average"
            target      = 1000
            target_unit = "ms"
            channels = [
              fusescope_notification_channel.team.display_name,
              fusescope_notification_channel.oncall.display_name,
            ]
          }
        ]
      }
    }
  }

  evaluation = {
    rolling = {
      kind = "rolling"
      spec = {
        eval_window = "10m"
        frequency   = "1m"
      }
    }
  }

  notification_settings = {
    group_by = ["service.name"]
  }

  schema_version = "v2alpha1"
}
