package db

import (
	"embed"
	"io/fs"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"

	"github.com/pkg/errors"
	"github.com/samber/lo"
)

//go:embed *.sql
var MigrationFS embed.FS

// pathRE is regex of a migration file
var pathRE = regexp.MustCompile(`(\d+)_migration.*`)

// FileLocation is location where all migration files live
const FileLocation = "."

// LatestMigrationVersion returns the latest available migration as found on the filesystem
func LatestMigrationVersion() (int, error) {
	paths, err := MigrationFS.ReadDir(FileLocation)
	if err != nil {
		return 0, errors.Wrap(err, "problem listing paths")
	}

	paths = lo.Filter(paths, func(item fs.DirEntry, _ int) bool {
		return pathRE.Match([]byte(item.Name()))
	})
	sort.Slice(paths, func(i, j int) bool {
		return paths[i].Name() < paths[j].Name()
	})

	latestPath := pathRE.FindSubmatch([]byte(paths[len(paths)-1].Name()))
	latestVersion, err := strconv.Atoi(string(latestPath[1]))
	if err != nil {
		return 0, err
	}

	return latestVersion, nil
}

// GetDockerInitPath returns the path to the db init file for use with docker
// CWD moves around depending on how tests are run and this provides stable path
func GetDockerInitPath() string {
	_, thisFilePath, _, _ := runtime.Caller(0)
	thisDirPath := filepath.Dir(thisFilePath)
	return filepath.Join(thisDirPath, "docker_init.sql")
}
