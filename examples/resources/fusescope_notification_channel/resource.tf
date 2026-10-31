variable "slack_webhook_url" {
  type      = string
  sensitive = true
}

resource "fusescope_notification_channel" "slack_oncall" {
  name         = "oncall-slack"
  display_name = "On-call Slack"

  config = {
    slack = {
      kind = "slack"
      spec = {
        api_url       = var.slack_webhook_url
        channel       = "#alerts-oncall"
        send_resolved = true
        title         = "[{{ .Status | toUpper }}] {{ .CommonLabels.alertname }}"
        text          = <<-EOT
          {{ range .Alerts -}}
          *Summary:* {{ .Annotations.summary }}
          *Service:* {{ index .Labels "service.name" }}
          {{ end }}
        EOT
      }
    }
  }
}
