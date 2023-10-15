//go:build mage

package main

import (
	"fmt"
	"os"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type (
	Run mg.Namespace
)

func init() {
	os.Setenv("MAGEFILE_VERBOSE", "true")
	os.Setenv("CGO_ENABLED", "0")
}

func (Run) Server() error {
	return sh.RunV("go", "run", "app/cmd/server/main.go")
}

func (Run) Db() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	return sh.RunV(
		"docker",
		"run",
		"--name", "darrell_db",
		"--rm",
		"--interactive",
		"--tty",
		// "--detach",
		"--env", "POSTGRES_USER=postgres",
		"--env", "POSTGRES_PASSWORD=postgres",
		"--env", "POSTGRES_DB=darrell",
		"--env", "PGDATA=/var/lib/postgresql/data",
		"--publish", "15432:5432",
		"--volume", fmt.Sprintf("%s/.pg/data:/var/lib/postgresql/data", pwd),
		"postgres:15.2",
	)
}

/*
docker run -it --rm \
	--env-file=.env \
	-e "PGHOST=host.docker.internal" \
	postgres:15.2 psql


docker run -it --rm \
	-e "PGHOST=host.docker.internal" \
	-e "PGPORT=15432" \
	-e "PGUSER=postgres" \
	-e "PGPASSWORD=postgres" \
	-e "PGDATABASE=darrell" \
	postgres:15.2 psql
*/
