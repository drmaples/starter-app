package objectstore

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
)

type fsStore struct {
	baseDir string
}

func newFS(baseDir string) *fsStore {
	return &fsStore{
		baseDir: baseDir,
	}
}

func (s *fsStore) PutObject(ctx context.Context, key string, obj PutObject) (Object, error) {
	fullPath, err := s.getFullPath(ctx, key)
	if err != nil {
		return Object{}, errors.Wrap(err, "cannot construct object path")
	}

	if err := ensureDir(ctx, s.baseDir, fullPath); err != nil {
		return Object{}, errors.Wrap(err, "cannot create object subdirectories")
	}

	if err := os.WriteFile(fullPath, obj.Data, 0o600); err != nil {
		return Object{}, errors.Wrap(err, "cannot write object data to a file")
	}

	if obj.MetaData != nil {
		b, err := json.Marshal(obj.MetaData)
		if err != nil {
			return Object{}, errors.Wrap(err, "cannot marshal object metadata")
		}

		if err := os.WriteFile(metaFilePath(fullPath), b, 0o600); err != nil {
			return Object{}, errors.Wrap(err, "cannot write object metadata to a file")
		}
	}

	return Object{PutObject: obj, Key: key}, nil
}

func (s *fsStore) GetDownloadURL(_ context.Context, key string) (string, error) {
	return fmt.Sprintf("%[1]s:%[2]d/%[3]s", LocalFileServerBaseURL, LocalFileServerPort, key), nil
}

func (s *fsStore) GetObject(ctx context.Context, key string) (Object, error) {
	if err := checkBaseDirExists(ctx, s.baseDir); err != nil {
		return Object{}, errors.Wrap(err, "error checking object store base directory")
	}

	fullPath, err := s.getFullPath(ctx, key)
	if err != nil {
		return Object{}, errors.Wrap(err, "cannot construct object path")
	}

	var (
		metaData map[string]string
		metaFile = metaFilePath(fullPath)
	)

	if exists, err := checkFileExists(ctx, metaFile); err != nil {
		return Object{}, errors.Wrap(err, "error checking meta file exists")
	} else if exists {
		contents, err := os.ReadFile(metaFile)
		if err != nil {
			return Object{}, errors.Wrap(err, "cannot read meta file contents")
		}
		if err := json.Unmarshal(contents, &metaData); err != nil {
			return Object{}, errors.Wrap(err, "cannot decode meta file contents")
		}
	}

	if exists, err := checkFileExists(ctx, fullPath); err != nil {
		return Object{}, errors.Wrap(err, "error checking data file exists")
	} else if exists {
		contents, err := os.ReadFile(fullPath)
		if err != nil {
			return Object{}, errors.Wrap(err, "cannot read data file contents")
		}
		return Object{
			PutObject: PutObject{
				Data:     contents,
				MetaData: metaData,
			},
			Key: key,
		}, nil
	}

	return Object{}, errors.New("object does not exist")
}

func (s *fsStore) DeleteObject(ctx context.Context, key string) error {
	if err := checkBaseDirExists(ctx, s.baseDir); err != nil {
		return err
	}

	fullPath := path.Join(s.baseDir, key)
	if fullPath == "" {
		return errors.New("cannot construct object path")
	}

	metaFile := metaFilePath(fullPath)
	if err := removeFile(ctx, metaFile); err != nil {
		return errors.Wrap(err, "cannot remove meta file")
	}

	if err := removeFile(ctx, fullPath); err != nil {
		return errors.Wrap(err, "cannot remove file")
	}

	return nil
}

func (s *fsStore) getFullPath(_ context.Context, key string) (string, error) {
	if key == "" {
		return "", errors.New("object key is empty")
	}

	if fullPath := path.Join(s.baseDir, key); fullPath == "" {
		return "", errors.New("cannot construct object path")
	} else {
		return fullPath, nil
	}
}

func removeFile(ctx context.Context, path string) error {
	if exists, err := checkFileExists(ctx, path); err != nil {
		return errors.Wrap(err, "error checking file before deletion")
	} else if exists {
		if err := os.Remove(path); err != nil {
			return errors.Wrap(err, "cannot remove file")
		}
	}
	return nil
}

func checkFileExists(_ context.Context, path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, errors.Wrap(err, "error getting file information")
	}

	if !fileInfo.Mode().IsRegular() {
		return false, errors.New("not a regular file")
	}

	return true, nil
}

func ensureDir(ctx context.Context, baseDir, fileName string) error {
	if err := checkBaseDirExists(ctx, baseDir); err != nil {
		return err
	}
	return os.MkdirAll(filepath.Dir(fileName), 0o755)
}

func checkBaseDirExists(_ context.Context, baseDir string) error {
	fileInfo, err := os.Stat(baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.Wrap(err, "base directory does not exist")
		}
		return errors.Wrap(err, "error getting file information")
	}

	if !fileInfo.IsDir() {
		return errors.New("not a directory")
	}

	return nil
}

func metaFilePath(path string) string {
	return fmt.Sprintf("%s.meta", path)
}

// LocalFileServer serves up files over http
func LocalFileServer(ctx context.Context, directory string) {
	if directory == "" {
		panic("local file server directory is not set")
	}

	fileServer := http.FileServer(http.Dir(directory))
	http.HandleFunc(
		"/",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")

			slog.InfoContext(r.Context(), "serving up file", slog.String("file", r.URL.String()))
			fileServer.ServeHTTP(w, r)
		},
	)

	slog.InfoContext(ctx, "local file server running",
		slog.Int("port", LocalFileServerPort),
		slog.String("directory", directory),
	)

	srv := &http.Server{
		Addr:        fmt.Sprintf(":%[1]d", LocalFileServerPort),
		ReadTimeout: 5 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		slog.ErrorContext(ctx, err.Error())
	}
}
