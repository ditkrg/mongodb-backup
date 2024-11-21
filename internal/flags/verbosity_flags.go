package flags

import (
	"fmt"

	mongoLog "github.com/mongodb/mongo-tools/common/log"
	"github.com/mongodb/mongo-tools/common/options"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type VerbosityFlags struct {
	Level int  `env:"LEVEL" default:"1" help:"log output will be at the specified level (between 1 and 3) the higher the number the more verbose"`
	Quiet bool `env:"QUIET" help:"hide all log output"`
}

func (v *VerbosityFlags) SetGlobalLogLevel() {
	var err error
	var level zerolog.Level

	if level, err = v.parse(); err != nil {
		log.Fatal().Err(err).Send()
	}

	zerolog.SetGlobalLevel(level)
	mongoLog.SetVerbosity(options.Verbosity{
		VLevel: v.Level,
		Quiet:  v.Quiet,
	})
	mongoLog.SetDateFormat("")
}

func (v *VerbosityFlags) parse() (zerolog.Level, error) {

	if v.Quiet {
		return zerolog.Disabled, nil
	}

	switch v.Level {
	case 1:
		return zerolog.InfoLevel, nil
	case 2:
		return zerolog.DebugLevel, nil
	case 3:
		return zerolog.TraceLevel, nil

	default:
		return 0, fmt.Errorf("invalid log level: %d", v.Level)
	}
}
