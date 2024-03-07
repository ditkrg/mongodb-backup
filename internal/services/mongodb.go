package services

import (
	"fmt"
	"strings"

	"github.com/ditkrg/mongodb-backup/internal/common"
	"github.com/mongodb/mongo-tools/common/db"
	"github.com/mongodb/mongo-tools/common/options"
	"github.com/mongodb/mongo-tools/mongodump"
	"github.com/rs/zerolog/log"
)

func StartDatabaseDump(backupOutDir string) error {
	log.Info().Msg("Starting database dump")

	// #############################
	// get environment variables
	// #############################
	mongodbUsername := common.GetRequiredEnv(common.MONGODB__USERNAME)
	mongodbPassword := common.GetRequiredEnv(common.MONGODB__PASSWORD)
	mongodbAuthMechanism := common.GetRequiredEnv(common.MONGODB__MECHANISM)
	mongodbAuthSource := common.GetRequiredEnv(common.MONGODB__AUTH_SOURCE)

	mongodbDatabaseToBackup := common.GetRequiredEnv(common.MONGODB__DB_TO_BACKUP)

	mongodbReplicaSet := common.GetRequiredEnv(common.MONGODB__REPLICA_SET)
	mongodbHosts := common.GetRequiredEnv(common.MONGODB__HOSTS)

	sslEnabled := common.GetBoolEnv(common.MONGODB__USE_SSL, true)

	// #############################
	// create a new mongodump instance
	// #############################
	dump := mongodump.MongoDump{
		ToolOptions: &options.ToolOptions{
			URI: &options.URI{},
			SSL: &options.SSL{},
			Auth: &options.Auth{
				Username:  mongodbUsername,
				Password:  mongodbPassword,
				Mechanism: mongodbAuthMechanism,
				Source:    mongodbAuthSource,
			},
			Namespace: &options.Namespace{
				DB: mongodbDatabaseToBackup,
			},
			Connection: &options.Connection{
				Host: fmt.Sprintf("%s/%s", mongodbReplicaSet, mongodbHosts),
			},
		},
		ProgressManager: &ProgressManager{},
		InputOptions:    &mongodump.InputOptions{},
		SessionProvider: &db.SessionProvider{},
		OutputOptions: &mongodump.OutputOptions{
			NumParallelCollections: 1,
			Out:                    backupOutDir,
			Gzip:                   false,
			DumpDBUsersAndRoles:    false,
		},
	}

	if sslEnabled {
		dump.ToolOptions.SSL.UseSSL = true
		dump.ToolOptions.SSL.SSLCAFile = common.GetRequiredEnv(common.MONGODB__TLS_CA_CERT_PATH)
	}

	if mongodbBackupCollection := common.GetEnv(common.MONGODB__COLLECTION_TO_BACKUP); mongodbBackupCollection != "" {
		dump.ToolOptions.Namespace.Collection = mongodbBackupCollection
	}

	if excludedCollections := common.GetEnv(common.MONGODB__EXCLUDED_COLLECTIONS); excludedCollections != "" {
		excludedCollectionsArray := strings.Split(excludedCollections, ",")
		dump.OutputOptions.ExcludedCollections = excludedCollectionsArray
	}

	if excludedCollectionPrefixes := common.GetEnv(common.MONGODB__EXCLUDED_COLLECTION_PREFIXES); excludedCollectionPrefixes != "" {
		excludedCollectionPrefixesArray := strings.Split(excludedCollectionPrefixes, ",")
		dump.OutputOptions.ExcludedCollectionPrefixes = excludedCollectionPrefixesArray
	}

	// #############################
	// validate options
	// #############################
	if err := dump.ValidateOptions(); err != nil {
		log.Err(err).Msg("Error validating options")
		return err
	}

	// #############################
	// initialize the dump
	// #############################
	if err := dump.Init(); err != nil {
		log.Err(err).Msg("Error initializing database dump")
		return err
	}

	// #############################
	// dump the database
	// #############################
	if err := dump.Dump(); err != nil {
		log.Err(err).Msg("Error dumping database")
		return err
	}

	log.Info().Msg("Database dump completed successfully")

	return nil
}
