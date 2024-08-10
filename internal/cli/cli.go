package cli

import (
	"github.com/alecthomas/kong"
	"github.com/ditkrg/mongodb-backup/internal/cli/commands"
)

type CLI struct {
	Version kong.VersionFlag `short:"v" help:"Print the version number"`

	List commands.List `cmd:"" help:"List backups"`

	DatabaseRestore commands.DatabaseRestore `cmd:"" name:"database-restore" help:"Restore a single database"`
	PitrRestore     commands.PitrRestore     `cmd:"" name:"pitr-restore" help:"Do a point-in-time restore"`
	FullRestore     commands.FullRestore     `cmd:"" name:"full-restore" help:"Do a full restore"`
}

func Run() {

	cli := &CLI{}

	ctx := kong.Parse(
		cli,
		kong.Description("A CLI tool for MongoDB backups restore."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact:             true,
			Summary:             true,
			NoExpandSubcommands: true,
		}),
		kong.Vars{
			"version":                 "0.0.1",
			"defaultHeight":           "0",
			"defaultWidth":            "0",
			"defaultAlign":            "left",
			"defaultBorder":           "none",
			"defaultBorderForeground": "",
			"defaultBorderBackground": "",
			"defaultBackground":       "",
			"defaultForeground":       "",
			"defaultMargin":           "0 0",
			"defaultPadding":          "0 0",
			"defaultUnderline":        "false",
			"defaultBold":             "false",
			"defaultFaint":            "false",
			"defaultItalic":           "false",
			"defaultStrikethrough":    "false",
		},
	)

	ctx.FatalIfErrorf(ctx.Run())
}
