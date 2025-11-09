package config

import (
	"errors"
	"fmt"
	"taskmaster2/internal/adapter/storage/sqlite3"
	"taskmaster2/pkg/httpserver"

	"github.com/ilyakaznacheev/cleanenv"
)

var (
	ErrReadingConfig = errors.New("config: failed to load config")
)

type Config struct {
	Server  httpserver.Config `yaml:"server"`
	SQLite3 sqlite3.Config    `yaml:"sqlite3"`
}

func New(path string) (Config, error) {
	var c Config
	err := cleanenv.ReadConfig(path, &c)
	if err != nil {
		return Config{}, fmt.Errorf("%w: %v", ErrReadingConfig, err)
	}
	return c, nil
}
