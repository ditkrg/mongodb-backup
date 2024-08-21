package flags

import (
	"fmt"
	"strings"
	"time"

	"github.com/ditkrg/mongodb-backup/internal/helpers"
	"github.com/mongodb/mongo-tools/common/options"
	"github.com/mongodb/mongo-tools/mongodump"
	"github.com/rs/zerolog/log"
)

type MongoDumpFlags struct {
	ConnectionString           string         `env:"CONNECTION_STRING" required:"" help:"The connection to the MongoDB instance to dump from"`
	Database                   string         `env:"DATABASE" help:"The database to dump"`
	BackupDir                  string         `env:"BACKUP_DIR" default:"/backup" help:"The directory to store the backup"`
	Collection                 string         `env:"COLLECTION" help:"The collection to dump"`
	Query                      string         `env:"QUERY" help:"The query to filter the documents to dump"`
	ExcludedCollections        []string       `env:"EXCLUDED_COLLECTIONS" help:"The collections to exclude from the dump"`
	ExcludedCollectionPrefixes []string       `env:"EXCLUDED_COLLECTION_PREFIXES" help:"The collection prefixes to exclude from the dump"`
	Gzip                       bool           `env:"GZIP" negatable:"" default:"true" help:"Compress the backup using gzip"`
	DumpDBUsersAndRoles        bool           `env:"DUMP_DB_USERS_AND_ROLES" help:"Dump the users and roles in the database"`
	OpLog                      bool           `env:"OPLOG" name:"oplog" help:"take an oplog dump"`
	SkipUsersAndRoles          bool           `env:"SKIP_USERS_AND_ROLES" help:"Skip dumping the users and roles in the database"`
	NumParallelCollections     int            `env:"NUM_PARALLEL_COLLECTIONS" default:"1" help:"The number of collections to dump in parallel"`
	KeepRecentN                int            `env:"KEEP_RECENT_N" default:"10" help:"The number of collections to dump in parallel"`
	Verbosity                  VerbosityFlags `embed:"" prefix:"verbosity-" envprefix:"VERBOSITY__"`
}

func (o *MongoDumpFlags) PrepareMongoDump() (*mongodump.MongoDump, error) {
	log.Info().Msg("Preparing mongodump")

	o.BackupDir = strings.TrimSuffix(o.BackupDir, "/")

	inputOptions := &mongodump.InputOptions{Query: o.Query}

	outputOptions := &mongodump.OutputOptions{
		Archive:                    fmt.Sprintf("%s/dump_%d", o.BackupDir, time.Now().Unix()),
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

	if err := toolOptions.NormalizeOptionsAndURI(); err != nil {
		log.Error().Err(err).Msg("Failed to normalize options and URI")
		return nil, err
	}

	mongodump := &mongodump.MongoDump{
		SkipUsersAndRoles: o.SkipUsersAndRoles,
		ToolOptions:       toolOptions,
		InputOptions:      inputOptions,
		OutputOptions:     outputOptions,
		ProgressManager:   &helpers.ProgressManager{},
	}

	if err := mongodump.ValidateOptions(); err != nil {
		log.Error().Err(err).Msg("Failed to validate mongodump options")
		return nil, err
	}

	log.Info().Msg("Prepared mongodump")
	return mongodump, nil
}
