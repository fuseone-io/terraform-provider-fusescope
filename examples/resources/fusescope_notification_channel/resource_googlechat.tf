variable "googlechat_webhook_url" {
  type      = string
  sensitive = true
}

resource "fusescope_notification_channel" "googlechat_platform" {
  name         = "platform-googlechat"
  display_name = "Platform Google Chat"

  config = {
    googlechat = {
      kind = "googlechat"
      spec = {
        webhook_url   = var.googlechat_webhook_url
        send_resolved = true
        title         = "[{{ .Status | toUpper }}] {{ .CommonLabels.alertname }}"
        text          = "{{ range .Alerts }}*{{ .Annotations.summary }}*\n{{ end }}"
      }
    }
  }
}
