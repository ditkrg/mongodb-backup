package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ditkrg/mongodb-backup/internal/flags"
	"github.com/ditkrg/mongodb-backup/internal/helpers"
	"github.com/ditkrg/mongodb-backup/internal/models"
	"github.com/ditkrg/mongodb-backup/internal/services"
	"github.com/mongodb/mongo-tools/mongodump"
	"github.com/rs/zerolog/log"
)

type DumpCommand struct {
	S3    flags.S3Flags        `embed:"" group:"Common S3 Flags:"`
	Mongo flags.MongoDumpFlags `embed:"" envprefix:"MONGO_DUMP__" group:"Common Mongo Dump Flags:"`
}

func (command DumpCommand) Run() error {

	if command.Mongo.OpLog {
		return startOplogBackup(&command)
	} else {
		return startBackup(&command)
	}
}

func startBackup(command *DumpCommand) error {
	var err error
	var mongoDump *mongodump.MongoDump

	s3FileKey := helpers.S3FileKey(true, command.Mongo.Gzip)
	s3FileKeyWithPrefix := fmt.Sprintf("%s/%s", helpers.S3BackupPrefix(command.S3.Prefix, command.Mongo.Database), s3FileKey)

	// ######################
	// Prepare MongoDump
	// ######################
	if mongoDump, err = command.Mongo.PrepareMongoDump(); err != nil {
		return err
	}

	// ######################
	// dump database
	// ######################
	log.Info().Msg("Starting database dump")

	if err := mongoDump.Init(); err != nil {
		log.Error().Err(err).Msg("Error initializing database dump")
		return err
	}

	if err := mongoDump.Dump(); err != nil {
		log.Error().Err(err).Msg("Error dumping database")
		return err
	}

	log.Info().Msg("Database dump completed successfully")

	// ######################
	// Prepare S3 Service
	// ######################
	ctx := context.Background()
	s3Service := services.NewS3Service(command.S3)

	// ######################
	// Upload backup to S3
	// ######################
	if err := s3Service.UploadFile(
		ctx,
		command.S3.Bucket,
		s3FileKeyWithPrefix,
		mongoDump.OutputOptions.Archive,
	); err != nil {
		return err
	}

	//  ######################
	//  Keep the latest N backups
	//  ######################
	if err := keepRecentBackups(ctx, s3Service, command); err != nil {
		return err
	}

	log.Info().Msg("Backup completed successfully")
	return nil
}

func startOplogBackup(command *DumpCommand) error {
	var mongoDump *mongodump.MongoDump
	var oplogConfig *models.OplogConfig
	var listResp *s3.ListObjectsV2Output
	var oplogListResp *s3.ListObjectsV2Output
	var exists bool
	var err error

	backupDir, _ := strings.CutSuffix(command.Mongo.BackupDir, "/")

	s3OpLogFileKey := fmt.Sprintf("%s.tar.gz", time.Now().Format("060102-150405"))
	s3OpLogFilePrefix := helpers.S3OplogPrefix(command.S3.Prefix)
	s3OpLogFileKeyWithPrefix := fmt.Sprintf("%s/%s", s3OpLogFilePrefix, s3OpLogFileKey)

	s3OplogConfigFileKeyWithPrefix := fmt.Sprintf("%s/%s", helpers.S3OplogPrefix(command.S3.Prefix), helpers.ConfigFileName)

	s3FullBackupPrefix := helpers.S3BackupPrefix(command.S3.Prefix, "")
	tarFileDir := backupDir + "/local"

	// ######################
	// Prepare S3 Service
	// ######################
	ctx := context.Background()
	s3Service := services.NewS3Service(command.S3)

	// ######################
	// Check if a backup Exists
	// ######################

	if exists, err = s3Service.ObjectsExistsAt(ctx, command.S3.Bucket, s3FullBackupPrefix); err != nil {
		return err
	}

	if !exists {
		log.Info().Msgf("no backups found in %s/%s, there must be a full backup before oplog backup", command.S3.Bucket, s3FullBackupPrefix)
		return nil
	}

	// ######################
	// Get the latest oplog config
	// ######################
	if oplogConfig, err = getOplogConfig(ctx, s3Service, command); err != nil {
		return err
	}

	// ######################
	// Prepare MongoDump
	// ######################
	if mongoDump, err = command.Mongo.PrepareMongoDump(); err != nil {
		return err
	}

	// ######################
	// dump oplog
	// ######################
	startTime := time.Now().Unix()

	log.Info().Msg("Starting oplog dump")

	if oplogConfig == nil {
		log.Info().Msg("Taking a full OpLog backup")
	} else {
		log.Info().Msgf("Taking OpLog from %d", oplogConfig.LastJobTime)
		mongoDump.InputOptions.Query = fmt.Sprintf(helpers.OplogQuery, oplogConfig.LastJobTime)
	}

	if err := mongoDump.Init(); err != nil {
		log.Error().Err(err).Msg("Error initializing oplog dump")
		return err
	}

	if err := mongoDump.Dump(); err != nil {
		log.Error().Err(err).Msg("Error dumping oplog")
		return err
	}

	log.Info().Msg("Oplog dump completed successfully")

	// ######################
	// Upload oplog to S3
	// ######################
	helpers.TarDirectory(tarFileDir, s3OpLogFileKey)

	s3Service.UploadFile(
		ctx,
		command.S3.Bucket,
		s3OpLogFileKeyWithPrefix,
		fmt.Sprintf("%s/%s", tarFileDir, s3OpLogFileKey),
	)

	// ######################
	// Upload a new oplog config
	// ######################
	log.Info().Msg("Uploading a new oplog config")

	oplogConfigByteArray, err := json.Marshal(&models.OplogConfig{LastJobTime: startTime})

	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal oplog config")
		return err
	}

	if _, err := s3Service.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(command.S3.Bucket),
		Key:    aws.String(s3OplogConfigFileKeyWithPrefix),
		Body:   bytes.NewReader(oplogConfigByteArray),
	}); err != nil {
		log.Error().Err(err).Msg("Failed to upload content")
		return err
	}

	// ######################
	// Keep Relative oplog backups
	// ######################
	log.Info().Msg("Keep Relative Oplog Backups")

	if listResp, err = s3Service.List(
		ctx,
		command.S3.Bucket,
		s3FullBackupPrefix,
	); err != nil {
		return err
	}

	sort.Slice(listResp.Contents, func(i, j int) bool {
		return listResp.Contents[i].LastModified.Before(*listResp.Contents[j].LastModified)
	})

	oldestBackup := listResp.Contents[0].LastModified

	// ######################
	// Get all oplog backups older than the oldest backup
	// ######################
	objectsToDelete := make([]types.ObjectIdentifier, 0)

	if oplogListResp, err = s3Service.List(ctx, command.S3.Bucket, s3OpLogFilePrefix); err != nil {
		return err
	}

	for _, obj := range oplogListResp.Contents {
		if obj.LastModified.Before(*oldestBackup) && (*obj.Key) != s3OplogConfigFileKeyWithPrefix {
			objectsToDelete = append(objectsToDelete, types.ObjectIdentifier{Key: obj.Key})
		}
	}

	if err := s3Service.Delete(ctx, command.S3.Bucket, objectsToDelete); err != nil {
		return err
	}

	return nil
}

func keepRecentBackups(ctx context.Context, s3Service *services.S3Service, command *DumpCommand) error {
	if command.Mongo.KeepRecentN > 0 {
		return nil
	}

	log.Info().Msgf("Keep most Recent %d Backups", command.Mongo.KeepRecentN)

	var err error
	var resp *s3.ListObjectsV2Output

	if resp, err = s3Service.List(
		ctx,
		command.S3.Bucket,
		helpers.S3BackupPrefix(command.S3.Prefix, command.Mongo.Database),
	); err != nil {
		log.Error().Err(err).Msg("Failed to list backups")
		return err
	}

	s3BackupCount := len(resp.Contents)

	log.Info().Msgf("Found %d backups", s3BackupCount)

	if s3BackupCount > command.Mongo.KeepRecentN {

		backupsToDeleteCount := s3BackupCount - command.Mongo.KeepRecentN
		objectsToDelete := make([]types.ObjectIdentifier, backupsToDeleteCount)

		sort.Slice(resp.Contents, func(i, j int) bool {
			return resp.Contents[i].LastModified.After(*resp.Contents[j].LastModified)
		})

		for i, obj := range resp.Contents[command.Mongo.KeepRecentN:] {
			objectsToDelete[i] = types.ObjectIdentifier{
				Key: obj.Key,
			}
		}

		if err := s3Service.Delete(ctx, command.S3.Bucket, objectsToDelete); err != nil {
			return err
		}
	}

	return nil
}

func getOplogConfig(ctx context.Context, s3Service *services.S3Service, command *DumpCommand) (*models.OplogConfig, error) {
	log.Info().Msg("Getting the latest oplog config")

	oplogKeyWithPrefix := fmt.Sprintf("%s/%s", helpers.S3OplogPrefix(command.S3.Prefix), helpers.ConfigFileName)

	resp, err := s3Service.Get(ctx, command.S3.Bucket, oplogKeyWithPrefix)

	if err != nil {
		var responseError *awshttp.ResponseError

		if ok := errors.As(err, &responseError); !ok {
			log.Error().Err(err).Msg("Failed to get config file from S3")
			return nil, err
		}

		if responseError.ResponseError.HTTPStatusCode() == http.StatusNotFound {
			log.Info().Msg("No oplog config found")
			return nil, nil
		}

		log.Error().Err(err).Msg("Failed to get config file from S3")
		return nil, err
	}

	defer resp.Body.Close()

	var oplogConfig models.OplogConfig
	if err := json.NewDecoder(resp.Body).Decode(&oplogConfig); err != nil {
		log.Error().Err(err).Msg("Failed to decode the config file")
		return nil, err
	}

	log.Info().Msg("Got the latest oplog config")
	return &oplogConfig, nil
}
