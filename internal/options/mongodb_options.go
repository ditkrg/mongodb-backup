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

type MongoDBOptions struct {
	ConnectionString           string    `env:"CONNECTION_STRING,required"`
	DatabaseToBackup           string    `env:"DB_TO_BACKUP,required"`
	BackupOutDir               string    `env:"BACKUP_OUT_DIR,default=/backup"`
	Gzip                       bool      `env:"GZIP,default=true"`
	OpLog                      bool      `env:"OPLOG,default=false"`
	DumpDBUsersAndRoles        bool      `env:"DUMP_DB_USERS_AND_ROLES,default=true"`
	SkipUsersAndRoles          bool      `env:"SKIP_USERS_AND_ROLES,default=false"`
	CollectionToBackup         string    `env:"COLLECTION_TO_BACKUP"`
	ExcludedCollections        []string  `env:"EXCLUDED_COLLECTIONS"`
	ExcludedCollectionPrefixes []string  `env:"EXCLUDED_COLLECTION_PREFIXES"`
	Query                      string    `env:"QUERY"`
	NumParallelCollections     int       `env:"NUM_PARALLEL_COLLECTIONS,default=1"`
	Verbosity                  Verbosity `env:",prefix=VERBOSITY__"`

	MongoDumpOptions  *mongodump.MongoDump
	BackupOutFilePath string
}

type Verbosity struct {
	Quiet bool `env:"QUIET,default=false"`
	Level int  `env:"LEVEL,default=0"`
}

func (o *MongoDBOptions) PrepareMongoDumpOptions() {
	o.BackupOutDir, _ = strings.CutSuffix(o.BackupOutDir, "/")

	o.BackupOutFilePath = fmt.Sprintf("%s/archive_dump_%s.gzip", o.BackupOutDir, time.Now().Format("060102-150405"))

	toolOptions := options.New("mongodb-backup", "", "", "", false, options.EnabledOptions{
		Auth:       true,
		Connection: false,
		Namespace:  false,
		URI:        false,
	})

	toolOptions.ConnectionString = o.ConnectionString
	toolOptions.Verbosity = &options.Verbosity{Quiet: o.Verbosity.Quiet, VLevel: o.Verbosity.Level}
	toolOptions.Namespace = &options.Namespace{DB: o.DatabaseToBackup, Collection: o.CollectionToBackup}
	toolOptions.NormalizeOptionsAndURI()

	inputOptions := &mongodump.InputOptions{
		Query: o.Query,
	}

	outputOptions := &mongodump.OutputOptions{
		Archive:                    o.BackupOutFilePath,
		NumParallelCollections:     o.NumParallelCollections,
		Gzip:                       o.Gzip,
		DumpDBUsersAndRoles:        o.DumpDBUsersAndRoles,
		ExcludedCollections:        o.ExcludedCollections,
		ExcludedCollectionPrefixes: o.ExcludedCollectionPrefixes,
		Oplog:                      o.OpLog,
	}

	o.MongoDumpOptions = &mongodump.MongoDump{
		SkipUsersAndRoles: o.SkipUsersAndRoles,
		ToolOptions:       toolOptions,
		InputOptions:      inputOptions,
		OutputOptions:     outputOptions,
		ProgressManager:   &helpers.ProgressManager{},
	}
}

func (o *MongoDBOptions) Validate() error {

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
