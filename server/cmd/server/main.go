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

func main() {
	cfg := config.LoadFromEnv()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("build app: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := application.Run(ctx); err != nil && err != http.ErrServerClosed {
		log.Fatalf("run server: %v", err)
	}
}
