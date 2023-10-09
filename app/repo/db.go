package repo

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/caarlos0/env/v9"
	_ "github.com/jackc/pgx/v5/stdlib" // postgres drivers
	"github.com/pkg/errors"
)

var (
	dbConnOnce sync.Once
	dbInst     *sql.DB
)

type config struct {
	Host     string `env:"PGHOST"`
	Port     int    `env:"PGPORT"`
	User     string `env:"PGUSER"`
	Password string `env:"PGPASSWORD"`
	Database string `env:"PGDATABASE"`
}

// DBConn is a singleton db connection
func DBConn() *sql.DB {
	dbConnOnce.Do(func() {
		var cfg config
		if err := env.Parse(&cfg); err != nil {
			panic(err)
		}

		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database,
		)
		db, err := sql.Open("pgx", dsn)
		if err != nil {
			panic(err)
		}

		// FIXME: do not forget to adjust these
		// db.SetMaxOpenConns(...)
		// db.SetMaxIdleConns(...)
		// db.SetConnMaxIdleTime(...)
		// db.SetConnMaxLifetime(...)

		dbInst = db
	})
	return dbInst
}

// Initialize sets up the models layer - db connection
func Initialize(ctx context.Context) error {
	if err := DBConn().PingContext(ctx); err != nil {
		return errors.Wrap(err, "cannot connect to db")
	}
	return nil
}
