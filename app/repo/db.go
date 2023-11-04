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

const (
	// DefaultSchema is default postgres schema where tables live
	DefaultSchema = "public"

	// Driver is the database driver to use
	Driver = "pgx"

	// DSNTemplate is template used to construct db dsn
	DSNTemplate = "host=%[1]s port=%[2]d user=%[3]s password=%[4]s dbname=%[5]s sslmode=disable"
)

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

// dbConn is a singleton db connection since sql.Open should be called once
func dbConn(cfg platform.DBConfig) *sql.DB {
	dbConnOnce.Do(func() {
		dsn := fmt.Sprintf(DSNTemplate, cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)
		db, err := sql.Open(Driver, dsn)
		if err != nil {
			panic(err)
		}

		db.SetMaxOpenConns(cfg.MaxOpenConns)
		db.SetMaxIdleConns(cfg.MaxIdleConns)
		db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
		db.SetConnMaxLifetime(cfg.ConnMaxLifeTime)

		dbInst = db
	})
	return dbInst
}

// Initialize sets up the models layer - db connection
func Initialize(ctx context.Context, cfg platform.DBConfig) (*sql.DB, error) {
	db := dbConn(cfg)
	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Wrap(err, "cannot connect to db")
	}
	return db, nil
}
