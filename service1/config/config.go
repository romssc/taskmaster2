package config

import (
	"errors"
	"fmt"
	"strings"

	"service1/internal/adapter/broker/kafkaa"
	"service1/internal/controller/httprouter"
	"service1/internal/server/httpserver"

	"github.com/spf13/viper"
)

var (
	ErrReadingConfig = errors.New("config: failed to configure")
)

type Config struct {
	Server httpserver.Config `mapstructure:"server"`
	Router httprouter.Config `mapstructure:"router"`
	Kafka  kafkaa.Config     `mapstructure:"kafka"`
}

func New() (Config, error) {
	v := viper.New()

	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", "8081")
	v.SetDefault("kafka.address", []string{"0.0.0.0:9092"})
	v.SetDefault("kafka.topic", "tasks")

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./service1/")

	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("%w: %v", ErrReadingConfig, err)
	}

	v.SetEnvPrefix("")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return Config{}, fmt.Errorf("%w: %v", ErrReadingConfig, err)
	}

	fmt.Println(c)

	return c, nil
}
