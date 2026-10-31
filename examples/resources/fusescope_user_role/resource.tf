resource "fusescope_user" "engineer" {
  email = "engineer.reader@example.com"
}

resource "fusescope_role" "serviceaccount_reader" {
  name        = "user-role-serviceaccount-reader"
  description = "Read-only access to service accounts"

  transaction_groups = [
    {
      relation = "read"

      object_group = {
        resource = {
          type = "serviceaccount"
          kind = "serviceaccount"
        }

        selectors = ["*"]
      }
    },
  ]
}

resource "fusescope_user_role" "engineer_reader" {
  user_id = fusescope_user.engineer.id
  role_id = fusescope_role.serviceaccount_reader.id
}
