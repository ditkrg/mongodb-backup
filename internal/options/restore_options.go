package options

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"github.com/sethvargo/go-envconfig"
)

var Restore *RestoreOptions

type RestoreOptions struct {
	MongoRestore MongoRestoreOptions `env:",prefix=MONGO_RESTORE__"`
	S3           S3Options           `env:",prefix=S3__"`
}

func LoadRestoreOptions() {
	envFilePath := os.Getenv("ENV_FILE_PATH")
	if envFilePath == "" {
		envFilePath = fmt.Sprintf("/home/%s/.mongodbBackup/.env", os.Getenv("USER"))
	}

	godotenv.Load(envFilePath, ".env")

	var config RestoreOptions

	if err := envconfig.Process(context.Background(), &config); err != nil {
		log.Fatal().Err(err).Msg("Failed to process environment variables")
	}

	Restore = &config
}
