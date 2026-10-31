provider "fusescope" {
  # Root URL of your FuseScope instance. Can also be set with the
  # FUSESCOPE_ENDPOINT environment variable.
  endpoint = "http://localhost:8080"

  # API key of a service account with the Admin role. Prefer the
  # FUSESCOPE_ACCESS_TOKEN environment variable to keep the secret out of
  # configuration and state.
  access_token = var.fusescope_access_token
}
