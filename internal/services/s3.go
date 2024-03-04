package services

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ditkrg/mongodb-backup/internal/common"
	"github.com/rs/zerolog/log"
)

func StartBackupUpload(backupDir string, fileName string) error {

	log.Info().Msg("Uploading the backup to S3")

	// ######################
	// Prepare env variables
	// ######################
	s3AccessKey := common.GetRequiredEnv(common.S3__ACCESS_KEY)
	s3SecretAccessKey := common.GetRequiredEnv(common.S3__SECRET_ACCESS_KEY)
	s3Endpoint := common.GetRequiredEnv(common.S3__ENDPOINT)
	keepResentN := common.GetIntEnv(common.S3__KEEP_RESENT_N, 10)
	s3Bucket := common.GetRequiredEnv(common.S3__BUCKET)
	jobDir := common.GetRequiredEnv(common.S3__JOB_DIR)
	context := context.TODO()

	// ######################
	// Create S3 client
	// ######################
	config := aws.Config{
		Credentials: credentials.NewStaticCredentialsProvider(s3AccessKey, s3SecretAccessKey, ""),
		Region:      "us-east-1",
	}

	s3Client := s3.NewFromConfig(config, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(s3Endpoint)
		o.UsePathStyle = true
	})

	// ######################
	// Upload backup to S3
	// ######################
	file, err := os.Open(fmt.Sprintf("%s/%s", backupDir, fileName))

	if err != nil {
		log.Err(err).Msg("Failed to open backup file")
		return err
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Err(err).Msg("Failed to close the zip file")
		}
	}()

	_, err = s3Client.PutObject(context, &s3.PutObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(fmt.Sprintf("%s/%s", jobDir, fileName)),
		Body:   file,
	})

	if err != nil {
		log.Err(err).Msg("Failed to upload backup to S3")
		return err
	}

	// ######################
	// keep only the last N backups
	// ######################
	resp, err := s3Client.ListObjectsV2(context, &s3.ListObjectsV2Input{
		Bucket: aws.String(s3Bucket),
		Prefix: aws.String(fmt.Sprintf("%s/", jobDir)),
	})

	if err != nil {
		log.Err(err).Msg("Failed to list objects in S3 bucket")
		return err
	}

	sort.Slice(resp.Contents, func(i, j int) bool {
		return resp.Contents[i].LastModified.After(*resp.Contents[j].LastModified)
	})

	objectsToDelete := make([]types.ObjectIdentifier, len(resp.Contents)-keepResentN)

	for i, obj := range resp.Contents[keepResentN:] {
		objectsToDelete[i] = types.ObjectIdentifier{
			Key: obj.Key,
		}
	}

	deleteObjectsOutput, err := s3Client.DeleteObjects(context, &s3.DeleteObjectsInput{
		Bucket: aws.String(s3Bucket),
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
		log.Err(err).Send()
		return err
	}

	if deleteObjectsOutput.Errors != nil {
		for _, err := range deleteObjectsOutput.Errors {
			s3Err := fmt.Errorf("error deleting object, code :%s, Key: %s, message :%s", *err.Code, *err.Key, *err.Message)
			log.Err(s3Err).Send()
		}

		return fmt.Errorf("failed to delete old backups from S3")
	}

	log.Info().Msg("Uploaded backup to S3")
	return nil
}
