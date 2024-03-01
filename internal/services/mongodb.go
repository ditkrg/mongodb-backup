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
	dump := mongodump.MongoDump{
		ToolOptions: &options.ToolOptions{
			URI: &options.URI{},
			SSL: &options.SSL{},
			Auth: &options.Auth{
				Username:  common.GetRequiredEnv(common.MONGODB__USERNAME),
				Password:  common.GetRequiredEnv(common.MONGODB__PASSWORD),
				Mechanism: common.GetRequiredEnv(common.MONGODB__MECHANISM),
				Source:    common.GetRequiredEnv(common.MONGODB__AUTH_SOURCE),
			},
			Namespace: &options.Namespace{
				DB: common.GetRequiredEnv(common.MONGODB__DB_TO_BACKUP),
			},
			Connection: &options.Connection{
				Host: fmt.Sprintf("%s/%s", common.GetRequiredEnv(common.MONGODB__REPLICA_SET), common.GetRequiredEnv(common.MONGODB__HOSTS)),
			},
			// ReplicaSetName: mongodbReplicaSet,
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

	sslEnabled := common.GetBoolEnv(common.MONGODB__USE_SSL, true)

	if sslEnabled {
		dump.ToolOptions.SSL.UseSSL = true
		dump.ToolOptions.SSL.SSLCAFile = common.GetRequiredEnv(common.MONGODB__TLS_CA_CERT_PATH)
	}

	if mongodbBackupCollection := common.GetEnv(common.MONGODB__COLLECTIONS_TO_BACKUP); mongodbBackupCollection != "" {
		dump.ToolOptions.Namespace.Collection = mongodbBackupCollection
	}

	if excludedCollections := common.GetEnv(common.MONGODB__EXCLUDED_COLLECTIONS); excludedCollections != "" {
		excludedCollectionsArray := strings.Split(excludedCollections, ",")
		dump.OutputOptions.ExcludedCollections = excludedCollectionsArray
	}

	if excludedCollectionPrefixes := common.GetEnv(common.MONGODB__EXCLUDED_COLLECTIONS); excludedCollectionPrefixes != "" {
		excludedCollectionPrefixesArray := strings.Split(excludedCollectionPrefixes, ",")
		dump.OutputOptions.ExcludedCollectionPrefixes = excludedCollectionPrefixesArray
	}

	if err := dump.ValidateOptions(); err != nil {
		log.Err(err).Msg("Error validating options")
		return err
	}

	if err := dump.Init(); err != nil {
		log.Err(err).Msg("Error initializing database dump")
		return err
	}

	if err := dump.Dump(); err != nil {
		log.Err(err).Msg("Error dumping database")
		return err
	}

	return nil
}
