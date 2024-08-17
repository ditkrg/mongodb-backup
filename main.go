package main

import (
	"context"

	"github.com/alecthomas/kong"
	"github.com/ditkrg/mongodb-backup/internal/commands"
	"github.com/ditkrg/mongodb-backup/internal/options"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sethvargo/go-envconfig"
)

type CLI struct {
	Version kong.VersionFlag `short:"v" help:"Print the version number"`

	Restore struct {
		List        commands.ListCommand            `cmd:"" name:"list" help:"List backups"`
		Database    commands.DatabaseRestoreCommand `cmd:"" name:"database" help:"Restore a backup"`
		PitrRestore commands.PitrRestoreCommand     `cmd:"" name:"pitr" help:"Perform a point-in-time restore"`
	} `cmd:""`

	Dump commands.DumpCommand `cmd:"" name:"dump" help:"Take a database or point-in-time backup"`
}

func main() {
	godotenv.Load(".env")
	setGlobalLogLevel()

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
		log.Fatal().Err(err).Msg("Failed to run the command")
	}
}

func setGlobalLogLevel() {
	var c options.LogLevel
	var err error
	var level zerolog.Level

	if err = envconfig.Process(context.Background(), &c); err != nil {
		log.Fatal().Err(err).Send()
	}

	if level, err = c.Parse(); err != nil {
		log.Fatal().Err(err).Send()
	}

	zerolog.SetGlobalLevel(level)
}
