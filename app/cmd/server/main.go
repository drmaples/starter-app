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

	platform.LoadEnv(ctx)

	if err := repo.Initialize(ctx); err != nil {
		panic(err)
	}

	e := controller.Initialize()
	slog.InfoContext(ctx, "starting server",
		slog.String("address", controller.GetServerAddress()),
	)
	e.Logger.Fatal(e.Start(controller.GetServerBindAddress()))
}
