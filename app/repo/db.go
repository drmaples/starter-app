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

// Querier represents a query-able database/sql object: sql.Tx, sql.DB, sql.Stmt
type Querier interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type config struct {
	Host     string `env:"PGHOST,required"`
	Port     int    `env:"PGPORT,required"`
	User     string `env:"PGUSER,required"`
	Password string `env:"PGPASSWORD,required"`
	Database string `env:"PGDATABASE,required"`
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
