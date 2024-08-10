package flags

type VerbosityFlags struct {
	Quiet bool `env:"QUIET" default:"false" help:"when enabled, log output will be suppressed (Default: false)"`
	Level int  `env:"LEVEL" default:"0" help:"log output will be at the specified level (Default: 0)"`
}
