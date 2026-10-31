terraform {
  required_providers {
    fusescope = {
      source = "fuseone-io/fusescope"
    }
  }
}

# Reads FUSESCOPE_ENDPOINT and FUSESCOPE_ACCESS_TOKEN from the environment.
provider "fusescope" {}
