package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"github.com/drmaples/starter-app/app/platform"
	"github.com/drmaples/starter-app/app/repo"
	"github.com/drmaples/starter-app/db"
)

// FIXME: fetching from env vars seems useful
// https://cli.urfave.org/v2/examples/flags/#values-from-the-environment

func rootCmd() *cli.App {
	root := &cli.App{
		Name:  "migrate",
		Usage: "cli for managing db migrations",
		Commands: []*cli.Command{
			currentCmd(),
			runCmd(),
		},
	}
	return root
}

func getMigrator() (*migrate.Migrate, error) {
	cfg, err := platform.NewDBConfig()
	if err != nil {
		return nil, err
	}
	return repo.NewMigrator(repo.GetConnectionURI(cfg))
}

func currentCmd() *cli.Command {
	cmd := &cli.Command{
		Name:  "current",
		Usage: "list current and latest db migration version",
		Action: func(cCtx *cli.Context) error {
			latestVersion, err := db.LatestMigrationVersion()
			if err != nil {
				return err
			}

			m, err := getMigrator()
			if err != nil {
				return errors.Wrap(err, "problem creating migrator")
			}

			currentVersion, dirty, err := m.Version()
			if err != nil {
				if !errors.Is(err, migrate.ErrNilVersion) {
					return errors.Wrap(err, "problem getting migration version")
				}
				slog.WarnContext(cCtx.Context, "no database migrations have ever been applied")
			}

			slog.InfoContext(cCtx.Context, "migration information",
				slog.Int("latest", latestVersion),
				slog.Int("current", int(currentVersion)),
				slog.Bool("dirty", dirty),
			)
			return nil
		},
	}
	return cmd
}

func runCmd() *cli.Command {
	cmd := &cli.Command{
		Name:  "run",
		Usage: "run up/down migration to specified version. no args runs up to latest",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "n",
				Aliases: []string{"number"},
				Value:   -1,
				Usage:   "migration to specific version or latest",
			},
		},
		Action: func(cCtx *cli.Context) error {
			m, err := getMigrator()
			if err != nil {
				return errors.Wrap(err, "problem creating migrator")
			}

			version := cCtx.Int("number")
			if version < 0 {
				slog.InfoContext(cCtx.Context, "migrating to latest version")
				err = m.Up()
			} else {
				slog.InfoContext(cCtx.Context, "migrating to specific version", slog.Int("version", version))
				err = m.Migrate(uint(version))
			}

			if err != nil {
				if errors.Is(err, migrate.ErrNoChange) {
					slog.WarnContext(cCtx.Context, "no changes in migration")
					return nil
				}
				return errors.Wrap(err, "problem running migrating")
			}

			slog.InfoContext(cCtx.Context, "migration successful")
			return nil
		},
	}
	return cmd
}

func main() {
	ctx := context.Background()
	if err := rootCmd().RunContext(ctx, os.Args); err != nil {
		panic(fmt.Sprintf("%+v", err))
	}
}
