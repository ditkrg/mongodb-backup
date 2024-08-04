package services

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ditkrg/mongodb-backup/internal/options"
	"github.com/rs/zerolog/log"
)

type S3Service struct {
	*s3.Client
}

func InitS3Service(options options.S3Options) *S3Service {
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

func (s3Service *S3Service) StartBackupUpload(ctx context.Context, options options.Options) error {
	log.Info().Msg("Uploading the backup to S3")

	file, err := os.Open(options.MongoDB.BackupOutFilePath)

	if err != nil {
		log.Err(err).Msg("Failed to open backup file")
		return err
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Err(err).Msg("Failed to close the zip file")
		}
	}()

	var backupFileName string

	if options.MongoDB.DatabaseToBackup == "" {
		backupFileName = fmt.Sprintf("%s_%s.gzip", "all_databases", time.Now().Format("060102-150405"))
	} else {
		backupFileName = fmt.Sprintf("%s_%s.gzip", options.MongoDB.DatabaseToBackup, time.Now().Format("060102-150405"))
	}

	_, err = s3Service.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(options.S3.Bucket),
		Key:    aws.String(fmt.Sprintf("%s/%s", options.S3.Prefix, backupFileName)),
		Body:   file,
	})

	if err != nil {
		log.Err(err).Msg("Failed to upload backup to S3")
		return err
	}

	log.Info().Msg("Uploaded backup to S3")

	return nil
}

func (s3Service *S3Service) KeepMostRecentN(ctx context.Context, options options.Options) error {
	log.Info().Msgf("Keep Latest %d backups", options.S3.KeepRecentN)

	resp, err := s3Service.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(options.S3.Bucket),
		Prefix: aws.String(fmt.Sprintf("%s/", options.S3.Prefix)),
	})

	if err != nil {
		log.Err(err).Msg("Failed to list objects in S3 bucket")
		return err
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

		deleteObjectsOutput, err := s3Service.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(options.S3.Bucket),
			Delete: &types.Delete{
				Objects: objectsToDelete,
			},
		})

		if err != nil {
			log.Err(err).Msg("Failed to delete old backups from S3")
			return err
		}

		if len(deleteObjectsOutput.Deleted) != len(objectsToDelete) {
			err := fmt.Errorf("failed to delete all old backups from S3")
			log.Err(err).Msg("Failed to delete old backups from S3")
			return err
		}

		if deleteObjectsOutput.Errors != nil {
			for _, err := range deleteObjectsOutput.Errors {
				s3Err := fmt.Errorf("error deleting object, code :%s, Key: %s, message :%s", *err.Code, *err.Key, *err.Message)
				log.Err(s3Err).Msg("Failed to delete old backups from S3")
			}

			return fmt.Errorf("failed to delete old backups from S3")
		}

	}

	log.Info().Msg("Removed old backups from S3")

	return nil
}
