package main

import (
	"log"

	"github.com/nullableocean/grpcservices/orderserviceclient/internal/cli"
)

func main() {
	err := cli.New().Execute()
	if err != nil {
		log.Fatal(err)
	}
}
