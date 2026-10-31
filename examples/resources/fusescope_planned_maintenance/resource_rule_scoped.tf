# Silence only the listed rules by referencing their IDs.
resource "fusescope_notification_channel" "db" {
  name         = "webhook-db"
  display_name = "webhook-db"

  config = {
    webhook = {
      kind = "webhook"
      spec = {
        url           = "https://example.com/webhook-db"
        send_resolved = true
      }
    }
  }
}

resource "fusescope_rule" "db_connections" {
  alert          = "Database connections saturated"
  alert_type     = "METRIC_BASED_ALERT"
  rule_type      = "promql_rule"
  schema_version = "v2alpha1"

  condition = {
    composite_query = {
      panel_type = "graph"
      query_type = "promql"

      queries = [
        {
          promql = {
            type = "promql"
            spec = {
              name  = "A"
              query = "sum by (\"db.system\") ({\"db.client.connection.count\", state=\"used\"})"
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
            target     = 90
            channels   = [fusescope_notification_channel.db.display_name]
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

  notification_settings = {}
}

resource "fusescope_planned_maintenance" "db_patching" {
  name        = "db-patching"
  description = "Silence database alerts during the weekly patching window."
  alert_ids   = [fusescope_rule.db_connections.id]

  schedule = {
    timezone   = "UTC"
    start_time = "2026-01-04T01:00:00Z"
    end_time   = "2026-01-04T02:00:00Z"

    recurrence = {
      repeat_type = "weekly"
      repeat_on   = ["sunday"]
      duration    = "1h0m0s"
    }
  }
}
