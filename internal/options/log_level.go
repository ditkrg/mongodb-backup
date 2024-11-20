package options

import (
	"fmt"

	"github.com/rs/zerolog"
)

const (
	debugLogLevel   = "Debug"
	infoLogLevel    = "Info"
	warnLogLevel    = "Warn"
	errorLogLevel   = "Error"
	fatalLogLevel   = "Fatal"
	panicLogLevel   = "Panic"
	disableLogLevel = "Disable"
	traceLogLevel   = "Trace"
)

type LogLevel struct {
	Level string `env:"LOG_LEVEL,default=Info"`
}

func (l *LogLevel) Parse() (zerolog.Level, error) {

	switch l.Level {
	case debugLogLevel:
		return zerolog.DebugLevel, nil
	case infoLogLevel:
		return zerolog.InfoLevel, nil
	case warnLogLevel:
		return zerolog.WarnLevel, nil
	case errorLogLevel:
		return zerolog.ErrorLevel, nil
	case fatalLogLevel:
		return zerolog.FatalLevel, nil
	case panicLogLevel:
		return zerolog.PanicLevel, nil
	case disableLogLevel:
		return zerolog.Disabled, nil
	case traceLogLevel:
		return zerolog.TraceLevel, nil
	default:
		return 0, fmt.Errorf("invalid log level: %s", l.Level)
	}
}
