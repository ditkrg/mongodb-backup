package services

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

func CompressDir(dir string, fileName string) error {
	log.Info().Msg("Compressing the backup directory")

	// #############################
	// create a zip file
	// #############################
	zipBackupPath := fmt.Sprintf("%s/%s", dir, fileName)

	file, err := os.Create(zipBackupPath)
	if err != nil {
		log.Err(err).Msg("Failed to create a zip file")
		return err
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Err(err).Msg("Failed to close the zip file")
		}
	}()

	// #############################
	// create a new zip archive
	// #############################
	w := zip.NewWriter(file)
	defer func() {
		if err := w.Close(); err != nil {
			log.Err(err).Msg("Failed to close the zip file writer")
		}
	}()

	// #############################
	// add files to the archive
	// #############################
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {

		log.Info().Msgf("Adding file to the archive: %s", path)

		if err != nil {
			log.Err(err).Msg("Failed to walk through the directory")
			return err
		}

		// #############################
		// skip the backup zip file and ignore directories
		// #############################
		if info.IsDir() || zipBackupPath == path {
			return nil
		}

		file, err := os.Open(path)

		if err != nil {
			log.Err(err).Msg("Failed to open file")
			return err
		}

		defer func() {
			if err := file.Close(); err != nil {
				log.Err(err).Msg("Failed to close the file")
			}
		}()

		// #############################
		// create the archive file
		// #############################
		archivePath := path[len(dir)+1:]
		f, err := w.Create(archivePath)

		if err != nil {
			log.Err(err).Msg("Failed to create the archived file")
			return err
		}

		// #############################
		// copy the file to the archive file
		// #############################
		if _, err = io.Copy(f, file); err != nil {
			log.Err(err).Msg("Failed to copy the file to the archive file")
			return err
		}

		log.Info().Msgf("File added to the archive: %s", path)
		return nil
	})

	if err != nil {
		log.Err(err).Msg("Failed to walk through the directory")
	}

	log.Info().Msg("Backup directory compressed")

	return err
}
