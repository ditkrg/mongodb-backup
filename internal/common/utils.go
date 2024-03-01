package common

import (
	"fmt"
	"os"
	"strconv"

	zerolog "github.com/rs/zerolog/log"
)

func GetRequiredEnv(variable string) string {

	environmentVariable := os.Getenv(variable)

	if environmentVariable == "" {
		err := fmt.Errorf("%s is not set", variable)
		zerolog.Panic().Err(err).Msg("missing env variable")
	}

	return environmentVariable
}

func GetEnv(variable string) string {
	environmentVariable := os.Getenv(variable)
	return environmentVariable
}

func GetBoolEnv(variable string, defaultValue bool) bool {

	environmentVariable := os.Getenv(variable)

	if environmentVariable == "" {
		return defaultValue
	}

	boolVal, err := strconv.ParseBool(environmentVariable)
	if err != nil {
		err := fmt.Errorf("%s is not set", variable)
		zerolog.Panic().Err(err).Msg("missing env variable")
	}

	return boolVal
}
