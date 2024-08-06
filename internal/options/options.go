package options

import (
	"context"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"github.com/sethvargo/go-envconfig"
)

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

func LoadConfig() *Options {
	godotenv.Load()

	var config Options

	if err := envconfig.Process(context.Background(), &config); err != nil {
		log.Fatal().Err(err).Msg("Failed to process environment variables")
	}

	config.MongoDB.PrepareMongoDumpOptions()

	if err := config.Validate(); err != nil {
		log.Fatal().Err(err).Msg("Invalid configuration")
	}

	return &config
}
