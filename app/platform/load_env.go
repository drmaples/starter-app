package platform

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

const envFile = ".env"

// LoadEnv loads from .env file. non-local envs will have env vars injected via docker/k8s
func LoadEnv(ctx context.Context) {
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
