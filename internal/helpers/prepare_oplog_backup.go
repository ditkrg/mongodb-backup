package helpers

import (
	"strings"
	"time"

	"github.com/ditkrg/mongodb-backup/internal/models"
	"github.com/rs/zerolog/log"
)

func PrepareOplogBackup(pitrKey string, prefix string) models.PitrBackup {

	fileName := strings.TrimPrefix(pitrKey, S3OplogPrefix(prefix))
	FileNameWithoutExtension := strings.TrimSuffix(fileName, ".tar.gz")
	timeStringArray := strings.Split(FileNameWithoutExtension, "_")
	var err error

	pitrBackup := models.PitrBackup{
		Key:                      pitrKey,
		FileName:                 fileName,
		FileNameWithoutExtension: FileNameWithoutExtension,
	}

	pitrBackup.FromTime, err = time.Parse(TimeFormat, timeStringArray[0])
	if err != nil {
		log.Panic().Err(err).Msg("Failed to parse time")
	}

	pitrBackup.ToTime, err = time.Parse(TimeFormat, timeStringArray[1])
	if err != nil {
		log.Panic().Err(err).Msg("Failed to parse time")
	}

	pitrBackup.FromString = pitrBackup.FromTime.Format(HumanReadableTimeFormat)
	pitrBackup.ToString = pitrBackup.ToTime.Format(HumanReadableTimeFormat)

	return pitrBackup
}
