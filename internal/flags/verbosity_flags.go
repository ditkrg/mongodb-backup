package flags

type VerbosityFlags struct {
	Quiet bool `env:"QUIET" help:"when enabled, log output will be suppressed"`
	Level int  `env:"LEVEL" default:"0" help:"log output will be at the specified level (Default: 0)"`
}
