package services

import (
	"context"
	"time"

	"github.com/ditkrg/mongodb-backup/internal/models"
	"github.com/ditkrg/mongodb-backup/internal/options"
	"github.com/rs/zerolog/log"
)

func StartBackupProcess() {
	if options.Dump.MongoDump.OpLog {
		startOplogBackup()
	} else {
		startBackup()
	}
}

func startBackup() {
	// ######################
	// dump database
	// ######################
	StartDatabaseDump()

	// ######################
	// Prepare S3 Service
	// ######################
	ctx := context.Background()
	s3Service := NewS3Service(options.Dump.S3)

	// ######################
	// Upload backup to S3
	// ######################
	s3Service.UploadBackup(ctx)

	//  ######################
	//  Keep the latest N backups
	//  ######################
	if options.Dump.S3.KeepRecentN > 0 {
		s3Service.KeepRecentBackups(ctx)
	}

	log.Info().Msg("Backup completed successfully")
}

func startOplogBackup() {
	// ######################
	// Prepare S3 Service
	// ######################
	ctx := context.Background()
	s3Service := NewS3Service(options.Dump.S3)

	// ######################
	// Check if a backup Exists
	// ######################
	backUpDir := options.Dump.S3.BackupDirPath("")
	if exists := s3Service.ExistsAt(ctx, options.Dump.S3.Bucket, backUpDir); !exists {
		log.Error().Msgf("no backups found in %s/%s, there must be a full backup before oplog backup", options.Dump.S3.Bucket, backUpDir)
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
	StartOplogDump(oplogConfig)

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
