package options

import (
	"fmt"
	"time"
)

type S3Options struct {
	EndPoint    string `env:"ENDPOINT,required"`
	AccessKey   string `env:"ACCESS_KEY,required"`
	SecretKey   string `env:"SECRET_ACCESS_KEY,required"`
	Bucket      string `env:"BUCKET,required"`
	Prefix      string `env:"PREFIX"`
	KeepRecentN int    `env:"KEEP_RECENT_N,default=5"`
}

func (options S3Options) OplogDir() string {
	if options.Prefix == "" {
		return "oplog"
	} else {
		return fmt.Sprintf("%s/oplog", options.Prefix)
	}
}

func (options S3Options) BackupDirPath(databaseName string) string {
	var dirPath string

	if databaseName == "" {
		dirPath = "full_backups"
	} else {
		dirPath = fmt.Sprintf("%s_database_backups", databaseName)
	}

	if options.Prefix == "" {
		return dirPath
	} else {
		return fmt.Sprintf("%s/%s", options.Prefix, dirPath)
	}
}

func (options S3Options) BackupFilePath(databaseName string) string {
	timeNow := time.Now().Format("060102-150405")
	fileDir := options.BackupDirPath(databaseName)
	return fmt.Sprintf("%s/archive_%s.gzip", fileDir, timeNow)
}
