package main

import (
	"os"

	"github.com/ditkrg/mongodb-backup/internal/common"
	"github.com/ditkrg/mongodb-backup/internal/services"
)

func main() {
	backupDir := common.BACKUP_DEFAULT_DIR
	if envBackupDir := common.GetEnv(common.MONGODB__BACKUP_OUT_DIR); envBackupDir != "" {
		backupDir = envBackupDir
	}

	if err := services.StartDatabaseDump(backupDir); err != nil {
		os.Exit(1)
	}

	if err := services.StartBackupUpload(backupDir); err != nil {
		os.Exit(1)
	}
}
