package cli

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/ditkrg/mongodb-backup/internal/cli/commands"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type CLI struct {
	Version kong.VersionFlag `short:"v" help:"Print the version number"`

	PitrRestore     commands.PitrRestore     `cmd:"" name:"pitr-restore" help:"Do a point-in-time restore"`
	DatabaseRestore commands.DatabaseRestore `cmd:"" name:"database-restore" help:"Restore a backup"`
}

func Run() {
	envFilePath := os.Getenv("ENV_FILE_PATH")
	if envFilePath == "" {
		envFilePath = fmt.Sprintf("%s/.mongoCli/.env", os.Getenv("HOME"))
	}

	godotenv.Load(envFilePath, ".env")

	cli := &CLI{}

	ctx := kong.Parse(
		cli,
		kong.Description("A CLI tool for MongoDB backups restore."),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
		kong.Vars{
			"version": "0.0.1",
		},
	)

	if err := ctx.Run(); err != nil {
		log.Fatal().Err(err).Msg("Failed to run the CLI")
	}
}
