package platform

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

const envFile = ".env"

// DBConfig struct for holding db config
type DBConfig struct {
	Host     string `env:"PGHOST,required"`
	Port     int    `env:"PGPORT,required"`
	User     string `env:"PGUSER,required"`
	Password string `env:"PGPASSWORD,required"`
	Name     string `env:"PGDATABASE,required"`
	SSLMode  string `env:"PGSSLMODE" envDefault:"disable"`
	Schema   string `env:"PGSCHEMA" envDefault:"public"`

	ConnMaxIdleTime time.Duration `env:"DB_CONN_MAX_IDLE_TIME" envDefault:"1m"`
	ConnMaxLifeTime time.Duration `env:"DB_CONN_MAX_LIFETIME" envDefault:"5m"`
	MaxOpenConns    int           `env:"DB_MAX_OPEN_CONNS" envDefault:"-1"`
	MaxIdleConns    int           `env:"DB_MAX_IDLE_CONNS" envDefault:"-1"`
}

// Config struct for holding app config
type Config struct {
	DB DBConfig

	GoogleClientID     string `env:"GOOGLE_CLIENT_ID,required"`
	GoogleClientSecret string `env:"GOOGLE_CLIENT_SECRET,required"`

	Environment string `env:"ENVIRONMENT,required"`

	ServerURL     string `env:"SERVER_URL" envDefault:"http://localhost"`
	ServerPort    int    `env:"SERVER_PORT" envDefault:"8000"`
	ServerAddress string `env:"SERVER_ADDRESS,expand" envDefault:"${SERVER_URL}:${SERVER_PORT}"`
	JWTSignKey    string `env:"JWT_SIGN_KEY" envDefault:"my-secret"` // FIXME: do not want default, make required
}

// NewDBConfig creates new db config. used by CMDs that do not need every setting
func NewDBConfig() (DBConfig, error) {
	if err := loadEnv(context.Background()); err != nil {
		return DBConfig{}, err
	}
	var cfg DBConfig
	if err := env.Parse(&cfg); err != nil {
		return DBConfig{}, err
	}
	return cfg, nil
}

// NewConfig creates new app config
func NewConfig() (Config, error) {
	if err := loadEnv(context.Background()); err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// loadEnv loads from .env file. non-local envs will have env vars injected via docker/k8s
func loadEnv(ctx context.Context) error {
	info, err := os.Stat(envFile)
	if os.IsNotExist(err) {
		slog.WarnContext(ctx, "env file not found", slog.String("file", envFile))
		return nil
	}
	if info.IsDir() {
		return errors.Wrap(err, "expected a file, got a directory")
	}

	if err := godotenv.Load(envFile); err != nil {
		return errors.Wrap(err, "problem loading env from file")
	}

	slog.InfoContext(ctx, "successfully loaded env", slog.String("file", envFile))
	return nil
}
