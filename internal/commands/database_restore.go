package commands

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/charmbracelet/huh"
	"github.com/ditkrg/mongodb-backup/internal/flags"
	"github.com/ditkrg/mongodb-backup/internal/helpers"
	"github.com/ditkrg/mongodb-backup/internal/models"
	"github.com/ditkrg/mongodb-backup/internal/services"
	"github.com/mongodb/mongo-tools/mongorestore"
	"github.com/rs/zerolog/log"
)

type DatabaseRestoreCommand struct {
	Key       string                  `optional:"" prefix:"s3-" help:"The key of the backup to restore."`
	S3        flags.S3Flags           `embed:"" group:"S3 Flags:"`
	Mongo     flags.MongoRestoreFlags `embed:"" envprefix:"MONGO_RESTORE__"`
	Verbosity flags.VerbosityFlags    `embed:"" prefix:"verbosity-" envprefix:"VERBOSITY__" group:"verbosity options"`
}

func (command DatabaseRestoreCommand) Run() error {
	command.Verbosity.SetGlobalLogLevel()

	ctx := context.Background()
	s3Service := services.NewS3Service(command.S3)

	var err error
	var mongoRestore *mongorestore.MongoRestore
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
	fileName := strconv.FormatInt(time.Now().Unix(), 10)
	if err := helpers.WriteToFile(backup.Body, backup.ContentLength, backupDir, fileName); err != nil {
		return err
	}

	log.Info().Msgf("Restoring backup %s", command.Key)

	// ########################
	// Prepare restore options
	// ########################
	if mongoRestore, err = command.Mongo.PrepareBackupMongoRestoreOptions(filepath.Join(backupDir, fileName)); err != nil {
		log.Err(err).Msg("Failed to prepare restore options")
		return err
	}

	// ########################
	// Restore backup
	// ########################
	result := mongoRestore.Restore()

	if result.Err != nil {
		log.Err(result.Err).Msg("Failed to restore backup")
		return result.Err
	}

	log.Info().Msgf("Successfully restored %d, Failed to restore %d", result.Successes, result.Failures)

	if !command.Mongo.InputOptions.OplogReplay {
		return nil
	}

	log.Info().Msg("Restoring Oplog")

	if err := command.RestoreOplog(ctx, s3Service, command.Key); err != nil {
		log.Err(err).Msg("Failed to restore oplog")
		return err
	}

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

func (command *DatabaseRestoreCommand) RestoreOplog(ctx context.Context, s3Service *services.S3Service, keyRestored string) error {

	keyPath := helpers.S3BackupPrefix(command.S3.Prefix, "")

	backupRestoreTime := strings.TrimPrefix(keyRestored, keyPath)
	backupRestoreTime = strings.TrimSuffix(backupRestoreTime, ".gzip")
	backupRestoreTime = strings.TrimSuffix(backupRestoreTime, ".archive")

	// ###############################
	// List all the backups
	// ###############################
	resp, err := s3Service.List(
		ctx,
		command.S3.Bucket,
		helpers.S3OplogPrefix(command.S3.Prefix),
	)

	if err != nil {
		return err
	}

	if len(resp.Contents) == 0 {
		log.Info().Msg("No Oplog backups found")
		return nil
	}

	// ###############################
	// Filter out the config file
	// ###############################
	resp.Contents = slices.DeleteFunc(resp.Contents, func(i types.Object) bool {
		return strings.Contains(*i.Key, helpers.ConfigFileName)
	})

	// ###############################
	// Prepare the list of oplog backups and oplog backups to be restored
	// ###############################
	oplogBackupList := make([]models.OplogBackup, len(resp.Contents))
	oplogToRestore := make([]models.OplogBackup, 0)

	// ###############################
	// Change backups to models.oplogBackup
	// ###############################
	for i, obj := range resp.Contents {
		key := *obj.Key
		oplogBackupList[i] = helpers.PrepareOplogBackup(key, command.S3.Prefix)
	}

	// ###############################
	// Sort the backups by ToTime
	// ###############################
	sort.Slice(oplogBackupList, func(i, j int) bool {
		iToTime := oplogBackupList[i].ToTime
		jToTime := oplogBackupList[j].ToTime
		return iToTime.Before(jToTime)
	})

	// ###############################
	// change the oplog limits to time
	// ###############################
	oplogLimitToTime, err := getOplogLimit(command)
	if err != nil {
		return err
	}

	// ###############################
	// parse the oplog backup from time to time
	// ###############################
	time, err := time.Parse(helpers.TimeFormat, backupRestoreTime)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse backup time into a from Limit")
		return err
	}
	oplogLimitFromTime := &time

	// ###############################
	// prepare the directories
	// ###############################
	downloadsDir := filepath.Join(command.Mongo.BackupDir, "downloads")
	restoreDir := filepath.Join(command.Mongo.BackupDir, "toBeRestored")
	outputDir := filepath.Join(restoreDir, "local")

	// ###############################
	// Download the backups
	// ###############################
	for _, oplogBackup := range oplogBackupList {

		// ###############################
		// Check if the backup should be restored
		// ###############################
		if !shouldRestoreBackup(oplogLimitFromTime, oplogLimitToTime, oplogBackup) {
			switch {
			case oplogLimitToTime != nil:
				log.Info().Msgf("Skipping backup %s as it is not in the range %s ~ %s", oplogBackup.Key, backupRestoreTime, command.Mongo.InputOptions.OplogLimit)
			default:
				log.Info().Msgf("Skipping backup %s as it is before %s", oplogBackup.Key, backupRestoreTime)

			}
			continue
		}

		// ###############################
		// Add the backup to the list of backups to be restored
		// ###############################
		oplogToRestore = append(oplogToRestore, oplogBackup)

		// ###############################
		// Download the backup
		// ###############################
		obj, err := s3Service.Get(ctx, command.S3.Bucket, oplogBackup.Key)
		if err != nil {
			return err
		}

		// ###############################
		// Write the backup to file
		// ###############################
		if err := helpers.WriteToFile(obj.Body, obj.ContentLength, downloadsDir, oplogBackup.FileName); err != nil {
			return err
		}
	}

	// ###############################
	// Restore the backups
	// ###############################
	for _, oplogBackup := range oplogToRestore {
		tarPath := filepath.Join(downloadsDir, oplogBackup.FileName)

		// ###############################
		// Extract the tar file
		// ###############################
		if err := helpers.ExtractTar(tarPath, outputDir); err != nil {
			return err
		}

		// ###############################
		// Restore the tar file
		// ###############################
		if err := os.RemoveAll(tarPath); err != nil {
			log.Error().Err(err).Msgf("failed to remove %s", tarPath)
		}

		// ###############################
		// Prepare the mongodb options
		// ###############################
		oplogOptions, err := command.Mongo.PrepareOplogMongoRestoreOptions(restoreDir, oplogLimitToTime)
		if err != nil {
			return err
		}

		// ###############################
		// Restore the oplog
		// ###############################
		log.Info().Msg("start mongodb restore")
		result := oplogOptions.Restore()

		if result.Err != nil {
			log.Err(result.Err).Msg("Failed to restore oplog")
			return result.Err
		}

		// ###############################
		// Remove the output directory
		// ###############################
		if err := os.RemoveAll(outputDir); err != nil {
			log.Error().Err(err).Msgf("failed to remove %s", outputDir)
		}
	}

	return nil
}

func getOplogLimit(command *DatabaseRestoreCommand) (*time.Time, error) {
	if command.Mongo.InputOptions.OplogLimit != "" {
		log.Info().Msg("Parsing oplog limits")
	}

	var oplogLimitToTime *time.Time = nil

	if command.Mongo.InputOptions.OplogLimit != "" {
		time, err := time.Parse(helpers.TimeFormat, command.Mongo.InputOptions.OplogLimit)
		if err != nil {
			log.Error().Err(err).Msg("Failed to parse provided To Limit")
			return nil, err
		}
		oplogLimitToTime = &time
	}

	return oplogLimitToTime, nil
}

func shouldRestoreBackup(from *time.Time, to *time.Time, backup models.OplogBackup) bool {
	if from != nil && to != nil {
		return (from.After(backup.FromTime) && from.Before(backup.ToTime)) ||
			(to.After(backup.FromTime) && to.Before(backup.ToTime)) ||
			(backup.FromTime.After(*from) && backup.ToTime.Before(*to))

	}

	if from != nil {
		return from.Before(backup.ToTime)
	}

	if to != nil {
		return to.After(backup.FromTime)
	}

	return true
}
