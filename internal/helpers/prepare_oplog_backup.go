package helpers

import (
	"strings"
	"time"

	"github.com/ditkrg/mongodb-backup/internal/models"
	"github.com/rs/zerolog/log"
)

func PrepareOplogBackup(oplogKey string, prefix string) models.OplogBackup {

	fileName := strings.TrimPrefix(oplogKey, S3OplogPrefix(prefix))
	FileNameWithoutExtension := strings.TrimSuffix(fileName, ".tar.gz")
	timeStringArray := strings.Split(FileNameWithoutExtension, "_")
	var err error

	oplogBackup := models.OplogBackup{
		Key:                      oplogKey,
		FileName:                 fileName,
		FileNameWithoutExtension: FileNameWithoutExtension,
	}

	oplogBackup.FromTime, err = time.Parse(TimeFormat, timeStringArray[0])
	if err != nil {
		log.Panic().Err(err).Msg("Failed to parse time")
	}

	oplogBackup.ToTime, err = time.Parse(TimeFormat, timeStringArray[1])
	if err != nil {
		log.Panic().Err(err).Msg("Failed to parse time")
	}

	oplogBackup.FromString = oplogBackup.FromTime.Format(HumanReadableTimeFormat)
	oplogBackup.ToString = oplogBackup.ToTime.Format(HumanReadableTimeFormat)

	return oplogBackup
}
