resource "fusescope_notification_channel" "errors" {
  name         = "webhook-errors"
  display_name = "webhook-errors"

  config = {
    webhook = {
      kind = "webhook"
      spec = {
        url           = "https://example.com/webhook-errors"
        send_resolved = true
      }
    }
  }
}

resource "fusescope_rule" "error_rate" {
  alert      = "Error rate per service"
  alert_type = "TRACES_BASED_ALERT"
  rule_type  = "threshold_rule"

  annotations = {
    summary     = "{{$service.name}} error rate is {{$value}}%"
    description = "{{$value}}% of {{$service.name}} server spans failed (threshold {{$threshold}}%)."
  }

  condition = {
    composite_query = {
      panel_type = "graph"
      query_type = "builder"
      unit       = "percent"

      queries = [
        {
          builder_query = {
            type = "builder_query"
            spec = {
              traces = {
                name     = "A"
                signal   = "traces"
                disabled = true

                aggregations = [
                  { expression = "count()" }
                ]

                filter = {
                  expression = "kind_string = 'Server' AND has_error = true"
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
        },
        {
          builder_query = {
            type = "builder_query"
            spec = {
              traces = {
                name     = "B"
                signal   = "traces"
                disabled = true

                aggregations = [
                  { expression = "count()" }
                ]

                filter = {
                  expression = "kind_string = 'Server'"
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
        },
        {
          builder_formula = {
            type = "builder_formula"
            spec = {
              name       = "F1"
              expression = "A * 100 / B"
            }
          }
        }
      ]
    }

    selected_query_name = "F1"

    thresholds = {
      basic = {
        kind = "basic"
        spec = [
          {
            name       = "critical"
            op         = "above"
            match_type = "on_average"
            target     = 5
            channels   = [fusescope_notification_channel.errors.display_name]
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
    group_by = ["service.name"]
  }

  schema_version = "v2alpha1"
}
