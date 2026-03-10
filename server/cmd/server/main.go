package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"on-my-interview/server/internal/app"
	"on-my-interview/server/internal/config"
)

const version = "v0.0.1"

func main() {
	cfg := config.LoadFromEnv()
	log.Printf("starting server version=%s", version)

	application, err := app.New(cfg, version)
	if err != nil {
		log.Fatalf("build app: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := application.Run(ctx); err != nil && err != http.ErrServerClosed {
		log.Fatalf("run server: %v", err)
	}
}
