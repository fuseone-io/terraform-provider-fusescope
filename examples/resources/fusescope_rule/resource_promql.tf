# Dotted OpenTelemetry names are quoted in PromQL: {"metric.name", "attr.name"="..."}.
resource "fusescope_notification_channel" "infra" {
  name         = "webhook-infra"
  display_name = "webhook-infra"

  config = {
    webhook = {
      kind = "webhook"
      spec = {
        url           = "https://example.com/webhook-infra"
        send_resolved = true
      }
    }
  }
}

resource "fusescope_rule" "host_memory" {
  alert      = "Host memory above 90%"
  alert_type = "METRIC_BASED_ALERT"
  rule_type  = "promql_rule"

  annotations = {
    summary     = "{{$host.name}} memory usage is {{$value}}"
    description = "Memory used on {{$host.name}} has stayed above {{$threshold}} for 10 minutes."
  }

  condition = {
    composite_query = {
      panel_type = "graph"
      query_type = "promql"
      unit       = "percentunit"

      queries = [
        {
          promql = {
            type = "promql"
            spec = {
              name  = "A"
              query = "max by (\"host.name\") ({\"system.memory.utilization\", state=\"used\"})"
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
            name       = "critical"
            op         = "above"
            match_type = "all_the_times"
            target     = 0.9
            channels   = [fusescope_notification_channel.infra.display_name]
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

  labels = {
    team = "infra"
  }

  notification_settings = {
    group_by = ["host.name"]
  }

  schema_version = "v2alpha1"
}
