package commands

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ditkrg/mongodb-backup/internal/flags"
	"github.com/ditkrg/mongodb-backup/internal/helpers"
	"github.com/ditkrg/mongodb-backup/internal/models"
	"github.com/ditkrg/mongodb-backup/internal/services"
	"github.com/rs/zerolog/log"
)

type PitrRestoreCommand struct {
	S3    flags.S3Flags               `embed:"" group:"Common S3 Flags:"`
	Mongo flags.MongoPitrRestoreFlags `embed:"" prefix:"mongo-" envprefix:"MONGO_PITR__" group:"Common MongoDB Flags:" `
}

func (command *PitrRestoreCommand) Run() error {

	// ###############################
	// Initialize the S3 Service
	// ###############################
	ctx := context.Background()
	s3Service := services.NewS3Service(command.S3)

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
	// Prepare the list of backups model and backups to be restored
	// ###############################
	pitrBackupList := make([]models.PitrBackup, len(resp.Contents))
	pitrToRestore := make([]models.PitrBackup, 0)

	// ###############################
	// Change backups to models.PitrBackup
	// ###############################
	for i, obj := range resp.Contents {
		key := *obj.Key
		pitrBackupList[i] = helpers.PrepareOplogBackup(key, command.S3.Prefix)
	}

	// ###############################
	// Sort the backups by ToTime
	// ###############################
	sort.Slice(pitrBackupList, func(i, j int) bool {
		iToTime := pitrBackupList[i].ToTime
		jToTime := pitrBackupList[j].ToTime
		return iToTime.Before(jToTime)
	})

	// ###############################
	// change the oplog limits to time
	// ###############################
	oplogLimitFromTime, oplogLimitToTime, err := getOplogLimit(command)
	if err != nil {
		return err
	}

	// ###############################
	// prepare the directories
	// ###############################
	downloadsDir := filepath.Join(command.Mongo.BackupDir, "downloads")
	restoreDir := filepath.Join(command.Mongo.BackupDir, "toBeRestored")
	outputDir := filepath.Join(restoreDir, "local")

	// ###############################
	// Download the backups
	// ###############################
	for _, pitrBackup := range pitrBackupList {

		// ###############################
		// Check if the backup should be restored
		// ###############################
		if !shouldRestoreBackup(oplogLimitFromTime, oplogLimitToTime, pitrBackup) {

			switch {
			case oplogLimitFromTime != nil && oplogLimitToTime != nil:
				log.Info().Msgf("Skipping backup %s as it is not in the range %s ~ %s", pitrBackup.Key, command.Mongo.OplogLimitFrom, command.Mongo.OplogLimitTo)
			case oplogLimitFromTime != nil:
				log.Info().Msgf("Skipping backup %s as it is before %s", pitrBackup.Key, command.Mongo.OplogLimitFrom)
			case oplogLimitToTime != nil:
				log.Info().Msgf("Skipping backup %s as it is after %s", pitrBackup.Key, command.Mongo.OplogLimitTo)
			default:
				log.Info().Msgf("Skipping backup %s as it is not in the range", pitrBackup.Key)
			}

			continue
		}

		// ###############################
		// Add the backup to the list of backups to be restored
		// ###############################
		pitrToRestore = append(pitrToRestore, pitrBackup)

		// ###############################
		// Download the backup
		// ###############################
		obj, err := s3Service.Get(ctx, command.S3.Bucket, pitrBackup.Key)
		if err != nil {
			return err
		}

		// ###############################
		// Write the backup to file
		// ###############################
		if err := helpers.WriteToFile(obj.Body, downloadsDir, pitrBackup.FileName); err != nil {
			return err
		}
	}

	// ###############################
	// Restore the backups
	// ###############################
	for _, pitrBackup := range pitrToRestore {
		tarPath := filepath.Join(downloadsDir, pitrBackup.FileName)

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
		pitrOptions, err := command.Mongo.PrepareOplogMongoRestoreOptions(restoreDir, oplogLimitToTime)
		if err != nil {
			return err
		}

		// ###############################
		// Restore the oplog
		// ###############################
		log.Info().Msg("start mongodb restore")
		result := pitrOptions.Restore()

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

func getOplogLimit(command *PitrRestoreCommand) (*time.Time, *time.Time, error) {
	if command.Mongo.OplogLimitFrom != "" || command.Mongo.OplogLimitTo != "" {
		log.Info().Msg("Parse oplog limits")
	}

	var oplogLimitFromTime *time.Time = nil
	var oplogLimitToTime *time.Time = nil

	if command.Mongo.OplogLimitFrom != "" {
		time, err := time.Parse(helpers.HumanReadableTimeFormat, command.Mongo.OplogLimitFrom)
		if err != nil {
			log.Error().Err(err).Msg("Failed to parse From Limit")
			return nil, nil, err
		}

		oplogLimitFromTime = &time
	}

	if command.Mongo.OplogLimitTo != "" {
		time, err := time.Parse(helpers.HumanReadableTimeFormat, command.Mongo.OplogLimitTo)
		if err != nil {
			log.Error().Err(err).Msg("Failed to parse To Limit")
			return nil, nil, err
		}
		oplogLimitToTime = &time
	}

	return oplogLimitFromTime, oplogLimitToTime, nil
}

func shouldRestoreBackup(from *time.Time, to *time.Time, backup models.PitrBackup) bool {
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

// func pickTime(from string, to string) (*time.Time, error) {
// 	var timeString string

// 	style := lipGloss.
// 		NewStyle().
// 		Foreground(lipGloss.Color("#FF8800")).
// 		Bold(true)

// 	message := fmt.Sprintf("please type a time between %s and %s to restore oplog, type the input in the following format %s", style.Render(from), style.Render(to), style.Render(helpers.HumanReadableTimeFormat))
// 	fmt.Println(message)

// 	if err := huh.NewInput().Value(&timeString).Run(); err != nil {
// 		log.Error().Err(err).Msg("Failed to get oplog limit")
// 		return nil, err
// 	}

// 	time, err := time.Parse(helpers.HumanReadableTimeFormat, timeString)
// 	if err != nil {
// 		log.Error().Err(err).Msg("Failed to parse time")
// 		return nil, err
// 	}

// 	return &time, err
// }
