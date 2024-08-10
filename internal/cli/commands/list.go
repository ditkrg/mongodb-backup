package commands

import (
	"context"
	"fmt"

	"github.com/ditkrg/mongodb-backup/internal/options"
	"github.com/ditkrg/mongodb-backup/internal/services"
)

type List struct {
	Bucket string `required:"" name:"bucket" help:"The S3 bucket to list backups from."`
	Prefix string `optional:"" name:"prefix" help:"The prefix to filter backups."`
}

func (list List) Run() error {

	ctx := context.Background()
	s3Service := services.NewS3Service(options.Restore.S3)

	response := s3Service.List(ctx, list.Bucket, list.Prefix)

	for _, object := range response.Contents {
		fmt.Println(*object.Key)
	}

	return nil
}
