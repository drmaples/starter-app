package db

import (
	"embed"
	"regexp"
)

//go:embed *.sql
var MigrationFS embed.FS

// PathRE is regex of a migration file
var PathRE = regexp.MustCompile(`(\d+)_migration.*`)

// FileLocation is location where all migration files live
const FileLocation = "."
