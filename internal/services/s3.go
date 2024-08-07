package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
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

func NewS3Service(options options.S3Options) *S3Service {
	config := aws.Config{
		Credentials: credentials.NewStaticCredentialsProvider(options.AccessKey, options.SecretKey, ""),
		Region:      "us-east-1",
	}

	return &S3Service{
		Client: s3.NewFromConfig(config, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(options.EndPoint)
			o.UsePathStyle = true
		}),
	}
}

func (s3Service *S3Service) StartBackupUpload(ctx context.Context, options *options.Options) {
	log.Info().Msg("Uploading the backup to S3")

	s3Service.uploadFile(
		ctx,
		options.S3.Bucket,
		options.S3.BackupFilePath(options.MongoDB.DatabaseToBackup),
		options.MongoDB.MongoDumpOptions.OutputOptions.Archive,
	)

	log.Info().Msg("Uploaded backup to S3")
}

func (s3Service *S3Service) KeepMostRecentN(ctx context.Context, options *options.Options) {
	log.Info().Msgf("Keep Latest %d backups", options.S3.KeepRecentN)

	resp, err := s3Service.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(options.S3.Bucket),
		Prefix: aws.String(options.S3.BackupDirPath(options.MongoDB.DatabaseToBackup)),
	})

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to list objects in S3 bucket")
	}

	sort.Slice(resp.Contents, func(i, j int) bool {
		return resp.Contents[i].LastModified.After(*resp.Contents[j].LastModified)
	})

	backupsN := len(resp.Contents)

	if backupsN > options.S3.KeepRecentN {
		objectsToDelete := make([]types.ObjectIdentifier, backupsN-options.S3.KeepRecentN)

		for i, obj := range resp.Contents[options.S3.KeepRecentN:] {
			objectsToDelete[i] = types.ObjectIdentifier{
				Key: obj.Key,
			}
		}

		s3Service.deleteObjects(ctx, options.S3.Bucket, objectsToDelete)
	}

	log.Info().Msg("Removed old backups from S3")
}

func (s3Service *S3Service) GetOplogConfig(ctx context.Context, options *options.Options) *models.OplogConfig {
	log.Info().Msg("Getting the latest oplog config")

	key := fmt.Sprintf("%s/%s", options.S3.OplogDir(), helpers.ConfigFileName)

	resp, err := s3Service.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(options.S3.Bucket),
		Key:    aws.String(key),
	})

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

func (s3Service *S3Service) UploadOplogConfig(ctx context.Context, options *options.Options, oplogConfig *models.OplogConfig) {
	oplogConfigArray, err := json.Marshal(oplogConfig)

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to marshal oplog config")
	}

	_, err = s3Service.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(options.S3.Bucket),
		Key:    aws.String(fmt.Sprintf("%s/%s", options.S3.OplogDir(), helpers.ConfigFileName)),
		Body:   strings.NewReader(string(oplogConfigArray)),
	})

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get config file from S3")
	}
}

func (s3Service *S3Service) UploadOplog(ctx context.Context, options *options.Options) {
	log.Info().Msg("Uploading the oplog to S3")

	fileName := fmt.Sprintf("%s.tar.gz", time.Now().Format("060102-150405"))
	dirToTar := options.MongoDB.BackupOutDir + "/local"

	helpers.TarDirectory(dirToTar, fileName)

	s3Service.uploadFile(
		ctx,
		options.S3.Bucket,
		fmt.Sprintf("%s/%s", options.S3.OplogDir(), fileName),
		fmt.Sprintf("%s/%s", dirToTar, fileName),
	)

	os.Remove(fmt.Sprintf("%s/%s", dirToTar, fileName))

	log.Info().Msg("Uploaded oplog to S3")
}

func (s3Service *S3Service) ObjectExistsAt(ctx context.Context, bucket string, prefix string) bool {

	log.Info().Msgf("Checking if objects exists in %s/%s", bucket, prefix)

	resp, err := s3Service.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to list objects in S3 bucket")
	}

	if len(resp.Contents) == 0 {
		log.Info().Msgf("No objects found in %s/%s", bucket, prefix)
		return false
	}

	log.Info().Msgf("Found %d objects in %s", len(resp.Contents), bucket)

	return true
}

func (s3Service *S3Service) KeepRelativeOplogBackups(ctx context.Context, options *options.Options) {
	log.Info().Msg("Keep Relative Oplog Backups")

	var err error
	var backupResponse *s3.ListObjectsV2Output
	var oplogResponse *s3.ListObjectsV2Output

	// ######################
	// Get the oldest backup
	// ######################
	backupResponse, err = s3Service.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(options.S3.Bucket),
		Prefix: aws.String(options.S3.BackupDirPath("")),
	})

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to list objects in S3 bucket")
	}

	sort.Slice(backupResponse.Contents, func(i, j int) bool {
		return backupResponse.Contents[i].LastModified.Before(*backupResponse.Contents[j].LastModified)
	})

	oldestBackup := backupResponse.Contents[0].LastModified

	// ######################
	// Get all oplog backups older than the oldest backup
	// ######################
	oplogResponse, err = s3Service.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(options.S3.Bucket),
		Prefix: aws.String(options.S3.OplogDir()),
	})

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to list objects in S3 bucket")
	}

	objectsToDelete := make([]types.ObjectIdentifier, 0)

	oplogFileKey := aws.String(fmt.Sprintf("%s/%s", options.S3.OplogDir(), helpers.ConfigFileName))
	for _, obj := range oplogResponse.Contents {

		if obj.LastModified.Before(*oldestBackup) && obj.Key != oplogFileKey {
			objectsToDelete = append(objectsToDelete, types.ObjectIdentifier{Key: obj.Key})
		}
	}

	s3Service.deleteObjects(ctx, options.S3.Bucket, objectsToDelete)

	log.Info().Msg("Removed old oplog backups from S3")
}

func (s3Service *S3Service) uploadFile(ctx context.Context, bucket string, key string, filePath string) error {
	file, err := os.Open(filePath)

	if err != nil {
		return err
	}

	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	_, err = s3Service.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Body:   file,
		Key:    aws.String(key),
	})

	if err != nil {
		return err
	}

	log.Info().Msgf("Uploaded %s to S3", fileInfo.Name())

	return nil
}

func (s3Service *S3Service) deleteObjects(ctx context.Context, bucket string, objectsToDelete []types.ObjectIdentifier) {

	if len(objectsToDelete) == 0 {
		return
	}

	deleteObjectsOutput, err := s3Service.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(bucket),
		Delete: &types.Delete{
			Objects: objectsToDelete,
		},
	})

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to delete object from S3")
	}

	if len(deleteObjectsOutput.Deleted) != len(objectsToDelete) {
		err := fmt.Errorf("failed to delete all object from S3")
		log.Fatal().Err(err).Msg("Failed to delete object from S3")
	}

	if deleteObjectsOutput.Errors != nil {
		for _, err := range deleteObjectsOutput.Errors {
			s3Err := fmt.Errorf("error deleting object, code :%s, Key: %s, message :%s", *err.Code, *err.Key, *err.Message)
			log.Fatal().Err(s3Err).Msg("Failed to delete object from S3")
		}
	}
}
