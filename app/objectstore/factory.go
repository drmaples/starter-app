package objectstore

import (
	"context"
	"log/slog"
	"os"
	"path"

	"github.com/pkg/errors"
)

type backend string

const (
	s3BackendType backend = "s3"
	fsBackendType backend = "fs"

	objectStoreProviderKey = "OBJECT_STORE_PROVIDER"

	// LocalFileServerPort port that local http file server runs on
	LocalFileServerPort = 9000

	// LocalFileServerBaseURL is used for download url of the file, to mimic s3 behavior
	LocalFileServerBaseURL = "http://localhost"
)

// New constructs a new object store
func New(ctx context.Context, bucket string) (Store, error) {
	if bucket == "" {
		return nil, errors.New("bucket cannot be empty")
	}

	be := os.Getenv(objectStoreProviderKey)
	switch be {
	case string(s3BackendType):
		return nil, errors.New("s3 not yet supported")
	case string(fsBackendType):
		cwd, err := os.Getwd()
		if err != nil {
			return nil, errors.New("failed to get current working dir")
		}
		baseDir := path.Join(cwd, bucket)
		if _, err := os.Stat(baseDir); os.IsNotExist(err) {
			slog.InfoContext(ctx, "creating local fs dir", slog.String("dir", bucket))
			if err := os.Mkdir(bucket, 0o755); err != nil {
				return nil, errors.New("problem creating local fs dir")
			}
		}
		return newFS(baseDir), nil
	}

	return nil, errors.New("unsupported object store backend")
}
