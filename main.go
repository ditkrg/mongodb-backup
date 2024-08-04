package main

import (
	"context"
	"os"

	"github.com/ditkrg/mongodb-backup/internal/options"
	"github.com/ditkrg/mongodb-backup/internal/services"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"github.com/sethvargo/go-envconfig"
)

func main() {
	// ######################
	// Load env variables
	// ######################
	godotenv.Load()

	var config options.Options

	if err := envconfig.Process(context.Background(), &config); err != nil {
		log.Fatal().Err(err).Msg("Failed to process environment variables")
	}

	config.MongoDB.PrepareMongoDumpOptions()

	if err := config.Validate(); err != nil {
		log.Fatal().Err(err).Msg("Invalid configuration")
	}

	// ######################
	// dump database
	// ######################

	if err := services.StartDatabaseDump(config.MongoDB); err != nil {
		log.Err(err).Msg("Failed to dump database")
		os.Exit(1)
	}

	// ######################
	// Prepare S3 Service
	// ######################
	ctx := context.Background()
	s3Service := services.InitS3Service(config.S3)

	// ######################
	// Upload backup to S3
	// ######################

	if err := s3Service.StartBackupUpload(ctx, config); err != nil {
		log.Err(err).Msg("Failed to upload backup to S3")
		os.Exit(1)
	}

	//  ######################
	//  Keep the latest N backups
	//  ######################
	if config.S3.KeepRecentN > 0 {
		if err := s3Service.KeepMostRecentN(ctx, config); err != nil {
			log.Err(err).Msg("Failed to keep the latest N backups")
			os.Exit(1)
		}
	}

	log.Info().Msg("Backup completed successfully")
}
