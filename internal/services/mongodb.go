package services

import (
	"fmt"

	"github.com/ditkrg/mongodb-backup/internal/helpers"
	"github.com/ditkrg/mongodb-backup/internal/models"
	"github.com/ditkrg/mongodb-backup/internal/options"
	"github.com/rs/zerolog/log"
)

func StartDatabaseDump() {
	log.Info().Msg("Starting database dump")

	if err := options.Config.MongoDB.MongoDumpOptions.Init(); err != nil {
		log.Fatal().Err(err).Msg("Error initializing database dump")
	}

	if err := options.Config.MongoDB.MongoDumpOptions.Dump(); err != nil {
		log.Fatal().Err(err).Msg("Error dumping database")
	}

	log.Info().Msg("Database dump completed successfully")
}

func StartOplogDump(oplogConfig *models.OplogConfig) {
	log.Info().Msg("Starting oplog dump")

	if oplogConfig != nil {
		options.Config.MongoDB.MongoDumpOptions.InputOptions.Query = fmt.Sprintf(helpers.OplogQuery, oplogConfig.LastJobTime)
	}

	if err := options.Config.MongoDB.MongoDumpOptions.Init(); err != nil {
		log.Fatal().Err(err).Msg("Error initializing oplog dump")
	}

	if err := options.Config.MongoDB.MongoDumpOptions.Dump(); err != nil {
		log.Fatal().Err(err).Msg("Error dumping oplog")
	}

	log.Info().Msg("Oplog dump completed successfully")
}
