package main

import (
	"context"
	"log"

	"github.com/nullableocean/grpcservices/spotinstrument/internal/app/fxrunner"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fxApp, err := fxrunner.FxAppRunner()
	if err != nil {
		log.Fatalf("failed to build app: %v", err)
	}

	if err := fxApp.Start(ctx); err != nil {
		log.Fatalf("failed to start app: %v", err)
	}

	<-fxApp.Wait()

	if err := fxApp.Stop(ctx); err != nil {
		log.Fatalf("failed to stop app: %v", err)
	}
}
