variable "opsgenie_api_key" {
  type      = string
  sensitive = true
}

resource "fusescope_notification_channel" "opsgenie_oncall" {
  name         = "oncall-opsgenie"
  display_name = "On-call Opsgenie"

  config = {
    opsgenie = {
      kind = "opsgenie"
      spec = {
        api_key       = var.opsgenie_api_key
        send_resolved = true
        message       = "{{ .CommonLabels.alertname }}"
        description   = "{{ range .Alerts }}{{ .Annotations.description }}\n{{ end }}"
        priority      = "{{ if eq (index .Alerts 0).Labels.severity \"critical\" }}P1{{ else }}P3{{ end }}"
      }
    }
  }
}
