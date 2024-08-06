package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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

	file, err := os.Open(options.MongoDB.MongoDumpOptions.OutputOptions.Archive)

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open backup file")
	}

	defer file.Close()

	_, err = s3Service.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(options.S3.Bucket),
		Body:   file,
		Key:    aws.String(options.S3.GetBackupFilePath(options.MongoDB.DatabaseToBackup)),
	})

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to upload backup to S3")
	}

	log.Info().Msg("Uploaded backup to S3")
}

func (s3Service *S3Service) KeepMostRecentN(ctx context.Context, options *options.Options) {
	log.Info().Msgf("Keep Latest %d backups", options.S3.KeepRecentN)

	listObject := &s3.ListObjectsV2Input{
		Bucket: aws.String(options.S3.Bucket),
	}

	if options.S3.Prefix != "" {
		listObject.Prefix = aws.String(options.S3.GetBackupDirPath(options.MongoDB.DatabaseToBackup))
	}

	resp, err := s3Service.ListObjectsV2(ctx, listObject)

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

		deleteObjectsOutput, err := s3Service.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(options.S3.Bucket),
			Delete: &types.Delete{
				Objects: objectsToDelete,
			},
		})

		if err != nil {
			log.Fatal().Err(err).Msg("Failed to delete old backups from S3")
		}

		if len(deleteObjectsOutput.Deleted) != len(objectsToDelete) {
			err := fmt.Errorf("failed to delete all old backups from S3")
			log.Fatal().Err(err).Msg("Failed to delete old backups from S3")
		}

		if deleteObjectsOutput.Errors != nil {
			for _, err := range deleteObjectsOutput.Errors {
				s3Err := fmt.Errorf("error deleting object, code :%s, Key: %s, message :%s", *err.Code, *err.Key, *err.Message)
				log.Fatal().Err(s3Err).Msg("Failed to delete old backups from S3")
			}
		}

	}

	log.Info().Msg("Removed old backups from S3")
}

func (s3Service *S3Service) GetOplogConfig(ctx context.Context, options *options.Options) *models.OplogConfig {
	log.Info().Msg("Getting the latest oplog config")

	resp, err := s3Service.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(options.S3.Bucket),
		Key:    aws.String(options.S3.GetOplogConfigFilePath()),
	})

	if err != nil {
		var re *awshttp.ResponseError

		if ok := errors.As(err, &re); !ok {
			log.Fatal().Err(err).Msg("Failed to get config file from S3")
		}

		if re.ResponseError.HTTPStatusCode() == http.StatusNotFound {
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

func (s3Service *S3Service) UpdateOplogConfig(ctx context.Context, options *options.Options, oplogConfig *models.OplogConfig) {
	oplogConfigArray, err := json.Marshal(oplogConfig)

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to marshal oplog config")
	}

	_, err = s3Service.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(options.S3.Bucket),
		Key:    aws.String(options.S3.GetOplogConfigFilePath()),
		Body:   strings.NewReader(string(oplogConfigArray)),
	})

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get config file from S3")
	}
}

func (s3Service *S3Service) StartOplogBackup(ctx context.Context, options *options.Options) {
	log.Info().Msg("Uploading the oplog to S3")

	s3Dir := options.S3.CreateNewOplogBackupDir()

	err := filepath.Walk(options.MongoDB.BackupOutDir+"/local", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return s3Service.uploadOplogFile(ctx, options, s3Dir, path)
		}
		return nil
	})

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to upload oplog to S3")
	}

	log.Info().Msg("Uploaded oplog to S3")
}

func (s3Service *S3Service) uploadOplogFile(ctx context.Context, options *options.Options, s3Dir string, filePath string) error {
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
		Bucket: aws.String(options.S3.Bucket),
		Body:   file,
		Key:    aws.String(fmt.Sprintf("%s/%s", s3Dir, fileInfo.Name())),
	})

	if err != nil {
		return err
	}

	log.Info().Msgf("Uploaded %s to S3", fileInfo.Name())

	return nil
}

func (s3Service *S3Service) CheckBackupExists(ctx context.Context, options *options.Options) bool {
	log.Info().Msgf("Checking if a backup exists")

	resp, err := s3Service.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(options.S3.Bucket),
		Prefix: aws.String(options.S3.GetBackupDirPath("")),
	})

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to list objects in S3 bucket")
	}

	if len(resp.Contents) == 0 {
		log.Info().Msg("No backup exists. there must be a backup before starting the oplog backup")
		return false
	}

	log.Info().Msgf("Found %d backups in %s", len(resp.Contents), options.S3.Bucket)
	return true
}

func (s3Service *S3Service) KeepRelativeOplogBackups(ctx context.Context, options *options.Options) {
	log.Info().Msg("Keep Relative Oplog Backups")

	listObject := &s3.ListObjectsV2Input{
		Bucket: aws.String(options.S3.Bucket),
		Prefix: aws.String(options.S3.GetParentOplogBackupDir()),
	}

	resp, err := s3Service.ListObjectsV2(ctx, listObject)

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to list objects in S3 bucket")
	}

	backupsN := len(resp.Contents)

	print("backupsN: ", backupsN)
}
