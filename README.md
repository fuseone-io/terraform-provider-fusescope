<p align="center">
  <img alt="FuseScope" src=".github/brand/fusescope-logo-universal.svg" width="440">
</p>

# Terraform Provider for FuseScope

Manage [FuseScope](https://github.com/fuseone-io/fusescope) as code with
[Terraform](https://terraform.io) or [OpenTofu](https://opentofu.org).

| Area | Resources (`fusescope_*`) |
|---|---|
| Alerting | `rule`, `notification_channel`, `route_policy`, `planned_maintenance` |
| Dashboards | `dashboard`, `public_dashboard`, `saved_view` |
| Real User Monitoring | `rum_application` |
| Access | `user`, `user_role`, `service_account`, `service_account_role`, `auth_domain` (SSO) |

Each resource has a matching data source, and `role` is a data source only. Reference:
[`docs/`](docs/).

## Usage

```terraform
terraform {
  required_providers {
    fusescope = {
      source = "fuseone-io/fusescope"
    }
  }
}

provider "fusescope" {
  endpoint = "https://fusescope.example.com" # or FUSESCOPE_ENDPOINT
}

resource "fusescope_rum_application" "shop" {
  name            = "shop-web"
  platform        = "web"
  allowed_origins = ["https://shop.example.com"]
}

output "shop_rum_key" {
  value     = fusescope_rum_application.shop.key
  sensitive = true
}
```

Authenticate with the API key of a service account whose role can manage the declared
resources. Set it in `FUSESCOPE_ACCESS_TOKEN` rather than in configuration.
`FUSESCOPE_HTTP_MAX_RETRY` and `FUSESCOPE_HTTP_TIMEOUT` tune the HTTP client.

Requires Terraform ≥ 1.4 or OpenTofu ≥ 1.6. Older Terraform versions show permanent drift on
nested attributes that the server fills in.

## Development

Go ≥ 1.25.

```shell
go install          # build into $GOPATH/bin
go test ./...
go generate ./...   # regenerate docs/ from the schema, templates/ and examples/
```

`docs/` is generated, so never edit it by hand. CI fails when it is out of date. The
integration suite runs every example and scenario against a live FuseScope instance; see
[`tests/README.md`](tests/README.md).

## License

[Mozilla Public License 2.0](LICENSE). This is FuseOne's independent fork of the
[Terraform SigNoz Provider](https://github.com/SigNoz/terraform-provider-signoz) and does
not track upstream; attribution is in [NOTICE](NOTICE).
