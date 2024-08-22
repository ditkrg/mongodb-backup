package helpers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

func WriteToFile(body io.ReadCloser, dir string, fileName string) error {
	defer body.Close()

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	filePath := filepath.Join(dir, fileName)

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
