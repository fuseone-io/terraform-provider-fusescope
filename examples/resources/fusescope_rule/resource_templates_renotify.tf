# Annotations are Go templates: {{$value}} and {{$threshold}} are the observed
# value and the target, {{$<label>}} (e.g. {{$service.name}}) reads an alert label.
resource "fusescope_notification_channel" "traffic" {
  name         = "webhook-traffic"
  display_name = "webhook-traffic"

  config = {
    webhook = {
      kind = "webhook"
      spec = {
        url           = "https://example.com/webhook-traffic"
        send_resolved = true
      }
    }
  }
}

resource "fusescope_rule" "traffic_drop" {
  alert       = "Request rate dropped"
  alert_type  = "TRACES_BASED_ALERT"
  rule_type   = "threshold_rule"
  description = "Server request rate per service fell below the expected floor."

  annotations = {
    summary     = "{{$service.name}} receives {{$value}} req/s (floor {{$threshold}})"
    description = <<-EOT
      Request rate of {{$service.name}} in {{$deployment.environment}} is {{$value}} req/s,
      below the floor of {{$threshold}} req/s. Check the load balancer and recent deploys.
    EOT
    runbook_url = "https://runbooks.example.com/traffic-drop"
  }

  labels = {
    team = "checkout"
  }

  condition = {
    alert_on_absent = true
    absent_for      = 15

    composite_query = {
      panel_type = "graph"
      query_type = "builder"
      unit       = "reqps"

      queries = [
        {
          builder_query = {
            type = "builder_query"
            spec = {
              traces = {
                name   = "A"
                signal = "traces"

                aggregations = [
                  { expression = "rate()" }
                ]

                filter = {
                  expression = "kind_string = 'Server' AND service.name = 'checkout'"
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
            op         = "below"
            match_type = "all_the_times"
            target     = 1
            channels   = [fusescope_notification_channel.traffic.display_name]
          }
        ]
      }
    }
  }

  evaluation = {
    rolling = {
      kind = "rolling"
      spec = {
        eval_window = "15m"
        frequency   = "5m"
      }
    }
  }

  notification_settings = {
    group_by             = ["service.name", "deployment.environment"]
    new_group_eval_delay = "10m"
    renotify = {
      enabled      = true
      interval     = "1h"
      alert_states = ["firing", "nodata"]
    }
  }

  schema_version = "v2alpha1"
}
