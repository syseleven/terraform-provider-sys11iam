package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/syseleven/terraform-provider-sys11iam/internal/provider"
)

func main() {
	opts := providerserver.ServeOpts{
		Address: "syseleven/sys11iam",
	}

	err := providerserver.Serve(context.Background(), provider.New(), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
