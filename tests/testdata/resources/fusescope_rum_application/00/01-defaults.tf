# Only the required attributes: every optional one takes the server default.
resource "fusescope_rum_application" "defaults" {
  name        = "testdata-rum-defaults"
  platform    = "web"
  environment = "test"
}
