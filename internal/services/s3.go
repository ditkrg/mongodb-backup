package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ditkrg/mongodb-backup/internal/helpers"
	"github.com/ditkrg/mongodb-backup/internal/models"
	"github.com/ditkrg/mongodb-backup/internal/options"
	"github.com/rs/zerolog/log"
)

type S3Service struct {
	*s3.Client
}

func NewS3Service() *S3Service {
	config := aws.Config{
		Credentials: credentials.NewStaticCredentialsProvider(options.Config.S3.AccessKey, options.Config.S3.SecretKey, ""),
		Region:      "us-east-1",
	}

	return &S3Service{
		Client: s3.NewFromConfig(config, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(options.Config.S3.EndPoint)
			o.UsePathStyle = true
		}),
	}
}

func (s3Service *S3Service) UploadBackup(ctx context.Context) {
	log.Info().Msg("Uploading the backup to S3")

	s3Service.UploadFile(
		ctx,
		options.Config.S3.Bucket,
		options.Config.S3.BackupFilePath(options.Config.MongoDB.DatabaseToBackup),
		options.Config.MongoDB.MongoDumpOptions.OutputOptions.Archive,
	)

	log.Info().Msg("Uploaded backup to S3")
}

func (s3Service *S3Service) KeepRecentBackups(ctx context.Context) {
	log.Info().Msg("Keep Recent Backups")

	resp := s3Service.List(
		ctx,
		options.Config.S3.Bucket,
		options.Config.S3.BackupDirPath(options.Config.MongoDB.DatabaseToBackup),
	)

	s3BackupCount := len(resp.Contents)

	log.Info().Msgf("Found %d backups, max backup to keep %d", s3BackupCount, options.Config.S3.KeepRecentN)

	if s3BackupCount > options.Config.S3.KeepRecentN {
		backupsToDeleteCount := s3BackupCount - options.Config.S3.KeepRecentN
		objectsToDelete := make([]types.ObjectIdentifier, backupsToDeleteCount)

		sort.Slice(resp.Contents, func(i, j int) bool {
			return resp.Contents[i].LastModified.After(*resp.Contents[j].LastModified)
		})

		for i, obj := range resp.Contents[options.Config.S3.KeepRecentN:] {
			objectsToDelete[i] = types.ObjectIdentifier{
				Key: obj.Key,
			}
		}

		s3Service.Delete(ctx, options.Config.S3.Bucket, objectsToDelete)
	}
}

func (s3Service *S3Service) GetOplogConfig(ctx context.Context) *models.OplogConfig {
	log.Info().Msg("Getting the latest oplog config")

	key := fmt.Sprintf("%s/%s", options.Config.S3.OplogDir(), helpers.ConfigFileName)

	resp, err := s3Service.Get(ctx, options.Config.S3.Bucket, key)

	if err != nil {
		var responseError *awshttp.ResponseError

		if ok := errors.As(err, &responseError); !ok {
			log.Fatal().Err(err).Msg("Failed to get config file from S3")
		}

		if responseError.ResponseError.HTTPStatusCode() == http.StatusNotFound {
			return nil
		}

		log.Fatal().Err(err).Msg("Failed to get config file from S3")
	}

	defer resp.Body.Close()

	var oplogConfig models.OplogConfig
	if err := json.NewDecoder(resp.Body).Decode(&oplogConfig); err != nil {
		log.Fatal().Err(err).Msg("Failed to decode the config file")
	}

	log.Info().Msg("Got the latest oplog config")
	return &oplogConfig
}

func (s3Service *S3Service) UploadOplogConfig(ctx context.Context, oplogConfig *models.OplogConfig) {
	oplogConfigArray, err := json.Marshal(oplogConfig)

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to marshal oplog config")
	}

	s3Service.UploadByteArray(
		ctx,
		options.Config.S3.Bucket,
		fmt.Sprintf("%s/%s", options.Config.S3.OplogDir(), helpers.ConfigFileName),
		oplogConfigArray,
	)
}

func (s3Service *S3Service) UploadOplog(ctx context.Context) {
	fileName := fmt.Sprintf("%s.tar.gz", time.Now().Format("060102-150405"))
	dirToTar := options.Config.MongoDB.BackupOutDir + "/local"

	helpers.TarDirectory(dirToTar, fileName)

	s3Service.UploadFile(
		ctx,
		options.Config.S3.Bucket,
		fmt.Sprintf("%s/%s", options.Config.S3.OplogDir(), fileName),
		fmt.Sprintf("%s/%s", dirToTar, fileName),
	)

	os.Remove(fmt.Sprintf("%s/%s", dirToTar, fileName))
}

func (s3Service *S3Service) KeepRelativeOplogBackups(ctx context.Context) {
	log.Info().Msg("Keep Relative Oplog Backups")

	// ######################
	// Get the oldest backup
	// ######################
	backupResponse := s3Service.List(
		ctx,
		options.Config.S3.Bucket,
		options.Config.S3.BackupDirPath(""),
	)

	sort.Slice(backupResponse.Contents, func(i, j int) bool {
		return backupResponse.Contents[i].LastModified.Before(*backupResponse.Contents[j].LastModified)
	})

	oldestBackup := backupResponse.Contents[0].LastModified

	// ######################
	// Get all oplog backups older than the oldest backup
	// ######################
	objectsToDelete := make([]types.ObjectIdentifier, 0)
	oplogFileKey := aws.String(fmt.Sprintf("%s/%s", options.Config.S3.OplogDir(), helpers.ConfigFileName))

	oplogResponse := s3Service.List(ctx, options.Config.S3.Bucket, options.Config.S3.OplogDir())

	for _, obj := range oplogResponse.Contents {
		if obj.LastModified.Before(*oldestBackup) && obj.Key != oplogFileKey {
			objectsToDelete = append(objectsToDelete, types.ObjectIdentifier{Key: obj.Key})
		}
	}

	s3Service.Delete(ctx, options.Config.S3.Bucket, objectsToDelete)
}
