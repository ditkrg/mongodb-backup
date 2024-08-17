package helpers

import (
	"fmt"
	"time"
)

func S3OplogPrefix(prefix string) string {
	if prefix == "" {
		return "oplog"
	} else {
		return fmt.Sprintf("%s/oplog", prefix)
	}
}

func S3BackupPrefix(prefix string, databaseName string) string {
	var backupKind string

	if databaseName == "" {
		backupKind = "full_backups"
	} else {
		backupKind = fmt.Sprintf("%s_database_backups", databaseName)
	}

	if prefix == "" {
		return backupKind
	} else {
		return fmt.Sprintf("%s/%s", prefix, backupKind)
	}
}

func S3FileKey(archive bool, gzip bool) string {
	timeNow := time.Now().Format("060102-150405")
	return fileWithSuffix(timeNow, archive, gzip)
}

func fileWithSuffix(fileName string, archive bool, gzip bool) string {
	fileNameWithSuffix := fileName

	if archive {
		fileNameWithSuffix = fmt.Sprintf("%s.archive", fileName)
	}

	if gzip {
		fileNameWithSuffix = fmt.Sprintf("%s.gzip", fileNameWithSuffix)
	}

	return fileNameWithSuffix
}
