package flags

type VerbosityFlags struct {
	Level int  `env:"LEVEL" default:"5" help:"log output will be at the specified level (Default: 5)"`
	Quiet bool `env:"QUIET" negatable:"" default:"true" help:"hide all log output"`
}
