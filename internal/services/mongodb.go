package services

import (
	"github.com/ditkrg/mongodb-backup/internal/options"
	"github.com/rs/zerolog/log"
)

func StartDatabaseDump(options options.MongoDBOptions) error {
	log.Info().Msg("Starting database dump")

	if err := options.MongoDumpOptions.Init(); err != nil {
		log.Err(err).Msg("Error initializing database dump")
		return err
	}

	if err := options.MongoDumpOptions.Dump(); err != nil {
		log.Err(err).Msg("Error dumping database")
		return err
	}

	log.Info().Msg("Database dump completed successfully")

	return nil
}
