package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ditkrg/mongodb-backup/internal/common"
	"github.com/ditkrg/mongodb-backup/internal/services"
)

func main() {
	// ######################
	// Check env variables
	// ######################
	checkEnvVariables()

	// ######################
	// prepare file name
	// ######################
	fileName := fmt.Sprintf("%s_%s.zip", common.GetRequiredEnv(common.MONGODB__DB_TO_BACKUP), time.Now().Format("060102-150405"))

	// ######################
	// prepare backup dir path
	// ######################
	backupDir := common.BACKUP_DEFAULT_DIR
	if envBackupDir := common.GetEnv(common.MONGODB__BACKUP_OUT_DIR); envBackupDir != "" {
		envBackupDir, _ := strings.CutSuffix(envBackupDir, "/")
		backupDir = envBackupDir
	}

	// ######################
	// dump database
	// ######################
	if err := services.StartDatabaseDump(backupDir); err != nil {
		os.Exit(1)
	}

	// ######################
	// compress backup
	// ######################
	if err := services.CompressDir(backupDir, fileName); err != nil {
		os.Exit(1)
	}

	// ######################
	// upload backup to S3
	// ######################
	if err := services.StartBackupUpload(backupDir, fileName); err != nil {
		os.Exit(1)
	}
}

func checkEnvVariables() {
	common.GetRequiredEnv(common.S3__ACCESS_KEY)
	common.GetRequiredEnv(common.S3__SECRET_ACCESS_KEY)
	common.GetRequiredEnv(common.S3__ENDPOINT)
	common.GetRequiredEnv(common.S3__BUCKET)
	common.GetRequiredEnv(common.S3__JOB_DIR)

	common.GetRequiredEnv(common.MONGODB__TLS_CA_CERT_PATH)
	common.GetRequiredEnv(common.MONGODB__USERNAME)
	common.GetRequiredEnv(common.MONGODB__PASSWORD)
	common.GetRequiredEnv(common.MONGODB__MECHANISM)
	common.GetRequiredEnv(common.MONGODB__AUTH_SOURCE)
	common.GetRequiredEnv(common.MONGODB__DB_TO_BACKUP)
	common.GetRequiredEnv(common.MONGODB__REPLICA_SET)
	common.GetRequiredEnv(common.MONGODB__HOSTS)

	common.GetIntEnv(common.S3__KEEP_RESENT_N, 10)
	common.GetBoolEnv(common.MONGODB__USE_SSL, true)
}
