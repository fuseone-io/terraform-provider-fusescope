resource "fusescope_rum_application" "shop" {
  name        = "shop-web"
  platform    = "web"
  environment = "production"

  allowed_origins        = ["https://shop.example.com", "https://*.shop.example.com"]
  trace_propagation_urls = ["https://api.example.com"]
  sample_rate            = 0.5

  pii_policy = {
    query_allowlist = ["page", "sort"]
    require_consent = true
  }
}

# The public key goes into the browser SDK configuration.
output "shop_rum_key" {
  value     = fusescope_rum_application.shop.key
  sensitive = true
}
