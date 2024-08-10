package helpers

import (
	"context"

	"github.com/charmbracelet/huh"
	"github.com/ditkrg/mongodb-backup/internal/services"
	"github.com/rs/zerolog/log"
)

func ChooseDatabaseToRestore(s3Service *services.S3Service, ctx context.Context, bucket string, prefix string, validator func(string) bool) (backupToRestore string) {
	response := s3Service.List(ctx, bucket, prefix)

	list := make([]huh.Option[string], 0)

	for _, object := range response.Contents {
		key := *object.Key
		if validator(key) {
			list = append(list, huh.NewOption(key, key))
		}
	}

	err := huh.NewSelect[string]().
		Title("Choose a backup to restore").
		Options(list...).
		Value(&backupToRestore).
		Run()

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to choose backup to restore")
	}

	return backupToRestore
}
