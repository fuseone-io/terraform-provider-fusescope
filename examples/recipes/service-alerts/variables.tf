variable "service_name" {
  description = "Value of the service.name resource attribute to alert on."
  type        = string
}

variable "environment" {
  description = "Value of the deployment.environment resource attribute to alert on."
  type        = string
  default     = "production"
}

variable "slack_webhook_url" {
  description = "Slack incoming webhook URL; receives warnings and criticals."
  type        = string
  sensitive   = true
}

variable "slack_channel" {
  type    = string
  default = "#alerts"
}

variable "pagerduty_routing_key" {
  description = "PagerDuty Events API v2 integration key; receives criticals."
  type        = string
  sensitive   = true
}

variable "latency_p99_ms" {
  description = "p99 latency warning and critical thresholds, in milliseconds."
  type = object({
    warning  = number
    critical = number
  })
  default = {
    warning  = 500
    critical = 1000
  }
}

variable "error_rate_percent" {
  description = "Server span error rate that pages, in percent."
  type        = number
  default     = 5
}

variable "min_request_rate" {
  description = "Request rate floor, in requests per second."
  type        = number
  default     = 1
}
