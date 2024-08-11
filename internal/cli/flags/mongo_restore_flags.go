package flags

import (
	"github.com/mongodb/mongo-tools/common/options"
	"github.com/mongodb/mongo-tools/mongorestore"
	"github.com/rs/zerolog/log"
)

type MongoRestoreFlags struct {
	ConnectionString         string         `required:"" env:"CONNECTION_STRING" help:"The connection to the MongoDB instance to restore to"`
	BackupDir                string         `required:"" env:"BACKUP_DIR" help:"The directory to download the backup to and restore from"`
	WriteConcern             string         `env:"WRITE_CONCERN" default:"majority" help:"Write concern for the restore operation"`
	OplogLimit               string         `env:"OPLOG_LIMIT" help:"only include oplog entries before the provided Timestamp"`
	NSExclude                []string       `env:"NS_EXCLUDE" help:"Namespaces (database.collection) to exclude from the restore"`
	NSInclude                []string       `env:"NS_INCLUDE" help:"Namespaces (database.collection) to include in the restore"`
	NumParallelCollections   int            `env:"NUM_PARALLEL_COLLECTIONS" default:"1" help:"Number of collections to restore in parallel"`
	NumInsertionWorkers      int            `env:"NUM_INSERTION_WORKERS" default:"1" help:"Number of insert operations to run concurrently per collection"`
	Gzip                     bool           `env:"GZIP" default:"true" help:"Whether the backup is gzipped (Default: true)"`
	RestoreDBUsersAndRoles   bool           `env:"RESTORE_DB_USERS_AND_ROLES" default:"true" help:"Testore user and role definitions for the given database (Default: true)"`
	SkipUsersAndRoles        bool           `env:"SKIP_USERS_AND_ROLES" default:"false" help:"Skip restoring users and roles, regardless of namespace, when true (Default: false)"`
	ObjectCheck              bool           `env:"OBJECT_CHECK" default:"true" help:"validate all objects before inserting (Default: true)"`
	Drop                     bool           `env:"DROP" default:"false" help:"Drop each collection before import (Default: false)"`
	DryRun                   bool           `env:"DRY_RUN" default:"false" help:"Run the restore in 'dry run' mode (Default: false)"`
	NoIndexRestore           bool           `env:"NO_INDEX_RESTORE" default:"false" help:"Don't restore indexes (Default: false)"`
	ConvertLegacyIndexes     bool           `env:"CONVERT_LEGACY_INDEXES" default:"false" help:"Removes invalid index options and rewrites legacy option values (e.g. true becomes 1) (Default: false)"`
	NoOptionsRestore         bool           `env:"NO_OPTIONS_RESTORE" default:"false" help:"Don't restore collection options (Default: false)"`
	KeepIndexVersion         bool           `env:"KEEP_INDEX_VERSION" default:"true" help:"Don't upgrade indexes to latest version (Default: true)"`
	MaintainInsertionOrder   bool           `env:"MAINTAIN_INSERTION_ORDER" default:"false" help:"restore the documents in the order of their appearance in the input source. By default the insertions will be performed in an arbitrary order. Setting this flag also enables the behavior of stopOnError and restricts NumInsertionWorkersPerCollection to 1. (Default: false)"`
	StopOnError              bool           `env:"STOP_ON_ERROR" default:"false" help:"Stop restoring at first error rather than continuing (Default: false)"`
	BypassDocumentValidation bool           `env:"BYPASS_DOCUMENT_VALIDATION" default:"false" help:"Bypass document validation (Default: false)"`
	PreserveUUID             bool           `env:"PRESERVE_UUID" default:"false" help:"preserve original collection UUIDs (requires drop) (Default: false)"`
	FixDottedHashedIndexes   bool           `env:"FIX_DOTTED_HASHED_INDEXES" default:"false" help:"when enabled, all the hashed indexes on dotted fields will be created as single field ascending indexes on the destination (Default: false)"`
	Verbosity                VerbosityFlags `embed:"" prefix:"verbosity-" envprefix:"VERBOSITY__"`
}

func (o *MongoRestoreFlags) PrepareBackupMongoRestoreOptions(filePath string) *mongorestore.MongoRestore {
	inputOptions := &mongorestore.InputOptions{
		Archive:                filePath,
		RestoreDBUsersAndRoles: o.RestoreDBUsersAndRoles,
		Objcheck:               o.ObjectCheck,
		Gzip:                   o.Gzip,
	}

	outputOptions := &mongorestore.OutputOptions{
		Drop:                     o.Drop,
		DryRun:                   o.DryRun,
		WriteConcern:             o.WriteConcern,
		NoIndexRestore:           o.NoIndexRestore,
		ConvertLegacyIndexes:     o.ConvertLegacyIndexes,
		NoOptionsRestore:         o.NoOptionsRestore,
		KeepIndexVersion:         o.KeepIndexVersion,
		MaintainInsertionOrder:   o.MaintainInsertionOrder,
		NumParallelCollections:   o.NumParallelCollections,
		NumInsertionWorkers:      o.NumInsertionWorkers,
		StopOnError:              o.StopOnError,
		BypassDocumentValidation: o.BypassDocumentValidation,
		PreserveUUID:             o.PreserveUUID,
		FixDottedHashedIndexes:   o.FixDottedHashedIndexes,
	}

	nsOptions := &mongorestore.NSOptions{
		NSExclude: o.NSExclude,
		NSInclude: o.NSInclude,
	}

	toolOptions := options.New("mongodb-restore", "", "", "", false, options.EnabledOptions{Auth: true})
	toolOptions.ConnectionString = o.ConnectionString
	toolOptions.Verbosity = &options.Verbosity{Quiet: o.Verbosity.Quiet, VLevel: o.Verbosity.Level}

	toolOptions.NormalizeOptionsAndURI()

	mongorestoreOptions, err := mongorestore.New(mongorestore.Options{
		ToolOptions:     toolOptions,
		OutputOptions:   outputOptions,
		NSOptions:       nsOptions,
		TargetDirectory: o.BackupDir,
		InputOptions:    inputOptions,
	})

	mongorestoreOptions.SkipUsersAndRoles = o.SkipUsersAndRoles

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create mongorestore options")
	}

	return mongorestoreOptions
}
