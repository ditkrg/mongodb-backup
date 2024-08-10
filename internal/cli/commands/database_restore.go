package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/ditkrg/mongodb-backup/internal/cli/helpers"
	"github.com/ditkrg/mongodb-backup/internal/options"
	"github.com/ditkrg/mongodb-backup/internal/services"
)

type DatabaseRestore struct {
	Bucket string `required:"" name:"bucket" help:"The S3 bucket to list backups from."`
	Prefix string `optional:"" name:"prefix" help:"The prefix to filter backups."`
	Key    string `optional:"" name:"key" help:"The key of the backup to restore."`
}

func (restore DatabaseRestore) Run() error {
	ctx := context.Background()
	s3Service := services.NewS3Service(options.Restore.S3)

	if restore.Key == "" {
		restore.Key = helpers.ChooseDatabaseToRestore(s3Service, ctx, restore.Bucket, restore.Prefix, func(key string) bool {
			return !strings.Contains(key, "oplog") && !strings.Contains(key, "full_backups")
		})
	}

	fmt.Println(restore.Key)

	return nil
}
