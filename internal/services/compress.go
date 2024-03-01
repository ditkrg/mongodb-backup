package services

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

func CompressDir(dir string) error {
	// #############################
	// create a zip file
	// #############################
	zipBackupPath := fmt.Sprintf("%s/backup.zip", dir)

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

		return nil
	})

	if err != nil {
		log.Err(err).Msg("Failed to walk through the directory")
	}

	return err
}
