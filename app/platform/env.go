package platform

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/caarlos0/env/v9"
	"github.com/joho/godotenv"
)

const envFile = ".env"

var (
	cfgOnce   sync.Once
	cfg       config
	dbCfgOnce sync.Once
	dbCfg     dbConfig
)

type dbConfig struct {
	DBHost     string `env:"PGHOST,required"`
	DBPort     int    `env:"PGPORT,required"`
	DBUser     string `env:"PGUSER,required"`
	DBPassword string `env:"PGPASSWORD,required"`
	DBName     string `env:"PGDATABASE,required"`

	DBConnMaxIdleTime time.Duration `env:"DB_CONN_MAX_IDLE_TIME" envDefault:"1m"`
	DBConnMaxLifeTime time.Duration `env:"DB_CONN_MAX_LIFETIME" envDefault:"5m"`
	DBMaxOpenConns    int           `env:"DB_MAX_OPEN_CONNS" envDefault:"-1"`
	DBMaxIdleConns    int           `env:"DB_MAX_IDLE_CONNS" envDefault:"-1"`
}

type config struct {
	dbConfig

	GoogleClientID     string `env:"GOOGLE_CLIENT_ID,required"`
	GoogleClientSecret string `env:"GOOGLE_CLIENT_SECRET,required"`

	Environment string `env:"ENVIRONMENT,required"`

	ServerURL  string `env:"SERVER_URL" envDefault:"http://localhost"`
	ServerPort int    `env:"SERVER_PORT" envDefault:"8000"`
	JWTSignKey string `env:"JWT_SIGN_KEY" envDefault:"my-secret"` // FIXME: do not want default, make required
}

// DBConfig is a singleton representing the db config. used by CMDs that do not need every setting
func DBConfig() dbConfig {
	dbCfgOnce.Do(func() {
		loadEnv(context.Background())
		var c dbConfig
		if err := env.Parse(&c); err != nil {
			panic(err)
		}
		dbCfg = c
	})
	return dbCfg
}

// Config is a singleton representing the app config
func Config() config {
	cfgOnce.Do(func() {
		loadEnv(context.Background())
		var c config
		if err := env.Parse(&c); err != nil {
			panic(err)
		}
		cfg = c
	})
	return cfg
}

// loadEnv loads from .env file. non-local envs will have env vars injected via docker/k8s
func loadEnv(ctx context.Context) {
	info, err := os.Stat(envFile)
	if os.IsNotExist(err) {
		slog.WarnContext(ctx, "env file not found", slog.String("file", envFile))
		return
	}
	if info.IsDir() {
		panic(fmt.Sprintf("%s is a dir, expected a file", envFile))
	}

	if err := godotenv.Load(envFile); err != nil {
		panic(err)
	}

	slog.InfoContext(ctx, "successfully loaded env", slog.String("file", envFile))
}
