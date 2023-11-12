package repo

import (
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // load postgres drivers
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/pkg/errors"

	"github.com/drmaples/starter-app/db"
)

// NewMigrator returns a new db migrator
func NewMigrator(dsn string) (*migrate.Migrate, error) {
	fs, err := iofs.New(db.MigrationFS, db.FileLocation)
	if err != nil {
		return nil, errors.Wrap(err, "problem setting up migration file system")
	}

	m, err := migrate.NewWithSourceInstance("iofs", fs, dsn)
	if err != nil {
		return nil, errors.Wrap(err, "problem creating migrate object")
	}
	return m, nil
}
