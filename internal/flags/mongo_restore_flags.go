package flags

import (
	"strconv"
	"time"

	"github.com/mongodb/mongo-tools/common/options"
	"github.com/mongodb/mongo-tools/mongorestore"
	"github.com/rs/zerolog/log"
)

type MongoRestoreFlags struct {
	ConnectionString string `required:"" env:"CONNECTION_STRING" help:"The connection to the MongoDB instance to restore to"`
	BackupDir        string `required:"" env:"BACKUP_DIR" help:"The directory to download the backup to and restore from"`

	NamespaceOptions struct {
		Database   string   `env:"DATABASE" short:"d" help:"The database to restore"`
		Collection string   `env:"COLLECTION" short:"c" help:"The collection to restore"`
		NSExclude  []string `env:"NS_EXCLUDE" help:"Namespaces (database.collection) to exclude from the restore"`
		NSInclude  []string `env:"NS_INCLUDE" help:"Namespaces (database.collection) to include in the restore"`
	} `embed:"" group:"namespace options"`

	RestoreOptions struct {
		Drop                     bool   `env:"DROP" help:"Drop each collection before import"`
		DryRun                   bool   `env:"DRY_RUN" help:"Run the restore in 'dry run' mode"`
		WriteConcern             string `env:"WRITE_CONCERN" default:"majority" help:"Write concern for the restore operation"`
		NoIndexRestore           bool   `env:"NO_INDEX_RESTORE" help:"Don't restore indexes"`
		ConvertLegacyIndexes     bool   `env:"CONVERT_LEGACY_INDEXES" help:"Removes invalid index options and rewrites legacy option values (e.g. true becomes 1)"`
		NoOptionsRestore         bool   `env:"NO_OPTIONS_RESTORE" help:"Don't restore collection options"`
		KeepIndexVersion         bool   `env:"KEEP_INDEX_VERSION" negatable:"" default:"true" help:"Don't upgrade indexes to latest version (Default: true)"`
		MaintainInsertionOrder   bool   `env:"MAINTAIN_INSERTION_ORDER" help:"restore the documents in the order of their appearance in the input source. By default the insertions will be performed in an arbitrary order. Setting this flag also enables the behavior of stopOnError and restricts NumInsertionWorkersPerCollection to 1."`
		NumParallelCollections   int    `env:"NUM_PARALLEL_COLLECTIONS" default:"1" help:"Number of collections to restore in parallel"`
		NumInsertionWorkers      int    `env:"NUM_INSERTION_WORKERS" default:"1" help:"Number of insert operations to run concurrently per collection"`
		StopOnError              bool   `env:"STOP_ON_ERROR" help:"Stop restoring at first error rather than continuing"`
		BypassDocumentValidation bool   `env:"BYPASS_DOCUMENT_VALIDATION" help:"Bypass document validation"`
		PreserveUUID             bool   `env:"PRESERVE_UUID" help:"preserve original collection UUIDs (requires drop)"`
		FixDottedHashedIndexes   bool   `env:"FIX_DOTTED_HASHED_INDEXES" help:"when enabled, all the hashed indexes on dotted fields will be created as single field ascending indexes on the destination"`
	} `embed:"" group:"restore options"`

	InputOptions struct {
		ObjectCheck            bool   `env:"OBJECT_CHECK" negatable:"" default:"true" help:"validate all objects before inserting (Default: true)"`
		OplogReplay            bool   `env:"OPLOG_REPLAY" negatable:"" default:"true" help:"replay the oplog backups (Default: true)"`
		OplogLimit             string `env:"OPLOG_LIMIT_TO" help:"The End time of the OpLog restore."`
		Gzip                   bool   `env:"GZIP" negatable:"" default:"true" help:"Whether the backup is gzipped (Default: true)"`
		RestoreDBUsersAndRoles bool   `env:"RESTORE_DB_USERS_AND_ROLES" help:"restore user and role definitions for the given database"`
		SkipUsersAndRoles      bool   `env:"SKIP_USERS_AND_ROLES" help:"Skip restoring users and roles, regardless of namespace, when true"`
	} `embed:"" group:"restore options"`
}

func (o *MongoRestoreFlags) PrepareBackupMongoRestoreOptions(filePath string) (*mongorestore.MongoRestore, error) {
	log.Info().Msg("preparing mongodb restore options")

	inputOptions := &mongorestore.InputOptions{
		Archive:                filePath,
		Objcheck:               o.InputOptions.ObjectCheck,
		Gzip:                   o.InputOptions.Gzip,
		RestoreDBUsersAndRoles: o.InputOptions.RestoreDBUsersAndRoles,
	}

	outputOptions := &mongorestore.OutputOptions{
		Drop:                     o.RestoreOptions.Drop,
		DryRun:                   o.RestoreOptions.DryRun,
		WriteConcern:             o.RestoreOptions.WriteConcern,
		NoIndexRestore:           o.RestoreOptions.NoIndexRestore,
		ConvertLegacyIndexes:     o.RestoreOptions.ConvertLegacyIndexes,
		NoOptionsRestore:         o.RestoreOptions.NoOptionsRestore,
		KeepIndexVersion:         o.RestoreOptions.KeepIndexVersion,
		MaintainInsertionOrder:   o.RestoreOptions.MaintainInsertionOrder,
		NumParallelCollections:   o.RestoreOptions.NumParallelCollections,
		NumInsertionWorkers:      o.RestoreOptions.NumInsertionWorkers,
		StopOnError:              o.RestoreOptions.StopOnError,
		BypassDocumentValidation: o.RestoreOptions.BypassDocumentValidation,
		PreserveUUID:             o.RestoreOptions.PreserveUUID,
		FixDottedHashedIndexes:   o.RestoreOptions.FixDottedHashedIndexes,
		TempUsersColl:            "tempusers",
		TempRolesColl:            "temproles",
	}

	nsOptions := &mongorestore.NSOptions{
		NSExclude: o.NamespaceOptions.NSExclude,
		NSInclude: o.NamespaceOptions.NSInclude,
	}

	toolOptions := options.New("mongodb-restore", "", "", "", false, options.EnabledOptions{Auth: true})
	toolOptions.ConnectionString = o.ConnectionString
	toolOptions.Namespace = &options.Namespace{DB: o.NamespaceOptions.Database, Collection: o.NamespaceOptions.Collection}

	if err := toolOptions.NormalizeOptionsAndURI(); err != nil {
		log.Error().Err(err).Msg("Failed to normalize options and URI")
		return nil, err
	}

	mongorestore, err := mongorestore.New(mongorestore.Options{
		ToolOptions:     toolOptions,
		OutputOptions:   outputOptions,
		NSOptions:       nsOptions,
		TargetDirectory: o.BackupDir,
		InputOptions:    inputOptions,
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to create mongorestore options")
		return nil, err
	}
	mongorestore.SkipUsersAndRoles = o.InputOptions.SkipUsersAndRoles

	if err := mongorestore.ParseAndValidateOptions(); err != nil {
		log.Err(err).Msg("Failed to parse and validate options")
		return nil, err
	}

	return mongorestore, nil
}

func (o *MongoRestoreFlags) PrepareOplogMongoRestoreOptions(backupDir string, to *time.Time) (*mongorestore.MongoRestore, error) {
	log.Info().Msg("preparing mongodb oplog restore options")

	inputOptions := &mongorestore.InputOptions{
		Directory: backupDir,
		// RestoreDBUsersAndRoles: o.RestoreDBUsersAndRoles,
		Objcheck:    o.InputOptions.ObjectCheck,
		Gzip:        o.InputOptions.Gzip,
		OplogReplay: true,
	}

	if to != nil {
		inputOptions.OplogLimit = strconv.FormatInt(to.Unix(), 10)
	}

	outputOptions := &mongorestore.OutputOptions{
		Drop:                     o.RestoreOptions.Drop,
		DryRun:                   o.RestoreOptions.DryRun,
		WriteConcern:             o.RestoreOptions.WriteConcern,
		NoIndexRestore:           o.RestoreOptions.NoIndexRestore,
		ConvertLegacyIndexes:     o.RestoreOptions.ConvertLegacyIndexes,
		NoOptionsRestore:         o.RestoreOptions.NoOptionsRestore,
		KeepIndexVersion:         o.RestoreOptions.KeepIndexVersion,
		MaintainInsertionOrder:   o.RestoreOptions.MaintainInsertionOrder,
		NumParallelCollections:   o.RestoreOptions.NumParallelCollections,
		NumInsertionWorkers:      o.RestoreOptions.NumInsertionWorkers,
		StopOnError:              o.RestoreOptions.StopOnError,
		BypassDocumentValidation: o.RestoreOptions.BypassDocumentValidation,
		PreserveUUID:             o.RestoreOptions.PreserveUUID,
		FixDottedHashedIndexes:   o.RestoreOptions.FixDottedHashedIndexes,
	}

	nsOptions := &mongorestore.NSOptions{
		NSExclude: o.NamespaceOptions.NSExclude,
		NSInclude: o.NamespaceOptions.NSInclude,
	}

	toolOptions := options.New("mongodb-restore", "", "", "", false, options.EnabledOptions{Auth: true})
	toolOptions.ConnectionString = o.ConnectionString

	if err := toolOptions.NormalizeOptionsAndURI(); err != nil {
		log.Error().Err(err).Msg("Failed to normalize options and URI")
		return nil, err
	}

	mongorestore, err := mongorestore.New(mongorestore.Options{
		ToolOptions:     toolOptions,
		OutputOptions:   outputOptions,
		NSOptions:       nsOptions,
		TargetDirectory: backupDir,
		InputOptions:    inputOptions,
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to create mongorestore options")
		return nil, err
	}

	mongorestore.SkipUsersAndRoles = o.InputOptions.SkipUsersAndRoles

	if err := mongorestore.ParseAndValidateOptions(); err != nil {
		log.Err(err).Msg("Failed to parse and validate options")
		return nil, err
	}

	log.Info().Msg("finished preparing oplog mongodb restore options")
	return mongorestore, nil
}
