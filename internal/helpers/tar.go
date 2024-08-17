package helpers

import (
	"archive/tar"
	"fmt"
	"io"

	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

func TarDirectory(sourceDirPath string, fileName string) error {

	outputFilePath := fmt.Sprintf("%s/%s", sourceDirPath, fileName)
	outFile, err := os.Create(outputFilePath)

	if err != nil {
		log.Error().Err(err).Msg("could not create tar.gz file")
		return err
	}

	defer outFile.Close()

	tarWriter := tar.NewWriter(outFile)
	defer tarWriter.Close()

	err = filepath.Walk(sourceDirPath, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			log.Err(err).Msg("Failed to walk through the directory")
			return err
		}

		if info.IsDir() || outputFilePath == path {
			return nil
		}

		log.Info().Msgf("Adding %s to %s", path, outputFilePath)

		file, err := os.Open(path)

		if err != nil {
			log.Err(err).Msg("Failed to open file")
			return err
		}

		defer file.Close()

		// Create tar header
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			log.Err(err).Msg("Failed to create tar header")
			return err
		}

		// Ensure the header has the correct name
		header.Name, err = filepath.Rel(sourceDirPath, path)
		if err != nil {
			log.Err(err).Msg("Failed to create tar header")
			return err
		}

		// Write header to tar archive
		if err := tarWriter.WriteHeader(header); err != nil {
			log.Err(err).Msg("Failed to create tar header")
			return err
		}

		if _, err = io.Copy(tarWriter, file); err != nil {
			log.Err(err).Msg("Failed to copy the file to the archive file")
			return err
		}

		log.Info().Msgf("File %s added to %s", path, outputFilePath)
		return nil
	})

	if err != nil {
		log.Error().Err(err).Msg("Failed to archive the directory")
		return err
	}

	log.Info().Msgf("Directory %s archived to %s", sourceDirPath, outputFilePath)
	return nil
}
