package main

import (
	"context"
	"flag"
	"log"

	"github.com/fuseone-io/terraform-provider-fusescope/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary.
	version string = "dev"

	// goreleaser can pass other information to the main package, such as the specific commit
	// https://goreleaser.com/cookbooks/using-main.version/
)

const (
	registry       = "registry.terraform.io/fuseone-io/fusescope"
	terraformAgent = "terraform"
	name           = "fusescope"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: registry,
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(terraformAgent, version, name), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
