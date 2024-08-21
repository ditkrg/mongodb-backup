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
	log.Info().Msgf("adding directory %s to %s", sourceDirPath, fileName)

	outputFilePath := sourceDirPath + fileName
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

	log.Info().Msgf("Directory %s added to %s", sourceDirPath, outputFilePath)
	return nil
}

func ExtractTar(tarPath, destDir string) error {
	log.Info().Msgf("Extracting %s to %s", tarPath, destDir)

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Open the tar file
	file, err := os.Open(tarPath)
	if err != nil {
		return fmt.Errorf("failed to open tar file: %w", err)
	}
	defer file.Close()

	// Create a tar reader
	tarReader := tar.NewReader(file)

	// Iterate through the files in the tar archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of tar archive
		}
		if err != nil {
			return fmt.Errorf("failed to read tar file: %w", err)
		}

		// Get the individual file path and extract it
		targetPath := filepath.Join(destDir, header.Name)

		// Handle directories
		switch header.Typeflag {
		case tar.TypeDir:
			// Create a directory
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			// Create a file
			outFile, err := os.Create(targetPath)
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			defer outFile.Close()

			// Copy the file content from the tar archive
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}

			// Restore file permissions
			if err := os.Chmod(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to set file permissions: %w", err)
			}
		default:
			// Skip unsupported file types
			continue
		}
	}

	log.Info().Msgf("Extracted %s to %s", tarPath, destDir)
	return nil
}
