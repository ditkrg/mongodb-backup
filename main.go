package main

import (
	"context"
	"fmt"
	"os"
	"time"

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
		os.Exit(1)
	}

	// ######################
	// compress backup
	// ######################

	var backupFileName string

	if config.MongoDB.DatabaseToBackup == "" {
		backupFileName = fmt.Sprintf("%s_%s.zip", "all_databases", time.Now().Format("060102-150405"))
	} else {
		backupFileName = fmt.Sprintf("%s_%s.tar", config.MongoDB.DatabaseToBackup, time.Now().Format("060102-150405"))
	}

	if err := services.CompressDir(config.MongoDB.BackupOutDir, backupFileName); err != nil {
		log.Err(err).Msg("Failed to compress backup")
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
	if err := s3Service.StartBackupUpload(ctx, config, backupFileName); err != nil {
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
