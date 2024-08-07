package main

import (
	"context"
	"time"

	"github.com/ditkrg/mongodb-backup/internal/models"
	"github.com/ditkrg/mongodb-backup/internal/options"
	"github.com/ditkrg/mongodb-backup/internal/services"
	"github.com/rs/zerolog/log"
)

func main() {
	config := options.LoadConfig()

	if config.MongoDB.OpLog {
		startOplogBackup(config)
	} else {
		startBackup(config)
	}
}

func startBackup(config *options.Options) {
	// ######################
	// dump database
	// ######################
	services.StartDatabaseDump(config.MongoDB)

	// ######################
	// Prepare S3 Service
	// ######################
	ctx := context.Background()
	s3Service := services.NewS3Service(config.S3)

	// ######################
	// Upload backup to S3
	// ######################
	s3Service.StartBackupUpload(ctx, config)

	//  ######################
	//  Keep the latest N backups
	//  ######################
	if config.S3.KeepRecentN > 0 {
		s3Service.KeepMostRecentN(ctx, config)
	}

	log.Info().Msg("Backup completed successfully")
}

func startOplogBackup(config *options.Options) {
	// ######################
	// Prepare S3 Service
	// ######################
	ctx := context.Background()
	s3Service := services.NewS3Service(config.S3)

	// ######################
	// Check if a backup Exists
	// ######################
	if exists := s3Service.CheckBackupExists(ctx, config); !exists {
		return
	}

	// ######################
	// Get the latest oplog config
	// ######################
	oplogConfig := s3Service.GetOplogConfig(ctx, config)

	// ######################
	// dump oplog
	// ######################
	startTime := time.Now().Unix()
	services.StartOplogDump(config.MongoDB, oplogConfig)

	// ######################
	// Upload oplog to S3
	// ######################
	s3Service.UploadOplog(ctx, config)

	// ######################
	// Update the latest oplog config
	// ######################
	s3Service.UploadOplogConfig(ctx, config, &models.OplogConfig{LastJobTime: startTime})

	// ######################
	// Keep Relative oplog backups
	// ######################
	s3Service.KeepRelativeOplogBackups(ctx, config)
}
