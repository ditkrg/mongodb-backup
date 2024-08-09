package options

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"github.com/sethvargo/go-envconfig"
)

var Config *Options

type Options struct {
	MongoDB MongoDBOptions `env:",prefix=MONGODB__"`
	S3      S3Options      `env:",prefix=S3__"`
}

func (o *Options) Validate() error {
	if err := o.MongoDB.Validate(); err != nil {
		return err
	}

	return nil
}

func LoadConfig() {

	envFilePath := os.Getenv("ENV_FILE_PATH")
	if envFilePath == "" {
		envFilePath = fmt.Sprintf("/home/%s/.mongodbBackup/.env", os.Getenv("USER"))
	}

	godotenv.Load(envFilePath, ".env")

	var config Options

	if err := envconfig.Process(context.Background(), &config); err != nil {
		log.Fatal().Err(err).Msg("Failed to process environment variables")
	}

	config.MongoDB.PrepareMongoDumpOptions()

	if err := config.Validate(); err != nil {
		log.Fatal().Err(err).Msg("Invalid configuration")
	}

	Config = &config
}
