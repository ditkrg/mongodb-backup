package options

type Verbosity struct {
	Quiet bool `env:"QUIET,default=false"`
	Level int  `env:"LEVEL,default=0"`
}
