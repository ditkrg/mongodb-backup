package helpers

import (
	"fmt"
)

func S3OplogPrefix(prefix string) string {
	if prefix == "" {
		return "oplog/"
	} else {
		return fmt.Sprintf("%s/oplog/", prefix)
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
		return fmt.Sprintf("%s/%s/", prefix, backupKind)
	}
}
