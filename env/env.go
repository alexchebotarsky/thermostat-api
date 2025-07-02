package env

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/alexchebotarsky/thermofridge-api/model/thermofridge"
	"github.com/joho/godotenv"
	envconfig "github.com/sethvargo/go-envconfig"
)

type Config struct {
	LogLevel  slog.Level `env:"LOG_LEVEL,default=debug"`
	LogFormat string     `env:"LOG_FORMAT,default=text"`

	Host string `env:"HOST,default=localhost"`
	Port uint16 `env:"PORT,default=8000"`

	StoragePath string `env:"STORAGE_PATH,default=./storage.db"`

	DefaultMode              thermofridge.Mode `env:"DEFAULT_MODE,default=OFF"`
	DefaultTargetTemperature int               `env:"DEFAULT_TARGET_TEMPERATURE,default=20"`

	PubSubHost     string `env:"PUBSUB_HOST,default=localhost"`
	PubSubPort     uint16 `env:"PUBSUB_PORT,default=1883"`
	PubSubClientID string `env:"PUBSUB_CLIENT_ID,default=thermofridge-api"`
	PubSubQoS      byte   `env:"PUBSUB_QOS,default=1"`
}

func LoadConfig(ctx context.Context) (*Config, error) {
	var c Config

	// We are loading env variables from .env file only for local development
	err := godotenv.Load(".env")
	if err != nil {
		slog.Debug(fmt.Sprintf("error loading .env file: %v", err))
	}

	err = envconfig.Process(ctx, &c)
	if err != nil {
		return nil, fmt.Errorf("error processing environment variables: %v", err)
	}

	return &c, nil
}
