package flags

import (
	"fmt"
	"strings"
	"time"

	"github.com/ditkrg/mongodb-backup/internal/helpers"
	"github.com/mongodb/mongo-tools/common/options"
	"github.com/mongodb/mongo-tools/mongodump"
	"github.com/rs/zerolog/log"

	mongoLog "github.com/mongodb/mongo-tools/common/log"
)

type MongoDumpFlags struct {
	ConnectionString string `env:"CONNECTION_STRING" required:"" help:"The connection to the MongoDB instance to dump from"`
	BackupDir        string `env:"BACKUP_DIR" default:"/backup" help:"The directory to store the backup"`

	NamespaceOptions struct {
		Database   string `env:"DATABASE" help:"The database to dump"`
		Collection string `env:"COLLECTION" help:"The collection to dump"`
	} `embed:"" group:"namespace options"`

	QueryOptions struct {
		Query          string `env:"QUERY" help:"The query to filter the documents to dump"`
		QueryFile      string `env:"QUERY_FILE" help:"Path to a file containing a query filter (v2 Extended JSON)"`
		ReadPreference string `env:"READ_PREFERENCE" help:"specify either a preference mode (e.g. 'nearest') or a preference json object"`
	} `embed:"" group:"query options"`

	OutputOptions struct {
		Gzip                       bool     `env:"GZIP" negatable:"" default:"true" help:"Compress the backup using gzip"`
		OpLog                      bool     `env:"OPLOG" name:"oplog" help:"take an oplog dump"`
		DumpDBUsersAndRoles        bool     `env:"DUMP_DB_USERS_AND_ROLES" help:"Dump the users and roles in the database"`
		SkipUsersAndRoles          bool     `env:"SKIP_USERS_AND_ROLES" help:"Skip dumping the users and roles in the database"`
		ExcludedCollections        []string `env:"EXCLUDED_COLLECTIONS" help:"The collections to exclude from the dump"`
		ExcludedCollectionPrefixes []string `env:"EXCLUDED_COLLECTION_PREFIXES" help:"The collection prefixes to exclude from the dump"`
		NumParallelCollections     int      `env:"NUM_PARALLEL_COLLECTIONS" default:"1" help:"The number of collections to dump in parallel"`
	} `embed:"" group:"output options"`

	KeepRecentN int            `env:"KEEP_RECENT_N" default:"10" help:"The number of collections to dump in parallel"`
	Verbosity   VerbosityFlags `embed:"" prefix:"verbosity-" envprefix:"VERBOSITY__"`
}

func (o *MongoDumpFlags) PrepareMongoDump() (*mongodump.MongoDump, error) {
	log.Info().Msg("Preparing mongodump")

	mongoLog.SetVerbosity(options.Verbosity{
		VLevel: o.Verbosity.Level,
		Quiet:  o.Verbosity.Quiet,
	})

	o.BackupDir = strings.TrimSuffix(o.BackupDir, "/")

	inputOptions := &mongodump.InputOptions{
		Query:          o.QueryOptions.Query,
		QueryFile:      o.QueryOptions.QueryFile,
		ReadPreference: o.QueryOptions.ReadPreference,
	}

	outputOptions := &mongodump.OutputOptions{
		Archive:                    fmt.Sprintf("%s/dump_%d", o.BackupDir, time.Now().Unix()),
		NumParallelCollections:     o.OutputOptions.NumParallelCollections,
		Gzip:                       o.OutputOptions.Gzip,
		DumpDBUsersAndRoles:        o.OutputOptions.DumpDBUsersAndRoles,
		ExcludedCollections:        o.OutputOptions.ExcludedCollections,
		ExcludedCollectionPrefixes: o.OutputOptions.ExcludedCollectionPrefixes,
	}

	toolOptions := options.New("mongodb-backup", "", "", "", false, options.EnabledOptions{Auth: true})
	toolOptions.ConnectionString = o.ConnectionString
	toolOptions.Namespace = &options.Namespace{DB: o.NamespaceOptions.Database, Collection: o.NamespaceOptions.Collection}

	if o.OutputOptions.OpLog {
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
		SkipUsersAndRoles: o.OutputOptions.SkipUsersAndRoles,
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
