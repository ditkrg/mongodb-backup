package options

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ditkrg/mongodb-backup/internal/helpers"

	"github.com/mongodb/mongo-tools/common/options"
	"github.com/mongodb/mongo-tools/mongodump"
)

type MongoDumpOptions struct {
	ConnectionString           string    `env:"CONNECTION_STRING"`
	Database                   string    `env:"DATABASE"`
	BackupDir                  string    `env:"BACKUP_DIR,default=/backup"`
	Gzip                       bool      `env:"GZIP,default=true"`
	OpLog                      bool      `env:"OPLOG,default=false"`
	DumpDBUsersAndRoles        bool      `env:"DUMP_DB_USERS_AND_ROLES,default=true"`
	SkipUsersAndRoles          bool      `env:"SKIP_USERS_AND_ROLES,default=false"`
	Collection                 string    `env:"COLLECTION"`
	ExcludedCollections        []string  `env:"EXCLUDED_COLLECTIONS"`
	ExcludedCollectionPrefixes []string  `env:"EXCLUDED_COLLECTION_PREFIXES"`
	Query                      string    `env:"QUERY"`
	NumParallelCollections     int       `env:"NUM_PARALLEL_COLLECTIONS,default=1"`
	Verbosity                  Verbosity `env:",prefix=VERBOSITY__"`

	MongoDumpOptions *mongodump.MongoDump
}

func (o *MongoDumpOptions) PrepareMongoDumpOptions() {
	o.BackupDir, _ = strings.CutSuffix(o.BackupDir, "/")

	inputOptions := &mongodump.InputOptions{Query: o.Query}

	outputOptions := &mongodump.OutputOptions{
		Archive:                    fmt.Sprintf("%s/archive_dump_%s.gzip", o.BackupDir, time.Now().Format("060102-150405")),
		NumParallelCollections:     o.NumParallelCollections,
		Gzip:                       o.Gzip,
		DumpDBUsersAndRoles:        o.DumpDBUsersAndRoles,
		ExcludedCollections:        o.ExcludedCollections,
		ExcludedCollectionPrefixes: o.ExcludedCollectionPrefixes,
	}

	toolOptions := options.New("mongodb-backup", "", "", "", false, options.EnabledOptions{Auth: true})
	toolOptions.ConnectionString = o.ConnectionString
	toolOptions.Verbosity = &options.Verbosity{Quiet: o.Verbosity.Quiet, VLevel: o.Verbosity.Level}
	toolOptions.Namespace = &options.Namespace{DB: o.Database, Collection: o.Collection}

	if o.OpLog {
		outputOptions.Archive = ""
		outputOptions.Out = o.BackupDir
		outputOptions.DumpDBUsersAndRoles = false
		toolOptions.Namespace = &options.Namespace{DB: "local", Collection: "oplog.rs"}
	}

	toolOptions.NormalizeOptionsAndURI()

	o.MongoDumpOptions = &mongodump.MongoDump{
		SkipUsersAndRoles: o.SkipUsersAndRoles,
		ToolOptions:       toolOptions,
		InputOptions:      inputOptions,
		OutputOptions:     outputOptions,
		ProgressManager:   &helpers.ProgressManager{},
	}
}

func (o *MongoDumpOptions) Validate() error {

	if o.Verbosity.Level < 0 || o.Verbosity.Level > 5 {
		return errors.New("verbosity level must be between 0 and 5")
	}

	if o.NumParallelCollections < 1 {
		return errors.New("numParallelCollections must be greater than 0")
	}

	if err := o.MongoDumpOptions.ValidateOptions(); err != nil {
		return err
	}

	return nil
}
