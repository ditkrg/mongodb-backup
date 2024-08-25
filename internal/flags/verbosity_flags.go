package flags

type VerbosityFlags struct {
	Level string `env:"LEVEL" default:"5" help:"log output will be at the specified level (Default: 5)"`
}
