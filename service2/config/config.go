package config

import (
	"errors"
	"fmt"
	"taskmaster2/service2/internal/pkg/broker/kafkaa"
	"taskmaster2/service2/internal/usecase/update"

	"github.com/ilyakaznacheev/cleanenv"
)

var (
	ErrReadingConfig = errors.New("config: failed to load config")
)

type Config struct {
	Kafka  kafkaa.Config `yaml:"kafka"`
	Router Router        `yaml:"router"`
}

type Router struct {
	Update update.Config `yaml:"update"`
}

func New(path string) (Config, error) {
	var c Config
	err := cleanenv.ReadConfig(path, &c)
	if err != nil {
		return Config{}, fmt.Errorf("%w: %v", ErrReadingConfig, err)
	}
	return c, nil
}
