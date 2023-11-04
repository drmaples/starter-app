package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/drmaples/starter-app/app/controller"
	"github.com/drmaples/starter-app/app/platform"
	"github.com/drmaples/starter-app/app/repo"
)

func main() {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	platform.Config() // ensure env vars exist

	if err := repo.Initialize(ctx); err != nil {
		panic(err)
	}

	controller.NewController().Run(ctx)
}
