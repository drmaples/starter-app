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
	return sh.RunV("go", "run", "app/cmd/main.go")
}

func (Run) Database() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	return sh.RunV(
		"docker",
		"run",
		"--name", "one_advisory_db",
		"--rm",
		"--interactive",
		"--tty",
		// "--detach",
		"--env", "POSTGRES_USER=postgres",
		"--env", "POSTGRES_PASSWORD=postgres",
		"--env", "POSTGRES_DB=one_advisory",
		"--env", "PGDATA=/var/lib/postgresql/data",
		"--publish", "15432:5432",
		"--volume", fmt.Sprintf("%s/.pg/data:/var/lib/postgresql/data", pwd),
		"--volume", fmt.Sprintf("%s/db/local/init.sql:/docker-entrypoint-initdb.d/init.sql", pwd),
		"postgres:15.2",
	)
}
