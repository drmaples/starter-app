package test_repo

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/pkg/errors"

	"github.com/drmaples/starter-app/app/platform"
	"github.com/drmaples/starter-app/app/repo"
)

const (
	pgImage = "postgres:15.2" // keep in sync with docker-compose.yml
	pgUser  = "postgres"
	pgPass  = "postgres"
	pgDB    = "test"
	pgSSL   = "disable"

	dockerPortName = "5432/tcp"

	maxRetryWait = 30
	hardKillSecs = 600 // single suite of tests should never take longer than this
)

// IPostgresContainer is the interface
type IPostgresContainer interface {
	GetDB() *sql.DB
	Setup() error
	TearDown() error
}

type postgresContainer struct {
	pool     *dockertest.Pool
	resource *dockertest.Resource
	db       *sql.DB
}

// NewPostgresContainer creates new postgres docker container for use with integration tests
func NewPostgresContainer() IPostgresContainer {
	return &postgresContainer{}
}

func (c *postgresContainer) GetDB() *sql.DB {
	return c.db
}

func (c *postgresContainer) Setup() error {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return errors.Wrap(err, "error creating docker test pool")
	}
	c.pool = pool

	if err := pool.Client.Ping(); err != nil {
		return errors.Wrap(err, "error connecting to docker client")
	}

	nuggets := strings.Split(pgImage, ":")
	resource, err := pool.RunWithOptions(
		&dockertest.RunOptions{
			Repository: nuggets[0],
			Tag:        nuggets[1],
			Env: []string{
				fmt.Sprintf("POSTGRES_USER=%s", pgUser),
				fmt.Sprintf("POSTGRES_PASSWORD=%s", pgPass),
				fmt.Sprintf("POSTGRES_DB=%s", pgDB),
			},
		},
		func(config *docker.HostConfig) {
			config.AutoRemove = true // makes stopped container go away by itself
			config.RestartPolicy = docker.RestartPolicy{Name: "no"}
		})
	if err != nil {
		return errors.Wrap(err, "error running container")
	}
	// tell docker to hard kill the container in N seconds
	if err := resource.Expire(hardKillSecs); err != nil {
		return err
	}
	c.resource = resource
	slog.Info("docker test container created", slog.String("name", resource.Container.Name), slog.String("image", resource.Container.Config.Image))

	port, err := strconv.Atoi(resource.GetPort(dockerPortName))
	if err != nil {
		return err
	}
	dsn := repo.GetConnectionURI(platform.DBConfig{
		Host:     resource.GetBoundIP(dockerPortName),
		Port:     port,
		User:     pgUser,
		Password: pgPass,
		Name:     pgDB,
		SSLMode:  pgSSL,
	})
	slog.Info("connecting to test database", slog.String("dsn", dsn))

	var db *sql.DB
	pool.MaxWait = maxRetryWait * time.Second
	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		db, err = sql.Open(repo.Driver, dsn)
		if err != nil {
			return errors.Wrap(err, "error connecting to database")
		}
		if err := db.Ping(); err != nil {
			return errors.Wrap(err, "error pinging database")
		}
		m, err := repo.NewMigrator(dsn)
		if err != nil {
			return errors.Wrap(err, "error creating db migrator")
		}
		if err := m.Up(); err != nil {
			slog.Error("problem running migrations", slog.Any("error", err))
			return err
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "error pinging database")
	}
	c.db = db

	return nil
}

func (c *postgresContainer) TearDown() error {
	slog.Info("purging docker container")
	if err := c.pool.Purge(c.resource); err != nil {
		return errors.Wrap(err, "could not purge docker container")
	}
	return nil
}
