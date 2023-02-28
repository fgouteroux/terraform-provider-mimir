package main

import (
	"flag"

	"github.com/fgouteroux/terraform-provider-mimir/mimir"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

//go:generate ./tools/generate-docs.sh

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary
	version string = "dev"
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{ProviderFunc: mimir.Provider(version), Debug: debugMode, ProviderAddr: "registry.terraform.io/fgouteroux/mimir"}
	plugin.Serve(opts)
}
