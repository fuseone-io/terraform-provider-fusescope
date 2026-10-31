# Exceptions-based rules query the exceptions table with ClickHouse SQL. The
# query must return `value` and `interval` columns; other columns become labels.
resource "fusescope_notification_channel" "exceptions" {
  name         = "webhook-exceptions"
  display_name = "webhook-exceptions"

  config = {
    webhook = {
      kind = "webhook"
      spec = {
        url           = "https://example.com/webhook-exceptions"
        send_resolved = true
      }
    }
  }
}

resource "fusescope_rule" "exceptions" {
  alert      = "Exceptions per service"
  alert_type = "EXCEPTIONS_BASED_ALERT"
  rule_type  = "threshold_rule"

  annotations = {
    summary     = "{{$serviceName}} raised {{$value}} {{$exceptionType}} exceptions"
    description = "{{$serviceName}} raised {{$value}} {{$exceptionType}} exceptions in the last 5 minutes (threshold {{$threshold}})."
  }

  condition = {
    composite_query = {
      panel_type = "graph"
      query_type = "clickhouse_sql"

      queries = [
        {
          clickhouse_sql = {
            type = "clickhouse_sql"
            spec = {
              name  = "A"
              query = <<-EOT
                SELECT
                  toStartOfInterval(timestamp, toIntervalMinute(1)) AS interval,
                  serviceName,
                  exceptionType,
                  toFloat64(count()) AS value
                FROM signoz_traces.distributed_signoz_error_index_v2
                WHERE timestamp BETWEEN {{.start_datetime}} AND {{.end_datetime}}
                GROUP BY serviceName, exceptionType, interval
              EOT
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
            target     = 10
            channels   = [fusescope_notification_channel.exceptions.display_name]
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
    group_by = ["serviceName", "exceptionType"]
  }

  schema_version = "v2alpha1"
}
