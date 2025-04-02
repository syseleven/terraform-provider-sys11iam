package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"gitlab.syseleven.de/ncs/terraform-provider-ncs/internal/provider"
)

func main() {
	opts := providerserver.ServeOpts{
		Address: "hashicorp.com/syseleven/ncs",
	}

	err := providerserver.Serve(context.Background(), provider.New(), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
