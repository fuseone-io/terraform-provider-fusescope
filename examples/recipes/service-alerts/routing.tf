# Route policies apply to rules with notification_settings.use_policy = true.
# A threshold's name becomes the alert's severity label.
resource "fusescope_route_policy" "critical" {
  name        = "${var.service_name}-critical"
  description = "Page on-call and notify Slack for critical ${var.service_name} alerts."
  kind        = "policy"
  expression  = "service.name == \"${var.service_name}\" && severity == \"critical\""
  channels = [
    fusescope_notification_channel.pagerduty.display_name,
    fusescope_notification_channel.slack.display_name,
  ]
}

resource "fusescope_route_policy" "warning" {
  name        = "${var.service_name}-warning"
  description = "Notify Slack for ${var.service_name} warnings."
  kind        = "policy"
  expression  = "service.name == \"${var.service_name}\" && severity == \"warning\""
  channels    = [fusescope_notification_channel.slack.display_name]
}
