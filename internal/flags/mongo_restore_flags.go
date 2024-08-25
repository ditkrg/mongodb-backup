package flags

import (
	"github.com/mongodb/mongo-tools/common/options"
	"github.com/mongodb/mongo-tools/mongorestore"
	"github.com/rs/zerolog/log"
)

type MongoRestoreFlags struct {
	ConnectionString         string         `required:"" env:"CONNECTION_STRING" help:"The connection to the MongoDB instance to restore to"`
	BackupDir                string         `required:"" env:"BACKUP_DIR" help:"The directory to download the backup to and restore from"`
	Database                 string         `env:"DATABASE" help:"The database to dump"`
	Collection               string         `env:"COLLECTION" help:"The collection to dump"`
	WriteConcern             string         `env:"WRITE_CONCERN" default:"majority" help:"Write concern for the restore operation"`
	OplogLimit               string         `env:"OPLOG_LIMIT" help:"only include oplog entries before the provided Timestamp"`
	NSExclude                []string       `env:"NS_EXCLUDE" help:"Namespaces (database.collection) to exclude from the restore"`
	NSInclude                []string       `env:"NS_INCLUDE" help:"Namespaces (database.collection) to include in the restore"`
	NumParallelCollections   int            `env:"NUM_PARALLEL_COLLECTIONS" default:"1" help:"Number of collections to restore in parallel"`
	NumInsertionWorkers      int            `env:"NUM_INSERTION_WORKERS" default:"1" help:"Number of insert operations to run concurrently per collection"`
	Gzip                     bool           `env:"GZIP" negatable:"" default:"true" help:"Whether the backup is gzipped (Default: true)"`
	SkipUsersAndRoles        bool           `env:"SKIP_USERS_AND_ROLES" help:"Skip restoring users and roles, regardless of namespace, when true (Default: false)"`
	RestoreDBUsersAndRoles   bool           `env:"RESTORE_DB_USERS_AND_ROLES" help:"restore user and role definitions for the given database"`
	ObjectCheck              bool           `env:"OBJECT_CHECK" negatable:"" default:"true" help:"validate all objects before inserting (Default: true)"`
	Drop                     bool           `env:"DROP" help:"Drop each collection before import (Default: false)"`
	DryRun                   bool           `env:"DRY_RUN" help:"Run the restore in 'dry run' mode (Default: false)"`
	NoIndexRestore           bool           `env:"NO_INDEX_RESTORE" help:"Don't restore indexes (Default: false)"`
	ConvertLegacyIndexes     bool           `env:"CONVERT_LEGACY_INDEXES" help:"Removes invalid index options and rewrites legacy option values (e.g. true becomes 1) (Default: false)"`
	NoOptionsRestore         bool           `env:"NO_OPTIONS_RESTORE" help:"Don't restore collection options (Default: false)"`
	KeepIndexVersion         bool           `env:"KEEP_INDEX_VERSION" negatable:"" default:"true" help:"Don't upgrade indexes to latest version (Default: true)"`
	MaintainInsertionOrder   bool           `env:"MAINTAIN_INSERTION_ORDER" help:"restore the documents in the order of their appearance in the input source. By default the insertions will be performed in an arbitrary order. Setting this flag also enables the behavior of stopOnError and restricts NumInsertionWorkersPerCollection to 1. (Default: false)"`
	StopOnError              bool           `env:"STOP_ON_ERROR" help:"Stop restoring at first error rather than continuing (Default: false)"`
	BypassDocumentValidation bool           `env:"BYPASS_DOCUMENT_VALIDATION" help:"Bypass document validation (Default: false)"`
	PreserveUUID             bool           `env:"PRESERVE_UUID" help:"preserve original collection UUIDs (requires drop) (Default: false)"`
	FixDottedHashedIndexes   bool           `env:"FIX_DOTTED_HASHED_INDEXES" help:"when enabled, all the hashed indexes on dotted fields will be created as single field ascending indexes on the destination (Default: false)"`
	Verbosity                VerbosityFlags `embed:"" prefix:"verbosity-" envprefix:"VERBOSITY__"`
}

func (o *MongoRestoreFlags) PrepareBackupMongoRestoreOptions(filePath string) (*mongorestore.MongoRestore, error) {
	inputOptions := &mongorestore.InputOptions{
		Archive:                filePath,
		Objcheck:               o.ObjectCheck,
		Gzip:                   o.Gzip,
		RestoreDBUsersAndRoles: o.RestoreDBUsersAndRoles,
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
		TempUsersColl:            "tempusers",
		TempRolesColl:            "temproles",
	}

	nsOptions := &mongorestore.NSOptions{
		NSExclude: o.NSExclude,
		NSInclude: o.NSInclude,
	}

	toolOptions := options.New("mongodb-restore", "", "", "", false, options.EnabledOptions{Auth: true})
	toolOptions.ConnectionString = o.ConnectionString
	toolOptions.SetVerbosity(o.Verbosity.Level)

	toolOptions.Namespace = &options.Namespace{DB: o.Database, Collection: o.Collection}

	if err := toolOptions.NormalizeOptionsAndURI(); err != nil {
		log.Error().Err(err).Msg("Failed to normalize options and URI")
		return nil, err
	}

	mongorestoreOptions, err := mongorestore.New(mongorestore.Options{
		ToolOptions:     toolOptions,
		OutputOptions:   outputOptions,
		NSOptions:       nsOptions,
		TargetDirectory: o.BackupDir,
		InputOptions:    inputOptions,
	})

	mongorestoreOptions.SkipUsersAndRoles = o.SkipUsersAndRoles

	if err != nil {
		log.Error().Err(err).Msg("Failed to create mongorestore options")
		return nil, err
	}

	if err := mongorestoreOptions.ParseAndValidateOptions(); err != nil {
		log.Err(err).Msg("Failed to parse and validate options")
		return nil, err
	}

	return mongorestoreOptions, nil
}
