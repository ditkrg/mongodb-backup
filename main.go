package main

import (
	"os"

	"github.com/ditkrg/mongodb-backup/internal/cli"
	"github.com/ditkrg/mongodb-backup/internal/options"
	"github.com/ditkrg/mongodb-backup/internal/services"
)

func main() {
	args := os.Args

	if len(args) == 2 && args[1] == "--server" {
		options.LoadDumpOptions()
		services.StartBackupProcess()
	} else {
		options.LoadRestoreOptions()
		cli.Run()
	}
}
