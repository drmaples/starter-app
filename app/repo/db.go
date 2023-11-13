package repo

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
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
		dsn := GetConnectionURI(cfg)
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

// GetConnectionURI returns the connection URI for a given schema
// https://www.postgresql.org/docs/14/libpq-connect.html#LIBPQ-CONNSTRING
func GetConnectionURI(cfg platform.DBConfig) string {
	p := url.Values{}
	p.Add("sslmode", cfg.SSLMode)
	p.Add("search_path", cfg.Schema)

	u := url.URL{
		Scheme:   "postgresql",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Path:     cfg.Name,
		RawQuery: p.Encode(),
	}
	return u.String()
}

// Initialize sets up the models layer - db connection
func Initialize(ctx context.Context, cfg platform.DBConfig) (*sql.DB, error) {
	db := dbConn(cfg)
	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Wrap(err, "cannot connect to db")
	}
	return db, nil
}
