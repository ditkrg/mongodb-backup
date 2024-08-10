package options

import (
	"fmt"
	"strings"

	"github.com/mongodb/mongo-tools/common/options"
	"github.com/mongodb/mongo-tools/mongorestore"
	"github.com/rs/zerolog/log"
)

type MongoRestoreOptions struct {
	ConnectionString           string    `env:"CONNECTION_STRING"`
	Database                   string    `env:"DATABASE"`
	Collection                 string    `env:"COLLECTION"`
	BackupDir                  string    `env:"BACKUP_DIR,default=/backup"`
	Gzip                       bool      `env:"GZIP,default=true"`
	RestoreDBUsersAndRoles     bool      `env:"RESTORE_DB_USERS_AND_ROLES,default=true"`
	SkipUsersAndRoles          bool      `env:"SKIP_USERS_AND_ROLES,default=false"`
	ObjectCheck                bool      `env:"OBJECT_CHECK,default=true"`
	Drop                       bool      `env:"DROP,default=false"`
	DryRun                     bool      `env:"DRY_RUN,default=false"`
	WriteConcern               string    `env:"WRITE_CONCERN,default=majority"`
	NoIndexRestore             bool      `env:"NO_INDEX_RESTORE,default=false"`
	ConvertLegacyIndexes       bool      `env:"CONVERT_LEGACY_INDEXES,default=false"`
	NoOptionsRestore           bool      `env:"NO_OPTIONS_RESTORE,default=false"`
	KeepIndexVersion           bool      `env:"KEEP_INDEX_VERSION,default=true"`
	MaintainInsertionOrder     bool      `env:"MAINTAIN_INSERTION_ORDER,default=true"`
	NumParallelCollections     int       `env:"NUM_PARALLEL_COLLECTIONS,default=1"`
	NumInsertionWorkers        int       `env:"NUM_INSERTION_WORKERS,default=1"`
	StopOnError                bool      `env:"STOP_ON_ERROR,default=false"`
	BypassDocumentValidation   bool      `env:"BYPASS_DOCUMENT_VALIDATION,default=false"`
	PreserveUUID               bool      `env:"PRESERVE_UUID,default=false"`
	FixDottedHashedIndexes     bool      `env:"FIX_DOTTED_HASHED_INDEXES,default=false"`
	ExcludedCollections        []string  `env:"EXCLUDED_COLLECTIONS"`
	ExcludedCollectionPrefixes []string  `env:"EXCLUDED_COLLECTION_PREFIXES"`
	NSExclude                  []string  `env:"NS_EXCLUDE"`
	NSInclude                  []string  `env:"NS_INCLUDE"`
	OplogLimit                 string    `env:"OPLOG_LIMIT"`
	Verbosity                  Verbosity `env:",prefix=VERBOSITY__"`
}

func (o *MongoRestoreOptions) PrepareBackupMongoRestoreOptions(filePath string) *mongorestore.MongoRestore {
	inputOptions := &mongorestore.InputOptions{
		Archive:                filePath,
		RestoreDBUsersAndRoles: o.RestoreDBUsersAndRoles,
		Objcheck:               o.ObjectCheck,
		Gzip:                   o.Gzip,
	}

	mongorestoreOptions := o.getMongoRestoreOptions()
	mongorestoreOptions.InputOptions = inputOptions

	return mongorestoreOptions

}

func (o *MongoRestoreOptions) PrepareOplogMongoRestoreOptions() *mongorestore.MongoRestore {
	o.BackupDir, _ = strings.CutSuffix(o.BackupDir, "/")

	inputOptions := &mongorestore.InputOptions{
		Directory:              o.BackupDir,
		RestoreDBUsersAndRoles: o.RestoreDBUsersAndRoles,
		Objcheck:               o.ObjectCheck,
		Gzip:                   o.Gzip,
		OplogLimit:             o.OplogLimit,
		OplogReplay:            true,
	}

	mongorestoreOptions := o.getMongoRestoreOptions()
	mongorestoreOptions.InputOptions = inputOptions
	return mongorestoreOptions
}

func (o *MongoRestoreOptions) getMongoRestoreOptions() *mongorestore.MongoRestore {
	fmt.Println(o.StopOnError)
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
		ExcludedCollections:        o.ExcludedCollections,
		ExcludedCollectionPrefixes: o.ExcludedCollectionPrefixes,
		NSExclude:                  o.NSExclude,
		NSInclude:                  o.NSInclude,
	}

	toolOptions := options.New("mongodb-restore", "", "", "", false, options.EnabledOptions{Auth: true})
	toolOptions.ConnectionString = o.ConnectionString
	toolOptions.Verbosity = &options.Verbosity{Quiet: o.Verbosity.Quiet, VLevel: o.Verbosity.Level}
	toolOptions.Namespace = &options.Namespace{DB: o.Database, Collection: o.Collection}

	toolOptions.NormalizeOptionsAndURI()

	mongorestoreOptions, err := mongorestore.New(mongorestore.Options{
		ToolOptions:     toolOptions,
		OutputOptions:   outputOptions,
		NSOptions:       nsOptions,
		TargetDirectory: o.BackupDir,
	})

	mongorestoreOptions.SkipUsersAndRoles = o.SkipUsersAndRoles
	// mongorestoreOptions.ProgressManager = &helpers.ProgressManager{}

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create mongorestore options")
	}

	return mongorestoreOptions
}
