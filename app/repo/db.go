package repo

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib" // postgres drivers
	"github.com/pkg/errors"

	"github.com/drmaples/starter-app/app/platform"
)

// DefaultSchema is default postgres schema where tables live
const DefaultSchema = "public"

var (
	dbConnOnce sync.Once
	dbInst     *sql.DB

	_ Querier = &sql.DB{}   // assert adheres to interface
	_ Querier = &sql.Conn{} // assert adheres to interface
	_ Querier = &sql.Tx{}   // assert adheres to interface
)

// Querier represents a query-able database/sql object: sql.DB, sql.Conn, sql.Tx
type Querier interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// dbConn is a singleton db connection. sql.Open should be called just once
func dbConn() *sql.DB {
	dbConnOnce.Do(func() {
		cfg := platform.DBConfig()
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
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
func Initialize(ctx context.Context) (*sql.DB, error) {
	db := dbConn()
	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Wrap(err, "cannot connect to db")
	}
	return db, nil
}
