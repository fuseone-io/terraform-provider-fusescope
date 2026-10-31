resource "fusescope_rum_application" "android" {
  name                  = "shop-android"
  platform              = "android"
  environment           = "production"
  allowed_package_names = ["com.example.shop"]

  source_maps = {
    retention = "1y"
  }
}
