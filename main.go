package main

import (
	"context"
	"flag"

	"github.com/fgouteroux/terraform-provider-mimir/mimir"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	debugFlag := flag.Bool("debug", false, "Start provider in stand-alone debug mode.")
	flag.Parse()

	serveOpts := &plugin.ServeOpts{
		ProviderFunc: mimir.Provider,
	}
	if debugFlag != nil && *debugFlag {
		plugin.Debug(context.Background(), "registry.terraform.io/fgouteroux/mimir", serveOpts)
	} else {
		plugin.Serve(serveOpts)
	}
}
