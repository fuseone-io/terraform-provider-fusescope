variable "msteams_webhook_url" {
  type      = string
  sensitive = true
}

resource "fusescope_notification_channel" "msteams_platform" {
  name         = "platform-msteams"
  display_name = "Platform Teams"

  config = {
    msteams = {
      kind = "msteams"
      spec = {
        webhook_url   = var.msteams_webhook_url
        send_resolved = true
        title         = "[{{ .Status | toUpper }}] {{ .CommonLabels.alertname }}"
        text          = "{{ range .Alerts }}{{ .Annotations.description }}\n{{ end }}"
      }
    }
  }
}
