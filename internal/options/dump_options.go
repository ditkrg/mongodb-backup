package options

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"github.com/sethvargo/go-envconfig"
)

var Dump *DumpOptions

type DumpOptions struct {
	MongoDump MongoDumpOptions `env:",prefix=MONGO_DUMP__"`
	S3        S3Options        `env:",prefix=S3__"`
}

func LoadDumpOptions() {
	envFilePath := os.Getenv("ENV_FILE_PATH")
	if envFilePath == "" {
		envFilePath = fmt.Sprintf("/home/%s/.mongodbBackup/.env", os.Getenv("USER"))
	}

	godotenv.Load(envFilePath, ".env")

	var config DumpOptions

	if err := envconfig.Process(context.Background(), &config); err != nil {
		log.Fatal().Err(err).Msg("Failed to process environment variables")
	}

	config.MongoDump.PrepareMongoDumpOptions()

	if err := config.MongoDump.Validate(); err != nil {
		log.Fatal().Err(err).Msg("Invalid configuration")
	}

	Dump = &config
}
