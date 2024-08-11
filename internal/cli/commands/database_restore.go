package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/ditkrg/mongodb-backup/internal/cli/flags"
	"github.com/ditkrg/mongodb-backup/internal/cli/helpers"
	"github.com/ditkrg/mongodb-backup/internal/options"
	"github.com/ditkrg/mongodb-backup/internal/services"
	"github.com/rs/zerolog/log"
)

type DatabaseRestore struct {
	Key   string                  `optional:"" prefix:"s3-" help:"The key of the backup to restore."`
	S3    flags.S3Flags           `embed:"" prefix:"s3-" envprefix:"S3__" group:"Common S3 Flags:" `
	Mongo flags.MongoRestoreFlags `embed:"" prefix:"mongo-" envprefix:"MONGO_RESTORE__" group:"Common Mongo Restore Flags:"`
}

func (restore DatabaseRestore) Run() error {
	ctx := context.Background()

	options := options.S3Options{
		EndPoint:  restore.S3.EndPoint,
		AccessKey: restore.S3.AccessKey,
		SecretKey: restore.S3.SecretKey,
		Bucket:    restore.S3.Bucket,
		Prefix:    restore.S3.Prefix,
	}

	s3Service := services.NewS3Service(options)

	if restore.Key == "" {
		restore.Key = helpers.ChooseDatabaseToRestore(s3Service, ctx, restore.S3.Bucket, restore.S3.Prefix, func(key string) bool {
			return !strings.Contains(key, "oplog")
		})
	}

	backup, err := s3Service.Get(ctx, restore.S3.Bucket, restore.Key)

	if err != nil {
		log.Err(err).Msgf("Failed to get backup %s", restore.Key)
		return err
	}

	sections := strings.Split(restore.Key, "/")
	fileName := sections[len(sections)-1]
	filePath := fmt.Sprintf("%s/%s", restore.Mongo.BackupDir, fileName)
	restore.Mongo.BackupDir, _ = strings.CutSuffix(restore.Mongo.BackupDir, "/")

	if err := helpers.WriteToFile(backup.Body, filePath); err != nil {
		return err
	}

	log.Info().Msgf("Restoring backup %s", restore.Key)
	restoreOptions := restore.Mongo.PrepareBackupMongoRestoreOptions(filePath)

	if err := restoreOptions.ParseAndValidateOptions(); err != nil {
		log.Err(err).Msg("Failed to parse and validate options")
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
