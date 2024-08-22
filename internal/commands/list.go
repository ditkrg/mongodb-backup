package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/charmbracelet/lipgloss/list"
	"github.com/ditkrg/mongodb-backup/internal/flags"
	"github.com/ditkrg/mongodb-backup/internal/helpers"
	"github.com/ditkrg/mongodb-backup/internal/services"
	"github.com/rs/zerolog/log"
)

type ListCommand struct {
	S3          flags.S3Flags `embed:"" group:"Common S3 Flags:"`
	Oplog       bool          `required:"" xor:"list" help:"List oplog backups"`
	FullBackups bool          `required:"" xor:"list" help:"List full backups"`
	Database    string        `required:"" xor:"list" help:"List backups for a specific database"`
}

func (command *ListCommand) Run() error {
	var resp *s3.ListObjectsV2Output
	var prefix string
	var err error

	ctx := context.Background()
	s3Service := services.NewS3Service(command.S3)

	if command.Oplog {
		prefix = helpers.S3OplogPrefix(command.S3.Prefix)
	} else if command.FullBackups {
		prefix = helpers.S3BackupPrefix(command.S3.Prefix, "")
	} else {
		prefix = helpers.S3BackupPrefix(command.S3.Prefix, command.Database)
	}

	if resp, err = s3Service.List(ctx, command.S3.Bucket, prefix); err != nil {
		return err
	}

	if len(resp.Contents) == 0 {
		message := "No backups found"
		log.Info().Msg(message)
		fmt.Println(message)
		return nil
	}

	list := list.New()

	for _, object := range resp.Contents {
		key := *object.Key
		if strings.Contains(key, helpers.ConfigFileName) {
			continue
		}

		if command.Oplog {
			list = list.Item(FormatOplogTime(key, command.S3.Prefix))
		} else {
			list = list.Item(key)
		}
	}

	fmt.Println("List of Available Backups:")
	fmt.Println(list)

	return nil

}

func FormatOplogTime(key string, prefix string) string {
	key = strings.TrimSuffix(key, ".tar.gz")
	key = strings.TrimPrefix(key, helpers.S3OplogPrefix(prefix))

	fromString := strings.Split(key, "_")[0]
	toString := strings.Split(key, "_")[1]

	fromTime, err := time.Parse(helpers.TimeFormat, fromString)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to parse time")
	}

	toTime, err := time.Parse(helpers.TimeFormat, toString)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to parse time")
	}

	return fmt.Sprintf("%s ~ %s", fromTime.Format(helpers.HumanReadableTimeFormat), toTime.Format(helpers.HumanReadableTimeFormat))
}
