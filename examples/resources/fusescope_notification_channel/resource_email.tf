resource "fusescope_notification_channel" "email_sre" {
  name         = "sre-email"
  display_name = "SRE email"

  config = {
    email = {
      kind = "email"
      spec = {
        to            = "sre@example.com"
        send_resolved = true
        headers = {
          Subject = "[{{ .Status | toUpper }}] {{ .CommonLabels.alertname }}"
        }
      }
    }
  }
}
