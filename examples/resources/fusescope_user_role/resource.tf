# FuseScope ships the managed roles signoz-admin, signoz-editor and
# signoz-viewer; GET /api/v1/roles lists their IDs.
variable "viewer_role_id" {
  type = string
}

resource "fusescope_user" "engineer" {
  email = "engineer.reader@example.com"
}

resource "fusescope_user_role" "engineer_reader" {
  user_id = fusescope_user.engineer.id
  role_id = var.viewer_role_id
}
