package helpers

import (
	"strings"

	"github.com/rs/zerolog/log"
)

type MongoLogger struct {
}

func (mongoLogger *MongoLogger) Write(message []byte) (int, error) {

	if len(message) == 0 {
		return 0, nil
	}

	cleanString := strings.TrimSpace(string(message))

	if len(cleanString) == 0 {
		return len(message), nil
	}

	log.Log().Msg(cleanString)
	return len(message), nil
}
