package helpers

import (
	"io"
	"os"

	"github.com/rs/zerolog/log"
)

func WriteToFile(body io.ReadCloser, filePath string) error {
	defer body.Close()

	log.Info().Msgf("Writing backup to %s", filePath)

	outFile, err := os.Create(filePath)

	if err != nil {
		log.Err(err).Msg("Failed to create output file")
		return err
	}

	defer outFile.Close()

	if _, err := io.Copy(outFile, body); err != nil {
		log.Err(err).Msg("Failed to write backup to file")
		return err
	}

	log.Info().Msg("Finished writing backup to file")

	return nil
}
