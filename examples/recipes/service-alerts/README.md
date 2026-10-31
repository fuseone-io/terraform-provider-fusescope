# Service alerts

Golden-signal alerts for one service, routed by severity:

| Rule | Signal | Severity |
|------|--------|----------|
| p99 latency | traces | warning, critical |
| error rate | traces | critical |
| traffic drop (or no data) | traces | warning |
| error logs | logs | warning |

Rules set `notification_settings.use_policy = true`, so the route policies pick the
channels: criticals page PagerDuty and post to Slack, warnings post to Slack only.
Alerts carry `service.name` and `deployment.environment` from the group-by, and
`severity` from the threshold name.

## Usage

```sh
export FUSESCOPE_ENDPOINT=https://fusescope.example.com
export FUSESCOPE_ACCESS_TOKEN=...   # service account API key
export TF_VAR_slack_webhook_url=...
export TF_VAR_pagerduty_routing_key=...

terraform init
terraform apply -var service_name=checkout -var environment=production
```

Thresholds are variables (`latency_p99_ms`, `error_rate_percent`, `min_request_rate`);
see `variables.tf`.
