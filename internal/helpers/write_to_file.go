package helpers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog/log"
)

func WriteToFile(body io.ReadCloser, contentLength *int64, dir string, fileName string) error {
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

	pw := &progressWriter{
		total:  int(*contentLength),
		file:   outFile,
		reader: body,
		onProgress: func(ratio float64) {
			teaProgram.Send(progressMsg(ratio))
		},
	}

	m := model{
		pw:       pw,
		progress: progress.New(progress.WithDefaultGradient()),
	}

	teaProgram = tea.NewProgram(m, tea.WithInput(nil), tea.WithoutSignalHandler())

	go pw.Start()

	if _, err := teaProgram.Run(); err != nil {
		log.Err(err).Msg("Failed to write backup to file")
		return err
	}

	log.Info().Msg("Finished writing backup to file")

	return nil
}
