package config

import (
	"errors"
	"fmt"
	"taskmaster2/service2/internal/pkg/broker/kafkaa"

	"github.com/ilyakaznacheev/cleanenv"
)

var (
	ErrReadingConfig = errors.New("config: failed to load config")
)

type Config struct {
	Kafka kafkaa.Config `yaml:"kafka"`
}

func New(path string) (Config, error) {
	var c Config
	err := cleanenv.ReadConfig(path, &c)
	if err != nil {
		return Config{}, fmt.Errorf("%w: %v", ErrReadingConfig, err)
	}
	return c, nil
}
