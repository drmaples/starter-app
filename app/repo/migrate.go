package repo

import (
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/pkg/errors"

	"github.com/drmaples/starter-app/db"
)

// NewMigrator returns a new db migrator
func NewMigrator(dbConn *sql.DB) (*migrate.Migrate, error) {
	driver, err := postgres.WithInstance(dbConn, &postgres.Config{})
	if err != nil {
		return nil, errors.Wrap(err, "problem getting driver")
	}

	fs, err := iofs.New(db.MigrationFS, db.FileLocation)
	if err != nil {
		return nil, errors.Wrap(err, "problem setting up migration file system")
	}

	m, err := migrate.NewWithInstance("iofs", fs, "postgres", driver)
	if err != nil {
		return nil, errors.Wrap(err, "problem creating migrate object")
	}
	return m, nil
}
