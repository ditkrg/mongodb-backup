package services

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/rs/zerolog/log"
)

func (s3Service *S3Service) List(ctx context.Context, bucket string, prefix string) *s3.ListObjectsV2Output {
	log.Info().Msgf("Listing objects in %s/%s", bucket, prefix)

	resp, err := s3Service.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to list objects in S3 bucket")
	}

	log.Info().Msgf("successfully listed objects in %s/%s", bucket, prefix)
	return resp
}

func (s3Service *S3Service) ExistsAt(ctx context.Context, bucket string, prefix string) bool {

	log.Info().Msgf("Checking if objects exists in %s/%s", bucket, prefix)

	resp := s3Service.List(ctx, bucket, prefix)

	if len(resp.Contents) == 0 {
		log.Info().Msgf("No objects found in %s/%s", bucket, prefix)
		return false
	}

	log.Info().Msgf("Found %d objects in %s", len(resp.Contents), bucket)
	return true
}

func (s3Service *S3Service) UploadFile(ctx context.Context, bucket string, key string, filePath string) {
	log.Info().Msgf("Uploading %s to S3", filePath)

	file, err := os.Open(filePath)

	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to open %s", filePath)
	}

	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to get file %s info", filePath)
	}

	_, err = s3Service.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Body:   file,
		Key:    aws.String(key),
	})

	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to open %s", filePath)
	}

	log.Info().Msgf("Uploaded %s to S3", fileInfo.Name())
}

func (s3Service *S3Service) UploadByteArray(ctx context.Context, bucket string, key string, byteArray []byte) {
	log.Info().Msgf("Uploading content to s3 %s", key)

	_, err := s3Service.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   strings.NewReader(string(byteArray)),
	})

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to upload")
	}

	log.Info().Msgf("Uploaded content to S3 %s", key)
}

func (s3Service *S3Service) Delete(ctx context.Context, bucket string, objectsToDelete []types.ObjectIdentifier) {

	if len(objectsToDelete) == 0 {
		return
	}

	log.Info().Msgf("Deleting %d objects from S3", len(objectsToDelete))

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

	log.Info().Msgf("Deleted %d objects from S3", len(objectsToDelete))
}

func (s3Service *S3Service) Get(ctx context.Context, bucket string, key string) (*s3.GetObjectOutput, error) {
	log.Info().Msgf("Getting object from S3 %s", key)

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
