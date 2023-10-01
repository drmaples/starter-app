package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"

	"github.com/drmaples/starter-app/app/controller"
	"github.com/drmaples/starter-app/app/models"
)

const (
	serverAddress = ":8000"
	envFile       = ".env"
)

// loadEnv loads from .env file. non-local envs will have env vars injected via docker/k8s
func loadEnv(ctx context.Context) {
	info, err := os.Stat(envFile)
	if os.IsNotExist(err) {
		slog.WarnContext(ctx, "env file not found", slog.String("file", envFile))
		return
	}
	if info.IsDir() {
		panic(fmt.Sprintf("%s is a dir, expected a file", envFile))
	}

	if err := godotenv.Load(envFile); err != nil {
		panic(err)
	}

	slog.InfoContext(ctx, "successfully loaded env file", slog.String("file", envFile))
}

func main() {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	loadEnv(ctx)

	if err := models.Initialize(ctx); err != nil {
		panic(err)
	}

	e := controller.Initialize()
	slog.InfoContext(ctx, "starting server",
		slog.String("address", serverAddress),
	)
	e.Logger.Fatal(e.Start(serverAddress))
}
