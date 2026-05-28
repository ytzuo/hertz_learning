package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"Hertz/bootstrap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	worker := bootstrap.NewWorker()
	if err := worker.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
