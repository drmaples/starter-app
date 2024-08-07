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
	Gen mg.Namespace
)

func init() {
	os.Setenv("MAGEFILE_VERBOSE", "true")
	os.Setenv("CGO_ENABLED", "0")
}

func (Run) Server() error {
	return sh.RunV("go", "run", "app/cmd/server/main.go")
}

func (Run) Db() error {
	return sh.RunV("docker-compose", "up", "--force-recreate")
}

func (Gen) Swagger() error {
	// install swag via asdf
	if err := sh.RunV("swag", "--version"); err != nil {
		return err
	}

	return sh.RunV("swag", "init", "--generalInfo", "app/controller/main.go")
}

// see https://github.com/KarnerTh/mermerd
func (Gen) Erd() error {
	// what is an erd? https://www.databasestar.com/entity-relationship-diagram/

	if err := sh.RunV("go", "install", "github.com/KarnerTh/mermerd@v0.11.0"); err != nil {
		return err
	}

	basePath, err := sh.Output("go", "env", "GOPATH")
	if err != nil {
		return err
	}

	exe := fmt.Sprintf("%s/bin/mermerd", basePath)
	if err := sh.RunV(exe, "version"); err != nil {
		return err
	}

	return sh.RunV(exe, "--runConfig", "mermerd.yml")
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
