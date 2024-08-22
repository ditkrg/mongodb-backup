package services

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ditkrg/mongodb-backup/internal/flags"
	"github.com/rs/zerolog/log"
)

type S3Service struct {
	*s3.Client
}

func NewS3Service(s3Options flags.S3Flags) *S3Service {
	config := aws.Config{
		Credentials: credentials.NewStaticCredentialsProvider(s3Options.AccessKey, s3Options.SecretKey, ""),
		Region:      "us-east-1",
	}

	return &S3Service{
		Client: s3.NewFromConfig(config, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(s3Options.EndPoint)
			o.UsePathStyle = true
		}),
	}
}

func (s3Service *S3Service) List(ctx context.Context, bucket string, prefix string) (*s3.ListObjectsV2Output, error) {
	log.Info().Msgf("Listing objects in %s/%s", bucket, prefix)

	resp, err := s3Service.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to list objects in S3 bucket")
		return nil, err
	}

	log.Info().Msgf("successfully listed objects in %s/%s", bucket, prefix)
	return resp, nil
}

func (s3Service *S3Service) ObjectsExistsAt(ctx context.Context, bucket string, prefix string) (bool, error) {
	log.Info().Msgf("Checking if objects exists in %s/%s", bucket, prefix)

	var resp *s3.ListObjectsV2Output
	var err error

	if resp, err = s3Service.List(ctx, bucket, prefix); err != nil {
		return false, err
	}

	if len(resp.Contents) == 0 {
		log.Info().Msgf("No objects found in %s/%s", bucket, prefix)
		return false, nil
	}

	log.Info().Msgf("Found %d objects in %s/%s", len(resp.Contents), bucket, prefix)
	return true, nil
}

func (s3Service *S3Service) UploadFile(ctx context.Context, bucket string, key string, filePath string) error {
	log.Info().Msgf("Uploading file %s to S3", filePath)

	file, err := os.Open(filePath)

	if err != nil {
		log.Error().Err(err).Msgf("Failed to open %s", filePath)
		return err
	}

	defer file.Close()

	if _, err := s3Service.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Body:   file,
		Key:    aws.String(key),
	}); err != nil {
		log.Error().Err(err).Msgf("Failed to open %s", filePath)
		return err
	}

	log.Info().Msgf("Uploaded %s to S3", filePath)
	return nil
}

func (s3Service *S3Service) Delete(ctx context.Context, bucket string, objectsToDelete []types.ObjectIdentifier) error {
	if len(objectsToDelete) == 0 {
		log.Info().Msg("No objects to delete from S3")
		return nil
	}

	log.Info().Msgf("Deleting %d objects from S3", len(objectsToDelete))

	deleteObjectsOutput, err := s3Service.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(bucket),
		Delete: &types.Delete{
			Objects: objectsToDelete,
		},
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to delete object from S3")
		return err
	}

	if len(deleteObjectsOutput.Deleted) != len(objectsToDelete) {
		err := fmt.Errorf("failed to delete all object from S3")
		log.Error().Err(err).Send()
		return err
	}

	if deleteObjectsOutput.Errors != nil {
		for _, err := range deleteObjectsOutput.Errors {
			s3Err := fmt.Errorf("error deleting object, code :%s, Key: %s, message :%s", *err.Code, *err.Key, *err.Message)
			log.Error().Err(s3Err).Send()
		}

		return errors.New("failed to delete all objects from S3")
	}

	log.Info().Msgf("Deleted %d objects from S3", len(objectsToDelete))
	return nil
}

func (s3Service *S3Service) Get(ctx context.Context, bucket string, key string) (*s3.GetObjectOutput, error) {
	log.Info().Msgf("Getting object %s from S3", key)

	resp, err := s3Service.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		log.Err(err).Msgf("Failed to get object %s from S3", key)
		return nil, err
	}

	log.Info().Msgf("Got object %s from S3", key)
	return resp, nil
}
