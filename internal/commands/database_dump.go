package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsHttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ditkrg/mongodb-backup/internal/flags"
	"github.com/ditkrg/mongodb-backup/internal/helpers"
	"github.com/ditkrg/mongodb-backup/internal/models"
	"github.com/ditkrg/mongodb-backup/internal/services"
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
	timeNow := time.Now().UTC().Format(helpers.TimeFormat)

	s3FileKey := fmt.Sprintf("%s.archive", timeNow)
	if command.Mongo.Gzip {
		s3FileKey = fmt.Sprintf("%s.gzip", timeNow)
	}

	s3FileKeyWithPrefix := helpers.S3BackupPrefix(command.S3.Prefix, command.Mongo.Database) + s3FileKey

	// ######################
	// Prepare MongoDump
	// ######################
	mongoDump, err := command.Mongo.PrepareMongoDump()
	if err != nil {
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

	os.Remove(mongoDump.OutputOptions.Archive)

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
	tarFileDir := strings.TrimSuffix(command.Mongo.BackupDir, "/") + "/local/"

	// ######################
	// Prepare S3 Service
	// ######################
	ctx := context.Background()
	s3Service := services.NewS3Service(command.S3)

	// ######################
	// Check if a backup Exists
	// ######################
	bucketObjects, err := s3Service.List(ctx, command.S3.Bucket, helpers.S3BackupPrefix(command.S3.Prefix, ""))
	if err != nil {
		return err
	}

	if len(bucketObjects.Contents) == 0 {
		log.Info().Msgf("no backups found in %s/%s, there must be a full backup before oplog backup", command.S3.Bucket, helpers.S3BackupPrefix(command.S3.Prefix, ""))
		return nil
	}

	log.Info().Msgf("Found %d objects in %s/%s", len(bucketObjects.Contents), command.S3.Bucket, helpers.S3BackupPrefix(command.S3.Prefix, ""))

	// ######################
	// Get the latest oplog config
	// ######################
	previousOplogRunInfo, err := getPreviousOplogRunData(ctx, s3Service, command)
	if err != nil {
		return err
	}

	// ######################
	// Prepare MongoDump
	// ######################
	mongoDump, err := command.Mongo.PrepareMongoDump()
	if err != nil {
		return err
	}

	// ######################
	// Prepare OpLog Backup Key
	// ######################
	startTime := time.Now().UTC().Format(helpers.TimeFormat)
	var s3OpLogBackupKey string

	if previousOplogRunInfo == nil {
		helpers.SortByKeyTimeStamp(bucketObjects.Contents, helpers.S3BackupPrefix(command.S3.Prefix, ""))
		key := *bucketObjects.Contents[0].Key
		key = strings.TrimPrefix(key, helpers.S3BackupPrefix(command.S3.Prefix, ""))
		key = strings.TrimSuffix(key, ".gzip")
		key = strings.TrimSuffix(key, ".archive")
		previousOplogRunInfo = &models.PreviousOplogRunInfo{OplogTakenFrom: "0", OplogTakenTo: key}
	}

	s3OpLogBackupKey = fmt.Sprintf("%s_%s.tar.gz", previousOplogRunInfo.OplogTakenTo, startTime)
	mongoDump.InputOptions.Query = fmt.Sprintf(helpers.OplogQuery, previousOplogRunInfo.OplogTakenTo, startTime)

	log.Info().Msgf("Taking OpLog from %s to %s", previousOplogRunInfo.OplogTakenTo, startTime)

	// ######################
	// dump oplog
	// ######################
	log.Info().Msg("Starting oplog dump")

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
	// Tar Oplog Directory
	// ######################
	if err := helpers.TarDirectory(tarFileDir, s3OpLogBackupKey); err != nil {
		return err
	}

	// ######################
	// Upload oplog to S3
	// ######################
	if err := s3Service.UploadFile(
		ctx,
		command.S3.Bucket,
		helpers.S3OplogPrefix(command.S3.Prefix)+s3OpLogBackupKey,
		tarFileDir+s3OpLogBackupKey,
	); err != nil {
		return err
	}

	os.RemoveAll(tarFileDir)

	// ######################
	// Upload a new oplog config
	// ######################
	log.Info().Msg("Upload the current oplog run info")

	oplogConfigByteArray, err := json.Marshal(&models.PreviousOplogRunInfo{OplogTakenFrom: previousOplogRunInfo.OplogTakenTo, OplogTakenTo: startTime})

	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal oplog config")
		return err
	}

	if _, err := s3Service.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(command.S3.Bucket),
		Key:    aws.String(helpers.S3OplogPrefix(command.S3.Prefix) + helpers.ConfigFileName),
		Body:   bytes.NewReader(oplogConfigByteArray),
	}); err != nil {
		log.Error().Err(err).Msg("Failed to upload content")
		return err
	}

	// ######################
	// Keep Relative oplog backups
	// ######################
	if err := keepRelativeOplogBackups(ctx, s3Service, command); err != nil {
		return err
	}

	return nil
}

func keepRecentBackups(ctx context.Context, s3Service *services.S3Service, command *DumpCommand) error {
	if command.Mongo.KeepRecentN <= 0 {
		return nil
	}

	log.Info().Msgf("Keep most Recent %d Backups", command.Mongo.KeepRecentN)

	resp, err := s3Service.List(
		ctx,
		command.S3.Bucket,
		helpers.S3BackupPrefix(command.S3.Prefix, command.Mongo.Database),
	)

	if err != nil {
		log.Error().Err(err).Msg("Failed to list backups")
		return err
	}

	s3BackupCount := len(resp.Contents)

	log.Info().Msgf("Found %d backups", s3BackupCount)

	if s3BackupCount > command.Mongo.KeepRecentN {

		backupsToDeleteCount := s3BackupCount - command.Mongo.KeepRecentN
		objectsToDelete := make([]types.ObjectIdentifier, backupsToDeleteCount)

		helpers.SortByKeyTimeStamp(resp.Contents, helpers.S3BackupPrefix(command.S3.Prefix, command.Mongo.Database))

		for i, obj := range resp.Contents[:backupsToDeleteCount] {
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

func keepRelativeOplogBackups(ctx context.Context, s3Service *services.S3Service, command *DumpCommand) error {
	log.Info().Msg("Keep Relative Oplog Backups")

	listResp, err := s3Service.List(ctx, command.S3.Bucket, helpers.S3BackupPrefix(command.S3.Prefix, ""))

	if err != nil {
		return err
	}

	helpers.SortByKeyTimeStamp(listResp.Contents, helpers.S3BackupPrefix(command.S3.Prefix, command.Mongo.Database))

	oldestBackupKey := strings.TrimPrefix(*listResp.Contents[0].Key, helpers.S3BackupPrefix(command.S3.Prefix, ""))
	oldestBackupKey = strings.TrimSuffix(oldestBackupKey, ".gzip")
	oldestBackupKey = strings.TrimSuffix(oldestBackupKey, ".archive")

	oldestBackupTime, err := time.Parse(helpers.TimeFormat, oldestBackupKey)

	if err != nil {
		log.Error().Err(err).Msg("Failed to parse the oldest backup time")
		return err
	}

	// ######################
	// Get all oplog backups older than the oldest backup
	// ######################
	oplogBackupListResp, err := s3Service.List(ctx, command.S3.Bucket, helpers.S3OplogPrefix(command.S3.Prefix))
	if err != nil {
		return err
	}

	objectsToDelete := make([]types.ObjectIdentifier, 0)

	for _, obj := range oplogBackupListResp.Contents {
		if (*obj.Key) == helpers.S3OplogPrefix(command.S3.Prefix)+helpers.ConfigFileName {
			continue
		}

		fileKey := strings.TrimPrefix(*obj.Key, helpers.S3OplogPrefix(command.S3.Prefix))
		fileKey = strings.TrimSuffix(fileKey, ".tar.gz")

		fileKey = strings.Split(fileKey, "_")[1]
		toTimeOfLastBackup, err := time.Parse(helpers.TimeFormat, fileKey)

		if err != nil {
			log.Error().Err(err).Msgf("Failed to parse the time of %s", fileKey)
			return err
		}

		shouldKeepObject := toTimeOfLastBackup.After(oldestBackupTime)
		if !shouldKeepObject {
			objectsToDelete = append(objectsToDelete, types.ObjectIdentifier{Key: obj.Key})
		}
	}

	if err := s3Service.Delete(ctx, command.S3.Bucket, objectsToDelete); err != nil {
		return err
	}

	return nil
}

func getPreviousOplogRunData(ctx context.Context, s3Service *services.S3Service, command *DumpCommand) (*models.PreviousOplogRunInfo, error) {
	log.Info().Msg("Getting the latest oplog config")

	oplogKeyWithPrefix := helpers.S3OplogPrefix(command.S3.Prefix) + helpers.ConfigFileName

	resp, err := s3Service.Get(ctx, command.S3.Bucket, oplogKeyWithPrefix)

	if err != nil {
		var responseError *awsHttp.ResponseError

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

	var oplogConfig models.PreviousOplogRunInfo
	if err := json.NewDecoder(resp.Body).Decode(&oplogConfig); err != nil {
		log.Error().Err(err).Msg("Failed to decode the config file")
		return nil, err
	}

	log.Info().Msg("Got the latest oplog config")
	return &oplogConfig, nil
}
