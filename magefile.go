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
	// swag is not in $PATH by default. https://github.com/swaggo/swag
	path, err := sh.Output("go", "env", "GOPATH")
	if err != nil {
		return err
	}
	swag_bin := fmt.Sprintf("%s/bin/swag", path)
	if err := sh.RunV(swag_bin, "--version"); err != nil {
		return err
	}

	return sh.RunV(swag_bin, "init", "--generalInfo", "app/controller/main.go")
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
