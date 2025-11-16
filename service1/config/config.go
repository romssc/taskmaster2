package config

import (
	"errors"
	"fmt"

	"service1/internal/adapter/broker/kafkaa"
	"service1/internal/pkg/server/httpserver"
	"service1/internal/usecase/create"
	"service1/internal/usecase/list"
	"service1/internal/usecase/listid"

	"github.com/ilyakaznacheev/cleanenv"
)

var (
	ErrReadingConfig = errors.New("config: failed to load config")
)

type Config struct {
	Server httpserver.Config `yaml:"server"`
	Router Router            `yaml:"router"`
	Kafka  kafkaa.Config     `yaml:"kafka"`
}

type Router struct {
	Create create.Config `yaml:"create"`
	List   list.Config   `yaml:"list"`
	ListID listid.Config `yaml:"list_id"`
}

func New(path string) (Config, error) {
	var c Config
	err := cleanenv.ReadConfig(path, &c)
	if err != nil {
		return Config{}, fmt.Errorf("%w: %v", ErrReadingConfig, err)
	}
	return c, nil
}
