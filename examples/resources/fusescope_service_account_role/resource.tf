# FuseScope ships the managed roles signoz-admin, signoz-editor and
# signoz-viewer; GET /api/v1/roles lists their IDs.
variable "viewer_role_id" {
  type = string
}

resource "fusescope_service_account" "deployer" {
  name = "deploy-bot"
}

resource "fusescope_service_account_role" "deployer_reader" {
  service_account_id = fusescope_service_account.deployer.id
  role_id            = var.viewer_role_id
}
