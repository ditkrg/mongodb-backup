package options

import (
	"fmt"
	"time"

	"github.com/ditkrg/mongodb-backup/internal/helpers"
)

type S3Options struct {
	EndPoint    string `env:"ENDPOINT,required"`
	AccessKey   string `env:"ACCESS_KEY,required"`
	SecretKey   string `env:"SECRET_ACCESS_KEY,required"`
	Bucket      string `env:"BUCKET,required"`
	Prefix      string `env:"PREFIX"`
	KeepRecentN int    `env:"KEEP_RECENT_N,default=5"`
}

func (options S3Options) GetOplogConfigDir() string {
	if options.Prefix == "" {
		return "oplog"
	} else {
		return fmt.Sprintf("%s/oplog", options.Prefix)
	}
}

func (options S3Options) CreateNewOplogBackupDir() string {
	timeNow := time.Now().Format("060102-150405")
	parentDir := options.GetParentOplogBackupDir()
	return fmt.Sprintf("%s/%s", parentDir, timeNow)
}

func (options S3Options) GetParentOplogBackupDir() string {
	if options.Prefix == "" {
		return "oplog"
	} else {
		return fmt.Sprintf("%s/oplog", options.Prefix)
	}
}

func (options S3Options) GetOplogConfigFilePath() string {
	fileDir := options.GetOplogConfigDir()

	if options.Prefix == "" {
		return fmt.Sprintf("%s/%s", fileDir, helpers.ConfigFileName)
	} else {
		return fmt.Sprintf("%s/%s", fileDir, helpers.ConfigFileName)
	}
}

func (options S3Options) GetBackupDirPath(databaseName string) string {
	var dirPath string

	if databaseName == "" {
		dirPath = "full_backups"
	} else {
		dirPath = fmt.Sprintf("%s_backups", databaseName)
	}

	if options.Prefix == "" {
		return dirPath
	} else {
		return fmt.Sprintf("%s/%s", options.Prefix, dirPath)
	}
}

func (options S3Options) GetBackupFilePath(databaseName string) string {
	timeNow := time.Now().Format("060102-150405")
	fileDir := options.GetBackupDirPath(databaseName)

	if databaseName == "" {
		return fmt.Sprintf("%s/archive_%s.gzip", fileDir, timeNow)
	} else {
		return fmt.Sprintf("%s/archive_%s.gzip", fileDir, timeNow)
	}
}
