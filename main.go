package main

import (
	"context"
	"flag"
	"log"

	"github.com/chmuri/terraform-provider-freeipa/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var (
	// version is populated by GoReleaser. Default is 1.1.1.
	version string = "1.1.1"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/chmuri/freeipa",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
