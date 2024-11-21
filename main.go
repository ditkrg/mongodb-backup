package main

import (
	"github.com/alecthomas/kong"
	"github.com/ditkrg/mongodb-backup/internal/commands"
	"github.com/ditkrg/mongodb-backup/internal/helpers"
	"github.com/joho/godotenv"
	mongoLog "github.com/mongodb/mongo-tools/common/log"
	"github.com/rs/zerolog/log"
)

type CLI struct {
	Version kong.VersionFlag `short:"v" help:"Print the version number"`

	List    commands.ListCommand            `cmd:"" name:"list" help:"List backups"`
	Dump    commands.DumpCommand            `cmd:"" name:"dump" help:"Take a database or point-in-time backup"`
	Restore commands.DatabaseRestoreCommand `cmd:"" name:"restore" help:"Restore a Database"`
}

func main() {
	// #############################
	// Load environment variables
	// #############################
	godotenv.Load(".env")

	// #############################
	// Set global MongoDB log writer
	// #############################
	mongoLog.SetWriter(&helpers.MongoLogger{})

	// #############################
	// Prepare CLI
	// #############################
	cli := &CLI{}

	ctx := kong.Parse(
		cli,
		kong.Description("A CLI tool for MongoDB backups restore."),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
		kong.Vars{
			"version": "0.1.0",
		},
	)

	// #############################
	// Run the command
	// #############################
	if err := ctx.Run(); err != nil {
		log.Fatal().Err(err).Msg("Failed to run the command")
	}
}
