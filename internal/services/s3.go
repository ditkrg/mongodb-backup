package services

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ditkrg/mongodb-backup/internal/common"
)

func StartBackupUpload(backupDir string) error {
	s3AccessKey := common.GetRequiredEnv(common.S3__ACCESS_KEY)
	s3SecretAccessKey := common.GetRequiredEnv(common.S3__SECRET_ACCESS_KEY)
	s3Endpoint := common.GetRequiredEnv(common.S3__ENDPOINT)
	s3Bucket := common.GetRequiredEnv(common.S3__BUCKET)
	jobDir := common.GetRequiredEnv(common.S3__JOB_DIR)

	config := aws.Config{
		Credentials: credentials.NewStaticCredentialsProvider(s3AccessKey, s3SecretAccessKey, ""),
		Region:      "us-east-1",
	}

	s3Client := s3.NewFromConfig(config, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(s3Endpoint)
		o.UsePathStyle = true
	})

	return nil
}
