package repo

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib" // postgres drivers
	"github.com/pkg/errors"
)

var (
	dbConnOnce sync.Once
	dbInst     *sql.DB
)

// DBConn is a singleton db connection
func DBConn() *sql.DB {
	dbConnOnce.Do(func() {
		port, err := strconv.Atoi(os.Getenv("PGPORT"))
		if err != nil {
			panic(err)
		}
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			os.Getenv("PGHOST"),
			port,
			os.Getenv("PGUSER"),
			os.Getenv("PGPASSWORD"),
			os.Getenv("PGDATABASE"),
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
