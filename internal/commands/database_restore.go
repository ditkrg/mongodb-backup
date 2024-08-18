package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/charmbracelet/huh"
	"github.com/ditkrg/mongodb-backup/internal/flags"
	"github.com/ditkrg/mongodb-backup/internal/helpers"
	"github.com/ditkrg/mongodb-backup/internal/services"
	"github.com/mongodb/mongo-tools/mongorestore"
	"github.com/rs/zerolog/log"
)

type DatabaseRestoreCommand struct {
	Key   string                  `optional:"" prefix:"s3-" help:"The key of the backup to restore."`
	S3    flags.S3Flags           `embed:"" group:"Common S3 Flags:"`
	Mongo flags.MongoRestoreFlags `embed:"" envprefix:"MONGO_RESTORE__" group:"Common Mongo Restore Flags:"`
}

func (command DatabaseRestoreCommand) Run() error {
	ctx := context.Background()
	s3Service := services.NewS3Service(command.S3)

	var err error
	var restoreOptions *mongorestore.MongoRestore
	var backup *s3.GetObjectOutput

	backupDir := strings.TrimSuffix(command.Mongo.BackupDir, "/")

	// ########################
	// If key is not provided, let user choose the backup to restore
	// ########################
	if command.Key == "" {
		if command.Key, err = chooseDatabaseToRestore(s3Service, ctx, command.S3.Bucket, command.S3.Prefix); err != nil {
			return err
		}
	}

	// ########################
	// Get backup from S3
	// ########################
	if backup, err = s3Service.Get(ctx, command.S3.Bucket, command.Key); err != nil {
		return err
	}

	// ########################
	// Write backup to file
	// ########################
	filePath := fmt.Sprintf("%s/%d", backupDir, time.Now().Unix())

	if err := helpers.WriteToFile(backup.Body, filePath); err != nil {
		return err
	}

	log.Info().Msgf("Restoring backup %s", command.Key)

	if restoreOptions, err = command.Mongo.PrepareBackupMongoRestoreOptions(filePath); err != nil {
		log.Err(err).Msg("Failed to prepare restore options")
		return err
	}

	result := restoreOptions.Restore()

	if result.Err != nil {
		log.Err(result.Err).Msg("Failed to restore backup")
		return result.Err
	}

	log.Info().Msgf("Successfully restored %d", result.Successes)
	log.Info().Msgf("Failed to restore %d", result.Failures)

	return nil
}

func chooseDatabaseToRestore(s3Service *services.S3Service, ctx context.Context, bucket string, prefix string) (string, error) {
	var response *s3.ListObjectsV2Output
	var backupToRestore string
	var err error

	if response, err = s3Service.List(ctx, bucket, prefix); err != nil {
		return "", err
	}

	list := make([]huh.Option[string], 0)

	for _, object := range response.Contents {
		key := *object.Key
		if !strings.Contains(key, "oplog") {
			list = append(list, huh.NewOption(key, key))
		}
	}

	err = huh.NewSelect[string]().
		Title("Choose a backup to restore").
		Options(list...).
		Value(&backupToRestore).
		Run()

	if err != nil {
		log.Error().Err(err).Msg("Failed to choose backup to restore")
		return "", err
	}

	return backupToRestore, nil
}
