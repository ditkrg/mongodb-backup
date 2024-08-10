package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/ditkrg/mongodb-backup/internal/cli/helpers"
	"github.com/ditkrg/mongodb-backup/internal/options"
	"github.com/ditkrg/mongodb-backup/internal/services"
	"github.com/rs/zerolog/log"
)

type FullRestore struct {
	Bucket string `required:"" name:"bucket" help:"The S3 bucket to list backups from."`
	Prefix string `optional:"" name:"prefix" help:"The prefix to filter backups."`
	Key    string `optional:"" name:"key" help:"The key of the backup to restore."`
}

func (restore FullRestore) Run() error {
	ctx := context.Background()
	s3Service := services.NewS3Service(options.Restore.S3)

	fmt.Println(options.Restore.MongoRestore.StopOnError)
	if restore.Key == "" {
		restore.Key = helpers.ChooseDatabaseToRestore(s3Service, ctx, restore.Bucket, restore.Prefix, func(key string) bool {
			return strings.Contains(key, "full_backups")
		})
	}

	backup, err := s3Service.Get(ctx, restore.Bucket, restore.Key)

	if err != nil {
		log.Err(err).Msgf("Failed to get backup %s", restore.Key)
		return err
	}

	defer backup.Body.Close()

	sections := strings.Split(restore.Key, "/")
	options.Restore.MongoRestore.BackupDir, _ = strings.CutSuffix(options.Restore.MongoRestore.BackupDir, "/")

	filePath := fmt.Sprintf("%s/%s", options.Restore.MongoRestore.BackupDir, sections[len(sections)-1])

	// log.Info().Msgf("Writing backup to %s", filePath)
	// outFile, err := os.Create(filePath)
	// if err != nil {
	// 	log.Err(err).Msg("Failed to create output file")
	// 	return err
	// }
	// defer outFile.Close()

	// _, err = io.Copy(outFile, backup.Body)
	// if err != nil {
	// 	log.Err(err).Msg("Failed to write backup to file")
	// 	return err
	// }

	log.Info().Msgf("Restoring backup %s", restore.Key)
	restoreOptions := options.Restore.MongoRestore.PrepareBackupMongoRestoreOptions(filePath)

	if err := restoreOptions.ParseAndValidateOptions(); err != nil {
		log.Err(err).Msg("Failed to parse and validate options")
		return err
	}

	result := restoreOptions.Restore()

	if result.Err != nil {
		log.Err(result.Err).Msg("Failed to restore backup")
		return result.Err
	}

	log.Info().Msg("Finished Restoring Backup")
	log.Info().Msgf("Successfully restored %d", result.Successes)
	log.Info().Msgf("Failed to restore %d", result.Failures)

	return nil
}
