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

	cfg, err := platform.NewConfig()
	if err != nil {
		panic(err)
	}
	dbConn, err := repo.Initialize(ctx, cfg.DB)
	if err != nil {
		panic(err)
	}

	con := controller.New(dbConn, cfg, repo.NewUserRepo())
	con.Run(ctx)
}
