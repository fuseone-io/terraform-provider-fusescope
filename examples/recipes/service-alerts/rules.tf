locals {
  server_spans = "service.name = '${var.service_name}' AND deployment.environment = '${var.environment}' AND kind_string = 'Server'"

  # Grouping by service.name puts it on every alert, which the route policies match on.
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
    group_by   = ["service.name", "deployment.environment"]
    use_policy = true
    renotify = {
      enabled      = true
      interval     = "1h"
      alert_states = ["firing"]
    }
  }

  labels = {
    service = var.service_name
  }
}

resource "fusescope_rule" "latency" {
  alert          = "${var.service_name}: p99 latency"
  alert_type     = "TRACES_BASED_ALERT"
  rule_type      = "threshold_rule"
  schema_version = "v2alpha1"
  labels         = local.labels

  annotations = {
    summary     = "{{$service.name}} p99 latency is {{$value}} (threshold {{$threshold}})"
    description = "p99 latency of {{$service.name}} server spans in {{$deployment.environment}} crossed {{$threshold}}."
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
                name         = "A"
                signal       = "traces"
                aggregations = [{ expression = "p99(duration_nano)" }]
                filter       = { expression = local.server_spans }
                group_by     = local.group_by
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
            name        = "warning"
            op          = "above"
            match_type  = "on_average"
            target      = var.latency_p99_ms.warning
            target_unit = "ms"
          },
          {
            name        = "critical"
            op          = "above"
            match_type  = "on_average"
            target      = var.latency_p99_ms.critical
            target_unit = "ms"
          }
        ]
      }
    }
  }

  evaluation            = local.evaluation
  notification_settings = local.notification_settings
}

resource "fusescope_rule" "error_rate" {
  alert          = "${var.service_name}: error rate"
  alert_type     = "TRACES_BASED_ALERT"
  rule_type      = "threshold_rule"
  schema_version = "v2alpha1"
  labels         = local.labels

  annotations = {
    summary     = "{{$service.name}} error rate is {{$value}}% (threshold {{$threshold}}%)"
    description = "{{$value}}% of {{$service.name}} server spans in {{$deployment.environment}} failed."
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
                name         = "A"
                signal       = "traces"
                disabled     = true
                aggregations = [{ expression = "count()" }]
                filter       = { expression = "${local.server_spans} AND has_error = true" }
                group_by     = local.group_by
              }
            }
          }
        },
        {
          builder_query = {
            type = "builder_query"
            spec = {
              traces = {
                name         = "B"
                signal       = "traces"
                disabled     = true
                aggregations = [{ expression = "count()" }]
                filter       = { expression = local.server_spans }
                group_by     = local.group_by
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
            target     = var.error_rate_percent
          }
        ]
      }
    }
  }

  evaluation            = local.evaluation
  notification_settings = local.notification_settings
}

resource "fusescope_rule" "traffic" {
  alert          = "${var.service_name}: traffic drop"
  alert_type     = "TRACES_BASED_ALERT"
  rule_type      = "threshold_rule"
  schema_version = "v2alpha1"
  labels         = local.labels

  annotations = {
    summary     = "{{$service.name}} receives {{$value}} req/s (floor {{$threshold}})"
    description = "Request rate of {{$service.name}} in {{$deployment.environment}} fell below {{$threshold}} req/s, or stopped reporting."
  }

  condition = {
    alert_on_absent = true
    absent_for      = 10

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
                name         = "A"
                signal       = "traces"
                aggregations = [{ expression = "rate()" }]
                filter       = { expression = local.server_spans }
                group_by     = local.group_by
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
            target     = var.min_request_rate
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

  notification_settings = local.notification_settings
}

resource "fusescope_rule" "error_logs" {
  alert          = "${var.service_name}: error logs"
  alert_type     = "LOGS_BASED_ALERT"
  rule_type      = "threshold_rule"
  schema_version = "v2alpha1"
  labels         = local.labels

  annotations = {
    summary     = "{{$service.name}} logged {{$value}} errors in 5m (threshold {{$threshold}})"
    description = "{{$service.name}} in {{$deployment.environment}} logged {{$value}} ERROR/FATAL lines in the last 5 minutes."
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
                name         = "A"
                signal       = "logs"
                aggregations = [{ expression = "count()" }]
                filter = {
                  expression = "service.name = '${var.service_name}' AND deployment.environment = '${var.environment}' AND severity_text IN ('ERROR', 'FATAL')"
                }
                group_by = local.group_by
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
          }
        ]
      }
    }
  }

  evaluation            = local.evaluation
  notification_settings = local.notification_settings
}
