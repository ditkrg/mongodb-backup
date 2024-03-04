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

func GetIntEnv(variable string, defaultValue int) int {

	environmentVariable := os.Getenv(variable)

	if environmentVariable == "" {
		return defaultValue
	}

	value, err := strconv.ParseInt(environmentVariable, 10, 64)
	if err != nil {
		err := fmt.Errorf("could not prase '%s' to int", environmentVariable)
		zerolog.Panic().Err(err).Msg("failed to parse env variable to int")
	}

	if value < 0 {
		err := fmt.Errorf("'%s' should be a positive number", environmentVariable)
		zerolog.Panic().Err(err).Msg("negative value for env variable")
	}

	return int(value)
}

func GetBoolEnv(variable string, defaultValue bool) bool {

	environmentVariable := os.Getenv(variable)

	if environmentVariable == "" {
		return defaultValue
	}

	boolVal, err := strconv.ParseBool(environmentVariable)
	if err != nil {
		err := fmt.Errorf("could not prase '%s' to bool", environmentVariable)
		zerolog.Panic().Err(err).Msg("failed to parse env variable to bool")
	}

	return boolVal
}
