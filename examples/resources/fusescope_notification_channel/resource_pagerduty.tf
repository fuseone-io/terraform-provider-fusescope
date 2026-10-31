variable "pagerduty_routing_key" {
  type      = string
  sensitive = true
}

resource "fusescope_notification_channel" "pagerduty_oncall" {
  name         = "oncall-pagerduty"
  display_name = "On-call PagerDuty"

  config = {
    pagerduty = {
      kind = "pagerduty"
      spec = {
        routing_key   = var.pagerduty_routing_key
        send_resolved = true
        description   = "[{{ .Status | toUpper }}] {{ .CommonLabels.alertname }}"
        severity      = "{{ (index .Alerts 0).Labels.severity }}"
        client        = "FuseScope"
        details = {
          num_firing = "{{ .Alerts.Firing | len }}"
        }
      }
    }
  }
}
