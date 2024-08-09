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
	options.LoadConfig()
	if options.Config.MongoDB.OpLog {
		startOplogBackup()
	} else {
		startBackup()
	}
}

func startBackup() {
	// ######################
	// dump database
	// ######################
	services.StartDatabaseDump()

	// ######################
	// Prepare S3 Service
	// ######################
	ctx := context.Background()
	s3Service := services.NewS3Service()

	// ######################
	// Upload backup to S3
	// ######################
	s3Service.UploadBackup(ctx)

	//  ######################
	//  Keep the latest N backups
	//  ######################
	if options.Config.S3.KeepRecentN > 0 {
		s3Service.KeepRecentBackups(ctx)
	}

	log.Info().Msg("Backup completed successfully")
}

func startOplogBackup() {
	// ######################
	// Prepare S3 Service
	// ######################
	ctx := context.Background()
	s3Service := services.NewS3Service()

	// ######################
	// Check if a backup Exists
	// ######################
	backUpDir := options.Config.S3.BackupDirPath("")
	if exists := s3Service.ExistsAt(ctx, options.Config.S3.Bucket, backUpDir); !exists {
		log.Error().Msgf("no backups found in %s/%s, there must be a full backup before oplog backup", options.Config.S3.Bucket, backUpDir)
		return
	}

	// ######################
	// Get the latest oplog config
	// ######################
	oplogConfig := s3Service.GetOplogConfig(ctx)

	// ######################
	// dump oplog
	// ######################
	startTime := time.Now().Unix()
	services.StartOplogDump(oplogConfig)

	// ######################
	// Upload oplog to S3
	// ######################
	s3Service.UploadOplog(ctx)

	// ######################
	// Update the latest oplog config
	// ######################
	s3Service.UploadOplogConfig(ctx, &models.OplogConfig{LastJobTime: startTime})

	// ######################
	// Keep Relative oplog backups
	// ######################
	s3Service.KeepRelativeOplogBackups(ctx)
}
