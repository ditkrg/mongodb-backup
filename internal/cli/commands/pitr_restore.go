package commands

type PitrRestore struct {
	Bucket    string `required:"" name:"bucket" help:"The S3 bucket to list backups from."`
	Prefix    string `required:"" name:"prefix" help:"The prefix to list backups from."`
	StartTime string `required:"" name:"start-time" help:"The start time of the PITR restore."`
	EndTime   string `required:"" name:"end-time" help:"The end time of the PITR restore."`
}

func (c *PitrRestore) Run() error {
	return nil
}
