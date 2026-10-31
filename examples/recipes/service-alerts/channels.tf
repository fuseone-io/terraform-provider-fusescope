resource "fusescope_notification_channel" "slack" {
  name         = "${var.service_name}-slack"
  display_name = "${var.service_name} Slack"

  config = {
    slack = {
      kind = "slack"
      spec = {
        api_url       = var.slack_webhook_url
        channel       = var.slack_channel
        send_resolved = true
        title         = "[{{ .Status | toUpper }}] {{ .CommonLabels.alertname }}"
        text          = "{{ range .Alerts }}{{ .Annotations.summary }}\n{{ end }}"
      }
    }
  }
}

resource "fusescope_notification_channel" "pagerduty" {
  name         = "${var.service_name}-pagerduty"
  display_name = "${var.service_name} PagerDuty"

  config = {
    pagerduty = {
      kind = "pagerduty"
      spec = {
        routing_key   = var.pagerduty_routing_key
        send_resolved = true
        description   = "{{ .CommonLabels.alertname }}: {{ .CommonAnnotations.summary }}"
        severity      = "{{ .CommonLabels.severity }}"
      }
    }
  }
}
